#ifndef SCRIPTGO_SQLITE_TYPES_H
#include "sqlite_types.h"
#endif

static int build_type_name(sqlite3_stmt *stmt, int col_count, char **out_type_name) {
    size_t total_len = 1;
    for (int i = 0; i < col_count; i++) {
        const char *name = sqlite3_column_name(stmt, i);
        total_len += (name ? strlen(name) : 0) + 1;
    }
    char *buf = malloc(total_len);
    if (buf == NULL) return -1;
    buf[0] = '\0';
    for (int i = 0; i < col_count; i++) {
        const char *name = sqlite3_column_name(stmt, i);
        if (i > 0) strcat(buf, ":");
        strcat(buf, name ? name : "");
    }
    *out_type_name = buf;
    return 0;
}

static int read_row(scriptgo_sqlite_stmt_t *stmt, void **out_row) {
    int col_count = sqlite3_column_count(stmt->stmt);
    if (stmt->return_arrays) {
        void *arr = NULL;
        if (scriptgo_array_new(0, (int64_t)sizeof(scriptgo_sqlite_unknown_t), &arr) != 0 ||
            scriptgo_array_set_tag(arr, 6) != 0) {
            return -1;
        }
        double len = 0;
        for (int i = 0; i < col_count; i++) {
            scriptgo_sqlite_unknown_t val;
            val.padding = 0;
            int type = sqlite3_column_type(stmt->stmt, i);
            if (type == SQLITE_NULL) {
                val.tag = 1; val.payload = 0;
            } else if (type == SQLITE_INTEGER) {
                if (stmt->read_bigints) {
                    val.tag = 7;
                    val.payload = (uint64_t)sqlite3_column_int64(stmt->stmt, i);
                } else {
                    val.tag = 3;
                    double num = (double)sqlite3_column_int64(stmt->stmt, i);
                    memcpy(&val.payload, &num, sizeof(num));
                }
            } else if (type == SQLITE_FLOAT) {
                val.tag = 3;
                double num = sqlite3_column_double(stmt->stmt, i);
                memcpy(&val.payload, &num, sizeof(num));
            } else if (type == SQLITE_TEXT) {
                val.tag = 4;
                val.payload = (uint64_t)(uintptr_t)strdup((const char *)sqlite3_column_text(stmt->stmt, i));
            } else if (type == SQLITE_BLOB) {
                const void *blob = sqlite3_column_blob(stmt->stmt, i);
                int bytes = sqlite3_column_bytes(stmt->stmt, i);
                void *buf = NULL;
                if (scriptgo_buffer_alloc((double)bytes, NULL, 0, 0, 0, &buf) == 0 && buf != NULL) {
                    scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)buf;
                    if (bytes > 0 && blob != NULL) memcpy(view->data, blob, (size_t)bytes);
                    val.tag = 6;
                    val.payload = (uint64_t)(uintptr_t)buf;
                } else {
                    val.tag = 1; val.payload = 0;
                }
            }
            if (scriptgo_array_push(arr, &val, &len) != 0) return -1;
        }
        *out_row = arr;
        return 0;
    }

    void *obj = NULL;
    if (scriptgo_object_new(col_count, &obj) != 0) return -1;
    char *type_name = NULL;
    if (build_type_name(stmt->stmt, col_count, &type_name) == 0 && type_name != NULL) {
        scriptgo_object_type_set(obj, type_name);
        free(type_name);
    }

    for (int i = 0; i < col_count; i++) {
        int type = sqlite3_column_type(stmt->stmt, i);
        if (type == SQLITE_NULL) {
            scriptgo_object_unknown_set(obj, i, 1, 0);
        } else if (type == SQLITE_INTEGER) {
            if (stmt->read_bigints) {
                scriptgo_object_bigint_set(obj, i, sqlite3_column_int64(stmt->stmt, i));
            } else {
                scriptgo_object_number_set(obj, i, (double)sqlite3_column_int64(stmt->stmt, i));
            }
        } else if (type == SQLITE_FLOAT) {
            scriptgo_object_number_set(obj, i, sqlite3_column_double(stmt->stmt, i));
        } else if (type == SQLITE_TEXT) {
            const char *text = (const char *)sqlite3_column_text(stmt->stmt, i);
            scriptgo_object_string_set(obj, i, text ? strdup(text) : "");
        } else if (type == SQLITE_BLOB) {
            const void *blob = sqlite3_column_blob(stmt->stmt, i);
            int bytes = sqlite3_column_bytes(stmt->stmt, i);
            void *buf = NULL;
            if (scriptgo_buffer_alloc((double)bytes, NULL, 0, 0, 0, &buf) == 0 && buf != NULL) {
                scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)buf;
                if (bytes > 0 && blob != NULL) memcpy(view->data, blob, (size_t)bytes);
                scriptgo_object_ptr_set(obj, i, buf);
            } else {
                scriptgo_object_unknown_set(obj, i, 1, 0);
            }
        }
    }
    *out_row = obj;
    return 0;
}

