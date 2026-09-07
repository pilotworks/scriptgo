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
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#if defined(__wasi__)
typedef int jmp_buf[16];
#define setjmp(env) (0)
#define longjmp(env, val) abort()
#else
#include <setjmp.h>
#endif
#include <math.h>

const char scriptgo_undefined_sentinel = 0;
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

int32_t scriptgo_is_truthy_unknown(const scriptgo_value *value) {
    unsigned int tag;
    unsigned long long payload;
    if (value == NULL || scriptgo_value_validate(value) != 0) return 0;
    tag = value->tag;
    payload = value->payload;
    switch (tag) {
    case SCRIPTGO_TAG_UNDEFINED:
    case SCRIPTGO_TAG_NULL:
        return 0;
    case SCRIPTGO_TAG_BOOLEAN:
        return payload ? 1 : 0;
    case SCRIPTGO_TAG_NUMBER: {
        union {
            unsigned long long u64;
            double d;
        } u;
        u.u64 = payload;
        return (u.d != 0.0 && !isnan(u.d)) ? 1 : 0;
    }
    case SCRIPTGO_TAG_STRING: {
        const char *s = (const char *)(uintptr_t)payload;
        return (s != NULL && s[0] != '\0') ? 1 : 0;
    }
    default:
        return payload != 0 ? 1 : 0;
    }
}

