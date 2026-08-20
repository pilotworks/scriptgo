#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__)
#define HAVE_ARC4RANDOM 1
#endif

int scriptgo_runtime_set_error(const char *message);

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
    // Set version 4 (0100xxxx)
    b[6] = (b[6] & 0x0f) | 0x40;
    // Set variant to RFC 4122 (10xxxxxx)
    b[8] = (b[8] & 0x3f) | 0x80;

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

// SHA-256 implementation
typedef struct {
    unsigned char data[64];
    unsigned int datalen;
    unsigned long long bitlen;
    unsigned int state[8];
} SHA256_CTX;

#define ROTLEFT(a,b) (((a) << (b)) | ((a) >> (32-(b))))
#define ROTRIGHT(a,b) (((a) >> (b)) | ((a) << (32-(b))))
#define CH(x,y,z) (((x) & (y)) ^ (~(x) & (z)))
#define MAJ(x,y,z) (((x) & (y)) ^ ((x) & (z)) ^ ((y) & (z)))
#define EP0(x) (ROTRIGHT(x,2) ^ ROTRIGHT(x,13) ^ ROTRIGHT(x,22))
#define EP1(x) (ROTRIGHT(x,6) ^ ROTRIGHT(x,11) ^ ROTRIGHT(x,25))
#define SIG0(x) (ROTRIGHT(x,7) ^ ROTRIGHT(x,18) ^ ((x) >> 3))
#define SIG1(x) (ROTRIGHT(x,17) ^ ROTRIGHT(x,19) ^ ((x) >> 10))

static const unsigned int k[64] = {
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
        m[i] = SIG1(m[i - 2]) + m[i - 7] + SIG0(m[i - 15]) + m[i - 16];

    a = ctx->state[0]; b = ctx->state[1]; c = ctx->state[2]; d = ctx->state[3];
    e = ctx->state[4]; f = ctx->state[5]; g = ctx->state[6]; h = ctx->state[7];

    for (i = 0; i < 64; ++i) {
        t1 = h + EP1(e) + CH(e,f,g) + k[i] + m[i];
        t2 = EP0(a) + MAJ(a,b,c);
        h = g; g = f; f = e; e = d + t1;
        d = c; c = b; b = a; a = t1 + t2;
    }

    ctx->state[0] += a; ctx->state[1] += b; ctx->state[2] += c; ctx->state[3] += d;
    ctx->state[4] += e; ctx->state[5] += f; ctx->state[6] += g; ctx->state[7] += h;
}

static void sha256_init(SHA256_CTX *ctx) {
    ctx->datalen = 0;
    ctx->bitlen = 0;
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

static void sha256_final(SHA256_CTX *ctx, unsigned char hash[]) {
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

int scriptgo_crypto_hash_digest(const char *algo, const char *data, const char *encoding, char **out_digest) {
    if (algo == NULL || data == NULL || out_digest == NULL) {
        return crypto_fail("scriptgo crypto invalid arguments");
    }
    unsigned char hash[32];
    SHA256_CTX ctx;
    sha256_init(&ctx);
    sha256_update(&ctx, (const unsigned char*)data, strlen(data));
    sha256_final(&ctx, hash);

    char *hex_str = (char *)malloc(65);
    if (hex_str == NULL) {
        return crypto_fail("scriptgo crypto allocation failed");
    }
    for (int i = 0; i < 32; i++) {
        snprintf(hex_str + (i * 2), 3, "%02x", hash[i]);
    }
    hex_str[64] = '\0';
    *out_digest = hex_str;
    return 0;
}

