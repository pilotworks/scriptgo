#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_map_set_string_string(void *handle, const char *key, const char *value, void **out_map);
int scriptgo_map_set_string_number(void *handle, const char *key, double value, void **out_map);
int scriptgo_map_set_string_bigint(void *handle, const char *key, int64_t value, void **out_map);
int scriptgo_map_set_string_ptr(void *handle, const char *key, void *value, void **out_map);

#define SCRIPTGO_MAGIC_MAP 0x4D415031 // "MAP1"

typedef enum {
    SCRIPTGO_MAP_VAL_NUMBER = 1,
    SCRIPTGO_MAP_VAL_STRING = 2,
    SCRIPTGO_MAP_VAL_PTR = 3,
    SCRIPTGO_MAP_VAL_BIGINT = 4
} scriptgo_map_val_type;

typedef struct {
    char *key_str;
    scriptgo_map_val_type val_type;
    double num_val;
    int64_t bigint_val;
    char *str_val;
    void *ptr_val;
} scriptgo_map_native_entry;

typedef struct {
    uint32_t magic;
    int64_t size;
    int64_t capacity;
    scriptgo_map_native_entry *entries;
} scriptgo_map_native;

static int map_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static char *num_to_str(double n) {
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

int scriptgo_map_new(void **out_map) {
    if (out_map == NULL) return map_fail("scriptgo map new: null out_map");
    scriptgo_map_native *m = calloc(1, sizeof(scriptgo_map_native));
    if (m == NULL) return map_fail("scriptgo map new: out of memory");
    m->magic = SCRIPTGO_MAGIC_MAP;
    m->size = 0;
    m->capacity = 8;
    m->entries = calloc(m->capacity, sizeof(scriptgo_map_native_entry));
    if (m->entries == NULL) {
        free(m);
        return map_fail("scriptgo map new: out of memory");
    }
    *out_map = m;
    return 0;
}

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_array_inner_map;

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    void *boxed_fields;
    uintptr_t fields[];
} scriptgo_object_inner_map;

