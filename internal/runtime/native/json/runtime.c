#include <ctype.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

static int json_fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_json_stringify_number(double value, char **out_str) {
    char buf[64];
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (isnan(value) || isinf(value)) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (value == (double)(int64_t)value && fabs(value) < 1e15) {
        snprintf(buf, sizeof(buf), "%lld", (long long)value);
    } else {
        snprintf(buf, sizeof(buf), "%g", value);
    }
    *out_str = strdup(buf);
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_json_stringify_bool(int value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    *out_str = strdup(value ? "true" : "false");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_json_stringify_string(const char *value, char **out_str) {
    if (value == NULL || out_str == NULL) return json_fail("scriptgo json invalid argument");
    size_t len = strlen(value);
    size_t cap = len * 2 + 3;
    char *buf = malloc(cap);
    if (buf == NULL) return json_fail("scriptgo json allocation failed");
    size_t pos = 0;
    buf[pos++] = '"';
    for (size_t i = 0; i < len; i++) {
        unsigned char c = (unsigned char)value[i];
        if (c == '"') {
            buf[pos++] = '\\'; buf[pos++] = '"';
        } else if (c == '\\') {
            buf[pos++] = '\\'; buf[pos++] = '\\';
        } else if (c == '\n') {
            buf[pos++] = '\\'; buf[pos++] = 'n';
        } else if (c == '\r') {
            buf[pos++] = '\\'; buf[pos++] = 'r';
        } else if (c == '\t') {
            buf[pos++] = '\\'; buf[pos++] = 't';
        } else {
            buf[pos++] = (char)c;
        }
    }
    buf[pos++] = '"';
    buf[pos] = '\0';
    *out_str = buf;
    return 0;
}

int scriptgo_array_join_number(void *handle, const char *separator, char **out_str);

int scriptgo_json_stringify_number_array(void *handle, char **out_str) {
    char *joined = NULL;
    if (handle == NULL || out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (scriptgo_array_join_number(handle, ",", &joined) != 0) return -1;
    size_t len = strlen(joined);
    char *res = malloc(len + 3);
    if (res == NULL) { free(joined); return json_fail("scriptgo json allocation failed"); }
    res[0] = '[';
    memcpy(res + 1, joined, len);
    res[len + 1] = ']';
    res[len + 2] = '\0';
    free(joined);
    *out_str = res;
    return 0;
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_array_internal;

int scriptgo_json_stringify_string_array(void *handle, char **out_str) {
    scriptgo_array_internal *array = handle;
    if (array == NULL || out_str == NULL || array->element_size != sizeof(char *)) return json_fail("scriptgo json invalid argument");
    size_t cap = 256, len = 0;
    char *buf = malloc(cap);
    if (buf == NULL) return json_fail("scriptgo json allocation failed");
    buf[len++] = '[';
    buf[len] = '\0';
    for (int64_t i = 0; i < array->length; i++) {
        const char *elem = *(const char **)(array->data + (size_t)i * sizeof(char *));
        char *str_json = NULL;
        if (scriptgo_json_stringify_string(elem != NULL ? elem : "", &str_json) != 0) { free(buf); return -1; }
        size_t s_len = strlen(str_json);
        while (len + s_len + 3 >= cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) { free(str_json); free(buf); return json_fail("scriptgo json allocation failed"); }
            buf = new_buf;
        }
        if (i > 0) {
            buf[len++] = ',';
        }
        memcpy(buf + len, str_json, s_len);
        len += s_len;
        free(str_json);
    }
    buf[len++] = ']';
    buf[len] = '\0';
    *out_str = buf;
    return 0;
}

int scriptgo_json_parse_string(const char *input, char **out_str) {
    if (input == NULL || out_str == NULL) return json_fail("scriptgo json invalid argument");
    while (*input == ' ' || *input == '\t' || *input == '\n' || *input == '\r') input++;
    if (*input == '"') {
        size_t len = strlen(input);
        if (len >= 2 && input[len - 1] == '"') {
            char *res = malloc(len - 1);
            if (res == NULL) return json_fail("scriptgo json allocation failed");
            size_t pos = 0;
            for (size_t i = 1; i < len - 1; i++) {
                if (input[i] == '\\' && i + 1 < len - 1) {
                    i++;
                    if (input[i] == 'n') res[pos++] = '\n';
                    else if (input[i] == 'r') res[pos++] = '\r';
                    else if (input[i] == 't') res[pos++] = '\t';
                    else if (input[i] == '"') res[pos++] = '"';
                    else if (input[i] == '\\') res[pos++] = '\\';
                    else res[pos++] = input[i];
                } else {
                    res[pos++] = input[i];
                }
            }
            res[pos] = '\0';
            *out_str = res;
            return 0;
        }
    }
    *out_str = strdup(input);
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_string_from_object(void *obj, char **out_str);

int scriptgo_json_stringify_unknown(uint32_t tag, uint32_t padding, uint64_t payload, char **out_str) {
    (void)padding;
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    switch (tag) {
    case 0: // undefined
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    case 1: // null
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    case 2: // boolean
        return scriptgo_json_stringify_bool((int)payload, out_str);
    case 3: { // number
        union {
            uint64_t u64;
            double d;
        } u;
        u.u64 = payload;
        return scriptgo_json_stringify_number(u.d, out_str);
    }
    case 4: // string
        return scriptgo_json_stringify_string((const char *)(uintptr_t)payload, out_str);
    default:
        if (payload == 0) {
            *out_str = strdup("null");
            return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
        }
        return scriptgo_string_from_object((void *)(uintptr_t)payload, out_str);
    }
}

