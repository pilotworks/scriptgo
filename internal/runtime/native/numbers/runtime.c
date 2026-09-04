#include <ctype.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

int scriptgo_number_parse_int_radix(const char *str, double radix, double *out_value) {
    if (str == NULL || out_value == NULL) return scriptgo_runtime_set_error("invalid argument to parseInt");
    while (isspace((unsigned char)*str)) str++;
    int r = 10;
    if (!isnan(radix) && radix != 0.0) {
        if (radix < 2.0 || radix > 36.0) {
            *out_value = NAN;
            return 0;
        }
        r = (int)radix;
    }
    char *endptr = NULL;
    long long val = strtoll(str, &endptr, r);
    if (endptr == str) {
        *out_value = NAN;
    } else {
        *out_value = (double)val;
    }
    return 0;
}

int scriptgo_number_parse_int(const char *str, double *out_value) {
    return scriptgo_number_parse_int_radix(str, 0.0, out_value);
}

int scriptgo_number_parse_float(const char *str, double *out_value) {
    if (str == NULL || out_value == NULL) return scriptgo_runtime_set_error("invalid argument to parseFloat");
    while (isspace((unsigned char)*str)) str++;
    char *endptr = NULL;
    double val = strtod(str, &endptr);
    if (endptr == str) {
        *out_value = NAN;
    } else {
        *out_value = val;
    }
    return 0;
}

