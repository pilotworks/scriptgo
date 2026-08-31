#ifndef SCRIPTGO_SQLITE_TYPES_H
#define SCRIPTGO_SQLITE_TYPES_H

#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>

#ifndef _SQLITE3_H_
#ifndef SQLITE3_H
#include "sqlite3.h"
#endif
#endif

#define SCRIPTGO_MAGIC_SQLITE_DB      0x53514C44u /* 'SQLD' */
#define SCRIPTGO_MAGIC_SQLITE_STMT    0x53514C53u /* 'SQLS' */
#define SCRIPTGO_MAGIC_SQLITE_SESSION 0x53514C4Eu /* 'SQLN' */

#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252u
#define SCRIPTGO_MAGIC_BUFFER     0x42554646u
#define SCRIPTGO_OBJECT_MAGIC     0x53474F424A454354ULL

typedef struct scriptgo_sqlite_stmt scriptgo_sqlite_stmt_t;
typedef struct scriptgo_sqlite_session scriptgo_sqlite_session_t;

typedef struct scriptgo_sqlite_db {
    uint32_t magic;
    sqlite3 *db;
    char *location;
    int is_open;
    int enable_foreign_keys;
    scriptgo_sqlite_stmt_t *stmts;
    scriptgo_sqlite_session_t *sessions;
} scriptgo_sqlite_db_t;

struct scriptgo_sqlite_stmt {
    uint32_t magic;
    sqlite3_stmt *stmt;
    scriptgo_sqlite_db_t *db;
    char *sql;
    int allow_bare_named;
    int allow_unknown_named;
    int return_arrays;
    int read_bigints;
    scriptgo_sqlite_stmt_t *next;
    scriptgo_sqlite_stmt_t *prev;
};

struct scriptgo_sqlite_session {
    uint32_t magic;
    sqlite3_session *session;
    scriptgo_sqlite_db_t *db;
    scriptgo_sqlite_session_t *next;
    scriptgo_sqlite_session_t *prev;
};

typedef struct {
    uint32_t tag;
    uint32_t padding;
    uint64_t payload;
} scriptgo_sqlite_unknown_t;

typedef struct {
    uint32_t magic;
    uint32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_sqlite_buffer_view_t;

typedef struct {
    void *fn_ptr;
    void *env;
} scriptgo_sqlite_closure_t;

/* Externs from scriptgo runtime */
int scriptgo_runtime_set_error(const char *message);
int scriptgo_object_new(int64_t field_count, void **out_object);
int scriptgo_object_type_set(void *handle, const char *type_name);
int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_bool_set(void *handle, int64_t index, int32_t value);
int scriptgo_object_bigint_set(void *handle, int64_t index, int64_t value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);
int scriptgo_object_unknown_set(void *handle, int64_t index, uint32_t tag, uint64_t payload);
int scriptgo_object_unknown_get(void *handle, int64_t index, uint32_t *out_tag, uint64_t *out_payload);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set_tag(void *handle, int64_t tag);
int scriptgo_array_push(void *handle, const void *value, double *out_length);
int scriptgo_buffer_alloc(double size, const char *fill_str, double fill_num, int has_fill, int is_str_fill, void **out_buf);
int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count);
int scriptgo_gc_unregister(void *ptr);

int scriptgo_sqlite_bind_all(scriptgo_sqlite_stmt_t *stmt, void *first_param, void *rest_params);

static inline int sqlite_fail(const char *prefix, const char *detail) {
    char msg[512];
    if (detail != NULL && detail[0] != '\0') {
        snprintf(msg, sizeof(msg), "%s: %s", prefix, detail);
    } else {
        snprintf(msg, sizeof(msg), "%s", prefix);
    }
    return scriptgo_runtime_set_error(msg);
}

#endif /* SCRIPTGO_SQLITE_TYPES_H */
