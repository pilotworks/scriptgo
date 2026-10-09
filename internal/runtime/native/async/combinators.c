/* Promise.all, Promise.allSettled, Promise.any and Promise.race.
 *
 * Each input element is read with its static element kind (input_tag: a
 * SCRIPTGO_TAG_* for typed arrays, SCRIPTGO_TAG_OBJECT for arrays of promise
 * or object references, or -1 for unknown[] whose elements carry their own
 * tags). A non-promise element is wrapped in a fulfilled promise so it
 * settles in a later microtask, as Promise.resolve(element) does. Results
 * are written in input order into the caller's tuple object (container) or
 * a new array whose element layout the caller chooses (output_size,
 * output_tag), matching the static result type. */

enum {
    SCRIPTGO_COMBINE_ALL = 0,
    SCRIPTGO_COMBINE_ALL_SETTLED = 1,
    SCRIPTGO_COMBINE_ANY = 2,
    SCRIPTGO_COMBINE_RACE = 3
};

typedef struct {
    int32_t kind;
    int done;
    int64_t remaining;
    /* An array, or with output_is_object a tuple object, in input order. */
    void *output;
    int output_is_object;
    void *errors;
    scriptgo_promise *result;
} scriptgo_aggregate_env;

int scriptgo_error_capture_stack(const char *name, const char *msg, char **out_stack);

static uint64_t scriptgo_aggregate_payload(const scriptgo_promise *p) {
    uint64_t payload = 0;
    if (p->tag == SCRIPTGO_TAG_NUMBER) {
        memcpy(&payload, &p->num_value, sizeof(payload));
    } else if (p->tag == SCRIPTGO_TAG_BOOLEAN || p->tag == SCRIPTGO_TAG_BIGINT) {
        payload = p->int_value;
    } else if (p->tag != SCRIPTGO_TAG_UNDEFINED && p->tag != SCRIPTGO_TAG_NULL) {
        payload = (uint64_t)(uintptr_t)p->ptr_value;
    }
    return payload;
}

static int scriptgo_aggregate_store(void *array_handle, int64_t index, uint32_t tag, uint64_t payload);

static void scriptgo_aggregate_value(scriptgo_value *out, uint32_t tag, uint64_t payload) {
    memset(out, 0, sizeof(*out));
    out->tag = tag;
    out->payload = payload;
    if (tag == SCRIPTGO_TAG_STRING && payload != 0) out->aux = strlen((const char *)(uintptr_t)payload);
}

/* Stores a settled value into slot index of the output: a tuple object's
 * field, or an array element in the array's layout. */
static int scriptgo_aggregate_output(scriptgo_aggregate_env *aggregate, int64_t index, uint32_t tag, uint64_t payload) {
    if (aggregate->output_is_object) {
        scriptgo_value value;
        scriptgo_aggregate_value(&value, tag, payload);
        return scriptgo_object_unknown_set(aggregate->output, index, &value);
    }
    return scriptgo_aggregate_store(aggregate->output, index, tag, payload);
}

/* Stores a settled value into slot index of an array in its element layout. */
static int scriptgo_aggregate_store(void *array_handle, int64_t index, uint32_t tag, uint64_t payload) {
    scriptgo_array_tagged_view *array = array_handle;
    unsigned char *slot = array->data + (size_t)index * (size_t)array->element_size;
    if (array->element_size == (int64_t)sizeof(scriptgo_value)) {
        scriptgo_aggregate_value((scriptgo_value *)slot, tag, payload);
    } else if (array->element_size == 1) {
        *slot = payload != 0;
    } else {
        memcpy(slot, &payload, (size_t)array->element_size < sizeof(payload) ? (size_t)array->element_size : sizeof(payload));
    }
    return 0;
}

/* A { status, value } or { status, reason } result, keyed by name like a
 * parsed JSON object so property reads and inspection see exactly those keys. */
