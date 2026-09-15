#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>

#define SCRIPTGO_WS_CONNECTING 0
#define SCRIPTGO_WS_OPEN       1
#define SCRIPTGO_WS_CLOSING    2
#define SCRIPTGO_WS_CLOSED     3

int scriptgo_websocket_connect(const char *url, const char *protocol, double *out_handle);
int scriptgo_websocket_send_text(double handle, const char *data, double *out_sent);
int scriptgo_websocket_send_binary(double handle, const void *data, double length, double *out_sent);
int scriptgo_websocket_close(double handle, double code, const char *reason);
int scriptgo_websocket_poll(double handle, double *out_event_type, char **out_data, double *out_code, char **out_reason);
int scriptgo_websocket_ready_state(double handle, double *out_state);

#if !defined(_WIN32)
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <poll.h>
#include <signal.h>
#else
#include <winsock2.h>
#include <ws2tcpip.h>
#endif

#if defined(SO_NOSIGPIPE)
#define WS_HAVE_NOSIGPIPE 1
#endif

#if defined(MSG_NOSIGNAL)
#define WS_SEND_FLAGS MSG_NOSIGNAL
#else
#define WS_SEND_FLAGS 0
#endif

int scriptgo_runtime_set_error(const char *message);

static int ws_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

// -------------------------------------------------------------
// Base64 encoding table and helper
// -------------------------------------------------------------
static const char ws_b64_table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

static void b64_encode(const unsigned char *data, size_t input_length, char *encoded_data) {
    size_t i = 0, j = 0;
    for (; i < input_length;) {
        size_t remaining = input_length - i;
        unsigned char octet_a = data[i++];
        unsigned char octet_b = remaining > 1 ? data[i++] : 0;
        unsigned char octet_c = remaining > 2 ? data[i++] : 0;
        uint32_t triple = ((uint32_t)octet_a << 16) | ((uint32_t)octet_b << 8) | octet_c;

        encoded_data[j++] = ws_b64_table[(triple >> 18) & 0x3F];
        encoded_data[j++] = ws_b64_table[(triple >> 12) & 0x3F];
        encoded_data[j++] = remaining > 1 ? ws_b64_table[(triple >> 6) & 0x3F] : '=';
        encoded_data[j++] = remaining > 2 ? ws_b64_table[triple & 0x3F] : '=';
    }
    encoded_data[j] = '\0';
}

// -------------------------------------------------------------
// RFC 3174 SHA-1 Implementation
// -------------------------------------------------------------
#define WS_ROTATE_LEFT(val, n) (((val) << (n)) | ((val) >> (32 - (n))))

typedef struct {
    uint32_t state[5];
    uint32_t count[2];
    unsigned char buffer[64];
} ws_sha1_ctx;

