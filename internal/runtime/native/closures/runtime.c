#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_push(void *handle, const void *value, double *out_length);
int scriptgo_gc_register(void *ptr, int tag, uint32_t field_count);

#define SCRIPTGO_CLOSURE_GC_TAG 3

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    unsigned char *data;
    void *owned_data;
} scriptgo_array_inner;

typedef struct {
    void *fn_ptr;
    void *env;
    void *invoke_ptr;
    int32_t return_tag;
} scriptgo_closure;

extern const char scriptgo_undefined_sentinel;

int scriptgo_closure_create(void *fn_ptr, void *env, void *invoke_ptr, int32_t return_tag, void **out_closure) {
    scriptgo_closure *c;
    if (out_closure == NULL) return scriptgo_runtime_set_error("scriptgo closure allocation failed");
    c = malloc(sizeof(scriptgo_closure));
    if (c == NULL) return scriptgo_runtime_set_error("scriptgo closure allocation failed");
    c->fn_ptr = fn_ptr;
    c->env = env;
    c->invoke_ptr = invoke_ptr;
    c->return_tag = return_tag;
    if (scriptgo_gc_register(c, SCRIPTGO_CLOSURE_GC_TAG, 0) != 0) {
        free(c);
        return scriptgo_runtime_set_error("scriptgo closure registration failed");
    }
    *out_closure = c;
    return 0;
}

void *scriptgo_closure_alloc(int64_t size) {
    if (size <= 0) size = 8;
    return malloc((size_t)size);
}

int scriptgo_closure_equals(void *h1, void *h2) {
    if (h1 == h2) return 1;
    if (h1 == NULL || h2 == NULL || h1 == &scriptgo_undefined_sentinel || h2 == &scriptgo_undefined_sentinel) return 0;
    scriptgo_closure *c1 = h1;
    scriptgo_closure *c2 = h2;
    return (c1->fn_ptr == c2->fn_ptr && c1->env == c2->env) ? 1 : 0;
}

int scriptgo_closure_invoke_value(void *closure_handle, int32_t arg_count, const scriptgo_value *a1, const scriptgo_value *a2, const scriptgo_value *a3, const scriptgo_value *a4, scriptgo_value *out_value) {
    scriptgo_value result = {0};
    if (out_value == NULL) return scriptgo_runtime_set_error("scriptgo closure result is null");
    *out_value = result;
    if (closure_handle == NULL || closure_handle == &scriptgo_undefined_sentinel) return 0;
    scriptgo_closure *c = closure_handle;
    if (c->fn_ptr == NULL) return 0;
    scriptgo_value dummy = {0};
    const scriptgo_value *v1 = (a1 != NULL && arg_count >= 1) ? a1 : &dummy;
    const scriptgo_value *v2 = (a2 != NULL && arg_count >= 2) ? a2 : &dummy;
    const scriptgo_value *v3 = (a3 != NULL && arg_count >= 3) ? a3 : &dummy;
    const scriptgo_value *v4 = (a4 != NULL && arg_count >= 4) ? a4 : &dummy;
    if (c->invoke_ptr != NULL) {
        void (*invoke)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, scriptgo_value *) =
            (void (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, scriptgo_value *))c->invoke_ptr;
        invoke(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload, &result);
        *out_value = result;
        return 0;
    }
    if (c->return_tag == -1) {
        scriptgo_value (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (scriptgo_value (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        result = fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload);
    } else if (c->return_tag == SCRIPTGO_TAG_UNDEFINED) {
        void (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (void (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload);
    } else if (c->return_tag == SCRIPTGO_TAG_NUMBER) {
        double (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (double (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        double value = fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload);
        result.tag = SCRIPTGO_TAG_NUMBER;
        memcpy(&result.payload, &value, sizeof(value));
    } else if (c->return_tag == SCRIPTGO_TAG_BOOLEAN) {
        _Bool (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (_Bool (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        result.tag = SCRIPTGO_TAG_BOOLEAN;
        result.payload = fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload) ? 1 : 0;
    } else if (c->return_tag == SCRIPTGO_TAG_BIGINT) {
        uint64_t (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (uint64_t (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        result.tag = SCRIPTGO_TAG_BIGINT;
        result.payload = fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload);
    } else if (c->return_tag >= SCRIPTGO_TAG_STRING && c->return_tag <= SCRIPTGO_TAG_SYMBOL) {
        void *(*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
            (void *(*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))c->fn_ptr;
        void *value = fn(c->env, v1->tag, v1->flags, v1->payload, v2->tag, v2->flags, v2->payload, v3->tag, v3->flags, v3->payload, v4->tag, v4->flags, v4->payload);
        result.tag = (uint32_t)c->return_tag;
        result.payload = (uint64_t)(uintptr_t)value;
        if (c->return_tag == SCRIPTGO_TAG_STRING && value != NULL && value != &scriptgo_undefined_sentinel) {
            result.aux = strlen((const char *)value);
        }
    } else {
        return scriptgo_runtime_set_error("scriptgo closure has invalid return tag");
    }
    *out_value = result;
    return 0;
}

int scriptgo_closure_invoke(void *closure_handle, int32_t arg_count, const scriptgo_value *a1, const scriptgo_value *a2, const scriptgo_value *a3, const scriptgo_value *a4) {
    scriptgo_value result = {0};
    int status = scriptgo_closure_invoke_value(closure_handle, arg_count, a1, a2, a3, a4, &result);
    if (status == 0) scriptgo_value_release(&result);
    return status;
}

int scriptgo_array_map_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        double mapped = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(double), &mapped, sizeof(double));
    }
    return 0;
}

int scriptgo_array_push(void *handle, const void *value, double *out_length);

int scriptgo_array_set_tag(void *handle, int64_t tag);

int scriptgo_array_flat_map_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array flatMap failed");
    }
    if (scriptgo_array_new(0, sizeof(double), out_array) != 0) {
        return -1;
    }
    scriptgo_array_set_tag(*out_array, 3);
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        void *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        void *res_ptr = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        double out_len;
        if (res_ptr != NULL) {
            scriptgo_array_inner *sub = res_ptr;
            if (sub->element_size == sizeof(double)) {
                for (int64_t j = 0; j < sub->length; j++) {
                    double v = *(double *)(sub->data + (size_t)j * sizeof(double));
                    scriptgo_array_push(*out_array, &v, &out_len);
                }
            } else {
                double v = (double)(uintptr_t)res_ptr;
                scriptgo_array_push(*out_array, &v, &out_len);
            }
        }
    }
    return 0;
}

int scriptgo_array_flat_map_number_scalar(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array flatMap failed");
    }
    if (scriptgo_array_new(0, sizeof(double), out_array) != 0) {
        return -1;
    }
    scriptgo_array_set_tag(*out_array, 3);
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        double res_val = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        double out_len;
        scriptgo_array_push(*out_array, &res_val, &out_len);
    }
    return 0;
}

int scriptgo_array_map_number_from_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        double mapped = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(double), &mapped, sizeof(double));
    }
    return 0;
}

