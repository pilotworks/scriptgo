#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef enum {
    SCRIPTGO_TYPE_OBJECT = 1,
    SCRIPTGO_TYPE_ARRAY,
    SCRIPTGO_TYPE_CLOSURE,
    SCRIPTGO_TYPE_MAP,
    SCRIPTGO_TYPE_SET,
    SCRIPTGO_TYPE_BUFFER,
    SCRIPTGO_TYPE_WEAKREF,
    SCRIPTGO_TYPE_WEAKMAP,
    SCRIPTGO_TYPE_WEAKSET,
    SCRIPTGO_TYPE_ARRAYBUFFER,
    SCRIPTGO_TYPE_SYMBOL,
    SCRIPTGO_TYPE_CLOSURE_ENV
} scriptgo_gc_type_tag;

typedef struct scriptgo_gc_header {
    union {
        struct {
            uint32_t type_tag : 8;
            uint32_t gc_mark  : 2; // 0 = White, 1 = Grey, 2 = Black
            uint32_t is_weak  : 1;
            uint32_t is_root  : 1;
            uint32_t reserved : 20;
        };
        uint32_t flags;
    };
    uint32_t field_count;
} scriptgo_gc_header;

#if defined(__wasi__)
typedef int jmp_buf[16];
#define setjmp(env) (0)
#else
#include <setjmp.h>
#endif

#if defined(__APPLE__)
#include <pthread.h>
static void *get_native_stack_bottom(void) {
    return pthread_get_stackaddr_np(pthread_self());
}
#elif defined(__linux__)
#include <pthread.h>
static void *get_native_stack_bottom(void) {
    pthread_attr_t attr;
    void *stackaddr = NULL;
    size_t stacksize = 0;
    if (pthread_getattr_np(pthread_self(), &attr) == 0) {
        pthread_attr_getstack(&attr, &stackaddr, &stacksize);
        pthread_attr_destroy(&attr);
        return (void *)((uintptr_t)stackaddr + stacksize);
    }
    return NULL;
}
#else
static void *get_native_stack_bottom(void) {
    return NULL;
}
#endif

int scriptgo_runtime_set_error(const char *message);

typedef struct gc_node {
    void *ptr;
    scriptgo_gc_header header;
    struct gc_node *next;
    struct gc_node *hash_next;
} gc_node;

typedef struct root_node {
    void *ptr;
    struct root_node *next;
} root_node;

typedef struct root_slot_node {
    void *slot;
    int64_t word_count;
    struct root_slot_node *next;
} root_slot_node;

typedef struct {
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
    uint8_t type_name_owned;
    uint32_t capacity;
    void *boxed_fields;
    uintptr_t fields[];
} gc_object_layout;

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
    int64_t element_tag;
} gc_array_layout;

#define GC_HASH_INITIAL_CAPACITY 65536

static gc_node *gc_head = NULL;
static gc_node *gc_node_freelist = NULL;
static root_node *root_head = NULL;
static root_slot_node *root_slot_head = NULL;
static gc_node **gc_hash_table = NULL;
static size_t gc_hash_capacity = 0;
static size_t gc_hash_mask = 0;
static int gc_hash_dirty = 1;

static int64_t total_allocated_bytes = 0;
static int64_t total_live_objects = 0;
#define SCRIPTGO_GC_DEFAULT_THRESHOLD 262144
static int64_t gc_threshold = SCRIPTGO_GC_DEFAULT_THRESHOLD;
static int gc_in_progress = 0;
static void *scriptgo_gc_stack_bottom = NULL;
static gc_node **gc_mark_stack = NULL;
static size_t gc_mark_stack_cap = 0;

static inline size_t gc_hash_ptr(void *ptr, size_t mask) {
    uintptr_t v = ((uintptr_t)ptr >> 4) * 0x9E3779B97F4A7C15ULL;
    return (size_t)(v & mask);
}

static inline int is_possible_heap_ptr(void *ptr) {
    uintptr_t v = (uintptr_t)ptr;
    return ((v & 0x7ULL) == 0) && (v > 4096ULL) && (v < 0x0000800000000000ULL);
}

