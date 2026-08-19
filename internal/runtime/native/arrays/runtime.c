#include <stdint.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
} scriptgo_array;

static int fail(const char *message) { return scriptgo_runtime_set_error(message); }

static int check_index(scriptgo_array *array, double index, size_t *offset) {
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    *offset = (size_t)index * (size_t)array->element_size;
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array) {
    scriptgo_array *array;
    size_t byte_length;
    if (out_array == NULL || length < 0 || element_size <= 0) {
        return fail("scriptgo array allocation failed");
    }
    if ((uint64_t)length > SIZE_MAX / (uint64_t)element_size) {
        return fail("scriptgo array allocation failed");
    }
    array = calloc(1, sizeof(*array));
    if (array == NULL) return fail("scriptgo array allocation failed");
    byte_length = (size_t)length * (size_t)element_size;
    if (byte_length != 0) {
        array->data = calloc(1, byte_length);
        if (array->data == NULL) {
            free(array);
            return fail("scriptgo array allocation failed");
        }
    }
    array->length = length;
    array->capacity = length;
    array->element_size = element_size;
    *out_array = array;
    return 0;
}

int scriptgo_array_get(void *handle, double index, void *out_value) {
    scriptgo_array *array = handle;
    size_t offset;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (check_index(array, index, &offset) != 0) return -1;
    memcpy(out_value, array->data + offset, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_set(void *handle, double index, const void *value) {
    scriptgo_array *array = handle;
    size_t offset;
    if (array == NULL || value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (check_index(array, index, &offset) != 0) return -1;
    memcpy(array->data + offset, value, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_length(void *handle, int64_t *out_length) {
    scriptgo_array *array = handle;
    if (array == NULL || out_length == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    *out_length = array->length;
    return 0;
}

int scriptgo_array_release(void *handle) {
    scriptgo_array *array = handle;
    if (array != NULL) {
        free(array->data);
        free(array);
    }
    return 0;
}
