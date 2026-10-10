/* String.prototype methods that take a regular expression: match,
 * matchAll, search, replace and split. They follow the RegExp.prototype
 * [Symbol.match] / [Symbol.matchAll] / [Symbol.search] / [Symbol.replace] /
 * [Symbol.split] algorithms of ECMA-262 over the matcher in runtime.c. */

int scriptgo_closure_invoke_value(void *closure_handle, int32_t arg_count, const scriptgo_value *a1,
                                  const scriptgo_value *a2, const scriptgo_value *a3, const scriptgo_value *a4,
                                  scriptgo_value *out_value);
int scriptgo_string_from_unknown(const scriptgo_value *value, char **out_str);

/* A growable list of pointers (match arrays or strings). */
typedef struct {
    void **items;
    int64_t count;
    int64_t capacity;
} regex_list;

static int regex_list_push(regex_list *list, void *item) {
    if (list->count == list->capacity) {
        int64_t capacity = list->capacity == 0 ? 8 : list->capacity * 2;
        void **grown = realloc(list->items, (size_t)capacity * sizeof(void *));
        if (grown == NULL) return regex_fail("regex: out of memory");
        list->items = grown;
        list->capacity = capacity;
    }
    list->items[list->count++] = item;
    return 0;
}

/* regex_list_array moves the list into a new array of 8-byte elements
 * tagged element_tag. */
static int regex_list_array(regex_list *list, int64_t element_tag, void **out_array) {
    int status = scriptgo_array_new_tagged(list->count, sizeof(void *), element_tag, out_array);
    for (int64_t i = 0; status == 0 && i < list->count; i++) scriptgo_array_set(*out_array, (double)i, &list->items[i]);
    free(list->items);
    list->items = NULL;
    list->count = list->capacity = 0;
    return status;
}

/* A growable UTF-16 buffer for building replace results. */
typedef struct {
    uint16_t *units;
    size_t length;
    size_t capacity;
} regex_builder;

static int regex_builder_append(regex_builder *builder, const uint16_t *units, size_t count) {
    if (builder->length + count > builder->capacity) {
        size_t capacity = builder->capacity == 0 ? 64 : builder->capacity;
        while (capacity < builder->length + count) capacity *= 2;
        uint16_t *grown = realloc(builder->units, capacity * sizeof(uint16_t));
        if (grown == NULL) return regex_fail("regex: out of memory");
        builder->units = grown;
        builder->capacity = capacity;
    }
    if (count > 0) memcpy(builder->units + builder->length, units, count * sizeof(uint16_t));
    builder->length += count;
    return 0;
}

static int regex_builder_append_text(regex_builder *builder, const char *text) {
    regex_subject converted;
    if (regex_subject_init(&converted, text) != 0) return -1;
    int status = regex_builder_append(builder, converted.units, converted.length);
    regex_subject_free(&converted);
    return status;
}

/* regex_scan calls visit for each successive match of a global pattern (or
 * the first match otherwise) from code unit 0, stepping past empty matches.
 * visit returns 0 to continue, 1 to stop and -1 on error. */
typedef int (*regex_visit_fn)(const regex_program *program, const regex_match *match, const regex_subject *subject,
                              void *context);

static int regex_scan(const regex_program *program, const regex_subject *subject, bool all, regex_visit_fn visit,
                      void *context) {
    regex_match match;
    if (regex_match_init(&match, program) != 0) return -1;
    size_t start = 0;
    int status = 0;
    while (start <= subject->length) {
        int found = regex_exec_at(program, subject, start, &match);
        if (found <= 0) {
            status = found;
            break;
        }
        status = visit(program, &match, subject, context);
        if (status != 0 || !all) break;
        size_t begin = regex_group_start(&match, subject, 0);
        size_t end = regex_group_end(&match, subject, 0);
        start = end > begin ? end : regex_advance(program, subject, end);
    }
    regex_match_free(&match);
    return status < 0 ? -1 : 0;
}

