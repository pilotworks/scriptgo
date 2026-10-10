#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* BigInt conversions to and from strings and numbers. */

int scriptgo_runtime_set_error(const char *message);

int scriptgo_string_from_bigint(long long value, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("invalid argument to fromBigInt");
    char buf[64];
    snprintf(buf, sizeof(buf), "%lld", value);
    *out_str = strdup(buf);
    if (*out_str == NULL) return scriptgo_runtime_set_error("fromBigInt allocation failed");
    return 0;
}

int scriptgo_string_from_bigint_locale(long long value, char **out_str) {
    char raw[64];
    size_t digits;
    size_t groups;
    size_t output_len;
    size_t src = 0;
    size_t dst = 0;
    char *result;

    if (out_str == NULL) return scriptgo_runtime_set_error("invalid argument to bigint locale formatting");
    snprintf(raw, sizeof(raw), "%lld", value);
    digits = raw[0] == '-' ? strlen(raw) - 1 : strlen(raw);
    groups = digits > 3 ? (digits - 1) / 3 : 0;
    output_len = strlen(raw) + groups;
    result = malloc(output_len + 1);
    if (result == NULL) return scriptgo_runtime_set_error("bigint locale allocation failed");
    if (raw[0] == '-') result[dst++] = raw[src++];
    for (size_t i = 0; i < digits; i++) {
        if (i > 0 && (digits - i) % 3 == 0) result[dst++] = ',';
        result[dst++] = raw[src++];
    }
    result[dst] = '\0';
    *out_str = result;
    return 0;
}

int scriptgo_bigint_from_number(double value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint fromNumber");
    *out_value = (long long)value;
    return 0;
}

int scriptgo_bigint_from_string(const char *str, long long *out_value) {
    if (str == NULL || out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint fromString");
    char *endptr = NULL;
    *out_value = strtoll(str, &endptr, 10);
    return 0;
}

int scriptgo_bigint_as_int_n(long long bits, long long value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint asIntN");
    if (bits <= 0) {
        *out_value = 0;
        return 0;
    }
    if (bits >= 64) {
        *out_value = value;
        return 0;
    }
    unsigned long long uval = (unsigned long long)value;
    unsigned long long mask = (1ULL << bits) - 1ULL;
    uval = uval & mask;
    if (uval & (1ULL << (bits - 1))) {
        uval |= (~mask);
    }
    *out_value = (long long)uval;
    return 0;
}

int scriptgo_bigint_as_uint_n(long long bits, long long value, long long *out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to bigint asUintN");
    if (bits <= 0) {
        *out_value = 0;
        return 0;
    }
    if (bits >= 64) {
        *out_value = value;
        return 0;
    }
    unsigned long long uval = (unsigned long long)value;
    unsigned long long mask = (1ULL << bits) - 1ULL;
    *out_value = (long long)(uval & mask);
    return 0;
}

long long scriptgo_bigint_pow(long long base, long long exp) {
    if (exp < 0) {
        if (base == 1) return 1;
        if (base == -1) return (exp % 2 == 0) ? 1 : -1;
        return 0;
    }
    if (exp == 0) return 1;
    long long result = 1;
    long long b = base;
    unsigned long long e = (unsigned long long)exp;
    while (e > 0) {
        if (e & 1) {
            result *= b;
        }
        b *= b;
        e >>= 1;
    }
    return result;
}