int scriptgo_sqlite_get(void *stmt_handle, void *first_param, void *rest_params, void **out_row) {
    if (out_row == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (stmt_handle == NULL) return sqlite_fail("sqlite: invalid statement handle", NULL);
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT || stmt->stmt == NULL) {
        return sqlite_fail("sqlite: statement is not prepared or closed", NULL);
    }

    if (scriptgo_sqlite_bind_all(stmt, first_param, rest_params) != 0) {
        return -1;
    }

    int rc = sqlite3_step(stmt->stmt);
    if (rc == SQLITE_ROW) {
        return read_row(stmt, out_row);
    }
    if (rc == SQLITE_DONE) {
        *out_row = NULL; /* undefined */
        return 0;
    }
    return sqlite_fail("sqlite get failed", sqlite3_errmsg(stmt->db->db));
}

int scriptgo_sqlite_all(void *stmt_handle, void *first_param, void *rest_params, void **out_rows) {
    if (out_rows == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (stmt_handle == NULL) return sqlite_fail("sqlite: invalid statement handle", NULL);
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT || stmt->stmt == NULL) {
        return sqlite_fail("sqlite: statement is not prepared or closed", NULL);
    }

    if (scriptgo_sqlite_bind_all(stmt, first_param, rest_params) != 0) {
        return -1;
    }

    void *arr = NULL;
    if (scriptgo_array_new(0, (int64_t)sizeof(void *), &arr) != 0 ||
        scriptgo_array_set_tag(arr, 5) != 0) {
        return -1;
    }

    double len = 0;
    while (1) {
        int rc = sqlite3_step(stmt->stmt);
        if (rc == SQLITE_ROW) {
            void *row = NULL;
            if (read_row(stmt, &row) != 0) return -1;
            if (scriptgo_array_push(arr, &row, &len) != 0) return -1;
        } else if (rc == SQLITE_DONE) {
            break;
        } else {
            return sqlite_fail("sqlite all failed", sqlite3_errmsg(stmt->db->db));
        }
    }
    *out_rows = arr;
    return 0;
}

int scriptgo_sqlite_columns(void *stmt_handle, void **out_cols) {
    if (out_cols == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (stmt_handle == NULL) return sqlite_fail("sqlite: invalid statement handle", NULL);
    scriptgo_sqlite_stmt_t *stmt = (scriptgo_sqlite_stmt_t *)stmt_handle;
    if (stmt->magic != SCRIPTGO_MAGIC_SQLITE_STMT || stmt->stmt == NULL) {
        return sqlite_fail("sqlite: statement is not prepared or closed", NULL);
    }

    int col_count = sqlite3_column_count(stmt->stmt);
    void *arr = NULL;
    if (scriptgo_array_new(0, (int64_t)sizeof(void *), &arr) != 0 ||
        scriptgo_array_set_tag(arr, 5) != 0) {
        return -1;
    }

    double len = 0;
    for (int i = 0; i < col_count; i++) {
        void *col_obj = NULL;
        if (scriptgo_object_new(5, &col_obj) != 0) return -1;
        scriptgo_object_type_set(col_obj, "column:database:name:table:type");

        const char *name = sqlite3_column_name(stmt->stmt, i);
        const char *orig_name = sqlite3_column_origin_name(stmt->stmt, i);
        const char *table_name = sqlite3_column_table_name(stmt->stmt, i);
        const char *db_name = sqlite3_column_database_name(stmt->stmt, i);
        const char *decl_type = sqlite3_column_decltype(stmt->stmt, i);

        scriptgo_object_string_set(col_obj, 0, orig_name ? strdup(orig_name) : (name ? strdup(name) : ""));
        scriptgo_object_string_set(col_obj, 1, db_name ? strdup(db_name) : "main");
        scriptgo_object_string_set(col_obj, 2, name ? strdup(name) : "");
        scriptgo_object_string_set(col_obj, 3, table_name ? strdup(table_name) : "");
        scriptgo_object_string_set(col_obj, 4, decl_type ? strdup(decl_type) : "");

        if (scriptgo_array_push(arr, &col_obj, &len) != 0) return -1;
    }
    *out_cols = arr;
    return 0;
}