/* regex_open compiles the pattern and converts the subject. */
static int regex_open(const char *str, const char *pattern, const char *flags, regex_program *program,
                      regex_subject *subject) {
    if (regex_compile(pattern, flags, program) != 0) return -1;
    return regex_subject_init(subject, str);
}

static int regex_collect_array(const regex_program *program, const regex_match *match, const regex_subject *subject,
                               void *context) {
    void *array = NULL;
    if (regex_match_array(program, match, subject, &array) != 0) return -1;
    return regex_list_push(context, array);
}

static int regex_collect_text(const regex_program *program, const regex_match *match, const regex_subject *subject,
                              void *context) {
    (void)program;
    return regex_list_push(context, (void *)regex_group_text(match, subject, 0));
}

/* scriptgo_string_match_all returns every match of a global pattern as an
 * array of match arrays (str.matchAll). */
int scriptgo_string_match_all(const char *str, const char *pattern, const char *flags, void **out_array) {
    if (out_array == NULL) return regex_fail("invalid argument to matchAll");
    regex_program program;
    regex_subject subject;
    if (regex_open(str, pattern, flags, &program, &subject) != 0) return -1;
    if (!(program.flags & REGEX_FLAG_GLOBAL)) {
        regex_subject_free(&subject);
        return regex_throw("TypeError: String.prototype.matchAll called with a non-global RegExp argument");
    }
    regex_list matches = {0};
    int status = regex_scan(&program, &subject, true, regex_collect_array, &matches);
    regex_subject_free(&subject);
    if (status != 0) {
        free(matches.items);
        return -1;
    }
    return regex_list_array(&matches, 6, out_array);
}

/* scriptgo_string_match is str.match: the first match with its properties,
 * or every matched string (null when none) for a global pattern. */
int scriptgo_string_match(const char *str, const char *pattern, const char *flags, void **out_array) {
    if (out_array == NULL) return regex_fail("invalid argument to match");
    *out_array = NULL;
    regex_program program;
    if (regex_compile(pattern, flags, &program) != 0) return -1;
    if (!(program.flags & REGEX_FLAG_GLOBAL)) return scriptgo_regex_exec(pattern, flags, str, out_array);
    regex_subject subject;
    if (regex_subject_init(&subject, str) != 0) return -1;
    regex_list texts = {0};
    int status = regex_scan(&program, &subject, true, regex_collect_text, &texts);
    regex_subject_free(&subject);
    if (status != 0 || texts.count == 0) {
        free(texts.items);
        return status;
    }
    return regex_list_array(&texts, 4, out_array);
}

/* scriptgo_string_search is the code unit index of the first match from
 * the start of the string, or -1. */
int scriptgo_string_search(const char *str, const char *pattern, const char *flags, double *out_index) {
    if (out_index == NULL) return regex_fail("invalid argument to search");
    regex_program program;
    regex_subject subject;
    regex_match match;
    if (regex_open(str, pattern, flags, &program, &subject) != 0) return -1;
    if (regex_match_init(&match, &program) != 0) {
        regex_subject_free(&subject);
        return -1;
    }
    int status = regex_exec_at(&program, &subject, 0, &match);
    if (status >= 0) *out_index = status > 0 ? (double)regex_group_start(&match, &subject, 0) : -1.0;
    regex_match_free(&match);
    regex_subject_free(&subject);
    return status < 0 ? -1 : 0;
}

/* Replacement state shared by the string and function forms of replace. */
typedef struct {
    regex_builder out;
    size_t last;           /* end of the previous match */
    const uint16_t *template_units; /* the replacement string, for $ patterns */
    size_t template_length;
    void *replacer;        /* the replacer function, or NULL */
} regex_replace_state;

