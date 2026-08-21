#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252 // "TARR"
#define SCRIPTGO_MAGIC_DATAVIEW   0x44564957 // "DVIW"

typedef struct {
    int64_t byte_length;
    unsigned char *data;
} scriptgo_array_buffer;

typedef enum {
    SCRIPTGO_TYPEDARRAY_INT8 = 1,
    SCRIPTGO_TYPEDARRAY_UINT8 = 2,
    SCRIPTGO_TYPEDARRAY_UINT8_CLAMPED = 3,
    SCRIPTGO_TYPEDARRAY_INT16 = 4,
    SCRIPTGO_TYPEDARRAY_UINT16 = 5,
    SCRIPTGO_TYPEDARRAY_INT32 = 6,
    SCRIPTGO_TYPEDARRAY_UINT32 = 7,
    SCRIPTGO_TYPEDARRAY_FLOAT32 = 8,
    SCRIPTGO_TYPEDARRAY_FLOAT64 = 9,
    SCRIPTGO_TYPEDARRAY_BIGINT64 = 10,
    SCRIPTGO_TYPEDARRAY_BIGUINT64 = 11
} scriptgo_typedarray_kind;

typedef struct {
    uint32_t magic;
    scriptgo_typedarray_kind kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    scriptgo_array_buffer *buffer;
    unsigned char *data;
} scriptgo_typed_array;

typedef struct {
    uint32_t magic;
    scriptgo_array_buffer *buffer;
    int64_t byte_offset;
    int64_t byte_length;
} scriptgo_data_view;

