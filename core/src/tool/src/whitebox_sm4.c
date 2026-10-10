/*
 * whitebox_sm4.c
 * 白盒 SM4 加解密工具，依赖 Tongsuo/OpenSSL libcrypto。
 *
 * ★ 密钥集成到二进制内部：
 *   加密密钥由 WB_MASTER（内嵌常量）+ salt（密文头随机）派生。
 *   不生成、不读取任何外部 .pass 文件。
 *
 * 用法：
 *   whitebox_sm4 <file>          加密：file 就地加密
 *   whitebox_sm4 <file> <out>    解密：file → out
 *
 * 文件格式：
 *   magic "WBX1"(4) | salt(16) | iv(16) | ciphertext | hmac_sm3(32)
 *
 * 退出码：
 *   0  成功
 *   1  加解密失败（文件缺失、格式错误、HMAC 不匹配等）
 *   2  参数错误
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/sha.h>
#include <openssl/hmac.h>

#define MAGIC        "WBX1"
#define SALT_LEN     16
#define IV_LEN       16
#define HMAC_LEN     32
#define SM4_KEY_LEN  16

/*
 * WB_MASTER —— 白盒主密钥，硬编码在二进制内。
 *
 * ★ 生产部署提示：
 *   建议在编译前替换为随机值（例如 `openssl rand -hex 32` 的结果），
 *   并由构建系统在 CI 阶段注入，避免源码库泄露导致密钥暴露。
 */
static const unsigned char WB_MASTER[32] = {
    0x31,0x32,0x33,0x34,0x35,0x36,0x37,0x38,
    0x39,0x30,0x41,0x42,0x43,0x44,0x45,0x46,
    0x10,0x20,0x30,0x40,0x50,0x60,0x70,0x80,
    0x90,0xA0,0xB0,0xC0,0xD0,0xE0,0xF0,0x11
};

/* ------------------------------------------------------------------------- */
/* 帮助                                                                      */
/* ------------------------------------------------------------------------- */

static void print_usage(const char *prog) {
    fprintf(stderr,
        "usage: %s <file>          # 加密：file 就地加密\n"
        "       %s <file> <out>    # 解密：file → out\n",
        prog, prog);
}

/* ------------------------------------------------------------------------- */
/* 文件 I/O                                                                  */
/* ------------------------------------------------------------------------- */

static int read_file(const char *path, unsigned char **buf, size_t *len) {
    FILE *f;
    long n;
    unsigned char *p;

    if (!path) { fprintf(stderr, "error: path is NULL\n"); return -1; }
    f = fopen(path, "rb");
    if (!f) {
        fprintf(stderr, "error: cannot open '%s': %s\n", path, strerror(errno));
        return -1;
    }
    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return -1; }
    n = ftell(f);
    if (n < 0) { fclose(f); return -1; }
    if (fseek(f, 0, SEEK_SET) != 0) { fclose(f); return -1; }

    p = malloc((size_t)n + 1);
    if (!p) { fclose(f); return -1; }
    if (n > 0 && fread(p, 1, (size_t)n, f) != (size_t)n) {
        fprintf(stderr, "error: short read on '%s'\n", path);
        free(p); fclose(f); return -1;
    }
    p[n] = 0;
    *buf = p;
    *len = (size_t)n;
    fclose(f);
    return 0;
}

