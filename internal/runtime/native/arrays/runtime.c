#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
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
    array->owned_data = NULL;
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


