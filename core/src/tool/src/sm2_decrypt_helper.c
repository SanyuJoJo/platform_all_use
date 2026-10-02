/*
 * sm2_decrypt_helper.c - SM2 私钥解密辅助程序（诊断增强版）
 *
 * 用途：
 *   用 SM2 私钥解密 SM2 密文，输出明文。
 *   支持 ASN.1 (DER) 与裸 C1C3C2 两种密文格式。
 *
 * 关键诊断：
 *   - 加载私钥后打印曲线名与公钥头部，便于核对私钥来源；
 *   - 识别密文形态（ASN.1 vs 裸密文），走对应路径；
 *   - 失败时给出明确的人工提示。
 *
 * 用法：
 *   sm2_decrypt_helper --key <priv.pem> --in <cipher.bin> --out <plain.bin>
 *
 * 退出码：
 *   0  成功
 *   1  参数错误
 *   2  私钥加载失败
 *   3  解密失败
 *   4  文件 IO 失败
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/err.h>
#include <openssl/obj_mac.h>
#include <openssl/core_names.h>

#ifndef EVP_PKEY_SM2
#define EVP_PKEY_SM2 NID_sm2
#endif

/* -----------------------------------------------------------------------------
 * 工具
 * ------------------------------------------------------------------------- */

static void print_openssl_error(const char *stage) {
    fprintf(stderr, "sm2_decrypt_helper: %s failed\n", stage);
    ERR_print_errors_fp(stderr);
}

static void hexdump_head(const char *label,
                         const unsigned char *buf, size_t len, size_t max) {
    if (!buf || len == 0) return;
    size_t n = len < max ? len : max;
    fprintf(stderr, "sm2_decrypt_helper: %s (%zu bytes):", label, n);
    for (size_t i = 0; i < n; i++) fprintf(stderr, " %02x", buf[i]);
    fprintf(stderr, "\n");
}

/* 打印私钥的相关信息：曲线名 + 公钥点头部，便于核对。 */
static void dump_private_key_info(EVP_PKEY *pkey) {
    char curve[128] = {0};
    size_t clen = 0;
    if (EVP_PKEY_get_utf8_string_param(pkey,
            OSSL_PKEY_PARAM_GROUP_NAME, curve, sizeof(curve) - 1, &clen) > 0) {
        fprintf(stderr, "sm2_decrypt_helper: private key curve = %s\n", curve);
    } else {
        fprintf(stderr, "sm2_decrypt_helper: private key curve = (unknown)\n");
        ERR_clear_error();
    }

    unsigned char pub[256] = {0};
    size_t publen = 0;
    if (EVP_PKEY_get_octet_string_param(pkey,
            OSSL_PKEY_PARAM_PUB_KEY, pub, sizeof(pub), &publen) > 0) {
        fprintf(stderr, "sm2_decrypt_helper: public key point len = %zu\n", publen);
        hexdump_head("public key head", pub, publen, 32);
    } else {
        fprintf(stderr, "sm2_decrypt_helper: cannot extract public key\n");
        ERR_clear_error();
    }
}

static EVP_PKEY *load_private_key(const char *key_path) {
    FILE *fp = fopen(key_path, "r");
    if (!fp) {
        fprintf(stderr, "sm2_decrypt_helper: cannot open key file: %s\n", key_path);
        return NULL;
    }
    EVP_PKEY *pkey = PEM_read_PrivateKey(fp, NULL, NULL, NULL);
    fclose(fp);
    if (!pkey) {
        print_openssl_error("PEM_read_PrivateKey");
        return NULL;
    }

    /* 兼容以通用 EC 格式存储的 SM2 私钥 */
    if (EVP_PKEY_set_alias_type(pkey, EVP_PKEY_SM2) <= 0) {
        fprintf(stderr, "sm2_decrypt_helper: warning: set alias SM2 failed\n");
        ERR_clear_error();
    }
    return pkey;
}

static unsigned char *read_file(const char *path, size_t *len) {
    FILE *fp = fopen(path, "rb");
    if (!fp) {
        fprintf(stderr, "sm2_decrypt_helper: cannot open input: %s\n", path);
        return NULL;
    }
    if (fseek(fp, 0, SEEK_END) != 0) { fclose(fp); return NULL; }
    long fsize = ftell(fp);
    if (fsize < 0) { fclose(fp); return NULL; }
    if (fseek(fp, 0, SEEK_SET) != 0) { fclose(fp); return NULL; }

    unsigned char *buf = (unsigned char *)malloc((size_t)fsize + 1);
    if (!buf) { fclose(fp); return NULL; }
    if (fsize > 0 && fread(buf, 1, (size_t)fsize, fp) != (size_t)fsize) {
        free(buf); fclose(fp); return NULL;
    }
    fclose(fp);
    *len = (size_t)fsize;
    return buf;
}

