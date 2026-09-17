#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>

#if defined(_WIN32)
#include <io.h>
#include <windows.h>
#elif !defined(__wasi__)
#include <termios.h>
#include <sys/ioctl.h>
#endif

int scriptgo_runtime_set_error(const char *message);

static int tty_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

int scriptgo_tty_isatty(double fd, double *out_isatty) {
    if (out_isatty == NULL) return tty_fail("scriptgo tty isatty: invalid arguments");
#if defined(__wasi__)
    *out_isatty = 0.0;
    return 0;
#elif defined(_WIN32)
    int fd_int = (int)fd;
    HANDLE h = (HANDLE)_get_osfhandle(fd_int);
    if (h == INVALID_HANDLE_VALUE) {
        *out_isatty = 0.0;
        return 0;
    }
    DWORD mode;
    *out_isatty = GetConsoleMode(h, &mode) ? 1.0 : 0.0;
    return 0;
#else
    int fd_int = (int)fd;
    int res = isatty(fd_int);
    *out_isatty = res ? 1.0 : 0.0;
    return 0;
#endif
}

int scriptgo_tty_get_window_size(double fd, double *out_cols, double *out_rows) {
    if (out_cols == NULL || out_rows == NULL) return tty_fail("scriptgo tty get_window_size: invalid arguments");
    double cols = 0.0;
    double rows = 0.0;
#if defined(__wasi__)
    /* Fall through to env / default fallback */
#elif defined(_WIN32)
    int fd_int = (int)fd;
    HANDLE h = (HANDLE)_get_osfhandle(fd_int);
    if (h != INVALID_HANDLE_VALUE) {
        CONSOLE_SCREEN_BUFFER_INFO csbi;
        if (GetConsoleScreenBufferInfo(h, &csbi)) {
            cols = (double)(csbi.srWindow.Right - csbi.srWindow.Left + 1);
            rows = (double)(csbi.srWindow.Bottom - csbi.srWindow.Top + 1);
        }
    }
#else
    int fd_int = (int)fd;
    struct winsize ws;
    if (ioctl(fd_int, TIOCGWINSZ, &ws) == 0 && ws.ws_col != 0) {
        cols = (double)ws.ws_col;
        rows = (double)ws.ws_row;
    }
#endif

    if (cols <= 0.0 || rows <= 0.0) {
        const char *env_cols = getenv("COLUMNS");
        const char *env_rows = getenv("LINES");
        if (env_cols) cols = atof(env_cols);
        if (env_rows) rows = atof(env_rows);
    }
    if (cols <= 0.0) cols = 80.0;
    if (rows <= 0.0) rows = 24.0;
    *out_cols = cols;
    *out_rows = rows;
    return 0;
}

#if !defined(__wasi__) && !defined(_WIN32)
static struct termios g_tty_orig_termios;
static int g_tty_orig_saved = 0;
#elif defined(_WIN32)
static DWORD g_tty_orig_mode = 0;
static int g_tty_orig_saved = 0;
#endif

int scriptgo_tty_set_raw_mode(double fd, double mode, double *out_success) {
#if defined(__wasi__)
    if (out_success) *out_success = 0.0;
    return 0;
#elif defined(_WIN32)
    int fd_int = (int)fd;
    HANDLE h = (HANDLE)_get_osfhandle(fd_int);
    if (h == INVALID_HANDLE_VALUE) {
        if (out_success) *out_success = 0.0;
        return 0;
    }
    DWORD current_mode;
    if (!GetConsoleMode(h, &current_mode)) {
        if (out_success) *out_success = 0.0;
        return 0;
    }
    if (!g_tty_orig_saved) {
        g_tty_orig_mode = current_mode;
        g_tty_orig_saved = 1;
    }
    if (mode != 0.0) {
        DWORD raw_mode = current_mode;
        raw_mode &= ~(ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT | ENABLE_PROCESSED_INPUT);
        SetConsoleMode(h, raw_mode);
    } else {
        SetConsoleMode(h, g_tty_orig_mode);
    }
    if (out_success) *out_success = 1.0;
    return 0;
#else
    int fd_int = (int)fd;
    if (!isatty(fd_int)) {
        if (out_success) *out_success = 0.0;
        return 0;
    }
    if (mode != 0.0) {
        struct termios raw;
        if (tcgetattr(fd_int, &raw) < 0) {
            if (out_success) *out_success = 0.0;
            return tty_fail("tcgetattr failed");
        }
        if (!g_tty_orig_saved) {
            g_tty_orig_termios = raw;
            g_tty_orig_saved = 1;
        }
        raw.c_iflag &= ~(BRKINT | ICRNL | INPCK | ISTRIP | IXON);
        raw.c_oflag &= ~(OPOST);
        raw.c_cflag |= (CS8);
        raw.c_lflag &= ~(ECHO | ICANON | IEXTEN | ISIG);
        raw.c_cc[VMIN] = 1;
        raw.c_cc[VTIME] = 0;
        if (tcsetattr(fd_int, TCSANOW, &raw) < 0) {
            if (out_success) *out_success = 0.0;
            return tty_fail("tcsetattr failed");
        }
    } else {
        if (g_tty_orig_saved) {
            tcsetattr(fd_int, TCSANOW, &g_tty_orig_termios);
        }
    }
    if (out_success) *out_success = 1.0;
    return 0;
#endif
}

int scriptgo_tty_read(double fd, double max_len, char **out_data, double *out_bytes_read) {
    if (out_data == NULL || out_bytes_read == NULL) return tty_fail("scriptgo tty read: invalid arguments");
    int fd_int = (int)fd;
    size_t lim = (size_t)max_len;
    if (lim == 0) lim = 4096;

    char *buf = (char *)malloc(lim + 1);
    if (buf == NULL) return tty_fail("scriptgo tty read: allocation failed");

    ssize_t n = read(fd_int, buf, lim);
    if (n < 0) {
        free(buf);
        *out_data = strdup("");
        *out_bytes_read = 0.0;
        return 0;
    }
    buf[n] = '\0';
    *out_data = buf;
    *out_bytes_read = (double)n;
    return 0;
}

int scriptgo_tty_read_line(double fd, char **out_line) {
    if (out_line == NULL) return tty_fail("scriptgo tty read_line: invalid arguments");
    int fd_int = (int)fd;
    size_t cap = 256;
    size_t len = 0;
    char *buf = (char *)malloc(cap);
    if (buf == NULL) return tty_fail("scriptgo tty read_line: allocation failed");

    char ch;
    while (1) {
        ssize_t n = read(fd_int, &ch, 1);
        if (n <= 0) {
            break;
        }
        if (ch == '\n') {
            break;
        }
        if (ch == '\r') {
            continue;
        }
        if (len + 1 >= cap) {
            cap *= 2;
            char *new_buf = (char *)realloc(buf, cap);
            if (new_buf == NULL) {
                free(buf);
                return tty_fail("scriptgo tty read_line: allocation failed");
            }
            buf = new_buf;
        }
        buf[len++] = ch;
    }
    buf[len] = '\0';
    *out_line = buf;
    return 0;
}

int scriptgo_tty_write(double fd, const char *data, double len, double *out_written) {
    if (data == NULL) return tty_fail("scriptgo tty write: invalid arguments");
    int fd_int = (int)fd;
    size_t write_len = (len > 0.0) ? (size_t)len : strlen(data);
    ssize_t n = write(fd_int, data, write_len);
    if (out_written) {
        *out_written = (n >= 0) ? (double)n : 0.0;
    }
    return 0;
}
