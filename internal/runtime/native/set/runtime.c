#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_set_add_number(void *handle, double value, void **out_set);
int scriptgo_set_add_string(void *handle, const char *value, void **out_set);
int scriptgo_set_add_ptr(void *handle, void *value, void **out_set);

#define SCRIPTGO_MAGIC_SET 0x53455431 // "SET1"

typedef enum {
    SCRIPTGO_SET_VAL_NUMBER = 1,
    SCRIPTGO_SET_VAL_STRING = 2,
    SCRIPTGO_SET_VAL_PTR = 3
} scriptgo_set_val_type;

typedef struct {
    char *key_str;
    scriptgo_set_val_type val_type;
    double num_val;
    char *str_val;
    void *ptr_val;
} scriptgo_set_native_entry;

typedef struct {
    uint32_t magic;
    int64_t size;
    int64_t capacity;
    scriptgo_set_native_entry *entries;
} scriptgo_set_native;

static int set_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static char *set_num_to_str(double n) {
    char buf[64];
    if (isnan(n)) {
        snprintf(buf, sizeof(buf), "NaN");
    } else if (isinf(n)) {
        snprintf(buf, sizeof(buf), n > 0 ? "Infinity" : "-Infinity");
    } else if (n == (double)(int64_t)n) {
        snprintf(buf, sizeof(buf), "%lld", (long long)n);
    } else {
        snprintf(buf, sizeof(buf), "%.14g", n);
    }
    return strdup(buf);
}

