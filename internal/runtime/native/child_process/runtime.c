#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <stdint.h>
#include <errno.h>

int scriptgo_runtime_set_error(const char *message);

typedef struct {
    int64_t length;
    int64_t capacity;
    int64_t element_size;
    void *data;
} scriptgo_array_raw_cp;

static int cp_fail(const char *message) { return scriptgo_runtime_set_error(message); }

static char *read_all_fd(int fd) {
    size_t cap = 4096;
    size_t len = 0;
    char *buf = malloc(cap);
    if (buf == NULL) return NULL;

    while (1) {
        if (len + 1024 > cap) {
            cap *= 2;
            char *new_buf = realloc(buf, cap);
            if (new_buf == NULL) {
                free(buf);
                return NULL;
            }
            buf = new_buf;
        }
        ssize_t n = read(fd, buf + len, cap - len - 1);
        if (n <= 0) {
            break;
        }
        len += (size_t)n;
    }
    buf[len] = '\0';
    return buf;
}

int scriptgo_child_process_exec_sync(const char *command, const char *cwd, const char *input,
                                     char **out_stdout, char **out_stderr, double *out_status) {
    if (command == NULL || out_stdout == NULL || out_stderr == NULL || out_status == NULL) {
        return cp_fail("scriptgo child_process invalid arguments");
    }

    int stdout_pipe[2];
    int stderr_pipe[2];
    int stdin_pipe[2];

    if (pipe(stdout_pipe) != 0 || pipe(stderr_pipe) != 0 || pipe(stdin_pipe) != 0) {
        return cp_fail("scriptgo child_process pipe creation failed");
    }

    pid_t pid = fork();
    if (pid < 0) {
        return cp_fail("scriptgo child_process fork failed");
    }

    if (pid == 0) {
        // Child process
        close(stdout_pipe[0]);
        close(stderr_pipe[0]);
        close(stdin_pipe[1]);

        dup2(stdout_pipe[1], STDOUT_FILENO);
        dup2(stderr_pipe[1], STDERR_FILENO);
        dup2(stdin_pipe[0], STDIN_FILENO);

        close(stdout_pipe[1]);
        close(stderr_pipe[1]);
        close(stdin_pipe[0]);

        if (cwd != NULL && strlen(cwd) > 0) {
            if (chdir(cwd) != 0) {
                _exit(127);
            }
        }

        execl("/bin/sh", "sh", "-c", command, (char *)NULL);
        _exit(127);
    }

    // Parent process
    close(stdout_pipe[1]);
    close(stderr_pipe[1]);
    close(stdin_pipe[0]);

    if (input != NULL && strlen(input) > 0) {
        write(stdin_pipe[1], input, strlen(input));
    }
    close(stdin_pipe[1]);

    char *stdout_str = read_all_fd(stdout_pipe[0]);
    char *stderr_str = read_all_fd(stderr_pipe[0]);
    close(stdout_pipe[0]);
    close(stderr_pipe[0]);

    int status = 0;
    waitpid(pid, &status, 0);

    if (stdout_str == NULL) stdout_str = strdup("");
    if (stderr_str == NULL) stderr_str = strdup("");

    *out_stdout = stdout_str;
    *out_stderr = stderr_str;
    if (WIFEXITED(status)) {
        *out_status = (double)WEXITSTATUS(status);
    } else {
        *out_status = 1.0;
    }
    return 0;
}

int scriptgo_child_process_spawn_sync(const char *command, void *args_handle, const char *cwd, const char *input,
                                      char **out_stdout, char **out_stderr, double *out_status) {
    if (command == NULL || out_stdout == NULL || out_stderr == NULL || out_status == NULL) {
        return cp_fail("scriptgo child_process spawn invalid arguments");
    }

    // Prepare argv
    int argc = 1;
    scriptgo_array_raw_cp *arr = (scriptgo_array_raw_cp *)args_handle;
    if (arr != NULL && arr->data != NULL) {
        argc += (int)arr->length;
    }

    char **argv = malloc(sizeof(char *) * (argc + 1));
    if (argv == NULL) {
        return cp_fail("scriptgo child_process argv allocation failed");
    }
    argv[0] = (char *)command;
    if (arr != NULL && arr->data != NULL) {
        char **entries = (char **)arr->data;
        for (int i = 0; i < (int)arr->length; i++) {
            argv[i + 1] = entries[i];
        }
    }
    argv[argc] = NULL;

    int stdout_pipe[2];
    int stderr_pipe[2];
    int stdin_pipe[2];

    if (pipe(stdout_pipe) != 0 || pipe(stderr_pipe) != 0 || pipe(stdin_pipe) != 0) {
        free(argv);
        return cp_fail("scriptgo child_process pipe creation failed");
    }

    pid_t pid = fork();
    if (pid < 0) {
        free(argv);
        return cp_fail("scriptgo child_process fork failed");
    }

    if (pid == 0) {
        // Child process
        close(stdout_pipe[0]);
        close(stderr_pipe[0]);
        close(stdin_pipe[1]);

        dup2(stdout_pipe[1], STDOUT_FILENO);
        dup2(stderr_pipe[1], STDERR_FILENO);
        dup2(stdin_pipe[0], STDIN_FILENO);

        close(stdout_pipe[1]);
        close(stderr_pipe[1]);
        close(stdin_pipe[0]);

        if (cwd != NULL && strlen(cwd) > 0) {
            if (chdir(cwd) != 0) {
                _exit(127);
            }
        }

        execvp(command, argv);
        _exit(127);
    }

    // Parent process
    free(argv);
    close(stdout_pipe[1]);
    close(stderr_pipe[1]);
    close(stdin_pipe[0]);

    if (input != NULL && strlen(input) > 0) {
        write(stdin_pipe[1], input, strlen(input));
    }
    close(stdin_pipe[1]);

    char *stdout_str = read_all_fd(stdout_pipe[0]);
    char *stderr_str = read_all_fd(stderr_pipe[0]);
    close(stdout_pipe[0]);
    close(stderr_pipe[0]);

    int status = 0;
    waitpid(pid, &status, 0);

    if (stdout_str == NULL) stdout_str = strdup("");
    if (stderr_str == NULL) stderr_str = strdup("");

    *out_stdout = stdout_str;
    *out_stderr = stderr_str;
    if (WIFEXITED(status)) {
        *out_status = (double)WEXITSTATUS(status);
    } else {
        *out_status = 1.0;
    }
    return 0;
}
