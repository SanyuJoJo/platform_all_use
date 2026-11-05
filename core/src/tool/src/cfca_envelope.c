/*
 * cfca_envelope.c — CFCA 数字信封 + CSR 生成/解析工具
 *
 * 5 段 header（80 字符，每字段 16 字符十进制）：
 *   [0,16)  version = 1
 *   [16,32) type    = 1
 *   [32,48) 0
 *   [48,64) 0
 *   [64,80) body_len = strlen("seg1,seg2,seg3,seg4,seg5,")
 *
 * EnvelopedKey 结构（与成功样本 enc_base64.txt 完全一致）：
 *   SEQUENCE { INTEGER 2, OCTET STRING <192 字节 raw C1C3C2> }
 *     192 = X(32) || Y(32) || C3(32) || C2(96)
 *
 * SM2 明文 96 字节布局：
 *   [0, 16)   SM4 密钥（随机会话密钥）
 *   [16, 48)  SM4-ECB(加密私钥 D, 32 字节)
 *   [48, 96)  0 填充
 *
 * ★ 临时密钥模式：
 *   - 默认（无 --tmp-key）：gen-csr 每次随机生成一对临时密钥，
 *     并强制输出 --out-tmp-key / --out-tmp-pub-der 供调用方保管；
 *   - 指定 --tmp-key：使用已有的临时私钥（平台内置固定），
 *     仅将其公钥写入 CSR attribute，不再生成新密钥对，
 *     也不再强制要求 --out-tmp-key / --out-tmp-pub-der。
 */
#define OPENSSL_SUPPRESS_DEPRECATED 1

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <stdint.h>

#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/x509.h>
#include <openssl/x509v3.h>
#include <openssl/asn1t.h>
#include <openssl/objects.h>
#include <openssl/rand.h>
#include <openssl/ec.h>
#include <openssl/bn.h>
#include <openssl/err.h>

#define OID_TMP_PUB  "1.2.840.113549.1.9.63"
#define SM4_KEY_LEN  16
#define SM4_BLOCK    16
#define SM2_D_LEN    32
#define ENC_PUB_LEN  65
#define CFCA_SEG_BYTES 48
#define CFCA_KEYBYTES_LEN 136
#define CFCA_PLAIN_LEN 96
#define CFCA_SYM_PRI_LEN 192

static void usage(const char *p) {
    fprintf(stderr,
        "Usage:\n"
        "  %s gen-csr        --subject \"/C=CN/O=Org/CN=test\" \\\n"
        "                    --out-csr <csr.pem> --out-sign-key <sign.key> \\\n"
        "                    --tmp-key <builtin_tmp.key> \\\n"
        "                    [--out-tmp-pub-der <tmp.pub.der>]\n"
        "                    # 若未提供 --tmp-key，则必须提供 --out-tmp-key 与 --out-tmp-pub-der\n"
        "  %s gen-plain-csr  --subject \"/C=CN/O=Org/CN=test\" \\\n"
        "                    --out-csr <csr.pem> --out-key <enc.key>\n"
        "  %s csr2pub        --in <csr.pem> --out <tmp.pub.der>\n"
        "  %s csr2signpub    --in <csr.pem> --out <sign.pub.pem>\n"
        "  %s csr2subj       --in <csr.pem> --out <subject.txt>\n"
        "  %s sign-cert      --csr <csr.pem> --ca-cert <ca.crt> --ca-key <ca.key> \\\n"
        "                    --out-cert <cert.pem> [--days 365] [--is-enc 0|1] \\\n"
        "                    [--subject \"/C=CN/O=Org/CN=test\"] [--san ...]\n"
        "  %s seal           --tmp-pub <pub.pem> --enc-key <key.pem> --out <env> \\\n"
        "                    [--format cfca|der]\n"
        "  %s unseal         --tmp-key <key.pem> --in <env> --out <key.pem>\n",
        p, p, p, p, p, p, p, p);
}

static EVP_PKEY *load_pub(const char *path) {
    FILE *f = fopen(path, "rb");
    if (!f) { perror(path); return NULL; }
    EVP_PKEY *k = PEM_read_PUBKEY(f, NULL, NULL, NULL);
    fclose(f);
    return k;
}

static EVP_PKEY *load_priv(const char *path) {
    FILE *f = fopen(path, "rb");
    if (!f) { perror(path); return NULL; }
    EVP_PKEY *k = PEM_read_PrivateKey(f, NULL, NULL, NULL);
    fclose(f);
    return k;
}

static int write_priv_pem(const char *path, EVP_PKEY *k) {
    FILE *f = fopen(path, "wb");
    if (!f) { perror(path); return -1; }
    int rc = PEM_write_PrivateKey(f, k, NULL, NULL, 0, NULL, NULL);
    fclose(f);
    return rc == 1 ? 0 : -1;
}

static int write_csr_pem(const char *path, X509_REQ *r) {
    FILE *f = fopen(path, "wb");
    if (!f) { perror(path); return -1; }
    int rc = PEM_write_X509_REQ(f, r);
    fclose(f);
    return rc == 1 ? 0 : -1;
}

static int write_bytes(const char *path, const unsigned char *buf, size_t len) {
    FILE *f = fopen(path, "wb");
    if (!f) { perror(path); return -1; }
    int ok = (len == 0) || (fwrite(buf, 1, len, f) == len);
    fclose(f);
    return ok ? 0 : -1;
}

static char *b64_encode_alloc(const unsigned char *in, size_t inlen) {
    size_t outlen = 4 * ((inlen + 2) / 3) + 1;
    char *out = malloc(outlen);
    if (!out) return NULL;
    EVP_EncodeBlock((unsigned char *)out, in, (int)inlen);
    return out;
}

static unsigned char *b64_decode_alloc(const char *in, size_t inlen, size_t *outlen) {
    if (inlen == 0 || inlen % 4 != 0) return NULL;
    unsigned char *out = OPENSSL_malloc(inlen / 4 * 3 + 4);
    if (!out) return NULL;
    int n = EVP_DecodeBlock(out, (const unsigned char *)in, (int)inlen);
    if (n < 0) { OPENSSL_free(out); return NULL; }
    while (inlen > 0 && in[inlen - 1] == '=') { n--; inlen--; }
    *outlen = (size_t)n;
    return out;
}

static EVP_PKEY *gen_sm2_key(void) {
    EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new_id(EVP_PKEY_SM2, NULL);
    if (!ctx) return NULL;
    EVP_PKEY *k = NULL;
    if (EVP_PKEY_keygen_init(ctx) <= 0 ||
        EVP_PKEY_keygen(ctx, &k) <= 0) {
        EVP_PKEY_CTX_free(ctx);
        return NULL;
    }
    EVP_PKEY_CTX_free(ctx);

    EC_KEY *ec = (EC_KEY *)EVP_PKEY_get0_EC_KEY(k);
    if (ec) EC_KEY_set_asn1_flag(ec, OPENSSL_EC_NAMED_CURVE);

    return k;
}

static int get_raw_d_32(EVP_PKEY *pkey, unsigned char out[SM2_D_LEN]) {
    BIGNUM *d = NULL;
    if (EVP_PKEY_get_bn_param(pkey, "priv", &d) != 1) {
        const EC_KEY *ec = EVP_PKEY_get0_EC_KEY(pkey);
        if (!ec) return -1;
        d = BN_dup(EC_KEY_get0_private_key(ec));
        if (!d) return -1;
    }
    if (BN_num_bytes(d) > SM2_D_LEN) { BN_free(d); return -1; }
    memset(out, 0, SM2_D_LEN);
    BN_bn2binpad(d, out, SM2_D_LEN);
    BN_free(d);
    return 0;
}

static EVP_PKEY *pkey_from_raw_d(const unsigned char d[SM2_D_LEN]) {
    EC_KEY *ec = EC_KEY_new_by_curve_name(NID_sm2);
    if (!ec) return NULL;
    EC_KEY_set_asn1_flag(ec, OPENSSL_EC_NAMED_CURVE);
    BIGNUM *bn = BN_bin2bn(d, SM2_D_LEN, NULL);
    if (!bn || EC_KEY_set_private_key(ec, bn) != 1) {
        BN_free(bn); EC_KEY_free(ec); return NULL;
    }
    const EC_GROUP *g = EC_KEY_get0_group(ec);
    EC_POINT *pub = EC_POINT_new(g);
    if (!pub || EC_POINT_mul(g, pub, bn, NULL, NULL, NULL) != 1) {
        BN_free(bn); EC_POINT_free(pub); EC_KEY_free(ec); return NULL;
    }
    EC_KEY_set_public_key(ec, pub);
    EC_POINT_free(pub);
    BN_free(bn);
    EVP_PKEY *pk = EVP_PKEY_new();
    if (!pk || EVP_PKEY_assign_EC_KEY(pk, ec) != 1) {
        if (pk) EVP_PKEY_free(pk);
        EC_KEY_free(ec);
        return NULL;
    }
    return pk;
}