static int typedarray_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static inline uint16_t swap16_if_be(uint16_t v, int is_le) {
    return is_le ? v : (uint16_t)((v >> 8) | (v << 8));
}
static inline uint32_t swap32_if_be(uint32_t v, int is_le) {
    return is_le ? v : (((v & 0xff000000) >> 24) | ((v & 0x00ff0000) >> 8) | ((v & 0x0000ff00) << 8) | ((v & 0x000000ff) << 24));
}
static inline uint64_t swap64_if_be(uint64_t v, int is_le) {
    return is_le ? v : (((v & 0xff00000000000000ULL) >> 56) |
                        ((v & 0x00ff000000000000ULL) >> 40) |
                        ((v & 0x0000ff0000000000ULL) >> 24) |
                        ((v & 0x000000ff00000000ULL) >> 8) |
                        ((v & 0x00000000ff000000ULL) << 8) |
                        ((v & 0x0000000000ff0000ULL) << 24) |
                        ((v & 0x000000000000ff00ULL) << 40) |
                        ((v & 0x00000000000000ffULL) << 56));
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

int scriptgo_arraybuffer_is_view(void *handle, int32_t *out_is_view) {
    if (out_is_view == NULL) return typedarray_fail("scriptgo ArrayBuffer.isView invalid args");
    if (handle == NULL) {
        *out_is_view = 0;
        return 0;
    }
    uint32_t magic = *(uint32_t *)handle;
    if (magic == SCRIPTGO_MAGIC_TYPEDARRAY || magic == SCRIPTGO_MAGIC_DATAVIEW) {
        *out_is_view = 1;
    } else {
        *out_is_view = 0;
    }
    return 0;
}

static int64_t typedarray_element_size(scriptgo_typedarray_kind kind) {
    switch (kind) {
    case SCRIPTGO_TYPEDARRAY_INT8:
    case SCRIPTGO_TYPEDARRAY_UINT8:
    case SCRIPTGO_TYPEDARRAY_UINT8_CLAMPED:
        return 1;
    case SCRIPTGO_TYPEDARRAY_INT16:
    case SCRIPTGO_TYPEDARRAY_UINT16:
        return 2;
    case SCRIPTGO_TYPEDARRAY_INT32:
    case SCRIPTGO_TYPEDARRAY_UINT32:
    case SCRIPTGO_TYPEDARRAY_FLOAT32:
        return 4;
    case SCRIPTGO_TYPEDARRAY_FLOAT64:
    case SCRIPTGO_TYPEDARRAY_BIGINT64:
    case SCRIPTGO_TYPEDARRAY_BIGUINT64:
        return 8;
    default:
        return 1;
    }
}

int scriptgo_typedarray_new(int64_t kind, int64_t length, void *buffer_handle, int64_t byte_offset, void **out_array) {
    if (out_array == NULL || length < 0 || byte_offset < 0) {
        return typedarray_fail("scriptgo TypedArray allocation failed");
    }
    int64_t elem_size = typedarray_element_size((scriptgo_typedarray_kind)kind);
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)buffer_handle;
    if (buf == NULL) {
        int64_t total_bytes = length * elem_size;
        void *created_buf = NULL;
        if (scriptgo_arraybuffer_new(total_bytes, &created_buf) != 0) return -1;
        buf = (scriptgo_array_buffer *)created_buf;
        byte_offset = 0;
    } else {
        if (length == 0 && byte_offset < buf->byte_length) {
            length = (buf->byte_length - byte_offset) / elem_size;
        }
        if (byte_offset + length * elem_size > buf->byte_length) {
            return typedarray_fail("scriptgo TypedArray range out of bounds");
        }
    }
    scriptgo_typed_array *ta = calloc(1, sizeof(scriptgo_typed_array));
    if (ta == NULL) return typedarray_fail("scriptgo TypedArray allocation failed");
    ta->magic = SCRIPTGO_MAGIC_TYPEDARRAY;
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
    case SCRIPTGO_TYPEDARRAY_INT8:
        *out_value = (double)(int8_t)ta->data[idx];
        break;
    case SCRIPTGO_TYPEDARRAY_UINT8:
    case SCRIPTGO_TYPEDARRAY_UINT8_CLAMPED:
        *out_value = (double)ta->data[idx];
        break;
    case SCRIPTGO_TYPEDARRAY_INT16: {
        int16_t val;
        memcpy(&val, ta->data + (idx * 2), sizeof(int16_t));
        *out_value = (double)val;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT16: {
        uint16_t val;
        memcpy(&val, ta->data + (idx * 2), sizeof(uint16_t));
        *out_value = (double)val;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_INT32: {
        int32_t val;
        memcpy(&val, ta->data + (idx * 4), sizeof(int32_t));
        *out_value = (double)val;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT32: {
        uint32_t val;
        memcpy(&val, ta->data + (idx * 4), sizeof(uint32_t));
        *out_value = (double)val;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_FLOAT32: {
        float val;
        memcpy(&val, ta->data + (idx * 4), sizeof(float));
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
        *out_value = 0.0;
        break;
    }
    return 0;
}

int scriptgo_typedarray_get_bigint(void *handle, double index, int64_t *out_value) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_value == NULL) return typedarray_fail("scriptgo TypedArray get bigint failed");
    if (index != index || index < 0 || index != (double)(int64_t)index || (int64_t)index >= ta->length) {
        *out_value = 0;
        return 0;
    }
    int64_t idx = (int64_t)index;
    if (ta->data == NULL) {
        *out_value = 0;
        return 0;
    }
    int64_t val;
    memcpy(&val, ta->data + (idx * 8), sizeof(int64_t));
    *out_value = val;
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
    case SCRIPTGO_TYPEDARRAY_INT8: {
        int8_t i8 = (int8_t)(int32_t)value;
        ta->data[idx] = (uint8_t)i8;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT8: {
        uint8_t u8 = (uint8_t)(uint32_t)value;
        ta->data[idx] = u8;
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT8_CLAMPED: {
        if (value != value || value <= 0.0) {
            ta->data[idx] = 0;
        } else if (value >= 255.0) {
            ta->data[idx] = 255;
        } else {
            ta->data[idx] = (uint8_t)rint(value);
        }
        break;
    }
    case SCRIPTGO_TYPEDARRAY_INT16: {
        int16_t i16 = (int16_t)(int32_t)value;
        memcpy(ta->data + (idx * 2), &i16, sizeof(int16_t));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT16: {
        uint16_t u16 = (uint16_t)(uint32_t)value;
        memcpy(ta->data + (idx * 2), &u16, sizeof(uint16_t));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_INT32: {
        int32_t i32 = (int32_t)value;
        memcpy(ta->data + (idx * 4), &i32, sizeof(int32_t));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_UINT32: {
        uint32_t u32 = (uint32_t)value;
        memcpy(ta->data + (idx * 4), &u32, sizeof(uint32_t));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_FLOAT32: {
        float f32 = (float)value;
        memcpy(ta->data + (idx * 4), &f32, sizeof(float));
        break;
    }
    case SCRIPTGO_TYPEDARRAY_FLOAT64: {
        memcpy(ta->data + (idx * 8), &value, sizeof(double));
        break;
    }
    default:
        break;
    }
    return 0;
}

int scriptgo_typedarray_set_bigint(void *handle, double index, int64_t value) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL) return typedarray_fail("scriptgo TypedArray set bigint failed");
    if (index != index || index < 0 || index != (double)(int64_t)index || (int64_t)index >= ta->length) {
        return 0;
    }
    int64_t idx = (int64_t)index;
    if (ta->data == NULL) return 0;
    memcpy(ta->data + (idx * 8), &value, sizeof(int64_t));
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
        if (kind == SCRIPTGO_TYPEDARRAY_BIGINT64 || kind == SCRIPTGO_TYPEDARRAY_BIGUINT64) {
            int64_t val = 0;
            if (arr->element_size == sizeof(int64_t)) {
                memcpy(&val, arr->data + i * sizeof(int64_t), sizeof(int64_t));
            }
            scriptgo_typedarray_set_bigint(ta, (double)i, val);
        } else {
            double val = 0.0;
            if (arr->element_size == sizeof(double)) {
                memcpy(&val, arr->data + i * sizeof(double), sizeof(double));
            }
            scriptgo_typedarray_set(ta, (double)i, val);
        }
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
        if (target->kind == SCRIPTGO_TYPEDARRAY_BIGINT64 || target->kind == SCRIPTGO_TYPEDARRAY_BIGUINT64) {
            int64_t val = 0;
            scriptgo_typedarray_get_bigint(src, (double)i, &val);
            scriptgo_typedarray_set_bigint(target, (double)(offset + i), val);
        } else {
            double val = 0.0;
            scriptgo_typedarray_get(src, (double)i, &val);
            scriptgo_typedarray_set(target, (double)(offset + i), val);
        }
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
        if (ta->kind == SCRIPTGO_TYPEDARRAY_BIGINT64 || ta->kind == SCRIPTGO_TYPEDARRAY_BIGUINT64) {
            scriptgo_typedarray_set_bigint(ta, (double)i, (int64_t)value);
        } else {
            scriptgo_typedarray_set(ta, (double)i, value);
        }
    }
    return 0;
}

static const char *typedarray_kind_name(scriptgo_typedarray_kind kind) {
    switch (kind) {
    case SCRIPTGO_TYPEDARRAY_INT8: return "Int8Array";
    case SCRIPTGO_TYPEDARRAY_UINT8: return "Uint8Array";
    case SCRIPTGO_TYPEDARRAY_UINT8_CLAMPED: return "Uint8ClampedArray";
    case SCRIPTGO_TYPEDARRAY_INT16: return "Int16Array";
    case SCRIPTGO_TYPEDARRAY_UINT16: return "Uint16Array";
    case SCRIPTGO_TYPEDARRAY_INT32: return "Int32Array";
    case SCRIPTGO_TYPEDARRAY_UINT32: return "Uint32Array";
    case SCRIPTGO_TYPEDARRAY_FLOAT32: return "Float32Array";
    case SCRIPTGO_TYPEDARRAY_FLOAT64: return "Float64Array";
    case SCRIPTGO_TYPEDARRAY_BIGINT64: return "BigInt64Array";
    case SCRIPTGO_TYPEDARRAY_BIGUINT64: return "BigUint64Array";
    default: return "TypedArray";
    }
}

int scriptgo_typedarray_to_string(void *handle, char **out_str) {
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta == NULL || out_str == NULL) return typedarray_fail("scriptgo TypedArray toString failed");
    const char *name = typedarray_kind_name(ta->kind);
    size_t cap = 64 + (size_t)ta->length * 28;
    char *buf = malloc(cap);
    if (buf == NULL) return typedarray_fail("scriptgo TypedArray toString allocation failed");
    int offset = snprintf(buf, cap, "%s(%lld) [ ", name, (long long)ta->length);
    for (int64_t i = 0; i < ta->length; i++) {
        if (i > 0) {
            offset += snprintf(buf + offset, cap - (size_t)offset, ", ");
        }
        if (ta->kind == SCRIPTGO_TYPEDARRAY_BIGINT64 || ta->kind == SCRIPTGO_TYPEDARRAY_BIGUINT64) {
            int64_t bval = 0;
            scriptgo_typedarray_get_bigint(ta, (double)i, &bval);
            offset += snprintf(buf + offset, cap - (size_t)offset, "%lldn", (long long)bval);
        } else {
            double val = 0.0;
            scriptgo_typedarray_get(ta, (double)i, &val);
            if (val == (double)(long long)val) {
                offset += snprintf(buf + offset, cap - (size_t)offset, "%lld", (long long)val);
            } else {
                offset += snprintf(buf + offset, cap - (size_t)offset, "%.15g", val);
            }
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

// -----------------------------------------------------------------------------
// DataView Implementation
// -----------------------------------------------------------------------------

int scriptgo_dataview_new(void *buffer_handle, double byte_offset_d, double byte_length_d, void **out_view) {
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)buffer_handle;
    if (buf == NULL || out_view == NULL) return typedarray_fail("scriptgo DataView new failed");
    int64_t byte_offset = (int64_t)byte_offset_d;
    if (byte_offset < 0 || byte_offset > buf->byte_length) {
        return typedarray_fail("scriptgo DataView byteOffset out of bounds");
    }
    int64_t byte_length;
    if (byte_length_d == 0.0 && byte_offset == 0) {
        byte_length = buf->byte_length;
    } else if (byte_length_d == 0.0) {
        byte_length = buf->byte_length - byte_offset;
    } else {
        byte_length = (int64_t)byte_length_d;
    }
    if (byte_length < 0 || byte_offset + byte_length > buf->byte_length) {
        return typedarray_fail("scriptgo DataView byteLength out of bounds");
    }
    scriptgo_data_view *dv = calloc(1, sizeof(scriptgo_data_view));
    if (dv == NULL) return typedarray_fail("scriptgo DataView allocation failed");
    dv->magic = SCRIPTGO_MAGIC_DATAVIEW;
    dv->buffer = buf;
    dv->byte_offset = byte_offset;
    dv->byte_length = byte_length;
    *out_view = dv;
    return 0;
}

int scriptgo_dataview_byte_length(void *handle, double *out_len) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    if (dv == NULL || out_len == NULL) return typedarray_fail("scriptgo DataView byteLength failed");
    *out_len = (double)dv->byte_length;
    return 0;
}

int scriptgo_dataview_byte_offset(void *handle, double *out_offset) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    if (dv == NULL || out_offset == NULL) return typedarray_fail("scriptgo DataView byteOffset failed");
    *out_offset = (double)dv->byte_offset;
    return 0;
}

int scriptgo_dataview_buffer(void *handle, void **out_buffer) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    if (dv == NULL || out_buffer == NULL) return typedarray_fail("scriptgo DataView buffer failed");
    *out_buffer = dv->buffer;
    return 0;
}

static inline unsigned char *dataview_ptr(scriptgo_data_view *dv, double offset_d, int64_t size) {
    if (dv == NULL || dv->buffer == NULL || dv->buffer->data == NULL) return NULL;
    int64_t off = (int64_t)offset_d;
    if (off < 0 || off + size > dv->byte_length) return NULL;
    return dv->buffer->data + dv->byte_offset + off;
}

int scriptgo_dataview_get_int8(void *handle, double byte_offset, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 1);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getInt8 range out of bounds");
    *out_val = (double)(int8_t)*p;
    return 0;
}

int scriptgo_dataview_set_int8(void *handle, double byte_offset, double value) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 1);
    if (p == NULL) return typedarray_fail("scriptgo DataView setInt8 range out of bounds");
    *p = (uint8_t)(int8_t)(int32_t)value;
    return 0;
}

