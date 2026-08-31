#ifndef SCRIPTGO_SQLITE_TYPES_H
#include "sqlite_types.h"
#endif

int scriptgo_sqlite_prepare(void *db_handle, const char *sql, void **out_stmt) {
    if (out_stmt == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }
    if (sql == NULL) return sqlite_fail("sqlite prepare: SQL string required", NULL);

    sqlite3_stmt *compiled_stmt = NULL;
    int rc = sqlite3_prepare_v2(db->db, sql, -1, &compiled_stmt, NULL);
    if (rc != SQLITE_OK) {
        char detail[512];
        snprintf(detail, sizeof(detail), "%s (in %s)", sqlite3_errmsg(db->db), sql);
        return sqlite_fail("sqlite prepare failed", detail);
    }

    scriptgo_sqlite_stmt_t *stmt = calloc(1, sizeof(*stmt));
    if (stmt == NULL) {
        sqlite3_finalize(compiled_stmt);
        return sqlite_fail("sqlite: allocation failed", NULL);
    }
    stmt->magic = SCRIPTGO_MAGIC_SQLITE_STMT;
    stmt->stmt = compiled_stmt;
    stmt->db = db;
    stmt->sql = strdup(sql);
    stmt->allow_bare_named = 0;
    stmt->allow_unknown_named = 0;
    stmt->return_arrays = 0;
    stmt->read_bigints = 0;

    /* Link into db->stmts */
    stmt->next = db->stmts;
    if (db->stmts != NULL) {
        db->stmts->prev = stmt;
    }
    db->stmts = stmt;

    scriptgo_gc_register(stmt, 21, 0);
    *out_stmt = stmt;
    return 0;
}

int scriptgo_sqlite_finalize(void *stmt_handle) {
    if (stmt_handle == NULL) return 0;
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT) return 0;

    if (stmt->stmt != NULL) {
        sqlite3_finalize(stmt->stmt);
        stmt->stmt = NULL;
    }

    /* Unlink from db list */
    if (stmt->db != NULL) {
        if (stmt->prev != NULL) {
            stmt->prev->next = stmt->next;
        } else if (stmt->db->stmts == stmt) {
            stmt->db->stmts = stmt->next;
        }
        if (stmt->next != NULL) {
            stmt->next->prev = stmt->prev;
        }
    }

    if (stmt->sql != NULL) {
        free(stmt->sql);
        stmt->sql = NULL;
    }
    stmt->magic = 0;
    return 0;
}

int scriptgo_sqlite_stmt_config(void *stmt_handle, double bare, double unk, double ret_arr, double read_bi) {
    if (stmt_handle == NULL) return 0;
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT) return 0;
    if (bare >= 0) stmt->allow_bare_named = (int)bare;
    if (unk >= 0) stmt->allow_unknown_named = (int)unk;
    if (ret_arr >= 0) stmt->return_arrays = (int)ret_arr;
    if (read_bi >= 0) stmt->read_bigints = (int)read_bi;
    return 0;
}

int scriptgo_sqlite_bind_param(sqlite3_stmt *stmt, int idx, uint32_t tag, uint64_t payload) {
    if (tag == 1 || tag == 0) { /* null or undefined */
        return sqlite3_bind_null(stmt, idx);
    }
    if (tag == 2) { /* boolean */
        return sqlite3_bind_int(stmt, idx, (payload & 1) ? 1 : 0);
    }
    if (tag == 3) { /* number */
        double val = 0;
        memcpy(&val, &payload, sizeof(val));
        if (floor(val) == val && val >= -9007199254740991.0 && val <= 9007199254740991.0) {
            return sqlite3_bind_int64(stmt, idx, (sqlite3_int64)val);
        }
        return sqlite3_bind_double(stmt, idx, val);
    }
    if (tag == 4) { /* string */
        const char *str = (const char *)(uintptr_t)payload;
        return sqlite3_bind_text(stmt, idx, str ? str : "", -1, SQLITE_TRANSIENT);
    }
    if (tag == 7) { /* bigint */
        int64_t val = (int64_t)payload;
        return sqlite3_bind_int64(stmt, idx, (sqlite3_int64)val);
    }
    if (tag == 5 || tag == 6) { /* object or buffer */
        void *ptr = (void *)(uintptr_t)payload;
        if (ptr != NULL) {
            uint32_t magic = *(uint32_t *)ptr;
            if (magic == SCRIPTGO_MAGIC_BUFFER || magic == SCRIPTGO_MAGIC_TYPEDARRAY) {
                scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)ptr;
                return sqlite3_bind_blob(stmt, idx, view->data, (int)view->length, SQLITE_TRANSIENT);
            }
        }
    }
    return sqlite3_bind_null(stmt, idx);
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
} scriptgo_sqlite_array_t;

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} scriptgo_sqlite_object_t;

static int find_object_property_index(const scriptgo_sqlite_object_t *obj, const char *prop) {
    if (obj == NULL || obj->type_name == NULL || prop == NULL) return -1;
    const char *t = obj->type_name;
    size_t plen = strlen(prop);

    if (strncmp(t, "__shape_", 8) == 0) {
        const char *cur = t + 8;
        int idx = 0;
        while (*cur != '\0') {
            const char *under1 = strchr(cur, '_');
            if (under1 == NULL) break;
            size_t name_len = (size_t)(under1 - cur);
            if (name_len == plen && strncmp(cur, prop, plen) == 0) {
                return idx;
            }
            const char *type_start = under1 + 1;
            const char *under2 = strchr(type_start, '_');
            if (under2 == NULL) break;
            cur = under2 + 1;
            idx++;
        }
        return -1;
    }

    const char *cur = t;
    if (*cur == ':') cur++;
    int idx = 0;
    while (*cur != '\0') {
        const char *colon = strchr(cur, ':');
        size_t name_len = colon ? (size_t)(colon - cur) : strlen(cur);
        if (name_len > 0) {
            if (name_len == plen && strncmp(cur, prop, plen) == 0) {
                return idx;
            }
            idx++;
        }
        if (colon == NULL) break;
        cur = colon + 1;
    }
    return -1;
}