static EVP_PKEY *pkey_from_point65(const unsigned char *point, size_t len) {
    if (len != ENC_PUB_LEN || point[0] != 0x04) return NULL;
    EC_KEY *ec = EC_KEY_new_by_curve_name(NID_sm2);
    if (!ec) return NULL;
    EC_KEY_set_asn1_flag(ec, OPENSSL_EC_NAMED_CURVE);
    const EC_GROUP *g = EC_KEY_get0_group(ec);
    EC_POINT *p = EC_POINT_new(g);
    if (!p || EC_POINT_oct2point(g, p, point, len, NULL) != 1) {
        EC_POINT_free(p); EC_KEY_free(ec); return NULL;
    }
    EC_KEY_set_public_key(ec, p);
    EC_POINT_free(p);
    EVP_PKEY *pk = EVP_PKEY_new();
    if (!pk || EVP_PKEY_assign_EC_KEY(pk, ec) != 1) {
        if (pk) EVP_PKEY_free(pk);
        EC_KEY_free(ec);
        return NULL;
    }
    return pk;
}

static int get_uncompressed_point(EVP_PKEY *k, unsigned char out[ENC_PUB_LEN]) {
    const EC_KEY *ec = EVP_PKEY_get0_EC_KEY(k);
    if (!ec) return -1;
    const EC_GROUP *g = EC_KEY_get0_group(ec);
    const EC_POINT *p = EC_KEY_get0_public_key(ec);
    if (!p) return -1;
    size_t n = EC_POINT_point2oct(g, p, POINT_CONVERSION_UNCOMPRESSED,
                                  out, ENC_PUB_LEN, NULL);
    return n == ENC_PUB_LEN ? 0 : -1;
}

static int sm3_digest(const unsigned char *data, size_t len, unsigned char out[32]) {
    unsigned int outlen = 0;
    EVP_MD_CTX *ctx = EVP_MD_CTX_new();
    if (!ctx) return -1;
    int ok = EVP_DigestInit_ex(ctx, EVP_sm3(), NULL) == 1 &&
             EVP_DigestUpdate(ctx, data, len) == 1 &&
             EVP_DigestFinal_ex(ctx, out, &outlen) == 1;
    EVP_MD_CTX_free(ctx);
    return ok ? 0 : -1;
}

static int sm2_kdf(const unsigned char *z, size_t zlen,
                   unsigned char *out, size_t outlen) {
    uint32_t ct = 1;
    size_t offset = 0;
    while (offset < outlen) {
        unsigned char buf[32];
        unsigned char *input = malloc(zlen + 4);
        if (!input) return -1;
        memcpy(input, z, zlen);
        input[zlen]     = (unsigned char)((ct >> 24) & 0xff);
        input[zlen + 1] = (unsigned char)((ct >> 16) & 0xff);
        input[zlen + 2] = (unsigned char)((ct >> 8) & 0xff);
        input[zlen + 3] = (unsigned char)(ct & 0xff);
        if (sm3_digest(input, zlen + 4, buf) != 0) { free(input); return -1; }
        free(input);
        size_t copy = outlen - offset;
        if (copy > 32) copy = 32;
        memcpy(out + offset, buf, copy);
        offset += copy;
        ct++;
    }
    return 0;
}

/* SM2 加密，C1C3C2 格式：X(32) || Y(32) || C3(32) || C2(inlen) */
static int sm2_encrypt_manual(EVP_PKEY *pub, const unsigned char *in, size_t inlen,
                               unsigned char **out, size_t *out_len) {
    const EC_KEY *ec = EVP_PKEY_get0_EC_KEY(pub);
    if (!ec) return -1;
    const EC_GROUP *group = EC_KEY_get0_group(ec);
    const EC_POINT *P = EC_KEY_get0_public_key(ec);
    const BIGNUM *order = EC_GROUP_get0_order(group);
    if (!group || !P || !order) return -1;

    BIGNUM *k = BN_new();
    if (!k) return -1;
    if (BN_rand_range(k, order) != 1) { BN_free(k); return -1; }

    EC_POINT *C1 = EC_POINT_new(group);
    EC_POINT *S  = EC_POINT_new(group);
    BIGNUM *x1 = BN_new(), *y1 = BN_new();
    BIGNUM *x2 = BN_new(), *y2 = BN_new();
    unsigned char *t = NULL, *C2 = NULL, *c3_in = NULL, *result = NULL;
    int ret = -1;

    if (!C1 || !S || !x1 || !y1 || !x2 || !y2) goto end;

    if (EC_POINT_mul(group, C1, k, NULL, NULL, NULL) != 1) goto end;
    if (EC_POINT_mul(group, S, NULL, P, k, NULL) != 1) goto end;
    if (EC_POINT_get_affine_coordinates(group, C1, x1, y1, NULL) != 1) goto end;
    if (EC_POINT_get_affine_coordinates(group, S,  x2, y2, NULL) != 1) goto end;

    unsigned char x1_buf[32], y1_buf[32], x2_buf[32], y2_buf[32];
    BN_bn2binpad(x1, x1_buf, 32);
    BN_bn2binpad(y1, y1_buf, 32);
    BN_bn2binpad(x2, x2_buf, 32);
    BN_bn2binpad(y2, y2_buf, 32);

    unsigned char z[64];
    memcpy(z,      x2_buf, 32);
    memcpy(z + 32, y2_buf, 32);
    t = malloc(inlen);
    if (!t) goto end;
    if (sm2_kdf(z, 64, t, inlen) != 0) goto end;

    C2 = malloc(inlen);
    if (!C2) goto end;
    for (size_t i = 0; i < inlen; i++) C2[i] = in[i] ^ t[i];

    c3_in = malloc(32 + inlen + 32);
    if (!c3_in) goto end;
    memcpy(c3_in,              x2_buf, 32);
    memcpy(c3_in + 32,         in,     inlen);
    memcpy(c3_in + 32 + inlen, y2_buf, 32);
    unsigned char C3[32];
    if (sm3_digest(c3_in, 32 + inlen + 32, C3) != 0) goto end;

    size_t total = 64 + 32 + inlen;
    result = malloc(total);
    if (!result) goto end;
    memcpy(result,        x1_buf, 32);
    memcpy(result + 32,   y1_buf, 32);
    memcpy(result + 64,   C3,     32);
    memcpy(result + 96,   C2,     inlen);

    *out = result;
    *out_len = total;
    ret = 0;

end:
    if (k) BN_free(k);
    if (C1) EC_POINT_free(C1);
    if (S) EC_POINT_free(S);
    if (x1) BN_free(x1);
    if (y1) BN_free(y1);
    if (x2) BN_free(x2);
    if (y2) BN_free(y2);
    if (t) free(t);
    if (C2) free(C2);
    if (c3_in) free(c3_in);
    return ret;
}