static void gc_hash_rebuild(void) {
    if (!gc_hash_dirty) return;
    size_t target_cap = GC_HASH_INITIAL_CAPACITY;
    while (target_cap < (size_t)total_live_objects * 2) {
        target_cap *= 2;
    }
    if (gc_hash_table == NULL || gc_hash_capacity != target_cap) {
        free(gc_hash_table);
        gc_hash_capacity = target_cap;
        gc_hash_mask = target_cap - 1;
        gc_hash_table = (gc_node **)calloc(gc_hash_capacity, sizeof(gc_node *));
    } else {
        memset(gc_hash_table, 0, gc_hash_capacity * sizeof(gc_node *));
    }
    if (__builtin_expect(gc_hash_table != NULL, 1)) {
        gc_node *curr = gc_head;
        while (curr != NULL) {
            size_t idx = gc_hash_ptr(curr->ptr, gc_hash_mask);
            curr->hash_next = gc_hash_table[idx];
            gc_hash_table[idx] = curr;
            curr = curr->next;
        }
    }
    gc_hash_dirty = 0;
}

static inline gc_node *find_node(void *ptr) {
    if (__builtin_expect(!is_possible_heap_ptr(ptr), 0)) return NULL;
    if (__builtin_expect(gc_hash_dirty, 0)) {
        gc_hash_rebuild();
    }
    if (__builtin_expect(gc_hash_table == NULL, 0)) return NULL;
    size_t idx = gc_hash_ptr(ptr, gc_hash_mask);
    gc_node *curr = gc_hash_table[idx];
    while (curr != NULL) {
        if (curr->ptr == ptr) {
            return curr;
        }
        curr = curr->hash_next;
    }
    return NULL;
}

void scriptgo_gc_init(void *stack_bottom) {
    void *native_bottom = get_native_stack_bottom();
    if (native_bottom != NULL) {
        scriptgo_gc_stack_bottom = native_bottom;
    } else if (stack_bottom != NULL) {
        scriptgo_gc_stack_bottom = stack_bottom;
    }
    if (gc_hash_table == NULL) {
        gc_hash_capacity = GC_HASH_INITIAL_CAPACITY;
        gc_hash_mask = GC_HASH_INITIAL_CAPACITY - 1;
        gc_hash_table = (gc_node **)calloc(gc_hash_capacity, sizeof(gc_node *));
    }
    const char *env_thresh = getenv("SCRIPTGO_GC_THRESHOLD");
    if (env_thresh != NULL) {
        long long t = atoll(env_thresh);
        if (t > 0) {
            gc_threshold = (int64_t)t;
        }
    }
}

void scriptgo_gc_set_threshold(int64_t threshold) {
    if (threshold > 0) {
        gc_threshold = threshold;
    }
}

int scriptgo_gc_collect(int64_t *out_collected_count);

static void *gc_in_flight_ptr = NULL;

static void __attribute__((noinline)) gc_trigger_collect(void *in_flight) {
    gc_in_flight_ptr = in_flight;
    scriptgo_gc_collect(NULL);
    gc_in_flight_ptr = NULL;
}

static inline __attribute__((always_inline)) int scriptgo_gc_register_fast(void *ptr, int tag, uint32_t field_count) {
    if (__builtin_expect(ptr == NULL, 0)) return 0;

    gc_node *node = gc_node_freelist;
    if (__builtin_expect(node != NULL, 1)) {
        gc_node_freelist = node->next;
    } else {
        const size_t GC_NODE_CHUNK_SIZE = 8192;
        gc_node *chunk = (gc_node *)malloc(sizeof(gc_node) * GC_NODE_CHUNK_SIZE);
        if (__builtin_expect(chunk == NULL, 0)) {
            return scriptgo_runtime_set_error("scriptgo gc node allocation failed");
        }
        for (size_t i = 1; i < GC_NODE_CHUNK_SIZE - 1; i++) {
            chunk[i].next = &chunk[i + 1];
        }
        chunk[GC_NODE_CHUNK_SIZE - 1].next = NULL;
        gc_node_freelist = &chunk[1];
        node = &chunk[0];
    }
    node->ptr = ptr;
    *(uint64_t *)&node->header = ((uint64_t)field_count << 32) | (uint32_t)(tag & 0xFF);
    node->next = gc_head;
    gc_head = node;
    total_live_objects++;
    gc_hash_dirty = 1;

    if (__builtin_expect(total_live_objects > gc_threshold && !gc_in_progress, 0)) {
        gc_trigger_collect(ptr);
    }
    return 0;
}

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count) {
    return scriptgo_gc_register_fast(ptr, tag, field_count);
}

int scriptgo_gc_is_registered(void *ptr) {
    if (ptr == NULL) return 0;
    return find_node(ptr) != NULL ? 1 : 0;
}