int scriptgo_dataview_get_uint8(void *handle, double byte_offset, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 1);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getUint8 range out of bounds");
    *out_val = (double)*p;
    return 0;
}

int scriptgo_dataview_set_uint8(void *handle, double byte_offset, double value) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 1);
    if (p == NULL) return typedarray_fail("scriptgo DataView setUint8 range out of bounds");
    *p = (uint8_t)(uint32_t)value;
    return 0;
}

int scriptgo_dataview_get_int16(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 2);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getInt16 range out of bounds");
    uint16_t u;
    memcpy(&u, p, 2);
    u = swap16_if_be(u, is_le);
    *out_val = (double)(int16_t)u;
    return 0;
}

int scriptgo_dataview_set_int16(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 2);
    if (p == NULL) return typedarray_fail("scriptgo DataView setInt16 range out of bounds");
    uint16_t u = (uint16_t)(int16_t)(int32_t)value;
    u = swap16_if_be(u, is_le);
    memcpy(p, &u, 2);
    return 0;
}

int scriptgo_dataview_get_uint16(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 2);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getUint16 range out of bounds");
    uint16_t u;
    memcpy(&u, p, 2);
    u = swap16_if_be(u, is_le);
    *out_val = (double)u;
    return 0;
}

int scriptgo_dataview_set_uint16(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 2);
    if (p == NULL) return typedarray_fail("scriptgo DataView setUint16 range out of bounds");
    uint16_t u = (uint16_t)(uint32_t)value;
    u = swap16_if_be(u, is_le);
    memcpy(p, &u, 2);
    return 0;
}

