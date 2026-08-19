#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} scriptgo_object;

int scriptgo_runtime_set_error(const char *message);

static int object_fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_object_new(int64_t field_count, void **out_object) {
    if (out_object == NULL || field_count < 0) {
        return object_fail("scriptgo object allocation failed");
    }
    scriptgo_object *object = calloc(1, sizeof(*object) + (size_t)field_count * sizeof(object->fields[0]));
    if (object == NULL) {
        return object_fail("scriptgo object allocation failed");
    }
    object->field_count = field_count;
    *out_object = object;
    return 0;
}

int scriptgo_object_number_set(void *handle, int64_t index, double value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    memcpy(&((scriptgo_object *)handle)->fields[index], &value, sizeof(value));
    return 0;
}

int scriptgo_object_number_get(void *handle, int64_t index, double *out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    memcpy(out_value, &((scriptgo_object *)handle)->fields[index], sizeof(*out_value));
    return 0;
}

int scriptgo_object_string_set(void *handle, int64_t index, const char *value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    ((scriptgo_object *)handle)->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_string_get(void *handle, int64_t index, const char **out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    *out_value = (const char *)((scriptgo_object *)handle)->fields[index];
    return 0;
}

int scriptgo_object_bool_set(void *handle, int64_t index, int32_t value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    ((scriptgo_object *)handle)->fields[index] = (uintptr_t)(value != 0 ? 1 : 0);
    return 0;
}

int scriptgo_object_bool_get(void *handle, int64_t index, int32_t *out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    *out_value = (int32_t)((scriptgo_object *)handle)->fields[index];
    return 0;
}

int scriptgo_object_ptr_set(void *handle, int64_t index, void *value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    ((scriptgo_object *)handle)->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_ptr_get(void *handle, int64_t index, void **out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    *out_value = (void *)((scriptgo_object *)handle)->fields[index];
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
    if (obj->type_name == NULL) {
        *out_result = 0;
        return 0;
    }
    char needle[256];
    snprintf(needle, sizeof(needle), ":%s:", class_name);
    *out_result = (strstr(obj->type_name, needle) != NULL) ? 1 : 0;
    return 0;
}

int scriptgo_object_release(void *handle) {
    if (handle == NULL) {
        return 0;
    }
    scriptgo_object *object = handle;
    free(object);
    return 0;
}
