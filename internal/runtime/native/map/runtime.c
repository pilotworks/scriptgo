#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

#define SCRIPTGO_MAGIC_MAP 0x4D415031 // "MAP1"

typedef enum {
    SCRIPTGO_MAP_VAL_NUMBER = 1,
    SCRIPTGO_MAP_VAL_STRING = 2,
    SCRIPTGO_MAP_VAL_PTR = 3,
    SCRIPTGO_MAP_VAL_BIGINT = 4,
    SCRIPTGO_MAP_VAL_BOOL = 5,  /* stored in bigint_val as 0 or 1 */
    SCRIPTGO_MAP_VAL_BOXED = 6  /* val_tag and its payload in bigint_val */
} scriptgo_map_val_type;

/* A key keeps its JavaScript type as a value tag and payload (the
 * scriptgo_value encoding): a number key never aliases its string form.
 * String keys own a copy of their text. */
typedef struct {
    uint32_t key_tag;
    uint64_t key_payload;
    scriptgo_map_val_type val_type;
    uint32_t val_tag;
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

int scriptgo_map_set_number(void *handle, uint32_t key_tag, uint64_t key_payload, double value, void **out_map);
int scriptgo_map_set_string(void *handle, uint32_t key_tag, uint64_t key_payload, const char *value, void **out_map);
int scriptgo_map_set_bigint(void *handle, uint32_t key_tag, uint64_t key_payload, int64_t value, void **out_map);
int scriptgo_map_set_ptr(void *handle, uint32_t key_tag, uint64_t key_payload, void *value, void **out_map);
int scriptgo_map_set_bool(void *handle, uint32_t key_tag, uint64_t key_payload, int32_t value, void **out_map);
int scriptgo_map_set_unknown(void *handle, uint32_t key_tag, uint64_t key_payload, uint32_t value_tag, uint64_t value_payload, void **out_map);

static int map_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static double map_bits_to_double(uint64_t bits) {
    double d;
    memcpy(&d, &bits, sizeof(d));
    return d;
}

static uint64_t map_double_to_bits(double d) {
    uint64_t bits;
    memcpy(&bits, &d, sizeof(bits));
    return bits;
}

/* map_key_is_reference reports tags compared by identity. */
static int map_key_is_reference(uint32_t tag) {
    return tag == SCRIPTGO_TAG_OBJECT || tag == SCRIPTGO_TAG_ARRAY || tag == SCRIPTGO_TAG_FUNCTION ||
           tag == SCRIPTGO_TAG_SYMBOL || tag == SCRIPTGO_TAG_PROMISE;
}

/* map_key_normalize applies SameValueZero's representation: -0 is stored as
 * +0, booleans as 0/1 and an absent string as "". */
static void map_key_normalize(uint32_t *tag, uint64_t *payload) {
    if (*tag == SCRIPTGO_TAG_NUMBER && map_bits_to_double(*payload) == 0.0) *payload = 0;
    if (*tag == SCRIPTGO_TAG_BOOLEAN) *payload = *payload != 0;
    if (*tag == SCRIPTGO_TAG_UNDEFINED || *tag == SCRIPTGO_TAG_NULL) *payload = 0;
    if (*tag == SCRIPTGO_TAG_STRING && *payload == 0) *payload = (uint64_t)(uintptr_t)"";
}

/* map_key_equals implements SameValueZero for normalized keys. */
static int map_key_equals(const scriptgo_map_native_entry *entry, uint32_t tag, uint64_t payload) {
    if (map_key_is_reference(entry->key_tag) && map_key_is_reference(tag)) return entry->key_payload == payload;
    if (entry->key_tag != tag) return 0;
    switch (tag) {
    case SCRIPTGO_TAG_NUMBER: {
        double a = map_bits_to_double(entry->key_payload), b = map_bits_to_double(payload);
        return a == b || (isnan(a) && isnan(b));
    }
    case SCRIPTGO_TAG_STRING:
        return strcmp((const char *)(uintptr_t)entry->key_payload, (const char *)(uintptr_t)payload) == 0;
    default:
        return entry->key_payload == payload;
    }
}

/* map_entry_release_value frees a value's owned string text. */
static void map_entry_release_value(scriptgo_map_native_entry *entry) {
    if (entry->val_type == SCRIPTGO_MAP_VAL_STRING) free(entry->str_val);
    if (entry->val_type == SCRIPTGO_MAP_VAL_BOXED && entry->val_tag == SCRIPTGO_TAG_STRING) {
        free((void *)(uintptr_t)entry->bigint_val);
    }
    entry->str_val = NULL;
}

static void map_entry_release(scriptgo_map_native_entry *entry) {
    if (entry->key_tag == SCRIPTGO_TAG_STRING) free((void *)(uintptr_t)entry->key_payload);
    map_entry_release_value(entry);
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

static int map_set_string_pair(scriptgo_map_native *m, const char *key, const char *value) {
    void *dummy;
    return scriptgo_map_set_string(m, SCRIPTGO_TAG_STRING, (uint64_t)(uintptr_t)key, value, &dummy);
}

int scriptgo_map_new_entries(void *entries_array, void **out_map) {
    if (scriptgo_map_new(out_map) != 0) return -1;
    if (entries_array == NULL) return 0;
    scriptgo_map_native *m = *out_map;
    uint32_t *magic_check = (uint32_t *)entries_array;
    if (*magic_check == SCRIPTGO_MAGIC_MAP) {
        scriptgo_map_native *src = (scriptgo_map_native *)entries_array;
        for (int64_t i = 0; i < src->size; i++) {
            scriptgo_map_native_entry *e = &src->entries[i];
            void *dummy;
            if (e->val_type == SCRIPTGO_MAP_VAL_NUMBER) {
                scriptgo_map_set_number(m, e->key_tag, e->key_payload, e->num_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_STRING) {
                scriptgo_map_set_string(m, e->key_tag, e->key_payload, e->str_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_BIGINT) {
                scriptgo_map_set_bigint(m, e->key_tag, e->key_payload, e->bigint_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_BOOL) {
                scriptgo_map_set_bool(m, e->key_tag, e->key_payload, (int32_t)e->bigint_val, &dummy);
            } else if (e->val_type == SCRIPTGO_MAP_VAL_BOXED) {
                scriptgo_map_set_unknown(m, e->key_tag, e->key_payload, e->val_tag, (uint64_t)e->bigint_val, &dummy);
            } else {
                scriptgo_map_set_ptr(m, e->key_tag, e->key_payload, e->ptr_val, &dummy);
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
                map_set_string_pair(m, (const char *)obj->fields[0], (const char *)obj->fields[1]);
            } else {
                scriptgo_array_inner_map *sub_arr = (scriptgo_array_inner_map *)item;
                if (sub_arr->length >= 2 && sub_arr->data != NULL) {
                    map_set_string_pair(m, *(const char **)(sub_arr->data), *(const char **)(sub_arr->data + sizeof(void *)));
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
                map_set_string_pair(m, (const char *)obj->fields[0], (const char *)obj->fields[1]);
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
                    map_set_string_pair(m, k, v);
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

static int64_t map_find_entry(scriptgo_map_native *m, uint32_t tag, uint64_t payload) {
    map_key_normalize(&tag, &payload);
    for (int64_t i = 0; i < m->size; i++) {
        if (map_key_equals(&m->entries[i], tag, payload)) return i;
    }
    return -1;
}

/* map_slot returns the entry for a key, appending a new one (with its own
 * copy of a string key) when the key is absent. The previous value of an
 * existing entry is released. */
static scriptgo_map_native_entry *map_slot(void *handle, uint32_t key_tag, uint64_t key_payload) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) {
        map_fail("scriptgo map set: invalid handle");
        return NULL;
    }
    int64_t idx = map_find_entry(m, key_tag, key_payload);
    if (idx >= 0) {
        scriptgo_map_native_entry *entry = &m->entries[idx];
        map_entry_release_value(entry);
        return entry;
    }
    if (map_ensure_capacity(m) != 0) return NULL;
    map_key_normalize(&key_tag, &key_payload);
    if (key_tag == SCRIPTGO_TAG_STRING) {
        char *copy = strdup((const char *)(uintptr_t)key_payload);
        if (copy == NULL) {
            map_fail("scriptgo map set: out of memory");
            return NULL;
        }
        key_payload = (uint64_t)(uintptr_t)copy;
    }
    scriptgo_map_native_entry *entry = &m->entries[m->size++];
    memset(entry, 0, sizeof(*entry));
    entry->key_tag = key_tag;
    entry->key_payload = key_payload;
    return entry;
}

int scriptgo_map_set_number(void *handle, uint32_t key_tag, uint64_t key_payload, double value, void **out_map) {
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) return -1;
    entry->val_type = SCRIPTGO_MAP_VAL_NUMBER;
    entry->num_val = value;
    if (out_map != NULL) *out_map = handle;
    return 0;
}

int scriptgo_map_set_string(void *handle, uint32_t key_tag, uint64_t key_payload, const char *value, void **out_map) {
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) return -1;
    entry->val_type = SCRIPTGO_MAP_VAL_STRING;
    entry->str_val = strdup(value != NULL ? value : "");
    if (out_map != NULL) *out_map = handle;
    return 0;
}

int scriptgo_map_set_bigint(void *handle, uint32_t key_tag, uint64_t key_payload, int64_t value, void **out_map) {
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) return -1;
    entry->val_type = SCRIPTGO_MAP_VAL_BIGINT;
    entry->bigint_val = value;
    if (out_map != NULL) *out_map = handle;
    return 0;
}

int scriptgo_map_set_ptr(void *handle, uint32_t key_tag, uint64_t key_payload, void *value, void **out_map) {
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) return -1;
    entry->val_type = SCRIPTGO_MAP_VAL_PTR;
    entry->ptr_val = value;
    if (out_map != NULL) *out_map = handle;
    return 0;
}

int scriptgo_map_set_bool(void *handle, uint32_t key_tag, uint64_t key_payload, int32_t value, void **out_map) {
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) return -1;
    entry->val_type = SCRIPTGO_MAP_VAL_BOOL;
    entry->bigint_val = value != 0;
    if (out_map != NULL) *out_map = handle;
    return 0;
}

int scriptgo_map_set_unknown(void *handle, uint32_t key_tag, uint64_t key_payload, uint32_t value_tag, uint64_t value_payload, void **out_map) {
    if (value_tag == SCRIPTGO_TAG_STRING) {
        char *copy = strdup(value_payload != 0 ? (const char *)(uintptr_t)value_payload : "");
        if (copy == NULL) return map_fail("scriptgo map set: out of memory");
        value_payload = (uint64_t)(uintptr_t)copy;
    }
    scriptgo_map_native_entry *entry = map_slot(handle, key_tag, key_payload);
    if (entry == NULL) {
        if (value_tag == SCRIPTGO_TAG_STRING) free((void *)(uintptr_t)value_payload);
        return -1;
    }
    entry->val_type = SCRIPTGO_MAP_VAL_BOXED;
    entry->val_tag = value_tag;
    entry->bigint_val = (int64_t)value_payload;
    if (out_map != NULL) *out_map = handle;
    return 0;
}

extern const char scriptgo_undefined_sentinel;

/* map_entry_value is an entry's value as a tagged value. */
static scriptgo_value map_entry_value(const scriptgo_map_native_entry *entry) {
    scriptgo_value value = {0};
    switch (entry->val_type) {
    case SCRIPTGO_MAP_VAL_NUMBER:
        value.tag = SCRIPTGO_TAG_NUMBER;
        value.payload = map_double_to_bits(entry->num_val);
        break;
    case SCRIPTGO_MAP_VAL_STRING:
        value.tag = SCRIPTGO_TAG_STRING;
        value.payload = (uint64_t)(uintptr_t)(entry->str_val ? entry->str_val : "");
        break;
    case SCRIPTGO_MAP_VAL_BIGINT:
        value.tag = SCRIPTGO_TAG_BIGINT;
        value.payload = (uint64_t)entry->bigint_val;
        break;
    case SCRIPTGO_MAP_VAL_BOOL:
        value.tag = SCRIPTGO_TAG_BOOLEAN;
        value.payload = (uint64_t)entry->bigint_val;
        break;
    case SCRIPTGO_MAP_VAL_BOXED:
        value.tag = entry->val_tag;
        value.payload = (uint64_t)entry->bigint_val;
        break;
    default:
        if (entry->ptr_val == (void *)&scriptgo_undefined_sentinel) {
            value.tag = SCRIPTGO_TAG_UNDEFINED;
        } else {
            value.tag = entry->ptr_val == NULL ? SCRIPTGO_TAG_NULL : SCRIPTGO_TAG_OBJECT;
            value.payload = (uint64_t)(uintptr_t)entry->ptr_val;
        }
        break;
    }
    return value;
}

/* map_lookup finds the entry for a key, or NULL when absent. */
static int map_lookup(void *handle, uint32_t key_tag, uint64_t key_payload, const void *out, const char *op, scriptgo_map_native_entry **found) {
    scriptgo_map_native *m = handle;
    *found = NULL;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail(op);
    if (out == NULL) return map_fail(op);
    int64_t idx = map_find_entry(m, key_tag, key_payload);
    if (idx >= 0) *found = &m->entries[idx];
    return 0;
}

int scriptgo_map_get_number(void *handle, uint32_t key_tag, uint64_t key_payload, double *out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = entry != NULL ? map_entry_value(entry) : (scriptgo_value){0};
    /* An absent key reads as undefined: number storage's undefined marker. */
    *out_val = map_bits_to_double(value.tag == SCRIPTGO_TAG_NUMBER ? value.payload : SCRIPTGO_NUMBER_UNDEFINED_BITS);
    return 0;
}

int scriptgo_map_get_string(void *handle, uint32_t key_tag, uint64_t key_payload, char **out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = entry != NULL ? map_entry_value(entry) : (scriptgo_value){0};
    *out_val = value.tag == SCRIPTGO_TAG_STRING ? (char *)(uintptr_t)value.payload : (char *)&scriptgo_undefined_sentinel;
    return 0;
}

int scriptgo_map_get_bigint(void *handle, uint32_t key_tag, uint64_t key_payload, int64_t *out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = entry != NULL ? map_entry_value(entry) : (scriptgo_value){0};
    *out_val = value.tag == SCRIPTGO_TAG_BIGINT ? (int64_t)value.payload : 0;
    return 0;
}

int scriptgo_map_get_bool(void *handle, uint32_t key_tag, uint64_t key_payload, int32_t *out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = entry != NULL ? map_entry_value(entry) : (scriptgo_value){0};
    *out_val = value.tag == SCRIPTGO_TAG_BOOLEAN ? (int32_t)value.payload : 0;
    return 0;
}

int scriptgo_map_get_ptr(void *handle, uint32_t key_tag, uint64_t key_payload, void **out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = entry != NULL ? map_entry_value(entry) : (scriptgo_value){0};
    if (entry == NULL || value.tag == SCRIPTGO_TAG_UNDEFINED) {
        *out_val = (void *)&scriptgo_undefined_sentinel;
    } else {
        *out_val = map_key_is_reference(value.tag) ? (void *)(uintptr_t)value.payload : NULL;
    }
    return 0;
}

/* scriptgo_map_get_unknown reads any value as a tagged value; an absent key
 * reads as undefined. */
int scriptgo_map_get_unknown(void *handle, uint32_t key_tag, uint64_t key_payload, scriptgo_value *out_val) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_val, "scriptgo map get: invalid arguments", &entry) != 0) return -1;
    scriptgo_value value = {0};
    if (entry != NULL) value = map_entry_value(entry);
    *out_val = value;
    return 0;
}