int scriptgo_map_new_entries(void *entries_array, void **out_map) {
    if (scriptgo_map_new(out_map) != 0) return -1;
    if (entries_array == NULL) return 0;
    scriptgo_map_native *m = *out_map;
    uint32_t *magic_check = (uint32_t *)entries_array;
    if (*magic_check == SCRIPTGO_MAGIC_MAP) {
        scriptgo_map_native *src = (scriptgo_map_native *)entries_array;
        for (int64_t i = 0; i < src->size; i++) {
            scriptgo_map_native_entry *e = &src->entries[i];
            if (e->key_str == NULL) continue;
            void *dummy;
            if (e->val_type == SCRIPTGO_MAP_VAL_NUMBER) {
                scriptgo_map_set_string_number(m, e->key_str, e->num_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_STRING) {
                scriptgo_map_set_string_string(m, e->key_str, e->str_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_BIGINT) {
                scriptgo_map_set_string_bigint(m, e->key_str, e->bigint_val, &dummy);
            } else {
                scriptgo_map_set_string_ptr(m, e->key_str, e->ptr_val, &dummy);
            }
        }
        return 0;
    }
    scriptgo_object_inner_map *root_obj = (scriptgo_object_inner_map *)entries_array;
    if (root_obj->magic == 0x53474F424A454354ULL) {
        for (int64_t i = 0; i < root_obj->field_count; i++) {
            void *item = (void *)root_obj->fields[i];
            if (item == NULL) continue;
            scriptgo_object_inner_map *obj = (scriptgo_object_inner_map *)item;
            if (obj->magic == 0x53474F424A454354ULL && obj->field_count >= 2) {
                const char *k = (const char *)obj->fields[0];
                const char *v = (const char *)obj->fields[1];
                void *dummy;
                scriptgo_map_set_string_string(m, k, v, &dummy);
            } else {
                scriptgo_array_inner_map *sub_arr = (scriptgo_array_inner_map *)item;
                if (sub_arr->length >= 2 && sub_arr->data != NULL) {
                    const char *k = *(const char **)(sub_arr->data);
                    const char *v = *(const char **)(sub_arr->data + sizeof(void *));
                    void *dummy;
                    scriptgo_map_set_string_string(m, k, v, &dummy);
                }
            }
        }
        return 0;
    }
    scriptgo_array_inner_map *arr = entries_array;
    if (arr->data == NULL) return 0;
    if (arr->element_size == sizeof(void *) || arr->element_size == sizeof(scriptgo_value)) {
        for (int64_t i = 0; i < arr->length; i++) {
            void *item = NULL;
            if (arr->element_size == sizeof(scriptgo_value)) {
                item = (void *)*(uintptr_t *)(arr->data + (size_t)i * sizeof(scriptgo_value) + 8);
            } else {
                item = *(void **)(arr->data + (size_t)i * sizeof(void *));
            }
            if (item == NULL) continue;
            scriptgo_object_inner_map *obj = (scriptgo_object_inner_map *)item;
            if (obj->magic == 0x53474F424A454354ULL && obj->field_count >= 2) {
                const char *k = (const char *)obj->fields[0];
                const char *v = (const char *)obj->fields[1];
                void *dummy;
                scriptgo_map_set_string_string(m, k, v, &dummy);
            } else {
                scriptgo_array_inner_map *sub_arr = (scriptgo_array_inner_map *)item;
                if (sub_arr->length >= 2 && sub_arr->data != NULL) {
                    const char *k = NULL;
                    const char *v = NULL;
                    if (sub_arr->element_size == sizeof(scriptgo_value)) {
                        k = (const char *)*(uintptr_t *)(sub_arr->data + 8);
                        v = (const char *)*(uintptr_t *)(sub_arr->data + sizeof(scriptgo_value) + 8);
                    } else {
                        k = *(const char **)(sub_arr->data);
                        v = *(const char **)(sub_arr->data + sizeof(void *));
                    }
                    void *dummy;
                    scriptgo_map_set_string_string(m, k, v, &dummy);
                }
            }
        }
    }
    return 0;
}

static int map_ensure_capacity(scriptgo_map_native *m) {
    if (m->size >= m->capacity) {
        int64_t new_cap = m->capacity * 2;
        if (new_cap < 8) new_cap = 8;
        scriptgo_map_native_entry *new_entries = realloc(m->entries, (size_t)new_cap * sizeof(scriptgo_map_native_entry));
        if (new_entries == NULL) return map_fail("scriptgo map set: out of memory");
        m->entries = new_entries;
        m->capacity = new_cap;
    }
    return 0;
}

static int64_t map_find_entry(scriptgo_map_native *m, const char *key_str) {
    if (m == NULL || key_str == NULL) return -1;
    for (int64_t i = 0; i < m->size; i++) {
        if (m->entries[i].key_str != NULL && strcmp(m->entries[i].key_str, key_str) == 0) {
            return i;
        }
    }
    return -1;
}

int scriptgo_map_set_string_number(void *handle, const char *key, double value, void **out_map) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map set: invalid handle");
    if (key == NULL) key = "";
    int64_t idx = map_find_entry(m, key);
    if (idx >= 0) {
        m->entries[idx].val_type = SCRIPTGO_MAP_VAL_NUMBER;
        m->entries[idx].num_val = value;
    } else {
        if (map_ensure_capacity(m) != 0) return -1;
        m->entries[m->size].key_str = strdup(key);
        m->entries[m->size].val_type = SCRIPTGO_MAP_VAL_NUMBER;
        m->entries[m->size].num_val = value;
        m->size++;
    }
    if (out_map != NULL) *out_map = m;
    return 0;
}

int scriptgo_map_set_string_string(void *handle, const char *key, const char *value, void **out_map) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map set: invalid handle");
    if (key == NULL) key = "";
    if (value == NULL) value = "";
    int64_t idx = map_find_entry(m, key);
    if (idx >= 0) {
        if (m->entries[idx].val_type == SCRIPTGO_MAP_VAL_STRING && m->entries[idx].str_val != NULL) {
            free(m->entries[idx].str_val);
        }
        m->entries[idx].val_type = SCRIPTGO_MAP_VAL_STRING;
        m->entries[idx].str_val = strdup(value);
    } else {
        if (map_ensure_capacity(m) != 0) return -1;
        m->entries[m->size].key_str = strdup(key);
        m->entries[m->size].val_type = SCRIPTGO_MAP_VAL_STRING;
        m->entries[m->size].str_val = strdup(value);
        m->size++;
    }
    if (out_map != NULL) *out_map = m;
    return 0;
}

