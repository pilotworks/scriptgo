#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct {
    int64_t byte_length;
    unsigned char *data;
} scriptgo_array_buffer;

typedef enum {
    SCRIPTGO_TYPEDARRAY_UINT8 = 1,
    SCRIPTGO_TYPEDARRAY_INT32 = 2,
    SCRIPTGO_TYPEDARRAY_FLOAT64 = 3
} scriptgo_typedarray_kind;

typedef struct {
    scriptgo_typedarray_kind kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    scriptgo_array_buffer *buffer;
    unsigned char *data;
} scriptgo_typed_array;

static int typedarray_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

int scriptgo_arraybuffer_new(int64_t byte_length, void **out_buffer) {
    if (out_buffer == NULL || byte_length < 0) {
        return typedarray_fail("scriptgo ArrayBuffer invalid length");
    }
    scriptgo_array_buffer *buf = calloc(1, sizeof(scriptgo_array_buffer));
    if (buf == NULL) return typedarray_fail("scriptgo ArrayBuffer allocation failed");
    buf->byte_length = byte_length;
    if (byte_length > 0) {
        buf->data = calloc(1, (size_t)byte_length);
        if (buf->data == NULL) {
            free(buf);
            return typedarray_fail("scriptgo ArrayBuffer allocation failed");
        }
    }
    *out_buffer = buf;
    return 0;
}

int scriptgo_arraybuffer_slice(void *buffer_handle, double begin, double end, void **out_buffer) {
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)buffer_handle;
    if (buf == NULL || out_buffer == NULL) return typedarray_fail("scriptgo ArrayBuffer slice failed");
    int64_t len = buf->byte_length;
    int64_t start_idx = (int64_t)begin;
    int64_t end_idx = (int64_t)end;
    if (start_idx < 0) {
        start_idx += len;
        if (start_idx < 0) start_idx = 0;
    } else if (start_idx > len) {
        start_idx = len;
    }
    if (end_idx < 0) {
        end_idx += len;
        if (end_idx < 0) end_idx = 0;
    } else if (end_idx > len) {
        end_idx = len;
    }
    int64_t new_len = end_idx > start_idx ? (end_idx - start_idx) : 0;
    void *new_buf_handle = NULL;
    if (scriptgo_arraybuffer_new(new_len, &new_buf_handle) != 0) return -1;
    scriptgo_array_buffer *new_buf = (scriptgo_array_buffer *)new_buf_handle;
    if (new_len > 0 && buf->data != NULL) {
        memcpy(new_buf->data, buf->data + start_idx, (size_t)new_len);
    }
    *out_buffer = new_buf;
    return 0;
}

int scriptgo_arraybuffer_byte_length(void *buffer_handle, double *out_len) {
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)buffer_handle;
    if (buf == NULL || out_len == NULL) return typedarray_fail("scriptgo ArrayBuffer byteLength failed");
    *out_len = (double)buf->byte_length;
    return 0;
}

int scriptgo_typedarray_new(int64_t kind, int64_t length, void *buffer_handle, int64_t byte_offset, void **out_array) {
    if (out_array == NULL || length < 0 || byte_offset < 0) {
        return typedarray_fail("scriptgo TypedArray allocation failed");
    }
    int64_t elem_size = 1;
    if (kind == SCRIPTGO_TYPEDARRAY_INT32) {
        elem_size = 4;
    } else if (kind == SCRIPTGO_TYPEDARRAY_FLOAT64) {
        elem_size = 8;
    }
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)buffer_handle;
    if (buf == NULL) {
        int64_t total_bytes = length * elem_size;
        void *created_buf = NULL;
        if (scriptgo_arraybuffer_new(total_bytes, &created_buf) != 0) return -1;
        buf = (scriptgo_array_buffer *)created_buf;
        byte_offset = 0;
    } else {
        if (byte_offset + length * elem_size > buf->byte_length) {
            return typedarray_fail("scriptgo TypedArray range out of bounds");
        }
    }
    scriptgo_typed_array *ta = calloc(1, sizeof(scriptgo_typed_array));
    if (ta == NULL) return typedarray_fail("scriptgo TypedArray allocation failed");
    ta->kind = (scriptgo_typedarray_kind)kind;
    ta->length = length;
    ta->byte_offset = byte_offset;
    ta->element_size = elem_size;
    ta->buffer = buf;
    ta->data = buf->data != NULL ? (buf->data + byte_offset) : NULL;
    *out_array = ta;
    return 0;
}

