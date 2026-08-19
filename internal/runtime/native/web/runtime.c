#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#if defined(__aarch64__) && defined(__ARM_NEON)
#include <arm_neon.h>
#define SCRIPTGO_HAS_NEON_BASE64 1
#endif

int scriptgo_runtime_set_error(const char *message);

static int web_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static const char b64_table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

#if defined(SCRIPTGO_HAS_NEON_BASE64)
static uint8x16_t b64_neon_map(uint8x16_t values) {
    uint8x16_t upper = vaddq_u8(values, vdupq_n_u8('A'));
    uint8x16_t lower = vaddq_u8(values, vdupq_n_u8('a' - 26));
    uint8x16_t digits = vsubq_u8(values, vdupq_n_u8(52 - '0'));
    uint8x16_t plus = vdupq_n_u8('+');
    uint8x16_t slash = vdupq_n_u8('/');
    uint8x16_t lt26 = vcltq_u8(values, vdupq_n_u8(26));
    uint8x16_t lt52 = vcltq_u8(values, vdupq_n_u8(52));
    uint8x16_t lt62 = vcltq_u8(values, vdupq_n_u8(62));
    uint8x16_t result = slash;

    result = vbslq_u8(lt62, digits, result);
    result = vbslq_u8(lt52, lower, result);
    result = vbslq_u8(lt26, upper, result);
    result = vbslq_u8(vceqq_u8(values, vdupq_n_u8(62)), plus, result);
    return result;
}

static void b64_neon_encode_block(const unsigned char *input, char *output) {
    static const uint8_t a_indices[] = {0, 3, 6, 9, 0, 3, 6, 9, 0, 3, 6, 9, 0, 3, 6, 9};
    static const uint8_t b_indices[] = {1, 4, 7, 10, 1, 4, 7, 10, 1, 4, 7, 10, 1, 4, 7, 10};
    static const uint8_t c_indices[] = {2, 5, 8, 11, 2, 5, 8, 11, 2, 5, 8, 11, 2, 5, 8, 11};
    uint8x16_t source = vld1q_u8(input);
    uint8x16_t a = vqtbl1q_u8(source, vld1q_u8(a_indices));
    uint8x16_t b = vqtbl1q_u8(source, vld1q_u8(b_indices));
    uint8x16_t c = vqtbl1q_u8(source, vld1q_u8(c_indices));
    uint8x16_t mask2 = vdupq_n_u8(3);
    uint8x16_t mask4 = vdupq_n_u8(15);
    uint8x16_t mask6 = vdupq_n_u8(63);
    uint8x16_t s0 = vshrq_n_u8(a, 2);
    uint8x16_t s1 = vorrq_u8(vshlq_n_u8(vandq_u8(a, mask2), 4), vshrq_n_u8(b, 4));
    uint8x16_t s2 = vorrq_u8(vshlq_n_u8(vandq_u8(b, mask4), 2), vshrq_n_u8(c, 6));
    uint8x16_t s3 = vandq_u8(c, mask6);

    uint8_t values0[16], values1[16], values2[16], values3[16], sextets[16];
    vst1q_u8(values0, s0);
    vst1q_u8(values1, s1);
    vst1q_u8(values2, s2);
    vst1q_u8(values3, s3);
    for (int group = 0; group < 4; group++) {
        sextets[group * 4] = values0[group];
        sextets[group * 4 + 1] = values1[group];
        sextets[group * 4 + 2] = values2[group];
        sextets[group * 4 + 3] = values3[group];
    }
    vst1q_u8((uint8_t *)output, b64_neon_map(vld1q_u8(sextets)));
}
#endif

int scriptgo_web_btoa(const char *input, char **out_base64) {
    if (input == NULL || out_base64 == NULL) {
        return web_fail("scriptgo btoa invalid arguments");
    }
    size_t in_len = strlen(input);
    size_t out_len = 4 * ((in_len + 2) / 3);
    char *encoded = (char *)malloc(out_len + 1);
    if (encoded == NULL) {
        return web_fail("scriptgo btoa allocation failed");
    }

    size_t i = 0, j = 0;
#if defined(SCRIPTGO_HAS_NEON_BASE64)
    while (in_len - i >= 16) {
        b64_neon_encode_block((const unsigned char *)input + i, encoded + j);
        i += 12;
        j += 16;
    }
#endif
    for (; i < in_len;) {
        size_t remaining = in_len - i;
        unsigned char octet_a = (unsigned char)input[i++];
        unsigned char octet_b = remaining > 1 ? (unsigned char)input[i++] : 0;
        unsigned char octet_c = remaining > 2 ? (unsigned char)input[i++] : 0;

        uint32_t triple = ((uint32_t)octet_a << 16) | ((uint32_t)octet_b << 8) | octet_c;

        encoded[j++] = b64_table[(triple >> 18) & 0x3F];
        encoded[j++] = b64_table[(triple >> 12) & 0x3F];
        encoded[j++] = remaining > 1 ? b64_table[(triple >> 6) & 0x3F] : '=';
        encoded[j++] = remaining > 2 ? b64_table[triple & 0x3F] : '=';
    }
    encoded[out_len] = '\0';
    *out_base64 = encoded;
    return 0;
}

