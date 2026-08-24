#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_push(void *handle, const void *value, double *out_length);

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
} scriptgo_closure;

int scriptgo_closure_create(void *fn_ptr, void *env, void **out_closure) {
    scriptgo_closure *c;
    if (out_closure == NULL) return scriptgo_runtime_set_error("scriptgo closure allocation failed");
    c = malloc(sizeof(scriptgo_closure));
    if (c == NULL) return scriptgo_runtime_set_error("scriptgo closure allocation failed");
    c->fn_ptr = fn_ptr;
    c->env = env;
    *out_closure = c;
    return 0;
}

int scriptgo_closure_invoke(void *closure_handle, int32_t arg_count, const scriptgo_boxed_value *a1, const scriptgo_boxed_value *a2, const scriptgo_boxed_value *a3, const scriptgo_boxed_value *a4) {
    scriptgo_closure *c = closure_handle;
    if (c == NULL || c->fn_ptr == NULL) return 0;
    scriptgo_boxed_value dummy = {0};
    const scriptgo_boxed_value *v1 = (a1 != NULL && arg_count >= 1) ? a1 : &dummy;
    const scriptgo_boxed_value *v2 = (a2 != NULL && arg_count >= 2) ? a2 : &dummy;
    const scriptgo_boxed_value *v3 = (a3 != NULL && arg_count >= 3) ? a3 : &dummy;
    const scriptgo_boxed_value *v4 = (a4 != NULL && arg_count >= 4) ? a4 : &dummy;
    void (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
        (void (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
    fn(c->env, v1->tag, v1->pad, v1->payload, v2->tag, v2->pad, v2->payload, v3->tag, v3->pad, v3->payload, v4->tag, v4->pad, v4->payload);
    return 0;
}

int scriptgo_array_map_number(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(double)) {
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

int scriptgo_array_map_string(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(char *)) {
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

int scriptgo_array_map_ptr(void *handle, void *closure_handle, void **out_array) {
    scriptgo_array_inner *array = handle;
    scriptgo_closure *c = closure_handle;
    scriptgo_array_inner *res;
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(void *)) {
        return scriptgo_runtime_set_error("scriptgo array map failed");
    }
    if (scriptgo_array_new(array->length, sizeof(void *), out_array) != 0) {
        return -1;
    }
    res = *out_array;
    for (int64_t i = 0; i < array->length; i++) {
        void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        void *(*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void *(*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        void *mapped = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
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
    if (array == NULL || c == NULL || out_array == NULL || array->element_size != sizeof(void *)) {
        return scriptgo_runtime_set_error("scriptgo array filter failed");
    }
    if (scriptgo_array_new(0, sizeof(void *), out_array) != 0) {
        return -1;
    }
    for (int64_t i = 0; i < array->length; i++) {
        void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        uint8_t (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (uint8_t (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        uint8_t keep = fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
        if (keep) {
            double dummy;
            if (scriptgo_array_push(*out_array, &item, &dummy) != 0) return -1;
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
    if (array == NULL || c == NULL || array->element_size != sizeof(void *)) {
        return scriptgo_runtime_set_error("scriptgo array forEach failed");
    }
    for (int64_t i = 0; i < array->length; i++) {
        void *item = *(void **)(array->data + (size_t)i * sizeof(void *));
        union { double d; int64_t i; } u_idx;
        u_idx.d = (double)i;
        void (*fn)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t) =
            (void (*)(void *, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t, int32_t, int32_t, int64_t))c->fn_ptr;
        fn(c->env, 5, 0, (int64_t)(uintptr_t)item, 3, 0, u_idx.i, 0, 0, 0, 0, 0, 0);
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

