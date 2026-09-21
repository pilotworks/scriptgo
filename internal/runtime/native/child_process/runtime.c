#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdint.h>
#include <errno.h>

#if !defined(__wasi__)
#include <sys/wait.h>
#include <sys/types.h>
#include <fcntl.h>
#include <signal.h>
#endif

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

#if defined(__wasi__)
int scriptgo_child_process_exec_sync(const char *command, const char *cwd, const char *input,
                                     char **out_stdout, char **out_stderr, double *out_status) {
    if (out_stdout) *out_stdout = strdup("");
    if (out_stderr) *out_stderr = strdup("child_process is not supported on WebAssembly/WASI");
    if (out_status) *out_status = 1.0;
    return cp_fail("child_process is not supported on WebAssembly/WASI");
}

int scriptgo_child_process_spawn_sync(const char *command, void *args_handle, const char *cwd,
                                      const char *input, char **out_stdout, char **out_stderr, double *out_status) {
    if (out_stdout) *out_stdout = strdup("");
    if (out_stderr) *out_stderr = strdup("child_process is not supported on WebAssembly/WASI");
    if (out_status) *out_status = 1.0;
    return cp_fail("child_process is not supported on WebAssembly/WASI");
}

int scriptgo_child_process_exec_file_sync(const char *command, void *args_handle, const char *cwd,
                                          const char *input, char **out_stdout, char **out_stderr, double *out_status) {
    if (out_stdout) *out_stdout = strdup("");
    if (out_stderr) *out_stderr = strdup("child_process is not supported on WebAssembly/WASI");
    if (out_status) *out_status = 1.0;
    return cp_fail("child_process is not supported on WebAssembly/WASI");
}

int scriptgo_child_process_spawn_async(const char *command, void *args_handle, const char *cwd,
                                       double *out_pid, double *out_stdin_fd, double *out_stdout_fd, double *out_stderr_fd) {
    if (out_pid) *out_pid = -1.0;
    if (out_stdin_fd) *out_stdin_fd = -1.0;
    if (out_stdout_fd) *out_stdout_fd = -1.0;
    if (out_stderr_fd) *out_stderr_fd = -1.0;
    return cp_fail("child_process is not supported on WebAssembly/WASI");
}

int scriptgo_child_process_poll_status(double pid, double *out_status, double *out_exited) {
    if (out_status) *out_status = 1.0;
    if (out_exited) *out_exited = 1.0;
    return 0;
}

int scriptgo_child_process_pipe_read(double fd, double max_len, char **out_data, double *out_bytes_read, double *out_eof) {
    if (out_data) *out_data = strdup("");
    if (out_bytes_read) *out_bytes_read = 0.0;
    if (out_eof) *out_eof = 1.0;
    return 0;
}

int scriptgo_child_process_pipe_write(double fd, const char *data, double len, double *out_written) {
    if (out_written) *out_written = 0.0;
    return 0;
}

int scriptgo_child_process_pipe_close(double fd) {
    return 0;
}

int scriptgo_child_process_kill(double pid, double sig, double *out_success) {
    if (out_success) *out_success = 0.0;
    return 0;
}
#else
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

