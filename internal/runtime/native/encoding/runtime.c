#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_typedarray_new(int64_t kind, int64_t length, void *buffer_handle, int64_t byte_offset, void **out_array);

#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252
#define SCRIPTGO_MAGIC_DATAVIEW   0x44564957
#define SCRIPTGO_MAGIC_TEXT_ENCODER 0x54454E43
#define SCRIPTGO_MAGIC_TEXT_DECODER 0x54444543

typedef struct {
    uint32_t magic;
} scriptgo_text_encoder_native;

typedef struct {
    uint32_t magic;
    int fatal;
    int ignore_bom;
} scriptgo_text_decoder_native;

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_typedarray_view_header;

typedef struct {
    uint32_t magic;
    void *buffer;
    int64_t byte_offset;
    int64_t byte_length;
} scriptgo_dataview_view_header;

typedef struct {
    int64_t byte_length;
    unsigned char *data;
} scriptgo_array_buffer_header;

static int encoding_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

int scriptgo_text_encoder_new(void **out_encoder) {
    if (out_encoder == NULL) return encoding_fail("scriptgo TextEncoder new: null output");
    scriptgo_text_encoder_native *enc = calloc(1, sizeof(scriptgo_text_encoder_native));
    if (enc == NULL) return encoding_fail("scriptgo TextEncoder new: out of memory");
    enc->magic = SCRIPTGO_MAGIC_TEXT_ENCODER;
    *out_encoder = enc;
    return 0;
}

int scriptgo_text_encoder_encoding(void *handle, const char **out_str) {
    if (out_str == NULL) return encoding_fail("scriptgo TextEncoder.encoding: null output");
    *out_str = "utf-8";
    return 0;
}

int scriptgo_text_encoder_encode(void *handle, const char *input, void **out_uint8array) {
    if (out_uint8array == NULL) return encoding_fail("scriptgo TextEncoder.encode: null output");
    if (input == NULL) input = "";
    size_t len = strlen(input);
    void *ta_ptr = NULL;
    if (scriptgo_typedarray_new(2 /* SCRIPTGO_TYPEDARRAY_UINT8 */, (int64_t)len, NULL, 0, &ta_ptr) != 0) {
        return encoding_fail("scriptgo TextEncoder.encode: failed to allocate Uint8Array");
    }
    scriptgo_typedarray_view_header *ta = (scriptgo_typedarray_view_header *)ta_ptr;
    if (len > 0 && ta != NULL && ta->data != NULL) {
        memcpy(ta->data, input, len);
    }
    *out_uint8array = ta_ptr;
    return 0;
}

int scriptgo_text_encoder_encode_into(void *handle, const char *source, void *destination_uint8array, double *out_read, double *out_written) {
    if (out_read == NULL || out_written == NULL) return encoding_fail("scriptgo TextEncoder.encodeInto: null output");
    if (destination_uint8array == NULL) return encoding_fail("scriptgo TextEncoder.encodeInto: null destination");
    
    scriptgo_typedarray_view_header *ta = (scriptgo_typedarray_view_header *)destination_uint8array;
    if (ta->magic != SCRIPTGO_MAGIC_TYPEDARRAY) {
        return encoding_fail("scriptgo TextEncoder.encodeInto: destination is not a TypedArray");
    }

    if (source == NULL) {
        *out_read = 0;
        *out_written = 0;
        return 0;
    }

    size_t dest_len = (size_t)ta->length;
    unsigned char *dest_data = ta->data;

    size_t src_len = strlen(source);
    size_t src_idx = 0;
    size_t written_idx = 0;
    size_t read_code_units = 0;

    while (src_idx < src_len) {
        unsigned char c = (unsigned char)source[src_idx];
        size_t char_bytes = 1;
        size_t code_units = 1;

        if (c < 0x80) {
            char_bytes = 1;
            code_units = 1;
        } else if ((c & 0xE0) == 0xC0) {
            char_bytes = 2;
            code_units = 1;
        } else if ((c & 0xF0) == 0xE0) {
            char_bytes = 3;
            code_units = 1;
        } else if ((c & 0xF8) == 0xF0) {
            char_bytes = 4;
            code_units = 2;
        }

        if (src_idx + char_bytes > src_len) {
            break;
        }
        if (written_idx + char_bytes > dest_len) {
            break;
        }

        for (size_t k = 0; k < char_bytes; k++) {
            dest_data[written_idx + k] = (unsigned char)source[src_idx + k];
        }

        written_idx += char_bytes;
        src_idx += char_bytes;
        read_code_units += code_units;
    }

    *out_read = (double)read_code_units;
    *out_written = (double)written_idx;
    return 0;
}

