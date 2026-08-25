#ifndef _GNU_SOURCE
#define _GNU_SOURCE 1
#endif
#ifndef _DEFAULT_SOURCE
#define _DEFAULT_SOURCE 1
#endif
#ifdef __APPLE__
#ifndef _DARWIN_C_SOURCE
#define _DARWIN_C_SOURCE 1
#endif
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <setjmp.h>

static char scriptgo_runtime_error[256];

int scriptgo_runtime_set_error(const char *message) {
    if (message == NULL) message = "scriptgo runtime error";
    strncpy(scriptgo_runtime_error, message, sizeof(scriptgo_runtime_error) - 1);
    scriptgo_runtime_error[sizeof(scriptgo_runtime_error) - 1] = '\0';
    return -1;
}

const char *scriptgo_runtime_last_error(void) { return scriptgo_runtime_error; }

void scriptgo_runtime_abort_if_failed(int status) {
    if (status != 0) {
        fputs(scriptgo_runtime_last_error(), stderr);
        fputc('\n', stderr);
        abort();
    }
}

typedef enum {
    SCRIPTGO_TAG_UNDEFINED = 0,
    SCRIPTGO_TAG_NULL      = 1,
    SCRIPTGO_TAG_BOOLEAN   = 2,
    SCRIPTGO_TAG_NUMBER    = 3,
    SCRIPTGO_TAG_STRING    = 4,
    SCRIPTGO_TAG_OBJECT    = 5,
    SCRIPTGO_TAG_ARRAY     = 6,
    SCRIPTGO_TAG_FUNCTION  = 7,
    SCRIPTGO_TAG_BIGINT    = 8,
    SCRIPTGO_TAG_SYMBOL    = 9
} ScriptGoTypeTag;

typedef struct {
    unsigned int tag;
    unsigned int flags;
    unsigned long long payload;
} ScriptGoUnknown;

typedef struct {
    int32_t tag;
    int32_t pad;
    int64_t payload;
} scriptgo_boxed_value;

static const char *scriptgo_tag_name(unsigned int tag) {
    switch (tag) {
    case SCRIPTGO_TAG_UNDEFINED: return "undefined";
    case SCRIPTGO_TAG_NULL:      return "null";
    case SCRIPTGO_TAG_BOOLEAN:   return "boolean";
    case SCRIPTGO_TAG_NUMBER:    return "number";
    case SCRIPTGO_TAG_STRING:    return "string";
    case SCRIPTGO_TAG_OBJECT:    return "object";
    case SCRIPTGO_TAG_ARRAY:     return "array";
    case SCRIPTGO_TAG_FUNCTION:  return "function";
    case SCRIPTGO_TAG_BIGINT:    return "bigint";
    case SCRIPTGO_TAG_SYMBOL:    return "symbol";
    default:                     return "unknown";
    }
}

void scriptgo_throw_string(const char *str);

void __scriptgo_fail_checked_cast(unsigned int actual_tag, unsigned int expected_tag, const char *span) {
    char msg[256];
    if (span != NULL && strlen(span) > 0) {
        snprintf(msg, sizeof(msg), "TypeError: SG4002: cannot cast %s to %s at %s",
                 scriptgo_tag_name(actual_tag),
                 scriptgo_tag_name(expected_tag),
                 span);
    } else {
        snprintf(msg, sizeof(msg), "TypeError: SG4002: cannot cast %s to %s",
                 scriptgo_tag_name(actual_tag),
                 scriptgo_tag_name(expected_tag));
    }
    scriptgo_throw_string(msg);
}

const char *__scriptgo_typeof_unknown(unsigned int tag) {
    switch (tag) {
    case SCRIPTGO_TAG_UNDEFINED: return "undefined";
    case SCRIPTGO_TAG_NULL:      return "object";
    case SCRIPTGO_TAG_BOOLEAN:   return "boolean";
    case SCRIPTGO_TAG_NUMBER:    return "number";
    case SCRIPTGO_TAG_STRING:    return "string";
    case SCRIPTGO_TAG_FUNCTION:  return "function";
    case SCRIPTGO_TAG_BIGINT:    return "bigint";
    case SCRIPTGO_TAG_SYMBOL:    return "symbol";
    default:                     return "object";
    }
}

int scriptgo_string_from_number(double value, char **out_value);
int scriptgo_string_from_bigint(long long value, char **out_str);
int scriptgo_string_from_object(void *obj, char **out_str);

int scriptgo_string_from_unknown(unsigned int tag, unsigned int padding, unsigned long long payload, char **out_str) {
    (void)padding;
    if (out_str == NULL) {
        return -1;
    }
    switch (tag) {
    case SCRIPTGO_TAG_UNDEFINED:
        *out_str = strdup("undefined");
        return 0;
    case SCRIPTGO_TAG_NULL:
        if (payload == 0) {
            *out_str = strdup("null");
        } else {
            *out_str = (char *)(uintptr_t)payload;
        }
        return 0;
    case SCRIPTGO_TAG_BOOLEAN: {
        int b = (int)payload;
        *out_str = strdup(b ? "true" : "false");
        return 0;
    }
    case SCRIPTGO_TAG_NUMBER: {
        union {
            unsigned long long u64;
            double d;
        } u;
        u.u64 = payload;
        return scriptgo_string_from_number(u.d, out_str);
    }
    case SCRIPTGO_TAG_STRING:
        *out_str = (char *)(uintptr_t)payload;
        return 0;
    case SCRIPTGO_TAG_BIGINT:
        return scriptgo_string_from_bigint((long long)payload, out_str);
    default:
        if (payload == 0) {
            *out_str = strdup("null");
            return 0;
        }
        return scriptgo_string_from_object((void *)(uintptr_t)payload, out_str);
    }
}

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} scriptgo_object_t;

