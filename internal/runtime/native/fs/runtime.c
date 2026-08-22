#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <dirent.h>
#include <unistd.h>
#include <errno.h>
#include <stdint.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_push(void *handle, const void *value, double *out_length);

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
        struct stat st;
        if (stat(path, &st) == 0) {
            *out_bool = 1.0;
        } else {
            *out_bool = 0.0;
        }
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

int scriptgo_fs_stat_sync(const char *path, double *out_size, double *out_mtime, double *out_birthtime, double *out_mode) {
    struct stat st;
    if (path == NULL || out_size == NULL || out_mtime == NULL || out_birthtime == NULL || out_mode == NULL) {
        return fs_fail("scriptgo fs stat invalid arguments");
    }
    if (stat(path, &st) != 0) {
        return fs_fail("scriptgo fs stat failed");
    }
    *out_size = (double)st.st_size;
    *out_mtime = (double)st.st_mtime * 1000.0;
#if defined(__APPLE__)
    *out_birthtime = (double)st.st_birthtimespec.tv_sec * 1000.0;
#else
    *out_birthtime = (double)st.st_ctime * 1000.0;
#endif
    *out_mode = (double)st.st_mode;
    return 0;
}

int scriptgo_fs_readdir_sync(const char *path, void **out_array) {
    DIR *d;
    struct dirent *dir;
    double dummy;

    if (path == NULL || out_array == NULL) {
        return fs_fail("scriptgo fs readdir invalid arguments");
    }
    if (scriptgo_array_new(0, sizeof(char *), out_array) != 0) {
        return fs_fail("scriptgo fs readdir array allocation failed");
    }
    d = opendir(path);
    if (d == NULL) {
        return fs_fail("scriptgo fs cannot open directory");
    }
    while ((dir = readdir(d)) != NULL) {
        if (strcmp(dir->d_name, ".") == 0 || strcmp(dir->d_name, "..") == 0) {
            continue;
        }
        char *name = strdup(dir->d_name);
        if (name != NULL) {
            scriptgo_array_push(*out_array, &name, &dummy);
        }
    }
    closedir(d);
    return 0;
}

int scriptgo_fs_copy_file_sync(const char *src, const char *dest) {
    FILE *fsrc, *fdest;
    char buffer[8192];
    size_t bytes;

    if (src == NULL || dest == NULL) {
        return fs_fail("scriptgo fs copyFile invalid arguments");
    }
    fsrc = fopen(src, "rb");
    if (fsrc == NULL) {
        return fs_fail("scriptgo fs copyFile cannot open source");
    }
    fdest = fopen(dest, "wb");
    if (fdest == NULL) {
        fclose(fsrc);
        return fs_fail("scriptgo fs copyFile cannot open destination");
    }
    while ((bytes = fread(buffer, 1, sizeof(buffer), fsrc)) > 0) {
        if (fwrite(buffer, 1, bytes, fdest) != bytes) {
            fclose(fsrc);
            fclose(fdest);
            return fs_fail("scriptgo fs copyFile write failed");
        }
    }
    fclose(fsrc);
    fclose(fdest);
    return 0;
}

int scriptgo_fs_rename_sync(const char *oldpath, const char *newpath) {
    if (oldpath == NULL || newpath == NULL) {
        return fs_fail("scriptgo fs rename invalid arguments");
    }
    if (rename(oldpath, newpath) != 0) {
        return fs_fail("scriptgo fs rename failed");
    }
    return 0;
}

int scriptgo_fs_append_file_sync(const char *path, const char *content) {
    FILE *f;
    size_t len;

    if (path == NULL || content == NULL) {
        return fs_fail("scriptgo fs appendFile invalid arguments");
    }
    f = fopen(path, "ab");
    if (f == NULL) {
        return fs_fail("scriptgo fs appendFile cannot open file");
    }
    len = strlen(content);
    if (len > 0) {
        if (fwrite(content, 1, len, f) != len) {
            fclose(f);
            return fs_fail("scriptgo fs appendFile write failed");
        }
    }
    fclose(f);
    return 0;
}

static int mkdir_recursive(const char *path) {
    char tmp[1024];
    char *p = NULL;
    size_t len;

    snprintf(tmp, sizeof(tmp), "%s", path);
    len = strlen(tmp);
    if (tmp[len - 1] == '/') {
        tmp[len - 1] = 0;
    }
    for (p = tmp + 1; *p; p++) {
        if (*p == '/') {
            *p = 0;
            mkdir(tmp, 0755);
            *p = '/';
        }
    }
    return mkdir(tmp, 0755);
}

int scriptgo_fs_mkdir_sync(const char *path, double is_recursive) {
    if (path == NULL) {
        return fs_fail("scriptgo fs mkdir invalid arguments");
    }
    if (is_recursive > 0.5) {
        mkdir_recursive(path);
    } else {
        if (mkdir(path, 0755) != 0 && errno != EEXIST) {
            return fs_fail("scriptgo fs mkdir failed");
        }
    }
    return 0;
}

static int rm_recursive(const char *path) {
    struct stat st;
    if (lstat(path, &st) != 0) {
        return -1;
    }
    if (S_ISDIR(st.st_mode)) {
        DIR *d = opendir(path);
        if (d == NULL) return -1;
        struct dirent *ent;
        while ((ent = readdir(d)) != NULL) {
            if (strcmp(ent->d_name, ".") == 0 || strcmp(ent->d_name, "..") == 0) continue;
            char subpath[1024];
            snprintf(subpath, sizeof(subpath), "%s/%s", path, ent->d_name);
            rm_recursive(subpath);
        }
        closedir(d);
        return rmdir(path);
    } else {
        return unlink(path);
    }
}

int scriptgo_fs_rm_sync(const char *path, double is_recursive, double is_force) {
    if (path == NULL) {
        return fs_fail("scriptgo fs rm invalid arguments");
    }
    int err;
    if (is_recursive > 0.5) {
        err = rm_recursive(path);
    } else {
        err = remove(path);
    }
    if (err != 0) {
        if (is_force <= 0.5) {
            return fs_fail("scriptgo fs rm failed");
        }
    }
    return 0;
}
