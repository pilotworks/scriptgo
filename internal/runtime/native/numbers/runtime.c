#include <ctype.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

int scriptgo_number_parse_int(const char *str, double *out_value) {
    if (str == NULL || out_value == NULL) return scriptgo_runtime_set_error("invalid argument to parseInt");
    while (isspace((unsigned char)*str)) str++;
    char *endptr = NULL;
    long long val = strtoll(str, &endptr, 10);
    if (endptr == str) {
        *out_value = NAN;
    } else {
        *out_value = (double)val;
    }
    return 0;
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
        snprintf(buf, sizeof(buf), "%.*f", d, val);
    }
    size_t len = strlen(buf);
    char *res = malloc(len + 1);
    if (res == NULL) return scriptgo_runtime_set_error("scriptgo string allocation failed");
    memcpy(res, buf, len + 1);
    *out_value = res;
    return 0;
}
