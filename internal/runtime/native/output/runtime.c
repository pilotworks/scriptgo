#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <time.h>

int scriptgo_runtime_set_error(const char *message);

static int scriptgo_console_indent_level = 0;

static void scriptgo_console_print_indent(FILE *stream) {
    for (int i = 0; i < scriptgo_console_indent_level * 2; i++) {
        fputc(' ', stream);
    }
}

static void scriptgo_format_double_shortest(char *buf, size_t size, double value) {
    if (isnan(value)) {
        snprintf(buf, size, "NaN");
        return;
    }
    if (isinf(value)) {
        if (value > 0) snprintf(buf, size, "Infinity");
        else snprintf(buf, size, "-Infinity");
        return;
    }
    if (value == 0.0 && signbit(value)) {
        snprintf(buf, size, "-0");
        return;
    }
    char b15[64], b16[64], b17[64];
    snprintf(b15, sizeof(b15), "%.15g", value);
    if (strtod(b15, NULL) == value) {
        snprintf(buf, size, "%s", b15);
        return;
    }
    snprintf(b16, sizeof(b16), "%.16g", value);
    if (strtod(b16, NULL) == value) {
        snprintf(buf, size, "%s", b16);
        return;
    }
    snprintf(b17, sizeof(b17), "%.17g", value);
    snprintf(buf, size, "%s", b17);
}

static int scriptgo_console_number(FILE *stream, double value) {
    scriptgo_console_print_indent(stream);
    int ret;
    if (value == 0.0 && signbit(value)) {
        ret = fprintf(stream, "-0\n");
    } else if (!isnan(value) && !isinf(value) && value == (double)(long long)value && value >= -9007199254740991.0 && value <= 9007199254740991.0) {
        ret = fprintf(stream, "%lld\n", (long long)value);
    } else {
        char buf[64];
        scriptgo_format_double_shortest(buf, sizeof(buf), value);
        ret = fprintf(stream, "%s\n", buf);
    }
    fflush(stream);
    if (ret < 0) return scriptgo_runtime_set_error("scriptgo number output failed");
    return 0;
}

extern const char scriptgo_undefined_sentinel;

static int scriptgo_console_string(FILE *stream, const char *value) {
    scriptgo_console_print_indent(stream);
    if (value == NULL) value = "null";
    else if (value == &scriptgo_undefined_sentinel) value = "undefined";
    if (fputs(value, stream) == EOF || fputc('\n', stream) == EOF) return scriptgo_runtime_set_error("scriptgo string output failed");
    fflush(stream);
    return 0;
}

static int scriptgo_console_bool(FILE *stream, int value) {
    scriptgo_console_print_indent(stream);
    if (fputs(value ? "true\n" : "false\n", stream) == EOF) return scriptgo_runtime_set_error("scriptgo boolean output failed");
    fflush(stream);
    return 0;
}

static int scriptgo_console_bigint(FILE *stream, long long value);

static int scriptgo_console_unknown(FILE *stream, unsigned int tag, unsigned int flags, unsigned long long payload) {
    if (tag == SCRIPTGO_TAG_UNDEFINED) {
        return scriptgo_console_string(stream, "undefined");
    } else if (tag == SCRIPTGO_TAG_NULL) {
        return scriptgo_console_string(stream, "null");
    } else if (tag == SCRIPTGO_TAG_BOOLEAN) {
        return scriptgo_console_bool(stream, (int)payload);
    } else if (tag == SCRIPTGO_TAG_NUMBER) {
        union { unsigned long long raw; double num; } u;
        u.raw = payload;
        return scriptgo_console_number(stream, u.num);
    } else if (tag == SCRIPTGO_TAG_STRING) {
        return scriptgo_console_string(stream, (const char *)payload);
    } else if (tag == SCRIPTGO_TAG_BIGINT) {
        return scriptgo_console_bigint(stream, (long long)payload);
    } else {
        return scriptgo_console_string(stream, "[object Object]");
    }
}

