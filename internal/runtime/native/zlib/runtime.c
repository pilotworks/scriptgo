#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#ifdef SCRIPTGO_HAS_ZLIB
#include <zlib.h>
#endif
#ifdef SCRIPTGO_HAS_BROTLI
#include <brotli/decode.h>
#include <brotli/encode.h>
#endif
#ifdef SCRIPTGO_HAS_ZSTD
#include <zstd.h>
#endif

int scriptgo_runtime_set_error(const char *message);
int scriptgo_buffer_alloc(double size, const char *fill_str, double fill_num, int has_fill, int is_str_fill, void **out_buf);

#define SCRIPTGO_MAGIC_TYPEDARRAY 0x54415252u
#define SCRIPTGO_MAGIC_BUFFER     0x42554646u

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_zlib_view;

static int zlib_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static int zlib_input(void *handle, const unsigned char **data, size_t *length) {
    if (handle == NULL || data == NULL || length == NULL) return zlib_fail("zlib: invalid input");
    scriptgo_zlib_view *view = (scriptgo_zlib_view *)handle;
    if (view->magic != SCRIPTGO_MAGIC_BUFFER && view->magic != SCRIPTGO_MAGIC_TYPEDARRAY) {
        return zlib_fail("zlib: input must be a Buffer or Uint8Array");
    }
    if (view->length < 0 || (view->length > 0 && view->data == NULL)) return zlib_fail("zlib: invalid input buffer");
    *data = view->data;
    *length = (size_t)view->length;
    return 0;
}

static int zlib_output(const unsigned char *data, size_t length, void **out) {
    if (out == NULL) return zlib_fail("zlib: invalid output");
    if (scriptgo_buffer_alloc((double)length, NULL, 0, 0, 0, out) != 0) return -1;
    if (length > 0 && data != NULL) {
        scriptgo_zlib_view *view = (scriptgo_zlib_view *)*out;
        memcpy(view->data, data, length);
    }
    return 0;
}

#ifdef SCRIPTGO_HAS_ZLIB
static int zlib_compress(const unsigned char *input, size_t input_length, int window_bits, void **out) {
    size_t capacity = compressBound((uLong)input_length) + 64u;
    unsigned char *buffer = (unsigned char *)malloc(capacity == 0 ? 1 : capacity);
    if (buffer == NULL) return zlib_fail("zlib: allocation failed");

    z_stream stream;
    memset(&stream, 0, sizeof(stream));
    int status = deflateInit2(&stream, Z_DEFAULT_COMPRESSION, Z_DEFLATED, window_bits, 8, Z_DEFAULT_STRATEGY);
    if (status != Z_OK) {
        free(buffer);
        return zlib_fail("zlib: compression initialization failed");
    }
    stream.next_in = (Bytef *)input;
    stream.avail_in = (uInt)input_length;
    stream.next_out = buffer;
    stream.avail_out = (uInt)capacity;
    status = deflate(&stream, Z_FINISH);
    if (status != Z_STREAM_END) {
        deflateEnd(&stream);
        free(buffer);
        return zlib_fail("zlib: compression failed");
    }
    size_t output_length = (size_t)stream.total_out;
    deflateEnd(&stream);
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}

