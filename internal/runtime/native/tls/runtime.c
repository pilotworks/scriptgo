#include <errno.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#if defined(_WIN32)
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#else
#include <arpa/inet.h>
#include <dirent.h>
#include <netdb.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <unistd.h>
#endif

#if defined(__APPLE__)
#include <Security/Security.h>
#endif

#if defined(SCRIPTGO_HAS_OPENSSL)

int scriptgo_runtime_set_error(const char *message);

static int tls_fail(const char *message) {
    return scriptgo_runtime_set_error(message);
}

static int tls_integer_in_range(double value, double minimum, double maximum) {
    if (value != value || value < minimum || value > maximum) return 0;
    return value == (double)(int64_t)value;
}

#include <openssl/err.h>
#include <openssl/core_names.h>
#include <openssl/evp.h>
#include <openssl/objects.h>
#include <openssl/pem.h>
#include <openssl/params.h>
#include <openssl/rand.h>
#include <openssl/ssl.h>
#include <openssl/x509.h>
#include <openssl/x509v3.h>
#if OPENSSL_VERSION_NUMBER < 0x30000000L
#include <openssl/hmac.h>
#endif

typedef struct scriptgo_tls_context scriptgo_tls_context;
typedef struct scriptgo_tls_socket scriptgo_tls_socket;
typedef struct scriptgo_tls_sni_entry scriptgo_tls_sni_entry;

extern const char *scriptgo_tls_bundled_root_certificates[];

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_tls_typed_array_view;

struct scriptgo_tls_context {
    SSL_CTX *ctx;
};

struct scriptgo_tls_socket {
    SSL *ssl;
    int fd;
    int is_pair;
    int refs;
    int close_count;
    int max_send_fragment;
};

#define SCRIPTGO_TLS_MAGIC_TYPEDARRAY 0x54415252
#define SCRIPTGO_TLS_MAGIC_BUFFER 0x42554646

static int tls_openssl_fail(const char *prefix);

static int tls_is_byte_view(const scriptgo_tls_typed_array_view *view) {
    return view != NULL && (view->magic == SCRIPTGO_TLS_MAGIC_TYPEDARRAY || view->magic == SCRIPTGO_TLS_MAGIC_BUFFER);
}

static int tls_set_session_from_view(SSL *ssl, void *session_data, int require_non_empty) {
    scriptgo_tls_typed_array_view *view = (scriptgo_tls_typed_array_view *)session_data;
    const unsigned char *cursor;
    SSL_SESSION *session;
    int result;

    if (session_data == NULL) return require_non_empty ? tls_fail("TLS session is missing") : 0;
    if (!tls_is_byte_view(view) || view->length < 0 || view->element_size != 1 ||
        (view->length != 0 && view->data == NULL)) {
        return tls_fail("TLS session must be a byte view");
    }
    if (view->length == 0) return require_non_empty ? tls_fail("TLS session is empty") : 0;
    if ((uint64_t)view->length > (uint64_t)LONG_MAX) return tls_fail("TLS session is too large");
    cursor = view->data;
    session = d2i_SSL_SESSION(NULL, &cursor, (long)view->length);
    if (session == NULL || cursor != view->data + view->length) {
        SSL_SESSION_free(session);
        return tls_openssl_fail("TLS session parsing failed");
    }
    result = SSL_set_session(ssl, session);
    SSL_SESSION_free(session);
    return result == 1 ? 0 : tls_openssl_fail("TLS session configuration failed");
}

struct scriptgo_tls_sni_entry {
    char *hostname;
    SSL_CTX *ctx;
    scriptgo_tls_sni_entry *next;
};

typedef struct {
    SSL_CTX *ctx;
    int fd;
    int request_cert;
    int reject;
    unsigned char ticket_keys[48];
    scriptgo_tls_sni_entry *sni;
} scriptgo_tls_server;

static int tls_ticket_ex_index = -1;

static int tls_ticket_ex_index_init(void) {
    if (tls_ticket_ex_index >= 0) return 0;
    tls_ticket_ex_index = SSL_get_ex_new_index(0, NULL, NULL, NULL, NULL);
    return tls_ticket_ex_index >= 0 ? 0 : -1;
}

#if OPENSSL_VERSION_NUMBER >= 0x30000000L
static int tls_ticket_key_callback(SSL *ssl, unsigned char *key_name, unsigned char *iv,
                                   EVP_CIPHER_CTX *cipher, EVP_MAC_CTX *mac, int encrypt) {
    scriptgo_tls_server *server = (scriptgo_tls_server *)SSL_get_ex_data(ssl, tls_ticket_ex_index);
    const EVP_CIPHER *ticket_cipher = EVP_aes_128_cbc();
    char digest_name[] = "SHA256";
    OSSL_PARAM params[3];

    if (server == NULL || key_name == NULL || iv == NULL || cipher == NULL || mac == NULL) return -1;
    params[0] = OSSL_PARAM_construct_octet_string(OSSL_MAC_PARAM_KEY, server->ticket_keys + 16, 16);
    params[1] = OSSL_PARAM_construct_utf8_string(OSSL_MAC_PARAM_DIGEST, digest_name, 0);
    params[2] = OSSL_PARAM_construct_end();
    if (encrypt) {
        if (RAND_bytes(iv, EVP_CIPHER_iv_length(ticket_cipher)) != 1) return -1;
        memcpy(key_name, server->ticket_keys, 16);
        if (EVP_EncryptInit_ex(cipher, ticket_cipher, NULL, server->ticket_keys + 32, iv) != 1) return -1;
    } else {
        if (memcmp(key_name, server->ticket_keys, 16) != 0) return 0;
        if (EVP_DecryptInit_ex(cipher, ticket_cipher, NULL, server->ticket_keys + 32, iv) != 1) return -1;
    }
    return EVP_MAC_CTX_set_params(mac, params) == 1 ? 1 : -1;
}

static int tls_configure_ticket_callback(SSL_CTX *ctx) {
    if (ctx == NULL || tls_ticket_ex_index_init() != 0) return tls_fail("TLS ticket callback initialization failed");
    return SSL_CTX_set_tlsext_ticket_key_evp_cb(ctx, tls_ticket_key_callback) == 1
        ? 0 : tls_openssl_fail("TLS ticket callback configuration failed");
}
#else
static int tls_ticket_key_callback(SSL *ssl, unsigned char *key_name, unsigned char *iv,
                                   EVP_CIPHER_CTX *cipher, HMAC_CTX *mac, int encrypt) {
    scriptgo_tls_server *server = (scriptgo_tls_server *)SSL_get_ex_data(ssl, tls_ticket_ex_index);
    const EVP_CIPHER *ticket_cipher = EVP_aes_128_cbc();

    if (server == NULL || key_name == NULL || iv == NULL || cipher == NULL || mac == NULL) return -1;
    if (encrypt) {
        if (RAND_bytes(iv, EVP_CIPHER_iv_length(ticket_cipher)) != 1) return -1;
        memcpy(key_name, server->ticket_keys, 16);
        if (EVP_EncryptInit_ex(cipher, ticket_cipher, NULL, server->ticket_keys + 32, iv) != 1) return -1;
    } else {
        if (memcmp(key_name, server->ticket_keys, 16) != 0) return 0;
        if (EVP_DecryptInit_ex(cipher, ticket_cipher, NULL, server->ticket_keys + 32, iv) != 1) return -1;
    }
    return HMAC_Init_ex(mac, server->ticket_keys + 16, 16, EVP_sha256(), NULL) == 1 ? 1 : -1;
}

static int tls_configure_ticket_callback(SSL_CTX *ctx) {
    if (ctx == NULL || tls_ticket_ex_index_init() != 0) return tls_fail("TLS ticket callback initialization failed");
    return SSL_CTX_set_tlsext_ticket_key_cb(ctx, tls_ticket_key_callback) > 0
        ? 0 : tls_openssl_fail("TLS ticket callback configuration failed");
}
#endif

static char *tls_strdup(const char *value) {
    size_t length = value == NULL ? 0 : strlen(value);
    char *copy = (char *)malloc(length + 1);
    if (copy == NULL) return NULL;
    if (length != 0) memcpy(copy, value, length);
    copy[length] = '\0';
    return copy;
}

static int tls_openssl_fail(const char *prefix) {
    char detail[256];
    unsigned long error = ERR_get_error();
    if (error != 0) {
        ERR_error_string_n(error, detail, sizeof(detail));
        char message[512];
        snprintf(message, sizeof(message), "%s: %s", prefix, detail);
        return tls_fail(message);
    }
    return tls_fail(prefix);
}

static int tls_append(char **buffer, size_t *length, size_t *capacity, const char *text) {
    if (buffer == NULL || length == NULL || capacity == NULL || text == NULL) return -1;
    size_t add = strlen(text);
    if (add > SIZE_MAX - *length - 1) return -1;
    size_t required = *length + add + 1;
    if (required > *capacity) {
        size_t next = *capacity == 0 ? 256 : *capacity;
        while (next < required) {
            if (next > SIZE_MAX / 2) {
                next = required;
                break;
            }
            next *= 2;
        }
        char *grown = (char *)realloc(*buffer, next);
        if (grown == NULL) return -1;
        *buffer = grown;
        *capacity = next;
    }
    memcpy(*buffer + *length, text, add);
    *length += add;
    (*buffer)[*length] = '\0';
    return 0;
}

static int tls_append_char(char **buffer, size_t *length, size_t *capacity, char value) {
    char text[2] = {value, '\0'};
    return tls_append(buffer, length, capacity, text);
}

static int tls_append_json_string(char **buffer, size_t *length, size_t *capacity, const char *value) {
    if (tls_append_char(buffer, length, capacity, '"') != 0) return -1;
    if (value != NULL) {
        for (const unsigned char *cursor = (const unsigned char *)value; *cursor != '\0'; cursor++) {
            switch (*cursor) {
            case '"': if (tls_append(buffer, length, capacity, "\\\"") != 0) return -1; break;
            case '\\': if (tls_append(buffer, length, capacity, "\\\\") != 0) return -1; break;
            case '\n': if (tls_append(buffer, length, capacity, "\\n") != 0) return -1; break;
            case '\r': if (tls_append(buffer, length, capacity, "\\r") != 0) return -1; break;
            case '\t': if (tls_append(buffer, length, capacity, "\\t") != 0) return -1; break;
            default:
                if (*cursor < 0x20) {
                    char escaped[7];
                    snprintf(escaped, sizeof(escaped), "\\u%04x", *cursor);
                    if (tls_append(buffer, length, capacity, escaped) != 0) return -1;
                } else if (tls_append_char(buffer, length, capacity, (char)*cursor) != 0) return -1;
            }
        }
    }
    return tls_append_char(buffer, length, capacity, '"');
}

static char *tls_json_empty_object(void) { return tls_strdup("{}"); }

static char *tls_hex(const unsigned char *data, size_t length) {
    if (length > (SIZE_MAX - 1) / 2 || (data == NULL && length != 0)) return NULL;
    char *result = (char *)malloc(length * 2 + 1);
    if (result == NULL) return NULL;
    for (size_t i = 0; i < length; i++) snprintf(result + i * 2, 3, "%02x", data[i]);
    result[length * 2] = '\0';
    return result;
}

static int tls_hex_digit(char value) {
    if (value >= '0' && value <= '9') return value - '0';
    if (value >= 'a' && value <= 'f') return value - 'a' + 10;
    if (value >= 'A' && value <= 'F') return value - 'A' + 10;
    return -1;
}

static int tls_hex_decode(const char *hex, unsigned char *out, size_t length) {
    if (hex == NULL || (out == NULL && length != 0) || length > (SIZE_MAX - 1) / 2 || strlen(hex) != length * 2) return -1;
    for (size_t i = 0; i < length; i++) {
        int high = tls_hex_digit(hex[i * 2]);
        int low = tls_hex_digit(hex[i * 2 + 1]);
        if (high < 0 || low < 0) return -1;
        out[i] = (unsigned char)((high << 4) | low);
    }
    return 0;
}

