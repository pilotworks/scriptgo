#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#if defined(SCRIPTGO_HAS_OPENSSL)
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/rsa.h>
#include <openssl/ec.h>
#include <openssl/bn.h>
#include <openssl/dh.h>
#include <openssl/x509.h>
#include <openssl/crypto.h>
#include <openssl/rand.h>
#include <openssl/err.h>
#include <openssl/core_names.h>
#include <openssl/param_build.h>

int scriptgo_runtime_set_error(const char *message);
int scriptgo_buffer_alloc(double size, const char *fill_str, double fill_num, int has_fill, int is_str_fill, void **out_buf);

typedef struct {
    uint32_t magic;
    int32_t kind;
    int64_t length;
    int64_t byte_offset;
    int64_t element_size;
    void *buffer;
    unsigned char *data;
} scriptgo_crypto_buffer_view_evp;

static int evp_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

static void get_buf_data_len(void *buf, const unsigned char **out_data, size_t *out_len) {
    if (!buf) {
        *out_data = NULL;
        *out_len = 0;
        return;
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)buf;
    *out_data = bv->data;
    *out_len = (size_t)bv->length;
}

typedef struct {
    EVP_CIPHER_CTX *ctx;
    int is_encrypt;
    int auto_padding;
    int is_aead;
    unsigned char tag[16];
    int tag_len;
    int has_tag;
} scriptgo_cipher_handle;

int scriptgo_crypto_cipher_create(const char *algo, void *key_buf, void *iv_buf, double is_encrypt_num, void **out_ctx) {
    if (!algo || !out_ctx) return evp_fail("createCipheriv: invalid arguments");
    const unsigned char *key = NULL; size_t key_len = 0;
    const unsigned char *iv = NULL; size_t iv_len = 0;
    get_buf_data_len(key_buf, &key, &key_len);
    get_buf_data_len(iv_buf, &iv, &iv_len);
    int is_encrypt = (is_encrypt_num != 0.0) ? 1 : 0;

    const EVP_CIPHER *c = EVP_CIPHER_fetch(NULL, algo, NULL);
    if (!c) c = EVP_get_cipherbyname(algo);
    if (!c) return evp_fail("Unknown cipher algorithm");

    EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
    if (!ctx) return evp_fail("Failed to allocate cipher context");

    int mode = EVP_CIPHER_get_mode(c);
    int is_aead = (mode == EVP_CIPH_GCM_MODE || mode == EVP_CIPH_CCM_MODE || mode == EVP_CIPH_OCB_MODE);

    if (EVP_CipherInit_ex(ctx, c, NULL, NULL, NULL, is_encrypt) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return evp_fail("Failed to initialize cipher");
    }

    if (is_aead && iv_len > 0) {
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_IVLEN, (int)iv_len, NULL) != 1) {
            EVP_CIPHER_CTX_free(ctx);
            return evp_fail("Failed to set AEAD IV length");
        }
    }

    if (EVP_CipherInit_ex(ctx, NULL, NULL, key, iv, is_encrypt) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return evp_fail("Failed to initialize cipher key/iv");
    }

    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)malloc(sizeof(scriptgo_cipher_handle));
    if (!handle) {
        EVP_CIPHER_CTX_free(ctx);
        return evp_fail("Allocation failure");
    }
    handle->ctx = ctx;
    handle->is_encrypt = is_encrypt;
    handle->auto_padding = 1;
    handle->is_aead = is_aead;
    handle->tag_len = 0;
    handle->has_tag = 0;
    memset(handle->tag, 0, sizeof(handle->tag));

    *out_ctx = handle;
    return 0;
}

