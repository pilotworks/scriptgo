#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL
#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252U
#define SCRIPTGO_MAGIC_DATAVIEW 0x44564957U
#define SCRIPTGO_MAGIC_BUFFER 0x42554646U
#define SCRIPTGO_ARRAYBUFFER_GC_TAG 10

extern int scriptgo_gc_get_tag(void *ptr);
extern int scriptgo_gc_is_registered(void *ptr);

extern const char scriptgo_undefined_sentinel;

static inline int is_invalid_object_handle(void *handle) {
    return handle == NULL || handle == (void *)&scriptgo_undefined_sentinel;
}

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    uintptr_t fields[];
} scriptgo_object;

typedef struct {
    uint32_t magic;
    void *buffer;
    int64_t byte_offset;
    int64_t byte_length;
} scriptgo_object_data_view;

typedef struct {
    uint32_t magic;
    uint32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_typed_array_view;

#define SCRIPTGO_OBJECT_TAG_STRING 4U
#define SCRIPTGO_OBJECT_TAG_OBJECT 5U
#define SCRIPTGO_OBJECT_TAG_ARRAY  6U

int scriptgo_runtime_set_error(const char *message);

static int object_fail(const char *message) { return scriptgo_runtime_set_error(message); }

static int is_class_descriptor(const char *type_name) {
    return type_name != NULL && strncmp(type_name, "__class__|", 10) == 0;
}

// Reads one length-prefixed class descriptor token without allocating. The
// descriptor is emitted by lowering and keeps hierarchy names separate from
// the runtime field order.
static int next_class_descriptor_token(const char **cursor, char *kind,
                                       const char **value, size_t *value_length) {
    const char *p;
    size_t length = 0;

    if (cursor == NULL || *cursor == NULL || kind == NULL || value == NULL || value_length == NULL) {
        return -1;
    }
    p = *cursor;
    if (*p == '\0') return 0;
    *kind = *p++;
    if (*kind != 'c' && *kind != 'b' && *kind != 'f') return -1;
    if (*p < '0' || *p > '9') return -1;
    while (*p >= '0' && *p <= '9') {
        size_t digit = (size_t)(*p - '0');
        if (length > (SIZE_MAX - digit) / 10) return -1;
        length = length * 10 + digit;
        p++;
    }
    if (*p++ != ':') return -1;
    *value = p;
    *value_length = length;
    p += length;
    if (*p == '|') {
        p++;
    } else if (*p != '\0') {
        return -1;
    }
    *cursor = p;
    return 1;
}

static int object_field_index(const scriptgo_object *object, const char *property) {
    const char *cursor;
    int index = 0;
    size_t property_len;

    if (object == NULL || object->type_name == NULL || property == NULL) {
        return -1;
    }
    property_len = strlen(property);

    if (is_class_descriptor(object->type_name)) {
        cursor = object->type_name + 10;
        for (;;) {
            char kind;
            const char *value;
            size_t value_length;
            int token = next_class_descriptor_token(&cursor, &kind, &value, &value_length);
            if (token <= 0) return -1;
            if (kind == 'f') {
                if (value_length == property_len && memcmp(value, property, value_length) == 0) {
                    return index;
                }
                index++;
            }
        }
    }

    if (strncmp(object->type_name, "__json__", 8) == 0) {
        const char *encoded = object->type_name + 8;
        int index = 0;
        while (*encoded != '\0') {
            size_t key_length = 0;
            if (*encoded++ != '|') return -1;
            if (*encoded < '0' || *encoded > '9') return -1;
            while (*encoded >= '0' && *encoded <= '9') {
                key_length = key_length * 10 + (size_t)(*encoded - '0');
                encoded++;
            }
            if (*encoded++ != ':') return -1;
            if (key_length == property_len && strncmp(encoded, property, key_length) == 0) return index;
            encoded += key_length;
            index++;
        }
        return -1;
    }

    cursor = object->type_name;
    /* Plain shape markers carry no fields; only their encoded extensions do. */
    {
        const char *extension = strchr(cursor, '|');
        const char *descriptor_colon = strchr(cursor, ':');
        if (extension != NULL && (descriptor_colon == NULL || descriptor_colon > extension)) {
            cursor = extension;
        } else if (extension == NULL && descriptor_colon == NULL) {
            return -1;
        }
    }
    while (*cursor != '\0') {
        const char *field_start;
        const char *field_end;
        size_t field_len;

        if (*cursor == '|') {
            size_t encoded_len = 0;
            cursor++;
            if (*cursor < '0' || *cursor > '9') return -1;
            while (*cursor >= '0' && *cursor <= '9') {
                encoded_len = encoded_len * 10 + (size_t)(*cursor - '0');
                cursor++;
            }
            if (*cursor++ != ':') return -1;
            if (encoded_len == property_len &&
                memcmp(cursor, property, encoded_len) == 0) {
                return index;
            }
            cursor += encoded_len;
            index++;
            continue;
        }

        if (*cursor == ':') {
            cursor++;
        }
        field_start = cursor;
        field_end = strchr(field_start, ':');
        if (field_end == NULL) {
            field_end = field_start + strlen(field_start);
        }
        field_len = (size_t)(field_end - field_start);
        if (field_len == property_len && strncmp(field_start, property, field_len) == 0) {
            return index;
        }
        if (*field_end == '\0') {
            break;
        }
        cursor = field_end + 1;
        index++;
    }
    return -1;
}

int scriptgo_unknown_number_property(uint32_t tag, uint64_t payload, const char *property, double *out_value) {
    if (property == NULL || out_value == NULL) {
        return object_fail("scriptgo unknown property invalid arguments");
    }
    *out_value = NAN;
    if (tag == SCRIPTGO_OBJECT_TAG_STRING) {
        if (strcmp(property, "length") == 0) {
            const char *value = (const char *)(uintptr_t)payload;
            *out_value = value == NULL ? 0.0 : (double)strlen(value);
        }
        return 0;
    }
    if ((tag != SCRIPTGO_OBJECT_TAG_OBJECT && tag != SCRIPTGO_OBJECT_TAG_ARRAY) || payload == 0) {
        return 0;
    }
    void *handle = (void *)(uintptr_t)payload;
    if (!scriptgo_gc_is_registered(handle)) {
        return 0;
    }
    uint32_t magic = *(uint32_t *)handle;
    if (magic == SCRIPTGO_MAGIC_TYPEDARRAY) {
        scriptgo_typed_array_view *view = (scriptgo_typed_array_view *)handle;
        if (strcmp(property, "length") == 0) {
            *out_value = (double)view->length;
        } else if (strcmp(property, "byteLength") == 0) {
            *out_value = (double)(view->length * view->element_size);
        } else if (strcmp(property, "byteOffset") == 0) {
            *out_value = (double)view->byte_offset;
        } else if (strcmp(property, "BYTES_PER_ELEMENT") == 0) {
            *out_value = (double)view->element_size;
        }
    } else if (magic == SCRIPTGO_MAGIC_DATAVIEW) {
        scriptgo_object_data_view *view = (scriptgo_object_data_view *)handle;
        if (strcmp(property, "byteLength") == 0) {
            *out_value = (double)view->byte_length;
        } else if (strcmp(property, "byteOffset") == 0) {
            *out_value = (double)view->byte_offset;
        }
    } else if (*(uint64_t *)handle == SCRIPTGO_OBJECT_MAGIC) {
        scriptgo_object *object = (scriptgo_object *)handle;
        int index = object_field_index(object, property);
        if (index >= 0 && index < object->field_count) {
            memcpy(out_value, &object->fields[index], sizeof(*out_value));
        }
    }
    return 0;
}

int scriptgo_object_is_number(double a, double b, int32_t *out_result) {
    if (out_result == NULL) {
        return object_fail("scriptgo Object.is null output");
    }
    if (isnan(a) && isnan(b)) {
        *out_result = 1;
    } else if (a == 0.0 && b == 0.0) {
        *out_result = (signbit(a) == signbit(b)) ? 1 : 0;
    } else {
        *out_result = (a == b) ? 1 : 0;
    }
    return 0;
}

int scriptgo_object_is_string(const char *a, const char *b, int32_t *out_result) {
    if (out_result == NULL) {
        return object_fail("scriptgo Object.is null output");
    }
    if (a == b) {
        *out_result = 1;
    } else if (a == NULL || b == NULL) {
        *out_result = 0;
    } else {
        *out_result = (strcmp(a, b) == 0) ? 1 : 0;
    }
    return 0;
}

int scriptgo_object_is_ptr(void *a, void *b, int32_t *out_result) {
    if (out_result == NULL) {
        return object_fail("scriptgo Object.is null output");
    }
    if (a == b) {
        *out_result = 1;
        return 0;
    }
    if (is_invalid_object_handle(a) || is_invalid_object_handle(b)) {
        *out_result = 0;
        return 0;
    }
    scriptgo_object *oa = (scriptgo_object *)a;
    scriptgo_object *ob = (scriptgo_object *)b;
    if (oa->magic == SCRIPTGO_OBJECT_MAGIC && ob->magic == SCRIPTGO_OBJECT_MAGIC) {
        if (oa->type_name != NULL && ob->type_name != NULL && strcmp(oa->type_name, ob->type_name) != 0) {
            *out_result = 0;
            return 0;
        }
        if (oa->field_count != ob->field_count) {
            *out_result = 0;
            return 0;
        }
        int64_t count = oa->field_count;
        for (int64_t i = 0; i < count; i++) {
            if (oa->fields[i] != ob->fields[i]) {
                *out_result = 0;
                return 0;
            }
        }
        *out_result = 1;
        return 0;
    }
    *out_result = (a == b) ? 1 : 0;
    return 0;
}

int scriptgo_object_is_unknown(uint32_t tag0, uint64_t payload0, uint32_t tag1, uint64_t payload1, int32_t *out_result) {
    if (out_result == NULL) return object_fail("scriptgo Object.is null output");
    if (tag0 != tag1) {
        *out_result = 0;
        return 0;
    }
    if (tag0 == SCRIPTGO_TAG_UNDEFINED || tag0 == SCRIPTGO_TAG_NULL) {
        *out_result = 1;
        return 0;
    }
    if (tag0 == SCRIPTGO_TAG_BOOLEAN) {
        *out_result = (payload0 == payload1) ? 1 : 0;
        return 0;
    }
    if (tag0 == SCRIPTGO_TAG_NUMBER) {
        union { uint64_t u; double d; } u0, u1;
        u0.u = payload0;
        u1.u = payload1;
        if (isnan(u0.d) && isnan(u1.d)) {
            *out_result = 1;
            return 0;
        }
        if (u0.d == 0.0 && u1.d == 0.0) {
            *out_result = (1.0 / u0.d == 1.0 / u1.d) ? 1 : 0;
            return 0;
        }
        *out_result = (u0.d == u1.d) ? 1 : 0;
        return 0;
    }
    if (tag0 == SCRIPTGO_TAG_STRING) {
        const char *s0 = (const char *)(uintptr_t)payload0;
        const char *s1 = (const char *)(uintptr_t)payload1;
        if (s0 == s1) { *out_result = 1; return 0; }
        if (s0 == NULL || s1 == NULL) { *out_result = 0; return 0; }
        *out_result = (strcmp(s0, s1) == 0) ? 1 : 0;
        return 0;
    }
    if (payload0 == payload1) {
        *out_result = 1;
        return 0;
    }
    if (payload0 == 0 || payload1 == 0) {
        *out_result = 0;
        return 0;
    }
    if (scriptgo_gc_is_registered((void *)(uintptr_t)payload0) &&
        scriptgo_gc_is_registered((void *)(uintptr_t)payload1)) {
        scriptgo_object *oa = (scriptgo_object *)(uintptr_t)payload0;
        scriptgo_object *ob = (scriptgo_object *)(uintptr_t)payload1;
        if (oa->magic == SCRIPTGO_OBJECT_MAGIC && ob->magic == SCRIPTGO_OBJECT_MAGIC) {
            if (oa->type_name != NULL && ob->type_name != NULL && strcmp(oa->type_name, ob->type_name) != 0) {
                *out_result = 0;
                return 0;
            }
            if (oa->field_count != ob->field_count) {
                *out_result = 0;
                return 0;
            }
            int64_t count = oa->field_count;
            for (int64_t i = 0; i < count; i++) {
                if (oa->fields[i] != ob->fields[i]) {
                    *out_result = 0;
                    return 0;
                }
            }
            *out_result = 1;
            return 0;
        }
    }
    *out_result = (payload0 == payload1) ? 1 : 0;
    return 0;
}

int scriptgo_object_equals_unknown(uint32_t tag0, uint64_t payload0,
                                   uint32_t tag1, uint64_t payload1,
                                   int32_t loose, int32_t *out_result) {
    if (out_result == NULL) return object_fail("scriptgo unknown equality null output");
    if (tag0 != tag1) {
        *out_result = (loose &&
                       ((tag0 == SCRIPTGO_TAG_UNDEFINED || tag0 == SCRIPTGO_TAG_NULL) &&
                        (tag1 == SCRIPTGO_TAG_UNDEFINED || tag1 == SCRIPTGO_TAG_NULL))) ? 1 : 0;
        return 0;
    }
    switch (tag0) {
    case SCRIPTGO_TAG_UNDEFINED:
    case SCRIPTGO_TAG_NULL:
        *out_result = 1;
        return 0;
    case SCRIPTGO_TAG_BOOLEAN:
        *out_result = payload0 == payload1 ? 1 : 0;
        return 0;
    case SCRIPTGO_TAG_NUMBER: {
        union { uint64_t u; double d; } u0, u1;
        u0.u = payload0;
        u1.u = payload1;
        *out_result = (!isnan(u0.d) && !isnan(u1.d) && u0.d == u1.d) ? 1 : 0;
        return 0;
    }
    case SCRIPTGO_TAG_STRING: {
        const char *s0 = (const char *)(uintptr_t)payload0;
        const char *s1 = (const char *)(uintptr_t)payload1;
        *out_result = (s0 != NULL && s1 != NULL && strcmp(s0, s1) == 0) ? 1 : (s0 == s1 ? 1 : 0);
        return 0;
    }
    default:
        *out_result = payload0 == payload1 ? 1 : 0;
        return 0;
    }
}

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count);
int scriptgo_gc_is_registered(void *ptr);
int scriptgo_gc_unregister(void *ptr);

