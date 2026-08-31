#ifndef SCRIPTGO_SQLITE_TYPES_H
#include "sqlite_types.h"
#endif

int scriptgo_sqlite_open(const char *location, double flags, void **out_db) {
    if (out_db == NULL) {
        return sqlite_fail("sqlite: invalid output pointer", NULL);
    }
    const char *loc = (location != NULL && location[0] != '\0') ? location : ":memory:";
    int open_flags = (flags > 0) ? (int)flags : (SQLITE_OPEN_READWRITE | SQLITE_OPEN_CREATE);

    scriptgo_sqlite_db_t *db = calloc(1, sizeof(*db));
    if (db == NULL) {
        return sqlite_fail("sqlite: allocation failed", NULL);
    }
    db->magic = SCRIPTGO_MAGIC_SQLITE_DB;
    db->location = strdup(loc);
    db->is_open = 0;

    int rc = sqlite3_open_v2(loc, &db->db, open_flags, NULL);
    if (rc != SQLITE_OK) {
        const char *err = (db->db != NULL) ? sqlite3_errmsg(db->db) : sqlite3_errstr(rc);
        char msg[256];
        snprintf(msg, sizeof(msg), "sqlite open failed: %s", err);
        if (db->db != NULL) {
            sqlite3_close_v2(db->db);
        }
        free(db->location);
        free(db);
        return sqlite_fail(msg, NULL);
    }

    db->is_open = 1;
    sqlite3_db_config(db->db, SQLITE_DBCONFIG_ENABLE_FKEY, 1, NULL);
    scriptgo_gc_register(db, 20, 0);
    *out_db = db;
    return 0;
}

int scriptgo_sqlite_close(void *db_handle) {
    if (db_handle == NULL) {
        return 0;
    }
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB) {
        return sqlite_fail("sqlite: invalid database handle", NULL);
    }
    if (!db->is_open) {
        return 0;
    }

    /* Finalize all attached statements */
    scriptgo_sqlite_stmt_t *stmt = db->stmts;
    while (stmt != NULL) {
        scriptgo_sqlite_stmt_t *next = stmt->next;
        if (stmt->stmt != NULL) {
            sqlite3_finalize(stmt->stmt);
            stmt->stmt = NULL;
        }
        stmt = next;
    }
    db->stmts = NULL;

    /* Delete sessions */
    scriptgo_sqlite_session_t *sess = db->sessions;
    while (sess != NULL) {
        scriptgo_sqlite_session_t *next = sess->next;
        if (sess->session != NULL) {
            sqlite3session_delete(sess->session);
            sess->session = NULL;
        }
        sess = next;
    }
    db->sessions = NULL;

    if (db->db != NULL) {
        sqlite3_close_v2(db->db);
        db->db = NULL;
    }
    db->is_open = 0;
    return 0;
}

int scriptgo_sqlite_exec(void *db_handle, const char *sql) {
    if (db_handle == NULL) {
        return sqlite_fail("sqlite: invalid database handle", NULL);
    }
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }
    if (sql == NULL) {
        return 0;
    }

    char *errmsg = NULL;
    int rc = sqlite3_exec(db->db, sql, NULL, NULL, &errmsg);
    if (rc != SQLITE_OK) {
        char detail[512];
        snprintf(detail, sizeof(detail), "%s", errmsg ? errmsg : sqlite3_errstr(rc));
        if (errmsg != NULL) {
            sqlite3_free(errmsg);
        }
        return sqlite_fail("sqlite exec failed", detail);
    }
    return 0;
}

int scriptgo_sqlite_enable_load_extension(void *db_handle, double enabled) {
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }
    int rc = sqlite3_enable_load_extension(db->db, (int)enabled);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite enableLoadExtension failed", sqlite3_errmsg(db->db));
    }
    return 0;
}

int scriptgo_sqlite_load_extension(void *db_handle, const char *path) {
    if (db_handle == NULL) return 0;
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return 0;
    }
    if (path == NULL || path[0] == '\0') return 0;
    char *errmsg = NULL;
    sqlite3_enable_load_extension(db->db, 1);
    sqlite3_load_extension(db->db, path, NULL, &errmsg);
    if (errmsg != NULL) {
        sqlite3_free(errmsg);
    }
    return 0;
}

int scriptgo_sqlite_location(void *db_handle, const char *db_name, const char **out_loc) {
    if (out_loc == NULL) return 0;
    *out_loc = "";
    if (db_handle == NULL) return 0;
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) return 0;
    const char *target = (db_name != NULL && db_name[0] != '\0') ? db_name : "main";
    const char *fn = sqlite3_db_filename(db->db, target);
    *out_loc = (fn != NULL && fn[0] != '\0') ? fn : "";
    return 0;
}

int scriptgo_sqlite_is_transaction(void *db_handle, double *out_tx) {
    if (out_tx == NULL) return 0;
    *out_tx = 0.0;
    if (db_handle == NULL) return 0;
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) return 0;
    *out_tx = (sqlite3_get_autocommit(db->db) == 0) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_sqlite_backup(void *src_handle, const char *dest_path, double *out_pages) {
    if (out_pages != NULL) *out_pages = 0.0;
    if (src_handle == NULL) return sqlite_fail("sqlite backup: invalid source handle", NULL);
    scriptgo_sqlite_db_t *src_db = (scriptgo_sqlite_db_t *)src_handle;
    if (src_db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !src_db->is_open || src_db->db == NULL) {
        return sqlite_fail("sqlite backup: source database is not open", NULL);
    }
    if (dest_path == NULL || dest_path[0] == '\0') {
        return sqlite_fail("sqlite backup: invalid destination path", NULL);
    }
    sqlite3 *dest_db = NULL;
    int rc = sqlite3_open(dest_path, &dest_db);
    if (rc != SQLITE_OK) {
        if (dest_db != NULL) sqlite3_close_v2(dest_db);
        return sqlite_fail("sqlite backup: failed to open destination", sqlite3_errstr(rc));
    }
    sqlite3_backup *b = sqlite3_backup_init(dest_db, "main", src_db->db, "main");
    if (b == NULL) {
        sqlite3_close_v2(dest_db);
        return sqlite_fail("sqlite backup: init failed", sqlite3_errmsg(dest_db));
    }
    rc = sqlite3_backup_step(b, -1);
    int total_pages = sqlite3_backup_pagecount(b);
    sqlite3_backup_finish(b);
    sqlite3_close_v2(dest_db);
    if (rc != SQLITE_DONE && rc != SQLITE_OK) {
        return sqlite_fail("sqlite backup step failed", sqlite3_errstr(rc));
    }
    if (out_pages != NULL) *out_pages = (double)total_pages;
    return 0;
}
