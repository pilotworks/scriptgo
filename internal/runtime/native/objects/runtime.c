#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} scriptgo_object;

int scriptgo_runtime_set_error(const char *message);

static int object_fail(const char *message) { return scriptgo_runtime_set_error(message); }

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
    if (a == NULL || b == NULL) {
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
    if (handle == NULL || index < 0 || index >= 64) {
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
    if (handle == NULL || index < 0 || index >= 64) {
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
    if (handle == NULL || index < 0 || index >= 64) {
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
    if (handle == NULL || index < 0 || index >= 64) {
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
    if (handle == NULL || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    o->fields[index] = (uintptr_t)(value != 0 ? 1 : 0);
    return 0;
}

int scriptgo_object_bool_get(void *handle, int64_t index, int32_t *out_value) {
    if (out_value == NULL) {
        return 0;
    }
    if (handle == NULL || index < 0 || index >= 64) {
        *out_value = 0;
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (o->magic != SCRIPTGO_OBJECT_MAGIC || index >= o->field_count) {
        *out_value = 0;
        return 0;
    }
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) {
        *out_value = 0;
    } else {
        *out_value = (int32_t)val;
    }
    return 0;
}

int scriptgo_object_ptr_set(void *handle, int64_t index, void *value) {
    if (handle == NULL || index < 0 || index >= 64) {
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
    if (handle == NULL || index < 0 || index >= 64) {
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
        *out_value = (void *)val;
    }
    return 0;
}

int scriptgo_object_unknown_set(void *handle, int64_t index, uint32_t tag, uint64_t payload) {
    if (handle == NULL || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    if (index >= o->field_count) {
        o->field_count = index + 1;
    }
    (void)tag;
    o->fields[index] = (uintptr_t)payload;
    return 0;
}

int scriptgo_object_unknown_get(void *handle, int64_t index, uint32_t *out_tag, uint64_t *out_payload) {
    if (out_tag == NULL || out_payload == NULL) {
        return 0;
    }
    *out_tag = 0;
    *out_payload = 0;
    if (handle == NULL || index < 0 || index >= 64) {
        return 0;
    }
    scriptgo_object *o = (scriptgo_object *)handle;
    uintptr_t val = o->fields[index];
    if (val == (uintptr_t)SCRIPTGO_OBJECT_NAN_BITS) {
        *out_tag = 0;
        *out_payload = 0;
    } else if (val == 0) {
        *out_tag = 0;
        *out_payload = 0;
    } else if ((val & 0xFFF8000000000000ULL) != 0) {
        *out_tag = 3;
        *out_payload = (uint64_t)val;
    } else if (val < 4096) {
        *out_tag = 2;
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
    if (handle == NULL) {
        return object_fail("scriptgo object type set failed");
    }
    ((scriptgo_object *)handle)->type_name = type_name;
    return 0;
}

int scriptgo_object_type_get(void *handle, const char **out_type) {
    if (handle == NULL || out_type == NULL) {
        return object_fail("scriptgo object type get failed");
    }
    *out_type = ((scriptgo_object *)handle)->type_name;
    return 0;
}

int scriptgo_object_instanceof(void *handle, const char *class_name, int32_t *out_result) {
    if (out_result == NULL) {
        return object_fail("scriptgo instanceof null output");
    }
    if (handle == NULL || class_name == NULL) {
        *out_result = 0;
        return 0;
    }
    scriptgo_object *obj = (scriptgo_object *)handle;
    if (obj->magic != SCRIPTGO_OBJECT_MAGIC || obj->type_name == NULL) {
        *out_result = 0;
        return 0;
    }
    if (strcmp(obj->type_name, class_name) == 0) {
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
    if (handle == NULL) {
        return 0;
    }
    scriptgo_object *object = handle;
    scriptgo_gc_unregister(object);
    free(object);
    return 0;
}
