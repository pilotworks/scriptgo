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
    SCRIPTGO_SET_VAL_EMPTY  = 0,
    SCRIPTGO_SET_VAL_NUMBER = 1,
    SCRIPTGO_SET_VAL_STRING = 2,
    SCRIPTGO_SET_VAL_PTR    = 3
} scriptgo_set_val_type;

typedef struct {
    uint64_t hash;
    scriptgo_set_val_type val_type;
    double num_val;
    char *str_val;
    void *ptr_val;
    int64_t next;
} scriptgo_set_native_entry;

typedef struct {
    uint32_t magic;
    int64_t size;            // Number of active, non-empty elements
    int64_t entry_count;     // Total allocated elements in entries (including tombstones)
    int64_t capacity;        // Allocated capacity of entries array
    scriptgo_set_native_entry *entries;
    int64_t bucket_mask;     // (bucket_count - 1), bucket_count is a power of 2
    int64_t *buckets;        // Head entry index for each bucket, or -1
} scriptgo_set_native;

static int set_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static inline uint64_t hash_number(double n) {
    if (isnan(n)) return 0x7ff8000000000000ULL;
    if (n == 0.0) n = 0.0; // normalize -0.0 to +0.0
    uint64_t u;
    memcpy(&u, &n, sizeof(u));
    u ^= u >> 33;
    u *= 0xff51afd7ed558ccdULL;
    u ^= u >> 33;
    u *= 0xc4ceb9fe1a85ec53ULL;
    u ^= u >> 33;
    return u;
}

static inline int number_equals(double a, double b) {
    if (isnan(a) && isnan(b)) return 1;
    return a == b;
}

static inline uint64_t hash_string(const char *s) {
    if (s == NULL) return 0;
    uint64_t h = 0xcbf29ce484222325ULL;
    while (*s) {
        h ^= (uint64_t)(unsigned char)*s++;
        h *= 0x100000001b3ULL;
    }
    return h;
}

static inline uint64_t hash_ptr(void *p) {
    uint64_t u = (uint64_t)(uintptr_t)p;
    u ^= u >> 33;
    u *= 0xff51afd7ed558ccdULL;
    u ^= u >> 33;
    return u;
}

static int scriptgo_set_new_with_capacity(void **out_set, int64_t min_capacity) {
    if (out_set == NULL) return set_fail("scriptgo set new: null out_set");
    scriptgo_set_native *s = calloc(1, sizeof(scriptgo_set_native));
    if (s == NULL) return set_fail("scriptgo set new: out of memory");
    s->magic = SCRIPTGO_MAGIC_SET;
    s->size = 0;
    s->entry_count = 0;
    s->capacity = 16;
    while (s->capacity < min_capacity) s->capacity *= 2;
    s->entries = calloc((size_t)s->capacity, sizeof(scriptgo_set_native_entry));
    if (s->entries == NULL) {
        free(s);
        return set_fail("scriptgo set new: out of memory");
    }
    int64_t bucket_count = s->capacity;
    s->bucket_mask = bucket_count - 1;
    s->buckets = malloc((size_t)bucket_count * sizeof(int64_t));
    if (s->buckets == NULL) {
        free(s->entries);
        free(s);
        return set_fail("scriptgo set new: out of memory");
    }
    memset(s->buckets, -1, (size_t)bucket_count * sizeof(int64_t));
    *out_set = s;
    return 0;
}

int scriptgo_set_new(void **out_set) {
    return scriptgo_set_new_with_capacity(out_set, 16);
}