/* SM2 解密，接收 C1C3C2 格式 */
static int sm2_decrypt_manual(EVP_PKEY *priv, const unsigned char *in, size_t inlen,
                               unsigned char **out, size_t *out_len) {
    if (inlen < 96) {
        fprintf(stderr, "cfca_envelope: SM2 密文太短 %zu\n", inlen);
        return -1;
    }
    const EC_KEY *ec = EVP_PKEY_get0_EC_KEY(priv);
    if (!ec) return -1;
    const EC_GROUP *group = EC_KEY_get0_group(ec);
    const BIGNUM *d = EC_KEY_get0_private_key(ec);
    if (!group || !d) return -1;

    const unsigned char *x1_buf = in;
    const unsigned char *y1_buf = in + 32;
    const unsigned char *C3     = in + 64;
    const unsigned char *C2     = in + 96;
    size_t c2_len = inlen - 96;

    BIGNUM *x1 = BN_bin2bn(x1_buf, 32, NULL);
    BIGNUM *y1 = BN_bin2bn(y1_buf, 32, NULL);
    if (!x1 || !y1) { BN_free(x1); BN_free(y1); return -1; }

    EC_POINT *C1 = EC_POINT_new(group);
    if (!C1 || EC_POINT_set_affine_coordinates(group, C1, x1, y1, NULL) != 1) {
        BN_free(x1); BN_free(y1); EC_POINT_free(C1); return -1;
    }
    BN_free(x1); BN_free(y1);

    EC_POINT *S = EC_POINT_new(group);
    BIGNUM *x2 = BN_new(), *y2 = BN_new();
    unsigned char *t = NULL, *M = NULL, *c3_in = NULL;
    int ret = -1;

    if (!S || !x2 || !y2) goto end;
    if (EC_POINT_mul(group, S, NULL, C1, d, NULL) != 1) goto end;
    if (EC_POINT_get_affine_coordinates(group, S, x2, y2, NULL) != 1) goto end;

    unsigned char x2_buf[32], y2_buf[32], z[64];
    BN_bn2binpad(x2, x2_buf, 32);
    BN_bn2binpad(y2, y2_buf, 32);
    memcpy(z, x2_buf, 32);
    memcpy(z + 32, y2_buf, 32);
    t = malloc(c2_len);
    M = malloc(c2_len);
    c3_in = malloc(32 + c2_len + 32);
    if (!t || !M || !c3_in) goto end;
    if (sm2_kdf(z, 64, t, c2_len) != 0) goto end;

    for (size_t i = 0; i < c2_len; i++) M[i] = C2[i] ^ t[i];

    memcpy(c3_in,               x2_buf, 32);
    memcpy(c3_in + 32,          M,      c2_len);
    memcpy(c3_in + 32 + c2_len, y2_buf, 32);
    unsigned char C3_calc[32];
    if (sm3_digest(c3_in, 32 + c2_len + 32, C3_calc) != 0) goto end;
    if (memcmp(C3_calc, C3, 32) != 0) {
        fprintf(stderr, "cfca_envelope: SM2 C3 校验失败\n");
        goto end;
    }

    *out = M;
    *out_len = c2_len;
    M = NULL;
    ret = 0;

end:
    if (C1) EC_POINT_free(C1);
    if (S) EC_POINT_free(S);
    if (x2) BN_free(x2);
    if (y2) BN_free(y2);
    if (t) free(t);
    if (M) free(M);
    if (c3_in) free(c3_in);
    return ret;
}

static int parse_tl_raw(const unsigned char *p, const unsigned char *end,
                        unsigned char *tag, size_t *len,
                        const unsigned char **content, const unsigned char **next) {
    if (p + 2 > end) return -1;
    unsigned char t = p[0];
    unsigned char l0 = p[1];
    p += 2;
    size_t l;
    if ((l0 & 0x80) == 0) {
        l = l0;
    } else {
        int nbytes = l0 & 0x7F;
        if (nbytes < 1 || nbytes > 4 || p + nbytes > end) return -1;
        l = 0;
        for (int i = 0; i < nbytes; i++) l = (l << 8) | p[i];
        p += nbytes;
    }
    if (p + l > end) return -1;
    *tag = t;
    *len = l;
    *content = p;
    *next = p + l;
    return 0;
}

static int sm4_ecb_nopad(const unsigned char key[SM4_KEY_LEN],
                         const unsigned char *in, size_t inlen, int encrypt,
                         unsigned char **out, size_t *outlen) {
    if (inlen % SM4_BLOCK != 0) return -1;
    EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
    if (!ctx) return -1;
    unsigned char *buf = malloc(inlen + SM4_BLOCK);
    if (!buf) { EVP_CIPHER_CTX_free(ctx); return -1; }
    int outl1 = 0, outl2 = 0;
    int ok = (encrypt ? EVP_EncryptInit_ex(ctx, EVP_sm4_ecb(), NULL, key, NULL)
                      : EVP_DecryptInit_ex(ctx, EVP_sm4_ecb(), NULL, key, NULL)) == 1 &&
             EVP_CIPHER_CTX_set_padding(ctx, 0) == 1 &&
             (encrypt ? EVP_EncryptUpdate(ctx, buf, &outl1, in, (int)inlen)
                      : EVP_DecryptUpdate(ctx, buf, &outl1, in, (int)inlen)) == 1 &&
             (encrypt ? EVP_EncryptFinal_ex(ctx, buf + outl1, &outl2)
                      : EVP_DecryptFinal_ex(ctx, buf + outl1, &outl2)) == 1;
    EVP_CIPHER_CTX_free(ctx);
    if (!ok) { free(buf); return -1; }
    *out = buf;
    *outlen = (size_t)(outl1 + outl2);
    return 0;
}

/* ==========================================================================
 * 构造简化格式 EnvelopedKey：
 *   SEQUENCE { INTEGER 2, OCTET STRING <X||Y||C3||C2 = 192字节> }
 * ========================================================================== */
static int build_enveloped_key_der(
    const unsigned char *raw_c1c3c2, size_t raw_len,
    unsigned char **out_der, int *out_len)
{
    unsigned char int_der[] = { 0x02, 0x01, 0x02 };

    int oct_content_len = (int)raw_len;
    int oct_hdr_len = (oct_content_len < 0x80) ? 2 :
                      (oct_content_len < 0x100) ? 3 : 4;
    int oct_der_len = oct_hdr_len + oct_content_len;

    unsigned char *oct_der = malloc((size_t)oct_der_len);
    if (!oct_der) return -1;
    int p = 0;
    oct_der[p++] = 0x04;
    if (oct_content_len < 0x80) {
        oct_der[p++] = (unsigned char)oct_content_len;
    } else if (oct_content_len < 0x100) {
        oct_der[p++] = 0x81;
        oct_der[p++] = (unsigned char)oct_content_len;
    } else {
        oct_der[p++] = 0x82;
        oct_der[p++] = (unsigned char)((oct_content_len >> 8) & 0xff);
        oct_der[p++] = (unsigned char)(oct_content_len & 0xff);
    }
    memcpy(oct_der + p, raw_c1c3c2, raw_len);
    p += (int)raw_len;

    int content_len = (int)sizeof(int_der) + oct_der_len;
    int seq_hdr_len = (content_len < 0x80) ? 2 :
                      (content_len < 0x100) ? 3 : 4;
    int total = seq_hdr_len + content_len;

    unsigned char *der = malloc((size_t)total);
    if (!der) { free(oct_der); return -1; }
    int q = 0;
    der[q++] = 0x30;
    if (content_len < 0x80) {
        der[q++] = (unsigned char)content_len;
    } else if (content_len < 0x100) {
        der[q++] = 0x81;
        der[q++] = (unsigned char)content_len;
    } else {
        der[q++] = 0x82;
        der[q++] = (unsigned char)((content_len >> 8) & 0xff);
        der[q++] = (unsigned char)(content_len & 0xff);
    }
    memcpy(der + q, int_der, sizeof(int_der)); q += (int)sizeof(int_der);
    memcpy(der + q, oct_der, (size_t)oct_der_len); q += oct_der_len;

    free(oct_der);

    *out_der = der;
    *out_len = total;
    return 0;
}

/* ==========================================================================
 * 解析简化格式 EnvelopedKey：
 *   SEQUENCE { INTEGER 2, OCTET STRING <192字节 raw C1C3C2> }
 * ========================================================================== */
static int parse_enveloped_key_der(
    const unsigned char *der, size_t der_len,
    unsigned char **out_sm2_raw, size_t *out_sm2_len)
{
    const unsigned char *p = der;
    const unsigned char *end = der + der_len;
    unsigned char tag; size_t len;
    const unsigned char *content, *next;

    if (parse_tl_raw(p, end, &tag, &len, &content, &next) != 0 || tag != 0x30) {
        fprintf(stderr, "cfca_envelope: EnvelopedKey 不是 SEQUENCE\n");
        return -1;
    }
    const unsigned char *seq_end = content + len;
    p = content;

    if (parse_tl_raw(p, seq_end, &tag, &len, &content, &next) != 0 || tag != 0x02) {
        fprintf(stderr, "cfca_envelope: 期望 INTEGER version\n");
        return -1;
    }
    p = next;

    if (parse_tl_raw(p, seq_end, &tag, &len, &content, &next) != 0 || tag != 0x04) {
        fprintf(stderr, "cfca_envelope: 期望 OCTET STRING\n");
        return -1;
    }
    unsigned char *raw = malloc(len);
    if (!raw) return -1;
    memcpy(raw, content, len);

    *out_sm2_raw = raw;
    *out_sm2_len = len;
    return 0;
}