int scriptgo_text_decoder_new(const char *label, int32_t fatal, int32_t ignore_bom, void **out_decoder) {
    if (out_decoder == NULL) return encoding_fail("scriptgo TextDecoder new: null output");
    scriptgo_text_decoder_native *dec = calloc(1, sizeof(scriptgo_text_decoder_native));
    if (dec == NULL) return encoding_fail("scriptgo TextDecoder new: out of memory");
    dec->magic = SCRIPTGO_MAGIC_TEXT_DECODER;
    dec->fatal = fatal ? 1 : 0;
    dec->ignore_bom = ignore_bom ? 1 : 0;
    *out_decoder = dec;
    return 0;
}

int scriptgo_text_decoder_encoding(void *handle, const char **out_str) {
    if (out_str == NULL) return encoding_fail("scriptgo TextDecoder.encoding: null output");
    *out_str = "utf-8";
    return 0;
}

int scriptgo_text_decoder_fatal(void *handle, int32_t *out_val) {
    if (out_val == NULL) return encoding_fail("scriptgo TextDecoder.fatal: null output");
    scriptgo_text_decoder_native *dec = (scriptgo_text_decoder_native *)handle;
    *out_val = (dec != NULL && dec->fatal) ? 1 : 0;
    return 0;
}

int scriptgo_text_decoder_ignore_bom(void *handle, int32_t *out_val) {
    if (out_val == NULL) return encoding_fail("scriptgo TextDecoder.ignoreBOM: null output");
    scriptgo_text_decoder_native *dec = (scriptgo_text_decoder_native *)handle;
    *out_val = (dec != NULL && dec->ignore_bom) ? 1 : 0;
    return 0;
}

static int is_valid_utf8(const unsigned char *data, size_t len) {
    size_t i = 0;
    while (i < len) {
        unsigned char c = data[i];
        if (c < 0x80) {
            i++;
        } else if ((c & 0xE0) == 0xC0) {
            if (i + 1 >= len || (data[i + 1] & 0xC0) != 0x80 || c < 0xC2) return 0;
            i += 2;
        } else if ((c & 0xF0) == 0xE0) {
            if (i + 2 >= len || (data[i + 1] & 0xC0) != 0x80 || (data[i + 2] & 0xC0) != 0x80) return 0;
            if (c == 0xE0 && data[i + 1] < 0xA0) return 0;
            if (c == 0xED && data[i + 1] >= 0xA0) return 0;
            i += 3;
        } else if ((c & 0xF8) == 0xF0) {
            if (i + 3 >= len || (data[i + 1] & 0xC0) != 0x80 || (data[i + 2] & 0xC0) != 0x80 || (data[i + 3] & 0xC0) != 0x80) return 0;
            if (c == 0xF0 && data[i + 1] < 0x90) return 0;
            if (c == 0xF4 && data[i + 1] > 0x8F) return 0;
            i += 4;
        } else {
            return 0;
        }
    }
    return 1;
}

int scriptgo_text_decoder_decode(void *handle, void *input, const char **out_str) {
    if (out_str == NULL) return encoding_fail("scriptgo TextDecoder.decode: null output");
    if (input == NULL) {
        *out_str = strdup("");
        return 0;
    }

    scriptgo_text_decoder_native *dec = (scriptgo_text_decoder_native *)handle;
    int fatal = dec != NULL ? dec->fatal : 0;
    int ignore_bom = dec != NULL ? dec->ignore_bom : 0;

    const unsigned char *data = NULL;
    size_t byte_len = 0;

    uint32_t magic = *(uint32_t *)input;
    if (magic == SCRIPTGO_MAGIC_TYPEDARRAY) {
        scriptgo_typedarray_view_header *ta = (scriptgo_typedarray_view_header *)input;
        data = ta->data;
        byte_len = (size_t)(ta->length * ta->element_size);
    } else if (magic == SCRIPTGO_MAGIC_DATAVIEW) {
        scriptgo_dataview_view_header *dv = (scriptgo_dataview_view_header *)input;
        scriptgo_array_buffer_header *ab = (scriptgo_array_buffer_header *)dv->buffer;
        data = ab != NULL && ab->data != NULL ? (ab->data + dv->byte_offset) : NULL;
        byte_len = (size_t)dv->byte_length;
    } else {
        // Assume ArrayBuffer
        scriptgo_array_buffer_header *ab = (scriptgo_array_buffer_header *)input;
        data = ab->data;
        byte_len = (size_t)ab->byte_length;
    }

    if (data == NULL || byte_len == 0) {
        *out_str = strdup("");
        return 0;
    }

    // Check BOM
    if (!ignore_bom && byte_len >= 3) {
        if (data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF) {
            data += 3;
            byte_len -= 3;
        }
    }

    if (fatal && !is_valid_utf8(data, byte_len)) {
        return encoding_fail("TypeError: The encoded data was not valid.");
    }

    char *res = malloc(byte_len + 1);
    if (res == NULL) return encoding_fail("scriptgo TextDecoder.decode: out of memory");
    if (byte_len > 0) {
        memcpy(res, data, byte_len);
    }
    res[byte_len] = '\0';
    *out_str = res;
    return 0;
}
