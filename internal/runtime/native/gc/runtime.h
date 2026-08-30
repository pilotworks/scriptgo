#ifndef SCRIPTGO_GC_RUNTIME_H
#define SCRIPTGO_GC_RUNTIME_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

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
    SCRIPTGO_TYPE_ARRAYBUFFER
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

void scriptgo_gc_init(void);
int scriptgo_gc_register(void *ptr, scriptgo_gc_type_tag tag, uint32_t field_count);
int scriptgo_gc_is_registered(void *ptr);
int scriptgo_gc_unregister(void *ptr);
int scriptgo_gc_add_root(void *ptr);
int scriptgo_gc_remove_root(void *ptr);
int scriptgo_gc_get_tag(void *ptr);
int scriptgo_gc_collect(int64_t *out_collected_count);
int scriptgo_gc_get_stats(int64_t *out_live_count, int64_t *out_heap_bytes);

#ifdef __cplusplus
}
#endif

#endif /* SCRIPTGO_GC_RUNTIME_H */
