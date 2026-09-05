#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#if !defined(__wasi__)
#include <setjmp.h>
#endif

int scriptgo_runtime_set_error(const char *message);

typedef enum {
    PROMISE_PENDING = 0,
    PROMISE_FULFILLED = 1,
    PROMISE_REJECTED = 2
} scriptgo_promise_state;

typedef struct {
	void *fn_ptr;
	void *env;
} scriptgo_closure_inner;

typedef struct scriptgo_promise scriptgo_promise;

typedef struct scriptgo_reaction {
	struct scriptgo_promise *result_promise;
    scriptgo_closure_inner *on_fulfilled;
    scriptgo_closure_inner *on_rejected;
	int is_resume;
	uint32_t result_tag;
	int is_aggregate;
	void *aggregate_env;
	int64_t aggregate_index;
    struct scriptgo_reaction *next;
} scriptgo_reaction;

typedef struct scriptgo_promise {
    scriptgo_promise_state state;
    void *ptr_value;
    double num_value;
    uint64_t int_value;
    uint32_t tag;
    int is_queued;
    scriptgo_reaction *reactions_head;
    scriptgo_reaction *reactions_tail;
    struct scriptgo_promise *all_next;
} scriptgo_promise;

typedef struct {
    void *fn_ptr;
    void *env;
} scriptgo_promise_closure;

typedef struct {
    uint32_t tag;
    uint32_t pad;
    uint64_t payload;
} scriptgo_async_slot;

typedef struct {
	size_t count;
	scriptgo_async_slot *slots;
} scriptgo_async_frame;

typedef struct {
	int64_t remaining;
	void *output;
	scriptgo_promise *result;
} scriptgo_all_numbers_env;

typedef struct {
	int64_t length;
	int64_t capacity;
	int64_t element_size;
	unsigned char *data;
	void *owned_data;
	int64_t element_tag;
} scriptgo_array_view;

int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_get(void *handle, double index, void *out_value);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_set_tag(void *handle, int64_t tag);

typedef struct {
    scriptgo_promise *promise;
    int reject;
} scriptgo_promise_resolver_env;

typedef struct scriptgo_microtask {
    scriptgo_promise *promise;
    scriptgo_closure_inner *closure;
    void *ptr_arg;
    double num_arg;
    int is_number;
    struct scriptgo_microtask *next;
} scriptgo_microtask;

static scriptgo_microtask *microtask_head = NULL;
static scriptgo_microtask *microtask_tail = NULL;
static scriptgo_promise *all_promises = NULL;

static void scriptgo_queue_promise_reactions(scriptgo_promise *p);

extern const char scriptgo_undefined_sentinel;

int scriptgo_async_frame_new(int64_t count, void **out_frame) {
    scriptgo_async_frame *frame;
    if (count < 0 || out_frame == NULL) return scriptgo_runtime_set_error("scriptgo async frame invalid argument");
    frame = calloc(1, sizeof(*frame));
    if (frame == NULL) return scriptgo_runtime_set_error("scriptgo async frame allocation failed");
    frame->count = (size_t)count;
    if (frame->count > 0) {
        frame->slots = calloc(frame->count, sizeof(*frame->slots));
        if (frame->slots == NULL) {
            free(frame);
            return scriptgo_runtime_set_error("scriptgo async frame allocation failed");
        }
    }
    *out_frame = frame;
    return 0;
}

int scriptgo_async_frame_set(void *frame_handle, int64_t index, uint32_t tag, uint64_t payload) {
    scriptgo_async_frame *frame = frame_handle;
    if (frame == NULL || index < 0 || (size_t)index >= frame->count) return scriptgo_runtime_set_error("scriptgo async frame set out of bounds");
    frame->slots[index].tag = tag;
    frame->slots[index].pad = 0;
    frame->slots[index].payload = payload;
    return 0;
}

