#include "scriptgo_value.h"

#include <stdlib.h>
#include <string.h>
#include <limits.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct scriptgo_engine_ref {
    uint64_t magic;
    uint32_t refs;
    uint32_t tag;
    scriptgo_dynamic_context *context;
    uint64_t handle;
    scriptgo_engine_retain_fn retain;
    scriptgo_engine_release_fn release;
    struct scriptgo_engine_ref *next;
} scriptgo_engine_ref;

struct scriptgo_dynamic_context {
    uint64_t magic;
    uint32_t live_refs;
    uint8_t shutting_down;
    void *engine;
    void *user_data;
    scriptgo_engine_retain_fn retain_ref;
    scriptgo_engine_release_fn release_ref;
    scriptgo_dynamic_call_fn call;
};

#define SCRIPTGO_ENGINE_REF_MAGIC 0x5347524546455231ULL
#define SCRIPTGO_CONTEXT_MAGIC 0x534743545856315FULL

static scriptgo_engine_ref *scriptgo_engine_refs = NULL;

static scriptgo_engine_ref *find_engine_ref(uint64_t payload) {
    scriptgo_engine_ref *ref = scriptgo_engine_refs;
    while (ref != NULL) {
        if ((uint64_t)(uintptr_t)ref == payload) return ref;
        ref = ref->next;
    }
    return NULL;
}

static void link_engine_ref(scriptgo_engine_ref *ref) {
    ref->next = scriptgo_engine_refs;
    scriptgo_engine_refs = ref;
}

static void unlink_engine_ref(scriptgo_engine_ref *ref) {
    scriptgo_engine_ref **cursor = &scriptgo_engine_refs;
    while (*cursor != NULL) {
        if (*cursor == ref) {
            *cursor = ref->next;
            ref->next = NULL;
            return;
        }
        cursor = &(*cursor)->next;
    }
}

static int value_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

void scriptgo_value_init_undefined(scriptgo_value *value) {
    if (value == NULL) return;
    value->tag = SCRIPTGO_TAG_UNDEFINED;
    value->flags = 0;
    value->payload = 0;
    value->aux = 0;
}

static int valid_tag(uint32_t tag) {
    return tag <= SCRIPTGO_TAG_SYMBOL;
}

int32_t scriptgo_value_validate(const scriptgo_value *value) {
    if (value == NULL) return value_fail("scriptgo value is null");
    if (!valid_tag(value->tag)) return value_fail("SG9001: invalid scriptgo value tag");
    if ((value->flags & ~(SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF)) != 0) {
        return value_fail("SG9001: invalid scriptgo value flags");
    }
    if ((value->flags & SCRIPTGO_VALUE_ENGINE_REF) != 0 &&
        (value->flags & SCRIPTGO_VALUE_OWNED) == 0) {
        return value_fail("SG9001: engine value must be owned");
    }
    switch (value->tag) {
    case SCRIPTGO_TAG_UNDEFINED:
    case SCRIPTGO_TAG_NULL:
        if (value->flags != 0 || value->payload != 0 || value->aux != 0) {
            return value_fail("SG9001: invalid nullish scriptgo value");
        }
        break;
    case SCRIPTGO_TAG_BOOLEAN:
        if (value->payload > 1 || value->aux != 0 || value->flags != 0) {
            return value_fail("SG9001: invalid boolean scriptgo value");
        }
        break;
    case SCRIPTGO_TAG_NUMBER:
        if (value->aux != 0 || value->flags != 0) {
            return value_fail("SG9001: invalid number scriptgo value");
        }
        break;
    case SCRIPTGO_TAG_STRING:
        if (value->payload == 0 && value->aux != 0) {
            return value_fail("SG9001: string length has no data");
        }
        if ((value->flags & SCRIPTGO_VALUE_ENGINE_REF) != 0) {
            return value_fail("SG9001: string cannot be an engine reference");
        }
        if ((value->flags & ~SCRIPTGO_VALUE_OWNED) != 0) {
            return value_fail("SG9001: invalid string ownership flags");
        }
        break;
    default:
        if ((value->flags & SCRIPTGO_VALUE_OWNED) != 0 &&
            (value->flags & SCRIPTGO_VALUE_ENGINE_REF) == 0) {
            return value_fail("SG9001: native references cannot be owned");
        }
        if ((value->flags & SCRIPTGO_VALUE_ENGINE_REF) != 0 && value->payload == 0) {
            return value_fail("SG9001: engine reference has no handle");
        }
        if ((value->flags & SCRIPTGO_VALUE_ENGINE_REF) == 0 && value->aux != 0) {
            return value_fail("SG9001: native reference has auxiliary data");
        }
        break;
    }
    return 0;
}