static void ws_sha1_transform(uint32_t state[5], const unsigned char buffer[64]) {
    uint32_t a = state[0], b = state[1], c = state[2], d = state[3], e = state[4];
    uint32_t block[80];
    for (int i = 0; i < 16; i++) {
        block[i] = ((uint32_t)buffer[i*4] << 24) | ((uint32_t)buffer[i*4+1] << 16) |
                   ((uint32_t)buffer[i*4+2] << 8)  | ((uint32_t)buffer[i*4+3]);
    }
    for (int i = 16; i < 80; i++) {
        block[i] = WS_ROTATE_LEFT(block[i-3] ^ block[i-8] ^ block[i-14] ^ block[i-16], 1);
    }
    for (int i = 0; i < 80; i++) {
        uint32_t f, k;
        if (i < 20) { f = (b & c) | ((~b) & d); k = 0x5A827999; }
        else if (i < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
        else if (i < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
        else { f = b ^ c ^ d; k = 0xCA62C1D6; }
        uint32_t temp = WS_ROTATE_LEFT(a, 5) + f + e + k + block[i];
        e = d; d = c; c = WS_ROTATE_LEFT(b, 30); b = a; a = temp;
    }
    state[0] += a; state[1] += b; state[2] += c; state[3] += d; state[4] += e;
}

static void ws_sha1_init(ws_sha1_ctx *ctx) {
    ctx->state[0] = 0x67452301; ctx->state[1] = 0xEFCDAB89;
    ctx->state[2] = 0x98BADCFE; ctx->state[3] = 0x10325476; ctx->state[4] = 0xC3D2E1F0;
    ctx->count[0] = ctx->count[1] = 0;
}

static void ws_sha1_update(ws_sha1_ctx *ctx, const unsigned char *data, size_t len) {
    size_t i, j = (ctx->count[0] >> 3) & 63;
    if ((ctx->count[0] += (uint32_t)(len << 3)) < (uint32_t)(len << 3)) ctx->count[1]++;
    ctx->count[1] += (uint32_t)(len >> 29);
    if ((j + len) > 63) {
        memcpy(&ctx->buffer[j], data, (i = 64 - j));
        ws_sha1_transform(ctx->state, ctx->buffer);
        for (; i + 63 < len; i += 64) ws_sha1_transform(ctx->state, &data[i]);
        j = 0;
    } else i = 0;
    memcpy(&ctx->buffer[j], &data[i], len - i);
}

static void ws_sha1_final(unsigned char digest[20], ws_sha1_ctx *ctx) {
    unsigned char finalcount[8];
    for (int i = 0; i < 8; i++) {
        finalcount[i] = (unsigned char)((ctx->count[(i >= 4 ? 0 : 1)] >> ((3 - (i & 3)) * 8)) & 255);
    }
    ws_sha1_update(ctx, (const unsigned char *)"\200", 1);
    while ((ctx->count[0] & 504) != 448) ws_sha1_update(ctx, (const unsigned char *)"\0", 1);
    ws_sha1_update(ctx, finalcount, 8);
    for (int i = 0; i < 20; i++) {
        digest[i] = (unsigned char)((ctx->state[i >> 2] >> ((3 - (i & 3)) * 8)) & 255);
    }
}

// -------------------------------------------------------------
// Connection Struct & State
// -------------------------------------------------------------
#define WS_MAX_CLIENTS 64
#define WS_GUID "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

typedef struct {
    int active;
    int fd;
    int ready_state;
    char *expected_accept;
    unsigned char recv_buf[16384];
    size_t recv_len;
    int close_code;
    char *close_reason;
    int pending_open_event;
    int pending_close_event;
    int pending_error_event;
    char *pending_error_message;
} scriptgo_ws_client;

static scriptgo_ws_client clients[WS_MAX_CLIENTS];
static int next_client_slot = 0;

static void compute_ws_accept(const char *key, char *out_b64) {
    char combined[128];
    snprintf(combined, sizeof(combined), "%s%s", key, WS_GUID);
    ws_sha1_ctx ctx;
    ws_sha1_init(&ctx);
    ws_sha1_update(&ctx, (const unsigned char *)combined, strlen(combined));
    unsigned char digest[20];
    ws_sha1_final(digest, &ctx);
    b64_encode(digest, 20, out_b64);
}

// -------------------------------------------------------------
// WebSocket Protocol Frame Parser & Framer (RFC 6455)
// -------------------------------------------------------------
static int ws_send_frame(int fd, uint8_t opcode, const unsigned char *payload, size_t len) {
    if (fd < 0) return -1;
    unsigned char header[14];
    size_t header_len = 2;
    header[0] = 0x80 | (opcode & 0x0F); // FIN + opcode

    // Client MUST mask frames
    header[1] = 0x80;
    if (len < 126) {
        header[1] |= (uint8_t)len;
    } else if (len <= 65535) {
        header[1] |= 126;
        header[2] = (uint8_t)((len >> 8) & 0xFF);
        header[3] = (uint8_t)(len & 0xFF);
        header_len = 4;
    } else {
        header[1] |= 127;
        uint64_t l64 = (uint64_t)len;
        for (int i = 0; i < 8; i++) {
            header[2 + i] = (uint8_t)((l64 >> ((7 - i) * 8)) & 0xFF);
        }
        header_len = 10;
    }

    // Generate 4-byte random masking key
    uint8_t mask[4];
    for (int i = 0; i < 4; i++) {
        mask[i] = (uint8_t)(rand() & 0xFF);
        header[header_len + i] = mask[i];
    }
    header_len += 4;

    // Send header
    if (send(fd, (const char *)header, header_len, WS_SEND_FLAGS) < 0) {
        return -1;
    }

    // Mask payload and send
    if (len > 0) {
        unsigned char *masked = (unsigned char *)malloc(len);
        if (masked == NULL) return -1;
        for (size_t i = 0; i < len; i++) {
            masked[i] = payload[i] ^ mask[i % 4];
        }
        ssize_t sent = send(fd, (const char *)masked, len, WS_SEND_FLAGS);
        free(masked);
        if (sent < 0) return -1;
    }
    return 0;
}

// -------------------------------------------------------------
// scriptgo_websocket_connect
// -------------------------------------------------------------
int scriptgo_websocket_connect(const char *url, const char *protocol, double *out_handle) {
    if (url == NULL || out_handle == NULL) {
        return ws_fail("WebSocket: missing URL");
    }

    int slot = -1;
    for (int i = 0; i < WS_MAX_CLIENTS; i++) {
        int idx = (next_client_slot + i) % WS_MAX_CLIENTS;
        if (!clients[idx].active) {
            slot = idx;
            next_client_slot = (idx + 1) % WS_MAX_CLIENTS;
            break;
        }
    }
    if (slot < 0) {
        return ws_fail("WebSocket: maximum concurrent connections reached");
    }

    scriptgo_ws_client *cli = &clients[slot];
    memset(cli, 0, sizeof(*cli));
    cli->active = 1;
    cli->ready_state = SCRIPTGO_WS_CONNECTING;
    cli->close_code = 1000;
    cli->close_reason = strdup("");

    // Parse ws://host[:port]/path
    int use_tls = 0;
    const char *p = url;
    if (strncmp(p, "ws://", 5) == 0) {
        p += 5;
    } else if (strncmp(p, "wss://", 6) == 0) {
        p += 6;
        use_tls = 1;
    } else {
        cli->active = 0;
        return ws_fail("WebSocket: URL must start with ws:// or wss://");
    }

    char host[256] = {0};
    int port = use_tls ? 443 : 80;
    const char *slash = strchr(p, '/');
    const char *colon = strchr(p, ':');
    const char *path = slash ? slash : "/";

    if (colon && (!slash || colon < slash)) {
        size_t hlen = colon - p;
        if (hlen >= sizeof(host)) hlen = sizeof(host) - 1;
        strncpy(host, p, hlen);
        port = atoi(colon + 1);
    } else if (slash) {
        size_t hlen = slash - p;
        if (hlen >= sizeof(host)) hlen = sizeof(host) - 1;
        strncpy(host, p, hlen);
    } else {
        strncpy(host, p, sizeof(host) - 1);
    }

    // Connect socket
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        cli->pending_error_event = 1;
        cli->pending_error_message = strdup("socket creation failed");
        cli->ready_state = SCRIPTGO_WS_CLOSED;
        *out_handle = (double)slot;
        return 0;
    }

    struct addrinfo hints, *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);

    if (getaddrinfo(host, port_str, &hints, &res) != 0 || res == NULL) {
        close(fd);
        cli->pending_error_event = 1;
        cli->pending_error_message = strdup("DNS resolution failed");
        cli->ready_state = SCRIPTGO_WS_CLOSED;
        *out_handle = (double)slot;
        return 0;
    }
    cli->fd = fd;

#if defined(WS_HAVE_NOSIGPIPE)
    int set_nosigpipe = 1;
    setsockopt(fd, SOL_SOCKET, SO_NOSIGPIPE, (void *)&set_nosigpipe, sizeof(int));
#endif

    // Set non-blocking socket BEFORE connect so connect doesn't block and initial state is CONNECTING
    int flags = fcntl(fd, F_GETFL, 0);
    fcntl(fd, F_SETFL, flags | O_NONBLOCK);

    int connected = -1;
    for (struct addrinfo *rp = res; rp != NULL; rp = rp->ai_next) {
        int rc = connect(fd, rp->ai_addr, rp->ai_addrlen);
        if (rc == 0 || (rc < 0 && (errno == EINPROGRESS || errno == EALREADY || errno == EWOULDBLOCK))) {
            connected = 0;
            break;
        }
    }
    freeaddrinfo(res);

#if !defined(_WIN32)
    signal(SIGPIPE, SIG_IGN);
#endif

    if (connected != 0) {
        cli->pending_error_event = 1;
        cli->pending_error_message = strdup("connection refused");
    } else {
        // Generate random 16-byte Sec-WebSocket-Key
        unsigned char raw_key[16];
        for (int i = 0; i < 16; i++) raw_key[i] = (unsigned char)(rand() & 0xFF);
        char ws_key[32];
        b64_encode(raw_key, 16, ws_key);

        char exp_accept[64];
        compute_ws_accept(ws_key, exp_accept);
        cli->expected_accept = strdup(exp_accept);

        // Send HTTP Upgrade handshake
        char handshake[1024];
        int hlen = snprintf(handshake, sizeof(handshake),
            "GET %s HTTP/1.1\r\n"
            "Host: %s:%d\r\n"
            "Upgrade: websocket\r\n"
            "Connection: Upgrade\r\n"
            "Sec-WebSocket-Key: %s\r\n"
            "Sec-WebSocket-Version: 13\r\n",
            path, host, port, ws_key);

        if (protocol != NULL && strlen(protocol) > 0) {
            hlen += snprintf(handshake + hlen, sizeof(handshake) - hlen,
                "Sec-WebSocket-Protocol: %s\r\n", protocol);
        }
        hlen += snprintf(handshake + hlen, sizeof(handshake) - hlen, "\r\n");

        send(fd, handshake, (size_t)hlen, WS_SEND_FLAGS);
    }

    *out_handle = (double)slot;
    return 0;
}

