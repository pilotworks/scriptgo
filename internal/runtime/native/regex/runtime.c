#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ECMAScript regular expressions on QuickJS-ng's libregexp, compiled as its
 * own translation unit with a scriptgo_re_ prefix (see
 * internal/runtime/regexp_engine.go). Subjects are matched as UTF-16 code
 * units, so every index (match index, lastIndex, search) is a JavaScript
 * string index. This file holds compilation, matching and match arrays;
 * methods.c holds the String.prototype methods built on them. */

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_new_tagged(int64_t length, int64_t element_size, int64_t element_tag, void **out_array);
int scriptgo_array_set(void *array, double index, const void *element);
int scriptgo_array_get(void *handle, double index, void *out_value);
int scriptgo_array_set_properties(void *handle, void *properties);
int scriptgo_array_properties(void *handle, void **out_properties);
int scriptgo_object_new_typed(int64_t field_count, const char *type_name, void **out_object);
int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);

extern const char scriptgo_undefined_sentinel;

/* libregexp's interface (libregexp.h). */
#define REGEX_FLAG_GLOBAL (1 << 0)
#define REGEX_FLAG_IGNORECASE (1 << 1)
#define REGEX_FLAG_MULTILINE (1 << 2)
#define REGEX_FLAG_DOTALL (1 << 3)
#define REGEX_FLAG_UNICODE (1 << 4)
#define REGEX_FLAG_STICKY (1 << 5)
#define REGEX_FLAG_INDICES (1 << 6)
#define REGEX_FLAG_NAMED_GROUPS (1 << 7)
#define REGEX_FLAG_UNICODE_SETS (1 << 8)

uint8_t *scriptgo_re_lre_compile(int *plen, char *error_msg, int error_msg_size, const char *buf,
                                 size_t buf_len, int re_flags, void *opaque);
int scriptgo_re_lre_get_capture_count(const uint8_t *bc_buf);
int scriptgo_re_lre_get_flags(const uint8_t *bc_buf);
const char *scriptgo_re_lre_get_groupnames(const uint8_t *bc_buf);
int scriptgo_re_lre_exec(uint8_t **capture, const uint8_t *bc_buf, const uint8_t *cbuf, int cindex, int clen,
                         int cbuf_type, void *opaque);

/* Matching recurses on the C stack; the engine fails a match rather than
 * use more than this much stack below its caller. */
#define REGEX_STACK_BUDGET (256 * 1024)

void scriptgo_throw_error_message(const char *text);

static int regex_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

/* regex_throw throws a JavaScript error ("SyntaxError: ..." and so on) that
 * a try statement can catch. */
static int regex_throw(const char *message) {
    scriptgo_throw_error_message(message);
    return -1;
}

/* regex_parse_flags maps a flags string to libregexp flags; a repeated or
 * unknown flag, or u with v, is a SyntaxError. */
static int regex_parse_flags(const char *flags, int *out) {
    int result = 0;
    for (const char *p = flags != NULL ? flags : ""; *p != 0; p++) {
        int bit;
        switch (*p) {
        case 'd': bit = REGEX_FLAG_INDICES; break;
        case 'g': bit = REGEX_FLAG_GLOBAL; break;
        case 'i': bit = REGEX_FLAG_IGNORECASE; break;
        case 'm': bit = REGEX_FLAG_MULTILINE; break;
        case 's': bit = REGEX_FLAG_DOTALL; break;
        case 'u': bit = REGEX_FLAG_UNICODE; break;
        case 'v': bit = REGEX_FLAG_UNICODE_SETS; break;
        case 'y': bit = REGEX_FLAG_STICKY; break;
        default: bit = 0; break;
        }
        if (bit == 0 || (result & bit) != 0) goto invalid;
        result |= bit;
    }
    if ((result & REGEX_FLAG_UNICODE) && (result & REGEX_FLAG_UNICODE_SETS)) goto invalid;
    *out = result;
    return 0;
invalid: {
        char message[160];
        snprintf(message, sizeof(message), "SyntaxError: Invalid flags supplied to RegExp constructor '%.64s'", flags);
        return regex_throw(message);
    }
}