int scriptgo_map_set_string_bigint(void *handle, const char *key, int64_t value, void **out_map) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map set: invalid handle");
    if (key == NULL) key = "";
    int64_t idx = map_find_entry(m, key);
    if (idx >= 0) {
        m->entries[idx].val_type = SCRIPTGO_MAP_VAL_BIGINT;
        m->entries[idx].bigint_val = value;
    } else {
        if (map_ensure_capacity(m) != 0) return -1;
        m->entries[m->size].key_str = strdup(key);
        m->entries[m->size].val_type = SCRIPTGO_MAP_VAL_BIGINT;
        m->entries[m->size].bigint_val = value;
        m->size++;
    }
    if (out_map != NULL) *out_map = m;
    return 0;
}

int scriptgo_map_set_string_ptr(void *handle, const char *key, void *value, void **out_map) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map set: invalid handle");
    if (key == NULL) key = "";
    int64_t idx = map_find_entry(m, key);
    if (idx >= 0) {
        m->entries[idx].val_type = SCRIPTGO_MAP_VAL_PTR;
        m->entries[idx].ptr_val = value;
    } else {
        if (map_ensure_capacity(m) != 0) return -1;
        m->entries[m->size].key_str = strdup(key);
        m->entries[m->size].val_type = SCRIPTGO_MAP_VAL_PTR;
        m->entries[m->size].ptr_val = value;
        m->size++;
    }
    if (out_map != NULL) *out_map = m;
    return 0;
}

int scriptgo_map_set_number_number(void *handle, double key, double value, void **out_map) {
    char *kstr = num_to_str(key);
    int res = scriptgo_map_set_string_number(handle, kstr, value, out_map);
    free(kstr);
    return res;
}

int scriptgo_map_set_number_string(void *handle, double key, const char *value, void **out_map) {
    char *kstr = num_to_str(key);
    int res = scriptgo_map_set_string_string(handle, kstr, value, out_map);
    free(kstr);
    return res;
}

int scriptgo_map_set_number_ptr(void *handle, double key, void *value, void **out_map) {
    char *kstr = num_to_str(key);
    int res = scriptgo_map_set_string_ptr(handle, kstr, value, out_map);
    free(kstr);
    return res;
}

int scriptgo_map_get_number(void *handle, const char *key_str, double key_num, int32_t key_is_str, double *out_val) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map get: invalid handle");
    if (out_val == NULL) return map_fail("scriptgo map get: null out_val");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    if (idx >= 0 && m->entries[idx].val_type == SCRIPTGO_MAP_VAL_NUMBER) {
        *out_val = m->entries[idx].num_val;
    } else {
        *out_val = nan("");
    }
    return 0;
}

int scriptgo_map_get_string(void *handle, const char *key_str, double key_num, int32_t key_is_str, char **out_val) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map get: invalid handle");
    if (out_val == NULL) return map_fail("scriptgo map get: null out_val");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    if (idx >= 0 && m->entries[idx].val_type == SCRIPTGO_MAP_VAL_STRING) {
        *out_val = m->entries[idx].str_val;
    } else {
        *out_val = "";
    }
    return 0;
}