int scriptgo_map_has(void *handle, uint32_t key_tag, uint64_t key_payload, int32_t *out_bool) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_bool, "scriptgo map has: invalid arguments", &entry) != 0) return -1;
    *out_bool = entry != NULL;
    return 0;
}

int scriptgo_map_delete(void *handle, uint32_t key_tag, uint64_t key_payload, int32_t *out_bool) {
    scriptgo_map_native_entry *entry;
    if (map_lookup(handle, key_tag, key_payload, out_bool, "scriptgo map delete: invalid arguments", &entry) != 0) return -1;
    if (entry == NULL) {
        *out_bool = 0;
        return 0;
    }
    scriptgo_map_native *m = handle;
    map_entry_release(entry);
    int64_t idx = entry - m->entries;
    memmove(&m->entries[idx], &m->entries[idx + 1], (size_t)(m->size - idx - 1) * sizeof(scriptgo_map_native_entry));
    m->size--;
    *out_bool = 1;
    return 0;
}

int scriptgo_map_clear(void *handle) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP) return map_fail("scriptgo map clear: invalid handle");
    for (int64_t i = 0; i < m->size; i++) map_entry_release(&m->entries[i]);
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
int scriptgo_symbol_to_string(void *symbol, char **out_string);
int scriptgo_gc_get_tag(void *ptr);

