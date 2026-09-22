/*
 * keycrypt.c - AES-256-GCM 私钥加解密工具
 *
 * 用法:
 *   keycrypt encrypt --master <hex-key-file> --in <plain>     --out <encrypted>
 *   keycrypt decrypt --master <hex-key-file> --in <encrypted> --out <plain>
 *
 * 主密钥文件内容为 64 个 hex 字符（32 字节），可含结尾换行。
 *
 * 输出文件格式:
 *   offset 0      : magic  "KC01"  (4 字节)
 *   offset 4      : IV              (12 字节)
 *   offset 16     : ciphertext      (N 字节)
 *   offset 16 + N : GCM tag         (16 字节)
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <openssl/evp.h>
#include <openssl/rand.h>

#define KC_MAGIC     "KC01"
#define KC_MAGIC_LEN 4
#define KC_IV_LEN    12
#define KC_TAG_LEN   16
#define KC_KEY_LEN   32

static void usage(const char *prog) {
    fprintf(stderr,
        "Usage:\n"
        "  %s encrypt --master <hex-key> --in <plain>     --out <encrypted>\n"
        "  %s decrypt --master <hex-key> --in <encrypted> --out <plain>\n",
        prog, prog);
}

static int read_file(const char *path, unsigned char **buf, size_t *len) {
    FILE *f = fopen(path, "rb");
    long n;
    unsigned char *p;
    if (!f) return -1;
    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return -1; }
    n = ftell(f);
    if (n < 0) { fclose(f); return -1; }
    if (fseek(f, 0, SEEK_SET) != 0) { fclose(f); return -1; }
    p = (unsigned char *)malloc((size_t)n + 1);
    if (!p) { fclose(f); return -1; }
    if (n > 0 && fread(p, 1, (size_t)n, f) != (size_t)n) {
        free(p); fclose(f); return -1;
    }
    p[n] = 0;
    fclose(f);
    *buf = p;
    *len = (size_t)n;
    return 0;
}

static int write_file(const char *path, const unsigned char *buf, size_t len) {
    FILE *f = fopen(path, "wb");
    if (!f) return -1;
    if (len > 0 && fwrite(buf, 1, len, f) != len) { fclose(f); return -1; }
    if (fclose(f) != 0) return -1;
    return 0;
}

static int parse_hex_key(const char *hex, size_t hexlen,
                         unsigned char *out, size_t outlen) {
    size_t i;
    while (hexlen > 0 && (hex[hexlen - 1] == '\n' ||
                          hex[hexlen - 1] == '\r' ||
                          hex[hexlen - 1] == ' '  ||
                          hex[hexlen - 1] == '\t')) {
        hexlen--;
    }
    if (hexlen != outlen * 2) return -1;
    for (i = 0; i < outlen; i++) {
        unsigned int b;
        if (sscanf(hex + i * 2, "%2x", &b) != 1) return -1;
        out[i] = (unsigned char)b;
    }
    return 0;
}

static int load_master_key(const char *path, unsigned char key[KC_KEY_LEN]) {
    unsigned char *buf = NULL;
    size_t len = 0;
    int rc;
    if (read_file(path, &buf, &len) != 0) {
        fprintf(stderr, "keycrypt: cannot read master key: %s\n", path);
        return -1;
    }
    rc = parse_hex_key((const char *)buf, len, key, KC_KEY_LEN);
    free(buf);
    if (rc != 0) {
        fprintf(stderr, "keycrypt: master key must be %d hex chars (32 bytes)\n",
                KC_KEY_LEN * 2);
        return -1;
    }
    return 0;
}

static int do_encrypt(const char *master_path,
                      const char *in_path, const char *out_path) {
    unsigned char key[KC_KEY_LEN];
    unsigned char iv[KC_IV_LEN];
    unsigned char tag[KC_TAG_LEN];
    unsigned char *plain = NULL, *cipher = NULL, *out = NULL;
    size_t plainlen = 0;
    int outl = 0, total = 0;
    size_t total_len;
    EVP_CIPHER_CTX *ctx = NULL;
    int rc = -1;

    if (load_master_key(master_path, key) != 0) return -1;
    if (read_file(in_path, &plain, &plainlen) != 0) {
        fprintf(stderr, "keycrypt: cannot read input: %s\n", in_path);
        return -1;
    }
    if (RAND_bytes(iv, KC_IV_LEN) != 1) {
        fprintf(stderr, "keycrypt: RAND_bytes failed\n");
        free(plain);
        return -1;
    }

    cipher = (unsigned char *)malloc(plainlen + 32);
    if (!cipher) { free(plain); return -1; }

    ctx = EVP_CIPHER_CTX_new();
    if (!ctx) { free(plain); free(cipher); return -1; }

    if (EVP_EncryptInit_ex(ctx, EVP_aes_256_gcm(), NULL, NULL, NULL) != 1) goto done;
    if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_IVLEN, KC_IV_LEN, NULL) != 1) goto done;
    if (EVP_EncryptInit_ex(ctx, NULL, NULL, key, iv) != 1) goto done;
    if (EVP_EncryptUpdate(ctx, cipher, &outl, plain, (int)plainlen) != 1) goto done;
    total = outl;
    if (EVP_EncryptFinal_ex(ctx, cipher + total, &outl) != 1) goto done;
    total += outl;
    if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_GET_TAG, KC_TAG_LEN, tag) != 1) goto done;

    total_len = KC_MAGIC_LEN + KC_IV_LEN + (size_t)total + KC_TAG_LEN;
    out = (unsigned char *)malloc(total_len);
    if (!out) goto done;
    memcpy(out, KC_MAGIC, KC_MAGIC_LEN);
    memcpy(out + KC_MAGIC_LEN, iv, KC_IV_LEN);
    memcpy(out + KC_MAGIC_LEN + KC_IV_LEN, cipher, (size_t)total);
    memcpy(out + KC_MAGIC_LEN + KC_IV_LEN + total, tag, KC_TAG_LEN);

    if (write_file(out_path, out, total_len) != 0) {
        fprintf(stderr, "keycrypt: cannot write output: %s\n", out_path);
        goto done;
    }
    rc = 0;

done:
    if (ctx) EVP_CIPHER_CTX_free(ctx);
    if (plain) free(plain);
    if (cipher) free(cipher);
    if (out) free(out);
    return rc;
}

static int do_decrypt(const char *master_path,
                      const char *in_path, const char *out_path) {
    unsigned char key[KC_KEY_LEN];
    unsigned char *in = NULL, *plain = NULL;
    size_t inlen = 0, ctlen;
    const unsigned char *iv, *ct, *tag;
    int outl = 0, total = 0;
    EVP_CIPHER_CTX *ctx = NULL;
    int rc = -1;

    if (load_master_key(master_path, key) != 0) return -1;
    if (read_file(in_path, &in, &inlen) != 0) {
        fprintf(stderr, "keycrypt: cannot read input: %s\n", in_path);
        return -1;
    }
    if (inlen < (size_t)(KC_MAGIC_LEN + KC_IV_LEN + KC_TAG_LEN)) {
        fprintf(stderr, "keycrypt: input too short\n");
        free(in);
        return -1;
    }
    if (memcmp(in, KC_MAGIC, KC_MAGIC_LEN) != 0) {
        fprintf(stderr, "keycrypt: bad magic, not a keycrypt file\n");
        free(in);
        return -1;
    }

    iv    = in + KC_MAGIC_LEN;
    ct    = in + KC_MAGIC_LEN + KC_IV_LEN;
    ctlen = inlen - KC_MAGIC_LEN - KC_IV_LEN - KC_TAG_LEN;
    tag   = in + KC_MAGIC_LEN + KC_IV_LEN + ctlen;

    plain = (unsigned char *)malloc(ctlen + 32);
    if (!plain) { free(in); return -1; }
    ctx = EVP_CIPHER_CTX_new();
    if (!ctx) { free(in); free(plain); return -1; }

    if (EVP_DecryptInit_ex(ctx, EVP_aes_256_gcm(), NULL, NULL, NULL) != 1) goto done;
    if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_IVLEN, KC_IV_LEN, NULL) != 1) goto done;
    if (EVP_DecryptInit_ex(ctx, NULL, NULL, key, iv) != 1) goto done;
    if (EVP_DecryptUpdate(ctx, plain, &outl, ct, (int)ctlen) != 1) goto done;
    total = outl;
    if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_TAG, KC_TAG_LEN, (void *)tag) != 1) goto done;
    if (EVP_DecryptFinal_ex(ctx, plain + total, &outl) != 1) {
        fprintf(stderr, "keycrypt: authentication failed (wrong key or tampered)\n");
        goto done;
    }
    total += outl;

    if (write_file(out_path, plain, (size_t)total) != 0) {
        fprintf(stderr, "keycrypt: cannot write output: %s\n", out_path);
        goto done;
    }
    rc = 0;

done:
    if (ctx) EVP_CIPHER_CTX_free(ctx);
    if (in) free(in);
    if (plain) free(plain);
    return rc;
}

int main(int argc, char **argv) {
    const char *mode, *master = NULL, *in = NULL, *out = NULL;
    int i;

    if (argc < 2) { usage(argv[0]); return 2; }
    mode = argv[1];

    for (i = 2; i < argc; i++) {
        if (strcmp(argv[i], "--master") == 0 && i + 1 < argc) master = argv[++i];
        else if (strcmp(argv[i], "--in")  == 0 && i + 1 < argc) in = argv[++i];
        else if (strcmp(argv[i], "--out") == 0 && i + 1 < argc) out = argv[++i];
        else { usage(argv[0]); return 2; }
    }
    if (!master || !in || !out) { usage(argv[0]); return 2; }

    if (strcmp(mode, "encrypt") == 0) return do_encrypt(master, in, out) == 0 ? 0 : 1;
    if (strcmp(mode, "decrypt") == 0) return do_decrypt(master, in, out) == 0 ? 0 : 1;

    usage(argv[0]);
    return 2;
}