#define SCRIPTGO_OBJECT_NAN_BITS 0x7FF8000000000000ULL

int scriptgo_object_new(int64_t field_count, void **out_object) {
    if (out_object == NULL || field_count < 0) {
        return object_fail("scriptgo object allocation failed");
    }
    int64_t capacity = field_count < 64 ? 64 : field_count;
    scriptgo_object *object = malloc(sizeof(*object) + (size_t)capacity * sizeof(object->fields[0]));
    if (object == NULL) {
        return object_fail("scriptgo object allocation failed");
    }
    object->magic = SCRIPTGO_OBJECT_MAGIC;
    object->field_count = field_count;
    object->type_name = NULL;
    object->extensible = 1;
    object->sealed = 0;
    object->frozen = 0;
    for (int64_t i = 0; i < capacity; i++) {
        uint64_t nan_bits = SCRIPTGO_OBJECT_NAN_BITS;
        memcpy(&object->fields[i], &nan_bits, sizeof(uint64_t));
    }
    scriptgo_gc_register(object, 1, (uint32_t)capacity);
    *out_object = object;
    return 0;
}

int scriptgo_object_freeze(void *handle, void **out_object) {
    if (out_object == NULL) return object_fail("scriptgo object freeze failed");
    if (is_invalid_object_handle(handle) || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) {
        return object_fail("scriptgo object freeze invalid object");
    }
    scriptgo_object *object = (scriptgo_object *)handle;
    object->extensible = 0;
    object->sealed = 1;
    object->frozen = 1;
    *out_object = handle;
    return 0;
}

