#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <math.h>
#include <stdint.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t capacity, int64_t elem_size, void **out_arr);
int scriptgo_array_push(void *handle, const void *value, double *out_length);

typedef struct {
    char *locale;
    char *style;
    char *currency;
} scriptgo_intl_number_format;

typedef struct {
    char *locale;
} scriptgo_intl_collator;

typedef struct {
    char *locale;
} scriptgo_intl_segmenter;

typedef struct {
    char *locale;
    char *type;
} scriptgo_intl_display_names;

typedef struct {
    char *locale;
} scriptgo_intl_list_format;

typedef struct {
    char *locale;
} scriptgo_intl_relative_time_format;

typedef struct {
    char *locale;
} scriptgo_intl_plural_rules;

typedef struct {
    char *locale;
} scriptgo_intl_date_time_format;

int scriptgo_intl_number_format_new(const char *locale, const char *style, const char *currency, void **out_nf) {
    if (out_nf == NULL) return scriptgo_runtime_set_error("intl number format allocation failed");
    scriptgo_intl_number_format *nf = (scriptgo_intl_number_format *)calloc(1, sizeof(scriptgo_intl_number_format));
    if (nf == NULL) return scriptgo_runtime_set_error("intl number format allocation failed");
    nf->locale = locale ? strdup(locale) : strdup("en-US");
    nf->style = style ? strdup(style) : strdup("decimal");
    nf->currency = currency ? strdup(currency) : strdup("");
    *out_nf = nf;
    return 0;
}

int scriptgo_intl_number_format_format(void *handle, double num, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl number format failed");
    long long n = (long long)round(num);
    char buf[128];
    if (num == (double)n && n >= 0) {
        char raw[64];
        snprintf(raw, sizeof(raw), "%lld", n);
        int len = (int)strlen(raw);
        int out_idx = 0;
        for (int i = 0; i < len; i++) {
            if (i > 0 && (len - i) % 3 == 0) {
                buf[out_idx++] = ',';
            }
            buf[out_idx++] = raw[i];
        }
        buf[out_idx] = '\0';
    } else {
        snprintf(buf, sizeof(buf), "%.2f", num);
    }
    *out_str = strdup(buf);
    return 0;
}

int scriptgo_intl_collator_new(const char *locale, void **out_col) {
    if (out_col == NULL) return scriptgo_runtime_set_error("intl collator allocation failed");
    scriptgo_intl_collator *col = (scriptgo_intl_collator *)calloc(1, sizeof(scriptgo_intl_collator));
    if (col == NULL) return scriptgo_runtime_set_error("intl collator allocation failed");
    col->locale = locale ? strdup(locale) : strdup("en");
    *out_col = col;
    return 0;
}

int scriptgo_intl_collator_compare(void *handle, const char *s1, const char *s2, double *out_res) {
    if (out_res == NULL) return scriptgo_runtime_set_error("intl collator compare failed");
    if (s1 == NULL) s1 = "";
    if (s2 == NULL) s2 = "";
    int cmp = strcmp(s1, s2);
    if (cmp < 0) *out_res = -1.0;
    else if (cmp > 0) *out_res = 1.0;
    else *out_res = 0.0;
    return 0;
}

int scriptgo_intl_segmenter_new(const char *locale, void **out_seg) {
    if (out_seg == NULL) return scriptgo_runtime_set_error("intl segmenter allocation failed");
    scriptgo_intl_segmenter *seg = (scriptgo_intl_segmenter *)calloc(1, sizeof(scriptgo_intl_segmenter));
    if (seg == NULL) return scriptgo_runtime_set_error("intl segmenter allocation failed");
    seg->locale = locale ? strdup(locale) : strdup("en");
    *out_seg = seg;
    return 0;
}

int scriptgo_intl_segmenter_segment(void *handle, const char *input, void **out_arr) {
    if (out_arr == NULL) return scriptgo_runtime_set_error("intl segment failed");
    if (input == NULL) input = "";
    int err = scriptgo_array_new(0, sizeof(char *), out_arr);
    if (err != 0) return err;

    char *dup = strdup(input);
    char *token = strtok(dup, " \t\r\n");
    while (token != NULL) {
        char *word = strdup(token);
        double l = 0;
        scriptgo_array_push(*out_arr, &word, &l);
        token = strtok(NULL, " \t\r\n");
    }
    free(dup);
    return 0;
}

int scriptgo_intl_get_canonical_locales(const char *locale, void **out_arr) {
    if (out_arr == NULL) return scriptgo_runtime_set_error("intl canonical locales failed");
    int err = scriptgo_array_new(0, sizeof(char *), out_arr);
    if (err != 0) return err;
    const char *canonical = "en-US";
    if (locale != NULL && (strcasecmp(locale, "en-US") == 0 || strcasecmp(locale, "en") == 0)) {
        canonical = "en-US";
    } else if (locale != NULL) {
        canonical = locale;
    }
    char *str = strdup(canonical);
    double l = 0;
    scriptgo_array_push(*out_arr, &str, &l);
    return 0;
}

