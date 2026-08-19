#include <stdint.h>
#include <stdlib.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct {
    int64_t length;
    int64_t capacity;
    double *data;
} scriptgo_array_number;

typedef struct {
    int64_t length;
    int64_t capacity;
    void **data;
} scriptgo_array_string;

static int fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_array_number_new(int64_t length, void **out_array) {
    if (out_array == NULL || length < 0) {
        return fail("scriptgo array allocation failed");
    }
    scriptgo_array_number *array = calloc(1, sizeof(*array));
    if (array == NULL) {
        return fail("scriptgo array allocation failed");
    }
    if (length > 0) {
        array->data = calloc((size_t)length, sizeof(*array->data));
        if (array->data == NULL) {
            free(array);
            return fail("scriptgo array allocation failed");
        }
    }
    array->length = length;
    array->capacity = length;
    *out_array = array;
    return 0;
}

int scriptgo_array_number_get(void *handle, double index, double *out_value) {
    if (handle == NULL || out_value == NULL) {
        return fail("scriptgo array access failed");
    }
    scriptgo_array_number *array = handle;
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    *out_value = array->data[(int64_t)index];
    return 0;
}

int scriptgo_array_number_set(void *handle, double index, double value) {
    if (handle == NULL) {
        return fail("scriptgo array access failed");
    }
    scriptgo_array_number *array = handle;
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    array->data[(int64_t)index] = value;
    return 0;
}

int scriptgo_array_number_length(void *handle, int64_t *out_length) {
    if (handle == NULL || out_length == NULL) {
        return fail("scriptgo array access failed");
    }
    *out_length = ((scriptgo_array_number *)handle)->length;
    return 0;
}

int scriptgo_array_length(void *handle, int64_t *out_length) {
    if (handle == NULL || out_length == NULL) {
        return fail("scriptgo array access failed");
    }
    *out_length = *(int64_t *)handle;
    return 0;
}

int scriptgo_array_number_release(void *handle) {
    if (handle != NULL) {
        scriptgo_array_number *array = handle;
        free(array->data);
        free(array);
    }
    return 0;
}

int scriptgo_array_string_new(int64_t length, void **out_array) {
    if (out_array == NULL || length < 0) {
        return fail("scriptgo array allocation failed");
    }
    scriptgo_array_string *array = calloc(1, sizeof(*array));
    if (array == NULL) {
        return fail("scriptgo array allocation failed");
    }
    if (length > 0) {
        array->data = calloc((size_t)length, sizeof(*array->data));
        if (array->data == NULL) {
            free(array);
            return fail("scriptgo array allocation failed");
        }
    }
    array->length = length;
    array->capacity = length;
    *out_array = array;
    return 0;
}

int scriptgo_array_string_get(void *handle, double index, void **out_value) {
    if (handle == NULL || out_value == NULL) {
        return fail("scriptgo array access failed");
    }
    scriptgo_array_string *array = handle;
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    *out_value = array->data[(int64_t)index];
    return 0;
}

int scriptgo_array_string_set(void *handle, double index, void *value) {
    if (handle == NULL) {
        return fail("scriptgo array access failed");
    }
    scriptgo_array_string *array = handle;
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    array->data[(int64_t)index] = value;
    return 0;
}

int scriptgo_array_string_length(void *handle, int64_t *out_length) {
    if (handle == NULL || out_length == NULL) {
        return fail("scriptgo array access failed");
    }
    *out_length = ((scriptgo_array_string *)handle)->length;
    return 0;
}

int scriptgo_array_string_release(void *handle) {
    if (handle != NULL) {
        scriptgo_array_string *array = handle;
        free(array->data);
        free(array);
    }
    return 0;
}
