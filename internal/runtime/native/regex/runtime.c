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

extern const char scriptgo_undefined_sentinel;

#define MAX_REGEX_GROUPS 64

typedef struct {
    char *pattern;
    int posix_group_count;
    int js_group_count;
    int is_capturing[MAX_REGEX_GROUPS];
    /* names[i] is the name of JS capture group i ((?<name>...)), or NULL. */
    char *names[MAX_REGEX_GROUPS];
    int named_count;
} normalized_regex_t;

static void normalized_regex_free_names(normalized_regex_t *norm) {
    for (int g = 0; g < MAX_REGEX_GROUPS; g++) {
        free(norm->names[g]);
        norm->names[g] = NULL;
    }
}

static int normalize_pattern_full(const char *pattern, normalized_regex_t *out) {
    if (pattern == NULL || out == NULL) return -1;
    size_t len = strlen(pattern);
    char *buf = malloc(len * 16 + 1);
    if (buf == NULL) return -1;

    out->posix_group_count = 0;
    out->js_group_count = 0;
    out->named_count = 0;
    for (int g = 0; g < MAX_REGEX_GROUPS; g++) {
        out->is_capturing[g] = 0;
        out->names[g] = NULL;
    }

    size_t out_idx = 0;
    size_t i = 0;
    int in_bracket = 0;

    while (i < len) {
        if (pattern[i] == '\\' && i + 1 < len) {
            char next = pattern[i + 1];
            if (!in_bracket) {
                if (next == 'd') {
                    const char *rep = "[0-9]";
                    size_t rlen = 5;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                } else if (next == 'D') {
                    const char *rep = "[^0-9]";
                    size_t rlen = 6;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                } else if (next == 'w') {
                    const char *rep = "[a-zA-Z0-9_]";
                    size_t rlen = 12;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                } else if (next == 'W') {
                    const char *rep = "[^a-zA-Z0-9_]";
                    size_t rlen = 13;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                } else if (next == 's') {
                    const char *rep = "[ \t\r\n\f\v]";
                    size_t rlen = 8;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                } else if (next == 'S') {
                    const char *rep = "[^ \t\r\n\f\v]";
                    size_t rlen = 9;
                    memcpy(buf + out_idx, rep, rlen);
                    out_idx += rlen;
                    i += 2;
                    continue;
                }
            }
            buf[out_idx++] = '\\';
            buf[out_idx++] = next;
            i += 2;
            continue;
        }

        if (pattern[i] == '[') {
            in_bracket = 1;
            buf[out_idx++] = pattern[i++];
            if (i < len && pattern[i] == '^') {
                buf[out_idx++] = pattern[i++];
            }
            if (i < len && pattern[i] == ']') {
                buf[out_idx++] = pattern[i++];
            }
            continue;
        }

        if (pattern[i] == ']' && in_bracket) {
            in_bracket = 0;
            buf[out_idx++] = pattern[i++];
            continue;
        }

        if (!in_bracket && pattern[i] == '(') {
            if (i + 2 < len && pattern[i + 1] == '?' && pattern[i + 2] == ':') {
                out->posix_group_count++;
                if (out->posix_group_count < MAX_REGEX_GROUPS) {
                    out->is_capturing[out->posix_group_count] = 0;
                }
                buf[out_idx++] = '(';
                i += 3;
                continue;
            } else {
                out->posix_group_count++;
                if (out->posix_group_count < MAX_REGEX_GROUPS) {
                    out->is_capturing[out->posix_group_count] = 1;
                }
                out->js_group_count++;
                buf[out_idx++] = '(';
                i++;
                /* A named group (?<name>...) captures like (...). */
                if (i + 1 < len && pattern[i] == '?' && pattern[i + 1] == '<' &&
                    i + 2 < len && pattern[i + 2] != '=' && pattern[i + 2] != '!') {
                    const char *name = pattern + i + 2;
                    const char *close = strchr(name, '>');
                    if (close != NULL) {
                        if (out->js_group_count < MAX_REGEX_GROUPS) {
                            out->names[out->js_group_count] = strndup(name, (size_t)(close - name));
                            out->named_count++;
                        }
                        i = (size_t)(close - pattern) + 1;
                    }
                }
                continue;
            }
        }

        buf[out_idx++] = pattern[i++];
    }

    buf[out_idx] = '\0';
    out->pattern = buf;
    return 0;
}

static char *normalize_pattern(const char *pattern) {
    normalized_regex_t norm;
    if (normalize_pattern_full(pattern, &norm) != 0) return NULL;
    normalized_regex_free_names(&norm);
    return norm.pattern;
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

int scriptgo_object_new_typed(int64_t field_count, const char *type_name, void **out_object);
int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);
int scriptgo_array_set_properties(void *handle, void *properties);
int scriptgo_array_get(void *handle, double index, void *out_value);
int scriptgo_array_new_tagged(int64_t length, int64_t element_size, int64_t element_tag, void **out_array);