int scriptgo_array_map_number_from_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(double), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        double mapped = fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(double), &mapped, sizeof(double));
    }
    return 0;
}

int scriptgo_array_map_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        char *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (char *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        char *mapped = fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(char *), &mapped, sizeof(char *));
    }
    return 0;
}

int scriptgo_array_map_string_from_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        char *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (char *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        char *mapped = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(char *), &mapped, sizeof(char *));
    }
    return 0;
}

int scriptgo_array_map_string_from_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(char *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        char *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (char *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        char *mapped = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(char *), &mapped, sizeof(char *));
    }
    return 0;
}

int scriptgo_array_map_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(void *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    void *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (void *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
    if (array->element_size == sizeof(scriptgo_value)) {
        typedef scriptgo_value boxed_val;
        for (int64_t i = 0; i < array->length; i++) {
            boxed_val item = *(boxed_val *)(array->data + (size_t)i * sizeof(scriptgo_value));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            void *mapped = fn(c->env, item.tag, item.flags, item.payload, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
            memcpy(res->data + (size_t)i * sizeof(void *), &mapped, sizeof(void *));
        }
    } else {
        for (int64_t i = 0; i < array->length; i++) {
            void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            void *mapped = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
            memcpy(res->data + (size_t)i * sizeof(void *), &mapped, sizeof(void *));
        }
    }
    return 0;
}

int scriptgo_array_map_ptr_from_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(void *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        void *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        void *mapped = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(void *), &mapped, sizeof(void *));
    }
    return 0;
}

int scriptgo_array_map_ptr_from_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(void *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        void *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        void *mapped = fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        memcpy(res->data + (size_t)i * sizeof(void *), &mapped, sizeof(void *));
    }
    return 0;
}

int scriptgo_array_filter_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array filter failed");
    }
    if (scriptgo_array_new(0, sizeof(double), out_array) != 0) {
        return -1;
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        uint8_t keep = fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        if (keep) {
            double dummy;
            if (scriptgo_array_push(*out_array, &item, &dummy) != 0) return -1;
        }
    }
    return 0;
}

