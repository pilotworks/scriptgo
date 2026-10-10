/* Console layout: breaks a value's single-line inspection into lines as
 * Node.js util.inspect does with its console defaults (breakLength 80,
 * compact 3). The inspection services render every value on one line
 * ("{ a: 1, b: [ 1, 2 ] }"); this pass parses that text back into
 * containers and entries and re-joins them with Node's rules:
 *
 *   - a container prints on one line when its entries fit in breakLength
 *     and its innermost nesting (the last container reached while
 *     formatting it) is fewer than compact levels deep;
 *   - otherwise each entry goes on its own line, indented two spaces per
 *     level;
 *   - arrays of more than six entries are grouped into aligned columns
 *     (numbers right-aligned);
 *   - strings holding line breaks that exceed the line are split into
 *     concatenated pieces, one per line.
 *
 * Text that does not parse as an inspection (a multi-line error, an
 * unbalanced bracket) is left as it is. */

#include <math.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#define CONSOLE_LAYOUT_BREAK_LENGTH 80
#define CONSOLE_LAYOUT_COMPACT 3
#define CONSOLE_LAYOUT_MIN_LINE_WIDTH 16

typedef struct {
    char *data;
    size_t length;
    size_t capacity;
    int failed;
} console_layout_text;

static void console_layout_append(console_layout_text *text, const char *data, size_t length) {
    if (text->failed) return;
    if (text->length + length + 1 > text->capacity) {
        size_t capacity = text->capacity == 0 ? 64 : text->capacity;
        while (text->length + length + 1 > capacity) capacity *= 2;
        char *grown = realloc(text->data, capacity);
        if (grown == NULL) {
            text->failed = 1;
            return;
        }
        text->data = grown;
        text->capacity = capacity;
    }
    memcpy(text->data + text->length, data, length);
    text->length += length;
    text->data[text->length] = '\0';
}

static void console_layout_append_str(console_layout_text *text, const char *data) {
    console_layout_append(text, data, strlen(data));
}

static void console_layout_spaces(console_layout_text *text, size_t count) {
    for (size_t i = 0; i < count; i++) console_layout_append(text, " ", 1);
}

/* console_layout_width is a string's length in UTF-16 code units, the
 * length util.inspect measures. */
static size_t console_layout_width(const char *data, size_t length) {
    size_t width = 0;
    for (size_t i = 0; i < length; i++) {
        unsigned char c = (unsigned char)data[i];
        if ((c & 0xC0) == 0x80) continue;
        width += c >= 0xF0 ? 2 : 1;
    }
    return width;
}

/* console_layout_skip_quoted returns the index just past the quoted string
 * starting at at, or 0 when it is unterminated. */
static size_t console_layout_skip_quoted(const char *s, size_t at, size_t end) {
    char quote = s[at];
    for (size_t i = at + 1; i < end; i++) {
        if (s[i] == '\\') {
            i++;
            continue;
        }
        if (s[i] == quote) return i + 1;
    }
    return 0;
}

static int console_layout_is_quote(char c) {
    return c == '\'' || c == '"' || c == '`';
}

/* console_layout_match returns the index of the bracket closing the one at
 * at, or 0 when the text is unbalanced. */
static size_t console_layout_match(const char *s, size_t at, size_t end) {
    int depth = 0;
    for (size_t i = at; i < end; i++) {
        char c = s[i];
        if (console_layout_is_quote(c)) {
            size_t next = console_layout_skip_quoted(s, i, end);
            if (next == 0) return 0;
            i = next - 1;
        } else if (c == '\\') {
            i++;
        } else if (c == '{' || c == '[' || c == '(') {
            depth++;
        } else if (c == '}' || c == ']' || c == ')') {
            depth--;
            if (depth == 0) return i;
            if (depth < 0) return 0;
        }
    }
    return 0;
}

/* console_layout_opener reports whether the bracket at at opens an
 * inspected container ("{ ", "[ ", "{}", "[]"), not an atom such as
 * "[Function: f]". */
static int console_layout_opener(const char *s, size_t at, size_t start, size_t end) {
    char c = s[at];
    if (c != '{' && c != '[') return 0;
    if (at > start && s[at - 1] != ' ') return 0;
    if (at + 1 >= end) return 0;
    char next = s[at + 1];
    return next == ' ' || (c == '{' && next == '}') || (c == '[' && next == ']');
}