/* regex_match_groups builds the groups object of a match with named
 * captures: its key list names the groups and each field holds the captured
 * text or undefined. */
static void *regex_match_groups(const normalized_regex_t *norm, char *const *captures) {
    if (norm->named_count == 0) return (void *)&scriptgo_undefined_sentinel;
    size_t layout_length = 2;
    for (int g = 1; g <= norm->js_group_count && g < MAX_REGEX_GROUPS; g++) {
        if (norm->names[g] != NULL) layout_length += strlen(norm->names[g]) + 1;
    }
    /* The layout string lives as long as the object (objects keep it). */
    char *layout = malloc(layout_length);
    if (layout == NULL) return NULL;
    size_t at = 0;
    layout[at++] = ':';
    for (int g = 1; g <= norm->js_group_count && g < MAX_REGEX_GROUPS; g++) {
        if (norm->names[g] == NULL) continue;
        size_t n = strlen(norm->names[g]);
        memcpy(layout + at, norm->names[g], n);
        at += n;
        layout[at++] = ':';
    }
    layout[at] = '\0';
    void *groups = NULL;
    if (scriptgo_object_new_typed(norm->named_count, layout, &groups) != 0) return NULL;
    int field = 0;
    for (int g = 1; g <= norm->js_group_count && g < MAX_REGEX_GROUPS; g++) {
        if (norm->names[g] == NULL) continue;
        scriptgo_object_string_set(groups, field++, captures[g] != NULL ? captures[g] : &scriptgo_undefined_sentinel);
    }
    return groups;
}

/* regex_exec_at matches pattern against str from byte offset start. On a
 * match it returns the RegExp match array ([match, ...captures] with index,
 * input and groups properties) and the byte offsets of the match; otherwise
 * *out_array is NULL. */
static int regex_exec_at(const char *pattern, const char *flags, const char *str, size_t start,
                         size_t *out_begin, size_t *out_end, void **out_array) {
    *out_array = NULL;
    int cflags = REG_EXTENDED;
    if (flags != NULL) {
        if (strchr(flags, 'i') != NULL) cflags |= REG_ICASE;
        if (strchr(flags, 'm') != NULL) cflags |= REG_NEWLINE;
    }
    regex_t re;
    normalized_regex_t norm;
    if (normalize_pattern_full(pattern, &norm) != 0) {
        return scriptgo_runtime_set_error("regex normalization failed");
    }
    if (regcomp(&re, norm.pattern, cflags) != 0) {
        free(norm.pattern);
        normalized_regex_free_names(&norm);
        return scriptgo_runtime_set_error("invalid regular expression");
    }
    free(norm.pattern);

    size_t nmatch = re.re_nsub + 1;
    regmatch_t *pmatch = malloc(nmatch * sizeof(regmatch_t));
    char **captures = calloc((size_t)norm.js_group_count + 1, sizeof(char *));
    if (pmatch == NULL || captures == NULL) {
        free(pmatch);
        free(captures);
        regfree(&re);
        normalized_regex_free_names(&norm);
        return scriptgo_runtime_set_error("regex match allocation failed");
    }
    int status = regexec(&re, str + start, nmatch, pmatch, start > 0 ? REG_NOTBOL : 0);
    regfree(&re);
    int err = 0;
    if (status != 0) goto done;
    *out_begin = start + (size_t)pmatch[0].rm_so;
    *out_end = start + (size_t)pmatch[0].rm_eo;

    captures[0] = strndup(str + *out_begin, (size_t)(pmatch[0].rm_eo - pmatch[0].rm_so));
    int js_idx = 1;
    for (int p = 1; p <= norm.posix_group_count && (size_t)p < nmatch && js_idx <= norm.js_group_count; p++) {
        if (p < MAX_REGEX_GROUPS && !norm.is_capturing[p]) continue;
        if (pmatch[p].rm_so != -1) {
            captures[js_idx] = strndup(str + start + pmatch[p].rm_so, (size_t)(pmatch[p].rm_eo - pmatch[p].rm_so));
        }
        js_idx++;
    }

    err = scriptgo_array_new_tagged(1 + norm.js_group_count, sizeof(const char *), SCRIPTGO_TAG_STRING, out_array);
    if (err != 0) goto done;
    for (int g = 0; g <= norm.js_group_count; g++) {
        const char *value = captures[g] != NULL ? captures[g] : &scriptgo_undefined_sentinel;
        scriptgo_array_set(*out_array, (double)g, &value);
    }
    void *properties = NULL;
    void *groups = regex_match_groups(&norm, captures);
    if (groups == NULL || scriptgo_object_new_typed(3, ":index:input:groups:", &properties) != 0) {
        err = scriptgo_runtime_set_error("regex match allocation failed");
        goto done;
    }
    scriptgo_object_number_set(properties, 0, (double)utf16_unit_offset(str, *out_begin));
    scriptgo_object_string_set(properties, 1, str);
    scriptgo_object_ptr_set(properties, 2, groups);
    scriptgo_array_set_properties(*out_array, properties);

done:
    /* The array and groups object keep the capture strings. */
    free(captures);
    free(pmatch);
    normalized_regex_free_names(&norm);
    return err;
}