static int set_ensure_capacity(scriptgo_set_native *s) {
    if (s->entry_count < s->capacity && s->entry_count < (s->bucket_mask + 1)) {
        return 0;
    }

    // If tombstones occupy more than 50% of the entries, compact without growing
    if (s->entry_count > s->size * 2) {
        int64_t new_count = 0;
        for (int64_t i = 0; i < s->entry_count; i++) {
            if (s->entries[i].val_type != SCRIPTGO_SET_VAL_EMPTY) {
                if (new_count != i) {
                    s->entries[new_count] = s->entries[i];
                }
                new_count++;
            }
        }
        s->entry_count = new_count;
        memset(s->buckets, -1, (size_t)(s->bucket_mask + 1) * sizeof(int64_t));
        for (int64_t i = 0; i < s->entry_count; i++) {
            int64_t b = (int64_t)(s->entries[i].hash & (uint64_t)s->bucket_mask);
            s->entries[i].next = s->buckets[b];
            s->buckets[b] = i;
        }
        return 0;
    }

    int64_t new_cap = s->capacity * 2;
    if (new_cap < 16) new_cap = 16;
    int64_t new_bcount = (s->bucket_mask + 1) * 2;
    if (new_bcount < 16) new_bcount = 16;

    scriptgo_set_native_entry *new_entries = calloc((size_t)new_cap, sizeof(scriptgo_set_native_entry));
    if (new_entries == NULL) return set_fail("scriptgo set grow: out of memory");

    int64_t *new_buckets = malloc((size_t)new_bcount * sizeof(int64_t));
    if (new_buckets == NULL) {
        free(new_entries);
        return set_fail("scriptgo set grow: out of memory");
    }
    memset(new_buckets, -1, (size_t)new_bcount * sizeof(int64_t));

    int64_t new_mask = new_bcount - 1;
    int64_t new_count = 0;
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type != SCRIPTGO_SET_VAL_EMPTY) {
            new_entries[new_count] = s->entries[i];
            int64_t b = (int64_t)(new_entries[new_count].hash & (uint64_t)new_mask);
            new_entries[new_count].next = new_buckets[b];
            new_buckets[b] = new_count;
            new_count++;
        }
    }

    free(s->entries);
    free(s->buckets);
    s->entries = new_entries;
    s->buckets = new_buckets;
    s->capacity = new_cap;
    s->bucket_mask = new_mask;
    s->entry_count = new_count;
    return 0;
}

static inline int64_t set_find_number(scriptgo_set_native *s, double val, uint64_t h) {
    if (s == NULL || s->buckets == NULL) return -1;
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    for (int64_t idx = s->buckets[b]; idx >= 0; idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_NUMBER &&
            s->entries[idx].hash == h &&
            number_equals(s->entries[idx].num_val, val)) {
            return idx;
        }
    }
    return -1;
}

static inline int64_t set_find_string(scriptgo_set_native *s, const char *val, uint64_t h) {
    if (s == NULL || s->buckets == NULL) return -1;
    if (val == NULL) val = "";
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    for (int64_t idx = s->buckets[b]; idx >= 0; idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_STRING &&
            s->entries[idx].hash == h &&
            strcmp(s->entries[idx].str_val, val) == 0) {
            return idx;
        }
    }
    return -1;
}

static inline int64_t set_find_ptr(scriptgo_set_native *s, void *val, uint64_t h) {
    if (s == NULL || s->buckets == NULL) return -1;
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    for (int64_t idx = s->buckets[b]; idx >= 0; idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_PTR &&
            s->entries[idx].hash == h &&
            s->entries[idx].ptr_val == val) {
            return idx;
        }
    }
    return -1;
}

static inline int64_t set_find_entry_copy(scriptgo_set_native *s, const scriptgo_set_native_entry *e) {
    if (e->val_type == SCRIPTGO_SET_VAL_NUMBER) return set_find_number(s, e->num_val, e->hash);
    if (e->val_type == SCRIPTGO_SET_VAL_STRING) return set_find_string(s, e->str_val, e->hash);
    if (e->val_type == SCRIPTGO_SET_VAL_PTR) return set_find_ptr(s, e->ptr_val, e->hash);
    return -1;
}

