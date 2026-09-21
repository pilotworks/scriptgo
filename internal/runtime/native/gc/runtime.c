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
    uint32_t type_tag : 8;
    uint32_t gc_mark  : 2; // 0 = White, 1 = Grey, 2 = Black
    uint32_t is_weak  : 1;
    uint32_t is_root  : 1;
    uint32_t reserved : 20;
    uint32_t field_count;
    struct scriptgo_gc_header *next;
    struct scriptgo_gc_header *prev;
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
    struct gc_node *prev;
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

#define GC_HASH_INITIAL_CAPACITY 65536

static gc_node *gc_head = NULL;
static gc_node *gc_node_freelist = NULL;
static root_node *root_head = NULL;
static root_slot_node *root_slot_head = NULL;
static gc_node **gc_hash_table = NULL;
static size_t gc_hash_capacity = 0;
static size_t gc_hash_count = 0;

static int64_t total_allocated_bytes = 0;
static int64_t total_live_objects = 0;
#define SCRIPTGO_GC_DEFAULT_THRESHOLD 32768
static int64_t gc_threshold = SCRIPTGO_GC_DEFAULT_THRESHOLD;
static int gc_in_progress = 0;
static void *scriptgo_gc_stack_bottom = NULL;

static inline size_t gc_hash_ptr(void *ptr, size_t mask) {
    uintptr_t v = (uintptr_t)ptr;
    v = (v >> 3) ^ (v >> 16) ^ (v >> 24);
    return (size_t)(v & mask);
}

static gc_node *find_node(void *ptr) {
    if (ptr == NULL || gc_hash_table == NULL || gc_hash_capacity == 0) return NULL;
    size_t idx = gc_hash_ptr(ptr, gc_hash_capacity - 1);
    gc_node *curr = gc_hash_table[idx];
    while (curr != NULL) {
        if (curr->ptr == ptr) {
            return curr;
        }
        curr = curr->hash_next;
    }
    return NULL;
}

static void hash_insert(gc_node *node) {
    if (gc_hash_capacity == 0) {
        gc_hash_capacity = GC_HASH_INITIAL_CAPACITY;
        gc_hash_table = (gc_node **)calloc(gc_hash_capacity, sizeof(gc_node *));
    } else if (gc_hash_count * 2 >= gc_hash_capacity) {
        size_t new_cap = gc_hash_capacity * 2;
        gc_node **new_table = (gc_node **)calloc(new_cap, sizeof(gc_node *));
        if (new_table != NULL) {
            for (size_t i = 0; i < gc_hash_capacity; i++) {
                gc_node *c = gc_hash_table[i];
                while (c != NULL) {
                    gc_node *next = c->hash_next;
                    size_t ni = gc_hash_ptr(c->ptr, new_cap - 1);
                    c->hash_next = new_table[ni];
                    new_table[ni] = c;
                    c = next;
                }
            }
            free(gc_hash_table);
            gc_hash_table = new_table;
            gc_hash_capacity = new_cap;
        }
    }
    if (gc_hash_table != NULL) {
        size_t idx = gc_hash_ptr(node->ptr, gc_hash_capacity - 1);
        node->hash_next = gc_hash_table[idx];
        gc_hash_table[idx] = node;
        gc_hash_count++;
    }
}

static void hash_remove(void *ptr) {
    if (ptr == NULL || gc_hash_table == NULL || gc_hash_capacity == 0) return;
    size_t idx = gc_hash_ptr(ptr, gc_hash_capacity - 1);
    gc_node **curr = &gc_hash_table[idx];
    while (*curr != NULL) {
        if ((*curr)->ptr == ptr) {
            gc_node *target = *curr;
            *curr = target->hash_next;
            target->hash_next = NULL;
            if (gc_hash_count > 0) gc_hash_count--;
            return;
        }
        curr = &(*curr)->hash_next;
    }
}

