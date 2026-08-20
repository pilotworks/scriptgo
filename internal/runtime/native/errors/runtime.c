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
    SCRIPTGO_TAG_FUNCTION  = 7
} ScriptGoTypeTag;

typedef struct {
    unsigned int tag;
    unsigned int flags;
    unsigned long long payload;
} ScriptGoUnknown;

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
    default:                     return "object";
    }
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

void scriptgo_throw_string(const char *str) {
    if (scriptgo_top_frame != NULL) {
        scriptgo_exception_frame_t *frame = scriptgo_top_frame;
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
        frame->thrown_bool = val;
        frame->thrown_type = 3;
        longjmp(frame->buf, 1);
    }
    fprintf(stderr, "Uncaught exception: %s\n", val ? "true" : "false");
    exit(1);
}

const char *scriptgo_exception_get_string(scriptgo_exception_frame_t *frame) {
    return frame->thrown_string ? frame->thrown_string : "";
}

double scriptgo_exception_get_number(scriptgo_exception_frame_t *frame) {
    return frame->thrown_number;
}

int scriptgo_exception_get_bool(scriptgo_exception_frame_t *frame) {
    return frame->thrown_bool;
}

void scriptgo_exception_rethrow(scriptgo_exception_frame_t *frame) {
    if (frame->thrown_type == 1) {
        scriptgo_throw_string(frame->thrown_string);
    } else if (frame->thrown_type == 2) {
        scriptgo_throw_number(frame->thrown_number);
    } else if (frame->thrown_type == 3) {
        scriptgo_throw_bool(frame->thrown_bool);
    }
}