typedef struct console_layout_node console_layout_node;

typedef struct {
    /* key is the "name: " text before the value, if any. */
    size_t key_start, key_length;
    console_layout_node *value;
    /* map_value is the value of a Map entry ("key => value"), whose key is
     * value. */
    console_layout_node *map_value;
} console_layout_entry;

struct console_layout_node {
    size_t start, end;      /* the node's text */
    size_t open;            /* index of the container bracket, if any */
    int container;          /* whether the node is a non-empty container */
    console_layout_entry *entries;
    size_t count;
};

static void console_layout_free(console_layout_node *node) {
    if (node == NULL) return;
    for (size_t i = 0; i < node->count; i++) {
        console_layout_free(node->entries[i].value);
        console_layout_free(node->entries[i].map_value);
    }
    free(node->entries);
    free(node);
}

static console_layout_node *console_layout_parse(const char *s, size_t start, size_t end);

/* console_layout_key_length is the length of a leading "name: " property
 * key (an identifier, a quoted string or a [Symbol(...)] key), or 0. */
static size_t console_layout_key_length(const char *s, size_t start, size_t end) {
    size_t i = start;
    if (i >= end) return 0;
    if (console_layout_is_quote(s[i])) {
        i = console_layout_skip_quoted(s, i, end);
        if (i == 0) return 0;
    } else if (s[i] == '[') {
        size_t close = console_layout_match(s, i, end);
        if (close == 0) return 0;
        i = close + 1;
    } else {
        while (i < end) {
            unsigned char c = (unsigned char)s[i];
            if (!((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$' || c >= 0x80)) break;
            i++;
        }
        if (i == start) return 0;
    }
    if (i + 1 < end && s[i] == ':' && s[i + 1] == ' ') return i + 2 - start;
    return 0;
}

/* console_layout_find_arrow returns the index of a top-level " => " (a Map
 * entry) in [start, end), or 0. */
static size_t console_layout_find_arrow(const char *s, size_t start, size_t end) {
    for (size_t i = start; i < end; i++) {
        char c = s[i];
        if (console_layout_is_quote(c)) {
            size_t next = console_layout_skip_quoted(s, i, end);
            if (next == 0) return 0;
            i = next - 1;
        } else if (c == '{' || c == '[' || c == '(') {
            size_t close = console_layout_match(s, i, end);
            if (close == 0) return 0;
            i = close;
        } else if (c == ' ' && i + 3 < end && s[i + 1] == '=' && s[i + 2] == '>' && s[i + 3] == ' ') {
            return i;
        }
    }
    return 0;
}

static int console_layout_parse_entry(const char *s, size_t start, size_t end, console_layout_entry *entry) {
    memset(entry, 0, sizeof(*entry));
    size_t key = console_layout_key_length(s, start, end);
    entry->key_start = start;
    entry->key_length = key;
    size_t value_start = start + key;
    size_t arrow = key == 0 ? console_layout_find_arrow(s, value_start, end) : 0;
    if (arrow != 0) {
        entry->value = console_layout_parse(s, value_start, arrow);
        entry->map_value = console_layout_parse(s, arrow + 4, end);
        return entry->value != NULL && entry->map_value != NULL;
    }
    entry->value = console_layout_parse(s, value_start, end);
    return entry->value != NULL;
}

/* console_layout_parse reads one inspected value: a prefix ("Map(1) ",
 * "[Object: null prototype] ") and a container, or an atom. */
static console_layout_node *console_layout_parse(const char *s, size_t start, size_t end) {
    console_layout_node *node = calloc(1, sizeof(*node));
    if (node == NULL) return NULL;
    node->start = start;
    node->end = end;
    size_t open = end;
    for (size_t i = start; i < end; i++) {
        if (console_layout_is_quote(s[i])) {
            size_t next = console_layout_skip_quoted(s, i, end);
            if (next == 0) break;
            i = next - 1;
        } else if (console_layout_opener(s, i, start, end)) {
            open = i;
            break;
        }
    }
    if (open == end) return node;
    size_t close = console_layout_match(s, open, end);
    /* An empty container ("{}", "Foo {}") or one followed by more text
     * stays as written. */
    if (close == 0 || close != end - 1 || close == open + 1) return node;
    if (close < open + 3 || s[open + 1] != ' ' || s[close - 1] != ' ') return node;
    node->open = open;
    node->container = 1;
    size_t at = open + 2;
    size_t content_end = close - 1;
    while (at <= content_end) {
        size_t entry_end = content_end;
        for (size_t i = at; i < content_end; i++) {
            char c = s[i];
            if (console_layout_is_quote(c)) {
                size_t next = console_layout_skip_quoted(s, i, content_end);
                if (next == 0) {
                    console_layout_free(node);
                    return NULL;
                }
                i = next - 1;
            } else if (c == '{' || c == '[' || c == '(') {
                size_t match = console_layout_match(s, i, content_end);
                if (match == 0) {
                    console_layout_free(node);
                    return NULL;
                }
                i = match;
            } else if (c == ',' && i + 1 < content_end && s[i + 1] == ' ') {
                entry_end = i;
                break;
            }
        }
        console_layout_entry *grown = realloc(node->entries, (node->count + 1) * sizeof(*grown));
        if (grown == NULL) {
            console_layout_free(node);
            return NULL;
        }
        node->entries = grown;
        if (!console_layout_parse_entry(s, at, entry_end, &node->entries[node->count++])) {
            console_layout_free(node);
            return NULL;
        }
        at = entry_end + 2;
    }
    return node;
}

typedef struct {
    char *text;
    size_t width;
} console_layout_piece;

static char *console_layout_format(const char *s, const console_layout_node *node, size_t indent, int depth, int *last_depth);

/* console_layout_numeric reports whether an entry prints a number or a
 * bigint, which grouped columns align to the right. */
static int console_layout_numeric(const char *text) {
    if (strcmp(text, "NaN") == 0 || strcmp(text, "Infinity") == 0 || strcmp(text, "-Infinity") == 0) return 1;
    const char *p = text;
    if (*p == '-') p++;
    if (*p < '0' || *p > '9') return 0;
    for (; *p != '\0'; p++) {
        if (!((*p >= '0' && *p <= '9') || *p == '.' || *p == 'e' || *p == '+' || *p == '-' || (*p == 'n' && p[1] == '\0'))) return 0;
    }
    return 1;
}

static int console_layout_more_items(const char *text) {
    size_t length = strlen(text);
    return strncmp(text, "... ", 4) == 0 &&
           ((length >= 10 && strcmp(text + length - 10, "more items") == 0) ||
            (length >= 9 && strcmp(text + length - 9, "more item") == 0));
}

/* console_layout_group arranges more than six array entries in columns
 * (util.inspect's groupArrayElements). It returns the rows, or NULL to keep
 * the entries as they are. */
static console_layout_piece *console_layout_group(console_layout_piece *output, size_t count, size_t indent, size_t *out_rows) {
    size_t output_length = count;
    if (count > 0 && console_layout_more_items(output[count - 1].text)) output_length--;
    if (output_length == 0) return NULL;
    const size_t separator_space = 2;
    size_t total_length = 0, max_length = 0;
    for (size_t i = 0; i < output_length; i++) {
        total_length += output[i].width + separator_space;
        if (output[i].width > max_length) max_length = output[i].width;
    }
    size_t actual_max = max_length + separator_space;
    if (!(actual_max * 3 + indent < CONSOLE_LAYOUT_BREAK_LENGTH &&
          ((double)total_length / (double)actual_max > 5 || max_length <= 6))) {
        return NULL;
    }
    double average_bias = sqrt((double)actual_max - (double)total_length / (double)count);
    double biased_max = fmax((double)actual_max - 3 - average_bias, 1);
    double columns_f = fmin(fmin(round(sqrt(2.5 * biased_max * (double)output_length) / biased_max),
                                 floor((double)(CONSOLE_LAYOUT_BREAK_LENGTH - indent) / (double)actual_max)),
                            fmin((double)(CONSOLE_LAYOUT_COMPACT * 4), 15));
    if (columns_f <= 1) return NULL;
    size_t columns = (size_t)columns_f;
    size_t *max_line_length = calloc(columns, sizeof(size_t));
    if (max_line_length == NULL) return NULL;
    for (size_t i = 0; i < columns; i++) {
        size_t line_length = 0;
        for (size_t j = i; j < output_length; j += columns) {
            if (output[j].width > line_length) line_length = output[j].width;
        }
        max_line_length[i] = line_length + separator_space;
    }
    int pad_start = 1;
    for (size_t i = 0; i < output_length; i++) {
        if (!console_layout_numeric(output[i].text)) {
            pad_start = 0;
            break;
        }
    }
    size_t rows_capacity = output_length / columns + 2;
    console_layout_piece *rows = calloc(rows_capacity, sizeof(*rows));
    if (rows == NULL) {
        free(max_line_length);
        return NULL;
    }
    size_t row_count = 0;
    for (size_t i = 0; i < output_length; i += columns) {
        size_t max = i + columns < output_length ? i + columns : output_length;
        console_layout_text row = {0};
        size_t j = i;
        for (; j + 1 < max; j++) {
            size_t cell = output[j].width + 2;
            size_t padding = max_line_length[j - i] > cell ? max_line_length[j - i] - cell : 0;
            if (pad_start) console_layout_spaces(&row, padding);
            console_layout_append_str(&row, output[j].text);
            console_layout_append(&row, ", ", 2);
            if (!pad_start) console_layout_spaces(&row, padding);
        }
        if (pad_start) {
            size_t target = max_line_length[j - i] - separator_space;
            if (target > output[j].width) console_layout_spaces(&row, target - output[j].width);
        }
        console_layout_append_str(&row, output[j].text);
        if (row.failed) {
            free(row.data);
            for (size_t k = 0; k < row_count; k++) free(rows[k].text);
            free(rows);
            free(max_line_length);
            return NULL;
        }
        rows[row_count].text = row.data;
        rows[row_count].width = console_layout_width(row.data, row.length);
        row_count++;
    }
    if (output_length < count) {
        rows[row_count].text = strdup(output[count - 1].text);
        rows[row_count].width = output[count - 1].width;
        row_count++;
    }
    free(max_line_length);
    *out_rows = row_count;
    return rows;
}

/* console_layout_string splits a quoted string atom holding line breaks
 * into one quoted piece per line when it exceeds the line
 * (util.inspect's formatPrimitive). It returns NULL to keep the atom. */
static char *console_layout_string(const char *s, size_t start, size_t end, size_t indent) {
    if (end - start < 2 || !console_layout_is_quote(s[start]) || console_layout_skip_quoted(s, start, end) != end) return NULL;
    size_t length = 0;
    int breaks = 0;
    for (size_t i = start + 1; i + 1 < end; i++) {
        if (s[i] == '\\') {
            if (s[i + 1] == 'n') breaks = 1;
            i++;
        }
        if (((unsigned char)s[i] & 0xC0) != 0x80) length++;
    }
    if (!breaks || length <= CONSOLE_LAYOUT_MIN_LINE_WIDTH || length + indent + 4 <= CONSOLE_LAYOUT_BREAK_LENGTH) return NULL;
    char quote = s[start];
    console_layout_text text = {0};
    size_t piece = start + 1;
    for (size_t i = start + 1; i + 1 < end; i++) {
        if (s[i] != '\\') continue;
        if (s[i + 1] == 'n' && i + 2 < end - 1) {
            console_layout_append(&text, &quote, 1);
            console_layout_append(&text, s + piece, i + 2 - piece);
            console_layout_append(&text, &quote, 1);
            console_layout_append_str(&text, " +\n");
            console_layout_spaces(&text, indent + 2);
            piece = i + 2;
        }
        i++;
    }
    console_layout_append(&text, &quote, 1);
    console_layout_append(&text, s + piece, end - 1 - piece);
    console_layout_append(&text, &quote, 1);
    if (text.failed) {
        free(text.data);
        return NULL;
    }
    return text.data;
}

static char *console_layout_entry_text(const char *s, const console_layout_entry *entry, size_t indent, int depth, int *last_depth) {
    console_layout_text text = {0};
    console_layout_append(&text, s + entry->key_start, entry->key_length);
    char *value = console_layout_format(s, entry->value, indent, depth, last_depth);
    if (value == NULL) {
        free(text.data);
        return NULL;
    }
    console_layout_append_str(&text, value);
    free(value);
    if (entry->map_value != NULL) {
        console_layout_append_str(&text, " => ");
        value = console_layout_format(s, entry->map_value, indent, depth, last_depth);
        if (value == NULL) {
            free(text.data);
            return NULL;
        }
        console_layout_append_str(&text, value);
        free(value);
    }
    if (text.failed) {
        free(text.data);
        return NULL;
    }
    return text.data;
}

/* console_layout_format lays out node, whose lines are indented by indent
 * spaces; depth is its nesting level and last_depth receives the level of
 * the last container formatted within it (util.inspect's currentDepth). */
static char *console_layout_format(const char *s, const console_layout_node *node, size_t indent, int depth, int *last_depth) {
    if (!node->container) {
        char *split = console_layout_string(s, node->start, node->end, indent);
        if (split != NULL) return split;
        char *atom = malloc(node->end - node->start + 1);
        if (atom == NULL) return NULL;
        memcpy(atom, s + node->start, node->end - node->start);
        atom[node->end - node->start] = '\0';
        return atom;
    }
    *last_depth = depth;
    console_layout_piece *output = calloc(node->count, sizeof(*output));
    if (output == NULL) return NULL;
    int has_newline = 0;
    for (size_t i = 0; i < node->count; i++) {
        output[i].text = console_layout_entry_text(s, &node->entries[i], indent + 2, depth + 1, last_depth);
        if (output[i].text == NULL) {
            for (size_t k = 0; k < i; k++) free(output[k].text);
            free(output);
            return NULL;
        }
        output[i].width = console_layout_width(output[i].text, strlen(output[i].text));
        if (strchr(output[i].text, '\n') != NULL) has_newline = 1;
    }
    size_t count = node->count;
    int grouped = 0;
    if (s[node->open] == '[' && count > 6) {
        size_t rows = 0;
        console_layout_piece *grouped_output = console_layout_group(output, count, indent, &rows);
        if (grouped_output != NULL) {
            for (size_t k = 0; k < count; k++) free(output[k].text);
            free(output);
            output = grouped_output;
            grouped = rows != count;
            count = rows;
        }
    }
    size_t brace_length = console_layout_width(s + node->start, node->open + 1 - node->start);
    char close = s[node->end - 1];
    console_layout_text text = {0};
    int single_line = 0;
    if (*last_depth - depth < CONSOLE_LAYOUT_COMPACT && !grouped && !has_newline) {
        size_t start = count + indent + brace_length + 10;
        size_t total = count + start;
        single_line = total + count <= CONSOLE_LAYOUT_BREAK_LENGTH;
        for (size_t i = 0; single_line && i < count; i++) {
            total += output[i].width;
            if (total > CONSOLE_LAYOUT_BREAK_LENGTH) single_line = 0;
        }
    }
    console_layout_append(&text, s + node->start, node->open + 1 - node->start);
    if (single_line) {
        for (size_t i = 0; i < count; i++) {
            console_layout_append(&text, i == 0 ? " " : ", ", i == 0 ? 1 : 2);
            console_layout_append_str(&text, output[i].text);
        }
        console_layout_append(&text, " ", 1);
    } else {
        for (size_t i = 0; i < count; i++) {
            console_layout_append(&text, i == 0 ? "\n" : ",\n", i == 0 ? 1 : 2);
            console_layout_spaces(&text, indent + 2);
            console_layout_append_str(&text, output[i].text);
        }
        console_layout_append(&text, "\n", 1);
        console_layout_spaces(&text, indent);
    }
    console_layout_append(&text, &close, 1);
    for (size_t i = 0; i < count; i++) free(output[i].text);
    free(output);
    if (text.failed) {
        free(text.data);
        return NULL;
    }
    return text.data;
}

/* scriptgo_console_layout lays out a single-line inspection as console.log
 * prints it. The result is a new string; text that does not parse as an
 * inspection is copied unchanged. */
int scriptgo_console_layout(const char *inspected, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("invalid console layout");
    *out_str = NULL;
    if (inspected == NULL) inspected = "";
    size_t length = strlen(inspected);
    char *result = NULL;
    if (memchr(inspected, '\n', length) == NULL) {
        console_layout_node *node = console_layout_parse(inspected, 0, length);
        if (node != NULL) {
            int last_depth = 0;
            result = console_layout_format(inspected, node, 0, 0, &last_depth);
            console_layout_free(node);
        }
    }
    if (result == NULL) result = strdup(inspected);
    if (result == NULL) return scriptgo_runtime_set_error("console layout allocation failed");
    *out_str = result;
    return 0;
}

