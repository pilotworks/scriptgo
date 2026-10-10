#include <ctype.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifndef YYJSON_H
#include "yyjson.h"
#endif

#ifndef SCRIPTGO_VALUE_H
#include "scriptgo_value.h"
#endif

int scriptgo_runtime_set_error(const char *message);

static int json_fail(const char *message) { return scriptgo_runtime_set_error(message); }

extern const char scriptgo_undefined_sentinel;

typedef scriptgo_value scriptgo_json_unknown;

typedef struct {
    char *buf;
    size_t len;
    size_t cap;
    /* inspect renders non-finite numbers as console.log does (NaN,
     * Infinity) instead of JSON's null. */
    int inspect;
} json_builder;

static inline int jb_reserve(json_builder *b, size_t extra) {
    if (b->len + extra + 1 > b->cap) {
        size_t next = b->cap == 0 ? 256 : b->cap * 2;
        while (next < b->len + extra + 1) next *= 2;
        char *grown = (char *)realloc(b->buf, next);
        if (grown == NULL) return -1;
        b->buf = grown;
        b->cap = next;
    }
    return 0;
}

static inline int jb_char(json_builder *b, char c) {
    if (jb_reserve(b, 1) != 0) return -1;
    b->buf[b->len++] = c;
    b->buf[b->len] = '\0';
    return 0;
}

static inline int jb_append(json_builder *b, const char *s, size_t len) {
    if (s == NULL || len == 0) return 0;
    if (jb_reserve(b, len) != 0) return -1;
    memcpy(b->buf + b->len, s, len);
    b->len += len;
    b->buf[b->len] = '\0';
    return 0;
}

static inline int jb_string(json_builder *b, const char *value, size_t len) {
    if (value == NULL) {
        return jb_append(b, "null", 4);
    }
    if (value == &scriptgo_undefined_sentinel || (len == 9 && strcmp(value, "undefined") == 0)) {
        return jb_append(b, "undefined", 9);
    }
    if (jb_reserve(b, len * 6 + 3) != 0) return -1;
    char *dst = b->buf + b->len;
    *dst++ = '"';
    for (size_t i = 0; i < len; i++) {
        unsigned char c = (unsigned char)value[i];
        if (c == '"') { *dst++ = '\\'; *dst++ = '"'; }
        else if (c == '\\') { *dst++ = '\\'; *dst++ = '\\'; }
        else if (c == '\n') { *dst++ = '\\'; *dst++ = 'n'; }
        else if (c == '\r') { *dst++ = '\\'; *dst++ = 'r'; }
        else if (c == '\t') { *dst++ = '\\'; *dst++ = 't'; }
        else if (c == '\b') { *dst++ = '\\'; *dst++ = 'b'; }
        else if (c == '\f') { *dst++ = '\\'; *dst++ = 'f'; }
        else if (c < 0x20) {
            dst += snprintf(dst, 7, "\\u%04x", c);
        } else {
            *dst++ = (char)c;
        }
    }
    *dst++ = '"';
    *dst = '\0';
    b->len = (size_t)(dst - b->buf);
    return 0;
}

static inline int jb_number(json_builder *b, double value) {
    if (isnan(value) || isinf(value)) {
        if (b->inspect) {
            if (isnan(value)) return jb_append(b, "NaN", 3);
            return value > 0 ? jb_append(b, "Infinity", 8) : jb_append(b, "-Infinity", 9);
        }
        return jb_append(b, "null", 4);
    }
    if (jb_reserve(b, 32) != 0) return -1;
    char number[32];
    scriptgo_number_format(value, number, sizeof(number));
    int written = snprintf(b->buf + b->len, 32, "%s", number);
    if (written > 0) {
        b->len += (size_t)written;
        b->buf[b->len] = '\0';
    }
    return 0;
}

