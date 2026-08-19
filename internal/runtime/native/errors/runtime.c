#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char scriptgo_runtime_error[256];

int scriptgo_runtime_set_error(const char *message) {
    if (message == NULL) message = "scriptgo runtime error";
    strncpy(scriptgo_runtime_error, message, sizeof(scriptgo_runtime_error) - 1);
    scriptgo_runtime_error[sizeof(scriptgo_runtime_error) - 1] = '\0';
    return -1;
}

const char *scriptgo_runtime_last_error(void) { return scriptgo_runtime_error; }

void scriptgo_runtime_abort_if_failed(int status) {
    if (status != 0) {
        fputs(scriptgo_runtime_last_error(), stderr);
        fputc('\n', stderr);
        abort();
    }
}
