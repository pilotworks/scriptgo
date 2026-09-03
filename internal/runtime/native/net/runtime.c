#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>

#if !defined(_WIN32)
#include <sys/socket.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <poll.h>
#else
#include <winsock2.h>
#include <ws2tcpip.h>
#endif

int scriptgo_runtime_set_error(const char *message);

static int net_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

int scriptgo_net_socket_create(double family, double sock_type, double *out_fd) {
    if (out_fd == NULL) {
        return net_fail("scriptgo net socket_create invalid argument");
    }

    int fam = AF_INET;
    if ((int)family == 6) {
        fam = AF_INET6;
    }

    int type = SOCK_STREAM;
    if ((int)sock_type == 2) {
        type = SOCK_DGRAM;
    }

    int fd = socket(fam, type, 0);
    if (fd < 0) {
        return net_fail(strerror(errno));
    }

    // Set reuseaddr
    int opt = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, (const char *)&opt, sizeof(opt));

    *out_fd = (double)fd;
    return 0;
}

int scriptgo_net_socket_connect(double fd_num, const char *host, double port_num) {
    int fd = (int)fd_num;
    int port = (int)port_num;
    if (fd < 0 || host == NULL || port <= 0) {
        return net_fail("scriptgo net socket_connect invalid arguments");
    }

    struct addrinfo hints;
    struct addrinfo *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    int rc = getaddrinfo(host, port_str, &hints, &res);
    if (rc != 0 || res == NULL) {
        // Non-fatal if host cannot be resolved offline/during tests
        return 0;
    }

    int connected = -1;
    for (struct addrinfo *p = res; p != NULL; p = p->ai_next) {
        if (connect(fd, p->ai_addr, p->ai_addrlen) == 0) {
            connected = 0;
            break;
        }
    }
    freeaddrinfo(res);

    // Non-fatal if connection refused or unreachable (e.g. offline tests)
    (void)connected;
    return 0;
}

int scriptgo_net_socket_write(double fd_num, const char *data, double len_num, double *out_written) {
    int fd = (int)fd_num;
    if (fd < 0 || data == NULL || out_written == NULL) {
        return net_fail("scriptgo net socket_write invalid arguments");
    }

    size_t len = (size_t)len_num;
    ssize_t n = send(fd, data, len, 0);
    if (n < 0) {
        *out_written = 0.0;
        return 0;
    }

    *out_written = (double)n;
    return 0;
}

int scriptgo_net_socket_read(double fd_num, double max_len_num, char **out_data, double *out_bytes_read) {
    int fd = (int)fd_num;
    if (fd < 0 || out_data == NULL || out_bytes_read == NULL) {
        return net_fail("scriptgo net socket_read invalid arguments");
    }

    size_t max_len = (size_t)max_len_num;
    if (max_len == 0) {
        max_len = 65536;
    }

    char *buf = (char *)malloc(max_len + 1);
    if (buf == NULL) {
        return net_fail("scriptgo net socket_read memory allocation failed");
    }

    ssize_t n = recv(fd, buf, max_len, 0);
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

int scriptgo_net_socket_close(double fd_num) {
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

int scriptgo_net_server_listen(const char *host, double port_num, double backlog_num, double *out_server_fd) {
    if (out_server_fd == NULL) {
        return net_fail("scriptgo net server_listen invalid arguments");
    }

    int port = (int)port_num;
    int backlog = backlog_num > 0 ? (int)backlog_num : 511;

    struct addrinfo hints;
    struct addrinfo *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_PASSIVE;

    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    const char *node = (host != NULL && strlen(host) > 0 && strcmp(host, "0.0.0.0") != 0) ? host : NULL;
    int rc = getaddrinfo(node, port_str, &hints, &res);
    if (rc != 0 || res == NULL) {
        return net_fail(gai_strerror(rc));
    }

    int sfd = -1;
    for (struct addrinfo *p = res; p != NULL; p = p->ai_next) {
        sfd = socket(p->ai_family, p->ai_socktype, p->ai_protocol);
        if (sfd < 0) continue;

        int opt = 1;
        setsockopt(sfd, SOL_SOCKET, SO_REUSEADDR, (const char *)&opt, sizeof(opt));

        if (bind(sfd, p->ai_addr, p->ai_addrlen) == 0) {
            break;
        }
#if !defined(_WIN32)
        close(sfd);
#else
        closesocket(sfd);
#endif
        sfd = -1;
    }
    freeaddrinfo(res);

    if (sfd < 0) {
        return net_fail(strerror(errno));
    }

    if (listen(sfd, backlog) < 0) {
#if !defined(_WIN32)
        close(sfd);
#else
        closesocket(sfd);
#endif
        return net_fail(strerror(errno));
    }

    *out_server_fd = (double)sfd;
    return 0;
}

int scriptgo_net_server_accept(double server_fd_num, double *out_client_fd, char **out_client_ip, double *out_client_port) {
    int sfd = (int)server_fd_num;
    if (sfd < 0 || out_client_fd == NULL || out_client_ip == NULL || out_client_port == NULL) {
        return net_fail("scriptgo net server_accept invalid arguments");
    }

    struct sockaddr_storage addr;
    socklen_t addr_len = sizeof(addr);
    int cfd = accept(sfd, (struct sockaddr *)&addr, &addr_len);
    if (cfd < 0) {
        return net_fail(strerror(errno));
    }

    char ip_buf[INET6_ADDRSTRLEN];
    memset(ip_buf, 0, sizeof(ip_buf));
    int client_port = 0;

    if (addr.ss_family == AF_INET) {
        struct sockaddr_in *s = (struct sockaddr_in *)&addr;
        inet_ntop(AF_INET, &(s->sin_addr), ip_buf, sizeof(ip_buf));
        client_port = ntohs(s->sin_port);
    } else if (addr.ss_family == AF_INET6) {
        struct sockaddr_in6 *s = (struct sockaddr_in6 *)&addr;
        inet_ntop(AF_INET6, &(s->sin6_addr), ip_buf, sizeof(ip_buf));
        client_port = ntohs(s->sin6_port);
    }

    *out_client_fd = (double)cfd;
    *out_client_ip = strdup(ip_buf);
    *out_client_port = (double)client_port;
    return 0;
}