/* ==========================================================================
 * 5 段格式：header 前 4 字段十进制 1/1/0/0，第 5 字段十进制 body_len
 * ========================================================================== */
static char *assemble_cfca_5seg(const unsigned char *ek_der, int ek_len,
                                size_t *out_len) {
    char *b64[5] = {0};

    for (int i = 0; i < 4; i++) {
        int off = i * CFCA_SEG_BYTES;
        if (off >= ek_len) { b64[i] = strdup(""); continue; }
        int seg = ek_len - off;
        if (seg > CFCA_SEG_BYTES) seg = CFCA_SEG_BYTES;
        b64[i] = b64_encode_alloc(ek_der + off, (size_t)seg);
        if (!b64[i]) {
            for (int j = 0; j < i; j++) free(b64[j]);
            return NULL;
        }
    }
    int off = 4 * CFCA_SEG_BYTES;
    if (ek_len > off) {
        b64[4] = b64_encode_alloc(ek_der + off, (size_t)(ek_len - off));
    } else {
        b64[4] = strdup("");
    }
    if (!b64[4]) {
        for (int i = 0; i < 4; i++) free(b64[i]);
        return NULL;
    }

    uint64_t body_len = 0;
    for (int i = 0; i < 5; i++) {
        body_len += (uint64_t)strlen(b64[i]);
        if (i < 4) body_len += 1;
    }
    body_len += 1;

    char hdr[81];
    snprintf(hdr,      17, "%016llu", (unsigned long long)1);
    snprintf(hdr + 16, 17, "%016llu", (unsigned long long)1);
    snprintf(hdr + 32, 17, "%016llu", (unsigned long long)0);
    snprintf(hdr + 48, 17, "%016llu", (unsigned long long)0);
    snprintf(hdr + 64, 17, "%016llu", (unsigned long long)body_len);

    size_t total = 80
                 + strlen(b64[0]) + 1 + strlen(b64[1]) + 1
                 + strlen(b64[2]) + 1 + strlen(b64[3]) + 1
                 + strlen(b64[4]) + 2;
    char *out = malloc(total);
    if (!out) {
        for (int i = 0; i < 5; i++) free(b64[i]);
        return NULL;
    }
    snprintf(out, total, "%s%s,%s,%s,%s,%s,",
             hdr, b64[0], b64[1], b64[2], b64[3], b64[4]);
    for (int i = 0; i < 5; i++) free(b64[i]);
    *out_len = strlen(out);
    return out;
}

static unsigned char *parse_cfca_5seg(const char *text, size_t text_len,
                                       size_t *out_der_len) {
    if (text_len < 80) return NULL;

    char len_buf[17];
    memcpy(len_buf, text + 64, 16);
    len_buf[16] = 0;

    uint64_t body_len = strtoull(len_buf, NULL, 10);
    if (body_len == 0 || body_len > (uint64_t)(text_len - 80)) {
        fprintf(stderr, "cfca_envelope: body_len=%llu 非法\n", (unsigned long long)body_len);
        return NULL;
    }

    const char *p = text + 80;
    const char *end = p + body_len;

    char *concat = malloc((size_t)body_len + 1);
    if (!concat) return NULL;
    size_t clen = 0;
    for (const char *q = p; q < end; q++) {
        if (*q == ',' || *q == '\n' || *q == '\r' || *q == ' ' || *q == '\t')
            continue;
        concat[clen++] = *q;
    }
    concat[clen] = 0;

    size_t ek_len = 0;
    unsigned char *ek_der = b64_decode_alloc(concat, clen, &ek_len);
    free(concat);
    if (!ek_der) return NULL;
    *out_der_len = ek_len;
    return ek_der;
}

/* ==========================================================================
 * do_seal：SM2 加密 96 字节明文，构造简化 EnvelopedKey
 * ========================================================================== */
static int do_seal(const char *sign_pub_path, const char *enc_key_path,
                   const char *out_path, int cfca_format) {
    int rc = -1;
    EVP_PKEY *sign_pub = NULL, *enc_key = NULL;
    unsigned char sm4_key[SM4_KEY_LEN];
    unsigned char raw_d[SM2_D_LEN];
    unsigned char plain96[CFCA_PLAIN_LEN];
    unsigned char *sm4_cipher_D = NULL;
    size_t sm4_cipher_D_len = 0;
    unsigned char *sm2_raw = NULL;
    size_t sm2_raw_len = 0;
    unsigned char *ek_der = NULL;
    int ek_len = 0;

    sign_pub = load_pub(sign_pub_path);
    if (!sign_pub) goto out;
    enc_key = load_priv(enc_key_path);
    if (!enc_key) goto out;

    if (RAND_bytes(sm4_key, SM4_KEY_LEN) != 1) goto out;
    if (get_raw_d_32(enc_key, raw_d) != 0) goto out;

    if (sm4_ecb_nopad(sm4_key, raw_d, SM2_D_LEN, 1,
                      &sm4_cipher_D, &sm4_cipher_D_len) != 0) goto out;
    if (sm4_cipher_D_len != 32) goto out;

    memset(plain96, 0, sizeof(plain96));
    memcpy(plain96,      sm4_key,        SM4_KEY_LEN);
    memcpy(plain96 + 16, sm4_cipher_D,   32);

    if (sm2_encrypt_manual(sign_pub, plain96, CFCA_PLAIN_LEN,
                           &sm2_raw, &sm2_raw_len) != 0) {
        fprintf(stderr, "cfca_envelope: SM2 加密失败\n");
        goto out;
    }
    fprintf(stderr, "cfca_envelope: SM2 raw = %zu bytes (plain=96)\n", sm2_raw_len);
    if (sm2_raw_len != CFCA_SYM_PRI_LEN) goto out;

    if (build_enveloped_key_der(sm2_raw, sm2_raw_len, &ek_der, &ek_len) != 0) {
        fprintf(stderr, "cfca_envelope: 构造 EnvelopedKey 失败\n");
        goto out;
    }
    fprintf(stderr, "cfca_envelope: EnvelopedKey DER = %d bytes\n", ek_len);

    if (cfca_format) {
        size_t txt_len = 0;
        char *txt = assemble_cfca_5seg(ek_der, ek_len, &txt_len);
        if (!txt) goto out;
        int ok = write_bytes(out_path, (const unsigned char *)txt, txt_len) == 0;
        free(txt);
        if (!ok) goto out;
    } else {
        if (write_bytes(out_path, ek_der, (size_t)ek_len) != 0) goto out;
    }
    rc = 0;
out:
    if (ek_der) free(ek_der);
    if (sm2_raw) free(sm2_raw);
    if (sm4_cipher_D) free(sm4_cipher_D);
    if (enc_key) EVP_PKEY_free(enc_key);
    if (sign_pub) EVP_PKEY_free(sign_pub);
    return rc;
}

/* ==========================================================================
 * unseal：SM2 解密 96 字节，取前 16 为 SM4 密钥，[16,48) 为 SM4-ECB 加密的 D
 * ========================================================================== */
static int unseal_from_ek_der(const unsigned char *ek_der, size_t ek_len,
                              EVP_PKEY *sign_priv, unsigned char **out_d) {
    unsigned char *sm2_raw = NULL;
    size_t sm2_raw_len = 0;

    if (parse_enveloped_key_der(ek_der, ek_len, &sm2_raw, &sm2_raw_len) != 0) {
        fprintf(stderr, "cfca_envelope: EnvelopedKey 解析失败\n");
        return -1;
    }

    unsigned char *plain = NULL;
    size_t plain_len = 0;
    if (sm2_decrypt_manual(sign_priv, sm2_raw, sm2_raw_len,
                           &plain, &plain_len) != 0) {
        free(sm2_raw);
        return -1;
    }
    free(sm2_raw);

    if (plain_len < SM4_KEY_LEN + SM2_D_LEN) {
        fprintf(stderr, "cfca_envelope: SM2 明文 %zu 字节，不足 48\n", plain_len);
        free(plain);
        return -1;
    }

    unsigned char *raw_d = NULL;
    size_t raw_d_len = 0;
    if (sm4_ecb_nopad(plain, plain + 16, 32, 0, &raw_d, &raw_d_len) != 0 ||
        raw_d_len != SM2_D_LEN) {
        fprintf(stderr, "cfca_envelope: SM4-ECB 解密失败\n");
        free(plain);
        if (raw_d) free(raw_d);
        return -1;
    }
    free(plain);

    *out_d = raw_d;
    return 0;
}