// -------------------------------------------------------------
// scriptgo_websocket_send_text & binary
// -------------------------------------------------------------
int scriptgo_websocket_send_text(double handle, const char *data, double *out_sent) {
    int slot = (int)handle;
    if (out_sent) *out_sent = 0.0;
    if (slot < 0 || slot >= WS_MAX_CLIENTS || !clients[slot].active || data == NULL) {
        return ws_fail("WebSocket.send: invalid state or handle");
    }
    scriptgo_ws_client *cli = &clients[slot];
    if (cli->ready_state != SCRIPTGO_WS_OPEN) {
        return ws_fail("WebSocket is not open");
    }

    size_t len = strlen(data);
    if (ws_send_frame(cli->fd, 0x01, (const unsigned char *)data, len) != 0) {
        return ws_fail("WebSocket send text failed");
    }
    if (out_sent) *out_sent = (double)len;
    return 0;
}

int scriptgo_websocket_send_binary(double handle, const void *data, double length, double *out_sent) {
    int slot = (int)handle;
    if (out_sent) *out_sent = 0.0;
    if (slot < 0 || slot >= WS_MAX_CLIENTS || !clients[slot].active) {
        return ws_fail("WebSocket.send: invalid state or handle");
    }
    scriptgo_ws_client *cli = &clients[slot];
    if (cli->ready_state != SCRIPTGO_WS_OPEN) {
        return ws_fail("WebSocket is not open");
    }

    size_t len = (size_t)length;
    if (ws_send_frame(cli->fd, 0x02, (const unsigned char *)data, len) != 0) {
        return ws_fail("WebSocket send binary failed");
    }
    if (out_sent) *out_sent = (double)len;
    return 0;
}