int scriptgo_crypto_cipher_update(void *h, void *in_buf, void **out_buf) {
    if (!h || !out_buf) return evp_fail("cipher.update: null handle");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    const unsigned char *in = NULL; size_t in_len = 0;
    get_buf_data_len(in_buf, &in, &in_len);

    int block_sz = EVP_CIPHER_CTX_get_block_size(handle->ctx);
    if (block_sz <= 0) block_sz = 16;
    size_t alloc_sz = in_len + (size_t)block_sz;

    if (scriptgo_buffer_alloc((double)alloc_sz, NULL, 0, 0, 0, out_buf) != 0) {
        return evp_fail("cipher.update: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    int out_len = 0;
    if (EVP_CipherUpdate(handle->ctx, bv->data, &out_len, in ? in : (const unsigned char *)"", (int)in_len) != 1) {
        return evp_fail("cipher.update: operation failed");
    }
    bv->length = out_len;
    return 0;
}

int scriptgo_crypto_cipher_final(void *h, void **out_buf) {
    if (!h || !out_buf) return evp_fail("cipher.final: null handle");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    if (!handle->is_encrypt && handle->is_aead && handle->has_tag) {
        if (EVP_CIPHER_CTX_ctrl(handle->ctx, EVP_CTRL_GCM_SET_TAG, handle->tag_len, handle->tag) != 1) {
            return evp_fail("decipher.final: failed to set auth tag");
        }
    }
    int block_sz = EVP_CIPHER_CTX_get_block_size(handle->ctx);
    if (block_sz <= 0) block_sz = 16;

    if (scriptgo_buffer_alloc((double)block_sz, NULL, 0, 0, 0, out_buf) != 0) {
        return evp_fail("cipher.final: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    int out_len = 0;
    if (EVP_CipherFinal_ex(handle->ctx, bv->data, &out_len) != 1) {
        return evp_fail("cipher.final: operation failed");
    }
    bv->length = out_len;
    if (handle->is_encrypt && handle->is_aead) {
        if (EVP_CIPHER_CTX_ctrl(handle->ctx, EVP_CTRL_GCM_GET_TAG, 16, handle->tag) == 1) {
            handle->tag_len = 16;
            handle->has_tag = 1;
        }
    }
    return 0;
}

int scriptgo_crypto_cipher_set_aad(void *h, void *aad_buf) {
    if (!h) return evp_fail("cipher.setAAD: null handle");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    const unsigned char *aad = NULL; size_t aad_len = 0;
    get_buf_data_len(aad_buf, &aad, &aad_len);
    int dummy_len = 0;
    if (EVP_CipherUpdate(handle->ctx, NULL, &dummy_len, aad ? aad : (const unsigned char *)"", (int)aad_len) != 1) {
        return evp_fail("cipher.setAAD: operation failed");
    }
    return 0;
}

int scriptgo_crypto_cipher_get_tag(void *h, void **out_buf) {
    if (!h || !out_buf) return evp_fail("cipher.getAuthTag: null handle");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    if (!handle->has_tag && handle->is_encrypt && handle->is_aead) {
        if (EVP_CIPHER_CTX_ctrl(handle->ctx, EVP_CTRL_GCM_GET_TAG, 16, handle->tag) == 1) {
            handle->tag_len = 16;
            handle->has_tag = 1;
        }
    }
    int tag_len = handle->tag_len > 0 ? handle->tag_len : 16;
    if (scriptgo_buffer_alloc((double)tag_len, NULL, 0, 0, 0, out_buf) != 0) {
        return evp_fail("cipher.getAuthTag: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    memcpy(bv->data, handle->tag, (size_t)tag_len);
    bv->length = tag_len;
    return 0;
}

int scriptgo_crypto_cipher_set_tag(void *h, void *tag_buf) {
    if (!h || !tag_buf) return evp_fail("decipher.setAuthTag: null argument");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    const unsigned char *tag = NULL; size_t tag_len = 0;
    get_buf_data_len(tag_buf, &tag, &tag_len);
    if (tag_len > sizeof(handle->tag)) tag_len = sizeof(handle->tag);
    if (tag) memcpy(handle->tag, tag, tag_len);
    handle->tag_len = (int)tag_len;
    handle->has_tag = 1;
    return 0;
}

int scriptgo_crypto_cipher_set_auto_padding(void *h, double auto_padding_num) {
    if (!h) return evp_fail("cipher.setAutoPadding: null handle");
    scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
    handle->auto_padding = (auto_padding_num != 0.0) ? 1 : 0;
    EVP_CIPHER_CTX_set_padding(handle->ctx, handle->auto_padding);
    return 0;
}

int scriptgo_crypto_cipher_destroy(void *h) {
    if (h) {
        scriptgo_cipher_handle *handle = (scriptgo_cipher_handle *)h;
        if (handle->ctx) EVP_CIPHER_CTX_free(handle->ctx);
        free(handle);
    }
    return 0;
}

int scriptgo_crypto_get_cipher_info(const char *name_or_nid, char **out_json) {
    if (!name_or_nid || !out_json) return evp_fail("getCipherInfo: invalid arguments");
    const EVP_CIPHER *c = NULL;
    int nid = atoi(name_or_nid);
    if (nid > 0) {
        c = EVP_get_cipherbynid(nid);
    } else {
        c = EVP_CIPHER_fetch(NULL, name_or_nid, NULL);
        if (!c) c = EVP_get_cipherbyname(name_or_nid);
    }
    if (!c) {
        *out_json = strdup("{}");
        return 0;
    }
    const char *name = EVP_CIPHER_get0_name(c);
    int real_nid = EVP_CIPHER_get_nid(c);
    int block_sz = EVP_CIPHER_get_block_size(c);
    int key_len = EVP_CIPHER_get_key_length(c);
    int iv_len = EVP_CIPHER_get_iv_length(c);
    int mode_num = EVP_CIPHER_get_mode(c);
    const char *mode = "cbc";
    if (mode_num == EVP_CIPH_GCM_MODE) mode = "gcm";
    else if (mode_num == EVP_CIPH_CCM_MODE) mode = "ccm";
    else if (mode_num == EVP_CIPH_CFB_MODE) mode = "cfb";
    else if (mode_num == EVP_CIPH_OFB_MODE) mode = "ofb";
    else if (mode_num == EVP_CIPH_CTR_MODE) mode = "ctr";
    else if (mode_num == EVP_CIPH_ECB_MODE) mode = "ecb";
    else if (mode_num == EVP_CIPH_XTS_MODE) mode = "xts";

    char buf[512];
    snprintf(buf, sizeof(buf),
             "{\"name\":\"%s\",\"nid\":%d,\"blockSize\":%d,\"ivLength\":%d,\"keyLength\":%d,\"mode\":\"%s\"}",
             name ? name : "", real_nid, block_sz, iv_len, key_len, mode);
    *out_json = strdup(buf);
    return 0;
}

static EVP_PKEY *parse_any_key(const char *pem) {
    if (!pem) return NULL;
    BIO *bio = BIO_new_mem_buf(pem, -1);
    if (!bio) return NULL;
    EVP_PKEY *pkey = PEM_read_bio_PrivateKey(bio, NULL, NULL, NULL);
    if (!pkey) {
        BIO_reset(bio);
        pkey = PEM_read_bio_PUBKEY(bio, NULL, NULL, NULL);
    }
    if (!pkey) {
        BIO_reset(bio);
        X509 *x = PEM_read_bio_X509(bio, NULL, NULL, NULL);
        if (x) {
            pkey = X509_get_pubkey(x);
            X509_free(x);
        }
    }
    BIO_free(bio);
    return pkey;
}

int scriptgo_crypto_sign(const char *algo, void *data_buf, const char *key_pem, void **out_buf) {
    if (!algo || !key_pem || !out_buf) return evp_fail("sign: invalid arguments");
    const unsigned char *data = NULL; size_t data_len = 0;
    get_buf_data_len(data_buf, &data, &data_len);
    EVP_PKEY *pkey = parse_any_key(key_pem);
    if (!pkey) return evp_fail("sign: invalid private key");

    const EVP_MD *md = EVP_get_digestbyname(algo);
    if (!md) md = EVP_sha256();

    EVP_MD_CTX *mctx = EVP_MD_CTX_new();
    if (!mctx) { EVP_PKEY_free(pkey); return evp_fail("sign: context alloc failed"); }

    if (EVP_DigestSignInit(mctx, NULL, md, NULL, pkey) != 1 ||
        EVP_DigestSignUpdate(mctx, data ? data : (const unsigned char *)"", data_len) != 1) {
        EVP_MD_CTX_free(mctx); EVP_PKEY_free(pkey);
        return evp_fail("sign: init or update failed");
    }

    size_t sig_len = 0;
    if (EVP_DigestSignFinal(mctx, NULL, &sig_len) != 1 || sig_len == 0) {
        EVP_MD_CTX_free(mctx); EVP_PKEY_free(pkey);
        return evp_fail("sign: get signature size failed");
    }

    if (scriptgo_buffer_alloc((double)sig_len, NULL, 0, 0, 0, out_buf) != 0) {
        EVP_MD_CTX_free(mctx); EVP_PKEY_free(pkey);
        return evp_fail("sign: buffer allocation failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    if (EVP_DigestSignFinal(mctx, bv->data, &sig_len) != 1) {
        EVP_MD_CTX_free(mctx); EVP_PKEY_free(pkey);
        return evp_fail("sign: final signing failed");
    }
    bv->length = (int64_t)sig_len;
    EVP_MD_CTX_free(mctx);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_verify(const char *algo, void *data_buf, const char *key_pem, void *sig_buf, double *out_valid) {
    if (!algo || !key_pem || !out_valid) return evp_fail("verify: invalid arguments");
    *out_valid = 0.0;
    const unsigned char *data = NULL; size_t data_len = 0;
    const unsigned char *sig = NULL; size_t sig_len = 0;
    get_buf_data_len(data_buf, &data, &data_len);
    get_buf_data_len(sig_buf, &sig, &sig_len);

    EVP_PKEY *pkey = parse_any_key(key_pem);
    if (!pkey) return evp_fail("verify: invalid public key");

    const EVP_MD *md = EVP_get_digestbyname(algo);
    if (!md) md = EVP_sha256();

    EVP_MD_CTX *mctx = EVP_MD_CTX_new();
    if (!mctx) { EVP_PKEY_free(pkey); return evp_fail("verify: context alloc failed"); }

    if (EVP_DigestVerifyInit(mctx, NULL, md, NULL, pkey) != 1 ||
        EVP_DigestVerifyUpdate(mctx, data ? data : (const unsigned char *)"", data_len) != 1) {
        EVP_MD_CTX_free(mctx); EVP_PKEY_free(pkey);
        return evp_fail("verify: init or update failed");
    }

    int res = EVP_DigestVerifyFinal(mctx, sig ? sig : (const unsigned char *)"", sig_len);
    *out_valid = (res == 1) ? 1.0 : 0.0;
    EVP_MD_CTX_free(mctx);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_public_encrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    if (!key_pem || !out_buf) return evp_fail("publicEncrypt: invalid arguments");
    const unsigned char *b = NULL; size_t b_len = 0;
    get_buf_data_len(buf, &b, &b_len);
    int padding = (int)padding_num;

    EVP_PKEY *pkey = parse_any_key(key_pem);
    if (!pkey) return evp_fail("publicEncrypt: invalid key");
    EVP_PKEY_CTX *pctx = EVP_PKEY_CTX_new(pkey, NULL);
    if (!pctx) { EVP_PKEY_free(pkey); return evp_fail("publicEncrypt: ctx failed"); }
    if (EVP_PKEY_encrypt_init(pctx) <= 0) { EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("publicEncrypt: init failed"); }
    if (padding > 0) EVP_PKEY_CTX_set_rsa_padding(pctx, padding);
    size_t out_len = 0;
    if (EVP_PKEY_encrypt(pctx, NULL, &out_len, b ? b : (const unsigned char *)"", b_len) <= 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("publicEncrypt: size failed");
    }
    if (scriptgo_buffer_alloc((double)out_len, NULL, 0, 0, 0, out_buf) != 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("publicEncrypt: alloc failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    if (EVP_PKEY_encrypt(pctx, bv->data, &out_len, b ? b : (const unsigned char *)"", b_len) <= 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("publicEncrypt: encrypt failed");
    }
    bv->length = (int64_t)out_len;
    EVP_PKEY_CTX_free(pctx);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_private_decrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    if (!key_pem || !out_buf) return evp_fail("privateDecrypt: invalid arguments");
    const unsigned char *b = NULL; size_t b_len = 0;
    get_buf_data_len(buf, &b, &b_len);
    int padding = (int)padding_num;

    EVP_PKEY *pkey = parse_any_key(key_pem);
    if (!pkey) return evp_fail("privateDecrypt: invalid key");
    EVP_PKEY_CTX *pctx = EVP_PKEY_CTX_new(pkey, NULL);
    if (!pctx) { EVP_PKEY_free(pkey); return evp_fail("privateDecrypt: ctx failed"); }
    if (EVP_PKEY_decrypt_init(pctx) <= 0) { EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("privateDecrypt: init failed"); }
    if (padding > 0) EVP_PKEY_CTX_set_rsa_padding(pctx, padding);
    size_t out_len = 0;
    if (EVP_PKEY_decrypt(pctx, NULL, &out_len, b ? b : (const unsigned char *)"", b_len) <= 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("privateDecrypt: size failed");
    }
    if (scriptgo_buffer_alloc((double)out_len, NULL, 0, 0, 0, out_buf) != 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("privateDecrypt: alloc failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_buf;
    if (EVP_PKEY_decrypt(pctx, bv->data, &out_len, b ? b : (const unsigned char *)"", b_len) <= 0) {
        EVP_PKEY_CTX_free(pctx); EVP_PKEY_free(pkey); return evp_fail("privateDecrypt: decrypt failed");
    }
    bv->length = (int64_t)out_len;
    EVP_PKEY_CTX_free(pctx);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_private_encrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    return scriptgo_crypto_public_encrypt(key_pem, buf, padding_num, out_buf);
}

int scriptgo_crypto_public_decrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    return scriptgo_crypto_private_decrypt(key_pem, buf, padding_num, out_buf);
}

static char *evp_json_escape(const char *src, size_t src_len) {
    if (!src) return strdup("");
    char *dest = (char *)malloc(src_len * 2 + 1);
    char *d = dest;
    for (size_t i = 0; i < src_len; i++) {
        if (src[i] == '\n') { *d++ = '\\'; *d++ = 'n'; }
        else if (src[i] == '\r') { *d++ = '\\'; *d++ = 'r'; }
        else if (src[i] == '\"') { *d++ = '\\'; *d++ = '\"'; }
        else if (src[i] == '\\') { *d++ = '\\'; *d++ = '\\'; }
        else { *d++ = src[i]; }
    }
    *d = '\0';
    return dest;
}

int scriptgo_crypto_generate_key_pair_sync(const char *type, double modulus_length_num, const char *named_curve, char **out_json) {
    if (!type || !out_json) return evp_fail("generateKeyPairSync: invalid arguments");
    EVP_PKEY *pkey = NULL;
    int modulus_length = (int)modulus_length_num;
    if (strcasecmp(type, "rsa") == 0) {
        size_t bits = modulus_length > 0 ? (size_t)modulus_length : 2048;
        pkey = EVP_PKEY_Q_keygen(NULL, NULL, "RSA", bits);
    } else if (strcasecmp(type, "ec") == 0) {
        const char *curve = (named_curve && strlen(named_curve) > 0) ? named_curve : "prime256v1";
        pkey = EVP_PKEY_Q_keygen(NULL, NULL, "EC", curve);
    } else if (strcasecmp(type, "ed25519") == 0) {
        pkey = EVP_PKEY_Q_keygen(NULL, NULL, "ED25519");
    }
    if (!pkey) return evp_fail("generateKeyPairSync: keygen failed");

    BIO *pub_bio = BIO_new(BIO_s_mem());
    PEM_write_bio_PUBKEY(pub_bio, pkey);
    char *pub_data = NULL;
    long pub_len = BIO_get_mem_data(pub_bio, &pub_data);

    BIO *priv_bio = BIO_new(BIO_s_mem());
    PEM_write_bio_PKCS8PrivateKey(priv_bio, pkey, NULL, NULL, 0, NULL, NULL);
    char *priv_data = NULL;
    long priv_len = BIO_get_mem_data(priv_bio, &priv_data);

    char *esc_pub = evp_json_escape(pub_data, (size_t)pub_len);
    char *esc_priv = evp_json_escape(priv_data, (size_t)priv_len);
    size_t total = strlen(esc_pub) + strlen(esc_priv) + 64;
    *out_json = (char *)malloc(total);
    snprintf(*out_json, total, "{\"publicKey\":\"%s\",\"privateKey\":\"%s\"}", esc_pub, esc_priv);

    free(esc_pub);
    free(esc_priv);
    BIO_free(pub_bio);
    BIO_free(priv_bio);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_key_details(const char *key_pem, char **out_json) {
    if (!key_pem || !out_json) return evp_fail("keyDetails: invalid arguments");
    BIO *bio = BIO_new_mem_buf(key_pem, -1);
    int is_priv = 1;
    EVP_PKEY *pkey = PEM_read_bio_PrivateKey(bio, NULL, NULL, NULL);
    if (!pkey) {
        BIO_reset(bio);
        is_priv = 0;
        pkey = PEM_read_bio_PUBKEY(bio, NULL, NULL, NULL);
    }
    BIO_free(bio);
    if (!pkey) {
        *out_json = strdup("{\"type\":\"secret\"}");
        return 0;
    }
    const char *ktype = is_priv ? "private" : "public";
    const char *asym_type = "rsa";
    int base_id = EVP_PKEY_get_base_id(pkey);
    if (base_id == EVP_PKEY_EC) asym_type = "ec";
    else if (base_id == EVP_PKEY_ED25519) asym_type = "ed25519";
    int bits = EVP_PKEY_get_bits(pkey);
    char buf[512];
    snprintf(buf, sizeof(buf), "{\"type\":\"%s\",\"asymmetricKeyType\":\"%s\",\"modulusLength\":%d}", ktype, asym_type, bits);
    *out_json = strdup(buf);
    EVP_PKEY_free(pkey);
    return 0;
}

typedef struct {
    EVP_PKEY *dh_key;
    BIGNUM *p;
    BIGNUM *g;
    BIGNUM *pub;
    BIGNUM *priv;
} scriptgo_dh_ctx;

int scriptgo_crypto_dh_create(void *prime_buf, void *gen_buf, void **out_dh) {
    if (!prime_buf || !out_dh) return evp_fail("createDiffieHellman: null prime");
    const unsigned char *prime = NULL; size_t prime_len = 0;
    const unsigned char *gen = NULL; size_t gen_len = 0;
    get_buf_data_len(prime_buf, &prime, &prime_len);
    get_buf_data_len(gen_buf, &gen, &gen_len);

    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)calloc(1, sizeof(scriptgo_dh_ctx));
    ctx->p = BN_bin2bn(prime, (int)prime_len, NULL);
    if (gen && gen_len > 0) {
        ctx->g = BN_bin2bn(gen, (int)gen_len, NULL);
    } else {
        ctx->g = BN_new();
        BN_set_word(ctx->g, 2);
    }
    *out_dh = ctx;
    return 0;
}

int scriptgo_crypto_dh_create_group(const char *group_name, void **out_dh) {
    if (!group_name || !out_dh) return evp_fail("createDiffieHellmanGroup: null name");
    // Standard RFC 3526 MODP group 14 (2048-bit prime)
    const char *hex_p = "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF";
    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)calloc(1, sizeof(scriptgo_dh_ctx));
    BN_hex2bn(&ctx->p, hex_p);
    ctx->g = BN_new();
    BN_set_word(ctx->g, 2);
    *out_dh = ctx;
    return 0;
}

static int dh_generate(scriptgo_dh_ctx *ctx) {
    if (ctx->dh_key) return 0;
    OSSL_PARAM_BLD *bld = OSSL_PARAM_BLD_new();
    OSSL_PARAM_BLD_push_BN(bld, "p", ctx->p);
    OSSL_PARAM_BLD_push_BN(bld, "g", ctx->g);
    OSSL_PARAM *params = OSSL_PARAM_BLD_to_param(bld);

    EVP_PKEY_CTX *pctx = EVP_PKEY_CTX_new_from_name(NULL, "DH", NULL);
    if (!pctx) { OSSL_PARAM_free(params); OSSL_PARAM_BLD_free(bld); return -1; }
    EVP_PKEY_fromdata_init(pctx);
    EVP_PKEY *param_key = NULL;
    EVP_PKEY_fromdata(pctx, &param_key, EVP_PKEY_KEY_PARAMETERS, params);
    EVP_PKEY_CTX_free(pctx);
    OSSL_PARAM_free(params);
    OSSL_PARAM_BLD_free(bld);

    if (!param_key) return -1;
    EVP_PKEY_CTX *kctx = EVP_PKEY_CTX_new(param_key, NULL);
    EVP_PKEY_keygen_init(kctx);
    EVP_PKEY_keygen(kctx, &ctx->dh_key);
    EVP_PKEY_CTX_free(kctx);
    EVP_PKEY_free(param_key);

    if (ctx->dh_key) {
        EVP_PKEY_get_bn_param(ctx->dh_key, "pub", &ctx->pub);
        EVP_PKEY_get_bn_param(ctx->dh_key, "priv", &ctx->priv);
        return 0;
    }
    return -1;
}

int scriptgo_crypto_dh_generate_keys(void *dh, void **out_pub) {
    if (!dh || !out_pub) return evp_fail("dh.generateKeys: null argument");
    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)dh;
    if (dh_generate(ctx) != 0) return evp_fail("dh.generateKeys: keygen failed");
    int num_bytes = BN_num_bytes(ctx->pub);
    if (scriptgo_buffer_alloc((double)num_bytes, NULL, 0, 0, 0, out_pub) != 0) return evp_fail("dh.generateKeys: alloc failed");
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_pub;
    BN_bn2bin(ctx->pub, bv->data);
    bv->length = num_bytes;
    return 0;
}

int scriptgo_crypto_dh_compute_secret(void *dh, void *other_pub_buf, void **out_secret) {
    if (!dh || !other_pub_buf || !out_secret) return evp_fail("dh.computeSecret: invalid arguments");
    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)dh;
    if (dh_generate(ctx) != 0) return evp_fail("dh.computeSecret: keygen failed");

    const unsigned char *other_pub = NULL; size_t other_pub_len = 0;
    get_buf_data_len(other_pub_buf, &other_pub, &other_pub_len);

    BIGNUM *other_bn = BN_bin2bn(other_pub, (int)other_pub_len, NULL);
    OSSL_PARAM_BLD *bld = OSSL_PARAM_BLD_new();
    OSSL_PARAM_BLD_push_BN(bld, "p", ctx->p);
    OSSL_PARAM_BLD_push_BN(bld, "g", ctx->g);
    OSSL_PARAM_BLD_push_BN(bld, "pub", other_bn);
    OSSL_PARAM *params = OSSL_PARAM_BLD_to_param(bld);

    EVP_PKEY_CTX *pctx = EVP_PKEY_CTX_new_from_name(NULL, "DH", NULL);
    EVP_PKEY_fromdata_init(pctx);
    EVP_PKEY *peer_key = NULL;
    EVP_PKEY_fromdata(pctx, &peer_key, EVP_PKEY_PUBLIC_KEY, params);
    EVP_PKEY_CTX_free(pctx);
    OSSL_PARAM_free(params);
    OSSL_PARAM_BLD_free(bld);
    BN_free(other_bn);

    if (!peer_key) return evp_fail("dh.computeSecret: peer key creation failed");
    EVP_PKEY_CTX *dctx = EVP_PKEY_CTX_new(ctx->dh_key, NULL);
    EVP_PKEY_derive_init(dctx);
    EVP_PKEY_derive_set_peer(dctx, peer_key);
    size_t sec_len = 0;
    EVP_PKEY_derive(dctx, NULL, &sec_len);
    if (scriptgo_buffer_alloc((double)sec_len, NULL, 0, 0, 0, out_secret) != 0) {
        EVP_PKEY_CTX_free(dctx); EVP_PKEY_free(peer_key); return evp_fail("dh.computeSecret: alloc failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_secret;
    EVP_PKEY_derive(dctx, bv->data, &sec_len);
    bv->length = (int64_t)sec_len;
    EVP_PKEY_CTX_free(dctx);
    EVP_PKEY_free(peer_key);
    return 0;
}

int scriptgo_crypto_dh_get_key(void *dh, double which_num, void **out_key) {
    if (!dh || !out_key) return evp_fail("dh.getKey: null");
    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)dh;
    int which = (int)which_num;
    BIGNUM *bn = NULL;
    if (which == 0) bn = ctx->p;
    else if (which == 1) bn = ctx->g;
    else if (which == 2) { dh_generate(ctx); bn = ctx->pub; }
    else if (which == 3) { dh_generate(ctx); bn = ctx->priv; }
    if (!bn) return evp_fail("dh.getKey: key not set");
    int num = BN_num_bytes(bn);
    if (scriptgo_buffer_alloc((double)num, NULL, 0, 0, 0, out_key) != 0) return evp_fail("dh.getKey: alloc failed");
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_key;
    BN_bn2bin(bn, bv->data);
    bv->length = num;
    return 0;
}

int scriptgo_crypto_dh_set_key(void *dh, double which_num, void *key_buf) {
    if (!dh || !key_buf) return evp_fail("dh.setKey: null");
    scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)dh;
    int which = (int)which_num;
    const unsigned char *key = NULL; size_t key_len = 0;
    get_buf_data_len(key_buf, &key, &key_len);
    BIGNUM *bn = BN_bin2bn(key, (int)key_len, NULL);
    if (which == 2) { if (ctx->pub) BN_free(ctx->pub); ctx->pub = bn; }
    else if (which == 3) { if (ctx->priv) BN_free(ctx->priv); ctx->priv = bn; }
    return 0;
}

int scriptgo_crypto_dh_destroy(void *dh) {
    if (dh) {
        scriptgo_dh_ctx *ctx = (scriptgo_dh_ctx *)dh;
        if (ctx->dh_key) EVP_PKEY_free(ctx->dh_key);
        if (ctx->p) BN_free(ctx->p);
        if (ctx->g) BN_free(ctx->g);
        if (ctx->pub) BN_free(ctx->pub);
        if (ctx->priv) BN_free(ctx->priv);
        free(ctx);
    }
    return 0;
}

typedef struct {
    char *curve;
    EVP_PKEY *ec_key;
} scriptgo_ecdh_ctx;

int scriptgo_crypto_ecdh_create(const char *curve_name, void **out_ecdh) {
    if (!curve_name || !out_ecdh) return evp_fail("createECDH: null curve");
    scriptgo_ecdh_ctx *ctx = (scriptgo_ecdh_ctx *)calloc(1, sizeof(scriptgo_ecdh_ctx));
    ctx->curve = strdup(curve_name);
    *out_ecdh = ctx;
    return 0;
}

int scriptgo_crypto_ecdh_generate_keys(void *ecdh, void **out_pub) {
    if (!ecdh || !out_pub) return evp_fail("ecdh.generateKeys: null argument");
    scriptgo_ecdh_ctx *ctx = (scriptgo_ecdh_ctx *)ecdh;
    if (!ctx->ec_key) {
        ctx->ec_key = EVP_PKEY_Q_keygen(NULL, NULL, "EC", ctx->curve ? ctx->curve : "prime256v1");
        if (!ctx->ec_key) return evp_fail("ecdh.generateKeys: keygen failed");
    }
    size_t pub_len = 0;
    EVP_PKEY_get_octet_string_param(ctx->ec_key, "pub", NULL, 0, &pub_len);
    if (pub_len == 0) pub_len = 65;
    if (scriptgo_buffer_alloc((double)pub_len, NULL, 0, 0, 0, out_pub) != 0) return evp_fail("ecdh.generateKeys: alloc failed");
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_pub;
    EVP_PKEY_get_octet_string_param(ctx->ec_key, "pub", bv->data, pub_len, &pub_len);
    bv->length = (int64_t)pub_len;
    return 0;
}

int scriptgo_crypto_ecdh_compute_secret(void *ecdh, void *other_pub_buf, void **out_secret) {
    if (!ecdh || !other_pub_buf || !out_secret) return evp_fail("ecdh.computeSecret: invalid arguments");
    scriptgo_ecdh_ctx *ctx = (scriptgo_ecdh_ctx *)ecdh;
    if (!ctx->ec_key) {
        ctx->ec_key = EVP_PKEY_Q_keygen(NULL, NULL, "EC", ctx->curve ? ctx->curve : "prime256v1");
        if (!ctx->ec_key) return evp_fail("ecdh.computeSecret: keygen failed");
    }
    const unsigned char *other_pub = NULL; size_t other_pub_len = 0;
    get_buf_data_len(other_pub_buf, &other_pub, &other_pub_len);

    OSSL_PARAM_BLD *bld = OSSL_PARAM_BLD_new();
    OSSL_PARAM_BLD_push_utf8_string(bld, "group", ctx->curve ? ctx->curve : "prime256v1", 0);
    OSSL_PARAM_BLD_push_octet_string(bld, "pub", other_pub, other_pub_len);
    OSSL_PARAM *params = OSSL_PARAM_BLD_to_param(bld);

    EVP_PKEY_CTX *pctx = EVP_PKEY_CTX_new_from_name(NULL, "EC", NULL);
    EVP_PKEY_fromdata_init(pctx);
    EVP_PKEY *peer_key = NULL;
    EVP_PKEY_fromdata(pctx, &peer_key, EVP_PKEY_PUBLIC_KEY, params);
    EVP_PKEY_CTX_free(pctx);
    OSSL_PARAM_free(params);
    OSSL_PARAM_BLD_free(bld);

    if (!peer_key) return evp_fail("ecdh.computeSecret: invalid peer key");
    EVP_PKEY_CTX *dctx = EVP_PKEY_CTX_new(ctx->ec_key, NULL);
    EVP_PKEY_derive_init(dctx);
    EVP_PKEY_derive_set_peer(dctx, peer_key);
    size_t sec_len = 0;
    EVP_PKEY_derive(dctx, NULL, &sec_len);
    if (scriptgo_buffer_alloc((double)sec_len, NULL, 0, 0, 0, out_secret) != 0) {
        EVP_PKEY_CTX_free(dctx); EVP_PKEY_free(peer_key); return evp_fail("ecdh.computeSecret: alloc failed");
    }
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_secret;
    EVP_PKEY_derive(dctx, bv->data, &sec_len);
    bv->length = (int64_t)sec_len;
    EVP_PKEY_CTX_free(dctx);
    EVP_PKEY_free(peer_key);
    return 0;
}

int scriptgo_crypto_ecdh_get_key(void *ecdh, double which_num, void **out_key) {
    if (!ecdh || !out_key) return evp_fail("ecdh.getKey: null");
    scriptgo_ecdh_ctx *ctx = (scriptgo_ecdh_ctx *)ecdh;
    int which = (int)which_num;
    if (!ctx->ec_key) {
        ctx->ec_key = EVP_PKEY_Q_keygen(NULL, NULL, "EC", ctx->curve ? ctx->curve : "prime256v1");
    }
    const char *param_name = (which == 0) ? "pub" : "priv";
    size_t sz = 0;
    if (which == 0) {
        EVP_PKEY_get_octet_string_param(ctx->ec_key, param_name, NULL, 0, &sz);
    } else {
        BIGNUM *bn = NULL;
        EVP_PKEY_get_bn_param(ctx->ec_key, param_name, &bn);
        if (bn) { sz = (size_t)BN_num_bytes(bn); BN_free(bn); }
    }
    if (sz == 0) sz = 32;
    if (scriptgo_buffer_alloc((double)sz, NULL, 0, 0, 0, out_key) != 0) return evp_fail("ecdh.getKey: alloc failed");
    scriptgo_crypto_buffer_view_evp *bv = (scriptgo_crypto_buffer_view_evp *)*out_key;
    if (which == 0) {
        EVP_PKEY_get_octet_string_param(ctx->ec_key, param_name, bv->data, sz, &sz);
    } else {
        BIGNUM *bn = NULL;
        EVP_PKEY_get_bn_param(ctx->ec_key, param_name, &bn);
        if (bn) { BN_bn2bin(bn, bv->data); BN_free(bn); }
    }
    bv->length = (int64_t)sz;
    return 0;
}

int scriptgo_crypto_ecdh_set_key(void *ecdh, double which_num, void *key_buf) {
    (void)ecdh; (void)which_num; (void)key_buf;
    return 0;
}

int scriptgo_crypto_ecdh_destroy(void *ecdh) {
    if (ecdh) {
        scriptgo_ecdh_ctx *ctx = (scriptgo_ecdh_ctx *)ecdh;
        if (ctx->curve) free(ctx->curve);
        if (ctx->ec_key) EVP_PKEY_free(ctx->ec_key);
        free(ctx);
    }
    return 0;
}

int scriptgo_crypto_spkac_verify(const char *spkac, double *out_valid) {
    if (!spkac || !out_valid) return evp_fail("verifySpkac: invalid arguments");
    *out_valid = 0.0;
    NETSCAPE_SPKI *spki = NETSCAPE_SPKI_b64_decode(spkac, -1);
    if (!spki) return 0;
    EVP_PKEY *pkey = NETSCAPE_SPKI_get_pubkey(spki);
    if (pkey) {
        *out_valid = (NETSCAPE_SPKI_verify(spki, pkey) == 1) ? 1.0 : 0.0;
        EVP_PKEY_free(pkey);
    }
    NETSCAPE_SPKI_free(spki);
    return 0;
}

int scriptgo_crypto_spkac_export_challenge(const char *spkac, char **out_challenge) {
    if (!spkac || !out_challenge) return evp_fail("exportChallenge: invalid arguments");
    *out_challenge = strdup("");
    NETSCAPE_SPKI *spki = NETSCAPE_SPKI_b64_decode(spkac, -1);
    if (!spki) return 0;
    if (spki->spkac && spki->spkac->challenge && spki->spkac->challenge->data) {
        free(*out_challenge);
        *out_challenge = strdup((const char *)spki->spkac->challenge->data);
    }
    NETSCAPE_SPKI_free(spki);
    return 0;
}

int scriptgo_crypto_spkac_export_public_key(const char *spkac, char **out_pub_pem) {
    if (!spkac || !out_pub_pem) return evp_fail("exportPublicKey: invalid arguments");
    *out_pub_pem = strdup("");
    NETSCAPE_SPKI *spki = NETSCAPE_SPKI_b64_decode(spkac, -1);
    if (!spki) return 0;
    EVP_PKEY *pkey = NETSCAPE_SPKI_get_pubkey(spki);
    if (pkey) {
        BIO *bio = BIO_new(BIO_s_mem());
        PEM_write_bio_PUBKEY(bio, pkey);
        char *data = NULL;
        long len = BIO_get_mem_data(bio, &data);
        if (len > 0 && data) {
            free(*out_pub_pem);
            *out_pub_pem = (char *)malloc((size_t)len + 1);
            memcpy(*out_pub_pem, data, (size_t)len);
            (*out_pub_pem)[len] = '\0';
        }
        BIO_free(bio);
        EVP_PKEY_free(pkey);
    }
    NETSCAPE_SPKI_free(spki);
    return 0;
}

int scriptgo_crypto_x509_check_private_key(const char *cert_pem, const char *key_pem, double *out_valid) {
    if (!cert_pem || !key_pem || !out_valid) return evp_fail("checkPrivateKey: invalid arguments");
    *out_valid = 0.0;
    BIO *cbio = BIO_new_mem_buf(cert_pem, -1);
    X509 *x = PEM_read_bio_X509(cbio, NULL, NULL, NULL);
    BIO_free(cbio);
    if (!x) return 0;
    BIO *kbio = BIO_new_mem_buf(key_pem, -1);
    EVP_PKEY *pkey = PEM_read_bio_PrivateKey(kbio, NULL, NULL, NULL);
    BIO_free(kbio);
    if (!pkey) { X509_free(x); return 0; }
    *out_valid = (X509_check_private_key(x, pkey) == 1) ? 1.0 : 0.0;
    X509_free(x);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_x509_verify(const char *cert_pem, const char *pubkey_pem, double *out_valid) {
    if (!cert_pem || !pubkey_pem || !out_valid) return evp_fail("verify: invalid arguments");
    *out_valid = 0.0;
    BIO *cbio = BIO_new_mem_buf(cert_pem, -1);
    X509 *x = PEM_read_bio_X509(cbio, NULL, NULL, NULL);
    BIO_free(cbio);
    if (!x) return 0;
    EVP_PKEY *pkey = parse_any_key(pubkey_pem);
    if (!pkey) { X509_free(x); return 0; }
    *out_valid = (X509_verify(x, pkey) == 1) ? 1.0 : 0.0;
    X509_free(x);
    EVP_PKEY_free(pkey);
    return 0;
}

int scriptgo_crypto_get_fips(double *out_fips) {
    if (!out_fips) return evp_fail("getFips: null pointer");
    *out_fips = (double)EVP_default_properties_is_fips_enabled(NULL);
    return 0;
}

int scriptgo_crypto_set_fips(double enable_num) {
    int enable = (enable_num != 0.0) ? 1 : 0;
    EVP_default_properties_enable_fips(NULL, enable);
    return 0;
}

int scriptgo_crypto_secure_heap_used(char **out_json) {
    if (!out_json) return evp_fail("secureHeapUsed: null pointer");
    size_t used = 0;
    if (CRYPTO_secure_malloc_initialized()) {
        used = CRYPTO_secure_used();
    }
    char buf[128];
    snprintf(buf, sizeof(buf), "{\"total\":0,\"min\":0,\"used\":%zu,\"util\":0}", used);
    *out_json = strdup(buf);
    return 0;
}

int scriptgo_crypto_set_engine(const char *engine, double flags_num) {
    (void)engine; (void)flags_num;
    return 0;
}

#else

int scriptgo_crypto_cipher_create(const char *algo, void *key_buf, void *iv_buf, double is_encrypt_num, void **out_ctx) {
    (void)algo; (void)key_buf; (void)iv_buf; (void)is_encrypt_num; (void)out_ctx;
    return -1;
}
int scriptgo_crypto_cipher_update(void *h, void *in_buf, void **out_buf) {
    (void)h; (void)in_buf; (void)out_buf; return -1;
}
int scriptgo_crypto_cipher_final(void *h, void **out_buf) { (void)h; (void)out_buf; return -1; }
int scriptgo_crypto_cipher_set_aad(void *h, void *aad_buf) { (void)h; (void)aad_buf; return -1; }
int scriptgo_crypto_cipher_get_tag(void *h, void **out_buf) { (void)h; (void)out_buf; return -1; }
int scriptgo_crypto_cipher_set_tag(void *h, void *tag_buf) { (void)h; (void)tag_buf; return -1; }
int scriptgo_crypto_cipher_set_auto_padding(void *h, double auto_padding_num) { (void)h; (void)auto_padding_num; return -1; }
int scriptgo_crypto_cipher_destroy(void *h) { (void)h; return 0; }
int scriptgo_crypto_get_cipher_info(const char *name_or_nid, char **out_json) { (void)name_or_nid; (void)out_json; return -1; }
int scriptgo_crypto_sign(const char *algo, void *data_buf, const char *key_pem, void **out_buf) {
    (void)algo; (void)data_buf; (void)key_pem; (void)out_buf; return -1;
}
int scriptgo_crypto_verify(const char *algo, void *data_buf, const char *key_pem, void *sig_buf, double *out_valid) {
    (void)algo; (void)data_buf; (void)key_pem; (void)sig_buf; (void)out_valid; return -1;
}
int scriptgo_crypto_public_encrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    (void)key_pem; (void)buf; (void)padding_num; (void)out_buf; return -1;
}
int scriptgo_crypto_private_decrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    (void)key_pem; (void)buf; (void)padding_num; (void)out_buf; return -1;
}
int scriptgo_crypto_private_encrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    (void)key_pem; (void)buf; (void)padding_num; (void)out_buf; return -1;
}
int scriptgo_crypto_public_decrypt(const char *key_pem, void *buf, double padding_num, void **out_buf) {
    (void)key_pem; (void)buf; (void)padding_num; (void)out_buf; return -1;
}
int scriptgo_crypto_generate_key_pair_sync(const char *type, double modulus_length_num, const char *named_curve, char **out_json) {
    (void)type; (void)modulus_length_num; (void)named_curve; (void)out_json; return -1;
}
int scriptgo_crypto_key_details(const char *key_pem, char **out_json) { (void)key_pem; (void)out_json; return -1; }
int scriptgo_crypto_dh_create(void *prime_buf, void *gen_buf, void **out_dh) {
    (void)prime_buf; (void)gen_buf; (void)out_dh; return -1;
}
int scriptgo_crypto_dh_create_group(const char *group_name, void **out_dh) { (void)group_name; (void)out_dh; return -1; }
int scriptgo_crypto_dh_generate_keys(void *dh, void **out_pub) { (void)dh; (void)out_pub; return -1; }
int scriptgo_crypto_dh_compute_secret(void *dh, void *other_pub_buf, void **out_secret) {
    (void)dh; (void)other_pub_buf; (void)out_secret; return -1;
}
int scriptgo_crypto_dh_get_key(void *dh, double which_num, void **out_key) { (void)dh; (void)which_num; (void)out_key; return -1; }
int scriptgo_crypto_dh_set_key(void *dh, double which_num, void *key_buf) { (void)dh; (void)which_num; (void)key_buf; return -1; }
int scriptgo_crypto_dh_destroy(void *dh) { (void)dh; return 0; }
int scriptgo_crypto_ecdh_create(const char *curve_name, void **out_ecdh) { (void)curve_name; (void)out_ecdh; return -1; }
int scriptgo_crypto_ecdh_generate_keys(void *ecdh, void **out_pub) { (void)ecdh; (void)out_pub; return -1; }
int scriptgo_crypto_ecdh_compute_secret(void *ecdh, void *other_pub_buf, void **out_secret) {
    (void)ecdh; (void)other_pub_buf; (void)out_secret; return -1;
}
int scriptgo_crypto_ecdh_get_key(void *ecdh, double which_num, void **out_key) { (void)ecdh; (void)which_num; (void)out_key; return -1; }
int scriptgo_crypto_ecdh_set_key(void *ecdh, double which_num, void *key_buf) { (void)ecdh; (void)which_num; (void)key_buf; return -1; }
int scriptgo_crypto_ecdh_destroy(void *ecdh) { (void)ecdh; return 0; }
int scriptgo_crypto_spkac_verify(const char *spkac, double *out_valid) { (void)spkac; (void)out_valid; return -1; }
int scriptgo_crypto_spkac_export_challenge(const char *spkac, char **out_challenge) { (void)spkac; (void)out_challenge; return -1; }
int scriptgo_crypto_spkac_export_public_key(const char *spkac, char **out_pub_pem) { (void)spkac; (void)out_pub_pem; return -1; }
int scriptgo_crypto_x509_check_private_key(const char *cert_pem, const char *key_pem, double *out_valid) { (void)cert_pem; (void)key_pem; (void)out_valid; return -1; }
int scriptgo_crypto_x509_verify(const char *cert_pem, const char *pubkey_pem, double *out_valid) { (void)cert_pem; (void)pubkey_pem; (void)out_valid; return -1; }
int scriptgo_crypto_get_fips(double *out_fips) { (void)out_fips; return -1; }
int scriptgo_crypto_set_fips(double enable_num) { (void)enable_num; return -1; }
int scriptgo_crypto_secure_heap_used(char **out_json) { (void)out_json; return -1; }
int scriptgo_crypto_set_engine(const char *engine, double flags_num) { (void)engine; (void)flags_num; return -1; }

#endif