static size_t regex_template_digits(const uint16_t *units, size_t length, size_t at, int group_count, int *group) {
    /* $n or $nn: two digits when they name a group, else one. */
    if (at >= length || units[at] < '0' || units[at] > '9') return 0;
    int one = units[at] - '0';
    if (at + 1 < length && units[at + 1] >= '0' && units[at + 1] <= '9') {
        int two = one * 10 + (units[at + 1] - '0');
        if (two >= 1 && two < group_count) {
            *group = two;
            return 2;
        }
    }
    if (one >= 1 && one < group_count) {
        *group = one;
        return 1;
    }
    return 0;
}

static int regex_append_group(regex_builder *out, const regex_match *match, const regex_subject *subject, int group) {
    if (!regex_group_matched(match, group)) return 0;
    size_t start = regex_group_start(match, subject, group);
    return regex_builder_append(out, subject->units + start, regex_group_end(match, subject, group) - start);
}

/* regex_substitute appends GetSubstitution for the replacement template:
 * $$, $&, $`, $', $n, $nn and $<name>. */
static int regex_substitute(regex_replace_state *state, const regex_program *program, const regex_match *match,
                            const regex_subject *subject) {
    const uint16_t *t = state->template_units;
    size_t n = state->template_length;
    const char **names = NULL;
    bool named = false;
    for (size_t i = 0; i < n; i++) {
        int status = 0;
        if (t[i] != '$' || i + 1 >= n) {
            status = regex_builder_append(&state->out, &t[i], 1);
        } else if (t[i + 1] == '$') {
            status = regex_builder_append(&state->out, &t[i], 1);
            i++;
        } else if (t[i + 1] == '&') {
            status = regex_append_group(&state->out, match, subject, 0);
            i++;
        } else if (t[i + 1] == '`') {
            status = regex_builder_append(&state->out, subject->units, regex_group_start(match, subject, 0));
            i++;
        } else if (t[i + 1] == '\'') {
            size_t end = regex_group_end(match, subject, 0);
            status = regex_builder_append(&state->out, subject->units + end, subject->length - end);
            i++;
        } else if (t[i + 1] == '<') {
            if (names == NULL) {
                names = calloc((size_t)program->capture_count, sizeof(char *));
                if (names == NULL) return regex_fail("regex: out of memory");
                named = regex_group_names(program, names);
            }
            size_t close = i + 2;
            while (close < n && t[close] != '>') close++;
            if (!named || close >= n) {
                status = regex_builder_append(&state->out, &t[i], 1);
            } else {
                char *name = regex_units_text(t + i + 2, close - i - 2);
                for (int group = 1; name != NULL && group < program->capture_count; group++) {
                    if (names[group] != NULL && strcmp(names[group], name) == 0) {
                        status = regex_append_group(&state->out, match, subject, group);
                        break;
                    }
                }
                free(name);
                i = close;
            }
        } else {
            int group = 0;
            size_t digits = regex_template_digits(t, n, i + 1, program->capture_count, &group);
            if (digits > 0) {
                status = regex_append_group(&state->out, match, subject, group);
                i += digits;
            } else {
                status = regex_builder_append(&state->out, &t[i], 1);
            }
        }
        if (status != 0) {
            free(names);
            return -1;
        }
    }
    free(names);
    return 0;
}

static scriptgo_value regex_value_text(const char *text) {
    scriptgo_value value = {0};
    if (text == &scriptgo_undefined_sentinel) return value;
    value.tag = SCRIPTGO_TAG_STRING;
    value.payload = (uint64_t)(uintptr_t)text;
    value.aux = strlen(text);
    return value;
}

/* regex_call_replacer appends the replacer's result for a match. It is
 * called with (match, ...captures, offset, string) cut to the four
 * arguments a native closure takes. */