const char *__scriptgo_typeof_unknown(const scriptgo_value *value) {
    unsigned int tag = value == NULL ? SCRIPTGO_TAG_UNDEFINED : value->tag;
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

int scriptgo_string_from_unknown(const scriptgo_value *value, char **out_str) {
    unsigned int tag;
    unsigned long long payload;
    if (value == NULL || scriptgo_value_validate(value) != 0 || out_str == NULL) {
        return -1;
    }
    tag = value->tag;
    payload = value->payload;
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
        {
            uint64_t length = value->aux;
            if (length == 0 && payload != 0) length = strlen((const char *)(uintptr_t)payload);
            if (length > (uint64_t)SIZE_MAX - 1) return -1;
            *out_str = malloc((size_t)length + 1);
            if (*out_str == NULL) return -1;
            if (length != 0) memcpy(*out_str, (const void *)(uintptr_t)payload, (size_t)length);
            (*out_str)[length] = '\0';
        }
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
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    uintptr_t fields[];
} scriptgo_object_t;

int scriptgo_gc_is_registered(void *ptr);

int scriptgo_string_from_object(void *obj, char **out_str) {
    if (out_str == NULL) {
        return -1;
    }
    if (obj == NULL) {
        *out_str = strdup("null");
        return 0;
    }
    if (obj == &scriptgo_undefined_sentinel) {
        *out_str = strdup("undefined");
        return 0;
    }
    uint32_t magic = *(const uint32_t *)obj;
    if (magic == 0x42554646) {
        struct {
            uint32_t magic;
            int32_t kind;
            int64_t length;
            int64_t byte_offset;
            int64_t element_size;
            void *buffer;
            unsigned char *data;
        } *bv = (void *)obj;
        char *s = malloc((size_t)bv->length + 1);
        if (s != NULL) {
            if (bv->data != NULL && bv->length > 0) {
                memcpy(s, bv->data, (size_t)bv->length);
            }
            s[bv->length] = '\0';
            *out_str = s;
            return 0;
        }
    }
    if (scriptgo_gc_is_registered(obj)) {
        scriptgo_object_t *o = (scriptgo_object_t *)obj;
        if (o->magic == SCRIPTGO_OBJECT_MAGIC) {
            if (o->type_name != NULL && (strcmp(o->type_name, "Error") == 0 || strstr(o->type_name, "Error") != NULL)) {
                if (o->field_count > 0 && o->fields[0] != 0 && o->fields[0] != 0x7FF8000000000000ULL) {
                    *out_str = strdup((const char *)o->fields[0]);
                    return 0;
                }
            }
            *out_str = strdup("[object Object]");
            return 0;
        }
    }
    *out_str = (char *)obj;
    return 0;
}

int scriptgo_error_to_string(void *obj, char **out_str) {
    scriptgo_object_t *error;
    const char *name;
    const char *message;
    size_t name_len;
    size_t message_len;
    char *result;

    if (obj == NULL || out_str == NULL || !scriptgo_gc_is_registered(obj)) {
        return scriptgo_runtime_set_error("invalid error object");
    }
    error = (scriptgo_object_t *)obj;
    if (error->magic != SCRIPTGO_OBJECT_MAGIC || error->field_count < 2) {
        return scriptgo_runtime_set_error("invalid error object");
    }
    name = error->fields[1] == 0 ? "Error" : (const char *)error->fields[1];
    message = error->fields[0] == 0 ? "" : (const char *)error->fields[0];
    name_len = strlen(name);
    message_len = strlen(message);
    if (name_len == 0) name = "Error";
    if (message_len == 0) {
        *out_str = strdup(name);
    } else if (name_len == 0) {
        *out_str = strdup(message);
    } else {
        result = malloc(name_len + message_len + 3);
        if (result == NULL) return scriptgo_runtime_set_error("error string allocation failed");
        memcpy(result, name, name_len);
        memcpy(result + name_len, ": ", 2);
        memcpy(result + name_len + 2, message, message_len + 1);
        *out_str = result;
    }
    return *out_str == NULL ? scriptgo_runtime_set_error("error string allocation failed") : 0;
}

struct scriptgo_exception_frame {
    jmp_buf buf;
    scriptgo_value thrown;
    struct scriptgo_exception_frame *prev;
};

static scriptgo_exception_frame_t *scriptgo_top_frame = NULL;
static _Thread_local char *scriptgo_exception_string_scratch = NULL;

void scriptgo_exception_push(scriptgo_exception_frame_t *frame) {
    scriptgo_value_init_undefined(&frame->thrown);
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
    scriptgo_value_release(&frame->thrown);
    free(frame);
}

static void scriptgo_uncaught_value(const scriptgo_value *value) {
    if (value == NULL) exit(1);
    if (value->tag == SCRIPTGO_TAG_OBJECT || value->tag == SCRIPTGO_TAG_ARRAY ||
        value->tag == SCRIPTGO_TAG_FUNCTION) {
        fprintf(stderr, "Uncaught exception object: %p\n", (void *)(uintptr_t)value->payload);
    } else if (value->tag == SCRIPTGO_TAG_NUMBER) {
        double number;
        memcpy(&number, &value->payload, sizeof(number));
        fprintf(stderr, "Uncaught exception: %g\n", number);
    } else if (value->tag == SCRIPTGO_TAG_BOOLEAN) {
        fprintf(stderr, "Uncaught exception: %s\n", value->payload ? "true" : "false");
    } else if (value->tag == SCRIPTGO_TAG_STRING) {
        fprintf(stderr, "Uncaught exception: %.*s\n", (int)value->aux,
                value->payload ? (const char *)(uintptr_t)value->payload : "");
    } else if (value->tag == SCRIPTGO_TAG_NULL) {
        fprintf(stderr, "Uncaught exception: null\n");
    } else {
        fprintf(stderr, "Uncaught exception: undefined\n");
    }
    exit(1);
}

void scriptgo_exception_throw_copy(const scriptgo_value *value) {
    scriptgo_value copied;
    scriptgo_value_init_undefined(&copied);
    if (scriptgo_value_clone(&copied, value) != 0) {
        fputs("scriptgo exception value clone failed\n", stderr);
        abort();
    }
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
        scriptgo_top_frame = frame->prev;
        scriptgo_value_release(&frame->thrown);
        frame->thrown = copied;
        longjmp(frame->buf, 1);
    }
    scriptgo_uncaught_value(&copied);
    scriptgo_value_release(&copied);
    abort();
}

void scriptgo_exception_throw_move(scriptgo_value *value) {
    scriptgo_value moved;
    scriptgo_value_init_undefined(&moved);
    if (scriptgo_value_move(&moved, value) != 0) {
        fputs("scriptgo exception value move failed\n", stderr);
        abort();
    }
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
        scriptgo_top_frame = frame->prev;
        scriptgo_value_release(&frame->thrown);
        frame->thrown = moved;
        longjmp(frame->buf, 1);
    }
    scriptgo_uncaught_value(&moved);
}

