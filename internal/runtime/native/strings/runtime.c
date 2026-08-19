#include <math.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>

static char *string_copy_range(const char *value, size_t start, size_t length) {
    char *result = malloc(length + 1);
    if (result == NULL) {
        abort();
    }
    memcpy(result, value + start, length);
    result[length] = '\0';
    return result;
}

char *scriptgo_string_concat(const char *left, const char *right) {
    size_t left_length = strlen(left);
    size_t right_length = strlen(right);
    char *result = malloc(left_length + right_length + 1);
    if (result == NULL) {
        abort();
    }
    memcpy(result, left, left_length);
    memcpy(result + left_length, right, right_length + 1);
    return result;
}

double scriptgo_string_length(const char *value) {
    return (double)strlen(value);
}

double scriptgo_string_last_index(const char *value, const char *needle, double position) {
    const char *last = NULL;
    const char *cursor = value;
    if (*needle == '\0') {
        return (double)strlen(value);
    }
    size_t limit = strlen(value);
    if (position >= 0.0 && (size_t)position < limit) {
        limit = (size_t)position + 1;
    }
    while ((cursor = strstr(cursor, needle)) != NULL && (size_t)(cursor - value) < limit) {
        last = cursor;
        cursor++;
    }
    return last == NULL ? -1.0 : (double)(last - value);
}

char *scriptgo_string_slice(const char *value, double start_value, double end_value) {
    size_t length = strlen(value);
    size_t start = (size_t)fmax(0.0, fmin(start_value, (double)length));
    size_t end = end_value < 0.0 ? length : (size_t)fmax(0.0, fmin(end_value, (double)length));
    if (end < start) {
        end = start;
    }
    return string_copy_range(value, start, end - start);
}