static int zlib_decompress(const unsigned char *input, size_t input_length, int window_bits, void **out) {
    size_t capacity = input_length > 128 ? input_length * 2u : 256u;
    if (capacity < 256u) capacity = 256u;
    unsigned char *buffer = (unsigned char *)malloc(capacity);
    if (buffer == NULL) return zlib_fail("zlib: allocation failed");

    z_stream stream;
    memset(&stream, 0, sizeof(stream));
    int status = inflateInit2(&stream, window_bits);
    if (status != Z_OK) {
        free(buffer);
        return zlib_fail("zlib: decompression initialization failed");
    }
    stream.next_in = (Bytef *)input;
    stream.avail_in = (uInt)input_length;
    for (;;) {
        if (stream.total_out == capacity) {
            if (capacity > (size_t)0x3fffffff) {
                inflateEnd(&stream);
                free(buffer);
                return zlib_fail("zlib: decompressed data is too large");
            }
            size_t next = capacity * 2u;
            unsigned char *grown = (unsigned char *)realloc(buffer, next);
            if (grown == NULL) {
                inflateEnd(&stream);
                free(buffer);
                return zlib_fail("zlib: allocation failed");
            }
            buffer = grown;
            capacity = next;
        }
        stream.next_out = buffer + stream.total_out;
        stream.avail_out = (uInt)(capacity - stream.total_out);
        status = inflate(&stream, Z_NO_FLUSH);
        if (status == Z_STREAM_END) break;
        if (status != Z_OK) {
            inflateEnd(&stream);
            free(buffer);
            return zlib_fail("zlib: decompression failed");
        }
        if (stream.avail_in == 0 && stream.avail_out > 0) {
            inflateEnd(&stream);
            free(buffer);
            return zlib_fail("zlib: truncated input");
        }
    }
    size_t output_length = (size_t)stream.total_out;
    inflateEnd(&stream);
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}
#endif

#ifdef SCRIPTGO_HAS_BROTLI
static int brotli_compress(const unsigned char *input, size_t input_length, void **out) {
    size_t capacity = BrotliEncoderMaxCompressedSize(input_length);
    if (capacity == 0) capacity = 1;
    unsigned char *buffer = (unsigned char *)malloc(capacity);
    if (buffer == NULL) return zlib_fail("brotli: allocation failed");
    size_t output_length = capacity;
    BROTLI_BOOL ok = BrotliEncoderCompress(BROTLI_DEFAULT_QUALITY, BROTLI_DEFAULT_WINDOW,
                                            BROTLI_MODE_GENERIC, input_length, input,
                                            &output_length, buffer);
    if (!ok) {
        free(buffer);
        return zlib_fail("brotli: compression failed");
    }
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}

static int brotli_decompress(const unsigned char *input, size_t input_length, void **out) {
    size_t capacity = input_length > 128 ? input_length * 3u : 256u;
    unsigned char *buffer = (unsigned char *)malloc(capacity);
    BrotliDecoderState *state = BrotliDecoderCreateInstance(NULL, NULL, NULL);
    if (buffer == NULL || state == NULL) {
        free(buffer);
        BrotliDecoderDestroyInstance(state);
        return zlib_fail("brotli: allocation failed");
    }
    const uint8_t *next_in = input;
    size_t available_in = input_length;
    size_t output_length = 0;
    for (;;) {
        size_t available_out = capacity - output_length;
        uint8_t *next_out = buffer + output_length;
        BrotliDecoderResult status = BrotliDecoderDecompressStream(
            state, &available_in, &next_in, &available_out, &next_out, NULL);
        output_length = capacity - available_out;
        if (status == BROTLI_DECODER_RESULT_SUCCESS) break;
        if (status != BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT) {
            BrotliDecoderDestroyInstance(state);
            free(buffer);
            return zlib_fail("brotli: decompression failed");
        }
        size_t next_capacity = capacity * 2u;
        unsigned char *grown = (unsigned char *)realloc(buffer, next_capacity);
        if (grown == NULL) {
            BrotliDecoderDestroyInstance(state);
            free(buffer);
            return zlib_fail("brotli: allocation failed");
        }
        buffer = grown;
        capacity = next_capacity;
    }
    BrotliDecoderDestroyInstance(state);
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}
#endif

#ifdef SCRIPTGO_HAS_ZSTD
static int zstd_compress(const unsigned char *input, size_t input_length, void **out) {
    size_t capacity = ZSTD_compressBound(input_length);
    unsigned char *buffer = (unsigned char *)malloc(capacity == 0 ? 1 : capacity);
    if (buffer == NULL) return zlib_fail("zstd: allocation failed");
    size_t output_length = ZSTD_compress(buffer, capacity, input, input_length, 3);
    if (ZSTD_isError(output_length)) {
        free(buffer);
        return zlib_fail("zstd: compression failed");
    }
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}