static int is_owned_string(const scriptgo_value *value) {
    return value != NULL && value->tag == SCRIPTGO_TAG_STRING &&
           (value->flags & SCRIPTGO_VALUE_OWNED) != 0;
}

static int value_shape_is_initialized(const scriptgo_value *value) {
    if (value == NULL || !valid_tag(value->tag) ||
        (value->flags & ~(SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF)) != 0) return 0;
    if (value->tag <= SCRIPTGO_TAG_NUMBER) {
        if (value->flags != 0) return 0;
        if (value->tag <= SCRIPTGO_TAG_BOOLEAN) return value->payload == 0 && value->aux == 0;
        return value->aux == 0;
    }
    if (value->tag == SCRIPTGO_TAG_STRING) {
        return (value->flags & SCRIPTGO_VALUE_ENGINE_REF) == 0 &&
               (value->payload != 0 || value->aux == 0);
    }
    return (value->flags & SCRIPTGO_VALUE_ENGINE_REF) == 0
        ? value->aux == 0 : value->payload != 0;
}

static void prepare_destination(scriptgo_value *value) {
    if (value_shape_is_initialized(value)) scriptgo_value_release(value);
    else scriptgo_value_init_undefined(value);
}

int32_t scriptgo_value_clone(scriptgo_value *out, const scriptgo_value *source) {
    if (out == NULL || source == NULL || out == source) return value_fail("SG9001: invalid value clone");
    if (scriptgo_value_validate(source) != 0) return -1;
    prepare_destination(out);
    if (is_owned_string(source)) {
        return scriptgo_value_string_copy((const void *)(uintptr_t)source->payload,
                                          source->aux, out);
    }
    if ((source->flags & SCRIPTGO_VALUE_ENGINE_REF) != 0) {
        scriptgo_engine_ref *ref = find_engine_ref(source->payload);
        if (ref == NULL || ref->magic != SCRIPTGO_ENGINE_REF_MAGIC || ref->refs == 0) {
            return value_fail("SG9001: invalid engine reference");
        }
        if (ref->retain != NULL) ref->retain(ref->context, ref->handle);
        ref->refs++;
        if (ref->context != NULL) ref->context->live_refs++;
    }
    *out = *source;
    return 0;
}

int32_t scriptgo_value_move(scriptgo_value *out, scriptgo_value *source) {
    if (out == NULL || source == NULL || out == source) return value_fail("SG9001: invalid value move");
    if (scriptgo_value_validate(source) != 0) return -1;
    prepare_destination(out);
    *out = *source;
    scriptgo_value_init_undefined(source);
    return 0;
}

int32_t scriptgo_value_release(scriptgo_value *value) {
    if (value == NULL) return 0;
    if (value->flags & SCRIPTGO_VALUE_ENGINE_REF) {
        scriptgo_engine_ref *ref = find_engine_ref(value->payload);
        if (ref != NULL && ref->magic == SCRIPTGO_ENGINE_REF_MAGIC && ref->refs > 0) {
            ref->refs--;
            if (ref->context != NULL && ref->context->live_refs > 0) ref->context->live_refs--;
            if (ref->release != NULL) ref->release(ref->context, ref->handle);
            if (ref->refs == 0) {
                unlink_engine_ref(ref);
                ref->magic = 0;
                free(ref);
            }
        }
    } else if (is_owned_string(value) && value->payload != 0) {
        free((void *)(uintptr_t)value->payload);
    }
    scriptgo_value_init_undefined(value);
    return 0;
}

int32_t scriptgo_value_string_borrow(const void *bytes, uint64_t length, scriptgo_value *out) {
    if (out == NULL || (bytes == NULL && length != 0)) return value_fail("SG9001: invalid borrowed string");
    prepare_destination(out);
    out->tag = SCRIPTGO_TAG_STRING;
    out->flags = 0;
    out->payload = (uint64_t)(uintptr_t)bytes;
    out->aux = length;
    return 0;
}

int32_t scriptgo_value_string_copy(const void *bytes, uint64_t length, scriptgo_value *out) {
    void *copy;
    if (out == NULL || (bytes == NULL && length != 0)) return value_fail("SG9001: invalid string copy");
    if (length == UINT64_MAX || length > (uint64_t)SIZE_MAX - 1) {
        return value_fail("SG9001: string is too large");
    }
    if (out == NULL) return value_fail("SG9001: invalid string copy");
    prepare_destination(out);
    copy = malloc((size_t)length + 1);
    if (copy == NULL) return value_fail("scriptgo value string allocation failed");
    if (length != 0) memcpy(copy, bytes, (size_t)length);
    ((unsigned char *)copy)[length] = 0;
    out->tag = SCRIPTGO_TAG_STRING;
    out->flags = SCRIPTGO_VALUE_OWNED;
    out->payload = (uint64_t)(uintptr_t)copy;
    out->aux = length;
    return 0;
}

