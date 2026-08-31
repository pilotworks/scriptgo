#ifndef SCRIPTGO_SQLITE_TYPES_H
#include "sqlite_types.h"
#endif

static void scriptgo_sqlite_scalar_callback(sqlite3_context *ctx, int argc, sqlite3_value **argv) {
    void *closure = sqlite3_user_data(ctx);
    if (closure == NULL) {
        sqlite3_result_null(ctx);
        return;
    }

    scriptgo_sqlite_unknown_t a[4];
    memset(a, 0, sizeof(a));
    int n = argc > 4 ? 4 : argc;
    for (int i = 0; i < n; i++) {
        int type = sqlite3_value_type(argv[i]);
        if (type == SQLITE_INTEGER) {
            a[i].tag = 3;
            double d = (double)sqlite3_value_int64(argv[i]);
            memcpy(&a[i].payload, &d, sizeof(d));
        } else if (type == SQLITE_FLOAT) {
            a[i].tag = 3;
            double d = sqlite3_value_double(argv[i]);
            memcpy(&a[i].payload, &d, sizeof(d));
        } else if (type == SQLITE_TEXT) {
            a[i].tag = 4;
            a[i].payload = (uintptr_t)sqlite3_value_text(argv[i]);
        } else if (type == SQLITE_BLOB) {
            a[i].tag = 6;
            a[i].payload = (uintptr_t)sqlite3_value_blob(argv[i]);
        } else {
            a[i].tag = 1;
            a[i].payload = 0;
        }
    }

    scriptgo_sqlite_closure_t *c = (scriptgo_sqlite_closure_t *)closure;
    if (c != NULL && c->fn_ptr != NULL) {
        double (*fn_dbl)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        double ret_d = fn_dbl(c->env, a[0].tag, 0, a[0].payload, a[1].tag, 0, a[1].payload, a[2].tag, 0, a[2].payload, a[3].tag, 0, a[3].payload);
        if (!isnan(ret_d)) {
            if (floor(ret_d) == ret_d && ret_d >= -9007199254740991.0 && ret_d <= 9007199254740991.0) {
                sqlite3_result_int64(ctx, (sqlite3_int64)ret_d);
            } else {
                sqlite3_result_double(ctx, ret_d);
            }
            return;
        }

        uint64_t (*fn_ptr)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint64_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        uint64_t ret_u = fn_ptr(c->env, a[0].tag, 0, a[0].payload, a[1].tag, 0, a[1].payload, a[2].tag, 0, a[2].payload, a[3].tag, 0, a[3].payload);
        if (ret_u != 0) {
            sqlite3_result_text(ctx, (const char *)(uintptr_t)ret_u, -1, SQLITE_TRANSIENT);
            return;
        }
    }
    sqlite3_result_null(ctx);
}

int scriptgo_sqlite_create_function(void *db_handle, const char *name, double deterministic, double direct_only, void *fn_closure) {
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }
    if (name == NULL || name[0] == '\0') return sqlite_fail("sqlite function: invalid name", NULL);

    int flags = SQLITE_UTF8;
    if (deterministic > 0) flags |= SQLITE_DETERMINISTIC;
    if (direct_only > 0) flags |= SQLITE_DIRECTONLY;

    void *closure = fn_closure;
    if (fn_closure != NULL) {
        scriptgo_sqlite_unknown_t *boxed = (scriptgo_sqlite_unknown_t *)fn_closure;
        if (boxed->payload != 0) {
            closure = (void *)(uintptr_t)boxed->payload;
        }
    }

    int rc = sqlite3_create_function_v2(db->db, name, -1, flags, closure, scriptgo_sqlite_scalar_callback, NULL, NULL, NULL);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite function registration failed", sqlite3_errmsg(db->db));
    }
    return 0;
}

static void scriptgo_sqlite_agg_step(sqlite3_context *ctx, int argc, sqlite3_value **argv) {
}

static void scriptgo_sqlite_agg_final(sqlite3_context *ctx) {
    sqlite3_result_null(ctx);
}

int scriptgo_sqlite_create_aggregate(void *db_handle, const char *name, double deterministic, double direct_only, void *options) {
    if (db_handle == NULL) return sqlite_fail("sqlite: invalid database handle", NULL);
    scriptgo_sqlite_db_t *db = (scriptgo_sqlite_db_t *)db_handle;
    if (db->magic != SCRIPTGO_MAGIC_SQLITE_DB || !db->is_open || db->db == NULL) {
        return sqlite_fail("sqlite: database is not open", NULL);
    }
    if (name == NULL || name[0] == '\0') return sqlite_fail("sqlite aggregate: invalid name", NULL);

    int flags = SQLITE_UTF8;
    if (deterministic > 0) flags |= SQLITE_DETERMINISTIC;
    if (direct_only > 0) flags |= SQLITE_DIRECTONLY;

    int rc = sqlite3_create_function_v2(db->db, name, -1, flags, options, NULL, scriptgo_sqlite_agg_step, scriptgo_sqlite_agg_final, NULL);
    if (rc != SQLITE_OK) {
        return sqlite_fail("sqlite aggregate registration failed", sqlite3_errmsg(db->db));
    }
    return 0;
}