int scriptgo_set_add_number(void *handle, double value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    uint64_t h = hash_number(value);
    int64_t idx = set_find_number(s, value, h);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) return -1;
        int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
        s->entries[s->entry_count].hash = h;
        s->entries[s->entry_count].val_type = SCRIPTGO_SET_VAL_NUMBER;
        s->entries[s->entry_count].num_val = value;
        s->entries[s->entry_count].str_val = NULL;
        s->entries[s->entry_count].ptr_val = NULL;
        s->entries[s->entry_count].next = s->buckets[b];
        s->buckets[b] = s->entry_count;
        s->entry_count++;
        s->size++;
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_add_string(void *handle, const char *value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    if (value == NULL) value = "";
    uint64_t h = hash_string(value);
    int64_t idx = set_find_string(s, value, h);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) return -1;
        int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
        s->entries[s->entry_count].hash = h;
        s->entries[s->entry_count].val_type = SCRIPTGO_SET_VAL_STRING;
        s->entries[s->entry_count].num_val = 0;
        s->entries[s->entry_count].str_val = strdup(value);
        s->entries[s->entry_count].ptr_val = NULL;
        s->entries[s->entry_count].next = s->buckets[b];
        s->buckets[b] = s->entry_count;
        s->entry_count++;
        s->size++;
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_add_ptr(void *handle, void *value, void **out_set) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set add: invalid handle");
    uint64_t h = hash_ptr(value);
    int64_t idx = set_find_ptr(s, value, h);
    if (idx < 0) {
        if (set_ensure_capacity(s) != 0) return -1;
        int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
        s->entries[s->entry_count].hash = h;
        s->entries[s->entry_count].val_type = SCRIPTGO_SET_VAL_PTR;
        s->entries[s->entry_count].num_val = 0;
        s->entries[s->entry_count].str_val = NULL;
        s->entries[s->entry_count].ptr_val = value;
        s->entries[s->entry_count].next = s->buckets[b];
        s->buckets[b] = s->entry_count;
        s->entry_count++;
        s->size++;
    }
    if (out_set != NULL) *out_set = s;
    return 0;
}