static int scriptgo_console_bigint(FILE *stream, long long value) {
    scriptgo_console_print_indent(stream);
    int ret = fprintf(stream, "%lldn\n", value);
    fflush(stream);
    if (ret < 0) return scriptgo_runtime_set_error("scriptgo bigint output failed");
    return 0;
}

int scriptgo_symbol_to_string(void *symbol, char **out_string);

static int scriptgo_console_symbol(FILE *stream, void *value) {
    if (value == NULL) {
        return scriptgo_console_string(stream, "Symbol()");
    }
    char *str = NULL;
    int err = scriptgo_symbol_to_string(value, &str);
    if (err != 0 || str == NULL) return err;
    int res = scriptgo_console_string(stream, str);
    free(str);
    return res;
}

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
} scriptgo_array_raw_t;

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} scriptgo_object_raw_t;

int scriptgo_string_from_object(void *obj, char **out_str);

static int scriptgo_console_array(FILE *stream, scriptgo_array_raw_t *arr) {
    if (arr == NULL) return scriptgo_console_string(stream, "null");
    if (arr == (scriptgo_array_raw_t *)&scriptgo_undefined_sentinel) return scriptgo_console_string(stream, "undefined");
    if (arr->length == 0) return scriptgo_console_string(stream, "[]");

    scriptgo_console_print_indent(stream);
    fprintf(stream, "[ ");
    for (int64_t i = 0; i < arr->length; i++) {
        if (i > 0) fprintf(stream, ", ");
        if (arr->element_size == 1) {
            uint8_t b = *(uint8_t *)(arr->data + (size_t)i);
            fprintf(stream, "%s", b ? "true" : "false");
        } else if (arr->element_size == sizeof(double)) {
            double d = *(double *)(arr->data + (size_t)i * sizeof(double));
            char buf[64];
            scriptgo_format_double_shortest(buf, sizeof(buf), d);
            fprintf(stream, "%s", buf);
        } else if (arr->element_size == sizeof(char *)) {
            const char *s = *(const char **)(arr->data + (size_t)i * sizeof(char *));
            if (s == NULL) fprintf(stream, "null");
            else if (s == &scriptgo_undefined_sentinel) fprintf(stream, "undefined");
            else fprintf(stream, "'%s'", s);
        } else if (arr->element_size == 16) {
            uint32_t tag = *(uint32_t *)(arr->data + (size_t)i * 16);
            uint64_t payload = *(uint64_t *)(arr->data + (size_t)i * 16 + 8);
            if (tag == SCRIPTGO_TAG_UNDEFINED) fprintf(stream, "undefined");
            else if (tag == SCRIPTGO_TAG_NULL) fprintf(stream, "null");
            else if (tag == SCRIPTGO_TAG_BOOLEAN) fprintf(stream, "%s", payload ? "true" : "false");
            else if (tag == SCRIPTGO_TAG_NUMBER) {
                union { unsigned long long raw; double num; } u;
                u.raw = payload;
                char buf[64];
                scriptgo_format_double_shortest(buf, sizeof(buf), u.num);
                fprintf(stream, "%s", buf);
            } else if (tag == SCRIPTGO_TAG_STRING) {
                fprintf(stream, "'%s'", (const char *)payload);
            } else {
                fprintf(stream, "[object Object]");
            }
        } else {
            fprintf(stream, "[object Object]");
        }
    }
    fprintf(stream, " ]\n");
    fflush(stream);
    return 0;
}

int scriptgo_gc_get_tag(void *ptr);

