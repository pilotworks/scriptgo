#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_arraybuffer_new(int64_t byte_length, void **out_buffer);
int scriptgo_typedarray_new(int64_t kind, int64_t length, void *buffer_handle, int64_t byte_offset, void **out_array);
int scriptgo_typedarray_from_array(int64_t kind, void *array_handle, void **out_array);

#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252 // "TARR"
#define SCRIPTGO_MAGIC_BUFFER     0x42554646 // "BUFF"
#define SCRIPTGO_MAGIC_DATAVIEW   0x44564957 // "DVIW"

typedef struct {
    int64_t byte_length;
    unsigned char *data;
} scriptgo_buffer_array_buffer;

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    scriptgo_buffer_array_buffer *buffer;
    unsigned char *data;
} scriptgo_buffer_view;

static int buffer_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static inline uint16_t b_swap16_if_be(uint16_t v, int is_le) {
    return is_le ? v : (uint16_t)((v >> 8) | (v << 8));
}
static inline uint32_t b_swap32_if_be(uint32_t v, int is_le) {
    return is_le ? v : (((v & 0xff000000) >> 24) | ((v & 0x00ff0000) >> 8) | ((v & 0x0000ff00) << 8) | ((v & 0x000000ff) << 24));
}
static inline uint64_t b_swap64_if_be(uint64_t v, int is_le) {
    return is_le ? v : (((v & 0xff00000000000000ULL) >> 56) |
                        ((v & 0x00ff000000000000ULL) >> 40) |
                        ((v & 0x0000ff0000000000ULL) >> 24) |
                        ((v & 0x000000ff00000000ULL) >> 8) |
                        ((v & 0x00000000ff000000ULL) << 8) |
                        ((v & 0x0000000000ff0000ULL) << 24) |
                        ((v & 0x000000000000ff00ULL) << 40) |
                        ((v & 0x00000000000000ffULL) << 56));
}

static const char b64_chars[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
static const char hex_chars[] = "0123456789abcdef";

static int b64_val(char c) {
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+' || c == '-') return 62;
    if (c == '/' || c == '_') return 63;
    return -1;
}

static int hex_val(char c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return -1;
}

typedef enum {
    ENC_UTF8,
    ENC_HEX,
    ENC_BASE64,
    ENC_ASCII,
    ENC_LATIN1
} scriptgo_encoding_t;

static scriptgo_encoding_t parse_encoding(const char *enc) {
    if (enc == NULL || *enc == '\0') return ENC_UTF8;
    if (strcasecmp(enc, "hex") == 0) return ENC_HEX;
    if (strcasecmp(enc, "base64") == 0 || strcasecmp(enc, "base64url") == 0) return ENC_BASE64;
    if (strcasecmp(enc, "ascii") == 0) return ENC_ASCII;
    if (strcasecmp(enc, "latin1") == 0 || strcasecmp(enc, "binary") == 0) return ENC_LATIN1;
    return ENC_UTF8;
}

int scriptgo_buffer_alloc(double size, const char *fill_str, double fill_num, int has_fill, int is_str_fill, void **out_buf) {
    if (out_buf == NULL) return buffer_fail("Buffer.alloc: null output");
    if (size < 0 || size != size) return buffer_fail("Buffer.alloc: invalid size");
    int64_t len = (int64_t)size;
    void *arr_ptr = NULL;
    if (scriptgo_typedarray_new(2 /* UINT8 */, len, NULL, 0, &arr_ptr) != 0) {
        return buffer_fail("Buffer.alloc: allocation failed");
    }
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)arr_ptr;
    bv->magic = SCRIPTGO_MAGIC_BUFFER;
    if (has_fill && len > 0 && bv->data != NULL) {
        if (is_str_fill && fill_str != NULL && *fill_str != '\0') {
            size_t flen = strlen(fill_str);
            for (int64_t i = 0; i < len; i++) {
                bv->data[i] = (unsigned char)fill_str[i % flen];
            }
        } else {
            unsigned char byte_val = (unsigned char)((int64_t)fill_num & 0xFF);
            memset(bv->data, byte_val, (size_t)len);
        }
    }
    *out_buf = bv;
    return 0;
}