void scriptgo_exception_take_value(scriptgo_exception_frame_t *frame, scriptgo_value *out) {
    if (frame == NULL || out == NULL) {
        fputs("scriptgo exception value take failed\n", stderr);
        abort();
    }
    if (scriptgo_value_move(out, &frame->thrown) != 0) abort();
}

void scriptgo_throw_string(const char *str) {
    scriptgo_value value;
    if (str != NULL && scriptgo_gc_is_registered((void *)str)) {
        value.tag = SCRIPTGO_TAG_OBJECT;
        value.flags = 0;
        value.payload = (uint64_t)(uintptr_t)str;
        value.aux = 0;
    } else {
        scriptgo_value_string_borrow(str ? str : "", str ? strlen(str) : 0, &value);
    }
    scriptgo_exception_throw_copy(&value);
}

void scriptgo_throw_number(double num) {
    scriptgo_value value;
    value.tag = SCRIPTGO_TAG_NUMBER;
    value.flags = 0;
    memcpy(&value.payload, &num, sizeof(num));
    value.aux = 0;
    scriptgo_exception_throw_copy(&value);
}

void scriptgo_throw_bool(int val) {
    scriptgo_value value;
    value.tag = SCRIPTGO_TAG_BOOLEAN;
    value.flags = 0;
    value.payload = val != 0;
    value.aux = 0;
    scriptgo_exception_throw_copy(&value);
}

const char *scriptgo_exception_get_string(scriptgo_exception_frame_t *frame) {
    const char *source;
    char *copy;
    if (frame == NULL) return "";
    if (frame->thrown.tag == SCRIPTGO_TAG_STRING || frame->thrown.tag == SCRIPTGO_TAG_OBJECT) {
        source = (const char *)(uintptr_t)frame->thrown.payload;
        if (frame->thrown.tag != SCRIPTGO_TAG_STRING ||
            (frame->thrown.flags & SCRIPTGO_VALUE_OWNED) == 0) {
            return source != NULL ? source : "";
        }
        if (frame->thrown.aux > (uint64_t)SIZE_MAX - 1) return "";
        free(scriptgo_exception_string_scratch);
        scriptgo_exception_string_scratch = NULL;
        copy = malloc((size_t)frame->thrown.aux + 1);
        if (copy == NULL) return "";
        if (frame->thrown.aux != 0 && source != NULL) {
            memcpy(copy, source, (size_t)frame->thrown.aux);
        }
        copy[frame->thrown.aux] = '\0';
        scriptgo_exception_string_scratch = copy;
        return scriptgo_exception_string_scratch;
    }
    return "";
}

double scriptgo_exception_get_number(scriptgo_exception_frame_t *frame) {
    double value = 0.0;
    if (frame != NULL && frame->thrown.tag == SCRIPTGO_TAG_NUMBER) {
        memcpy(&value, &frame->thrown.payload, sizeof(value));
    }
    return value;
}

int scriptgo_exception_get_bool(scriptgo_exception_frame_t *frame) {
    return frame != NULL && frame->thrown.tag == SCRIPTGO_TAG_BOOLEAN && frame->thrown.payload != 0;
}

int scriptgo_exception_get_tag_payload(scriptgo_exception_frame_t *frame, uint32_t *out_tag, uint64_t *out_payload) {
    if (frame == NULL || out_tag == NULL || out_payload == NULL) return -1;
    *out_tag = frame->thrown.tag;
    *out_payload = frame->thrown.payload;
    return 0;
}

void scriptgo_exception_rethrow(scriptgo_exception_frame_t *frame) {
    if (frame == NULL) return;
    scriptgo_value value;
    scriptgo_value_init_undefined(&value);
    scriptgo_exception_take_value(frame, &value);
    scriptgo_exception_frame_free(frame);
    scriptgo_exception_throw_move(&value);
}

void scriptgo_debugger_break(const char *file, int line) {
    (void)file;
    (void)line;
}