int scriptgo_async_frame_get(void *frame_handle, int64_t index, uint32_t *out_tag, uint64_t *out_payload) {
    scriptgo_async_frame *frame = frame_handle;
    if (frame == NULL || index < 0 || (size_t)index >= frame->count || out_tag == NULL || out_payload == NULL) {
        return scriptgo_runtime_set_error("scriptgo async frame get out of bounds");
    }
    *out_tag = frame->slots[index].tag;
    *out_payload = frame->slots[index].payload;
    return 0;
}

int scriptgo_async_frame_release(void *frame_handle) {
    scriptgo_async_frame *frame = frame_handle;
    if (frame == NULL) return 0;
    free(frame->slots);
    free(frame);
    return 0;
}

int scriptgo_closure_invoke(void *closure_handle, int32_t arg_count, const scriptgo_boxed_value *a1, const scriptgo_boxed_value *a2, const scriptgo_boxed_value *a3, const scriptgo_boxed_value *a4);

static int scriptgo_promise_set_boxed(scriptgo_promise *p, int rejected, uint32_t tag, uint64_t payload);
int scriptgo_promise_resolve_existing(void *promise_handle, void *value);
int scriptgo_promise_resolve_existing_number(void *promise_handle, double value);
int scriptgo_promise_resolve_existing_array(void *promise_handle, void *value);

static uint64_t scriptgo_promise_payload(const scriptgo_promise *p) {
    uint64_t payload = 0;
    if (p->tag == 3) {
        memcpy(&payload, &p->num_value, sizeof(payload));
    } else {
        payload = (uint64_t)(uintptr_t)p->ptr_value;
    }
    return payload;
}

void scriptgo_throw_bool(int val);
int scriptgo_string_from_bigint(long long value, char **out_str);
int scriptgo_string_from_object(void *obj, char **out_str);
typedef struct scriptgo_exception_frame scriptgo_exception_frame_t;
scriptgo_exception_frame_t *scriptgo_exception_frame_new(void);
void *scriptgo_exception_buf(scriptgo_exception_frame_t *frame);
void scriptgo_exception_frame_free(scriptgo_exception_frame_t *frame);
int scriptgo_exception_get_tag_payload(scriptgo_exception_frame_t *frame, uint32_t *out_tag, uint64_t *out_payload);

typedef struct {
    int threw;
    uint32_t tag;
    uint64_t payload;
} scriptgo_reaction_result;