int scriptgo_map_get_bigint(void *handle, const char *key_str, double key_num, int32_t key_is_str, int64_t *out_val) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map get: invalid handle");
    if (out_val == NULL) return map_fail("scriptgo map get: null out_val");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    if (idx >= 0 && m->entries[idx].val_type == SCRIPTGO_MAP_VAL_BIGINT) {
        *out_val = m->entries[idx].bigint_val;
    } else {
        *out_val = 0;
    }
    return 0;
}

int scriptgo_map_get_ptr(void *handle, const char *key_str, double key_num, int32_t key_is_str, void **out_val) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map get: invalid handle");
    if (out_val == NULL) return map_fail("scriptgo map get: null out_val");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    if (idx >= 0 && m->entries[idx].val_type == SCRIPTGO_MAP_VAL_PTR) {
        *out_val = m->entries[idx].ptr_val;
    } else {
        *out_val = NULL;
    }
    return 0;
}

int scriptgo_map_has(void *handle, const char *key_str, double key_num, int32_t key_is_str, int32_t *out_bool) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map has: invalid handle");
    if (out_bool == NULL) return map_fail("scriptgo map has: null out_bool");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    *out_bool = (idx >= 0) ? 1 : 0;
    return 0;
}

int scriptgo_map_delete(void *handle, const char *key_str, double key_num, int32_t key_is_str, int32_t *out_bool) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map delete: invalid handle");
    if (out_bool == NULL) return map_fail("scriptgo map delete: null out_bool");
    char *alloc_key = NULL;
    const char *lookup_key = key_str;
    if (!key_is_str) {
        alloc_key = num_to_str(key_num);
        lookup_key = alloc_key;
    }
    int64_t idx = map_find_entry(m, lookup_key);
    if (alloc_key != NULL) free(alloc_key);
    if (idx < 0) {
        *out_bool = 0;
        return 0;
    }
    if (m->entries[idx].key_str != NULL) free(m->entries[idx].key_str);
    if (m->entries[idx].val_type == SCRIPTGO_MAP_VAL_STRING && m->entries[idx].str_val != NULL) {
        free(m->entries[idx].str_val);
    }
    for (int64_t i = idx; i < m->size - 1; i++) {
        m->entries[i] = m->entries[i + 1];
    }
    m->size--;
    *out_bool = 1;
    return 0;
}

int scriptgo_map_clear(void *handle) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map clear: invalid handle");
    for (int64_t i = 0; i < m->size; i++) {
        if (m->entries[i].key_str != NULL) free(m->entries[i].key_str);
        if (m->entries[i].val_type == SCRIPTGO_MAP_VAL_STRING && m->entries[i].str_val != NULL) {
            free(m->entries[i].str_val);
        }
    }
    m->size = 0;
    return 0;
}

int scriptgo_map_size(void *handle, double *out_size) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map size: invalid handle");
    if (out_size == NULL) return map_fail("scriptgo map size: null out_size");
    *out_size = (double)m->size;
    return 0;
}

int scriptgo_console_inspect_array(void *value, char **out_str);
int scriptgo_json_inspect_object(void *handle, char **out_str);
int scriptgo_gc_get_tag(void *ptr);
extern const char scriptgo_undefined_sentinel;