static int tls_version(const char *value) {
    if (value == NULL || *value == '\0') return 0;
    if (strcmp(value, "TLSv1.2") == 0) return TLS1_2_VERSION;
    if (strcmp(value, "TLSv1.3") == 0) return TLS1_3_VERSION;
    if (strcmp(value, "TLSv1.1") == 0) return TLS1_1_VERSION;
    if (strcmp(value, "TLSv1") == 0 || strcmp(value, "TLSv1.0") == 0) return TLS1_VERSION;
    return -1;
}

static int tls_load_certificate(SSL_CTX *ctx, const char *pem) {
    if (pem == NULL || *pem == '\0') return 0;
    BIO *bio = BIO_new_mem_buf(pem, -1);
    if (bio == NULL) return tls_openssl_fail("TLS certificate allocation failed");
    X509 *certificate = PEM_read_bio_X509(bio, NULL, NULL, NULL);
    BIO_free(bio);
    if (certificate == NULL) return tls_openssl_fail("TLS certificate parsing failed");
    int result = SSL_CTX_use_certificate(ctx, certificate);
    X509_free(certificate);
    return result == 1 ? 0 : tls_openssl_fail("TLS certificate configuration failed");
}

static int tls_load_private_key(SSL_CTX *ctx, const char *pem) {
    if (pem == NULL || *pem == '\0') return 0;
    BIO *bio = BIO_new_mem_buf(pem, -1);
    if (bio == NULL) return tls_openssl_fail("TLS private key allocation failed");
    EVP_PKEY *key = PEM_read_bio_PrivateKey(bio, NULL, NULL, NULL);
    BIO_free(bio);
    if (key == NULL) return tls_openssl_fail("TLS private key parsing failed");
    int result = SSL_CTX_use_PrivateKey(ctx, key);
    EVP_PKEY_free(key);
    if (result != 1) return tls_openssl_fail("TLS private key configuration failed");
    return SSL_CTX_check_private_key(ctx) == 1 ? 0 : tls_openssl_fail("TLS certificate and private key do not match");
}

static int tls_load_ca(SSL_CTX *ctx, const char *pem, int ca_provided) {
    X509_STORE *store = SSL_CTX_get_cert_store(ctx);
    if (ca_provided || (pem != NULL && *pem != '\0')) {
        if (pem == NULL || *pem == '\0') return 0;
        BIO *bio = BIO_new_mem_buf(pem, -1);
        if (bio == NULL) return tls_openssl_fail("TLS CA allocation failed");
        int count = 0;
        X509 *certificate;
        while ((certificate = PEM_read_bio_X509(bio, NULL, NULL, NULL)) != NULL) {
            if (X509_STORE_add_cert(store, certificate) == 1) count++;
            X509_free(certificate);
        }
        BIO_free(bio);
        if (count == 0) return tls_openssl_fail("TLS CA certificate parsing failed");
        return 0;
    }
    if (SSL_CTX_set_default_verify_paths(ctx) != 1) return tls_openssl_fail("TLS default verify paths configuration failed");
    return 0;
}

static int tls_configure_context(SSL_CTX *ctx, const char *cert, const char *key, const char *ca, int ca_provided,
                                 const char *min_version, const char *max_version, const char *ciphers) {
    int min = tls_version(min_version);
    int max = tls_version(max_version);
    if (min < 0 || max < 0 || (min != 0 && max != 0 && min > max)) return tls_fail("TLS protocol version range is invalid");
    if (min != 0 && SSL_CTX_set_min_proto_version(ctx, min) != 1) return tls_openssl_fail("TLS minimum version configuration failed");
    if (max != 0 && SSL_CTX_set_max_proto_version(ctx, max) != 1) return tls_openssl_fail("TLS maximum version configuration failed");
    if (ciphers != NULL && *ciphers != '\0') {
        int modern = SSL_CTX_set_ciphersuites(ctx, ciphers);
        int legacy = SSL_CTX_set_cipher_list(ctx, ciphers);
        if (modern != 1 && legacy != 1) return tls_openssl_fail("TLS cipher configuration failed");
        ERR_clear_error();
    }
    if (tls_load_certificate(ctx, cert) != 0 || tls_load_private_key(ctx, key) != 0) return -1;
    if (tls_load_ca(ctx, ca, ca_provided) != 0) return -1;
    return 0;
}

static scriptgo_tls_context *tls_context_from_handle(double handle) {
    return (scriptgo_tls_context *)(uintptr_t)handle;
}

static scriptgo_tls_socket *tls_socket_from_handle(double handle) {
    return (scriptgo_tls_socket *)(uintptr_t)handle;
}

static scriptgo_tls_server *tls_server_from_handle(double handle) {
    return (scriptgo_tls_server *)(uintptr_t)handle;
}

static SSL_CTX *tls_context_value(double handle) {
    scriptgo_tls_context *context = tls_context_from_handle(handle);
    return context == NULL ? NULL : context->ctx;
}

static void tls_sni_free(scriptgo_tls_sni_entry *entry) {
    while (entry != NULL) {
        scriptgo_tls_sni_entry *next = entry->next;
        free(entry->hostname);
        SSL_CTX_free(entry->ctx);
        free(entry);
        entry = next;
    }
}

static int tls_hostname_equal(const char *left, const char *right) {
#if defined(_WIN32)
    return _stricmp(left, right) == 0;
#else
    return strcasecmp(left, right) == 0;
#endif
}

static const char *tls_signature_name(int signature) {
    switch (signature) {
    case TLSEXT_signature_rsa: return "RSA";
    case TLSEXT_signature_dsa: return "DSA";
    case TLSEXT_signature_ecdsa: return "ECDSA";
    default: {
        const char *name = OBJ_nid2sn(signature);
        return name == NULL ? "UNKNOWN" : name;
    }
    }
}

static const char *tls_hash_name(int hash) {
    switch (hash) {
    case TLSEXT_hash_md5: return "MD5";
    case TLSEXT_hash_sha1: return "SHA1";
    case TLSEXT_hash_sha224: return "SHA224";
    case TLSEXT_hash_sha256: return "SHA256";
    case TLSEXT_hash_sha384: return "SHA384";
    case TLSEXT_hash_sha512: return "SHA512";
    default: return "UNDEF";
    }
}

static void tls_trace_callback(int write_p, int version, int content_type,
                               const void *buffer, size_t length, SSL *ssl, void *argument) {
    (void)ssl;
    (void)argument;
    const unsigned char *bytes = (const unsigned char *)buffer;
    fprintf(stderr, "TLS trace %s version=0x%x type=%d length=%zu",
            write_p ? "write" : "read", version, content_type, length);
    size_t shown = length < 32 ? length : 32;
    if (shown != 0) {
        fputs(" data=", stderr);
        for (size_t i = 0; i < shown; i++) fprintf(stderr, "%02x", bytes[i]);
        if (shown != length) fputs("...", stderr);
    }
    fputc('\n', stderr);
}

static int tls_sni_callback(SSL *ssl, int *alert, void *argument) {
    (void)alert;
    scriptgo_tls_server *server = (scriptgo_tls_server *)argument;
    const char *hostname = SSL_get_servername(ssl, TLSEXT_NAMETYPE_host_name);
    if (server == NULL || hostname == NULL) return SSL_TLSEXT_ERR_NOACK;
    for (scriptgo_tls_sni_entry *entry = server->sni; entry != NULL; entry = entry->next) {
        if (tls_hostname_equal(hostname, entry->hostname)) {
            SSL_set_SSL_CTX(ssl, entry->ctx);
            return SSL_TLSEXT_ERR_OK;
        }
    }
    return SSL_TLSEXT_ERR_NOACK;
}

static int tls_socket_configure_verify(SSL *ssl, const char *servername, int reject) {
    if (reject) {
        SSL_set_verify(ssl, SSL_VERIFY_PEER, NULL);
        if (servername != NULL && *servername != '\0' && SSL_set1_host(ssl, servername) != 1) return -1;
    } else {
        SSL_set_verify(ssl, SSL_VERIFY_NONE, NULL);
    }
    return 0;
}

static int tls_close_fd(int fd) {
    if (fd < 0) return 0;
#if defined(_WIN32)
    return closesocket((SOCKET)fd);
#else
    return close(fd);
#endif
}

static int tls_socket_address(int fd, int local, char *address, size_t address_size, double *port, const char **family) {
    struct sockaddr_storage storage;
    socklen_t length = sizeof(storage);
    int result = local ? getsockname(fd, (struct sockaddr *)&storage, &length) : getpeername(fd, (struct sockaddr *)&storage, &length);
    if (result != 0) return -1;
    if (storage.ss_family == AF_INET) {
        struct sockaddr_in *value = (struct sockaddr_in *)&storage;
        inet_ntop(AF_INET, &value->sin_addr, address, address_size);
        *port = (double)ntohs(value->sin_port);
        *family = "IPv4";
        return 0;
    }
    if (storage.ss_family == AF_INET6) {
        struct sockaddr_in6 *value = (struct sockaddr_in6 *)&storage;
        inet_ntop(AF_INET6, &value->sin6_addr, address, address_size);
        *port = (double)ntohs(value->sin6_port);
        *family = "IPv6";
        return 0;
    }
    return -1;
}

static int tls_socket_new(SSL *ssl, int fd, int is_pair, int refs, double *out_handle) {
    if (ssl == NULL || out_handle == NULL || refs <= 0) return tls_fail("TLS socket creation arguments are invalid");
    scriptgo_tls_socket *socket = (scriptgo_tls_socket *)calloc(1, sizeof(*socket));
    if (socket == NULL) return tls_fail("TLS socket allocation failed");
    socket->ssl = ssl;
    socket->fd = fd;
    socket->is_pair = is_pair;
    socket->refs = refs;
    socket->close_count = 0;
    socket->max_send_fragment = 16384;
    *out_handle = (double)(uintptr_t)socket;
    return 0;
}

int scriptgo_tls_context_create(const char *cert, const char *key, const char *ca,
                               const char *min_version, const char *max_version,
                               const char *ciphers, double ca_provided, double *out_handle) {
    if (out_handle == NULL) return tls_fail("TLS context output is null");
    SSL_CTX *ssl_context = SSL_CTX_new(TLS_method());
    if (ssl_context == NULL) return tls_openssl_fail("TLS context allocation failed");
    if (tls_configure_context(ssl_context, cert, key, ca, ca_provided != 0, min_version, max_version, ciphers) != 0) {
        SSL_CTX_free(ssl_context);
        return -1;
    }
    scriptgo_tls_context *context = (scriptgo_tls_context *)calloc(1, sizeof(*context));
    if (context == NULL) {
        SSL_CTX_free(ssl_context);
        return tls_fail("TLS context allocation failed");
    }
    context->ctx = ssl_context;
    *out_handle = (double)(uintptr_t)context;
    return 0;
}

int scriptgo_tls_socket_create(double context_handle, double is_server, double *out_handle) {
    SSL_CTX *ctx = tls_context_value(context_handle);
    if (ctx == NULL || out_handle == NULL) return tls_fail("TLS socket creation arguments are invalid");
    SSL *ssl = SSL_new(ctx);
    BIO *read_bio = BIO_new(BIO_s_mem());
    BIO *write_bio = BIO_new(BIO_s_mem());
    if (ssl == NULL || read_bio == NULL || write_bio == NULL) {
        SSL_free(ssl);
        BIO_free(read_bio);
        BIO_free(write_bio);
        return tls_openssl_fail("TLS socket allocation failed");
    }
    SSL_set_bio(ssl, read_bio, write_bio);
    if (is_server != 0) SSL_set_accept_state(ssl);
    else SSL_set_connect_state(ssl);
    if (tls_socket_new(ssl, -1, 0, 1, out_handle) != 0) {
        SSL_free(ssl);
        return -1;
    }
    return 0;
}