static int zstd_decompress(const unsigned char *input, size_t input_length, void **out) {
    size_t capacity = 256u;
    unsigned long long frame_size = ZSTD_getFrameContentSize(input, input_length);
    if (frame_size != ZSTD_CONTENTSIZE_UNKNOWN && frame_size != ZSTD_CONTENTSIZE_ERROR) {
        capacity = frame_size == 0 ? 1 : (size_t)frame_size;
    }
    unsigned char *buffer = (unsigned char *)malloc(capacity);
    ZSTD_DStream *stream = ZSTD_createDStream();
    if (buffer == NULL || stream == NULL || ZSTD_isError(ZSTD_initDStream(stream))) {
        free(buffer);
        ZSTD_freeDStream(stream);
        return zlib_fail("zstd: allocation failed");
    }
    ZSTD_inBuffer in_buffer = {input, input_length, 0};
    size_t output_length = 0;
    size_t remaining = 1;
    while (in_buffer.pos < in_buffer.size || remaining != 0) {
        if (output_length == capacity) {
            size_t next_capacity = capacity * 2u;
            unsigned char *grown = (unsigned char *)realloc(buffer, next_capacity);
            if (grown == NULL) {
                ZSTD_freeDStream(stream);
                free(buffer);
                return zlib_fail("zstd: allocation failed");
            }
            buffer = grown;
            capacity = next_capacity;
        }
        ZSTD_outBuffer out_buffer = {buffer + output_length, capacity - output_length, 0};
        remaining = ZSTD_decompressStream(stream, &out_buffer, &in_buffer);
        if (ZSTD_isError(remaining)) {
            ZSTD_freeDStream(stream);
            free(buffer);
            return zlib_fail("zstd: decompression failed");
        }
        output_length += out_buffer.pos;
        if (in_buffer.pos == in_buffer.size && remaining != 0 && out_buffer.pos == 0) {
            ZSTD_freeDStream(stream);
            free(buffer);
            return zlib_fail("zstd: truncated input");
        }
    }
    ZSTD_freeDStream(stream);
    int result = zlib_output(buffer, output_length, out);
    free(buffer);
    return result;
}
#endif

static int scriptgo_zlib_transform(const unsigned char *input, size_t input_length, double mode, void **out) {
#ifdef SCRIPTGO_HAS_ZLIB
    int operation = (int)mode;
    switch (operation) {
    case 0: return zlib_compress(input, input_length, 15, out);
    case 1: return zlib_compress(input, input_length, -15, out);
    case 2: return zlib_compress(input, input_length, 31, out);
    case 3: return zlib_decompress(input, input_length, 15, out);
    case 4: return zlib_decompress(input, input_length, -15, out);
    case 5: return zlib_decompress(input, input_length, 47, out);
#ifdef SCRIPTGO_HAS_BROTLI
    case 6: return brotli_compress(input, input_length, out);
    case 7: return brotli_decompress(input, input_length, out);
#else
    case 6:
    case 7: return zlib_fail("zlib: Brotli codec runtime is unavailable for this target");
#endif
#ifdef SCRIPTGO_HAS_ZSTD
    case 8: return zstd_compress(input, input_length, out);
    case 9: return zstd_decompress(input, input_length, out);
#else
    case 8:
    case 9: return zlib_fail("zlib: Zstandard codec runtime is unavailable for this target");
#endif
    default: return zlib_fail("zlib: unknown transform mode");
    }
#else
    (void)input;
    (void)input_length;
    (void)mode;
    (void)out;
    return zlib_fail("zlib: codec runtime is unavailable for this target");
#endif
}

int scriptgo_zlib_transform_string(const char *input, double mode, void **out) {
    if (input == NULL) input = "";
    return scriptgo_zlib_transform((const unsigned char *)input, strlen(input), mode, out);
}

int scriptgo_zlib_transform_buffer(void *handle, double mode, void **out) {
    const unsigned char *input = NULL;
    size_t length = 0;
    if (zlib_input(handle, &input, &length) != 0) return -1;
    return scriptgo_zlib_transform(input, length, mode, out);
}