// -------------------------------------------------------------
// scriptgo_websocket_close
// -------------------------------------------------------------
int scriptgo_websocket_close(double handle, double code, const char *reason) {
    int slot = (int)handle;
    if (slot < 0 || slot >= WS_MAX_CLIENTS || !clients[slot].active) {
        return 0;
    }
    scriptgo_ws_client *cli = &clients[slot];
    if (cli->ready_state == SCRIPTGO_WS_CLOSING || cli->ready_state == SCRIPTGO_WS_CLOSED) {
        return 0;
    }

    uint16_t c = (code > 0) ? (uint16_t)code : 1000;
    cli->close_code = (int)c;
    if (cli->close_reason) free(cli->close_reason);
    cli->close_reason = reason ? strdup(reason) : strdup("");

    // Send Close Frame (0x08)
    unsigned char payload[128];
    payload[0] = (uint8_t)((c >> 8) & 0xFF);
    payload[1] = (uint8_t)(c & 0xFF);
    size_t rlen = reason ? strlen(reason) : 0;
    if (rlen > 123) rlen = 123;
    if (rlen > 0) memcpy(payload + 2, reason, rlen);

    if (cli->fd >= 0) {
        ws_send_frame(cli->fd, 0x08, payload, 2 + rlen);
    }
    cli->ready_state = SCRIPTGO_WS_CLOSING;
    cli->pending_close_event = 1;
    return 0;
}