static int scriptgo_console_object(FILE *stream, void *value) {
    if (value == NULL) {
        return scriptgo_console_string(stream, "null");
    }
    if (value == &scriptgo_undefined_sentinel) {
        return scriptgo_console_string(stream, "undefined");
    }
    int tag = scriptgo_gc_get_tag(value);
    if (tag == 1 /* SCRIPTGO_TYPE_OBJECT */) {
        scriptgo_object_raw_t *o = (scriptgo_object_raw_t *)value;
        if (o->magic == SCRIPTGO_OBJECT_MAGIC) {
            if (o->type_name != NULL && strstr(o->type_name, "Error") != NULL) {
                if (o->field_count > 0 && o->fields[0] != 0 && o->fields[0] != 0x7FF8000000000000ULL) {
                    return scriptgo_console_string(stream, (const char *)o->fields[0]);
                }
            }
            return scriptgo_console_string(stream, "[object Object]");
        }
    } else if (tag == 2 /* SCRIPTGO_TYPE_ARRAY */) {
        scriptgo_array_raw_t *arr = (scriptgo_array_raw_t *)value;
        return scriptgo_console_array(stream, arr);
    }
    char *str = NULL;
    int err = scriptgo_string_from_object(value, &str);
    if (err == 0 && str != NULL) {
        return scriptgo_console_string(stream, str);
    }
    return scriptgo_console_string(stream, (const char *)value);
}

#define SCRIPTGO_CONSOLE_METHOD(name, stream) \
    int scriptgo_console_##name##_number(double value) { return scriptgo_console_number(stream, value); } \
    int scriptgo_console_##name##_bigint(long long value) { return scriptgo_console_bigint(stream, value); } \
    int scriptgo_console_##name##_symbol(void *value) { return scriptgo_console_symbol(stream, value); } \
    int scriptgo_console_##name##_string(const char *value) { return scriptgo_console_string(stream, value); } \
    int scriptgo_console_##name##_bool(int value) { return scriptgo_console_bool(stream, value); } \
    int scriptgo_console_##name##_unknown(unsigned int tag, unsigned int flags, unsigned long long payload) { return scriptgo_console_unknown(stream, tag, flags, payload); } \
    int scriptgo_console_##name##_object(void *value) { return scriptgo_console_object(stream, value); }

SCRIPTGO_CONSOLE_METHOD(log, stdout)
SCRIPTGO_CONSOLE_METHOD(info, stdout)
SCRIPTGO_CONSOLE_METHOD(debug, stdout)
SCRIPTGO_CONSOLE_METHOD(warn, stderr)
SCRIPTGO_CONSOLE_METHOD(error, stderr)

int scriptgo_console_clear(void) {
    fputs("\x1b[2J\x1b[0f", stdout);
    fflush(stdout);
    return 0;
}

int scriptgo_console_group(void) {
    scriptgo_console_indent_level++;
    return 0;
}

int scriptgo_console_group_end(void) {
    if (scriptgo_console_indent_level > 0) {
        scriptgo_console_indent_level--;
    }
    return 0;
}

typedef struct {
    char *label;
    int count;
} ScriptGoConsoleCounter;

static ScriptGoConsoleCounter scriptgo_counters[256];
static int scriptgo_counter_count = 0;

int scriptgo_console_count(const char *label) {
    const char *lbl = (label == NULL || strlen(label) == 0) ? "default" : label;
    for (int i = 0; i < scriptgo_counter_count; i++) {
        if (strcmp(scriptgo_counters[i].label, lbl) == 0) {
            scriptgo_counters[i].count++;
            scriptgo_console_print_indent(stdout);
            fprintf(stdout, "%s: %d\n", lbl, scriptgo_counters[i].count);
            fflush(stdout);
            return 0;
        }
    }
    if (scriptgo_counter_count < 256) {
        scriptgo_counters[scriptgo_counter_count].label = strdup(lbl);
        scriptgo_counters[scriptgo_counter_count].count = 1;
        scriptgo_counter_count++;
        scriptgo_console_print_indent(stdout);
        fprintf(stdout, "%s: 1\n", lbl);
        fflush(stdout);
        return 0;
    }
    return 0;
}