static int write_file(const char *path, const unsigned char *buf, size_t len) {
    FILE *fp = fopen(path, "wb");
    if (!fp) return -1;
    if (len > 0 && fwrite(buf, 1, len, fp) != len) { fclose(fp); return -1; }
    if (fclose(fp) != 0) return -1;
    return 0;
}

static void usage(const char *prog) {
    fprintf(stderr,
        "Usage: %s --key <priv.pem> --in <cipher.bin> --out <plain.bin>\n",
        prog);
}

/* -----------------------------------------------------------------------------
 * 单次解密尝试
 * ------------------------------------------------------------------------- */
static int try_decrypt_once(
    EVP_PKEY *pkey,
    const unsigned char *cipher, size_t cipher_len,
    unsigned char **plain_out, size_t *plain_len_out,
    const char **err_stage
) {
    EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new(pkey, NULL);
    if (!ctx) { *err_stage = "EVP_PKEY_CTX_new"; return -1; }

    if (EVP_PKEY_decrypt_init(ctx) <= 0) {
        *err_stage = "EVP_PKEY_decrypt_init";
        EVP_PKEY_CTX_free(ctx);
        return -1;
    }

    size_t outlen = 0;
    if (EVP_PKEY_decrypt(ctx, NULL, &outlen, cipher, cipher_len) <= 0) {
        *err_stage = "EVP_PKEY_decrypt (size query)";
        EVP_PKEY_CTX_free(ctx);
        return -1;
    }

    unsigned char *plain = (unsigned char *)malloc(outlen + 256);
    if (!plain) {
        *err_stage = "malloc";
        EVP_PKEY_CTX_free(ctx);
        return -1;
    }

    size_t actual_len = outlen;
    if (EVP_PKEY_decrypt(ctx, plain, &actual_len, cipher, cipher_len) <= 0) {
        *err_stage = "EVP_PKEY_decrypt (actual)";
        free(plain);
        EVP_PKEY_CTX_free(ctx);
        return -1;
    }

    EVP_PKEY_CTX_free(ctx);
    *plain_out = plain;
    *plain_len_out = actual_len;
    return 0;
}

/* -----------------------------------------------------------------------------
 * 密文变体构造
 * ------------------------------------------------------------------------- */

/* 剥离 0x04 前缀（若存在） */
static unsigned char *variant_strip_04(const unsigned char *in, size_t inlen, size_t *outlen) {
    if (inlen == 0 || in[0] != 0x04) return NULL;
    size_t n = inlen - 1;
    unsigned char *buf = (unsigned char *)malloc(n > 0 ? n : 1);
    if (!buf) return NULL;
    if (n > 0) memcpy(buf, in + 1, n);
    *outlen = n;
    return buf;
}

/* 前置 0x04 前缀 */
static unsigned char *variant_add_04(const unsigned char *in, size_t inlen, size_t *outlen) {
    unsigned char *buf = (unsigned char *)malloc(inlen + 1);
    if (!buf) return NULL;
    buf[0] = 0x04;
    if (inlen > 0) memcpy(buf + 1, in, inlen);
    *outlen = inlen + 1;
    return buf;
}

/* C1C2C3 ⇄ C1C3C2 重排，C1 长度由参数指定 */
static unsigned char *variant_rearrange_c2c3(
    const unsigned char *in, size_t inlen, size_t c1_len
) {
    if (inlen < c1_len + 32) return NULL;
    size_t c2c3_len = inlen - c1_len;
    size_t c2_len = c2c3_len - 32;
    if (c2_len == 0 || c2_len > c2c3_len) return NULL;

    unsigned char *buf = (unsigned char *)malloc(inlen);
    if (!buf) return NULL;

    memcpy(buf, in, c1_len);
    memcpy(buf + c1_len, in + c1_len + c2_len, 32);
    memcpy(buf + c1_len + 32, in + c1_len, c2_len);
    return buf;
}

/* -----------------------------------------------------------------------------
 * 主流程
 * ------------------------------------------------------------------------- */

typedef struct {
    const char *name;
    unsigned char *buf;
    size_t len;
    int owned;
} variant_t;