static scriptgo_reaction_result scriptgo_invoke_reaction(
    scriptgo_closure_inner *closure,
    uint32_t input_tag,
    uint64_t input_payload,
    uint32_t result_tag) {
    scriptgo_reaction_result result = {0, 0, 0};
    scriptgo_exception_frame_t *frame = scriptgo_exception_frame_new();
    if (frame == NULL) {
        result.threw = 1;
        result.tag = 4;
        result.payload = (uint64_t)(uintptr_t)"exception frame allocation failed";
        return result;
    }
#if defined(__wasi__)
    int jumped = 0;
#else
    int jumped = setjmp(*(jmp_buf *)scriptgo_exception_buf(frame));
#endif
    if (jumped != 0) {
        result.threw = 1;
        scriptgo_exception_get_tag_payload(frame, &result.tag, &result.payload);
        scriptgo_exception_frame_free(frame);
        return result;
    }
    if (result_tag == 3) {
        double (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (double (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))closure->fn_ptr;
        double value = fn(closure->env, input_tag, 0, input_payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
        memcpy(&result.payload, &value, sizeof(value));
    } else if (result_tag == 0) {
        struct scriptgo_boxed_result { uint32_t tag; uint32_t pad; uint64_t payload; };
        struct scriptgo_boxed_result (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (struct scriptgo_boxed_result (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))closure->fn_ptr;
        (void)fn(closure->env, input_tag, 0, input_payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
    } else {
        uint64_t (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (uint64_t (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))closure->fn_ptr;
        result.payload = fn(closure->env, input_tag, 0, input_payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
    }
    result.tag = result_tag;
    scriptgo_exception_frame_free(frame);
    return result;
}

static void scriptgo_promise_resolver_callback(
    void *env_ptr,
    int32_t tag0, int32_t pad0, int64_t payload0,
    int32_t tag1, int32_t pad1, int64_t payload1,
    int32_t tag2, int32_t pad2, int64_t payload2,
    int32_t tag3, int32_t pad3, int64_t payload3);

static int scriptgo_promise_resolver_create_internal(scriptgo_promise *promise, int reject, void **out_closure) {
    scriptgo_promise_resolver_env *env;
    scriptgo_promise_closure *closure;
    if (promise == NULL || out_closure == NULL) {
        return scriptgo_runtime_set_error("scriptgo promise resolver invalid argument");
    }
    env = calloc(1, sizeof(*env));
    closure = calloc(1, sizeof(*closure));
    if (env == NULL || closure == NULL) {
        return scriptgo_runtime_set_error("scriptgo promise resolver allocation failed");
    }
    env->promise = promise;
    env->reject = reject ? 1 : 0;
    closure->fn_ptr = (void *)scriptgo_promise_resolver_callback;
    closure->env = env;
    *out_closure = closure;
    return 0;
}

static void scriptgo_promise_resolver_callback(
    void *env_ptr,
    int32_t tag0, int32_t pad0, int64_t payload0,
    int32_t tag1, int32_t pad1, int64_t payload1,
    int32_t tag2, int32_t pad2, int64_t payload2,
    int32_t tag3, int32_t pad3, int64_t payload3) {
    (void)tag1; (void)pad1; (void)payload1;
    (void)tag2; (void)pad2; (void)payload2;
    (void)tag3; (void)pad3; (void)payload3;
    (void)pad0;
    scriptgo_promise_resolver_env *env = (scriptgo_promise_resolver_env *)env_ptr;
    if (env == NULL || env->promise == NULL) return;
    (void)scriptgo_promise_set_boxed(env->promise, env->reject, (uint32_t)tag0, (uint64_t)payload0);
}

static void queue_microtask_node(scriptgo_microtask *task) {
    task->next = NULL;
    if (microtask_tail != NULL) {
        microtask_tail->next = task;
        microtask_tail = task;
    } else {
        microtask_head = task;
        microtask_tail = task;
    }
}

static scriptgo_promise *scriptgo_find_promise_payload(uint64_t payload) {

	scriptgo_promise *candidate = all_promises;
	while (candidate != NULL) {
		if ((uint64_t)(uintptr_t)candidate == payload) return candidate;
		candidate = candidate->all_next;
	}
	return NULL;
}

static int scriptgo_promise_adopt(scriptgo_promise *target, scriptgo_promise *source) {
	scriptgo_reaction *reaction;
	if (target == NULL || source == NULL) return scriptgo_runtime_set_error("scriptgo promise adoption failed");
	if (target == source) return scriptgo_runtime_set_error("scriptgo promise self-resolution");
	reaction = calloc(1, sizeof(*reaction));
	if (reaction == NULL) return scriptgo_runtime_set_error("scriptgo promise reaction allocation failed");
	reaction->result_promise = target;
	reaction->result_tag = 5;
	if (source->reactions_tail != NULL) {
		source->reactions_tail->next = reaction;
		source->reactions_tail = reaction;
	} else {
		source->reactions_head = reaction;
		source->reactions_tail = reaction;
	}
	if (source->state != PROMISE_PENDING) scriptgo_queue_promise_reactions(source);
	return 0;
}

int scriptgo_queue_microtask(void *closure_handle, void *arg) {
    if (closure_handle == NULL) return 0;
    scriptgo_microtask *task = malloc(sizeof(scriptgo_microtask));
    if (task == NULL) return scriptgo_runtime_set_error("scriptgo microtask allocation failed");
    task->promise = NULL;
    task->closure = (scriptgo_closure_inner *)closure_handle;
    task->ptr_arg = arg;
    task->num_arg = 0.0;
    task->is_number = 0;
    queue_microtask_node(task);
    return 0;
}

int scriptgo_queue_microtask_number(void *closure_handle, double value) {
    if (closure_handle == NULL) return 0;
    scriptgo_microtask *task = malloc(sizeof(scriptgo_microtask));
    if (task == NULL) return scriptgo_runtime_set_error("scriptgo microtask allocation failed");
    task->promise = NULL;
    task->closure = (scriptgo_closure_inner *)closure_handle;
    task->ptr_arg = NULL;
    task->num_arg = value;
    task->is_number = 1;
    queue_microtask_node(task);
    return 0;
}

static void scriptgo_queue_promise_reactions(scriptgo_promise *p) {
    if (p == NULL || p->reactions_head == NULL || p->is_queued) return;
    scriptgo_microtask *task = malloc(sizeof(scriptgo_microtask));
    if (task == NULL) return;
    p->is_queued = 1;
    task->promise = p;
    task->closure = NULL;
    task->ptr_arg = NULL;
    task->num_arg = 0.0;
    task->is_number = 0;
    queue_microtask_node(task);
}

static int scriptgo_promise_schedule_resume_internal(scriptgo_promise *p, scriptgo_closure_inner *closure) {
    scriptgo_reaction *r;
    if (p == NULL || closure == NULL) return scriptgo_runtime_set_error("scriptgo promise resume invalid argument");
    r = calloc(1, sizeof(*r));
    if (r == NULL) return scriptgo_runtime_set_error("scriptgo promise reaction allocation failed");
	r->result_promise = NULL;
	r->on_fulfilled = closure;
    r->on_rejected = closure;
    r->is_resume = 1;
    if (p->reactions_tail != NULL) {
        p->reactions_tail->next = r;
        p->reactions_tail = r;
    } else {
        p->reactions_head = r;
        p->reactions_tail = r;
    }
    if (p->state != PROMISE_PENDING) scriptgo_queue_promise_reactions(p);
    return 0;
}

int scriptgo_promise_schedule_resume(void *promise_handle, void *closure_handle) {
    return scriptgo_promise_schedule_resume_internal(
        (scriptgo_promise *)promise_handle,
        (scriptgo_closure_inner *)closure_handle);
}

int scriptgo_promise_schedule_resume_pair(void *promise_handle, void *fulfilled_handle, void *rejected_handle) {
    scriptgo_promise *p = promise_handle;
    scriptgo_reaction *r;
    if (p == NULL || (fulfilled_handle == NULL && rejected_handle == NULL)) {
        return scriptgo_runtime_set_error("scriptgo promise resume pair invalid argument");
    }
    r = calloc(1, sizeof(*r));
    if (r == NULL) return scriptgo_runtime_set_error("scriptgo promise reaction allocation failed");
	r->on_fulfilled = (scriptgo_closure_inner *)fulfilled_handle;
	r->on_rejected = (scriptgo_closure_inner *)rejected_handle;
	r->result_promise = NULL;
    r->is_resume = 1;
    if (p->reactions_tail != NULL) {
        p->reactions_tail->next = r;
        p->reactions_tail = r;
    } else {
        p->reactions_head = r;
        p->reactions_tail = r;
    }
    if (p->state != PROMISE_PENDING) scriptgo_queue_promise_reactions(p);
    return 0;
}

int scriptgo_event_loop_run(void) {
    while (microtask_head != NULL) {
        scriptgo_microtask *task = microtask_head;
        microtask_head = task->next;
        if (microtask_head == NULL) {
            microtask_tail = NULL;
        }
        if (task->promise != NULL) {
            scriptgo_promise *p = task->promise;
            p->is_queued = 0;
            scriptgo_reaction *r = p->reactions_head;
			if (r != NULL) {
                p->reactions_head = r->next;
                if (p->reactions_head == NULL) {
                    p->reactions_tail = NULL;
                }
				if (r->is_aggregate) {
					scriptgo_all_numbers_env *aggregate = r->aggregate_env;
					if (aggregate != NULL && p->state == PROMISE_FULFILLED && p->tag == 3) {
						double value = p->num_value;
						if (scriptgo_array_set(aggregate->output, (double)r->aggregate_index, &value) != 0) {
							return -1;
						}
						aggregate->remaining--;
						if (aggregate->remaining == 0) {
							scriptgo_promise_resolve_existing_array(aggregate->result, aggregate->output);
						}
					} else if (aggregate != NULL) {
						scriptgo_promise_set_boxed(aggregate->result, 1, p->tag ? p->tag : 5, scriptgo_promise_payload(p));
					}
					free(r);
					if (p->reactions_head != NULL) scriptgo_queue_promise_reactions(p);
					continue;
				}
					if (r->is_resume) {
                    uint64_t payload = 0;
                    uint32_t tag = p->tag;
					if (tag == 3) {
                        memcpy(&payload, &p->num_value, sizeof(double));
                    } else if (tag == 2 || tag == 8) {
                        payload = p->int_value;
                    } else if (p->ptr_value != NULL) {
                        payload = (uint64_t)(uintptr_t)p->ptr_value;
                    }
                    if (p->state == PROMISE_REJECTED) tag = p->tag ? p->tag : 5;
                    scriptgo_closure_inner *resume = p->state == PROMISE_REJECTED ? r->on_rejected : r->on_fulfilled;
                    if (resume != NULL && resume->fn_ptr != NULL) {
                        void (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                            (void (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))resume->fn_ptr;
                        fn(resume->env, tag, 0, payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
                    }
				} else {
					scriptgo_closure_inner *handler = p->state == PROMISE_FULFILLED ? r->on_fulfilled : r->on_rejected;
					uint32_t source_tag = p->tag ? p->tag : 5;
					uint64_t source_payload = scriptgo_promise_payload(p);
					if (handler == NULL || handler->fn_ptr == NULL) {
						scriptgo_promise_set_boxed(r->result_promise, p->state == PROMISE_REJECTED, source_tag, source_payload);
					} else {
						scriptgo_reaction_result callback = scriptgo_invoke_reaction(handler, source_tag, source_payload, r->result_tag);
						if (callback.threw) {
							scriptgo_promise_set_boxed(r->result_promise, 1, callback.tag, callback.payload);
						} else if (r->result_tag == 0) {
							scriptgo_promise_set_boxed(r->result_promise, 0, 0, 0);
						} else {
							scriptgo_promise_set_boxed(r->result_promise, 0, callback.tag, callback.payload);
						}
					}
				}
				free(r);
                if (p->reactions_head != NULL) {
                    scriptgo_queue_promise_reactions(p);
                }
            }
        } else if (task->closure != NULL && task->closure->fn_ptr != NULL) {
            uint64_t payload = 0;
            uint32_t tag = 0;
            if (task->is_number) {
                tag = 3; // number
                memcpy(&payload, &task->num_arg, sizeof(double));
            } else if (task->ptr_arg != NULL) {
                tag = 4; // string/object
                payload = (uint64_t)(uintptr_t)task->ptr_arg;
            }
            void (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                (void (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))task->closure->fn_ptr;
            fn(task->closure->env, tag, 0, payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
        }
        free(task);
    }
    return 0;
}

int scriptgo_promise_create(void **out_promise) {
    scriptgo_promise *p;
    if (out_promise == NULL) return scriptgo_runtime_set_error("scriptgo promise allocation failed");
    p = calloc(1, sizeof(scriptgo_promise));
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise allocation failed");
    p->state = PROMISE_PENDING;
    p->ptr_value = NULL;
    p->num_value = 0.0;
    p->int_value = 0;
    p->tag = 0;
    p->reactions_head = NULL;
    p->reactions_tail = NULL;
    p->all_next = all_promises;
    all_promises = p;
    *out_promise = p;
    return 0;
}

int scriptgo_promise_construct(void *executor_handle, void **out_promise) {
    scriptgo_promise *promise = NULL;
    void *resolve = NULL;
    void *reject = NULL;
    scriptgo_boxed_value resolve_value;
    scriptgo_boxed_value reject_value;

    if (executor_handle == NULL || out_promise == NULL) {
        return scriptgo_runtime_set_error("scriptgo promise constructor invalid argument");
    }
    if (scriptgo_promise_create((void **)&promise) != 0) return -1;

	if (scriptgo_promise_resolver_create_internal(promise, 0, &resolve) != 0 ||
		scriptgo_promise_resolver_create_internal(promise, 1, &reject) != 0) {
		return -1;
	}

    resolve_value.tag = 7;
    resolve_value.pad = 0;
	resolve_value.payload = (int64_t)(uintptr_t)resolve;
    reject_value.tag = 7;
    reject_value.pad = 0;
    reject_value.payload = (int64_t)(uintptr_t)reject;
    if (scriptgo_closure_invoke(executor_handle, 2, &resolve_value, &reject_value, NULL, NULL) != 0) {
        return -1;
    }
    *out_promise = promise;
	return 0;
}

int scriptgo_promise_all_numbers(void *array_handle, void **out_promise) {
	scriptgo_array_view *input = array_handle;
	scriptgo_all_numbers_env *aggregate;
	void *output = NULL;
	if (input == NULL || out_promise == NULL || input->element_size != (int64_t)sizeof(void *)) {
		return scriptgo_runtime_set_error("Promise.all requires an array of number promises");
	}
	if (scriptgo_promise_create(out_promise) != 0) return -1;
	aggregate = calloc(1, sizeof(*aggregate));
	if (aggregate == NULL) return scriptgo_runtime_set_error("Promise.all allocation failed");
	if (scriptgo_array_new(input->length, (int64_t)sizeof(double), &output) != 0) return -1;
	if (scriptgo_array_set_tag(output, 6) != 0) return -1;
	aggregate->remaining = input->length;
	aggregate->output = output;
	aggregate->result = (scriptgo_promise *)*out_promise;
	if (input->length == 0) {
		scriptgo_promise_resolve_existing_array(aggregate->result, output);
		free(aggregate);
		return 0;
	}
	for (int64_t index = 0; index < input->length; index++) {
		void *promise_handle = NULL;
		if (scriptgo_array_get(input, (double)index, &promise_handle) != 0 || promise_handle == NULL) return -1;
		scriptgo_promise *promise = promise_handle;
		scriptgo_reaction *reaction = calloc(1, sizeof(*reaction));
		if (reaction == NULL) return scriptgo_runtime_set_error("Promise.all reaction allocation failed");
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

int scriptgo_promise_resolver_create(void *promise_handle, int reject, void **out_closure) {
    return scriptgo_promise_resolver_create_internal((scriptgo_promise *)promise_handle, reject, out_closure);
}

static int scriptgo_promise_set_boxed(scriptgo_promise *p, int rejected, uint32_t tag, uint64_t payload) {
	scriptgo_promise *source;
	if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise resolve failed");
	if (p->state != PROMISE_PENDING) return 0;
	if (!rejected && tag == 5 && (source = scriptgo_find_promise_payload(payload)) != NULL) {
		return scriptgo_promise_adopt(p, source);
	}
    p->state = rejected ? PROMISE_REJECTED : PROMISE_FULFILLED;
    p->tag = tag;
    p->ptr_value = NULL;
    if (tag == 3) {
        memcpy(&p->num_value, &payload, sizeof(p->num_value));
    } else if (tag == 2 || tag == 8) {
        p->int_value = payload;
    } else if (tag == 0) {
        p->ptr_value = (void *)&scriptgo_undefined_sentinel;
    } else if (tag != 1) {
        p->ptr_value = (void *)(uintptr_t)payload;
    }
    scriptgo_queue_promise_reactions(p);
    return 0;
}

void scriptgo_throw_string(const char *str);
void scriptgo_throw_number(double num);

int scriptgo_promise_resolve(void *promise_handle, void *value) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 5, (uint64_t)(uintptr_t)value);
}

int scriptgo_promise_resolve_number(void *promise_handle, double value) {
    uint64_t payload = 0;
    memcpy(&payload, &value, sizeof(payload));
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 3, payload);
}

int scriptgo_promise_reject(void *promise_handle, void *reason) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 1, 5, (uint64_t)(uintptr_t)reason);
}

int scriptgo_promise_resolve_existing(void *promise_handle, void *value) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 5, (uint64_t)(uintptr_t)value);
}

int scriptgo_promise_reject_existing(void *promise_handle, void *reason) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 1, 5, (uint64_t)(uintptr_t)reason);
}

int scriptgo_promise_resolve_existing_number(void *promise_handle, double value) {
    uint64_t payload = 0;
    memcpy(&payload, &value, sizeof(payload));
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 3, payload);
}

int scriptgo_promise_resolve_existing_array(void *promise_handle, void *value) {
	return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 6, (uint64_t)(uintptr_t)value);
}

int scriptgo_promise_resolve_existing_bool(void *promise_handle, int value) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 2, (uint64_t)(value != 0));
}