// -------------------------------------------------------------
// scriptgo_websocket_poll (Event Loop Hook)
// out_event_type: 0 = none, 1 = open, 2 = message, 3 = close, 4 = error
// -------------------------------------------------------------
int scriptgo_websocket_poll(double handle, double *out_event_type, char **out_data, double *out_code, char **out_reason) {
    int slot = (int)handle;
    if (out_event_type) *out_event_type = 0.0;
    if (out_data) *out_data = strdup("");
    if (out_code) *out_code = 1000.0;
    if (out_reason) *out_reason = strdup("");

    if (slot < 0 || slot >= WS_MAX_CLIENTS || !clients[slot].active) {
        return 0;
    }
    scriptgo_ws_client *cli = &clients[slot];

    // Priority 1: Pending error
    if (cli->pending_error_event) {
        cli->pending_error_event = 0;
        if (out_event_type) *out_event_type = 4.0;
        if (out_data && cli->pending_error_message) {
            free(*out_data);
            *out_data = strdup(cli->pending_error_message);
        }
        return 0;
    }

    // Priority 2: Pending open
    if (cli->pending_open_event) {
        cli->pending_open_event = 0;
        if (out_event_type) *out_event_type = 1.0;
        return 0;
    }

    // Priority 3: Pending close
    if (cli->pending_close_event) {
        cli->pending_close_event = 0;
        cli->ready_state = SCRIPTGO_WS_CLOSED;
        if (out_event_type) *out_event_type = 3.0;
        if (out_code) *out_code = (double)cli->close_code;
        if (out_reason) {
            free(*out_reason);
            *out_reason = strdup(cli->close_reason ? cli->close_reason : "");
        }
        if (cli->fd >= 0) {
            close(cli->fd);
            cli->fd = -1;
        }
        return 0;
    }

    // If socket is closed or disconnected, no IO
    if (cli->fd < 0 || cli->ready_state == SCRIPTGO_WS_CLOSED) {
        return 0;
    }

    // Non-blocking read from socket
    unsigned char temp[4096];
    ssize_t n = recv(cli->fd, (char *)temp, sizeof(temp), 0);
    if (n > 0) {
        if (cli->recv_len + (size_t)n <= sizeof(cli->recv_buf)) {
            memcpy(cli->recv_buf + cli->recv_len, temp, (size_t)n);
            cli->recv_len += (size_t)n;
        }
    } else if (n == 0) {
        // Peer disconnected
        cli->ready_state = SCRIPTGO_WS_CLOSED;
        if (out_event_type) *out_event_type = 3.0;
        if (out_code) *out_code = 1006.0; // Abnormal closure
        if (out_reason) {
            free(*out_reason);
            *out_reason = strdup("peer closed connection");
        }
        close(cli->fd);
        cli->fd = -1;
        return 0;
    } else {
        if (errno != EAGAIN && errno != EWOULDBLOCK) {
            cli->ready_state = SCRIPTGO_WS_CLOSED;
            if (out_event_type) *out_event_type = 4.0;
            if (out_data) {
                free(*out_data);
                *out_data = strdup(strerror(errno));
            }
            close(cli->fd);
            cli->fd = -1;
            return 0;
        }
    }

    // If in CONNECTING state: parse HTTP Handshake response
    if (cli->ready_state == SCRIPTGO_WS_CONNECTING) {
        char *header_end = strstr((char *)cli->recv_buf, "\r\n\r\n");
        if (header_end != NULL) {
            size_t header_len = (size_t)(header_end - (char *)cli->recv_buf) + 4;
            // Check for HTTP/1.1 101 Switching Protocols
            if (strstr((char *)cli->recv_buf, "101 Switching Protocols") != NULL) {
                cli->ready_state = SCRIPTGO_WS_OPEN;
                // Move leftover bytes forward
                size_t leftover = cli->recv_len - header_len;
                if (leftover > 0) {
                    memmove(cli->recv_buf, cli->recv_buf + header_len, leftover);
                }
                cli->recv_len = leftover;

                if (out_event_type) *out_event_type = 1.0; // "open"
                return 0;
            } else {
                cli->ready_state = SCRIPTGO_WS_CLOSED;
                close(cli->fd);
                cli->fd = -1;
                if (out_event_type) *out_event_type = 4.0;
                if (out_data) {
                    free(*out_data);
                    *out_data = strdup("handshake failed: response not 101");
                }
                return 0;
            }
        }
        return 0;
    }

    // In OPEN or CLOSING state: parse WebSocket frame(s)
    if (cli->recv_len >= 2) {
        uint8_t b0 = cli->recv_buf[0];
        uint8_t b1 = cli->recv_buf[1];
        uint8_t opcode = b0 & 0x0F;
        int has_mask = (b1 & 0x80) != 0;
        uint64_t payload_len = b1 & 0x7F;
        size_t offset = 2;

        if (payload_len == 126) {
            if (cli->recv_len < 4) return 0;
            payload_len = ((uint64_t)cli->recv_buf[2] << 8) | cli->recv_buf[3];
            offset = 4;
        } else if (payload_len == 127) {
            if (cli->recv_len < 10) return 0;
            payload_len = 0;
            for (int i = 0; i < 8; i++) {
                payload_len = (payload_len << 8) | cli->recv_buf[2 + i];
            }
            offset = 10;
        }

        uint8_t mask[4] = {0};
        if (has_mask) {
            if (cli->recv_len < offset + 4) return 0;
            memcpy(mask, cli->recv_buf + offset, 4);
            offset += 4;
        }

        if (cli->recv_len >= offset + (size_t)payload_len) {
            unsigned char *frame_data = (unsigned char *)malloc((size_t)payload_len + 1);
            if (frame_data == NULL) return 0;
            memcpy(frame_data, cli->recv_buf + offset, (size_t)payload_len);
            if (has_mask) {
                for (size_t i = 0; i < (size_t)payload_len; i++) {
                    frame_data[i] ^= mask[i % 4];
                }
            }
            frame_data[payload_len] = '\0';

            // Shift buffer
            size_t total_frame_size = offset + (size_t)payload_len;
            size_t leftover = cli->recv_len - total_frame_size;
            if (leftover > 0) {
                memmove(cli->recv_buf, cli->recv_buf + total_frame_size, leftover);
            }
            cli->recv_len = leftover;

            // Handle Opcode
            if (opcode == 0x01 || opcode == 0x02) { // Text or Binary
                if (out_event_type) *out_event_type = 2.0; // "message"
                if (out_data) {
                    free(*out_data);
                    *out_data = (char *)frame_data;
                } else {
                    free(frame_data);
                }
                return 0;
            } else if (opcode == 0x08) { // Close
                uint16_t code = 1000;
                char reason_buf[128] = {0};
                if (payload_len >= 2) {
                    code = ((uint16_t)frame_data[0] << 8) | frame_data[1];
                    size_t rlen = (size_t)payload_len - 2;
                    if (rlen > sizeof(reason_buf) - 1) rlen = sizeof(reason_buf) - 1;
                    memcpy(reason_buf, frame_data + 2, rlen);
                }
                free(frame_data);

                // Echo close frame if we haven't closed yet
                if (cli->ready_state != SCRIPTGO_WS_CLOSING && cli->fd >= 0) {
                    ws_send_frame(cli->fd, 0x08, (const unsigned char *)"\x03\xe8", 2);
                }
                cli->ready_state = SCRIPTGO_WS_CLOSED;
                close(cli->fd);
                cli->fd = -1;

                if (out_event_type) *out_event_type = 3.0; // "close"
                if (out_code) *out_code = (double)code;
                if (out_reason) {
                    free(*out_reason);
                    *out_reason = strdup(reason_buf);
                }
                return 0;
            } else if (opcode == 0x09) { // Ping -> reply Pong
                if (cli->fd >= 0) {
                    ws_send_frame(cli->fd, 0x0A, frame_data, (size_t)payload_len);
                }
                free(frame_data);
                return 0;
            } else if (opcode == 0x0A) { // Pong
                free(frame_data);
                return 0;
            } else {
                free(frame_data);
            }
        }
    }
    return 0;
}

int scriptgo_websocket_ready_state(double handle, double *out_state) {
    int slot = (int)handle;
    if (out_state) *out_state = (double)SCRIPTGO_WS_CLOSED;
    if (slot < 0 || slot >= WS_MAX_CLIENTS || !clients[slot].active) {
        return 0;
    }
    if (out_state) *out_state = (double)clients[slot].ready_state;
    return 0;
}
