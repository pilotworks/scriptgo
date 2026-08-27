#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <pthread.h>

int scriptgo_runtime_set_error(const char *message);

static int atomics_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static pthread_mutex_t atomics_wait_mutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t atomics_wait_cond = PTHREAD_COND_INITIALIZER;

int scriptgo_shared_array_buffer_new(int64_t byte_length, void **out_buffer) {
    if (out_buffer == NULL) return atomics_fail("SharedArrayBuffer: null output buffer pointer");
    if (byte_length < 0) return atomics_fail("SharedArrayBuffer: invalid length");
    scriptgo_array_buffer *buf = (scriptgo_array_buffer *)malloc(sizeof(scriptgo_array_buffer));
    if (buf == NULL) return atomics_fail("SharedArrayBuffer allocation failed");
    buf->byte_length = byte_length;
    buf->data = (byte_length > 0) ? (unsigned char *)calloc(1, (size_t)byte_length) : NULL;
    *out_buffer = buf;
    return 0;
}

int scriptgo_atomics_is_lock_free(double size, int32_t *out_res) {
    if (out_res == NULL) return atomics_fail("Atomics.isLockFree: null result");
    int64_t s = (int64_t)size;
    *out_res = (s == 1 || s == 2 || s == 4 || s == 8) ? 1 : 0;
    return 0;
}

static void *get_element_ptr(void *handle, double index, int64_t expected_size) {
    if (handle == NULL) return NULL;
    scriptgo_typed_array *ta = (scriptgo_typed_array *)handle;
    if (ta->magic != SCRIPTGO_MAGIC_TYPEDARRAY) return NULL;
    int64_t idx = (int64_t)index;
    if (idx < 0 || idx >= ta->length) return NULL;
    if (ta->element_size != expected_size) return NULL;
    return (void *)(ta->data + (idx * ta->element_size));
}

int scriptgo_atomics_add(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.add: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old = __atomic_fetch_add(ptr, val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_sub(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.sub: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old = __atomic_fetch_sub(ptr, val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_and(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.and: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old = __atomic_fetch_and(ptr, val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_or(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.or: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old = __atomic_fetch_or(ptr, val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_xor(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.xor: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old = __atomic_fetch_xor(ptr, val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_load(void *handle, double index, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.load: invalid TypedArray or index out of range");
    int32_t val;
    __atomic_load(ptr, &val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)val;
    return 0;
}

int scriptgo_atomics_store(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.store: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    __atomic_store(ptr, &val, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)val;
    return 0;
}

int scriptgo_atomics_exchange(void *handle, double index, double value, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.exchange: invalid TypedArray or index out of range");
    int32_t val = (int32_t)value;
    int32_t old;
    __atomic_exchange(ptr, &val, &old, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)old;
    return 0;
}

int scriptgo_atomics_compare_exchange(void *handle, double index, double expected, double replacement, double *out_val) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.compareExchange: invalid TypedArray or index out of range");
    int32_t exp = (int32_t)expected;
    int32_t des = (int32_t)replacement;
    __atomic_compare_exchange(ptr, &exp, &des, 0, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST);
    if (out_val) *out_val = (double)exp;
    return 0;
}

int scriptgo_atomics_wait(void *handle, double index, double value, double timeout_ms, void **out_str) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.wait: invalid TypedArray or index out of range");
    int32_t expected = (int32_t)value;
    pthread_mutex_lock(&atomics_wait_mutex);
    int32_t current;
    __atomic_load(ptr, &current, __ATOMIC_SEQ_CST);
    if (current != expected) {
        pthread_mutex_unlock(&atomics_wait_mutex);
        *out_str = strdup("not-equal");
        return 0;
    }
    if (timeout_ms <= 0.0) {
        pthread_cond_wait(&atomics_wait_cond, &atomics_wait_mutex);
        pthread_mutex_unlock(&atomics_wait_mutex);
        *out_str = strdup("ok");
        return 0;
    }
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    long sec = (long)(timeout_ms / 1000.0);
    long nsec = (long)((timeout_ms - (double)(sec * 1000)) * 1000000.0);
    ts.tv_sec += sec;
    ts.tv_nsec += nsec;
    if (ts.tv_nsec >= 1000000000L) {
        ts.tv_sec += 1;
        ts.tv_nsec -= 1000000000L;
    }
    int rc = pthread_cond_timedwait(&atomics_wait_cond, &atomics_wait_mutex, &ts);
    pthread_mutex_unlock(&atomics_wait_mutex);
    if (rc != 0) {
        *out_str = strdup("timed-out");
        return 0;
    }
    *out_str = strdup("ok");
    return 0;
}

int scriptgo_atomics_notify(void *handle, double index, double count, double *out_woken) {
    int32_t *ptr = (int32_t *)get_element_ptr(handle, index, 4);
    if (ptr == NULL) return atomics_fail("Atomics.notify: invalid TypedArray or index out of range");
    pthread_mutex_lock(&atomics_wait_mutex);
    pthread_cond_broadcast(&atomics_wait_cond);
    pthread_mutex_unlock(&atomics_wait_mutex);
    if (out_woken) *out_woken = (count > 0) ? count : 1.0;
    return 0;
}
