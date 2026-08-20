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
    if (regcomp(&re, pattern, cflags) != 0) {
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    regmatch_t pmatch[1];
    int status = regexec(&re, str, 1, pmatch, 0);
    regfree(&re);
    *out_bool = (status == 0) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_regex_exec(const char *pattern, const char *flags, const char *str, void **out_array) {
    if (pattern == NULL || str == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("invalid argument to regex exec");
    }
    int cflags = REG_EXTENDED;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    }
    regex_t re;
    if (regcomp(&re, pattern, cflags) != 0) {
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    regmatch_t pmatch[16];
    int status = regexec(&re, str, 16, pmatch, 0);
    if (status != 0) {
        regfree(&re);
        return scriptgo_array_new(0, sizeof(const char*), out_array);
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
            memcpy(sub, str + pmatch[i].rm_so, len);
            sub[len] = '\0';
            scriptgo_array_set(*out_array, (double)i, &sub);
        }
    }
    return 0;
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
    if (regcomp(&re, pattern, cflags) != 0) {
        return scriptgo_runtime_set_error("invalid regular expression");
    }
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
    int err = scriptgo_array_new(count, sizeof(const char*), out_array);
    if (err != 0) return err;
    if (count == 0) return 0;

    regcomp(&re, pattern, cflags);
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
    if (regcomp(&re, pattern, cflags) != 0) {
        return scriptgo_runtime_set_error("invalid regular expression");
    }
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
    if (regcomp(&re, pattern, cflags) != 0) {
        return scriptgo_runtime_set_error("invalid regular expression");
    }
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
