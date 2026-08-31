#ifndef SCRIPTGO_SQLITE_TYPES_H
#include "sqlite_types.h"
#endif

int scriptgo_sqlite_session_create(void *db_handle, const char *table, void **out_session) {
    if (out_session == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }

    sqlite3_session *s = NULL;
    int rc = sqlite3session_create(db->db, "main", &s);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite createSession failed", sqlite3_errmsg(db->db));
    }
    sqlite3session_attach(s, (table != NULL && table[0] != '\0') ? table : NULL);

    scriptgo_sqlite_session_t *sess = calloc(1, sizeof(*sess));
    if (sess == NULL) {
        sqlite3session_delete(s);
        return sqlite_fail("sqlite: allocation failed", NULL);
    }
    sess->magic = SCRIPTGO_MAGIC_SQLITE_SESSION;
    sess->session = s;
    sess->db = db;

    sess->next = db->sessions;
    if (db->sessions != NULL) db->sessions->prev = sess;
    db->sessions = sess;

    scriptgo_gc_register(sess, 22, 0);
    *out_session = sess;
    return 0;
}

int scriptgo_sqlite_session_close(void *sess_handle) {
    if (sess_handle == NULL) return 0;
    scriptgo_sqlite_session_t *sess = (scriptgo_sqlite_session_t *)sess_handle;
    if (sess->magic != SCRIPTGO_MAGIC_SQLITE_SESSION) return 0;

    if (sess->session != NULL) {
        sqlite3session_delete(sess->session);
        sess->session = NULL;
    }
    if (sess->db != NULL) {
        if (sess->prev != NULL) sess->prev->next = sess->next;
        else if (sess->db->sessions == sess) sess->db->sessions = sess->next;
        if (sess->next != NULL) sess->next->prev = sess->prev;
    }
    sess->magic = 0;
    return 0;
}

int scriptgo_sqlite_session_changeset(void *sess_handle, void **out_buf) {
    if (out_buf == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (sess_handle == NULL) return sqlite_fail("sqlite: invalid session handle", NULL);
    scriptgo_sqlite_session_t *sess = (scriptgo_sqlite_session_t *)sess_handle;
    if (sess->magic != SCRIPTGO_MAGIC_SQLITE_SESSION || sess->session == NULL) {
        return sqlite_fail("sqlite: session is closed", NULL);
    }

    int nchangeset = 0;
    void *pchangeset = NULL;
    int rc = sqlite3session_changeset(sess->session, &nchangeset, &pchangeset);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite changeset failed", sqlite3_errmsg(sess->db->db));
    }

    void *buf = NULL;
    if (scriptgo_buffer_alloc((double)nchangeset, NULL, 0, 0, 0, &buf) == 0 && buf != NULL) {
        scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)buf;
        if (nchangeset > 0 && pchangeset != NULL) {
            memcpy(view->data, pchangeset, (size_t)nchangeset);
        }
        *out_buf = buf;
    }
    if (pchangeset != NULL) sqlite3_free(pchangeset);
    return 0;
}

int scriptgo_sqlite_session_patchset(void *sess_handle, void **out_buf) {
    if (out_buf == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (sess_handle == NULL) return sqlite_fail("sqlite: invalid session handle", NULL);
    scriptgo_sqlite_session_t *sess = (scriptgo_sqlite_session_t *)sess_handle;
    if (sess->magic != SCRIPTGO_MAGIC_SQLITE_SESSION || sess->session == NULL) {
        return sqlite_fail("sqlite: session is closed", NULL);
    }

    int npatchset = 0;
    void *ppatchset = NULL;
    int rc = sqlite3session_patchset(sess->session, &npatchset, &ppatchset);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite patchset failed", sqlite3_errmsg(sess->db->db));
    }

    void *buf = NULL;
    if (scriptgo_buffer_alloc((double)npatchset, NULL, 0, 0, 0, &buf) == 0 && buf != NULL) {
        scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)buf;
        if (npatchset > 0 && ppatchset != NULL) {
            memcpy(view->data, ppatchset, (size_t)npatchset);
        }
        *out_buf = buf;
    }
    if (ppatchset != NULL) sqlite3_free(ppatchset);
    return 0;
}

static int default_conflict_handler(void *pCtx, int eConflict, sqlite3_changeset_iter *p) {
    return SQLITE_CHANGESET_REPLACE;
}

int scriptgo_sqlite_apply_changeset(void *db_handle, void *changeset_buf, double on_conflict, int32_t *out_result) {
    if (out_result == NULL) return sqlite_fail("sqlite: invalid output pointer", NULL);
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }

    if (changeset_buf == NULL) {
        *out_result = 1;
        return 0;
    }

    scriptgo_sqlite_buffer_view_t *view = (scriptgo_sqlite_buffer_view_t *)changeset_buf;
    if (view->length <= 0 || view->data == NULL) {
        *out_result = 1;
        return 0;
    }

    int rc = sqlite3changeset_apply(db->db, (int)view->length, view->data, NULL, default_conflict_handler, NULL);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite applyChangeset failed", sqlite3_errmsg(db->db));
    }
    *out_result = 1;
    return 0;
}
