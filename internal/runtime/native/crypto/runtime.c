#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <math.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__)
#define HAVE_ARC4RANDOM 1
#endif

int scriptgo_runtime_set_error(const char *message);
int scriptgo_buffer_alloc(double size, const char *fill_str, double fill_num, int has_fill, int is_str_fill, void **out_buf);

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_crypto_buffer_view;

static int crypto_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static int get_random_bytes(unsigned char *buf, size_t n) {
#if defined(HAVE_ARC4RANDOM)
    arc4random_buf(buf, n);
    return 0;
#else
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd < 0) return -1;
    size_t total = 0;
    while (total < n) {
        ssize_t r = read(fd, buf + total, n - total);
        if (r <= 0) {
            close(fd);
            return -1;
        }
        total += (size_t)r;
    }
    close(fd);
    return 0;
#endif
}

int scriptgo_crypto_random_uuid(char **out_uuid) {
    if (out_uuid == NULL) {
        return crypto_fail("scriptgo crypto invalid arguments");
    }
    unsigned char b[16];
    if (get_random_bytes(b, 16) != 0) {
        return crypto_fail("scriptgo crypto failed to generate random bytes");
    }
    b[6] = (b[6] & 0x0f) | 0x40; // Version 4
    b[8] = (b[8] & 0x3f) | 0x80; // Variant RFC 4122

    char *uuid = (char *)malloc(37);
    if (uuid == NULL) {
        return crypto_fail("scriptgo crypto allocation failed");
    }
    snprintf(uuid, 37,
        "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
        b[0], b[1], b[2], b[3],
        b[4], b[5],
        b[6], b[7],
        b[8], b[9],
        b[10], b[11], b[12], b[13], b[14], b[15]
    );
    *out_uuid = uuid;
    return 0;
}