static int do_unseal(const char *sign_key_path, const char *in_path,
                     const char *out_path) {
    int rc = -1;
    EVP_PKEY *sign_priv = NULL, *enc_key = NULL;
    unsigned char *env_bytes = NULL;
    long env_len = 0;
    unsigned char *raw_d = NULL;

    sign_priv = load_priv(sign_key_path);
    if (!sign_priv) goto out;

    {
        FILE *f = fopen(in_path, "rb");
        if (!f) { perror(in_path); goto out; }
        fseek(f, 0, SEEK_END);
        env_len = ftell(f);
        fseek(f, 0, SEEK_SET);
        env_bytes = malloc((size_t)env_len + 1);
        if (!env_bytes || fread(env_bytes, 1, (size_t)env_len, f) != (size_t)env_len) {
            fclose(f); goto out;
        }
        env_bytes[env_len] = 0;
        fclose(f);
    }

    int is_cfca = 0;
    if (env_len > 80) {
        int all_digit = 1;
        for (int i = 0; i < 80; i++) {
            char c = (char)env_bytes[i];
            if (c < '0' || c > '9') { all_digit = 0; break; }
        }
        if (all_digit && memchr(env_bytes, ',', (size_t)env_len)) is_cfca = 1;
    }

    if (is_cfca) {
        size_t ek_len = 0;
        unsigned char *ek_der = parse_cfca_5seg((const char *)env_bytes,
                                                 (size_t)env_len, &ek_len);
        if (!ek_der) { goto out; }
        if (unseal_from_ek_der(ek_der, ek_len, sign_priv, &raw_d) != 0) {
            free(ek_der); goto out;
        }
        free(ek_der);
    } else {
        if (unseal_from_ek_der(env_bytes, (size_t)env_len, sign_priv, &raw_d) != 0)
            goto out;
    }

    enc_key = pkey_from_raw_d(raw_d);
    if (!enc_key) goto out;
    if (write_priv_pem(out_path, enc_key) != 0) goto out;
    rc = 0;
out:
    if (enc_key) EVP_PKEY_free(enc_key);
    if (raw_d) free(raw_d);
    if (env_bytes) free(env_bytes);
    if (sign_priv) EVP_PKEY_free(sign_priv);
    return rc;
}

/* ==========================================================================
 * CSR 相关
 * ========================================================================== */
static X509_NAME *parse_subject(const char *subject) {
    X509_NAME *name = X509_NAME_new();
    if (!name) return NULL;
    const char *p = subject;
    if (*p == '/') p++;

    char buf[1024];
    size_t n = strlen(p);
    if (n >= sizeof(buf)) { X509_NAME_free(name); return NULL; }
    memcpy(buf, p, n);
    buf[n] = 0;

    char *saveptr = NULL;
    char *tok = strtok_r(buf, "/,", &saveptr);
    while (tok) {
        char *eq = strchr(tok, '=');
        if (eq) {
            *eq = 0;
            int nid = OBJ_txt2nid(tok);
            if (nid == NID_undef) { X509_NAME_free(name); return NULL; }
            if (X509_NAME_add_entry_by_NID(name, nid, MBSTRING_UTF8,
                                           (unsigned char *)(eq + 1), -1, -1, 0) != 1) {
                X509_NAME_free(name);
                return NULL;
            }
        }
        tok = strtok_r(NULL, "/,", &saveptr);
    }
    return name;
}

/* ==========================================================================
 * do_gen_csr —— 支持内置固定临时密钥
 *
 *   --tmp-key <path>       使用已有的临时私钥（不再生成新密钥对）
 *                          此时 --out-tmp-key / --out-tmp-pub-der 均为可选
 *
 *   --out-tmp-key <path>   生成新临时密钥对时必填
 *   --out-tmp-pub-der <path> 生成新临时密钥对时必填
 * ========================================================================== */
static int do_gen_csr(int argc, char **argv) {
    const char *subject = NULL;
    const char *out_csr = NULL, *out_sign_key = NULL;
    const char *out_tmp_key = NULL, *out_tmp_pub_der = NULL;
    const char *in_tmp_key = NULL;                 /* ★ 内置固定临时私钥 */

    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--subject") && i + 1 < argc)               subject = argv[++i];
        else if (!strcmp(argv[i], "--out-csr") && i + 1 < argc)          out_csr = argv[++i];
        else if (!strcmp(argv[i], "--out-sign-key") && i + 1 < argc)     out_sign_key = argv[++i];
        else if (!strcmp(argv[i], "--out-tmp-key") && i + 1 < argc)      out_tmp_key = argv[++i];
        else if (!strcmp(argv[i], "--out-tmp-pub-der") && i + 1 < argc)  out_tmp_pub_der = argv[++i];
        else if (!strcmp(argv[i], "--tmp-key") && i + 1 < argc)          in_tmp_key = argv[++i];    /* ★ */
        else if (!strcmp(argv[i], "--challenge-password") && i + 1 < argc) { ++i; }
        else { usage(argv[0]); return 2; }
    }

    if (!subject || !out_csr || !out_sign_key) { usage(argv[0]); return 2; }
    /* 若未提供内置临时私钥，则必须输出新生成的临时密钥对 */
    if (!in_tmp_key && (!out_tmp_key || !out_tmp_pub_der)) {
        usage(argv[0]); return 2;
    }

    int rc = -1;
    EVP_PKEY *sign_key = NULL, *tmp_key = NULL;
    X509_REQ *req = NULL;
    X509_NAME *name = NULL;
    unsigned char *tmp_pub_der = NULL;
    int tmp_pub_der_len = 0;
    int tmp_key_generated = 0;

    sign_key = gen_sm2_key();
    if (!sign_key) goto out;

    /* ★ 临时密钥：优先使用内置固定私钥；否则随机生成 */
    if (in_tmp_key) {
        tmp_key = load_priv(in_tmp_key);
        if (!tmp_key) {
            fprintf(stderr, "cfca_envelope: 加载内置临时私钥失败: %s\n", in_tmp_key);
            goto out;
        }
    } else {
        tmp_key = gen_sm2_key();
        if (!tmp_key) goto out;
        tmp_key_generated = 1;
    }

    tmp_pub_der_len = i2d_PUBKEY(tmp_key, &tmp_pub_der);
    if (tmp_pub_der_len <= 0) goto out;

    req = X509_REQ_new();
    if (!req) goto out;
    X509_REQ_set_version(req, 0);

    name = parse_subject(subject);
    if (!name) goto out;
    X509_REQ_set_subject_name(req, name);
    X509_REQ_set_pubkey(req, sign_key);

    unsigned char pub_point[ENC_PUB_LEN];
    if (get_uncompressed_point(tmp_key, pub_point) != 0) goto out;

    unsigned char keybytes[CFCA_KEYBYTES_LEN];
    memset(keybytes, 0, sizeof(keybytes));
    keybytes[0] = 0x00; keybytes[1] = 0xB4;
    keybytes[2] = 0x00; keybytes[3] = 0x00;
    keybytes[4] = 0x00; keybytes[5] = 0x01;
    keybytes[6] = 0x00; keybytes[7] = 0x00;
    memcpy(keybytes + 8,  pub_point + 1,  32);
    memcpy(keybytes + 72, pub_point + 33, 32);

    unsigned char seq_der[145];
    int dpos = 0;
    seq_der[dpos++] = 0x30;
    seq_der[dpos++] = 0x81;
    seq_der[dpos++] = 0x8E;
    seq_der[dpos++] = 0x02;
    seq_der[dpos++] = 0x01;
    seq_der[dpos++] = 0x01;
    seq_der[dpos++] = 0x04;
    seq_der[dpos++] = 0x81;
    seq_der[dpos++] = 0x88;
    memcpy(seq_der + dpos, keybytes, CFCA_KEYBYTES_LEN);
    dpos += CFCA_KEYBYTES_LEN;

    {
        ASN1_OBJECT *tmp_oid = OBJ_txt2obj(OID_TMP_PUB, 1);
        if (!tmp_oid) goto out;
        X509_ATTRIBUTE *attr = X509_ATTRIBUTE_create_by_OBJ(
            NULL, tmp_oid, V_ASN1_OCTET_STRING, seq_der, dpos);
        ASN1_OBJECT_free(tmp_oid);
        if (!attr) goto out;
        if (X509_REQ_add1_attr(req, attr) != 1) {
            X509_ATTRIBUTE_free(attr);
            goto out;
        }
        X509_ATTRIBUTE_free(attr);
    }

    if (X509_REQ_sign(req, sign_key, EVP_sm3()) <= 0) goto out;
    if (write_csr_pem(out_csr, req) != 0) goto out;
    if (write_priv_pem(out_sign_key, sign_key) != 0) goto out;

    /* ★ 仅当本次生成了新临时密钥时，才输出临时私钥 */
    if (tmp_key_generated && out_tmp_key) {
        if (write_priv_pem(out_tmp_key, tmp_key) != 0) goto out;
    }
    /* 临时公钥 DER：可选输出 */
    if (out_tmp_pub_der) {
        if (write_bytes(out_tmp_pub_der, tmp_pub_der, (size_t)tmp_pub_der_len) != 0) goto out;
    }

    rc = 0;