int32_t scriptgo_value_adopt_engine_ref(scriptgo_dynamic_context *context,
                                        uint32_t semantic_tag,
                                        uint64_t handle,
                                        scriptgo_value *out) {
    scriptgo_engine_ref *ref;
    if (out == NULL || context == NULL || handle == 0 || semantic_tag < SCRIPTGO_TAG_OBJECT ||
        !valid_tag(semantic_tag)) {
        return value_fail("SG9001: invalid engine reference");
    }
    if (context->magic != SCRIPTGO_CONTEXT_MAGIC || context->shutting_down) {
        return value_fail("SG9001: dynamic context is unavailable");
    }
    prepare_destination(out);
    ref = (scriptgo_engine_ref *)calloc(1, sizeof(*ref));
    if (ref == NULL) return value_fail("scriptgo engine reference allocation failed");
    ref->magic = SCRIPTGO_ENGINE_REF_MAGIC;
    ref->refs = 1;
    ref->tag = semantic_tag;
    ref->context = context;
    ref->handle = handle;
    ref->retain = context->retain_ref;
    ref->release = context->release_ref;
    link_engine_ref(ref);
    context->live_refs++;
    out->tag = semantic_tag;
    out->flags = SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF;
    out->payload = (uint64_t)(uintptr_t)ref;
    out->aux = 0;
    return 0;
}

scriptgo_dynamic_context *scriptgo_dynamic_context_new(
    void *engine,
    void *user_data,
    scriptgo_engine_retain_fn retain_ref,
    scriptgo_engine_release_fn release_ref,
    scriptgo_dynamic_call_fn call) {
    scriptgo_dynamic_context *context = calloc(1, sizeof(*context));
    if (context == NULL) {
        value_fail("SG9001: dynamic context allocation failed");
        return NULL;
    }
    context->magic = SCRIPTGO_CONTEXT_MAGIC;
    context->engine = engine;
    context->user_data = user_data;
    context->retain_ref = retain_ref;
    context->release_ref = release_ref;
    context->call = call;
    return context;
}

uint32_t scriptgo_dynamic_context_live_refs(const scriptgo_dynamic_context *context) {
    if (context == NULL || context->magic != SCRIPTGO_CONTEXT_MAGIC) return 0;
    return context->live_refs;
}

int32_t scriptgo_dynamic_context_shutdown(scriptgo_dynamic_context *context) {
    if (context == NULL || context->magic != SCRIPTGO_CONTEXT_MAGIC) {
        return value_fail("SG9001: invalid dynamic context");
    }
    if (context->live_refs != 0) {
        return value_fail("SG9001: dynamic context has outstanding engine references");
    }
    context->shutting_down = 1;
    return 0;
}

int32_t scriptgo_dynamic_context_destroy(scriptgo_dynamic_context *context) {
    if (context == NULL) return 0;
    if (context->magic != SCRIPTGO_CONTEXT_MAGIC || context->live_refs != 0 || !context->shutting_down) {
        return value_fail("SG9001: dynamic context is not ready to destroy");
    }
    context->magic = 0;
    free(context);
    return 0;
}

static uint64_t tag_bit(uint32_t tag) {
    return tag < 64 ? (UINT64_C(1) << tag) : 0;
}

static int is_boundary_value(const scriptgo_value *value) {
    if (value == NULL || scriptgo_value_validate(value) != 0) return 0;
    if (value->tag > SCRIPTGO_TAG_STRING) return 0;
    return value->flags == 0 || (value->tag == SCRIPTGO_TAG_STRING &&
                                 value->flags == SCRIPTGO_VALUE_OWNED);
}

static int boundary_matches(const scriptgo_value *value, scriptgo_value_constraint constraint) {
    return is_boundary_value(value) && (constraint.allowed_tags & tag_bit(value->tag)) != 0;
}

static int valid_constraint(scriptgo_value_constraint constraint) {
    return constraint.allowed_tags != 0 &&
           (constraint.allowed_tags & ~((UINT64_C(1) << 10) - 1)) == 0;
}

static int boundary_type_error(scriptgo_value *out_exception, const char *message) {
    scriptgo_value_init_undefined(out_exception);
    return scriptgo_value_string_copy(message, strlen(message), out_exception) == 0
        ? SCRIPTGO_CALL_THROWN : SCRIPTGO_CALL_FATAL;
}