int scriptgo_typedarray_get(void *handle, double index, double *out_value) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_value == NULL) return typedarray_fail("scriptgo TypedArray get failed");
    if (index != index || index < 0 || index != (double)(int64_t)index || (int64_t)index >= ta->length) {
        *out_value = 0.0;
        return 0;
    }
    int64_t idx = (int64_t)index;
    if (ta->data == NULL) {
        *out_value = 0.0;
        return 0;
    }
    switch (ta->kind) {
    case SCRIPTGO_TYPEDARRAY_UINT8:
        *out_value = (double)ta->data[idx];
        break;
    case SCRIPTGO_TYPEDARRAY_INT32: {
        int32_t val;
        memcpy(&val, ta->data + (idx * 4), sizeof(int32_t));
        *out_value = (double)val;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_FLOAT64: {
        double val;
        memcpy(&val, ta->data + (idx * 8), sizeof(double));
        *out_value = val;
        break;
    }
    default:
        return typedarray_fail("unsupported typedarray kind");
    }
    return 0;
}

int scriptgo_typedarray_set(void *handle, double index, double value) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL) return typedarray_fail("scriptgo TypedArray set failed");
    if (index != index || index < 0 || index != (double)(int64_t)index || (int64_t)index >= ta->length) {
        return 0;
    }
    int64_t idx = (int64_t)index;
    if (ta->data == NULL) return 0;
    switch (ta->kind) {
    case SCRIPTGO_TYPEDARRAY_UINT8: {
        uint8_t u8 = (uint8_t)(uint32_t)value;
        ta->data[idx] = u8;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_INT32: {
        int32_t i32 = (int32_t)value;
        memcpy(ta->data + (idx * 4), &i32, sizeof(int32_t));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_FLOAT64: {
        memcpy(ta->data + (idx * 8), &value, sizeof(double));
        break;
    }
    default:
        return typedarray_fail("unsupported typedarray kind");
    }
    return 0;
}

int scriptgo_typedarray_length(void *handle, double *out_length) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_length == NULL) return typedarray_fail("scriptgo TypedArray length failed");
    *out_length = (double)ta->length;
    return 0;
}

int scriptgo_typedarray_byte_length(void *handle, double *out_byte_length) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_byte_length == NULL) return typedarray_fail("scriptgo TypedArray byteLength failed");
    *out_byte_length = (double)(ta->length * ta->element_size);
    return 0;
}

int scriptgo_typedarray_byte_offset(void *handle, double *out_byte_offset) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_byte_offset == NULL) return typedarray_fail("scriptgo TypedArray byteOffset failed");
    *out_byte_offset = (double)ta->byte_offset;
    return 0;
}

int scriptgo_typedarray_buffer(void *handle, void **out_buffer) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_buffer == NULL) return typedarray_fail("scriptgo TypedArray buffer failed");
    *out_buffer = ta->buffer;
    return 0;
}

int scriptgo_typedarray_subarray(void *handle, double begin, double end, void **out_array) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_array == NULL) return typedarray_fail("scriptgo TypedArray subarray failed");
    int64_t len = ta->length;
    int64_t start_idx = (int64_t)begin;
    int64_t end_idx = (int64_t)end;
    if (start_idx < 0) {
        start_idx += len;
        if (start_idx < 0) start_idx = 0;
    } else if (start_idx > len) {
        start_idx = len;
    }
    if (end_idx < 0) {
        end_idx += len;
        if (end_idx < 0) end_idx = 0;
    } else if (end_idx > len) {
        end_idx = len;
    }
    int64_t sub_len = end_idx > start_idx ? (end_idx - start_idx) : 0;
    int64_t new_offset = ta->byte_offset + (start_idx * ta->element_size);
    return scriptgo_typedarray_new((int64_t)ta->kind, sub_len, ta->buffer, new_offset, out_array);
}

int scriptgo_typedarray_slice(void *handle, double begin, double end, void **out_array) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_array == NULL) return typedarray_fail("scriptgo TypedArray slice failed");
    void *sub_handle = NULL;
    if (scriptgo_typedarray_subarray(handle, begin, end, &sub_handle) != 0) return -1;
    scriptgo_typed_array *sub = (scriptgo_typed_array *)sub_handle;
    void *new_ta_handle = NULL;
    if (scriptgo_typedarray_new((int64_t)ta->kind, sub->length, NULL, 0, &new_ta_handle) != 0) return -1;
    scriptgo_typed_array *new_ta = (scriptgo_typed_array *)new_ta_handle;
    if (sub->length > 0 && sub->data != NULL && new_ta->data != NULL) {
        memcpy(new_ta->data, sub->data, (size_t)(sub->length * sub->element_size));
    }
    free(sub);
    *out_array = new_ta;
    return 0;
}