static char *map_quote(const char *text) {
    if (text == NULL) text = "";
    size_t length = strlen(text);
    char *quoted = malloc(length + 3);
    if (quoted == NULL) return NULL;
    quoted[0] = '\'';
    memcpy(quoted + 1, text, length);
    quoted[length + 1] = '\'';
    quoted[length + 2] = '\0';
    return quoted;
}

static char *map_inspect_number(double value) {
    char buf[64];
    scriptgo_number_format(value, buf, sizeof(buf));
    return strdup(buf);
}

static char *map_inspect_bigint(int64_t value) {
    char buf[32];
    snprintf(buf, sizeof(buf), "%lldn", (long long)value);
    return strdup(buf);
}

/* map_inspect_reference renders a heap value as console.log does. */
static char *map_inspect_reference(void *value) {
    char *inspected = NULL;
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

/* map_inspect_tagged renders a tagged key or value as console.log does. */
static char *map_inspect_tagged(uint32_t tag, uint64_t payload) {
    char *text = NULL;
    switch (tag) {
    case SCRIPTGO_TAG_UNDEFINED:
        return strdup("undefined");
    case SCRIPTGO_TAG_NULL:
        return strdup("null");
    case SCRIPTGO_TAG_BOOLEAN:
        return strdup(payload ? "true" : "false");
    case SCRIPTGO_TAG_NUMBER:
        return map_inspect_number(map_bits_to_double(payload));
    case SCRIPTGO_TAG_STRING:
        return map_quote((const char *)(uintptr_t)payload);
    case SCRIPTGO_TAG_BIGINT:
        return map_inspect_bigint((int64_t)payload);
    case SCRIPTGO_TAG_SYMBOL:
        if (scriptgo_symbol_to_string((void *)(uintptr_t)payload, &text) == 0 && text != NULL) return text;
        return strdup("Symbol()");
    default:
        return map_inspect_reference((void *)(uintptr_t)payload);
    }
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
        scriptgo_value tagged = map_entry_value(&m->entries[i]);
        char *key = map_inspect_tagged(m->entries[i].key_tag, m->entries[i].key_payload);
        char *value = map_inspect_tagged(tagged.tag, tagged.payload);
        int failed = key == NULL || value == NULL ||
                     map_append(&buf, &length, &capacity, i == 0 ? " " : ", ") != 0 ||
                     map_append(&buf, &length, &capacity, key) != 0 ||
                     map_append(&buf, &length, &capacity, " => ") != 0 ||
                     map_append(&buf, &length, &capacity, value) != 0;
        free(key);
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
        scriptgo_boxed_value a1 = map_entry_value(entry);

        scriptgo_boxed_value a2 = {0};
        a2.tag = entry->key_tag;
        a2.payload = entry->key_payload;

        scriptgo_boxed_value a3 = {0};
        a3.tag = SCRIPTGO_TAG_OBJECT;
        a3.payload = (uint64_t)(uintptr_t)m;

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

/* map_key_copy returns a key's payload for storage outside the map: string
 * keys are copied because deleting the entry frees its text. */
static uint64_t map_key_copy(const scriptgo_map_native_entry *entry) {
    if (entry->key_tag == SCRIPTGO_TAG_STRING) {
        return (uint64_t)(uintptr_t)strdup((const char *)(uintptr_t)entry->key_payload);
    }
    return entry->key_payload;
}

/* scriptgo_map_keys lays the keys out with the element size of the array
 * type the program expects: boxed values for a mixed key type, one byte for
 * booleans, and the raw 8-byte payload (double bits, string or heap pointer,
 * bigint) otherwise. */
int scriptgo_map_keys(void *handle, int64_t element_size, void **out_array) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || out_array == NULL) return map_fail("invalid map handle");
    if (element_size != 1 && element_size != (int64_t)sizeof(scriptgo_value)) element_size = sizeof(uint64_t);
    if (scriptgo_array_new(m->size, element_size, out_array) != 0) return -1;
    scriptgo_array_header *arr = *out_array;
    for (int64_t i = 0; i < m->size; i++) {
        uint64_t payload = map_key_copy(&m->entries[i]);
        unsigned char *slot = arr->data + (size_t)i * (size_t)element_size;
        if (element_size == 1) {
            *slot = payload != 0;
        } else if (element_size == (int64_t)sizeof(scriptgo_value)) {
            scriptgo_value boxed = {0};
            boxed.tag = m->entries[i].key_tag;
            boxed.payload = payload;
            memcpy(slot, &boxed, sizeof(boxed));
        } else {
            memcpy(slot, &payload, sizeof(payload));
        }
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

int scriptgo_object_number_set(void *handle, int64_t index, double value);
int scriptgo_object_string_set(void *handle, int64_t index, const char *value);
int scriptgo_object_ptr_set(void *handle, int64_t index, void *value);
int scriptgo_object_unknown_set(void *handle, int64_t index, const scriptgo_value *value);
int scriptgo_object_new_typed(int64_t field_count, const char *type_name, void **out_object);

int scriptgo_map_entries(void *handle, void **out_array) {
    scriptgo_map_native *m = handle;
    if (m == NULL || m->magic != SCRIPTGO_MAGIC_MAP || out_array == NULL) return map_fail("invalid map handle");
    if (scriptgo_array_new(m->size, sizeof(void *), out_array) != 0) return -1;
    scriptgo_array_header *arr = *out_array;
    for (int64_t i = 0; i < m->size; i++) {
        scriptgo_map_native_entry *entry = &m->entries[i];
        void *tup = NULL;
        /* A [key, value] entry is a tuple (see the backend's objectLayoutName). */
        if (scriptgo_object_new_typed(2, "::0:1:", &tup) != 0) return -1;
        if (entry->key_tag == SCRIPTGO_TAG_STRING) {
            scriptgo_object_string_set(tup, 0, (const char *)(uintptr_t)entry->key_payload);
        } else {
            scriptgo_value key = {0};
            key.tag = entry->key_tag;
            key.payload = entry->key_payload;
            scriptgo_object_unknown_set(tup, 0, &key);
        }
        if (entry->val_type == SCRIPTGO_MAP_VAL_NUMBER) {
            scriptgo_object_number_set(tup, 1, entry->num_val);
        } else if (entry->val_type == SCRIPTGO_MAP_VAL_STRING) {
            scriptgo_object_string_set(tup, 1, entry->str_val ? entry->str_val : "");
        } else if (entry->val_type == SCRIPTGO_MAP_VAL_PTR) {
            scriptgo_object_ptr_set(tup, 1, entry->ptr_val);
        } else {
            scriptgo_value value = map_entry_value(entry);
            scriptgo_object_unknown_set(tup, 1, &value);
        }
        memcpy(arr->data + (size_t)i * sizeof(void *), &tup, sizeof(void *));
    }
    return 0;
}