static int regex_call_replacer(regex_replace_state *state, const regex_program *program, const regex_match *match,
                               const regex_subject *subject) {
    scriptgo_value args[4];
    int count = 0;
    for (int group = 0; group < program->capture_count && count < 4; group++) {
        args[count++] = regex_value_text(regex_group_text(match, subject, group));
    }
    if (count < 4) {
        scriptgo_value offset = {0};
        double position = (double)regex_group_start(match, subject, 0);
        offset.tag = SCRIPTGO_TAG_NUMBER;
        memcpy(&offset.payload, &position, sizeof(position));
        args[count++] = offset;
    }
    if (count < 4) args[count++] = regex_value_text(subject->text);
    scriptgo_value result = {0};
    if (scriptgo_closure_invoke_value(state->replacer, count, &args[0], count > 1 ? &args[1] : NULL,
                                      count > 2 ? &args[2] : NULL, count > 3 ? &args[3] : NULL, &result) != 0) {
        return -1;
    }
    char *text = NULL;
    if (scriptgo_string_from_unknown(&result, &text) != 0 || text == NULL) return -1;
    return regex_builder_append_text(&state->out, text);
}

static int regex_replace_visit(const regex_program *program, const regex_match *match, const regex_subject *subject,
                               void *context) {
    regex_replace_state *state = context;
    size_t begin = regex_group_start(match, subject, 0);
    if (regex_builder_append(&state->out, subject->units + state->last, begin - state->last) != 0) return -1;
    int status = state->replacer != NULL ? regex_call_replacer(state, program, match, subject)
                                         : regex_substitute(state, program, match, subject);
    state->last = regex_group_end(match, subject, 0);
    return status;
}

static int regex_replace(const char *str, const char *pattern, const char *flags, const char *template_text,
                         void *replacer, char **out_str) {
    if (out_str == NULL) return regex_fail("invalid argument to replace");
    regex_program program;
    regex_subject subject;
    regex_subject template_subject = {0};
    if (regex_open(str, pattern, flags, &program, &subject) != 0) return -1;
    /* The replacer may compile other patterns and evict this one. */
    if (replacer != NULL && regex_program_own(&program) != 0) {
        regex_subject_free(&subject);
        return -1;
    }
    regex_replace_state state = {0};
    state.replacer = replacer;
    if (replacer == NULL) {
        if (regex_subject_init(&template_subject, template_text) != 0) {
            regex_subject_free(&subject);
            return -1;
        }
        state.template_units = template_subject.units;
        state.template_length = template_subject.length;
    }
    int status = regex_scan(&program, &subject, (program.flags & REGEX_FLAG_GLOBAL) != 0, regex_replace_visit, &state);
    if (status == 0) status = regex_builder_append(&state.out, subject.units + state.last, subject.length - state.last);
    if (status == 0) {
        *out_str = regex_units_text(state.out.units, state.out.length);
        if (*out_str == NULL) status = regex_fail("regex: out of memory");
    }
    free(state.out.units);
    if (replacer != NULL) free((void *)program.bytecode);
    if (replacer == NULL) regex_subject_free(&template_subject);
    regex_subject_free(&subject);
    return status;
}

/* scriptgo_string_replace_regex is str.replace(regexp, replacement) with the
 * $ patterns of GetSubstitution; a global pattern replaces every match. */
int scriptgo_string_replace_regex(const char *str, const char *pattern, const char *flags, const char *repl,
                                  char **out_str) {
    return regex_replace(str, pattern, flags, repl != NULL ? repl : "", NULL, out_str);
}

/* scriptgo_string_replace_regex_fn is str.replace(regexp, replacer). */
int scriptgo_string_replace_regex_fn(const char *str, const char *pattern, const char *flags, void *replacer,
                                     char **out_str) {
    if (replacer == NULL) return regex_throw("TypeError: replacer is not a function");
    return regex_replace(str, pattern, flags, NULL, replacer, out_str);
}

/* scriptgo_string_replace_all_regex is str.replaceAll(regexp, replacement),
 * which requires a global pattern. */