static int scriptgo_settled_result(int rejected, uint32_t tag, uint64_t payload, void **out_object) {
    const char *type_name = rejected ? "__json__|6:status|6:reason" : "__json__|6:status|5:value";
    if (scriptgo_object_new_typed(2, type_name, out_object) != 0) return -1;
    scriptgo_value status;
    scriptgo_aggregate_value(&status, SCRIPTGO_TAG_STRING, (uint64_t)(uintptr_t)(rejected ? "rejected" : "fulfilled"));
    scriptgo_value value;
    scriptgo_aggregate_value(&value, tag, payload);
    if (scriptgo_object_unknown_set(*out_object, 0, &status) != 0) return -1;
    return scriptgo_object_unknown_set(*out_object, 1, &value);
}

/* new AggregateError(errors, "All promises were rejected") in the built-in
 * Error layout (message, name, stack, cause) followed by errors. */
static int scriptgo_aggregate_error(void *errors, void **out_object) {
    static const char message[] = "All promises were rejected";
    char *stack = NULL;
    if (scriptgo_object_new_typed(5, "__class__|c14:AggregateError|b5:Error", out_object) != 0) return -1;
    if (scriptgo_error_capture_stack("AggregateError", message, &stack) != 0) stack = NULL;
    scriptgo_object *error = *out_object;
    error->fields[0] = (uintptr_t)message;
    error->fields[1] = (uintptr_t)"AggregateError";
    error->fields[2] = (uintptr_t)(stack != NULL ? stack : "");
    error->fields[3] = (uintptr_t)"";
    error->fields[4] = (uintptr_t)errors;
    return 0;
}

static int scriptgo_aggregate_reject_all(scriptgo_aggregate_env *aggregate) {
    void *error = NULL;
    if (scriptgo_aggregate_error(aggregate->errors, &error) != 0) return -1;
    return scriptgo_promise_set_boxed(aggregate->result, 1, SCRIPTGO_TAG_OBJECT, (uint64_t)(uintptr_t)error);
}

static int scriptgo_aggregate_resolve(scriptgo_aggregate_env *aggregate) {
    uint32_t tag = aggregate->output_is_object ? SCRIPTGO_TAG_OBJECT : SCRIPTGO_TAG_ARRAY;
    return scriptgo_promise_set_boxed(aggregate->result, 0, tag, (uint64_t)(uintptr_t)aggregate->output);
}

/* Called from the event loop when input promise index settles. */
static int scriptgo_promise_aggregate_settle(void *env, int64_t index, scriptgo_promise *p) {
    scriptgo_aggregate_env *aggregate = env;
    if (aggregate == NULL || aggregate->done) return 0;
    int rejected = p->state == PROMISE_REJECTED;
    uint32_t tag = p->tag;
    uint64_t payload = scriptgo_aggregate_payload(p);
    switch (aggregate->kind) {
    case SCRIPTGO_COMBINE_ALL:
        if (rejected) {
            aggregate->done = 1;
            return scriptgo_promise_set_boxed(aggregate->result, 1, tag ? tag : SCRIPTGO_TAG_OBJECT, payload);
        }
        if (scriptgo_aggregate_output(aggregate, index, tag, payload) != 0) return -1;
        break;
    case SCRIPTGO_COMBINE_ALL_SETTLED: {
        void *settled = NULL;
        if (scriptgo_settled_result(rejected, tag, payload, &settled) != 0) return -1;
        if (scriptgo_aggregate_output(aggregate, index, SCRIPTGO_TAG_OBJECT, (uint64_t)(uintptr_t)settled) != 0) return -1;
        break;
    }
    case SCRIPTGO_COMBINE_ANY:
        if (!rejected) {
            aggregate->done = 1;
            return scriptgo_promise_set_boxed(aggregate->result, 0, tag, payload);
        }
        if (scriptgo_aggregate_store(aggregate->errors, index, tag, payload) != 0) return -1;
        break;
    default:
        aggregate->done = 1;
        return scriptgo_promise_set_boxed(aggregate->result, rejected, rejected && tag == 0 ? SCRIPTGO_TAG_OBJECT : tag, payload);
    }
    if (--aggregate->remaining > 0) return 0;
    aggregate->done = 1;
    if (aggregate->kind == SCRIPTGO_COMBINE_ANY) return scriptgo_aggregate_reject_all(aggregate);
    return scriptgo_aggregate_resolve(aggregate);
}

