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
    SCRIPTGO_TYPE_WEAKSET
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

int scriptgo_runtime_set_error(const char *message);

typedef struct gc_node {
    void *ptr;
    scriptgo_gc_header header;
    struct gc_node *next;
    struct gc_node *prev;
} gc_node;

typedef struct root_node {
    void *ptr;
    struct root_node *next;
} root_node;

static gc_node *gc_head = NULL;
static root_node *root_head = NULL;
static int64_t total_allocated_bytes = 0;
static int64_t total_live_objects = 0;

void scriptgo_gc_init(void) {
    // Initialized on demand
}

int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count) {
    if (ptr == NULL) return 0;
    gc_node *node = (gc_node *)malloc(sizeof(gc_node));
    if (node == NULL) {
        return scriptgo_runtime_set_error("scriptgo gc node allocation failed");
    }
    node->ptr = ptr;
    node->header.type_tag = (uint32_t)tag;
    node->header.gc_mark = 0;
    node->header.is_weak = 0;
    node->header.is_root = 0;
    node->header.field_count = field_count;
    node->header.next = NULL;
    node->header.prev = NULL;

    node->next = gc_head;
    node->prev = NULL;
    if (gc_head != NULL) {
        gc_head->prev = node;
    }
    gc_head = node;
    total_live_objects++;
    return 0;
}

static gc_node *find_node(void *ptr) {
    gc_node *curr = gc_head;
    while (curr != NULL) {
        if (curr->ptr == ptr) {
            return curr;
        }
        curr = curr->next;
    }
    return NULL;
}

int scriptgo_gc_is_registered(void *ptr) {
    if (ptr == NULL) return 0;
    return find_node(ptr) != NULL ? 1 : 0;
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

typedef struct {
    int64_t field_count;
    const char *type_name;
    uintptr_t fields[];
} gc_object_layout;

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

int scriptgo_gc_collect(int64_t *out_collected_count) {
    // 1. Reset all marks to 0 (White)
    gc_node *curr = gc_head;
    while (curr != NULL) {
        curr->header.gc_mark = 0;
        curr = curr->next;
    }

    // 2. Setup Mark Stack
    gc_node **mark_stack = (gc_node **)malloc(sizeof(gc_node *) * (size_t)(total_live_objects + 16));
    if (mark_stack == NULL) return -1;
    size_t stack_top = 0;

    // Push roots
    root_node *r = root_head;
    while (r != NULL) {
        gc_node *n = find_node(r->ptr);
        if (n != NULL && n->header.gc_mark == 0) {
            n->header.gc_mark = 1; // Grey
            mark_stack[stack_top++] = n;
        }
        r = r->next;
    }

    // 3. Mark phase (DFS tracing through pointer fields)
    while (stack_top > 0) {
        gc_node *node = mark_stack[--stack_top];
        node->header.gc_mark = 2; // Black (Reachable)

        if (node->header.type_tag == SCRIPTGO_TYPE_OBJECT) {
            gc_object_layout *obj = (gc_object_layout *)node->ptr;
            if (obj != NULL) {
                for (int64_t i = 0; i < obj->field_count; i++) {
                    void *child_ptr = (void *)obj->fields[i];
                    if (child_ptr != NULL) {
                        gc_node *child_node = find_node(child_ptr);
                        if (child_node != NULL && child_node->header.gc_mark == 0) {
                            child_node->header.gc_mark = 1; // Grey
                            mark_stack[stack_top++] = child_node;
                        }
                    }
                }
            }
        }
    }
    free(mark_stack);

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
            // Unlink
            if (curr->prev != NULL) {
                curr->prev->next = curr->next;
            } else {
                gc_head = curr->next;
            }
            if (curr->next != NULL) {
                curr->next->prev = curr->prev;
            }
            // Free the object payload
            if (curr->ptr != NULL) {
                free(curr->ptr);
            }
            free(curr);
            total_live_objects--;
            collected++;
        }
        curr = next;
    }

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
