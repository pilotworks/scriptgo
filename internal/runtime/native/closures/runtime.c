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

typedef struct {
    int32_t tag;
    int32_t pad;
    int64_t payload;
} scriptgo_boxed_value;

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
    if (arg_count == 0) {
        void (*fn)(void *) = (void (*)(void *))c->fn_ptr;
        fn(c->env);
    } else if (arg_count == 1) {
        if (a1 == NULL) return 0;
        if (a1->tag == 3) {
            double num;
            memcpy(&num, &a1->payload, sizeof(double));
            void (*fn)(void *, double) = (void (*)(void *, double))c->fn_ptr;
            fn(c->env, num);
        } else {
            void *ptr = (void *)(uintptr_t)a1->payload;
            void (*fn)(void *, void *) = (void (*)(void *, void *))c->fn_ptr;
            fn(c->env, ptr);
        }
    } else if (arg_count == 2) {
        if (a1 == NULL || a2 == NULL) return 0;
        if (a1->tag == 3 && a2->tag == 3) {
            double n1, n2;
            memcpy(&n1, &a1->payload, sizeof(double));
            memcpy(&n2, &a2->payload, sizeof(double));
            void (*fn)(void *, double, double) = (void (*)(void *, double, double))c->fn_ptr;
            fn(c->env, n1, n2);
        } else if (a1->tag == 3 && a2->tag != 3) {
            double n1;
            void *p2 = (void *)(uintptr_t)a2->payload;
            memcpy(&n1, &a1->payload, sizeof(double));
            void (*fn)(void *, double, void *) = (void (*)(void *, double, void *))c->fn_ptr;
            fn(c->env, n1, p2);
        } else if (a1->tag != 3 && a2->tag == 3) {
            void *p1 = (void *)(uintptr_t)a1->payload;
            double n2;
            memcpy(&n2, &a2->payload, sizeof(double));
            void (*fn)(void *, void *, double) = (void (*)(void *, void *, double))c->fn_ptr;
            fn(c->env, p1, n2);
        } else {
            void *p1 = (void *)(uintptr_t)a1->payload;
            void *p2 = (void *)(uintptr_t)a2->payload;
            void (*fn)(void *, void *, void *) = (void (*)(void *, void *, void *))c->fn_ptr;
            fn(c->env, p1, p2);
        }
    }
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
        double (*fn)(void *, double, double) = (double (*)(void *, double, double))c->fn_ptr;
        double mapped = fn(c->env, item, (double)i);
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
        char *(*fn)(void *, char *, double) = (char *(*)(void *, char *, double))c->fn_ptr;
        char *mapped = fn(c->env, item, (double)i);
        memcpy(res->data + (size_t)i * sizeof(char *), &mapped, sizeof(char *));
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        uint8_t keep = fn(c->env, item, (double)i);
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
        uint8_t (*fn)(void *, char *, double) = (uint8_t (*)(void *, char *, double))c->fn_ptr;
        uint8_t keep = fn(c->env, item, (double)i);
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
        void (*fn)(void *, double, double) = (void (*)(void *, double, double))c->fn_ptr;
        fn(c->env, item, (double)i);
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
        void (*fn)(void *, char *, double) = (void (*)(void *, char *, double))c->fn_ptr;
        fn(c->env, item, (double)i);
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
        double (*fn)(void *, double, double, double) = (double (*)(void *, double, double, double))c->fn_ptr;
        acc = fn(c->env, acc, item, (double)i);
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (!fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, char *, double) = (uint8_t (*)(void *, char *, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, double, double) = (uint8_t (*)(void *, double, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        uint8_t (*fn)(void *, char *, double) = (uint8_t (*)(void *, char *, double))c->fn_ptr;
        if (fn(c->env, item, (double)i)) {
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
        double (*fn)(void *, double, double, double) = (double (*)(void *, double, double, double))c->fn_ptr;
        acc = fn(c->env, acc, item, (double)i);
    }
    *out_res = acc;
    return 0;
}

