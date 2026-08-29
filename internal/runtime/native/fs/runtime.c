#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/statvfs.h>
#include <sys/time.h>
#include <utime.h>
#include <dirent.h>
#include <unistd.h>
#include <fcntl.h>
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
        return fs_fail("scriptgo fs read_file_sync invalid arguments");
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
        return fs_fail("scriptgo fs write_file_sync invalid arguments");
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
        return fs_fail("scriptgo fs exists_sync invalid arguments");
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
        return fs_fail("scriptgo fs unlink_sync invalid arguments");
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

int scriptgo_fs_access_sync(const char *path, double mode, double *out_ok) {
    if (path == NULL || out_ok == NULL) {
        return fs_fail("scriptgo fs access invalid arguments");
    }
    int amode = (int)mode;
    if (access(path, amode) == 0) {
        *out_ok = 1.0;
    } else {
        *out_ok = 0.0;
    }
    return 0;
}

int scriptgo_fs_chmod_sync(const char *path, double mode) {
    if (path == NULL) {
        return fs_fail("scriptgo fs chmod invalid arguments");
    }
    if (chmod(path, (mode_t)mode) != 0) {
        return fs_fail("scriptgo fs chmod failed");
    }
    return 0;
}

int scriptgo_fs_realpath_sync(const char *path, char **out_path) {
    if (path == NULL || out_path == NULL) {
        return fs_fail("scriptgo fs realpath invalid arguments");
    }
    char *res = realpath(path, NULL);
    if (res == NULL) {
        return fs_fail("scriptgo fs realpath failed");
    }
    *out_path = res;
    return 0;
}

int scriptgo_fs_truncate_sync(const char *path, double len) {
    if (path == NULL) {
        return fs_fail("scriptgo fs truncate invalid arguments");
    }
    if (truncate(path, (off_t)len) != 0) {
        return fs_fail("scriptgo fs truncate failed");
    }
    return 0;
}

int scriptgo_fs_mkdtemp_sync(const char *prefix, char **out_path) {
    if (prefix == NULL || out_path == NULL) {
        return fs_fail("scriptgo fs mkdtemp invalid arguments");
    }
    size_t plen = strlen(prefix);
    char *template_str = (char *)malloc(plen + 7);
    if (template_str == NULL) {
        return fs_fail("scriptgo fs mkdtemp memory allocation failed");
    }
    snprintf(template_str, plen + 7, "%sXXXXXX", prefix);
    char *res = mkdtemp(template_str);
    if (res == NULL) {
        free(template_str);
        return fs_fail("scriptgo fs mkdtemp failed");
    }
    *out_path = res;
    return 0;
}

static int parse_open_flags(const char *flags) {
    if (flags == NULL || *flags == '\0' || strcmp(flags, "r") == 0) return O_RDONLY;
    if (strcmp(flags, "r+") == 0) return O_RDWR;
    if (strcmp(flags, "w") == 0) return O_WRONLY | O_CREAT | O_TRUNC;
    if (strcmp(flags, "w+") == 0) return O_RDWR | O_CREAT | O_TRUNC;
    if (strcmp(flags, "a") == 0) return O_WRONLY | O_CREAT | O_APPEND;
    if (strcmp(flags, "a+") == 0) return O_RDWR | O_CREAT | O_APPEND;
    if (strcmp(flags, "wx") == 0 || strcmp(flags, "xw") == 0) return O_WRONLY | O_CREAT | O_TRUNC | O_EXCL;
    if (strcmp(flags, "ax") == 0 || strcmp(flags, "xa") == 0) return O_WRONLY | O_CREAT | O_APPEND | O_EXCL;
    return O_RDONLY;
}

int scriptgo_fs_open_sync(const char *path, const char *flags, double mode, double *out_fd) {
    if (path == NULL || out_fd == NULL) {
        return fs_fail("scriptgo fs open invalid arguments");
    }
    int oflag = parse_open_flags(flags);
    mode_t cmode = (mode_t)mode;
    if (cmode == 0) cmode = 0666;
    int fd = open(path, oflag, cmode);
    if (fd < 0) {
        return fs_fail("scriptgo fs open failed");
    }
    *out_fd = (double)fd;
    return 0;
}