/* A program's bytecode belongs to the compile cache and stays valid until
 * the next compile; code that runs user code while matching (a replacer
 * function) takes its own copy with regex_program_own. */
typedef struct {
    const uint8_t *bytecode;
    size_t length;
    int flags;
    int capture_count;
} regex_program;

typedef struct {
    char *pattern;
    char *flags;
    uint8_t *bytecode;
    size_t length;
} regex_cache_entry;

/* Compiled patterns are cached, so a loop over exec or test compiles once. */
#define REGEX_CACHE_SIZE 64
static regex_cache_entry regex_cache[REGEX_CACHE_SIZE];
static unsigned regex_cache_next;

static uintptr_t regex_stack_limit(void) {
    char probe;
    uintptr_t here = (uintptr_t)&probe;
    return here > REGEX_STACK_BUDGET ? here - REGEX_STACK_BUDGET : 0;
}

/* regex_compile returns the program for pattern and flags, compiling and
 * caching it on first use. */
static int regex_compile(const char *pattern, const char *flags, regex_program *out) {
    if (pattern == NULL) pattern = "";
    if (flags == NULL) flags = "";
    for (unsigned i = 0; i < REGEX_CACHE_SIZE; i++) {
        regex_cache_entry *entry = &regex_cache[i];
        if (entry->bytecode != NULL && strcmp(entry->pattern, pattern) == 0 && strcmp(entry->flags, flags) == 0) {
            out->bytecode = entry->bytecode;
            out->length = entry->length;
            out->flags = scriptgo_re_lre_get_flags(entry->bytecode);
            out->capture_count = scriptgo_re_lre_get_capture_count(entry->bytecode);
            return 0;
        }
    }
    int re_flags;
    if (regex_parse_flags(flags, &re_flags) != 0) return -1;
    char error[64];
    int length = 0;
    uintptr_t limit = regex_stack_limit();
    uint8_t *bytecode = scriptgo_re_lre_compile(&length, error, sizeof(error), pattern, strlen(pattern), re_flags, &limit);
    if (bytecode == NULL) {
        char message[256];
        snprintf(message, sizeof(message), "SyntaxError: Invalid regular expression: /%.96s/%.16s: %s", pattern, flags, error);
        return regex_throw(message);
    }
    char *pattern_copy = strdup(pattern);
    char *flags_copy = strdup(flags);
    if (pattern_copy == NULL || flags_copy == NULL) {
        free(pattern_copy);
        free(flags_copy);
        free(bytecode);
        return regex_fail("regex compile: out of memory");
    }
    regex_cache_entry *slot = &regex_cache[regex_cache_next++ % REGEX_CACHE_SIZE];
    free(slot->pattern);
    free(slot->flags);
    free(slot->bytecode);
    slot->pattern = pattern_copy;
    slot->flags = flags_copy;
    slot->bytecode = bytecode;
    slot->length = (size_t)length;
    out->bytecode = bytecode;
    out->length = (size_t)length;
    out->flags = scriptgo_re_lre_get_flags(bytecode);
    out->capture_count = scriptgo_re_lre_get_capture_count(bytecode);
    return 0;
}

/* regex_program_own copies the program's bytecode out of the cache; the
 * caller frees it with free((void *)program->bytecode). */
static int regex_program_own(regex_program *program) {
    uint8_t *copy = malloc(program->length);
    if (copy == NULL) return regex_fail("regex: out of memory");
    memcpy(copy, program->bytecode, program->length);
    program->bytecode = copy;
    return 0;
}

/* A subject is the UTF-8 string and its UTF-16 code units. */
typedef struct {
    const char *text;
    uint16_t *units;
    size_t length;
} regex_subject;