int scriptgo_gc_get_tag(void *ptr) {
    if (ptr == NULL) return 0;
    gc_node *node = find_node(ptr);
    return node != NULL ? (int)node->header.type_tag : 0;
}

int scriptgo_gc_unregister(void *ptr) {
    if (ptr == NULL) return 0;
    gc_node **prev = &gc_head;
    gc_node *curr = *prev;
    while (curr != NULL) {
        if (curr->ptr == ptr) {
            *prev = curr->next;
            curr->next = gc_node_freelist;
            gc_node_freelist = curr;
            total_live_objects--;
            gc_hash_dirty = 1;
            return 0;
        }
        prev = &curr->next;
        curr = *prev;
    }
    return 0;
}

int scriptgo_gc_add_root(void *ptr) {
    if (ptr == NULL) return 0;
    root_node *r = (root_node *)malloc(sizeof(root_node));
    if (r == NULL) return -1;
    r->ptr = ptr;
    r->next = root_head;
    root_head = r;
    return 0;
}

int scriptgo_gc_remove_root(void *ptr) {
    if (ptr == NULL) return 0;
    root_node **curr = &root_head;
    while (*curr != NULL) {
        if ((*curr)->ptr == ptr) {
            root_node *tmp = *curr;
            *curr = (*curr)->next;
            free(tmp);
            return 0;
        }
        curr = &(*curr)->next;
    }
    return 0;
}

int scriptgo_gc_add_root_slot(void *slot, int64_t word_count) {
    if (slot == NULL || word_count <= 0) return 0;
    root_slot_node *r = (root_slot_node *)malloc(sizeof(root_slot_node));
    if (r == NULL) return -1;
    r->slot = slot;
    r->word_count = word_count;
    r->next = root_slot_head;
    root_slot_head = r;
    return 0;
}

typedef struct {
    void *fn_ptr;
    void *env;
    void *invoke_ptr;
    int32_t return_tag;
} gc_closure_layout;

// Weak references hooks
typedef void (*weak_clean_fn)(void *weak_obj, int (*is_alive)(void *ptr));
static weak_clean_fn weak_cleaners[16];
static int weak_cleaner_count = 0;

void scriptgo_gc_register_weak_cleaner(weak_clean_fn fn) {
    if (weak_cleaner_count < 16) {
        weak_cleaners[weak_cleaner_count++] = fn;
    }
}

static int is_object_alive(void *ptr) {
    if (ptr == NULL) return 0;
    gc_node *n = find_node(ptr);
    if (n == NULL) return 1; // External/non-managed is considered alive
    return n->header.gc_mark > 0;
}

#if defined(__clang__) || defined(__GNUC__)
__attribute__((no_sanitize("address", "undefined")))
#endif
static void gc_scan_memory(void *start, void *end, gc_node **mark_stack, size_t *stack_top, size_t stack_cap) {
    uintptr_t low = (uintptr_t)start;
    uintptr_t high = (uintptr_t)end;
    if (low > high) {
        uintptr_t tmp = low; low = high; high = tmp;
    }
    low = low & ~(sizeof(void *) - 1);
    high = high & ~(sizeof(void *) - 1);

    if (high - low > 32 * 1024 * 1024) return;

    for (uintptr_t *p = (uintptr_t *)low; p < (uintptr_t *)high; p++) {
        void *val = (void *)(*p);
        if (is_possible_heap_ptr(val)) {
            gc_node *n = find_node(val);
            if (n != NULL && n->header.gc_mark == 0) {
                n->header.gc_mark = 1;
                if (*stack_top < stack_cap) {
                    mark_stack[(*stack_top)++] = n;
                }
            }
        }
    }
}