int scriptgo_promise_resolve_existing_boxed(void *promise_handle, uint32_t tag, uint64_t payload) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, tag, payload);
}

int scriptgo_promise_reject_existing_boxed(void *promise_handle, uint32_t tag, uint64_t payload) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 1, tag, payload);
}

int scriptgo_promise_resolve_bool(void *promise_handle, int value) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 2, (uint64_t)(value != 0));
}

int scriptgo_promise_resolve_bigint(void *promise_handle, int64_t value) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, 8, (uint64_t)value);
}

int scriptgo_promise_resolve_boxed(void *promise_handle, uint32_t tag, uint64_t payload) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 0, tag, payload);
}

int scriptgo_promise_resolve_unknown(uint32_t tag, uint64_t payload, void **out_promise) {
	scriptgo_promise *promise;
	if (out_promise == NULL) return scriptgo_runtime_set_error("scriptgo unknown promise output missing");
	if (scriptgo_promise_create((void **)&promise) != 0) return -1;
	if (scriptgo_promise_set_boxed(promise, 0, tag, payload) != 0) return -1;
	*out_promise = promise;
	return 0;
}

int scriptgo_promise_reject_boxed(void *promise_handle, uint32_t tag, uint64_t payload) {
    return scriptgo_promise_set_boxed((scriptgo_promise *)promise_handle, 1, tag, payload);
}