static int regex_subject_init(regex_subject *subject, const char *text) {
    if (text == NULL) text = "";
    size_t bytes = strlen(text);
    subject->text = text;
    subject->length = 0;
    subject->units = malloc((bytes + 1) * sizeof(uint16_t));
    if (subject->units == NULL) return regex_fail("regex: out of memory");
    const unsigned char *p = (const unsigned char *)text;
    while (*p != 0) {
        size_t step;
        uint32_t cp = utf8_decode(p, &step);
        if (cp >= 0x10000) {
            subject->units[subject->length++] = (uint16_t)(0xD800 + ((cp - 0x10000) >> 10));
            subject->units[subject->length++] = (uint16_t)(0xDC00 + ((cp - 0x10000) & 0x3FF));
        } else {
            subject->units[subject->length++] = (uint16_t)cp;
        }
        p += step;
    }
    return 0;
}

static void regex_subject_free(regex_subject *subject) {
    free(subject->units);
    subject->units = NULL;
}

/* regex_units_text writes code units as UTF-8, pairing surrogates; a lone
 * surrogate is written in its 3-byte form. */
static char *regex_units_text(const uint16_t *units, size_t count) {
    char *out = malloc(count * 3 + 1);
    size_t written = 0;
    if (out == NULL) return NULL;
    for (size_t i = 0; i < count; i++) {
        uint32_t cp = units[i];
        if (cp >= 0xD800 && cp <= 0xDBFF && i + 1 < count && units[i + 1] >= 0xDC00 && units[i + 1] <= 0xDFFF) {
            cp = 0x10000 + ((cp - 0xD800) << 10) + (units[i + 1] - 0xDC00);
            i++;
        }
        written += utf8_encode(cp, out + written);
    }
    out[written] = '\0';
    return out;
}

/* regex_advance is AdvanceStringIndex: past one code unit, or one code
 * point in unicode mode. */
static size_t regex_advance(const regex_program *program, const regex_subject *subject, size_t index) {
    if ((program->flags & (REGEX_FLAG_UNICODE | REGEX_FLAG_UNICODE_SETS)) && index + 1 < subject->length &&
        subject->units[index] >= 0xD800 && subject->units[index] <= 0xDBFF &&
        subject->units[index + 1] >= 0xDC00 && subject->units[index + 1] <= 0xDFFF) {
        return index + 2;
    }
    return index + 1;
}

/* A match holds 2 * capture_count positions into the subject's units; an
 * unmatched group has NULL positions. */
typedef struct {
    uint8_t **capture;
    int count;
} regex_match;

static int regex_match_init(regex_match *match, const regex_program *program) {
    match->count = program->capture_count;
    match->capture = calloc((size_t)program->capture_count * 2, sizeof(uint8_t *));
    return match->capture == NULL ? regex_fail("regex: out of memory") : 0;
}

static void regex_match_free(regex_match *match) {
    free(match->capture);
    match->capture = NULL;
}

/* regex_exec_at matches from code unit start: 1 on a match, 0 on none, -1
 * on an engine failure (the error is set). A sticky program matches only
 * at start. */
static int regex_exec_at(const regex_program *program, const regex_subject *subject, size_t start, regex_match *match) {
    if (start > subject->length) return 0;
    /* Type 1 is a 16-bit buffer; the engine reads surrogate pairs as code
     * points itself when the program has the u or v flag. */
    uintptr_t limit = regex_stack_limit();
    int status = scriptgo_re_lre_exec(match->capture, program->bytecode, (const uint8_t *)subject->units, (int)start,
                                      (int)subject->length, 1, &limit);
    if (status < 0) return regex_throw("RangeError: Maximum call stack size exceeded");
    return status;
}

static bool regex_group_matched(const regex_match *match, int group) {
    return match->capture[2 * group] != NULL && match->capture[2 * group + 1] != NULL;
}

static size_t regex_group_start(const regex_match *match, const regex_subject *subject, int group) {
    return (size_t)((const uint16_t *)match->capture[2 * group] - subject->units);
}

static size_t regex_group_end(const regex_match *match, const regex_subject *subject, int group) {
    return (size_t)((const uint16_t *)match->capture[2 * group + 1] - subject->units);
}

/* regex_group_text is the text group captured, or the undefined sentinel
 * when it did not participate. */