int scriptgo_buffer_from_string(const char *str, const char *encoding_str, void **out_buf) {
    if (out_buf == NULL) return buffer_fail("Buffer.from: null output");
    if (str == NULL) str = "";
    scriptgo_encoding_t enc = parse_encoding(encoding_str);
    size_t in_len = strlen(str);

    switch (enc) {
    case ENC_HEX: {
        size_t byte_len = in_len / 2;
        void *arr_ptr = NULL;
        if (scriptgo_typedarray_new(2, (int64_t)byte_len, NULL, 0, &arr_ptr) != 0) return -1;
        scriptgo_buffer_view *bv = (scriptgo_buffer_view *)arr_ptr;
        bv->magic = SCRIPTGO_MAGIC_BUFFER;
        for (size_t i = 0; i < byte_len; i++) {
            int hi = hex_val(str[i * 2]);
            int lo = hex_val(str[i * 2 + 1]);
            if (hi < 0 || lo < 0) break;
            bv->data[i] = (unsigned char)((hi << 4) | lo);
        }
        *out_buf = bv;
        return 0;
    }
    case ENC_BASE64: {
        // Base64 decode
        size_t max_out = (in_len * 3) / 4 + 4;
        unsigned char *tmp = malloc(max_out > 0 ? max_out : 1);
        if (tmp == NULL) return buffer_fail("Buffer.from base64 out of memory");
        size_t out_len = 0;
        uint32_t buf = 0;
        int bits = 0;
        for (size_t i = 0; i < in_len; i++) {
            char c = str[i];
            if (c == '=') break;
            int val = b64_val(c);
            if (val < 0) continue;
            buf = (buf << 6) | (uint32_t)val;
            bits += 6;
            if (bits >= 8) {
                bits -= 8;
                tmp[out_len++] = (unsigned char)((buf >> bits) & 0xFF);
            }
        }
        void *arr_ptr = NULL;
        if (scriptgo_typedarray_new(2, (int64_t)out_len, NULL, 0, &arr_ptr) != 0) {
            free(tmp);
            return -1;
        }
        scriptgo_buffer_view *bv = (scriptgo_buffer_view *)arr_ptr;
        bv->magic = SCRIPTGO_MAGIC_BUFFER;
        if (out_len > 0) memcpy(bv->data, tmp, out_len);
        free(tmp);
        *out_buf = bv;
        return 0;
    }
    case ENC_ASCII: {
        void *arr_ptr = NULL;
        if (scriptgo_typedarray_new(2, (int64_t)in_len, NULL, 0, &arr_ptr) != 0) return -1;
        scriptgo_buffer_view *bv = (scriptgo_buffer_view *)arr_ptr;
        bv->magic = SCRIPTGO_MAGIC_BUFFER;
        for (size_t i = 0; i < in_len; i++) {
            bv->data[i] = (unsigned char)(str[i] & 0x7F);
        }
        *out_buf = bv;
        return 0;
    }
    case ENC_LATIN1:
    case ENC_UTF8:
    default: {
        void *arr_ptr = NULL;
        if (scriptgo_typedarray_new(2, (int64_t)in_len, NULL, 0, &arr_ptr) != 0) return -1;
        scriptgo_buffer_view *bv = (scriptgo_buffer_view *)arr_ptr;
        bv->magic = SCRIPTGO_MAGIC_BUFFER;
        if (in_len > 0) memcpy(bv->data, str, in_len);
        *out_buf = bv;
        return 0;
    }
    }
}

int scriptgo_buffer_from_array(void *arr_handle, void **out_buf) {
    if (out_buf == NULL) return buffer_fail("Buffer.from: null output");
    if (arr_handle == NULL) {
        return scriptgo_buffer_alloc(0, NULL, 0, 0, 0, out_buf);
    }
    uint32_t magic = *(uint32_t *)arr_handle;
    if (magic == SCRIPTGO_MAGIC_TYPEDARRAY || magic == SCRIPTGO_MAGIC_BUFFER) {
        scriptgo_buffer_view *src = (scriptgo_buffer_view *)arr_handle;
        void *created = NULL;
        if (scriptgo_typedarray_new(2, src->length, NULL, 0, &created) != 0) return -1;
        scriptgo_buffer_view *dst = (scriptgo_buffer_view *)created;
        dst->magic = SCRIPTGO_MAGIC_BUFFER;
        if (src->length > 0 && src->data != NULL && dst->data != NULL) {
            memcpy(dst->data, src->data, (size_t)src->length);
        }
        *out_buf = dst;
        return 0;
    }
    void *created = NULL;
    if (scriptgo_typedarray_from_array(2, arr_handle, &created) == 0 && created != NULL) {
        scriptgo_buffer_view *dst = (scriptgo_buffer_view *)created;
        dst->magic = SCRIPTGO_MAGIC_BUFFER;
        *out_buf = dst;
        return 0;
    }
    return scriptgo_buffer_alloc(0, NULL, 0, 0, 0, out_buf);
}