int scriptgo_promise_then(void *promise_handle, void *on_fulfilled_closure, void *on_rejected_closure, uint32_t result_tag, void **out_result) {
	scriptgo_promise *p = promise_handle;
	scriptgo_promise *result = NULL;
	if (p == NULL || out_result == NULL) return scriptgo_runtime_set_error("scriptgo promise then failed");
	if (scriptgo_promise_create((void **)&result) != 0) return -1;
	scriptgo_reaction *r = malloc(sizeof(scriptgo_reaction));
    if (r == NULL) return scriptgo_runtime_set_error("scriptgo promise reaction allocation failed");
	r->on_fulfilled = (scriptgo_closure_inner *)on_fulfilled_closure;
	r->on_rejected = (scriptgo_closure_inner *)on_rejected_closure;
	r->result_promise = result;
	r->result_tag = result_tag;
    r->is_resume = 0;
    r->next = NULL;
    if (p->reactions_tail != NULL) {
        p->reactions_tail->next = r;
        p->reactions_tail = r;
    } else {
        p->reactions_head = r;
        p->reactions_tail = r;
    }
	if (p->state != PROMISE_PENDING) {
		scriptgo_queue_promise_reactions(p);
	}
	*out_result = result;
	return 0;
}

static int scriptgo_promise_wait(scriptgo_promise *p) {
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (p->state == PROMISE_PENDING) {
        scriptgo_event_loop_run();
    }
    if (p->state != PROMISE_REJECTED) return 0;
    switch (p->tag) {
    case 2:
        scriptgo_throw_bool((int)p->int_value);
        break;
    case 3:
        scriptgo_throw_number(p->num_value);
        break;
    case 8: {
        char *text = NULL;
        if (scriptgo_string_from_bigint((long long)p->int_value, &text) == 0) {
            scriptgo_throw_string(text);
        }
        break;
    }
    case 4:
        scriptgo_throw_string(p->ptr_value ? (const char *)p->ptr_value : "");
        break;
    case 0:
        scriptgo_throw_string("undefined");
        break;
    case 1:
        scriptgo_throw_string("null");
        break;
    default: {
        char *text = NULL;
        if (scriptgo_string_from_object(p->ptr_value, &text) == 0) {
            scriptgo_throw_string(text);
        }
        break;
    }
    }
    return -1;
}