int scriptgo_object_seal(void *handle, void **out_object) {
    if (out_object == NULL) return object_fail("scriptgo object seal failed");
    if (is_invalid_object_handle(handle) || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) {
        return object_fail("scriptgo object seal invalid object");
    }
    scriptgo_object *object = (scriptgo_object *)handle;
    object->extensible = 0;
    object->sealed = 1;
    *out_object = handle;
    return 0;
}

int scriptgo_object_prevent_extensions(void *handle, void **out_object) {
    if (out_object == NULL) return object_fail("scriptgo object preventExtensions failed");
    if (is_invalid_object_handle(handle) || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) {
        return object_fail("scriptgo object preventExtensions invalid object");
    }
    ((scriptgo_object *)handle)->extensible = 0;
    *out_object = handle;
    return 0;
}

int scriptgo_object_is_frozen(void *handle, int32_t *out_result) {
    if (out_result == NULL) return object_fail("scriptgo object isFrozen failed");
    *out_result = !is_invalid_object_handle(handle) && ((scriptgo_object *)handle)->magic == SCRIPTGO_OBJECT_MAGIC && ((scriptgo_object *)handle)->frozen;
    return 0;
}

int scriptgo_object_is_sealed(void *handle, int32_t *out_result) {
    if (out_result == NULL) return object_fail("scriptgo object isSealed failed");
    *out_result = !is_invalid_object_handle(handle) && ((scriptgo_object *)handle)->magic == SCRIPTGO_OBJECT_MAGIC && ((scriptgo_object *)handle)->sealed;
    return 0;
}