int scriptgo_fs_close_sync(double fd) {
    int ifd = (int)fd;
    if (ifd < 0) {
        return fs_fail("scriptgo fs close invalid fd");
    }
    if (close(ifd) != 0) {
        return fs_fail("scriptgo fs close failed");
    }
    return 0;
}

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_fs_buffer_view_header;

static unsigned char *fs_extract_buffer_data(void *handle, size_t *out_cap) {
    if (handle == NULL) return NULL;
    scriptgo_fs_buffer_view_header *bv = (scriptgo_fs_buffer_view_header *)handle;
    if (bv->magic == 0x42554646 || bv->magic == 0x54415252) { // "BUFF" or "TARR"
        if (out_cap != NULL) *out_cap = (size_t)bv->length;
        return bv->data;
    }
    if (out_cap != NULL) *out_cap = strlen((const char *)handle);
    return (unsigned char *)handle;
}

int scriptgo_fs_read_fd_sync(double fd, void *buffer_handle, double offset, double length, double position, double *out_bytes_read) {
    int ifd = (int)fd;
    if (ifd < 0 || buffer_handle == NULL || out_bytes_read == NULL) {
        return fs_fail("scriptgo fs read invalid arguments");
    }
    size_t buf_cap = 0;
    unsigned char *raw_buf = fs_extract_buffer_data(buffer_handle, &buf_cap);
    if (raw_buf == NULL) {
        return fs_fail("scriptgo fs read invalid buffer");
    }
    size_t len = (size_t)length;
    size_t off = (size_t)offset;
    if (len == 0 && buf_cap > off) {
        len = buf_cap - off;
    }
    ssize_t n;
    if (position >= 0) {
        n = pread(ifd, raw_buf + off, len, (off_t)position);
    } else {
        n = read(ifd, raw_buf + off, len);
    }
    if (n < 0) {
        return fs_fail("scriptgo fs read failed");
    }
    *out_bytes_read = (double)n;
    return 0;
}

int scriptgo_fs_write_fd_sync(double fd, void *data_handle, double offset, double length, double *out_bytes_written) {
    int ifd = (int)fd;
    if (ifd < 0 || data_handle == NULL || out_bytes_written == NULL) {
        return fs_fail("scriptgo fs write invalid arguments");
    }
    size_t buf_cap = 0;
    unsigned char *raw_buf = fs_extract_buffer_data(data_handle, &buf_cap);
    if (raw_buf == NULL) {
        return fs_fail("scriptgo fs write invalid data");
    }
    size_t len = (size_t)length;
    size_t off = (size_t)offset;
    if (len == 0 && buf_cap > off) {
        len = buf_cap - off;
    }
    ssize_t n = write(ifd, raw_buf + off, len);
    if (n < 0) {
        return fs_fail("scriptgo fs write failed");
    }
    *out_bytes_written = (double)n;
    return 0;
}

int scriptgo_fs_opendir_sync(const char *path, void **out_names, void **out_types) {
    DIR *d;
    struct dirent *dir;
    double dummy;

    if (path == NULL || out_names == NULL || out_types == NULL) {
        return fs_fail("scriptgo fs opendir invalid arguments");
    }
    if (scriptgo_array_new(0, sizeof(char *), out_names) != 0 ||
        scriptgo_array_new(0, sizeof(double), out_types) != 0) {
        return fs_fail("scriptgo fs opendir array allocation failed");
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
        double type_num = 0.0;
#if defined(DT_DIR)
        if (dir->d_type == DT_DIR) type_num = 2.0;
        else if (dir->d_type == DT_REG) type_num = 1.0;
        else if (dir->d_type == DT_LNK) type_num = 3.0;
#endif
        if (name != NULL) {
            scriptgo_array_push(*out_names, &name, &dummy);
            scriptgo_array_push(*out_types, &type_num, &dummy);
        }
    }
    closedir(d);
    return 0;
}

