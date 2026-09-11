#ifndef SCRIPTGO_VALUE_H
#define SCRIPTGO_VALUE_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    SCRIPTGO_TAG_UNDEFINED = 0,
    SCRIPTGO_TAG_NULL      = 1,
    SCRIPTGO_TAG_BOOLEAN   = 2,
    SCRIPTGO_TAG_NUMBER    = 3,
    SCRIPTGO_TAG_STRING    = 4,
    SCRIPTGO_TAG_OBJECT    = 5,
    SCRIPTGO_TAG_ARRAY     = 6,
    SCRIPTGO_TAG_FUNCTION  = 7,
    SCRIPTGO_TAG_BIGINT    = 8,
    SCRIPTGO_TAG_SYMBOL    = 9
} scriptgo_value_tag;

enum {
    SCRIPTGO_VALUE_OWNED     = 1u << 0,
    SCRIPTGO_VALUE_ENGINE_REF = 1u << 1
};

typedef struct {
    uint32_t tag;
    uint32_t flags;
    uint64_t payload;
    uint64_t aux;
} scriptgo_value;

#if defined(__cplusplus)
static_assert(sizeof(scriptgo_value) == 24, "scriptgo_value ABI v1 size");
static_assert(alignof(scriptgo_value) == 8, "scriptgo_value ABI v1 alignment");
#else
_Static_assert(sizeof(scriptgo_value) == 24, "scriptgo_value ABI v1 size");
_Static_assert(_Alignof(scriptgo_value) == 8, "scriptgo_value ABI v1 alignment");
#endif

/* Kept as a source-compatible name while existing runtime families migrate. */
typedef scriptgo_value scriptgo_boxed_value;

typedef struct scriptgo_dynamic_context scriptgo_dynamic_context;
typedef struct scriptgo_exception_frame scriptgo_exception_frame_t;

typedef void (*scriptgo_engine_retain_fn)(scriptgo_dynamic_context *context,
                                          uint64_t handle);
typedef void (*scriptgo_engine_release_fn)(scriptgo_dynamic_context *context,
                                           uint64_t handle);

typedef struct {
    uint64_t allowed_tags;
} scriptgo_value_constraint;

typedef struct {
    uint32_t format;
    uint32_t parameter_count;
    const scriptgo_value_constraint *parameters;
    scriptgo_value_constraint this_value;
    scriptgo_value_constraint result;
    const char *name;
    const char *source_path;
    uint32_t source_start;
    uint32_t source_length;
} scriptgo_boundary_descriptor;

const char *scriptgo_runtime_last_error(void);

typedef int32_t (*scriptgo_dynamic_call_fn)(
    scriptgo_dynamic_context *context,
    uint64_t callable_handle,
    const scriptgo_value *this_value,
    const scriptgo_value *arguments,
    uint32_t argument_count,
    const scriptgo_boundary_descriptor *descriptor,
    scriptgo_value *out_result,
    scriptgo_value *out_exception,
    void *user_data);

enum {
    SCRIPTGO_BOUNDARY_FORMAT_V1 = 1
};

enum {
    SCRIPTGO_CALL_OK = 0,
    SCRIPTGO_CALL_THROWN = 1,
    SCRIPTGO_CALL_FATAL = -1
};

void scriptgo_value_init_undefined(scriptgo_value *value);
int32_t scriptgo_value_validate(const scriptgo_value *value);
int32_t scriptgo_value_clone(scriptgo_value *out, const scriptgo_value *source);
int32_t scriptgo_value_move(scriptgo_value *out, scriptgo_value *source);
int32_t scriptgo_value_release(scriptgo_value *value);
int32_t scriptgo_value_string_borrow(const void *bytes, uint64_t length, scriptgo_value *out);
int32_t scriptgo_value_string_copy(const void *bytes, uint64_t length, scriptgo_value *out);
int32_t scriptgo_value_adopt_engine_ref(scriptgo_dynamic_context *context,
                                        uint32_t semantic_tag,
                                        uint64_t handle,
                                        scriptgo_value *out);

/* Canonical Promise boundary operations. Typed helpers remain internal fast paths. */
int32_t scriptgo_promise_resolve_value(void *promise_handle, const scriptgo_value *value);
int32_t scriptgo_promise_reject_value(void *promise_handle, const scriptgo_value *value);
int32_t scriptgo_promise_resolve_existing_value(void *promise_handle, const scriptgo_value *value);
int32_t scriptgo_promise_reject_existing_value(void *promise_handle, const scriptgo_value *value);
int32_t scriptgo_promise_resolve_unknown_value(const scriptgo_value *value, void **out_promise);
int32_t scriptgo_promise_await_value(void *promise_handle, scriptgo_value *out_value);
int32_t scriptgo_promise_await_unknown_value(const scriptgo_value *value, scriptgo_value *out_value);

scriptgo_dynamic_context *scriptgo_dynamic_context_new(
    void *engine,
    void *user_data,
    scriptgo_engine_retain_fn retain_ref,
    scriptgo_engine_release_fn release_ref,
    scriptgo_dynamic_call_fn call);
int32_t scriptgo_dynamic_context_shutdown(scriptgo_dynamic_context *context);
int32_t scriptgo_dynamic_context_destroy(scriptgo_dynamic_context *context);
uint32_t scriptgo_dynamic_context_live_refs(const scriptgo_dynamic_context *context);
/* Releases engine values still owned by the embedding at process teardown. */
void scriptgo_dynamic_context_release_all(scriptgo_dynamic_context *context);
int32_t scriptgo_dynamic_call(scriptgo_dynamic_context *context,
                              const scriptgo_value *callable,
                              const scriptgo_value *this_value,
                              const scriptgo_value *arguments,
                              uint32_t argument_count,
                              const scriptgo_boundary_descriptor *descriptor,
                              scriptgo_value *out_result,
                              scriptgo_value *out_exception);

scriptgo_exception_frame_t *scriptgo_exception_frame_new(void);
void *scriptgo_exception_buf(scriptgo_exception_frame_t *frame);
void scriptgo_exception_frame_free(scriptgo_exception_frame_t *frame);
void scriptgo_exception_throw_copy(const scriptgo_value *value);
void scriptgo_exception_throw_move(scriptgo_value *value);
void scriptgo_exception_take_value(scriptgo_exception_frame_t *frame, scriptgo_value *out);
void scriptgo_exception_rethrow(scriptgo_exception_frame_t *frame);

#ifdef __cplusplus
}
#endif

#endif
