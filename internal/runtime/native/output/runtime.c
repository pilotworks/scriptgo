#include <stdio.h>

int scriptgo_runtime_set_error(const char *message);

static int scriptgo_console_number(FILE *stream, double value) {
    int ret;
    if (value == (double)(long long)value && value >= -9007199254740991.0 && value <= 9007199254740991.0) {
        ret = fprintf(stream, "%lld\n", (long long)value);
    } else {
        ret = fprintf(stream, "%.15g\n", value);
    }
    fflush(stream);
    if (ret < 0) return scriptgo_runtime_set_error("scriptgo number output failed");
    return 0;
}

static int scriptgo_console_string(FILE *stream, const char *value) {
    if (value == NULL || fputs(value, stream) == EOF || fputc('\n', stream) == EOF) return scriptgo_runtime_set_error("scriptgo string output failed");
    fflush(stream);
    return 0;
}

static int scriptgo_console_bool(FILE *stream, int value) {
    if (fputs(value ? "true\n" : "false\n", stream) == EOF) return scriptgo_runtime_set_error("scriptgo boolean output failed");
    fflush(stream);
    return 0;
}

static int scriptgo_console_unknown(FILE *stream, ScriptGoUnknown value) {
    if (value.tag == SCRIPTGO_TAG_UNDEFINED) {
        return scriptgo_console_string(stream, "undefined");
    } else if (value.tag == SCRIPTGO_TAG_NULL) {
        return scriptgo_console_string(stream, "null");
    } else if (value.tag == SCRIPTGO_TAG_BOOLEAN) {
        return scriptgo_console_bool(stream, (int)value.payload);
    } else if (value.tag == SCRIPTGO_TAG_NUMBER) {
        union { unsigned long long raw; double num; } u;
        u.raw = value.payload;
        return scriptgo_console_number(stream, u.num);
    } else if (value.tag == SCRIPTGO_TAG_STRING) {
        return scriptgo_console_string(stream, (const char *)value.payload);
    } else {
        return scriptgo_console_string(stream, "[object Object]");
    }
}

static int scriptgo_console_bigint(FILE *stream, long long value) {
    int ret = fprintf(stream, "%lldn\n", value);
    fflush(stream);
    if (ret < 0) return scriptgo_runtime_set_error("scriptgo bigint output failed");
    return 0;
}

#define SCRIPTGO_CONSOLE_METHOD(name, stream) \
    int scriptgo_console_##name##_number(double value) { return scriptgo_console_number(stream, value); } \
    int scriptgo_console_##name##_bigint(long long value) { return scriptgo_console_bigint(stream, value); } \
    int scriptgo_console_##name##_string(const char *value) { return scriptgo_console_string(stream, value); } \
    int scriptgo_console_##name##_bool(int value) { return scriptgo_console_bool(stream, value); } \
    int scriptgo_console_##name##_unknown(ScriptGoUnknown value) { return scriptgo_console_unknown(stream, value); }

SCRIPTGO_CONSOLE_METHOD(log, stdout)
SCRIPTGO_CONSOLE_METHOD(info, stdout)
SCRIPTGO_CONSOLE_METHOD(warn, stderr)
SCRIPTGO_CONSOLE_METHOD(error, stderr)
