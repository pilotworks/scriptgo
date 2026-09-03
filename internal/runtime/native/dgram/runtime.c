#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <unistd.h>
#include <errno.h>

#if !defined(_WIN32)
#include <sys/socket.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#else
#include <winsock2.h>
#include <ws2tcpip.h>
#endif

int scriptgo_runtime_set_error(const char *message);

static int dgram_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

int scriptgo_dgram_socket_create(double family, double *out_fd) {
    if (out_fd == NULL) {
        return dgram_fail("scriptgo dgram socket_create invalid argument");
    }

    int fam = AF_INET;
    if ((int)family == 6) {
        fam = AF_INET6;
    }

    int fd = socket(fam, SOCK_DGRAM, 0);
    if (fd < 0) {
        return dgram_fail(strerror(errno));
    }

    int opt = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, (const char *)&opt, sizeof(opt));

    *out_fd = (double)fd;
    return 0;
}

int scriptgo_dgram_bind(double fd_num, const char *address, double port_num) {
    int fd = (int)fd_num;
    int port = (int)port_num;
    if (fd < 0) {
        return dgram_fail("scriptgo dgram bind invalid socket descriptor");
    }

    struct addrinfo hints;
    struct addrinfo *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_DGRAM;
    hints.ai_flags = AI_PASSIVE;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    const char *node = (address != NULL && strlen(address) > 0 && strcmp(address, "0.0.0.0") != 0) ? address : NULL;
    int rc = getaddrinfo(node, port_str, &hints, &res);
    if (rc != 0 || res == NULL) {
        return dgram_fail(gai_strerror(rc));
    }

    int bound = -1;
    for (struct addrinfo *p = res; p != NULL; p = p->ai_next) {
        if (bind(fd, p->ai_addr, p->ai_addrlen) == 0) {
            bound = 0;
            break;
        }
    }
    freeaddrinfo(res);

    if (bound != 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_send(double fd_num, const char *data, double len_num, double port_num, const char *address, double *out_sent) {
    int fd = (int)fd_num;
    int port = (int)port_num;
    if (fd < 0 || data == NULL || address == NULL || out_sent == NULL) {
        return dgram_fail("scriptgo dgram send invalid arguments");
    }

    struct addrinfo hints;
    struct addrinfo *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_DGRAM;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    int rc = getaddrinfo(address, port_str, &hints, &res);
    if (rc != 0 || res == NULL) {
        return dgram_fail(gai_strerror(rc));
    }

    size_t len = (size_t)len_num;
    ssize_t n = sendto(fd, data, len, 0, res->ai_addr, res->ai_addrlen);
    freeaddrinfo(res);

    if (n < 0) {
        return dgram_fail(strerror(errno));
    }

    *out_sent = (double)n;
    return 0;
}

int scriptgo_dgram_recv(double fd_num, double max_len_num, char **out_data, double *out_read,
                        char **out_rinfo_ip, double *out_rinfo_port, double *out_rinfo_family) {
    int fd = (int)fd_num;
    if (fd < 0 || out_data == NULL || out_read == NULL || out_rinfo_ip == NULL || out_rinfo_port == NULL || out_rinfo_family == NULL) {
        return dgram_fail("scriptgo dgram recv invalid arguments");
    }

    size_t max_len = max_len_num > 0 ? (size_t)max_len_num : 65536;
    char *buf = (char *)malloc(max_len + 1);
    if (buf == NULL) {
        return dgram_fail("scriptgo dgram recv memory allocation failed");
    }

    struct sockaddr_storage src_addr;
    socklen_t addr_len = sizeof(src_addr);
    ssize_t n = recvfrom(fd, buf, max_len, 0, (struct sockaddr *)&src_addr, &addr_len);
    if (n < 0) {
        free(buf);
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            *out_data = strdup("");
            *out_read = 0.0;
            *out_rinfo_ip = strdup("127.0.0.1");
            *out_rinfo_port = 0.0;
            *out_rinfo_family = 4.0;
            return 0;
        }
        return dgram_fail(strerror(errno));
    }

    buf[n] = '\0';
    *out_data = buf;
    *out_read = (double)n;

    char ip_buf[INET6_ADDRSTRLEN];
    memset(ip_buf, 0, sizeof(ip_buf));
    int src_port = 0;
    double fam = 4.0;

    if (src_addr.ss_family == AF_INET) {
        struct sockaddr_in *s = (struct sockaddr_in *)&src_addr;
        inet_ntop(AF_INET, &(s->sin_addr), ip_buf, sizeof(ip_buf));
        src_port = ntohs(s->sin_port);
        fam = 4.0;
    } else if (src_addr.ss_family == AF_INET6) {
        struct sockaddr_in6 *s = (struct sockaddr_in6 *)&src_addr;
        inet_ntop(AF_INET6, &(s->sin6_addr), ip_buf, sizeof(ip_buf));
        src_port = ntohs(s->sin6_port);
        fam = 6.0;
    }

    *out_rinfo_ip = strdup(ip_buf);
    *out_rinfo_port = (double)src_port;
    *out_rinfo_family = fam;
    return 0;
}

int scriptgo_dgram_set_broadcast(double fd_num, double flag_num) {
    int fd = (int)fd_num;
    int opt = (int)flag_num ? 1 : 0;
    if (setsockopt(fd, SOL_SOCKET, SO_BROADCAST, (const char *)&opt, sizeof(opt)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_set_multicast_ttl(double fd_num, double ttl_num) {
    int fd = (int)fd_num;
    int ttl = (int)ttl_num;
    if (setsockopt(fd, IPPROTO_IP, IP_MULTICAST_TTL, (const char *)&ttl, sizeof(ttl)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_set_multicast_loopback(double fd_num, double flag_num) {
    int fd = (int)fd_num;
    int opt = (int)flag_num ? 1 : 0;
    if (setsockopt(fd, IPPROTO_IP, IP_MULTICAST_LOOP, (const char *)&opt, sizeof(opt)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_set_recv_buffer_size(double fd_num, double size_num) {
    int fd = (int)fd_num;
    int size = (int)size_num;
    if (setsockopt(fd, SOL_SOCKET, SO_RCVBUF, (const char *)&size, sizeof(size)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_set_send_buffer_size(double fd_num, double size_num) {
    int fd = (int)fd_num;
    int size = (int)size_num;
    if (setsockopt(fd, SOL_SOCKET, SO_SNDBUF, (const char *)&size, sizeof(size)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_get_recv_buffer_size(double fd_num, double *out_size) {
    int fd = (int)fd_num;
    if (out_size == NULL) return dgram_fail("invalid argument");
    int size = 0;
    socklen_t len = sizeof(size);
    if (getsockopt(fd, SOL_SOCKET, SO_RCVBUF, (char *)&size, &len) < 0) {
        *out_size = 65536.0;
        return 0;
    }
    *out_size = (double)size;
    return 0;
}

int scriptgo_dgram_get_send_buffer_size(double fd_num, double *out_size) {
    int fd = (int)fd_num;
    if (out_size == NULL) return dgram_fail("invalid argument");
    int size = 0;
    socklen_t len = sizeof(size);
    if (getsockopt(fd, SOL_SOCKET, SO_SNDBUF, (char *)&size, &len) < 0) {
        *out_size = 65536.0;
        return 0;
    }
    *out_size = (double)size;
    return 0;
}

int scriptgo_dgram_set_ttl(double fd_num, double ttl_num) {
    int fd = (int)fd_num;
    int ttl = (int)ttl_num;
    if (setsockopt(fd, IPPROTO_IP, IP_TTL, (const char *)&ttl, sizeof(ttl)) < 0) {
        return dgram_fail(strerror(errno));
    }
    return 0;
}

int scriptgo_dgram_set_multicast_interface(double fd_num, const char *iface_addr) {
    int fd = (int)fd_num;
    if (iface_addr == NULL || strlen(iface_addr) == 0) return 0;
    struct in_addr addr;
    if (inet_pton(AF_INET, iface_addr, &addr) > 0) {
        setsockopt(fd, IPPROTO_IP, IP_MULTICAST_IF, (const char *)&addr, sizeof(addr));
    }
    return 0;
}

int scriptgo_dgram_add_membership(double fd_num, const char *mcast_addr, const char *iface_addr) {
    int fd = (int)fd_num;
    if (mcast_addr == NULL || strlen(mcast_addr) == 0) return 0;
    struct ip_mreq mreq;
    memset(&mreq, 0, sizeof(mreq));
    if (inet_pton(AF_INET, mcast_addr, &mreq.imr_multiaddr) <= 0) return 0;
    if (iface_addr != NULL && strlen(iface_addr) > 0) {
        inet_pton(AF_INET, iface_addr, &mreq.imr_interface);
    } else {
        mreq.imr_interface.s_addr = htonl(INADDR_ANY);
    }
    setsockopt(fd, IPPROTO_IP, IP_ADD_MEMBERSHIP, (const char *)&mreq, sizeof(mreq));
    return 0;
}

int scriptgo_dgram_drop_membership(double fd_num, const char *mcast_addr, const char *iface_addr) {
    int fd = (int)fd_num;
    if (mcast_addr == NULL || strlen(mcast_addr) == 0) return 0;
    struct ip_mreq mreq;
    memset(&mreq, 0, sizeof(mreq));
    if (inet_pton(AF_INET, mcast_addr, &mreq.imr_multiaddr) <= 0) return 0;
    if (iface_addr != NULL && strlen(iface_addr) > 0) {
        inet_pton(AF_INET, iface_addr, &mreq.imr_interface);
    } else {
        mreq.imr_interface.s_addr = htonl(INADDR_ANY);
    }
    setsockopt(fd, IPPROTO_IP, IP_DROP_MEMBERSHIP, (const char *)&mreq, sizeof(mreq));
    return 0;
}

int scriptgo_dgram_add_source_specific_membership(double fd_num, const char *src_addr, const char *group_addr, const char *iface_addr) {
    int fd = (int)fd_num;
    if (src_addr == NULL || group_addr == NULL) return 0;
#if defined(IP_ADD_SOURCE_MEMBERSHIP)
    struct ip_mreq_source mreq;
    memset(&mreq, 0, sizeof(mreq));
    if (inet_pton(AF_INET, group_addr, &mreq.imr_multiaddr) <= 0) return 0;
    if (inet_pton(AF_INET, src_addr, &mreq.imr_sourceaddr) <= 0) return 0;
    if (iface_addr != NULL && strlen(iface_addr) > 0) {
        inet_pton(AF_INET, iface_addr, &mreq.imr_interface);
    } else {
        mreq.imr_interface.s_addr = htonl(INADDR_ANY);
    }
    setsockopt(fd, IPPROTO_IP, IP_ADD_SOURCE_MEMBERSHIP, (const char *)&mreq, sizeof(mreq));
#endif
    return 0;
}

int scriptgo_dgram_drop_source_specific_membership(double fd_num, const char *src_addr, const char *group_addr, const char *iface_addr) {
    int fd = (int)fd_num;
    if (src_addr == NULL || group_addr == NULL) return 0;
#if defined(IP_DROP_SOURCE_MEMBERSHIP)
    struct ip_mreq_source mreq;
    memset(&mreq, 0, sizeof(mreq));
    if (inet_pton(AF_INET, group_addr, &mreq.imr_multiaddr) <= 0) return 0;
    if (inet_pton(AF_INET, src_addr, &mreq.imr_sourceaddr) <= 0) return 0;
    if (iface_addr != NULL && strlen(iface_addr) > 0) {
        inet_pton(AF_INET, iface_addr, &mreq.imr_interface);
    } else {
        mreq.imr_interface.s_addr = htonl(INADDR_ANY);
    }
    setsockopt(fd, IPPROTO_IP, IP_DROP_SOURCE_MEMBERSHIP, (const char *)&mreq, sizeof(mreq));
#endif
    return 0;
}

int scriptgo_dgram_connect(double fd_num, const char *address, double port_num) {
    int fd = (int)fd_num;
    int port = (int)port_num;
    if (fd < 0 || address == NULL) return 0;

    struct addrinfo hints;
    struct addrinfo *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_DGRAM;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    int rc = getaddrinfo(address, port_str, &hints, &res);
    if (rc == 0 && res != NULL) {
        connect(fd, res->ai_addr, res->ai_addrlen);
        freeaddrinfo(res);
    }
    return 0;
}

int scriptgo_dgram_disconnect(double fd_num) {
    int fd = (int)fd_num;
    if (fd < 0) return 0;
    struct sockaddr sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_family = AF_UNSPEC;
    connect(fd, &sa, sizeof(sa));
    return 0;
}

int scriptgo_dgram_close(double fd_num) {
    int fd = (int)fd_num;
    if (fd >= 0) {
#if !defined(_WIN32)
        close(fd);
#else
        closesocket(fd);
#endif
    }
    return 0;
}
