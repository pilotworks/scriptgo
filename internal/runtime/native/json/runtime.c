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
        return jb_append(b, "null", 4);
    }
    if (jb_reserve(b, 32) != 0) return -1;
    int written;
    if (value == (double)(int64_t)value && fabs(value) < 1e15) {
        written = snprintf(b->buf + b->len, 32, "%lld", (long long)value);
    } else {
        written = snprintf(b->buf + b->len, 32, "%g", value);
    }
    if (written > 0) {
        b->len += (size_t)written;
        b->buf[b->len] = '\0';
    }
    return 0;
}

int scriptgo_json_stringify_number(double value, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
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

// Parsed JSON objects store plain runtime values, so avoid the generic accessor's
// cloning and validation path while preserving it for objects with boxed fields.
static inline void json_object_field_value(const scriptgo_json_object *obj, int64_t index,
                                           scriptgo_value *out_value) {
    memset(out_value, 0, sizeof(*out_value));
    if (index < 0 || index >= obj->field_count) return;

    uintptr_t value = obj->fields[index];
    if (value == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) return;
    if (value == 0) {
        out_value->tag = SCRIPTGO_TAG_NULL;
    } else if (((uint64_t)value >> 32) == 2) {
        out_value->tag = SCRIPTGO_TAG_BOOLEAN;
        out_value->payload = value & 1;
    } else if ((value & 0xFFF8000000000000ULL) != 0) {
        out_value->tag = SCRIPTGO_TAG_NUMBER;
        out_value->payload = (uint64_t)value;
    } else {
        int gc_tag = scriptgo_gc_get_tag((void *)value);
        out_value->tag = gc_tag == 2 ? SCRIPTGO_TAG_ARRAY :
                         gc_tag == 3 ? SCRIPTGO_TAG_FUNCTION :
                         gc_tag == 11 ? SCRIPTGO_TAG_SYMBOL :
                         gc_tag != 0 ? SCRIPTGO_TAG_OBJECT : SCRIPTGO_TAG_STRING;
        out_value->payload = (uint64_t)value;
    }
}

static inline void json_object_store_value(scriptgo_json_object *obj, int64_t index,
                                           const scriptgo_value *value) {
    if (value->tag == SCRIPTGO_TAG_BOOLEAN) {
        obj->fields[index] = (uintptr_t)((2ULL << 32) | (value->payload != 0));
    } else if (value->tag == SCRIPTGO_TAG_NULL) {
        obj->fields[index] = 0;
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

static int json_builder_object(json_builder *b, void *handle) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel) {
        return jb_append(b, "null", 4);
    }
    scriptgo_json_object *obj = (scriptgo_json_object *)handle;
    if (obj->magic == SCRIPTGO_OBJECT_MAGIC) {
        if (jb_char(b, '{') != 0) return -1;
        int has_fields = 0;
        const char *type_name = obj->type_name;
        if (type_name != NULL && strncmp(type_name, "__json__", 8) == 0) {
            const char *cursor = type_name + 8;
            int64_t field_idx = 0;
            while (*cursor == '|' && field_idx < obj->field_count) {
                cursor++; // skip '|'
                size_t key_len = 0;
                while (*cursor >= '0' && *cursor <= '9') {
                    key_len = key_len * 10 + (size_t)(*cursor - '0');
                    cursor++;
                }
                if (*cursor != ':') break;
                cursor++; // skip ':'
                const char *key_str = cursor;
                cursor += key_len;

                scriptgo_value value;
                if (obj->boxed_fields == NULL) {
                    json_object_field_value(obj, field_idx++, &value);
                } else {
                    scriptgo_value_init_undefined(&value);
                    scriptgo_object_unknown_get(handle, field_idx++, &value);
                }
                if (value.tag == SCRIPTGO_TAG_UNDEFINED || value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL) continue;
                if (has_fields && jb_char(b, ',') != 0) return -1;
                if (jb_string(b, key_str, key_len) != 0 ||
                    jb_char(b, ':') != 0 ||
                    json_builder_value(b, &value) != 0) return -1;
                has_fields = 1;
            }
        } else if (type_name != NULL && type_name[0] == ':') {
            const char *cursor = type_name;
            int64_t field_idx = 0;
            while (*cursor != '\0' && field_idx < obj->field_count) {
                if (*cursor == ':') cursor++;
                if (*cursor == '\0') break;
                const char *field_start = cursor;
                const char *field_end = strchr(field_start, ':');
                if (field_end == NULL) break;
                size_t key_len = (size_t)(field_end - field_start);
                cursor = field_end + 1;
                if (key_len == 0) continue;

                scriptgo_value value;
                scriptgo_value_init_undefined(&value);
                scriptgo_object_unknown_get(handle, field_idx++, &value);
                if (value.tag == SCRIPTGO_TAG_UNDEFINED || value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL) continue;
                if (has_fields && jb_char(b, ',') != 0) return -1;
                if (jb_string(b, field_start, key_len) != 0 ||
                    jb_char(b, ':') != 0 ||
                    json_builder_value(b, &value) != 0) return -1;
                has_fields = 1;
            }
        } else if (type_name != NULL && strncmp(type_name, "__class__|", 10) == 0) {
            const char *cursor = type_name + 10;
            int64_t field_idx = 0;
            while (*cursor != '\0' && field_idx < obj->field_count) {
                char kind = *cursor++;
                if (kind != 'c' && kind != 'b' && kind != 'f') break;
                size_t key_len = 0;
                while (*cursor >= '0' && *cursor <= '9') {
                    key_len = key_len * 10 + (size_t)(*cursor - '0');
                    cursor++;
                }
                if (*cursor != ':') break;
                cursor++;
                const char *key_str = cursor;
                cursor += key_len;
                if (*cursor == '|') cursor++;
                if (kind != 'f') continue;

                scriptgo_value value;
                scriptgo_value_init_undefined(&value);
                scriptgo_object_unknown_get(handle, field_idx++, &value);
                if (value.tag == SCRIPTGO_TAG_UNDEFINED || value.tag == SCRIPTGO_TAG_FUNCTION || value.tag == SCRIPTGO_TAG_SYMBOL) continue;
                if (has_fields && jb_char(b, ',') != 0) return -1;
                if (jb_string(b, key_str, key_len) != 0 ||
                    jb_char(b, ':') != 0 ||
                    json_builder_value(b, &value) != 0) return -1;
                has_fields = 1;
            }
        }
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

static int json_builder_value(json_builder *b, const scriptgo_value *value) {
    if (value == NULL) return jb_append(b, "null", 4);
    uint32_t tag = value->tag;
    uint64_t payload = value->payload;
    switch (tag) {
    case 0: // undefined
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
        if (arr->element_tag == 4) {
            return json_builder_string_array(b, arr);
        } else if (arr->element_size == sizeof(scriptgo_value)) {
            if (jb_char(b, '[') != 0) return -1;
            for (int64_t i = 0; i < arr->length; i++) {
                if (i > 0 && jb_char(b, ',') != 0) return -1;
                scriptgo_value *elem = (scriptgo_value *)(arr->data + (size_t)i * sizeof(scriptgo_value));
                if (json_builder_value(b, elem) != 0) return -1;
            }
            return jb_char(b, ']');
        } else {
            return json_builder_number_array(b, arr);
        }
    }
    case 5: // object
        return json_builder_object(b, (void *)(uintptr_t)payload);
    default:
        if (payload == 0) return jb_append(b, "null", 4);
        char *str = NULL;
        if (scriptgo_string_from_object((void *)(uintptr_t)payload, &str) != 0) return -1;
        int r = jb_string(b, str, strlen(str));
        free(str);
        return r;
    }
}

static int json_stringify_object(void *handle, char **out_str) {
    if (out_str == NULL) return json_fail("scriptgo json invalid argument");
    json_builder b = {0};
    if (json_builder_object(&b, handle) != 0) {
        free(b.buf);
        return -1;
    }
    *out_str = b.buf ? b.buf : strdup("{}");
    return *out_str == NULL ? json_fail("scriptgo json allocation failed") : 0;
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

static int convert_yyjson_val(yyjson_val *val, scriptgo_json_unknown *out) {
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
        char *copy = (char *)malloc(len + 1);
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
            if (convert_yyjson_val(elem, &elem_val) != 0) {
                return -1;
            }
            memcpy(native_array->data + (size_t)idx * sizeof(elem_val), &elem_val, sizeof(elem_val));
            idx++;
        }
        out->tag = 6;
        out->payload = (uint64_t)(uintptr_t)array;
        return 0;
    }
    case YYJSON_TYPE_OBJ: {
        size_t count = yyjson_obj_size(val);
        void *object = NULL;
        if (scriptgo_object_new((int64_t)count, &object) != 0) return -1;
        
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
            if (convert_yyjson_val(child, &child_val) != 0) {
                if (type_name != stack_buf) free(type_name);
                return -1;
            }
            json_object_store_value((scriptgo_json_object *)object, idx++, &child_val);
        }
        ((scriptgo_json_object *)object)->field_count = idx;
        if (scriptgo_object_type_set(object, type_name) != 0) {
            if (type_name != stack_buf) free(type_name);
            return -1;
        }
        if (type_name != stack_buf) {
            free(type_name);
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
    int status = convert_yyjson_val(root, out_value);
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

int scriptgo_json_inspect_object(void *handle, char **out_str) {
    char *json = NULL;
    char *out = NULL;
    size_t length = 0, capacity = 0;
    int in_string = 0;
    int empty = 0;
    if (out_str == NULL || json_stringify_object(handle, &json) != 0) return -1;
    empty = strcmp(json, "{}") == 0;
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
                int identifier = start < end;
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
        if (!in_string && c == '{' && !empty) {
            if (inspect_append(&out, &length, &capacity, '{') != 0 || inspect_append(&out, &length, &capacity, ' ') != 0) goto fail;
        } else if (!in_string && c == '}' && !empty) {
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