int scriptgo_tls_socket_connect(double context_handle, const char *host, double port,
                               const char *servername, double reject, void *session_data,
                               double *out_handle) {
    if (host == NULL || out_handle == NULL || !tls_integer_in_range(port, 0, 65535)) return tls_fail("TLS connect arguments are invalid");
    SSL_CTX *ctx = tls_context_value(context_handle);
    if (ctx == NULL) return tls_fail("TLS context is invalid");
    struct addrinfo hints;
    struct addrinfo *results = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_family = AF_UNSPEC;
    char port_text[32];
    snprintf(port_text, sizeof(port_text), "%d", (int)port);
    int lookup = getaddrinfo(host, port_text, &hints, &results);
    if (lookup != 0 || results == NULL) return tls_fail(gai_strerror(lookup));
    int fd = -1;
    for (struct addrinfo *entry = results; entry != NULL; entry = entry->ai_next) {
        fd = (int)socket(entry->ai_family, entry->ai_socktype, entry->ai_protocol);
        if (fd < 0) continue;
        if (connect(fd, entry->ai_addr, entry->ai_addrlen) == 0) break;
        tls_close_fd(fd);
        fd = -1;
    }
    freeaddrinfo(results);
    if (fd < 0) return tls_fail(strerror(errno));
    SSL *ssl = SSL_new(ctx);
    if (ssl == NULL || SSL_set_fd(ssl, fd) != 1) {
        if (ssl != NULL) SSL_free(ssl);
        tls_close_fd(fd);
        return tls_openssl_fail("TLS socket allocation failed");
    }
    const char *verify_name = servername != NULL && *servername != '\0' ? servername : host;
    if ((servername != NULL && *servername != '\0' && SSL_set_tlsext_host_name(ssl, servername) != 1) ||
        tls_socket_configure_verify(ssl, verify_name, reject != 0) != 0) {
        SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS client configuration failed");
    }
    if (tls_set_session_from_view(ssl, session_data, 0) != 0) {
        SSL_free(ssl); tls_close_fd(fd); return -1;
    }
    if (SSL_connect(ssl) != 1) {
        SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS handshake failed");
    }
    int result = tls_socket_new(ssl, fd, 0, 1, out_handle);
    if (result != 0) {
        SSL_free(ssl);
        tls_close_fd(fd);
    }
    return result;
}

int scriptgo_tls_socket_adopt(double context_handle, double fd_num, const char *servername,
                             double is_server, double request_cert, double reject, void *session_data,
                             double *out_handle) {
    SSL_CTX *ctx = tls_context_value(context_handle);
    if (ctx == NULL || out_handle == NULL || !tls_integer_in_range(fd_num, 0, INT_MAX)) return tls_fail("TLS socket adoption arguments are invalid");
    int fd = (int)fd_num;
    SSL *ssl = SSL_new(ctx);
    if (ssl == NULL || SSL_set_fd(ssl, fd) != 1) {
        if (ssl != NULL) SSL_free(ssl);
        return tls_openssl_fail("TLS socket adoption failed");
    }
    if (tls_set_session_from_view(ssl, session_data, 0) != 0) {
        SSL_free(ssl);
        tls_close_fd(fd);
        return -1;
    }
    if (is_server) {
        SSL_set_accept_state(ssl);
        int verify_mode = request_cert != 0
            ? SSL_VERIFY_PEER | (reject != 0 ? SSL_VERIFY_FAIL_IF_NO_PEER_CERT : 0)
            : SSL_VERIFY_NONE;
        SSL_set_verify(ssl, verify_mode, NULL);
        if (SSL_accept(ssl) != 1) { SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS server handshake failed"); }
    } else {
        SSL_set_connect_state(ssl);
        if (servername != NULL && *servername != '\0') SSL_set_tlsext_host_name(ssl, servername);
        if (tls_socket_configure_verify(ssl, servername, reject != 0) != 0 || SSL_connect(ssl) != 1) {
            SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS client handshake failed");
        }
    }
    int result = tls_socket_new(ssl, fd, 0, 1, out_handle);
    if (result != 0) {
        SSL_free(ssl);
        tls_close_fd(fd);
    }
    return result;
}

static int tls_pair_advance(scriptgo_tls_socket *socket) {
    if (socket == NULL || socket->ssl == NULL) return tls_fail("TLS pair is invalid");
    int result = SSL_do_handshake(socket->ssl);
    int error = SSL_get_error(socket->ssl, result);
    if (result == 1 || error == SSL_ERROR_WANT_READ || error == SSL_ERROR_WANT_WRITE) return 0;
    return tls_openssl_fail("TLS pair handshake failed");
}

int scriptgo_tls_pair_create(double context_handle, double is_server, double request_cert,
                             double reject, double *out_handle) {
    SSL_CTX *ctx = tls_context_value(context_handle);
    if (ctx == NULL || out_handle == NULL) return tls_fail("TLS pair context is invalid");
    SSL *ssl = SSL_new(ctx);
    BIO *read_bio = BIO_new(BIO_s_mem());
    BIO *write_bio = BIO_new(BIO_s_mem());
    if (ssl == NULL || read_bio == NULL || write_bio == NULL) {
        SSL_free(ssl);
        BIO_free(read_bio);
        BIO_free(write_bio);
        return tls_openssl_fail("TLS pair allocation failed");
    }
    SSL_set_bio(ssl, read_bio, write_bio);
    if (is_server) {
        SSL_set_accept_state(ssl);
        SSL_set_verify(ssl, reject ? SSL_VERIFY_PEER | (request_cert ? SSL_VERIFY_FAIL_IF_NO_PEER_CERT : 0) : SSL_VERIFY_NONE, NULL);
    } else {
        SSL_set_connect_state(ssl);
        SSL_set_verify(ssl, reject ? SSL_VERIFY_PEER : SSL_VERIFY_NONE, NULL);
    }
    if (tls_socket_new(ssl, -1, 1, 2, out_handle) != 0) { SSL_free(ssl); return -1; }
    return tls_pair_advance(tls_socket_from_handle(*out_handle));
}

int scriptgo_tls_socket_write(double handle, const char *data, double length, double *out_written) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || socket->is_pair || out_written == NULL || !tls_integer_in_range(length, 0, INT_MAX) || (data == NULL && length != 0)) return tls_fail("TLS socket write arguments are invalid");
    if (length == 0) {
        *out_written = 0;
        return 0;
    }
    size_t total = 0;
    while (total < (size_t)length) {
        int result = SSL_write(socket->ssl, data + total, (int)((size_t)length - total));
        if (result <= 0) return tls_openssl_fail("TLS socket write failed");
        total += (size_t)result;
    }
    *out_written = (double)total;
    return 0;
}

int scriptgo_tls_socket_write_bytes(double handle, void *view_data, double *out_written) {
    scriptgo_tls_typed_array_view *view = (scriptgo_tls_typed_array_view *)view_data;
    if (!tls_is_byte_view(view) || view->length < 0 || view->element_size <= 0 || (view->length != 0 && view->data == NULL)) return tls_fail("TLS socket byte write arguments are invalid");
    if ((uint64_t)view->length > SIZE_MAX / (uint64_t)view->element_size) return tls_fail("TLS socket byte write is too large");
    size_t length = (size_t)view->length * (size_t)view->element_size;
    if (length > INT_MAX) return tls_fail("TLS socket byte write is too large");
    return scriptgo_tls_socket_write(handle, (const char *)view->data, (double)length, out_written);
}

int scriptgo_tls_socket_read(double handle, double max_length, char **out_data) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || socket->is_pair || out_data == NULL || !tls_integer_in_range(max_length, 0, 16 * 1024 * 1024)) return tls_fail("TLS socket read arguments are invalid");
    size_t capacity = (size_t)max_length;
    if (capacity == 0) capacity = 65536;
    if (capacity > 16 * 1024 * 1024) return tls_fail("TLS socket read size is too large");
    char *data = (char *)malloc(capacity + 1);
    if (data == NULL) return tls_fail("TLS socket read allocation failed");
    int result = SSL_read(socket->ssl, data, (int)capacity);
    if (result <= 0) {
        int error = SSL_get_error(socket->ssl, result);
        if (error == SSL_ERROR_ZERO_RETURN) { data[0] = '\0'; *out_data = data; return 0; }
        free(data);
        return tls_openssl_fail("TLS socket read failed");
    }
    data[result] = '\0';
    *out_data = data;
    return 0;
}

static int tls_pair_write_data(scriptgo_tls_socket *socket, double mode, const char *data, size_t length,
                               int encoded, double *out_written) {
    if (socket == NULL || socket->ssl == NULL || !socket->is_pair || out_written == NULL || length > INT_MAX || (data == NULL && length != 0)) return tls_fail("TLS pair write arguments are invalid");
    if (mode != 0 && mode != 1) return tls_fail("TLS pair mode is invalid");
    if (length == 0) {
        *out_written = 0;
        return 0;
    }
    if (mode == 1) {
        /* The string ABI is NUL-terminated, so carry encrypted BIO bytes as hex. */
        const unsigned char *input = (const unsigned char *)data;
        unsigned char *decoded = NULL;
        size_t input_length = length;
        if (encoded) {
            if ((length & 1) != 0) return tls_fail("TLS pair encrypted input is not valid hexadecimal");
            input_length = length / 2;
            decoded = (unsigned char *)malloc(input_length == 0 ? 1 : input_length);
            if (decoded == NULL || tls_hex_decode(data, decoded, input_length) != 0) {
                free(decoded);
                return tls_fail("TLS pair encrypted input is not valid hexadecimal");
            }
            input = decoded;
        }
        if (input_length > INT_MAX) {
            free(decoded);
            return tls_fail("TLS pair encrypted input is too large");
        }
        size_t total = 0;
        while (total < input_length) {
            int result = BIO_write(SSL_get_rbio(socket->ssl), input + total, (int)(input_length - total));
            if (result <= 0) {
                free(decoded);
                return tls_openssl_fail("TLS pair encrypted write failed");
            }
            total += (size_t)result;
        }
        free(decoded);
        if (tls_pair_advance(socket) != 0) return -1;
        *out_written = (double)input_length;
        return 0;
    }
    if (tls_pair_advance(socket) != 0) return -1;
    size_t total = 0;
    while (total < length) {
        int result = SSL_write(socket->ssl, data + total, (int)(length - total));
        if (result <= 0) return tls_openssl_fail("TLS pair cleartext write failed");
        total += (size_t)result;
    }
    *out_written = (double)total;
    return 0;
}

int scriptgo_tls_socket_pair_write(double handle, double mode, const char *data, double length, double *out_written) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (!tls_integer_in_range(length, 0, INT_MAX)) return tls_fail("TLS pair write length is invalid");
    return tls_pair_write_data(socket, mode, data, (size_t)length, 1, out_written);
}

int scriptgo_tls_socket_pair_write_bytes(double handle, double mode, void *view_data, double *out_written) {
    scriptgo_tls_typed_array_view *view = (scriptgo_tls_typed_array_view *)view_data;
    if (!tls_is_byte_view(view) || view->length < 0 || view->element_size <= 0 || (view->length != 0 && view->data == NULL)) return tls_fail("TLS pair byte write arguments are invalid");
    if ((uint64_t)view->length > SIZE_MAX / (uint64_t)view->element_size) return tls_fail("TLS pair byte write is too large");
    return tls_pair_write_data(tls_socket_from_handle(handle), mode, (const char *)view->data,
                               (size_t)view->length * (size_t)view->element_size, 0, out_written);
}