static int b64_char_value(char c) {
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+') return 62;
    if (c == '/') return 63;
    return -1;
}

#if defined(SCRIPTGO_HAS_NEON_BASE64)
static void b64_neon_decode_block(const char *input, unsigned char *output) {
    uint8x16_t chars = vld1q_u8((const uint8_t *)input);
    uint8x16_t upper = vsubq_u8(chars, vdupq_n_u8('A'));
    uint8x16_t lower = vaddq_u8(vsubq_u8(chars, vdupq_n_u8('a')), vdupq_n_u8(26));
    uint8x16_t digits = vaddq_u8(vsubq_u8(chars, vdupq_n_u8('0')), vdupq_n_u8(52));
    uint8x16_t values = vdupq_n_u8(63);
    uint8x16_t is_upper = vandq_u8(vcgeq_u8(chars, vdupq_n_u8('A')), vcleq_u8(chars, vdupq_n_u8('Z')));
    uint8x16_t is_lower = vandq_u8(vcgeq_u8(chars, vdupq_n_u8('a')), vcleq_u8(chars, vdupq_n_u8('z')));
    uint8x16_t is_digit = vandq_u8(vcgeq_u8(chars, vdupq_n_u8('0')), vcleq_u8(chars, vdupq_n_u8('9')));
    uint8x16_t is_plus = vceqq_u8(chars, vdupq_n_u8('+'));
    uint8x16_t is_slash = vceqq_u8(chars, vdupq_n_u8('/'));
    values = vbslq_u8(is_upper, upper, values);
    values = vbslq_u8(is_lower, lower, values);
    values = vbslq_u8(is_digit, digits, values);
    values = vbslq_u8(is_plus, vdupq_n_u8(62), values);
    values = vbslq_u8(is_slash, vdupq_n_u8(63), values);

    uint8_t sextets[16];
    vst1q_u8(sextets, values);
    for (int group = 0; group < 4; group++) {
        uint8_t a = sextets[group * 4];
        uint8_t b = sextets[group * 4 + 1];
        uint8_t c = sextets[group * 4 + 2];
        uint8_t d = sextets[group * 4 + 3];
        output[group * 3] = (uint8_t)((a << 2) | (b >> 4));
        output[group * 3 + 1] = (uint8_t)((b << 4) | (c >> 2));
        output[group * 3 + 2] = (uint8_t)((c << 6) | d);
    }
}
#endif

int scriptgo_web_atob(const char *input, char **out_decoded) {
    if (input == NULL || out_decoded == NULL) {
        return web_fail("scriptgo atob invalid arguments");
    }
    size_t in_len = strlen(input);
    if (in_len % 4 != 0) {
        return web_fail("InvalidCharacterError: The string to be decoded is not correctly encoded.");
    }
    size_t pad = 0;
    if (in_len > 0) {
        if (input[in_len - 1] == '=') pad++;
        if (in_len > 1 && input[in_len - 2] == '=') pad++;
    }
    size_t out_len = (in_len / 4) * 3 - pad;
    char *decoded = (char *)malloc(out_len + 1);
    if (decoded == NULL) {
        return web_fail("scriptgo atob allocation failed");
    }

    size_t j = 0;
    size_t i = 0;
#if defined(SCRIPTGO_HAS_NEON_BASE64)
    while (in_len - i >= 16 && memchr(input + i, '=', 16) == NULL) {
        int valid = 1;
        for (size_t k = 0; k < 16; k++) {
            if (b64_char_value(input[i + k]) < 0) {
                valid = 0;
                break;
            }
        }
        if (!valid) break;
        b64_neon_decode_block(input + i, (unsigned char *)decoded + j);
        i += 16;
        j += 12;
    }
#endif
    for (; i < in_len; i += 4) {
        int v0 = b64_char_value(input[i]);
        int v1 = b64_char_value(input[i + 1]);
        int v2 = input[i + 2] == '=' ? 0 : b64_char_value(input[i + 2]);
        int v3 = input[i + 3] == '=' ? 0 : b64_char_value(input[i + 3]);

        if (v0 < 0 || v1 < 0 || (input[i + 2] != '=' && v2 < 0) || (input[i + 3] != '=' && v3 < 0)) {
            free(decoded);
            return web_fail("InvalidCharacterError: The string to be decoded contains invalid characters.");
        }

        uint32_t triple = ((uint32_t)v0 << 18) | ((uint32_t)v1 << 12) | ((uint32_t)v2 << 6) | (uint32_t)v3;

        if (j < out_len) decoded[j++] = (char)((triple >> 16) & 0xFF);
        if (j < out_len) decoded[j++] = (char)((triple >> 8) & 0xFF);
        if (j < out_len) decoded[j++] = (char)(triple & 0xFF);
    }
    decoded[out_len] = '\0';
    *out_decoded = decoded;
    return 0;
}

int scriptgo_performance_now(double *out_ms) {
    if (out_ms == NULL) {
        return web_fail("scriptgo performance invalid arguments");
    }
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    *out_ms = ((double)ts.tv_sec * 1000.0) + ((double)ts.tv_nsec / 1000000.0);
    return 0;
}
