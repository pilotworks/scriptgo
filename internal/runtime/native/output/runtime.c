#include <stdio.h>

int scriptgo_runtime_set_error(const char *message);

int scriptgo_print_number(double value) {
    if (printf("%g\n", value) < 0) return scriptgo_runtime_set_error("scriptgo number output failed");
    return 0;
}

int scriptgo_print_string(const char *value) {
    if (value == NULL || puts(value) == EOF) return scriptgo_runtime_set_error("scriptgo string output failed");
    return 0;
}

int scriptgo_print_bool(int value) {
    if (puts(value ? "true" : "false") == EOF) return scriptgo_runtime_set_error("scriptgo boolean output failed");
    return 0;
}