int scriptgo_promise_await_number(void *promise_handle, double *out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (scriptgo_promise_wait(p) != 0) return -1;
    if (p->state == PROMISE_FULFILLED) {
        *out_val = p->num_value;
    } else {
        *out_val = 0.0;
    }
    return 0;
}

int scriptgo_promise_await_ptr(void *promise_handle, void **out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (scriptgo_promise_wait(p) != 0) return -1;
    if (p->state == PROMISE_FULFILLED) {
        *out_val = p->ptr_value;
    } else {
        *out_val = NULL;
    }
    return 0;
}

int scriptgo_promise_await_bool(void *promise_handle, int32_t *out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (scriptgo_promise_wait(p) != 0) return -1;
    if (p->tag != 2) return scriptgo_runtime_set_error("scriptgo promise fulfilled with non-boolean value");
    *out_val = p->int_value != 0;
    return 0;
}

int scriptgo_promise_await_bigint(void *promise_handle, int64_t *out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (scriptgo_promise_wait(p) != 0) return -1;
    if (p->tag != 8) return scriptgo_runtime_set_error("scriptgo promise fulfilled with non-bigint value");
    *out_val = (int64_t)p->int_value;
    return 0;
}

int scriptgo_promise_await_boxed(void *promise_handle, uint32_t *out_tag, uint64_t *out_payload) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_tag == NULL || out_payload == NULL) {
        return scriptgo_runtime_set_error("scriptgo promise await output missing");
    }
    if (scriptgo_promise_wait(p) != 0) return -1;
    *out_tag = p->tag;
    if (p->tag == 3) {
        memcpy(out_payload, &p->num_value, sizeof(*out_payload));
    } else if (p->tag == 2 || p->tag == 8) {
        *out_payload = p->int_value;
    } else if (p->tag == 0 || p->tag == 1) {
        *out_payload = 0;
    } else {
        *out_payload = (uint64_t)(uintptr_t)p->ptr_value;
    }
    return 0;
}