static const char *regex_group_text(const regex_match *match, const regex_subject *subject, int group) {
    if (!regex_group_matched(match, group)) return &scriptgo_undefined_sentinel;
    size_t start = regex_group_start(match, subject, group);
    char *text = regex_units_text(subject->units + start, regex_group_end(match, subject, group) - start);
    return text != NULL ? text : &scriptgo_undefined_sentinel;
}

/* regex_group_names lists the name of each capture group 1..count-1 (NULL
 * for an unnamed group) and reports whether any is named. */
static bool regex_group_names(const regex_program *program, const char **names) {
    for (int i = 0; i < program->capture_count; i++) names[i] = NULL;
    if (!(program->flags & REGEX_FLAG_NAMED_GROUPS)) return false;
    const char *cursor = scriptgo_re_lre_get_groupnames(program->bytecode);
    bool any = false;
    for (int group = 1; cursor != NULL && group < program->capture_count; group++) {
        if (*cursor != 0) {
            names[group] = cursor;
            any = true;
        }
        cursor += strlen(cursor) + 1;
    }
    return any;
}

/* regex_groups_object builds the groups object of a match: a null-prototype
 * object (the __null_prototype shape, its keys in the |length:name extension
 * encoding) whose fields hold the captured text or undefined. */
static void *regex_groups_object(const regex_program *program, const regex_match *match, const regex_subject *subject) {
    static const char shape[] = "__null_prototype";
    const char **names = calloc((size_t)program->capture_count, sizeof(char *));
    if (names == NULL) return NULL;
    if (!regex_group_names(program, names)) {
        free(names);
        return (void *)&scriptgo_undefined_sentinel;
    }
    size_t layout_length = sizeof(shape);
    int named = 0;
    for (int group = 1; group < program->capture_count; group++) {
        if (names[group] == NULL) continue;
        layout_length += strlen(names[group]) + 24;
        named++;
    }
    /* Objects keep their layout string, so it is not freed. */
    char *layout = malloc(layout_length);
    void *groups = NULL;
    if (layout != NULL) {
        size_t at = sizeof(shape) - 1;
        memcpy(layout, shape, at);
        for (int group = 1; group < program->capture_count; group++) {
            if (names[group] == NULL) continue;
            at += (size_t)snprintf(layout + at, layout_length - at, "|%zu:%s", strlen(names[group]), names[group]);
        }
        layout[at] = '\0';
        if (scriptgo_object_new_typed(named, layout, &groups) == 0) {
            int field = 0;
            for (int group = 1; group < program->capture_count; group++) {
                if (names[group] == NULL) continue;
                scriptgo_object_string_set(groups, field++, regex_group_text(match, subject, group));
            }
        }
    }
    free(names);
    return groups;
}

/* regex_match_array is the RegExp exec result: [match, ...captures] as a
 * string array whose properties object holds index, input and groups. */
static int regex_match_array(const regex_program *program, const regex_match *match, const regex_subject *subject,
                             void **out_array) {
    if (scriptgo_array_new_tagged(program->capture_count, sizeof(char *), 4, out_array) != 0) return -1;
    for (int group = 0; group < program->capture_count; group++) {
        const char *text = regex_group_text(match, subject, group);
        scriptgo_array_set(*out_array, (double)group, &text);
    }
    void *properties = NULL;
    void *groups = regex_groups_object(program, match, subject);
    if (groups == NULL || scriptgo_object_new_typed(3, ":index:input:groups:", &properties) != 0) {
        return regex_fail("regex match: out of memory");
    }
    scriptgo_object_number_set(properties, 0, (double)regex_group_start(match, subject, 0));
    scriptgo_object_string_set(properties, 1, subject->text);
    scriptgo_object_ptr_set(properties, 2, groups);
    return scriptgo_array_set_properties(*out_array, properties);
}

/* scriptgo_regex_match_properties is the index/input/groups object of a
 * match array; an array without one (a global match's strings) reads all
 * three as undefined. */