int32_t scriptgo_dynamic_call(scriptgo_dynamic_context *context,
                              const scriptgo_value *callable,
                              const scriptgo_value *this_value,
                              const scriptgo_value *arguments,
                              uint32_t argument_count,
                              const scriptgo_boundary_descriptor *descriptor,
                              scriptgo_value *out_result,
                              scriptgo_value *out_exception) {
    scriptgo_engine_ref *ref;
    uint32_t i;
    int32_t status;
    if (out_result == NULL || out_exception == NULL || out_result == out_exception) {
        return value_fail("SG9001: malformed dynamic call outputs");
    }
    scriptgo_value_init_undefined(out_result);
    scriptgo_value_init_undefined(out_exception);
    if (context == NULL || context->magic != SCRIPTGO_CONTEXT_MAGIC || context->shutting_down ||
        context->call == NULL || callable == NULL || this_value == NULL || descriptor == NULL ||
        (argument_count != 0 && arguments == NULL) || descriptor->format != SCRIPTGO_BOUNDARY_FORMAT_V1 ||
        (descriptor->parameter_count != 0 && descriptor->parameters == NULL) ||
        descriptor->source_path == NULL || descriptor->name == NULL ||
        descriptor->source_length == 0 || !valid_constraint(descriptor->this_value) ||
        !valid_constraint(descriptor->result)) {
        return value_fail("SG9001: malformed dynamic call descriptor");
    }
    if (scriptgo_value_validate(callable) != 0 ||
        (callable->tag != SCRIPTGO_TAG_FUNCTION) ||
        (callable->flags != (SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF))) {
        return value_fail("SG9001: dynamic callable is not an owned engine reference");
    }
    ref = find_engine_ref(callable->payload);
    if (ref == NULL || ref->magic != SCRIPTGO_ENGINE_REF_MAGIC || ref->context != context ||
        ref->tag != SCRIPTGO_TAG_FUNCTION || ref->refs == 0) {
        return value_fail("SG9001: dynamic callable belongs to another context");
    }
    if (scriptgo_value_validate(this_value) != 0) {
        return value_fail("SG9001: malformed Dynamic this value");
    }
    if (descriptor->parameter_count != argument_count) {
        return boundary_type_error(out_exception, "TypeError: SG5002: Dynamic call arity mismatch");
    }
    if (!boundary_matches(this_value, descriptor->this_value)) {
        return boundary_type_error(out_exception, "TypeError: SG5002: invalid Dynamic this value");
    }
    for (i = 0; i < argument_count; i++) {
        if (scriptgo_value_validate(&arguments[i]) != 0) {
            return value_fail("SG9001: malformed Dynamic argument value");
        }
        if (!boundary_matches(&arguments[i], descriptor->parameters[i])) {
            return boundary_type_error(out_exception, "TypeError: SG5002: invalid Dynamic argument");
        }
    }
    status = context->call(context, ref->handle, this_value, arguments, argument_count,
                           descriptor, out_result, out_exception, context->user_data);
    if (status != SCRIPTGO_CALL_OK && status != SCRIPTGO_CALL_THROWN && status != SCRIPTGO_CALL_FATAL) {
        status = SCRIPTGO_CALL_FATAL;
    }
    if (status == SCRIPTGO_CALL_FATAL) {
        scriptgo_value_release(out_result);
        scriptgo_value_release(out_exception);
        return value_fail("SG9001: Dynamic adapter returned an invalid call status");
    }
    if (status == SCRIPTGO_CALL_THROWN) {
        if (scriptgo_value_validate(out_result) != 0 ||
            scriptgo_value_validate(out_exception) != 0 || out_exception->tag == SCRIPTGO_TAG_UNDEFINED ||
            out_result->tag != SCRIPTGO_TAG_UNDEFINED) {
            scriptgo_value_release(out_result);
            scriptgo_value_release(out_exception);
            return value_fail("SG9001: malformed Dynamic exception output");
        }
        if ((out_exception->flags & SCRIPTGO_VALUE_ENGINE_REF) != 0) {
            scriptgo_engine_ref *exception_ref = find_engine_ref(out_exception->payload);
            if (exception_ref == NULL || exception_ref->context != context) {
                scriptgo_value_release(out_exception);
                return value_fail("SG9001: Dynamic exception belongs to another context");
            }
        }
        return SCRIPTGO_CALL_THROWN;
    }
    if (scriptgo_value_validate(out_result) != 0 ||
        scriptgo_value_validate(out_exception) != 0 || out_exception->tag != SCRIPTGO_TAG_UNDEFINED) {
        scriptgo_value_release(out_result);
        scriptgo_value_release(out_exception);
        return value_fail("SG9001: malformed Dynamic result output");
    }
    if (!boundary_matches(out_result, descriptor->result)) {
        scriptgo_value_release(out_result);
        return boundary_type_error(out_exception, "TypeError: SG5003: invalid Dynamic result");
    }
    return SCRIPTGO_CALL_OK;
}