int scriptgo_array_filter_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(char *)) {
        return scriptgo_runtime_set_error("scriptgo array filter failed");
    }
    if (scriptgo_array_new(0, sizeof(char *), out_array) != 0) {
        return -1;
    }
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        uint8_t keep = fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        if (keep) {
            double dummy;
            if (scriptgo_array_push(*out_array, &item, &dummy) != 0) return -1;
        }
    }
    return 0;
}

int scriptgo_array_filter_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_array == NULL) {
        return scriptgo_runtime_set_error("scriptgo array filter failed");
    }
    if (array->element_size != sizeof(void *) && array->element_size != sizeof(scriptgo_value)) {
        return scriptgo_runtime_set_error("scriptgo array filter failed");
    }
    if (scriptgo_array_new(0, array->element_size, out_array) != 0) {
        return -1;
    }
    uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;

    if (array->element_size == sizeof(scriptgo_value)) {
        typedef scriptgo_value boxed_val;
        for (int64_t i = 0; i < array->length; i++) {
            boxed_val item = *(boxed_val *)(array->data + (size_t)i * sizeof(scriptgo_value));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            uint8_t keep = fn(c->env, item.tag, item.flags, item.payload, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
            if (keep) {
                double dummy;
                if (scriptgo_array_push(*out_array, &item, &dummy) != 0) return -1;
            }
        }
    } else {
        for (int64_t i = 0; i < array->length; i++) {
            void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            uint8_t keep = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
            if (keep) {
                double dummy;
                if (scriptgo_array_push(*out_array, &item, &dummy) != 0) return -1;
            }
        }
    }
    return 0;
}

int scriptgo_array_for_each_number(void *handle, void *closure_handle) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array forEach failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        void (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
    }
    return 0;
}

int scriptgo_array_for_each_string(void *handle, void *closure_handle) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || array->element_size != sizeof(char *)) {
        return scriptgo_runtime_set_error("scriptgo array forEach failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        void (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
    }
    return 0;
}

int scriptgo_array_for_each_ptr(void *handle, void *closure_handle) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL) {
        return scriptgo_runtime_set_error("scriptgo array forEach failed");
    }
    if (array->element_size != sizeof(void *) && array->element_size != sizeof(scriptgo_value)) {
        return scriptgo_runtime_set_error("scriptgo array forEach failed");
    }
    void (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (void (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
    if (array->element_size == sizeof(scriptgo_value)) {
        typedef scriptgo_value boxed_val;
        for (int64_t i = 0; i < array->length; i++) {
            boxed_val item = *(boxed_val *)(array->data + (size_t)i * sizeof(scriptgo_value));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            fn(c->env, item.tag, item.flags, item.payload, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        }
    } else {
        for (int64_t i = 0; i < array->length; i++) {
            void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
            union { double d; int64_t i; } u_idx;
            u_idx.d = (double)i;
            fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        }
    }
    return 0;
}


int scriptgo_array_reduce_number(void *handle, void *closure_handle, double init_val, double *out_res) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    double acc = init_val;
    if (array == NULL || c == NULL || out_res == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array reduce failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_acc, u_item, u_idx;
        u_acc.d = acc;
        u_item.d = item;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        acc = fn(c->env, 3, 0, u_acc.i, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0);
    }
    *out_res = acc;
    return 0;
}

int scriptgo_array_find_number(void *handle, void *closure_handle, double *out_val) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_val == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array find failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_val = item;
            return 0;
        }
    }
    *out_val = 0.0;
    return 0;
}

int scriptgo_array_some_number(void *handle, void *closure_handle, int32_t *out_bool) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_bool == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array some failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_bool = 1;
            return 0;
        }
    }
    *out_bool = 0;
    return 0;
}

int scriptgo_array_every_number(void *handle, void *closure_handle, int32_t *out_bool) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_bool == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array every failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (!fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_bool = 0;
            return 0;
        }
    }
    *out_bool = 1;
    return 0;
}

int scriptgo_array_find_index_number(void *handle, void *closure_handle, double *out_idx) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_idx == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array findIndex failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_idx = (double)i;
            return 0;
        }
    }
    *out_idx = -1.0;
    return 0;
}

int scriptgo_array_find_index_string(void *handle, void *closure_handle, double *out_idx) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_idx == NULL || array->element_size != sizeof(char *)) {
        return scriptgo_runtime_set_error("scriptgo array findIndex failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_idx = (double)i;
            return 0;
        }
    }
    *out_idx = -1.0;
    return 0;
}