/* map_inspect_value renders an entry's value as console.log does. */
static char *map_inspect_value(const scriptgo_map_native_entry *entry) {
    char buf[64];
    char *inspected = NULL;
    switch (entry->val_type) {
    case SCRIPTGO_MAP_VAL_NUMBER:
        return num_to_str(entry->num_val);
    case SCRIPTGO_MAP_VAL_BIGINT:
        snprintf(buf, sizeof(buf), "%lldn", (long long)entry->bigint_val);
        return strdup(buf);
    case SCRIPTGO_MAP_VAL_STRING: {
        const char *text = entry->str_val ? entry->str_val : "";
        size_t length = strlen(text);
        char *quoted = malloc(length + 3);
        if (quoted == NULL) return NULL;
        quoted[0] = '\'';
        memcpy(quoted + 1, text, length);
        quoted[length + 1] = '\'';
        quoted[length + 2] = '\0';
        return quoted;
    }
    default:
        break;
    }
    void *value = entry->ptr_val;
    if (value == NULL) return strdup("null");
    if (value == (void *)&scriptgo_undefined_sentinel) return strdup("undefined");
    int gc_tag = scriptgo_gc_get_tag(value);
    if (gc_tag == 2 && scriptgo_console_inspect_array(value, &inspected) == 0 && inspected != NULL) return inspected;
    if (gc_tag == 3) return strdup("[Function (anonymous)]");
    if (*(uint64_t *)value == 0x53474F424A454354ULL && scriptgo_json_inspect_object(value, &inspected) == 0 && inspected != NULL) {
        return inspected;
    }
    return strdup("[object]");
}

static int map_append(char **buf, size_t *length, size_t *capacity, const char *text) {
    size_t extra = strlen(text);
    if (*length + extra + 1 > *capacity) {
        size_t next = (*capacity == 0 ? 64 : *capacity * 2);
        while (next < *length + extra + 1) next *= 2;
        char *grown = realloc(*buf, next);
        if (grown == NULL) return -1;
        *buf = grown;
        *capacity = next;
    }
    memcpy(*buf + *length, text, extra + 1);
    *length += extra;
    return 0;
}

int scriptgo_map_to_string(void *handle, char **out_str) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) {
        if (out_str != NULL) *out_str = strdup("Map(0) {}");
        return 0;
    }
    if (out_str == NULL) return map_fail("scriptgo map toString: null out_str");
    char *buf = NULL;
    size_t length = 0, capacity = 0;
    char header[48];
    snprintf(header, sizeof(header), "Map(%lld) {", (long long)m->size);
    if (map_append(&buf, &length, &capacity, header) != 0) goto fail;
    for (int64_t i = 0; i < m->size; i++) {
        char *value = map_inspect_value(&m->entries[i]);
        if (value == NULL) goto fail;
        const char *key = m->entries[i].key_str ? m->entries[i].key_str : "";
        int failed = map_append(&buf, &length, &capacity, i == 0 ? " '" : ", '") != 0 ||
                     map_append(&buf, &length, &capacity, key) != 0 ||
                     map_append(&buf, &length, &capacity, "' => ") != 0 ||
                     map_append(&buf, &length, &capacity, value) != 0;
        free(value);
        if (failed) goto fail;
    }
    if (map_append(&buf, &length, &capacity, m->size > 0 ? " }" : "}") != 0) goto fail;
    *out_str = buf;
    return 0;
fail:
    free(buf);
    return map_fail("scriptgo map toString: out of memory");
}

int scriptgo_closure_invoke(void *closure_handle, int32_t arg_count, const scriptgo_boxed_value *a1, const scriptgo_boxed_value *a2, const scriptgo_boxed_value *a3, const scriptgo_boxed_value *a4);

