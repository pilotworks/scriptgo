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
        const char *marker = scriptgo_number_marker_name(val);
        snprintf(buf, sizeof(buf), "%s", marker != NULL ? marker : "NaN");
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

/* A number-storage marker formats as the value it stands for. */
void scriptgo_number_format(double value, char *buf, size_t size) {
    if (isnan(value)) {
        const char *marker = scriptgo_number_marker_name(value);
        snprintf(buf, size, "%s", marker != NULL ? marker : "NaN");
        return;
    }
    if (value == 0.0) { snprintf(buf, size, "0"); return; }
    if (isinf(value)) { snprintf(buf, size, value > 0 ? "Infinity" : "-Infinity"); return; }
    char sign[2] = {0, 0};
    if (value < 0) { sign[0] = '-'; value = -value; }
    /* Shortest digit string s (k digits) and exponent n with value = 0.s * 10^n. */
    char sci[40];
    int k = 1;
    for (; k <= 17; k++) {
        snprintf(sci, sizeof(sci), "%.*e", k - 1, value);
        if (strtod(sci, NULL) == value) break;
    }
    char digits[20];
    int len = 0;
    const char *p = sci;
    for (; *p != 'e' && *p != '\0'; p++) {
        if (*p >= '0' && *p <= '9' && len < 19) digits[len++] = *p;
    }
    while (len > 1 && digits[len - 1] == '0') len--;
    digits[len] = '\0';
    int n = (*p == 'e' ? atoi(p + 1) : 0) + 1;
    char out[64];
    if (len <= n && n <= 21) {
        snprintf(out, sizeof(out), "%s%s%0*d", sign, digits, n - len, 0);
        if (n - len == 0) snprintf(out, sizeof(out), "%s%s", sign, digits);
    } else if (0 < n && n <= 21) {
        snprintf(out, sizeof(out), "%s%.*s.%s", sign, n, digits, digits + n);
    } else if (-6 < n && n <= 0) {
        snprintf(out, sizeof(out), "%s0.%0*d%s", sign, -n, 0, digits);
        if (n == 0) snprintf(out, sizeof(out), "%s0.%s", sign, digits);
    } else {
        int e = n - 1;
        if (len == 1) snprintf(out, sizeof(out), "%s%se%c%d", sign, digits, e < 0 ? '-' : '+', e < 0 ? -e : e);
        else snprintf(out, sizeof(out), "%s%c.%se%c%d", sign, digits[0], digits + 1, e < 0 ? '-' : '+', e < 0 ? -e : e);
    }
    snprintf(buf, size, "%s", out);
}

/* StringToNumber (ECMA-262 7.1.4.1.1): surrounding whitespace is ignored,
 * an empty string is 0, 0x/0o/0b prefixes are integer literals, "Infinity"
 * may be signed, and anything not fully a numeric literal is NaN. */
double scriptgo_string_to_number(const char *str) {
    if (str == NULL) return NAN;
    const char *begin = str;
    while (*begin && isspace((unsigned char)*begin)) begin++;
    const char *end = begin + strlen(begin);
    while (end > begin && isspace((unsigned char)end[-1])) end--;
    size_t len = (size_t)(end - begin);
    if (len == 0) return 0.0;
    char buf[512];
    if (len >= sizeof(buf)) return NAN;
    memcpy(buf, begin, len);
    buf[len] = '\0';
    if (len > 2 && buf[0] == '0' && (buf[1] == 'x' || buf[1] == 'X' || buf[1] == 'o' || buf[1] == 'O' || buf[1] == 'b' || buf[1] == 'B')) {
        int radix = (buf[1] == 'x' || buf[1] == 'X') ? 16 : (buf[1] == 'o' || buf[1] == 'O') ? 8 : 2;
        double value = 0.0;
        for (size_t i = 2; i < len; i++) {
            int digit;
            char c = buf[i];
            if (c >= '0' && c <= '9') digit = c - '0';
            else if (c >= 'a' && c <= 'z') digit = c - 'a' + 10;
            else if (c >= 'A' && c <= 'Z') digit = c - 'A' + 10;
            else return NAN;
            if (digit >= radix) return NAN;
            value = value * radix + digit;
        }
        return value;
    }
    const char *body = buf;
    int negative = 0;
    if (*body == '+' || *body == '-') {
        negative = *body == '-';
        body++;
    }
    if (strcmp(body, "Infinity") == 0) return negative ? -INFINITY : INFINITY;
    /* strtod also accepts inf/nan/hex floats; only decimal literals remain. */
    for (const char *c = body; *c; c++) {
        if (!(isdigit((unsigned char)*c) || *c == '.' || *c == 'e' || *c == 'E' || *c == '+' || *c == '-')) return NAN;
    }
    if (*body == '\0') return NAN;
    char *parsed_end = NULL;
    double value = strtod(buf, &parsed_end);
    if (parsed_end == NULL || *parsed_end != '\0') return NAN;
    return value;
}

int scriptgo_number_to_string(double val, double radix, char **out_value) {
    if (out_value == NULL) return scriptgo_runtime_set_error("invalid argument to toString");
    int r = 10;
    if (!isnan(radix) && radix >= 2.0 && radix <= 36.0) {
        r = (int)radix;
    }
    char buf[128];
    if (isnan(val)) {
        const char *marker = scriptgo_number_marker_name(val);
        snprintf(buf, sizeof(buf), "%s", marker != NULL ? marker : "NaN");
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
        scriptgo_number_format(val, buf, sizeof(buf));
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
        const char *marker = scriptgo_number_marker_name(val);
        snprintf(buf, sizeof(buf), "%s", marker != NULL ? marker : "NaN");
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
        const char *marker = scriptgo_number_marker_name(val);
        snprintf(buf, sizeof(buf), "%s", marker != NULL ? marker : "NaN");
    } else if (isinf(val)) {
        if (val > 0) snprintf(buf, sizeof(buf), "Infinity");
        else snprintf(buf, sizeof(buf), "-Infinity");
    } else if (!isnan(precision) && precision > 0.0) {
        /* Number.prototype.toPrecision: p significant digits, keeping
         * trailing zeros; exponential when e < -6 or e >= p. */
        int p = (int)precision;
        if (p > 100) p = 100;
        char sci[160];
        snprintf(sci, sizeof(sci), "%.*e", p - 1, val);
        int e = atoi(strchr(sci, 'e') + 1);
        if (e < -6 || e >= p) {
            char *exp = strchr(sci, 'e');
            *exp = '\0';
            snprintf(buf, sizeof(buf), "%se%c%d", sci, e < 0 ? '-' : '+', e < 0 ? -e : e);
        } else {
            snprintf(buf, sizeof(buf), "%.*f", p - 1 - e, val);
        }
    } else {
        scriptgo_number_format(val, buf, sizeof(buf));
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