int scriptgo_dataview_get_int32(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getInt32 range out of bounds");
    uint32_t u;
    memcpy(&u, p, 4);
    u = swap32_if_be(u, is_le);
    *out_val = (double)(int32_t)u;
    return 0;
}

int scriptgo_dataview_set_int32(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL) return typedarray_fail("scriptgo DataView setInt32 range out of bounds");
    uint32_t u = (uint32_t)(int32_t)value;
    u = swap32_if_be(u, is_le);
    memcpy(p, &u, 4);
    return 0;
}

int scriptgo_dataview_get_uint32(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getUint32 range out of bounds");
    uint32_t u;
    memcpy(&u, p, 4);
    u = swap32_if_be(u, is_le);
    *out_val = (double)u;
    return 0;
}

int scriptgo_dataview_set_uint32(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL) return typedarray_fail("scriptgo DataView setUint32 range out of bounds");
    uint32_t u = (uint32_t)value;
    u = swap32_if_be(u, is_le);
    memcpy(p, &u, 4);
    return 0;
}

int scriptgo_dataview_get_float32(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getFloat32 range out of bounds");
    uint32_t u;
    memcpy(&u, p, 4);
    u = swap32_if_be(u, is_le);
    float f;
    memcpy(&f, &u, 4);
    *out_val = (double)f;
    return 0;
}