int scriptgo_tls_socket_pair_read(double handle, double mode, double max_length, char **out_data) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || !socket->is_pair || out_data == NULL || !tls_integer_in_range(max_length, 0, INT_MAX)) return tls_fail("TLS pair read arguments are invalid");
    if (mode != 0 && mode != 1) return tls_fail("TLS pair mode is invalid");
    size_t capacity = max_length > 0 ? (size_t)max_length : 65536;
    char *data = (char *)malloc(capacity + 1);
    if (data == NULL) return tls_fail("TLS pair read allocation failed");
    int result;
    if (mode == 1) {
        result = BIO_read(SSL_get_wbio(socket->ssl), data, (int)capacity);
        if (result <= 0) {
            free(data);
            *out_data = tls_strdup("");
            return *out_data == NULL ? tls_fail("TLS pair encrypted read allocation failed") : 0;
        }
        char *encoded = tls_hex((const unsigned char *)data, (size_t)result);
        free(data);
        if (encoded == NULL) return tls_fail("TLS pair encrypted read allocation failed");
        *out_data = encoded;
        return 0;
    } else {
        if (tls_pair_advance(socket) != 0) { free(data); return -1; }
        result = SSL_read(socket->ssl, data, (int)capacity);
        if (result <= 0) {
            int error = SSL_get_error(socket->ssl, result);
            if (error == SSL_ERROR_WANT_READ || error == SSL_ERROR_WANT_WRITE) result = 0;
            else { free(data); return tls_openssl_fail("TLS pair cleartext read failed"); }
        }
    }
    data[result] = '\0';
    *out_data = data;
    return 0;
}

int scriptgo_tls_socket_close(double handle) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL) return 0;
    if (socket->is_pair && socket->close_count < socket->refs) {
        socket->close_count++;
        if (socket->close_count < socket->refs) return 0;
    } else if (!socket->is_pair && socket->close_count != 0) {
        return 0;
    } else if (!socket->is_pair) {
        socket->close_count = 1;
    }
    if (socket->ssl != NULL) {
        if (socket->fd >= 0) SSL_shutdown(socket->ssl);
        SSL_free(socket->ssl);
    }
    if (socket->fd >= 0) tls_close_fd(socket->fd);
    socket->ssl = NULL;
    free(socket);
    return 0;
}

static char *tls_asn1_time(const ASN1_TIME *time) {
    if (time == NULL) return NULL;
    BIO *bio = BIO_new(BIO_s_mem());
    if (bio == NULL || ASN1_TIME_print(bio, time) != 1) { BIO_free(bio); return NULL; }
    char *data = NULL;
    long length = BIO_get_mem_data(bio, &data);
    if (length < 0 || (data == NULL && length != 0) || (size_t)length > SIZE_MAX - 1) { BIO_free(bio); return NULL; }
    char *result = (char *)malloc((size_t)length + 1);
    if (result != NULL) { memcpy(result, data, (size_t)length); result[length] = '\0'; }
    BIO_free(bio);
    return result;
}

static char *tls_certificate_pem(X509 *certificate) {
    if (certificate == NULL) return NULL;
    BIO *bio = BIO_new(BIO_s_mem());
    if (bio == NULL || PEM_write_bio_X509(bio, certificate) != 1) { BIO_free(bio); return NULL; }
    char *data = NULL;
    long length = BIO_get_mem_data(bio, &data);
    if (length < 0 || (data == NULL && length != 0) || (size_t)length > SIZE_MAX - 1) { BIO_free(bio); return NULL; }
    char *result = (char *)malloc((size_t)length + 1);
    if (result != NULL) { memcpy(result, data, (size_t)length); result[length] = '\0'; }
    BIO_free(bio);
    return result;
}

static char *tls_certificate_der(X509 *certificate) {
    if (certificate == NULL) return NULL;
    int length = i2d_X509(certificate, NULL);
    if (length <= 0) return NULL;
    unsigned char *data = (unsigned char *)malloc((size_t)length);
    if (data == NULL) return NULL;
    unsigned char *cursor = data;
    if (i2d_X509(certificate, &cursor) != length) {
        free(data);
        return NULL;
    }
    char *result = tls_hex(data, (size_t)length);
    free(data);
    return result;
}

static int tls_append_bytes(char **buffer, size_t *length, size_t *capacity,
                            const unsigned char *value, size_t value_length) {
    if (buffer == NULL || length == NULL || capacity == NULL || (value == NULL && value_length != 0) || value_length > SIZE_MAX - *length - 1) return -1;
    size_t required = *length + value_length + 1;
    if (required > *capacity) {
        size_t next = *capacity == 0 ? 256 : *capacity;
        while (next < required) {
            if (next > SIZE_MAX / 2) {
                next = required;
                break;
            }
            next *= 2;
        }
        char *grown = (char *)realloc(*buffer, next);
        if (grown == NULL) return -1;
        *buffer = grown;
        *capacity = next;
    }
    if (value_length != 0) memcpy(*buffer + *length, value, value_length);
    *length += value_length;
    (*buffer)[*length] = '\0';
    return 0;
}

typedef struct {
    char *key;
    char *value;
} tls_name_entry;

static void tls_name_entries_free(tls_name_entry *entries, int count) {
    if (entries == NULL) return;
    for (int i = 0; i < count; i++) {
        free(entries[i].key);
        free(entries[i].value);
    }
    free(entries);
}

static int tls_append_name_json(char **buffer, size_t *length, size_t *capacity, X509_NAME *name) {
    if (tls_append(buffer, length, capacity, "{") != 0) return -1;
    int count = name == NULL ? 0 : X509_NAME_entry_count(name);
    tls_name_entry *entries = count == 0 ? NULL : (tls_name_entry *)calloc((size_t)count, sizeof(*entries));
    if (count != 0 && entries == NULL) return -1;
    for (int i = 0; i < count; i++) {
        X509_NAME_ENTRY *entry = X509_NAME_get_entry(name, i);
        ASN1_OBJECT *object = entry == NULL ? NULL : X509_NAME_ENTRY_get_object(entry);
        ASN1_STRING *string = entry == NULL ? NULL : X509_NAME_ENTRY_get_data(entry);
        if (object == NULL || string == NULL) {
            tls_name_entries_free(entries, count);
            return -1;
        }
        int nid = OBJ_obj2nid(object);
        const char *short_name = nid == NID_undef ? NULL : OBJ_nid2sn(nid);
        int key_length = short_name == NULL ? OBJ_obj2txt(NULL, 0, object, 1) : (int)strlen(short_name);
        if (key_length <= 0) {
            tls_name_entries_free(entries, count);
            return -1;
        }
        entries[i].key = (char *)malloc((size_t)key_length + 1);
        if (entries[i].key == NULL) {
            tls_name_entries_free(entries, count);
            return -1;
        }
        if (short_name != NULL) {
            memcpy(entries[i].key, short_name, (size_t)key_length);
            entries[i].key[key_length] = '\0';
        } else if (OBJ_obj2txt(entries[i].key, key_length + 1, object, 1) != key_length) {
            tls_name_entries_free(entries, count);
            return -1;
        }
        unsigned char *utf8 = NULL;
        int value_length = ASN1_STRING_to_UTF8(&utf8, string);
        if (value_length < 0) {
            OPENSSL_free(utf8);
            tls_name_entries_free(entries, count);
            return -1;
        }
        entries[i].value = (char *)malloc((size_t)value_length + 1);
        if (entries[i].value == NULL) {
            OPENSSL_free(utf8);
            tls_name_entries_free(entries, count);
            return -1;
        }
        if (value_length != 0) memcpy(entries[i].value, utf8, (size_t)value_length);
        entries[i].value[value_length] = '\0';
        OPENSSL_free(utf8);
    }
    int first = 1;
    for (int i = 0; i < count; i++) {
        int seen = 0;
        for (int previous = 0; previous < i; previous++) {
            if (strcmp(entries[previous].key, entries[i].key) == 0) {
                seen = 1;
                break;
            }
        }
        if (seen) continue;
        int occurrences = 0;
        for (int j = i; j < count; j++) if (strcmp(entries[i].key, entries[j].key) == 0) occurrences++;
        if (!first && tls_append(buffer, length, capacity, ",") != 0) goto name_fail;
        first = 0;
        if (tls_append_json_string(buffer, length, capacity, entries[i].key) != 0 || tls_append(buffer, length, capacity, ":") != 0) goto name_fail;
        if (occurrences == 1) {
            if (tls_append_json_string(buffer, length, capacity, entries[i].value) != 0) goto name_fail;
        } else {
            if (tls_append(buffer, length, capacity, "[") != 0) goto name_fail;
            int value_first = 1;
            for (int j = i; j < count; j++) {
                if (strcmp(entries[i].key, entries[j].key) != 0) continue;
                if (!value_first && tls_append(buffer, length, capacity, ",") != 0) goto name_fail;
                value_first = 0;
                if (tls_append_json_string(buffer, length, capacity, entries[j].value) != 0) goto name_fail;
            }
            if (tls_append(buffer, length, capacity, "]") != 0) goto name_fail;
        }
    }
    tls_name_entries_free(entries, count);
    return tls_append(buffer, length, capacity, "}");

name_fail:
    tls_name_entries_free(entries, count);
    return -1;
}

static int tls_append_alt_name(char **buffer, size_t *length, size_t *capacity, int *first,
                               const char *prefix, const unsigned char *value, size_t value_length) {
    if (first == NULL || prefix == NULL) return -1;
    if (!*first && tls_append(buffer, length, capacity, ", ") != 0) return -1;
    if (tls_append(buffer, length, capacity, prefix) != 0 || tls_append_bytes(buffer, length, capacity, value, value_length) != 0) return -1;
    *first = 0;
    return 0;
}