int scriptgo_object_is_extensible(void *handle, int32_t *out_result) {
    if (out_result == NULL) return object_fail("scriptgo object isExtensible failed");
    *out_result = !is_invalid_object_handle(handle) && ((scriptgo_object *)handle)->magic == SCRIPTGO_OBJECT_MAGIC && ((scriptgo_object *)handle)->extensible;
    return 0;
}

int scriptgo_object_number_set(void *handle, int64_t index, double value) {
	if (is_invalid_object_handle(handle) || !scriptgo_gc_is_registered(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    memcpy(&o->fields[index], &value, sizeof(value));
    return 0;
}

int scriptgo_object_number_get(void *handle, int64_t index, double *out_value) {
    if (out_value == NULL) {
        return 0;
    }
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        *out_value = NAN;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = NAN;
        return 0;
    }
    memcpy(out_value, &o->fields[index], sizeof(*out_value));
    return 0;
}

int scriptgo_object_string_set(void *handle, int64_t index, const char *value) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    o->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_string_get(void *handle, int64_t index, const char **out_value) {
    if (out_value == NULL) {
        return 0;
    }
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        *out_value = &scriptgo_undefined_sentinel;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = &scriptgo_undefined_sentinel;
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS || val == 0) {
        *out_value = &scriptgo_undefined_sentinel;
    } else {
        *out_value = (const char *)val;
    }
    return 0;
}

int scriptgo_object_bool_set(void *handle, int64_t index, int32_t value) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    o->fields[index] = (uintptr_t)((2ULL << 32) | (value != 0 ? 1 : 0));
    return 0;
}

int scriptgo_object_bool_get(void *handle, int64_t index, int32_t *out_value) {
    if (out_value == NULL) {
        return 0;
    }
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        *out_value = 0;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = 0;
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS || val == 0) {
        *out_value = 0;
    } else if ((val >> 32) == 2) {
        *out_value = (int32_t)(val & 1);
    } else {
        *out_value = (int32_t)(val != 0 ? 1 : 0);
    }
    return 0;
}

int scriptgo_object_bigint_set(void *handle, int64_t index, int64_t value) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    o->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_bigint_get(void *handle, int64_t index, int64_t *out_value) {
    if (out_value == NULL) {
        return 0;
    }
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        *out_value = 0;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = 0;
        return 0;
    }
    *out_value = (int64_t)o->fields[index];
    return 0;
}

int scriptgo_object_ptr_set(void *handle, int64_t index, void *value) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    o->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_ptr_get(void *handle, int64_t index, void **out_value) {
    if (out_value == NULL) {
        return 0;
    }
	if (is_invalid_object_handle(handle) || !scriptgo_gc_is_registered(handle) || index < 0 || index >= 64) {
		*out_value = (void *)&scriptgo_undefined_sentinel;
		return 0;
	}
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = (void *)&scriptgo_undefined_sentinel;
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) {
        *out_value = (void *)&scriptgo_undefined_sentinel;
    } else {
        *out_value = (void *)val;
    }
    return 0;
}

int scriptgo_object_unknown_set(void *handle, int64_t index, uint32_t tag, uint64_t payload) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    if (tag == 2) {
        o->fields[index] = (uintptr_t)((2ULL << 32) | (payload != 0 ? 1 : 0));
    } else if (tag == 1) {
        o->fields[index] = 0;
    } else if (tag == 0) {
        o->fields[index] = (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS;
    } else {
        o->fields[index] = (uintptr_t)payload;
    }
    return 0;
}