int scriptgo_regex_match_properties(void *array, void **out_properties) {
    static void *absent = NULL;
    if (out_properties == NULL) return regex_fail("invalid argument to match properties");
    if (scriptgo_array_properties(array, out_properties) != 0) return -1;
    if (*out_properties != NULL) return 0;
    if (absent == NULL) {
        if (scriptgo_object_new_typed(3, ":index:input:groups:", &absent) != 0) return -1;
        for (int64_t i = 0; i < 3; i++) scriptgo_object_ptr_set(absent, i, (void *)&scriptgo_undefined_sentinel);
    }
    *out_properties = absent;
    return 0;
}

int scriptgo_regex_test(const char *pattern, const char *flags, const char *str, double *out_bool) {
    if (out_bool == NULL) return regex_fail("invalid argument to regex test");
    regex_program program;
    regex_subject subject;
    regex_match match;
    if (regex_compile(pattern, flags, &program) != 0) return -1;
    if (regex_subject_init(&subject, str) != 0) return -1;
    if (regex_match_init(&match, &program) != 0) {
        regex_subject_free(&subject);
        return -1;
    }
    int status = regex_exec_at(&program, &subject, 0, &match);
    regex_match_free(&match);
    regex_subject_free(&subject);
    if (status < 0) return -1;
    *out_bool = status > 0 ? 1.0 : 0.0;
    return 0;
}

/* regex_exec_stateful is RegExp.prototype.exec. A global or sticky pattern
 * starts at *inout_last_index (a code unit index) and stores the end of the
 * match there, or 0 when there is none. With out_array NULL only *out_found
 * is reported (RegExp.prototype.test). */
static int regex_exec_stateful(const char *pattern, const char *flags, const char *str, double *inout_last_index,
                               void **out_array, bool *out_found) {
    regex_program program;
    regex_subject subject;
    regex_match match;
    if (regex_compile(pattern, flags, &program) != 0) return -1;
    bool stateful = (program.flags & (REGEX_FLAG_GLOBAL | REGEX_FLAG_STICKY)) && inout_last_index != NULL;
    size_t start = 0;
    if (stateful && *inout_last_index == *inout_last_index && *inout_last_index > 0.0) {
        start = *inout_last_index > 9007199254740991.0 ? (size_t)-1 : (size_t)*inout_last_index;
    }
    if (regex_subject_init(&subject, str) != 0) return -1;
    if (regex_match_init(&match, &program) != 0) {
        regex_subject_free(&subject);
        return -1;
    }
    int status = regex_exec_at(&program, &subject, start, &match);
    if (status > 0 && out_array != NULL) status = regex_match_array(&program, &match, &subject, out_array) == 0 ? 1 : -1;
    if (stateful && status >= 0) *inout_last_index = status > 0 ? (double)regex_group_end(&match, &subject, 0) : 0.0;
    *out_found = status > 0;
    regex_match_free(&match);
    regex_subject_free(&subject);
    return status < 0 ? -1 : 0;
}

int scriptgo_regex_exec_stateful(const char *pattern, const char *flags, const char *str, double *inout_last_index,
                                 void **out_array) {
    if (out_array == NULL) return regex_fail("invalid argument to regex exec");
    bool found;
    *out_array = NULL;
    return regex_exec_stateful(pattern, flags, str, inout_last_index, out_array, &found);
}

/* scriptgo_regex_test_stateful is RegExp.prototype.test, which advances
 * lastIndex like exec. */
int scriptgo_regex_test_stateful(const char *pattern, const char *flags, const char *str, double *inout_last_index,
                                 double *out_bool) {
    if (out_bool == NULL) return regex_fail("invalid argument to regex test");
    bool found = false;
    if (regex_exec_stateful(pattern, flags, str, inout_last_index, NULL, &found) != 0) return -1;
    *out_bool = found ? 1.0 : 0.0;
    return 0;
}

/* scriptgo_regex_validate compiles a pattern for the RegExp constructor, so
 * an invalid pattern or flags is a SyntaxError where it is created. */
int scriptgo_regex_validate(const char *pattern, const char *flags) {
    regex_program program;
    return regex_compile(pattern, flags, &program);
}

int scriptgo_regex_exec(const char *pattern, const char *flags, const char *str, void **out_array) {
    return scriptgo_regex_exec_stateful(pattern, flags, str, NULL, out_array);
}