int scriptgo_sqlite_bind_all(scriptgo_sqlite_stmt_t *stmt, void *first_param, void *rest_params) {
    if (stmt == NULL || stmt->stmt == NULL) return 0;
    sqlite3_reset(stmt->stmt);
    sqlite3_clear_bindings(stmt->stmt);

    if (first_param == NULL && rest_params == NULL) return 0;

    int next_idx = 1;

    /* 1. Check first_param */
    if (first_param != NULL) {
        scriptgo_sqlite_unknown_t *boxed = (scriptgo_sqlite_unknown_t *)first_param;
        if (boxed->tag == 5) { /* Object: could be named params object */
            void *ptr = (void *)(uintptr_t)boxed->payload;
            if (ptr != NULL) {
                uint64_t *p64 = (uint64_t *)ptr;
                uint32_t *p32 = (uint32_t *)ptr;
                if (*p64 == SCRIPTGO_OBJECT_MAGIC && *p32 != SCRIPTGO_MAGIC_BUFFER && *p32 != SCRIPTGO_MAGIC_TYPEDARRAY) {
                    scriptgo_sqlite_object_t *obj = (scriptgo_sqlite_object_t *)ptr;
                    int param_count = sqlite3_bind_parameter_count(stmt->stmt);
                    if (param_count > 0) {
                        for (int p = 1; p <= param_count; p++) {
                            const char *pname = sqlite3_bind_parameter_name(stmt->stmt, p);
                            if (pname != NULL) {
                                const char *clean_name = pname;
                                if (*clean_name == ':' || *clean_name == '@' || *clean_name == '$') {
                                    clean_name++;
                                }
                                int field_idx = find_object_property_index(obj, clean_name);
                                if (field_idx >= 0) {
                                    uint32_t tag = 0;
                                    uint64_t payload = 0;
                                    scriptgo_object_unknown_get(ptr, field_idx, &tag, &payload);
                                    scriptgo_sqlite_bind_param(stmt->stmt, p, tag, payload);
                                }
                            }
                        }
                    }
                } else {
                    scriptgo_sqlite_bind_param(stmt->stmt, next_idx++, boxed->tag, boxed->payload);
                }
            }
        } else {
            scriptgo_sqlite_bind_param(stmt->stmt, next_idx++, boxed->tag, boxed->payload);
        }
    }

    /* 2. Rest positional parameters from array */
    if (rest_params != NULL) {
        scriptgo_sqlite_array_t *arr = (scriptgo_sqlite_array_t *)rest_params;
        for (int64_t i = 0; i < arr->length; i++) {
            if (arr->element_size == sizeof(scriptgo_sqlite_unknown_t)) {
                scriptgo_sqlite_unknown_t *elem = (scriptgo_sqlite_unknown_t *)(arr->data + i * arr->element_size);
                scriptgo_sqlite_bind_param(stmt->stmt, next_idx++, elem->tag, elem->payload);
            } else if (arr->element_size == sizeof(double)) {
                double *val = (double *)(arr->data + i * arr->element_size);
                sqlite3_bind_double(stmt->stmt, next_idx++, *val);
            } else if (arr->element_size == sizeof(char *)) {
                char **str = (char **)(arr->data + i * arr->element_size);
                sqlite3_bind_text(stmt->stmt, next_idx++, *str ? *str : "", -1, SQLITE_TRANSIENT);
            }
        }
    }
    return 0;
}

int scriptgo_sqlite_expanded_sql(void *stmt_handle, const char **out_sql) {
    if (out_sql == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (stmt_handle == NULL) return sqlite_fail("sqlite: invalid statement handle", NULL);
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT || stmt->stmt == NULL) {
        *out_sql = "";
        return 0;
    }
    char *expanded = sqlite3_expanded_sql(stmt->stmt);
    if (expanded != NULL) {
        *out_sql = expanded;
    } else {
        *out_sql = stmt->sql ? stmt->sql : "";
    }
    return 0;
}

int scriptgo_sqlite_run(void *stmt_handle, void *first_param, void *rest_params, void **out_result) {
    if (out_result == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (stmt_handle == NULL) return sqlite_fail("sqlite: invalid statement handle", NULL);
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT || stmt->stmt == NULL) {
        return sqlite_fail("sqlite: statement is not prepared or closed", NULL);
    }

    if (scriptgo_sqlite_bind_all(stmt, first_param, rest_params) != 0) {
        return -1;
    }

    int rc = sqlite3_step(stmt->stmt);
    if (rc != SQLITE_DONE && rc != SQLITE_ROW) {
        char detail[512];
        snprintf(detail, sizeof(detail), "%s", sqlite3_errmsg(stmt->db->db));
        return sqlite_fail("sqlite step failed", detail);
    }

    double changes = (double)sqlite3_changes(stmt->db->db);
    double last_id = (double)sqlite3_last_insert_rowid(stmt->db->db);

    void *obj = NULL;
    if (scriptgo_object_new(2, &obj) != 0) return -1;
    scriptgo_object_type_set(obj, "changes:lastInsertRowid");
    scriptgo_object_number_set(obj, 0, changes);
    scriptgo_object_number_set(obj, 1, last_id);

    *out_result = obj;
    return 0;
}