int main(int argc, char **argv) {
    const char *key_path = NULL;
    const char *in_path  = NULL;
    const char *out_path = NULL;

    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--key") == 0 && i + 1 < argc) {
            key_path = argv[++i];
        } else if (strcmp(argv[i], "--in") == 0 && i + 1 < argc) {
            in_path = argv[++i];
        } else if (strcmp(argv[i], "--out") == 0 && i + 1 < argc) {
            out_path = argv[++i];
        } else if (strcmp(argv[i], "-h") == 0 || strcmp(argv[i], "--help") == 0) {
            usage(argv[0]);
            return 0;
        } else {
            fprintf(stderr, "sm2_decrypt_helper: unknown argument: %s\n", argv[i]);
            usage(argv[0]);
            return 1;
        }
    }

    if (!key_path || !in_path || !out_path) {
        usage(argv[0]);
        return 1;
    }

    /* 1. 加载私钥并打印诊断 */
    EVP_PKEY *pkey = load_private_key(key_path);
    if (!pkey) return 2;
    dump_private_key_info(pkey);

    /* 2. 读取密文 */
    size_t cipher_len = 0;
    unsigned char *cipher_buf = read_file(in_path, &cipher_len);
    if (!cipher_buf || cipher_len == 0) {
        fprintf(stderr, "sm2_decrypt_helper: empty or unreadable ciphertext\n");
        free(cipher_buf);
        EVP_PKEY_free(pkey);
        return 4;
    }
    hexdump_head("cipher head", cipher_buf, cipher_len, 32);

    /* 3. 判定密文形态 */
    int is_asn1 = (cipher_buf[0] == 0x30);
    fprintf(stderr, "sm2_decrypt_helper: ciphertext form = %s\n",
            is_asn1 ? "ASN.1(DER)" : "bare");

    /* 4. 构造变体列表 */
    variant_t variants[8];
    int n = 0;

    if (is_asn1) {
        /* ASN.1 形态：原样 + 剥离可能误加的前缀 */
        variants[n++] = (variant_t){"asn1-as-is", cipher_buf, cipher_len, 0};
    } else {
        /* 裸密文形态：原样 + 加/去 0x04 + C2C3 重排 */
        variants[n++] = (variant_t){"bare-as-is", cipher_buf, cipher_len, 0};

        if (cipher_buf[0] == 0x04) {
            size_t l = 0;
            unsigned char *b = variant_strip_04(cipher_buf, cipher_len, &l);
            if (b) variants[n++] = (variant_t){"bare-strip-04", b, l, 1};
        } else {
            size_t l = 0;
            unsigned char *b = variant_add_04(cipher_buf, cipher_len, &l);
            if (b) variants[n++] = (variant_t){"bare-add-04", b, l, 1};
        }

        {
            unsigned char *b = variant_rearrange_c2c3(cipher_buf, cipher_len, 64);
            if (b) variants[n++] = (variant_t){"bare-C2C3-swap", b, cipher_len, 1};
        }
        if (cipher_buf[0] == 0x04) {
            unsigned char *b = variant_rearrange_c2c3(cipher_buf, cipher_len, 65);
            if (b) variants[n++] = (variant_t){"bare-C2C3-swap-C1(65)", b, cipher_len, 1};
        }
    }

    /* 5. 逐一尝试 */
    unsigned char *plain = NULL;
    size_t plain_len = 0;
    int ok = 0;
    const char *last_stage = NULL;

    for (int i = 0; i < n; i++) {
        const char *stage = NULL;
        fprintf(stderr, "sm2_decrypt_helper: trying variant[%d] %s (len=%zu)\n",
                i, variants[i].name, variants[i].len);

        if (try_decrypt_once(pkey, variants[i].buf, variants[i].len,
                             &plain, &plain_len, &stage) == 0) {
            fprintf(stderr,
                "sm2_decrypt_helper: decrypt ok (variant=%s, plain_len=%zu)\n",
                variants[i].name, plain_len);
            ok = 1;
            break;
        }

        fprintf(stderr, "sm2_decrypt_helper: variant[%d] %s failed at %s\n",
                i, variants[i].name, stage ? stage : "unknown");
        ERR_print_errors_fp(stderr);
        ERR_clear_error();
        last_stage = stage;
    }

    /* 6. 清理变体缓冲 */
    for (int i = 0; i < n; i++) {
        if (variants[i].owned) free(variants[i].buf);
    }

    if (!ok) {
        fprintf(stderr, "sm2_decrypt_helper: all variants failed\n");
        if (is_asn1) {
            fprintf(stderr,
                "sm2_decrypt_helper: HINT: ciphertext is ASN.1 encoded; "
                "ASN.1 parsing succeeded, but SM2 internal decryption failed. "
                "The most likely cause is that the private key does NOT match "
                "the public key which was used to encrypt the symmetric key. "
                "Please verify that the private key is the SIGNING private key "
                "(not the encryption private key), and that it matches the "
                "signing certificate that was used when issuing the envelope.\n");
        }
        free(cipher_buf);
        EVP_PKEY_free(pkey);
        return 3;
    }

    /* 7. 写出明文 */
    if (write_file(out_path, plain, plain_len) != 0) {
        fprintf(stderr, "sm2_decrypt_helper: cannot write output: %s\n", out_path);
        free(plain);
        free(cipher_buf);
        EVP_PKEY_free(pkey);
        return 4;
    }

    free(plain);
    free(cipher_buf);
    EVP_PKEY_free(pkey);
    return 0;
}