int scriptgo_buffer_concat(void *list_handle, double total_length_opt, void **out_buf) {
    if (out_buf == NULL) return buffer_fail("Buffer.concat: null output");
    if (list_handle == NULL) return scriptgo_buffer_alloc(0, NULL, 0, 0, 0, out_buf);

    // list_handle is an array of Buffer pointers
    typedef struct {
        int64_t length;
        int64_t capacity;
        int64_t element_size;
        void **data;
    } scriptgo_array_raw;

    scriptgo_array_raw *arr = (scriptgo_array_raw *)list_handle;
    int64_t n = arr->length;
    int64_t calc_len = 0;
    for (int64_t i = 0; i < n; i++) {
        if (arr->data[i] != NULL) {
            scriptgo_buffer_view *item = (scriptgo_buffer_view *)arr->data[i];
            calc_len += item->length;
        }
    }
    int64_t final_len = calc_len;
    if (total_length_opt >= 0 && total_length_opt == total_length_opt) {
        final_len = (int64_t)total_length_opt;
    }
    void *created = NULL;
    if (scriptgo_typedarray_new(2, final_len, NULL, 0, &created) != 0) return -1;
    scriptgo_buffer_view *dst = (scriptgo_buffer_view *)created;
    dst->magic = SCRIPTGO_MAGIC_BUFFER;
    
    int64_t off = 0;
    for (int64_t i = 0; i < n && off < final_len; i++) {
        if (arr->data[i] != NULL) {
            scriptgo_buffer_view *item = (scriptgo_buffer_view *)arr->data[i];
            int64_t copy_bytes = item->length;
            if (off + copy_bytes > final_len) copy_bytes = final_len - off;
            if (copy_bytes > 0 && item->data != NULL) {
                memcpy(dst->data + off, item->data, (size_t)copy_bytes);
            }
            off += copy_bytes;
        }
    }
    *out_buf = dst;
    return 0;
}

int scriptgo_buffer_is_buffer(void *handle, int32_t *out_is_buf) {
    if (out_is_buf == NULL) return buffer_fail("Buffer.isBuffer: null output");
    if (handle == NULL) {
        *out_is_buf = 0;
        return 0;
    }
    uint32_t magic = *(uint32_t *)handle;
    *out_is_buf = (magic == SCRIPTGO_MAGIC_BUFFER) ? 1 : 0;
    return 0;
}

int scriptgo_buffer_byte_length(const char *str, const char *encoding_str, double *out_len) {
    if (out_len == NULL) return buffer_fail("Buffer.byteLength: null output");
    if (str == NULL) str = "";
    scriptgo_encoding_t enc = parse_encoding(encoding_str);
    size_t in_len = strlen(str);
    switch (enc) {
    case ENC_HEX:
        *out_len = (double)(in_len / 2);
        return 0;
    case ENC_BASE64: {
        size_t out_len_cnt = 0;
        int bits = 0;
        for (size_t i = 0; i < in_len; i++) {
            if (str[i] == '=') break;
            if (b64_val(str[i]) >= 0) {
                bits += 6;
                if (bits >= 8) {
                    bits -= 8;
                    out_len_cnt++;
                }
            }
        }
        *out_len = (double)out_len_cnt;
        return 0;
    }
    default:
        *out_len = (double)in_len;
        return 0;
    }
}

