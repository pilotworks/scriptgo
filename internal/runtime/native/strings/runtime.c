#include <math.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

static int string_fail(const char *message) { return scriptgo_runtime_set_error(message); }

static size_t normalize_position(double value, size_t length) {
    if (isnan(value) || value <= 0.0) return 0;
    if (value >= (double)length) return length;
    return (size_t)value;
}

static int string_copy_range(const char *value, size_t start, size_t length, char **out) {
    char *result;
    if (value == NULL || out == NULL) return string_fail("scriptgo string argument is invalid");
    result = malloc(length + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, value + start, length);
    result[length] = '\0';
    *out = result;
    return 0;
}

int scriptgo_string_concat(const char *left, const char *right, char **out_value) {
    size_t left_length, right_length;
    char *result;
    if (left == NULL || right == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    left_length = strlen(left);
    right_length = strlen(right);
    result = malloc(left_length + right_length + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, left, left_length);
    memcpy(result + left_length, right, right_length + 1);
    *out_value = result;
    return 0;
}

int scriptgo_string_length(const char *value, double *out_length) {
    if (value == NULL || out_length == NULL) return string_fail("scriptgo string argument is invalid");
    *out_length = (double)strlen(value);
    return 0;
}

int scriptgo_string_last_index(const char *value, const char *needle, double position, double *out_index) {
    const char *last = NULL;
    const char *cursor;
    size_t limit, normalized;
    if (value == NULL || needle == NULL || out_index == NULL) return string_fail("scriptgo string argument is invalid");
    limit = strlen(value);
    if (position >= 0.0) {
        normalized = normalize_position(position, limit);
        limit = normalized + (normalized < limit ? 1 : 0);
    }
    if (*needle == '\0') {
        *out_index = position >= 0.0 ? (double)normalized : (double)limit;
        return 0;
    }
    cursor = value;
    while ((cursor = strstr(cursor, needle)) != NULL && (size_t)(cursor - value) < limit) {
        last = cursor;
        cursor++;
    }
    *out_index = last == NULL ? -1.0 : (double)(last - value);
    return 0;
}

int scriptgo_string_slice(const char *value, double start_value, double end_value, char **out_value) {
    size_t length, start, end;
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    length = strlen(value);
    start = normalize_position(start_value, length);
    if (end_value < 0.0) {
        end = length;
    } else {
        end = normalize_position(end_value, length);
    }
    if (end < start) end = start;
    return string_copy_range(value, start, end - start, out_value);
}

int scriptgo_string_release(char *value) {
    free(value);
    return 0;
}
