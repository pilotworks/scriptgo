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

int scriptgo_array_push(void *handle, const void *value, double *out_length) {
    scriptgo_array *array = handle;
    if (array == NULL || value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (array->length >= array->capacity) {
        int64_t new_cap = array->capacity <= 0 ? 4 : array->capacity * 2;
        unsigned char *new_data = realloc(array->data, (size_t)new_cap * (size_t)array->element_size);
        if (new_data == NULL) {
            return fail("scriptgo array reallocation failed");
        }
        array->data = new_data;
        array->capacity = new_cap;
    }
    memcpy(array->data + (size_t)array->length * (size_t)array->element_size, value, (size_t)array->element_size);
    array->length++;
    if (out_length != NULL) {
        *out_length = (double)array->length;
    }
    return 0;
}

int scriptgo_array_pop(void *handle, void *out_value) {
    scriptgo_array *array = handle;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (array->length <= 0) {
        memset(out_value, 0, (size_t)array->element_size);
        return 0;
    }
    array->length--;
    memcpy(out_value, array->data + (size_t)array->length * (size_t)array->element_size, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_slice(void *handle, double start_val, double end_val, void **out_array) {
    scriptgo_array *array = handle;
    int64_t length, start, end, new_len;
    if (array == NULL || out_array == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    length = array->length;
    if (start_val < 0.0) {
        start = length + (int64_t)start_val;
        if (start < 0) start = 0;
    } else {
        start = (int64_t)start_val;
        if (start > length) start = length;
    }
    if (end_val < 0.0) {
        end = length;
    } else {
        end = (int64_t)end_val;
        if (end > length) end = length;
    }
    if (end < start) end = start;
    new_len = end - start;
    if (scriptgo_array_new(new_len, array->element_size, out_array) != 0) {
        return -1;
    }
    if (new_len > 0) {
        scriptgo_array *res = *out_array;
        memcpy(res->data, array->data + (size_t)start * (size_t)array->element_size, (size_t)new_len * (size_t)array->element_size);
    }
    return 0;
}

int scriptgo_array_index_of_number(void *handle, double target, double from_index, double *out_index) {
    scriptgo_array *array = handle;
    int64_t start, i;
    if (array == NULL || out_index == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array access failed");
    }
    start = from_index < 0.0 ? 0 : (int64_t)from_index;
    for (i = start; i < array->length; i++) {
        double val = *(double *)(array->data + (size_t)i * sizeof(double));
        if (val == target) {
            *out_index = (double)i;
            return 0;
        }
    }
    *out_index = -1.0;
    return 0;
}

int scriptgo_array_index_of_string(void *handle, const char *target, double from_index, double *out_index) {
    scriptgo_array *array = handle;
    int64_t start, i;
    if (array == NULL || target == NULL || out_index == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array access failed");
    }
    start = from_index < 0.0 ? 0 : (int64_t)from_index;
    for (i = start; i < array->length; i++) {
        const char *val = *(const char **)(array->data + (size_t)i * sizeof(char *));
        if (val != NULL && strcmp(val, target) == 0) {
            *out_index = (double)i;
            return 0;
        }
    }
    *out_index = -1.0;
    return 0;
}

int scriptgo_array_includes_number(void *handle, double target, double *out_bool) {
    double idx = -1.0;
    if (scriptgo_array_index_of_number(handle, target, 0.0, &idx) != 0) return -1;
    *out_bool = idx >= 0.0 ? 1.0 : 0.0;
    return 0;
}

int scriptgo_array_includes_string(void *handle, const char *target, double *out_bool) {
    double idx = -1.0;
    if (scriptgo_array_index_of_string(handle, target, 0.0, &idx) != 0) return -1;
    *out_bool = idx >= 0.0 ? 1.0 : 0.0;
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