int scriptgo_buffer_to_string(void *handle, const char *encoding_str, double start_opt, double end_opt, int has_start, int has_end, const char **out_str) {
    if (out_str == NULL) return buffer_fail("Buffer.toString: null output");
    if (handle == NULL) {
        *out_str = "";
        return 0;
    }
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t len = bv->length;
    int64_t start = has_start ? (int64_t)start_opt : 0;
    int64_t end = has_end ? (int64_t)end_opt : len;
    if (start < 0) start = 0;
    if (start > len) start = len;
    if (end < start) end = start;
    if (end > len) end = len;
    int64_t slice_len = end - start;

    if (slice_len <= 0 || bv->data == NULL) {
        *out_str = "";
        return 0;
    }

    const unsigned char *data = bv->data + start;
    scriptgo_encoding_t enc = parse_encoding(encoding_str);

    switch (enc) {
    case ENC_HEX: {
        char *hex_str = malloc((size_t)slice_len * 2 + 1);
        if (hex_str == NULL) return buffer_fail("Buffer.toString hex out of memory");
        for (int64_t i = 0; i < slice_len; i++) {
            hex_str[i * 2] = hex_chars[(data[i] >> 4) & 0x0F];
            hex_str[i * 2 + 1] = hex_chars[data[i] & 0x0F];
        }
        hex_str[slice_len * 2] = '\0';
        *out_str = hex_str;
        return 0;
    }
    case ENC_BASE64: {
        size_t b64_len = 4 * ((slice_len + 2) / 3);
        char *b64_str = malloc(b64_len + 1);
        if (b64_str == NULL) return buffer_fail("Buffer.toString base64 out of memory");
        size_t o = 0;
        for (int64_t i = 0; i < slice_len; i += 3) {
            uint32_t octet_a = data[i];
            uint32_t octet_b = (i + 1 < slice_len) ? data[i + 1] : 0;
            uint32_t octet_c = (i + 2 < slice_len) ? data[i + 2] : 0;
            uint32_t triple = (octet_a << 16) | (octet_b << 8) | octet_c;
            b64_str[o++] = b64_chars[(triple >> 18) & 0x3F];
            b64_str[o++] = b64_chars[(triple >> 12) & 0x3F];
            b64_str[o++] = (i + 1 < slice_len) ? b64_chars[(triple >> 6) & 0x3F] : '=';
            b64_str[o++] = (i + 2 < slice_len) ? b64_chars[triple & 0x3F] : '=';
        }
        b64_str[o] = '\0';
        *out_str = b64_str;
        return 0;
    }
    case ENC_ASCII:
    case ENC_LATIN1:
    case ENC_UTF8:
    default: {
        char *res = malloc((size_t)slice_len + 1);
        if (res == NULL) return buffer_fail("Buffer.toString out of memory");
        memcpy(res, data, (size_t)slice_len);
        res[slice_len] = '\0';
        *out_str = res;
        return 0;
    }
    }
}

int scriptgo_buffer_copy(void *src_handle, void *target_handle, double target_start, double source_start, double source_end, int has_ts, int has_ss, int has_se, double *out_copied) {
    if (out_copied == NULL) return buffer_fail("Buffer.copy: null output");
    if (src_handle == NULL || target_handle == NULL) {
        *out_copied = 0;
        return 0;
    }
    scriptgo_buffer_view *src = (scriptgo_buffer_view *)src_handle;
    scriptgo_buffer_view *dst = (scriptgo_buffer_view *)target_handle;
    int64_t tstart = has_ts ? (int64_t)target_start : 0;
    int64_t sstart = has_ss ? (int64_t)source_start : 0;
    int64_t send = has_se ? (int64_t)source_end : src->length;

    if (tstart < 0) tstart = 0;
    if (sstart < 0) sstart = 0;
    if (send > src->length) send = src->length;
    if (sstart >= send || tstart >= dst->length) {
        *out_copied = 0;
        return 0;
    }
    int64_t to_copy = send - sstart;
    if (tstart + to_copy > dst->length) {
        to_copy = dst->length - tstart;
    }
    if (to_copy > 0 && src->data != NULL && dst->data != NULL) {
        memmove(dst->data + tstart, src->data + sstart, (size_t)to_copy);
    }
    *out_copied = (double)to_copy;
    return 0;
}

int scriptgo_buffer_fill(void *handle, const char *fill_str, double fill_num, int is_str, double start_opt, double end_opt, int has_s, int has_e, void **out_buf) {
    if (out_buf == NULL) return buffer_fail("Buffer.fill: null output");
    if (handle == NULL) return buffer_fail("Buffer.fill: null target");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t start = has_s ? (int64_t)start_opt : 0;
    int64_t end = has_e ? (int64_t)end_opt : bv->length;
    if (start < 0) start = 0;
    if (end > bv->length) end = bv->length;
    if (start < end && bv->data != NULL) {
        if (is_str && fill_str != NULL && *fill_str != '\0') {
            size_t flen = strlen(fill_str);
            for (int64_t i = start; i < end; i++) {
                bv->data[i] = (unsigned char)fill_str[(i - start) % flen];
            }
        } else {
            unsigned char b = (unsigned char)((int64_t)fill_num & 0xFF);
            memset(bv->data + start, b, (size_t)(end - start));
        }
    }
    *out_buf = bv;
    return 0;
}

