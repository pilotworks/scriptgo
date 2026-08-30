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

static int object_field_index(const scriptgo_object *object, const char *property) {
    const char *cursor;
    int index = 0;
    size_t property_len;

    if (object == NULL || object->type_name == NULL || property == NULL) {
        return -1;
    }
    property_len = strlen(property);
    cursor = object->type_name;
    while (*cursor != '\0') {
        const char *field_start;
        const char *field_end;
        size_t field_len;

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
    for (int64_t i = 0; i < capacity; i++) {
        uint64_t nan_bits = SCRIPTGO_OBJECT_NAN_BITS;
        memcpy(&object->fields[i], &nan_bits, sizeof(uint64_t));
    }
    scriptgo_gc_register(object, 1, (uint32_t)capacity);
    *out_object = object;
    return 0;
}

int scriptgo_object_number_set(void *handle, int64_t index, double value) {
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
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
        *out_value = NULL;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = NULL;
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) {
        *out_value = NULL;
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
    if (is_invalid_object_handle(handle) || index < 0 || index >= 64) {
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
        o->fields[index] = (uintptr_t)(1ULL << 32);
    } else if (tag == 0) {
        o->fields[index] = 0;
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
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS || val == 0) {
        *out_tag = 0;
        *out_payload = 0;
    } else if ((val >> 32) == 2) {
        *out_tag = 2;
        *out_payload = (val & 1);
    } else if ((val >> 32) == 1 && (val & 0xFFFFFFFF) == 0) {
        *out_tag = 1;
        *out_payload = 0;
    } else if ((val & 0xFFF8000000000000ULL) != 0) {
        *out_tag = 3;
        *out_payload = (uint64_t)val;
    } else {
        if (scriptgo_gc_is_registered((void *)val)) {
            *out_tag = 5;
        } else {
            *out_tag = 4;
        }
        *out_payload = (uint64_t)val;
    }
    return 0;
}

int scriptgo_object_type_set(void *handle, const char *type_name) {
    if (is_invalid_object_handle(handle)) {
        return object_fail("scriptgo object type set failed");
    }
    ((scriptgo_object *)handle)->type_name = type_name;
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

int scriptgo_object_keys(void *handle, void **out_array) {
    (void)handle;
    return scriptgo_array_new(0, 8, out_array);
}

int scriptgo_object_release(void *handle) {
    if (is_invalid_object_handle(handle)) {
        return 0;
    }
    scriptgo_object *object = handle;
    scriptgo_gc_unregister(object);
    free(object);
    return 0;
}
