#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

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

typedef struct scriptgo_promise {
    scriptgo_promise_state state;
    void *ptr_value;
    double num_value;
    int is_number;
    scriptgo_closure_inner *on_fulfilled;
    scriptgo_closure_inner *on_rejected;
} scriptgo_promise;

typedef struct scriptgo_microtask {
    scriptgo_closure_inner *closure;
    void *ptr_arg;
    double num_arg;
    int is_number;
    struct scriptgo_microtask *next;
} scriptgo_microtask;

static scriptgo_microtask *microtask_head = NULL;
static scriptgo_microtask *microtask_tail = NULL;

int scriptgo_queue_microtask(void *closure_handle, void *arg) {
    scriptgo_microtask *task;
    if (closure_handle == NULL) return 0;
    task = malloc(sizeof(scriptgo_microtask));
    if (task == NULL) return scriptgo_runtime_set_error("scriptgo microtask allocation failed");
    task->closure = (scriptgo_closure_inner *)closure_handle;
    task->ptr_arg = arg;
    task->num_arg = 0.0;
    task->is_number = 0;
    task->next = NULL;
    if (microtask_tail != NULL) {
        microtask_tail->next = task;
        microtask_tail = task;
    } else {
        microtask_head = task;
        microtask_tail = task;
    }
    return 0;
}

int scriptgo_queue_microtask_number(void *closure_handle, double value) {
    scriptgo_microtask *task;
    if (closure_handle == NULL) return 0;
    task = malloc(sizeof(scriptgo_microtask));
    if (task == NULL) return scriptgo_runtime_set_error("scriptgo microtask allocation failed");
    task->closure = (scriptgo_closure_inner *)closure_handle;
    task->ptr_arg = NULL;
    task->num_arg = value;
    task->is_number = 1;
    task->next = NULL;
    if (microtask_tail != NULL) {
        microtask_tail->next = task;
        microtask_tail = task;
    } else {
        microtask_head = task;
        microtask_tail = task;
    }
    return 0;
}

int scriptgo_event_loop_run(void) {
    while (microtask_head != NULL) {
        scriptgo_microtask *task = microtask_head;
        microtask_head = task->next;
        if (microtask_head == NULL) {
            microtask_tail = NULL;
        }
        if (task->closure != NULL && task->closure->fn_ptr != NULL) {
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
    p->is_number = 0;
    p->on_fulfilled = NULL;
    p->on_rejected = NULL;
    *out_promise = p;
    return 0;
}

void scriptgo_throw_string(const char *str);
void scriptgo_throw_number(double num);

int scriptgo_promise_resolve(void *promise_handle, void *value) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise resolve failed");
    if (p->state != PROMISE_PENDING) return 0;
    p->state = PROMISE_FULFILLED;
    p->ptr_value = value;
    p->is_number = 0;
    if (p->on_fulfilled != NULL) {
        scriptgo_queue_microtask(p->on_fulfilled, value);
    }
    return 0;
}

int scriptgo_promise_resolve_number(void *promise_handle, double value) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise resolve failed");
    if (p->state != PROMISE_PENDING) return 0;
    p->state = PROMISE_FULFILLED;
    p->num_value = value;
    p->is_number = 1;
    if (p->on_fulfilled != NULL) {
        scriptgo_queue_microtask_number(p->on_fulfilled, value);
    }
    return 0;
}

int scriptgo_promise_reject(void *promise_handle, void *reason) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise reject failed");
    if (p->state != PROMISE_PENDING) return 0;
    p->state = PROMISE_REJECTED;
    p->ptr_value = reason;
    p->is_number = 0;
    if (p->on_rejected != NULL) {
        scriptgo_queue_microtask(p->on_rejected, reason);
    }
    return 0;
}

int scriptgo_promise_then(void *promise_handle, void *on_fulfilled_closure, void **out_promise) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise then failed");
    if (out_promise != NULL) {
        if (scriptgo_promise_create(out_promise) != 0) return -1;
    }
    p->on_fulfilled = (scriptgo_closure_inner *)on_fulfilled_closure;
    if (p->state == PROMISE_FULFILLED) {
        if (p->is_number) {
            scriptgo_queue_microtask_number(on_fulfilled_closure, p->num_value);
        } else {
            scriptgo_queue_microtask(on_fulfilled_closure, p->ptr_value);
        }
    }
    return 0;
}

int scriptgo_promise_await_number(void *promise_handle, double *out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (p->state == PROMISE_PENDING) {
        scriptgo_event_loop_run();
    }
    if (p->state == PROMISE_REJECTED) {
        if (p->is_number) {
            scriptgo_throw_number(p->num_value);
        } else {
            scriptgo_throw_string(p->ptr_value ? (const char *)p->ptr_value : "Promise rejected");
        }
        return -1;
    }
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
    if (p->state == PROMISE_PENDING) {
        scriptgo_event_loop_run();
    }
    if (p->state == PROMISE_REJECTED) {
        scriptgo_throw_string(p->ptr_value ? (const char *)p->ptr_value : "Promise rejected");
        return -1;
    }
    if (p->state == PROMISE_FULFILLED) {
        *out_val = p->ptr_value;
    } else {
        *out_val = NULL;
    }
    return 0;
}