int scriptgo_child_process_spawn_async(const char *command, void *args_handle, const char *cwd,
                                       double *out_pid, double *out_stdin_fd, double *out_stdout_fd, double *out_stderr_fd) {
    if (command == NULL || out_pid == NULL || out_stdin_fd == NULL || out_stdout_fd == NULL || out_stderr_fd == NULL) {
        return cp_fail("spawn_async invalid arguments");
    }

    int argc = 1;
    scriptgo_array_raw_cp *arr = (scriptgo_array_raw_cp *)args_handle;
    if (arr != NULL && arr->data != NULL) {
        argc += (int)arr->length;
    }

    char **argv = malloc(sizeof(char *) * (argc + 1));
    if (argv == NULL) {
        return cp_fail("spawn_async argv allocation failed");
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
        return cp_fail("spawn_async pipe creation failed");
    }

    pid_t pid = fork();
    if (pid < 0) {
        free(argv);
        close(stdout_pipe[0]); close(stdout_pipe[1]);
        close(stderr_pipe[0]); close(stderr_pipe[1]);
        close(stdin_pipe[0]); close(stdin_pipe[1]);
        return cp_fail("spawn_async fork failed");
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

    // Set non-blocking on stdout and stderr read ends
    int flags = fcntl(stdout_pipe[0], F_GETFL, 0);
    fcntl(stdout_pipe[0], F_SETFL, flags | O_NONBLOCK);
    flags = fcntl(stderr_pipe[0], F_GETFL, 0);
    fcntl(stderr_pipe[0], F_SETFL, flags | O_NONBLOCK);

    *out_pid = (double)pid;
    *out_stdin_fd = (double)stdin_pipe[1];
    *out_stdout_fd = (double)stdout_pipe[0];
    *out_stderr_fd = (double)stderr_pipe[0];
    return 0;
}

int scriptgo_child_process_poll_status(double pid_num, double *out_status, double *out_exited) {
    if (out_status == NULL || out_exited == NULL) {
        return cp_fail("poll_status invalid arguments");
    }
    pid_t pid = (pid_t)pid_num;
    if (pid <= 0) {
        *out_status = 1.0;
        *out_exited = 1.0;
        return 0;
    }
    int status = 0;
    pid_t res = waitpid(pid, &status, WNOHANG);
    if (res == 0) {
        *out_status = 0.0;
        *out_exited = 0.0;
        return 0;
    } else if (res > 0) {
        *out_exited = 1.0;
        if (WIFEXITED(status)) {
            *out_status = (double)WEXITSTATUS(status);
        } else if (WIFSIGNALED(status)) {
            *out_status = (double)(128 + WTERMSIG(status));
        } else {
            *out_status = 1.0;
        }
        return 0;
    } else {
        *out_exited = 1.0;
        *out_status = 0.0;
        return 0;
    }
}

int scriptgo_child_process_pipe_read(double fd_num, double max_len, char **out_data, double *out_bytes_read, double *out_eof) {
    if (out_data == NULL || out_bytes_read == NULL || out_eof == NULL) {
        return cp_fail("pipe_read invalid arguments");
    }
    int fd = (int)fd_num;
    size_t limit = (size_t)max_len;
    if (limit <= 0) limit = 65536;
    char *buf = malloc(limit + 1);
    if (buf == NULL) return cp_fail("pipe_read malloc failed");

    ssize_t n = read(fd, buf, limit);
    if (n > 0) {
        buf[n] = '\0';
        *out_data = buf;
        *out_bytes_read = (double)n;
        *out_eof = 0.0;
        return 0;
    } else if (n == 0) {
        buf[0] = '\0';
        *out_data = buf;
        *out_bytes_read = 0.0;
        *out_eof = 1.0;
        return 0;
    } else {
        buf[0] = '\0';
        *out_data = buf;
        *out_bytes_read = 0.0;
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            *out_eof = 0.0;
            return 0;
        }
        *out_eof = 1.0;
        return 0;
    }
}

int scriptgo_child_process_pipe_write(double fd_num, const char *data, double len, double *out_written) {
    if (out_written == NULL) return cp_fail("pipe_write invalid arguments");
    int fd = (int)fd_num;
    if (data == NULL || len <= 0.0) {
        *out_written = 0.0;
        return 0;
    }
    ssize_t n = write(fd, data, (size_t)len);
    if (n < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            *out_written = 0.0;
            return 0;
        }
        *out_written = -1.0;
        return cp_fail("pipe_write failed");
    }
    *out_written = (double)n;
    return 0;
}

int scriptgo_child_process_pipe_close(double fd_num) {
    int fd = (int)fd_num;
    if (fd >= 0) {
        close(fd);
    }
    return 0;
}

int scriptgo_child_process_kill(double pid_num, double sig_num, double *out_success) {
    if (out_success == NULL) return cp_fail("kill invalid arguments");
    pid_t pid = (pid_t)pid_num;
    int sig = (int)sig_num;
    if (sig <= 0) sig = SIGTERM;
    int res = kill(pid, sig);
    *out_success = (res == 0) ? 1.0 : 0.0;
    return 0;
}
#endif