static char *tls_certificate_json(X509 *certificate, int include_pem) {
    if (certificate == NULL) return tls_json_empty_object();
    char subject[512] = "";
    char issuer[512] = "";
    if (X509_NAME_oneline(X509_get_subject_name(certificate), subject, sizeof(subject)) == NULL ||
        X509_NAME_oneline(X509_get_issuer_name(certificate), issuer, sizeof(issuer)) == NULL) return NULL;
    char *valid_from = tls_asn1_time(X509_get0_notBefore(certificate));
    char *valid_to = tls_asn1_time(X509_get0_notAfter(certificate));
    char *serial = NULL;
    BIGNUM *serial_bn = ASN1_INTEGER_to_BN(X509_get0_serialNumber(certificate), NULL);
    if (serial_bn != NULL) { serial = BN_bn2hex(serial_bn); BN_free(serial_bn); }
    if (valid_from == NULL || valid_to == NULL || serial == NULL) {
        free(valid_from); free(valid_to); OPENSSL_free(serial);
        return NULL;
    }
    char *alt_names = NULL; size_t alt_name_length = 0; size_t alt_name_capacity = 0; int first_alt_name = 1;
    unsigned char digest[EVP_MAX_MD_SIZE]; unsigned int digest_len = 0;
    char *fingerprint = NULL; char *fingerprint256 = NULL; char *fingerprint512 = NULL;
    char *raw = NULL;
    if (X509_digest(certificate, EVP_sha1(), digest, &digest_len) != 1) goto certificate_fail;
    fingerprint = tls_hex(digest, digest_len);
    if (X509_digest(certificate, EVP_sha256(), digest, &digest_len) != 1) goto certificate_fail;
    fingerprint256 = tls_hex(digest, digest_len);
    if (X509_digest(certificate, EVP_sha512(), digest, &digest_len) != 1) goto certificate_fail;
    fingerprint512 = tls_hex(digest, digest_len);
    if (fingerprint == NULL || fingerprint256 == NULL || fingerprint512 == NULL) goto certificate_fail;

    GENERAL_NAMES *names = (GENERAL_NAMES *)X509_get_ext_d2i(certificate, NID_subject_alt_name, NULL, NULL);
    if (names != NULL) {
        for (int i = 0; i < sk_GENERAL_NAME_num(names); i++) {
            GENERAL_NAME *name = sk_GENERAL_NAME_value(names, i);
            if (name->type == GEN_DNS) {
                ASN1_STRING *value = name->d.dNSName;
                if (tls_append_alt_name(&alt_names, &alt_name_length, &alt_name_capacity, &first_alt_name, "DNS:",
                                        ASN1_STRING_get0_data(value), (size_t)ASN1_STRING_length(value)) != 0) goto certificate_fail;
            } else if (name->type == GEN_IPADD) {
                char ip[INET6_ADDRSTRLEN];
                const unsigned char *raw = ASN1_STRING_get0_data(name->d.iPAddress);
                int address_family = ASN1_STRING_length(name->d.iPAddress) == 4 ? AF_INET : AF_INET6;
                if (inet_ntop(address_family, raw, ip, sizeof(ip)) != NULL) {
                    if (tls_append_alt_name(&alt_names, &alt_name_length, &alt_name_capacity, &first_alt_name, "IP Address:",
                                            (const unsigned char *)ip, strlen(ip)) != 0) goto certificate_fail;
                }
            } else if (name->type == GEN_EMAIL) {
                ASN1_STRING *value = name->d.rfc822Name;
                if (tls_append_alt_name(&alt_names, &alt_name_length, &alt_name_capacity, &first_alt_name, "email:",
                                        ASN1_STRING_get0_data(value), (size_t)ASN1_STRING_length(value)) != 0) goto certificate_fail;
            } else if (name->type == GEN_URI) {
                ASN1_STRING *value = name->d.uniformResourceIdentifier;
                if (tls_append_alt_name(&alt_names, &alt_name_length, &alt_name_capacity, &first_alt_name, "URI:",
                                        ASN1_STRING_get0_data(value), (size_t)ASN1_STRING_length(value)) != 0) goto certificate_fail;
            }
        }
        GENERAL_NAMES_free(names);
    }

    char *pem = include_pem ? tls_certificate_pem(certificate) : NULL;
    raw = tls_certificate_der(certificate);
    if (raw == NULL) goto certificate_fail;
    if (include_pem && pem == NULL) goto certificate_fail;
    char *result = NULL; size_t length = 0; size_t capacity = 0;
    if (tls_append(&result, &length, &capacity, "{") != 0) goto certificate_fail_with_pem;
    const char *keys[] = {"subject", "issuer", "subjectAltName", "validFrom", "validTo", "fingerprint", "fingerprint256", "fingerprint512", "serialNumber"};
    const char *values[] = {subject, issuer, alt_names, valid_from, valid_to, fingerprint, fingerprint256, fingerprint512, serial};
    for (size_t i = 0; i < sizeof(keys) / sizeof(keys[0]); i++) {
        if (i != 0 && tls_append(&result, &length, &capacity, ",") != 0) goto certificate_fail_with_pem;
        if (tls_append_json_string(&result, &length, &capacity, keys[i]) != 0 ||
            tls_append(&result, &length, &capacity, ":") != 0 ||
            tls_append_json_string(&result, &length, &capacity, values[i] == NULL ? "" : values[i]) != 0) goto certificate_fail_with_pem;
    }
    if (include_pem) {
        if (tls_append(&result, &length, &capacity, ",\"pem\":") != 0 || tls_append_json_string(&result, &length, &capacity, pem) != 0) goto certificate_fail_with_pem;
    }
    if (tls_append(&result, &length, &capacity, ",\"ca\":") != 0 ||
        tls_append(&result, &length, &capacity, X509_check_ca(certificate) > 0 ? "true" : "false") != 0 ||
        tls_append(&result, &length, &capacity, ",\"raw\":") != 0 ||
        tls_append_json_string(&result, &length, &capacity, raw) != 0 ||
        tls_append(&result, &length, &capacity, ",\"subjectObject\":") != 0 ||
        tls_append_name_json(&result, &length, &capacity, X509_get_subject_name(certificate)) != 0 ||
        tls_append(&result, &length, &capacity, ",\"issuerObject\":") != 0 ||
        tls_append_name_json(&result, &length, &capacity, X509_get_issuer_name(certificate)) != 0) goto certificate_fail_with_pem;
    if (tls_append(&result, &length, &capacity, "}") != 0) goto certificate_fail_with_pem;
    free(valid_from); free(valid_to); OPENSSL_free(serial); free(fingerprint); free(fingerprint256); free(fingerprint512); free(pem); free(raw); free(alt_names);
    return result;

certificate_fail_with_pem:
    free(result);
    free(pem);
certificate_fail:
    free(valid_from); free(valid_to); OPENSSL_free(serial); free(fingerprint); free(fingerprint256); free(fingerprint512); free(raw); free(alt_names);
    return NULL;
}

static int tls_session_hex(SSL *ssl, int ticket, char **out_value) {
    if (out_value == NULL) return tls_fail("TLS session output is null");
    if (ssl == NULL || !SSL_is_init_finished(ssl) || (ticket && SSL_is_server(ssl))) {
        *out_value = tls_strdup("");
        return *out_value == NULL ? tls_fail("TLS session allocation failed") : 0;
    }
    SSL_SESSION *session = SSL_get1_session(ssl);
    if (session == NULL) { *out_value = tls_strdup(""); return 0; }
    unsigned char *data = NULL; int length = 0;
    if (ticket) {
        const unsigned char *ticket_data = NULL; size_t ticket_length = 0;
        SSL_SESSION_get0_ticket(session, &ticket_data, &ticket_length);
        *out_value = tls_hex(ticket_data, ticket_length);
        SSL_SESSION_free(session);
        return *out_value == NULL ? tls_fail("TLS ticket allocation failed") : 0;
    }
    length = i2d_SSL_SESSION(session, &data);
    *out_value = length > 0 ? tls_hex(data, (size_t)length) : tls_strdup("");
    OPENSSL_free(data); SSL_SESSION_free(session);
    return *out_value == NULL ? tls_fail("TLS session allocation failed") : 0;
}

int scriptgo_tls_socket_info(double handle, const char *property, char **out_value) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (out_value == NULL || property == NULL) return tls_fail("TLS socket info arguments are invalid");
    int known = strcmp(property, "protocol") == 0 || strcmp(property, "cipher") == 0 ||
                strcmp(property, "certificate") == 0 || strcmp(property, "peerCertificate") == 0 ||
                strcmp(property, "peerCertificateDetailed") == 0 || strcmp(property, "finished") == 0 ||
                strcmp(property, "peerFinished") == 0 || strcmp(property, "session") == 0 ||
                strcmp(property, "ticket") == 0 || strcmp(property, "sharedSigalgs") == 0 ||
                strcmp(property, "localAddress") == 0 || strcmp(property, "remoteAddress") == 0 ||
                strcmp(property, "localFamily") == 0 || strcmp(property, "remoteFamily") == 0 ||
                strcmp(property, "authorizationError") == 0 || strcmp(property, "ephemeral") == 0;
    if (!known) return tls_fail("TLS socket info property is unknown");
    if (socket == NULL || socket->ssl == NULL) {
        if (strcmp(property, "certificate") == 0 || strcmp(property, "peerCertificate") == 0 || strcmp(property, "peerCertificateDetailed") == 0 || strcmp(property, "cipher") == 0 || strcmp(property, "ephemeral") == 0) {
            *out_value = tls_strdup("{}");
        } else if (strcmp(property, "sharedSigalgs") == 0) {
            *out_value = tls_strdup("[]");
        } else {
            *out_value = tls_strdup("");
        }
        return *out_value == NULL ? tls_fail("TLS socket info allocation failed") : 0;
    }
    if (strcmp(property, "protocol") == 0) {
        *out_value = tls_strdup(SSL_get_version(socket->ssl));
        return *out_value == NULL ? tls_fail("TLS protocol allocation failed") : 0;
    }
    if (strcmp(property, "cipher") == 0) {
        const SSL_CIPHER *cipher = SSL_get_current_cipher(socket->ssl);
        char *result = NULL; size_t length = 0; size_t capacity = 0;
        const char *standard_name = cipher == NULL ? "" : SSL_CIPHER_standard_name(cipher);
        if (tls_append(&result, &length, &capacity, "{\"name\":") != 0 ||
            tls_append_json_string(&result, &length, &capacity, cipher == NULL ? "" : SSL_CIPHER_get_name(cipher)) != 0 ||
            tls_append(&result, &length, &capacity, ",\"standardName\":") != 0 ||
            tls_append_json_string(&result, &length, &capacity, standard_name == NULL ? "" : standard_name) != 0 ||
            tls_append(&result, &length, &capacity, ",\"version\":") != 0 ||
            tls_append_json_string(&result, &length, &capacity, cipher == NULL ? "" : SSL_CIPHER_get_version(cipher)) != 0 ||
            tls_append(&result, &length, &capacity, "}") != 0) {
            free(result);
            return tls_fail("TLS cipher allocation failed");
        }
        *out_value = result; return 0;
    }
    if (strcmp(property, "certificate") == 0) {
        X509 *certificate = SSL_get_certificate(socket->ssl);
        *out_value = tls_certificate_json(certificate, 1);
        return *out_value == NULL ? tls_fail("TLS certificate allocation failed") : 0;
    }
    if (strcmp(property, "peerCertificate") == 0 || strcmp(property, "peerCertificateDetailed") == 0) {
        X509 *certificate = SSL_get_peer_certificate(socket->ssl);
        int include_pem = strcmp(property, "peerCertificateDetailed") == 0;
        *out_value = tls_certificate_json(certificate, include_pem);
        X509_free(certificate);
        return *out_value == NULL ? tls_fail("TLS peer certificate allocation failed") : 0;
    }
    if (strcmp(property, "finished") == 0 || strcmp(property, "peerFinished") == 0) {
        unsigned char data[256]; size_t length = strcmp(property, "finished") == 0 ? SSL_get_finished(socket->ssl, data, sizeof(data)) : SSL_get_peer_finished(socket->ssl, data, sizeof(data));
        *out_value = tls_hex(data, length); return *out_value == NULL ? tls_fail("TLS finished message allocation failed") : 0;
    }
    if (strcmp(property, "session") == 0) return tls_session_hex(socket->ssl, 0, out_value);
    if (strcmp(property, "ticket") == 0) return tls_session_hex(socket->ssl, 1, out_value);
    if (strcmp(property, "sharedSigalgs") == 0) {
        char *result = NULL; size_t length = 0; size_t capacity = 0; tls_append(&result, &length, &capacity, "[");
        int first = 1;
        for (int i = 0;; i++) {
            int sign = 0, hash = 0, sig = 0; unsigned char raw_sig = 0, raw_hash = 0;
            if (SSL_get_shared_sigalgs(socket->ssl, i, &sign, &hash, &sig, &raw_sig, &raw_hash) <= 0) break;
            (void)sig; (void)raw_sig; (void)raw_hash;
            char value[128]; snprintf(value, sizeof(value), "%s+%s", tls_signature_name(sign), tls_hash_name(hash));
            if (!first && tls_append(&result, &length, &capacity, ",") != 0) { free(result); return tls_fail("TLS signature algorithm allocation failed"); }
            first = 0;
            if (tls_append_json_string(&result, &length, &capacity, value) != 0) { free(result); return tls_fail("TLS signature algorithm allocation failed"); }
        }
        if (tls_append(&result, &length, &capacity, "]") != 0) { free(result); return tls_fail("TLS signature algorithm allocation failed"); }
        *out_value = result; return 0;
    }
    if (strcmp(property, "localAddress") == 0 || strcmp(property, "remoteAddress") == 0 || strcmp(property, "localFamily") == 0 || strcmp(property, "remoteFamily") == 0) {
        char address[INET6_ADDRSTRLEN] = "";
        const char *family = "IPv4";
        double port = 0;
        int local = strcmp(property, "localAddress") == 0 || strcmp(property, "localFamily") == 0;
        if (tls_socket_address(socket->fd, local, address, sizeof(address), &port, &family) != 0) {
            *out_value = tls_strdup("");
            return *out_value == NULL ? tls_fail("TLS socket address allocation failed") : 0;
        }
        *out_value = tls_strdup(strcmp(property, "localFamily") == 0 || strcmp(property, "remoteFamily") == 0 ? family : address);
        return *out_value == NULL ? tls_fail("TLS socket address allocation failed") : 0;
    }
    if (strcmp(property, "authorizationError") == 0) {
        if (!SSL_is_init_finished(socket->ssl)) {
            *out_value = tls_strdup("");
            return *out_value == NULL ? tls_fail("TLS authorization error allocation failed") : 0;
        }
        long verify_result = SSL_get_verify_result(socket->ssl);
        *out_value = tls_strdup(verify_result == X509_V_OK ? "" : X509_verify_cert_error_string(verify_result));
        return *out_value == NULL ? tls_fail("TLS authorization error allocation failed") : 0;
    }
    if (strcmp(property, "ephemeral") == 0) {
        if (SSL_is_server(socket->ssl)) {
            *out_value = tls_strdup("{\"server\":true}");
            return *out_value == NULL ? tls_fail("TLS ephemeral key allocation failed") : 0;
        }
        if (!SSL_is_init_finished(socket->ssl)) {
            *out_value = tls_strdup("{}");
            return *out_value == NULL ? tls_fail("TLS ephemeral key allocation failed") : 0;
        }
        EVP_PKEY *key = NULL;
        char name[128] = "";
        const char *type = "";
        int size = 0;
        if (SSL_get_server_tmp_key(socket->ssl, &key) == 1 && key != NULL) {
            int key_type = EVP_PKEY_get_base_id(key);
            if (key_type == EVP_PKEY_EC) {
                type = "ECDH";
#if OPENSSL_VERSION_NUMBER >= 0x30000000L
                size_t name_length = 0;
                EVP_PKEY_get_group_name(key, name, sizeof(name), &name_length);
#endif
            } else if (key_type == EVP_PKEY_DH) {
                type = "DH";
            } else if (key_type == EVP_PKEY_RSA || key_type == EVP_PKEY_RSA_PSS) {
                type = "RSA";
            }
            size = EVP_PKEY_get_bits(key);
            EVP_PKEY_free(key);
        }
        if (type[0] == '\0') {
            *out_value = tls_strdup("{}");
            return *out_value == NULL ? tls_fail("TLS ephemeral key allocation failed") : 0;
        }
        char value[256];
        snprintf(value, sizeof(value), "{\"type\":\"%s\",\"name\":\"%s\",\"size\":%d}", type, name, size);
        *out_value = tls_strdup(value);
        return *out_value == NULL ? tls_fail("TLS ephemeral key allocation failed") : 0;
    }
    return tls_fail("TLS socket info property is unavailable");
}