int scriptgo_number_is_nan(double val, double *out_bool) {
    if (out_bool == NULL) return scriptgo_runtime_set_error("invalid argument to isNaN");
    *out_bool = isnan(val) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_number_is_finite(double val, double *out_bool) {
    if (out_bool == NULL) return scriptgo_runtime_set_error("invalid argument to isFinite");
    *out_bool = isfinite(val) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_number_is_integer(double val, double *out_bool) {
    if (out_bool == NULL) return scriptgo_runtime_set_error("invalid argument to isInteger");
    *out_bool = (isfinite(val) && trunc(val) == val) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_number_is_safe_integer(double val, double *out_bool) {
    if (out_bool == NULL) return scriptgo_runtime_set_error("invalid argument to isSafeInteger");
    *out_bool = (isfinite(val) && trunc(val) == val && fabs(val) <= 9007199254740991.0) ? 1.0 : 0.0;
    return 0;
}

int scriptgo_number_to_fixed(double val, double digits, char **out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to toFixed");
    int d = 0;
    if (!isnan(digits) && digits > 0.0) {
        d = (int)digits;
        if (d > 20) d = 20;
    }
    char buf[128];
    if (isnan(val)) {
        snprintf(buf, sizeof(buf), "NaN");
    } else if (isinf(val)) {
        if (val > 0) snprintf(buf, sizeof(buf), "Infinity");
        else snprintf(buf, sizeof(buf), "-Infinity");
    } else {
        double factor = pow(10.0, d);
        double rounded = round(val * factor) / factor;
        snprintf(buf, sizeof(buf), "%.*f", d, rounded);
    }
    size_t len = strlen(buf);
    char *res = malloc(len + 1);
    if (res == NULL) return scriptgo_runtime_set_error("scriptgo string allocation failed");
    memcpy(res, buf, len + 1);
    *out_value = res;
    return 0;
}

int scriptgo_number_to_string(double val, double radix, char **out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to toString");
    int r = 10;
    if (!isnan(radix) && radix >= 2.0 && radix <= 36.0) {
        r = (int)radix;
    }
    char buf[128];
    if (isnan(val)) {
        snprintf(buf, sizeof(buf), "NaN");
    } else if (isinf(val)) {
        if (val > 0) snprintf(buf, sizeof(buf), "Infinity");
        else snprintf(buf, sizeof(buf), "-Infinity");
    } else if (r != 10 && trunc(val) == val) {
        long long num = (long long)val;
        int is_neg = num < 0;
        unsigned long long uval = is_neg ? -num : num;
        if (uval == 0) {
            snprintf(buf, sizeof(buf), "0");
        } else {
            char temp[65];
            int pos = 64;
            temp[pos] = '\0';
            const char digits[] = "0123456789abcdefghijklmnopqrstuvwxyz";
            while (uval > 0 && pos > 0) {
                temp[--pos] = digits[uval % r];
                uval /= r;
            }
            if (is_neg && pos > 0) {
                temp[--pos] = '-';
            }
            snprintf(buf, sizeof(buf), "%s", &temp[pos]);
        }
    } else {
        if (trunc(val) == val) {
            snprintf(buf, sizeof(buf), "%.0f", val);
        } else {
            snprintf(buf, sizeof(buf), "%g", val);
        }
    }
    size_t len = strlen(buf);
    char *res = malloc(len + 1);
    if (res == NULL) return scriptgo_runtime_set_error("scriptgo string allocation failed");
    memcpy(res, buf, len + 1);
    *out_value = res;
    return 0;
}

int scriptgo_number_to_exponential(double val, double fractionDigits, char **out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to toExponential");
    char buf[128];
    if (isnan(val)) {
        snprintf(buf, sizeof(buf), "NaN");
    } else if (isinf(val)) {
        if (val > 0) snprintf(buf, sizeof(buf), "Infinity");
        else snprintf(buf, sizeof(buf), "-Infinity");
    } else if (!isnan(fractionDigits) && fractionDigits >= 0.0) {
        int d = (int)fractionDigits;
        if (d > 20) d = 20;
        snprintf(buf, sizeof(buf), "%.*e", d, val);
    } else {
        snprintf(buf, sizeof(buf), "%e", val);
    }
    char *e_pos = strchr(buf, 'e');
    if (e_pos != NULL) {
        char *sign = e_pos + 1;
        if (*sign == '+' || *sign == '-') {
            char *digits = sign + 1;
            if (*digits == '0' && *(digits + 1) != '\0') {
                memmove(digits, digits + 1, strlen(digits));
            }
        }
    }
    size_t len = strlen(buf);
    char *res = malloc(len + 1);
    if (res == NULL) return scriptgo_runtime_set_error("scriptgo string allocation failed");
    memcpy(res, buf, len + 1);
    *out_value = res;
    return 0;
}

int scriptgo_number_to_precision(double val, double precision, char **out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to toPrecision");
    char buf[128];
    if (isnan(val)) {
        snprintf(buf, sizeof(buf), "NaN");
    } else if (isinf(val)) {
        if (val > 0) snprintf(buf, sizeof(buf), "Infinity");
        else snprintf(buf, sizeof(buf), "-Infinity");
    } else if (!isnan(precision) && precision > 0.0) {
        int p = (int)precision;
        if (p > 21) p = 21;
        snprintf(buf, sizeof(buf), "%.*g", p, val);
    } else {
        snprintf(buf, sizeof(buf), "%g", val);
    }
    size_t len = strlen(buf);
    char *res = malloc(len + 1);
    if (res == NULL) return scriptgo_runtime_set_error("scriptgo string allocation failed");
    memcpy(res, buf, len + 1);
    *out_value = res;
    return 0;
}

int scriptgo_number_to_locale_string(double val, char **out_value) {
    return scriptgo_number_to_string(val, 10.0, out_value);
}

int32_t scriptgo_to_int32(double val) {
    if (isnan(val) || isinf(val) || val == 0.0) {
        return 0;
    }
    if (fabs(val) < 2147483648.0) {
        return (int32_t)val;
    }
    double two32 = 4294967296.0;
    double two31 = 2147483648.0;
    double res = fmod(trunc(val), two32);
    if (res < 0.0) res += two32;
    if (res >= two31) res -= two32;
    return (int32_t)res;
}

double scriptgo_math_round(double x) {
    if (isnan(x) || isinf(x) || x == 0.0) return x;
    if (x >= -0.5 && x < 0.0) return -0.0;
    return floor(x + 0.5);
}

double scriptgo_math_pow(double x, double y) {
    if (isnan(y)) return NAN;
    if (y == 0.0) return 1.0;
    if (isnan(x)) return NAN;
    if (fabs(x) == 1.0 && isinf(y)) return NAN;
    return pow(x, y);
}