int scriptgo_crypto_random_bytes(double size, void **out_buf) {
    if (out_buf == NULL || size < 0 || size != size) {
        return crypto_fail("crypto.randomBytes: invalid arguments");
    }
    size_t n = (size_t)size;
    if (scriptgo_buffer_alloc(size, NULL, 0, 0, 0, out_buf) != 0) {
        return crypto_fail("crypto.randomBytes: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view *bv = (scriptgo_crypto_buffer_view *)*out_buf;
    if (n > 0 && bv != NULL && bv->data != NULL) {
        if (get_random_bytes(bv->data, n) != 0) {
            return crypto_fail("crypto.randomBytes: failed to generate random bytes");
        }
    }
    return 0;
}

int scriptgo_crypto_random_int(double min_val, double max_val, double *out_val) {
    if (out_val == NULL) return crypto_fail("crypto.randomInt: invalid arguments");
    int64_t min_i = (int64_t)min_val;
    int64_t max_i = (int64_t)max_val;
    if (min_i >= max_i) {
        return crypto_fail("crypto.randomInt: min must be less than max");
    }
    uint64_t range = (uint64_t)(max_i - min_i);
    uint64_t r = 0;
    if (get_random_bytes((unsigned char *)&r, sizeof(r)) != 0) {
        return crypto_fail("crypto.randomInt: CSPRNG error");
    }
    uint64_t val = (r % range);
    *out_val = (double)(min_i + (int64_t)val);
    return 0;
}

int scriptgo_crypto_random_fill(void *buffer_handle, double offset_opt, double size_opt) {
    if (buffer_handle == NULL) return crypto_fail("crypto.randomFillSync: null buffer");
    scriptgo_crypto_buffer_view *bv = (scriptgo_crypto_buffer_view *)buffer_handle;
    if (bv->data == NULL) return 0;
    int64_t off = (int64_t)offset_opt;
    if (off < 0) off = 0;
    int64_t sz = (int64_t)size_opt;
    if (sz <= 0 || off + sz > bv->length) {
        sz = bv->length - off;
    }
    if (sz > 0 && off < bv->length) {
        if (get_random_bytes(bv->data + off, (size_t)sz) != 0) {
            return crypto_fail("crypto.randomFillSync: failed to generate random bytes");
        }
    }
    return 0;
}

int scriptgo_crypto_timing_safe_equal(void *a_handle, void *b_handle, double *out_eq) {
    if (out_eq == NULL) return crypto_fail("crypto.timingSafeEqual: invalid arguments");
    if (a_handle == NULL || b_handle == NULL) {
        *out_eq = 0.0;
        return 0;
    }
    scriptgo_crypto_buffer_view *a = (scriptgo_crypto_buffer_view *)a_handle;
    scriptgo_crypto_buffer_view *b = (scriptgo_crypto_buffer_view *)b_handle;
    if (a->length != b->length) {
        return crypto_fail("crypto.timingSafeEqual: input buffers must have the same length");
    }
    unsigned char result = 0;
    for (int64_t i = 0; i < a->length; i++) {
        result |= (a->data[i] ^ b->data[i]);
    }
    *out_eq = (result == 0) ? 1.0 : 0.0;
    return 0;
}

// -------------------------------------------------------------
// Base64 & Hex Encoding Helpers
// -------------------------------------------------------------
static const char crypto_b64_table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

static char *encode_base64(const unsigned char *data, size_t input_length) {
    size_t output_length = 4 * ((input_length + 2) / 3);
    char *encoded = (char *)malloc(output_length + 1);
    if (encoded == NULL) return NULL;

    size_t i, j = 0;
    for (i = 0; i < input_length;) {
        uint32_t octet_a = i < input_length ? data[i++] : 0;
        uint32_t octet_b = i < input_length ? data[i++] : 0;
        uint32_t octet_c = i < input_length ? data[i++] : 0;
        uint32_t triple = (octet_a << 16) | (octet_b << 8) | octet_c;

        encoded[j++] = crypto_b64_table[(triple >> 18) & 0x3F];
        encoded[j++] = crypto_b64_table[(triple >> 12) & 0x3F];
        encoded[j++] = (i > input_length + 1) ? '=' : crypto_b64_table[(triple >> 6) & 0x3F];
        encoded[j++] = (i > input_length) ? '=' : crypto_b64_table[triple & 0x3F];
    }
    encoded[output_length] = '\0';
    return encoded;
}

static char *encode_hex(const unsigned char *data, size_t len) {
    char *hex = (char *)malloc(len * 2 + 1);
    if (hex == NULL) return NULL;
    for (size_t i = 0; i < len; i++) {
        snprintf(hex + (i * 2), 3, "%02x", data[i]);
    }
    hex[len * 2] = '\0';
    return hex;
}

static char *format_digest_output(const unsigned char *data, size_t len, const char *encoding) {
    if (encoding != NULL && (strcasecmp(encoding, "base64") == 0 || strcasecmp(encoding, "base64url") == 0)) {
        return encode_base64(data, len);
    }
    if (encoding != NULL && (strcasecmp(encoding, "binary") == 0 || strcasecmp(encoding, "latin1") == 0)) {
        char *bin = (char *)malloc(len + 1);
        if (bin == NULL) return NULL;
        memcpy(bin, data, len);
        bin[len] = '\0';
        return bin;
    }
    return encode_hex(data, len);
}

// -------------------------------------------------------------
// MD5 Implementation (RFC 1321)
// -------------------------------------------------------------
typedef struct {
    uint32_t state[4];
    uint32_t count[2];
    unsigned char buffer[64];
} MD5_CTX;

#define F(x, y, z) (((x) & (y)) | ((~x) & (z)))
#define G(x, y, z) (((x) & (z)) | ((y) & (~z)))
#define H(x, y, z) ((x) ^ (y) ^ (z))
#define I(x, y, z) ((y) ^ ((x) | (~z)))
#define ROTATE_LEFT(x, n) (((x) << (n)) | ((x) >> (32-(n))))
#define FF(a, b, c, d, x, s, ac) { (a) += F ((b), (c), (d)) + (x) + (uint32_t)(ac); (a) = ROTATE_LEFT ((a), (s)); (a) += (b); }
#define GG(a, b, c, d, x, s, ac) { (a) += G ((b), (c), (d)) + (x) + (uint32_t)(ac); (a) = ROTATE_LEFT ((a), (s)); (a) += (b); }
#define HH(a, b, c, d, x, s, ac) { (a) += H ((b), (c), (d)) + (x) + (uint32_t)(ac); (a) = ROTATE_LEFT ((a), (s)); (a) += (b); }
#define II(a, b, c, d, x, s, ac) { (a) += I ((b), (c), (d)) + (x) + (uint32_t)(ac); (a) = ROTATE_LEFT ((a), (s)); (a) += (b); }

static void md5_transform(uint32_t state[4], const unsigned char block[64]) {
    uint32_t a = state[0], b = state[1], c = state[2], d = state[3], x[16];
    for (int i = 0, j = 0; i < 16; i++, j += 4)
        x[i] = ((uint32_t)block[j]) | (((uint32_t)block[j+1]) << 8) | (((uint32_t)block[j+2]) << 16) | (((uint32_t)block[j+3]) << 24);

    FF (a, b, c, d, x[ 0], 7, 0xd76aa478); FF (d, a, b, c, x[ 1], 12, 0xe8c7b756); FF (c, d, a, b, x[ 2], 17, 0x242070db); FF (b, c, d, a, x[ 3], 22, 0xc1bdceee);
    FF (a, b, c, d, x[ 4], 7, 0xf57c0faf); FF (d, a, b, c, x[ 5], 12, 0x4787c62a); FF (c, d, a, b, x[ 6], 17, 0xa8304613); FF (b, c, d, a, x[ 7], 22, 0xfd469501);
    FF (a, b, c, d, x[ 8], 7, 0x698098d8); FF (d, a, b, c, x[ 9], 12, 0x8b44f7af); FF (c, d, a, b, x[10], 17, 0xffff5bb1); FF (b, c, d, a, x[11], 22, 0x895cd7be);
    FF (a, b, c, d, x[12], 7, 0x6b901122); FF (d, a, b, c, x[13], 12, 0xfd987193); FF (c, d, a, b, x[14], 17, 0xa679438e); FF (b, c, d, a, x[15], 22, 0x49b40821);

    GG (a, b, c, d, x[ 1], 5, 0xf61e2562); GG (d, a, b, c, x[ 6], 9, 0xc040b340); GG (c, d, a, b, x[11], 14, 0x265e5a51); GG (b, c, d, a, x[ 0], 20, 0xe9b6c7aa);
    GG (a, b, c, d, x[ 5], 5, 0xd62f105d); GG (d, a, b, c, x[10], 9,  0x2441453); GG (c, d, a, b, x[15], 14, 0xd8a1e681); GG (b, c, d, a, x[ 4], 20, 0xe7d3fbc8);
    GG (a, b, c, d, x[ 9], 5, 0x21e1cde6); GG (d, a, b, c, x[14], 9, 0xc33707d6); GG (c, d, a, b, x[ 3], 14, 0xf4d50d87); GG (b, c, d, a, x[ 8], 20, 0x455a14ed);
    GG (a, b, c, d, x[13], 5, 0xa9e3e905); GG (d, a, b, c, x[ 2], 9, 0xfcefa3f8); GG (c, d, a, b, x[ 7], 14, 0x676f02d9); GG (b, c, d, a, x[12], 20, 0x8d2a4c8a);

    HH (a, b, c, d, x[ 5], 4, 0xfffa3942); HH (d, a, b, c, x[ 8], 11, 0x8771f681); HH (c, d, a, b, x[11], 16, 0x6d9d6122); HH (b, c, d, a, x[14], 23, 0xfde5380c);
    HH (a, b, c, d, x[ 1], 4, 0xa4beea44); HH (d, a, b, c, x[ 4], 11, 0x4bdecfa9); HH (c, d, a, b, x[ 7], 16, 0xf6bb4b60); HH (b, c, d, a, x[10], 23, 0xbebfbc70);
    HH (a, b, c, d, x[13], 4, 0x289b7ec6); HH (d, a, b, c, x[ 0], 11, 0xeaa127fa); HH (c, d, a, b, x[ 3], 16, 0xd4ef3085); HH (b, c, d, a, x[ 6], 23,  0x4881d05);
    HH (a, b, c, d, x[ 9], 4, 0xd9d4d039); HH (d, a, b, c, x[12], 11, 0xe6db99e5); HH (c, d, a, b, x[15], 16, 0x1fa27cf8); HH (b, c, d, a, x[ 2], 23, 0xc4ac5665);

    II (a, b, c, d, x[ 0], 6, 0xf4292244); II (d, a, b, c, x[ 7], 10, 0x432aff97); II (c, d, a, b, x[14], 15, 0xab9423a7); II (b, c, d, a, x[ 5], 21, 0xfc93a039);
    II (a, b, c, d, x[12], 6, 0x655b59c3); II (d, a, b, c, x[ 3], 10, 0x8f0ccc92); II (c, d, a, b, x[10], 15, 0xffeff47d); II (b, c, d, a, x[ 1], 21, 0x85845dd1);
    II (a, b, c, d, x[ 8], 6, 0x6fa87e4f); II (d, a, b, c, x[15], 10, 0xfe2ce6e0); II (c, d, a, b, x[ 6], 15, 0xa3014314); II (b, c, d, a, x[13], 21, 0x4e0811a1);
    II (a, b, c, d, x[ 4], 6, 0xf7537e82); II (d, a, b, c, x[11], 10, 0xbd3af235); II (c, d, a, b, x[ 2], 15, 0x2ad7d2bb); II (b, c, d, a, x[ 9], 21, 0xeb86d391);

    state[0] += a; state[1] += b; state[2] += c; state[3] += d;
}

static void md5_init(MD5_CTX *ctx) {
    ctx->count[0] = ctx->count[1] = 0;
    ctx->state[0] = 0x67452301; ctx->state[1] = 0xefcdab89;
    ctx->state[2] = 0x98badcfe; ctx->state[3] = 0x10325476;
}

static void md5_update(MD5_CTX *ctx, const unsigned char *input, size_t inputLen) {
    size_t i, index, partLen;
    index = (size_t)((ctx->count[0] >> 3) & 0x3F);
    if ((ctx->count[0] += ((uint32_t)inputLen << 3)) < ((uint32_t)inputLen << 3))
        ctx->count[1]++;
    ctx->count[1] += ((uint32_t)inputLen >> 29);
    partLen = 64 - index;
    if (inputLen >= partLen) {
        memcpy(&ctx->buffer[index], input, partLen);
        md5_transform(ctx->state, ctx->buffer);
        for (i = partLen; i + 63 < inputLen; i += 64)
            md5_transform(ctx->state, &input[i]);
        index = 0;
    } else i = 0;
    memcpy(&ctx->buffer[index], &input[i], inputLen - i);
}

static void md5_final(unsigned char digest[16], MD5_CTX *ctx) {
    unsigned char bits[8];
    for (int i = 0, j = 0; i < 2; i++, j += 4) {
        bits[j] = (unsigned char)(ctx->count[i] & 0xff);
        bits[j+1] = (unsigned char)((ctx->count[i] >> 8) & 0xff);
        bits[j+2] = (unsigned char)((ctx->count[i] >> 16) & 0xff);
        bits[j+3] = (unsigned char)((ctx->count[i] >> 24) & 0xff);
    }
    size_t index = (size_t)((ctx->count[0] >> 3) & 0x3f);
    size_t padLen = (index < 56) ? (56 - index) : (120 - index);
    static unsigned char PADDING[64] = { 0x80, 0 };
    md5_update(ctx, PADDING, padLen);
    md5_update(ctx, bits, 8);
    for (int i = 0, j = 0; i < 4; i++, j += 4) {
        digest[j] = (unsigned char)(ctx->state[i] & 0xff);
        digest[j+1] = (unsigned char)((ctx->state[i] >> 8) & 0xff);
        digest[j+2] = (unsigned char)((ctx->state[i] >> 16) & 0xff);
        digest[j+3] = (unsigned char)((ctx->state[i] >> 24) & 0xff);
    }
}

// -------------------------------------------------------------
// SHA-1 Implementation (RFC 3174)
// -------------------------------------------------------------
typedef struct {
    uint32_t state[5];
    uint32_t count[2];
    unsigned char buffer[64];
} SHA1_CTX;

static void sha1_transform(uint32_t state[5], const unsigned char buffer[64]) {
    uint32_t a = state[0], b = state[1], c = state[2], d = state[3], e = state[4];
    uint32_t block[80];
    for (int i = 0; i < 16; i++)
        block[i] = ((uint32_t)buffer[i*4] << 24) | ((uint32_t)buffer[i*4+1] << 16) | ((uint32_t)buffer[i*4+2] << 8) | ((uint32_t)buffer[i*4+3]);
    for (int i = 16; i < 80; i++)
        block[i] = ROTATE_LEFT(block[i-3] ^ block[i-8] ^ block[i-14] ^ block[i-16], 1);

    for (int i = 0; i < 80; i++) {
        uint32_t f, k;
        if (i < 20) { f = (b & c) | ((~b) & d); k = 0x5A827999; }
        else if (i < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
        else if (i < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
        else { f = b ^ c ^ d; k = 0xCA62C1D6; }
        uint32_t temp = ROTATE_LEFT(a, 5) + f + e + k + block[i];
        e = d; d = c; c = ROTATE_LEFT(b, 30); b = a; a = temp;
    }
    state[0] += a; state[1] += b; state[2] += c; state[3] += d; state[4] += e;
}

static void sha1_init(SHA1_CTX *ctx) {
    ctx->state[0] = 0x67452301; ctx->state[1] = 0xEFCDAB89;
    ctx->state[2] = 0x98BADCFE; ctx->state[3] = 0x10325476; ctx->state[4] = 0xC3D2E1F0;
    ctx->count[0] = ctx->count[1] = 0;
}

static void sha1_update(SHA1_CTX *ctx, const unsigned char *data, size_t len) {
    size_t i, j = (ctx->count[0] >> 3) & 63;
    if ((ctx->count[0] += (uint32_t)(len << 3)) < (uint32_t)(len << 3)) ctx->count[1]++;
    ctx->count[1] += (uint32_t)(len >> 29);
    if ((j + len) > 63) {
        memcpy(&ctx->buffer[j], data, (i = 64 - j));
        sha1_transform(ctx->state, ctx->buffer);
        for (; i + 63 < len; i += 64) sha1_transform(ctx->state, &data[i]);
        j = 0;
    } else i = 0;
    memcpy(&ctx->buffer[j], &data[i], len - i);
}

static void sha1_final(unsigned char digest[20], SHA1_CTX *ctx) {
    unsigned char finalcount[8];
    for (int i = 0; i < 8; i++)
        finalcount[i] = (unsigned char)((ctx->count[(i >= 4 ? 0 : 1)] >> ((3 - (i & 3)) * 8)) & 255);
    sha1_update(ctx, (const unsigned char *)"\200", 1);
    while ((ctx->count[0] & 504) != 448) sha1_update(ctx, (const unsigned char *)"\0", 1);
    sha1_update(ctx, finalcount, 8);
    for (int i = 0; i < 20; i++)
        digest[i] = (unsigned char)((ctx->state[i >> 2] >> ((3 - (i & 3)) * 8)) & 255);
}

// -------------------------------------------------------------
// SHA-256 Implementation
// -------------------------------------------------------------
typedef struct {
    unsigned char data[64];
    unsigned int datalen;
    unsigned long long bitlen;
    unsigned int state[8];
} SHA256_CTX;

#define ROTRIGHT(a,b) (((a) >> (b)) | ((a) << (32-(b))))
#define CH_256(x,y,z) (((x) & (y)) ^ (~(x) & (z)))
#define MAJ_256(x,y,z) (((x) & (y)) ^ ((x) & (z)) ^ ((y) & (z)))
#define EP0_256(x) (ROTRIGHT(x,2) ^ ROTRIGHT(x,13) ^ ROTRIGHT(x,22))
#define EP1_256(x) (ROTRIGHT(x,6) ^ ROTRIGHT(x,11) ^ ROTRIGHT(x,25))
#define SIG0_256(x) (ROTRIGHT(x,7) ^ ROTRIGHT(x,18) ^ ((x) >> 3))
#define SIG1_256(x) (ROTRIGHT(x,17) ^ ROTRIGHT(x,19) ^ ((x) >> 10))

static const unsigned int k256[64] = {
    0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,
    0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,
    0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,
    0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,
    0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,
    0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,
    0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,
    0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2
};

static void sha256_transform(SHA256_CTX *ctx, const unsigned char data[]) {
    unsigned int a, b, c, d, e, f, g, h, i, j, t1, t2, m[64];
    for (i = 0, j = 0; i < 16; ++i, j += 4)
        m[i] = ((unsigned int)data[j] << 24) | ((unsigned int)data[j + 1] << 16) | ((unsigned int)data[j + 2] << 8) | ((unsigned int)data[j + 3]);
    for ( ; i < 64; ++i)
        m[i] = SIG1_256(m[i - 2]) + m[i - 7] + SIG0_256(m[i - 15]) + m[i - 16];

    a = ctx->state[0]; b = ctx->state[1]; c = ctx->state[2]; d = ctx->state[3];
    e = ctx->state[4]; f = ctx->state[5]; g = ctx->state[6]; h = ctx->state[7];

    for (i = 0; i < 64; ++i) {
        t1 = h + EP1_256(e) + CH_256(e,f,g) + k256[i] + m[i];
        t2 = EP0_256(a) + MAJ_256(a,b,c);
        h = g; g = f; f = e; e = d + t1;
        d = c; c = b; b = a; a = t1 + t2;
    }
    ctx->state[0] += a; ctx->state[1] += b; ctx->state[2] += c; ctx->state[3] += d;
    ctx->state[4] += e; ctx->state[5] += f; ctx->state[6] += g; ctx->state[7] += h;
}

static void sha256_init(SHA256_CTX *ctx) {
    ctx->datalen = 0; ctx->bitlen = 0;
    ctx->state[0] = 0x6a09e667; ctx->state[1] = 0xbb67ae85;
    ctx->state[2] = 0x3c6ef372; ctx->state[3] = 0xa54ff53a;
    ctx->state[4] = 0x510e527f; ctx->state[5] = 0x9b05688c;
    ctx->state[6] = 0x1f83d9ab; ctx->state[7] = 0x5be0cd19;
}

static void sha256_update(SHA256_CTX *ctx, const unsigned char data[], size_t len) {
    for (size_t i = 0; i < len; ++i) {
        ctx->data[ctx->datalen] = data[i];
        ctx->datalen++;
        if (ctx->datalen == 64) {
            sha256_transform(ctx, ctx->data);
            ctx->bitlen += 512;
            ctx->datalen = 0;
        }
    }
}

static void sha256_final(unsigned char hash[32], SHA256_CTX *ctx) {
    unsigned int i = ctx->datalen;
    if (ctx->datalen < 56) {
        ctx->data[i++] = 0x80;
        while (i < 56) ctx->data[i++] = 0x00;
    } else {
        ctx->data[i++] = 0x80;
        while (i < 64) ctx->data[i++] = 0x00;
        sha256_transform(ctx, ctx->data);
        memset(ctx->data, 0, 56);
    }
    ctx->bitlen += (unsigned long long)ctx->datalen * 8;
    ctx->data[63] = (unsigned char)(ctx->bitlen);
    ctx->data[62] = (unsigned char)(ctx->bitlen >> 8);
    ctx->data[61] = (unsigned char)(ctx->bitlen >> 16);
    ctx->data[60] = (unsigned char)(ctx->bitlen >> 24);
    ctx->data[59] = (unsigned char)(ctx->bitlen >> 32);
    ctx->data[58] = (unsigned char)(ctx->bitlen >> 40);
    ctx->data[57] = (unsigned char)(ctx->bitlen >> 48);
    ctx->data[56] = (unsigned char)(ctx->bitlen >> 56);
    sha256_transform(ctx, ctx->data);
    for (i = 0; i < 4; ++i) {
        hash[i]      = (unsigned char)((ctx->state[0] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 4]  = (unsigned char)((ctx->state[1] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 8]  = (unsigned char)((ctx->state[2] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 12] = (unsigned char)((ctx->state[3] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 16] = (unsigned char)((ctx->state[4] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 20] = (unsigned char)((ctx->state[5] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 24] = (unsigned char)((ctx->state[6] >> (24 - i * 8)) & 0x000000ff);
        hash[i + 28] = (unsigned char)((ctx->state[7] >> (24 - i * 8)) & 0x000000ff);
    }
}

// -------------------------------------------------------------
// SHA-512 Implementation (FIPS PUB 180-2)
// -------------------------------------------------------------
typedef struct {
    uint64_t state[8];
    uint64_t bitcount[2];
    unsigned char buffer[128];
} SHA512_CTX;

#define ROTR64(x, n) (((x) >> (n)) | ((x) << (64 - (n))))
#define CH_512(x, y, z) (((x) & (y)) ^ (~(x) & (z)))
#define MAJ_512(x, y, z) (((x) & (y)) ^ ((x) & (z)) ^ ((y) & (z)))
#define SIGMA0_512(x) (ROTR64(x, 28) ^ ROTR64(x, 34) ^ ROTR64(x, 39))
#define SIGMA1_512(x) (ROTR64(x, 14) ^ ROTR64(x, 18) ^ ROTR64(x, 41))
#define SIG0_512(x) (ROTR64(x, 1) ^ ROTR64(x, 8) ^ ((x) >> 7))
#define SIG1_512(x) (ROTR64(x, 19) ^ ROTR64(x, 61) ^ ((x) >> 6))

static const uint64_t K512[80] = {
    0x428a2f98d728ae22ULL, 0x7137449123ef65cdULL, 0xb5c0fbcfec4d3b2fULL, 0xe9b5dba58189dbbcULL,
    0x3956c25bf348b538ULL, 0x59f111f1b605d019ULL, 0x923f82a4af194f9bULL, 0xab1c5ed5da6d8118ULL,
    0xd807aa98a3030242ULL, 0x12835b0145706fbeULL, 0x243185be4ee4b28cULL, 0x550c7dc3d5ffb4e2ULL,
    0x72be5d74f27b896fULL, 0x80deb1fe3b1696b1ULL, 0x9bdc06a725c71235ULL, 0xc19bf174cf692694ULL,
    0xe49b69c19ef14ad2ULL, 0xefbe4786384f25e3ULL, 0x0fc19dc68b8cd5b5ULL, 0x240ca1cc77ac9c65ULL,
    0x2de92c6f592b0275ULL, 0x4a7484aa6ea6e483ULL, 0x5cb0a9dcbd41fbd4ULL, 0x76f988da831153b5ULL,
    0x983e5152ee66dfabULL, 0xa831c66d2db43210ULL, 0xb00327c898fb213fULL, 0xbf597fc7beef0ee4ULL,
    0xc6e00bf33da88fc2ULL, 0xd5a79147930aa725ULL, 0x06ca6351e003826fULL, 0x142929670a0e6e70ULL,
    0x27b70a8546d22ffcULL, 0x2e1b21385c26c926ULL, 0x4d2c6dfc5ac42aedULL, 0x53380d139d95b3dfULL,
    0x650a73548baf63deULL, 0x766a0abb3c77b2a8ULL, 0x81c2c92e47867a6eULL, 0x92722c851482353bULL,
    0xa2bfe8a14cf10364ULL, 0xa81a664bbc423001ULL, 0xc24b8b70d0f89791ULL, 0xc76c51a30654be30ULL,
    0xd192e819d6ef5218ULL, 0xd69906245565a910ULL, 0xf40e35855771202aULL, 0x106aa07032bbd1b8ULL,
    0x19a4c116b8d2d0c8ULL, 0x1e376c085141ab53ULL, 0x2748774cdf8eeb99ULL, 0x34b0bcb5e19b48a8ULL,
    0x391c0cb3c5c95a63ULL, 0x4ed8aa4ae3418acbULL, 0x5b9cca4f7763e373ULL, 0x682e6ff3d6b2b8a3ULL,
    0x748f82ee5defb2fcULL, 0x78a5636f43172f60ULL, 0x84c87814a1f0ab72ULL, 0x8cc702081a6439ecULL,
    0x90befffa23631e28ULL, 0xa4506cebde82bde9ULL, 0xbef9a3f7b2c67915ULL, 0xc67178f2e372532bULL,
    0xca273eceea26619cULL, 0xd186b8c721c0c207ULL, 0xeada7dd6cde0eb1eULL, 0xf57d4f7fee6ed178ULL,
    0x06f067aa72176fbaULL, 0x0a637dc5a2c898a6ULL, 0x113f9804bef90daeULL, 0x1b710b35131c471bULL,
    0x28db77f523047d84ULL, 0x32caab7b40c72493ULL, 0x3c9ebe0a15c9bebcULL, 0x431d67c49c100d4cULL,
    0x4cc5d4becb3e42b6ULL, 0x597f299cfc657e2aULL, 0x5fcb6fab3ad6faecULL, 0x6c44198c4a475817ULL
};

static void sha512_transform(SHA512_CTX *ctx, const unsigned char *buffer) {
    uint64_t a = ctx->state[0], b = ctx->state[1], c = ctx->state[2], d = ctx->state[3];
    uint64_t e = ctx->state[4], f = ctx->state[5], g = ctx->state[6], h = ctx->state[7];
    uint64_t W[80];
    for (int i = 0; i < 16; i++) {
        W[i] = ((uint64_t)buffer[i*8] << 56) | ((uint64_t)buffer[i*8+1] << 48) |
               ((uint64_t)buffer[i*8+2] << 40) | ((uint64_t)buffer[i*8+3] << 32) |
               ((uint64_t)buffer[i*8+4] << 24) | ((uint64_t)buffer[i*8+5] << 16) |
               ((uint64_t)buffer[i*8+6] << 8) | ((uint64_t)buffer[i*8+7]);
    }
    for (int i = 16; i < 80; i++) {
        W[i] = SIG1_512(W[i-2]) + W[i-7] + SIG0_512(W[i-15]) + W[i-16];
    }
    for (int i = 0; i < 80; i++) {
        uint64_t T1 = h + SIGMA1_512(e) + CH_512(e, f, g) + K512[i] + W[i];
        uint64_t T2 = SIGMA0_512(a) + MAJ_512(a, b, c);
        h = g; g = f; f = e; e = d + T1;
        d = c; c = b; b = a; a = T1 + T2;
    }
    ctx->state[0] += a; ctx->state[1] += b; ctx->state[2] += c; ctx->state[3] += d;
    ctx->state[4] += e; ctx->state[5] += f; ctx->state[6] += g; ctx->state[7] += h;
}

static void sha512_init(SHA512_CTX *ctx) {
    ctx->state[0] = 0x6a09e667f3bcc908ULL; ctx->state[1] = 0xbb67ae8584caa73bULL;
    ctx->state[2] = 0x3c6ef372fe94f82bULL; ctx->state[3] = 0xa54ff53a5f1d36f1ULL;
    ctx->state[4] = 0x510e527fade682d1ULL; ctx->state[5] = 0x9b05688c2b3e6c1fULL;
    ctx->state[6] = 0x1f83d9abfb41bd6bULL; ctx->state[7] = 0x5be0cd19137e2179ULL;
    ctx->bitcount[0] = ctx->bitcount[1] = 0;
}

static void sha512_update(SHA512_CTX *ctx, const unsigned char *data, size_t len) {
    size_t left = (ctx->bitcount[0] >> 3) & 127;
    size_t fill = 128 - left;
    ctx->bitcount[0] += (uint64_t)len << 3;
    if (ctx->bitcount[0] < ((uint64_t)len << 3)) ctx->bitcount[1]++;
    ctx->bitcount[1] += (uint64_t)len >> 61;
    if (left && len >= fill) {
        memcpy(ctx->buffer + left, data, fill);
        sha512_transform(ctx, ctx->buffer);
        data += fill; len -= fill; left = 0;
    }
    while (len >= 128) {
        sha512_transform(ctx, data);
        data += 128; len -= 128;
    }
    if (len) memcpy(ctx->buffer + left, data, len);
}

static void sha512_final(unsigned char digest[64], SHA512_CTX *ctx) {
    size_t last = (ctx->bitcount[0] >> 3) & 127;
    size_t padn = (last < 112) ? (112 - last) : (240 - last);
    static const unsigned char PADDING[128] = { 0x80 };
    sha512_update(ctx, PADDING, padn);
    unsigned char bits[16];
    for (int i = 0; i < 8; i++) {
        bits[i] = (unsigned char)((ctx->bitcount[1] >> ((7 - i) * 8)) & 255);
        bits[i+8] = (unsigned char)((ctx->bitcount[0] >> ((7 - i) * 8)) & 255);
    }
    sha512_update(ctx, bits, 16);
    for (int i = 0; i < 8; i++) {
        for (int j = 0; j < 8; j++) {
            digest[i*8 + j] = (unsigned char)((ctx->state[i] >> ((7 - j) * 8)) & 255);
        }
    }
}

// -------------------------------------------------------------
// Generic Hash Digest Function
// -------------------------------------------------------------
static int compute_raw_hash(const char *algo, const unsigned char *data, size_t len, unsigned char *out_hash, size_t *out_len) {
    if (strcasecmp(algo, "sha256") == 0 || strcasecmp(algo, "sha-256") == 0) {
        SHA256_CTX ctx;
        sha256_init(&ctx);
        sha256_update(&ctx, data, len);
        sha256_final(out_hash, &ctx);
        *out_len = 32;
        return 0;
    }
    if (strcasecmp(algo, "sha512") == 0 || strcasecmp(algo, "sha-512") == 0) {
        SHA512_CTX ctx;
        sha512_init(&ctx);
        sha512_update(&ctx, data, len);
        sha512_final(out_hash, &ctx);
        *out_len = 64;
        return 0;
    }
    if (strcasecmp(algo, "sha1") == 0 || strcasecmp(algo, "sha-1") == 0) {
        SHA1_CTX ctx;
        sha1_init(&ctx);
        sha1_update(&ctx, data, len);
        sha1_final(out_hash, &ctx);
        *out_len = 20;
        return 0;
    }
    if (strcasecmp(algo, "md5") == 0) {
        MD5_CTX ctx;
        md5_init(&ctx);
        md5_update(&ctx, data, len);
        md5_final(out_hash, &ctx);
        *out_len = 16;
        return 0;
    }
    return -1;
}

int scriptgo_crypto_hash_digest(const char *algo, const char *data, const char *encoding, char **out_digest) {
    if (algo == NULL || data == NULL || out_digest == NULL) {
        return crypto_fail("crypto.createHash: invalid arguments");
    }
    unsigned char hash[64];
    size_t hash_len = 0;
    if (compute_raw_hash(algo, (const unsigned char *)data, strlen(data), hash, &hash_len) != 0) {
        return crypto_fail("crypto.createHash: unsupported digest algorithm");
    }
    char *res = format_digest_output(hash, hash_len, encoding);
    if (res == NULL) {
        return crypto_fail("crypto.createHash: format output failed");
    }
    *out_digest = res;
    return 0;
}

int scriptgo_crypto_hash_digest_buffer(const char *algo, void *data_handle, const char *encoding, char **out_digest) {
    if (algo == NULL || data_handle == NULL || out_digest == NULL) {
        return crypto_fail("crypto.createHash: invalid arguments");
    }
    scriptgo_crypto_buffer_view *view = (scriptgo_crypto_buffer_view *)data_handle;
    if (view->length < 0 || (view->length > 0 && view->data == NULL)) {
        return crypto_fail("crypto.createHash: invalid input buffer");
    }
    unsigned char hash[64];
    size_t hash_len = 0;
    if (compute_raw_hash(algo, view->data, (size_t)view->length, hash, &hash_len) != 0) {
        return crypto_fail("crypto.createHash: unsupported digest algorithm");
    }
    char *res = format_digest_output(hash, hash_len, encoding);
    if (res == NULL) return crypto_fail("crypto.createHash: format output failed");
    *out_digest = res;
    return 0;
}

// -------------------------------------------------------------
// HMAC Implementation
// -------------------------------------------------------------
static int compute_raw_hmac(const char *algo, const unsigned char *key, size_t key_len, const unsigned char *data, size_t data_len, unsigned char *out_hmac, size_t *out_len) {
    size_t block_size = 64;
    if (strcasecmp(algo, "sha512") == 0 || strcasecmp(algo, "sha-512") == 0) {
        block_size = 128;
    }
    unsigned char k_pad[128];
    memset(k_pad, 0, sizeof(k_pad));

    if (key_len > block_size) {
        size_t temp_len = 0;
        if (compute_raw_hash(algo, key, key_len, k_pad, &temp_len) != 0) return -1;
    } else {
        memcpy(k_pad, key, key_len);
    }

    unsigned char o_key_pad[128];
    unsigned char i_key_pad[128];
    for (size_t i = 0; i < block_size; i++) {
        o_key_pad[i] = k_pad[i] ^ 0x5c;
        i_key_pad[i] = k_pad[i] ^ 0x36;
    }

    // Inner hash = H(i_key_pad || data)
    size_t inner_in_len = block_size + data_len;
    unsigned char *inner_in = (unsigned char *)malloc(inner_in_len);
    if (inner_in == NULL) return -1;
    memcpy(inner_in, i_key_pad, block_size);
    memcpy(inner_in + block_size, data, data_len);

    unsigned char inner_hash[64];
    size_t inner_len = 0;
    if (compute_raw_hash(algo, inner_in, inner_in_len, inner_hash, &inner_len) != 0) {
        free(inner_in);
        return -1;
    }
    free(inner_in);

    // Outer hash = H(o_key_pad || inner_hash)
    size_t outer_in_len = block_size + inner_len;
    unsigned char *outer_in = (unsigned char *)malloc(outer_in_len);
    if (outer_in == NULL) return -1;
    memcpy(outer_in, o_key_pad, block_size);
    memcpy(outer_in + block_size, inner_hash, inner_len);

    if (compute_raw_hash(algo, outer_in, outer_in_len, out_hmac, out_len) != 0) {
        free(outer_in);
        return -1;
    }
    free(outer_in);
    return 0;
}

int scriptgo_crypto_hmac_digest(const char *algo, const char *key, const char *data, const char *encoding, char **out_digest) {
    if (algo == NULL || key == NULL || data == NULL || out_digest == NULL) {
        return crypto_fail("crypto.createHmac: invalid arguments");
    }
    unsigned char hmac[64];
    size_t hmac_len = 0;
    if (compute_raw_hmac(algo, (const unsigned char *)key, strlen(key), (const unsigned char *)data, strlen(data), hmac, &hmac_len) != 0) {
        return crypto_fail("crypto.createHmac: computation failed");
    }
    char *res = format_digest_output(hmac, hmac_len, encoding);
    if (res == NULL) {
        return crypto_fail("crypto.createHmac: format output failed");
    }
    *out_digest = res;
    return 0;
}

int scriptgo_crypto_hmac_digest_buffer(const char *algo, void *key_handle, void *data_handle, const char *encoding, char **out_digest) {
    if (algo == NULL || key_handle == NULL || data_handle == NULL || out_digest == NULL) {
        return crypto_fail("crypto.createHmac: invalid arguments");
    }
    scriptgo_crypto_buffer_view *key = (scriptgo_crypto_buffer_view *)key_handle;
    scriptgo_crypto_buffer_view *data = (scriptgo_crypto_buffer_view *)data_handle;
    if (key->length < 0 || data->length < 0 ||
        (key->length > 0 && key->data == NULL) || (data->length > 0 && data->data == NULL)) {
        return crypto_fail("crypto.createHmac: invalid input buffer");
    }
    unsigned char hmac[64];
    size_t hmac_len = 0;
    if (compute_raw_hmac(algo, key->data, (size_t)key->length, data->data, (size_t)data->length, hmac, &hmac_len) != 0) {
        return crypto_fail("crypto.createHmac: computation failed");
    }
    char *res = format_digest_output(hmac, hmac_len, encoding);
    if (res == NULL) return crypto_fail("crypto.createHmac: format output failed");
    *out_digest = res;
    return 0;
}

// -------------------------------------------------------------
// PBKDF2 Implementation
// -------------------------------------------------------------
int scriptgo_crypto_pbkdf2_sync(const char *password, const char *salt, double iterations, double keylen, const char *digest, void **out_buffer) {
    if (password == NULL || salt == NULL || out_buffer == NULL || keylen <= 0 || iterations <= 0) {
        return crypto_fail("crypto.pbkdf2Sync: invalid arguments");
    }
    const char *algo = (digest != NULL && *digest != '\0') ? digest : "sha1";
    size_t dk_len = (size_t)keylen;
    int iter = (int)iterations;

    if (scriptgo_buffer_alloc((double)dk_len, NULL, 0, 0, 0, out_buffer) != 0) {
        return crypto_fail("crypto.pbkdf2Sync: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view *bv = (scriptgo_crypto_buffer_view *)*out_buffer;
    unsigned char *dk = bv->data;

    size_t h_len = 0;
    unsigned char dummy[64];
    if (compute_raw_hash(algo, (const unsigned char *)"", 0, dummy, &h_len) != 0) {
        return crypto_fail("crypto.pbkdf2Sync: unsupported digest algorithm");
    }

    size_t salt_len = strlen(salt);
    size_t pass_len = strlen(password);
    uint32_t block_count = (uint32_t)((dk_len + h_len - 1) / h_len);
    unsigned char *u_block = (unsigned char *)malloc(h_len);
    unsigned char *f_block = (unsigned char *)malloc(h_len);
    unsigned char *salt_and_idx = (unsigned char *)malloc(salt_len + 4);
    if (u_block == NULL || f_block == NULL || salt_and_idx == NULL) {
        free(u_block); free(f_block); free(salt_and_idx);
        return crypto_fail("crypto.pbkdf2Sync: memory allocation failed");
    }

    memcpy(salt_and_idx, salt, salt_len);
    for (uint32_t b = 1; b <= block_count; b++) {
        salt_and_idx[salt_len] = (unsigned char)((b >> 24) & 0xff);
        salt_and_idx[salt_len + 1] = (unsigned char)((b >> 16) & 0xff);
        salt_and_idx[salt_len + 2] = (unsigned char)((b >> 8) & 0xff);
        salt_and_idx[salt_len + 3] = (unsigned char)(b & 0xff);

        size_t u_len = 0;
        compute_raw_hmac(algo, (const unsigned char *)password, pass_len, salt_and_idx, salt_len + 4, u_block, &u_len);
        memcpy(f_block, u_block, h_len);

        for (int c = 1; c < iter; c++) {
            compute_raw_hmac(algo, (const unsigned char *)password, pass_len, u_block, h_len, u_block, &u_len);
            for (size_t k = 0; k < h_len; k++) {
                f_block[k] ^= u_block[k];
            }
        }
        size_t offset = (b - 1) * h_len;
        size_t to_copy = (offset + h_len > dk_len) ? (dk_len - offset) : h_len;
        memcpy(dk + offset, f_block, to_copy);
    }

    free(u_block);
    free(f_block);
    free(salt_and_idx);
    return 0;
}

typedef struct {
    int64_t byte_length;
    unsigned char *data;
} scriptgo_crypto_array_buffer;

int scriptgo_arraybuffer_new(int64_t byte_length, void **out_buffer);

int scriptgo_crypto_hkdf_sync(const char *digest, const char *ikm, const char *salt, const char *info, double keylen, void **out_buffer) {
    const char *algo = (digest != NULL && *digest != '\0') ? digest : "sha256";
    size_t output_len = (size_t)keylen;
    unsigned char hash_probe[64];
    size_t hash_len = 0;
    if (ikm == NULL || salt == NULL || info == NULL || out_buffer == NULL || keylen < 0 || keylen != keylen) {
        return crypto_fail("crypto.hkdfSync: invalid arguments");
    }
    if (compute_raw_hash(algo, (const unsigned char *)"", 0, hash_probe, &hash_len) != 0 || hash_len == 0) {
        return crypto_fail("crypto.hkdfSync: unsupported digest algorithm");
    }
    if (output_len > 255 * hash_len) {
        return crypto_fail("crypto.hkdfSync: key length too large");
    }
    if (scriptgo_arraybuffer_new((int64_t)output_len, out_buffer) != 0) {
        return crypto_fail("crypto.hkdfSync: buffer allocation failed");
    }
    scriptgo_crypto_array_buffer *buffer = (scriptgo_crypto_array_buffer *)*out_buffer;
    unsigned char zero_salt[64];
    unsigned char prk[64];
    memset(zero_salt, 0, sizeof(zero_salt));
    const unsigned char *salt_bytes = (const unsigned char *)salt;
    size_t salt_len = strlen(salt);
    if (salt_len == 0) { salt_bytes = zero_salt; salt_len = hash_len; }
    if (compute_raw_hmac(algo, salt_bytes, salt_len, (const unsigned char *)ikm, strlen(ikm), prk, &hash_len) != 0) {
        return crypto_fail("crypto.hkdfSync: extract failed");
    }
    unsigned char previous[64];
    size_t previous_len = 0;
    size_t produced = 0;
    size_t info_len = strlen(info);
    for (unsigned int block = 1; produced < output_len; block++) {
        size_t input_len = previous_len + info_len + 1;
        unsigned char *input = (unsigned char *)malloc(input_len);
        if (input == NULL) return crypto_fail("crypto.hkdfSync: allocation failed");
        memcpy(input, previous, previous_len);
        memcpy(input + previous_len, info, info_len);
        input[input_len - 1] = (unsigned char)block;
        if (compute_raw_hmac(algo, prk, hash_len, input, input_len, previous, &previous_len) != 0) {
            free(input);
            return crypto_fail("crypto.hkdfSync: expand failed");
        }
        free(input);
        size_t copy_len = output_len - produced < previous_len ? output_len - produced : previous_len;
        if (copy_len > 0) memcpy(buffer->data + produced, previous, copy_len);
        produced += copy_len;
    }
    return 0;
}

static int crypto_pbkdf2_raw(const char *algo, const unsigned char *password, size_t password_len,
                             const unsigned char *salt, size_t salt_len, int iterations,
                             unsigned char *derived, size_t derived_len) {
    unsigned char probe[64];
    size_t hash_len = 0;
    if (iterations <= 0 || compute_raw_hash(algo, (const unsigned char *)"", 0, probe, &hash_len) != 0 || hash_len == 0) return -1;
    size_t block_count = (derived_len + hash_len - 1) / hash_len;
    unsigned char *u = malloc(hash_len);
    unsigned char *f = malloc(hash_len);
    unsigned char *salt_block = malloc(salt_len + 4);
    if (u == NULL || f == NULL || salt_block == NULL) { free(u); free(f); free(salt_block); return -1; }
    memcpy(salt_block, salt, salt_len);
    for (size_t block = 1; block <= block_count; block++) {
        salt_block[salt_len] = (unsigned char)((block >> 24) & 0xff);
        salt_block[salt_len + 1] = (unsigned char)((block >> 16) & 0xff);
        salt_block[salt_len + 2] = (unsigned char)((block >> 8) & 0xff);
        salt_block[salt_len + 3] = (unsigned char)(block & 0xff);
        size_t u_len = 0;
        if (compute_raw_hmac(algo, password, password_len, salt_block, salt_len + 4, u, &u_len) != 0) { free(u); free(f); free(salt_block); return -1; }
        memcpy(f, u, hash_len);
        for (int iteration = 1; iteration < iterations; iteration++) {
            if (compute_raw_hmac(algo, password, password_len, u, hash_len, u, &u_len) != 0) { free(u); free(f); free(salt_block); return -1; }
            for (size_t i = 0; i < hash_len; i++) f[i] ^= u[i];
        }
        size_t offset = (block - 1) * hash_len;
        size_t copy_len = derived_len - offset < hash_len ? derived_len - offset : hash_len;
        memcpy(derived + offset, f, copy_len);
    }
    free(u); free(f); free(salt_block);
    return 0;
}

static uint32_t crypto_load32(const unsigned char *p) {
    return ((uint32_t)p[0]) | ((uint32_t)p[1] << 8) | ((uint32_t)p[2] << 16) | ((uint32_t)p[3] << 24);
}

static void crypto_store32(unsigned char *p, uint32_t value) {
    p[0] = (unsigned char)value;
    p[1] = (unsigned char)(value >> 8);
    p[2] = (unsigned char)(value >> 16);
    p[3] = (unsigned char)(value >> 24);
}

static uint32_t crypto_rotl32(uint32_t value, int shift) {
    return (value << shift) | (value >> (32 - shift));
}

static void crypto_salsa20_8(unsigned char block[64]) {
    uint32_t x[16], original[16];
    for (int i = 0; i < 16; i++) x[i] = original[i] = crypto_load32(block + i * 4);
    for (int round = 0; round < 8; round += 2) {
        x[4] ^= crypto_rotl32(x[0] + x[12], 7); x[8] ^= crypto_rotl32(x[4] + x[0], 9);
        x[12] ^= crypto_rotl32(x[8] + x[4], 13); x[0] ^= crypto_rotl32(x[12] + x[8], 18);
        x[9] ^= crypto_rotl32(x[5] + x[1], 7); x[13] ^= crypto_rotl32(x[9] + x[5], 9);
        x[1] ^= crypto_rotl32(x[13] + x[9], 13); x[5] ^= crypto_rotl32(x[1] + x[13], 18);
        x[14] ^= crypto_rotl32(x[10] + x[6], 7); x[2] ^= crypto_rotl32(x[14] + x[10], 9);
        x[6] ^= crypto_rotl32(x[2] + x[14], 13); x[10] ^= crypto_rotl32(x[6] + x[2], 18);
        x[3] ^= crypto_rotl32(x[15] + x[11], 7); x[7] ^= crypto_rotl32(x[3] + x[15], 9);
        x[11] ^= crypto_rotl32(x[7] + x[3], 13); x[15] ^= crypto_rotl32(x[11] + x[7], 18);
        x[1] ^= crypto_rotl32(x[0] + x[3], 7); x[2] ^= crypto_rotl32(x[1] + x[0], 9);
        x[3] ^= crypto_rotl32(x[2] + x[1], 13); x[0] ^= crypto_rotl32(x[3] + x[2], 18);
        x[6] ^= crypto_rotl32(x[5] + x[4], 7); x[7] ^= crypto_rotl32(x[6] + x[5], 9);
        x[4] ^= crypto_rotl32(x[7] + x[6], 13); x[5] ^= crypto_rotl32(x[4] + x[7], 18);
        x[11] ^= crypto_rotl32(x[10] + x[9], 7); x[8] ^= crypto_rotl32(x[11] + x[10], 9);
        x[9] ^= crypto_rotl32(x[8] + x[11], 13); x[10] ^= crypto_rotl32(x[9] + x[8], 18);
        x[12] ^= crypto_rotl32(x[15] + x[14], 7); x[13] ^= crypto_rotl32(x[12] + x[15], 9);
        x[14] ^= crypto_rotl32(x[13] + x[12], 13); x[15] ^= crypto_rotl32(x[14] + x[13], 18);
    }
    for (int i = 0; i < 16; i++) crypto_store32(block + i * 4, x[i] + original[i]);
}

static void crypto_blockmix(const unsigned char *input, unsigned char *output, size_t r) {
    size_t blocks = 2 * r;
    unsigned char x[64];
    unsigned char *y = malloc(blocks * 64);
    if (y == NULL) return;
    memcpy(x, input + (blocks - 1) * 64, 64);
    for (size_t i = 0; i < blocks; i++) {
        for (size_t j = 0; j < 64; j++) x[j] ^= input[i * 64 + j];
        crypto_salsa20_8(x);
        memcpy(y + i * 64, x, 64);
    }
    for (size_t i = 0; i < r; i++) memcpy(output + i * 64, y + (2 * i) * 64, 64);
    for (size_t i = 0; i < r; i++) memcpy(output + (r + i) * 64, y + (2 * i + 1) * 64, 64);
    free(y);
}

static uint64_t crypto_integerify(const unsigned char *block, size_t r) {
    const unsigned char *last = block + (2 * r - 1) * 64;
    uint64_t value = 0;
    for (int i = 0; i < 8; i++) value |= ((uint64_t)last[i]) << (8 * i);
    return value;
}

int scriptgo_crypto_scrypt_sync(const char *password, const char *salt, double keylen, void **out_buffer) {
    const size_t n = 16384, r = 8, p = 1;
    const size_t block_len = 128 * r;
    size_t output_len = (size_t)keylen;
    if (password == NULL || salt == NULL || out_buffer == NULL || keylen < 0 || keylen != keylen || output_len > 1024 * 1024) {
        return crypto_fail("crypto.scryptSync: invalid arguments");
    }
    if (scriptgo_buffer_alloc((double)output_len, NULL, 0, 0, 0, out_buffer) != 0) return crypto_fail("crypto.scryptSync: buffer allocation failed");
    scriptgo_crypto_buffer_view *result = (scriptgo_crypto_buffer_view *)*out_buffer;
    size_t b_len = block_len * p;
    unsigned char *b = malloc(b_len);
    unsigned char *v = malloc(block_len * n);
    unsigned char *x = malloc(block_len);
    unsigned char *y = malloc(block_len);
    if (b == NULL || v == NULL || x == NULL || y == NULL) { free(b); free(v); free(x); free(y); return crypto_fail("crypto.scryptSync: memory allocation failed"); }
    if (crypto_pbkdf2_raw("sha256", (const unsigned char *)password, strlen(password), (const unsigned char *)salt, strlen(salt), 1, b, b_len) != 0) {
        free(b); free(v); free(x); free(y); return crypto_fail("crypto.scryptSync: initial derivation failed");
    }
    for (size_t part = 0; part < p; part++) {
        unsigned char *part_b = b + part * block_len;
        memcpy(x, part_b, block_len);
        for (size_t i = 0; i < n; i++) { memcpy(v + i * block_len, x, block_len); crypto_blockmix(x, y, r); unsigned char *tmp = x; x = y; y = tmp; }
        for (size_t i = 0; i < n; i++) {
            size_t j = (size_t)(crypto_integerify(x, r) & (n - 1));
            for (size_t k = 0; k < block_len; k++) x[k] ^= v[j * block_len + k];
            crypto_blockmix(x, y, r); unsigned char *tmp = x; x = y; y = tmp;
        }
        memcpy(part_b, x, block_len);
    }
    if (crypto_pbkdf2_raw("sha256", (const unsigned char *)password, strlen(password), b, b_len, 1, result->data, output_len) != 0) {
        free(b); free(v); free(x); free(y); return crypto_fail("crypto.scryptSync: final derivation failed");
    }
    free(b); free(v); free(x); free(y);
    return 0;
}