int scriptgo_tls_socket_number(double handle, const char *property, double *out_value) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (out_value == NULL || property == NULL) return tls_fail("TLS socket number arguments are invalid");
    if (strcmp(property, "localPort") != 0 && strcmp(property, "remotePort") != 0 && strcmp(property, "maxSendFragment") != 0) {
        return tls_fail("TLS socket number property is unknown");
    }
    *out_value = 0;
    if (socket == NULL || socket->ssl == NULL) return 0;
    char address[INET6_ADDRSTRLEN]; const char *family = "IPv4"; double port = 0;
    if (strcmp(property, "localPort") == 0) { tls_socket_address(socket->fd, 1, address, sizeof(address), &port, &family); *out_value = port; }
    else if (strcmp(property, "remotePort") == 0) { tls_socket_address(socket->fd, 0, address, sizeof(address), &port, &family); *out_value = port; }
    else if (strcmp(property, "maxSendFragment") == 0) *out_value = (double)socket->max_send_fragment;
    return 0;
}

int scriptgo_tls_socket_bool(double handle, const char *property, int32_t *out_value) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (out_value == NULL || property == NULL) return tls_fail("TLS socket bool arguments are invalid");
    if (strcmp(property, "authorized") != 0 && strcmp(property, "sessionReused") != 0 && strcmp(property, "handshakeFinished") != 0) {
        return tls_fail("TLS socket bool property is unknown");
    }
    *out_value = 0;
    if (socket == NULL || socket->ssl == NULL) return 0;
    if (strcmp(property, "authorized") == 0) {
        X509 *certificate = SSL_get_peer_certificate(socket->ssl);
        *out_value = SSL_get_verify_result(socket->ssl) == X509_V_OK && certificate != NULL;
        X509_free(certificate);
    }
    else if (strcmp(property, "sessionReused") == 0) *out_value = SSL_session_reused(socket->ssl) != 0;
    else if (strcmp(property, "handshakeFinished") == 0) *out_value = SSL_is_init_finished(socket->ssl) != 0;
    return 0;
}

int scriptgo_tls_socket_export_keying_material(double handle, double length, const char *label,
                                               void *context_data, char **out_hex) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || label == NULL || out_hex == NULL || !tls_integer_in_range(length, 0, 65536)) return tls_fail("TLS keying material arguments are invalid");
    scriptgo_tls_typed_array_view *context_view = (scriptgo_tls_typed_array_view *)context_data;
    const unsigned char *context = NULL;
    size_t context_length = 0;
    int use_context = 0;
    if (context_data != NULL) {
        if (!tls_is_byte_view(context_view) || context_view->length < 0 || context_view->element_size <= 0 ||
            (context_view->length != 0 && context_view->data == NULL) ||
            (uint64_t)context_view->length > SIZE_MAX / (uint64_t)context_view->element_size) {
            return tls_fail("TLS keying material context must be a byte view");
        }
        context = context_view->data;
        context_length = (size_t)context_view->length * (size_t)context_view->element_size;
        use_context = 1;
    }
    size_t size = (size_t)length; unsigned char *data = (unsigned char *)malloc(size == 0 ? 1 : size);
    if (data == NULL) return tls_fail("TLS keying material allocation failed");
    int result = SSL_export_keying_material(socket->ssl, data, size, label, strlen(label), context, context_length, use_context);
    if (result != 1) { free(data); return tls_openssl_fail("TLS keying material export failed"); }
    *out_hex = tls_hex(data, size); free(data);
    return *out_hex == NULL ? tls_fail("TLS keying material allocation failed") : 0;
}

int scriptgo_tls_socket_set_option(double handle, const char *option, double value) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || option == NULL) return tls_fail("TLS socket option arguments are invalid");
    if (strcmp(option, "disableRenegotiation") == 0) { SSL_set_options(socket->ssl, SSL_OP_NO_RENEGOTIATION); return 0; }
    if (strcmp(option, "trace") == 0) {
        SSL_set_msg_callback(socket->ssl, value != 0 ? tls_trace_callback : NULL);
        SSL_set_msg_callback_arg(socket->ssl, NULL);
        return 0;
    }
    if (strcmp(option, "maxSendFragment") == 0) {
        if (value < 512 || value > 16384 || value != (double)(int)value) return tls_fail("TLS max send fragment is out of range");
        SSL_set_max_send_fragment(socket->ssl, (uint16_t)value);
        socket->max_send_fragment = (int)value;
        return 0;
    }
    if (strcmp(option, "requestCert") == 0) {
        SSL_set_verify(socket->ssl, value != 0 ? SSL_VERIFY_PEER | SSL_VERIFY_FAIL_IF_NO_PEER_CERT : SSL_VERIFY_NONE, NULL);
        return 0;
    }
    if (strcmp(option, "rejectUnauthorized") == 0) {
        SSL_set_verify(socket->ssl, value != 0 ? SSL_VERIFY_PEER : SSL_VERIFY_NONE, NULL);
        return 0;
    }
    return tls_fail("TLS socket option is unknown");
}

int scriptgo_tls_socket_set_servername(double handle, const char *servername) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || servername == NULL || *servername == '\0') {
        return tls_fail("TLS server name arguments are invalid");
    }
    if (SSL_is_init_finished(socket->ssl)) return tls_fail("TLS server name must be set before the handshake");
    if (SSL_set_tlsext_host_name(socket->ssl, servername) != 1 || SSL_set1_host(socket->ssl, servername) != 1) {
        return tls_openssl_fail("TLS server name configuration failed");
    }
    return 0;
}

int scriptgo_tls_socket_set_session(double handle, void *session_data) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL) return tls_fail("TLS session arguments are invalid");
    if (SSL_is_init_finished(socket->ssl)) return tls_fail("TLS session must be set before the handshake");
    return tls_set_session_from_view(socket->ssl, session_data, 1);
}

int scriptgo_tls_socket_set_key_cert(double handle, const char *cert, const char *key) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || cert == NULL || key == NULL) return tls_fail("TLS key and certificate arguments are invalid");
    BIO *cert_bio = BIO_new_mem_buf(cert, -1); BIO *key_bio = BIO_new_mem_buf(key, -1);
    X509 *certificate = cert_bio == NULL ? NULL : PEM_read_bio_X509(cert_bio, NULL, NULL, NULL);
    EVP_PKEY *private_key = key_bio == NULL ? NULL : PEM_read_bio_PrivateKey(key_bio, NULL, NULL, NULL);
    BIO_free(cert_bio); BIO_free(key_bio);
    if (certificate == NULL || private_key == NULL) { X509_free(certificate); EVP_PKEY_free(private_key); return tls_openssl_fail("TLS key and certificate parsing failed"); }
    int result = SSL_use_certificate(socket->ssl, certificate) == 1 && SSL_use_PrivateKey(socket->ssl, private_key) == 1;
    X509_free(certificate); EVP_PKEY_free(private_key);
    if (!result || SSL_check_private_key(socket->ssl) != 1) return tls_openssl_fail("TLS key and certificate configuration failed");
    return 0;
}

int scriptgo_tls_socket_renegotiate(double handle, int32_t *out_value) {
    scriptgo_tls_socket *socket = tls_socket_from_handle(handle);
    if (socket == NULL || socket->ssl == NULL || out_value == NULL) return tls_fail("TLS renegotiation arguments are invalid");
    *out_value = SSL_renegotiate(socket->ssl) == 1 ? 1 : 0;
    return 0;
}

int scriptgo_tls_server_listen(double context_handle, double request_cert, double reject,
                              const char *host, double port, double backlog, double *out_handle) {
    SSL_CTX *ctx = tls_context_value(context_handle);
    if (ctx == NULL || out_handle == NULL || !tls_integer_in_range(port, 0, 65535) || !tls_integer_in_range(backlog, 1, INT_MAX)) return tls_fail("TLS server listen arguments are invalid");
    struct addrinfo hints; struct addrinfo *results = NULL; memset(&hints, 0, sizeof(hints));
    hints.ai_socktype = SOCK_STREAM; hints.ai_family = AF_UNSPEC; hints.ai_flags = AI_PASSIVE;
    char port_text[32]; snprintf(port_text, sizeof(port_text), "%d", (int)port);
    int lookup = getaddrinfo(host == NULL || *host == '\0' ? NULL : host, port_text, &hints, &results);
    if (lookup != 0 || results == NULL) return tls_fail(gai_strerror(lookup));
    int fd = -1;
    for (struct addrinfo *entry = results; entry != NULL; entry = entry->ai_next) {
        fd = (int)socket(entry->ai_family, entry->ai_socktype, entry->ai_protocol); if (fd < 0) continue;
        int reuse = 1; setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, (const char *)&reuse, sizeof(reuse));
        if (bind(fd, entry->ai_addr, entry->ai_addrlen) == 0 && listen(fd, (int)backlog) == 0) break;
        tls_close_fd(fd); fd = -1;
    }
    freeaddrinfo(results); if (fd < 0) return tls_fail(strerror(errno));
    scriptgo_tls_server *server = (scriptgo_tls_server *)calloc(1, sizeof(*server));
    if (server == NULL) { tls_close_fd(fd); return tls_fail("TLS server allocation failed"); }
    server->ctx = ctx; SSL_CTX_up_ref(ctx); server->fd = fd; server->request_cert = request_cert != 0; server->reject = reject != 0;
    if (RAND_bytes(server->ticket_keys, sizeof(server->ticket_keys)) != 1) {
        int result = tls_openssl_fail("TLS ticket key generation failed");
        SSL_CTX_free(server->ctx);
        tls_close_fd(fd);
        free(server);
        return result;
    }
    if (tls_configure_ticket_callback(ctx) != 0) {
        SSL_CTX_free(server->ctx);
        tls_close_fd(fd);
        free(server);
        return -1;
    }
    SSL_CTX_set_tlsext_servername_callback(ctx, tls_sni_callback); SSL_CTX_set_tlsext_servername_arg(ctx, server);
    *out_handle = (double)(uintptr_t)server; return 0;
}

