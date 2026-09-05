#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_set_length(void *handle, double length);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
} scriptgo_array;

typedef struct {
    uint32_t tag;
    uint32_t padding;
    uint64_t payload;
} scriptgo_array_unknown;

static int fail(const char *message) { return scriptgo_runtime_set_error(message); }

static int check_index(scriptgo_array *array, double index, size_t *offset) {
    if (index != index || index < 0 || index != (double)(int64_t)index ||
        (int64_t)index >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    *offset = (size_t)index * (size_t)array->element_size;
    return 0;
}

int scriptgo_array_set_tag(void *handle, int64_t tag) {
    scriptgo_array *array = handle;
    if (array == NULL) return fail("scriptgo array null");
    array->element_tag = tag;
    return 0;
}

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count);
int scriptgo_gc_unregister(void *ptr);

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
    array->element_tag = 0;
    array->owned_data = NULL;
    scriptgo_gc_register(array, 2, 0);
    *out_array = array;
    return 0;
}

#define SCRIPTGO_OBJECT_MAGIC 0x53474F424A454354ULL
extern const char scriptgo_undefined_sentinel;
typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    uintptr_t fields[];
} scriptgo_runtime_object_header;

int scriptgo_array_get(void *handle, double index, void *out_value) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel || out_value == NULL) {
        return fail("scriptgo array access failed");
    }
    if (*(uint64_t *)handle == SCRIPTGO_OBJECT_MAGIC) {
        scriptgo_runtime_object_header *obj = handle;
        int64_t idx = (int64_t)index;
        if (index != index || index < 0 || idx >= obj->field_count) {
            return fail("scriptgo array index out of bounds");
        }
        memcpy(out_value, &obj->fields[idx], sizeof(uintptr_t));
        return 0;
    }
    scriptgo_array *array = handle;
    size_t offset;
    if (array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (check_index(array, index, &offset) != 0) {
        return -1;
    }
    if (array->element_size == 16) {
        memcpy(out_value, array->data + offset + 8, 8);
        return 0;
    }
    memcpy(out_value, array->data + offset, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_get_unknown(void *handle, double index, void *out_value) {
    scriptgo_array *array = handle;
    size_t offset;
    uint32_t *tag_ptr;
    uint64_t *payload_ptr;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    tag_ptr = (uint32_t *)out_value;
    tag_ptr[0] = 0; // UNDEFINED
    tag_ptr[1] = 0; // padding
    payload_ptr = (uint64_t *)((char *)out_value + 8);
    *payload_ptr = 0;
    if (index != index || index < 0 || index != (double)(int64_t)index) {
        return fail("scriptgo array index out of bounds");
    }
    if ((int64_t)index >= array->length) {
        return 0;
    }
    offset = (size_t)index * (size_t)array->element_size;
    if (array->element_size == 16) {
        memcpy(out_value, array->data + offset, 16);
        return 0;
    }
    if (array->element_size == 8) {
        uint64_t val;
        memcpy(&val, array->data + offset, 8);
        *payload_ptr = val;
        if (array->element_tag > 0) {
            *tag_ptr = (uint32_t)array->element_tag;
        } else if (val == 0) {
            *tag_ptr = 1; // NULL
        } else {
            *tag_ptr = 5; // OBJECT / POINTER / STRING
        }
        return 0;
    }
    if (array->element_size == 1) {
        uint8_t val = *(array->data + offset);
        *tag_ptr = 2; // BOOLEAN
        *payload_ptr = (uint64_t)val;
        return 0;
    }
    memcpy(out_value, array->data + offset, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_set(void *handle, double index, const void *value) {
    scriptgo_array *array = handle;
    size_t offset;
    int64_t idx;
    if (array == NULL || handle == (void *)&scriptgo_undefined_sentinel || value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (index != index || index < 0 || index != (double)(int64_t)index) {
        return fail("scriptgo array index out of bounds");
    }
    idx = (int64_t)index;
    if (idx >= array->length) {
        if (scriptgo_array_set_length(handle, (double)(idx + 1)) != 0) {
            return -1;
        }
    }
    offset = (size_t)idx * (size_t)array->element_size;
    memcpy(array->data + offset, value, (size_t)array->element_size);
    return 0;
}

extern const char scriptgo_undefined_sentinel;

int scriptgo_array_set_typed(void *handle, double index, const void *value,
                             int64_t value_size, int64_t tag) {
    scriptgo_array *array = handle;
    int64_t idx;
    size_t offset;
    if (array == NULL || handle == (void *)&scriptgo_undefined_sentinel || value == NULL ||
        array->element_size <= 0 || value_size <= 0 || value_size > 16) {
        return fail("scriptgo typed array assignment failed");
    }
    if (index != index || index < 0 || index != (double)(int64_t)index) {
        return fail("scriptgo array index out of bounds");
    }
    idx = (int64_t)index;
    if (idx >= array->length && scriptgo_array_set_length(handle, (double)(idx + 1)) != 0) {
        return -1;
    }
    offset = (size_t)idx * (size_t)array->element_size;
    if (array->element_size == 16 && value_size < 16) {
        uint32_t element_tag = (uint32_t)tag;
        uint32_t padding = 0;
        uint64_t payload = 0;
        if (value_size == 1) {
            uint8_t byte_value = *(const uint8_t *)value;
            payload = byte_value;
        } else {
            memcpy(&payload, value, (size_t)value_size);
        }
        if (value_size == sizeof(uintptr_t)) {
            void *pointer_value = NULL;
            memcpy(&pointer_value, value, sizeof(pointer_value));
            if (pointer_value == (void *)&scriptgo_undefined_sentinel) {
                element_tag = 0;
                payload = 0;
            } else if (pointer_value == NULL) {
                element_tag = 1;
                payload = 0;
            }
        }
        memcpy(array->data + offset, &element_tag, sizeof(element_tag));
        memcpy(array->data + offset + sizeof(element_tag), &padding, sizeof(padding));
        memcpy(array->data + offset + 8, &payload, sizeof(payload));
        return 0;
    }
    if (value_size != array->element_size) {
        return fail("scriptgo typed array assignment size mismatch");
    }
    memcpy(array->data + offset, value, (size_t)value_size);
    return 0;
}

int scriptgo_array_length(void *handle, int64_t *out_length) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel || out_length == NULL) {
        return fail("scriptgo array access failed");
    }
    if (*(uint64_t *)handle == SCRIPTGO_OBJECT_MAGIC) {
        scriptgo_runtime_object_header *obj = handle;
        *out_length = obj->field_count;
        return 0;
    }
    scriptgo_array *array = handle;
    if (array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    *out_length = array->length;
    return 0;
}

int scriptgo_array_set_length(void *handle, double length) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    int64_t new_len = (int64_t)length;
    if (new_len < 0) new_len = 0;
    if (new_len <= array->length) {
        array->length = new_len;
        return 0;
    }
    if (new_len > array->capacity) {
        int64_t new_cap = new_len;
        unsigned char *new_data = realloc(array->data, (size_t)new_cap * (size_t)array->element_size);
        if (new_data == NULL) return fail("scriptgo array reallocation failed");
        array->data = new_data;
        array->capacity = new_cap;
    }
    memset(array->data + (size_t)array->length * (size_t)array->element_size, 0, (size_t)(new_len - array->length) * (size_t)array->element_size);
    array->length = new_len;
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
    size_t offset;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (array->length == 0) {
        memset(out_value, 0, (size_t)array->element_size);
        return 0;
    }
    offset = (size_t)(array->length - 1) * (size_t)array->element_size;
    memcpy(out_value, array->data + offset, (size_t)array->element_size);
    array->length--;
    return 0;
}

int scriptgo_array_slice_with_size(void *handle, double start_val, double end_val, int64_t target_element_size, void **out_array) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel || out_array == NULL) {
        return fail("scriptgo array access failed");
    }
    int64_t length, start, end, new_len, element_size;
    void *src_data;
    scriptgo_array *source_array = NULL;
    int is_object = 0;
    if (*(uint64_t *)handle == SCRIPTGO_OBJECT_MAGIC) {
        scriptgo_runtime_object_header *obj = handle;
        length = obj->field_count;
        element_size = target_element_size > 0 ? target_element_size : (int64_t)sizeof(uintptr_t);
        src_data = obj->fields;
        is_object = 1;
    } else {
        scriptgo_array *array = handle;
        if (array->element_size <= 0) {
            return fail("scriptgo array access failed");
        }
        length = array->length;
        source_array = array;
        element_size = array->element_size == (int64_t)sizeof(scriptgo_array_unknown) &&
                       target_element_size > 0 && target_element_size != array->element_size
                           ? target_element_size : array->element_size;
        src_data = array->data;
    }
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
    if (scriptgo_array_new(new_len, element_size, out_array) != 0) {
        return -1;
    }
    if (new_len > 0) {
        scriptgo_array *res = *out_array;
        if (is_object) {
            scriptgo_runtime_object_header *obj = handle;
            if (element_size == 1) {
                for (int64_t i = 0; i < new_len; i++) {
                    uint8_t val = (uint8_t)obj->fields[start + i];
                    res->data[i] = val;
                }
            } else {
                memcpy(res->data, (unsigned char *)src_data + (size_t)start * (size_t)element_size, (size_t)new_len * (size_t)element_size);
            }
        } else {
            if (source_array != NULL && source_array->element_size == (int64_t)sizeof(scriptgo_array_unknown) &&
                target_element_size > 0 && target_element_size != source_array->element_size) {
                scriptgo_array_unknown *source = (scriptgo_array_unknown *)src_data + start;
                for (int64_t i = 0; i < new_len; i++) {
                    if (target_element_size == (int64_t)sizeof(uint8_t)) {
                        res->data[i] = source[i].payload != 0 ? 1 : 0;
                    } else if (target_element_size == (int64_t)sizeof(uint64_t)) {
                        memcpy(res->data + (size_t)i * (size_t)target_element_size, &source[i].payload, sizeof(source[i].payload));
                    } else {
                        scriptgo_gc_unregister(res);
                        free(res->data);
                        free(res);
                        return fail("scriptgo array slice element size is unsupported");
                    }
                }
            } else {
                memcpy(res->data, (unsigned char *)src_data + (size_t)start * (size_t)element_size, (size_t)new_len * (size_t)element_size);
            }
        }
    }
    return 0;
}

int scriptgo_array_slice(void *handle, double start_val, double end_val, void **out_array) {
    return scriptgo_array_slice_with_size(handle, start_val, end_val, 0, out_array);
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

int scriptgo_array_index_of_ptr(void *handle, const void *target, double from_index, double *out_index) {
    scriptgo_array *array = handle;
    int64_t start, i;
    if (array == NULL || out_index == NULL || array->element_size != sizeof(void *)) {
        return fail("scriptgo array access failed");
    }
    start = from_index < 0.0 ? 0 : (int64_t)from_index;
    for (i = start; i < array->length; i++) {
        const void *val = *(const void **)(array->data + (size_t)i * sizeof(void *));
        if (val == target) {
            *out_index = (double)i;
            return 0;
        }
    }
    *out_index = -1.0;
    return 0;
}

int scriptgo_array_includes_number(void *handle, double target, double *out_bool) {
    if (out_bool == NULL) return -1;
    scriptgo_array *array = handle;
    if (array == NULL) {
        *out_bool = 0.0;
        return 0;
    }
    double *data = (double *)array->data;
    int target_is_nan = isnan(target);
    for (size_t i = 0; i < array->length; i++) {
        if (target_is_nan) {
            if (isnan(data[i])) {
                *out_bool = 1.0;
                return 0;
            }
        } else if (data[i] == target) {
            *out_bool = 1.0;
            return 0;
        }
    }
    *out_bool = 0.0;
    return 0;
}

int scriptgo_array_includes_string(void *handle, const char *target, double *out_bool) {
    double idx = -1.0;
    if (scriptgo_array_index_of_string(handle, target, 0.0, &idx) != 0) return -1;
    *out_bool = idx >= 0.0 ? 1.0 : 0.0;
    return 0;
}

int scriptgo_array_includes_ptr(void *handle, const void *target, double *out_bool) {
    double idx = -1.0;
    if (scriptgo_array_index_of_ptr(handle, target, 0.0, &idx) != 0) return -1;
    *out_bool = idx >= 0.0 ? 1.0 : 0.0;
    return 0;
}

int scriptgo_array_release(void *handle) {
    scriptgo_array *array = handle;
    if (array != NULL) {
        scriptgo_gc_unregister(array);
        free(array->owned_data);
        free(array->data);
        free(array);
    }
    return 0;
}

int scriptgo_array_set_owned_data(void *handle, void *owned_data) {
    scriptgo_array *array = handle;
    if (array == NULL) {
        return fail("scriptgo array access failed");
    }
    array->owned_data = owned_data;
    return 0;
}

int scriptgo_array_at(void *handle, double index, void *out_value) {
    scriptgo_array *array = handle;
    int64_t idx;
    size_t offset;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (index != index || index != (double)(int64_t)index) {
        return fail("scriptgo array index out of bounds");
    }
    idx = (int64_t)index;
    if (idx < 0) {
        idx = array->length + idx;
    }
    if (idx < 0 || idx >= array->length) {
        memset(out_value, 0, (size_t)array->element_size);
        return 0;
    }
    offset = (size_t)idx * (size_t)array->element_size;
    memcpy(out_value, array->data + offset, (size_t)array->element_size);
    return 0;
}

int scriptgo_array_shift(void *handle, void *out_value) {
    scriptgo_array *array = handle;
    if (array == NULL || out_value == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (array->length <= 0) {
        memset(out_value, 0, (size_t)array->element_size);
        return 0;
    }
    memcpy(out_value, array->data, (size_t)array->element_size);
    if (array->length > 1) {
        memmove(array->data, array->data + (size_t)array->element_size, (size_t)(array->length - 1) * (size_t)array->element_size);
    }
    array->length--;
    return 0;
}

int scriptgo_array_unshift(void *handle, const void *value, double *out_length) {
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
    if (array->length > 0) {
        memmove(array->data + (size_t)array->element_size, array->data, (size_t)array->length * (size_t)array->element_size);
    }
    memcpy(array->data, value, (size_t)array->element_size);
    array->length++;
    if (out_length != NULL) {
        *out_length = (double)array->length;
    }
    return 0;
}

int scriptgo_array_reverse(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    int64_t i, j;
    unsigned char temp[64];
    if (array == NULL || array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    if (array->length > 1) {
        size_t elem_sz = (size_t)array->element_size;
        void *tmp_buf = elem_sz <= sizeof(temp) ? temp : malloc(elem_sz);
        if (tmp_buf == NULL) return fail("scriptgo array allocation failed");
        for (i = 0, j = array->length - 1; i < j; i++, j--) {
            memcpy(tmp_buf, array->data + (size_t)i * elem_sz, elem_sz);
            memcpy(array->data + (size_t)i * elem_sz, array->data + (size_t)j * elem_sz, elem_sz);
            memcpy(array->data + (size_t)j * elem_sz, tmp_buf, elem_sz);
        }
        if (tmp_buf != temp) free(tmp_buf);
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_concat(void *handle1, void *handle2, void **out_array) {
    scriptgo_array *arr1 = handle1;
    scriptgo_array *arr2 = handle2;
    int64_t total_len;
    scriptgo_array *res;
    if (arr1 == NULL || arr2 == NULL || out_array == NULL || arr1->element_size != arr2->element_size) {
        return fail("scriptgo array access failed");
    }
    total_len = arr1->length + arr2->length;
    if (scriptgo_array_new(total_len, arr1->element_size, out_array) != 0) {
        return -1;
    }
    res = *out_array;
    if (arr1->length > 0) {
        memcpy(res->data, arr1->data, (size_t)arr1->length * (size_t)arr1->element_size);
    }
    if (arr2->length > 0) {
        memcpy(res->data + (size_t)arr1->length * (size_t)arr1->element_size, arr2->data, (size_t)arr2->length * (size_t)arr2->element_size);
    }
    return 0;
}

int scriptgo_array_splice(void *handle, double start_val, double delete_count_val, void **out_array) {
    scriptgo_array *array = handle;
    int64_t length, start, delete_count, remaining;
    scriptgo_array *deleted;
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
    if (delete_count_val < 0.0) {
        delete_count = 0;
    } else {
        delete_count = (int64_t)delete_count_val;
        if (start + delete_count > length) {
            delete_count = length - start;
        }
    }
    if (scriptgo_array_new(delete_count, array->element_size, out_array) != 0) {
        return -1;
    }
    deleted = *out_array;
    if (delete_count > 0) {
        memcpy(deleted->data, array->data + (size_t)start * (size_t)array->element_size, (size_t)delete_count * (size_t)array->element_size);
        remaining = length - (start + delete_count);
        if (remaining > 0) {
            memmove(array->data + (size_t)start * (size_t)array->element_size, array->data + (size_t)(start + delete_count) * (size_t)array->element_size, (size_t)remaining * (size_t)array->element_size);
        }
        array->length -= delete_count;
    }
    return 0;
}

int scriptgo_array_join_number(void *handle, const char *separator, char **out_str) {
    scriptgo_array *array = handle;
    size_t cap = 256, len = 0;
    char *buf;
    const char *sep = separator != NULL ? separator : ",";
    size_t sep_len = strlen(sep);
    if (array == NULL || out_str == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array access failed");
    }
    buf = malloc(cap);
    if (buf == NULL) return fail("scriptgo string allocation failed");
    buf[0] = '\0';
    for (int64_t i = 0; i < array->length; i++) {
        double val = *(double *)(array->data + (size_t)i * sizeof(double));
        char num_buf[64];
        if (val == (double)(int64_t)val && fabs(val) < 1e15) {
            snprintf(num_buf, sizeof(num_buf), "%lld", (long long)val);
        } else {
            snprintf(num_buf, sizeof(num_buf), "%g", val);
        }
        size_t n_len = strlen(num_buf);
        while (len + n_len + sep_len + 1 >= cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) { free(buf); return fail("scriptgo string allocation failed"); }
            buf = new_buf;
        }
        if (i > 0) {
            memcpy(buf + len, sep, sep_len);
            len += sep_len;
        }
        memcpy(buf + len, num_buf, n_len);
        len += n_len;
        buf[len] = '\0';
    }
    *out_str = buf;
    return 0;
}

int scriptgo_array_join_string(void *handle, const char *separator, char **out_str) {
    scriptgo_array *array = handle;
    size_t cap = 256, len = 0;
    char *buf;
    const char *sep = separator != NULL ? separator : ",";
    size_t sep_len = strlen(sep);
    if (array == NULL || out_str == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array access failed");
    }
    buf = malloc(cap);
    if (buf == NULL) return fail("scriptgo string allocation failed");
    buf[0] = '\0';
    for (int64_t i = 0; i < array->length; i++) {
        const char *val = *(const char **)(array->data + (size_t)i * sizeof(char *));
        if (val == NULL) val = "";
        size_t s_len = strlen(val);
        while (len + s_len + sep_len + 1 >= cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) { free(buf); return fail("scriptgo string allocation failed"); }
            buf = new_buf;
        }
        if (i > 0) {
            memcpy(buf + len, sep, sep_len);
            len += sep_len;
        }
        memcpy(buf + len, val, s_len);
        len += s_len;
        buf[len] = '\0';
    }
    *out_str = buf;
    return 0;
}

int scriptgo_array_join_bigint(void *handle, const char *separator, char **out_str) {
    scriptgo_array *array = handle;
    size_t cap = 256, len = 0;
    char *buf;
    const char *sep = separator != NULL ? separator : ",";
    size_t sep_len = strlen(sep);
    if (array == NULL || out_str == NULL || array->element_size != sizeof(int64_t)) {
        return fail("scriptgo array access failed");
    }
    buf = malloc(cap);
    if (buf == NULL) return fail("scriptgo string allocation failed");
    buf[0] = '\0';
    for (int64_t i = 0; i < array->length; i++) {
        int64_t val = *(int64_t *)(array->data + (size_t)i * sizeof(int64_t));
        char num_buf[64];
        snprintf(num_buf, sizeof(num_buf), "%lld", (long long)val);
        size_t n_len = strlen(num_buf);
        while (len + n_len + sep_len + 1 >= cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) { free(buf); return fail("scriptgo string allocation failed"); }
            buf = new_buf;
        }
        if (i > 0) {
            memcpy(buf + len, sep, sep_len);
            len += sep_len;
        }
        memcpy(buf + len, num_buf, n_len);
        len += n_len;
        buf[len] = '\0';
    }
    *out_str = buf;
    return 0;
}

typedef struct {
    uint32_t tag;
    uint32_t padding;
    uint64_t payload;
} scriptgo_boxed_unknown_t;

int scriptgo_string_from_unknown(unsigned int tag, unsigned int padding, unsigned long long payload, char **out_str);
int scriptgo_string_from_number(double value, char **out_value);

int scriptgo_array_join_unknown(void *handle, const char *separator, char **out_str) {
    if (handle == NULL || handle == (void *)&scriptgo_undefined_sentinel || out_str == NULL) {
        return fail("scriptgo array access failed");
    }
    if (*(uint64_t *)handle == SCRIPTGO_OBJECT_MAGIC) {
        scriptgo_runtime_object_header *obj = handle;
        size_t cap = 256, len = 0;
        char *buf = malloc(cap);
        if (buf == NULL) return fail("scriptgo string allocation failed");
        buf[0] = '\0';
        const char *sep = separator != NULL ? separator : ",";
        size_t sep_len = strlen(sep);
        for (int64_t i = 0; i < obj->field_count; i++) {
            char *val_str = (char *)obj->fields[i];
            if (val_str == NULL) val_str = "";
            size_t s_len = strlen(val_str);
            while (len + s_len + sep_len + 1 >= cap) {
                cap *= 2;
                char *new_buf = realloc(buf, cap);
                if (new_buf == NULL) { free(buf); return fail("scriptgo string allocation failed"); }
                buf = new_buf;
            }
            if (i > 0) {
                memcpy(buf + len, sep, sep_len);
                len += sep_len;
            }
            memcpy(buf + len, val_str, s_len);
            len += s_len;
            buf[len] = '\0';
        }
        *out_str = buf;
        return 0;
    }
    scriptgo_array *array = handle;
    if (array->element_size <= 0) {
        return fail("scriptgo array access failed");
    }
    size_t cap = 256, len = 0;
    char *buf;
    const char *sep = separator != NULL ? separator : ",";
    size_t sep_len = strlen(sep);
    buf = malloc(cap);
    if (buf == NULL) return fail("scriptgo string allocation failed");
    buf[0] = '\0';
    for (int64_t i = 0; i < array->length; i++) {
        char *val_str = NULL;
        int need_free = 0;
        if (array->element_size == sizeof(scriptgo_boxed_unknown_t)) {
            scriptgo_boxed_unknown_t *item = (scriptgo_boxed_unknown_t *)(array->data + (size_t)i * sizeof(scriptgo_boxed_unknown_t));
            if (scriptgo_string_from_unknown(item->tag, item->padding, item->payload, &val_str) != 0 || val_str == NULL) {
                val_str = "";
            } else {
                need_free = 1;
            }
        } else if (array->element_tag == 3) {
            double d = *(double *)(array->data + (size_t)i * sizeof(double));
            if (scriptgo_string_from_number(d, &val_str) == 0 && val_str != NULL) {
                need_free = 1;
            } else {
                val_str = "";
            }
        } else if (array->element_tag == 4 || array->element_size == sizeof(char *)) {
            val_str = *(char **)(array->data + (size_t)i * sizeof(char *));
            if (val_str == NULL) val_str = "";
        } else {
            val_str = "";
        }
        size_t s_len = strlen(val_str);
        while (len + s_len + sep_len + 1 >= cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) {
                if (need_free) free(val_str);
                free(buf);
                return fail("scriptgo string allocation failed");
            }
            buf = new_buf;
        }
        if (i > 0) {
            memcpy(buf + len, sep, sep_len);
            len += sep_len;
        }
        memcpy(buf + len, val_str, s_len);
        len += s_len;
        buf[len] = '\0';
        if (need_free) {
            free(val_str);
        }
    }
    *out_str = buf;
    return 0;
}

static int cmp_doubles(const void *a, const void *b) {
    double da = *(const double *)a;
    double db = *(const double *)b;
    if (da < db) return -1;
    if (da > db) return 1;
    return 0;
}

static int cmp_strings(const void *a, const void *b) {
    const char *sa = *(const char * const *)a;
    const char *sb = *(const char * const *)b;
    if (sa == NULL) sa = "";
    if (sb == NULL) sb = "";
    return strcmp(sa, sb);
}

int scriptgo_array_fill_number(void *handle, double value, double start_val, double end_val, int32_t has_start, int32_t has_end, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array fill failed");
    }
    int64_t len = array->length;
    int64_t start = 0;
    int64_t end = len;
    if (has_start) {
        start = (int64_t)start_val;
        if (start < 0) { start = len + start; if (start < 0) start = 0; }
        else if (start > len) start = len;
    }
    if (has_end) {
        end = (int64_t)end_val;
        if (end < 0) { end = len + end; if (end < 0) end = 0; }
        else if (end > len) end = len;
    }
    for (int64_t i = start; i < end; i++) {
        *(double *)(array->data + (size_t)i * sizeof(double)) = value;
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_fill_string(void *handle, const char *value, double start_val, double end_val, int32_t has_start, int32_t has_end, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array fill failed");
    }
    int64_t len = array->length;
    int64_t start = 0;
    int64_t end = len;
    if (has_start) {
        start = (int64_t)start_val;
        if (start < 0) { start = len + start; if (start < 0) start = 0; }
        else if (start > len) start = len;
    }
    if (has_end) {
        end = (int64_t)end_val;
        if (end < 0) { end = len + end; if (end < 0) end = 0; }
        else if (end > len) end = len;
    }
    for (int64_t i = start; i < end; i++) {
        *(const char **)(array->data + (size_t)i * sizeof(char *)) = value;
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_to_reversed(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size <= 0) {
        return fail("scriptgo array toReversed failed");
    }
    if (scriptgo_array_new(array->length, array->element_size, out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    size_t elem_sz = (size_t)array->element_size;
    for (int64_t i = 0; i < array->length; i++) {
        memcpy(res->data + (size_t)i * elem_sz, array->data + (size_t)(array->length - 1 - i) * elem_sz, elem_sz);
    }
    return 0;
}

int scriptgo_array_to_sorted_number(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array toSorted failed");
    }
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    if (array->length > 0) {
        memcpy(res->data, array->data, (size_t)array->length * sizeof(double));
        qsort(res->data, (size_t)res->length, sizeof(double), cmp_doubles);
    }
    return 0;
}

int scriptgo_array_to_sorted_string(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array toSorted failed");
    }
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    if (array->length > 0) {
        memcpy(res->data, array->data, (size_t)array->length * sizeof(char *));
        qsort(res->data, (size_t)res->length, sizeof(char *), cmp_strings);
    }
    return 0;
}

int scriptgo_array_last_index_of_number(void *handle, double target, double from_index, double *out_index) {
    scriptgo_array *array = handle;
    int64_t start, i;
    if (array == NULL || out_index == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array access failed");
    }
    start = array->length - 1;
    if (from_index >= 0.0) {
        start = (int64_t)from_index;
        if (start >= array->length) start = array->length - 1;
    } else {
        start = array->length + (int64_t)from_index;
    }
    for (i = start; i >= 0; i--) {
        double val = *(double *)(array->data + (size_t)i * sizeof(double));
        if (val == target) {
            *out_index = (double)i;
            return 0;
        }
    }
    *out_index = -1.0;
    return 0;
}

int scriptgo_array_last_index_of_string(void *handle, const char *target, double from_index, double *out_index) {
    scriptgo_array *array = handle;
    int64_t start, i;
    if (array == NULL || target == NULL || out_index == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array access failed");
    }
    start = array->length - 1;
    if (from_index >= 0.0) {
        start = (int64_t)from_index;
        if (start >= array->length) start = array->length - 1;
    } else {
        start = array->length + (int64_t)from_index;
    }
    for (i = start; i >= 0; i--) {
        const char *val = *(const char **)(array->data + (size_t)i * sizeof(char *));
        if (val != NULL && strcmp(val, target) == 0) {
            *out_index = (double)i;
            return 0;
        }
    }
    *out_index = -1.0;
    return 0;
}

int scriptgo_array_copy_within(void *handle, double target_val, double start_val, double end_val, int32_t has_start, int32_t has_end, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size <= 0) {
        return fail("scriptgo array copyWithin failed");
    }
    int64_t len = array->length;
    int64_t target = (int64_t)target_val;
    if (target < 0) { target = len + target; if (target < 0) target = 0; }
    else if (target > len) target = len;

    int64_t start = 0;
    if (has_start) {
        start = (int64_t)start_val;
        if (start < 0) { start = len + start; if (start < 0) start = 0; }
        else if (start > len) start = len;
    }
    int64_t end = len;
    if (has_end) {
        end = (int64_t)end_val;
        if (end < 0) { end = len + end; if (end < 0) end = 0; }
        else if (end > len) end = len;
    }
    int64_t count = end - start;
    if (count > len - target) count = len - target;
    if (count > 0 && start < len) {
        size_t elem_sz = (size_t)array->element_size;
        memmove(array->data + (size_t)target * elem_sz, array->data + (size_t)start * elem_sz, (size_t)count * elem_sz);
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_with_number(void *handle, double index, double value, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array with failed");
    }
    int64_t idx = (int64_t)index;
    if (idx < 0) idx = array->length + idx;
    if (idx < 0 || idx >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    memcpy(res->data, array->data, (size_t)array->length * sizeof(double));
    *(double *)(res->data + (size_t)idx * sizeof(double)) = value;
    return 0;
}

int scriptgo_array_with_string(void *handle, double index, const char *value, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array with failed");
    }
    int64_t idx = (int64_t)index;
    if (idx < 0) idx = array->length + idx;
    if (idx < 0 || idx >= array->length) {
        return fail("scriptgo array index out of bounds");
    }
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    memcpy(res->data, array->data, (size_t)array->length * sizeof(char *));
    *(const char **)(res->data + (size_t)idx * sizeof(char *)) = value;
    return 0;
}

int scriptgo_array_to_spliced(void *handle, double start_val, double delete_count_val, int32_t has_delete_count, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL || array->element_size <= 0) {
        return fail("scriptgo array toSpliced failed");
    }
    int64_t length = array->length;
    int64_t start, delete_count;
    if (start_val < 0.0) {
        start = length + (int64_t)start_val;
        if (start < 0) start = 0;
    } else {
        start = (int64_t)start_val;
        if (start > length) start = length;
    }
    if (!has_delete_count) {
        delete_count = length - start;
    } else if (delete_count_val < 0.0) {
        delete_count = 0;
    } else {
        delete_count = (int64_t)delete_count_val;
        if (start + delete_count > length) {
            delete_count = length - start;
        }
    }
    int64_t new_len = length - delete_count;
    if (scriptgo_array_new(new_len, array->element_size, out_array) != 0) {
        return -1;
    }
    scriptgo_array *res = *out_array;
    size_t elem_sz = (size_t)array->element_size;
    if (start > 0) {
        memcpy(res->data, array->data, (size_t)start * elem_sz);
    }
    int64_t remaining = length - (start + delete_count);
    if (remaining > 0) {
        memcpy(res->data + (size_t)start * elem_sz, array->data + (size_t)(start + delete_count) * elem_sz, (size_t)remaining * elem_sz);
    }
    return 0;
}

int scriptgo_array_sort_number(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size != sizeof(double)) {
        return fail("scriptgo array sort failed");
    }
    if (array->length > 0) {
        qsort(array->data, (size_t)array->length, sizeof(double), cmp_doubles);
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_sort_string(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || array->element_size != sizeof(char *)) {
        return fail("scriptgo array sort failed");
    }
    if (array->length > 0) {
        qsort(array->data, (size_t)array->length, sizeof(char *), cmp_strings);
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_is_array(void *handle, double *out_bool) {
    if (out_bool == NULL) return fail("scriptgo isArray failed");
    *out_bool = handle != NULL ? 1.0 : 0.0;
    return 0;
}

int scriptgo_array_keys(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL) return fail("scriptgo array keys failed");
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) return -1;
    scriptgo_array *res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        double d = (double)i;
        memcpy(res->data + (size_t)i * sizeof(double), &d, sizeof(double));
    }
    return 0;
}

int scriptgo_array_entries(void *handle, void **out_array) {
    scriptgo_array *array = handle;
    if (array == NULL || out_array == NULL) return fail("scriptgo array entries failed");
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) return -1;
    scriptgo_array *res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        char buf[256];
        if (array->element_size == sizeof(char *)) {
            const char *val = *(const char **)(array->data + (size_t)i * sizeof(char *));
            snprintf(buf, sizeof(buf), "[ %lld, '%s' ]", (long long)i, val ? val : "undefined");
        } else if (array->element_size == sizeof(double)) {
            double val = *(double *)(array->data + (size_t)i * sizeof(double));
            if (val == (double)(int64_t)val) {
                snprintf(buf, sizeof(buf), "[ %lld, %lld ]", (long long)i, (long long)val);
            } else {
                snprintf(buf, sizeof(buf), "[ %lld, %g ]", (long long)i, val);
            }
        } else {
            snprintf(buf, sizeof(buf), "[ %lld, <item> ]", (long long)i);
        }
        char *entry_str = strdup(buf);
        memcpy(res->data + (size_t)i * sizeof(char *), &entry_str, sizeof(char *));
    }
    return 0;
}