/* 原子写：先写 <path>.tmp.<pid>，再 rename。保留原权限。 */
static int write_file_atomic(const char *path, const unsigned char *buf,
                             size_t len, mode_t mode) {
    char tmppath[4096];
    FILE *f;
    int fd;
    struct stat st;
    mode_t final_mode = mode;

    if (snprintf(tmppath, sizeof(tmppath), "%s.tmp.%d",
                 path, (int)getpid()) >= (int)sizeof(tmppath)) {
        fprintf(stderr, "error: path too long: %s\n", path);
        return -1;
    }
    if (stat(path, &st) == 0) {
        final_mode = st.st_mode & 0777;
    }
    fd = open(tmppath, O_WRONLY | O_CREAT | O_TRUNC, final_mode);
    if (fd < 0) {
        fprintf(stderr, "error: cannot create '%s': %s\n",
                tmppath, strerror(errno));
        return -1;
    }
    f = fdopen(fd, "wb");
    if (!f) { close(fd); unlink(tmppath); return -1; }
    if (len && fwrite(buf, 1, len, f) != len) {
        fprintf(stderr, "error: short write on '%s'\n", tmppath);
        fclose(f); unlink(tmppath); return -1;
    }
    if (fclose(f) != 0) { unlink(tmppath); return -1; }
    if (rename(tmppath, path) != 0) {
        fprintf(stderr, "error: rename '%s' → '%s': %s\n",
                tmppath, path, strerror(errno));
        unlink(tmppath);
        return -1;
    }
    return 0;
}

/* ------------------------------------------------------------------------- */
/* KDF / 加解密原语                                                          */
/* ------------------------------------------------------------------------- */

/* ★ 密钥派生（无外部口令）：
 *   key_material = SHA256(WB_MASTER || salt)
 *   sm4key       = key_material[0:16]
 *   hmackey      = SHA256(key_material)  （32 字节）
 */
static void derive_key(const unsigned char *salt,
                       unsigned char sm4key[SM4_KEY_LEN],
                       unsigned char hmackey[32]) {
    unsigned char in[32 + SALT_LEN];
    unsigned char out[32];

    memcpy(in, WB_MASTER, 32);
    memcpy(in + 32, salt, SALT_LEN);
    SHA256(in, sizeof(in), out);
    memcpy(sm4key, out, SM4_KEY_LEN);
    SHA256(out, 32, hmackey);
}

static int sm4_cbc_crypt(int enc,
                         const unsigned char *key, const unsigned char *iv,
                         const unsigned char *in, int inlen,
                         unsigned char *out) {
    EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
    int outl = 0, total = 0;

    if (!ctx) return -1;
    if (enc) {
        if (EVP_EncryptInit_ex(ctx, EVP_sm4_cbc(), NULL, key, iv) != 1) goto err;
        if (EVP_EncryptUpdate(ctx, out, &outl, in, inlen) != 1) goto err;
        total = outl;
        if (EVP_EncryptFinal_ex(ctx, out + total, &outl) != 1) goto err;
        total += outl;
    } else {
        if (EVP_DecryptInit_ex(ctx, EVP_sm4_cbc(), NULL, key, iv) != 1) goto err;
        if (EVP_DecryptUpdate(ctx, out, &outl, in, inlen) != 1) goto err;
        total = outl;
        if (EVP_DecryptFinal_ex(ctx, out + total, &outl) != 1) goto err;
        total += outl;
    }
    EVP_CIPHER_CTX_free(ctx);
    return total;
err:
    EVP_CIPHER_CTX_free(ctx);
    return -1;
}

/* ------------------------------------------------------------------------- */
/* 加密（就地）                                                              */
/* ------------------------------------------------------------------------- */