int scriptgo_dataview_set_float32(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 4);
    if (p == NULL) return typedarray_fail("scriptgo DataView setFloat32 range out of bounds");
    float f = (float)value;
    uint32_t u;
    memcpy(&u, &f, 4);
    u = swap32_if_be(u, is_le);
    memcpy(p, &u, 4);
    return 0;
}

int scriptgo_dataview_get_float64(void *handle, double byte_offset, int32_t is_le, double *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 8);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getFloat64 range out of bounds");
    uint64_t u;
    memcpy(&u, p, 8);
    u = swap64_if_be(u, is_le);
    double d;
    memcpy(&d, &u, 8);
    *out_val = d;
    return 0;
}

int scriptgo_dataview_set_float64(void *handle, double byte_offset, double value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 8);
    if (p == NULL) return typedarray_fail("scriptgo DataView setFloat64 range out of bounds");
    uint64_t u;
    memcpy(&u, &value, 8);
    u = swap64_if_be(u, is_le);
    memcpy(p, &u, 8);
    return 0;
}

int scriptgo_dataview_get_bigint64(void *handle, double byte_offset, int32_t is_le, int64_t *out_val) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 8);
    if (p == NULL || out_val == NULL) return typedarray_fail("scriptgo DataView getBigInt64 range out of bounds");
    uint64_t u;
    memcpy(&u, p, 8);
    u = swap64_if_be(u, is_le);
    *out_val = (int64_t)u;
    return 0;
}

int scriptgo_dataview_set_bigint64(void *handle, double byte_offset, int64_t value, int32_t is_le) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    unsigned char *p = dataview_ptr(dv, byte_offset, 8);
    if (p == NULL) return typedarray_fail("scriptgo DataView setBigInt64 range out of bounds");
    uint64_t u = (uint64_t)value;
    u = swap64_if_be(u, is_le);
    memcpy(p, &u, 8);
    return 0;
}

int scriptgo_dataview_get_biguint64(void *handle, double byte_offset, int32_t is_le, int64_t *out_val) {
    return scriptgo_dataview_get_bigint64(handle, byte_offset, is_le, out_val);
}

int scriptgo_dataview_set_biguint64(void *handle, double byte_offset, int64_t value, int32_t is_le) {
    return scriptgo_dataview_set_bigint64(handle, byte_offset, value, is_le);
}

int scriptgo_dataview_to_string(void *handle, char **out_str) {
    scriptgo_data_view *dv = (scriptgo_data_view *)handle;
    if (dv == NULL || out_str == NULL) return typedarray_fail("scriptgo DataView toString failed");
    char *res = malloc(96);
    if (res == NULL) return typedarray_fail("scriptgo DataView toString allocation failed");
    snprintf(res, 96, "DataView { byteLength: %lld, byteOffset: %lld }", (long long)dv->byte_length, (long long)dv->byte_offset);
    *out_str = res;
    return 0;
}
