#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_runtime_set_error(const char *message);

static int fs_fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_fs_read_file_sync(const char *path, char **out_content) {
    FILE *f;
    long size;
    char *buf;
    size_t read_bytes;

    if (path == NULL || out_content == NULL) {
        return fs_fail("scriptgo fs invalid arguments");
    }
    f = fopen(path, "rb");
    if (f == NULL) {
        return fs_fail("scriptgo fs cannot open file for reading");
    }
    if (fseek(f, 0, SEEK_END) != 0) {
        fclose(f);
        return fs_fail("scriptgo fs seek failed");
    }
    size = ftell(f);
    if (size < 0) {
        fclose(f);
        return fs_fail("scriptgo fs ftell failed");
    }
    rewind(f);

    buf = malloc((size_t)size + 1);
    if (buf == NULL) {
        fclose(f);
        return fs_fail("scriptgo fs buffer allocation failed");
    }
    read_bytes = fread(buf, 1, (size_t)size, f);
    buf[read_bytes] = '\0';
    fclose(f);

    *out_content = buf;
    return 0;
}

int scriptgo_fs_write_file_sync(const char *path, const char *content) {
    FILE *f;
    size_t len;

    if (path == NULL || content == NULL) {
        return fs_fail("scriptgo fs invalid arguments");
    }
    f = fopen(path, "wb");
    if (f == NULL) {
        return fs_fail("scriptgo fs cannot open file for writing");
    }
    len = strlen(content);
    if (len > 0) {
        if (fwrite(content, 1, len, f) != len) {
            fclose(f);
            return fs_fail("scriptgo fs write failed");
        }
    }
    fclose(f);
    return 0;
}

int scriptgo_fs_exists_sync(const char *path, double *out_bool) {
    FILE *f;
    if (path == NULL || out_bool == NULL) {
        return fs_fail("scriptgo fs invalid arguments");
    }
    f = fopen(path, "rb");
    if (f != NULL) {
        fclose(f);
        *out_bool = 1.0;
    } else {
        *out_bool = 0.0;
    }
    return 0;
}

int scriptgo_fs_unlink_sync(const char *path) {
    if (path == NULL) {
        return fs_fail("scriptgo fs invalid arguments");
    }
    remove(path);
    return 0;
}