int scriptgo_gc_collect(int64_t *out_collected_count) {
    if (gc_in_progress) return 0;
    gc_in_progress = 1;
    gc_hash_rebuild();
    gc_node *curr = NULL;

    // 1. Setup Mark Stack
    size_t stack_cap = (size_t)(total_live_objects + 1024);
    if (stack_cap > gc_mark_stack_cap) {
        gc_node **new_stack = (gc_node **)realloc(gc_mark_stack, sizeof(gc_node *) * stack_cap);
        if (new_stack == NULL) {
            gc_in_progress = 0;
            return -1;
        }
        gc_mark_stack = new_stack;
        gc_mark_stack_cap = stack_cap;
    }
    gc_node **mark_stack = gc_mark_stack;
    size_t stack_top = 0;

    #define GC_PUSH(n) do { \
        gc_node *_target = (n); \
        if (_target != NULL && _target->header.gc_mark == 0) { \
            _target->header.gc_mark = 1; /* Grey */ \
            if (stack_top < stack_cap) { \
                mark_stack[stack_top++] = _target; \
            } \
        } \
    } while (0)

    // Push explicit roots
    root_node *r = root_head;
    while (r != NULL) {
        gc_node *n = find_node(r->ptr);
        GC_PUSH(n);
        r = r->next;
    }

    if (gc_in_flight_ptr != NULL) {
        gc_node *n = find_node(gc_in_flight_ptr);
        GC_PUSH(n);
    }

    // Push global root slots
    root_slot_node *rs = root_slot_head;
    while (rs != NULL) {
        if (rs->slot != NULL && rs->word_count > 0) {
            void **words = (void **)rs->slot;
            for (int64_t i = 0; i < rs->word_count; i++) {
                void *val = words[i];
                if (is_possible_heap_ptr(val)) {
                    gc_node *child = find_node(val);
                    GC_PUSH(child);
                }
            }
        }
        rs = rs->next;
    }

    // 2. Scan native stack & CPU registers
    jmp_buf cpu_regs;
    setjmp(cpu_regs);
    volatile uintptr_t stack_top_ptr = (uintptr_t)&cpu_regs;
    gc_scan_memory((void *)cpu_regs, (void *)((uintptr_t)cpu_regs + sizeof(cpu_regs)), mark_stack, &stack_top, stack_cap);

    void *bottom = scriptgo_gc_stack_bottom;
    if (bottom != NULL) {
        gc_scan_memory((void *)stack_top_ptr, bottom, mark_stack, &stack_top, stack_cap);
    }

    // 3. Mark phase (DFS tracing through pointer fields)
    while (stack_top > 0) {
        gc_node *node = mark_stack[--stack_top];
        node->header.gc_mark = 2; // Black (Reachable)

        if (node->header.type_tag == SCRIPTGO_TYPE_OBJECT) {
            gc_object_layout *obj = (gc_object_layout *)node->ptr;
            if (obj != NULL) {
                for (int64_t i = 0; i < obj->field_count; i++) {
                    void *raw = (void *)obj->fields[i];
                    if (is_possible_heap_ptr(raw)) {
                        gc_node *child = find_node(raw);
                        GC_PUSH(child);
                    }
                }
                if (__builtin_expect(obj->boxed_fields != NULL, 0)) {
                    scriptgo_value *boxed = (scriptgo_value *)obj->boxed_fields;
                    for (int64_t i = 0; i < obj->field_count; i++) {
                        void *raw = (void *)(uintptr_t)boxed[i].payload;
                        if (is_possible_heap_ptr(raw)) {
                            gc_node *child = find_node(raw);
                            GC_PUSH(child);
                        }
                    }
                }
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_ARRAY) {
            gc_array_layout *arr = (gc_array_layout *)node->ptr;
            if (arr != NULL && arr->data != NULL) {
                if (arr->element_size == (int64_t)sizeof(scriptgo_value)) {
                    for (int64_t i = 0; i < arr->length; i++) {
                        scriptgo_value *val = (scriptgo_value *)(arr->data + (size_t)i * sizeof(scriptgo_value));
                        void *raw = (void *)(uintptr_t)val->payload;
                        if (is_possible_heap_ptr(raw)) {
                            gc_node *child = find_node(raw);
                            GC_PUSH(child);
                        }
                    }
                } else if (arr->element_size == (int64_t)sizeof(void *)) {
                    for (int64_t i = 0; i < arr->length; i++) {
                        void *ptr_val = *(void **)(arr->data + (size_t)i * sizeof(void *));
                        if (is_possible_heap_ptr(ptr_val)) {
                            gc_node *child = find_node(ptr_val);
                            GC_PUSH(child);
                        }
                    }
                }
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_CLOSURE) {
            gc_closure_layout *c = (gc_closure_layout *)node->ptr;
            if (c != NULL && is_possible_heap_ptr(c->env)) {
                gc_node *child = find_node(c->env);
                GC_PUSH(child);
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_CLOSURE_ENV) {
            void **words = (void **)node->ptr;
            if (words != NULL) {
                for (uint32_t i = 0; i < node->header.field_count; i++) {
                    void *ptr_val = words[i];
                    if (is_possible_heap_ptr(ptr_val)) {
                        gc_node *child = find_node(ptr_val);
                        GC_PUSH(child);
                    }
                }
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_BUFFER) {
            typedef struct {
                uint32_t magic;
                int32_t kind;
                int64_t length;
                int64_t byte_offset;
                int64_t element_size;
                void *buffer;
                unsigned char *data;
            } gc_typedarray_layout;
            gc_typedarray_layout *ta = (gc_typedarray_layout *)node->ptr;
            if (ta != NULL && is_possible_heap_ptr(ta->buffer)) {
                gc_node *child = find_node(ta->buffer);
                GC_PUSH(child);
            }
        }
    }
    #undef GC_PUSH

    // 4. Run Weak reference cleaners
    for (int i = 0; i < weak_cleaner_count; i++) {
        gc_node *c = gc_head;
        while (c != NULL) {
            if (c->header.is_weak) {
                weak_cleaners[i](c->ptr, is_object_alive);
            }
            c = c->next;
        }
    }

    // 5. Sweep phase: Collect and free all remaining White objects (unreachable / cyclic)
    int64_t collected = 0;
    gc_node **prev = &gc_head;
    curr = *prev;
    while (curr != NULL) {
        gc_node *next = curr->next;
        if (curr->header.gc_mark == 0) {
            *prev = next;
            // Free the object payload
            if (curr->ptr != NULL) {
                if (curr->header.type_tag == SCRIPTGO_TYPE_OBJECT) {
                    gc_object_layout *o = (gc_object_layout *)curr->ptr;
                    extern void *scriptgo_object_freelist_8;
                    if (__builtin_expect(o != NULL && o->magic == 0x53474F424A454354ULL && o->capacity == 8 && o->boxed_fields == NULL && !o->type_name_owned, 1)) {
                        o->fields[0] = (uintptr_t)scriptgo_object_freelist_8;
                        scriptgo_object_freelist_8 = o;
                    } else {
                        void scriptgo_object_free(void *handle);
                        scriptgo_object_free(curr->ptr);
                    }
                } else if (curr->header.type_tag == SCRIPTGO_TYPE_CLOSURE) {
                    void scriptgo_closure_free(void *ptr);
                    scriptgo_closure_free(curr->ptr);
                } else if (curr->header.type_tag == SCRIPTGO_TYPE_ARRAY) {
                    gc_array_layout *arr = (gc_array_layout *)curr->ptr;
                    free(arr->owned_data);
                    free(arr->data);
                    free(curr->ptr);
                } else if (curr->header.type_tag == SCRIPTGO_TYPE_ARRAYBUFFER) {
                    typedef struct {
                        int64_t byte_length;
                        unsigned char *data;
                    } gc_array_buffer_layout;
                    gc_array_buffer_layout *buf = (gc_array_buffer_layout *)curr->ptr;
                    if (buf != NULL && buf->data != NULL) {
                        free(buf->data);
                    }
                    free(curr->ptr);
                } else if (curr->header.type_tag == SCRIPTGO_TYPE_SYMBOL) {
                    typedef struct { uint64_t id; char *description; } sym_payload;
                    sym_payload *sym = (sym_payload *)curr->ptr;
                    if (sym->description != NULL) {
                        free(sym->description);
                    }
                    free(curr->ptr);
                } else if (curr->header.type_tag == SCRIPTGO_TYPE_SET) {
                    void scriptgo_set_free(void *handle);
                    scriptgo_set_free(curr->ptr);
                } else {
                    free(curr->ptr);
                }
            }
            curr->next = gc_node_freelist;
            gc_node_freelist = curr;
            total_live_objects--;
            collected++;
        } else {
            // Live object: reset mark to 0 for next collection
            curr->header.gc_mark = 0;
            prev = &curr->next;
        }
        curr = next;
    }
    gc_hash_dirty = 1;

    if (total_live_objects * 2 > gc_threshold) {
        gc_threshold = total_live_objects * 2;
    }
    if (gc_threshold < SCRIPTGO_GC_DEFAULT_THRESHOLD) {
        gc_threshold = SCRIPTGO_GC_DEFAULT_THRESHOLD;
    }

    gc_in_progress = 0;
    if (out_collected_count != NULL) {
        *out_collected_count = collected;
    }
    return 0;
}

int scriptgo_gc_get_stats(int64_t *out_live_count, int64_t *out_heap_bytes) {
    if (out_live_count != NULL) {
        *out_live_count = total_live_objects;
    }
    if (out_heap_bytes != NULL) {
        *out_heap_bytes = total_allocated_bytes;
    }
    return 0;
}
