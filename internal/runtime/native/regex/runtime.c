#include <ctype.h>
#include <regex.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set(void *array, double index, const void *element);

static char *normalize_pattern(const char *pattern) {
    if (pattern == NULL) return NULL;
    size_t len = strlen(pattern);
    char *buf = malloc(len * 16 + 1);
    if (buf == NULL) return NULL;
    size_t out = 0;
    for (size_t i = 0; i < len; i++) {
        if (pattern[i] == '\\' && i + 1 < len) {
            char next = pattern[i + 1];
            if (next == 'd') {
                const char *rep = "[0-9]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            } else if (next == 'D') {
                const char *rep = "[^0-9]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            } else if (next == 'w') {
                const char *rep = "[a-zA-Z0-9_]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            } else if (next == 'W') {
                const char *rep = "[^a-zA-Z0-9_]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            } else if (next == 's') {
                const char *rep = "[ \t\r\n\f\v]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            } else if (next == 'S') {
                const char *rep = "[^ \t\r\n\f\v]";
                size_t rlen = strlen(rep);
                memcpy(buf + out, rep, rlen);
                out += rlen;
                i++;
                continue;
            }
        }
        buf[out++] = pattern[i];
    }
    buf[out] = '\0';
    return buf;
}

int scriptgo_regex_test(const char *pattern, const char *flags, const char *str, double *out_bool) {
    if (pattern == NULL || str == NULL || out_bool == NULL) {
        return scriptgo_runtime_set_error("invalid argument to regex test");
    }
    int cflags = REG_EXTENDED;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    }
    regex_t re;
    char *norm = normalize_pattern(pattern);
    if (regcomp(&re, norm ? norm : pattern, cflags) != 0) {
        if (norm) free(norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    if (norm) free(norm);
    regmatch_t pmatch[1];
    int status = regexec(&re, str, 1, pmatch, 0);
    regfree(&re);
    *out_bool = (status == 0) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_regex_exec_stateful(const char *pattern, const char *flags, const char *str, double *inout_last_index, void **out_array) {
    if (pattern == NULL || str == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("invalid argument to regex exec");
    }
    int is_global = (flags != NULL && (strchr(flags, 'g') != NULL || strchr(flags, 'y') != NULL));
    size_t last_idx = 0;
    if (is_global && inout_last_index != NULL && !isnan(*inout_last_index) && *inout_last_index > 0.0) {
        last_idx = (size_t)*inout_last_index;
    }
    size_t str_len = strlen(str);
    if (last_idx > str_len) {
        if (is_global && inout_last_index != NULL) *inout_last_index = 0.0;
        *out_array = NULL;
        return 0;
    }
    int cflags = REG_EXTENDED;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    }
    regex_t re;
    char *norm = normalize_pattern(pattern);
    if (regcomp(&re, norm ? norm : pattern, cflags) != 0) {
        if (norm) free(norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    if (norm) free(norm);
    regmatch_t pmatch[16];
    int status = regexec(&re, str + last_idx, 16, pmatch, 0);
    if (status != 0) {
        regfree(&re);
        if (is_global && inout_last_index != NULL) *inout_last_index = 0.0;
        *out_array = NULL;
        return 0;
    }
    if (is_global && inout_last_index != NULL) {
        *inout_last_index = (double)(last_idx + pmatch[0].rm_eo);
    }
    int count = 0;
    for (int i = 0; i < 16; i++) {
        if (pmatch[i].rm_so != -1) count++;
    }
    regfree(&re);
    int err = scriptgo_array_new(count, sizeof(const char*), out_array);
    if (err != 0) return err;
    for (int i = 0; i < count; i++) {
        int len = pmatch[i].rm_eo - pmatch[i].rm_so;
        char *sub = malloc(len + 1);
        if (sub != NULL) {
            memcpy(sub, str + last_idx + pmatch[i].rm_so, len);
            sub[len] = '\0';
            scriptgo_array_set(*out_array, (double)i, &sub);
        }
    }
    return 0;
}

int scriptgo_regex_exec(const char *pattern, const char *flags, const char *str, void **out_array) {
    return scriptgo_regex_exec_stateful(pattern, flags, str, NULL, out_array);
}

int scriptgo_string_match(const char *str, const char *pattern, const char *flags, void **out_array) {
    if (pattern == NULL || flags == NULL || str == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("invalid argument to match");
    }
    int is_global = (strchr(flags, 'g') != NULL);
    if (!is_global) {
        return scriptgo_regex_exec(pattern, flags, str, out_array);
    }
    int cflags = REG_EXTENDED;
    if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
    if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    regex_t re;
    char *norm = normalize_pattern(pattern);
    if (regcomp(&re, norm ? norm : pattern, cflags) != 0) {
        if (norm) free(norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    if (norm) free(norm);
    const char *cursor = str;
    regmatch_t pmatch[1];
    int count = 0;
    while (*cursor && regexec(&re, cursor, 1, pmatch, 0) == 0) {
        count++;
        int advance = pmatch[0].rm_eo;
        if (advance == 0) advance = 1;
        cursor += advance;
    }
    regfree(&re);
    if (count == 0) {
        *out_array = NULL;
        return 0;
    }
    int err = scriptgo_array_new(count, sizeof(const char*), out_array);
    if (err != 0) return err;

    char *norm2 = normalize_pattern(pattern);
    regcomp(&re, norm2 ? norm2 : pattern, cflags);
    if (norm2) free(norm2);
    cursor = str;
    int idx = 0;
    while (*cursor && regexec(&re, cursor, 1, pmatch, 0) == 0 && idx < count) {
        int len = pmatch[0].rm_eo - pmatch[0].rm_so;
        char *sub = malloc(len + 1);
        if (sub != NULL) {
            memcpy(sub, cursor + pmatch[0].rm_so, len);
            sub[len] = '\0';
            scriptgo_array_set(*out_array, (double)idx, &sub);
        }
        idx++;
        int advance = pmatch[0].rm_eo;
        if (advance == 0) advance = 1;
        cursor += advance;
    }
    regfree(&re);
    return 0;
}

int scriptgo_string_search(const char *str, const char *pattern, const char *flags, double *out_index) {
    if (pattern == NULL || str == NULL || out_index == NULL) {
        return scriptgo_runtime_set_error("invalid argument to search");
    }
    int cflags = REG_EXTENDED;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    }
    regex_t re;
    char *norm = normalize_pattern(pattern);
    if (regcomp(&re, norm ? norm : pattern, cflags) != 0) {
        if (norm) free(norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    if (norm) free(norm);
    regmatch_t pmatch[1];
    int status = regexec(&re, str, 1, pmatch, 0);
    regfree(&re);
    if (status == 0) {
        *out_index = (double)pmatch[0].rm_so;
    } else {
        *out_index = -1.0;
    }
    return 0;
}

int scriptgo_string_replace_regex(const char *str, const char *pattern, const char *flags, const char *repl, char **out_str) {
    if (str == NULL || pattern == NULL || repl == NULL || out_str == NULL) {
        return scriptgo_runtime_set_error("invalid argument to replace");
    }
    int cflags = REG_EXTENDED;
    int is_global = 0;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
        if (strchr(flags, 'g') != NULL) is_global = 1;
    }
    regex_t re;
    char *norm = normalize_pattern(pattern);
    if (regcomp(&re, norm ? norm : pattern, cflags) != 0) {
        if (norm) free(norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    if (norm) free(norm);

    regmatch_t pmatch[1];
    if (regexec(&re, str, 1, pmatch, 0) != 0) {
        regfree(&re);
        *out_str = strdup(str);
        return 0;
    }
    if (!is_global) {
        regfree(&re);
        size_t prefix_len = pmatch[0].rm_so;
        size_t repl_len = strlen(repl);
        size_t suffix_len = strlen(str + pmatch[0].rm_eo);
        char *res = malloc(prefix_len + repl_len + suffix_len + 1);
        if (res == NULL) return scriptgo_runtime_set_error("replace allocation failed");
        memcpy(res, str, prefix_len);
        memcpy(res + prefix_len, repl, repl_len);
        memcpy(res + prefix_len + repl_len, str + pmatch[0].rm_eo, suffix_len + 1);
        *out_str = res;
        return 0;
    }
    size_t cap = strlen(str) * 2 + strlen(repl) * 4 + 64;
    char *buf = malloc(cap);
    if (buf == NULL) {
        regfree(&re);
        return scriptgo_runtime_set_error("replace allocation failed");
    }
    size_t len = 0;
    const char *cursor = str;
    size_t repl_len = strlen(repl);
    while (*cursor && regexec(&re, cursor, 1, pmatch, 0) == 0) {
        size_t pfx = pmatch[0].rm_so;
        while (len + pfx + repl_len + 1 >= cap) {
            cap *= 2;
            buf = realloc(buf, cap);
        }
        memcpy(buf + len, cursor, pfx);
        len += pfx;
        memcpy(buf + len, repl, repl_len);
        len += repl_len;
        int adv = pmatch[0].rm_eo;
        if (adv == 0) adv = 1;
        cursor += adv;
    }
    size_t rem = strlen(cursor);
    while (len + rem + 1 >= cap) {
        cap *= 2;
        buf = realloc(buf, cap);
    }
    memcpy(buf + len, cursor, rem);
    len += rem;
    buf[len] = '\0';
    regfree(&re);
    *out_str = buf;
    return 0;
}

int scriptgo_string_from_bigint(long long value, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("invalid argument to fromBigInt");
    char buf[64];
    snprintf(buf, sizeof(buf), "%lld", value);
    *out_str = strdup(buf);
    if (*out_str == NULL) return scriptgo_runtime_set_error("fromBigInt allocation failed");
    return 0;
}

int scriptgo_string_from_bigint_locale(long long value, char **out_str) {
    char raw[64];
    size_t digits;
    size_t groups;
    size_t output_len;
    size_t src = 0;
    size_t dst = 0;
    char *result;

    if (out_str == NULL) return scriptgo_runtime_set_error("invalid argument to bigint locale formatting");
    snprintf(raw, sizeof(raw), "%lld", value);
    digits = raw[0] == '-' ? strlen(raw) - 1 : strlen(raw);
    groups = digits > 3 ? (digits - 1) / 3 : 0;
    output_len = strlen(raw) + groups;
    result = malloc(output_len + 1);
    if (result == NULL) return scriptgo_runtime_set_error("bigint locale allocation failed");
    if (raw[0] == '-') result[dst++] = raw[src++];
    for (size_t i = 0; i < digits; i++) {
        if (i > 0 && (digits - i) % 3 == 0) result[dst++] = ',';
        result[dst++] = raw[src++];
    }
    result[dst] = '\0';
    *out_str = result;
    return 0;
}

int scriptgo_bigint_from_number(double value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint fromNumber");
    *out_value = (long long)value;
    return 0;
}

int scriptgo_bigint_from_string(const char *str, long long *out_value) {
    if (str == NULL || out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint fromString");
    char *endptr = NULL;
    *out_value = strtoll(str, &endptr, 10);
    return 0;
}

int scriptgo_bigint_as_int_n(long long bits, long long value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint asIntN");
    if (bits <= 0) {
        *out_value = 0;
        return 0;
    }
    if (bits >= 64) {
        *out_value = value;
        return 0;
    }
    unsigned long long uval = (unsigned long long)value;
    unsigned long long mask = (1ULL << bits) - 1ULL;
    uval = uval & mask;
    if (uval & (1ULL << (bits - 1))) {
        uval |= (~mask);
    }
    *out_value = (long long)uval;
    return 0;
}

int scriptgo_bigint_as_uint_n(long long bits, long long value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint asUintN");
    if (bits <= 0) {
        *out_value = 0;
        return 0;
    }
    if (bits >= 64) {
        *out_value = value;
        return 0;
    }
    unsigned long long uval = (unsigned long long)value;
    unsigned long long mask = (1ULL << bits) - 1ULL;
    *out_value = (long long)(uval & mask);
    return 0;
}

long long scriptgo_bigint_pow(long long base, long long exp) {
    if (exp < 0) {
        if (base == 1) return 1;
        if (base == -1) return (exp % 2 == 0) ? 1 : -1;
        return 0;
    }
    if (exp == 0) return 1;
    long long result = 1;
    long long b = base;
    unsigned long long e = (unsigned long long)exp;
    while (e > 0) {
        if (e & 1) {
            result *= b;
        }
        b *= b;
        e >>= 1;
    }
    return result;
}