int scriptgo_typedarray_from_array(int64_t kind, void *array_handle, void **out_array) {
    if (array_handle == NULL || out_array == NULL) return typedarray_fail("scriptgo TypedArray from array failed");
    // Handle standard scriptgo_array (struct with length and data)
    typedef struct {
        int64_t length;
        int64_t capacity;
        int64_t element_size;
        unsigned char *data;
    } temp_arr_header;
    temp_arr_header *arr = (temp_arr_header *)array_handle;
    void *new_ta_handle = NULL;
    if (scriptgo_typedarray_new(kind, arr->length, NULL, 0, &new_ta_handle) != 0) return -1;
    scriptgo_typed_array *ta = (scriptgo_typed_array *)new_ta_handle;
    for (int64_t i = 0; i < arr->length; i++) {
        double val = 0.0;
        if (arr->element_size == sizeof(double)) {
            memcpy(&val, arr->data + i * sizeof(double), sizeof(double));
        }
        scriptgo_typedarray_set(ta, (double)i, val);
    }
    *out_array = ta;
    return 0;
}

int scriptgo_typedarray_set_array(void *target_handle, void *src_handle, double offset_d) {
    scriptgo_typed_array *target = (scriptgo_typed_array *)target_handle;
    scriptgo_typed_array *src = (scriptgo_typed_array *)src_handle;
    if (target == NULL || src == NULL) return typedarray_fail("scriptgo TypedArray set failed");
    int64_t offset = (int64_t)offset_d;
    if (offset < 0 || offset + src->length > target->length) {
        return typedarray_fail("scriptgo TypedArray set offset out of bounds");
    }
    for (int64_t i = 0; i < src->length; i++) {
        double val = 0.0;
        scriptgo_typedarray_get(src, (double)i, &val);
        scriptgo_typedarray_set(target, (double)(offset + i), val);
    }
    return 0;
}

int scriptgo_typedarray_fill(void *handle, double value, double start_d, double end_d) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL) return typedarray_fail("scriptgo TypedArray fill failed");
    int64_t len = ta->length;
    int64_t start_idx = (int64_t)start_d;
    int64_t end_idx = (int64_t)end_d;
    if (start_idx < 0) {
        start_idx += len;
        if (start_idx < 0) start_idx = 0;
    } else if (start_idx > len) {
        start_idx = len;
    }
    if (end_idx < 0) {
        end_idx += len;
        if (end_idx < 0) end_idx = 0;
    } else if (end_idx > len) {
        end_idx = len;
    }
    for (int64_t i = start_idx; i < end_idx; i++) {
        scriptgo_typedarray_set(ta, (double)i, value);
    }
    return 0;
}

int scriptgo_typedarray_to_string(void *handle, char **out_str) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_str == NULL) return typedarray_fail("scriptgo TypedArray toString failed");
    const char *name = "Uint8Array";
    if (ta->kind == SCRIPTGO_TYPEDARRAY_INT32) name = "Int32Array";
    else if (ta->kind == SCRIPTGO_TYPEDARRAY_FLOAT64) name = "Float64Array";
    
    size_t cap = 64 + (size_t)ta->length * 24;
    char *buf = malloc(cap);
    if (buf == NULL) return typedarray_fail("scriptgo TypedArray toString allocation failed");
    int offset = snprintf(buf, cap, "%s(%lld) [ ", name, (long long)ta->length);
    for (int64_t i = 0; i < ta->length; i++) {
        double val = 0.0;
        scriptgo_typedarray_get(ta, (double)i, &val);
        if (i > 0) {
            offset += snprintf(buf + offset, cap - (size_t)offset, ", ");
        }
        if (val == (double)(long long)val) {
            offset += snprintf(buf + offset, cap - (size_t)offset, "%lld", (long long)val);
        } else {
            offset += snprintf(buf + offset, cap - (size_t)offset, "%.15g", val);
        }
    }
    snprintf(buf + offset, cap - (size_t)offset, " ]");
    *out_str = buf;
    return 0;
}

int scriptgo_arraybuffer_to_string(void *handle, char **out_str) {
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)handle;
    if (buf == NULL || out_str == NULL) return typedarray_fail("scriptgo ArrayBuffer toString failed");
    char *res = malloc(64);
    if (res == NULL) return typedarray_fail("scriptgo ArrayBuffer toString allocation failed");
    snprintf(res, 64, "ArrayBuffer { byteLength: %lld }", (long long)buf->byte_length);
    *out_str = res;
    return 0;
}