int scriptgo_console_count_reset(const char *label) {
    const char *lbl = (label == NULL || strlen(label) == 0) ? "default" : label;
    for (int i = 0; i < scriptgo_counter_count; i++) {
        if (strcmp(scriptgo_counters[i].label, lbl) == 0) {
            scriptgo_counters[i].count = 0;
            return 0;
        }
    }
    fprintf(stderr, "Count for '%s' does not exist\n", lbl);
    fflush(stderr);
    return 0;
}

typedef struct {
    char *label;
    double start_ms;
    int active;
} ScriptGoConsoleTimer;

static ScriptGoConsoleTimer scriptgo_timers[256];
static int scriptgo_timer_count = 0;

static double scriptgo_console_now_ms(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (double)tv.tv_sec * 1000.0 + (double)tv.tv_usec / 1000.0;
}

int scriptgo_console_time(const char *label) {
    const char *lbl = (label == NULL || strlen(label) == 0) ? "default" : label;
    for (int i = 0; i < scriptgo_timer_count; i++) {
        if (scriptgo_timers[i].active && strcmp(scriptgo_timers[i].label, lbl) == 0) {
            fprintf(stderr, "Timer '%s' already exists\n", lbl);
            fflush(stderr);
            return 0;
        }
    }
    double now = scriptgo_console_now_ms();
    for (int i = 0; i < scriptgo_timer_count; i++) {
        if (!scriptgo_timers[i].active) {
            free(scriptgo_timers[i].label);
            scriptgo_timers[i].label = strdup(lbl);
            scriptgo_timers[i].start_ms = now;
            scriptgo_timers[i].active = 1;
            return 0;
        }
    }
    if (scriptgo_timer_count < 256) {
        scriptgo_timers[scriptgo_timer_count].label = strdup(lbl);
        scriptgo_timers[scriptgo_timer_count].start_ms = now;
        scriptgo_timers[scriptgo_timer_count].active = 1;
        scriptgo_timer_count++;
        return 0;
    }
    return 0;
}

int scriptgo_console_time_log(const char *label, const char *data) {
    const char *lbl = (label == NULL || strlen(label) == 0) ? "default" : label;
    double now = scriptgo_console_now_ms();
    for (int i = 0; i < scriptgo_timer_count; i++) {
        if (scriptgo_timers[i].active && strcmp(scriptgo_timers[i].label, lbl) == 0) {
            double elapsed = now - scriptgo_timers[i].start_ms;
            scriptgo_console_print_indent(stdout);
            if (data && strlen(data) > 0) {
                fprintf(stdout, "%s: %.3fms %s\n", lbl, elapsed, data);
            } else {
                fprintf(stdout, "%s: %.3fms\n", lbl, elapsed);
            }
            fflush(stdout);
            return 0;
        }
    }
    fprintf(stderr, "Timer '%s' does not exist\n", lbl);
    fflush(stderr);
    return 0;
}

int scriptgo_console_time_end(const char *label) {
    const char *lbl = (label == NULL || strlen(label) == 0) ? "default" : label;
    double now = scriptgo_console_now_ms();
    for (int i = 0; i < scriptgo_timer_count; i++) {
        if (scriptgo_timers[i].active && strcmp(scriptgo_timers[i].label, lbl) == 0) {
            double elapsed = now - scriptgo_timers[i].start_ms;
            scriptgo_timers[i].active = 0;
            scriptgo_console_print_indent(stdout);
            fprintf(stdout, "%s: %.3fms\n", lbl, elapsed);
            fflush(stdout);
            return 0;
        }
    }
    fprintf(stderr, "Timer '%s' does not exist\n", lbl);
    fflush(stderr);
    return 0;
}

int scriptgo_console_trace(const char *msg) {
    scriptgo_console_print_indent(stderr);
    if (msg && strlen(msg) > 0) {
        fprintf(stderr, "Trace: %s\n", msg);
    } else {
        fprintf(stderr, "Trace\n");
    }
    fflush(stderr);
    return 0;
}