int scriptgo_string_replace_all_regex(const char *str, const char *pattern, const char *flags, const char *repl,
                                      char **out_str) {
    if (flags == NULL || strchr(flags, 'g') == NULL) {
        return regex_throw("TypeError: String.prototype.replaceAll called with a non-global RegExp argument");
    }
    return scriptgo_string_replace_regex(str, pattern, flags, repl, out_str);
}

int scriptgo_string_replace_all_regex_fn(const char *str, const char *pattern, const char *flags, void *replacer,
                                         char **out_str) {
    if (flags == NULL || strchr(flags, 'g') == NULL) {
        return regex_throw("TypeError: String.prototype.replaceAll called with a non-global RegExp argument");
    }
    return scriptgo_string_replace_regex_fn(str, pattern, flags, replacer, out_str);
}

/* scriptgo_string_split_regex is str.split(regexp, limit): the pieces
 * between matches tried at each position, with the captures of each match
 * spliced in. A limit below 0 (the omitted argument) or undefined is 2**32-1. */
int scriptgo_string_split_regex(const char *str, const char *pattern, const char *flags, double limit,
                                void **out_array) {
    if (out_array == NULL) return regex_fail("invalid argument to split");
    uint32_t lim = 0xFFFFFFFFu;
    if (limit == limit && limit >= 0.0) lim = limit >= 4294967295.0 ? 0xFFFFFFFFu : (uint32_t)limit;
    if (lim == 0) return scriptgo_array_new_tagged(0, sizeof(char *), 4, out_array);
    /* The splitter is the pattern made sticky, so it only matches where it
     * is tried. */
    char sticky[16];
    size_t flag_count = 0;
    for (const char *p = flags != NULL ? flags : ""; *p != 0 && flag_count + 2 < sizeof(sticky); p++) {
        if (*p != 'y') sticky[flag_count++] = *p;
    }
    sticky[flag_count++] = 'y';
    sticky[flag_count] = '\0';
    regex_program program;
    regex_subject subject;
    regex_match match;
    if (regex_open(str, pattern, sticky, &program, &subject) != 0) return -1;
    if (regex_match_init(&match, &program) != 0) {
        regex_subject_free(&subject);
        return -1;
    }
    regex_list pieces = {0};
    int status = 0;
    if (subject.length == 0) {
        status = regex_exec_at(&program, &subject, 0, &match);
        if (status == 0) status = regex_list_push(&pieces, (void *)subject.text);
    } else {
        size_t p = 0, q = 0;
        while (status >= 0 && q < subject.length) {
            int found = regex_exec_at(&program, &subject, q, &match);
            if (found < 0) {
                status = -1;
                break;
            }
            size_t e = found > 0 ? regex_group_end(&match, &subject, 0) : 0;
            if (found == 0 || e == p) {
                q = regex_advance(&program, &subject, q);
                continue;
            }
            char *piece = regex_units_text(subject.units + p, q - p);
            if (piece == NULL || regex_list_push(&pieces, piece) != 0) {
                status = -1;
                break;
            }
            if ((uint32_t)pieces.count == lim) break;
            p = e;
            for (int group = 1; group < program.capture_count; group++) {
                if (regex_list_push(&pieces, (void *)regex_group_text(&match, &subject, group)) != 0) {
                    status = -1;
                    break;
                }
                if ((uint32_t)pieces.count == lim) break;
            }
            if (status < 0 || (uint32_t)pieces.count == lim) break;
            q = p;
        }
        if (status >= 0 && (uint32_t)pieces.count < lim) {
            char *rest = regex_units_text(subject.units + p, subject.length - p);
            if (rest == NULL || regex_list_push(&pieces, rest) != 0) status = -1;
        }
    }
    regex_match_free(&match);
    regex_subject_free(&subject);
    if (status < 0) {
        free(pieces.items);
        return -1;
    }
    return regex_list_array(&pieces, 4, out_array);
}
