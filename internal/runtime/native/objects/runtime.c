#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    uint32_t refcount;
    int64_t field_count;
    uintptr_t fields[];
} scriptgo_object_v2;

static int object_fail(const char *message) {
    fputs(message, stderr);
    fputc('\n', stderr);
    exit(1);
    return -1;
}

int scriptgo_object_new_v2(int64_t field_count, void **out_object) {
    if (out_object == NULL || field_count < 0) {
        return object_fail("scriptgo object allocation failed");
    }
    scriptgo_object_v2 *object = calloc(1, sizeof(*object) + (size_t)field_count * sizeof(object->fields[0]));
    if (object == NULL) {
        return object_fail("scriptgo object allocation failed");
    }
    object->refcount = 1;
    object->field_count = field_count;
    *out_object = object;
    return 0;
}

int scriptgo_object_number_set_v2(void *handle, int64_t index, double value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object_v2 *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    memcpy(&((scriptgo_object_v2 *)handle)->fields[index], &value, sizeof(value));
    return 0;
}

int scriptgo_object_number_get_v2(void *handle, int64_t index, double *out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object_v2 *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    memcpy(out_value, &((scriptgo_object_v2 *)handle)->fields[index], sizeof(*out_value));
    return 0;
}

int scriptgo_object_string_set_v2(void *handle, int64_t index, const char *value) {
    if (handle == NULL || index < 0 || index >= ((scriptgo_object_v2 *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    ((scriptgo_object_v2 *)handle)->fields[index] = (uintptr_t)value;
    return 0;
}

int scriptgo_object_string_get_v2(void *handle, int64_t index, const char **out_value) {
    if (handle == NULL || out_value == NULL || index < 0 || index >= ((scriptgo_object_v2 *)handle)->field_count) {
        return object_fail("scriptgo object field access failed");
    }
    *out_value = (const char *)((scriptgo_object_v2 *)handle)->fields[index];
    return 0;
}

int scriptgo_object_release_v2(void *handle) {
    if (handle == NULL) {
        return 0;
    }
    scriptgo_object_v2 *object = handle;
    if (--object->refcount == 0) {
        free(object);
    }
    return 0;
}