out:
    if (tmp_pub_der) OPENSSL_free(tmp_pub_der);
    if (name) X509_NAME_free(name);
    if (req) X509_REQ_free(req);
    if (sign_key) EVP_PKEY_free(sign_key);
    if (tmp_key) EVP_PKEY_free(tmp_key);
    return rc;
}

static int do_gen_plain_csr(int argc, char **argv) {
    const char *subject = NULL;
    const char *out_csr = NULL, *out_key = NULL;

    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--subject") && i + 1 < argc)  subject = argv[++i];
        else if (!strcmp(argv[i], "--out-csr") && i + 1 < argc) out_csr = argv[++i];
        else if (!strcmp(argv[i], "--out-key") && i + 1 < argc) out_key = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!subject || !out_csr || !out_key) { usage(argv[0]); return 2; }

    int rc = -1;
    EVP_PKEY *key = NULL;
    X509_REQ *req = NULL;
    X509_NAME *name = NULL;

    key = gen_sm2_key();
    if (!key) goto out;

    req = X509_REQ_new();
    if (!req) goto out;
    X509_REQ_set_version(req, 0);

    name = parse_subject(subject);
    if (!name) goto out;
    X509_REQ_set_subject_name(req, name);
    X509_REQ_set_pubkey(req, key);

    if (X509_REQ_sign(req, key, EVP_sm3()) <= 0) goto out;
    if (write_csr_pem(out_csr, req) != 0) goto out;
    if (write_priv_pem(out_key, key) != 0) goto out;
    rc = 0;
out:
    if (name) X509_NAME_free(name);
    if (req) X509_REQ_free(req);
    if (key) EVP_PKEY_free(key);
    return rc;
}