int scriptgo_tls_server_accept(double server_handle, double *out_handle) {
    scriptgo_tls_server *server = tls_server_from_handle(server_handle);
    if (server == NULL || server->fd < 0 || out_handle == NULL) return tls_fail("TLS server accept arguments are invalid");
    struct sockaddr_storage address; socklen_t length = sizeof(address);
    int fd = (int)accept(server->fd, (struct sockaddr *)&address, &length);
    if (fd < 0) return tls_fail(strerror(errno));
    SSL *ssl = SSL_new(server->ctx);
    if (ssl == NULL || SSL_set_fd(ssl, fd) != 1) { SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS accepted socket allocation failed"); }
    if (tls_ticket_ex_index_init() != 0 || SSL_set_ex_data(ssl, tls_ticket_ex_index, server) != 1) {
        SSL_free(ssl);
        tls_close_fd(fd);
        return tls_fail("TLS ticket callback state configuration failed");
    }
    SSL_set_accept_state(ssl); SSL_set_verify(ssl, server->reject ? SSL_VERIFY_PEER | (server->request_cert ? SSL_VERIFY_FAIL_IF_NO_PEER_CERT : 0) : SSL_VERIFY_NONE, NULL);
    if (SSL_accept(ssl) != 1) { SSL_free(ssl); tls_close_fd(fd); return tls_openssl_fail("TLS accepted socket handshake failed"); }
    int result = tls_socket_new(ssl, fd, 0, 1, out_handle);
    if (result != 0) {
        SSL_free(ssl);
        tls_close_fd(fd);
    }
    return result;
}

int scriptgo_tls_server_close(double handle) {
    scriptgo_tls_server *server = tls_server_from_handle(handle);
    if (server == NULL) return 0;
    if (server->fd >= 0) tls_close_fd(server->fd);
    tls_sni_free(server->sni); SSL_CTX_free(server->ctx); free(server); return 0;
}

int scriptgo_tls_server_info(double handle, const char *property, char **out_value) {
    scriptgo_tls_server *server = tls_server_from_handle(handle);
    if (server == NULL || property == NULL || out_value == NULL) return tls_fail("TLS server info arguments are invalid");
    if (strcmp(property, "ticketKeys") == 0) { *out_value = tls_hex(server->ticket_keys, sizeof(server->ticket_keys)); return 0; }
    if (strcmp(property, "address") == 0) {
        char address[INET6_ADDRSTRLEN] = ""; const char *family = "IPv4"; double port = 0;
        if (tls_socket_address(server->fd, 1, address, sizeof(address), &port, &family) != 0) return tls_fail("TLS server address lookup failed");
        char value[256]; snprintf(value, sizeof(value), "{\"port\":%.0f,\"family\":\"%s\",\"address\":\"%s\"}", port, family, address);
        *out_value = tls_strdup(value); return 0;
    }
    return tls_fail("TLS server info property is unknown");
}

int scriptgo_tls_server_set_context(double handle, double context_handle, double request_cert, double reject) {
    scriptgo_tls_server *server = tls_server_from_handle(handle); SSL_CTX *ctx = tls_context_value(context_handle);
    if (server == NULL || ctx == NULL) return tls_fail("TLS server context arguments are invalid");
    if (tls_configure_ticket_callback(ctx) != 0) return -1;
    SSL_CTX_up_ref(ctx); SSL_CTX *old = server->ctx; server->ctx = ctx; server->request_cert = request_cert != 0; server->reject = reject != 0;
    SSL_CTX_set_tlsext_servername_callback(ctx, tls_sni_callback); SSL_CTX_set_tlsext_servername_arg(ctx, server); SSL_CTX_free(old); return 0;
}

int scriptgo_tls_server_add_context(double handle, const char *hostname, double context_handle) {
    scriptgo_tls_server *server = tls_server_from_handle(handle); SSL_CTX *ctx = tls_context_value(context_handle);
    if (server == NULL || ctx == NULL || hostname == NULL || *hostname == '\0') return tls_fail("TLS SNI context arguments are invalid");
    if (tls_configure_ticket_callback(ctx) != 0) return -1;
    scriptgo_tls_sni_entry *entry = (scriptgo_tls_sni_entry *)calloc(1, sizeof(*entry)); if (entry == NULL) return tls_fail("TLS SNI context allocation failed");
    entry->hostname = tls_strdup(hostname); if (entry->hostname == NULL) { free(entry); return tls_fail("TLS SNI hostname allocation failed"); }
    entry->ctx = ctx; SSL_CTX_up_ref(ctx); entry->next = server->sni; server->sni = entry; return 0;
}

int scriptgo_tls_server_set_ticket_keys(double handle, const char *hex) {
    scriptgo_tls_server *server = tls_server_from_handle(handle);
    unsigned char keys[sizeof(server->ticket_keys)];
    if (server == NULL || tls_hex_decode(hex, keys, sizeof(keys)) != 0) return tls_fail("TLS ticket keys must contain 48 bytes");
    memcpy(server->ticket_keys, keys, sizeof(keys));
#if OPENSSL_VERSION_NUMBER >= 0x30000000L
    return 0;
#else
    return SSL_CTX_set_tlsext_ticket_keys(server->ctx, server->ticket_keys, sizeof(server->ticket_keys)) == 1 ? 0 : tls_openssl_fail("TLS ticket key configuration failed");
#endif
}

int scriptgo_tls_x509_parse_pem(const char *pem, char **out_json) {
    if (pem == NULL || out_json == NULL) return tls_fail("X509 certificate arguments are invalid");
    BIO *bio = BIO_new_mem_buf(pem, -1); X509 *certificate = bio == NULL ? NULL : PEM_read_bio_X509(bio, NULL, NULL, NULL); BIO_free(bio);
    if (certificate == NULL) return tls_openssl_fail("X509 certificate parsing failed");
    *out_json = tls_certificate_json(certificate, 1); X509_free(certificate); return *out_json == NULL ? tls_fail("X509 certificate allocation failed") : 0;
}

int scriptgo_tls_x509_parse_bytes(void *data, char **out_json) {
    scriptgo_tls_typed_array_view *view = (scriptgo_tls_typed_array_view *)data;
    if (!tls_is_byte_view(view) || out_json == NULL || view->length <= 0 || view->element_size != 1 || view->data == NULL) return tls_fail("X509 certificate bytes are invalid");
    const unsigned char *cursor = view->data; X509 *certificate = d2i_X509(NULL, &cursor, view->length);
    if (certificate == NULL) return tls_openssl_fail("X509 DER certificate parsing failed");
    *out_json = tls_certificate_json(certificate, 1); X509_free(certificate); return *out_json == NULL ? tls_fail("X509 certificate allocation failed") : 0;
}

int scriptgo_tls_ciphers(char **out_json) {
    if (out_json == NULL) return tls_fail("TLS cipher output is null");
    SSL_CTX *ctx = SSL_CTX_new(TLS_method()); if (ctx == NULL) return tls_openssl_fail("TLS cipher context allocation failed");
    STACK_OF(SSL_CIPHER) *ciphers = SSL_CTX_get_ciphers(ctx); char *result = NULL; size_t length = 0; size_t capacity = 0;
    if (tls_append(&result, &length, &capacity, "[") != 0) { SSL_CTX_free(ctx); return tls_fail("TLS cipher allocation failed"); }
    for (int i = 0; i < sk_SSL_CIPHER_num(ciphers); i++) {
        if (i != 0 && tls_append(&result, &length, &capacity, ",") != 0) { free(result); SSL_CTX_free(ctx); return tls_fail("TLS cipher allocation failed"); }
        if (tls_append_json_string(&result, &length, &capacity, SSL_CIPHER_get_name(sk_SSL_CIPHER_value(ciphers, i))) != 0) {
            free(result); SSL_CTX_free(ctx); return tls_fail("TLS cipher allocation failed");
        }
    }
    if (tls_append(&result, &length, &capacity, "]") != 0) { free(result); SSL_CTX_free(ctx); return tls_fail("TLS cipher allocation failed"); }
    SSL_CTX_free(ctx); *out_json = result; return 0;
}

static int tls_append_bundle(char **result, size_t *length, size_t *capacity, const char *path, int *count) {
    FILE *file = fopen(path, "rb"); if (file == NULL) return 0;
    if (fseek(file, 0, SEEK_END) != 0) { fclose(file); return -1; }
    long size = ftell(file);
    if (size < 0 || fseek(file, 0, SEEK_SET) != 0) { fclose(file); return -1; }
    if (size <= 0 || size > 64 * 1024 * 1024) { fclose(file); return 0; }
    char *data = (char *)malloc((size_t)size + 1); if (data == NULL) { fclose(file); return -1; }
    size_t read_size = fread(data, 1, (size_t)size, file); fclose(file);
    if (read_size != (size_t)size) { free(data); return -1; }
    data[size] = '\0';
    const char *cursor = data;
    while ((cursor = strstr(cursor, "-----BEGIN CERTIFICATE-----")) != NULL) {
        const char *end = strstr(cursor, "-----END CERTIFICATE-----"); if (end == NULL) break; end += strlen("-----END CERTIFICATE-----");
        size_t cert_len = (size_t)(end - cursor); char *certificate = (char *)malloc(cert_len + 2); if (certificate == NULL) { free(data); return -1; }
        memcpy(certificate, cursor, cert_len); certificate[cert_len] = '\n'; certificate[cert_len + 1] = '\0';
        if (*count != 0 && tls_append(result, length, capacity, ",") != 0) { free(certificate); free(data); return -1; }
        if (tls_append_json_string(result, length, capacity, certificate) != 0) { free(certificate); free(data); return -1; }
        (*count)++; free(certificate); cursor = end;
    }
    free(data); return 0;
}

static int tls_append_bundle_directory(char **result, size_t *length, size_t *capacity,
                                        const char *path, int *count) {
#if defined(_WIN32)
    if (path == NULL || *path == '\0') return 0;
    size_t directory_length = strlen(path);
    if (directory_length > SIZE_MAX - 3) return -1;
    char *pattern = (char *)malloc(directory_length + 3);
    if (pattern == NULL) return -1;
    size_t pattern_length = directory_length;
    memcpy(pattern, path, directory_length);
    if (pattern_length == 0 || (pattern[pattern_length - 1] != '\\' && pattern[pattern_length - 1] != '/')) {
        pattern[pattern_length++] = '\\';
    }
    pattern[pattern_length++] = '*';
    pattern[pattern_length] = '\0';

    WIN32_FIND_DATAA entry_data;
    HANDLE search = FindFirstFileA(pattern, &entry_data);
    free(pattern);
    if (search == INVALID_HANDLE_VALUE) return 0;
    do {
        if ((entry_data.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0 || entry_data.cFileName[0] == '.') continue;
        size_t name_length = strlen(entry_data.cFileName);
        size_t separator = directory_length > 0 && (path[directory_length - 1] == '\\' || path[directory_length - 1] == '/') ? 0 : 1;
        if (directory_length > SIZE_MAX - name_length - separator - 1) {
            FindClose(search);
            return -1;
        }
        char *file_path = (char *)malloc(directory_length + separator + name_length + 1);
        if (file_path == NULL) {
            FindClose(search);
            return -1;
        }
        memcpy(file_path, path, directory_length);
        if (separator != 0) file_path[directory_length] = '\\';
        memcpy(file_path + directory_length + separator, entry_data.cFileName, name_length + 1);
        int result_code = tls_append_bundle(result, length, capacity, file_path, count);
        free(file_path);
        if (result_code != 0) {
            FindClose(search);
            return -1;
        }
    } while (FindNextFileA(search, &entry_data) != 0);
    FindClose(search);
    return 0;
#else
    if (path == NULL || *path == '\0') return 0;
    DIR *directory = opendir(path);
    if (directory == NULL) return 0;
    struct dirent *entry;
    while ((entry = readdir(directory)) != NULL) {
        if (entry->d_name[0] == '.') continue;
        size_t path_length = strlen(path);
        size_t name_length = strlen(entry->d_name);
        if (path_length > SIZE_MAX - name_length - 2) { closedir(directory); return -1; }
        char *file_path = (char *)malloc(path_length + name_length + 2);
        if (file_path == NULL) { closedir(directory); return -1; }
        memcpy(file_path, path, path_length);
        file_path[path_length] = '/';
        memcpy(file_path + path_length + 1, entry->d_name, name_length + 1);
        struct stat status;
        int is_regular = stat(file_path, &status) == 0 && S_ISREG(status.st_mode);
        int result_code = is_regular ? tls_append_bundle(result, length, capacity, file_path, count) : 0;
        free(file_path);
        if (result_code != 0) { closedir(directory); return -1; }
    }
    closedir(directory);
    return 0;
#endif
}

#if defined(__APPLE__)
typedef enum {
    TLS_TRUST_UNSPECIFIED = 0,
    TLS_TRUST_TRUSTED = 1,
    TLS_TRUST_DISTRUSTED = 2,
} tls_trust_status;

static int tls_macos_trust_dictionary(CFDictionaryRef trust_dict, int self_issued) {
    if (trust_dict == NULL) return TLS_TRUST_UNSPECIFIED;
    if (CFDictionaryContainsKey(trust_dict, kSecTrustSettingsApplication) ||
        CFDictionaryContainsKey(trust_dict, kSecTrustSettingsPolicyString)) {
        return TLS_TRUST_UNSPECIFIED;
    }
    if (CFDictionaryContainsKey(trust_dict, kSecTrustSettingsPolicy)) {
        SecPolicyRef policy = (SecPolicyRef)CFDictionaryGetValue(trust_dict, kSecTrustSettingsPolicy);
        if (policy == NULL) return TLS_TRUST_UNSPECIFIED;
        CFDictionaryRef properties = SecPolicyCopyProperties(policy);
        CFStringRef policy_oid = properties == NULL
            ? NULL
            : (CFStringRef)CFDictionaryGetValue(properties, kSecPolicyOid);
        int matches_ssl = policy_oid != NULL && CFEqual(policy_oid, kSecPolicyAppleSSL);
        if (properties != NULL) CFRelease(properties);
        if (!matches_ssl) return TLS_TRUST_UNSPECIFIED;
    }
    int trust_result = kSecTrustSettingsResultTrustRoot;
    if (CFDictionaryContainsKey(trust_dict, kSecTrustSettingsResult)) {
        CFNumberRef value = (CFNumberRef)CFDictionaryGetValue(trust_dict, kSecTrustSettingsResult);
        if (value == NULL || !CFNumberGetValue(value, kCFNumberIntType, &trust_result)) return TLS_TRUST_UNSPECIFIED;
        if (trust_result == kSecTrustSettingsResultDeny) return TLS_TRUST_DISTRUSTED;
        if (self_issued) {
            return trust_result == kSecTrustSettingsResultTrustRoot || trust_result == kSecTrustSettingsResultTrustAsRoot
                ? TLS_TRUST_TRUSTED : TLS_TRUST_UNSPECIFIED;
        }
        return trust_result == kSecTrustSettingsResultTrustAsRoot ? TLS_TRUST_TRUSTED : TLS_TRUST_UNSPECIFIED;
    }
    return TLS_TRUST_UNSPECIFIED;
}

static int tls_macos_trust_settings(CFArrayRef trust_settings, int self_issued) {
    if (trust_settings == NULL) return TLS_TRUST_UNSPECIFIED;
    if (CFArrayGetCount(trust_settings) == 0) return self_issued ? TLS_TRUST_TRUSTED : TLS_TRUST_UNSPECIFIED;
    for (CFIndex i = 0; i < CFArrayGetCount(trust_settings); i++) {
        int result = tls_macos_trust_dictionary((CFDictionaryRef)CFArrayGetValueAtIndex(trust_settings, i), self_issued);
        if (result == TLS_TRUST_TRUSTED || result == TLS_TRUST_DISTRUSTED) return result;
    }
    return TLS_TRUST_UNSPECIFIED;
}

static int tls_macos_certificate_trusted(SecCertificateRef certificate) {
    CFMutableArrayRef certificates = CFArrayCreateMutable(NULL, 1, &kCFTypeArrayCallBacks);
    if (certificates == NULL) return 0;
    CFArrayAppendValue(certificates, certificate);
    SecPolicyRef policy = SecPolicyCreateSSL(false, NULL);
    SecTrustRef trust = NULL;
    int trusted = 0;
    if (policy != NULL && SecTrustCreateWithCertificates(certificates, policy, &trust) == errSecSuccess) {
        trusted = SecTrustEvaluateWithError(trust, NULL) ? 1 : 0;
    }
    if (trust != NULL) CFRelease(trust);
    if (policy != NULL) CFRelease(policy);
    CFRelease(certificates);
    return trusted;
}

static int tls_macos_certificate_trusted_for_policy(X509 *certificate, SecCertificateRef reference) {
    int self_issued = X509_NAME_cmp(X509_get_subject_name(certificate), X509_get_issuer_name(certificate)) == 0;
    int trust_evaluated = 0;
    const SecTrustSettingsDomain domains[] = {kSecTrustSettingsDomainUser, kSecTrustSettingsDomainAdmin};
    for (size_t i = 0; i < sizeof(domains) / sizeof(domains[0]); i++) {
        CFArrayRef trust_settings = NULL;
        OSStatus status = SecTrustSettingsCopyTrustSettings(reference, domains[i], &trust_settings);
        if (status != errSecSuccess && status != errSecItemNotFound) continue;
        if (status == errSecSuccess && trust_settings != NULL) {
            int result = tls_macos_trust_settings(trust_settings, self_issued);
            if (result == TLS_TRUST_TRUSTED || result == TLS_TRUST_DISTRUSTED) {
                CFRelease(trust_settings);
                return result == TLS_TRUST_TRUSTED;
            }
        }
        if (trust_settings == NULL && !trust_evaluated) {
            if (tls_macos_certificate_trusted(reference)) return 1;
            trust_evaluated = 1;
        } else if (trust_settings != NULL) {
            CFRelease(trust_settings);
        }
    }
    return 0;
}

static int tls_append_macos_system_certificates(char **result, size_t *length, size_t *capacity, int *count) {
    const void *keys[] = {kSecClass, kSecMatchLimit, kSecReturnRef};
    const void *values[] = {kSecClassCertificate, kSecMatchLimitAll, kCFBooleanTrue};
    CFDictionaryRef query = CFDictionaryCreate(NULL, keys, values, 3,
                                                &kCFTypeDictionaryKeyCallBacks,
                                                &kCFTypeDictionaryValueCallBacks);
    if (query == NULL) return -1;
    CFArrayRef certificates = NULL;
    OSStatus status = SecItemCopyMatching(query, (CFTypeRef *)&certificates);
    CFRelease(query);
    if (status == errSecItemNotFound) return 0;
    if (status != errSecSuccess || certificates == NULL) return -1;
    for (CFIndex i = 0; i < CFArrayGetCount(certificates); i++) {
        SecCertificateRef reference = (SecCertificateRef)CFArrayGetValueAtIndex(certificates, i);
        CFDataRef der = SecCertificateCopyData(reference);
        if (der == NULL) continue;
        const unsigned char *cursor = CFDataGetBytePtr(der);
        X509 *certificate = d2i_X509(NULL, &cursor, CFDataGetLength(der));
        CFRelease(der);
        if (certificate == NULL) continue;
        if (tls_macos_certificate_trusted_for_policy(certificate, reference)) {
            char *pem = tls_certificate_pem(certificate);
            if (pem == NULL || (*count != 0 && tls_append(result, length, capacity, ",") != 0) ||
                (pem != NULL && tls_append_json_string(result, length, capacity, pem) != 0)) {
                free(pem);
                X509_free(certificate);
                CFRelease(certificates);
                return -1;
            }
            (*count)++;
            free(pem);
        }
        X509_free(certificate);
    }
    CFRelease(certificates);
    return 0;
}
#endif

static int tls_append_system_directory_list(char **result, size_t *length, size_t *capacity,
                                             const char *paths, int *count) {
    if (paths == NULL || *paths == '\0') return 0;
    char *copy = tls_strdup(paths);
    if (copy == NULL) return -1;
#if defined(_WIN32)
    const char separator = ';';
#else
    const char separator = ':';
#endif
    char *cursor = copy;
    while (cursor != NULL) {
        char *next = strchr(cursor, separator);
        if (next != NULL) *next++ = '\0';
        if (tls_append_bundle_directory(result, length, capacity, cursor, count) != 0) {
            free(copy);
            return -1;
        }
        cursor = next;
    }
    free(copy);
    return 0;
}

int scriptgo_tls_root_certificates(char **out_json) {
    if (out_json == NULL) return tls_fail("TLS root certificate output is null");
    char *result = NULL; size_t length = 0; size_t capacity = 0; int count = 0;
    if (tls_append(&result, &length, &capacity, "[") != 0) return tls_fail("TLS root certificate allocation failed");
    size_t bundled_count = sizeof(scriptgo_tls_bundled_root_certificates) / sizeof(scriptgo_tls_bundled_root_certificates[0]);
    for (size_t i = 0; i < bundled_count; i++) {
        if (count != 0 && tls_append(&result, &length, &capacity, ",") != 0) { free(result); return tls_fail("TLS root certificate allocation failed"); }
        if (tls_append_json_string(&result, &length, &capacity, scriptgo_tls_bundled_root_certificates[i]) != 0) { free(result); return tls_fail("TLS root certificate allocation failed"); }
        count++;
    }
    if (tls_append(&result, &length, &capacity, "]") != 0) { free(result); return tls_fail("TLS root certificate allocation failed"); }
    *out_json = result; return 0;
}

int scriptgo_tls_system_certificates(char **out_json) {
    if (out_json == NULL) return tls_fail("TLS system certificate output is null");
    char *result = NULL; size_t length = 0; size_t capacity = 0; int count = 0;
    if (tls_append(&result, &length, &capacity, "[") != 0) return tls_fail("TLS system certificate allocation failed");
#if defined(__APPLE__)
    if (tls_append_macos_system_certificates(&result, &length, &capacity, &count) != 0) { free(result); return tls_fail("TLS system certificate allocation failed"); }
#else
    const char *file = getenv("SSL_CERT_FILE");
    if (file == NULL || *file == '\0') file = X509_get_default_cert_file();
    if (tls_append_bundle(&result, &length, &capacity, file, &count) != 0) { free(result); return tls_fail("TLS system certificate allocation failed"); }
    const char *directory = getenv("SSL_CERT_DIR");
    if (directory == NULL || *directory == '\0') directory = X509_get_default_cert_dir();
    if (tls_append_system_directory_list(&result, &length, &capacity, directory, &count) != 0) { free(result); return tls_fail("TLS system certificate allocation failed"); }
#endif
    if (tls_append(&result, &length, &capacity, "]") != 0) { free(result); return tls_fail("TLS system certificate allocation failed"); }
    *out_json = result; return 0;
}

int scriptgo_tls_extra_certificates(char **out_json) {
    if (out_json == NULL) return tls_fail("TLS extra certificate output is null");
    char *result = NULL; size_t length = 0; size_t capacity = 0; int count = 0;
    if (tls_append(&result, &length, &capacity, "[") != 0) return tls_fail("TLS extra certificate allocation failed");
    const char *path = getenv("NODE_EXTRA_CA_CERTS");
    if (path != NULL && *path != '\0' && tls_append_bundle(&result, &length, &capacity, path, &count) != 0) {
        free(result);
        return tls_fail("TLS extra certificate allocation failed");
    }
    if (tls_append(&result, &length, &capacity, "]") != 0) { free(result); return tls_fail("TLS extra certificate allocation failed"); }
    *out_json = result; return 0;
}

#endif
