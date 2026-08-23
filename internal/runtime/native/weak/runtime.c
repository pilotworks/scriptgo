#include <stdint.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

static int weak_fail(const char *msg) { return scriptgo_runtime_set_error(msg); }

// --- WeakRef ---

typedef struct {
    void *target;
} scriptgo_weakref;

int scriptgo_weakref_new(void *target, void **out_ref) {
    if (out_ref == NULL) return weak_fail("scriptgo weakref allocation failed");
    scriptgo_weakref *ref = (scriptgo_weakref *)malloc(sizeof(scriptgo_weakref));
    if (ref == NULL) return weak_fail("scriptgo weakref allocation failed");
    ref->target = target;
    *out_ref = ref;
    return 0;
}

int scriptgo_weakref_deref(void *handle, void **out_target) {
    if (handle == NULL || out_target == NULL) return weak_fail("scriptgo weakref deref failed");
    scriptgo_weakref *ref = (scriptgo_weakref *)handle;
    *out_target = ref->target;
    return 0;
}

// --- WeakMap ---

typedef struct {
    void *key;
    uintptr_t val;
} weakmap_entry;

typedef struct {
    int64_t count;
    int64_t capacity;
    weakmap_entry *entries;
} scriptgo_weakmap;

int scriptgo_weakmap_new(void **out_map) {
    if (out_map == NULL) return weak_fail("scriptgo weakmap allocation failed");
    scriptgo_weakmap *m = (scriptgo_weakmap *)calloc(1, sizeof(scriptgo_weakmap));
    if (m == NULL) return weak_fail("scriptgo weakmap allocation failed");
    m->capacity = 8;
    m->entries = (weakmap_entry *)calloc((size_t)m->capacity, sizeof(weakmap_entry));
    if (m->entries == NULL) {
        free(m);
        return weak_fail("scriptgo weakmap allocation failed");
    }
    *out_map = m;
    return 0;
}

int scriptgo_weakmap_set(void *handle, void *key, void *value) {
    if (handle == NULL || key == NULL) return weak_fail("scriptgo weakmap set failed: key must be an object");
    scriptgo_weakmap *m = (scriptgo_weakmap *)handle;
    for (int64_t i = 0; i < m->count; i++) {
        if (m->entries[i].key == key) {
            m->entries[i].val = (uintptr_t)value;
            return 0;
        }
    }
    if (m->count >= m->capacity) {
        int64_t new_cap = m->capacity * 2;
        weakmap_entry *new_entries = (weakmap_entry *)realloc(m->entries, (size_t)new_cap * sizeof(weakmap_entry));
        if (new_entries == NULL) return weak_fail("scriptgo weakmap realloc failed");
        m->entries = new_entries;
        m->capacity = new_cap;
    }
    m->entries[m->count].key = key;
    m->entries[m->count].val = (uintptr_t)value;
    m->count++;
    return 0;
}

int scriptgo_weakmap_get(void *handle, void *key, void **out_value) {
    if (handle == NULL || out_value == NULL) return weak_fail("scriptgo weakmap get failed");
    if (key == NULL) {
        *out_value = NULL;
        return 0;
    }
    scriptgo_weakmap *m = (scriptgo_weakmap *)handle;
    for (int64_t i = 0; i < m->count; i++) {
        if (m->entries[i].key == key) {
            *out_value = (void *)m->entries[i].val;
            return 0;
        }
    }
    *out_value = NULL;
    return 0;
}

int scriptgo_weakmap_has(void *handle, void *key, int32_t *out_has) {
    if (handle == NULL || out_has == NULL) return weak_fail("scriptgo weakmap has failed");
    if (key == NULL) {
        *out_has = 0;
        return 0;
    }
    scriptgo_weakmap *m = (scriptgo_weakmap *)handle;
    for (int64_t i = 0; i < m->count; i++) {
        if (m->entries[i].key == key) {
            *out_has = 1;
            return 0;
        }
    }
    *out_has = 0;
    return 0;
}

int scriptgo_weakmap_delete(void *handle, void *key, int32_t *out_deleted) {
    if (handle == NULL || out_deleted == NULL) return weak_fail("scriptgo weakmap delete failed");
    if (key == NULL) {
        *out_deleted = 0;
        return 0;
    }
    scriptgo_weakmap *m = (scriptgo_weakmap *)handle;
    for (int64_t i = 0; i < m->count; i++) {
        if (m->entries[i].key == key) {
            m->entries[i] = m->entries[m->count - 1];
            m->count--;
            *out_deleted = 1;
            return 0;
        }
    }
    *out_deleted = 0;
    return 0;
}

// --- WeakSet ---

typedef struct {
    int64_t count;
    int64_t capacity;
    void **items;
} scriptgo_weakset;

int scriptgo_weakset_new(void **out_set) {
    if (out_set == NULL) return weak_fail("scriptgo weakset allocation failed");
    scriptgo_weakset *s = (scriptgo_weakset *)calloc(1, sizeof(scriptgo_weakset));
    if (s == NULL) return weak_fail("scriptgo weakset allocation failed");
    s->capacity = 8;
    s->items = (void **)calloc((size_t)s->capacity, sizeof(void *));
    if (s->items == NULL) {
        free(s);
        return weak_fail("scriptgo weakset allocation failed");
    }
    *out_set = s;
    return 0;
}

int scriptgo_weakset_add(void *handle, void *value) {
    if (handle == NULL || value == NULL) return weak_fail("scriptgo weakset add failed: value must be an object");
    scriptgo_weakset *s = (scriptgo_weakset *)handle;
    for (int64_t i = 0; i < s->count; i++) {
        if (s->items[i] == value) return 0;
    }
    if (s->count >= s->capacity) {
        int64_t new_cap = s->capacity * 2;
        void **new_items = (void **)realloc(s->items, (size_t)new_cap * sizeof(void *));
        if (new_items == NULL) return weak_fail("scriptgo weakset realloc failed");
        s->items = new_items;
        s->capacity = new_cap;
    }
    s->items[s->count++] = value;
    return 0;
}

int scriptgo_weakset_has(void *handle, void *value, int32_t *out_has) {
    if (handle == NULL || out_has == NULL) return weak_fail("scriptgo weakset has failed");
    if (value == NULL) {
        *out_has = 0;
        return 0;
    }
    scriptgo_weakset *s = (scriptgo_weakset *)handle;
    for (int64_t i = 0; i < s->count; i++) {
        if (s->items[i] == value) {
            *out_has = 1;
            return 0;
        }
    }
    *out_has = 0;
    return 0;
}

int scriptgo_weakset_delete(void *handle, void *value, int32_t *out_deleted) {
    if (handle == NULL || out_deleted == NULL) return weak_fail("scriptgo weakset delete failed");
    if (value == NULL) {
        *out_deleted = 0;
        return 0;
    }
    scriptgo_weakset *s = (scriptgo_weakset *)handle;
    for (int64_t i = 0; i < s->count; i++) {
        if (s->items[i] == value) {
            s->items[i] = s->items[s->count - 1];
            s->count--;
            *out_deleted = 1;
            return 0;
        }
    }
    *out_deleted = 0;
    return 0;
}
