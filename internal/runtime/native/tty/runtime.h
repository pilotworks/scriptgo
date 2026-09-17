#ifndef SCRIPTGO_TTY_RUNTIME_H
#define SCRIPTGO_TTY_RUNTIME_H

#ifdef __cplusplus
extern "C" {
#endif

int scriptgo_tty_isatty(double fd, double *out_isatty);
int scriptgo_tty_get_window_size(double fd, double *out_cols, double *out_rows);
int scriptgo_tty_set_raw_mode(double fd, double mode, double *out_success);
int scriptgo_tty_read(double fd, double max_len, char **out_data, double *out_bytes_read);
int scriptgo_tty_read_line(double fd, char **out_line);
int scriptgo_tty_write(double fd, const char *data, double len, double *out_written);

#ifdef __cplusplus
}
#endif

#endif /* SCRIPTGO_TTY_RUNTIME_H */