static int do_encrypt_inplace(const char *path) {
    unsigned char *plain = NULL;
    size_t plainlen = 0;
    unsigned char salt[SALT_LEN], iv[IV_LEN];
    unsigned char sm4key[SM4_KEY_LEN], hmackey[32];
    unsigned char *cipher = NULL, *out = NULL;
    unsigned int hmaclen = 0;
    int cipherlen, rc;

    if (read_file(path, &plain, &plainlen) != 0) return -1;
    if (plainlen == 0) {
        fprintf(stderr, "error: '%s' is empty\n", path);
        free(plain);
        return -1;
    }

    /* 已经是白盒密文 → 拒绝重复加密 */
    if (plainlen >= 4 && memcmp(plain, MAGIC, 4) == 0) {
        fprintf(stderr, "error: '%s' already encrypted (magic 'WBX1')\n", path);
        free(plain);
        return -1;
    }

    if (RAND_bytes(salt, sizeof(salt)) != 1) { free(plain); return -1; }
    if (RAND_bytes(iv, sizeof(iv)) != 1)     { free(plain); return -1; }

    derive_key(salt, sm4key, hmackey);

    cipher = malloc(plainlen + 32);
    if (!cipher) { free(plain); return -1; }
    cipherlen = sm4_cbc_crypt(1, sm4key, iv, plain, (int)plainlen, cipher);
    if (cipherlen <= 0) { free(plain); free(cipher); return -1; }

    out = malloc(4 + SALT_LEN + IV_LEN + cipherlen + HMAC_LEN);
    if (!out) { free(plain); free(cipher); return -1; }

    memcpy(out, MAGIC, 4);
    memcpy(out + 4, salt, SALT_LEN);
    memcpy(out + 4 + SALT_LEN, iv, IV_LEN);
    memcpy(out + 4 + SALT_LEN + IV_LEN, cipher, cipherlen);

    if (HMAC(EVP_sm3(), hmackey, 32,
             out, 4 + SALT_LEN + IV_LEN + cipherlen,
             out + 4 + SALT_LEN + IV_LEN + cipherlen, &hmaclen) == NULL) {
        free(plain); free(cipher); free(out);
        return -1;
    }

    rc = write_file_atomic(path, out,
                           4 + SALT_LEN + IV_LEN + cipherlen + HMAC_LEN,
                           0600);

    free(plain);
    free(cipher);
    free(out);
    return rc;
}

/* ------------------------------------------------------------------------- */
/* 解密                                                                      */
/* ------------------------------------------------------------------------- */

static int do_decrypt(const char *inpath, const char *outpath) {
    unsigned char *in = NULL;
    size_t inlen = 0;
    unsigned char sm4key[SM4_KEY_LEN], hmackey[32];
    unsigned char calc[32];
    unsigned int hmaclen = 0;
    int cipherlen, plainlen;
    unsigned char *plain = NULL;
    int rc;

    if (read_file(inpath, &in, &inlen) != 0) return -1;

    if (inlen < 4 + SALT_LEN + IV_LEN + HMAC_LEN) {
        fprintf(stderr, "error: '%s' too short (%zu bytes)\n", inpath, inlen);
        free(in);
        return -1;
    }
    if (memcmp(in, MAGIC, 4) != 0) {
        fprintf(stderr, "error: '%s' has invalid magic (expected 'WBX1')\n", inpath);
        free(in);
        return -1;
    }

    derive_key(in + 4, sm4key, hmackey);

    cipherlen = (int)(inlen - 4 - SALT_LEN - IV_LEN - HMAC_LEN);

    if (HMAC(EVP_sm3(), hmackey, 32,
             in, 4 + SALT_LEN + IV_LEN + cipherlen,
             calc, &hmaclen) == NULL) {
        free(in);
        return -1;
    }
    if (memcmp(calc, in + 4 + SALT_LEN + IV_LEN + cipherlen, HMAC_LEN) != 0) {
        fprintf(stderr, "error: HMAC mismatch — file corrupted or wrong WB_MASTER\n");
        free(in);
        return -1;
    }

    plain = malloc(cipherlen + 32);
    if (!plain) { free(in); return -1; }

    plainlen = sm4_cbc_crypt(0, sm4key, in + 4 + SALT_LEN,
                             in + 4 + SALT_LEN + IV_LEN, cipherlen, plain);
    if (plainlen < 0) { free(in); free(plain); return -1; }

    rc = write_file_atomic(outpath, plain, (size_t)plainlen, 0600);

    free(in);
    free(plain);
    return rc;
}

/* ------------------------------------------------------------------------- */
/* main                                                                      */
/* ------------------------------------------------------------------------- */

int main(int argc, char **argv) {
    if (argc == 2) {
        return do_encrypt_inplace(argv[1]) == 0 ? 0 : 1;
    }
    if (argc == 3) {
        return do_decrypt(argv[1], argv[2]) == 0 ? 0 : 1;
    }
    fprintf(stderr, "error: wrong number of arguments (got %d)\n", argc - 1);
    print_usage(argv[0]);
    return 2;
}