int scriptgo_object_unknown_get(void *handle, int64_t index, uint32_t *out_tag, uint64_t *out_payload) {
    if (out_tag == NULL || out_payload == NULL) {
        return 0;
    }
    *out_tag = 0;
    *out_payload = 0;
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) {
        *out_tag = 0;
        *out_payload = 0;
    } else if (val == 0) {
        *out_tag = 1;
        *out_payload = 0;
    } else if ((val >> 32) == 2) {
        *out_tag = 2;
        *out_payload = (val & 1);
    } else if ((val & 0xFFF8000000000000ULL) != 0) {
        *out_tag = 3;
        *out_payload = (uint64_t)val;
    } else {
        int gc_tag = scriptgo_gc_get_tag((void *)val);
        if (gc_tag == 2) {
            *out_tag = 6;
        } else if (gc_tag == 3) {
            *out_tag = 7;
        } else if (gc_tag == 11) {
            *out_tag = 9; // SCRIPTGO_TAG_SYMBOL
        } else if (gc_tag != 0) {
            *out_tag = 5;
        } else {
            *out_tag = 4;
        }
        *out_payload = (uint64_t)val;
    }
    return 0;
}

int scriptgo_object_property_unknown_get(void *handle, const char *property,
                                         uint32_t *out_tag, uint64_t *out_payload) {
    int index;
    if (out_tag == NULL || out_payload == NULL) {
        return object_fail("scriptgo object property output is invalid");
    }
    *out_tag = 0;
    *out_payload = 0;
    if (is_invalid_object_handle(handle) || property == NULL) return 0;
    if (((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    index = object_field_index((const scriptgo_object *)handle, property);
    if (index < 0) return 0;
    return scriptgo_object_unknown_get(handle, index, out_tag, out_payload);
}

static int object_property_index_for_set(void *handle, const char *property) {
    scriptgo_object *object = (scriptgo_object *)handle;
    int index;
    size_t property_length;
    size_t type_length;
    size_t offset;
    char *type_name;

    if (is_invalid_object_handle(handle) || property == NULL || object->magic != SCRIPTGO_OBJECT_MAGIC) {
        return -1;
    }
    if (object->type_name == NULL) {
        object->type_name = strdup("");
        if (object->type_name == NULL) return -1;
    }
    index = object_field_index(object, property);
    if (index >= 0) return index;
    if (is_class_descriptor(object->type_name)) return -1;
    /* Empty object literals become dictionary-backed once a dynamic key is assigned. */
    if (strcmp(object->type_name, "__shape_empty") == 0) {
        char *dictionary_name = strdup("__json__");
        if (dictionary_name == NULL) return -1;
        free((void *)object->type_name);
        object->type_name = dictionary_name;
    }
	if (object->field_count >= 64) {
		return -1;
	}
	property_length = strlen(property);
	type_length = strlen(object->type_name);
	/* Dynamic fields use a length-prefixed extension so property names may
	 * contain the ':' separator used by legacy fixed-layout descriptors. */
	offset = type_length;
	type_name = (char *)malloc(type_length + property_length + 32);
	if (type_name == NULL) return -1;
	memcpy(type_name, object->type_name, type_length);
	offset += (size_t)snprintf(type_name + offset, property_length + 32, "|%zu:", property_length);
	memcpy(type_name + offset, property, property_length);
	offset += property_length;
	type_name[offset] = '\0';
    free((void *)object->type_name);
    object->type_name = type_name;
    return (int)object->field_count;
}

int scriptgo_object_property_number_get(void *handle, const char *property, double *out_value) {
    int index;
    if (out_value == NULL) return object_fail("scriptgo object property number output is invalid");
    *out_value = NAN;
	if (is_invalid_object_handle(handle) || property == NULL || !scriptgo_gc_is_registered(handle) || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    index = object_field_index((const scriptgo_object *)handle, property);
    return index < 0 ? 0 : scriptgo_object_number_get(handle, index, out_value);
}

int scriptgo_object_property_string_get(void *handle, const char *property, const char **out_value) {
    uint32_t tag;
    uint64_t payload;
    if (out_value == NULL) return object_fail("scriptgo object property string output is invalid");
    *out_value = &scriptgo_undefined_sentinel;
    if (is_invalid_object_handle(handle) || property == NULL || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    if (scriptgo_object_property_unknown_get(handle, property, &tag, &payload) != 0) return -1;
    if (tag == SCRIPTGO_TAG_STRING) {
        *out_value = (const char *)(uintptr_t)payload;
    } else if (tag == SCRIPTGO_TAG_NULL) {
        *out_value = NULL;
    }
    return 0;
}

int scriptgo_object_property_string_set(void *handle, const char *property, const char *value) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property string set failed") : scriptgo_object_string_set(handle, index, value);
}

int scriptgo_object_property_bool_get(void *handle, const char *property, int32_t *out_value) {
    int index;
    if (out_value == NULL) return object_fail("scriptgo object property bool output is invalid");
    *out_value = 0;
    if (is_invalid_object_handle(handle) || property == NULL || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    index = object_field_index((const scriptgo_object *)handle, property);
    return index < 0 ? 0 : scriptgo_object_bool_get(handle, index, out_value);
}

int scriptgo_object_property_bigint_get(void *handle, const char *property, int64_t *out_value) {
    int index;
    if (out_value == NULL) return object_fail("scriptgo object property bigint output is invalid");
    *out_value = 0;
    if (is_invalid_object_handle(handle) || property == NULL || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    index = object_field_index((const scriptgo_object *)handle, property);
    return index < 0 ? 0 : scriptgo_object_bigint_get(handle, index, out_value);
}

int scriptgo_object_property_ptr_get(void *handle, const char *property, void **out_value) {
    int index;
    if (out_value == NULL) return object_fail("scriptgo object property pointer output is invalid");
    *out_value = (void *)&scriptgo_undefined_sentinel;
    if (is_invalid_object_handle(handle) || property == NULL || ((scriptgo_object *)handle)->magic != SCRIPTGO_OBJECT_MAGIC) return 0;
    index = object_field_index((const scriptgo_object *)handle, property);
    return index < 0 ? 0 : scriptgo_object_ptr_get(handle, index, out_value);
}

int scriptgo_object_property_number_set(void *handle, const char *property, double value) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property number set failed") : scriptgo_object_number_set(handle, index, value);
}

int scriptgo_object_property_bool_set(void *handle, const char *property, int32_t value) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property bool set failed") : scriptgo_object_bool_set(handle, index, value);
}

int scriptgo_object_property_bigint_set(void *handle, const char *property, int64_t value) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property bigint set failed") : scriptgo_object_bigint_set(handle, index, value);
}

int scriptgo_object_property_ptr_set(void *handle, const char *property, void *value) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property pointer set failed") : scriptgo_object_ptr_set(handle, index, value);
}

int scriptgo_object_property_unknown_set(void *handle, const char *property,
                                         uint32_t tag, uint64_t payload) {
    int index = object_property_index_for_set(handle, property);
    return index < 0 ? object_fail("scriptgo object property set failed") : scriptgo_object_unknown_set(handle, index, tag, payload);
}

int scriptgo_object_type_set(void *handle, const char *type_name) {
    char *owned_type_name = NULL;
    if (is_invalid_object_handle(handle)) {
        return object_fail("scriptgo object type set failed");
    }
    if (type_name != NULL) {
        owned_type_name = strdup(type_name);
        if (owned_type_name == NULL) return object_fail("scriptgo object type allocation failed");
    }
    free((void *)((scriptgo_object *)handle)->type_name);
    ((scriptgo_object *)handle)->type_name = owned_type_name;
    return 0;
}

int scriptgo_object_type_get(void *handle, const char **out_type) {
    if (is_invalid_object_handle(handle) || out_type == NULL) {
        return object_fail("scriptgo object type get failed");
    }
    *out_type = ((scriptgo_object *)handle)->type_name;
    return 0;
}

int scriptgo_object_instanceof(void *handle, const char *class_name, int32_t *out_result) {
    if (out_result == NULL) {
        return object_fail("scriptgo instanceof null output");
    }
    if (is_invalid_object_handle(handle) || class_name == NULL) {
        *out_result = 0;
        return 0;
    }
    if (!scriptgo_gc_is_registered(handle)) {
        *out_result = 0;
        return 0;
    }
    if (strcmp(class_name, "Object") == 0) {
        *out_result = 1;
        return 0;
    }
    // Typed arrays and buffers have their own native header rather than a
    // scriptgo_object type name. Their GC tag still lets instanceof preserve
    // the JavaScript built-in identity without dereferencing an object layout.
    int gc_tag = scriptgo_gc_get_tag(handle);
    if (gc_tag == SCRIPTGO_ARRAYBUFFER_GC_TAG) {
        *out_result = strcmp(class_name, "ArrayBuffer") == 0 ? 1 : 0;
        return 0;
    }
    if (gc_tag == 6 && class_name != NULL) {
        uint32_t magic = *(const uint32_t *)handle;
        if (magic == SCRIPTGO_MAGIC_TYPEDARRAY || magic == SCRIPTGO_MAGIC_BUFFER) {
            if (strcmp(class_name, "Buffer") == 0 ||
                strcmp(class_name, "Uint8Array") == 0 ||
                strcmp(class_name, "Int8Array") == 0 ||
                strcmp(class_name, "Uint8ClampedArray") == 0 ||
                strcmp(class_name, "Int16Array") == 0 ||
                strcmp(class_name, "Uint16Array") == 0 ||
                strcmp(class_name, "Int32Array") == 0 ||
                strcmp(class_name, "Uint32Array") == 0 ||
                strcmp(class_name, "Float32Array") == 0 ||
                strcmp(class_name, "Float64Array") == 0 ||
                strcmp(class_name, "BigInt64Array") == 0 ||
                strcmp(class_name, "BigUint64Array") == 0) {
                *out_result = 1;
                return 0;
            }
        } else if (magic == SCRIPTGO_MAGIC_DATAVIEW && strcmp(class_name, "DataView") == 0) {
            *out_result = 1;
            return 0;
        }
    }
    scriptgo_object *obj = (scriptgo_object *)handle;
    if (obj->magic != SCRIPTGO_OBJECT_MAGIC || obj->type_name == NULL) {
        *out_result = 0;
        return 0;
    }
    if (is_class_descriptor(obj->type_name)) {
        const char *cursor = obj->type_name + 10;
        size_t class_name_length = strlen(class_name);
        for (;;) {
            char kind;
            const char *value;
            size_t value_length;
            int token = next_class_descriptor_token(&cursor, &kind, &value, &value_length);
            if (token <= 0) {
                *out_result = 0;
                return 0;
            }
            if ((kind == 'c' || kind == 'b' || kind == 'f') && value_length == class_name_length &&
                memcmp(value, class_name, value_length) == 0) {
                *out_result = 1;
                return 0;
            }
        }
    }
    size_t cn_len = strlen(class_name);
    if (strncmp(obj->type_name, class_name, cn_len) == 0 &&
        (obj->type_name[cn_len] == '\0' || (obj->type_name[cn_len] == '_' && obj->type_name[cn_len+1] == '_'))) {
        *out_result = 1;
        return 0;
    }
    char needle[256];
    snprintf(needle, sizeof(needle), ":%s:", class_name);
    *out_result = (strstr(obj->type_name, needle) != NULL) ? 1 : 0;
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set_tag(void *handle, int64_t tag);
int scriptgo_array_release(void *handle);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_set_owned_data(void *handle, void *owned_data);
int scriptgo_array_push(void *handle, const void *value, double *out_length);

static int object_key_count(const scriptgo_object *object) {
    const char *cursor;
    int count = 0;

    if (object == NULL || object->type_name == NULL) return 0;
    if (is_class_descriptor(object->type_name)) {
        cursor = object->type_name + 10;
        for (;;) {
            char kind;
            const char *value;
            size_t value_length;
            int token = next_class_descriptor_token(&cursor, &kind, &value, &value_length);
            if (token <= 0) return token < 0 ? -1 : count;
            if (kind == 'f') count++;
        }
    }
    if (strncmp(object->type_name, "__json__", 8) == 0) {
        cursor = object->type_name + 8;
        while (*cursor != '\0') {
            size_t key_length = 0;
            if (*cursor++ != '|') return -1;
            if (*cursor < '0' || *cursor > '9') return -1;
            while (*cursor >= '0' && *cursor <= '9') {
                key_length = key_length * 10 + (size_t)(*cursor - '0');
                cursor++;
            }
            if (*cursor++ != ':' || strlen(cursor) < key_length) return -1;
            cursor += key_length;
            count++;
        }
        return count;
    }
    if (object->type_name[0] != ':') return 0;
    cursor = object->type_name;
    while (*cursor != '\0') {
        const char *field_start;
        const char *field_end;
        if (*cursor == ':') cursor++;
        field_start = cursor;
        field_end = strchr(field_start, ':');
        if (field_end == NULL) return -1;
        if (field_end != field_start) count++;
        cursor = field_end + 1;
    }
    return count;
}

static int object_key_storage_size(const scriptgo_object *object, size_t *out_size) {
    const char *cursor;
    size_t total = 0;

    if (out_size == NULL) return -1;
    *out_size = 0;
    if (object == NULL || object->type_name == NULL) return 0;
    if (is_class_descriptor(object->type_name)) {
        cursor = object->type_name + 10;
        for (;;) {
            char kind;
            const char *value;
            size_t value_length;
            int token = next_class_descriptor_token(&cursor, &kind, &value, &value_length);
            if (token <= 0) return token < 0 ? -1 : (*out_size = total, 0);
            if (kind == 'f') {
                if (value_length > SIZE_MAX - total - 1) return -1;
                total += value_length + 1;
            }
        }
    }
    if (strncmp(object->type_name, "__json__", 8) == 0) {
        cursor = object->type_name + 8;
        while (*cursor != '\0') {
            size_t key_length = 0;
            if (*cursor++ != '|') return -1;
            if (*cursor < '0' || *cursor > '9') return -1;
            while (*cursor >= '0' && *cursor <= '9') {
                key_length = key_length * 10 + (size_t)(*cursor - '0');
                cursor++;
            }
            if (*cursor++ != ':' || strlen(cursor) < key_length) return -1;
            if (key_length > SIZE_MAX - total - 1) return -1;
            total += key_length + 1;
            cursor += key_length;
        }
        *out_size = total;
        return 0;
    }
    if (object->type_name[0] != ':') return 0;
    cursor = object->type_name;
    while (*cursor != '\0') {
        const char *field_start;
        const char *field_end;
        if (*cursor == ':') cursor++;
        field_start = cursor;
        field_end = strchr(field_start, ':');
        if (field_end == NULL) return -1;
        if (field_end != field_start) {
            size_t field_length = (size_t)(field_end - field_start);
            if (field_length > SIZE_MAX - total - 1) return -1;
            total += field_length + 1;
        }
        cursor = field_end + 1;
    }
    *out_size = total;
    return 0;
}

int scriptgo_object_keys(void *handle, void **out_array) {
    scriptgo_object *object = (scriptgo_object *)handle;
    int count;
    int index = 0;
    size_t storage_size;
    size_t storage_offset = 0;
    char *storage = NULL;

    if (out_array == NULL || is_invalid_object_handle(handle) || object->magic != SCRIPTGO_OBJECT_MAGIC) {
        return object_fail("scriptgo object keys arguments are invalid");
    }
    count = object_key_count(object);
    if (count < 0) return object_fail("scriptgo object key metadata is invalid");
    if (object_key_storage_size(object, &storage_size) != 0) return object_fail("scriptgo object key metadata is invalid");
    if (scriptgo_array_new(count, (int64_t)sizeof(char *), out_array) != 0) return -1;
    if (scriptgo_array_set_tag(*out_array, SCRIPTGO_OBJECT_TAG_STRING) != 0) {
        scriptgo_array_release(*out_array);
        *out_array = NULL;
        return -1;
    }
    if (storage_size != 0) {
        storage = (char *)malloc(storage_size);
        if (storage == NULL || scriptgo_array_set_owned_data(*out_array, storage) != 0) {
            free(storage);
            scriptgo_array_release(*out_array);
            *out_array = NULL;
            return object_fail("scriptgo object key allocation failed");
        }
    }
    if (object->type_name != NULL && count > 0) {
        const char *cursor = object->type_name;
        if (is_class_descriptor(object->type_name)) {
            cursor += 10;
            while (*cursor != '\0') {
                char kind;
                const char *value;
                size_t value_length;
                if (next_class_descriptor_token(&cursor, &kind, &value, &value_length) <= 0) break;
                if (kind == 'f') {
                    char *key = storage + storage_offset;
                    memcpy(key, value, value_length);
                    key[value_length] = '\0';
                    if (scriptgo_array_set(*out_array, (double)index, &key) != 0) {
                        scriptgo_array_release(*out_array);
                        *out_array = NULL;
                        return object_fail("scriptgo object key allocation failed");
                    }
                    storage_offset += value_length + 1;
                    index++;
                }
            }
        } else if (strncmp(object->type_name, "__json__", 8) == 0) {
            cursor += 8;
            while (*cursor != '\0') {
                size_t key_length = 0;
                cursor++;
                while (*cursor >= '0' && *cursor <= '9') {
                    key_length = key_length * 10 + (size_t)(*cursor - '0');
                    cursor++;
                }
                cursor++;
                char *key = storage + storage_offset;
                memcpy(key, cursor, key_length);
                key[key_length] = '\0';
                if (scriptgo_array_set(*out_array, (double)index, &key) != 0) {
                    scriptgo_array_release(*out_array);
                    *out_array = NULL;
                    return object_fail("scriptgo object key allocation failed");
                }
                storage_offset += key_length + 1;
                index++;
                cursor += key_length;
            }
        } else if (object->type_name[0] == ':') {
            while (*cursor != '\0') {
                const char *start = cursor;
                const char *end = strchr(start, ':');
                if (end == NULL) break;
                if (end != start) {
                    size_t key_length = (size_t)(end - start);
                    char *key = storage + storage_offset;
                    memcpy(key, start, key_length);
                    key[key_length] = '\0';
                    if (scriptgo_array_set(*out_array, (double)index, &key) != 0) {
                        scriptgo_array_release(*out_array);
                        *out_array = NULL;
                        return object_fail("scriptgo object key allocation failed");
                    }
                    storage_offset += key_length + 1;
                    index++;
                }
                cursor = end + 1;
            }
        }
    }
    return 0;
}

int scriptgo_object_release(void *handle) {
    if (is_invalid_object_handle(handle)) {
        return 0;
    }
    scriptgo_object *object = handle;
    scriptgo_gc_unregister(object);
    free((void *)object->type_name);
    free(object);
    return 0;
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
} scriptgo_grp_arr;

typedef struct {
    void *fn_ptr;
    void *env;
} scriptgo_grp_closure;

int scriptgo_object_group_by(void *handle, void *closure_handle, void **out_object) {
    if (out_object == NULL) return object_fail("scriptgo object groupBy output is invalid");
    if (scriptgo_object_new(0, out_object) != 0) return -1;
    if (handle == NULL || closure_handle == NULL) return 0;

    scriptgo_grp_arr *array = (scriptgo_grp_arr *)handle;
    scriptgo_grp_closure *c = (scriptgo_grp_closure *)closure_handle;
    if (array->length <= 0) return 0;

    for (int64_t i = 0; i < array->length; i++) {
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        char *key = NULL;
        if (array->element_size == sizeof(double)) {
            double item = *(double *)(array->data + (size_t)i * sizeof(double));
            union { double d; int64_t i; } u_item;
            u_item.d = item;
            char *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
                (char *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
            key = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        } else {
            void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
            char *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
                (char *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
            key = fn(c->env, 4, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        }
        if (key == NULL) key = "undefined";
        void *sub_arr = NULL;
        scriptgo_object_property_ptr_get(*out_object, key, &sub_arr);
        if (sub_arr == NULL || sub_arr == (void *)&scriptgo_undefined_sentinel) {
            if (scriptgo_array_new(0, array->element_size, &sub_arr) != 0) return -1;
            if (array->element_tag > 0) scriptgo_array_set_tag(sub_arr, array->element_tag);
            scriptgo_object_property_ptr_set(*out_object, key, sub_arr);
        }
        double dummy;
        void *val_ptr = array->data + (size_t)i * (size_t)array->element_size;
        scriptgo_array_push(sub_arr, val_ptr, &dummy);
    }
    return 0;
}