int scriptgo_buffer_equals(void *a_handle, void *b_handle, int32_t *out_eq) {
    if (out_eq == NULL) return buffer_fail("Buffer.equals: null output");
    if (a_handle == b_handle) {
        *out_eq = 1;
        return 0;
    }
    if (a_handle == NULL || b_handle == NULL) {
        *out_eq = 0;
        return 0;
    }
    scriptgo_buffer_view *a = (scriptgo_buffer_view *)a_handle;
    scriptgo_buffer_view *b = (scriptgo_buffer_view *)b_handle;
    if (a->length != b->length) {
        *out_eq = 0;
        return 0;
    }
    if (a->length == 0) {
        *out_eq = 1;
        return 0;
    }
    *out_eq = (memcmp(a->data, b->data, (size_t)a->length) == 0) ? 1 : 0;
    return 0;
}

int scriptgo_buffer_compare(void *a_handle, void *b_handle, double *out_cmp) {
    if (out_cmp == NULL) return buffer_fail("Buffer.compare: null output");
    if (a_handle == b_handle) {
        *out_cmp = 0;
        return 0;
    }
    if (a_handle == NULL) {
        *out_cmp = -1;
        return 0;
    }
    if (b_handle == NULL) {
        *out_cmp = 1;
        return 0;
    }
    scriptgo_buffer_view *a = (scriptgo_buffer_view *)a_handle;
    scriptgo_buffer_view *b = (scriptgo_buffer_view *)b_handle;
    int64_t min_len = a->length < b->length ? a->length : b->length;
    int c = 0;
    if (min_len > 0 && a->data != NULL && b->data != NULL) {
        c = memcmp(a->data, b->data, (size_t)min_len);
    }
    if (c != 0) {
        *out_cmp = c < 0 ? -1 : 1;
    } else if (a->length < b->length) {
        *out_cmp = -1;
    } else if (a->length > b->length) {
        *out_cmp = 1;
    } else {
        *out_cmp = 0;
    }
    return 0;
}

int scriptgo_buffer_index_of(void *handle, const char *val_str, double val_num, int is_str, double byte_offset, int has_offset, double *out_idx) {
    if (out_idx == NULL) return buffer_fail("Buffer.indexOf: null output");
    if (handle == NULL) {
        *out_idx = -1;
        return 0;
    }
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = has_offset ? (int64_t)byte_offset : 0;
    if (off < 0) off += bv->length;
    if (off < 0) off = 0;
    if (off >= bv->length || bv->data == NULL) {
        *out_idx = -1;
        return 0;
    }

    if (is_str) {
        if (val_str == NULL || *val_str == '\0') {
            *out_idx = (double)off;
            return 0;
        }
        size_t nlen = strlen(val_str);
        if (nlen > (size_t)(bv->length - off)) {
            *out_idx = -1;
            return 0;
        }
        for (int64_t i = off; i <= bv->length - (int64_t)nlen; i++) {
            if (memcmp(bv->data + i, val_str, nlen) == 0) {
                *out_idx = (double)i;
                return 0;
            }
        }
        *out_idx = -1;
        return 0;
    } else {
        unsigned char target_byte = (unsigned char)((int64_t)val_num & 0xFF);
        for (int64_t i = off; i < bv->length; i++) {
            if (bv->data[i] == target_byte) {
                *out_idx = (double)i;
                return 0;
            }
        }
        *out_idx = -1;
        return 0;
    }
}

// Binary Read / Write functions
int scriptgo_buffer_read_u8(void *handle, double offset, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readUInt8: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off >= bv->length) {
        return buffer_fail("Buffer.readUInt8: out of range");
    }
    *out_val = (double)bv->data[off];
    return 0;
}

int scriptgo_buffer_write_u8(void *handle, double val, double offset, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off >= bv->length) {
        return buffer_fail("Buffer.writeUInt8: out of range");
    }
    bv->data[off] = (unsigned char)((int64_t)val & 0xFF);
    if (out_written) *out_written = (double)(off + 1);
    return 0;
}

int scriptgo_buffer_read_i8(void *handle, double offset, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readInt8: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off >= bv->length) {
        return buffer_fail("Buffer.readInt8: out of range");
    }
    *out_val = (double)(int8_t)bv->data[off];
    return 0;
}

int scriptgo_buffer_write_i8(void *handle, double val, double offset, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off >= bv->length) {
        return buffer_fail("Buffer.writeInt8: out of range");
    }
    bv->data[off] = (unsigned char)(int8_t)(int64_t)val;
    if (out_written) *out_written = (double)(off + 1);
    return 0;
}

