#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set(void *handle, double index, const void *value);

static int process_fail(const char *message) { return scriptgo_runtime_set_error(message); }

static int g_argc = 0;
static char **g_argv = NULL;

void scriptgo_process_init(int argc, char **argv) {
    g_argc = argc;
    g_argv = argv;
}

int scriptgo_process_exit(double code) {
    exit((int)code);
    return 0;
}

int scriptgo_process_cwd(char **out_cwd) {
    char buf[4096];
    char *res;
    if (out_cwd == NULL) return process_fail("scriptgo process invalid arguments");
    if (getcwd(buf, sizeof(buf)) == NULL) {
        return process_fail("scriptgo process getcwd failed");
    }
    res = strdup(buf);
    if (res == NULL) return process_fail("scriptgo process allocation failed");
    *out_cwd = res;
    return 0;
}

int scriptgo_process_argv(void **out_array) {
    if (out_array == NULL) return process_fail("scriptgo process invalid arguments");
    if (scriptgo_array_new((int64_t)g_argc, sizeof(char *), out_array) != 0) {
        return -1;
    }
    for (int i = 0; i < g_argc; i++) {
        char *arg = strdup(g_argv[i]);
        if (arg == NULL) return process_fail("scriptgo process allocation failed");
        if (scriptgo_array_set(*out_array, (double)i, &arg) != 0) {
            return -1;
        }
    }
    return 0;
}

int scriptgo_process_env(const char *key, char **out_value) {
    const char *val;
    char *res;
    if (key == NULL || out_value == NULL) return process_fail("scriptgo process invalid arguments");
    val = getenv(key);
    if (val == NULL) {
        val = "";
    }
    res = strdup(val);
    if (res == NULL) return process_fail("scriptgo process allocation failed");
    *out_value = res;
    return 0;
}

int scriptgo_process_pid(double *out_pid) {
    if (out_pid == NULL) return process_fail("scriptgo process invalid arguments");
    *out_pid = (double)getpid();
    return 0;
}

int scriptgo_process_ppid(double *out_ppid) {
    if (out_ppid == NULL) return process_fail("scriptgo process invalid arguments");
    *out_ppid = (double)getppid();
    return 0;
}

int scriptgo_process_version(char **out_version) {
    if (out_version == NULL) return process_fail("scriptgo process invalid arguments");
    *out_version = strdup("v0.1.0");
    return 0;
}