static scriptgo_promise *scriptgo_find_promise(void *handle) {
    scriptgo_promise *p = all_promises;
    while (p != NULL) {
        if ((void *)p == handle) return p;
        p = p->all_next;
    }
    return NULL;
}

/* Await an unknown value without mistaking ordinary values for Promise handles. */
int scriptgo_promise_await_unknown(uint32_t tag, uint64_t payload, uint32_t *out_tag, uint64_t *out_payload) {
    if (out_tag == NULL || out_payload == NULL) {
        return scriptgo_runtime_set_error("scriptgo unknown await output missing");
    }

    scriptgo_promise *p = NULL;
    if (tag == 5 && payload != 0) {
        p = scriptgo_find_promise((void *)(uintptr_t)payload);
    }
    if (p == NULL) {
        *out_tag = tag;
        *out_payload = payload;
        return 0;
    }

    if (scriptgo_promise_wait(p) != 0) return -1;

    if (p->tag == 3) {
        uint64_t bits = 0;
        memcpy(&bits, &p->num_value, sizeof(bits));
        *out_tag = 3;
        *out_payload = bits;
    } else if (p->tag == 2 || p->tag == 8) {
        *out_tag = p->tag;
        *out_payload = p->int_value;
    } else if (p->ptr_value == NULL) {
        *out_tag = 1;
        *out_payload = 0;
    } else if (p->ptr_value == (void *)&scriptgo_undefined_sentinel) {
        *out_tag = 0;
        *out_payload = 0;
    } else {
        *out_tag = p->tag == 4 ? 4 : 5;
        *out_payload = (uint64_t)(uintptr_t)p->ptr_value;
    }
    return 0;
}