int scriptgo_fs_fstat_sync(double fd, double *out_size, double *out_mtime, double *out_birthtime, double *out_mode, double *out_uid, double *out_gid, double *out_ino, double *out_dev) {
    int ifd = (int)fd;
    struct stat st;
    if (ifd < 0 || out_size == NULL || out_mtime == NULL || out_birthtime == NULL || out_mode == NULL) {
        return fs_fail("scriptgo fs fstat invalid arguments");
    }
    if (fstat(ifd, &st) != 0) {
        return fs_fail("scriptgo fs fstat failed");
    }
    *out_size = (double)st.st_size;
    *out_mtime = (double)st.st_mtime * 1000.0;
#if defined(__APPLE__)
    *out_birthtime = (double)st.st_birthtimespec.tv_sec * 1000.0 + (double)st.st_birthtimespec.tv_nsec / 1000000.0;
#else
    *out_birthtime = (double)st.st_ctime * 1000.0;
#endif
    *out_mode = (double)st.st_mode;
    if (out_uid != NULL) *out_uid = (double)st.st_uid;
    if (out_gid != NULL) *out_gid = (double)st.st_gid;
    if (out_ino != NULL) *out_ino = (double)st.st_ino;
    if (out_dev != NULL) *out_dev = (double)st.st_dev;
    return 0;
}

int scriptgo_fs_statfs_sync(const char *path, double *out_bsize, double *out_blocks, double *out_bfree, double *out_bavail, double *out_files, double *out_ffree) {
    struct statvfs st;
    if (path == NULL || out_bsize == NULL || out_blocks == NULL || out_bfree == NULL || out_bavail == NULL || out_files == NULL || out_ffree == NULL) {
        return fs_fail("scriptgo fs statfs invalid arguments");
    }
    if (statvfs(path, &st) != 0) {
        return fs_fail("scriptgo fs statfs failed");
    }
    *out_bsize = (double)st.f_bsize;
    *out_blocks = (double)st.f_blocks;
    *out_bfree = (double)st.f_bfree;
    *out_bavail = (double)st.f_bavail;
    *out_files = (double)st.f_files;
    *out_ffree = (double)st.f_ffree;
    return 0;
}

int scriptgo_fs_chown_sync(const char *path, double uid, double gid) {
    if (path == NULL) return fs_fail("scriptgo fs chown invalid arguments");
    if (chown(path, (uid_t)(int)uid, (gid_t)(int)gid) != 0) {
        return fs_fail("scriptgo fs chown failed");
    }
    return 0;
}

int scriptgo_fs_lchown_sync(const char *path, double uid, double gid) {
    if (path == NULL) return fs_fail("scriptgo fs lchown invalid arguments");
    if (lchown(path, (uid_t)(int)uid, (gid_t)(int)gid) != 0) {
        return fs_fail("scriptgo fs lchown failed");
    }
    return 0;
}

int scriptgo_fs_fchown_sync(double fd, double uid, double gid) {
    int ifd = (int)fd;
    if (ifd < 0) return fs_fail("scriptgo fs fchown invalid arguments");
    if (fchown(ifd, (uid_t)(int)uid, (gid_t)(int)gid) != 0) {
        return fs_fail("scriptgo fs fchown failed");
    }
    return 0;
}

int scriptgo_fs_fchmod_sync(double fd, double mode) {
    int ifd = (int)fd;
    if (ifd < 0) return fs_fail("scriptgo fs fchmod invalid arguments");
    if (fchmod(ifd, (mode_t)mode) != 0) {
        return fs_fail("scriptgo fs fchmod failed");
    }
    return 0;
}

int scriptgo_fs_link_sync(const char *existing_path, const char *new_path) {
    if (existing_path == NULL || new_path == NULL) return fs_fail("scriptgo fs link invalid arguments");
    if (link(existing_path, new_path) != 0) {
        return fs_fail("scriptgo fs link failed");
    }
    return 0;
}

int scriptgo_fs_symlink_sync(const char *target, const char *path) {
    if (target == NULL || path == NULL) return fs_fail("scriptgo fs symlink invalid arguments");
    if (symlink(target, path) != 0) {
        return fs_fail("scriptgo fs symlink failed");
    }
    return 0;
}

