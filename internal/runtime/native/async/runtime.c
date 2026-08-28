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

typedef struct scriptgo_reaction {
    scriptgo_closure_inner *on_fulfilled;
    scriptgo_closure_inner *on_rejected;
    struct scriptgo_reaction *next;
} scriptgo_reaction;

typedef struct scriptgo_promise {
    scriptgo_promise_state state;
    void *ptr_value;
    double num_value;
    uint32_t tag;
    int is_queued;
    scriptgo_reaction *reactions_head;
    scriptgo_reaction *reactions_tail;
} scriptgo_promise;

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
                if (p->state == PROMISE_FULFILLED && r->on_fulfilled != NULL && r->on_fulfilled->fn_ptr != NULL) {
                    uint64_t payload = 0;
                    uint32_t tag = p->tag;
                    if (tag == 3) {
                        memcpy(&payload, &p->num_value, sizeof(double));
                    } else if (p->ptr_value != NULL) {
                        payload = (uint64_t)(uintptr_t)p->ptr_value;
                    }
                    if (tag == 3) {
                        double (*fn_d)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                            (double (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))r->on_fulfilled->fn_ptr;
                        double res_num = fn_d(r->on_fulfilled->env, tag, 0, payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
                        p->num_value = res_num;
                        p->tag = 3;
                    } else {
                        uint64_t (*fn_ptr)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                            (uint64_t (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))r->on_fulfilled->fn_ptr;
                        uint64_t res_p = fn_ptr(r->on_fulfilled->env, tag, 0, payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
                        p->ptr_value = (void *)(uintptr_t)res_p;
                        p->tag = 4;
                    }
                } else if (p->state == PROMISE_REJECTED && r->on_rejected != NULL && r->on_rejected->fn_ptr != NULL) {
                    uint64_t payload = 0;
                    uint32_t tag = p->tag ? p->tag : 5; // object/error
                    if (tag == 3) {
                        memcpy(&payload, &p->num_value, sizeof(double));
                    } else if (p->ptr_value != NULL) {
                        payload = (uint64_t)(uintptr_t)p->ptr_value;
                    }
                    uint64_t (*fn_ptr)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                        (uint64_t (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))r->on_rejected->fn_ptr;
                    uint64_t res_p = fn_ptr(r->on_rejected->env, tag, 0, payload, 0, 0, 0, 0, 0, 0, 0, 0, 0);
                    p->state = PROMISE_FULFILLED;
                    p->ptr_value = (void *)(uintptr_t)res_p;
                    p->tag = 4;
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
    p->tag = 0;
    p->reactions_head = NULL;
    p->reactions_tail = NULL;
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
    p->tag = 5;
    scriptgo_queue_promise_reactions(p);
    return 0;
}

int scriptgo_promise_resolve_number(void *promise_handle, double value) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise resolve failed");
    if (p->state != PROMISE_PENDING) return 0;
    p->state = PROMISE_FULFILLED;
    p->num_value = value;
    p->tag = 3;
    scriptgo_queue_promise_reactions(p);
    return 0;
}

int scriptgo_promise_reject(void *promise_handle, void *reason) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise reject failed");
    if (p->state != PROMISE_PENDING) return 0;
    p->state = PROMISE_REJECTED;
    p->ptr_value = reason;
    p->tag = 5;
    scriptgo_queue_promise_reactions(p);
    return 0;
}

int scriptgo_promise_then(void *promise_handle, void *on_fulfilled_closure, void *on_rejected_closure) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL) return scriptgo_runtime_set_error("scriptgo promise then failed");
    scriptgo_reaction *r = malloc(sizeof(scriptgo_reaction));
    if (r == NULL) return scriptgo_runtime_set_error("scriptgo promise reaction allocation failed");
    r->on_fulfilled = (scriptgo_closure_inner *)on_fulfilled_closure;
    r->on_rejected = (scriptgo_closure_inner *)on_rejected_closure;
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
    return 0;
}

int scriptgo_promise_await_number(void *promise_handle, double *out_val) {
    scriptgo_promise *p = promise_handle;
    if (p == NULL || out_val == NULL) return scriptgo_runtime_set_error("scriptgo promise await failed");
    if (p->state == PROMISE_PENDING) {
        scriptgo_event_loop_run();
    }
    if (p->state == PROMISE_REJECTED) {
        if (p->tag == 3) {
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


