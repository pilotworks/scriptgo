#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

int scriptgo_array_new(int64_t, int64_t, void **);
int scriptgo_array_set(void *, double, const void *);
int scriptgo_array_index_of_number(void *, double, double, double *);
int scriptgo_array_release(void *);
int scriptgo_string_split(const char *, const char *, double, void **);
int scriptgo_web_btoa(const char *, char **);
int scriptgo_web_atob(const char *, char **);

static uint64_t now_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

static void print_result(const char *name, int64_t operations, uint64_t elapsed, uint64_t checksum) {
    printf("%s operations=%lld ns/op=%.2f checksum=%llu\n",
        name, (long long)operations, (double)elapsed / (double)operations,
        (unsigned long long)checksum);
}

static int bench_array_index(void) {
    const int64_t length = 4096;
    const int64_t operations = 1000000;
    void *array = NULL;
    double target = 0, result = 0;
    uint64_t checksum = 0;
    if (scriptgo_array_new(length, sizeof(double), &array) != 0) return 1;
    for (int64_t i = 0; i < length; i++) {
        double value = (double)i;
        if (scriptgo_array_set(array, (double)i, &value) != 0) return 2;
    }
    target = (double)(length - 1);
    uint64_t start = now_ns();
    for (int64_t i = 0; i < operations; i++) {
        if (scriptgo_array_index_of_number(array, target, 0, &result) != 0) return 3;
        checksum += (uint64_t)result;
    }
    print_result("array.indexOf(number)", operations, now_ns() - start, checksum);
    scriptgo_array_release(array);
    return 0;
}

static int bench_string_split(void) {
    const int64_t operations = 100000;
    const char *value = "alpha,beta,gamma,delta,epsilon,zeta,eta,theta,iota,kappa";
    void *array = NULL;
    uint64_t checksum = 0;
    uint64_t start = now_ns();
    for (int64_t i = 0; i < operations; i++) {
        if (scriptgo_string_split(value, ",", -1.0, &array) != 0) return 1;
        checksum += 10;
        scriptgo_array_release(array);
    }
    print_result("string.split", operations, now_ns() - start, checksum);
    return 0;
}

static int bench_btoa(void) {
    const int64_t operations = 100000;
    char input[1025];
    char *encoded = NULL;
    uint64_t checksum = 0;
    for (size_t i = 0; i < sizeof(input) - 1; i++) input[i] = (char)('A' + (i % 26));
    input[sizeof(input) - 1] = '\0';
    uint64_t start = now_ns();
    for (int64_t i = 0; i < operations; i++) {
        if (scriptgo_web_btoa(input, &encoded) != 0) return 1;
        checksum += (unsigned char)encoded[0] + (unsigned char)encoded[100];
        free(encoded);
    }
    print_result("web.btoa", operations, now_ns() - start, checksum);
    return 0;
}

static int bench_atob(void) {
    const int64_t operations = 100000;
    const char *input = "QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVo=";
    char *decoded = NULL;
    uint64_t checksum = 0;
    uint64_t start = now_ns();
    for (int64_t i = 0; i < operations; i++) {
        if (scriptgo_web_atob(input, &decoded) != 0) return 1;
        checksum += (unsigned char)decoded[0] + (unsigned char)decoded[20];
        free(decoded);
    }
    print_result("web.atob", operations, now_ns() - start, checksum);
    return 0;
}

int main(void) {
    if (bench_array_index() != 0) return 1;
    if (bench_string_split() != 0) return 2;
    if (bench_btoa() != 0) return 3;
    if (bench_atob() != 0) return 4;
    return 0;
}
