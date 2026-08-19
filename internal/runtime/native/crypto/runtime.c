#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__)
#define HAVE_ARC4RANDOM 1
#endif

int scriptgo_runtime_set_error(const char *message);

static int crypto_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static int get_random_bytes(unsigned char *buf, size_t n) {
#if defined(HAVE_ARC4RANDOM)
    arc4random_buf(buf, n);
    return 0;
#else
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd < 0) return -1;
    size_t total = 0;
    while (total < n) {
        ssize_t r = read(fd, buf + total, n - total);
        if (r <= 0) {
            close(fd);
            return -1;
        }
        total += (size_t)r;
    }
    close(fd);
    return 0;
#endif
}

int scriptgo_crypto_random_uuid(char **out_uuid) {
    if (out_uuid == NULL) {
        return crypto_fail("scriptgo crypto invalid arguments");
    }
    unsigned char b[16];
    if (get_random_bytes(b, 16) != 0) {
        return crypto_fail("scriptgo crypto failed to generate random bytes");
    }
    // Set version 4 (0100xxxx)
    b[6] = (b[6] & 0x0f) | 0x40;
    // Set variant to RFC 4122 (10xxxxxx)
    b[8] = (b[8] & 0x3f) | 0x80;

    char *uuid = (char *)malloc(37);
    if (uuid == NULL) {
        return crypto_fail("scriptgo crypto allocation failed");
    }
    snprintf(uuid, 37,
        "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
        b[0], b[1], b[2], b[3],
        b[4], b[5],
        b[6], b[7],
        b[8], b[9],
        b[10], b[11], b[12], b[13], b[14], b[15]
    );
    *out_uuid = uuid;
    return 0;
}