/* Reads input element index as the promise it settles through. */
static int scriptgo_aggregate_input(scriptgo_array_tagged_view *input, int64_t input_tag, int64_t index, scriptgo_promise **out) {
    scriptgo_value value;
    if (input_tag < 0) {
        scriptgo_array_element_value(input, index, &value);
    } else {
        uint64_t raw = 0;
        const unsigned char *slot = input->data + (size_t)index * (size_t)input->element_size;
        memcpy(&raw, slot, (size_t)input->element_size < sizeof(raw) ? (size_t)input->element_size : sizeof(raw));
        if (input_tag == SCRIPTGO_TAG_BOOLEAN && input->element_size == 1) raw &= 0xff;
        scriptgo_aggregate_value(&value, (uint32_t)input_tag, raw);
        if (input_tag == SCRIPTGO_TAG_STRING && raw == (uint64_t)(uintptr_t)&scriptgo_undefined_sentinel) value.tag = SCRIPTGO_TAG_UNDEFINED;
        if ((input_tag == SCRIPTGO_TAG_OBJECT || input_tag == SCRIPTGO_TAG_STRING) && raw == 0) value.tag = SCRIPTGO_TAG_NULL;
    }
    if (value.tag == SCRIPTGO_TAG_OBJECT) {
        scriptgo_promise *promise = scriptgo_find_promise_payload(value.payload);
        if (promise != NULL) {
            *out = promise;
            return 0;
        }
    }
    return scriptgo_promise_resolve_unknown(value.tag, value.payload, (void **)out);
}

int scriptgo_promise_combine(int32_t kind, void *input_handle, int64_t input_tag, void *container, int64_t output_size, int64_t output_tag, void **out_promise) {
    scriptgo_array_tagged_view *input = input_handle;
    if (input == NULL || out_promise == NULL || kind < SCRIPTGO_COMBINE_ALL || kind > SCRIPTGO_COMBINE_RACE) {
        return scriptgo_runtime_set_error("Promise combinator requires an array");
    }
    if (scriptgo_promise_create(out_promise) != 0) return -1;
    scriptgo_aggregate_env *aggregate = calloc(1, sizeof(*aggregate));
    if (aggregate == NULL) return scriptgo_runtime_set_error("Promise combinator allocation failed");
    aggregate->kind = kind;
    aggregate->remaining = input->length;
    aggregate->result = (scriptgo_promise *)*out_promise;
    if (container != NULL) {
        aggregate->output = container;
        aggregate->output_is_object = 1;
    } else if (kind == SCRIPTGO_COMBINE_ALL || kind == SCRIPTGO_COMBINE_ALL_SETTLED) {
        if (scriptgo_array_new_tagged(input->length, output_size, output_tag, &aggregate->output) != 0) return -1;
    } else if (kind == SCRIPTGO_COMBINE_ANY) {
        if (scriptgo_array_new(input->length, (int64_t)sizeof(scriptgo_value), &aggregate->errors) != 0) return -1;
    }
    if (input->length == 0) {
        /* race never settles on an empty input. */
        if (kind == SCRIPTGO_COMBINE_ANY) return scriptgo_aggregate_reject_all(aggregate);
        if (kind != SCRIPTGO_COMBINE_RACE) return scriptgo_aggregate_resolve(aggregate);
        return 0;
    }
    for (int64_t index = 0; index < input->length; index++) {
        scriptgo_promise *promise = NULL;
        if (scriptgo_aggregate_input(input, input_tag, index, &promise) != 0) return -1;
        scriptgo_reaction *reaction = calloc(1, sizeof(*reaction));
        if (reaction == NULL) return scriptgo_runtime_set_error("Promise combinator reaction allocation failed");
        reaction->is_aggregate = 1;
        reaction->aggregate_env = aggregate;
        reaction->aggregate_index = index;
        if (promise->reactions_tail != NULL) {
            promise->reactions_tail->next = reaction;
            promise->reactions_tail = reaction;
        } else {
            promise->reactions_head = reaction;
            promise->reactions_tail = reaction;
        }
        if (promise->state != PROMISE_PENDING) scriptgo_queue_promise_reactions(promise);
    }
    return 0;
}
