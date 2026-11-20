/*
 * Copyright 2019-2021 The OpenSSL Project Authors. All Rights Reserved.
 *
 * Licensed under the Apache License 2.0 (the "License").  You may not use
 * this file except in compliance with the License.  You can obtain a copy
 * in the file LICENSE in the source distribution or at
 * https://www.openssl.org/source/license.html
 */

#include <openssl/aes.h>
#include "prov/ciphercommon.h"
#include "prov/ciphercommon_gcm.h"
#include "crypto/aes_platform.h"

typedef struct prov_aes_gcm_ctx_st {
    PROV_GCM_CTX base;          /* must be first entry in struct */
    union {
        OSSL_UNION_ALIGN;
        AES_KEY ks;
    } ks;                       /* AES key schedule to use */

    /* Platform specific data */
    union {
        int dummy;
#if defined(OPENSSL_CPUID_OBJ) && defined(__s390__)
        struct {
            union {
                OSSL_UNION_ALIGN;
                S390X_KMA_PARAMS kma;
            } param;
            unsigned int fc;
            unsigned int hsflag;    /* hash subkey set flag */
            unsigned char ares[16];
            unsigned char mres[16];
            unsigned char kres[16];
            int areslen;
            int mreslen;
            int kreslen;
            int res;
        } s390x;
#endif /* defined(OPENSSL_CPUID_OBJ) && defined(__s390__) */
    } plat;
} PROV_AES_GCM_CTX;

const PROV_GCM_HW *ossl_prov_aes_hw_gcm(size_t keybits);

/*
 * See crypto/modes/asm/aes-gcm-avx512.pl for further details.
 */
void ossl_aes_gcm_encrypt_avx512 (const void* aes_keys, 
                                  void *gcm128ctx,
                                  unsigned int *pblocklen,
                                  const unsigned char *in,
                                  size_t len,
                                  unsigned char *out);

void ossl_aes_gcm_decrypt_avx512 (const void* keys,
                                  void *gcm128ctx,
                                  unsigned int *pblocklen,
                                  const unsigned char *in,
                                  size_t len,
                                  unsigned char *out);

void ossl_aes_gcm_init_avx512(const void *ks, void *gcm128ctx);
void ossl_aes_gcm_setiv_avx512(const void *ks, void *gcm128ctx,
                               const unsigned char *iv, size_t ivlen);
void ossl_aes_gcm_update_aad_avx512(void *gcm128ctx, const unsigned char *aad,
                                    size_t aadlen);
void ossl_aes_gcm_finalize_avx512(void *gcm128ctx, unsigned int pblocklen);

void ossl_gcm_gmult_avx512(u64 Xi[2], const void *gcm128ctx);

/* Returns non-zero when AVX512F + VAES + VPCLMULDQD combination is available */
int ossl_vaes_vpclmulqdq_capable(void);