int scriptgo_set_new(void **out_set) {
    if (out_set == NULL) return set_fail("scriptgo set new: null out_set");
    scriptgo_set_native *s = calloc(1, sizeof(scriptgo_set_native));
    if (s == NULL) return set_fail("scriptgo set new: out of memory");
    s->magic = SCRIPTGO_MAGIC_SET;
    s->size = 0;
    s->capacity = 8;
    s->entries = calloc(s->capacity, sizeof(scriptgo_set_native_entry));
    if (s->entries == NULL) {
        free(s);
        return set_fail("scriptgo set new: out of memory");
    }
    *out_set = s;
    return 0;
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_array_inner_set;

int scriptgo_set_new_values_number(void *values_array, void **out_set) {
    if (scriptgo_set_new(out_set) != 0) return -1;
    if (values_array == NULL) return 0;
    scriptgo_set_native *s = *out_set;
    scriptgo_array_inner_set *arr = values_array;
    for (int64_t i = 0; i < arr->length; i++) {
        double v = *(double *)(arr->data + (size_t)i * sizeof(double));
        void *dummy;
        scriptgo_set_add_number(s, v, &dummy);
    }
    return 0;
}

int scriptgo_set_new_values_string(void *values_array, void **out_set) {
    if (scriptgo_set_new(out_set) != 0) return -1;
    if (values_array == NULL) return 0;
    scriptgo_set_native *s = *out_set;
    scriptgo_array_inner_set *arr = values_array;
    for (int64_t i = 0; i < arr->length; i++) {
        char *v = *(char **)(arr->data + (size_t)i * sizeof(char *));
        void *dummy;
        scriptgo_set_add_string(s, v, &dummy);
    }
    return 0;
}

int scriptgo_set_new_values_ptr(void *values_array, void **out_set) {
    if (scriptgo_set_new(out_set) != 0) return -1;
    if (values_array == NULL) return 0;
    scriptgo_set_native *s = *out_set;
    scriptgo_array_inner_set *arr = values_array;
    for (int64_t i = 0; i < arr->length; i++) {
        void *v = *(void **)(arr->data + (size_t)i * sizeof(void *));
        void *dummy;
        scriptgo_set_add_ptr(s, v, &dummy);
    }
    return 0;
}

static int set_ensure_capacity(scriptgo_set_native *s) {
    if (s->size >= s->capacity) {
        int64_t new_cap = s->capacity * 2;
        if (new_cap < 8) new_cap = 8;
        scriptgo_set_native_entry *new_entries = realloc(s->entries, (size_t)new_cap * sizeof(scriptgo_set_native_entry));
        if (new_entries == NULL) return set_fail("scriptgo set add: out of memory");
        s->entries = new_entries;
        s->capacity = new_cap;
    }
    return 0;
}

static int64_t set_find_entry(scriptgo_set_native *s, const char *key_str) {
    if (s == NULL || key_str == NULL) return -1;
    for (int64_t i = 0; i < s->size; i++) {
        if (s->entries[i].key_str != NULL && strcmp(s->entries[i].key_str, key_str) == 0) {
            return i;
        }
    }
    return -1;
}

int scriptgo_set_add_number(void *handle, double value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    char *kstr = set_num_to_str(value);
    int64_t idx = set_find_entry(s, kstr);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) {
            free(kstr);
            return -1;
        }
        s->entries[s->size].key_str = kstr;
        s->entries[s->size].val_type = SCRIPTGO_SET_VAL_NUMBER;
        s->entries[s->size].num_val = value;
        s->size++;
    } else {
        free(kstr);
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_add_string(void *handle, const char *value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    if (value == NULL) value = "";
    int64_t idx = set_find_entry(s, value);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) return -1;
        s->entries[s->size].key_str = strdup(value);
        s->entries[s->size].val_type = SCRIPTGO_SET_VAL_STRING;
        s->entries[s->size].str_val = strdup(value);
        s->size++;
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_add_ptr(void *handle, void *value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    char buf[64];
    snprintf(buf, sizeof(buf), "p:%p", value);
    int64_t idx = set_find_entry(s, buf);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) return -1;
        s->entries[s->size].key_str = strdup(buf);
        s->entries[s->size].val_type = SCRIPTGO_SET_VAL_PTR;
        s->entries[s->size].ptr_val = value;
        s->size++;
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_has_number(void *handle, double value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    char *kstr = set_num_to_str(value);
    int64_t idx = set_find_entry(s, kstr);
    free(kstr);
    *out_bool = (idx >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_has_string(void *handle, const char *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    if (value == NULL) value = "";
    int64_t idx = set_find_entry(s, value);
    *out_bool = (idx >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_has_ptr(void *handle, void *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    char buf[64];
    snprintf(buf, sizeof(buf), "p:%p", value);
    int64_t idx = set_find_entry(s, buf);
    *out_bool = (idx >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_delete_number(void *handle, double value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    char *kstr = set_num_to_str(value);
    int64_t idx = set_find_entry(s, kstr);
    free(kstr);
    if (idx < 0) {
        *out_bool = 0;
        return 0;
    }
    if (s->entries[idx].key_str != NULL) free(s->entries[idx].key_str);
    if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_STRING && s->entries[idx].str_val != NULL) {
        free(s->entries[idx].str_val);
    }
    for (int64_t i = idx; i < s->size - 1; i++) {
        s->entries[i] = s->entries[i + 1];
    }
    s->size--;
    *out_bool = 1;
    return 0;
}

int scriptgo_set_delete_string(void *handle, const char *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    if (value == NULL) value = "";
    int64_t idx = set_find_entry(s, value);
    if (idx < 0) {
        *out_bool = 0;
        return 0;
    }
    if (s->entries[idx].key_str != NULL) free(s->entries[idx].key_str);
    if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_STRING && s->entries[idx].str_val != NULL) {
        free(s->entries[idx].str_val);
    }
    for (int64_t i = idx; i < s->size - 1; i++) {
        s->entries[i] = s->entries[i + 1];
    }
    s->size--;
    *out_bool = 1;
    return 0;
}

int scriptgo_set_delete_ptr(void *handle, void *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    char buf[64];
    snprintf(buf, sizeof(buf), "p:%p", value);
    int64_t idx = set_find_entry(s, buf);
    if (idx < 0) {
        *out_bool = 0;
        return 0;
    }
    if (s->entries[idx].key_str != NULL) free(s->entries[idx].key_str);
    for (int64_t i = idx; i < s->size - 1; i++) {
        s->entries[i] = s->entries[i + 1];
    }
    s->size--;
    *out_bool = 1;
    return 0;
}

int scriptgo_set_clear(void *handle) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set clear: invalid handle");
    for (int64_t i = 0; i < s->size; i++) {
        if (s->entries[i].key_str != NULL) free(s->entries[i].key_str);
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING && s->entries[i].str_val != NULL) {
            free(s->entries[i].str_val);
        }
    }
    s->size = 0;
    return 0;
}

int scriptgo_set_size(void *handle, double *out_size) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set size: invalid handle");
    if (out_size == NULL) return set_fail("scriptgo set size: null out_size");
    *out_size = (double)s->size;
    return 0;
}

int scriptgo_set_to_string(void *handle, char **out_str) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) {
        if (out_str != NULL) *out_str = strdup("Set(0) {}");
        return 0;
    }
    if (out_str == NULL) return set_fail("scriptgo set toString: null out_str");
    size_t cap = 256;
    char *buf = malloc(cap);
    if (buf == NULL) return set_fail("scriptgo set toString: out of memory");
    snprintf(buf, cap, "Set(%lld) {", (long long)s->size);
    for (int64_t i = 0; i < s->size; i++) {
        char val_buf[64];
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_NUMBER) {
            double n = s->entries[i].num_val;
            if (n == (double)(int64_t)n) {
                snprintf(val_buf, sizeof(val_buf), "%lld", (long long)n);
            } else {
                snprintf(val_buf, sizeof(val_buf), "%.14g", n);
            }
        } else if (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING) {
            snprintf(val_buf, sizeof(val_buf), "'%s'", s->entries[i].str_val ? s->entries[i].str_val : "");
        } else {
            snprintf(val_buf, sizeof(val_buf), "[object]");
        }
        char entry_buf[128];
        snprintf(entry_buf, sizeof(entry_buf), "%s%s", (i == 0 ? " " : ", "), val_buf);
        size_t needed = strlen(buf) + strlen(entry_buf) + 4;
        if (needed >= cap) {
            cap = needed * 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) {
                free(buf);
                return set_fail("scriptgo set toString: out of memory");
            }
            buf = new_buf;
        }
        strcat(buf, entry_buf);
    }
    if (s->size > 0) {
        strcat(buf, " ");
    }
    strcat(buf, "}");
    *out_str = buf;
    return 0;
}