int scriptgo_string_from_object(void *obj, char **out_str) {
    if (out_str == NULL) {
        return -1;
    }
    if (obj == NULL) {
        *out_str = strdup("null");
        return 0;
    }
    if ((uintptr_t)obj > 0x00007FFFFFFFFFFFULL) {
        union {
            unsigned long long u64;
            double d;
        } u;
        u.u64 = (uintptr_t)obj;
        return scriptgo_string_from_number(u.d, out_str);
    }
    scriptgo_object_t *o = (scriptgo_object_t *)obj;
    if (o->magic == SCRIPTGO_OBJECT_MAGIC) {
        if (o->field_count > 0) {
            uintptr_t f0 = o->fields[0];
            if (f0 != 0x7FF8000000000000ULL && f0 > 0x1000 && f0 <= 0x00007FFFFFFFFFFFULL) {
                *out_str = strdup((char *)f0);
                return 0;
            }
        }
        *out_str = strdup("[object Object]");
        return 0;
    }
    *out_str = (char *)obj;
    return 0;
}

typedef struct scriptgo_exception_frame {
    jmp_buf buf;
    const char *thrown_string;
    double thrown_number;
    int thrown_bool;
    int thrown_type;
    struct scriptgo_exception_frame *prev;
} scriptgo_exception_frame_t;

static scriptgo_exception_frame_t *scriptgo_top_frame = NULL;

void scriptgo_exception_push(scriptgo_exception_frame_t *frame) {
    frame->thrown_string = NULL;
    frame->thrown_number = 0.0;
    frame->thrown_bool = 0;
    frame->thrown_type = 0;
    frame->prev = scriptgo_top_frame;
    scriptgo_top_frame = frame;
}

void scriptgo_exception_pop(scriptgo_exception_frame_t *frame) {
    if (scriptgo_top_frame == frame) {
        scriptgo_top_frame = frame->prev;
    }
}

void *scriptgo_exception_buf(scriptgo_exception_frame_t *frame) {
    return (void*)frame->buf;
}

scriptgo_exception_frame_t *scriptgo_exception_frame_new(void) {
    scriptgo_exception_frame_t *frame = malloc(sizeof(scriptgo_exception_frame_t));
    if (frame == NULL) return NULL;
    scriptgo_exception_push(frame);
    return frame;
}

void scriptgo_exception_frame_free(scriptgo_exception_frame_t *frame) {
    if (frame == NULL) return;
    scriptgo_exception_pop(frame);
    free(frame);
}

void scriptgo_throw_string(const char *str) {
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
        scriptgo_top_frame = frame->prev;
        frame->thrown_string = str;
        frame->thrown_type = 1;
        longjmp(frame->buf, 1);
    }
    fprintf(stderr, "Uncaught exception: %s\n", str ? str : "");
    exit(1);
}

void scriptgo_throw_number(double num) {
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
        scriptgo_top_frame = frame->prev;
        frame->thrown_number = num;
        frame->thrown_type = 2;
        longjmp(frame->buf, 1);
    }
    fprintf(stderr, "Uncaught exception: %g\n", num);
    exit(1);
}

void scriptgo_throw_bool(int val) {
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
        scriptgo_top_frame = frame->prev;
        frame->thrown_bool = val;
        frame->thrown_type = 3;
        longjmp(frame->buf, 1);
    }
    fprintf(stderr, "Uncaught exception: %s\n", val ? "true" : "false");
    exit(1);
}

const char *scriptgo_exception_get_string(scriptgo_exception_frame_t *frame) {
    return (frame && frame->thrown_string) ? frame->thrown_string : "";
}

double scriptgo_exception_get_number(scriptgo_exception_frame_t *frame) {
    return frame->thrown_number;
}

int scriptgo_exception_get_bool(scriptgo_exception_frame_t *frame) {
    return frame->thrown_bool;
}

void scriptgo_exception_rethrow(scriptgo_exception_frame_t *frame) {
    if (frame == NULL) return;
    int type = frame->thrown_type;
    const char *str = frame->thrown_string;
    double num = frame->thrown_number;
    int b = frame->thrown_bool;
    scriptgo_exception_frame_free(frame);
    if (type == 1) {
        scriptgo_throw_string(str);
    } else if (type == 2) {
        scriptgo_throw_number(num);
    } else if (type == 3) {
        scriptgo_throw_bool(b);
    }
}

void scriptgo_debugger_break(const char *file, int line) {
    (void)file;
    (void)line;
}