int scriptgo_array_properties(void *handle, void **out_properties);

/* scriptgo_regex_match_properties is the index/input/groups object of a
 * match array; an array without one (a global match's strings) reads all
 * three as undefined. */
int scriptgo_regex_match_properties(void *array, void **out_properties) {
    static void *absent = NULL;
    if (out_properties == NULL) return scriptgo_runtime_set_error("invalid argument to match properties");
    if (scriptgo_array_properties(array, out_properties) != 0) return -1;
    if (*out_properties != NULL) return 0;
    if (absent == NULL) {
        if (scriptgo_object_new_typed(3, ":index:input:groups:", &absent) != 0) return -1;
        for (int64_t i = 0; i < 3; i++) scriptgo_object_ptr_set(absent, i, (void *)&scriptgo_undefined_sentinel);
    }
    *out_properties = absent;
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
    if (last_idx > strlen(str)) {
        if (is_global && inout_last_index != NULL) *inout_last_index = 0.0;
        *out_array = NULL;
        return 0;
    }
    size_t begin = 0, end = 0;
    if (regex_exec_at(pattern, flags, str, last_idx, &begin, &end, out_array) != 0) return -1;
    if (is_global && inout_last_index != NULL) {
        *inout_last_index = *out_array != NULL ? (double)end : 0.0;
    }
    return 0;
}

int scriptgo_regex_exec(const char *pattern, const char *flags, const char *str, void **out_array) {
    return scriptgo_regex_exec_stateful(pattern, flags, str, NULL, out_array);
}

/* regex_next_start is where a global search resumes after a match ending at
 * end: past an empty match by one UTF-8 character (past the end of str for
 * an empty match at its end, which ends the search). */
static size_t regex_next_start(const char *str, size_t begin, size_t end) {
    size_t units;
    if (end > begin) return end;
    if (str[end] == '\0') return end + 1;
    return end + utf8_step((const unsigned char *)str + end, &units);
}

/* scriptgo_string_match_all returns every match of a global pattern as an
 * array of match arrays (str.matchAll). */
int scriptgo_string_match_all(const char *str, const char *pattern, const char *flags, void **out_array) {
    if (pattern == NULL || str == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("invalid argument to matchAll");
    }
    if (flags == NULL || strchr(flags, 'g') == NULL) {
        return scriptgo_runtime_set_error("TypeError: String.prototype.matchAll called with a non-global RegExp argument");
    }
    void **matches = NULL;
    int64_t count = 0, capacity = 0;
    size_t start = 0, length = strlen(str);
    while (start <= length) {
        size_t begin = 0, end = 0;
        void *match = NULL;
        if (regex_exec_at(pattern, flags, str, start, &begin, &end, &match) != 0) {
            free(matches);
            return -1;
        }
        if (match == NULL) break;
        if (count == capacity) {
            capacity = capacity == 0 ? 8 : capacity * 2;
            void **grown = realloc(matches, (size_t)capacity * sizeof(void *));
            if (grown == NULL) {
                free(matches);
                return scriptgo_runtime_set_error("regex match allocation failed");
            }
            matches = grown;
        }
        matches[count++] = match;
        start = regex_next_start(str, begin, end);
    }
    int err = scriptgo_array_new_tagged(count, sizeof(void *), SCRIPTGO_TAG_ARRAY, out_array);
    for (int64_t i = 0; err == 0 && i < count; i++) {
        scriptgo_array_set(*out_array, (double)i, &matches[i]);
    }
    free(matches);
    return err;
}

/* scriptgo_string_match is str.match: the first match with its properties,
 * or every matched string for a global pattern. */
int scriptgo_string_match(const char *str, const char *pattern, const char *flags, void **out_array) {
    if (pattern == NULL || flags == NULL || str == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("invalid argument to match");
    }
    if (strchr(flags, 'g') == NULL) {
        return scriptgo_regex_exec(pattern, flags, str, out_array);
    }
    void *all = NULL;
    if (scriptgo_string_match_all(str, pattern, flags, &all) != 0) return -1;
    int64_t count = ((int64_t *)all)[0];
    *out_array = NULL;
    if (count == 0) return 0;
    if (scriptgo_array_new_tagged(count, sizeof(const char *), SCRIPTGO_TAG_STRING, out_array) != 0) return -1;
    for (int64_t i = 0; i < count; i++) {
        void *match = NULL;
        const char *text = NULL;
        scriptgo_array_get(all, (double)i, &match);
        scriptgo_array_get(match, 0.0, &text);
        scriptgo_array_set(*out_array, (double)i, &text);
    }
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
        *out_index = (double)utf16_unit_offset(str, (size_t)pmatch[0].rm_so);
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