void scriptgo_gc_init(void *stack_bottom) {
    void *native_bottom = get_native_stack_bottom();
    if (native_bottom != NULL) {
        scriptgo_gc_stack_bottom = native_bottom;
    } else if (stack_bottom != NULL) {
        scriptgo_gc_stack_bottom = stack_bottom;
    }
}

void scriptgo_gc_set_threshold(int64_t threshold) {
    if (threshold > 0) {
        gc_threshold = threshold;
    }
}

int scriptgo_gc_collect(int64_t *out_collected_count);

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count) {
    if (ptr == NULL) return 0;

    gc_node *node = gc_node_freelist;
    if (node != NULL) {
        gc_node_freelist = node->next;
    } else {
        node = (gc_node *)malloc(sizeof(gc_node));
        if (node == NULL) {
            return scriptgo_runtime_set_error("scriptgo gc node allocation failed");
        }
    }
    node->ptr = ptr;
    node->header.type_tag = (uint32_t)tag;
    node->header.gc_mark = 0;
    node->header.is_weak = 0;
    node->header.is_root = 0;
    node->header.reserved = 0;
    node->header.field_count = field_count;
    node->header.next = NULL;
    node->header.prev = NULL;
    node->hash_next = NULL;

    node->next = gc_head;
    node->prev = NULL;
    if (gc_head != NULL) {
        gc_head->prev = node;
    }
    gc_head = node;
    hash_insert(node);
    total_live_objects++;

    if (total_live_objects > gc_threshold && !gc_in_progress) {
        scriptgo_gc_collect(NULL);
    }
    return 0;
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
    gc_node *node = find_node(ptr);
    if (node != NULL) {
        if (node->prev != NULL) {
            node->prev->next = node->next;
        } else {
            gc_head = node->next;
        }
        if (node->next != NULL) {
            node->next->prev = node->prev;
        }
        hash_remove(ptr);
        free(node);
        total_live_objects--;
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
    uint64_t magic;
    int64_t field_count;
    const char *type_name;
    uint8_t extensible;
    uint8_t sealed;
    uint8_t frozen;
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
        if ((uintptr_t)val > 4096) {
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

    // 1. Reset all marks to 0 (White)
    gc_node *curr = gc_head;
    while (curr != NULL) {
        curr->header.gc_mark = 0;
        curr = curr->next;
    }

    // 2. Setup Mark Stack
    size_t stack_cap = (size_t)(total_live_objects + 1024);
    gc_node **mark_stack = (gc_node **)malloc(sizeof(gc_node *) * stack_cap);
    if (mark_stack == NULL) {
        gc_in_progress = 0;
        return -1;
    }
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

    // Push global root slots
    root_slot_node *s = root_slot_head;
    while (s != NULL) {
        if (s->slot != NULL) {
            void **words = (void **)s->slot;
            for (int64_t wi = 0; wi < s->word_count; wi++) {
                void *val = words[wi];
                if ((uintptr_t)val > 4096) {
                    gc_node *child = find_node(val);
                    GC_PUSH(child);
                }
            }
        }
        s = s->next;
    }

    // Conservative stack and register scanning
    jmp_buf registers;
    memset(&registers, 0, sizeof(registers));
    setjmp(registers);

    // Scan registers buffer
    gc_scan_memory(&registers, (char *)&registers + sizeof(registers), mark_stack, &stack_top, stack_cap);

    // Scan execution stack (from caller frame up to stack bottom)
    volatile void *stack_top_ptr = __builtin_frame_address(0);
    void *bottom = scriptgo_gc_stack_bottom;
    if (bottom == NULL) {
        bottom = get_native_stack_bottom();
    }
    if (bottom != NULL) {
        gc_scan_memory((void *)stack_top_ptr, bottom, mark_stack, &stack_top, stack_cap);
    }

    // 3. Mark phase (DFS tracing through pointer fields)
    while (stack_top > 0) {
        gc_node *node = mark_stack[--stack_top];
        node->header.gc_mark = 2; // Black (Reachable)

        if (node->header.type_tag == SCRIPTGO_TYPE_OBJECT) {
            gc_object_layout *obj = (gc_object_layout *)node->ptr;
            if (obj != NULL && obj->magic == 0x53474F424A454354ULL) {
                for (int64_t i = 0; i < obj->field_count; i++) {
                    uint64_t raw = (uint64_t)obj->fields[i];
                    if (raw != 0x7FF8000000000000ULL && raw > 4096) {
                        gc_node *child = find_node((void *)(uintptr_t)raw);
                        GC_PUSH(child);
                    }
                }
                if (obj->boxed_fields != NULL) {
                    scriptgo_value *boxed = (scriptgo_value *)obj->boxed_fields;
                    for (int64_t i = 0; i < obj->field_count; i++) {
                        if (boxed[i].payload > 4096) {
                            gc_node *child = find_node((void *)(uintptr_t)boxed[i].payload);
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
                        if (val->payload > 4096) {
                            gc_node *child = find_node((void *)(uintptr_t)val->payload);
                            GC_PUSH(child);
                        }
                    }
                } else if (arr->element_size == (int64_t)sizeof(void *)) {
                    for (int64_t i = 0; i < arr->length; i++) {
                        void *ptr_val = *(void **)(arr->data + (size_t)i * sizeof(void *));
                        if ((uintptr_t)ptr_val > 4096) {
                            gc_node *child = find_node(ptr_val);
                            GC_PUSH(child);
                        }
                    }
                }
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_CLOSURE) {
            gc_closure_layout *c = (gc_closure_layout *)node->ptr;
            if (c != NULL && c->env != NULL && (uintptr_t)c->env > 4096) {
                gc_node *child = find_node(c->env);
                GC_PUSH(child);
            }
        } else if (node->header.type_tag == SCRIPTGO_TYPE_CLOSURE_ENV) {
            void **words = (void **)node->ptr;
            if (words != NULL) {
                for (uint32_t i = 0; i < node->header.field_count; i++) {
                    void *ptr_val = words[i];
                    if ((uintptr_t)ptr_val > 4096) {
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
            if (ta != NULL && ta->buffer != NULL && (uintptr_t)ta->buffer > 4096) {
                gc_node *child = find_node(ta->buffer);
                GC_PUSH(child);
            }
        }
    }
    free(mark_stack);
    #undef GC_PUSH

    // 4. Run Weak reference cleaners
    for (int i = 0; i < weak_cleaner_count; i++) {
        curr = gc_head;
        while (curr != NULL) {
            if (curr->header.is_weak) {
                weak_cleaners[i](curr->ptr, is_object_alive);
            }
            curr = curr->next;
        }
    }

    // 5. Sweep phase: Collect and free all remaining White objects (unreachable / cyclic)
    int64_t collected = 0;
    curr = gc_head;
    while (curr != NULL) {
        gc_node *next = curr->next;
        if (curr->header.gc_mark == 0) {
            // Unlink from all-nodes list
            if (curr->prev != NULL) {
                curr->prev->next = curr->next;
            } else {
                gc_head = curr->next;
            }
            if (curr->next != NULL) {
                curr->next->prev = curr->prev;
            }
            // Remove from hash table
            hash_remove(curr->ptr);

            // Free the object payload
            if (curr->ptr != NULL) {
                if (curr->header.type_tag == SCRIPTGO_TYPE_OBJECT) {
                    gc_object_layout *object = (gc_object_layout *)curr->ptr;
                    if (object->magic == 0x53474F424A454354ULL) {
                        free((void *)object->type_name);
                        if (object->boxed_fields != NULL) {
                            free(object->boxed_fields);
                        }
                    }
                    free(curr->ptr);
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
                } else {
                    free(curr->ptr);
                }
            }
            curr->next = gc_node_freelist;
            gc_node_freelist = curr;
            total_live_objects--;
            collected++;
        }
        curr = next;
    }

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