int scriptgo_fs_readlink_sync(const char *path, char **out_link) {
    char buf[1024];
    ssize_t len;
    if (path == NULL || out_link == NULL) return fs_fail("scriptgo fs readlink invalid arguments");
    len = readlink(path, buf, sizeof(buf) - 1);
    if (len < 0) {
        return fs_fail("scriptgo fs readlink failed");
    }
    buf[len] = '\0';
    *out_link = strdup(buf);
    return 0;
}

int scriptgo_fs_utimes_sync(const char *path, double atime, double mtime) {
    struct timeval tv[2];
    if (path == NULL) return fs_fail("scriptgo fs utimes invalid arguments");
    tv[0].tv_sec = (time_t)atime;
    tv[0].tv_usec = (suseconds_t)((atime - (time_t)atime) * 1000000.0);
    tv[1].tv_sec = (time_t)mtime;
    tv[1].tv_usec = (suseconds_t)((mtime - (time_t)mtime) * 1000000.0);
    if (utimes(path, tv) != 0) {
        return fs_fail("scriptgo fs utimes failed");
    }
    return 0;
}

int scriptgo_fs_lutimes_sync(const char *path, double atime, double mtime) {
    struct timeval tv[2];
    if (path == NULL) return fs_fail("scriptgo fs lutimes invalid arguments");
    tv[0].tv_sec = (time_t)atime;
    tv[0].tv_usec = (suseconds_t)((atime - (time_t)atime) * 1000000.0);
    tv[1].tv_sec = (time_t)mtime;
    tv[1].tv_usec = (suseconds_t)((mtime - (time_t)mtime) * 1000000.0);
#if defined(__APPLE__) || defined(__FreeBSD__)
    if (lutimes(path, tv) != 0) {
        return fs_fail("scriptgo fs lutimes failed");
    }
#else
    if (utimes(path, tv) != 0) {
        return fs_fail("scriptgo fs lutimes failed");
    }
#endif
    return 0;
}

int scriptgo_fs_futimes_sync(double fd, double atime, double mtime) {
    int ifd = (int)fd;
    struct timeval tv[2];
    if (ifd < 0) return fs_fail("scriptgo fs futimes invalid arguments");
    tv[0].tv_sec = (time_t)atime;
    tv[0].tv_usec = (suseconds_t)((atime - (time_t)atime) * 1000000.0);
    tv[1].tv_sec = (time_t)mtime;
    tv[1].tv_usec = (suseconds_t)((mtime - (time_t)mtime) * 1000000.0);
    if (futimes(ifd, tv) != 0) {
        return fs_fail("scriptgo fs futimes failed");
    }
    return 0;
}

int scriptgo_fs_fsync_sync(double fd) {
    int ifd = (int)fd;
    if (ifd < 0) return fs_fail("scriptgo fs fsync invalid arguments");
    if (fsync(ifd) != 0) {
        return fs_fail("scriptgo fs fsync failed");
    }
    return 0;
}

int scriptgo_fs_fdatasync_sync(double fd) {
    int ifd = (int)fd;
    if (ifd < 0) return fs_fail("scriptgo fs fdatasync invalid arguments");
#if defined(__APPLE__)
    if (fsync(ifd) != 0) {
        return fs_fail("scriptgo fs fdatasync failed");
    }
#else
    if (fdatasync(ifd) != 0) {
        return fs_fail("scriptgo fs fdatasync failed");
    }
#endif
    return 0;
}

int scriptgo_fs_ftruncate_sync(double fd, double len) {
    int ifd = (int)fd;
    if (ifd < 0) return fs_fail("scriptgo fs ftruncate invalid arguments");
    if (ftruncate(ifd, (off_t)len) != 0) {
        return fs_fail("scriptgo fs ftruncate failed");
    }
    return 0;
}

int scriptgo_fs_rmdir_sync(const char *path) {
    if (path == NULL) return fs_fail("scriptgo fs rmdir invalid arguments");
    if (rmdir(path) != 0) {
        return fs_fail("scriptgo fs rmdir failed");
    }
    return 0;
}