int scriptgo_map_for_each(void *handle, void *closure_handle) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || closure_handle == NULL) {
        return map_fail("scriptgo map forEach: invalid arguments");
    }
    for (int64_t i = 0; i < m->size; i++) {
        scriptgo_map_native_entry *entry = &m->entries[i];
        scriptgo_boxed_value a1 = {0};
        if (entry->val_type == SCRIPTGO_MAP_VAL_NUMBER) {
            union { double d; int64_t i; } u;
            u.d = entry->num_val;
            a1.tag = 3;
            a1.payload = u.i;
        } else if (entry->val_type == SCRIPTGO_MAP_VAL_STRING) {
            a1.tag = 4;
            a1.payload = (int64_t)(uintptr_t)(entry->str_val ? entry->str_val : "");
        } else if (entry->val_type == SCRIPTGO_MAP_VAL_BIGINT) {
            a1.tag = 8;
            a1.payload = entry->bigint_val;
        } else {
            a1.tag = 5;
            a1.payload = (int64_t)(uintptr_t)entry->ptr_val;
        }

        scriptgo_boxed_value a2 = {0};
        a2.tag = 4;
        a2.payload = (int64_t)(uintptr_t)(entry->key_str ? entry->key_str : "");

        scriptgo_boxed_value a3 = {0};
        a3.tag = 5;
        a3.payload = (int64_t)(uintptr_t)m;

        scriptgo_boxed_value a4 = {0};

        scriptgo_closure_invoke(closure_handle, 3, &a1, &a2, &a3, &a4);
    }
    return 0;
}

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_array_header;

typedef struct {
    void *key;
    void *val;
} scriptgo_map_entry_tuple;

int scriptgo_map_keys(void *handle, void **out_array) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || out_array == NULL) return map_fail("invalid map handle");
    if (scriptgo_array_new(m->size, sizeof(char *), out_array) != 0) return -1;
    scriptgo_array_header *arr = *out_array;
    for (int64_t i = 0; i < m->size; i++) {
        char *k = strdup(m->entries[i].key_str ? m->entries[i].key_str : "");
        memcpy(arr->data + (size_t)i * sizeof(char *), &k, sizeof(char *));
    }
    return 0;
}

int scriptgo_map_values(void *handle, void **out_array) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || out_array == NULL) return map_fail("invalid map handle");
    int has_num = 0;
    for (int64_t i = 0; i < m->size; i++) {
        if (m->entries[i].val_type == SCRIPTGO_MAP_VAL_NUMBER) { has_num = 1; break; }
    }
    if (has_num) {
        if (scriptgo_array_new(m->size, sizeof(double), out_array) != 0) return -1;
        scriptgo_array_header *arr = *out_array;
        for (int64_t i = 0; i < m->size; i++) {
            double d = m->entries[i].num_val;
            memcpy(arr->data + (size_t)i * sizeof(double), &d, sizeof(double));
        }
        return 0;
    }
    if (scriptgo_array_new(m->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_array_header *arr = *out_array;
    for (int64_t i = 0; i < m->size; i++) {
        void *val = (m->entries[i].val_type == SCRIPTGO_MAP_VAL_STRING) ? (void *)m->entries[i].str_val : m->entries[i].ptr_val;
        memcpy(arr->data + (size_t)i * sizeof(void *), &val, sizeof(void *));
    }
    return 0;
}

int scriptgo_object_new(int64_t field_count, void **out_object);
int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);

int scriptgo_map_entries(void *handle, void **out_array) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || out_array == NULL) return map_fail("invalid map handle");
    if (scriptgo_array_new(m->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_array_header *arr = *out_array;
    for (int64_t i = 0; i < m->size; i++) {
        void *tup = NULL;
        if (scriptgo_object_new(2, &tup) != 0) return -1;
        scriptgo_object_string_set(tup, 0, m->entries[i].key_str ? m->entries[i].key_str : "");
        if (m->entries[i].val_type == SCRIPTGO_MAP_VAL_NUMBER) {
            scriptgo_object_number_set(tup, 1, m->entries[i].num_val);
        } else if (m->entries[i].val_type == SCRIPTGO_MAP_VAL_STRING) {
            scriptgo_object_string_set(tup, 1, m->entries[i].str_val ? m->entries[i].str_val : "");
        } else {
            scriptgo_object_ptr_set(tup, 1, m->entries[i].ptr_val);
        }
        memcpy(arr->data + (size_t)i * sizeof(void *), &tup, sizeof(void *));
    }
    return 0;
}