int scriptgo_array_find_last_number(void *handle, void *closure_handle, double *out_val) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_val == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array findLast failed");
    }
    for (int64_t i = array->length - 1; i >= 0; i--) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_val = item;
            return 0;
        }
    }
    *out_val = 0.0;
    return 0;
}

int scriptgo_array_find_last_index_number(void *handle, void *closure_handle, double *out_idx) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_idx == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array findLastIndex failed");
    }
    for (int64_t i = array->length - 1; i >= 0; i--) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_item, u_idx;
        u_item.d = item;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_idx = (double)i;
            return 0;
        }
    }
    *out_idx = -1.0;
    return 0;
}

int scriptgo_array_find_last_index_string(void *handle, void *closure_handle, double *out_idx) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || c == NULL || out_idx == NULL || array->element_size != sizeof(char *)) {
        return scriptgo_runtime_set_error("scriptgo array findLastIndex failed");
    }
    for (int64_t i = array->length - 1; i >= 0; i--) {
        char *item = *(char **)(array->data + (size_t)i * sizeof(char *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        if (fn(c->env, 4, 0, (int64_t)(uintptr_t)(item ? item : ""), 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0)) {
            *out_idx = (double)i;
            return 0;
        }
    }
    *out_idx = -1.0;
    return 0;
}

int scriptgo_array_reduce_right_number(void *handle, void *closure_handle, double init_val, double *out_res) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    double acc = init_val;
    if (array == NULL || c == NULL || out_res == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array reduceRight failed");
    }
    for (int64_t i = array->length - 1; i >= 0; i--) {
        double item = *(double *)(array->data + (size_t)i * sizeof(double));
        union { double d; int64_t i; } u_acc, u_item, u_idx;
        u_acc.d = acc;
        u_item.d = item;
        u_idx.d = (double)i;
        double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        acc = fn(c->env, 3, 0, u_acc.i, 3, 0, u_item.i, 3, 0, u_idx.i, 0, 0, 0);
    }
    *out_res = acc;
    return 0;
}

int scriptgo_array_sort_number(void *handle, void **out_array);
int scriptgo_array_sort_string(void *handle, void **out_array);

int scriptgo_array_sort_closure_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || array->element_size != sizeof(void *)) {
        return scriptgo_runtime_set_error("scriptgo array sort failed");
    }
    if (c == NULL) {
        if (out_array != NULL) *out_array = array;
        return 0;
    }
    double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;

    void **items = (void **)array->data;
    int64_t n = array->length;
    for (int64_t i = 1; i < n; i++) {
        void *key = items[i];
        int64_t j = i - 1;
        while (j >= 0) {
            double cmp = fn(c->env, 5, 0, (int64_t)items[j], 5, 0, (int64_t)key, 0, 0, 0, 0, 0, 0);
            if (cmp > 0.0) {
                items[j + 1] = items[j];
                j--;
            } else {
                break;
            }
        }
        items[j + 1] = key;
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_sort_closure_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || array->element_size != sizeof(double)) {
        return scriptgo_runtime_set_error("scriptgo array sort failed");
    }
    if (c == NULL) {
        return scriptgo_array_sort_number(handle, out_array);
    }
    double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;

    double *items = (double *)array->data;
    int64_t n = array->length;
    for (int64_t i = 1; i < n; i++) {
        double key = items[i];
        union { double d; int64_t i; } u_key;
        u_key.d = key;
        int64_t j = i - 1;
        while (j >= 0) {
            union { double d; int64_t i; } u_j;
            u_j.d = items[j];
            double cmp = fn(c->env, 3, 0, u_j.i, 3, 0, u_key.i, 0, 0, 0, 0, 0, 0);
            if (cmp > 0.0) {
                items[j + 1] = items[j];
                j--;
            } else {
                break;
            }
        }
        items[j + 1] = key;
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}

int scriptgo_array_sort_closure_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    if (array == NULL || array->element_size != sizeof(char *)) {
        return scriptgo_runtime_set_error("scriptgo array sort failed");
    }
    if (c == NULL) {
        return scriptgo_array_sort_string(handle, out_array);
    }
    double (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (double (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;

    char **items = (char **)array->data;
    int64_t n = array->length;
    for (int64_t i = 1; i < n; i++) {
        char *key = items[i];
        int64_t j = i - 1;
        while (j >= 0) {
            double cmp = fn(c->env, 4, 0, (int64_t)items[j], 4, 0, (int64_t)key, 0, 0, 0, 0, 0, 0);
            if (cmp > 0.0) {
                items[j + 1] = items[j];
                j--;
            } else {
                break;
            }
        }
        items[j + 1] = key;
    }
    if (out_array != NULL) {
        *out_array = array;
    }
    return 0;
}
