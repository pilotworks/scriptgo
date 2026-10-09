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

extern const char scriptgo_undefined_sentinel;

int scriptgo_string_compare(const char *left, const char *right) {
    if (left == right) return 0;
    if (left == &scriptgo_undefined_sentinel) left = "undefined";
    if (right == &scriptgo_undefined_sentinel) right = "undefined";
    if (left == NULL) return -1;
    if (right == NULL) return 1;
    return strcmp(left, right);
}

int scriptgo_string_concat(const char *left, const char *right, char **out_value) {
    if (out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (left == NULL) left = "null";
    else if (left == &scriptgo_undefined_sentinel) left = "undefined";
    if (right == NULL) right = "null";
    else if (right == &scriptgo_undefined_sentinel) right = "undefined";
    size_t left_length = strlen(left);
    size_t right_length = strlen(right);
    char *result = malloc(left_length + right_length + 1);
    if (result == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(result, left, left_length);
    memcpy(result + left_length, right, right_length + 1);
    *out_value = result;
    return 0;
}

static size_t utf8_to_utf16_length(const char *s) {
    if (s == NULL) return 0;
    size_t u16_len = 0;
    const unsigned char *p = (const unsigned char *)s;
    while (*p) {
        if (*p < 0x80) {
            u16_len++;
            p++;
        } else if ((*p & 0xE0) == 0xC0) {
            u16_len++;
            p += (p[1] ? 2 : 1);
        } else if ((*p & 0xF0) == 0xE0) {
            u16_len++;
            p += (p[1] && p[2] ? 3 : 1);
        } else if ((*p & 0xF8) == 0xF0) {
            u16_len += 2;
            p += (p[1] && p[2] && p[3] ? 4 : 1);
        } else {
            u16_len++;
            p++;
        }
    }
    return u16_len;
}

int scriptgo_string_length(const char *value, double *out_length) {
    if (out_length == NULL) return string_fail("scriptgo string argument is invalid");
    if (value == NULL) {
        *out_length = 0.0;
        return 0;
    }
    *out_length = (double)utf8_to_utf16_length(value);
    return 0;
}

int scriptgo_string_index_of(const char *value, const char *needle, double position, double *out_index) {
    if (position != position) position = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    const char *found;
    size_t start, length;
    if (value == NULL || needle == NULL || out_index == NULL) return string_fail("scriptgo string argument is invalid");
    length = utf16_length(value);
    start = normalize_position(position, length);
    if (*needle == '\0') {
        *out_index = (double)start;
        return 0;
    }
    found = strstr(value + utf16_byte_offset(value, start), needle);
    *out_index = found == NULL ? -1.0 : (double)utf16_unit_offset(value, (size_t)(found - value));
    return 0;
}

int scriptgo_string_last_index(const char *value, const char *needle, double position, double *out_index) {
    if (position != position) position = INFINITY; /* lastIndexOf: NaN position searches from the end */
    if (value == NULL || needle == NULL || out_index == NULL) return string_fail("scriptgo string argument is invalid");
    size_t length = utf16_length(value);
    size_t from = normalize_position(position, length);
    if (*needle == '\0') {
        *out_index = (double)from;
        return 0;
    }
    /* A match may start at any code unit up to `from`. */
    size_t max_start = utf16_byte_offset(value, from);
    const char *last = NULL;
    for (const char *cursor = value; (cursor = strstr(cursor, needle)) != NULL && (size_t)(cursor - value) <= max_start; cursor++) {
        last = cursor;
    }
    *out_index = last == NULL ? -1.0 : (double)utf16_unit_offset(value, (size_t)(last - value));
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
        const char *marker = scriptgo_number_marker_name(value);
        strcpy(buf, marker != NULL ? marker : "NaN");
    } else if (isinf(value)) {
        if (value > 0) strcpy(buf, "Infinity");
        else strcpy(buf, "-Infinity");
    } else {
        scriptgo_number_format(value, buf, sizeof(buf));
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
    if (start_value != start_value) start_value = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    size_t length, start, end;
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    length = utf16_length(value);
    int64_t s = (int64_t)start_value;
    if (s < 0) s = (int64_t)length + s;
    if (s < 0) s = 0;
    if ((size_t)s > length) s = (int64_t)length;
    start = (size_t)s;

    if (end_value >= 1e8) {
        end = length;
    } else {
        int64_t e = (int64_t)end_value;
        if (e < 0) e = (int64_t)length + e;
        if (e < 0) e = 0;
        if ((size_t)e > length) e = (int64_t)length;
        end = (size_t)e;
    }
    if (end < start) end = start;
    if (utf16_slice(value, start, end, out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
}

/* String.prototype.substring: indices clamp to [0, length] and the smaller
 * one starts the result. */
int scriptgo_string_substring(const char *value, double start_value, double end_value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t length = utf16_length(value);
    size_t start = normalize_position(start_value, length);
    size_t end = end_value != end_value ? 0 : normalize_position(end_value, length);
    if (start > end) {
        size_t swap = start;
        start = end;
        end = swap;
    }
    if (utf16_slice(value, start, end, out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
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

static char *expand_replacement(const char *rep, const char *match, size_t match_len, const char *full_str, size_t match_pos) {
    if (rep == NULL) return NULL;
    if (strchr(rep, '$') == NULL) {
        char *dup = malloc(strlen(rep) + 1);
        if (dup) strcpy(dup, rep);
        return dup;
    }
    size_t full_len = full_str ? strlen(full_str) : 0;
    size_t cap = strlen(rep) + match_len + full_len + 64;
    char *buf = malloc(cap);
    if (!buf) return NULL;
    size_t out_len = 0;
    const char *p = rep;
    while (*p) {
        if (*p == '$' && p[1] != '\0') {
            if (p[1] == '$') {
                buf[out_len++] = '$';
                p += 2;
            } else if (p[1] == '&') {
                if (match && match_len > 0) {
                    memcpy(buf + out_len, match, match_len);
                    out_len += match_len;
                }
                p += 2;
            } else if (p[1] == '`') {
                if (full_str && match_pos > 0) {
                    memcpy(buf + out_len, full_str, match_pos);
                    out_len += match_pos;
                }
                p += 2;
            } else if (p[1] == '\'') {
                if (full_str) {
                    size_t after_pos = match_pos + match_len;
                    size_t after_len = full_len > after_pos ? full_len - after_pos : 0;
                    memcpy(buf + out_len, full_str + after_pos, after_len);
                    out_len += after_len;
                }
                p += 2;
            } else {
                buf[out_len++] = *p++;
            }
        } else {
            buf[out_len++] = *p++;
        }
    }
    buf[out_len] = '\0';
    return buf;
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
    val_len = strlen(value);
    char *expanded = expand_replacement(replacement, search, search_len, value, prefix_len);
    if (expanded == NULL) return string_fail("scriptgo string allocation failed");
    rep_len = strlen(expanded);
    result = malloc(prefix_len + rep_len + (val_len - prefix_len - search_len) + 1);
    if (result == NULL) {
        free(expanded);
        return string_fail("scriptgo string allocation failed");
    }
    memcpy(result, value, prefix_len);
    memcpy(result + prefix_len, expanded, rep_len);
    memcpy(result + prefix_len + rep_len, found + search_len, val_len - prefix_len - search_len + 1);
    free(expanded);
    *out_value = result;
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_set_owned_data(void *handle, void *owned_data);
int scriptgo_array_release(void *handle);

extern const char scriptgo_undefined_sentinel;

int scriptgo_string_split(const char *value, const char *separator, double limit, void **out_array) {
    if (value == NULL || out_array == NULL) return string_fail("scriptgo string argument is invalid");
    /* An undefined separator does not split: the result is [value]. */
    if (separator == &scriptgo_undefined_sentinel) separator = NULL;
    if (!isnan(limit) && limit == 0.0) {
        return scriptgo_array_new(0, sizeof(char *), out_array);
    }
    size_t val_len = strlen(value);
    if (separator == NULL) {
        if (scriptgo_array_new(1, sizeof(char *), out_array) != 0) return -1;
        char *buffer = malloc(val_len + 1);
        if (buffer == NULL) {
            scriptgo_array_release(*out_array);
            return string_fail("scriptgo string allocation failed");
        }
        memcpy(buffer, value, val_len + 1);
        if (scriptgo_array_set_owned_data(*out_array, buffer) != 0) {
            free(buffer);
            scriptgo_array_release(*out_array);
            return -1;
        }
        char **data = (char **)((scriptgo_array *)(*out_array))->data;
        data[0] = buffer;
        return 0;
    }
    
    int has_capture = 0;
    const char *actual_sep = separator;
    char cap_buf[64] = {0};
    if (separator[0] == '(' && separator[strlen(separator)-1] == ')') {
        has_capture = 1;
        strncpy(cap_buf, separator + 1, strlen(separator) - 2);
        actual_sep = cap_buf;
    } else if (separator[0] == '/' && separator[1] == '(') {
        const char *close_p = strstr(separator, ")/");
        if (close_p != NULL) {
            has_capture = 1;
            strncpy(cap_buf, separator + 2, close_p - (separator + 2));
            actual_sep = cap_buf;
        }
    }
    
    size_t sep_len = strlen(actual_sep), count = 1, buffer_size;
    char *buffer;

    if (sep_len == 0) {
        count = utf16_length(value);
    } else {
        const char *p = value;
        while ((p = sep_len == 1 ? strchr(p, actual_sep[0]) : strstr(p, actual_sep)) != NULL) {
            count += (has_capture ? 2 : 1);
            p += sep_len;
        }
    }
    if (!isnan(limit) && limit > 0.0 && count > (size_t)limit) {
        count = (size_t)limit;
    }

    if (scriptgo_array_new((int64_t)count, sizeof(char *), out_array) != 0) return -1;
    buffer_size = val_len * 4 + 32;
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
        /* split("") yields each UTF-16 code unit; buffer holds val_len * 4 + 32
         * bytes, room for every unit (at most 3 bytes) and its terminator. */
        size_t offset = 0;
        for (size_t i = 0; i < count; i++) {
            char *unit = NULL;
            if (utf16_slice(value, i, i + 1, &unit) != 0) return string_fail("scriptgo string allocation failed");
            size_t unit_len = strlen(unit);
            memcpy(buffer + offset, unit, unit_len + 1);
            free(unit);
            char *sub = buffer + offset;
            offset += unit_len + 1;
            if (scriptgo_array_set(*out_array, (double)i, &sub) != 0) return -1;
        }
        return 0;
    }

    size_t idx = 0;
    size_t output_offset = 0;
    const char *p = value;
    const char *next;
    while (idx < count && (next = sep_len == 1 ? strchr(p, actual_sep[0]) : strstr(p, actual_sep)) != NULL) {
        size_t part_len = (size_t)(next - p);
        memcpy(buffer + output_offset, p, part_len);
        buffer[output_offset + part_len] = '\0';
        char *sub = buffer + output_offset;
        if (scriptgo_array_set(*out_array, (double)idx++, &sub) != 0) return -1;
        output_offset += part_len + 1;
        
        if (has_capture && idx < count) {
            memcpy(buffer + output_offset, actual_sep, sep_len);
            buffer[output_offset + sep_len] = '\0';
            char *cap_sub = buffer + output_offset;
            if (scriptgo_array_set(*out_array, (double)idx++, &cap_sub) != 0) return -1;
            output_offset += sep_len + 1;
        }
        p = next + sep_len;
    }
    if (idx < count) {
        size_t tail_len = val_len - (size_t)(p - value);
        memcpy(buffer + output_offset, p, tail_len);
        buffer[output_offset + tail_len] = '\0';
        char *tail = buffer + output_offset;
        if (scriptgo_array_set(*out_array, (double)idx, &tail) != 0) return -1;
    }
    return 0;
}

int scriptgo_string_char_at(const char *value, double pos, char **out_value) {
    if (pos != pos) pos = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = utf16_length(value);
    if (isnan(pos) || pos < 0.0 || pos >= (double)len) {
        return string_copy_range(value, 0, 0, out_value);
    }
    if (utf16_slice(value, (size_t)pos, (size_t)pos + 1, out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
}

int scriptgo_string_char_code_at(const char *value, double pos, double *out_code) {
    if (pos != pos) pos = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_code == NULL) return string_fail("scriptgo string argument is invalid");
    if (isnan(pos) || pos < 0.0) {
        *out_code = NAN;
        return 0;
    }
    size_t target_idx = (size_t)pos;
    size_t current_idx = 0;
    const unsigned char *p = (const unsigned char *)value;
    while (*p) {
        uint32_t cp = 0;
        size_t step = 0;
        if (*p < 0x80) {
            cp = *p;
            step = 1;
        } else if ((*p & 0xE0) == 0xC0 && p[1]) {
            cp = ((*p & 0x1F) << 6) | (p[1] & 0x3F);
            step = 2;
        } else if ((*p & 0xF0) == 0xE0 && p[1] && p[2]) {
            cp = ((*p & 0x0F) << 12) | ((p[1] & 0x3F) << 6) | (p[2] & 0x3F);
            step = 3;
        } else if ((*p & 0xF8) == 0xF0 && p[1] && p[2] && p[3]) {
            cp = ((*p & 0x07) << 18) | ((p[1] & 0x3F) << 12) | ((p[2] & 0x3F) << 6) | (p[3] & 0x3F);
            step = 4;
        } else {
            cp = *p;
            step = 1;
        }
        p += step;

        if (cp <= 0xFFFF) {
            if (current_idx == target_idx) {
                *out_code = (double)cp;
                return 0;
            }
            current_idx++;
        } else {
            uint32_t high = 0xD800 + ((cp - 0x10000) >> 10);
            uint32_t low = 0xDC00 + ((cp - 0x10000) & 0x3FF);
            if (current_idx == target_idx) {
                *out_code = (double)high;
                return 0;
            }
            current_idx++;
            if (current_idx == target_idx) {
                *out_code = (double)low;
                return 0;
            }
            current_idx++;
        }
    }
    *out_code = NAN;
    return 0;
}

int scriptgo_string_includes(const char *value, const char *search, double pos, double *out_bool) {
    if (pos != pos) pos = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || search == NULL || out_bool == NULL) return string_fail("scriptgo string argument is invalid");
    size_t start = utf16_byte_offset(value, normalize_position(pos, utf16_length(value)));
    if (*search == '\0') {
        *out_bool = 1.0;
        return 0;
    }
    *out_bool = strstr(value + start, search) != NULL ? 1.0 : 0.0;
    return 0;
}

int scriptgo_string_to_lower(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (utf8_map_case(value, unicode_to_lower, out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
}

int scriptgo_string_to_upper(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (utf8_map_case(value, unicode_to_upper, out_value) != 0) return string_fail("scriptgo string allocation failed");
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

/* String.prototype.repeat: ToIntegerOrInfinity(count); a negative or
 * infinite count, or a result longer than the maximum string length, throws
 * RangeError. Repeating "" (or zero times) is "" without iterating, and the
 * result is filled by doubling copies. */
#define SCRIPTGO_MAX_STRING_LENGTH ((size_t)0x1FFFFFE8)
void scriptgo_throw_error_message(const char *text);

int scriptgo_string_repeat(const char *value, double count, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    count = count != count ? 0.0 : trunc(count);
    if (count < 0.0 || isinf(count)) {
        char message[96];
        if (isinf(count)) {
            snprintf(message, sizeof(message), "RangeError: Invalid count value: %sInfinity", count < 0 ? "-" : "");
        } else {
            snprintf(message, sizeof(message), "RangeError: Invalid count value: %.0f", count);
        }
        scriptgo_throw_error_message(message);
        return string_fail("invalid repeat count");
    }
    size_t len = strlen(value);
    if (len == 0 || count == 0.0) {
        return string_copy_range(value, 0, 0, out_value);
    }
    if (count > (double)(SCRIPTGO_MAX_STRING_LENGTH / len)) {
        scriptgo_throw_error_message("RangeError: Invalid string length");
        return string_fail("repeat result too long");
    }
    size_t total = len * (size_t)count;
    char *res = malloc(total + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(res, value, len);
    size_t filled = len;
    while (filled < total) {
        size_t chunk = filled <= total - filled ? filled : total - filled;
        memcpy(res + filled, res, chunk);
        filled += chunk;
    }
    res[total] = '\0';
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

/* string_pad pads value to target_len UTF-16 code units with repetitions of
 * pad_str, in front (at_start) or behind. */
static int string_pad(const char *value, double target_len, const char *pad_str, int at_start, char **out_value) {
    if (target_len != target_len) target_len = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (pad_str == NULL || pad_str == &scriptgo_undefined_sentinel) pad_str = " ";
    size_t value_units = utf16_length(value);
    size_t pad_units = utf16_length(pad_str);
    if (target_len <= (double)value_units || pad_units == 0) {
        return string_copy_range(value, 0, strlen(value), out_value);
    }
    if (target_len > 1e9) return string_fail("RangeError: Invalid string length");
    size_t missing = (size_t)target_len - value_units;
    size_t repeats = missing / pad_units + 1;
    size_t pad_bytes = strlen(pad_str);
    char *repeated = malloc(repeats * pad_bytes + 1);
    if (repeated == NULL) return string_fail("scriptgo string allocation failed");
    for (size_t i = 0; i < repeats; i++) memcpy(repeated + i * pad_bytes, pad_str, pad_bytes);
    repeated[repeats * pad_bytes] = '\0';
    char *filler = NULL;
    int status = utf16_slice(repeated, 0, missing, &filler);
    free(repeated);
    if (status != 0) return string_fail("scriptgo string allocation failed");
    status = at_start ? scriptgo_string_concat(filler, value, out_value) : scriptgo_string_concat(value, filler, out_value);
    free(filler);
    return status;
}

int scriptgo_string_pad_start(const char *value, double target_len, const char *pad_str, char **out_value) {
    return string_pad(value, target_len, pad_str, 1, out_value);
}

int scriptgo_string_pad_end(const char *value, double target_len, const char *pad_str, char **out_value) {
    return string_pad(value, target_len, pad_str, 0, out_value);
}

int scriptgo_string_code_point_at(const char *value, double pos, double *out_code_point) {
    if (pos != pos) pos = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_code_point == NULL) return string_fail("scriptgo string argument is invalid");
    if (isnan(pos) || pos < 0.0) {
        *out_code_point = NAN;
        return 0;
    }
    size_t target_idx = (size_t)pos;
    size_t current_idx = 0;
    const unsigned char *p = (const unsigned char *)value;
    while (*p) {
        uint32_t cp = 0;
        size_t step = 0;
        if (*p < 0x80) {
            cp = *p;
            step = 1;
        } else if ((*p & 0xE0) == 0xC0 && p[1]) {
            cp = ((*p & 0x1F) << 6) | (p[1] & 0x3F);
            step = 2;
        } else if ((*p & 0xF0) == 0xE0 && p[1] && p[2]) {
            cp = ((*p & 0x0F) << 12) | ((p[1] & 0x3F) << 6) | (p[2] & 0x3F);
            step = 3;
        } else if ((*p & 0xF8) == 0xF0 && p[1] && p[2] && p[3]) {
            cp = ((*p & 0x07) << 18) | ((p[1] & 0x3F) << 12) | ((p[2] & 0x3F) << 6) | (p[3] & 0x3F);
            step = 4;
        } else {
            cp = *p;
            step = 1;
        }
        p += step;

        if (cp <= 0xFFFF) {
            if (current_idx == target_idx) {
                *out_code_point = (double)cp;
                return 0;
            }
            current_idx++;
        } else {
            if (current_idx == target_idx) {
                *out_code_point = (double)cp;
                return 0;
            }
            current_idx++;
            if (current_idx == target_idx) {
                uint32_t low = 0xDC00 + ((cp - 0x10000) & 0x3FF);
                *out_code_point = (double)low;
                return 0;
            }
            current_idx++;
        }
    }
    *out_code_point = NAN;
    return 0;
}

int scriptgo_string_from_code_point(double code_point, char **out_value) {
    if (out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (isnan(code_point) || code_point < 0.0 || code_point > 0x10FFFF) {
        return string_fail("Invalid code point");
    }
    uint32_t cp = (uint32_t)code_point;
    char buf[5] = {0};
    if (cp <= 0x7F) {
        buf[0] = (char)cp;
    } else if (cp <= 0x7FF) {
        buf[0] = (char)(0xC0 | (cp >> 6));
        buf[1] = (char)(0x80 | (cp & 0x3F));
    } else if (cp <= 0xFFFF) {
        buf[0] = (char)(0xE0 | (cp >> 12));
        buf[1] = (char)(0x80 | ((cp >> 6) & 0x3F));
        buf[2] = (char)(0x80 | (cp & 0x3F));
    } else {
        buf[0] = (char)(0xF0 | (cp >> 18));
        buf[1] = (char)(0x80 | ((cp >> 12) & 0x3F));
        buf[2] = (char)(0x80 | ((cp >> 6) & 0x3F));
        buf[3] = (char)(0x80 | (cp & 0x3F));
    }
    size_t blen = strlen(buf);
    char *res = malloc(blen + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    memcpy(res, buf, blen + 1);
    *out_value = res;
    return 0;
}

int scriptgo_string_is_well_formed(const char *value, double *out_bool) {
    if (value == NULL || out_bool == NULL) return string_fail("scriptgo string argument is invalid");
    const unsigned char *s = (const unsigned char *)value;
    int well_formed = 1;
    while (*s) {
        if (*s < 0x80) {
            s++;
        } else if ((*s & 0xE0) == 0xC0) {
            if ((s[1] & 0xC0) != 0x80 || (*s & 0x1E) == 0) { well_formed = 0; break; }
            s += 2;
        } else if ((*s & 0xF0) == 0xE0) {
            if ((s[1] & 0xC0) != 0x80 || (s[2] & 0xC0) != 0x80) { well_formed = 0; break; }
            uint32_t cp = ((*s & 0x0F) << 12) | ((s[1] & 0x3F) << 6) | (s[2] & 0x3F);
            if (cp >= 0xD800 && cp <= 0xDFFF) { well_formed = 0; break; }
            s += 3;
        } else if ((*s & 0xF8) == 0xF0) {
            if ((s[1] & 0xC0) != 0x80 || (s[2] & 0xC0) != 0x80 || (s[3] & 0xC0) != 0x80) { well_formed = 0; break; }
            s += 4;
        } else {
            well_formed = 0;
            break;
        }
    }
    *out_bool = well_formed ? 1.0 : 0.0;
    return 0;
}

int scriptgo_string_to_well_formed(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    char *res = malloc(len * 3 + 4);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    const unsigned char *s = (const unsigned char *)value;
    char *d = res;
    while (*s) {
        if (*s < 0x80) {
            *d++ = *s++;
        } else if ((*s & 0xE0) == 0xC0 && (s[1] & 0xC0) == 0x80) {
            *d++ = *s++;
            *d++ = *s++;
        } else if ((*s & 0xF0) == 0xE0 && (s[1] & 0xC0) == 0x80 && (s[2] & 0xC0) == 0x80) {
            uint32_t cp = ((*s & 0x0F) << 12) | ((s[1] & 0x3F) << 6) | (s[2] & 0x3F);
            if (cp >= 0xD800 && cp <= 0xDFFF) {
                *d++ = (char)0xEF; *d++ = (char)0xBF; *d++ = (char)0xBD;
                s += 3;
            } else {
                *d++ = *s++; *d++ = *s++; *d++ = *s++;
            }
        } else if ((*s & 0xF8) == 0xF0 && (s[1] & 0xC0) == 0x80 && (s[2] & 0xC0) == 0x80 && (s[3] & 0xC0) == 0x80) {
            *d++ = *s++; *d++ = *s++; *d++ = *s++; *d++ = *s++;
        } else {
            *d++ = (char)0xEF; *d++ = (char)0xBF; *d++ = (char)0xBD;
            s++;
        }
    }
    *d = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_release(char *value) {
    free(value);
    return 0;
}

int scriptgo_string_substr(const char *value, double start_val, double length_val, char **out_value) {
    if (start_val != start_val) start_val = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t length = utf16_length(value);
    int64_t start = (int64_t)start_val;
    if (start < 0) start = (int64_t)length + start;
    if (start < 0) start = 0;
    if ((size_t)start >= length) return string_copy_range(value, 0, 0, out_value);

    int64_t len = (int64_t)length_val;
    if (length_val >= 1e8) len = (int64_t)length - start;
    if (len <= 0) return string_copy_range(value, 0, 0, out_value);
    if (start + len > (int64_t)length) len = (int64_t)length - start;

    if (utf16_slice(value, (size_t)start, (size_t)(start + len), out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
}

int scriptgo_string_from_char_codes(const double *codes, int64_t count, char **out_value) {
    if (out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (count <= 0 || codes == NULL) {
        return string_copy_range("", 0, 0, out_value);
    }
    uint32_t *units = malloc((size_t)count * sizeof(uint32_t));
    if (units == NULL) return string_fail("scriptgo string allocation failed");
    for (int64_t i = 0; i < count; i++) {
        /* ToUint16 */
        double code = codes[i];
        units[i] = isfinite(code) ? (uint32_t)((int64_t)code & 0xFFFF) : 0;
    }
    int status = utf16_encode(units, (size_t)count, out_value);
    free(units);
    return status != 0 ? string_fail("scriptgo string allocation failed") : 0;
}

int scriptgo_string_at(const char *value, double pos, char **out_value) {
    if (pos != pos) pos = 0.0; /* ToIntegerOrInfinity(NaN) is 0 */
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = utf16_length(value);
    int64_t idx = (int64_t)pos;
    if (idx < 0) idx = (int64_t)len + idx;
    if (idx < 0 || (size_t)idx >= len) {
        return string_copy_range(value, 0, 0, out_value);
    }
    if (utf16_slice(value, (size_t)idx, (size_t)idx + 1, out_value) != 0) return string_fail("scriptgo string allocation failed");
    return 0;
}

int scriptgo_string_anchor(const char *value, const char *name, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (name == NULL) name = "";
    char buf[512];
    snprintf(buf, sizeof(buf), "<a name=\"%s\">%s</a>", name, value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_big(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<big>%s</big>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_blink(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<blink>%s</blink>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_bold(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<b>%s</b>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_fixed(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<tt>%s</tt>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_fontcolor(const char *value, const char *color, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (color == NULL) color = "";
    char buf[512];
    snprintf(buf, sizeof(buf), "<font color=\"%s\">%s</font>", color, value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_fontsize(const char *value, const char *size, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (size == NULL) size = "";
    char buf[512];
    snprintf(buf, sizeof(buf), "<font size=\"%s\">%s</font>", size, value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_italics(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<i>%s</i>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_link(const char *value, const char *url, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (url == NULL) url = "";
    char buf[512];
    snprintf(buf, sizeof(buf), "<a href=\"%s\">%s</a>", url, value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_small(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<small>%s</small>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_strike(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<strike>%s</strike>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_sub(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<sub>%s</sub>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_string_sup(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    char buf[512];
    snprintf(buf, sizeof(buf), "<sup>%s</sup>", value);
    char *res = strdup(buf);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    *out_value = res;
    return 0;
}

static int is_uri_component_unescaped(unsigned char c) {
    if ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) return 1;
    switch (c) {
        case '-': case '_': case '.': case '!': case '~': case '*': case '\'': case '(': case ')':
            return 1;
        default:
            return 0;
    }
}

static int is_uri_unescaped(unsigned char c) {
    if (is_uri_component_unescaped(c)) return 1;
    switch (c) {
        case ';': case ',': case '/': case '?': case ':': case '@': case '&': case '=': case '+': case '$': case '#':
            return 1;
        default:
            return 0;
    }
}

static int hex_digit_to_int(char c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    return -1;
}

int scriptgo_string_encode_uri_component(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    size_t cap = len * 3 + 1;
    char *res = malloc(cap);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    size_t out_idx = 0;
    static const char hex[] = "0123456789ABCDEF";
    for (size_t i = 0; i < len; i++) {
        unsigned char c = (unsigned char)value[i];
        if (is_uri_component_unescaped(c)) {
            res[out_idx++] = (char)c;
        } else {
            res[out_idx++] = '%';
            res[out_idx++] = hex[(c >> 4) & 0x0F];
            res[out_idx++] = hex[c & 0x0F];
        }
    }
    res[out_idx] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_decode_uri_component(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    char *res = malloc(len + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    size_t out_idx = 0;
    for (size_t i = 0; i < len; ) {
        if (value[i] == '%' && i + 2 < len) {
            int h1 = hex_digit_to_int(value[i + 1]);
            int h2 = hex_digit_to_int(value[i + 2]);
            if (h1 >= 0 && h2 >= 0) {
                res[out_idx++] = (char)((h1 << 4) | h2);
                i += 3;
                continue;
            }
        }
        res[out_idx++] = value[i++];
    }
    res[out_idx] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_encode_uri(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    size_t cap = len * 3 + 1;
    char *res = malloc(cap);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    size_t out_idx = 0;
    static const char hex[] = "0123456789ABCDEF";
    for (size_t i = 0; i < len; i++) {
        unsigned char c = (unsigned char)value[i];
        if (is_uri_unescaped(c)) {
            res[out_idx++] = (char)c;
        } else {
            res[out_idx++] = '%';
            res[out_idx++] = hex[(c >> 4) & 0x0F];
            res[out_idx++] = hex[c & 0x0F];
        }
    }
    res[out_idx] = '\0';
    *out_value = res;
    return 0;
}

int scriptgo_string_decode_uri(const char *value, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    size_t len = strlen(value);
    char *res = malloc(len + 1);
    if (res == NULL) return string_fail("scriptgo string allocation failed");
    size_t out_idx = 0;
    for (size_t i = 0; i < len; ) {
        if (value[i] == '%' && i + 2 < len) {
            int h1 = hex_digit_to_int(value[i + 1]);
            int h2 = hex_digit_to_int(value[i + 2]);
            if (h1 >= 0 && h2 >= 0) {
                unsigned char c = (unsigned char)((h1 << 4) | h2);
                if (c == ';' || c == ',' || c == '/' || c == '?' || c == ':' || c == '@' || c == '&' || c == '=' || c == '+' || c == '$' || c == '#') {
                    res[out_idx++] = value[i++];
                    res[out_idx++] = value[i++];
                    res[out_idx++] = value[i++];
                    continue;
                }
                res[out_idx++] = (char)c;
                i += 3;
                continue;
            }
        }
        res[out_idx++] = value[i++];
    }
    res[out_idx] = '\0';
    *out_value = res;
    return 0;
}



void scriptgo_throw_error_message(const char *text);

/* scriptgo_string_normalize implements String.prototype.normalize for the
 * input the runtime can decide without Unicode decomposition tables: the
 * form is validated (RangeError, as in JavaScript) and ASCII text, which
 * every normalization form leaves unchanged, is returned as is. Non-ASCII
 * text throws instead of returning an unnormalized string. */
int scriptgo_string_normalize(const char *value, const char *form, char **out_value) {
    if (value == NULL || out_value == NULL) return string_fail("scriptgo string argument is invalid");
    if (form != NULL && strcmp(form, "NFC") != 0 && strcmp(form, "NFD") != 0 &&
        strcmp(form, "NFKC") != 0 && strcmp(form, "NFKD") != 0) {
        scriptgo_throw_error_message("RangeError: The normalization form should be one of NFC, NFD, NFKC, NFKD.");
        return string_fail("normalize form is invalid");
    }
    for (const unsigned char *s = (const unsigned char *)value; *s; s++) {
        if (*s >= 0x80) {
            scriptgo_throw_error_message("String.prototype.normalize of non-ASCII text is not supported in the native subset");
            return string_fail("normalize of non-ASCII text is not supported");
        }
    }
    *out_value = strdup(value);
    if (*out_value == NULL) return string_fail("scriptgo string allocation failed");
    return 0;
}

/* scriptgo_string_code_points splits value into its code points, the
 * sequence `for..of` and spreading a string produce. */
int scriptgo_string_code_points(const char *value, void **out_array) {
    if (value == NULL || out_array == NULL) return string_fail("scriptgo string argument is invalid");
    size_t count = 0;
    for (const unsigned char *p = (const unsigned char *)value; *p != 0; count++) {
        size_t units;
        p += utf8_step(p, &units);
    }
    if (scriptgo_array_new((int64_t)count, sizeof(char *), out_array) != 0) return -1;
    size_t bytes = strlen(value);
    char *buffer = malloc(bytes + count + 1);
    if (buffer == NULL || scriptgo_array_set_owned_data(*out_array, buffer) != 0) {
        free(buffer);
        scriptgo_array_release(*out_array);
        return string_fail("scriptgo string allocation failed");
    }
    const unsigned char *p = (const unsigned char *)value;
    size_t offset = 0;
    for (size_t i = 0; i < count; i++) {
        size_t units;
        size_t step = utf8_step(p, &units);
        char *item = buffer + offset;
        memcpy(item, p, step);
        item[step] = '\0';
        offset += step + 1;
        p += step;
        if (scriptgo_array_set(*out_array, (double)i, &item) != 0) return -1;
    }
    return 0;
}

/* String.prototype.localeCompare with the root-locale ordering of
 * collation_compare. */
int scriptgo_string_locale_compare(const char *value, const char *other, double *out_result) {
    if (out_result == NULL) return string_fail("scriptgo string argument is invalid");
    if (value == NULL) value = "null";
    if (other == NULL) other = "null";
    if (other == &scriptgo_undefined_sentinel) other = "undefined";
    *out_result = (double)collation_compare(value, other);
    return 0;
}