int scriptgo_json_stringify_number(double value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    const char *marker = scriptgo_number_marker_name(value);
    if (marker != NULL && strcmp(marker, "undefined") == 0) {
        /* undefined has no JSON text; callers omit the property. */
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    json_builder b = {0};
    if (jb_number(&b, value) != 0) { free(b.buf); return json_fail("scriptgo json allocation failed"); }
    *out_str = b.buf ? b.buf : strdup("null");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_json_stringify_bool(int value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    *out_str = strdup(value ? "true" : "false");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_json_stringify_string(const char *value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    if (value == NULL) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (value == &scriptgo_undefined_sentinel || strcmp(value, "undefined") == 0) {
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (jb_string(&b, value, strlen(value)) != 0) { free(b.buf); return json_fail("scriptgo json allocation failed"); }
    *out_str = b.buf ? b.buf : strdup("\"\"");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
    int64_t integrity;
    void *properties;
} scriptgo_array_internal;

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set_tag(void *handle, int64_t tag);

static int json_builder_number_array(json_builder *b, const scriptgo_array_internal *arr) {
    if (jb_char(b, '[') != 0) return -1;
    const double *data = (const double *)arr->data;
    for (int64_t i = 0; i < arr->length; i++) {
        if (i > 0 && jb_char(b, ',') != 0) return -1;
        if (jb_number(b, data[i]) != 0) return -1;
    }
    return jb_char(b, ']');
}

static int json_builder_string_array(json_builder *b, const scriptgo_array_internal *arr) {
    if (jb_char(b, '[') != 0) return -1;
    for (int64_t i = 0; i < arr->length; i++) {
        if (i > 0 && jb_char(b, ',') != 0) return -1;
        const char *elem = *(const char **)(arr->data + (size_t)i * sizeof(char *));
        const char *s = elem != NULL ? elem : "";
        if (jb_string(b, s, strlen(s)) != 0) return -1;
    }
    return jb_char(b, ']');
}
int scriptgo_array_push(void *handle, const void *value, double *out_length);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_get(void *handle, double index, void *out_value);
int scriptgo_array_length(void *handle, int64_t *out_length);
int scriptgo_array_release(void *handle);
int scriptgo_object_new(int64_t field_count, void **out_object);
int scriptgo_object_type_set(void *handle, const char *type_name);
int scriptgo_object_unknown_set(void *handle, int64_t index, const scriptgo_value *value);
int scriptgo_object_unknown_get(void *handle, int64_t index, scriptgo_value *out_value);
int scriptgo_object_keys(void *handle, void **out_array);
int scriptgo_object_property_unknown_get(void *handle, const char *property, scriptgo_value *out_value);
int scriptgo_string_from_object(void *obj, char **out_str);
int scriptgo_object_region_contains(const void *ptr);
int scriptgo_json_arena_contains(const void *ptr);

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    uint8_t type_name_owned;
    uint32_t capacity;
    scriptgo_value *boxed_fields;
    uintptr_t fields[];
} scriptgo_json_object;

int scriptgo_json_stringify_unknown(const scriptgo_value *value, char **out_str);
static int json_builder_value(json_builder *b, const scriptgo_value *value);
static int json_builder_object(json_builder *b, void *handle);

#ifndef SCRIPTGO_OBJECT_NAN_BITS
#define SCRIPTGO_OBJECT_NAN_BITS SCRIPTGO_NUMBER_UNDEFINED_BITS
#endif
#ifndef SCRIPTGO_OBJECT_NULL_BITS
#define SCRIPTGO_OBJECT_NULL_BITS SCRIPTGO_NUMBER_NULL_BITS
#endif

// Parsed JSON objects store plain runtime values, so avoid the generic accessor's
// cloning and validation path while preserving it for objects with boxed fields.
static inline void json_object_field_value(const scriptgo_json_object *obj, int64_t index,
                                           scriptgo_value *out_value) {
    memset(out_value, 0, sizeof(*out_value));
    if (index < 0 || index >= obj->field_count) return;

    uintptr_t value = obj->fields[index];
    if (value == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) return;
    if (value == (uintptr_t)SCRIPTGO_OBJECT_NULL_BITS) {
        out_value->tag = SCRIPTGO_TAG_NULL;
    } else if (((uint64_t)value >> 32) == 2) {
        out_value->tag = SCRIPTGO_TAG_BOOLEAN;
        out_value->payload = value & 1;
    } else if (value == 0 || (value & 0xFFF0000000000000ULL) != 0) {
        out_value->tag = SCRIPTGO_TAG_NUMBER;
        out_value->payload = (uint64_t)value;
    } else {
        int gc_tag = scriptgo_gc_get_tag((void *)value);
        if (gc_tag == 2) {
            out_value->tag = SCRIPTGO_TAG_ARRAY;
        } else if (gc_tag == 3) {
            out_value->tag = SCRIPTGO_TAG_FUNCTION;
        } else if (gc_tag == 11) {
            out_value->tag = SCRIPTGO_TAG_SYMBOL;
        } else if (gc_tag != 0) {
            out_value->tag = SCRIPTGO_TAG_OBJECT;
        } else if ((scriptgo_object_region_contains((void *)value) || scriptgo_json_arena_contains((void *)value)) &&
                   *(uint64_t *)value == SCRIPTGO_OBJECT_MAGIC) {
            out_value->tag = SCRIPTGO_TAG_OBJECT;
        } else {
            out_value->tag = SCRIPTGO_TAG_STRING;
        }
        out_value->payload = (uint64_t)value;
    }
}

static inline void json_object_store_value(scriptgo_json_object *obj, int64_t index,
                                           const scriptgo_value *value) {
    if (value->tag == SCRIPTGO_TAG_BOOLEAN) {
        obj->fields[index] = (uintptr_t)((2ULL << 32) | (value->payload != 0));
    } else if (value->tag == SCRIPTGO_TAG_NULL) {
        obj->fields[index] = (uintptr_t)SCRIPTGO_OBJECT_NULL_BITS;
    } else if (value->tag == SCRIPTGO_TAG_UNDEFINED) {
        obj->fields[index] = (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS;
    } else {
        obj->fields[index] = (uintptr_t)value->payload;
    }
}

int scriptgo_json_stringify_number_array(void *handle, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (handle == NULL) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (handle == &scriptgo_undefined_sentinel) {
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    scriptgo_array_internal *arr = (scriptgo_array_internal *)handle;
    json_builder b = {0};
    if (json_builder_number_array(&b, arr) != 0) goto fail;
    *out_str = b.buf ? b.buf : strdup("[]");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
fail:
    free(b.buf);
    return json_fail("scriptgo json allocation failed");
}

int scriptgo_json_stringify_bool_array(void *handle, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (handle == NULL) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (handle == &scriptgo_undefined_sentinel) {
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    scriptgo_array_internal *array = (scriptgo_array_internal *)handle;
    if (array->element_size <= 0) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    if (jb_char(&b, '[') != 0) goto fail;
    for (int64_t i = 0; i < array->length; i++) {
        if (i > 0 && jb_char(&b, ',') != 0) goto fail;
        uint8_t elem = *(uint8_t *)(array->data + (size_t)i);
        if (jb_append(&b, elem ? "true" : "false", elem ? 4 : 5) != 0) goto fail;
    }
    if (jb_char(&b, ']') != 0) goto fail;
    *out_str = b.buf ? b.buf : strdup("[]");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
fail:
    free(b.buf);
    return json_fail("scriptgo json allocation failed");
}

int scriptgo_json_stringify_string_array(void *handle, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (handle == NULL) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (handle == &scriptgo_undefined_sentinel) {
        *out_str = strdup("undefined");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    scriptgo_array_internal *array = (scriptgo_array_internal *)handle;
    if (array->element_size <= 0) return json_fail("scriptgo json invalid argument");
    if (array->element_size == 1) return scriptgo_json_stringify_bool_array(handle, out_str);
    json_builder b = {0};
    if (json_builder_string_array(&b, array) != 0) goto fail;
    *out_str = b.buf ? b.buf : strdup("[]");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
fail:
    free(b.buf);
    return json_fail("scriptgo json allocation failed");
}

// Object arrays are common for typed JSON payloads. Serialize them directly
// into one growing buffer rather than lowering every element into temporary
// strings and repeatedly concatenating those strings.
int scriptgo_json_stringify_object_array(void *handle, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (handle == NULL) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    scriptgo_array_internal *array = (scriptgo_array_internal *)handle;
    if (array->element_size != (int64_t)sizeof(void *)) return json_fail("scriptgo json invalid object array");
    json_builder b = {0};
    if (jb_char(&b, '[') != 0) goto fail;
    for (int64_t i = 0; i < array->length; i++) {
        if (i > 0 && jb_char(&b, ',') != 0) goto fail;
        void *object = *(void **)(array->data + (size_t)i * sizeof(void *));
        if (json_builder_object(&b, object) != 0) goto fail;
    }
    if (jb_char(&b, ']') != 0) goto fail;
    *out_str = b.buf ? b.buf : strdup("[]");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
fail:
    free(b.buf);
    return json_fail("scriptgo json allocation failed");
}

typedef struct {
    json_builder *builder;
    void *handle;
    scriptgo_json_object *object;
    int dictionary;
    int has_fields;
    int failed;
    /* keep_undefined writes undefined fields (console output of a match
     * array's properties and of null-prototype objects) instead of omitting
     * them. */
    int keep_undefined;
} json_object_writer;

void *scriptgo_symbol_for_property_key(const char *key, size_t length);
int scriptgo_symbol_to_string(void *symbol, char **out_string);

static int json_write_object_field(const char *key, size_t key_len, int index, void *context) {
    json_object_writer *writer = context;
    if (index >= writer->object->field_count) return 1;
    char *symbol_label = NULL;
    if (key_len > 0 && key[0] == '\x01') {
        /* JSON omits symbol-keyed properties; console output shows them as
         * [Symbol(description)] (a \x01-prefixed key the inspect pass
         * brackets). */
        void *symbol = scriptgo_symbol_for_property_key(key, key_len);
        char *text = NULL;
        if (!writer->builder->inspect || symbol == NULL || scriptgo_symbol_to_string(symbol, &text) != 0 || text == NULL) return 0;
        size_t text_len = strlen(text);
        symbol_label = malloc(text_len + 2);
        if (symbol_label == NULL) {
            free(text);
            writer->failed = 1;
            return 1;
        }
        symbol_label[0] = '\x01';
        memcpy(symbol_label + 1, text, text_len + 1);
        free(text);
        key = symbol_label;
        key_len = text_len + 1;
    }
    scriptgo_value value;
    scriptgo_value_init_undefined(&value);
    if (writer->dictionary && writer->object->boxed_fields == NULL) {
        json_object_field_value(writer->object, index, &value);
    } else {
        scriptgo_object_unknown_get(writer->handle, index, &value);
    }
    int omit = value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL ||
               (value.tag == SCRIPTGO_TAG_UNDEFINED && !(writer->keep_undefined && writer->builder->inspect));
    if (omit) {
        free(symbol_label);
        return 0;
    }
    if ((writer->has_fields && jb_char(writer->builder, ',') != 0) ||
        jb_string(writer->builder, key, key_len) != 0 ||
        jb_char(writer->builder, ':') != 0 ||
        json_builder_value(writer->builder, &value) != 0) {
        free(symbol_label);
        writer->failed = 1;
        return 1;
    }
    free(symbol_label);
    writer->has_fields = 1;
    return 0;
}

typedef struct {
    json_builder *builder;
    const scriptgo_array_internal *names;
    const scriptgo_array_internal *texts;
    int has_fields;
    int failed;
} json_assembled_object_writer;

static int json_write_assembled_field(const char *key, size_t key_len, int index, void *context) {
    json_assembled_object_writer *writer = context;
    (void)index;
    for (int64_t i = 0; i < writer->names->length && i < writer->texts->length; i++) {
        const char *name = *(const char **)(writer->names->data + (size_t)i * sizeof(char *));
        if (name == NULL || strlen(name) != key_len || memcmp(name, key, key_len) != 0) continue;
        const char *text = *(const char **)(writer->texts->data + (size_t)i * sizeof(char *));
        /* An undefined value (or a function) has no JSON text. */
        if (text == NULL || text == &scriptgo_undefined_sentinel || strcmp(text, "undefined") == 0) return 0;
        if ((writer->has_fields && jb_char(writer->builder, ',') != 0) ||
            jb_string(writer->builder, key, key_len) != 0 ||
            jb_char(writer->builder, ':') != 0 ||
            jb_append(writer->builder, text, strlen(text)) != 0) {
            writer->failed = 1;
            return 1;
        }
        writer->has_fields = 1;
        return 0;
    }
    return 0;
}

/* scriptgo_json_assemble_object writes the object handle's properties in
 * their key order (insertion order), each as the JSON text the compiler
 * produced from its static type: texts[i] is the text for names[i], and
 * "undefined" omits the property. */
int scriptgo_json_assemble_object(void *handle, void *names_handle, void *texts_handle, char **out_str) {
    if (out_str == NULL || names_handle == NULL || texts_handle == NULL) return json_fail("scriptgo json invalid argument");
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel) {
        *out_str = strdup("null");
        return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
    }
    if (((scriptgo_json_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return json_fail("scriptgo json invalid object");
    json_builder b = {0};
    json_assembled_object_writer writer = {&b, names_handle, texts_handle, 0, 0};
    if (jb_char(&b, '{') != 0) goto fail;
    object_field_visit((const scriptgo_object *)handle, json_write_assembled_field, &writer);
    if (writer.failed || jb_char(&b, '}') != 0) goto fail;
    *out_str = b.buf;
    return 0;
fail:
    free(b.buf);
    return json_fail("scriptgo json allocation failed");
}

/* A tuple's key list is "0", "1", ... marked by a leading empty segment
 * (`::0:1:`, see the backend's objectLayoutName); it serializes and
 * inspects as an array. */
static int json_tuple_key(const char *key, size_t key_len, int index, void *context) {
    char expected[24];
    int written = snprintf(expected, sizeof(expected), "%d", index);
    int *count = context;
    if (written < 0 || (size_t)written != key_len || memcmp(expected, key, key_len) != 0) {
        *count = -1;
        return 1;
    }
    *count = index + 1;
    return 0;
}

static int json_object_is_tuple(const scriptgo_json_object *obj) {
    int count = 0;
    if (obj->type_name == NULL || strncmp(obj->type_name, "::", 2) != 0 || obj->field_count == 0) return 0;
    object_field_visit((const scriptgo_object *)obj, json_tuple_key, &count);
    return count > 0 && count == obj->field_count;
}

static int json_write_tuple_element(const char *key, size_t key_len, int index, void *context) {
    json_object_writer *writer = context;
    (void)key;
    (void)key_len;
    scriptgo_value value;
    scriptgo_value_init_undefined(&value);
    scriptgo_object_unknown_get(writer->handle, index, &value);
    if ((index > 0 && jb_char(writer->builder, ',') != 0)) {
        writer->failed = 1;
        return 1;
    }
    /* In JSON an array element without a JSON value (undefined, a function)
     * is null; console output prints the value itself. */
    if (!writer->builder->inspect &&
        (value.tag == SCRIPTGO_TAG_UNDEFINED || value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL)) {
        if (jb_append(writer->builder, "null", 4) != 0) writer->failed = 1;
        return writer->failed;
    }
    if (json_builder_value(writer->builder, &value) != 0) {
        writer->failed = 1;
        return 1;
    }
    return 0;
}

/* json_append_class_name writes "Name " before a class instance in console
 * output (the first c token of its __class__ descriptor). */
static int json_append_class_name(json_builder *b, const char *type_name) {
    if (type_name == NULL || strncmp(type_name, "__class__|", 10) != 0) return 0;
    const char *cursor = type_name + 10;
    char kind;
    const char *value;
    size_t value_length;
    while (next_class_descriptor_token(&cursor, &kind, &value, &value_length) > 0) {
        if (kind != 'c') continue;
        /* Generic specializations and module-qualified names print their
         * source name (Box__number is Box). */
        size_t length = 0;
        while (length < value_length && !(value[length] == '_' && length + 1 < value_length && value[length + 1] == '_')) length++;
        if (jb_append(b, value, length) != 0) return -1;
        return jb_char(b, ' ');
    }
    return 0;
}

/* A null-prototype object's shape is __null_prototype, optionally followed
 * by its |length:name key extensions (see regex_groups_object). Console
 * output tags it with JSON_INSPECT_NULL_PROTOTYPE, which json_inspect_text
 * expands to Node's "[Object: null prototype] " prefix. */
#define JSON_NULL_PROTOTYPE_SHAPE "__null_prototype"
#define JSON_INSPECT_NULL_PROTOTYPE '\x02'

static int json_is_null_prototype(const char *type_name) {
    size_t length = sizeof(JSON_NULL_PROTOTYPE_SHAPE) - 1;
    return type_name != NULL && strncmp(type_name, JSON_NULL_PROTOTYPE_SHAPE, length) == 0 &&
           (type_name[length] == '\0' || type_name[length] == '|');
}

static int json_builder_object(json_builder *b, void *handle) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel) {
        return jb_append(b, "null", 4);
    }
    scriptgo_json_object *obj = (scriptgo_json_object *)handle;
    if (obj->magic == SCRIPTGO_OBJECT_MAGIC && json_object_is_tuple(obj)) {
        json_object_writer writer = {b, handle, obj, 0, 0, 0};
        if (jb_char(b, '[') != 0) return -1;
        object_field_visit((const scriptgo_object *)handle, json_write_tuple_element, &writer);
        if (writer.failed) return -1;
        return jb_char(b, ']');
    }
    if (obj->magic == SCRIPTGO_OBJECT_MAGIC) {
        int null_prototype = json_is_null_prototype(obj->type_name);
        if (b->inspect && json_append_class_name(b, obj->type_name) != 0) return -1;
        if (b->inspect && null_prototype && jb_char(b, JSON_INSPECT_NULL_PROTOTYPE) != 0) return -1;
        if (jb_char(b, '{') != 0) return -1;
        /* The field names come from the shared layout walk, so key lists
         * with dynamic extensions, dictionaries and class layouts agree with
         * property access. A null-prototype object (a match's groups) owns
         * every key it lists, so its undefined fields print. */
        json_object_writer writer = {b, handle, obj, obj->type_name != NULL && strncmp(obj->type_name, "__json__", 8) == 0, 0, 0, null_prototype};
        object_field_visit((const scriptgo_object *)handle, json_write_object_field, &writer);
        if (writer.failed) return -1;
        return jb_char(b, '}');
    }

    void *keys = NULL;
    int64_t key_count = 0;
    if (scriptgo_object_keys(handle, &keys) != 0 || scriptgo_array_length(keys, &key_count) != 0) {
        scriptgo_array_release(keys);
        return -1;
    }
    if (jb_char(b, '{') != 0) { scriptgo_array_release(keys); return -1; }
    int has_fields = 0;
    for (int64_t i = 0; i < key_count; i++) {
        const char *key = NULL;
        scriptgo_value value;
        scriptgo_value_init_undefined(&value);
        if (scriptgo_array_get(keys, (double)i, &key) != 0 || key == NULL ||
            scriptgo_object_property_unknown_get(handle, key, &value) != 0) {
            scriptgo_array_release(keys);
            return -1;
        }
        if (value.tag == SCRIPTGO_TAG_UNDEFINED || value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL) continue;
        if (has_fields && jb_char(b, ',') != 0) { scriptgo_array_release(keys); return -1; }
        if (jb_string(b, key, strlen(key)) != 0 ||
            jb_char(b, ':') != 0 ||
            json_builder_value(b, &value) != 0) {
            scriptgo_array_release(keys);
            return -1;
        }
        has_fields = 1;
    }
    scriptgo_array_release(keys);
    return jb_char(b, '}');
}

/* json_builder_array writes an array in its element layout: boxed values,
 * strings, bools, bigints, numbers, or pointers whose element tag names
 * them (objects, arrays, functions, symbols). */
static int json_builder_array_elements(json_builder *b, const scriptgo_array_internal *arr);

/* json_builder_array_properties appends an array's named properties (a
 * RegExp match's index, input and groups) after its elements, as console.log
 * shows them: `[ 'a', index: 0, input: 'a', groups: undefined ]`. */
static int json_builder_array_properties(json_builder *b, const scriptgo_array_internal *arr) {
    json_builder fields = {0};
    fields.inspect = 1;
    json_object_writer writer = {&fields, arr->properties, (scriptgo_json_object *)arr->properties, 0, 0, 0, 1};
    if (jb_char(&fields, '{') != 0) return -1;
    object_field_visit((const scriptgo_object *)arr->properties, json_write_object_field, &writer);
    if (writer.failed || jb_char(&fields, '}') != 0) {
        free(fields.buf);
        return -1;
    }
    /* fields.buf is "{...}"; splice its members before the closing ']'. */
    int failed = 0;
    if (fields.len > 2 && fields.buf[0] == '{' && fields.buf[fields.len - 1] == '}') {
        b->len--;
        failed = (arr->length > 0 && jb_char(b, ',') != 0) ||
                 jb_append(b, fields.buf + 1, fields.len - 2) != 0 ||
                 jb_char(b, ']') != 0;
    }
    free(fields.buf);
    return failed ? -1 : 0;
}

static int json_builder_array(json_builder *b, const scriptgo_array_internal *arr) {
    if (json_builder_array_elements(b, arr) != 0) return -1;
    if (b->inspect && arr->properties != NULL) return json_builder_array_properties(b, arr);
    return 0;
}

static int json_builder_array_elements(json_builder *b, const scriptgo_array_internal *arr) {
    if (arr->element_tag == SCRIPTGO_TAG_STRING && arr->element_size == sizeof(char *)) {
        return json_builder_string_array(b, arr);
    }
    if (jb_char(b, '[') != 0) return -1;
    for (int64_t i = 0; i < arr->length; i++) {
        if (i > 0 && jb_char(b, ',') != 0) return -1;
        scriptgo_value element;
        memset(&element, 0, sizeof(element));
        if (arr->element_size == sizeof(scriptgo_value)) {
            memcpy(&element, arr->data + (size_t)i * sizeof(scriptgo_value), sizeof(element));
        } else if (arr->element_size == 1) {
            element.tag = SCRIPTGO_TAG_BOOLEAN;
            element.payload = arr->data[i] != 0;
        } else {
            uint64_t bits;
            memcpy(&bits, arr->data + (size_t)i * 8, sizeof(bits));
            element.payload = bits;
            switch (arr->element_tag) {
            case SCRIPTGO_TAG_OBJECT:
            case SCRIPTGO_TAG_ARRAY:
            case SCRIPTGO_TAG_FUNCTION:
            case SCRIPTGO_TAG_SYMBOL:
            case SCRIPTGO_TAG_BIGINT:
                element.tag = (uint32_t)arr->element_tag;
                if ((element.tag == SCRIPTGO_TAG_OBJECT || element.tag == SCRIPTGO_TAG_ARRAY) && bits == 0) element.tag = SCRIPTGO_TAG_NULL;
                if (bits == (uint64_t)(uintptr_t)&scriptgo_undefined_sentinel) element.tag = SCRIPTGO_TAG_UNDEFINED;
                break;
            default:
                element.tag = SCRIPTGO_TAG_NUMBER;
                break;
            }
        }
        if (element.tag == SCRIPTGO_TAG_NUMBER) {
            double number;
            memcpy(&number, &element.payload, sizeof(number));
            const char *marker = scriptgo_number_marker_name(number);
            if (marker != NULL) {
                /* undefined and null held in number storage */
                const char *text = b->inspect ? marker : "null";
                if (jb_append(b, text, strlen(text)) != 0) return -1;
                continue;
            }
        }
        if (json_builder_value(b, &element) != 0) return -1;
    }
    return jb_char(b, ']');
}

static int json_builder_value(json_builder *b, const scriptgo_value *value) {
    if (value == NULL) return jb_append(b, "null", 4);
    uint32_t tag = value->tag;
    uint64_t payload = value->payload;
    switch (tag) {
    case 0: // undefined
        /* console output shows undefined; JSON writes null in arrays. */
        return b->inspect ? jb_append(b, "undefined", 9) : jb_append(b, "null", 4);
    case 1: // null
        return jb_append(b, "null", 4);
    case 2: // boolean
        return jb_append(b, payload ? "true" : "false", payload ? 4 : 5);
    case 3: { // number
        union { uint64_t u64; double d; } u;
        u.u64 = payload;
        return jb_number(b, u.d);
    }
    case 4: { // string
        const char *s = (const char *)(uintptr_t)payload;
        if (s == NULL) return jb_append(b, "null", 4);
        return jb_string(b, s, strlen(s));
    }
    case 6: { // array
        scriptgo_array_internal *arr = (scriptgo_array_internal *)(uintptr_t)payload;
        if (arr == NULL) return jb_append(b, "null", 4);
        return json_builder_array(b, arr);
    }
    case 5: // object
        return json_builder_object(b, (void *)(uintptr_t)payload);
    case 7: // function
        return b->inspect ? jb_append(b, "[Function (anonymous)]", 22) : jb_append(b, "null", 4);
    case 8: // bigint
        if (b->inspect) {
            char text[32];
            int written = snprintf(text, sizeof(text), "%lldn", (long long)payload);
            return jb_append(b, text, (size_t)written);
        }
        break;
    case 9: // symbol
        if (b->inspect) {
            char *text = NULL;
            if (scriptgo_symbol_to_string((void *)(uintptr_t)payload, &text) != 0 || text == NULL) return -1;
            int r = jb_append(b, text, strlen(text));
            free(text);
            return r;
        }
        return jb_append(b, "null", 4);
    default:
        break;
    }
    {
        if (payload == 0) return jb_append(b, "null", 4);
        char *str = NULL;
        if (scriptgo_string_from_object((void *)(uintptr_t)payload, &str) != 0) return -1;
        int r = jb_string(b, str, strlen(str));
        free(str);
        return r;
    }
}

static int json_render_object(void *handle, int inspect, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    b.inspect = inspect;
    if (json_builder_object(&b, handle) != 0) {
        free(b.buf);
        return -1;
    }
    *out_str = b.buf ? b.buf : strdup("{}");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

static int json_stringify_object(void *handle, char **out_str) {
    return json_render_object(handle, 0, out_str);
}

int scriptgo_json_stringify_unknown(const scriptgo_value *value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    if (value == NULL || scriptgo_value_validate(value) != 0) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    if (json_builder_value(&b, value) != 0) {
        free(b.buf);
        return -1;
    }
    *out_str = b.buf ? b.buf : strdup("null");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
}

int scriptgo_object_region_active(void);
void *scriptgo_object_region_alloc(size_t size);
int scriptgo_gc_register_fast(void *ptr, int tag, uint32_t field_count);

typedef struct json_arena_chunk {
    struct json_arena_chunk *next;
    size_t used;
    size_t capacity;
    unsigned char data[] __attribute__((aligned(8)));
} json_arena_chunk;

typedef struct json_arena {
    json_arena_chunk *chunks;
    json_arena_chunk *current;
} json_arena;

static void *json_arena_alloc(json_arena *arena, size_t size) {
    if (arena == NULL) return NULL;
    size = (size + sizeof(void *) - 1) & ~(sizeof(void *) - 1);
    json_arena_chunk *chunk = arena->current;
    if (__builtin_expect(chunk != NULL && chunk->capacity - chunk->used >= size, 1)) {
        void *res = chunk->data + chunk->used;
        chunk->used += size;
        return res;
    }
    size_t capacity = chunk != NULL ? chunk->capacity * 2 : 256 * 1024;
    if (capacity > 4 * 1024 * 1024) capacity = 4 * 1024 * 1024;
    if (capacity < size) capacity = size;
    json_arena_chunk *nc = (json_arena_chunk *)malloc(sizeof(*nc) + capacity);
    if (nc == NULL) return NULL;
    nc->next = NULL;
    nc->used = size;
    nc->capacity = capacity;
    if (chunk != NULL) {
        chunk->next = nc;
    } else {
        arena->chunks = nc;
    }
    arena->current = nc;
    return nc->data;
}

static void json_arena_free(json_arena *arena) {
    if (arena == NULL) return;
    json_arena_chunk *c = arena->chunks;
    while (c != NULL) {
        json_arena_chunk *next = c->next;
        free(c);
        c = next;
    }
    free(arena);
}

typedef struct json_arena_tracker {
    void *root;
    json_arena *arena;
    struct json_arena_tracker *next;
} json_arena_tracker;

static json_arena_tracker *active_json_arenas = NULL;
static int json_cleaner_registered = 0;

static void json_arena_gc_cleaner(void *weak_obj, int (*is_alive)(void *ptr)) {
    (void)weak_obj;
    json_arena_tracker **prev = &active_json_arenas;
    json_arena_tracker *curr = *prev;
    while (curr != NULL) {
        if (!is_alive(curr->root)) {
            *prev = curr->next;
            json_arena_free(curr->arena);
            json_arena_tracker *to_free = curr;
            curr = curr->next;
            free(to_free);
        } else {
            prev = &curr->next;
            curr = curr->next;
        }
    }
}

static void register_json_arena(void *root, json_arena *arena) {
    if (!json_cleaner_registered) {
        void scriptgo_gc_register_weak_cleaner(void (*fn)(void *weak_obj, int (*is_alive)(void *ptr)));
        scriptgo_gc_register_weak_cleaner(json_arena_gc_cleaner);
        json_cleaner_registered = 1;
    }
    json_arena_tracker *t = (json_arena_tracker *)malloc(sizeof(json_arena_tracker));
    if (t == NULL) return;
    t->root = root;
    t->arena = arena;
    t->next = active_json_arenas;
    active_json_arenas = t;
}

int scriptgo_json_arena_contains(const void *ptr) {
    if (active_json_arenas == NULL || ptr == NULL) return 0;
    for (json_arena_tracker *t = active_json_arenas; t != NULL; t = t->next) {
        if (t->arena == NULL) continue;
        for (json_arena_chunk *c = t->arena->chunks; c != NULL; c = c->next) {
            if ((const unsigned char *)ptr >= c->data && (const unsigned char *)ptr < c->data + c->used) {
                return 1;
            }
        }
    }
    return 0;
}

static inline void *json_val_alloc(json_arena *arena, size_t size) {
    if (scriptgo_object_region_active()) {
        return scriptgo_object_region_alloc(size);
    }
    return json_arena_alloc(arena, size);
}

static int convert_yyjson_val_arena(yyjson_val *val, scriptgo_json_unknown *out, json_arena *arena, int is_root) {
    if (val == NULL || out == NULL) return -1;
    out->flags = 0;
    out->aux = 0;
    yyjson_type type = yyjson_get_type(val);
    switch (type) {
    case YYJSON_TYPE_NULL:
        out->tag = 1;
        out->payload = 0;
        return 0;
    case YYJSON_TYPE_BOOL:
        out->tag = 2;
        out->payload = yyjson_get_bool(val) ? 1 : 0;
        return 0;
    case YYJSON_TYPE_NUM: {
        out->tag = 3;
        double d = yyjson_get_num(val);
        memcpy(&out->payload, &d, sizeof(double));
        return 0;
    }
    case YYJSON_TYPE_STR: {
        out->tag = 4;
        size_t len = yyjson_get_len(val);
        const char *s = yyjson_get_str(val);
        char *copy = (char *)json_val_alloc(arena, len + 1);
        if (copy == NULL) return json_fail("scriptgo json allocation failed");
        memcpy(copy, s, len);
        copy[len] = '\0';
        out->payload = (uint64_t)(uintptr_t)copy;
        return 0;
    }
    case YYJSON_TYPE_ARR: {
        size_t count = yyjson_arr_size(val);
        void *array = NULL;
        if (scriptgo_array_new((int64_t)count, (int64_t)sizeof(scriptgo_json_unknown), &array) != 0 ||
            scriptgo_array_set_tag(array, 6) != 0) {
            return -1;
        }
        yyjson_val *elem;
        yyjson_arr_iter iter;
        yyjson_arr_iter_init(val, &iter);
        scriptgo_array_internal *native_array = (scriptgo_array_internal *)array;
        int64_t idx = 0;
        while ((elem = yyjson_arr_iter_next(&iter))) {
            scriptgo_json_unknown elem_val;
            if (convert_yyjson_val_arena(elem, &elem_val, arena, 0) != 0) {
                return -1;
            }
            memcpy(native_array->data + (size_t)idx * sizeof(elem_val), &elem_val, sizeof(elem_val));
            idx++;
        }
        if (is_root && arena != NULL) {
            register_json_arena(native_array, arena);
        }
        out->tag = 6;
        out->payload = (uint64_t)(uintptr_t)array;
        return 0;
    }
    case YYJSON_TYPE_OBJ: {
        size_t count = yyjson_obj_size(val);
        uint32_t capacity = (uint32_t)count;
        if (capacity < 1) capacity = 1;
        scriptgo_json_object *object = (scriptgo_json_object *)json_val_alloc(arena, sizeof(scriptgo_json_object) + (size_t)capacity * sizeof(uintptr_t));
        if (object == NULL) return json_fail("scriptgo json allocation failed");
        object->magic = SCRIPTGO_OBJECT_MAGIC;
        object->field_count = (int64_t)count;
        object->extensible = 1;
        object->sealed = 0;
        object->frozen = 0;
        object->type_name_owned = 0;
        object->capacity = capacity;
        object->boxed_fields = NULL;
        for (uint32_t i = 0; i < capacity; i++) {
            object->fields[i] = (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS;
        }

        char stack_buf[512];
        char *type_name = stack_buf;
        size_t type_name_cap = sizeof(stack_buf);
        size_t type_name_len = 8;
        memcpy(type_name, "__json__", 8);
        type_name[8] = '\0';

        yyjson_val *key;
        yyjson_obj_iter iter;
        yyjson_obj_iter_init(val, &iter);
        int64_t idx = 0;
        while ((key = yyjson_obj_iter_next(&iter))) {
            yyjson_val *child = yyjson_obj_iter_get_val(key);
            const char *key_str = yyjson_get_str(key);
            size_t key_len = yyjson_get_len(key);

            char digits[32];
            int dlen = snprintf(digits, sizeof(digits), "|%zu:", key_len);
            size_t req = type_name_len + (size_t)dlen + key_len + 1;
            if (req > type_name_cap) {
                size_t new_cap = type_name_cap * 2;
                if (new_cap < req + 256) new_cap = req + 256;
                char *grown;
                if (type_name == stack_buf) {
                    grown = (char *)malloc(new_cap);
                    if (grown == NULL) return json_fail("scriptgo json allocation failed");
                    memcpy(grown, stack_buf, type_name_len);
                } else {
                    grown = (char *)realloc(type_name, new_cap);
                    if (grown == NULL) { free(type_name); return json_fail("scriptgo json allocation failed"); }
                }
                type_name = grown;
                type_name_cap = new_cap;
            }
            memcpy(type_name + type_name_len, digits, (size_t)dlen);
            type_name_len += (size_t)dlen;
            memcpy(type_name + type_name_len, key_str, key_len);
            type_name_len += key_len;
            type_name[type_name_len] = '\0';

            scriptgo_json_unknown child_val;
            if (convert_yyjson_val_arena(child, &child_val, arena, 0) != 0) {
                if (type_name != stack_buf) free(type_name);
                return -1;
            }
            json_object_store_value(object, idx++, &child_val);
        }
        object->field_count = idx;
        char *tn_copy = (char *)json_val_alloc(arena, type_name_len + 1);
        if (tn_copy != NULL) {
            memcpy(tn_copy, type_name, type_name_len + 1);
            object->type_name = tn_copy;
        }
        if (type_name != stack_buf) {
            free(type_name);
        }
        if (is_root && arena != NULL) {
            scriptgo_gc_register_fast(object, 1, object->capacity);
            register_json_arena(object, arena);
        }
        out->tag = 5;
        out->payload = (uint64_t)(uintptr_t)object;
        return 0;
    }
    default:
        out->tag = 0;
        out->payload = 0;
        return 0;
    }
}

int scriptgo_json_parse_unknown(const char *input, scriptgo_json_unknown *out_value) {
    if (input == NULL || out_value == NULL) return json_fail("scriptgo json invalid argument");
    const char *p = input;
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    if (*p == '\0') return json_fail("scriptgo json invalid value");

    size_t in_len = strlen(input);
    yyjson_read_err err;
    yyjson_doc *doc = yyjson_read_opts((char *)input, in_len, YYJSON_READ_NOFLAG, NULL, &err);
    if (doc == NULL) {
        return json_fail("scriptgo json invalid value");
    }
    yyjson_val *root = yyjson_doc_get_root(doc);
    if (root == NULL) {
        yyjson_doc_free(doc);
        return json_fail("scriptgo json invalid value");
    }
    json_arena *arena = NULL;
    if (!scriptgo_object_region_active()) {
        arena = (json_arena *)calloc(1, sizeof(json_arena));
    }
    int status = convert_yyjson_val_arena(root, out_value, arena, 1);
    if (status != 0 && arena != NULL) {
        json_arena_free(arena);
    }
    yyjson_doc_free(doc);
    return status;
}

static int inspect_append(char **out, size_t *length, size_t *capacity, char value) {
    if (*length + 2 > *capacity) {
        size_t next = *capacity == 0 ? 64 : *capacity * 2;
        char *grown = realloc(*out, next);
        if (grown == NULL) return json_fail("scriptgo inspect allocation failed");
        *out = grown;
        *capacity = next;
    }
    (*out)[(*length)++] = value;
    (*out)[*length] = '\0';
    return 0;
}

/* json_inspect_text converts JSON text rendered in inspect mode into
 * console.log's format: unquoted identifier keys, single-quoted strings,
 * spaces inside braces and brackets and after commas and colons. */
static int json_inspect_text(char *json, char **out_str) {
    char *out = NULL;
    size_t length = 0, capacity = 0;
    int in_string = 0;
    for (size_t i = 0; json[i] != '\0'; i++) {
        char c = json[i];
        if (c == '"') {
            size_t end = i + 1;
            while (json[end] != '\0') {
                if (json[end] == '"' && json[end - 1] != '\\') break;
                end++;
            }
            int is_key = json[end + 1] == ':';
            if (is_key) {
                size_t start = i + 1;
                if (end - start >= 6 && strncmp(json + start, "\\u0001", 6) == 0) {
                    /* A symbol key: [Symbol(description)]. */
                    if (inspect_append(&out, &length, &capacity, '[') != 0) goto fail;
                    for (size_t j = start + 6; j < end; j++) {
                        if (json[j] == '\\' && json[j + 1] == '"') j++;
                        if (inspect_append(&out, &length, &capacity, json[j]) != 0) goto fail;
                    }
                    if (inspect_append(&out, &length, &capacity, ']') != 0) goto fail;
                    i = end;
                    continue;
                }
                /* Node quotes keys that are not identifiers, numeric keys included. */
                int identifier = start < end && !(json[start] >= '0' && json[start] <= '9');
                for (size_t j = start; j < end; j++) {
                    char k = json[j];
                    if (!((k >= 'a' && k <= 'z') || (k >= 'A' && k <= 'Z') ||
                          (k >= '0' && k <= '9') || k == '_' || k == '$')) {
                        identifier = 0;
                        break;
                    }
                }
                if (identifier) {
                    for (size_t j = start; j < end; j++) {
                        if (inspect_append(&out, &length, &capacity, json[j]) != 0) goto fail;
                    }
                } else {
                    if (inspect_append(&out, &length, &capacity, '\'') != 0) goto fail;
                    for (size_t j = start; j < end; j++) {
                        if (inspect_append(&out, &length, &capacity, json[j]) != 0) goto fail;
                    }
                    if (inspect_append(&out, &length, &capacity, '\'') != 0) goto fail;
                }
            } else {
                if (inspect_append(&out, &length, &capacity, '\'') != 0) goto fail;
                for (size_t j = i + 1; j < end; j++) {
                    if (json[j] == '\\' && json[j + 1] == '"') j++;
                    if (inspect_append(&out, &length, &capacity, json[j]) != 0) goto fail;
                }
                if (inspect_append(&out, &length, &capacity, '\'') != 0) goto fail;
            }
            i = end;
            in_string = 0;
            continue;
        }
        if (c == JSON_INSPECT_NULL_PROTOTYPE) {
            static const char prefix[] = "[Object: null prototype] ";
            for (size_t j = 0; j + 1 < sizeof(prefix); j++) {
                if (inspect_append(&out, &length, &capacity, prefix[j]) != 0) goto fail;
            }
            continue;
        }
        if (!in_string && c == '{' && json[i + 1] == '}') {
            if (inspect_append(&out, &length, &capacity, '{') != 0 || inspect_append(&out, &length, &capacity, '}') != 0) goto fail;
            i++;
        } else if (!in_string && c == '{') {
            if (inspect_append(&out, &length, &capacity, '{') != 0 || inspect_append(&out, &length, &capacity, ' ') != 0) goto fail;
        } else if (!in_string && c == '}') {
            if (inspect_append(&out, &length, &capacity, ' ') != 0 || inspect_append(&out, &length, &capacity, '}') != 0) goto fail;
        } else if (!in_string && c == '[' && json[i + 1] != ']') {
            if (inspect_append(&out, &length, &capacity, '[') != 0 || inspect_append(&out, &length, &capacity, ' ') != 0) goto fail;
        } else if (!in_string && c == ']' && i > 0 && json[i - 1] != '[') {
            if (inspect_append(&out, &length, &capacity, ' ') != 0 || inspect_append(&out, &length, &capacity, ']') != 0) goto fail;
        } else if (!in_string && c == ',') {
            if (inspect_append(&out, &length, &capacity, ',') != 0 || inspect_append(&out, &length, &capacity, ' ') != 0) goto fail;
        } else if (!in_string && c == ':') {
            if (inspect_append(&out, &length, &capacity, ':') != 0 || inspect_append(&out, &length, &capacity, ' ') != 0) goto fail;
        } else if (inspect_append(&out, &length, &capacity, c) != 0) {
            goto fail;
        }
    }
    free(json);
    *out_str = out;
    return 0;
fail:
    free(json);
    free(out);
    return -1;
}

int scriptgo_json_inspect_object(void *handle, char **out_str) {
    char *json = NULL;
    if (out_str == NULL || json_render_object(handle, 1, &json) != 0) return -1;
    return json_inspect_text(json, out_str);
}

/* scriptgo_json_inspect_array renders an array as console.log shows it. */
int scriptgo_json_inspect_array(void *handle, char **out_str) {
    if (out_str == NULL || handle == NULL) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    b.inspect = 1;
    if (json_builder_array(&b, (const scriptgo_array_internal *)handle) != 0) {
        free(b.buf);
        return -1;
    }
    return json_inspect_text(b.buf, out_str);
}