int scriptgo_buffer_read_u16(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readUInt16: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 2 > bv->length) {
        return buffer_fail("Buffer.readUInt16: out of range");
    }
    uint16_t raw;
    memcpy(&raw, bv->data + off, 2);
    *out_val = (double)b_swap16_if_be(raw, is_le);
    return 0;
}

int scriptgo_buffer_write_u16(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 2 > bv->length) {
        return buffer_fail("Buffer.writeUInt16: out of range");
    }
    uint16_t raw = b_swap16_if_be((uint16_t)(uint32_t)val, is_le);
    memcpy(bv->data + off, &raw, 2);
    if (out_written) *out_written = (double)(off + 2);
    return 0;
}

int scriptgo_buffer_read_i16(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readInt16: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 2 > bv->length) {
        return buffer_fail("Buffer.readInt16: out of range");
    }
    uint16_t raw;
    memcpy(&raw, bv->data + off, 2);
    *out_val = (double)(int16_t)b_swap16_if_be(raw, is_le);
    return 0;
}

int scriptgo_buffer_write_i16(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 2 > bv->length) {
        return buffer_fail("Buffer.writeInt16: out of range");
    }
    uint16_t raw = b_swap16_if_be((uint16_t)(int16_t)(int32_t)val, is_le);
    memcpy(bv->data + off, &raw, 2);
    if (out_written) *out_written = (double)(off + 2);
    return 0;
}

int scriptgo_buffer_read_u32(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readUInt32: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.readUInt32: out of range");
    }
    uint32_t raw;
    memcpy(&raw, bv->data + off, 4);
    *out_val = (double)b_swap32_if_be(raw, is_le);
    return 0;
}

int scriptgo_buffer_write_u32(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.writeUInt32: out of range");
    }
    uint32_t raw = b_swap32_if_be((uint32_t)val, is_le);
    memcpy(bv->data + off, &raw, 4);
    if (out_written) *out_written = (double)(off + 4);
    return 0;
}

int scriptgo_buffer_read_i32(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readInt32: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.readInt32: out of range");
    }
    uint32_t raw;
    memcpy(&raw, bv->data + off, 4);
    *out_val = (double)(int32_t)b_swap32_if_be(raw, is_le);
    return 0;
}

int scriptgo_buffer_write_i32(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.writeInt32: out of range");
    }
    uint32_t raw = b_swap32_if_be((uint32_t)(int32_t)val, is_le);
    memcpy(bv->data + off, &raw, 4);
    if (out_written) *out_written = (double)(off + 4);
    return 0;
}

int scriptgo_buffer_read_float(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readFloat: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.readFloat: out of range");
    }
    uint32_t raw;
    memcpy(&raw, bv->data + off, 4);
    raw = b_swap32_if_be(raw, is_le);
    float f;
    memcpy(&f, &raw, 4);
    *out_val = (double)f;
    return 0;
}

int scriptgo_buffer_write_float(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 4 > bv->length) {
        return buffer_fail("Buffer.writeFloat: out of range");
    }
    float f = (float)val;
    uint32_t raw;
    memcpy(&raw, &f, 4);
    raw = b_swap32_if_be(raw, is_le);
    memcpy(bv->data + off, &raw, 4);
    if (out_written) *out_written = (double)(off + 4);
    return 0;
}

int scriptgo_buffer_read_double(void *handle, double offset, int is_le, double *out_val) {
    if (out_val == NULL) return buffer_fail("Buffer.readDouble: null output");
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 8 > bv->length) {
        return buffer_fail("Buffer.readDouble: out of range");
    }
    uint64_t raw;
    memcpy(&raw, bv->data + off, 8);
    raw = b_swap64_if_be(raw, is_le);
    double d;
    memcpy(&d, &raw, 8);
    *out_val = d;
    return 0;
}

int scriptgo_buffer_write_double(void *handle, double val, double offset, int is_le, double *out_written) {
    scriptgo_buffer_view *bv = (scriptgo_buffer_view *)handle;
    int64_t off = (int64_t)offset;
    if (bv == NULL || bv->data == NULL || off < 0 || off + 8 > bv->length) {
        return buffer_fail("Buffer.writeDouble: out of range");
    }
    double d = val;
    uint64_t raw;
    memcpy(&raw, &d, 8);
    raw = b_swap64_if_be(raw, is_le);
    memcpy(bv->data + off, &raw, 8);
    if (out_written) *out_written = (double)(off + 8);
    return 0;
}
