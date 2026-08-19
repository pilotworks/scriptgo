#include <ctype.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
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

int scriptgo_string_index_of(const char *value, const char *needle, double position, double *out_index) {
    const char *found;
    size_t start, length;
    if (value == NULL || needle == NULL || out_index == NULL) return string_fail("scriptgo string argument is invalid");
    length = strlen(value);
    start = normalize_position(position, length);
    if (start > length) {
        *out_index = -1.0;
        return 0;
    }
    if (*needle == '\0') {
        *out_index = (double)start;
        return 0;
    }
    found = strstr(value + start, needle);
    *out_index = found == NULL ? -1.0 : (double)(found - value);
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
    if (needle[1] == '\0') {
        const char *found = strrchr(value, needle[0]);
        if (found != NULL && (size_t)(found - value) < limit) {
            *out_index = (double)(found - value);
            return 0;
        }
        *out_index = -1.0;
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

int scriptgo_string_starts_with(const char *value, const char *prefix, double *out_bool) {
    size_t prefix_len, value_len;
    if (value == NULL || prefix == NULL || out_bool == NULL) return string_fail("scriptgo string argument is invalid");
    prefix_len = strlen(prefix);
    value_len = strlen(value);
    if (prefix_len > value_len) {
        *out_bool = 0.0;
        return 0;
    }
    *out_bool = strncmp(value, prefix, prefix_len) == 0 ? 1.0 : 0.0;
    return 0;
}

int scriptgo_string_ends_with(const char *value, const char *suffix, double *out_bool) {
    size_t suffix_len, value_len;
    if (value == NULL || suffix == NULL || out_bool == NULL) return string_fail("scriptgo string argument is invalid");
    suffix_len = strlen(suffix);
    value_len = strlen(value);
    if (suffix_len > value_len) {
        *out_bool = 0.0;
        return 0;
    }
    *out_bool = strcmp(value + value_len - suffix_len, suffix) == 0 ? 1.0 : 0.0;
    return 0;
}

int scriptgo_string_from_number(double value, char **out_value) {
    char buf[64];
    size_t length;
    char *result;
    if (out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (isnan(value)) {
        strcpy(buf, "NaN");
    } else if (isinf(value)) {
        if (value > 0) strcpy(buf, "Infinity");
        else strcpy(buf, "-Infinity");
    } else if (value == (double)(int64_t)value && fabs(value) < 1e15) {
        snprintf(buf, sizeof(buf), "%lld", (long long)value);
    } else {
        snprintf(buf, sizeof(buf), "%g", value);
    }
    length = strlen(buf);
    result = malloc(length + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, buf, length + 1);
    *out_value = result;
    return 0;
}

int scriptgo_string_from_bool(int value, char **out_value) {
    const char *str = value ? "true" : "false";
    size_t length = strlen(str);
    char *result;
    if (out_value == NULL) return string_fail("scriptgo string argument is invalid");
    result = malloc(length + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, str, length + 1);
    *out_value = result;
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

int scriptgo_string_trim(const char *value, char **out_value) {
    size_t start = 0, end;
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    while (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
        start++;
    }
    end = strlen(value);
    while (end > start && (value[end - 1] == ' ' || value[end - 1] == '\t' || value[end - 1] == '\n' || value[end - 1] == '\r')) {
        end--;
    }
    return string_copy_range(value, start, end - start, out_value);
}

int scriptgo_string_replace(const char *value, const char *search, const char *replacement, char **out_value) {
    const char *found;
    size_t val_len, search_len, rep_len, prefix_len;
    char *result;
    if (value == NULL || search == NULL || replacement == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    found = strstr(value, search);
    if (found == NULL) {
        return string_copy_range(value, 0, strlen(value), out_value);
    }
    prefix_len = (size_t)(found - value);
    search_len = strlen(search);
    rep_len = strlen(replacement);
    val_len = strlen(value);
    result = malloc(prefix_len + rep_len + (val_len - prefix_len - search_len) + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, value, prefix_len);
    memcpy(result + prefix_len, replacement, rep_len);
    memcpy(result + prefix_len + rep_len, found + search_len, val_len - prefix_len - search_len + 1);
    *out_value = result;
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_set_owned_data(void *handle, void *owned_data);
int scriptgo_array_release(void *handle);

int scriptgo_string_split(const char *value, const char *separator, void **out_array) {
    size_t val_len, sep_len, count = 1, buffer_size;
    char *buffer;
    if (value == NULL || separator == NULL || out_array == NULL) return string_fail("scriptgo string argument is invalid");
    val_len = strlen(value);
    sep_len = strlen(separator);

    if (sep_len == 0) {
        count = val_len;
    } else {
        const char *p = value;
        while ((p = sep_len == 1 ? strchr(p, separator[0]) : strstr(p, separator)) != NULL) {
            count++;
            p += sep_len;
        }
    }

    if (scriptgo_array_new((int64_t)count, sizeof(char *), out_array) != 0) return -1;
    if (sep_len == 0 && val_len > SIZE_MAX / 2) {
        scriptgo_array_release(*out_array);
        return string_fail("scriptgo string allocation failed");
    }
    buffer_size = sep_len == 0 ? val_len * 2 : val_len + 1;
    if (buffer_size == 0) buffer_size = 1;
    buffer = malloc(buffer_size);
    if (buffer == NULL) {
        scriptgo_array_release(*out_array);
        return string_fail("scriptgo string allocation failed");
    }
    if (scriptgo_array_set_owned_data(*out_array, buffer) != 0) {
        free(buffer);
        scriptgo_array_release(*out_array);
        return -1;
    }

    if (sep_len == 0) {
        for (size_t i = 0; i < val_len; i++) {
            buffer[i * 2] = value[i];
            buffer[i * 2 + 1] = '\0';
            char *sub = buffer + i * 2;
            if (scriptgo_array_set(*out_array, (double)i, &sub) != 0) return -1;
        }
        return 0;
    }

    size_t idx = 0;
    size_t output_offset = 0;
    const char *p = value;
    const char *next;
    while ((next = sep_len == 1 ? strchr(p, separator[0]) : strstr(p, separator)) != NULL) {
        size_t part_len = (size_t)(next - p);
        memcpy(buffer + output_offset, p, part_len);
        buffer[output_offset + part_len] = '\0';
        char *sub = buffer + output_offset;
        if (scriptgo_array_set(*out_array, (double)idx++, &sub) != 0) return -1;
        output_offset += part_len + 1;
        p = next + sep_len;
    }
    size_t tail_len = val_len - (size_t)(p - value);
    memcpy(buffer + output_offset, p, tail_len);
    buffer[output_offset + tail_len] = '\0';
    char *tail = buffer + output_offset;
    if (scriptgo_array_set(*out_array, (double)idx, &tail) != 0) return -1;
    return 0;
}

int scriptgo_string_char_at(const char *value, double pos, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    if (isnan(pos) || pos < 0.0 || pos >= (double)len) {
        return string_copy_range(value, 0, 0, out_value);
    }
    return string_copy_range(value, (size_t)pos, 1, out_value);
}

int scriptgo_string_char_code_at(const char *value, double pos, double *out_code) {
    if (value == NULL || out_code == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    if (isnan(pos) || pos < 0.0 || pos >= (double)len) {
        *out_code = NAN;
        return 0;
    }
    *out_code = (double)(unsigned char)value[(size_t)pos];
    return 0;
}

int scriptgo_string_includes(const char *value, const char *search, double pos, double *out_bool) {
    if (value == NULL || search == NULL || out_bool == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    size_t start = normalize_position(pos, len);
    if (start > len) {
        *out_bool = 0.0;
        return 0;
    }
    if (*search == '\0') {
        *out_bool = 1.0;
        return 0;
    }
    *out_bool = strstr(value + start, search) != NULL ? 1.0 : 0.0;
    return 0;
}

int scriptgo_string_to_lower(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    char *res = malloc(len + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    for (size_t i = 0; i < len; i++) {
        res[i] = (char)tolower((unsigned char)value[i]);
    }
    res[len] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_to_upper(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    char *res = malloc(len + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    for (size_t i = 0; i < len; i++) {
        res[i] = (char)toupper((unsigned char)value[i]);
    }
    res[len] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_trim_start(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t start = 0;
    while (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
        start++;
    }
    size_t len = strlen(value);
    return string_copy_range(value, start, len - start, out_value);
}

int scriptgo_string_trim_end(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t end = strlen(value);
    while (end > 0 && (value[end - 1] == ' ' || value[end - 1] == '\t' || value[end - 1] == '\n' || value[end - 1] == '\r')) {
        end--;
    }
    return string_copy_range(value, 0, end, out_value);
}

int scriptgo_string_repeat(const char *value, double count, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (isnan(count) || count <= 0.0) {
        return string_copy_range(value, 0, 0, out_value);
    }
    size_t c = (size_t)count;
    size_t len = strlen(value);
    char *res = malloc(len * c + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    for (size_t i = 0; i < c; i++) {
        memcpy(res + i * len, value, len);
    }
    res[len * c] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_replace_all(const char *value, const char *search, const char *replacement, char **out_value) {
    if (value == NULL || search == NULL || replacement == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t search_len = strlen(search);
    if (search_len == 0) {
        return string_copy_range(value, 0, strlen(value), out_value);
    }
    size_t rep_len = strlen(replacement);
    size_t count = 0;
    const char *p = value;
    while ((p = strstr(p, search)) != NULL) {
        count++;
        p += search_len;
    }
    if (count == 0) {
        return string_copy_range(value, 0, strlen(value), out_value);
    }
    size_t val_len = strlen(value);
    size_t new_len = val_len + count * rep_len - count * search_len;
    char *res = malloc(new_len + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    char *dst = res;
    p = value;
    const char *next;
    while ((next = strstr(p, search)) != NULL) {
        size_t part_len = (size_t)(next - p);
        memcpy(dst, p, part_len);
        dst += part_len;
        memcpy(dst, replacement, rep_len);
        dst += rep_len;
        p = next + search_len;
    }
    strcpy(dst, p);
    *out_value = res;
    return 0;
}

int scriptgo_string_pad_start(const char *value, double target_len, const char *pad_str, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (pad_str == NULL || *pad_str == '\0') pad_str = " ";
    size_t val_len = strlen(value);
    if (isnan(target_len) || target_len <= (double)val_len) {
        return string_copy_range(value, 0, val_len, out_value);
    }
    size_t target = (size_t)target_len;
    size_t diff = target - val_len;
    size_t pad_len = strlen(pad_str);
    char *res = malloc(target + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    for (size_t i = 0; i < diff; i++) {
        res[i] = pad_str[i % pad_len];
    }
    memcpy(res + diff, value, val_len + 1);
    *out_value = res;
    return 0;
}

int scriptgo_string_pad_end(const char *value, double target_len, const char *pad_str, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (pad_str == NULL || *pad_str == '\0') pad_str = " ";
    size_t val_len = strlen(value);
    if (isnan(target_len) || target_len <= (double)val_len) {
        return string_copy_range(value, 0, val_len, out_value);
    }
    size_t target = (size_t)target_len;
    size_t diff = target - val_len;
    size_t pad_len = strlen(pad_str);
    char *res = malloc(target + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(res, value, val_len);
    for (size_t i = 0; i < diff; i++) {
        res[val_len + i] = pad_str[i % pad_len];
    }
    res[target] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_release(char *value) {
    free(value);
    return 0;
}