int scriptgo_set_has_number(void *handle, double value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    uint64_t h = hash_number(value);
    *out_bool = (set_find_number(s, value, h) >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_has_string(void *handle, const char *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    if (value == NULL) value = "";
    uint64_t h = hash_string(value);
    *out_bool = (set_find_string(s, value, h) >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_has_ptr(void *handle, void *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set has: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set has: null out_bool");
    uint64_t h = hash_ptr(value);
    *out_bool = (set_find_ptr(s, value, h) >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_set_delete_number(void *handle, double value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    uint64_t h = hash_number(value);
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    int64_t prev = -1;
    for (int64_t idx = s->buckets[b]; idx >= 0; prev = idx, idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_NUMBER &&
            s->entries[idx].hash == h &&
            number_equals(s->entries[idx].num_val, value)) {
            if (prev >= 0) {
                s->entries[prev].next = s->entries[idx].next;
            } else {
                s->buckets[b] = s->entries[idx].next;
            }
            s->entries[idx].val_type = SCRIPTGO_SET_VAL_EMPTY;
            s->entries[idx].next = -1;
            s->size--;
            *out_bool = 1;
            return 0;
        }
    }
    *out_bool = 0;
    return 0;
}

int scriptgo_set_delete_string(void *handle, const char *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    if (value == NULL) value = "";
    uint64_t h = hash_string(value);
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    int64_t prev = -1;
    for (int64_t idx = s->buckets[b]; idx >= 0; prev = idx, idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_STRING &&
            s->entries[idx].hash == h &&
            strcmp(s->entries[idx].str_val, value) == 0) {
            if (prev >= 0) {
                s->entries[prev].next = s->entries[idx].next;
            } else {
                s->buckets[b] = s->entries[idx].next;
            }
            if (s->entries[idx].str_val != NULL) free(s->entries[idx].str_val);
            s->entries[idx].str_val = NULL;
            s->entries[idx].val_type = SCRIPTGO_SET_VAL_EMPTY;
            s->entries[idx].next = -1;
            s->size--;
            *out_bool = 1;
            return 0;
        }
    }
    *out_bool = 0;
    return 0;
}

int scriptgo_set_delete_ptr(void *handle, void *value, int32_t *out_bool) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set delete: invalid handle");
    if (out_bool == NULL) return set_fail("scriptgo set delete: null out_bool");
    uint64_t h = hash_ptr(value);
    int64_t b = (int64_t)(h & (uint64_t)s->bucket_mask);
    int64_t prev = -1;
    for (int64_t idx = s->buckets[b]; idx >= 0; prev = idx, idx = s->entries[idx].next) {
        if (s->entries[idx].val_type == SCRIPTGO_SET_VAL_PTR &&
            s->entries[idx].hash == h &&
            s->entries[idx].ptr_val == value) {
            if (prev >= 0) {
                s->entries[prev].next = s->entries[idx].next;
            } else {
                s->buckets[b] = s->entries[idx].next;
            }
            s->entries[idx].val_type = SCRIPTGO_SET_VAL_EMPTY;
            s->entries[idx].next = -1;
            s->size--;
            *out_bool = 1;
            return 0;
        }
    }
    *out_bool = 0;
    return 0;
}

int scriptgo_set_clear(void *handle) {
    scriptgo_set_native *s = handle;
    if (s == NULL || s->magic != SCRIPTGO_MAGIC_SET) return set_fail("scriptgo set clear: invalid handle");
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING && s->entries[i].str_val != NULL) {
            free(s->entries[i].str_val);
        }
    }
    s->size = 0;
    s->entry_count = 0;
    if (s->buckets != NULL) {
        memset(s->buckets, -1, (size_t)(s->bucket_mask + 1) * sizeof(int64_t));
    }
    return 0;
}

void scriptgo_set_free(void *handle) {
    if (handle == NULL) return;
    scriptgo_set_native *s = handle;
    if (s->magic != SCRIPTGO_MAGIC_SET) {
        free(handle);
        return;
    }
    scriptgo_set_clear(s);
    if (s->entries != NULL) free(s->entries);
    if (s->buckets != NULL) free(s->buckets);
    free(s);
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
    int first = 1;
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
        char val_buf[64];
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_NUMBER) {
            double n = s->entries[i].num_val;
            if (isnan(n)) snprintf(val_buf, sizeof(val_buf), "NaN");
            else if (isinf(n)) snprintf(val_buf, sizeof(val_buf), n > 0 ? "Infinity" : "-Infinity");
            else if (n == (double)(int64_t)n) snprintf(val_buf, sizeof(val_buf), "%lld", (long long)n);
            else snprintf(val_buf, sizeof(val_buf), "%.14g", n);
        } else if (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING) {
            snprintf(val_buf, sizeof(val_buf), "'%s'", s->entries[i].str_val ? s->entries[i].str_val : "");
        } else {
            snprintf(val_buf, sizeof(val_buf), "[object]");
        }
        char entry_buf[128];
        snprintf(entry_buf, sizeof(entry_buf), "%s%s", (first ? " " : ", "), val_buf);
        first = 0;
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
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
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
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_NUMBER) { has_num = 1; break; }
    }
    if (has_num) {
        if (scriptgo_array_new(s->size, sizeof(double), out_array) != 0) return -1;
        scriptgo_set_array_header *arr = *out_array;
        int64_t dst_idx = 0;
        for (int64_t i = 0; i < s->entry_count; i++) {
            if (s->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
            double d = s->entries[i].num_val;
            memcpy(arr->data + (size_t)dst_idx * sizeof(double), &d, sizeof(double));
            dst_idx++;
        }
        return 0;
    }
    if (scriptgo_array_new(s->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_set_array_header *arr = *out_array;
    int64_t dst_idx = 0;
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
        void *val = (s->entries[i].val_type == SCRIPTGO_SET_VAL_STRING) ? (void *)s->entries[i].str_val : s->entries[i].ptr_val;
        memcpy(arr->data + (size_t)dst_idx * sizeof(void *), &val, sizeof(void *));
        dst_idx++;
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
    int64_t dst_idx = 0;
    for (int64_t i = 0; i < s->entry_count; i++) {
        if (s->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
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
        memcpy(arr->data + (size_t)dst_idx * sizeof(void *), &tup, sizeof(void *));
        dst_idx++;
    }
    return 0;
}

int scriptgo_set_new_values_number(void *values_array, void **out_set) {
    if (scriptgo_set_new(out_set) != 0) return -1;
    if (values_array == NULL) return 0;
    scriptgo_set_native *s = *out_set;
    scriptgo_set_array_header *arr = values_array;
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
    scriptgo_set_array_header *arr = values_array;
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
    scriptgo_set_array_header *arr = values_array;
    for (int64_t i = 0; i < arr->length; i++) {
        void *v = *(void **)(arr->data + (size_t)i * sizeof(void *));
        void *dummy;
        scriptgo_set_add_ptr(s, v, &dummy);
    }
    return 0;
}

static int set_add_entry_copy(scriptgo_set_native *dst, const scriptgo_set_native_entry *e) {
    if (dst == NULL || e == NULL || e->val_type == SCRIPTGO_SET_VAL_EMPTY) return 0;
    int64_t idx = set_find_entry_copy(dst, e);
    if (idx >= 0) return 0;
    if (set_ensure_capacity(dst) != 0) return -1;
    int64_t b = (int64_t)(e->hash & (uint64_t)dst->bucket_mask);
    dst->entries[dst->entry_count].hash = e->hash;
    dst->entries[dst->entry_count].val_type = e->val_type;
    dst->entries[dst->entry_count].num_val = e->num_val;
    dst->entries[dst->entry_count].str_val = e->str_val ? strdup(e->str_val) : NULL;
    dst->entries[dst->entry_count].ptr_val = e->ptr_val;
    dst->entries[dst->entry_count].next = dst->buckets[b];
    dst->buckets[b] = dst->entry_count;
    dst->entry_count++;
    dst->size++;
    return 0;
}

int scriptgo_set_union(void *handle_a, void *handle_b, void **out_set) {
    if (out_set == NULL) return set_fail("scriptgo_set_union: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    int64_t capacity = (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET ? sa->size : 0) +
                       (sb != NULL && sb->magic == SCRIPTGO_MAGIC_SET ? sb->size : 0);
    if (scriptgo_set_new_with_capacity(out_set, capacity) != 0) return -1;
    scriptgo_set_native *dst = *out_set;
    if (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sa->entry_count; i++) {
            set_add_entry_copy(dst, &sa->entries[i]);
        }
    }
    if (sb != NULL && sb->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sb->entry_count; i++) {
            set_add_entry_copy(dst, &sb->entries[i]);
        }
    }
    return 0;
}

int scriptgo_set_intersection(void *handle_a, void *handle_b, void **out_set) {
    if (out_set == NULL) return set_fail("scriptgo_set_intersection: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    int64_t capacity = (sa != NULL && sb != NULL && sa->magic == SCRIPTGO_MAGIC_SET && sb->magic == SCRIPTGO_MAGIC_SET) ?
                       (sa->size < sb->size ? sa->size : sb->size) : 0;
    if (scriptgo_set_new_with_capacity(out_set, capacity) != 0) return -1;
    scriptgo_set_native *dst = *out_set;
    if (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET && sb != NULL && sb->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sa->entry_count; i++) {
            if (sa->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
            if (set_find_entry_copy(sb, &sa->entries[i]) >= 0) {
                set_add_entry_copy(dst, &sa->entries[i]);
            }
        }
    }
    return 0;
}

int scriptgo_set_difference(void *handle_a, void *handle_b, void **out_set) {
    if (out_set == NULL) return set_fail("scriptgo_set_difference: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    int64_t capacity = (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET) ? sa->size : 0;
    if (scriptgo_set_new_with_capacity(out_set, capacity) != 0) return -1;
    scriptgo_set_native *dst = *out_set;
    if (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sa->entry_count; i++) {
            if (sa->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
            if (sb == NULL || sb->magic != SCRIPTGO_MAGIC_SET || set_find_entry_copy(sb, &sa->entries[i]) < 0) {
                set_add_entry_copy(dst, &sa->entries[i]);
            }
        }
    }
    return 0;
}

int scriptgo_set_symmetric_difference(void *handle_a, void *handle_b, void **out_set) {
    if (out_set == NULL) return set_fail("scriptgo_set_symmetric_difference: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    int64_t capacity = (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET ? sa->size : 0) +
                       (sb != NULL && sb->magic == SCRIPTGO_MAGIC_SET ? sb->size : 0);
    if (scriptgo_set_new_with_capacity(out_set, capacity) != 0) return -1;
    scriptgo_set_native *dst = *out_set;
    if (sa != NULL && sa->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sa->entry_count; i++) {
            if (sa->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
            if (sb == NULL || sb->magic != SCRIPTGO_MAGIC_SET || set_find_entry_copy(sb, &sa->entries[i]) < 0) {
                set_add_entry_copy(dst, &sa->entries[i]);
            }
        }
    }
    if (sb != NULL && sb->magic == SCRIPTGO_MAGIC_SET) {
        for (int64_t i = 0; i < sb->entry_count; i++) {
            if (sb->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
            if (sa == NULL || sa->magic != SCRIPTGO_MAGIC_SET || set_find_entry_copy(sa, &sb->entries[i]) < 0) {
                set_add_entry_copy(dst, &sb->entries[i]);
            }
        }
    }
    return 0;
}

int scriptgo_set_is_subset_of(void *handle_a, void *handle_b, int32_t *out_bool) {
    if (out_bool == NULL) return set_fail("scriptgo_set_is_subset_of: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    if (sa == NULL || sa->magic != SCRIPTGO_MAGIC_SET) {
        *out_bool = 1;
        return 0;
    }
    if (sb == NULL || sb->magic != SCRIPTGO_MAGIC_SET) {
        *out_bool = (sa->size == 0) ? 1 : 0;
        return 0;
    }
    if (sa->size > sb->size) {
        *out_bool = 0;
        return 0;
    }
    for (int64_t i = 0; i < sa->entry_count; i++) {
        if (sa->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
        if (set_find_entry_copy(sb, &sa->entries[i]) < 0) {
            *out_bool = 0;
            return 0;
        }
    }
    *out_bool = 1;
    return 0;
}

int scriptgo_set_is_superset_of(void *handle_a, void *handle_b, int32_t *out_bool) {
    return scriptgo_set_is_subset_of(handle_b, handle_a, out_bool);
}

int scriptgo_set_is_disjoint_from(void *handle_a, void *handle_b, int32_t *out_bool) {
    if (out_bool == NULL) return set_fail("scriptgo_set_is_disjoint_from: null output");
    scriptgo_set_native *sa = handle_a;
    scriptgo_set_native *sb = handle_b;
    if (sa == NULL || sa->magic != SCRIPTGO_MAGIC_SET || sb == NULL || sb->magic != SCRIPTGO_MAGIC_SET) {
        *out_bool = 1;
        return 0;
    }
    for (int64_t i = 0; i < sa->entry_count; i++) {
        if (sa->entries[i].val_type == SCRIPTGO_SET_VAL_EMPTY) continue;
        if (set_find_entry_copy(sb, &sa->entries[i]) >= 0) {
            *out_bool = 0;
            return 0;
        }
    }
    *out_bool = 1;
    return 0;
}