int scriptgo_closure_invoke(void *closure_handle, int32_t arg_count, const scriptgo_boxed_value *a1, const scriptgo_boxed_value *a2, const scriptgo_boxed_value *a3, const scriptgo_boxed_value *a4);

int scriptgo_set_for_each(void *handle, void *closure_handle) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET || closure_handle == NULL) {
        return set_fail("scriptgo set forEach: invalid arguments");
    }
    for (int64_t i = 0; i < s->size; i++) {
        scriptgo_set_native_entry *entry = &s->entries[i];
        scriptgo_boxed_value a1 = {0};
        if (entry->val_type == SCRIPTGO_SET_VAL_NUMBER) {
            union { double d; int64_t i; } u;
            u.d = entry->num_val;
            a1.tag = 3;
            a1.payload = u.i;
        } else if (entry->val_type == SCRIPTGO_SET_VAL_STRING) {
            a1.tag = 4;
            a1.payload = (int64_t)(uintptr_t)(entry->str_val ? entry->str_val : "");
        } else {
            a1.tag = 5;
            a1.payload = (int64_t)(uintptr_t)entry->ptr_val;
        }

        scriptgo_boxed_value a2 = a1;

        scriptgo_boxed_value a3 = {0};
        a3.tag = 5;
        a3.payload = (int64_t)(uintptr_t)s;

        scriptgo_boxed_value a4 = {0};

        scriptgo_closure_invoke(closure_handle, 3, &a1, &a2, &a3, &a4);
    }
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_object_new(int64_t field_count, void **out_object);
int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_set_array_header;

int scriptgo_set_values(void *handle, void **out_array) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET || out_array == NULL) return set_fail("invalid set handle");
    int has_num = 0;
    for (int64_t i = 0; i < s->size; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_NUMBER) { has_num = 1; break; }
    }
    if (has_num) {
        if (scriptgo_array_new(s->size, sizeof(double), out_array) != 0) return -1;
        scriptgo_set_array_header *arr = *out_array;
        for (int64_t i = 0; i < s->size; i++) {
            double d = s->entries[i].num_val;
            memcpy(arr->data + (size_t)i * sizeof(double), &d, sizeof(double));
        }
        return 0;
    }
    if (scriptgo_array_new(s->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_set_array_header *arr = *out_array;
    for (int64_t i = 0; i < s->size; i++) {
        void *val = (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING) ? (void *)s->entries[i].str_val : s->entries[i].ptr_val;
        memcpy(arr->data + (size_t)i * sizeof(void *), &val, sizeof(void *));
    }
    return 0;
}

int scriptgo_set_keys(void *handle, void **out_array) {
    return scriptgo_set_values(handle, out_array);
}

int scriptgo_set_entries(void *handle, void **out_array) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET || out_array == NULL) return set_fail("invalid set handle");
    if (scriptgo_array_new(s->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_set_array_header *arr = *out_array;
    for (int64_t i = 0; i < s->size; i++) {
        void *tup = NULL;
        if (scriptgo_object_new(2, &tup) != 0) return -1;
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_NUMBER) {
            scriptgo_object_number_set(tup, 0, s->entries[i].num_val);
            scriptgo_object_number_set(tup, 1, s->entries[i].num_val);
        } else if (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING) {
            scriptgo_object_string_set(tup, 0, s->entries[i].str_val ? s->entries[i].str_val : "");
            scriptgo_object_string_set(tup, 1, s->entries[i].str_val ? s->entries[i].str_val : "");
        } else {
            scriptgo_object_ptr_set(tup, 0, s->entries[i].ptr_val);
            scriptgo_object_ptr_set(tup, 1, s->entries[i].ptr_val);
        }
        memcpy(arr->data + (size_t)i * sizeof(void *), &tup, sizeof(void *));
    }
    return 0;
}