static unsigned char *load_pem_body_der(const char *path, size_t *out_len) {
    FILE *f = fopen(path, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long n = ftell(f);
    fseek(f, 0, SEEK_SET);
    if (n <= 0) { fclose(f); return NULL; }
    unsigned char *buf = malloc((size_t)n + 1);
    if (!buf) { fclose(f); return NULL; }
    size_t rd = fread(buf, 1, (size_t)n, f);
    fclose(f);
    if (rd == 0) { free(buf); return NULL; }
    buf[rd] = 0;

    if (rd > 11 && memcmp(buf, "-----BEGIN ", 11) == 0) {
        BIO *bio = BIO_new_mem_buf(buf, (int)rd);
        if (!bio) { free(buf); return NULL; }
        char *name = NULL, *header = NULL;
        unsigned char *data = NULL;
        long data_len = 0;
        int ok = PEM_read_bio(bio, &name, &header, &data, &data_len) != NULL;
        BIO_free(bio);
        OPENSSL_free(name);
        OPENSSL_free(header);
        free(buf);
        if (!ok || data_len <= 0) {
            if (data) OPENSSL_free(data);
            return NULL;
        }
        *out_len = (size_t)data_len;
        return data;
    }

    *out_len = rd;
    unsigned char *der = OPENSSL_malloc(rd);
    if (!der) { free(buf); return NULL; }
    memcpy(der, buf, rd);
    free(buf);
    return der;
}

static EVP_PKEY *scan_point_in_bytes(const unsigned char *data, int data_len) {
    if (!data || data_len <= 0) return NULL;
    const int offsets[] = {0, 8, 17};
    const int n_offsets = (int)(sizeof(offsets) / sizeof(offsets[0]));
    for (int i = 0; i < n_offsets; i++) {
        int off = offsets[i];
        if (off >= data_len) continue;
        if (data_len >= off + ENC_PUB_LEN && data[off] == 0x04) {
            EVP_PKEY *pk = pkey_from_point65(data + off, ENC_PUB_LEN);
            if (pk) return pk;
        }
        if (data_len >= off + 64) {
            unsigned char point[ENC_PUB_LEN];
            point[0] = 0x04;
            memcpy(point + 1,  data + off,      32);
            memcpy(point + 33, data + off + 32, 32);
            EVP_PKEY *pk = pkey_from_point65(point, ENC_PUB_LEN);
            if (pk) return pk;
        }
        if (data_len >= off + 32 + 32 + 32) {
            unsigned char point[ENC_PUB_LEN];
            point[0] = 0x04;
            memcpy(point + 1,  data + off,           32);
            memcpy(point + 33, data + off + 32 + 32, 32);
            EVP_PKEY *pk = pkey_from_point65(point, ENC_PUB_LEN);
            if (pk) return pk;
        }
    }
    return NULL;
}

static EVP_PKEY *parse_cfca_keybytes(const unsigned char *kp, size_t kp_len) {
    if (kp_len < 8 + 32 + 32 + 32) return NULL;
    unsigned char point[ENC_PUB_LEN];
    point[0] = 0x04;
    memcpy(point + 1,  kp + 8,  32);
    memcpy(point + 33, kp + 72, 32);
    return pkey_from_point65(point, ENC_PUB_LEN);
}

static EVP_PKEY *extract_tmp_pub_from_der(const unsigned char *der, size_t der_len) {
    static const unsigned char oid[] = {
        0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x09, 0x3f
    };
    const size_t oid_len = sizeof(oid);
    if (der_len < oid_len) return NULL;

    for (size_t i = 0; i + oid_len <= der_len; i++) {
        if (memcmp(der + i, oid, oid_len) != 0) continue;
        const unsigned char *p   = der + i + oid_len;
        const unsigned char *end = der + der_len;
        unsigned char tag; size_t len;
        const unsigned char *content, *next;
        if (parse_tl_raw(p, end, &tag, &len, &content, &next) != 0) continue;
        if (tag == 0x31) {
            unsigned char it; size_t il;
            const unsigned char *ic, *in;
            if (parse_tl_raw(content, content + len, &it, &il, &ic, &in) != 0) continue;
            content = ic; len = il; tag = it;
        }
        if (tag == 0x03) {
            if (len > 0 && content[0] == 0) { content++; len--; }
            tag = 0x04;
        }
        if (tag != 0x04) continue;
        {
            unsigned char t1; size_t l1;
            const unsigned char *c1, *n1;
            if (parse_tl_raw(content, content + len, &t1, &l1, &c1, &n1) == 0 && t1 == 0x30) {
                const unsigned char *q = c1, *qend = c1 + l1;
                unsigned char t2; size_t l2;
                const unsigned char *c2, *n2;
                if (parse_tl_raw(q, qend, &t2, &l2, &c2, &n2) == 0 && t2 == 0x02) {
                    q = n2;
                    if (parse_tl_raw(q, qend, &t2, &l2, &c2, &n2) == 0 && t2 == 0x04) {
                        EVP_PKEY *pk = parse_cfca_keybytes(c2, l2);
                        if (pk) return pk;
                    }
                }
            }
        }
        {
            EVP_PKEY *pk = scan_point_in_bytes(content, (int)len);
            if (pk) return pk;
        }
    }
    return NULL;
}

static X509_NAME *extract_subject_from_der(const unsigned char *der, size_t der_len) {
    unsigned char tag; size_t len;
    const unsigned char *content, *next;
    const unsigned char *end = der + der_len;
    if (parse_tl_raw(der, end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    const unsigned char *l1_end = next;
    if (parse_tl_raw(content, l1_end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    const unsigned char *info_start = content;
    const unsigned char *info_end = next;
    if (parse_tl_raw(info_start, info_end, &tag, &len, &content, &next) != 0 || tag != 0x02) return NULL;
    const unsigned char *name_start = next;
    if (parse_tl_raw(name_start, info_end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    long name_total_len = (long)(next - name_start);
    const unsigned char *p = name_start;
    return d2i_X509_NAME(NULL, &p, name_total_len);
}

static EVP_PKEY *extract_spki_from_der(const unsigned char *der, size_t der_len,
                                       X509_NAME **out_subject) {
    unsigned char tag; size_t len;
    const unsigned char *content, *next;
    const unsigned char *end = der + der_len;
    if (parse_tl_raw(der, end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    const unsigned char *l1_end = next;
    if (parse_tl_raw(content, l1_end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    const unsigned char *info_start = content;
    const unsigned char *info_end = next;
    if (parse_tl_raw(info_start, info_end, &tag, &len, &content, &next) != 0 || tag != 0x02) return NULL;
    const unsigned char *name_start = next;
    if (parse_tl_raw(name_start, info_end, &tag, &len, &content, &next) != 0 || tag != 0x30) return NULL;
    const unsigned char *name_next = next;

    if (out_subject) {
        long name_len = (long)(name_next - name_start);
        const unsigned char *np = name_start;
        *out_subject = d2i_X509_NAME(NULL, &np, name_len);
        if (!*out_subject) return NULL;
    }

    const unsigned char *spki_start = name_next;
    if (parse_tl_raw(spki_start, info_end, &tag, &len, &content, &next) != 0 || tag != 0x30) {
        if (out_subject && *out_subject) { X509_NAME_free(*out_subject); *out_subject = NULL; }
        return NULL;
    }
    long spki_len = (long)(next - spki_start);
    const unsigned char *sp = spki_start;
    EVP_PKEY *pk = d2i_PUBKEY(NULL, &sp, spki_len);
    if (!pk && out_subject && *out_subject) { X509_NAME_free(*out_subject); *out_subject = NULL; }
    return pk;
}

static int do_csr2pub(int argc, char **argv) {
    const char *in_path = NULL, *out_path = NULL;
    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--in") && i + 1 < argc)  in_path = argv[++i];
        else if (!strcmp(argv[i], "--out") && i + 1 < argc) out_path = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!in_path || !out_path) { usage(argv[0]); return 2; }

    int rc = -1;
    unsigned char *der = NULL;
    size_t der_len = 0;
    EVP_PKEY *tmp_pub = NULL;
    unsigned char *spki_der = NULL;
    int spki_len = 0;

    der = load_pem_body_der(in_path, &der_len);
    if (!der || der_len == 0) goto out;
    tmp_pub = extract_tmp_pub_from_der(der, der_len);
    if (!tmp_pub) goto out;
    spki_len = i2d_PUBKEY(tmp_pub, &spki_der);
    if (spki_len <= 0) goto out;
    if (write_bytes(out_path, spki_der, (size_t)spki_len) != 0) goto out;
    rc = 0;
out:
    if (spki_der) OPENSSL_free(spki_der);
    if (tmp_pub) EVP_PKEY_free(tmp_pub);
    if (der) OPENSSL_free(der);
    return rc;
}

static int do_csr2signpub(int argc, char **argv) {
    const char *in_path = NULL, *out_path = NULL;
    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--in") && i + 1 < argc)  in_path = argv[++i];
        else if (!strcmp(argv[i], "--out") && i + 1 < argc) out_path = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!in_path || !out_path) { usage(argv[0]); return 2; }

    int rc = -1;
    EVP_PKEY *pubkey = NULL;
    X509_NAME *subject = NULL;
    unsigned char *csr_der = NULL;
    BUF_MEM *mem = NULL;
    BIO *bio = NULL;

    {
        FILE *f = fopen(in_path, "rb");
        if (f) {
            X509_REQ *req = PEM_read_X509_REQ(f, NULL, NULL, NULL);
            fclose(f);
            if (req) {
                pubkey = X509_REQ_get_pubkey(req);
                X509_REQ_free(req);
            } else {
                ERR_clear_error();
            }
        }
    }

    if (!pubkey) {
        size_t der_len = 0;
        csr_der = load_pem_body_der(in_path, &der_len);
        if (!csr_der || der_len == 0) goto out;
        pubkey = extract_spki_from_der(csr_der, der_len, &subject);
        if (!pubkey) goto out;
    }

    bio = BIO_new(BIO_s_mem());
    if (!bio) goto out;
    if (PEM_write_bio_PUBKEY(bio, pubkey) != 1) goto out;
    BIO_get_mem_ptr(bio, &mem);
    if (!mem || !mem->data || mem->length == 0) goto out;
    if (write_bytes(out_path, (const unsigned char *)mem->data, mem->length) != 0)
        goto out;
    rc = 0;
out:
    if (bio) BIO_free(bio);
    if (csr_der) OPENSSL_free(csr_der);
    if (subject) X509_NAME_free(subject);
    if (pubkey) EVP_PKEY_free(pubkey);
    return rc;
}

static int do_csr2subj(int argc, char **argv) {
    const char *in_path = NULL, *out_path = NULL;
    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--in") && i + 1 < argc)  in_path = argv[++i];
        else if (!strcmp(argv[i], "--out") && i + 1 < argc) out_path = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!in_path || !out_path) { usage(argv[0]); return 2; }

    int rc = -1;
    unsigned char *der = NULL;
    size_t der_len = 0;
    X509_NAME *name = NULL;
    BIO *out = NULL;

    der = load_pem_body_der(in_path, &der_len);
    if (!der || der_len == 0) goto out;
    name = extract_subject_from_der(der, der_len);
    if (!name) goto out;
    out = BIO_new_file(out_path, "w");
    if (!out) goto out;
    X509_NAME_print_ex(out, name, 0, XN_FLAG_RFC2253);
    BIO_printf(out, "\n");
    BIO_flush(out);
    rc = 0;
out:
    if (out) BIO_free(out);
    if (name) X509_NAME_free(name);
    if (der) OPENSSL_free(der);
    return rc;
}

static int do_sign_cert(int argc, char **argv) {
    const char *csr_path = NULL, *ca_cert_path = NULL, *ca_key_path = NULL;
    const char *out_path = NULL, *san_str = NULL, *subject_str = NULL;
    int days = 365;
    int is_enc = 0;

    for (int i = 2; i < argc; i++) {
        if (!strcmp(argv[i], "--csr") && i + 1 < argc)          csr_path = argv[++i];
        else if (!strcmp(argv[i], "--ca-cert") && i + 1 < argc) ca_cert_path = argv[++i];
        else if (!strcmp(argv[i], "--ca-key") && i + 1 < argc)  ca_key_path = argv[++i];
        else if (!strcmp(argv[i], "--out-cert") && i + 1 < argc) out_path = argv[++i];
        else if (!strcmp(argv[i], "--days") && i + 1 < argc)    days = atoi(argv[++i]);
        else if (!strcmp(argv[i], "--is-enc") && i + 1 < argc)  is_enc = atoi(argv[++i]);
        else if (!strcmp(argv[i], "--san") && i + 1 < argc)     san_str = argv[++i];
        else if (!strcmp(argv[i], "--subject") && i + 1 < argc) subject_str = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!csr_path || !ca_cert_path || !ca_key_path || !out_path) { usage(argv[0]); return 2; }

    int rc = -1;
    EVP_PKEY *pubkey = NULL;
    X509_NAME *csr_subject = NULL;
    X509_NAME *subject = NULL;
    X509 *ca_cert = NULL;
    EVP_PKEY *ca_key = NULL;
    X509 *cert = NULL;
    unsigned char *csr_der = NULL;

    /* 1. 从 CSR 提取公钥 + subject */
    {
        FILE *f = fopen(csr_path, "rb");
        if (f) {
            X509_REQ *req = PEM_read_X509_REQ(f, NULL, NULL, NULL);
            fclose(f);
            if (req) {
                csr_subject = X509_NAME_dup(X509_REQ_get_subject_name(req));
                pubkey = X509_REQ_get_pubkey(req);
                X509_REQ_free(req);
            } else {
                ERR_clear_error();
            }
        }
    }
    if (!pubkey || !csr_subject) {
        if (csr_subject) { X509_NAME_free(csr_subject); csr_subject = NULL; }
        if (pubkey) { EVP_PKEY_free(pubkey); pubkey = NULL; }
        size_t der_len = 0;
        csr_der = load_pem_body_der(csr_path, &der_len);
        if (!csr_der || der_len == 0) {
            fprintf(stderr, "cfca_envelope: 读 CSR DER 失败\n");
            goto out;
        }
        pubkey = extract_spki_from_der(csr_der, der_len, &csr_subject);
        if (!pubkey || !csr_subject) {
            fprintf(stderr, "cfca_envelope: 从 CSR 提取公钥失败\n");
            goto out;
        }
    }

    /* 2. subject：--subject 优先，否则用 CSR 里的 */
    if (subject_str && *subject_str) {
        subject = parse_subject(subject_str);
        if (!subject) {
            fprintf(stderr, "cfca_envelope: 解析 --subject 失败: %s\n", subject_str);
            goto out;
        }
        fprintf(stderr, "cfca_envelope: sign-cert 用 --subject 覆盖 CSR subject\n");
    } else {
        subject = csr_subject;
        csr_subject = NULL;
    }

    /* 3. 加载 CA */
    { FILE *f = fopen(ca_cert_path, "rb"); if (!f) goto out;
      ca_cert = PEM_read_X509(f, NULL, NULL, NULL); fclose(f); if (!ca_cert) goto out; }
    ca_key = load_priv(ca_key_path);
    if (!ca_key) goto out;

    /* 4. 构造并签发 */
    cert = X509_new();
    if (!cert) goto out;
    X509_set_version(cert, 2);

    { BIGNUM *serial = BN_new(); if (!serial) goto out;
      BN_rand(serial, 128, BN_RAND_TOP_ANY, BN_RAND_BOTTOM_ANY);
      BN_to_ASN1_INTEGER(serial, X509_get_serialNumber(cert));
      BN_free(serial); }

    X509_set_subject_name(cert, subject);
    X509_set_issuer_name(cert, X509_get_subject_name(ca_cert));
    X509_set_pubkey(cert, pubkey);
    X509_gmtime_adj(X509_get_notBefore(cert), 0);
    X509_gmtime_adj(X509_get_notAfter(cert), (long)days * 86400L);

    {
        X509V3_CTX ctx;
        X509V3_set_ctx_nodb(&ctx);
        X509V3_set_ctx(&ctx, ca_cert, cert, NULL, NULL, 0);
        X509_EXTENSION *ext;

        const char *ku = is_enc
            ? "critical,keyEncipherment,dataEncipherment"
            : "critical,digitalSignature,nonRepudiation";
        ext = X509V3_EXT_conf_nid(NULL, &ctx, NID_key_usage, ku);
        if (!ext) goto out;
        X509_add_ext(cert, ext, -1);
        X509_EXTENSION_free(ext);

        ext = X509V3_EXT_conf_nid(NULL, &ctx, NID_ext_key_usage, "serverAuth");
        if (ext) { X509_add_ext(cert, ext, -1); X509_EXTENSION_free(ext); }

        ext = X509V3_EXT_conf_nid(NULL, &ctx, NID_subject_key_identifier, "hash");
        if (ext) { X509_add_ext(cert, ext, -1); X509_EXTENSION_free(ext); }

        ext = X509V3_EXT_conf_nid(NULL, &ctx, NID_authority_key_identifier, "keyid,issuer");
        if (ext) { X509_add_ext(cert, ext, -1); X509_EXTENSION_free(ext); }

        if (san_str && *san_str) {
            char san_buf[4096];
            char *o = san_buf;
            size_t remain = sizeof(san_buf);
            const char *pp = san_str;
            int first = 1;
            while (*pp && remain > 8) {
                const char *comma = strchr(pp, ',');
                size_t item_len = comma ? (size_t)(comma - pp) : strlen(pp);
                while (item_len > 0 && (*pp == ' ' || *pp == '\t')) { pp++; item_len--; }
                while (item_len > 0 && (pp[item_len-1] == ' ' || pp[item_len-1] == '\t')) item_len--;
                if (item_len > 0) {
                    int n = snprintf(o, remain, "%sDNS:%.*s",
                                     first ? "" : ",", (int)item_len, pp);
                    if (n > 0) { o += n; remain -= (size_t)n; first = 0; }
                }
                if (!comma) break;
                pp = comma + 1;
            }
            *o = 0;
            if (san_buf[0]) {
                ext = X509V3_EXT_conf_nid(NULL, &ctx, NID_subject_alt_name, san_buf);
                if (ext) { X509_add_ext(cert, ext, -1); X509_EXTENSION_free(ext); }
            }
        }
    }

    if (X509_sign(cert, ca_key, EVP_sm3()) <= 0) goto out;

    { FILE *f = fopen(out_path, "wb"); if (!f) goto out;
      int ok = PEM_write_X509(f, cert) == 1; fclose(f); if (!ok) goto out; }
    rc = 0;

out:
    if (cert) X509_free(cert);
    if (ca_key) EVP_PKEY_free(ca_key);
    if (ca_cert) X509_free(ca_cert);
    if (pubkey) EVP_PKEY_free(pubkey);
    if (subject) X509_NAME_free(subject);
    if (csr_subject) X509_NAME_free(csr_subject);
    if (csr_der) OPENSSL_free(csr_der);
    return rc;
}

int main(int argc, char **argv) {
    if (argc < 2) { usage(argv[0]); return 2; }
    if (!strcmp(argv[1], "gen-csr"))       return do_gen_csr(argc, argv) == 0 ? 0 : 1;
    if (!strcmp(argv[1], "gen-plain-csr")) return do_gen_plain_csr(argc, argv) == 0 ? 0 : 1;
    if (!strcmp(argv[1], "csr2pub"))       return do_csr2pub(argc, argv) == 0 ? 0 : 1;
    if (!strcmp(argv[1], "csr2signpub"))   return do_csr2signpub(argc, argv) == 0 ? 0 : 1;
    if (!strcmp(argv[1], "csr2subj"))      return do_csr2subj(argc, argv) == 0 ? 0 : 1;
    if (!strcmp(argv[1], "sign-cert"))     return do_sign_cert(argc, argv) == 0 ? 0 : 1;

    if (!strcmp(argv[1], "seal")) {
        const char *a = NULL, *b = NULL, *c = NULL;
        int cfca_format = 0;
        for (int i = 2; i < argc; i++) {
            if (!strcmp(argv[i], "--tmp-pub") && i + 1 < argc) a = argv[++i];
            else if (!strcmp(argv[i], "--enc-key") && i + 1 < argc) b = argv[++i];
            else if (!strcmp(argv[i], "--out") && i + 1 < argc) c = argv[++i];
            else if (!strcmp(argv[i], "--format") && i + 1 < argc)
                cfca_format = (strcmp(argv[++i], "cfca") == 0);
            else if (!strcmp(argv[i], "--cipher-order") && i + 1 < argc)
                { ++i; /* 忽略，简化格式固定 C1C3C2 */ }
            else { usage(argv[0]); return 2; }
        }
        if (!a || !b || !c) { usage(argv[0]); return 2; }
        return do_seal(a, b, c, cfca_format) == 0 ? 0 : 1;
    }

    if (!strcmp(argv[1], "unseal")) {
        const char *a = NULL, *b = NULL, *c = NULL;
        for (int i = 2; i < argc; i++) {
            if (!strcmp(argv[i], "--tmp-key") && i + 1 < argc) a = argv[++i];
            else if (!strcmp(argv[i], "--in") && i + 1 < argc) b = argv[++i];
            else if (!strcmp(argv[i], "--out") && i + 1 < argc) c = argv[++i];
            else { usage(argv[0]); return 2; }
        }
        if (!a || !b || !c) { usage(argv[0]); return 2; }
        return do_unseal(a, b, c) == 0 ? 0 : 1;
    }

    usage(argv[0]);
    return 2;
}