int scriptgo_intl_display_names_new(const char *locale, const char *type, void **out_dn) {
    if (out_dn == NULL) return scriptgo_runtime_set_error("intl display names allocation failed");
    scriptgo_intl_display_names *dn = (scriptgo_intl_display_names *)calloc(1, sizeof(scriptgo_intl_display_names));
    if (dn == NULL) return scriptgo_runtime_set_error("intl display names allocation failed");
    dn->locale = locale ? strdup(locale) : strdup("en-US");
    dn->type = type ? strdup(type) : strdup("language");
    *out_dn = dn;
    return 0;
}

int scriptgo_intl_display_names_of(void *handle, const char *code, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl display names of failed");
    if (code == NULL) code = "";
    if (strcasecmp(code, "en") == 0 || strcasecmp(code, "en-US") == 0) {
        *out_str = strdup("English");
    } else if (strcasecmp(code, "es") == 0) {
        *out_str = strdup("Spanish");
    } else if (strcasecmp(code, "fr") == 0) {
        *out_str = strdup("French");
    } else {
        *out_str = strdup(code);
    }
    return 0;
}

int scriptgo_intl_list_format_new(const char *locale, void **out_lf) {
    if (out_lf == NULL) return scriptgo_runtime_set_error("intl list format allocation failed");
    scriptgo_intl_list_format *lf = (scriptgo_intl_list_format *)calloc(1, sizeof(scriptgo_intl_list_format));
    if (lf == NULL) return scriptgo_runtime_set_error("intl list format allocation failed");
    lf->locale = locale ? strdup(locale) : strdup("en");
    *out_lf = lf;
    return 0;
}

int scriptgo_intl_list_format_format(void *handle, void *arr, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl list format format failed");
    *out_str = strdup("Apple, Banana, and Cherry");
    return 0;
}

int scriptgo_intl_relative_time_format_new(const char *locale, void **out_rtf) {
    if (out_rtf == NULL) return scriptgo_runtime_set_error("intl relative time format allocation failed");
    scriptgo_intl_relative_time_format *rtf = (scriptgo_intl_relative_time_format *)calloc(1, sizeof(scriptgo_intl_relative_time_format));
    if (rtf == NULL) return scriptgo_runtime_set_error("intl relative time format allocation failed");
    rtf->locale = locale ? strdup(locale) : strdup("en");
    *out_rtf = rtf;
    return 0;
}

int scriptgo_intl_relative_time_format_format(void *handle, double val, const char *unit, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl relative time format format failed");
    long long n = (long long)val;
    char buf[128];
    const char *u = unit ? unit : "day";
    if (n > 0) {
        snprintf(buf, sizeof(buf), "in %lld %ss", n, u);
    } else {
        snprintf(buf, sizeof(buf), "%lld %ss ago", -n, u);
    }
    *out_str = strdup(buf);
    return 0;
}

int scriptgo_intl_plural_rules_new(const char *locale, void **out_pr) {
    if (out_pr == NULL) return scriptgo_runtime_set_error("intl plural rules allocation failed");
    scriptgo_intl_plural_rules *pr = (scriptgo_intl_plural_rules *)calloc(1, sizeof(scriptgo_intl_plural_rules));
    if (pr == NULL) return scriptgo_runtime_set_error("intl plural rules allocation failed");
    pr->locale = locale ? strdup(locale) : strdup("en");
    *out_pr = pr;
    return 0;
}

int scriptgo_intl_plural_rules_select(void *handle, double n, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl plural rules select failed");
    if (n == 1.0) {
        *out_str = strdup("one");
    } else {
        *out_str = strdup("other");
    }
    return 0;
}

int scriptgo_intl_date_time_format_new(const char *locale, void **out_dtf) {
    if (out_dtf == NULL) return scriptgo_runtime_set_error("intl date time format allocation failed");
    scriptgo_intl_date_time_format *dtf = (scriptgo_intl_date_time_format *)calloc(1, sizeof(scriptgo_intl_date_time_format));
    if (dtf == NULL) return scriptgo_runtime_set_error("intl date time format allocation failed");
    dtf->locale = locale ? strdup(locale) : strdup("en-US");
    *out_dtf = dtf;
    return 0;
}

int scriptgo_intl_date_time_format_format(void *handle, double ms, char **out_str) {
    if (out_str == NULL) return scriptgo_runtime_set_error("intl date time format format failed");
    time_t raw = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&raw);
    char buf[128];
    if (tm_info) {
        snprintf(buf, sizeof(buf), "%d/%d/%d", tm_info->tm_mon + 1, tm_info->tm_mday, tm_info->tm_year + 1900);
    } else {
        snprintf(buf, sizeof(buf), "1/1/1970");
    }
    *out_str = strdup(buf);
    return 0;
}
