#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <ctype.h>

int scriptgo_runtime_set_error(const char *message);

static int web_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static const char b64_table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

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
    for (i = 0; i < in_len;) {
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
    for (size_t i = 0; i < in_len; i += 4) {
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
