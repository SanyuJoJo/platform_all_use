#! /usr/bin/env perl
# -*- mode: perl -*-

package configdata;

use strict;
use warnings;

use Exporter;
our @ISA = qw(Exporter);
our @EXPORT = qw(
    %config %target %disabled %withargs %unified_info
    @disablables @disablables_int
);

our %config = (
    "AR" => "ar",
    "ARFLAGS" => [
        "qc"
    ],
    "ASFLAGS" => [],
    "CC" => "gcc",
    "CFLAGS" => [
        "-Wall -O3"
    ],
    "CPPDEFINES" => [],
    "CPPFLAGS" => [],
    "CPPINCLUDES" => [],
    "CXX" => "g++",
    "CXXFLAGS" => [
        "-Wall -O3"
    ],
    "FIPSKEY" => "f4556650ac31d35461610bac4ed81b1a181b2d8a43ea2854cbae22ca74560813",
    "FIPS_VENDOR" => "OpenSSL non-compliant FIPS Provider",
    "HASHBANGPERL" => "/usr/bin/env perl",
    "LDFLAGS" => [],
    "LDLIBS" => [],
    "OBJCOPY" => "objcopy",
    "PERL" => "/usr/bin/perl",
    "RANLIB" => "ranlib",
    "RC" => "windres",
    "RCFLAGS" => [],
    "SMTCPASSWD" => "Tongsuo123",
    "SMTCPUBKEY" => "-----BEGIN PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAERjiZ5ubxrnOZnjhvqvuJ5UcdRI64\\nsBEVwF0UztQK9eYzqOsFEm0PKkCjoYkdmiZ+Und0agHk94eFKhtUYsu0bw==\\n-----END PUBLIC KEY-----",
    "afalgeng" => "",
    "api" => "30500",
    "b32" => "0",
    "b64" => "0",
    "b64l" => "1",
    "bn_ll" => "0",
    "build_file" => "Makefile",
    "build_file_templates" => [
        "Configurations/common0.tmpl",
        "Configurations/unix-Makefile.tmpl"
    ],
    "build_infos" => [
        "./build.info",
        "crypto/build.info",
        "ssl/build.info",
        "apps/build.info",
        "util/build.info",
        "fuzz/build.info",
        "providers/build.info",
        "engines/build.info",
        "exporters/build.info",
        "crypto/objects/build.info",
        "crypto/buffer/build.info",
        "crypto/bio/build.info",
        "crypto/stack/build.info",
        "crypto/lhash/build.info",
        "crypto/hashtable/build.info",
        "crypto/rand/build.info",
        "crypto/evp/build.info",
        "crypto/asn1/build.info",
        "crypto/pem/build.info",
        "crypto/x509/build.info",
        "crypto/conf/build.info",
        "crypto/txt_db/build.info",
        "crypto/pkcs7/build.info",
        "crypto/pkcs12/build.info",
        "crypto/ui/build.info",
        "crypto/kdf/build.info",
        "crypto/store/build.info",
        "crypto/property/build.info",
        "crypto/md5/build.info",
        "crypto/sha/build.info",
        "crypto/ml_kem/build.info",
        "crypto/hmac/build.info",
        "crypto/poly1305/build.info",
        "crypto/eia3/build.info",
        "crypto/siphash/build.info",
        "crypto/sm3/build.info",
        "crypto/des/build.info",
        "crypto/aes/build.info",
        "crypto/rc4/build.info",
        "crypto/zuc/build.info",
        "crypto/sm4/build.info",
        "crypto/chacha/build.info",
        "crypto/modes/build.info",
        "crypto/bn/build.info",
        "crypto/ec/build.info",
        "crypto/rsa/build.info",
        "crypto/dsa/build.info",
        "crypto/dh/build.info",
        "crypto/sm2/build.info",
        "crypto/dso/build.info",
        "crypto/engine/build.info",
        "crypto/err/build.info",
        "crypto/comp/build.info",
        "crypto/http/build.info",
        "crypto/ocsp/build.info",
        "crypto/cms/build.info",
        "crypto/ts/build.info",
        "crypto/srp/build.info",
        "crypto/cmac/build.info",
        "crypto/ct/build.info",
        "crypto/async/build.info",
        "crypto/ess/build.info",
        "crypto/crmf/build.info",
        "crypto/cmp/build.info",
        "crypto/encode_decode/build.info",
        "crypto/sdf/build.info",
        "crypto/tsapi/build.info",
        "crypto/zkp/build.info",
        "crypto/ffc/build.info",
        "crypto/hpke/build.info",
        "crypto/thread/build.info",
        "crypto/ml_dsa/build.info",
        "crypto/slh_dsa/build.info",
        "ssl/record/build.info",
        "ssl/rio/build.info",
        "ssl/quic/build.info",
        "apps/lib/build.info",
        "providers/common/build.info",
        "providers/implementations/build.info",
        "ssl/record/methods/build.info",
        "providers/common/der/build.info",
        "providers/implementations/digests/build.info",
        "providers/implementations/ciphers/build.info",
        "providers/implementations/rands/build.info",
        "providers/implementations/macs/build.info",
        "providers/implementations/kdfs/build.info",
        "providers/implementations/exchange/build.info",
        "providers/implementations/keymgmt/build.info",
        "providers/implementations/signature/build.info",
        "providers/implementations/asymciphers/build.info",
        "providers/implementations/encode_decode/build.info",
        "providers/implementations/storemgmt/build.info",
        "providers/implementations/kem/build.info",
        "providers/implementations/skeymgmt/build.info",
        "providers/implementations/rands/seeding/build.info"
    ],
    "build_metadata" => "",
    "build_type" => "release",
    "builddir" => ".",
    "cflags" => [
        "-Wa,--noexecstack"
    ],
    "conf_files" => [
        "Configurations/00-base-templates.conf",
        "Configurations/10-main.conf"
    ],
    "cppflags" => [],
    "cxxflags" => [],
    "defines" => [
        "NDEBUG"
    ],
    "dynamic_engines" => "1",
    "engdirs" => [
        "afalg"
    ],
    "ex_libs" => [],
    "full_version" => "3.5.4",
    "includes" => [],
    "lflags" => [],
    "lib_defines" => [
        "OPENSSL_PIC"
    ],
    "libdir" => "",
    "major" => "3",
    "makedep_scheme" => "gcc",
    "minor" => "5",
    "openssl_api_defines" => [
        "OPENSSL_CONFIGURED_API=30500"
    ],
    "openssl_feature_defines" => [
        "OPENSSL_RAND_SEED_OS",
        "OPENSSL_THREADS",
        "OPENSSL_NO_ACVP_TESTS",
        "OPENSSL_NO_ARGON2",
        "OPENSSL_NO_ARIA",
        "OPENSSL_NO_ASAN",
        "OPENSSL_NO_ATF_SLIBCE",
        "OPENSSL_NO_BF",
        "OPENSSL_NO_BLAKE2",
        "OPENSSL_NO_BN_METHOD",
        "OPENSSL_NO_BROTLI",
        "OPENSSL_NO_BROTLI_DYNAMIC",
        "OPENSSL_NO_BULLETPROOFS",
        "OPENSSL_NO_CAMELLIA",
        "OPENSSL_NO_CAST",
        "OPENSSL_NO_CRYPTO_MDEBUG",
        "OPENSSL_NO_CRYPTO_MDEBUG_BACKTRACE",
        "OPENSSL_NO_CRYPTO_MDEBUG_COUNT",
        "OPENSSL_NO_DELEGATED_CREDENTIAL",
        "OPENSSL_NO_DEMOS",
        "OPENSSL_NO_DEVCRYPTOENG",
        "OPENSSL_NO_EC_ELGAMAL",
        "OPENSSL_NO_EC_NISTP_64_GCC_128",
        "OPENSSL_NO_EC_SM2P_64_GCC_128",
        "OPENSSL_NO_EGD",
        "OPENSSL_NO_EVP_CIPHER_API_COMPAT",
        "OPENSSL_NO_EXTERNAL_TESTS",
        "OPENSSL_NO_FIPS_JITTER",
        "OPENSSL_NO_FIPS_POST",
        "OPENSSL_NO_FIPS_SECURITYCHECKS",
        "OPENSSL_NO_FUZZ_AFL",
        "OPENSSL_NO_FUZZ_LIBFUZZER",
        "OPENSSL_NO_GOST",
        "OPENSSL_NO_H3DEMO",
        "OPENSSL_NO_HQINTEROP",
        "OPENSSL_NO_IDEA",
        "OPENSSL_NO_JITTER",
        "OPENSSL_NO_KTLS",
        "OPENSSL_NO_MD2",
        "OPENSSL_NO_MD4",
        "OPENSSL_NO_MDC2",
        "OPENSSL_NO_MSAN",
        "OPENSSL_NO_NIZK",
        "OPENSSL_NO_OPTIMIZE_CHACHA_CHOOSE",
        "OPENSSL_NO_PAILLIER",
        "OPENSSL_NO_PIE",
        "OPENSSL_NO_RC2",
        "OPENSSL_NO_RC5",
        "OPENSSL_NO_RIPEMD",
        "OPENSSL_NO_RMD160",
        "OPENSSL_NO_SCTP",
        "OPENSSL_NO_SDF_LIB",
        "OPENSSL_NO_SDF_LIB_DYNAMIC",
        "OPENSSL_NO_SEED",
        "OPENSSL_NO_SM2_THRESHOLD",
        "OPENSSL_NO_SMTC",
        "OPENSSL_NO_SMTC_DEBUG",
        "OPENSSL_NO_SSL3",
        "OPENSSL_NO_SSL3_METHOD",
        "OPENSSL_NO_SSLKEYLOG",
        "OPENSSL_NO_STATUS",
        "OPENSSL_NO_TESTS",
        "OPENSSL_NO_TFO",
        "OPENSSL_NO_TRACE",
        "OPENSSL_NO_TWISTED_EC_ELGAMAL",
        "OPENSSL_NO_UBSAN",
        "OPENSSL_NO_UNIT_TEST",
        "OPENSSL_NO_UPLINK",
        "OPENSSL_NO_WBSM4_BAIWU",
        "OPENSSL_NO_WBSM4_WSISE",
        "OPENSSL_NO_WBSM4_XIAOLAI",
        "OPENSSL_NO_WHIRLPOOL",
        "OPENSSL_NO_WINSTORE",
        "OPENSSL_NO_ZKP_GADGET",
        "OPENSSL_NO_ZKP_TRANSCRIPT",
        "OPENSSL_NO_ZLIB",
        "OPENSSL_NO_ZLIB_DYNAMIC",
        "OPENSSL_NO_ZSTD",
        "OPENSSL_NO_ZSTD_DYNAMIC",
        "OPENSSL_NO_STATIC_ENGINE"
    ],
    "openssl_other_defines" => [
        "OPENSSL_NO_KTLS"
    ],
    "openssl_sys_defines" => [],
    "openssldir" => "/opt/plats_tool/core/libs/tongsuo/ssl",
    "options" => "--prefix=/opt/plats_tool/core/libs/tongsuo --openssldir=/opt/plats_tool/core/libs/tongsuo/ssl enable-ntls enable-weak-ssl-ciphers enable-shared no-acvp-tests no-argon2 no-aria no-asan no-atf_slibce no-bf no-blake2 no-bn-method no-brotli no-brotli-dynamic no-buildtest-c++ no-bulletproofs no-camellia no-cast no-crypto-mdebug no-crypto-mdebug-backtrace no-crypto-mdebug-count no-delegated-credential no-demos no-devcryptoeng no-ec_elgamal no-ec_nistp_64_gcc_128 no-ec_sm2p_64_gcc_128 no-egd no-evp-cipher-api-compat no-external-tests no-fips no-fips-jitter no-fips-post no-fips-securitychecks no-fuzz-afl no-fuzz-libfuzzer no-gost no-h3demo no-hqinterop no-idea no-jitter no-ktls no-md2 no-md4 no-mdc2 no-msan no-nizk no-optimize-chacha-choose no-paillier no-pie no-rc2 no-rc5 no-ripemd no-rmd160 no-sctp no-sdf-lib no-sdf-lib-dynamic no-seed no-sm2_threshold no-smtc no-smtc-debug no-ssl3 no-ssl3-method no-sslkeylog no-status no-tests no-tfo no-trace no-twisted_ec_elgamal no-ubsan no-unit-test no-uplink no-wbsm4-baiwu no-wbsm4-wsise no-wbsm4-xiaolai no-whirlpool no-winstore no-zkp-gadget no-zkp-transcript no-zlib no-zlib-dynamic no-zstd no-zstd-dynamic",
    "patch" => "4",
    "perl_archname" => "x86_64-linux-gnu-thread-multi",
    "perl_cmd" => "/usr/bin/perl",
    "perl_version" => "5.30.0",
    "perlargv" => [
        "--prefix=/opt/plats_tool/core/libs/tongsuo",
        "--openssldir=/opt/plats_tool/core/libs/tongsuo/ssl",
        "enable-ntls",
        "enable-weak-ssl-ciphers",
        "no-tests",
        "shared"
    ],
    "perlenv" => {
        "AR" => undef,
        "ARFLAGS" => undef,
        "AS" => undef,
        "ASFLAGS" => undef,
        "BUILDFILE" => undef,
        "CC" => undef,
        "CFLAGS" => undef,
        "CPP" => undef,
        "CPPDEFINES" => undef,
        "CPPFLAGS" => undef,
        "CPPINCLUDES" => undef,
        "CROSS_COMPILE" => undef,
        "CXX" => undef,
        "CXXFLAGS" => undef,
        "HASHBANGPERL" => undef,
        "LD" => undef,
        "LDFLAGS" => undef,
        "LDLIBS" => undef,
        "MT" => undef,
        "MTFLAGS" => undef,
        "OBJCOPY" => undef,
        "OPENSSL_LOCAL_CONFIG_DIR" => undef,
        "PERL" => undef,
        "RANLIB" => undef,
        "RC" => undef,
        "RCFLAGS" => undef,
        "RM" => undef,
        "WINDRES" => undef,
        "__CNF_CFLAGS" => undef,
        "__CNF_CPPDEFINES" => undef,
        "__CNF_CPPFLAGS" => undef,
        "__CNF_CPPINCLUDES" => undef,
        "__CNF_CXXFLAGS" => undef,
        "__CNF_LDFLAGS" => undef,
        "__CNF_LDLIBS" => undef
    },
    "prefix" => "/opt/plats_tool/core/libs/tongsuo",
    "prerelease" => "",
    "processor" => "",
    "rc4_int" => "unsigned int",
    "release_date" => "3 Aug 2026",
    "shlib_version" => "3",
    "sourcedir" => ".",
    "symbol_prefix" => "",
    "system_ciphers_file" => "",
    "target" => "linux-x86_64",
    "tongsuo_full_version" => "8.5.0-pre2",
    "tongsuo_major" => "8",
    "tongsuo_minor" => "5",
    "tongsuo_patch" => "0",
    "tongsuo_prerelease" => "-pre2",
    "tongsuo_version" => "8.5.0",
    "version" => "3.5.4"
);
our %target = (
    "AR" => "ar",
    "ARFLAGS" => "qc",
    "CC" => "gcc",
    "CFLAGS" => "-Wall -O3",
    "CXX" => "g++",
    "CXXFLAGS" => "-Wall -O3",
    "HASHBANGPERL" => "/usr/bin/env perl",
    "OBJCOPY" => "objcopy",
    "RANLIB" => "ranlib",
    "RC" => "windres",
    "_conf_fname_int" => [
        "Configurations/00-base-templates.conf",
        "Configurations/00-base-templates.conf",
        "Configurations/10-main.conf",
        "Configurations/10-main.conf",
        "Configurations/10-main.conf",
        "Configurations/shared-info.pl"
    ],
    "asm_arch" => "x86_64",
    "bin_cflags" => "",
    "bin_lflags" => "",
    "bn_ops" => "SIXTY_FOUR_BIT_LONG",
    "build_file" => "Makefile",
    "build_scheme" => [
        "unified",
        "unix"
    ],
    "cflags" => "-pthread -m64",
    "cppflags" => "",
    "cxxflags" => "-std=c++11 -pthread -m64",
    "defines" => [
        "OPENSSL_BUILDING_OPENSSL"
    ],
    "disable" => [],
    "dso_ldflags" => "-Wl,-z,defs",
    "dso_scheme" => "dlfcn",
    "enable" => [
        "afalgeng"
    ],
    "ex_libs" => "-ldl -pthread",
    "includes" => [],
    "lflags" => "",
    "lib_cflags" => "",
    "lib_cppflags" => "-DOPENSSL_USE_NODELETE -DL_ENDIAN",
    "lib_defines" => [],
    "module_cflags" => "-fPIC",
    "module_cxxflags" => undef,
    "module_ldflags" => "-Wl,-znodelete -shared -Wl,-Bsymbolic",
    "multilib" => "64",
    "perl_platform" => "Unix",
    "perlasm_scheme" => "elf",
    "shared_cflag" => "-fPIC",
    "shared_defflag" => "-Wl,--version-script=",
    "shared_defines" => [],
    "shared_ldflag" => "-Wl,-znodelete -shared -Wl,-Bsymbolic",
    "shared_rcflag" => "",
    "shared_sonameflag" => "-Wl,-soname=",
    "shared_target" => "linux-shared",
    "template" => "1",
    "thread_defines" => [],
    "thread_scheme" => "pthreads",
    "unistd" => "<unistd.h>"
);
our @disablables = (
    "acvp-tests",
    "afalgeng",
    "atf_slibce",
    "apps",
    "asan",
    "asm",
    "async",
    "atexit",
    "autoalginit",
    "autoerrinit",
    "autoload-config",
    "brotli",
    "brotli-dynamic",
    "buildtest-c++",
    "bulk",
    "cached-fetch",
    "capieng",
    "winstore",
    "chacha",
    "cmac",
    "cmp",
    "cms",
    "comp",
    "crypto-mdebug",
    "ct",
    "delegated-credential",
    "default-thread-pool",
    "demos",
    "h3demo",
    "hqinterop",
    "deprecated",
    "des",
    "devcryptoeng",
    "dgram",
    "dh",
    "dsa",
    "dso",
    "dtls",
    "dynamic-engine",
    "ec",
    "ec2m",
    "ec_nistp_64_gcc_128",
    "ec_sm2p_64_gcc_128",
    "ec_elgamal",
    "twisted_ec_elgamal",
    "ecdh",
    "ecdsa",
    "ecx",
    "egd",
    "engine",
    "err",
    "external-tests",
    "filenames",
    "fips",
    "fips-securitychecks",
    "fips-post",
    "fips-jitter",
    "fuzz-afl",
    "fuzz-libfuzzer",
    "http",
    "integrity-only-ciphers",
    "jitter",
    "ktls",
    "legacy",
    "loadereng",
    "makedepend",
    "ml-dsa",
    "ml-kem",
    "module",
    "msan",
    "multiblock",
    "nextprotoneg",
    "ntls",
    "ocb",
    "ocsp",
    "padlockeng",
    "pic",
    "pie",
    "pinshared",
    "poly1305",
    "posix-io",
    "psk",
    "quic",
    "unstable-qlog",
    "rc4",
    "rc5",
    "rdrand",
    "rfc3779",
    "smtc",
    "smtc-debug",
    "scrypt",
    "sctp",
    "sdf-lib",
    "sdf-lib-dynamic",
    "secure-memory",
    "shared",
    "siphash",
    "siv",
    "slh-dsa",
    "sm2",
    "sm2-precomp",
    "sm2_threshold",
    "sm3",
    "sm4",
    "wbsm4-xiaolai",
    "wbsm4-baiwu",
    "wbsm4-wsise",
    "zuc",
    "sock",
    "srp",
    "srtp",
    "sse2",
    "ssl",
    "ssl-trace",
    "static-engine",
    "stdio",
    "sslkeylog",
    "tests",
    "tfo",
    "thread-pool",
    "threads",
    "tls",
    "tls-deprecated-ec",
    "trace",
    "ts",
    "ubsan",
    "ui-console",
    "unit-test",
    "uplink",
    "weak-ssl-ciphers",
    "zlib",
    "zlib-dynamic",
    "zstd",
    "zstd-dynamic",
    "skip-scsv",
    "session-lookup",
    "dynamic-ciphers",
    "verify-sni",
    "rsa-multi-prime-key-compat",
    "session-reused-type",
    "optimize-chacha-choose",
    "status",
    "evp-cipher-api-compat",
    "crypto-mdebug-count",
    "paillier",
    "bulletproofs",
    "nizk",
    "zkp-gadget",
    "zkp-transcript",
    "bn-method",
    "ssl3",
    "ssl3-method",
    "tls1",
    "tls1-method",
    "tls1_1",
    "tls1_1-method",
    "tls1_2",
    "tls1_2-method",
    "tls1_3",
    "dtls1",
    "dtls1-method",
    "dtls1_2",
    "dtls1_2-method"
);
our @disablables_int = (
    "crmf"
);
our %disabled = (
    "acvp-tests" => "cascade",
    "argon2" => "default",
    "aria" => "default",
    "asan" => "default",
    "atf_slibce" => "default",
    "bf" => "default",
    "blake2" => "default",
    "bn-method" => "default",
    "brotli" => "default",
    "brotli-dynamic" => "default",
    "buildtest-c++" => "default",
    "bulletproofs" => "default",
    "camellia" => "default",
    "cast" => "default",
    "crypto-mdebug" => "default",
    "crypto-mdebug-backtrace" => "default",
    "crypto-mdebug-count" => "default",
    "delegated-credential" => "default",
    "demos" => "default",
    "devcryptoeng" => "default",
    "ec_elgamal" => "default",
    "ec_nistp_64_gcc_128" => "default",
    "ec_sm2p_64_gcc_128" => "default",
    "egd" => "default",
    "evp-cipher-api-compat" => "default",
    "external-tests" => "default",
    "fips" => "default",
    "fips-jitter" => "default",
    "fips-post" => "cascade",
    "fips-securitychecks" => "cascade",
    "fuzz-afl" => "default",
    "fuzz-libfuzzer" => "default",
    "gost" => "default",
    "h3demo" => "default",
    "hqinterop" => "default",
    "idea" => "default",
    "jitter" => "default",
    "ktls" => "default",
    "md2" => "default",
    "md4" => "default",
    "mdc2" => "default",
    "msan" => "default",
    "nizk" => "default",
    "optimize-chacha-choose" => "default",
    "paillier" => "default",
    "pie" => "default",
    "rc2" => "default",
    "rc5" => "default",
    "ripemd" => "default",
    "rmd160" => "default",
    "sctp" => "default",
    "sdf-lib" => "default",
    "sdf-lib-dynamic" => "default",
    "seed" => "default",
    "sm2_threshold" => "default",
    "smtc" => "default",
    "smtc-debug" => "default",
    "ssl3" => "default",
    "ssl3-method" => "default",
    "sslkeylog" => "default",
    "status" => "default",
    "tests" => "option",
    "tfo" => "default",
    "trace" => "default",
    "twisted_ec_elgamal" => "default",
    "ubsan" => "default",
    "unit-test" => "default",
    "uplink" => "no uplink_arch",
    "wbsm4-baiwu" => "default",
    "wbsm4-wsise" => "default",
    "wbsm4-xiaolai" => "default",
    "whirlpool" => "default",
    "winstore" => "not-windows",
    "zkp-gadget" => "default",
    "zkp-transcript" => "default",
    "zlib" => "default",
    "zlib-dynamic" => "default",
    "zstd" => "default",
    "zstd-dynamic" => "default"
);
our %withargs = ();
our %unified_info = (
    "attributes" => {
        "depends" => {
            "providers/libcommon.a" => {
                "libcrypto" => {
                    "weak" => "1"
                }
            }
        },
        "generate" => {
            "exporters/OpenSSLConfig.cmake" => {
                "exporter" => "cmake"
            },
            "exporters/OpenSSLConfigVersion.cmake" => {
                "exporter" => "cmake"
            },
            "exporters/libcrypto.pc" => {
                "exporter" => "pkg-config"
            },
            "exporters/libssl.pc" => {
                "exporter" => "pkg-config"
            },
            "exporters/openssl.pc" => {
                "exporter" => "pkg-config"
            },
            "include/openssl/configuration.h" => {
                "skip" => "1"
            }
        },
        "libraries" => {
            "apps/libapps.a" => {
                "noinst" => "1"
            },
            "providers/libcommon.a" => {
                "noinst" => "1"
            },
            "providers/libdefault.a" => {
                "noinst" => "1"
            },
            "providers/liblegacy.a" => {
                "noinst" => "1"
            },
            "providers/libtemplate.a" => {
                "noinst" => "1"
            }
        },
        "modules" => {
            "engines/afalg" => {
                "engine" => "1"
            },
            "engines/capi" => {
                "engine" => "1"
            },
            "engines/dasync" => {
                "engine" => "1",
                "noinst" => "1"
            },
            "engines/ecptest" => {
                "engine" => "1"
            },
            "engines/loader_attic" => {
                "engine" => "1"
            },
            "engines/ossltest" => {
                "engine" => "1",
                "noinst" => "1"
            },
            "engines/padlock" => {
                "engine" => "1"
            }
        },
        "scripts" => {
            "apps/CA.pl" => {
                "misc" => "1"
            },
            "apps/tsget.pl" => {
                "linkname" => "tsget",
                "misc" => "1"
            },
            "util/shlib_wrap.sh" => {
                "noinst" => "1"
            },
            "util/wrap.pl" => {
                "noinst" => "1"
            }
        },
        "sources" => {
            "apps/openssl" => {
                "apps/openssl-bin-progs.o" => {
                    "nocheck" => "1"
                }
            },
            "apps/openssl-bin-progs.o" => {
                "apps/progs.c" => {
                    "nocheck" => "1"
                }
            },
            "apps/progs.o" => {}
        }
    },
    "defines" => {
        "engines/loader_attic" => [
            "OPENSSL_NO_PROVIDER_CODE"
        ],
        "engines/padlock" => [
            "PADLOCK_ASM"
        ],
        "libcrypto" => [
            "AES_ASM",
            "BSAES_ASM",
            "ECP_NISTZ256_ASM",
            "GHASH_ASM",
            "KECCAK1600_ASM",
            "MD5_ASM",
            "OPENSSL_BN_ASM_GF2m",
            "OPENSSL_BN_ASM_MONT",
            "OPENSSL_BN_ASM_MONT5",
            "OPENSSL_CPUID_OBJ",
            "OPENSSL_IA32_SSE2",
            "POLY1305_ASM",
            "RC4_ASM",
            "SHA1_ASM",
            "SHA256_ASM",
            "SHA512_ASM",
            "VPAES_ASM",
            "X25519_ASM"
        ],
        "libssl" => [
            "AES_ASM"
        ],
        "providers/legacy" => [
            "OPENSSL_CPUID_OBJ"
        ],
        "providers/libcommon.a" => [
            "OPENSSL_BN_ASM_GF2m",
            "OPENSSL_BN_ASM_MONT",
            "OPENSSL_BN_ASM_MONT5",
            "OPENSSL_CPUID_OBJ",
            "OPENSSL_IA32_SSE2"
        ],
        "providers/libdefault.a" => [
            "AES_ASM",
            "BSAES_ASM",
            "ECP_NISTZ256_ASM",
            "KECCAK1600_ASM",
            "OPENSSL_CPUID_OBJ",
            "OPENSSL_IA32_SSE2",
            "SHA1_ASM",
            "SHA256_ASM",
            "SHA512_ASM",
            "VPAES_ASM",
            "X25519_ASM"
        ],
        "providers/libfips.a" => [
            "AES_ASM",
            "BSAES_ASM",
            "ECP_NISTZ256_ASM",
            "FIPS_MODULE",
            "GHASH_ASM",
            "KECCAK1600_ASM",
            "OPENSSL_BN_ASM_GF2m",
            "OPENSSL_BN_ASM_MONT",
            "OPENSSL_BN_ASM_MONT5",
            "OPENSSL_CPUID_OBJ",
            "OPENSSL_IA32_SSE2",
            "SHA1_ASM",
            "SHA256_ASM",
            "SHA512_ASM",
            "VPAES_ASM",
            "X25519_ASM"
        ],
        "providers/liblegacy.a" => [
            "AES_ASM",
            "BSAES_ASM",
            "ECP_NISTZ256_ASM",
            "KECCAK1600_ASM",
            "MD5_ASM",
            "RC4_ASM",
            "SHA1_ASM",
            "SHA256_ASM",
            "SHA512_ASM",
            "VPAES_ASM",
            "X25519_ASM"
        ]
    },
    "depends" => {
        "" => [
            "OpenSSLConfigVersion.cmake",
            "crypto/params_idx.c",
            "exporters/OpenSSLConfigVersion.cmake",
            "exporters/openssl.pc",
            "include/crypto/bn_conf.h",
            "include/crypto/dso_conf.h",
            "include/internal/param_names.h",
            "include/openssl/asn1.h",
            "include/openssl/asn1t.h",
            "include/openssl/bio.h",
            "include/openssl/cmp.h",
            "include/openssl/cms.h",
            "include/openssl/comp.h",
            "include/openssl/conf.h",
            "include/openssl/core_names.h",
            "include/openssl/crmf.h",
            "include/openssl/crypto.h",
            "include/openssl/ct.h",
            "include/openssl/err.h",
            "include/openssl/ess.h",
            "include/openssl/fipskey.h",
            "include/openssl/lhash.h",
            "include/openssl/ocsp.h",
            "include/openssl/opensslv.h",
            "include/openssl/pkcs12.h",
            "include/openssl/pkcs7.h",
            "include/openssl/safestack.h",
            "include/openssl/srp.h",
            "include/openssl/ssl.h",
            "include/openssl/symbol_prefix.h",
            "include/openssl/ui.h",
            "include/openssl/x509.h",
            "include/openssl/x509_acert.h",
            "include/openssl/x509_vfy.h",
            "include/openssl/x509v3.h",
            "openssl.pc"
        ],
        "OpenSSLConfig.cmake" => [
            "builddata.pm"
        ],
        "OpenSSLConfigVersion.cmake" => [
            "OpenSSLConfig.cmake",
            "builddata.pm"
        ],
        "apps/lib/openssl-bin-cmp_mock_srv.o" => [
            "apps/progs.h"
        ],
        "apps/openssl" => [
            "apps/libapps.a",
            "libssl"
        ],
        "apps/openssl-bin-asn1parse.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ca.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ciphers.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-cmp.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-cms.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-crl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-crl2pkcs7.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-dgst.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-dhparam.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-dsa.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-dsaparam.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ec.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ecparam.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-enc.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-engine.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-errstr.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-fipsinstall.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-gendsa.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-genpkey.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-genrsa.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-info.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-kdf.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-list.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-mac.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-nseq.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ocsp.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-openssl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-passwd.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkcs12.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkcs7.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkcs8.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkey.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkeyparam.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-pkeyutl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-prime.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-progs.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-rand.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-rehash.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-req.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-rsa.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-rsautl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-s_client.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-s_server.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-s_time.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-sess_id.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-skeyutl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-smime.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-speed.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-spkac.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-srp.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-storeutl.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-ts.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-verify.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-version.o" => [
            "apps/progs.h"
        ],
        "apps/openssl-bin-x509.o" => [
            "apps/progs.h"
        ],
        "apps/progs.c" => [
            "configdata.pm"
        ],
        "apps/progs.h" => [
            "apps/progs.c"
        ],
        "crypto/aes/aes-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/aes/aesni-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/aes/vpaes-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/bn/bn-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/bn/co-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/bn/x86-gf2m.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/bn/x86-mont.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/des/crypt586.S" => [
            "crypto/perlasm/cbc.pl",
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/des/des-586.S" => [
            "crypto/perlasm/cbc.pl",
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/libcrypto-lib-cversion.o" => [
            "crypto/buildinf.h"
        ],
        "crypto/libcrypto-lib-info.o" => [
            "crypto/buildinf.h"
        ],
        "crypto/libcrypto-shlib-cversion.o" => [
            "crypto/buildinf.h"
        ],
        "crypto/libcrypto-shlib-info.o" => [
            "crypto/buildinf.h"
        ],
        "crypto/params_idx.c" => [
            "util/perl|OpenSSL/paramnames.pm"
        ],
        "crypto/rc4/rc4-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/sha/sha1-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/sha/sha256-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/sha/sha512-586.S" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "crypto/x86cpuid.s" => [
            "crypto/perlasm/x86asm.pl"
        ],
        "engines/afalg" => [
            "libcrypto"
        ],
        "engines/capi" => [
            "libcrypto"
        ],
        "engines/dasync" => [
            "libcrypto"
        ],
        "engines/ecptest" => [
            "libcrypto"
        ],
        "engines/loader_attic" => [
            "libcrypto"
        ],
        "engines/ossltest" => [
            "libcrypto"
        ],
        "engines/padlock" => [
            "libcrypto"
        ],
        "exporters/OpenSSLConfig.cmake" => [
            "installdata.pm"
        ],
        "exporters/OpenSSLConfigVersion.cmake" => [
            "exporters/OpenSSLConfig.cmake",
            "installdata.pm"
        ],
        "exporters/libcrypto.pc" => [
            "installdata.pm"
        ],
        "exporters/libssl.pc" => [
            "installdata.pm"
        ],
        "exporters/openssl.pc" => [
            "exporters/libcrypto.pc",
            "exporters/libssl.pc",
            "installdata.pm"
        ],
        "include/internal/param_names.h" => [
            "util/perl|OpenSSL/paramnames.pm"
        ],
        "include/openssl/core_names.h" => [
            "util/perl|OpenSSL/paramnames.pm"
        ],
        "include/openssl/symbol_prefix.h" => [
            "configdata.pm",
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_dsa.h",
            "providers/common/include/prov/der_ec.h",
            "providers/common/include/prov/der_ecx.h",
            "providers/common/include/prov/der_rsa.h",
            "providers/common/include/prov/der_sm2.h",
            "providers/common/include/prov/der_wrap.h",
            "util/perl/OpenSSL/ParseC.pm"
        ],
        "libcrypto.ld" => [
            "configdata.pm",
            "util/perl/OpenSSL/Ordinals.pm"
        ],
        "libcrypto.pc" => [
            "builddata.pm"
        ],
        "libssl" => [
            "libcrypto"
        ],
        "libssl.ld" => [
            "configdata.pm",
            "util/perl/OpenSSL/Ordinals.pm"
        ],
        "libssl.pc" => [
            "builddata.pm"
        ],
        "openssl.pc" => [
            "builddata.pm",
            "libcrypto.pc",
            "libssl.pc"
        ],
        "providers/common/der/der_digests_gen.c" => [
            "providers/common/der/DIGESTS.asn1",
            "providers/common/der/NIST.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_dsa_gen.c" => [
            "providers/common/der/DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_ec_gen.c" => [
            "providers/common/der/EC.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_ecx_gen.c" => [
            "providers/common/der/ECX.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_ml_dsa_gen.c" => [
            "providers/common/der/ML_DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_rsa_gen.c" => [
            "providers/common/der/NIST.asn1",
            "providers/common/der/RSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_slh_dsa_gen.c" => [
            "providers/common/der/SLH_DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_sm2_gen.c" => [
            "providers/common/der/SM2.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/der/der_wrap_gen.c" => [
            "providers/common/der/oids_to_c.pm",
            "providers/common/der/wrap.asn1"
        ],
        "providers/common/der/libcommon-lib-der_digests_gen.o" => [
            "providers/common/include/prov/der_digests.h"
        ],
        "providers/common/der/libcommon-lib-der_dsa_gen.o" => [
            "providers/common/include/prov/der_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_dsa_key.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_dsa_sig.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_ec_gen.o" => [
            "providers/common/include/prov/der_ec.h"
        ],
        "providers/common/der/libcommon-lib-der_ec_key.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_ec.h"
        ],
        "providers/common/der/libcommon-lib-der_ec_sig.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_ec.h"
        ],
        "providers/common/der/libcommon-lib-der_ecx_gen.o" => [
            "providers/common/include/prov/der_ecx.h"
        ],
        "providers/common/der/libcommon-lib-der_ecx_key.o" => [
            "providers/common/include/prov/der_ecx.h"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_gen.o" => [
            "providers/common/include/prov/der_ml_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_key.o" => [
            "providers/common/include/prov/der_ml_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_rsa_gen.o" => [
            "providers/common/include/prov/der_rsa.h"
        ],
        "providers/common/der/libcommon-lib-der_rsa_key.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_rsa.h"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_gen.o" => [
            "providers/common/include/prov/der_slh_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_key.o" => [
            "providers/common/include/prov/der_slh_dsa.h"
        ],
        "providers/common/der/libcommon-lib-der_wrap_gen.o" => [
            "providers/common/include/prov/der_wrap.h"
        ],
        "providers/common/der/libdefault-lib-der_rsa_sig.o" => [
            "providers/common/include/prov/der_digests.h",
            "providers/common/include/prov/der_rsa.h"
        ],
        "providers/common/der/libdefault-lib-der_sm2_gen.o" => [
            "providers/common/include/prov/der_sm2.h"
        ],
        "providers/common/der/libdefault-lib-der_sm2_key.o" => [
            "providers/common/include/prov/der_ec.h",
            "providers/common/include/prov/der_sm2.h"
        ],
        "providers/common/der/libdefault-lib-der_sm2_sig.o" => [
            "providers/common/include/prov/der_ec.h",
            "providers/common/include/prov/der_sm2.h"
        ],
        "providers/common/include/prov/der_digests.h" => [
            "providers/common/der/DIGESTS.asn1",
            "providers/common/der/NIST.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_dsa.h" => [
            "providers/common/der/DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_ec.h" => [
            "providers/common/der/EC.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_ecx.h" => [
            "providers/common/der/ECX.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_ml_dsa.h" => [
            "providers/common/der/ML_DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_rsa.h" => [
            "providers/common/der/NIST.asn1",
            "providers/common/der/RSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_slh_dsa.h" => [
            "providers/common/der/SLH_DSA.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_sm2.h" => [
            "providers/common/der/SM2.asn1",
            "providers/common/der/oids_to_c.pm"
        ],
        "providers/common/include/prov/der_wrap.h" => [
            "providers/common/der/oids_to_c.pm",
            "providers/common/der/wrap.asn1"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2any.o" => [
            "providers/common/include/prov/der_rsa.h"
        ],
        "providers/implementations/kdfs/libdefault-lib-x942kdf.o" => [
            "providers/common/include/prov/der_wrap.h"
        ],
        "providers/implementations/signature/libdefault-lib-dsa_sig.o" => [
            "providers/common/include/prov/der_dsa.h"
        ],
        "providers/implementations/signature/libdefault-lib-ecdsa_sig.o" => [
            "providers/common/include/prov/der_ec.h"
        ],
        "providers/implementations/signature/libdefault-lib-eddsa_sig.o" => [
            "providers/common/include/prov/der_ecx.h"
        ],
        "providers/implementations/signature/libdefault-lib-ml_dsa_sig.o" => [
            "providers/common/include/prov/der_ml_dsa.h"
        ],
        "providers/implementations/signature/libdefault-lib-rsa_sig.o" => [
            "providers/common/include/prov/der_rsa.h"
        ],
        "providers/implementations/signature/libdefault-lib-slh_dsa_sig.o" => [
            "providers/common/include/prov/der_slh_dsa.h"
        ],
        "providers/implementations/signature/libdefault-lib-sm2_sig.o" => [
            "providers/common/include/prov/der_sm2.h"
        ],
        "providers/legacy" => [
            "libcrypto",
            "providers/liblegacy.a"
        ],
        "providers/libcommon.a" => [
            "libcrypto"
        ],
        "providers/libdefault.a" => [
            "providers/libcommon.a"
        ],
        "providers/liblegacy.a" => [
            "providers/libcommon.a"
        ],
        "providers/libsmtc.a" => [
            "providers/libcommon.a"
        ],
        "util/wrap.pl" => [
            "configdata.pm"
        ]
    },
    "dirinfo" => {
        "apps" => {
            "products" => {
                "bin" => [
                    "apps/openssl"
                ],
                "script" => [
                    "apps/CA.pl",
                    "apps/tsget.pl"
                ]
            }
        },
        "apps/lib" => {
            "deps" => [
                "apps/lib/openssl-bin-cmp_mock_srv.o",
                "apps/lib/libapps-lib-app_libctx.o",
                "apps/lib/libapps-lib-app_params.o",
                "apps/lib/libapps-lib-app_provider.o",
                "apps/lib/libapps-lib-app_rand.o",
                "apps/lib/libapps-lib-apps.o",
                "apps/lib/libapps-lib-apps_opt_printf.o",
                "apps/lib/libapps-lib-apps_ui.o",
                "apps/lib/libapps-lib-columns.o",
                "apps/lib/libapps-lib-engine.o",
                "apps/lib/libapps-lib-engine_loader.o",
                "apps/lib/libapps-lib-fmt.o",
                "apps/lib/libapps-lib-http_server.o",
                "apps/lib/libapps-lib-log.o",
                "apps/lib/libapps-lib-names.o",
                "apps/lib/libapps-lib-opt.o",
                "apps/lib/libapps-lib-s_cb.o",
                "apps/lib/libapps-lib-s_socket.o",
                "apps/lib/libapps-lib-tlssrp_depr.o"
            ],
            "products" => {
                "bin" => [
                    "apps/openssl"
                ],
                "lib" => [
                    "apps/libapps.a"
                ]
            }
        },
        "crypto" => {
            "deps" => [
                "crypto/legacy-dso-cpuid.o",
                "crypto/legacy-dso-ctype.o",
                "crypto/legacy-dso-x86_64cpuid.o",
                "crypto/libcrypto-lib-asn1_dsa.o",
                "crypto/libcrypto-lib-bsearch.o",
                "crypto/libcrypto-lib-comp_methods.o",
                "crypto/libcrypto-lib-context.o",
                "crypto/libcrypto-lib-core_algorithm.o",
                "crypto/libcrypto-lib-core_fetch.o",
                "crypto/libcrypto-lib-core_namemap.o",
                "crypto/libcrypto-lib-cpt_err.o",
                "crypto/libcrypto-lib-cpuid.o",
                "crypto/libcrypto-lib-cryptlib.o",
                "crypto/libcrypto-lib-ctype.o",
                "crypto/libcrypto-lib-cversion.o",
                "crypto/libcrypto-lib-defaults.o",
                "crypto/libcrypto-lib-der_writer.o",
                "crypto/libcrypto-lib-deterministic_nonce.o",
                "crypto/libcrypto-lib-ebcdic.o",
                "crypto/libcrypto-lib-ex_data.o",
                "crypto/libcrypto-lib-getenv.o",
                "crypto/libcrypto-lib-indicator_core.o",
                "crypto/libcrypto-lib-info.o",
                "crypto/libcrypto-lib-init.o",
                "crypto/libcrypto-lib-initthread.o",
                "crypto/libcrypto-lib-mem.o",
                "crypto/libcrypto-lib-mem_sec.o",
                "crypto/libcrypto-lib-o_dir.o",
                "crypto/libcrypto-lib-o_fopen.o",
                "crypto/libcrypto-lib-o_init.o",
                "crypto/libcrypto-lib-o_str.o",
                "crypto/libcrypto-lib-o_syslog.o",
                "crypto/libcrypto-lib-o_time.o",
                "crypto/libcrypto-lib-packet.o",
                "crypto/libcrypto-lib-param_build.o",
                "crypto/libcrypto-lib-param_build_set.o",
                "crypto/libcrypto-lib-params.o",
                "crypto/libcrypto-lib-params_dup.o",
                "crypto/libcrypto-lib-params_from_text.o",
                "crypto/libcrypto-lib-params_idx.o",
                "crypto/libcrypto-lib-passphrase.o",
                "crypto/libcrypto-lib-provider.o",
                "crypto/libcrypto-lib-provider_child.o",
                "crypto/libcrypto-lib-provider_conf.o",
                "crypto/libcrypto-lib-provider_core.o",
                "crypto/libcrypto-lib-provider_predefined.o",
                "crypto/libcrypto-lib-punycode.o",
                "crypto/libcrypto-lib-quic_vlint.o",
                "crypto/libcrypto-lib-self_test_core.o",
                "crypto/libcrypto-lib-sleep.o",
                "crypto/libcrypto-lib-sparse_array.o",
                "crypto/libcrypto-lib-ssl_err.o",
                "crypto/libcrypto-lib-threads_lib.o",
                "crypto/libcrypto-lib-threads_none.o",
                "crypto/libcrypto-lib-threads_pthread.o",
                "crypto/libcrypto-lib-threads_win.o",
                "crypto/libcrypto-lib-time.o",
                "crypto/libcrypto-lib-trace.o",
                "crypto/libcrypto-lib-uid.o",
                "crypto/libcrypto-lib-x86_64cpuid.o",
                "crypto/libcrypto-shlib-asn1_dsa.o",
                "crypto/libcrypto-shlib-bsearch.o",
                "crypto/libcrypto-shlib-comp_methods.o",
                "crypto/libcrypto-shlib-context.o",
                "crypto/libcrypto-shlib-core_algorithm.o",
                "crypto/libcrypto-shlib-core_fetch.o",
                "crypto/libcrypto-shlib-core_namemap.o",
                "crypto/libcrypto-shlib-cpt_err.o",
                "crypto/libcrypto-shlib-cpuid.o",
                "crypto/libcrypto-shlib-cryptlib.o",
                "crypto/libcrypto-shlib-ctype.o",
                "crypto/libcrypto-shlib-cversion.o",
                "crypto/libcrypto-shlib-defaults.o",
                "crypto/libcrypto-shlib-der_writer.o",
                "crypto/libcrypto-shlib-deterministic_nonce.o",
                "crypto/libcrypto-shlib-ebcdic.o",
                "crypto/libcrypto-shlib-ex_data.o",
                "crypto/libcrypto-shlib-getenv.o",
                "crypto/libcrypto-shlib-indicator_core.o",
                "crypto/libcrypto-shlib-info.o",
                "crypto/libcrypto-shlib-init.o",
                "crypto/libcrypto-shlib-initthread.o",
                "crypto/libcrypto-shlib-mem.o",
                "crypto/libcrypto-shlib-mem_sec.o",
                "crypto/libcrypto-shlib-o_dir.o",
                "crypto/libcrypto-shlib-o_fopen.o",
                "crypto/libcrypto-shlib-o_init.o",
                "crypto/libcrypto-shlib-o_str.o",
                "crypto/libcrypto-shlib-o_syslog.o",
                "crypto/libcrypto-shlib-o_time.o",
                "crypto/libcrypto-shlib-packet.o",
                "crypto/libcrypto-shlib-param_build.o",
                "crypto/libcrypto-shlib-param_build_set.o",
                "crypto/libcrypto-shlib-params.o",
                "crypto/libcrypto-shlib-params_dup.o",
                "crypto/libcrypto-shlib-params_from_text.o",
                "crypto/libcrypto-shlib-params_idx.o",
                "crypto/libcrypto-shlib-passphrase.o",
                "crypto/libcrypto-shlib-provider.o",
                "crypto/libcrypto-shlib-provider_child.o",
                "crypto/libcrypto-shlib-provider_conf.o",
                "crypto/libcrypto-shlib-provider_core.o",
                "crypto/libcrypto-shlib-provider_predefined.o",
                "crypto/libcrypto-shlib-punycode.o",
                "crypto/libcrypto-shlib-quic_vlint.o",
                "crypto/libcrypto-shlib-self_test_core.o",
                "crypto/libcrypto-shlib-sleep.o",
                "crypto/libcrypto-shlib-sparse_array.o",
                "crypto/libcrypto-shlib-ssl_err.o",
                "crypto/libcrypto-shlib-threads_lib.o",
                "crypto/libcrypto-shlib-threads_none.o",
                "crypto/libcrypto-shlib-threads_pthread.o",
                "crypto/libcrypto-shlib-threads_win.o",
                "crypto/libcrypto-shlib-time.o",
                "crypto/libcrypto-shlib-trace.o",
                "crypto/libcrypto-shlib-uid.o",
                "crypto/libcrypto-shlib-x86_64cpuid.o",
                "crypto/libssl-shlib-ctype.o",
                "crypto/libssl-shlib-getenv.o",
                "crypto/libssl-shlib-packet.o",
                "crypto/libssl-shlib-quic_vlint.o",
                "crypto/libssl-shlib-time.o"
            ],
            "products" => {
                "dso" => [
                    "providers/legacy"
                ],
                "lib" => [
                    "libcrypto",
                    "libssl"
                ]
            }
        },
        "crypto/aes" => {
            "deps" => [
                "crypto/aes/libcrypto-lib-aes-x86_64.o",
                "crypto/aes/libcrypto-lib-aes_cfb.o",
                "crypto/aes/libcrypto-lib-aes_ecb.o",
                "crypto/aes/libcrypto-lib-aes_ige.o",
                "crypto/aes/libcrypto-lib-aes_misc.o",
                "crypto/aes/libcrypto-lib-aes_ofb.o",
                "crypto/aes/libcrypto-lib-aes_wrap.o",
                "crypto/aes/libcrypto-lib-aesni-mb-x86_64.o",
                "crypto/aes/libcrypto-lib-aesni-sha1-x86_64.o",
                "crypto/aes/libcrypto-lib-aesni-sha256-x86_64.o",
                "crypto/aes/libcrypto-lib-aesni-x86_64.o",
                "crypto/aes/libcrypto-lib-aesni-xts-avx512.o",
                "crypto/aes/libcrypto-lib-bsaes-x86_64.o",
                "crypto/aes/libcrypto-lib-vpaes-x86_64.o",
                "crypto/aes/libcrypto-shlib-aes-x86_64.o",
                "crypto/aes/libcrypto-shlib-aes_cfb.o",
                "crypto/aes/libcrypto-shlib-aes_ecb.o",
                "crypto/aes/libcrypto-shlib-aes_ige.o",
                "crypto/aes/libcrypto-shlib-aes_misc.o",
                "crypto/aes/libcrypto-shlib-aes_ofb.o",
                "crypto/aes/libcrypto-shlib-aes_wrap.o",
                "crypto/aes/libcrypto-shlib-aesni-mb-x86_64.o",
                "crypto/aes/libcrypto-shlib-aesni-sha1-x86_64.o",
                "crypto/aes/libcrypto-shlib-aesni-sha256-x86_64.o",
                "crypto/aes/libcrypto-shlib-aesni-x86_64.o",
                "crypto/aes/libcrypto-shlib-aesni-xts-avx512.o",
                "crypto/aes/libcrypto-shlib-bsaes-x86_64.o",
                "crypto/aes/libcrypto-shlib-vpaes-x86_64.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/asn1" => {
            "deps" => [
                "crypto/asn1/libcrypto-lib-a_bitstr.o",
                "crypto/asn1/libcrypto-lib-a_d2i_fp.o",
                "crypto/asn1/libcrypto-lib-a_digest.o",
                "crypto/asn1/libcrypto-lib-a_dup.o",
                "crypto/asn1/libcrypto-lib-a_gentm.o",
                "crypto/asn1/libcrypto-lib-a_i2d_fp.o",
                "crypto/asn1/libcrypto-lib-a_int.o",
                "crypto/asn1/libcrypto-lib-a_mbstr.o",
                "crypto/asn1/libcrypto-lib-a_object.o",
                "crypto/asn1/libcrypto-lib-a_octet.o",
                "crypto/asn1/libcrypto-lib-a_print.o",
                "crypto/asn1/libcrypto-lib-a_sign.o",
                "crypto/asn1/libcrypto-lib-a_strex.o",
                "crypto/asn1/libcrypto-lib-a_strnid.o",
                "crypto/asn1/libcrypto-lib-a_time.o",
                "crypto/asn1/libcrypto-lib-a_type.o",
                "crypto/asn1/libcrypto-lib-a_utctm.o",
                "crypto/asn1/libcrypto-lib-a_utf8.o",
                "crypto/asn1/libcrypto-lib-a_verify.o",
                "crypto/asn1/libcrypto-lib-ameth_lib.o",
                "crypto/asn1/libcrypto-lib-asn1_err.o",
                "crypto/asn1/libcrypto-lib-asn1_gen.o",
                "crypto/asn1/libcrypto-lib-asn1_item_list.o",
                "crypto/asn1/libcrypto-lib-asn1_lib.o",
                "crypto/asn1/libcrypto-lib-asn1_parse.o",
                "crypto/asn1/libcrypto-lib-asn_mime.o",
                "crypto/asn1/libcrypto-lib-asn_moid.o",
                "crypto/asn1/libcrypto-lib-asn_mstbl.o",
                "crypto/asn1/libcrypto-lib-asn_pack.o",
                "crypto/asn1/libcrypto-lib-bio_asn1.o",
                "crypto/asn1/libcrypto-lib-bio_ndef.o",
                "crypto/asn1/libcrypto-lib-d2i_param.o",
                "crypto/asn1/libcrypto-lib-d2i_pr.o",
                "crypto/asn1/libcrypto-lib-d2i_pu.o",
                "crypto/asn1/libcrypto-lib-evp_asn1.o",
                "crypto/asn1/libcrypto-lib-f_int.o",
                "crypto/asn1/libcrypto-lib-f_string.o",
                "crypto/asn1/libcrypto-lib-i2d_evp.o",
                "crypto/asn1/libcrypto-lib-n_pkey.o",
                "crypto/asn1/libcrypto-lib-nsseq.o",
                "crypto/asn1/libcrypto-lib-p5_pbe.o",
                "crypto/asn1/libcrypto-lib-p5_pbev2.o",
                "crypto/asn1/libcrypto-lib-p5_scrypt.o",
                "crypto/asn1/libcrypto-lib-p8_pkey.o",
                "crypto/asn1/libcrypto-lib-t_bitst.o",
                "crypto/asn1/libcrypto-lib-t_pkey.o",
                "crypto/asn1/libcrypto-lib-t_spki.o",
                "crypto/asn1/libcrypto-lib-tasn_dec.o",
                "crypto/asn1/libcrypto-lib-tasn_enc.o",
                "crypto/asn1/libcrypto-lib-tasn_fre.o",
                "crypto/asn1/libcrypto-lib-tasn_new.o",
                "crypto/asn1/libcrypto-lib-tasn_prn.o",
                "crypto/asn1/libcrypto-lib-tasn_scn.o",
                "crypto/asn1/libcrypto-lib-tasn_typ.o",
                "crypto/asn1/libcrypto-lib-tasn_utl.o",
                "crypto/asn1/libcrypto-lib-x_algor.o",
                "crypto/asn1/libcrypto-lib-x_bignum.o",
                "crypto/asn1/libcrypto-lib-x_info.o",
                "crypto/asn1/libcrypto-lib-x_int64.o",
                "crypto/asn1/libcrypto-lib-x_long.o",
                "crypto/asn1/libcrypto-lib-x_pkey.o",
                "crypto/asn1/libcrypto-lib-x_sig.o",
                "crypto/asn1/libcrypto-lib-x_spki.o",
                "crypto/asn1/libcrypto-lib-x_val.o",
                "crypto/asn1/libcrypto-shlib-a_bitstr.o",
                "crypto/asn1/libcrypto-shlib-a_d2i_fp.o",
                "crypto/asn1/libcrypto-shlib-a_digest.o",
                "crypto/asn1/libcrypto-shlib-a_dup.o",
                "crypto/asn1/libcrypto-shlib-a_gentm.o",
                "crypto/asn1/libcrypto-shlib-a_i2d_fp.o",
                "crypto/asn1/libcrypto-shlib-a_int.o",
                "crypto/asn1/libcrypto-shlib-a_mbstr.o",
                "crypto/asn1/libcrypto-shlib-a_object.o",
                "crypto/asn1/libcrypto-shlib-a_octet.o",
                "crypto/asn1/libcrypto-shlib-a_print.o",
                "crypto/asn1/libcrypto-shlib-a_sign.o",
                "crypto/asn1/libcrypto-shlib-a_strex.o",
                "crypto/asn1/libcrypto-shlib-a_strnid.o",
                "crypto/asn1/libcrypto-shlib-a_time.o",
                "crypto/asn1/libcrypto-shlib-a_type.o",
                "crypto/asn1/libcrypto-shlib-a_utctm.o",
                "crypto/asn1/libcrypto-shlib-a_utf8.o",
                "crypto/asn1/libcrypto-shlib-a_verify.o",
                "crypto/asn1/libcrypto-shlib-ameth_lib.o",
                "crypto/asn1/libcrypto-shlib-asn1_err.o",
                "crypto/asn1/libcrypto-shlib-asn1_gen.o",
                "crypto/asn1/libcrypto-shlib-asn1_item_list.o",
                "crypto/asn1/libcrypto-shlib-asn1_lib.o",
                "crypto/asn1/libcrypto-shlib-asn1_parse.o",
                "crypto/asn1/libcrypto-shlib-asn_mime.o",
                "crypto/asn1/libcrypto-shlib-asn_moid.o",
                "crypto/asn1/libcrypto-shlib-asn_mstbl.o",
                "crypto/asn1/libcrypto-shlib-asn_pack.o",
                "crypto/asn1/libcrypto-shlib-bio_asn1.o",
                "crypto/asn1/libcrypto-shlib-bio_ndef.o",
                "crypto/asn1/libcrypto-shlib-d2i_param.o",
                "crypto/asn1/libcrypto-shlib-d2i_pr.o",
                "crypto/asn1/libcrypto-shlib-d2i_pu.o",
                "crypto/asn1/libcrypto-shlib-evp_asn1.o",
                "crypto/asn1/libcrypto-shlib-f_int.o",
                "crypto/asn1/libcrypto-shlib-f_string.o",
                "crypto/asn1/libcrypto-shlib-i2d_evp.o",
                "crypto/asn1/libcrypto-shlib-n_pkey.o",
                "crypto/asn1/libcrypto-shlib-nsseq.o",
                "crypto/asn1/libcrypto-shlib-p5_pbe.o",
                "crypto/asn1/libcrypto-shlib-p5_pbev2.o",
                "crypto/asn1/libcrypto-shlib-p5_scrypt.o",
                "crypto/asn1/libcrypto-shlib-p8_pkey.o",
                "crypto/asn1/libcrypto-shlib-t_bitst.o",
                "crypto/asn1/libcrypto-shlib-t_pkey.o",
                "crypto/asn1/libcrypto-shlib-t_spki.o",
                "crypto/asn1/libcrypto-shlib-tasn_dec.o",
                "crypto/asn1/libcrypto-shlib-tasn_enc.o",
                "crypto/asn1/libcrypto-shlib-tasn_fre.o",
                "crypto/asn1/libcrypto-shlib-tasn_new.o",
                "crypto/asn1/libcrypto-shlib-tasn_prn.o",
                "crypto/asn1/libcrypto-shlib-tasn_scn.o",
                "crypto/asn1/libcrypto-shlib-tasn_typ.o",
                "crypto/asn1/libcrypto-shlib-tasn_utl.o",
                "crypto/asn1/libcrypto-shlib-x_algor.o",
                "crypto/asn1/libcrypto-shlib-x_bignum.o",
                "crypto/asn1/libcrypto-shlib-x_info.o",
                "crypto/asn1/libcrypto-shlib-x_int64.o",
                "crypto/asn1/libcrypto-shlib-x_long.o",
                "crypto/asn1/libcrypto-shlib-x_pkey.o",
                "crypto/asn1/libcrypto-shlib-x_sig.o",
                "crypto/asn1/libcrypto-shlib-x_spki.o",
                "crypto/asn1/libcrypto-shlib-x_val.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/async" => {
            "deps" => [
                "crypto/async/libcrypto-lib-async.o",
                "crypto/async/libcrypto-lib-async_err.o",
                "crypto/async/libcrypto-lib-async_wait.o",
                "crypto/async/libcrypto-shlib-async.o",
                "crypto/async/libcrypto-shlib-async_err.o",
                "crypto/async/libcrypto-shlib-async_wait.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/async/arch" => {
            "deps" => [
                "crypto/async/arch/libcrypto-lib-async_null.o",
                "crypto/async/arch/libcrypto-lib-async_posix.o",
                "crypto/async/arch/libcrypto-lib-async_win.o",
                "crypto/async/arch/libcrypto-shlib-async_null.o",
                "crypto/async/arch/libcrypto-shlib-async_posix.o",
                "crypto/async/arch/libcrypto-shlib-async_win.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/bio" => {
            "deps" => [
                "crypto/bio/libcrypto-lib-bf_buff.o",
                "crypto/bio/libcrypto-lib-bf_lbuf.o",
                "crypto/bio/libcrypto-lib-bf_nbio.o",
                "crypto/bio/libcrypto-lib-bf_null.o",
                "crypto/bio/libcrypto-lib-bf_prefix.o",
                "crypto/bio/libcrypto-lib-bf_readbuff.o",
                "crypto/bio/libcrypto-lib-bio_addr.o",
                "crypto/bio/libcrypto-lib-bio_cb.o",
                "crypto/bio/libcrypto-lib-bio_dump.o",
                "crypto/bio/libcrypto-lib-bio_err.o",
                "crypto/bio/libcrypto-lib-bio_lib.o",
                "crypto/bio/libcrypto-lib-bio_meth.o",
                "crypto/bio/libcrypto-lib-bio_print.o",
                "crypto/bio/libcrypto-lib-bio_sock.o",
                "crypto/bio/libcrypto-lib-bio_sock2.o",
                "crypto/bio/libcrypto-lib-bss_acpt.o",
                "crypto/bio/libcrypto-lib-bss_bio.o",
                "crypto/bio/libcrypto-lib-bss_conn.o",
                "crypto/bio/libcrypto-lib-bss_core.o",
                "crypto/bio/libcrypto-lib-bss_dgram.o",
                "crypto/bio/libcrypto-lib-bss_dgram_pair.o",
                "crypto/bio/libcrypto-lib-bss_fd.o",
                "crypto/bio/libcrypto-lib-bss_file.o",
                "crypto/bio/libcrypto-lib-bss_log.o",
                "crypto/bio/libcrypto-lib-bss_mem.o",
                "crypto/bio/libcrypto-lib-bss_null.o",
                "crypto/bio/libcrypto-lib-bss_sock.o",
                "crypto/bio/libcrypto-lib-ossl_core_bio.o",
                "crypto/bio/libcrypto-shlib-bf_buff.o",
                "crypto/bio/libcrypto-shlib-bf_lbuf.o",
                "crypto/bio/libcrypto-shlib-bf_nbio.o",
                "crypto/bio/libcrypto-shlib-bf_null.o",
                "crypto/bio/libcrypto-shlib-bf_prefix.o",
                "crypto/bio/libcrypto-shlib-bf_readbuff.o",
                "crypto/bio/libcrypto-shlib-bio_addr.o",
                "crypto/bio/libcrypto-shlib-bio_cb.o",
                "crypto/bio/libcrypto-shlib-bio_dump.o",
                "crypto/bio/libcrypto-shlib-bio_err.o",
                "crypto/bio/libcrypto-shlib-bio_lib.o",
                "crypto/bio/libcrypto-shlib-bio_meth.o",
                "crypto/bio/libcrypto-shlib-bio_print.o",
                "crypto/bio/libcrypto-shlib-bio_sock.o",
                "crypto/bio/libcrypto-shlib-bio_sock2.o",
                "crypto/bio/libcrypto-shlib-bss_acpt.o",
                "crypto/bio/libcrypto-shlib-bss_bio.o",
                "crypto/bio/libcrypto-shlib-bss_conn.o",
                "crypto/bio/libcrypto-shlib-bss_core.o",
                "crypto/bio/libcrypto-shlib-bss_dgram.o",
                "crypto/bio/libcrypto-shlib-bss_dgram_pair.o",
                "crypto/bio/libcrypto-shlib-bss_fd.o",
                "crypto/bio/libcrypto-shlib-bss_file.o",
                "crypto/bio/libcrypto-shlib-bss_log.o",
                "crypto/bio/libcrypto-shlib-bss_mem.o",
                "crypto/bio/libcrypto-shlib-bss_null.o",
                "crypto/bio/libcrypto-shlib-bss_sock.o",
                "crypto/bio/libcrypto-shlib-ossl_core_bio.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/bn" => {
            "deps" => [
                "crypto/bn/libcrypto-lib-bn_add.o",
                "crypto/bn/libcrypto-lib-bn_blind.o",
                "crypto/bn/libcrypto-lib-bn_const.o",
                "crypto/bn/libcrypto-lib-bn_conv.o",
                "crypto/bn/libcrypto-lib-bn_ctx.o",
                "crypto/bn/libcrypto-lib-bn_depr.o",
                "crypto/bn/libcrypto-lib-bn_dh.o",
                "crypto/bn/libcrypto-lib-bn_div.o",
                "crypto/bn/libcrypto-lib-bn_err.o",
                "crypto/bn/libcrypto-lib-bn_exp.o",
                "crypto/bn/libcrypto-lib-bn_exp2.o",
                "crypto/bn/libcrypto-lib-bn_gcd.o",
                "crypto/bn/libcrypto-lib-bn_gf2m.o",
                "crypto/bn/libcrypto-lib-bn_intern.o",
                "crypto/bn/libcrypto-lib-bn_kron.o",
                "crypto/bn/libcrypto-lib-bn_lib.o",
                "crypto/bn/libcrypto-lib-bn_mod.o",
                "crypto/bn/libcrypto-lib-bn_mont.o",
                "crypto/bn/libcrypto-lib-bn_mpi.o",
                "crypto/bn/libcrypto-lib-bn_mul.o",
                "crypto/bn/libcrypto-lib-bn_nist.o",
                "crypto/bn/libcrypto-lib-bn_prime.o",
                "crypto/bn/libcrypto-lib-bn_print.o",
                "crypto/bn/libcrypto-lib-bn_rand.o",
                "crypto/bn/libcrypto-lib-bn_recp.o",
                "crypto/bn/libcrypto-lib-bn_rsa_fips186_4.o",
                "crypto/bn/libcrypto-lib-bn_shift.o",
                "crypto/bn/libcrypto-lib-bn_sm2.o",
                "crypto/bn/libcrypto-lib-bn_sqr.o",
                "crypto/bn/libcrypto-lib-bn_sqrt.o",
                "crypto/bn/libcrypto-lib-bn_srp.o",
                "crypto/bn/libcrypto-lib-bn_word.o",
                "crypto/bn/libcrypto-lib-bn_x931p.o",
                "crypto/bn/libcrypto-lib-rsaz-2k-avx512.o",
                "crypto/bn/libcrypto-lib-rsaz-2k-avxifma.o",
                "crypto/bn/libcrypto-lib-rsaz-3k-avx512.o",
                "crypto/bn/libcrypto-lib-rsaz-3k-avxifma.o",
                "crypto/bn/libcrypto-lib-rsaz-4k-avx512.o",
                "crypto/bn/libcrypto-lib-rsaz-4k-avxifma.o",
                "crypto/bn/libcrypto-lib-rsaz-avx2.o",
                "crypto/bn/libcrypto-lib-rsaz-x86_64.o",
                "crypto/bn/libcrypto-lib-rsaz_exp.o",
                "crypto/bn/libcrypto-lib-rsaz_exp_x2.o",
                "crypto/bn/libcrypto-lib-x86_64-gf2m.o",
                "crypto/bn/libcrypto-lib-x86_64-mont.o",
                "crypto/bn/libcrypto-lib-x86_64-mont5.o",
                "crypto/bn/libcrypto-shlib-bn_add.o",
                "crypto/bn/libcrypto-shlib-bn_blind.o",
                "crypto/bn/libcrypto-shlib-bn_const.o",
                "crypto/bn/libcrypto-shlib-bn_conv.o",
                "crypto/bn/libcrypto-shlib-bn_ctx.o",
                "crypto/bn/libcrypto-shlib-bn_depr.o",
                "crypto/bn/libcrypto-shlib-bn_dh.o",
                "crypto/bn/libcrypto-shlib-bn_div.o",
                "crypto/bn/libcrypto-shlib-bn_err.o",
                "crypto/bn/libcrypto-shlib-bn_exp.o",
                "crypto/bn/libcrypto-shlib-bn_exp2.o",
                "crypto/bn/libcrypto-shlib-bn_gcd.o",
                "crypto/bn/libcrypto-shlib-bn_gf2m.o",
                "crypto/bn/libcrypto-shlib-bn_intern.o",
                "crypto/bn/libcrypto-shlib-bn_kron.o",
                "crypto/bn/libcrypto-shlib-bn_lib.o",
                "crypto/bn/libcrypto-shlib-bn_mod.o",
                "crypto/bn/libcrypto-shlib-bn_mont.o",
                "crypto/bn/libcrypto-shlib-bn_mpi.o",
                "crypto/bn/libcrypto-shlib-bn_mul.o",
                "crypto/bn/libcrypto-shlib-bn_nist.o",
                "crypto/bn/libcrypto-shlib-bn_prime.o",
                "crypto/bn/libcrypto-shlib-bn_print.o",
                "crypto/bn/libcrypto-shlib-bn_rand.o",
                "crypto/bn/libcrypto-shlib-bn_recp.o",
                "crypto/bn/libcrypto-shlib-bn_rsa_fips186_4.o",
                "crypto/bn/libcrypto-shlib-bn_shift.o",
                "crypto/bn/libcrypto-shlib-bn_sm2.o",
                "crypto/bn/libcrypto-shlib-bn_sqr.o",
                "crypto/bn/libcrypto-shlib-bn_sqrt.o",
                "crypto/bn/libcrypto-shlib-bn_srp.o",
                "crypto/bn/libcrypto-shlib-bn_word.o",
                "crypto/bn/libcrypto-shlib-bn_x931p.o",
                "crypto/bn/libcrypto-shlib-rsaz-2k-avx512.o",
                "crypto/bn/libcrypto-shlib-rsaz-2k-avxifma.o",
                "crypto/bn/libcrypto-shlib-rsaz-3k-avx512.o",
                "crypto/bn/libcrypto-shlib-rsaz-3k-avxifma.o",
                "crypto/bn/libcrypto-shlib-rsaz-4k-avx512.o",
                "crypto/bn/libcrypto-shlib-rsaz-4k-avxifma.o",
                "crypto/bn/libcrypto-shlib-rsaz-avx2.o",
                "crypto/bn/libcrypto-shlib-rsaz-x86_64.o",
                "crypto/bn/libcrypto-shlib-rsaz_exp.o",
                "crypto/bn/libcrypto-shlib-rsaz_exp_x2.o",
                "crypto/bn/libcrypto-shlib-x86_64-gf2m.o",
                "crypto/bn/libcrypto-shlib-x86_64-mont.o",
                "crypto/bn/libcrypto-shlib-x86_64-mont5.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/bn/asm" => {
            "deps" => [
                "crypto/bn/asm/libcrypto-lib-x86_64-gcc.o",
                "crypto/bn/asm/libcrypto-shlib-x86_64-gcc.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/buffer" => {
            "deps" => [
                "crypto/buffer/libcrypto-lib-buf_err.o",
                "crypto/buffer/libcrypto-lib-buffer.o",
                "crypto/buffer/libcrypto-shlib-buf_err.o",
                "crypto/buffer/libcrypto-shlib-buffer.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/chacha" => {
            "deps" => [
                "crypto/chacha/libcrypto-lib-chacha-x86_64.o",
                "crypto/chacha/libcrypto-shlib-chacha-x86_64.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/cmac" => {
            "deps" => [
                "crypto/cmac/libcrypto-lib-cmac.o",
                "crypto/cmac/libcrypto-shlib-cmac.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/cmp" => {
            "deps" => [
                "crypto/cmp/libcrypto-lib-cmp_asn.o",
                "crypto/cmp/libcrypto-lib-cmp_client.o",
                "crypto/cmp/libcrypto-lib-cmp_ctx.o",
                "crypto/cmp/libcrypto-lib-cmp_err.o",
                "crypto/cmp/libcrypto-lib-cmp_genm.o",
                "crypto/cmp/libcrypto-lib-cmp_hdr.o",
                "crypto/cmp/libcrypto-lib-cmp_http.o",
                "crypto/cmp/libcrypto-lib-cmp_msg.o",
                "crypto/cmp/libcrypto-lib-cmp_protect.o",
                "crypto/cmp/libcrypto-lib-cmp_server.o",
                "crypto/cmp/libcrypto-lib-cmp_status.o",
                "crypto/cmp/libcrypto-lib-cmp_util.o",
                "crypto/cmp/libcrypto-lib-cmp_vfy.o",
                "crypto/cmp/libcrypto-shlib-cmp_asn.o",
                "crypto/cmp/libcrypto-shlib-cmp_client.o",
                "crypto/cmp/libcrypto-shlib-cmp_ctx.o",
                "crypto/cmp/libcrypto-shlib-cmp_err.o",
                "crypto/cmp/libcrypto-shlib-cmp_genm.o",
                "crypto/cmp/libcrypto-shlib-cmp_hdr.o",
                "crypto/cmp/libcrypto-shlib-cmp_http.o",
                "crypto/cmp/libcrypto-shlib-cmp_msg.o",
                "crypto/cmp/libcrypto-shlib-cmp_protect.o",
                "crypto/cmp/libcrypto-shlib-cmp_server.o",
                "crypto/cmp/libcrypto-shlib-cmp_status.o",
                "crypto/cmp/libcrypto-shlib-cmp_util.o",
                "crypto/cmp/libcrypto-shlib-cmp_vfy.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/cms" => {
            "deps" => [
                "crypto/cms/libcrypto-lib-cms_asn1.o",
                "crypto/cms/libcrypto-lib-cms_att.o",
                "crypto/cms/libcrypto-lib-cms_cd.o",
                "crypto/cms/libcrypto-lib-cms_dd.o",
                "crypto/cms/libcrypto-lib-cms_dh.o",
                "crypto/cms/libcrypto-lib-cms_ec.o",
                "crypto/cms/libcrypto-lib-cms_enc.o",
                "crypto/cms/libcrypto-lib-cms_env.o",
                "crypto/cms/libcrypto-lib-cms_err.o",
                "crypto/cms/libcrypto-lib-cms_ess.o",
                "crypto/cms/libcrypto-lib-cms_io.o",
                "crypto/cms/libcrypto-lib-cms_kari.o",
                "crypto/cms/libcrypto-lib-cms_lib.o",
                "crypto/cms/libcrypto-lib-cms_pwri.o",
                "crypto/cms/libcrypto-lib-cms_rsa.o",
                "crypto/cms/libcrypto-lib-cms_sd.o",
                "crypto/cms/libcrypto-lib-cms_smime.o",
                "crypto/cms/libcrypto-shlib-cms_asn1.o",
                "crypto/cms/libcrypto-shlib-cms_att.o",
                "crypto/cms/libcrypto-shlib-cms_cd.o",
                "crypto/cms/libcrypto-shlib-cms_dd.o",
                "crypto/cms/libcrypto-shlib-cms_dh.o",
                "crypto/cms/libcrypto-shlib-cms_ec.o",
                "crypto/cms/libcrypto-shlib-cms_enc.o",
                "crypto/cms/libcrypto-shlib-cms_env.o",
                "crypto/cms/libcrypto-shlib-cms_err.o",
                "crypto/cms/libcrypto-shlib-cms_ess.o",
                "crypto/cms/libcrypto-shlib-cms_io.o",
                "crypto/cms/libcrypto-shlib-cms_kari.o",
                "crypto/cms/libcrypto-shlib-cms_lib.o",
                "crypto/cms/libcrypto-shlib-cms_pwri.o",
                "crypto/cms/libcrypto-shlib-cms_rsa.o",
                "crypto/cms/libcrypto-shlib-cms_sd.o",
                "crypto/cms/libcrypto-shlib-cms_smime.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/comp" => {
            "deps" => [
                "crypto/comp/libcrypto-lib-c_brotli.o",
                "crypto/comp/libcrypto-lib-c_zlib.o",
                "crypto/comp/libcrypto-lib-c_zstd.o",
                "crypto/comp/libcrypto-lib-comp_err.o",
                "crypto/comp/libcrypto-lib-comp_lib.o",
                "crypto/comp/libcrypto-shlib-c_brotli.o",
                "crypto/comp/libcrypto-shlib-c_zlib.o",
                "crypto/comp/libcrypto-shlib-c_zstd.o",
                "crypto/comp/libcrypto-shlib-comp_err.o",
                "crypto/comp/libcrypto-shlib-comp_lib.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/conf" => {
            "deps" => [
                "crypto/conf/libcrypto-lib-conf_api.o",
                "crypto/conf/libcrypto-lib-conf_def.o",
                "crypto/conf/libcrypto-lib-conf_err.o",
                "crypto/conf/libcrypto-lib-conf_lib.o",
                "crypto/conf/libcrypto-lib-conf_mall.o",
                "crypto/conf/libcrypto-lib-conf_mod.o",
                "crypto/conf/libcrypto-lib-conf_sap.o",
                "crypto/conf/libcrypto-lib-conf_ssl.o",
                "crypto/conf/libcrypto-shlib-conf_api.o",
                "crypto/conf/libcrypto-shlib-conf_def.o",
                "crypto/conf/libcrypto-shlib-conf_err.o",
                "crypto/conf/libcrypto-shlib-conf_lib.o",
                "crypto/conf/libcrypto-shlib-conf_mall.o",
                "crypto/conf/libcrypto-shlib-conf_mod.o",
                "crypto/conf/libcrypto-shlib-conf_sap.o",
                "crypto/conf/libcrypto-shlib-conf_ssl.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/crmf" => {
            "deps" => [
                "crypto/crmf/libcrypto-lib-crmf_asn.o",
                "crypto/crmf/libcrypto-lib-crmf_err.o",
                "crypto/crmf/libcrypto-lib-crmf_lib.o",
                "crypto/crmf/libcrypto-lib-crmf_pbm.o",
                "crypto/crmf/libcrypto-shlib-crmf_asn.o",
                "crypto/crmf/libcrypto-shlib-crmf_err.o",
                "crypto/crmf/libcrypto-shlib-crmf_lib.o",
                "crypto/crmf/libcrypto-shlib-crmf_pbm.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ct" => {
            "deps" => [
                "crypto/ct/libcrypto-lib-ct_b64.o",
                "crypto/ct/libcrypto-lib-ct_err.o",
                "crypto/ct/libcrypto-lib-ct_log.o",
                "crypto/ct/libcrypto-lib-ct_oct.o",
                "crypto/ct/libcrypto-lib-ct_policy.o",
                "crypto/ct/libcrypto-lib-ct_prn.o",
                "crypto/ct/libcrypto-lib-ct_sct.o",
                "crypto/ct/libcrypto-lib-ct_sct_ctx.o",
                "crypto/ct/libcrypto-lib-ct_vfy.o",
                "crypto/ct/libcrypto-lib-ct_x509v3.o",
                "crypto/ct/libcrypto-shlib-ct_b64.o",
                "crypto/ct/libcrypto-shlib-ct_err.o",
                "crypto/ct/libcrypto-shlib-ct_log.o",
                "crypto/ct/libcrypto-shlib-ct_oct.o",
                "crypto/ct/libcrypto-shlib-ct_policy.o",
                "crypto/ct/libcrypto-shlib-ct_prn.o",
                "crypto/ct/libcrypto-shlib-ct_sct.o",
                "crypto/ct/libcrypto-shlib-ct_sct_ctx.o",
                "crypto/ct/libcrypto-shlib-ct_vfy.o",
                "crypto/ct/libcrypto-shlib-ct_x509v3.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/des" => {
            "deps" => [
                "crypto/des/libcrypto-lib-cbc_cksm.o",
                "crypto/des/libcrypto-lib-cbc_enc.o",
                "crypto/des/libcrypto-lib-cfb64ede.o",
                "crypto/des/libcrypto-lib-cfb64enc.o",
                "crypto/des/libcrypto-lib-cfb_enc.o",
                "crypto/des/libcrypto-lib-des_enc.o",
                "crypto/des/libcrypto-lib-ecb3_enc.o",
                "crypto/des/libcrypto-lib-ecb_enc.o",
                "crypto/des/libcrypto-lib-fcrypt.o",
                "crypto/des/libcrypto-lib-fcrypt_b.o",
                "crypto/des/libcrypto-lib-ofb64ede.o",
                "crypto/des/libcrypto-lib-ofb64enc.o",
                "crypto/des/libcrypto-lib-ofb_enc.o",
                "crypto/des/libcrypto-lib-pcbc_enc.o",
                "crypto/des/libcrypto-lib-qud_cksm.o",
                "crypto/des/libcrypto-lib-rand_key.o",
                "crypto/des/libcrypto-lib-set_key.o",
                "crypto/des/libcrypto-lib-str2key.o",
                "crypto/des/libcrypto-lib-xcbc_enc.o",
                "crypto/des/libcrypto-shlib-cbc_cksm.o",
                "crypto/des/libcrypto-shlib-cbc_enc.o",
                "crypto/des/libcrypto-shlib-cfb64ede.o",
                "crypto/des/libcrypto-shlib-cfb64enc.o",
                "crypto/des/libcrypto-shlib-cfb_enc.o",
                "crypto/des/libcrypto-shlib-des_enc.o",
                "crypto/des/libcrypto-shlib-ecb3_enc.o",
                "crypto/des/libcrypto-shlib-ecb_enc.o",
                "crypto/des/libcrypto-shlib-fcrypt.o",
                "crypto/des/libcrypto-shlib-fcrypt_b.o",
                "crypto/des/libcrypto-shlib-ofb64ede.o",
                "crypto/des/libcrypto-shlib-ofb64enc.o",
                "crypto/des/libcrypto-shlib-ofb_enc.o",
                "crypto/des/libcrypto-shlib-pcbc_enc.o",
                "crypto/des/libcrypto-shlib-qud_cksm.o",
                "crypto/des/libcrypto-shlib-rand_key.o",
                "crypto/des/libcrypto-shlib-set_key.o",
                "crypto/des/libcrypto-shlib-str2key.o",
                "crypto/des/libcrypto-shlib-xcbc_enc.o",
                "crypto/des/liblegacy-lib-des_enc.o",
                "crypto/des/liblegacy-lib-fcrypt_b.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "providers/liblegacy.a"
                ]
            }
        },
        "crypto/dh" => {
            "deps" => [
                "crypto/dh/libcrypto-lib-dh_ameth.o",
                "crypto/dh/libcrypto-lib-dh_asn1.o",
                "crypto/dh/libcrypto-lib-dh_backend.o",
                "crypto/dh/libcrypto-lib-dh_check.o",
                "crypto/dh/libcrypto-lib-dh_depr.o",
                "crypto/dh/libcrypto-lib-dh_err.o",
                "crypto/dh/libcrypto-lib-dh_gen.o",
                "crypto/dh/libcrypto-lib-dh_group_params.o",
                "crypto/dh/libcrypto-lib-dh_kdf.o",
                "crypto/dh/libcrypto-lib-dh_key.o",
                "crypto/dh/libcrypto-lib-dh_lib.o",
                "crypto/dh/libcrypto-lib-dh_meth.o",
                "crypto/dh/libcrypto-lib-dh_pmeth.o",
                "crypto/dh/libcrypto-lib-dh_prn.o",
                "crypto/dh/libcrypto-lib-dh_rfc5114.o",
                "crypto/dh/libcrypto-shlib-dh_ameth.o",
                "crypto/dh/libcrypto-shlib-dh_asn1.o",
                "crypto/dh/libcrypto-shlib-dh_backend.o",
                "crypto/dh/libcrypto-shlib-dh_check.o",
                "crypto/dh/libcrypto-shlib-dh_depr.o",
                "crypto/dh/libcrypto-shlib-dh_err.o",
                "crypto/dh/libcrypto-shlib-dh_gen.o",
                "crypto/dh/libcrypto-shlib-dh_group_params.o",
                "crypto/dh/libcrypto-shlib-dh_kdf.o",
                "crypto/dh/libcrypto-shlib-dh_key.o",
                "crypto/dh/libcrypto-shlib-dh_lib.o",
                "crypto/dh/libcrypto-shlib-dh_meth.o",
                "crypto/dh/libcrypto-shlib-dh_pmeth.o",
                "crypto/dh/libcrypto-shlib-dh_prn.o",
                "crypto/dh/libcrypto-shlib-dh_rfc5114.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/dsa" => {
            "deps" => [
                "crypto/dsa/libcrypto-lib-dsa_ameth.o",
                "crypto/dsa/libcrypto-lib-dsa_asn1.o",
                "crypto/dsa/libcrypto-lib-dsa_backend.o",
                "crypto/dsa/libcrypto-lib-dsa_check.o",
                "crypto/dsa/libcrypto-lib-dsa_depr.o",
                "crypto/dsa/libcrypto-lib-dsa_err.o",
                "crypto/dsa/libcrypto-lib-dsa_gen.o",
                "crypto/dsa/libcrypto-lib-dsa_key.o",
                "crypto/dsa/libcrypto-lib-dsa_lib.o",
                "crypto/dsa/libcrypto-lib-dsa_meth.o",
                "crypto/dsa/libcrypto-lib-dsa_ossl.o",
                "crypto/dsa/libcrypto-lib-dsa_pmeth.o",
                "crypto/dsa/libcrypto-lib-dsa_prn.o",
                "crypto/dsa/libcrypto-lib-dsa_sign.o",
                "crypto/dsa/libcrypto-lib-dsa_vrf.o",
                "crypto/dsa/libcrypto-shlib-dsa_ameth.o",
                "crypto/dsa/libcrypto-shlib-dsa_asn1.o",
                "crypto/dsa/libcrypto-shlib-dsa_backend.o",
                "crypto/dsa/libcrypto-shlib-dsa_check.o",
                "crypto/dsa/libcrypto-shlib-dsa_depr.o",
                "crypto/dsa/libcrypto-shlib-dsa_err.o",
                "crypto/dsa/libcrypto-shlib-dsa_gen.o",
                "crypto/dsa/libcrypto-shlib-dsa_key.o",
                "crypto/dsa/libcrypto-shlib-dsa_lib.o",
                "crypto/dsa/libcrypto-shlib-dsa_meth.o",
                "crypto/dsa/libcrypto-shlib-dsa_ossl.o",
                "crypto/dsa/libcrypto-shlib-dsa_pmeth.o",
                "crypto/dsa/libcrypto-shlib-dsa_prn.o",
                "crypto/dsa/libcrypto-shlib-dsa_sign.o",
                "crypto/dsa/libcrypto-shlib-dsa_vrf.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/dso" => {
            "deps" => [
                "crypto/dso/libcrypto-lib-dso_dl.o",
                "crypto/dso/libcrypto-lib-dso_dlfcn.o",
                "crypto/dso/libcrypto-lib-dso_err.o",
                "crypto/dso/libcrypto-lib-dso_lib.o",
                "crypto/dso/libcrypto-lib-dso_openssl.o",
                "crypto/dso/libcrypto-lib-dso_win32.o",
                "crypto/dso/libcrypto-shlib-dso_dl.o",
                "crypto/dso/libcrypto-shlib-dso_dlfcn.o",
                "crypto/dso/libcrypto-shlib-dso_err.o",
                "crypto/dso/libcrypto-shlib-dso_lib.o",
                "crypto/dso/libcrypto-shlib-dso_openssl.o",
                "crypto/dso/libcrypto-shlib-dso_win32.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ec" => {
            "deps" => [
                "crypto/ec/libcrypto-lib-curve25519.o",
                "crypto/ec/libcrypto-lib-ec2_oct.o",
                "crypto/ec/libcrypto-lib-ec2_smpl.o",
                "crypto/ec/libcrypto-lib-ec_ameth.o",
                "crypto/ec/libcrypto-lib-ec_asn1.o",
                "crypto/ec/libcrypto-lib-ec_backend.o",
                "crypto/ec/libcrypto-lib-ec_check.o",
                "crypto/ec/libcrypto-lib-ec_curve.o",
                "crypto/ec/libcrypto-lib-ec_cvt.o",
                "crypto/ec/libcrypto-lib-ec_deprecated.o",
                "crypto/ec/libcrypto-lib-ec_err.o",
                "crypto/ec/libcrypto-lib-ec_key.o",
                "crypto/ec/libcrypto-lib-ec_kmeth.o",
                "crypto/ec/libcrypto-lib-ec_lib.o",
                "crypto/ec/libcrypto-lib-ec_mult.o",
                "crypto/ec/libcrypto-lib-ec_oct.o",
                "crypto/ec/libcrypto-lib-ec_pmeth.o",
                "crypto/ec/libcrypto-lib-ec_print.o",
                "crypto/ec/libcrypto-lib-ecdh_kdf.o",
                "crypto/ec/libcrypto-lib-ecdh_ossl.o",
                "crypto/ec/libcrypto-lib-ecdsa_ossl.o",
                "crypto/ec/libcrypto-lib-ecdsa_sign.o",
                "crypto/ec/libcrypto-lib-ecdsa_vrf.o",
                "crypto/ec/libcrypto-lib-eck_prn.o",
                "crypto/ec/libcrypto-lib-ecp_meth.o",
                "crypto/ec/libcrypto-lib-ecp_mont.o",
                "crypto/ec/libcrypto-lib-ecp_nist.o",
                "crypto/ec/libcrypto-lib-ecp_nistz256-x86_64.o",
                "crypto/ec/libcrypto-lib-ecp_nistz256.o",
                "crypto/ec/libcrypto-lib-ecp_oct.o",
                "crypto/ec/libcrypto-lib-ecp_smpl.o",
                "crypto/ec/libcrypto-lib-ecx_backend.o",
                "crypto/ec/libcrypto-lib-ecx_key.o",
                "crypto/ec/libcrypto-lib-ecx_meth.o",
                "crypto/ec/libcrypto-lib-x25519-x86_64.o",
                "crypto/ec/libcrypto-shlib-curve25519.o",
                "crypto/ec/libcrypto-shlib-ec2_oct.o",
                "crypto/ec/libcrypto-shlib-ec2_smpl.o",
                "crypto/ec/libcrypto-shlib-ec_ameth.o",
                "crypto/ec/libcrypto-shlib-ec_asn1.o",
                "crypto/ec/libcrypto-shlib-ec_backend.o",
                "crypto/ec/libcrypto-shlib-ec_check.o",
                "crypto/ec/libcrypto-shlib-ec_curve.o",
                "crypto/ec/libcrypto-shlib-ec_cvt.o",
                "crypto/ec/libcrypto-shlib-ec_deprecated.o",
                "crypto/ec/libcrypto-shlib-ec_err.o",
                "crypto/ec/libcrypto-shlib-ec_key.o",
                "crypto/ec/libcrypto-shlib-ec_kmeth.o",
                "crypto/ec/libcrypto-shlib-ec_lib.o",
                "crypto/ec/libcrypto-shlib-ec_mult.o",
                "crypto/ec/libcrypto-shlib-ec_oct.o",
                "crypto/ec/libcrypto-shlib-ec_pmeth.o",
                "crypto/ec/libcrypto-shlib-ec_print.o",
                "crypto/ec/libcrypto-shlib-ecdh_kdf.o",
                "crypto/ec/libcrypto-shlib-ecdh_ossl.o",
                "crypto/ec/libcrypto-shlib-ecdsa_ossl.o",
                "crypto/ec/libcrypto-shlib-ecdsa_sign.o",
                "crypto/ec/libcrypto-shlib-ecdsa_vrf.o",
                "crypto/ec/libcrypto-shlib-eck_prn.o",
                "crypto/ec/libcrypto-shlib-ecp_meth.o",
                "crypto/ec/libcrypto-shlib-ecp_mont.o",
                "crypto/ec/libcrypto-shlib-ecp_nist.o",
                "crypto/ec/libcrypto-shlib-ecp_nistz256-x86_64.o",
                "crypto/ec/libcrypto-shlib-ecp_nistz256.o",
                "crypto/ec/libcrypto-shlib-ecp_oct.o",
                "crypto/ec/libcrypto-shlib-ecp_smpl.o",
                "crypto/ec/libcrypto-shlib-ecx_backend.o",
                "crypto/ec/libcrypto-shlib-ecx_key.o",
                "crypto/ec/libcrypto-shlib-ecx_meth.o",
                "crypto/ec/libcrypto-shlib-x25519-x86_64.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ec/curve448" => {
            "deps" => [
                "crypto/ec/curve448/libcrypto-lib-curve448.o",
                "crypto/ec/curve448/libcrypto-lib-curve448_tables.o",
                "crypto/ec/curve448/libcrypto-lib-eddsa.o",
                "crypto/ec/curve448/libcrypto-lib-f_generic.o",
                "crypto/ec/curve448/libcrypto-lib-scalar.o",
                "crypto/ec/curve448/libcrypto-shlib-curve448.o",
                "crypto/ec/curve448/libcrypto-shlib-curve448_tables.o",
                "crypto/ec/curve448/libcrypto-shlib-eddsa.o",
                "crypto/ec/curve448/libcrypto-shlib-f_generic.o",
                "crypto/ec/curve448/libcrypto-shlib-scalar.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ec/curve448/arch_32" => {
            "deps" => [
                "crypto/ec/curve448/arch_32/libcrypto-lib-f_impl32.o",
                "crypto/ec/curve448/arch_32/libcrypto-shlib-f_impl32.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ec/curve448/arch_64" => {
            "deps" => [
                "crypto/ec/curve448/arch_64/libcrypto-lib-f_impl64.o",
                "crypto/ec/curve448/arch_64/libcrypto-shlib-f_impl64.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/eia3" => {
            "deps" => [
                "crypto/eia3/libcrypto-lib-eia3.o",
                "crypto/eia3/libcrypto-shlib-eia3.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/encode_decode" => {
            "deps" => [
                "crypto/encode_decode/libcrypto-lib-decoder_err.o",
                "crypto/encode_decode/libcrypto-lib-decoder_lib.o",
                "crypto/encode_decode/libcrypto-lib-decoder_meth.o",
                "crypto/encode_decode/libcrypto-lib-decoder_pkey.o",
                "crypto/encode_decode/libcrypto-lib-encoder_err.o",
                "crypto/encode_decode/libcrypto-lib-encoder_lib.o",
                "crypto/encode_decode/libcrypto-lib-encoder_meth.o",
                "crypto/encode_decode/libcrypto-lib-encoder_pkey.o",
                "crypto/encode_decode/libcrypto-shlib-decoder_err.o",
                "crypto/encode_decode/libcrypto-shlib-decoder_lib.o",
                "crypto/encode_decode/libcrypto-shlib-decoder_meth.o",
                "crypto/encode_decode/libcrypto-shlib-decoder_pkey.o",
                "crypto/encode_decode/libcrypto-shlib-encoder_err.o",
                "crypto/encode_decode/libcrypto-shlib-encoder_lib.o",
                "crypto/encode_decode/libcrypto-shlib-encoder_meth.o",
                "crypto/encode_decode/libcrypto-shlib-encoder_pkey.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/engine" => {
            "deps" => [
                "crypto/engine/libcrypto-lib-eng_all.o",
                "crypto/engine/libcrypto-lib-eng_cnf.o",
                "crypto/engine/libcrypto-lib-eng_ctrl.o",
                "crypto/engine/libcrypto-lib-eng_dyn.o",
                "crypto/engine/libcrypto-lib-eng_err.o",
                "crypto/engine/libcrypto-lib-eng_fat.o",
                "crypto/engine/libcrypto-lib-eng_init.o",
                "crypto/engine/libcrypto-lib-eng_lib.o",
                "crypto/engine/libcrypto-lib-eng_list.o",
                "crypto/engine/libcrypto-lib-eng_openssl.o",
                "crypto/engine/libcrypto-lib-eng_pkey.o",
                "crypto/engine/libcrypto-lib-eng_rdrand.o",
                "crypto/engine/libcrypto-lib-eng_table.o",
                "crypto/engine/libcrypto-lib-tb_asnmth.o",
                "crypto/engine/libcrypto-lib-tb_cipher.o",
                "crypto/engine/libcrypto-lib-tb_dh.o",
                "crypto/engine/libcrypto-lib-tb_digest.o",
                "crypto/engine/libcrypto-lib-tb_dsa.o",
                "crypto/engine/libcrypto-lib-tb_eckey.o",
                "crypto/engine/libcrypto-lib-tb_ecpmeth.o",
                "crypto/engine/libcrypto-lib-tb_pkmeth.o",
                "crypto/engine/libcrypto-lib-tb_rand.o",
                "crypto/engine/libcrypto-lib-tb_rsa.o",
                "crypto/engine/libcrypto-shlib-eng_all.o",
                "crypto/engine/libcrypto-shlib-eng_cnf.o",
                "crypto/engine/libcrypto-shlib-eng_ctrl.o",
                "crypto/engine/libcrypto-shlib-eng_dyn.o",
                "crypto/engine/libcrypto-shlib-eng_err.o",
                "crypto/engine/libcrypto-shlib-eng_fat.o",
                "crypto/engine/libcrypto-shlib-eng_init.o",
                "crypto/engine/libcrypto-shlib-eng_lib.o",
                "crypto/engine/libcrypto-shlib-eng_list.o",
                "crypto/engine/libcrypto-shlib-eng_openssl.o",
                "crypto/engine/libcrypto-shlib-eng_pkey.o",
                "crypto/engine/libcrypto-shlib-eng_rdrand.o",
                "crypto/engine/libcrypto-shlib-eng_table.o",
                "crypto/engine/libcrypto-shlib-tb_asnmth.o",
                "crypto/engine/libcrypto-shlib-tb_cipher.o",
                "crypto/engine/libcrypto-shlib-tb_dh.o",
                "crypto/engine/libcrypto-shlib-tb_digest.o",
                "crypto/engine/libcrypto-shlib-tb_dsa.o",
                "crypto/engine/libcrypto-shlib-tb_eckey.o",
                "crypto/engine/libcrypto-shlib-tb_ecpmeth.o",
                "crypto/engine/libcrypto-shlib-tb_pkmeth.o",
                "crypto/engine/libcrypto-shlib-tb_rand.o",
                "crypto/engine/libcrypto-shlib-tb_rsa.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/err" => {
            "deps" => [
                "crypto/err/libcrypto-lib-err.o",
                "crypto/err/libcrypto-lib-err_all.o",
                "crypto/err/libcrypto-lib-err_all_legacy.o",
                "crypto/err/libcrypto-lib-err_blocks.o",
                "crypto/err/libcrypto-lib-err_mark.o",
                "crypto/err/libcrypto-lib-err_prn.o",
                "crypto/err/libcrypto-lib-err_save.o",
                "crypto/err/libcrypto-shlib-err.o",
                "crypto/err/libcrypto-shlib-err_all.o",
                "crypto/err/libcrypto-shlib-err_all_legacy.o",
                "crypto/err/libcrypto-shlib-err_blocks.o",
                "crypto/err/libcrypto-shlib-err_mark.o",
                "crypto/err/libcrypto-shlib-err_prn.o",
                "crypto/err/libcrypto-shlib-err_save.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ess" => {
            "deps" => [
                "crypto/ess/libcrypto-lib-ess_asn1.o",
                "crypto/ess/libcrypto-lib-ess_err.o",
                "crypto/ess/libcrypto-lib-ess_lib.o",
                "crypto/ess/libcrypto-shlib-ess_asn1.o",
                "crypto/ess/libcrypto-shlib-ess_err.o",
                "crypto/ess/libcrypto-shlib-ess_lib.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/evp" => {
            "deps" => [
                "crypto/evp/libcrypto-lib-asymcipher.o",
                "crypto/evp/libcrypto-lib-bio_b64.o",
                "crypto/evp/libcrypto-lib-bio_enc.o",
                "crypto/evp/libcrypto-lib-bio_md.o",
                "crypto/evp/libcrypto-lib-bio_ok.o",
                "crypto/evp/libcrypto-lib-c_allc.o",
                "crypto/evp/libcrypto-lib-c_alld.o",
                "crypto/evp/libcrypto-lib-cmeth_lib.o",
                "crypto/evp/libcrypto-lib-ctrl_params_translate.o",
                "crypto/evp/libcrypto-lib-dh_ctrl.o",
                "crypto/evp/libcrypto-lib-dh_support.o",
                "crypto/evp/libcrypto-lib-digest.o",
                "crypto/evp/libcrypto-lib-dsa_ctrl.o",
                "crypto/evp/libcrypto-lib-e_aes.o",
                "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha1.o",
                "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha256.o",
                "crypto/evp/libcrypto-lib-e_chacha20_poly1305.o",
                "crypto/evp/libcrypto-lib-e_des.o",
                "crypto/evp/libcrypto-lib-e_des3.o",
                "crypto/evp/libcrypto-lib-e_eea3.o",
                "crypto/evp/libcrypto-lib-e_null.o",
                "crypto/evp/libcrypto-lib-e_old.o",
                "crypto/evp/libcrypto-lib-e_rc4.o",
                "crypto/evp/libcrypto-lib-e_rc4_hmac_md5.o",
                "crypto/evp/libcrypto-lib-e_rc5.o",
                "crypto/evp/libcrypto-lib-e_sm4.o",
                "crypto/evp/libcrypto-lib-e_wbsm4_baiwu.o",
                "crypto/evp/libcrypto-lib-e_wbsm4_wsise.o",
                "crypto/evp/libcrypto-lib-e_wbsm4_xiaolai.o",
                "crypto/evp/libcrypto-lib-e_xcbc_d.o",
                "crypto/evp/libcrypto-lib-ec_ctrl.o",
                "crypto/evp/libcrypto-lib-ec_support.o",
                "crypto/evp/libcrypto-lib-encode.o",
                "crypto/evp/libcrypto-lib-evp_cnf.o",
                "crypto/evp/libcrypto-lib-evp_enc.o",
                "crypto/evp/libcrypto-lib-evp_err.o",
                "crypto/evp/libcrypto-lib-evp_fetch.o",
                "crypto/evp/libcrypto-lib-evp_key.o",
                "crypto/evp/libcrypto-lib-evp_lib.o",
                "crypto/evp/libcrypto-lib-evp_pbe.o",
                "crypto/evp/libcrypto-lib-evp_pkey.o",
                "crypto/evp/libcrypto-lib-evp_rand.o",
                "crypto/evp/libcrypto-lib-evp_utils.o",
                "crypto/evp/libcrypto-lib-exchange.o",
                "crypto/evp/libcrypto-lib-kdf_lib.o",
                "crypto/evp/libcrypto-lib-kdf_meth.o",
                "crypto/evp/libcrypto-lib-kem.o",
                "crypto/evp/libcrypto-lib-keymgmt_lib.o",
                "crypto/evp/libcrypto-lib-keymgmt_meth.o",
                "crypto/evp/libcrypto-lib-legacy_md5.o",
                "crypto/evp/libcrypto-lib-legacy_md5_sha1.o",
                "crypto/evp/libcrypto-lib-legacy_sha.o",
                "crypto/evp/libcrypto-lib-m_null.o",
                "crypto/evp/libcrypto-lib-m_sigver.o",
                "crypto/evp/libcrypto-lib-mac_lib.o",
                "crypto/evp/libcrypto-lib-mac_meth.o",
                "crypto/evp/libcrypto-lib-names.o",
                "crypto/evp/libcrypto-lib-p5_crpt.o",
                "crypto/evp/libcrypto-lib-p5_crpt2.o",
                "crypto/evp/libcrypto-lib-p_dec.o",
                "crypto/evp/libcrypto-lib-p_enc.o",
                "crypto/evp/libcrypto-lib-p_legacy.o",
                "crypto/evp/libcrypto-lib-p_lib.o",
                "crypto/evp/libcrypto-lib-p_open.o",
                "crypto/evp/libcrypto-lib-p_seal.o",
                "crypto/evp/libcrypto-lib-p_sign.o",
                "crypto/evp/libcrypto-lib-p_verify.o",
                "crypto/evp/libcrypto-lib-pbe_scrypt.o",
                "crypto/evp/libcrypto-lib-pmeth_check.o",
                "crypto/evp/libcrypto-lib-pmeth_gn.o",
                "crypto/evp/libcrypto-lib-pmeth_lib.o",
                "crypto/evp/libcrypto-lib-s_lib.o",
                "crypto/evp/libcrypto-lib-signature.o",
                "crypto/evp/libcrypto-lib-skeymgmt_meth.o",
                "crypto/evp/libcrypto-shlib-asymcipher.o",
                "crypto/evp/libcrypto-shlib-bio_b64.o",
                "crypto/evp/libcrypto-shlib-bio_enc.o",
                "crypto/evp/libcrypto-shlib-bio_md.o",
                "crypto/evp/libcrypto-shlib-bio_ok.o",
                "crypto/evp/libcrypto-shlib-c_allc.o",
                "crypto/evp/libcrypto-shlib-c_alld.o",
                "crypto/evp/libcrypto-shlib-cmeth_lib.o",
                "crypto/evp/libcrypto-shlib-ctrl_params_translate.o",
                "crypto/evp/libcrypto-shlib-dh_ctrl.o",
                "crypto/evp/libcrypto-shlib-dh_support.o",
                "crypto/evp/libcrypto-shlib-digest.o",
                "crypto/evp/libcrypto-shlib-dsa_ctrl.o",
                "crypto/evp/libcrypto-shlib-e_aes.o",
                "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha1.o",
                "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha256.o",
                "crypto/evp/libcrypto-shlib-e_chacha20_poly1305.o",
                "crypto/evp/libcrypto-shlib-e_des.o",
                "crypto/evp/libcrypto-shlib-e_des3.o",
                "crypto/evp/libcrypto-shlib-e_eea3.o",
                "crypto/evp/libcrypto-shlib-e_null.o",
                "crypto/evp/libcrypto-shlib-e_old.o",
                "crypto/evp/libcrypto-shlib-e_rc4.o",
                "crypto/evp/libcrypto-shlib-e_rc4_hmac_md5.o",
                "crypto/evp/libcrypto-shlib-e_rc5.o",
                "crypto/evp/libcrypto-shlib-e_sm4.o",
                "crypto/evp/libcrypto-shlib-e_wbsm4_baiwu.o",
                "crypto/evp/libcrypto-shlib-e_wbsm4_wsise.o",
                "crypto/evp/libcrypto-shlib-e_wbsm4_xiaolai.o",
                "crypto/evp/libcrypto-shlib-e_xcbc_d.o",
                "crypto/evp/libcrypto-shlib-ec_ctrl.o",
                "crypto/evp/libcrypto-shlib-ec_support.o",
                "crypto/evp/libcrypto-shlib-encode.o",
                "crypto/evp/libcrypto-shlib-evp_cnf.o",
                "crypto/evp/libcrypto-shlib-evp_enc.o",
                "crypto/evp/libcrypto-shlib-evp_err.o",
                "crypto/evp/libcrypto-shlib-evp_fetch.o",
                "crypto/evp/libcrypto-shlib-evp_key.o",
                "crypto/evp/libcrypto-shlib-evp_lib.o",
                "crypto/evp/libcrypto-shlib-evp_pbe.o",
                "crypto/evp/libcrypto-shlib-evp_pkey.o",
                "crypto/evp/libcrypto-shlib-evp_rand.o",
                "crypto/evp/libcrypto-shlib-evp_utils.o",
                "crypto/evp/libcrypto-shlib-exchange.o",
                "crypto/evp/libcrypto-shlib-kdf_lib.o",
                "crypto/evp/libcrypto-shlib-kdf_meth.o",
                "crypto/evp/libcrypto-shlib-kem.o",
                "crypto/evp/libcrypto-shlib-keymgmt_lib.o",
                "crypto/evp/libcrypto-shlib-keymgmt_meth.o",
                "crypto/evp/libcrypto-shlib-legacy_md5.o",
                "crypto/evp/libcrypto-shlib-legacy_md5_sha1.o",
                "crypto/evp/libcrypto-shlib-legacy_sha.o",
                "crypto/evp/libcrypto-shlib-m_null.o",
                "crypto/evp/libcrypto-shlib-m_sigver.o",
                "crypto/evp/libcrypto-shlib-mac_lib.o",
                "crypto/evp/libcrypto-shlib-mac_meth.o",
                "crypto/evp/libcrypto-shlib-names.o",
                "crypto/evp/libcrypto-shlib-p5_crpt.o",
                "crypto/evp/libcrypto-shlib-p5_crpt2.o",
                "crypto/evp/libcrypto-shlib-p_dec.o",
                "crypto/evp/libcrypto-shlib-p_enc.o",
                "crypto/evp/libcrypto-shlib-p_legacy.o",
                "crypto/evp/libcrypto-shlib-p_lib.o",
                "crypto/evp/libcrypto-shlib-p_open.o",
                "crypto/evp/libcrypto-shlib-p_seal.o",
                "crypto/evp/libcrypto-shlib-p_sign.o",
                "crypto/evp/libcrypto-shlib-p_verify.o",
                "crypto/evp/libcrypto-shlib-pbe_scrypt.o",
                "crypto/evp/libcrypto-shlib-pmeth_check.o",
                "crypto/evp/libcrypto-shlib-pmeth_gn.o",
                "crypto/evp/libcrypto-shlib-pmeth_lib.o",
                "crypto/evp/libcrypto-shlib-s_lib.o",
                "crypto/evp/libcrypto-shlib-signature.o",
                "crypto/evp/libcrypto-shlib-skeymgmt_meth.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ffc" => {
            "deps" => [
                "crypto/ffc/libcrypto-lib-ffc_backend.o",
                "crypto/ffc/libcrypto-lib-ffc_dh.o",
                "crypto/ffc/libcrypto-lib-ffc_key_generate.o",
                "crypto/ffc/libcrypto-lib-ffc_key_validate.o",
                "crypto/ffc/libcrypto-lib-ffc_params.o",
                "crypto/ffc/libcrypto-lib-ffc_params_generate.o",
                "crypto/ffc/libcrypto-lib-ffc_params_validate.o",
                "crypto/ffc/libcrypto-shlib-ffc_backend.o",
                "crypto/ffc/libcrypto-shlib-ffc_dh.o",
                "crypto/ffc/libcrypto-shlib-ffc_key_generate.o",
                "crypto/ffc/libcrypto-shlib-ffc_key_validate.o",
                "crypto/ffc/libcrypto-shlib-ffc_params.o",
                "crypto/ffc/libcrypto-shlib-ffc_params_generate.o",
                "crypto/ffc/libcrypto-shlib-ffc_params_validate.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/hashtable" => {
            "deps" => [
                "crypto/hashtable/libcrypto-lib-hashfunc.o",
                "crypto/hashtable/libcrypto-lib-hashtable.o",
                "crypto/hashtable/libcrypto-shlib-hashfunc.o",
                "crypto/hashtable/libcrypto-shlib-hashtable.o",
                "crypto/hashtable/libssl-shlib-hashfunc.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "libssl"
                ]
            }
        },
        "crypto/hmac" => {
            "deps" => [
                "crypto/hmac/libcrypto-lib-hmac.o",
                "crypto/hmac/libcrypto-shlib-hmac.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/hpke" => {
            "deps" => [
                "crypto/hpke/libcrypto-lib-hpke.o",
                "crypto/hpke/libcrypto-lib-hpke_util.o",
                "crypto/hpke/libcrypto-shlib-hpke.o",
                "crypto/hpke/libcrypto-shlib-hpke_util.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/http" => {
            "deps" => [
                "crypto/http/libcrypto-lib-http_client.o",
                "crypto/http/libcrypto-lib-http_err.o",
                "crypto/http/libcrypto-lib-http_lib.o",
                "crypto/http/libcrypto-shlib-http_client.o",
                "crypto/http/libcrypto-shlib-http_err.o",
                "crypto/http/libcrypto-shlib-http_lib.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/kdf" => {
            "deps" => [
                "crypto/kdf/libcrypto-lib-kdf_err.o",
                "crypto/kdf/libcrypto-shlib-kdf_err.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/lhash" => {
            "deps" => [
                "crypto/lhash/libcrypto-lib-lh_stats.o",
                "crypto/lhash/libcrypto-lib-lhash.o",
                "crypto/lhash/libcrypto-shlib-lh_stats.o",
                "crypto/lhash/libcrypto-shlib-lhash.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/md5" => {
            "deps" => [
                "crypto/md5/libcrypto-lib-md5-x86_64.o",
                "crypto/md5/libcrypto-lib-md5_dgst.o",
                "crypto/md5/libcrypto-lib-md5_one.o",
                "crypto/md5/libcrypto-lib-md5_sha1.o",
                "crypto/md5/libcrypto-shlib-md5-x86_64.o",
                "crypto/md5/libcrypto-shlib-md5_dgst.o",
                "crypto/md5/libcrypto-shlib-md5_one.o",
                "crypto/md5/libcrypto-shlib-md5_sha1.o",
                "crypto/md5/liblegacy-lib-md5-x86_64.o",
                "crypto/md5/liblegacy-lib-md5_dgst.o",
                "crypto/md5/liblegacy-lib-md5_one.o",
                "crypto/md5/liblegacy-lib-md5_sha1.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "providers/liblegacy.a"
                ]
            }
        },
        "crypto/ml_dsa" => {
            "deps" => [
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_encoders.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_key.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_key_compress.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_matrix.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_ntt.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_params.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_sample.o",
                "crypto/ml_dsa/libcrypto-lib-ml_dsa_sign.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_encoders.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key_compress.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_matrix.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_ntt.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_params.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sample.o",
                "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sign.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ml_kem" => {
            "deps" => [
                "crypto/ml_kem/libcrypto-lib-ml_kem.o",
                "crypto/ml_kem/libcrypto-shlib-ml_kem.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/modes" => {
            "deps" => [
                "crypto/modes/libcrypto-lib-aes-gcm-avx512.o",
                "crypto/modes/libcrypto-lib-aesni-gcm-x86_64.o",
                "crypto/modes/libcrypto-lib-cbc128.o",
                "crypto/modes/libcrypto-lib-ccm128.o",
                "crypto/modes/libcrypto-lib-cfb128.o",
                "crypto/modes/libcrypto-lib-ctr128.o",
                "crypto/modes/libcrypto-lib-cts128.o",
                "crypto/modes/libcrypto-lib-gcm128.o",
                "crypto/modes/libcrypto-lib-ghash-x86_64.o",
                "crypto/modes/libcrypto-lib-ocb128.o",
                "crypto/modes/libcrypto-lib-ofb128.o",
                "crypto/modes/libcrypto-lib-siv128.o",
                "crypto/modes/libcrypto-lib-wrap128.o",
                "crypto/modes/libcrypto-lib-xts128.o",
                "crypto/modes/libcrypto-lib-xts128gb.o",
                "crypto/modes/libcrypto-shlib-aes-gcm-avx512.o",
                "crypto/modes/libcrypto-shlib-aesni-gcm-x86_64.o",
                "crypto/modes/libcrypto-shlib-cbc128.o",
                "crypto/modes/libcrypto-shlib-ccm128.o",
                "crypto/modes/libcrypto-shlib-cfb128.o",
                "crypto/modes/libcrypto-shlib-ctr128.o",
                "crypto/modes/libcrypto-shlib-cts128.o",
                "crypto/modes/libcrypto-shlib-gcm128.o",
                "crypto/modes/libcrypto-shlib-ghash-x86_64.o",
                "crypto/modes/libcrypto-shlib-ocb128.o",
                "crypto/modes/libcrypto-shlib-ofb128.o",
                "crypto/modes/libcrypto-shlib-siv128.o",
                "crypto/modes/libcrypto-shlib-wrap128.o",
                "crypto/modes/libcrypto-shlib-xts128.o",
                "crypto/modes/libcrypto-shlib-xts128gb.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/objects" => {
            "deps" => [
                "crypto/objects/libcrypto-lib-o_names.o",
                "crypto/objects/libcrypto-lib-obj_dat.o",
                "crypto/objects/libcrypto-lib-obj_err.o",
                "crypto/objects/libcrypto-lib-obj_lib.o",
                "crypto/objects/libcrypto-lib-obj_xref.o",
                "crypto/objects/libcrypto-shlib-o_names.o",
                "crypto/objects/libcrypto-shlib-obj_dat.o",
                "crypto/objects/libcrypto-shlib-obj_err.o",
                "crypto/objects/libcrypto-shlib-obj_lib.o",
                "crypto/objects/libcrypto-shlib-obj_xref.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ocsp" => {
            "deps" => [
                "crypto/ocsp/libcrypto-lib-ocsp_asn.o",
                "crypto/ocsp/libcrypto-lib-ocsp_cl.o",
                "crypto/ocsp/libcrypto-lib-ocsp_err.o",
                "crypto/ocsp/libcrypto-lib-ocsp_ext.o",
                "crypto/ocsp/libcrypto-lib-ocsp_http.o",
                "crypto/ocsp/libcrypto-lib-ocsp_lib.o",
                "crypto/ocsp/libcrypto-lib-ocsp_prn.o",
                "crypto/ocsp/libcrypto-lib-ocsp_srv.o",
                "crypto/ocsp/libcrypto-lib-ocsp_vfy.o",
                "crypto/ocsp/libcrypto-lib-v3_ocsp.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_asn.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_cl.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_err.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_ext.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_http.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_lib.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_prn.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_srv.o",
                "crypto/ocsp/libcrypto-shlib-ocsp_vfy.o",
                "crypto/ocsp/libcrypto-shlib-v3_ocsp.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/pem" => {
            "deps" => [
                "crypto/pem/loader_attic-dso-pvkfmt.o",
                "crypto/pem/libcrypto-lib-pem_all.o",
                "crypto/pem/libcrypto-lib-pem_err.o",
                "crypto/pem/libcrypto-lib-pem_info.o",
                "crypto/pem/libcrypto-lib-pem_lib.o",
                "crypto/pem/libcrypto-lib-pem_oth.o",
                "crypto/pem/libcrypto-lib-pem_pk8.o",
                "crypto/pem/libcrypto-lib-pem_pkey.o",
                "crypto/pem/libcrypto-lib-pem_sign.o",
                "crypto/pem/libcrypto-lib-pem_x509.o",
                "crypto/pem/libcrypto-lib-pem_xaux.o",
                "crypto/pem/libcrypto-lib-pvkfmt.o",
                "crypto/pem/libcrypto-shlib-pem_all.o",
                "crypto/pem/libcrypto-shlib-pem_err.o",
                "crypto/pem/libcrypto-shlib-pem_info.o",
                "crypto/pem/libcrypto-shlib-pem_lib.o",
                "crypto/pem/libcrypto-shlib-pem_oth.o",
                "crypto/pem/libcrypto-shlib-pem_pk8.o",
                "crypto/pem/libcrypto-shlib-pem_pkey.o",
                "crypto/pem/libcrypto-shlib-pem_sign.o",
                "crypto/pem/libcrypto-shlib-pem_x509.o",
                "crypto/pem/libcrypto-shlib-pem_xaux.o",
                "crypto/pem/libcrypto-shlib-pvkfmt.o"
            ],
            "products" => {
                "dso" => [
                    "engines/loader_attic"
                ],
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/pkcs12" => {
            "deps" => [
                "crypto/pkcs12/libcrypto-lib-p12_add.o",
                "crypto/pkcs12/libcrypto-lib-p12_asn.o",
                "crypto/pkcs12/libcrypto-lib-p12_attr.o",
                "crypto/pkcs12/libcrypto-lib-p12_crpt.o",
                "crypto/pkcs12/libcrypto-lib-p12_crt.o",
                "crypto/pkcs12/libcrypto-lib-p12_decr.o",
                "crypto/pkcs12/libcrypto-lib-p12_init.o",
                "crypto/pkcs12/libcrypto-lib-p12_key.o",
                "crypto/pkcs12/libcrypto-lib-p12_kiss.o",
                "crypto/pkcs12/libcrypto-lib-p12_mutl.o",
                "crypto/pkcs12/libcrypto-lib-p12_npas.o",
                "crypto/pkcs12/libcrypto-lib-p12_p8d.o",
                "crypto/pkcs12/libcrypto-lib-p12_p8e.o",
                "crypto/pkcs12/libcrypto-lib-p12_sbag.o",
                "crypto/pkcs12/libcrypto-lib-p12_utl.o",
                "crypto/pkcs12/libcrypto-lib-pk12err.o",
                "crypto/pkcs12/libcrypto-shlib-p12_add.o",
                "crypto/pkcs12/libcrypto-shlib-p12_asn.o",
                "crypto/pkcs12/libcrypto-shlib-p12_attr.o",
                "crypto/pkcs12/libcrypto-shlib-p12_crpt.o",
                "crypto/pkcs12/libcrypto-shlib-p12_crt.o",
                "crypto/pkcs12/libcrypto-shlib-p12_decr.o",
                "crypto/pkcs12/libcrypto-shlib-p12_init.o",
                "crypto/pkcs12/libcrypto-shlib-p12_key.o",
                "crypto/pkcs12/libcrypto-shlib-p12_kiss.o",
                "crypto/pkcs12/libcrypto-shlib-p12_mutl.o",
                "crypto/pkcs12/libcrypto-shlib-p12_npas.o",
                "crypto/pkcs12/libcrypto-shlib-p12_p8d.o",
                "crypto/pkcs12/libcrypto-shlib-p12_p8e.o",
                "crypto/pkcs12/libcrypto-shlib-p12_sbag.o",
                "crypto/pkcs12/libcrypto-shlib-p12_utl.o",
                "crypto/pkcs12/libcrypto-shlib-pk12err.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/pkcs7" => {
            "deps" => [
                "crypto/pkcs7/libcrypto-lib-bio_pk7.o",
                "crypto/pkcs7/libcrypto-lib-pk7_asn1.o",
                "crypto/pkcs7/libcrypto-lib-pk7_attr.o",
                "crypto/pkcs7/libcrypto-lib-pk7_doit.o",
                "crypto/pkcs7/libcrypto-lib-pk7_lib.o",
                "crypto/pkcs7/libcrypto-lib-pk7_mime.o",
                "crypto/pkcs7/libcrypto-lib-pk7_smime.o",
                "crypto/pkcs7/libcrypto-lib-pkcs7err.o",
                "crypto/pkcs7/libcrypto-shlib-bio_pk7.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_asn1.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_attr.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_doit.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_lib.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_mime.o",
                "crypto/pkcs7/libcrypto-shlib-pk7_smime.o",
                "crypto/pkcs7/libcrypto-shlib-pkcs7err.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/poly1305" => {
            "deps" => [
                "crypto/poly1305/libcrypto-lib-poly1305-x86_64.o",
                "crypto/poly1305/libcrypto-lib-poly1305.o",
                "crypto/poly1305/libcrypto-shlib-poly1305-x86_64.o",
                "crypto/poly1305/libcrypto-shlib-poly1305.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/property" => {
            "deps" => [
                "crypto/property/libcrypto-lib-defn_cache.o",
                "crypto/property/libcrypto-lib-property.o",
                "crypto/property/libcrypto-lib-property_err.o",
                "crypto/property/libcrypto-lib-property_parse.o",
                "crypto/property/libcrypto-lib-property_query.o",
                "crypto/property/libcrypto-lib-property_string.o",
                "crypto/property/libcrypto-shlib-defn_cache.o",
                "crypto/property/libcrypto-shlib-property.o",
                "crypto/property/libcrypto-shlib-property_err.o",
                "crypto/property/libcrypto-shlib-property_parse.o",
                "crypto/property/libcrypto-shlib-property_query.o",
                "crypto/property/libcrypto-shlib-property_string.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/rand" => {
            "deps" => [
                "crypto/rand/libcrypto-lib-prov_seed.o",
                "crypto/rand/libcrypto-lib-rand_deprecated.o",
                "crypto/rand/libcrypto-lib-rand_err.o",
                "crypto/rand/libcrypto-lib-rand_lib.o",
                "crypto/rand/libcrypto-lib-rand_meth.o",
                "crypto/rand/libcrypto-lib-rand_pool.o",
                "crypto/rand/libcrypto-lib-rand_uniform.o",
                "crypto/rand/libcrypto-lib-randfile.o",
                "crypto/rand/libcrypto-shlib-prov_seed.o",
                "crypto/rand/libcrypto-shlib-rand_deprecated.o",
                "crypto/rand/libcrypto-shlib-rand_err.o",
                "crypto/rand/libcrypto-shlib-rand_lib.o",
                "crypto/rand/libcrypto-shlib-rand_meth.o",
                "crypto/rand/libcrypto-shlib-rand_pool.o",
                "crypto/rand/libcrypto-shlib-rand_uniform.o",
                "crypto/rand/libcrypto-shlib-randfile.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/rc4" => {
            "deps" => [
                "crypto/rc4/libcrypto-lib-rc4-md5-x86_64.o",
                "crypto/rc4/libcrypto-lib-rc4-x86_64.o",
                "crypto/rc4/libcrypto-shlib-rc4-md5-x86_64.o",
                "crypto/rc4/libcrypto-shlib-rc4-x86_64.o",
                "crypto/rc4/liblegacy-lib-rc4-md5-x86_64.o",
                "crypto/rc4/liblegacy-lib-rc4-x86_64.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "providers/liblegacy.a"
                ]
            }
        },
        "crypto/rsa" => {
            "deps" => [
                "crypto/rsa/libcrypto-lib-rsa_ameth.o",
                "crypto/rsa/libcrypto-lib-rsa_asn1.o",
                "crypto/rsa/libcrypto-lib-rsa_backend.o",
                "crypto/rsa/libcrypto-lib-rsa_chk.o",
                "crypto/rsa/libcrypto-lib-rsa_crpt.o",
                "crypto/rsa/libcrypto-lib-rsa_depr.o",
                "crypto/rsa/libcrypto-lib-rsa_err.o",
                "crypto/rsa/libcrypto-lib-rsa_gen.o",
                "crypto/rsa/libcrypto-lib-rsa_lib.o",
                "crypto/rsa/libcrypto-lib-rsa_meth.o",
                "crypto/rsa/libcrypto-lib-rsa_mp.o",
                "crypto/rsa/libcrypto-lib-rsa_mp_names.o",
                "crypto/rsa/libcrypto-lib-rsa_none.o",
                "crypto/rsa/libcrypto-lib-rsa_oaep.o",
                "crypto/rsa/libcrypto-lib-rsa_ossl.o",
                "crypto/rsa/libcrypto-lib-rsa_pk1.o",
                "crypto/rsa/libcrypto-lib-rsa_pmeth.o",
                "crypto/rsa/libcrypto-lib-rsa_prn.o",
                "crypto/rsa/libcrypto-lib-rsa_pss.o",
                "crypto/rsa/libcrypto-lib-rsa_saos.o",
                "crypto/rsa/libcrypto-lib-rsa_schemes.o",
                "crypto/rsa/libcrypto-lib-rsa_sign.o",
                "crypto/rsa/libcrypto-lib-rsa_sp800_56b_check.o",
                "crypto/rsa/libcrypto-lib-rsa_sp800_56b_gen.o",
                "crypto/rsa/libcrypto-lib-rsa_x931.o",
                "crypto/rsa/libcrypto-lib-rsa_x931g.o",
                "crypto/rsa/libcrypto-shlib-rsa_ameth.o",
                "crypto/rsa/libcrypto-shlib-rsa_asn1.o",
                "crypto/rsa/libcrypto-shlib-rsa_backend.o",
                "crypto/rsa/libcrypto-shlib-rsa_chk.o",
                "crypto/rsa/libcrypto-shlib-rsa_crpt.o",
                "crypto/rsa/libcrypto-shlib-rsa_depr.o",
                "crypto/rsa/libcrypto-shlib-rsa_err.o",
                "crypto/rsa/libcrypto-shlib-rsa_gen.o",
                "crypto/rsa/libcrypto-shlib-rsa_lib.o",
                "crypto/rsa/libcrypto-shlib-rsa_meth.o",
                "crypto/rsa/libcrypto-shlib-rsa_mp.o",
                "crypto/rsa/libcrypto-shlib-rsa_mp_names.o",
                "crypto/rsa/libcrypto-shlib-rsa_none.o",
                "crypto/rsa/libcrypto-shlib-rsa_oaep.o",
                "crypto/rsa/libcrypto-shlib-rsa_ossl.o",
                "crypto/rsa/libcrypto-shlib-rsa_pk1.o",
                "crypto/rsa/libcrypto-shlib-rsa_pmeth.o",
                "crypto/rsa/libcrypto-shlib-rsa_prn.o",
                "crypto/rsa/libcrypto-shlib-rsa_pss.o",
                "crypto/rsa/libcrypto-shlib-rsa_saos.o",
                "crypto/rsa/libcrypto-shlib-rsa_schemes.o",
                "crypto/rsa/libcrypto-shlib-rsa_sign.o",
                "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_check.o",
                "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_gen.o",
                "crypto/rsa/libcrypto-shlib-rsa_x931.o",
                "crypto/rsa/libcrypto-shlib-rsa_x931g.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/sdf" => {
            "deps" => [
                "crypto/sdf/libcrypto-lib-sdf_lib.o",
                "crypto/sdf/libcrypto-lib-sdf_meth.o",
                "crypto/sdf/libcrypto-shlib-sdf_lib.o",
                "crypto/sdf/libcrypto-shlib-sdf_meth.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/sha" => {
            "deps" => [
                "crypto/sha/libcrypto-lib-keccak1600-x86_64.o",
                "crypto/sha/libcrypto-lib-sha1-mb-x86_64.o",
                "crypto/sha/libcrypto-lib-sha1-x86_64.o",
                "crypto/sha/libcrypto-lib-sha1_one.o",
                "crypto/sha/libcrypto-lib-sha1dgst.o",
                "crypto/sha/libcrypto-lib-sha256-mb-x86_64.o",
                "crypto/sha/libcrypto-lib-sha256-x86_64.o",
                "crypto/sha/libcrypto-lib-sha256.o",
                "crypto/sha/libcrypto-lib-sha3.o",
                "crypto/sha/libcrypto-lib-sha512-x86_64.o",
                "crypto/sha/libcrypto-lib-sha512.o",
                "crypto/sha/libcrypto-shlib-keccak1600-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha1-mb-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha1-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha1_one.o",
                "crypto/sha/libcrypto-shlib-sha1dgst.o",
                "crypto/sha/libcrypto-shlib-sha256-mb-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha256-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha256.o",
                "crypto/sha/libcrypto-shlib-sha3.o",
                "crypto/sha/libcrypto-shlib-sha512-x86_64.o",
                "crypto/sha/libcrypto-shlib-sha512.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/siphash" => {
            "deps" => [
                "crypto/siphash/libcrypto-lib-siphash.o",
                "crypto/siphash/libcrypto-shlib-siphash.o",
                "crypto/siphash/libssl-shlib-siphash.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "libssl"
                ]
            }
        },
        "crypto/slh_dsa" => {
            "deps" => [
                "crypto/slh_dsa/libcrypto-lib-slh_adrs.o",
                "crypto/slh_dsa/libcrypto-lib-slh_dsa.o",
                "crypto/slh_dsa/libcrypto-lib-slh_dsa_hash_ctx.o",
                "crypto/slh_dsa/libcrypto-lib-slh_dsa_key.o",
                "crypto/slh_dsa/libcrypto-lib-slh_fors.o",
                "crypto/slh_dsa/libcrypto-lib-slh_hash.o",
                "crypto/slh_dsa/libcrypto-lib-slh_hypertree.o",
                "crypto/slh_dsa/libcrypto-lib-slh_params.o",
                "crypto/slh_dsa/libcrypto-lib-slh_wots.o",
                "crypto/slh_dsa/libcrypto-lib-slh_xmss.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_adrs.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_dsa.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_dsa_hash_ctx.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_dsa_key.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_fors.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_hash.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_hypertree.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_params.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_wots.o",
                "crypto/slh_dsa/libcrypto-shlib-slh_xmss.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/sm2" => {
            "deps" => [
                "crypto/sm2/libcrypto-lib-sm2_crypt.o",
                "crypto/sm2/libcrypto-lib-sm2_err.o",
                "crypto/sm2/libcrypto-lib-sm2_key.o",
                "crypto/sm2/libcrypto-lib-sm2_kmeth.o",
                "crypto/sm2/libcrypto-lib-sm2_sign.o",
                "crypto/sm2/libcrypto-shlib-sm2_crypt.o",
                "crypto/sm2/libcrypto-shlib-sm2_err.o",
                "crypto/sm2/libcrypto-shlib-sm2_key.o",
                "crypto/sm2/libcrypto-shlib-sm2_kmeth.o",
                "crypto/sm2/libcrypto-shlib-sm2_sign.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/sm3" => {
            "deps" => [
                "crypto/sm3/libcrypto-lib-legacy_sm3.o",
                "crypto/sm3/libcrypto-lib-sm3.o",
                "crypto/sm3/libcrypto-shlib-legacy_sm3.o",
                "crypto/sm3/libcrypto-shlib-sm3.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/sm4" => {
            "deps" => [
                "crypto/sm4/libcrypto-lib-sm4.o",
                "crypto/sm4/libcrypto-shlib-sm4.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/srp" => {
            "deps" => [
                "crypto/srp/libcrypto-lib-srp_lib.o",
                "crypto/srp/libcrypto-lib-srp_vfy.o",
                "crypto/srp/libcrypto-shlib-srp_lib.o",
                "crypto/srp/libcrypto-shlib-srp_vfy.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/stack" => {
            "deps" => [
                "crypto/stack/libcrypto-lib-stack.o",
                "crypto/stack/libcrypto-shlib-stack.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/store" => {
            "deps" => [
                "crypto/store/libcrypto-lib-store_err.o",
                "crypto/store/libcrypto-lib-store_init.o",
                "crypto/store/libcrypto-lib-store_lib.o",
                "crypto/store/libcrypto-lib-store_meth.o",
                "crypto/store/libcrypto-lib-store_register.o",
                "crypto/store/libcrypto-lib-store_result.o",
                "crypto/store/libcrypto-lib-store_strings.o",
                "crypto/store/libcrypto-shlib-store_err.o",
                "crypto/store/libcrypto-shlib-store_init.o",
                "crypto/store/libcrypto-shlib-store_lib.o",
                "crypto/store/libcrypto-shlib-store_meth.o",
                "crypto/store/libcrypto-shlib-store_register.o",
                "crypto/store/libcrypto-shlib-store_result.o",
                "crypto/store/libcrypto-shlib-store_strings.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/thread" => {
            "deps" => [
                "crypto/thread/libcrypto-lib-api.o",
                "crypto/thread/libcrypto-lib-arch.o",
                "crypto/thread/libcrypto-lib-internal.o",
                "crypto/thread/libcrypto-shlib-api.o",
                "crypto/thread/libcrypto-shlib-arch.o",
                "crypto/thread/libcrypto-shlib-internal.o",
                "crypto/thread/libssl-shlib-arch.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "libssl"
                ]
            }
        },
        "crypto/thread/arch" => {
            "deps" => [
                "crypto/thread/arch/libcrypto-lib-thread_none.o",
                "crypto/thread/arch/libcrypto-lib-thread_posix.o",
                "crypto/thread/arch/libcrypto-lib-thread_win.o",
                "crypto/thread/arch/libcrypto-shlib-thread_none.o",
                "crypto/thread/arch/libcrypto-shlib-thread_posix.o",
                "crypto/thread/arch/libcrypto-shlib-thread_win.o",
                "crypto/thread/arch/libssl-shlib-thread_none.o",
                "crypto/thread/arch/libssl-shlib-thread_posix.o",
                "crypto/thread/arch/libssl-shlib-thread_win.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto",
                    "libssl"
                ]
            }
        },
        "crypto/ts" => {
            "deps" => [
                "crypto/ts/libcrypto-lib-ts_asn1.o",
                "crypto/ts/libcrypto-lib-ts_conf.o",
                "crypto/ts/libcrypto-lib-ts_err.o",
                "crypto/ts/libcrypto-lib-ts_lib.o",
                "crypto/ts/libcrypto-lib-ts_req_print.o",
                "crypto/ts/libcrypto-lib-ts_req_utils.o",
                "crypto/ts/libcrypto-lib-ts_rsp_print.o",
                "crypto/ts/libcrypto-lib-ts_rsp_sign.o",
                "crypto/ts/libcrypto-lib-ts_rsp_utils.o",
                "crypto/ts/libcrypto-lib-ts_rsp_verify.o",
                "crypto/ts/libcrypto-lib-ts_verify_ctx.o",
                "crypto/ts/libcrypto-shlib-ts_asn1.o",
                "crypto/ts/libcrypto-shlib-ts_conf.o",
                "crypto/ts/libcrypto-shlib-ts_err.o",
                "crypto/ts/libcrypto-shlib-ts_lib.o",
                "crypto/ts/libcrypto-shlib-ts_req_print.o",
                "crypto/ts/libcrypto-shlib-ts_req_utils.o",
                "crypto/ts/libcrypto-shlib-ts_rsp_print.o",
                "crypto/ts/libcrypto-shlib-ts_rsp_sign.o",
                "crypto/ts/libcrypto-shlib-ts_rsp_utils.o",
                "crypto/ts/libcrypto-shlib-ts_rsp_verify.o",
                "crypto/ts/libcrypto-shlib-ts_verify_ctx.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/tsapi" => {
            "deps" => [
                "crypto/tsapi/libcrypto-lib-tsapi_lib.o",
                "crypto/tsapi/libcrypto-shlib-tsapi_lib.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/txt_db" => {
            "deps" => [
                "crypto/txt_db/libcrypto-lib-txt_db.o",
                "crypto/txt_db/libcrypto-shlib-txt_db.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/ui" => {
            "deps" => [
                "crypto/ui/libcrypto-lib-ui_err.o",
                "crypto/ui/libcrypto-lib-ui_lib.o",
                "crypto/ui/libcrypto-lib-ui_null.o",
                "crypto/ui/libcrypto-lib-ui_openssl.o",
                "crypto/ui/libcrypto-lib-ui_util.o",
                "crypto/ui/libcrypto-shlib-ui_err.o",
                "crypto/ui/libcrypto-shlib-ui_lib.o",
                "crypto/ui/libcrypto-shlib-ui_null.o",
                "crypto/ui/libcrypto-shlib-ui_openssl.o",
                "crypto/ui/libcrypto-shlib-ui_util.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/x509" => {
            "deps" => [
                "crypto/x509/libcrypto-lib-by_dir.o",
                "crypto/x509/libcrypto-lib-by_file.o",
                "crypto/x509/libcrypto-lib-by_store.o",
                "crypto/x509/libcrypto-lib-pcy_cache.o",
                "crypto/x509/libcrypto-lib-pcy_data.o",
                "crypto/x509/libcrypto-lib-pcy_lib.o",
                "crypto/x509/libcrypto-lib-pcy_map.o",
                "crypto/x509/libcrypto-lib-pcy_node.o",
                "crypto/x509/libcrypto-lib-pcy_tree.o",
                "crypto/x509/libcrypto-lib-t_acert.o",
                "crypto/x509/libcrypto-lib-t_crl.o",
                "crypto/x509/libcrypto-lib-t_req.o",
                "crypto/x509/libcrypto-lib-t_x509.o",
                "crypto/x509/libcrypto-lib-v3_aaa.o",
                "crypto/x509/libcrypto-lib-v3_ac_tgt.o",
                "crypto/x509/libcrypto-lib-v3_addr.o",
                "crypto/x509/libcrypto-lib-v3_admis.o",
                "crypto/x509/libcrypto-lib-v3_akeya.o",
                "crypto/x509/libcrypto-lib-v3_akid.o",
                "crypto/x509/libcrypto-lib-v3_asid.o",
                "crypto/x509/libcrypto-lib-v3_attrdesc.o",
                "crypto/x509/libcrypto-lib-v3_attrmap.o",
                "crypto/x509/libcrypto-lib-v3_audit_id.o",
                "crypto/x509/libcrypto-lib-v3_authattid.o",
                "crypto/x509/libcrypto-lib-v3_battcons.o",
                "crypto/x509/libcrypto-lib-v3_bcons.o",
                "crypto/x509/libcrypto-lib-v3_bitst.o",
                "crypto/x509/libcrypto-lib-v3_conf.o",
                "crypto/x509/libcrypto-lib-v3_cpols.o",
                "crypto/x509/libcrypto-lib-v3_crld.o",
                "crypto/x509/libcrypto-lib-v3_enum.o",
                "crypto/x509/libcrypto-lib-v3_extku.o",
                "crypto/x509/libcrypto-lib-v3_genn.o",
                "crypto/x509/libcrypto-lib-v3_group_ac.o",
                "crypto/x509/libcrypto-lib-v3_ia5.o",
                "crypto/x509/libcrypto-lib-v3_ind_iss.o",
                "crypto/x509/libcrypto-lib-v3_info.o",
                "crypto/x509/libcrypto-lib-v3_int.o",
                "crypto/x509/libcrypto-lib-v3_iobo.o",
                "crypto/x509/libcrypto-lib-v3_lib.o",
                "crypto/x509/libcrypto-lib-v3_ncons.o",
                "crypto/x509/libcrypto-lib-v3_no_ass.o",
                "crypto/x509/libcrypto-lib-v3_no_rev_avail.o",
                "crypto/x509/libcrypto-lib-v3_pci.o",
                "crypto/x509/libcrypto-lib-v3_pcia.o",
                "crypto/x509/libcrypto-lib-v3_pcons.o",
                "crypto/x509/libcrypto-lib-v3_pku.o",
                "crypto/x509/libcrypto-lib-v3_pmaps.o",
                "crypto/x509/libcrypto-lib-v3_prn.o",
                "crypto/x509/libcrypto-lib-v3_purp.o",
                "crypto/x509/libcrypto-lib-v3_rolespec.o",
                "crypto/x509/libcrypto-lib-v3_san.o",
                "crypto/x509/libcrypto-lib-v3_sda.o",
                "crypto/x509/libcrypto-lib-v3_single_use.o",
                "crypto/x509/libcrypto-lib-v3_skid.o",
                "crypto/x509/libcrypto-lib-v3_soa_id.o",
                "crypto/x509/libcrypto-lib-v3_sxnet.o",
                "crypto/x509/libcrypto-lib-v3_timespec.o",
                "crypto/x509/libcrypto-lib-v3_tlsf.o",
                "crypto/x509/libcrypto-lib-v3_usernotice.o",
                "crypto/x509/libcrypto-lib-v3_utl.o",
                "crypto/x509/libcrypto-lib-v3err.o",
                "crypto/x509/libcrypto-lib-x509_acert.o",
                "crypto/x509/libcrypto-lib-x509_att.o",
                "crypto/x509/libcrypto-lib-x509_cmp.o",
                "crypto/x509/libcrypto-lib-x509_d2.o",
                "crypto/x509/libcrypto-lib-x509_def.o",
                "crypto/x509/libcrypto-lib-x509_err.o",
                "crypto/x509/libcrypto-lib-x509_ext.o",
                "crypto/x509/libcrypto-lib-x509_lu.o",
                "crypto/x509/libcrypto-lib-x509_meth.o",
                "crypto/x509/libcrypto-lib-x509_obj.o",
                "crypto/x509/libcrypto-lib-x509_r2x.o",
                "crypto/x509/libcrypto-lib-x509_req.o",
                "crypto/x509/libcrypto-lib-x509_set.o",
                "crypto/x509/libcrypto-lib-x509_trust.o",
                "crypto/x509/libcrypto-lib-x509_txt.o",
                "crypto/x509/libcrypto-lib-x509_v3.o",
                "crypto/x509/libcrypto-lib-x509_vfy.o",
                "crypto/x509/libcrypto-lib-x509_vpm.o",
                "crypto/x509/libcrypto-lib-x509aset.o",
                "crypto/x509/libcrypto-lib-x509cset.o",
                "crypto/x509/libcrypto-lib-x509name.o",
                "crypto/x509/libcrypto-lib-x509rset.o",
                "crypto/x509/libcrypto-lib-x509spki.o",
                "crypto/x509/libcrypto-lib-x509type.o",
                "crypto/x509/libcrypto-lib-x_all.o",
                "crypto/x509/libcrypto-lib-x_attrib.o",
                "crypto/x509/libcrypto-lib-x_crl.o",
                "crypto/x509/libcrypto-lib-x_exten.o",
                "crypto/x509/libcrypto-lib-x_ietfatt.o",
                "crypto/x509/libcrypto-lib-x_name.o",
                "crypto/x509/libcrypto-lib-x_pubkey.o",
                "crypto/x509/libcrypto-lib-x_req.o",
                "crypto/x509/libcrypto-lib-x_x509.o",
                "crypto/x509/libcrypto-lib-x_x509a.o",
                "crypto/x509/libcrypto-shlib-by_dir.o",
                "crypto/x509/libcrypto-shlib-by_file.o",
                "crypto/x509/libcrypto-shlib-by_store.o",
                "crypto/x509/libcrypto-shlib-pcy_cache.o",
                "crypto/x509/libcrypto-shlib-pcy_data.o",
                "crypto/x509/libcrypto-shlib-pcy_lib.o",
                "crypto/x509/libcrypto-shlib-pcy_map.o",
                "crypto/x509/libcrypto-shlib-pcy_node.o",
                "crypto/x509/libcrypto-shlib-pcy_tree.o",
                "crypto/x509/libcrypto-shlib-t_acert.o",
                "crypto/x509/libcrypto-shlib-t_crl.o",
                "crypto/x509/libcrypto-shlib-t_req.o",
                "crypto/x509/libcrypto-shlib-t_x509.o",
                "crypto/x509/libcrypto-shlib-v3_aaa.o",
                "crypto/x509/libcrypto-shlib-v3_ac_tgt.o",
                "crypto/x509/libcrypto-shlib-v3_addr.o",
                "crypto/x509/libcrypto-shlib-v3_admis.o",
                "crypto/x509/libcrypto-shlib-v3_akeya.o",
                "crypto/x509/libcrypto-shlib-v3_akid.o",
                "crypto/x509/libcrypto-shlib-v3_asid.o",
                "crypto/x509/libcrypto-shlib-v3_attrdesc.o",
                "crypto/x509/libcrypto-shlib-v3_attrmap.o",
                "crypto/x509/libcrypto-shlib-v3_audit_id.o",
                "crypto/x509/libcrypto-shlib-v3_authattid.o",
                "crypto/x509/libcrypto-shlib-v3_battcons.o",
                "crypto/x509/libcrypto-shlib-v3_bcons.o",
                "crypto/x509/libcrypto-shlib-v3_bitst.o",
                "crypto/x509/libcrypto-shlib-v3_conf.o",
                "crypto/x509/libcrypto-shlib-v3_cpols.o",
                "crypto/x509/libcrypto-shlib-v3_crld.o",
                "crypto/x509/libcrypto-shlib-v3_enum.o",
                "crypto/x509/libcrypto-shlib-v3_extku.o",
                "crypto/x509/libcrypto-shlib-v3_genn.o",
                "crypto/x509/libcrypto-shlib-v3_group_ac.o",
                "crypto/x509/libcrypto-shlib-v3_ia5.o",
                "crypto/x509/libcrypto-shlib-v3_ind_iss.o",
                "crypto/x509/libcrypto-shlib-v3_info.o",
                "crypto/x509/libcrypto-shlib-v3_int.o",
                "crypto/x509/libcrypto-shlib-v3_iobo.o",
                "crypto/x509/libcrypto-shlib-v3_lib.o",
                "crypto/x509/libcrypto-shlib-v3_ncons.o",
                "crypto/x509/libcrypto-shlib-v3_no_ass.o",
                "crypto/x509/libcrypto-shlib-v3_no_rev_avail.o",
                "crypto/x509/libcrypto-shlib-v3_pci.o",
                "crypto/x509/libcrypto-shlib-v3_pcia.o",
                "crypto/x509/libcrypto-shlib-v3_pcons.o",
                "crypto/x509/libcrypto-shlib-v3_pku.o",
                "crypto/x509/libcrypto-shlib-v3_pmaps.o",
                "crypto/x509/libcrypto-shlib-v3_prn.o",
                "crypto/x509/libcrypto-shlib-v3_purp.o",
                "crypto/x509/libcrypto-shlib-v3_rolespec.o",
                "crypto/x509/libcrypto-shlib-v3_san.o",
                "crypto/x509/libcrypto-shlib-v3_sda.o",
                "crypto/x509/libcrypto-shlib-v3_single_use.o",
                "crypto/x509/libcrypto-shlib-v3_skid.o",
                "crypto/x509/libcrypto-shlib-v3_soa_id.o",
                "crypto/x509/libcrypto-shlib-v3_sxnet.o",
                "crypto/x509/libcrypto-shlib-v3_timespec.o",
                "crypto/x509/libcrypto-shlib-v3_tlsf.o",
                "crypto/x509/libcrypto-shlib-v3_usernotice.o",
                "crypto/x509/libcrypto-shlib-v3_utl.o",
                "crypto/x509/libcrypto-shlib-v3err.o",
                "crypto/x509/libcrypto-shlib-x509_acert.o",
                "crypto/x509/libcrypto-shlib-x509_att.o",
                "crypto/x509/libcrypto-shlib-x509_cmp.o",
                "crypto/x509/libcrypto-shlib-x509_d2.o",
                "crypto/x509/libcrypto-shlib-x509_def.o",
                "crypto/x509/libcrypto-shlib-x509_err.o",
                "crypto/x509/libcrypto-shlib-x509_ext.o",
                "crypto/x509/libcrypto-shlib-x509_lu.o",
                "crypto/x509/libcrypto-shlib-x509_meth.o",
                "crypto/x509/libcrypto-shlib-x509_obj.o",
                "crypto/x509/libcrypto-shlib-x509_r2x.o",
                "crypto/x509/libcrypto-shlib-x509_req.o",
                "crypto/x509/libcrypto-shlib-x509_set.o",
                "crypto/x509/libcrypto-shlib-x509_trust.o",
                "crypto/x509/libcrypto-shlib-x509_txt.o",
                "crypto/x509/libcrypto-shlib-x509_v3.o",
                "crypto/x509/libcrypto-shlib-x509_vfy.o",
                "crypto/x509/libcrypto-shlib-x509_vpm.o",
                "crypto/x509/libcrypto-shlib-x509aset.o",
                "crypto/x509/libcrypto-shlib-x509cset.o",
                "crypto/x509/libcrypto-shlib-x509name.o",
                "crypto/x509/libcrypto-shlib-x509rset.o",
                "crypto/x509/libcrypto-shlib-x509spki.o",
                "crypto/x509/libcrypto-shlib-x509type.o",
                "crypto/x509/libcrypto-shlib-x_all.o",
                "crypto/x509/libcrypto-shlib-x_attrib.o",
                "crypto/x509/libcrypto-shlib-x_crl.o",
                "crypto/x509/libcrypto-shlib-x_exten.o",
                "crypto/x509/libcrypto-shlib-x_ietfatt.o",
                "crypto/x509/libcrypto-shlib-x_name.o",
                "crypto/x509/libcrypto-shlib-x_pubkey.o",
                "crypto/x509/libcrypto-shlib-x_req.o",
                "crypto/x509/libcrypto-shlib-x_x509.o",
                "crypto/x509/libcrypto-shlib-x_x509a.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "crypto/zuc" => {
            "deps" => [
                "crypto/zuc/libcrypto-lib-zuc.o",
                "crypto/zuc/libcrypto-shlib-zuc.o"
            ],
            "products" => {
                "lib" => [
                    "libcrypto"
                ]
            }
        },
        "engines" => {
            "products" => {
                "dso" => [
                    "engines/afalg",
                    "engines/capi",
                    "engines/dasync",
                    "engines/ecptest",
                    "engines/loader_attic",
                    "engines/ossltest",
                    "engines/padlock"
                ]
            }
        },
        "providers" => {
            "deps" => [
                "providers/libcrypto-lib-baseprov.o",
                "providers/libcrypto-lib-defltprov.o",
                "providers/libcrypto-lib-nullprov.o",
                "providers/libcrypto-lib-prov_running.o",
                "providers/libdefault.a",
                "providers/libcrypto-shlib-baseprov.o",
                "providers/libcrypto-shlib-defltprov.o",
                "providers/libcrypto-shlib-nullprov.o",
                "providers/libcrypto-shlib-prov_running.o",
                "providers/libdefault.a"
            ],
            "products" => {
                "dso" => [
                    "providers/legacy"
                ],
                "lib" => [
                    "libcrypto",
                    "providers/liblegacy.a"
                ]
            }
        },
        "providers/common" => {
            "deps" => [
                "providers/common/libcommon-lib-provider_ctx.o",
                "providers/common/libcommon-lib-provider_err.o",
                "providers/common/libdefault-lib-bio_prov.o",
                "providers/common/libdefault-lib-capabilities.o",
                "providers/common/libdefault-lib-digest_to_nid.o",
                "providers/common/libdefault-lib-provider_seeding.o",
                "providers/common/libdefault-lib-provider_util.o",
                "providers/common/libdefault-lib-securitycheck.o",
                "providers/common/libdefault-lib-securitycheck_default.o",
                "providers/common/liblegacy-lib-provider_util.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libcommon.a",
                    "providers/libdefault.a",
                    "providers/liblegacy.a"
                ]
            }
        },
        "providers/common/der" => {
            "deps" => [
                "providers/common/der/libcommon-lib-der_digests_gen.o",
                "providers/common/der/libcommon-lib-der_dsa_gen.o",
                "providers/common/der/libcommon-lib-der_dsa_key.o",
                "providers/common/der/libcommon-lib-der_dsa_sig.o",
                "providers/common/der/libcommon-lib-der_ec_gen.o",
                "providers/common/der/libcommon-lib-der_ec_key.o",
                "providers/common/der/libcommon-lib-der_ec_sig.o",
                "providers/common/der/libcommon-lib-der_ecx_gen.o",
                "providers/common/der/libcommon-lib-der_ecx_key.o",
                "providers/common/der/libcommon-lib-der_ml_dsa_gen.o",
                "providers/common/der/libcommon-lib-der_ml_dsa_key.o",
                "providers/common/der/libcommon-lib-der_rsa_gen.o",
                "providers/common/der/libcommon-lib-der_rsa_key.o",
                "providers/common/der/libcommon-lib-der_slh_dsa_gen.o",
                "providers/common/der/libcommon-lib-der_slh_dsa_key.o",
                "providers/common/der/libcommon-lib-der_wrap_gen.o",
                "providers/common/der/libdefault-lib-der_rsa_sig.o",
                "providers/common/der/libdefault-lib-der_sm2_gen.o",
                "providers/common/der/libdefault-lib-der_sm2_key.o",
                "providers/common/der/libdefault-lib-der_sm2_sig.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libcommon.a",
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/asymciphers" => {
            "deps" => [
                "providers/implementations/asymciphers/libdefault-lib-rsa_enc.o",
                "providers/implementations/asymciphers/libdefault-lib-sm2_enc.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/ciphers" => {
            "deps" => [
                "providers/implementations/ciphers/libcommon-lib-ciphercommon.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_block.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm_hw.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm_hw.o",
                "providers/implementations/ciphers/libcommon-lib-ciphercommon_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha1_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha256_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_polyval.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_wrp.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_fips.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_chacha20.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_cts.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_null.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_common.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap_hw.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3.o",
                "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3_hw.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_des.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_des_hw.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_desx.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_desx_hw.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_rc4.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5_hw.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hw.o",
                "providers/implementations/ciphers/liblegacy-lib-cipher_tdes_common.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libcommon.a",
                    "providers/libdefault.a",
                    "providers/liblegacy.a"
                ]
            }
        },
        "providers/implementations/digests" => {
            "deps" => [
                "providers/implementations/digests/libcommon-lib-digestcommon.o",
                "providers/implementations/digests/libdefault-lib-md5_prov.o",
                "providers/implementations/digests/libdefault-lib-md5_sha1_prov.o",
                "providers/implementations/digests/libdefault-lib-null_prov.o",
                "providers/implementations/digests/libdefault-lib-sha2_prov.o",
                "providers/implementations/digests/libdefault-lib-sha3_prov.o",
                "providers/implementations/digests/libdefault-lib-sm3_prov.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libcommon.a",
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/encode_decode" => {
            "deps" => [
                "providers/implementations/encode_decode/libdefault-lib-decode_der2key.o",
                "providers/implementations/encode_decode/libdefault-lib-decode_epki2pki.o",
                "providers/implementations/encode_decode/libdefault-lib-decode_msblob2key.o",
                "providers/implementations/encode_decode/libdefault-lib-decode_pem2der.o",
                "providers/implementations/encode_decode/libdefault-lib-decode_pvk2key.o",
                "providers/implementations/encode_decode/libdefault-lib-decode_spki2typespki.o",
                "providers/implementations/encode_decode/libdefault-lib-encode_key2any.o",
                "providers/implementations/encode_decode/libdefault-lib-encode_key2blob.o",
                "providers/implementations/encode_decode/libdefault-lib-encode_key2ms.o",
                "providers/implementations/encode_decode/libdefault-lib-encode_key2text.o",
                "providers/implementations/encode_decode/libdefault-lib-endecoder_common.o",
                "providers/implementations/encode_decode/libdefault-lib-ml_common_codecs.o",
                "providers/implementations/encode_decode/libdefault-lib-ml_dsa_codecs.o",
                "providers/implementations/encode_decode/libdefault-lib-ml_kem_codecs.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/exchange" => {
            "deps" => [
                "providers/implementations/exchange/libdefault-lib-dh_exch.o",
                "providers/implementations/exchange/libdefault-lib-ecdh_exch.o",
                "providers/implementations/exchange/libdefault-lib-ecx_exch.o",
                "providers/implementations/exchange/libdefault-lib-kdf_exch.o",
                "providers/implementations/exchange/libdefault-lib-sm2dh_exch.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/kdfs" => {
            "deps" => [
                "providers/implementations/kdfs/libdefault-lib-hkdf.o",
                "providers/implementations/kdfs/libdefault-lib-hmacdrbg_kdf.o",
                "providers/implementations/kdfs/libdefault-lib-kbkdf.o",
                "providers/implementations/kdfs/libdefault-lib-krb5kdf.o",
                "providers/implementations/kdfs/libdefault-lib-pbkdf2.o",
                "providers/implementations/kdfs/libdefault-lib-pbkdf2_fips.o",
                "providers/implementations/kdfs/libdefault-lib-pkcs12kdf.o",
                "providers/implementations/kdfs/libdefault-lib-scrypt.o",
                "providers/implementations/kdfs/libdefault-lib-sshkdf.o",
                "providers/implementations/kdfs/libdefault-lib-sskdf.o",
                "providers/implementations/kdfs/libdefault-lib-tls1_prf.o",
                "providers/implementations/kdfs/libdefault-lib-wbsm4kdf.o",
                "providers/implementations/kdfs/libdefault-lib-x942kdf.o",
                "providers/implementations/kdfs/liblegacy-lib-pbkdf1.o",
                "providers/implementations/kdfs/liblegacy-lib-pvkkdf.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a",
                    "providers/liblegacy.a"
                ]
            }
        },
        "providers/implementations/kem" => {
            "deps" => [
                "providers/implementations/kem/libdefault-lib-ec_kem.o",
                "providers/implementations/kem/libdefault-lib-ecx_kem.o",
                "providers/implementations/kem/libdefault-lib-kem_util.o",
                "providers/implementations/kem/libdefault-lib-ml_kem_kem.o",
                "providers/implementations/kem/libdefault-lib-mlx_kem.o",
                "providers/implementations/kem/libdefault-lib-rsa_kem.o",
                "providers/implementations/kem/libtemplate-lib-template_kem.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a",
                    "providers/libtemplate.a"
                ]
            }
        },
        "providers/implementations/keymgmt" => {
            "deps" => [
                "providers/implementations/keymgmt/libdefault-lib-dh_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-dsa_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-ec_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-ecx_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-kdf_legacy_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-mac_legacy_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-ml_dsa_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-ml_kem_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-mlx_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-rsa_kmgmt.o",
                "providers/implementations/keymgmt/libdefault-lib-slh_dsa_kmgmt.o",
                "providers/implementations/keymgmt/libtemplate-lib-template_kmgmt.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a",
                    "providers/libtemplate.a"
                ]
            }
        },
        "providers/implementations/macs" => {
            "deps" => [
                "providers/implementations/macs/libdefault-lib-cmac_prov.o",
                "providers/implementations/macs/libdefault-lib-eia3_prov.o",
                "providers/implementations/macs/libdefault-lib-gmac_prov.o",
                "providers/implementations/macs/libdefault-lib-hmac_prov.o",
                "providers/implementations/macs/libdefault-lib-kmac_prov.o",
                "providers/implementations/macs/libdefault-lib-poly1305_prov.o",
                "providers/implementations/macs/libdefault-lib-siphash_prov.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/rands" => {
            "deps" => [
                "providers/implementations/rands/libdefault-lib-drbg.o",
                "providers/implementations/rands/libdefault-lib-drbg_ctr.o",
                "providers/implementations/rands/libdefault-lib-drbg_hash.o",
                "providers/implementations/rands/libdefault-lib-drbg_hmac.o",
                "providers/implementations/rands/libdefault-lib-seed_src.o",
                "providers/implementations/rands/libdefault-lib-seed_src_jitter.o",
                "providers/implementations/rands/libdefault-lib-test_rng.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/rands/seeding" => {
            "deps" => [
                "providers/implementations/rands/seeding/libdefault-lib-rand_cpu_x86.o",
                "providers/implementations/rands/seeding/libdefault-lib-rand_tsc.o",
                "providers/implementations/rands/seeding/libdefault-lib-rand_unix.o",
                "providers/implementations/rands/seeding/libdefault-lib-rand_win.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/signature" => {
            "deps" => [
                "providers/implementations/signature/libdefault-lib-dsa_sig.o",
                "providers/implementations/signature/libdefault-lib-ecdsa_sig.o",
                "providers/implementations/signature/libdefault-lib-eddsa_sig.o",
                "providers/implementations/signature/libdefault-lib-mac_legacy_sig.o",
                "providers/implementations/signature/libdefault-lib-ml_dsa_sig.o",
                "providers/implementations/signature/libdefault-lib-rsa_sig.o",
                "providers/implementations/signature/libdefault-lib-slh_dsa_sig.o",
                "providers/implementations/signature/libdefault-lib-sm2_sig.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/skeymgmt" => {
            "deps" => [
                "providers/implementations/skeymgmt/libdefault-lib-aes_skmgmt.o",
                "providers/implementations/skeymgmt/libdefault-lib-generic.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "providers/implementations/storemgmt" => {
            "deps" => [
                "providers/implementations/storemgmt/libdefault-lib-file_store.o",
                "providers/implementations/storemgmt/libdefault-lib-file_store_any2obj.o"
            ],
            "products" => {
                "lib" => [
                    "providers/libdefault.a"
                ]
            }
        },
        "ssl" => {
            "deps" => [
                "ssl/libssl-lib-bio_ssl.o",
                "ssl/libssl-lib-d1_lib.o",
                "ssl/libssl-lib-d1_msg.o",
                "ssl/libssl-lib-d1_srtp.o",
                "ssl/libssl-lib-methods.o",
                "ssl/libssl-lib-pqueue.o",
                "ssl/libssl-lib-priority_queue.o",
                "ssl/libssl-lib-s3_enc.o",
                "ssl/libssl-lib-s3_lib.o",
                "ssl/libssl-lib-s3_msg.o",
                "ssl/libssl-lib-ssl_asn1.o",
                "ssl/libssl-lib-ssl_cert.o",
                "ssl/libssl-lib-ssl_cert_comp.o",
                "ssl/libssl-lib-ssl_ciph.o",
                "ssl/libssl-lib-ssl_conf.o",
                "ssl/libssl-lib-ssl_err_legacy.o",
                "ssl/libssl-lib-ssl_init.o",
                "ssl/libssl-lib-ssl_lib.o",
                "ssl/libssl-lib-ssl_mcnf.o",
                "ssl/libssl-lib-ssl_rsa.o",
                "ssl/libssl-lib-ssl_rsa_legacy.o",
                "ssl/libssl-lib-ssl_sess.o",
                "ssl/libssl-lib-ssl_stat.o",
                "ssl/libssl-lib-ssl_txt.o",
                "ssl/libssl-lib-ssl_utst.o",
                "ssl/libssl-lib-t1_enc.o",
                "ssl/libssl-lib-t1_lib.o",
                "ssl/libssl-lib-t1_trce.o",
                "ssl/libssl-lib-tls13_enc.o",
                "ssl/libssl-lib-tls_depr.o",
                "ssl/libssl-lib-tls_srp.o",
                "ssl/libssl-shlib-bio_ssl.o",
                "ssl/libssl-shlib-d1_lib.o",
                "ssl/libssl-shlib-d1_msg.o",
                "ssl/libssl-shlib-d1_srtp.o",
                "ssl/libssl-shlib-methods.o",
                "ssl/libssl-shlib-pqueue.o",
                "ssl/libssl-shlib-priority_queue.o",
                "ssl/libssl-shlib-s3_enc.o",
                "ssl/libssl-shlib-s3_lib.o",
                "ssl/libssl-shlib-s3_msg.o",
                "ssl/libssl-shlib-ssl_asn1.o",
                "ssl/libssl-shlib-ssl_cert.o",
                "ssl/libssl-shlib-ssl_cert_comp.o",
                "ssl/libssl-shlib-ssl_ciph.o",
                "ssl/libssl-shlib-ssl_conf.o",
                "ssl/libssl-shlib-ssl_err_legacy.o",
                "ssl/libssl-shlib-ssl_init.o",
                "ssl/libssl-shlib-ssl_lib.o",
                "ssl/libssl-shlib-ssl_mcnf.o",
                "ssl/libssl-shlib-ssl_rsa.o",
                "ssl/libssl-shlib-ssl_rsa_legacy.o",
                "ssl/libssl-shlib-ssl_sess.o",
                "ssl/libssl-shlib-ssl_stat.o",
                "ssl/libssl-shlib-ssl_txt.o",
                "ssl/libssl-shlib-ssl_utst.o",
                "ssl/libssl-shlib-t1_enc.o",
                "ssl/libssl-shlib-t1_lib.o",
                "ssl/libssl-shlib-t1_trce.o",
                "ssl/libssl-shlib-tls13_enc.o",
                "ssl/libssl-shlib-tls_depr.o",
                "ssl/libssl-shlib-tls_srp.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "ssl/quic" => {
            "deps" => [
                "ssl/quic/libssl-lib-cc_newreno.o",
                "ssl/quic/libssl-lib-json_enc.o",
                "ssl/quic/libssl-lib-qlog.o",
                "ssl/quic/libssl-lib-qlog_event_helpers.o",
                "ssl/quic/libssl-lib-quic_ackm.o",
                "ssl/quic/libssl-lib-quic_cfq.o",
                "ssl/quic/libssl-lib-quic_channel.o",
                "ssl/quic/libssl-lib-quic_demux.o",
                "ssl/quic/libssl-lib-quic_engine.o",
                "ssl/quic/libssl-lib-quic_fc.o",
                "ssl/quic/libssl-lib-quic_fifd.o",
                "ssl/quic/libssl-lib-quic_impl.o",
                "ssl/quic/libssl-lib-quic_lcidm.o",
                "ssl/quic/libssl-lib-quic_method.o",
                "ssl/quic/libssl-lib-quic_obj.o",
                "ssl/quic/libssl-lib-quic_port.o",
                "ssl/quic/libssl-lib-quic_rcidm.o",
                "ssl/quic/libssl-lib-quic_reactor.o",
                "ssl/quic/libssl-lib-quic_reactor_wait_ctx.o",
                "ssl/quic/libssl-lib-quic_record_rx.o",
                "ssl/quic/libssl-lib-quic_record_shared.o",
                "ssl/quic/libssl-lib-quic_record_tx.o",
                "ssl/quic/libssl-lib-quic_record_util.o",
                "ssl/quic/libssl-lib-quic_rstream.o",
                "ssl/quic/libssl-lib-quic_rx_depack.o",
                "ssl/quic/libssl-lib-quic_sf_list.o",
                "ssl/quic/libssl-lib-quic_srt_gen.o",
                "ssl/quic/libssl-lib-quic_srtm.o",
                "ssl/quic/libssl-lib-quic_sstream.o",
                "ssl/quic/libssl-lib-quic_statm.o",
                "ssl/quic/libssl-lib-quic_stream_map.o",
                "ssl/quic/libssl-lib-quic_thread_assist.o",
                "ssl/quic/libssl-lib-quic_tls.o",
                "ssl/quic/libssl-lib-quic_tls_api.o",
                "ssl/quic/libssl-lib-quic_tls_api_method.o",
                "ssl/quic/libssl-lib-quic_trace.o",
                "ssl/quic/libssl-lib-quic_tserver.o",
                "ssl/quic/libssl-lib-quic_txp.o",
                "ssl/quic/libssl-lib-quic_txpim.o",
                "ssl/quic/libssl-lib-quic_types.o",
                "ssl/quic/libssl-lib-quic_wire.o",
                "ssl/quic/libssl-lib-quic_wire_pkt.o",
                "ssl/quic/libssl-lib-uint_set.o",
                "ssl/quic/libssl-shlib-cc_newreno.o",
                "ssl/quic/libssl-shlib-json_enc.o",
                "ssl/quic/libssl-shlib-qlog.o",
                "ssl/quic/libssl-shlib-qlog_event_helpers.o",
                "ssl/quic/libssl-shlib-quic_ackm.o",
                "ssl/quic/libssl-shlib-quic_cfq.o",
                "ssl/quic/libssl-shlib-quic_channel.o",
                "ssl/quic/libssl-shlib-quic_demux.o",
                "ssl/quic/libssl-shlib-quic_engine.o",
                "ssl/quic/libssl-shlib-quic_fc.o",
                "ssl/quic/libssl-shlib-quic_fifd.o",
                "ssl/quic/libssl-shlib-quic_impl.o",
                "ssl/quic/libssl-shlib-quic_lcidm.o",
                "ssl/quic/libssl-shlib-quic_method.o",
                "ssl/quic/libssl-shlib-quic_obj.o",
                "ssl/quic/libssl-shlib-quic_port.o",
                "ssl/quic/libssl-shlib-quic_rcidm.o",
                "ssl/quic/libssl-shlib-quic_reactor.o",
                "ssl/quic/libssl-shlib-quic_reactor_wait_ctx.o",
                "ssl/quic/libssl-shlib-quic_record_rx.o",
                "ssl/quic/libssl-shlib-quic_record_shared.o",
                "ssl/quic/libssl-shlib-quic_record_tx.o",
                "ssl/quic/libssl-shlib-quic_record_util.o",
                "ssl/quic/libssl-shlib-quic_rstream.o",
                "ssl/quic/libssl-shlib-quic_rx_depack.o",
                "ssl/quic/libssl-shlib-quic_sf_list.o",
                "ssl/quic/libssl-shlib-quic_srt_gen.o",
                "ssl/quic/libssl-shlib-quic_srtm.o",
                "ssl/quic/libssl-shlib-quic_sstream.o",
                "ssl/quic/libssl-shlib-quic_statm.o",
                "ssl/quic/libssl-shlib-quic_stream_map.o",
                "ssl/quic/libssl-shlib-quic_thread_assist.o",
                "ssl/quic/libssl-shlib-quic_tls.o",
                "ssl/quic/libssl-shlib-quic_tls_api.o",
                "ssl/quic/libssl-shlib-quic_tls_api_method.o",
                "ssl/quic/libssl-shlib-quic_trace.o",
                "ssl/quic/libssl-shlib-quic_tserver.o",
                "ssl/quic/libssl-shlib-quic_txp.o",
                "ssl/quic/libssl-shlib-quic_txpim.o",
                "ssl/quic/libssl-shlib-quic_types.o",
                "ssl/quic/libssl-shlib-quic_wire.o",
                "ssl/quic/libssl-shlib-quic_wire_pkt.o",
                "ssl/quic/libssl-shlib-uint_set.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "ssl/record" => {
            "deps" => [
                "ssl/record/libssl-lib-rec_layer_d1.o",
                "ssl/record/libssl-lib-rec_layer_s3.o",
                "ssl/record/libssl-shlib-rec_layer_d1.o",
                "ssl/record/libssl-shlib-rec_layer_s3.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "ssl/record/methods" => {
            "deps" => [
                "ssl/record/methods/libssl-lib-dtls_meth.o",
                "ssl/record/methods/libssl-lib-ssl3_meth.o",
                "ssl/record/methods/libssl-lib-tls13_meth.o",
                "ssl/record/methods/libssl-lib-tls1_meth.o",
                "ssl/record/methods/libssl-lib-tls_common.o",
                "ssl/record/methods/libssl-lib-tls_multib.o",
                "ssl/record/methods/libssl-lib-tlsany_meth.o",
                "ssl/record/methods/libssl-shlib-dtls_meth.o",
                "ssl/record/methods/libssl-shlib-ssl3_cbc.o",
                "ssl/record/methods/libssl-shlib-ssl3_meth.o",
                "ssl/record/methods/libssl-shlib-tls13_meth.o",
                "ssl/record/methods/libssl-shlib-tls1_meth.o",
                "ssl/record/methods/libssl-shlib-tls_common.o",
                "ssl/record/methods/libssl-shlib-tls_multib.o",
                "ssl/record/methods/libssl-shlib-tls_pad.o",
                "ssl/record/methods/libssl-shlib-tlsany_meth.o",
                "ssl/record/methods/libcommon-lib-tls_pad.o",
                "ssl/record/methods/libdefault-lib-ssl3_cbc.o"
            ],
            "products" => {
                "lib" => [
                    "libssl",
                    "providers/libcommon.a",
                    "providers/libdefault.a"
                ]
            }
        },
        "ssl/rio" => {
            "deps" => [
                "ssl/rio/libssl-lib-poll_builder.o",
                "ssl/rio/libssl-lib-poll_immediate.o",
                "ssl/rio/libssl-lib-rio_notifier.o",
                "ssl/rio/libssl-shlib-poll_builder.o",
                "ssl/rio/libssl-shlib-poll_immediate.o",
                "ssl/rio/libssl-shlib-rio_notifier.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "ssl/statem" => {
            "deps" => [
                "ssl/statem/libssl-lib-extensions.o",
                "ssl/statem/libssl-lib-extensions_clnt.o",
                "ssl/statem/libssl-lib-extensions_cust.o",
                "ssl/statem/libssl-lib-extensions_srvr.o",
                "ssl/statem/libssl-lib-statem.o",
                "ssl/statem/libssl-lib-statem_clnt.o",
                "ssl/statem/libssl-lib-statem_dtls.o",
                "ssl/statem/libssl-lib-statem_lib.o",
                "ssl/statem/libssl-lib-statem_srvr.o",
                "ssl/statem/libssl-shlib-extensions.o",
                "ssl/statem/libssl-shlib-extensions_clnt.o",
                "ssl/statem/libssl-shlib-extensions_cust.o",
                "ssl/statem/libssl-shlib-extensions_srvr.o",
                "ssl/statem/libssl-shlib-statem.o",
                "ssl/statem/libssl-shlib-statem_clnt.o",
                "ssl/statem/libssl-shlib-statem_dtls.o",
                "ssl/statem/libssl-shlib-statem_lib.o",
                "ssl/statem/libssl-shlib-statem_srvr.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "ssl/statem_ntls" => {
            "deps" => [
                "ssl/statem_ntls/libssl-lib-ntls_extensions.o",
                "ssl/statem_ntls/libssl-lib-ntls_extensions_clnt.o",
                "ssl/statem_ntls/libssl-lib-ntls_extensions_cust.o",
                "ssl/statem_ntls/libssl-lib-ntls_extensions_srvr.o",
                "ssl/statem_ntls/libssl-lib-ntls_statem.o",
                "ssl/statem_ntls/libssl-lib-ntls_statem_clnt.o",
                "ssl/statem_ntls/libssl-lib-ntls_statem_lib.o",
                "ssl/statem_ntls/libssl-lib-ntls_statem_srvr.o",
                "ssl/statem_ntls/libssl-shlib-ntls_extensions.o",
                "ssl/statem_ntls/libssl-shlib-ntls_extensions_clnt.o",
                "ssl/statem_ntls/libssl-shlib-ntls_extensions_cust.o",
                "ssl/statem_ntls/libssl-shlib-ntls_extensions_srvr.o",
                "ssl/statem_ntls/libssl-shlib-ntls_statem.o",
                "ssl/statem_ntls/libssl-shlib-ntls_statem_clnt.o",
                "ssl/statem_ntls/libssl-shlib-ntls_statem_lib.o",
                "ssl/statem_ntls/libssl-shlib-ntls_statem_srvr.o"
            ],
            "products" => {
                "lib" => [
                    "libssl"
                ]
            }
        },
        "util" => {
            "products" => {
                "script" => [
                    "util/shlib_wrap.sh",
                    "util/wrap.pl"
                ]
            }
        }
    },
    "generate" => {
        "OpenSSLConfig.cmake" => [
            "exporters/cmake/OpenSSLConfig.cmake.in"
        ],
        "OpenSSLConfigVersion.cmake" => [
            "exporters/cmake/OpenSSLConfigVersion.cmake.in"
        ],
        "apps/progs.c" => [
            "apps/progs.pl",
            "\"-C\"",
            "\$(APPS_OPENSSL)"
        ],
        "apps/progs.h" => [
            "apps/progs.pl",
            "\"-H\"",
            "\$(APPS_OPENSSL)"
        ],
        "builddata.pm" => [
            "util/mkinstallvars.pl",
            "PREFIX=.",
            "BINDIR=apps",
            "APPLINKDIR=ms",
            "LIBDIR=",
            "INCLUDEDIR=include",
            "\"INCLUDEDIR=\$(SRCDIR)/include\"",
            "ENGINESDIR=engines",
            "MODULESDIR=providers",
            "\"VERSION=\$(VERSION)\"",
            "\"LDLIBS=\$(LIB_EX_LIBS)\""
        ],
        "crypto/aes/aes-586.S" => [
            "crypto/aes/asm/aes-586.pl"
        ],
        "crypto/aes/aes-armv4.S" => [
            "crypto/aes/asm/aes-armv4.pl"
        ],
        "crypto/aes/aes-c64xplus.S" => [
            "crypto/aes/asm/aes-c64xplus.pl"
        ],
        "crypto/aes/aes-ia64.s" => [
            "crypto/aes/asm/aes-ia64.S"
        ],
        "crypto/aes/aes-mips.S" => [
            "crypto/aes/asm/aes-mips.pl"
        ],
        "crypto/aes/aes-ppc.s" => [
            "crypto/aes/asm/aes-ppc.pl"
        ],
        "crypto/aes/aes-riscv32-zkn.s" => [
            "crypto/aes/asm/aes-riscv32-zkn.pl"
        ],
        "crypto/aes/aes-riscv64-zkn.s" => [
            "crypto/aes/asm/aes-riscv64-zkn.pl"
        ],
        "crypto/aes/aes-riscv64-zvbb-zvkg-zvkned.s" => [
            "crypto/aes/asm/aes-riscv64-zvbb-zvkg-zvkned.pl"
        ],
        "crypto/aes/aes-riscv64-zvkb-zvkned.s" => [
            "crypto/aes/asm/aes-riscv64-zvkb-zvkned.pl"
        ],
        "crypto/aes/aes-riscv64-zvkned.s" => [
            "crypto/aes/asm/aes-riscv64-zvkned.pl"
        ],
        "crypto/aes/aes-riscv64.s" => [
            "crypto/aes/asm/aes-riscv64.pl"
        ],
        "crypto/aes/aes-s390x.S" => [
            "crypto/aes/asm/aes-s390x.pl"
        ],
        "crypto/aes/aes-x86_64.s" => [
            "crypto/aes/asm/aes-x86_64.pl"
        ],
        "crypto/aes/aesni-mb-x86_64.s" => [
            "crypto/aes/asm/aesni-mb-x86_64.pl"
        ],
        "crypto/aes/aesni-sha1-x86_64.s" => [
            "crypto/aes/asm/aesni-sha1-x86_64.pl"
        ],
        "crypto/aes/aesni-sha256-x86_64.s" => [
            "crypto/aes/asm/aesni-sha256-x86_64.pl"
        ],
        "crypto/aes/aesni-x86.S" => [
            "crypto/aes/asm/aesni-x86.pl"
        ],
        "crypto/aes/aesni-x86_64.s" => [
            "crypto/aes/asm/aesni-x86_64.pl"
        ],
        "crypto/aes/aesni-xts-avx512.s" => [
            "crypto/aes/asm/aesni-xts-avx512.pl"
        ],
        "crypto/aes/aesp8-ppc.s" => [
            "crypto/aes/asm/aesp8-ppc.pl"
        ],
        "crypto/aes/aesv8-armx.S" => [
            "crypto/aes/asm/aesv8-armx.pl"
        ],
        "crypto/aes/bsaes-armv7.S" => [
            "crypto/aes/asm/bsaes-armv7.pl"
        ],
        "crypto/aes/bsaes-armv8.S" => [
            "crypto/aes/asm/bsaes-armv8.pl"
        ],
        "crypto/aes/bsaes-x86_64.s" => [
            "crypto/aes/asm/bsaes-x86_64.pl"
        ],
        "crypto/aes/vpaes-armv8.S" => [
            "crypto/aes/asm/vpaes-armv8.pl"
        ],
        "crypto/aes/vpaes-loongarch64.S" => [
            "crypto/aes/asm/vpaes-loongarch64.pl"
        ],
        "crypto/aes/vpaes-ppc.s" => [
            "crypto/aes/asm/vpaes-ppc.pl"
        ],
        "crypto/aes/vpaes-x86.S" => [
            "crypto/aes/asm/vpaes-x86.pl"
        ],
        "crypto/aes/vpaes-x86_64.s" => [
            "crypto/aes/asm/vpaes-x86_64.pl"
        ],
        "crypto/alphacpuid.s" => [
            "crypto/alphacpuid.pl"
        ],
        "crypto/arm64cpuid.S" => [
            "crypto/arm64cpuid.pl"
        ],
        "crypto/armv4cpuid.S" => [
            "crypto/armv4cpuid.pl"
        ],
        "crypto/bn/alpha-mont.S" => [
            "crypto/bn/asm/alpha-mont.pl"
        ],
        "crypto/bn/armv4-gf2m.S" => [
            "crypto/bn/asm/armv4-gf2m.pl"
        ],
        "crypto/bn/armv4-mont.S" => [
            "crypto/bn/asm/armv4-mont.pl"
        ],
        "crypto/bn/armv8-mont.S" => [
            "crypto/bn/asm/armv8-mont.pl"
        ],
        "crypto/bn/bn-586.S" => [
            "crypto/bn/asm/bn-586.pl"
        ],
        "crypto/bn/bn-ia64.s" => [
            "crypto/bn/asm/ia64.S"
        ],
        "crypto/bn/bn-mips.S" => [
            "crypto/bn/asm/mips.pl"
        ],
        "crypto/bn/bn-ppc.s" => [
            "crypto/bn/asm/ppc.pl"
        ],
        "crypto/bn/co-586.S" => [
            "crypto/bn/asm/co-586.pl"
        ],
        "crypto/bn/ia64-mont.s" => [
            "crypto/bn/asm/ia64-mont.pl"
        ],
        "crypto/bn/loongarch-mont.S" => [
            "crypto/bn/asm/loongarch-mont.pl"
        ],
        "crypto/bn/mips-mont.S" => [
            "crypto/bn/asm/mips-mont.pl"
        ],
        "crypto/bn/ppc-mont.s" => [
            "crypto/bn/asm/ppc-mont.pl"
        ],
        "crypto/bn/ppc64-mont-fixed.s" => [
            "crypto/bn/asm/ppc64-mont-fixed.pl"
        ],
        "crypto/bn/ppc64-mont.s" => [
            "crypto/bn/asm/ppc64-mont.pl"
        ],
        "crypto/bn/rsaz-2k-avx512.s" => [
            "crypto/bn/asm/rsaz-2k-avx512.pl"
        ],
        "crypto/bn/rsaz-2k-avxifma.s" => [
            "crypto/bn/asm/rsaz-2k-avxifma.pl"
        ],
        "crypto/bn/rsaz-3k-avx512.s" => [
            "crypto/bn/asm/rsaz-3k-avx512.pl"
        ],
        "crypto/bn/rsaz-3k-avxifma.s" => [
            "crypto/bn/asm/rsaz-3k-avxifma.pl"
        ],
        "crypto/bn/rsaz-4k-avx512.s" => [
            "crypto/bn/asm/rsaz-4k-avx512.pl"
        ],
        "crypto/bn/rsaz-4k-avxifma.s" => [
            "crypto/bn/asm/rsaz-4k-avxifma.pl"
        ],
        "crypto/bn/rsaz-avx2.s" => [
            "crypto/bn/asm/rsaz-avx2.pl"
        ],
        "crypto/bn/rsaz-x86_64.s" => [
            "crypto/bn/asm/rsaz-x86_64.pl"
        ],
        "crypto/bn/s390x-gf2m.s" => [
            "crypto/bn/asm/s390x-gf2m.pl"
        ],
        "crypto/bn/s390x-mont.S" => [
            "crypto/bn/asm/s390x-mont.pl"
        ],
        "crypto/bn/x86-gf2m.S" => [
            "crypto/bn/asm/x86-gf2m.pl"
        ],
        "crypto/bn/x86-mont.S" => [
            "crypto/bn/asm/x86-mont.pl"
        ],
        "crypto/bn/x86_64-gf2m.s" => [
            "crypto/bn/asm/x86_64-gf2m.pl"
        ],
        "crypto/bn/x86_64-mont.s" => [
            "crypto/bn/asm/x86_64-mont.pl"
        ],
        "crypto/bn/x86_64-mont5.s" => [
            "crypto/bn/asm/x86_64-mont5.pl"
        ],
        "crypto/buildinf.h" => [
            "util/mkbuildinf.pl",
            "\"\$(CC)",
            "\$(LIB_CFLAGS)",
            "\$(CPPFLAGS_Q)\"",
            "\"\$(PLATFORM)\""
        ],
        "crypto/chacha/chacha-armv4.S" => [
            "crypto/chacha/asm/chacha-armv4.pl"
        ],
        "crypto/chacha/chacha-armv8-sve.S" => [
            "crypto/chacha/asm/chacha-armv8-sve.pl"
        ],
        "crypto/chacha/chacha-armv8.S" => [
            "crypto/chacha/asm/chacha-armv8.pl"
        ],
        "crypto/chacha/chacha-c64xplus.S" => [
            "crypto/chacha/asm/chacha-c64xplus.pl"
        ],
        "crypto/chacha/chacha-ia64.S" => [
            "crypto/chacha/asm/chacha-ia64.pl"
        ],
        "crypto/chacha/chacha-ia64.s" => [
            "crypto/chacha/chacha-ia64.S"
        ],
        "crypto/chacha/chacha-loongarch64.S" => [
            "crypto/chacha/asm/chacha-loongarch64.pl"
        ],
        "crypto/chacha/chacha-ppc.s" => [
            "crypto/chacha/asm/chacha-ppc.pl"
        ],
        "crypto/chacha/chacha-riscv64-v-zbb-zvkb.s" => [
            "crypto/chacha/asm/chacha-riscv64-v-zbb.pl",
            "zvkb"
        ],
        "crypto/chacha/chacha-riscv64-v-zbb.s" => [
            "crypto/chacha/asm/chacha-riscv64-v-zbb.pl"
        ],
        "crypto/chacha/chacha-s390x.S" => [
            "crypto/chacha/asm/chacha-s390x.pl"
        ],
        "crypto/chacha/chacha-x86.S" => [
            "crypto/chacha/asm/chacha-x86.pl"
        ],
        "crypto/chacha/chacha-x86_64.s" => [
            "crypto/chacha/asm/chacha-x86_64.pl"
        ],
        "crypto/chacha/chachap10-ppc.s" => [
            "crypto/chacha/asm/chachap10-ppc.pl"
        ],
        "crypto/des/crypt586.S" => [
            "crypto/des/asm/crypt586.pl"
        ],
        "crypto/des/des-586.S" => [
            "crypto/des/asm/des-586.pl"
        ],
        "crypto/ec/ecp_nistp384-ppc64.s" => [
            "crypto/ec/asm/ecp_nistp384-ppc64.pl"
        ],
        "crypto/ec/ecp_nistp521-ppc64.s" => [
            "crypto/ec/asm/ecp_nistp521-ppc64.pl"
        ],
        "crypto/ec/ecp_nistz256-armv4.S" => [
            "crypto/ec/asm/ecp_nistz256-armv4.pl"
        ],
        "crypto/ec/ecp_nistz256-armv8.S" => [
            "crypto/ec/asm/ecp_nistz256-armv8.pl"
        ],
        "crypto/ec/ecp_nistz256-avx2.s" => [
            "crypto/ec/asm/ecp_nistz256-avx2.pl"
        ],
        "crypto/ec/ecp_nistz256-ppc64.s" => [
            "crypto/ec/asm/ecp_nistz256-ppc64.pl"
        ],
        "crypto/ec/ecp_nistz256-x86.S" => [
            "crypto/ec/asm/ecp_nistz256-x86.pl"
        ],
        "crypto/ec/ecp_nistz256-x86_64.s" => [
            "crypto/ec/asm/ecp_nistz256-x86_64.pl"
        ],
        "crypto/ec/ecp_sm2p256-armv8.S" => [
            "crypto/ec/asm/ecp_sm2p256-armv8.pl"
        ],
        "crypto/ec/x25519-ppc64.s" => [
            "crypto/ec/asm/x25519-ppc64.pl"
        ],
        "crypto/ec/x25519-x86_64.s" => [
            "crypto/ec/asm/x25519-x86_64.pl"
        ],
        "crypto/ia64cpuid.s" => [
            "crypto/ia64cpuid.S"
        ],
        "crypto/loongarch64cpuid.s" => [
            "crypto/loongarch64cpuid.pl"
        ],
        "crypto/md5/md5-586.S" => [
            "crypto/md5/asm/md5-586.pl"
        ],
        "crypto/md5/md5-aarch64.S" => [
            "crypto/md5/asm/md5-aarch64.pl"
        ],
        "crypto/md5/md5-loongarch64.S" => [
            "crypto/md5/asm/md5-loongarch64.pl"
        ],
        "crypto/md5/md5-x86_64.s" => [
            "crypto/md5/asm/md5-x86_64.pl"
        ],
        "crypto/modes/aes-gcm-armv8-unroll8_64.S" => [
            "crypto/modes/asm/aes-gcm-armv8-unroll8_64.pl"
        ],
        "crypto/modes/aes-gcm-armv8_64.S" => [
            "crypto/modes/asm/aes-gcm-armv8_64.pl"
        ],
        "crypto/modes/aes-gcm-avx512.s" => [
            "crypto/modes/asm/aes-gcm-avx512.pl"
        ],
        "crypto/modes/aes-gcm-ppc.s" => [
            "crypto/modes/asm/aes-gcm-ppc.pl"
        ],
        "crypto/modes/aes-gcm-riscv64-zvkb-zvkg-zvkned.s" => [
            "crypto/modes/asm/aes-gcm-riscv64-zvkb-zvkg-zvkned.pl"
        ],
        "crypto/modes/aesni-gcm-x86_64.s" => [
            "crypto/modes/asm/aesni-gcm-x86_64.pl"
        ],
        "crypto/modes/ghash-alpha.S" => [
            "crypto/modes/asm/ghash-alpha.pl"
        ],
        "crypto/modes/ghash-armv4.S" => [
            "crypto/modes/asm/ghash-armv4.pl"
        ],
        "crypto/modes/ghash-c64xplus.S" => [
            "crypto/modes/asm/ghash-c64xplus.pl"
        ],
        "crypto/modes/ghash-ia64.s" => [
            "crypto/modes/asm/ghash-ia64.pl"
        ],
        "crypto/modes/ghash-riscv64-zvkb-zvbc.s" => [
            "crypto/modes/asm/ghash-riscv64-zvkb-zvbc.pl"
        ],
        "crypto/modes/ghash-riscv64-zvkg.s" => [
            "crypto/modes/asm/ghash-riscv64-zvkg.pl"
        ],
        "crypto/modes/ghash-riscv64.s" => [
            "crypto/modes/asm/ghash-riscv64.pl"
        ],
        "crypto/modes/ghash-s390x.S" => [
            "crypto/modes/asm/ghash-s390x.pl"
        ],
        "crypto/modes/ghash-x86.S" => [
            "crypto/modes/asm/ghash-x86.pl"
        ],
        "crypto/modes/ghash-x86_64.s" => [
            "crypto/modes/asm/ghash-x86_64.pl"
        ],
        "crypto/modes/ghashp8-ppc.s" => [
            "crypto/modes/asm/ghashp8-ppc.pl"
        ],
        "crypto/modes/ghashv8-armx.S" => [
            "crypto/modes/asm/ghashv8-armx.pl"
        ],
        "crypto/params_idx.c" => [
            "crypto/params_idx.c.in"
        ],
        "crypto/poly1305/poly1305-armv4.S" => [
            "crypto/poly1305/asm/poly1305-armv4.pl"
        ],
        "crypto/poly1305/poly1305-armv8.S" => [
            "crypto/poly1305/asm/poly1305-armv8.pl"
        ],
        "crypto/poly1305/poly1305-c64xplus.S" => [
            "crypto/poly1305/asm/poly1305-c64xplus.pl"
        ],
        "crypto/poly1305/poly1305-ia64.s" => [
            "crypto/poly1305/asm/poly1305-ia64.S"
        ],
        "crypto/poly1305/poly1305-mips.S" => [
            "crypto/poly1305/asm/poly1305-mips.pl"
        ],
        "crypto/poly1305/poly1305-ppc.s" => [
            "crypto/poly1305/asm/poly1305-ppc.pl"
        ],
        "crypto/poly1305/poly1305-ppcfp.s" => [
            "crypto/poly1305/asm/poly1305-ppcfp.pl"
        ],
        "crypto/poly1305/poly1305-s390x.S" => [
            "crypto/poly1305/asm/poly1305-s390x.pl"
        ],
        "crypto/poly1305/poly1305-x86.S" => [
            "crypto/poly1305/asm/poly1305-x86.pl"
        ],
        "crypto/poly1305/poly1305-x86_64.s" => [
            "crypto/poly1305/asm/poly1305-x86_64.pl"
        ],
        "crypto/ppccpuid.s" => [
            "crypto/ppccpuid.pl"
        ],
        "crypto/rc4/rc4-586.S" => [
            "crypto/rc4/asm/rc4-586.pl"
        ],
        "crypto/rc4/rc4-c64xplus.s" => [
            "crypto/rc4/asm/rc4-c64xplus.pl"
        ],
        "crypto/rc4/rc4-md5-x86_64.s" => [
            "crypto/rc4/asm/rc4-md5-x86_64.pl"
        ],
        "crypto/rc4/rc4-s390x.s" => [
            "crypto/rc4/asm/rc4-s390x.pl"
        ],
        "crypto/rc4/rc4-x86_64.s" => [
            "crypto/rc4/asm/rc4-x86_64.pl"
        ],
        "crypto/riscv32cpuid.s" => [
            "crypto/riscv32cpuid.pl"
        ],
        "crypto/riscv64cpuid.s" => [
            "crypto/riscv64cpuid.pl"
        ],
        "crypto/s390xcpuid.S" => [
            "crypto/s390xcpuid.pl"
        ],
        "crypto/sha/keccak1600-armv4.S" => [
            "crypto/sha/asm/keccak1600-armv4.pl"
        ],
        "crypto/sha/keccak1600-armv8.S" => [
            "crypto/sha/asm/keccak1600-armv8.pl"
        ],
        "crypto/sha/keccak1600-avx2.S" => [
            "crypto/sha/asm/keccak1600-avx2.pl"
        ],
        "crypto/sha/keccak1600-avx512.S" => [
            "crypto/sha/asm/keccak1600-avx512.pl"
        ],
        "crypto/sha/keccak1600-avx512vl.S" => [
            "crypto/sha/asm/keccak1600-avx512vl.pl"
        ],
        "crypto/sha/keccak1600-c64x.S" => [
            "crypto/sha/asm/keccak1600-c64x.pl"
        ],
        "crypto/sha/keccak1600-mmx.S" => [
            "crypto/sha/asm/keccak1600-mmx.pl"
        ],
        "crypto/sha/keccak1600-ppc64.s" => [
            "crypto/sha/asm/keccak1600-ppc64.pl"
        ],
        "crypto/sha/keccak1600-s390x.S" => [
            "crypto/sha/asm/keccak1600-s390x.pl"
        ],
        "crypto/sha/keccak1600-x86_64.s" => [
            "crypto/sha/asm/keccak1600-x86_64.pl"
        ],
        "crypto/sha/keccak1600p8-ppc.S" => [
            "crypto/sha/asm/keccak1600p8-ppc.pl"
        ],
        "crypto/sha/sha1-586.S" => [
            "crypto/sha/asm/sha1-586.pl"
        ],
        "crypto/sha/sha1-alpha.S" => [
            "crypto/sha/asm/sha1-alpha.pl"
        ],
        "crypto/sha/sha1-armv4-large.S" => [
            "crypto/sha/asm/sha1-armv4-large.pl"
        ],
        "crypto/sha/sha1-armv8.S" => [
            "crypto/sha/asm/sha1-armv8.pl"
        ],
        "crypto/sha/sha1-c64xplus.S" => [
            "crypto/sha/asm/sha1-c64xplus.pl"
        ],
        "crypto/sha/sha1-ia64.s" => [
            "crypto/sha/asm/sha1-ia64.pl"
        ],
        "crypto/sha/sha1-mb-x86_64.s" => [
            "crypto/sha/asm/sha1-mb-x86_64.pl"
        ],
        "crypto/sha/sha1-mips.S" => [
            "crypto/sha/asm/sha1-mips.pl"
        ],
        "crypto/sha/sha1-ppc.s" => [
            "crypto/sha/asm/sha1-ppc.pl"
        ],
        "crypto/sha/sha1-s390x.S" => [
            "crypto/sha/asm/sha1-s390x.pl"
        ],
        "crypto/sha/sha1-thumb.S" => [
            "crypto/sha/asm/sha1-thumb.pl"
        ],
        "crypto/sha/sha1-x86_64.s" => [
            "crypto/sha/asm/sha1-x86_64.pl"
        ],
        "crypto/sha/sha256-586.S" => [
            "crypto/sha/asm/sha256-586.pl"
        ],
        "crypto/sha/sha256-armv4.S" => [
            "crypto/sha/asm/sha256-armv4.pl"
        ],
        "crypto/sha/sha256-armv8.S" => [
            "crypto/sha/asm/sha512-armv8.pl"
        ],
        "crypto/sha/sha256-c64xplus.S" => [
            "crypto/sha/asm/sha256-c64xplus.pl"
        ],
        "crypto/sha/sha256-ia64.s" => [
            "crypto/sha/asm/sha512-ia64.pl"
        ],
        "crypto/sha/sha256-mb-x86_64.s" => [
            "crypto/sha/asm/sha256-mb-x86_64.pl"
        ],
        "crypto/sha/sha256-mips.S" => [
            "crypto/sha/asm/sha512-mips.pl"
        ],
        "crypto/sha/sha256-ppc.s" => [
            "crypto/sha/asm/sha512-ppc.pl"
        ],
        "crypto/sha/sha256-riscv64-zvkb-zvknha_or_zvknhb.S" => [
            "crypto/sha/asm/sha256-riscv64-zvkb-zvknha_or_zvknhb.pl"
        ],
        "crypto/sha/sha256-s390x.S" => [
            "crypto/sha/asm/sha512-s390x.pl"
        ],
        "crypto/sha/sha256-x86_64.s" => [
            "crypto/sha/asm/sha512-x86_64.pl"
        ],
        "crypto/sha/sha256p8-ppc.s" => [
            "crypto/sha/asm/sha512p8-ppc.pl"
        ],
        "crypto/sha/sha512-586.S" => [
            "crypto/sha/asm/sha512-586.pl"
        ],
        "crypto/sha/sha512-armv4.S" => [
            "crypto/sha/asm/sha512-armv4.pl"
        ],
        "crypto/sha/sha512-armv8.S" => [
            "crypto/sha/asm/sha512-armv8.pl"
        ],
        "crypto/sha/sha512-c64xplus.S" => [
            "crypto/sha/asm/sha512-c64xplus.pl"
        ],
        "crypto/sha/sha512-ia64.s" => [
            "crypto/sha/asm/sha512-ia64.pl"
        ],
        "crypto/sha/sha512-mips.S" => [
            "crypto/sha/asm/sha512-mips.pl"
        ],
        "crypto/sha/sha512-ppc.s" => [
            "crypto/sha/asm/sha512-ppc.pl"
        ],
        "crypto/sha/sha512-riscv64-zvkb-zvknhb.S" => [
            "crypto/sha/asm/sha512-riscv64-zvkb-zvknhb.pl"
        ],
        "crypto/sha/sha512-s390x.S" => [
            "crypto/sha/asm/sha512-s390x.pl"
        ],
        "crypto/sha/sha512-x86_64.s" => [
            "crypto/sha/asm/sha512-x86_64.pl"
        ],
        "crypto/sha/sha512p8-ppc.s" => [
            "crypto/sha/asm/sha512p8-ppc.pl"
        ],
        "crypto/sm3/sm3-armv8.S" => [
            "crypto/sm3/asm/sm3-armv8.pl"
        ],
        "crypto/sm3/sm3-riscv64-zvksh.S" => [
            "crypto/sm3/asm/sm3-riscv64-zvksh.pl"
        ],
        "crypto/sm4/sm4-armv8.S" => [
            "crypto/sm4/asm/sm4-armv8.pl"
        ],
        "crypto/sm4/sm4-riscv64-zvksed.s" => [
            "crypto/sm4/asm/sm4-riscv64-zvksed.pl"
        ],
        "crypto/sm4/vpsm4-armv8.S" => [
            "crypto/sm4/asm/vpsm4-armv8.pl"
        ],
        "crypto/sm4/vpsm4_ex-armv8.S" => [
            "crypto/sm4/asm/vpsm4_ex-armv8.pl"
        ],
        "crypto/uplink-ia64.s" => [
            "ms/uplink-ia64.pl"
        ],
        "crypto/uplink-x86.S" => [
            "ms/uplink-x86.pl"
        ],
        "crypto/uplink-x86_64.s" => [
            "ms/uplink-x86_64.pl"
        ],
        "crypto/x86_64cpuid.s" => [
            "crypto/x86_64cpuid.pl"
        ],
        "crypto/x86cpuid.S" => [
            "crypto/x86cpuid.pl"
        ],
        "engines/afalg.ld" => [
            "util/engines.num"
        ],
        "engines/capi.ld" => [
            "util/engines.num"
        ],
        "engines/dasync.ld" => [
            "util/engines.num"
        ],
        "engines/e_padlock-x86.S" => [
            "engines/asm/e_padlock-x86.pl"
        ],
        "engines/e_padlock-x86_64.s" => [
            "engines/asm/e_padlock-x86_64.pl"
        ],
        "engines/ecptest.ld" => [
            "util/engines.num"
        ],
        "engines/loader_attic.ld" => [
            "util/engines.num"
        ],
        "engines/ossltest.ld" => [
            "util/engines.num"
        ],
        "engines/padlock.ld" => [
            "util/engines.num"
        ],
        "exporters/OpenSSLConfig.cmake" => [
            "exporters/cmake/OpenSSLConfig.cmake.in"
        ],
        "exporters/OpenSSLConfigVersion.cmake" => [
            "exporters/cmake/OpenSSLConfigVersion.cmake.in"
        ],
        "exporters/libcrypto.pc" => [
            "exporters/pkg-config/libcrypto.pc.in"
        ],
        "exporters/libssl.pc" => [
            "exporters/pkg-config/libssl.pc.in"
        ],
        "exporters/openssl.pc" => [
            "exporters/pkg-config/openssl.pc.in"
        ],
        "include/crypto/bn_conf.h" => [
            "include/crypto/bn_conf.h.in"
        ],
        "include/crypto/dso_conf.h" => [
            "include/crypto/dso_conf.h.in"
        ],
        "include/internal/param_names.h" => [
            "include/internal/param_names.h.in"
        ],
        "include/openssl/asn1.h" => [
            "include/openssl/asn1.h.in"
        ],
        "include/openssl/asn1t.h" => [
            "include/openssl/asn1t.h.in"
        ],
        "include/openssl/bio.h" => [
            "include/openssl/bio.h.in"
        ],
        "include/openssl/cmp.h" => [
            "include/openssl/cmp.h.in"
        ],
        "include/openssl/cms.h" => [
            "include/openssl/cms.h.in"
        ],
        "include/openssl/comp.h" => [
            "include/openssl/comp.h.in"
        ],
        "include/openssl/conf.h" => [
            "include/openssl/conf.h.in"
        ],
        "include/openssl/configuration.h" => [
            "include/openssl/configuration.h.in"
        ],
        "include/openssl/core_names.h" => [
            "include/openssl/core_names.h.in"
        ],
        "include/openssl/crmf.h" => [
            "include/openssl/crmf.h.in"
        ],
        "include/openssl/crypto.h" => [
            "include/openssl/crypto.h.in"
        ],
        "include/openssl/ct.h" => [
            "include/openssl/ct.h.in"
        ],
        "include/openssl/err.h" => [
            "include/openssl/err.h.in"
        ],
        "include/openssl/ess.h" => [
            "include/openssl/ess.h.in"
        ],
        "include/openssl/fipskey.h" => [
            "include/openssl/fipskey.h.in"
        ],
        "include/openssl/lhash.h" => [
            "include/openssl/lhash.h.in"
        ],
        "include/openssl/ocsp.h" => [
            "include/openssl/ocsp.h.in"
        ],
        "include/openssl/opensslv.h" => [
            "include/openssl/opensslv.h.in"
        ],
        "include/openssl/pkcs12.h" => [
            "include/openssl/pkcs12.h.in"
        ],
        "include/openssl/pkcs7.h" => [
            "include/openssl/pkcs7.h.in"
        ],
        "include/openssl/safestack.h" => [
            "include/openssl/safestack.h.in"
        ],
        "include/openssl/srp.h" => [
            "include/openssl/srp.h.in"
        ],
        "include/openssl/ssl.h" => [
            "include/openssl/ssl.h.in"
        ],
        "include/openssl/symbol_prefix.h" => [
            "util/mk_symbol_prefix.pl"
        ],
        "include/openssl/ui.h" => [
            "include/openssl/ui.h.in"
        ],
        "include/openssl/x509.h" => [
            "include/openssl/x509.h.in"
        ],
        "include/openssl/x509_acert.h" => [
            "include/openssl/x509_acert.h.in"
        ],
        "include/openssl/x509_vfy.h" => [
            "include/openssl/x509_vfy.h.in"
        ],
        "include/openssl/x509v3.h" => [
            "include/openssl/x509v3.h.in"
        ],
        "installdata.pm" => [
            "util/mkinstallvars.pl",
            "\"PREFIX=\$(INSTALLTOP)\"",
            "BINDIR=bin",
            "\"LIBDIR=\$(LIBDIR)\"",
            "\"libdir=\$(libdir)\"",
            "INCLUDEDIR=include",
            "APPLINKDIR=include/openssl",
            "\"ENGINESDIR=\$(ENGINESDIR)\"",
            "\"MODULESDIR=\$(MODULESDIR)\"",
            "\"PKGCONFIGDIR=\$(PKGCONFIGDIR)\"",
            "\"CMAKECONFIGDIR=\$(CMAKECONFIGDIR)\"",
            "\"LDLIBS=\$(LIB_EX_LIBS)\"",
            "\"VERSION=\$(VERSION)\""
        ],
        "libcrypto.ld" => [
            "util/libcrypto.num",
            "libcrypto"
        ],
        "libcrypto.pc" => [
            "exporters/pkg-config/libcrypto.pc.in"
        ],
        "libssl.ld" => [
            "util/libssl.num",
            "libssl"
        ],
        "libssl.pc" => [
            "exporters/pkg-config/libssl.pc.in"
        ],
        "openssl.pc" => [
            "exporters/pkg-config/openssl.pc.in"
        ],
        "providers/common/der/der_digests_gen.c" => [
            "providers/common/der/der_digests_gen.c.in"
        ],
        "providers/common/der/der_dsa_gen.c" => [
            "providers/common/der/der_dsa_gen.c.in"
        ],
        "providers/common/der/der_ec_gen.c" => [
            "providers/common/der/der_ec_gen.c.in"
        ],
        "providers/common/der/der_ecx_gen.c" => [
            "providers/common/der/der_ecx_gen.c.in"
        ],
        "providers/common/der/der_ml_dsa_gen.c" => [
            "providers/common/der/der_ml_dsa_gen.c.in"
        ],
        "providers/common/der/der_rsa_gen.c" => [
            "providers/common/der/der_rsa_gen.c.in"
        ],
        "providers/common/der/der_slh_dsa_gen.c" => [
            "providers/common/der/der_slh_dsa_gen.c.in"
        ],
        "providers/common/der/der_sm2_gen.c" => [
            "providers/common/der/der_sm2_gen.c.in"
        ],
        "providers/common/der/der_wrap_gen.c" => [
            "providers/common/der/der_wrap_gen.c.in"
        ],
        "providers/common/include/prov/der_digests.h" => [
            "providers/common/include/prov/der_digests.h.in"
        ],
        "providers/common/include/prov/der_dsa.h" => [
            "providers/common/include/prov/der_dsa.h.in"
        ],
        "providers/common/include/prov/der_ec.h" => [
            "providers/common/include/prov/der_ec.h.in"
        ],
        "providers/common/include/prov/der_ecx.h" => [
            "providers/common/include/prov/der_ecx.h.in"
        ],
        "providers/common/include/prov/der_ml_dsa.h" => [
            "providers/common/include/prov/der_ml_dsa.h.in"
        ],
        "providers/common/include/prov/der_rsa.h" => [
            "providers/common/include/prov/der_rsa.h.in"
        ],
        "providers/common/include/prov/der_slh_dsa.h" => [
            "providers/common/include/prov/der_slh_dsa.h.in"
        ],
        "providers/common/include/prov/der_sm2.h" => [
            "providers/common/include/prov/der_sm2.h.in"
        ],
        "providers/common/include/prov/der_wrap.h" => [
            "providers/common/include/prov/der_wrap.h.in"
        ],
        "providers/legacy.ld" => [
            "util/providers.num"
        ]
    },
    "includes" => {
        "OpenSSLConfig.cmake" => [
            "."
        ],
        "OpenSSLConfigVersion.cmake" => [
            "."
        ],
        "apps/asn1parse.o" => [
            "apps"
        ],
        "apps/ca.o" => [
            "apps"
        ],
        "apps/ciphers.o" => [
            "apps"
        ],
        "apps/cmp.o" => [
            "apps"
        ],
        "apps/cms.o" => [
            "apps"
        ],
        "apps/crl.o" => [
            "apps"
        ],
        "apps/crl2pkcs7.o" => [
            "apps"
        ],
        "apps/dgst.o" => [
            "apps"
        ],
        "apps/dhparam.o" => [
            "apps"
        ],
        "apps/dsa.o" => [
            "apps"
        ],
        "apps/dsaparam.o" => [
            "apps"
        ],
        "apps/ec.o" => [
            "apps"
        ],
        "apps/ecparam.o" => [
            "apps"
        ],
        "apps/enc.o" => [
            "apps"
        ],
        "apps/engine.o" => [
            "apps"
        ],
        "apps/errstr.o" => [
            "apps"
        ],
        "apps/fipsinstall.o" => [
            "apps"
        ],
        "apps/gendsa.o" => [
            "apps"
        ],
        "apps/genpkey.o" => [
            "apps"
        ],
        "apps/genrsa.o" => [
            "apps"
        ],
        "apps/info.o" => [
            "apps"
        ],
        "apps/kdf.o" => [
            "apps"
        ],
        "apps/lib/cmp_mock_srv.o" => [
            "apps"
        ],
        "apps/lib/openssl-bin-cmp_mock_srv.o" => [
            "apps"
        ],
        "apps/libapps.a" => [
            ".",
            "include",
            "apps/include"
        ],
        "apps/list.o" => [
            "apps"
        ],
        "apps/mac.o" => [
            "apps"
        ],
        "apps/nseq.o" => [
            "apps"
        ],
        "apps/ocsp.o" => [
            "apps"
        ],
        "apps/openssl" => [
            ".",
            "include",
            "apps/include"
        ],
        "apps/openssl-bin-asn1parse.o" => [
            "apps"
        ],
        "apps/openssl-bin-ca.o" => [
            "apps"
        ],
        "apps/openssl-bin-ciphers.o" => [
            "apps"
        ],
        "apps/openssl-bin-cmp.o" => [
            "apps"
        ],
        "apps/openssl-bin-cms.o" => [
            "apps"
        ],
        "apps/openssl-bin-crl.o" => [
            "apps"
        ],
        "apps/openssl-bin-crl2pkcs7.o" => [
            "apps"
        ],
        "apps/openssl-bin-dgst.o" => [
            "apps"
        ],
        "apps/openssl-bin-dhparam.o" => [
            "apps"
        ],
        "apps/openssl-bin-dsa.o" => [
            "apps"
        ],
        "apps/openssl-bin-dsaparam.o" => [
            "apps"
        ],
        "apps/openssl-bin-ec.o" => [
            "apps"
        ],
        "apps/openssl-bin-ecparam.o" => [
            "apps"
        ],
        "apps/openssl-bin-enc.o" => [
            "apps"
        ],
        "apps/openssl-bin-engine.o" => [
            "apps"
        ],
        "apps/openssl-bin-errstr.o" => [
            "apps"
        ],
        "apps/openssl-bin-fipsinstall.o" => [
            "apps"
        ],
        "apps/openssl-bin-gendsa.o" => [
            "apps"
        ],
        "apps/openssl-bin-genpkey.o" => [
            "apps"
        ],
        "apps/openssl-bin-genrsa.o" => [
            "apps"
        ],
        "apps/openssl-bin-info.o" => [
            "apps"
        ],
        "apps/openssl-bin-kdf.o" => [
            "apps"
        ],
        "apps/openssl-bin-list.o" => [
            "apps"
        ],
        "apps/openssl-bin-mac.o" => [
            "apps"
        ],
        "apps/openssl-bin-nseq.o" => [
            "apps"
        ],
        "apps/openssl-bin-ocsp.o" => [
            "apps"
        ],
        "apps/openssl-bin-openssl.o" => [
            "apps"
        ],
        "apps/openssl-bin-passwd.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkcs12.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkcs7.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkcs8.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkey.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkeyparam.o" => [
            "apps"
        ],
        "apps/openssl-bin-pkeyutl.o" => [
            "apps"
        ],
        "apps/openssl-bin-prime.o" => [
            "apps"
        ],
        "apps/openssl-bin-progs.o" => [
            "apps"
        ],
        "apps/openssl-bin-rand.o" => [
            "apps"
        ],
        "apps/openssl-bin-rehash.o" => [
            "apps"
        ],
        "apps/openssl-bin-req.o" => [
            "apps"
        ],
        "apps/openssl-bin-rsa.o" => [
            "apps"
        ],
        "apps/openssl-bin-rsautl.o" => [
            "apps"
        ],
        "apps/openssl-bin-s_client.o" => [
            "apps"
        ],
        "apps/openssl-bin-s_server.o" => [
            "apps"
        ],
        "apps/openssl-bin-s_time.o" => [
            "apps"
        ],
        "apps/openssl-bin-sess_id.o" => [
            "apps"
        ],
        "apps/openssl-bin-skeyutl.o" => [
            "apps"
        ],
        "apps/openssl-bin-smime.o" => [
            "apps"
        ],
        "apps/openssl-bin-speed.o" => [
            "apps"
        ],
        "apps/openssl-bin-spkac.o" => [
            "apps"
        ],
        "apps/openssl-bin-srp.o" => [
            "apps"
        ],
        "apps/openssl-bin-storeutl.o" => [
            "apps"
        ],
        "apps/openssl-bin-ts.o" => [
            "apps"
        ],
        "apps/openssl-bin-verify.o" => [
            "apps"
        ],
        "apps/openssl-bin-version.o" => [
            "apps"
        ],
        "apps/openssl-bin-x509.o" => [
            "apps"
        ],
        "apps/openssl.o" => [
            "apps"
        ],
        "apps/passwd.o" => [
            "apps"
        ],
        "apps/pkcs12.o" => [
            "apps"
        ],
        "apps/pkcs7.o" => [
            "apps"
        ],
        "apps/pkcs8.o" => [
            "apps"
        ],
        "apps/pkey.o" => [
            "apps"
        ],
        "apps/pkeyparam.o" => [
            "apps"
        ],
        "apps/pkeyutl.o" => [
            "apps"
        ],
        "apps/prime.o" => [
            "apps"
        ],
        "apps/progs.c" => [
            "."
        ],
        "apps/progs.o" => [
            "apps"
        ],
        "apps/rand.o" => [
            "apps"
        ],
        "apps/rehash.o" => [
            "apps"
        ],
        "apps/req.o" => [
            "apps"
        ],
        "apps/rsa.o" => [
            "apps"
        ],
        "apps/rsautl.o" => [
            "apps"
        ],
        "apps/s_client.o" => [
            "apps"
        ],
        "apps/s_server.o" => [
            "apps"
        ],
        "apps/s_time.o" => [
            "apps"
        ],
        "apps/sess_id.o" => [
            "apps"
        ],
        "apps/skeyutl.o" => [
            "apps"
        ],
        "apps/smime.o" => [
            "apps"
        ],
        "apps/speed.o" => [
            "apps"
        ],
        "apps/spkac.o" => [
            "apps"
        ],
        "apps/srp.o" => [
            "apps"
        ],
        "apps/storeutl.o" => [
            "apps"
        ],
        "apps/ts.o" => [
            "apps"
        ],
        "apps/verify.o" => [
            "apps"
        ],
        "apps/version.o" => [
            "apps"
        ],
        "apps/x509.o" => [
            "apps"
        ],
        "crypto/aes/aes-armv4.o" => [
            "crypto"
        ],
        "crypto/aes/aes-mips.o" => [
            "crypto"
        ],
        "crypto/aes/aes-s390x.o" => [
            "crypto"
        ],
        "crypto/aes/aesv8-armx.o" => [
            "crypto"
        ],
        "crypto/aes/bsaes-armv7.o" => [
            "crypto"
        ],
        "crypto/aes/vpaes-armv8.o" => [
            "crypto"
        ],
        "crypto/aes/vpaes-loongarch64.o" => [
            "crypto"
        ],
        "crypto/arm64cpuid.o" => [
            "crypto"
        ],
        "crypto/armv4cpuid.o" => [
            "crypto"
        ],
        "crypto/bn/armv4-gf2m.o" => [
            "crypto"
        ],
        "crypto/bn/armv4-mont.o" => [
            "crypto"
        ],
        "crypto/bn/armv8-mont.o" => [
            "crypto"
        ],
        "crypto/bn/bn-mips.o" => [
            "crypto"
        ],
        "crypto/bn/bn_exp.o" => [
            "crypto"
        ],
        "crypto/bn/libcrypto-lib-bn_exp.o" => [
            "crypto"
        ],
        "crypto/bn/libcrypto-shlib-bn_exp.o" => [
            "crypto"
        ],
        "crypto/bn/mips-mont.o" => [
            "crypto"
        ],
        "crypto/chacha/chacha-armv4.o" => [
            "crypto"
        ],
        "crypto/chacha/chacha-armv8-sve.o" => [
            "crypto"
        ],
        "crypto/chacha/chacha-armv8.o" => [
            "crypto"
        ],
        "crypto/chacha/chacha-loongarch64.o" => [
            "crypto"
        ],
        "crypto/chacha/chacha-s390x.o" => [
            "crypto"
        ],
        "crypto/cpuid.o" => [
            "."
        ],
        "crypto/cversion.o" => [
            "crypto"
        ],
        "crypto/ec/ecp_nistz256-armv4.o" => [
            "crypto"
        ],
        "crypto/ec/ecp_nistz256-armv8.o" => [
            "crypto"
        ],
        "crypto/ec/ecp_s390x_nistp.o" => [
            "crypto"
        ],
        "crypto/ec/ecp_sm2p256-armv8.o" => [
            "crypto"
        ],
        "crypto/ec/ecx_key.o" => [
            "crypto"
        ],
        "crypto/ec/ecx_meth.o" => [
            "crypto"
        ],
        "crypto/ec/ecx_s390x.o" => [
            "crypto"
        ],
        "crypto/ec/libcrypto-lib-ecx_key.o" => [
            "crypto"
        ],
        "crypto/ec/libcrypto-lib-ecx_meth.o" => [
            "crypto"
        ],
        "crypto/ec/libcrypto-shlib-ecx_key.o" => [
            "crypto"
        ],
        "crypto/ec/libcrypto-shlib-ecx_meth.o" => [
            "crypto"
        ],
        "crypto/evp/e_aes.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/evp/e_aes_cbc_hmac_sha1.o" => [
            "crypto/modes"
        ],
        "crypto/evp/e_aes_cbc_hmac_sha256.o" => [
            "crypto/modes"
        ],
        "crypto/evp/e_des.o" => [
            "crypto"
        ],
        "crypto/evp/e_des3.o" => [
            "crypto"
        ],
        "crypto/evp/e_sm4.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-lib-e_aes.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha1.o" => [
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha256.o" => [
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-lib-e_des.o" => [
            "crypto"
        ],
        "crypto/evp/libcrypto-lib-e_des3.o" => [
            "crypto"
        ],
        "crypto/evp/libcrypto-lib-e_sm4.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-shlib-e_aes.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha1.o" => [
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha256.o" => [
            "crypto/modes"
        ],
        "crypto/evp/libcrypto-shlib-e_des.o" => [
            "crypto"
        ],
        "crypto/evp/libcrypto-shlib-e_des3.o" => [
            "crypto"
        ],
        "crypto/evp/libcrypto-shlib-e_sm4.o" => [
            "crypto",
            "crypto/modes"
        ],
        "crypto/info.o" => [
            "crypto"
        ],
        "crypto/legacy-dso-cpuid.o" => [
            "."
        ],
        "crypto/libcrypto-lib-cpuid.o" => [
            "."
        ],
        "crypto/libcrypto-lib-cversion.o" => [
            "crypto"
        ],
        "crypto/libcrypto-lib-info.o" => [
            "crypto"
        ],
        "crypto/libcrypto-shlib-cpuid.o" => [
            "."
        ],
        "crypto/libcrypto-shlib-cversion.o" => [
            "crypto"
        ],
        "crypto/libcrypto-shlib-info.o" => [
            "crypto"
        ],
        "crypto/md5/md5-aarch64.o" => [
            "crypto"
        ],
        "crypto/md5/md5-loongarch64.o" => [
            "crypto"
        ],
        "crypto/modes/aes-gcm-armv8-unroll8_64.o" => [
            "crypto"
        ],
        "crypto/modes/aes-gcm-armv8_64.o" => [
            "crypto"
        ],
        "crypto/modes/gcm128.o" => [
            "crypto"
        ],
        "crypto/modes/ghash-armv4.o" => [
            "crypto"
        ],
        "crypto/modes/ghash-s390x.o" => [
            "crypto"
        ],
        "crypto/modes/ghashv8-armx.o" => [
            "crypto"
        ],
        "crypto/modes/libcrypto-lib-gcm128.o" => [
            "crypto"
        ],
        "crypto/modes/libcrypto-shlib-gcm128.o" => [
            "crypto"
        ],
        "crypto/params_idx.c" => [
            "util/perl"
        ],
        "crypto/poly1305/poly1305-armv4.o" => [
            "crypto"
        ],
        "crypto/poly1305/poly1305-armv8.o" => [
            "crypto"
        ],
        "crypto/poly1305/poly1305-mips.o" => [
            "crypto"
        ],
        "crypto/poly1305/poly1305-s390x.o" => [
            "crypto"
        ],
        "crypto/s390xcpuid.o" => [
            "crypto"
        ],
        "crypto/sha/keccak1600-armv4.o" => [
            "crypto"
        ],
        "crypto/sha/keccak1600-armv8.o" => [
            "crypto"
        ],
        "crypto/sha/sha1-armv4-large.o" => [
            "crypto"
        ],
        "crypto/sha/sha1-armv8.o" => [
            "crypto"
        ],
        "crypto/sha/sha1-mips.o" => [
            "crypto"
        ],
        "crypto/sha/sha1-s390x.o" => [
            "crypto"
        ],
        "crypto/sha/sha256-armv4.o" => [
            "crypto"
        ],
        "crypto/sha/sha256-armv8.o" => [
            "crypto"
        ],
        "crypto/sha/sha256-mips.o" => [
            "crypto"
        ],
        "crypto/sha/sha256-s390x.o" => [
            "crypto"
        ],
        "crypto/sha/sha512-armv4.o" => [
            "crypto"
        ],
        "crypto/sha/sha512-armv8.o" => [
            "crypto"
        ],
        "crypto/sha/sha512-mips.o" => [
            "crypto"
        ],
        "crypto/sha/sha512-s390x.o" => [
            "crypto"
        ],
        "crypto/sm3/sm3-armv8.o" => [
            "crypto"
        ],
        "crypto/sm4/sm4-armv8.o" => [
            "crypto"
        ],
        "crypto/sm4/vpsm4-armv8.o" => [
            "crypto"
        ],
        "crypto/sm4/vpsm4_ex-armv8.o" => [
            "crypto"
        ],
        "engines/afalg" => [
            "include"
        ],
        "engines/capi" => [
            "include"
        ],
        "engines/dasync" => [
            "include"
        ],
        "engines/ecptest" => [
            "include"
        ],
        "engines/loader_attic" => [
            "include"
        ],
        "engines/ossltest" => [
            "include"
        ],
        "engines/padlock" => [
            "include"
        ],
        "exporters/OpenSSLConfig.cmake" => [
            "."
        ],
        "exporters/OpenSSLConfigVersion.cmake" => [
            "."
        ],
        "exporters/libcrypto.pc" => [
            "."
        ],
        "exporters/libssl.pc" => [
            "."
        ],
        "exporters/openssl.pc" => [
            "."
        ],
        "include/internal/param_names.h" => [
            "util/perl"
        ],
        "include/openssl/core_names.h" => [
            "util/perl"
        ],
        "include/openssl/symbol_prefix.h" => [
            "providers/common/include/prov",
            ".",
            "util/perl/OpenSSL"
        ],
        "libcrypto" => [
            ".",
            "include",
            "providers/common/include",
            "providers/implementations/include"
        ],
        "libcrypto.ld" => [
            ".",
            "util/perl/OpenSSL"
        ],
        "libcrypto.pc" => [
            "."
        ],
        "libssl" => [
            ".",
            "include"
        ],
        "libssl.ld" => [
            ".",
            "util/perl/OpenSSL"
        ],
        "libssl.pc" => [
            "."
        ],
        "openssl.pc" => [
            "."
        ],
        "providers/common/der/der_digests_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_digests_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_dsa_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ec_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_ec_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ec_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ec_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ecx_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_ecx_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ecx_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ml_dsa_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_ml_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_ml_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_rsa_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_rsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_rsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_rsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_slh_dsa_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_slh_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_slh_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_sm2_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_sm2_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_sm2_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_sm2_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/der_wrap_gen.c" => [
            "providers/common/der"
        ],
        "providers/common/der/der_wrap_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_digests_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ec_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ec_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ec_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ecx_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ecx_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_rsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_rsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libcommon-lib-der_wrap_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libdefault-lib-der_rsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libdefault-lib-der_sm2_gen.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libdefault-lib-der_sm2_key.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/der/libdefault-lib-der_sm2_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/common/include/prov/der_digests.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_dsa.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_ec.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_ecx.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_ml_dsa.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_rsa.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_slh_dsa.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_sm2.h" => [
            "providers/common/der"
        ],
        "providers/common/include/prov/der_wrap.h" => [
            "providers/common/der"
        ],
        "providers/implementations/encode_decode/encode_key2any.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2any.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/kdfs/libdefault-lib-x942kdf.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/kdfs/x942kdf.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/ecdsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/eddsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-ecdsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-eddsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-ml_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-rsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-slh_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/libdefault-lib-sm2_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/ml_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/rsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/slh_dsa_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/implementations/signature/sm2_sig.o" => [
            "providers/common/include/prov"
        ],
        "providers/legacy" => [
            "include",
            "providers/implementations/include",
            "providers/common/include"
        ],
        "providers/libcommon.a" => [
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "providers/libdefault.a" => [
            ".",
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "providers/libfips.a" => [
            ".",
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "providers/liblegacy.a" => [
            ".",
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "providers/libsmtc.a" => [
            ".",
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "providers/libtemplate.a" => [
            "crypto",
            "include",
            "providers/implementations/include",
            "providers/common/include",
            "providers/fips/include"
        ],
        "util/wrap.pl" => [
            "."
        ]
    },
    "ldadd" => {},
    "libraries" => [
        "apps/libapps.a",
        "libcrypto",
        "libssl",
        "providers/libcommon.a",
        "providers/libdefault.a",
        "providers/liblegacy.a",
        "providers/libtemplate.a"
    ],
    "modules" => [
        "engines/afalg",
        "engines/capi",
        "engines/dasync",
        "engines/ecptest",
        "engines/loader_attic",
        "engines/ossltest",
        "engines/padlock",
        "providers/legacy"
    ],
    "programs" => [
        "apps/openssl"
    ],
    "scripts" => [
        "apps/CA.pl",
        "apps/tsget.pl",
        "util/shlib_wrap.sh",
        "util/wrap.pl"
    ],
    "shared_sources" => {
        "libcrypto" => [
            "crypto/aes/libcrypto-shlib-aes-x86_64.o",
            "crypto/aes/libcrypto-shlib-aes_cfb.o",
            "crypto/aes/libcrypto-shlib-aes_ecb.o",
            "crypto/aes/libcrypto-shlib-aes_ige.o",
            "crypto/aes/libcrypto-shlib-aes_misc.o",
            "crypto/aes/libcrypto-shlib-aes_ofb.o",
            "crypto/aes/libcrypto-shlib-aes_wrap.o",
            "crypto/aes/libcrypto-shlib-aesni-mb-x86_64.o",
            "crypto/aes/libcrypto-shlib-aesni-sha1-x86_64.o",
            "crypto/aes/libcrypto-shlib-aesni-sha256-x86_64.o",
            "crypto/aes/libcrypto-shlib-aesni-x86_64.o",
            "crypto/aes/libcrypto-shlib-aesni-xts-avx512.o",
            "crypto/aes/libcrypto-shlib-bsaes-x86_64.o",
            "crypto/aes/libcrypto-shlib-vpaes-x86_64.o",
            "crypto/asn1/libcrypto-shlib-a_bitstr.o",
            "crypto/asn1/libcrypto-shlib-a_d2i_fp.o",
            "crypto/asn1/libcrypto-shlib-a_digest.o",
            "crypto/asn1/libcrypto-shlib-a_dup.o",
            "crypto/asn1/libcrypto-shlib-a_gentm.o",
            "crypto/asn1/libcrypto-shlib-a_i2d_fp.o",
            "crypto/asn1/libcrypto-shlib-a_int.o",
            "crypto/asn1/libcrypto-shlib-a_mbstr.o",
            "crypto/asn1/libcrypto-shlib-a_object.o",
            "crypto/asn1/libcrypto-shlib-a_octet.o",
            "crypto/asn1/libcrypto-shlib-a_print.o",
            "crypto/asn1/libcrypto-shlib-a_sign.o",
            "crypto/asn1/libcrypto-shlib-a_strex.o",
            "crypto/asn1/libcrypto-shlib-a_strnid.o",
            "crypto/asn1/libcrypto-shlib-a_time.o",
            "crypto/asn1/libcrypto-shlib-a_type.o",
            "crypto/asn1/libcrypto-shlib-a_utctm.o",
            "crypto/asn1/libcrypto-shlib-a_utf8.o",
            "crypto/asn1/libcrypto-shlib-a_verify.o",
            "crypto/asn1/libcrypto-shlib-ameth_lib.o",
            "crypto/asn1/libcrypto-shlib-asn1_err.o",
            "crypto/asn1/libcrypto-shlib-asn1_gen.o",
            "crypto/asn1/libcrypto-shlib-asn1_item_list.o",
            "crypto/asn1/libcrypto-shlib-asn1_lib.o",
            "crypto/asn1/libcrypto-shlib-asn1_parse.o",
            "crypto/asn1/libcrypto-shlib-asn_mime.o",
            "crypto/asn1/libcrypto-shlib-asn_moid.o",
            "crypto/asn1/libcrypto-shlib-asn_mstbl.o",
            "crypto/asn1/libcrypto-shlib-asn_pack.o",
            "crypto/asn1/libcrypto-shlib-bio_asn1.o",
            "crypto/asn1/libcrypto-shlib-bio_ndef.o",
            "crypto/asn1/libcrypto-shlib-d2i_param.o",
            "crypto/asn1/libcrypto-shlib-d2i_pr.o",
            "crypto/asn1/libcrypto-shlib-d2i_pu.o",
            "crypto/asn1/libcrypto-shlib-evp_asn1.o",
            "crypto/asn1/libcrypto-shlib-f_int.o",
            "crypto/asn1/libcrypto-shlib-f_string.o",
            "crypto/asn1/libcrypto-shlib-i2d_evp.o",
            "crypto/asn1/libcrypto-shlib-n_pkey.o",
            "crypto/asn1/libcrypto-shlib-nsseq.o",
            "crypto/asn1/libcrypto-shlib-p5_pbe.o",
            "crypto/asn1/libcrypto-shlib-p5_pbev2.o",
            "crypto/asn1/libcrypto-shlib-p5_scrypt.o",
            "crypto/asn1/libcrypto-shlib-p8_pkey.o",
            "crypto/asn1/libcrypto-shlib-t_bitst.o",
            "crypto/asn1/libcrypto-shlib-t_pkey.o",
            "crypto/asn1/libcrypto-shlib-t_spki.o",
            "crypto/asn1/libcrypto-shlib-tasn_dec.o",
            "crypto/asn1/libcrypto-shlib-tasn_enc.o",
            "crypto/asn1/libcrypto-shlib-tasn_fre.o",
            "crypto/asn1/libcrypto-shlib-tasn_new.o",
            "crypto/asn1/libcrypto-shlib-tasn_prn.o",
            "crypto/asn1/libcrypto-shlib-tasn_scn.o",
            "crypto/asn1/libcrypto-shlib-tasn_typ.o",
            "crypto/asn1/libcrypto-shlib-tasn_utl.o",
            "crypto/asn1/libcrypto-shlib-x_algor.o",
            "crypto/asn1/libcrypto-shlib-x_bignum.o",
            "crypto/asn1/libcrypto-shlib-x_info.o",
            "crypto/asn1/libcrypto-shlib-x_int64.o",
            "crypto/asn1/libcrypto-shlib-x_long.o",
            "crypto/asn1/libcrypto-shlib-x_pkey.o",
            "crypto/asn1/libcrypto-shlib-x_sig.o",
            "crypto/asn1/libcrypto-shlib-x_spki.o",
            "crypto/asn1/libcrypto-shlib-x_val.o",
            "crypto/async/arch/libcrypto-shlib-async_null.o",
            "crypto/async/arch/libcrypto-shlib-async_posix.o",
            "crypto/async/arch/libcrypto-shlib-async_win.o",
            "crypto/async/libcrypto-shlib-async.o",
            "crypto/async/libcrypto-shlib-async_err.o",
            "crypto/async/libcrypto-shlib-async_wait.o",
            "crypto/bio/libcrypto-shlib-bf_buff.o",
            "crypto/bio/libcrypto-shlib-bf_lbuf.o",
            "crypto/bio/libcrypto-shlib-bf_nbio.o",
            "crypto/bio/libcrypto-shlib-bf_null.o",
            "crypto/bio/libcrypto-shlib-bf_prefix.o",
            "crypto/bio/libcrypto-shlib-bf_readbuff.o",
            "crypto/bio/libcrypto-shlib-bio_addr.o",
            "crypto/bio/libcrypto-shlib-bio_cb.o",
            "crypto/bio/libcrypto-shlib-bio_dump.o",
            "crypto/bio/libcrypto-shlib-bio_err.o",
            "crypto/bio/libcrypto-shlib-bio_lib.o",
            "crypto/bio/libcrypto-shlib-bio_meth.o",
            "crypto/bio/libcrypto-shlib-bio_print.o",
            "crypto/bio/libcrypto-shlib-bio_sock.o",
            "crypto/bio/libcrypto-shlib-bio_sock2.o",
            "crypto/bio/libcrypto-shlib-bss_acpt.o",
            "crypto/bio/libcrypto-shlib-bss_bio.o",
            "crypto/bio/libcrypto-shlib-bss_conn.o",
            "crypto/bio/libcrypto-shlib-bss_core.o",
            "crypto/bio/libcrypto-shlib-bss_dgram.o",
            "crypto/bio/libcrypto-shlib-bss_dgram_pair.o",
            "crypto/bio/libcrypto-shlib-bss_fd.o",
            "crypto/bio/libcrypto-shlib-bss_file.o",
            "crypto/bio/libcrypto-shlib-bss_log.o",
            "crypto/bio/libcrypto-shlib-bss_mem.o",
            "crypto/bio/libcrypto-shlib-bss_null.o",
            "crypto/bio/libcrypto-shlib-bss_sock.o",
            "crypto/bio/libcrypto-shlib-ossl_core_bio.o",
            "crypto/bn/asm/libcrypto-shlib-x86_64-gcc.o",
            "crypto/bn/libcrypto-shlib-bn_add.o",
            "crypto/bn/libcrypto-shlib-bn_blind.o",
            "crypto/bn/libcrypto-shlib-bn_const.o",
            "crypto/bn/libcrypto-shlib-bn_conv.o",
            "crypto/bn/libcrypto-shlib-bn_ctx.o",
            "crypto/bn/libcrypto-shlib-bn_depr.o",
            "crypto/bn/libcrypto-shlib-bn_dh.o",
            "crypto/bn/libcrypto-shlib-bn_div.o",
            "crypto/bn/libcrypto-shlib-bn_err.o",
            "crypto/bn/libcrypto-shlib-bn_exp.o",
            "crypto/bn/libcrypto-shlib-bn_exp2.o",
            "crypto/bn/libcrypto-shlib-bn_gcd.o",
            "crypto/bn/libcrypto-shlib-bn_gf2m.o",
            "crypto/bn/libcrypto-shlib-bn_intern.o",
            "crypto/bn/libcrypto-shlib-bn_kron.o",
            "crypto/bn/libcrypto-shlib-bn_lib.o",
            "crypto/bn/libcrypto-shlib-bn_mod.o",
            "crypto/bn/libcrypto-shlib-bn_mont.o",
            "crypto/bn/libcrypto-shlib-bn_mpi.o",
            "crypto/bn/libcrypto-shlib-bn_mul.o",
            "crypto/bn/libcrypto-shlib-bn_nist.o",
            "crypto/bn/libcrypto-shlib-bn_prime.o",
            "crypto/bn/libcrypto-shlib-bn_print.o",
            "crypto/bn/libcrypto-shlib-bn_rand.o",
            "crypto/bn/libcrypto-shlib-bn_recp.o",
            "crypto/bn/libcrypto-shlib-bn_rsa_fips186_4.o",
            "crypto/bn/libcrypto-shlib-bn_shift.o",
            "crypto/bn/libcrypto-shlib-bn_sm2.o",
            "crypto/bn/libcrypto-shlib-bn_sqr.o",
            "crypto/bn/libcrypto-shlib-bn_sqrt.o",
            "crypto/bn/libcrypto-shlib-bn_srp.o",
            "crypto/bn/libcrypto-shlib-bn_word.o",
            "crypto/bn/libcrypto-shlib-bn_x931p.o",
            "crypto/bn/libcrypto-shlib-rsaz-2k-avx512.o",
            "crypto/bn/libcrypto-shlib-rsaz-2k-avxifma.o",
            "crypto/bn/libcrypto-shlib-rsaz-3k-avx512.o",
            "crypto/bn/libcrypto-shlib-rsaz-3k-avxifma.o",
            "crypto/bn/libcrypto-shlib-rsaz-4k-avx512.o",
            "crypto/bn/libcrypto-shlib-rsaz-4k-avxifma.o",
            "crypto/bn/libcrypto-shlib-rsaz-avx2.o",
            "crypto/bn/libcrypto-shlib-rsaz-x86_64.o",
            "crypto/bn/libcrypto-shlib-rsaz_exp.o",
            "crypto/bn/libcrypto-shlib-rsaz_exp_x2.o",
            "crypto/bn/libcrypto-shlib-x86_64-gf2m.o",
            "crypto/bn/libcrypto-shlib-x86_64-mont.o",
            "crypto/bn/libcrypto-shlib-x86_64-mont5.o",
            "crypto/buffer/libcrypto-shlib-buf_err.o",
            "crypto/buffer/libcrypto-shlib-buffer.o",
            "crypto/chacha/libcrypto-shlib-chacha-x86_64.o",
            "crypto/cmac/libcrypto-shlib-cmac.o",
            "crypto/cmp/libcrypto-shlib-cmp_asn.o",
            "crypto/cmp/libcrypto-shlib-cmp_client.o",
            "crypto/cmp/libcrypto-shlib-cmp_ctx.o",
            "crypto/cmp/libcrypto-shlib-cmp_err.o",
            "crypto/cmp/libcrypto-shlib-cmp_genm.o",
            "crypto/cmp/libcrypto-shlib-cmp_hdr.o",
            "crypto/cmp/libcrypto-shlib-cmp_http.o",
            "crypto/cmp/libcrypto-shlib-cmp_msg.o",
            "crypto/cmp/libcrypto-shlib-cmp_protect.o",
            "crypto/cmp/libcrypto-shlib-cmp_server.o",
            "crypto/cmp/libcrypto-shlib-cmp_status.o",
            "crypto/cmp/libcrypto-shlib-cmp_util.o",
            "crypto/cmp/libcrypto-shlib-cmp_vfy.o",
            "crypto/cms/libcrypto-shlib-cms_asn1.o",
            "crypto/cms/libcrypto-shlib-cms_att.o",
            "crypto/cms/libcrypto-shlib-cms_cd.o",
            "crypto/cms/libcrypto-shlib-cms_dd.o",
            "crypto/cms/libcrypto-shlib-cms_dh.o",
            "crypto/cms/libcrypto-shlib-cms_ec.o",
            "crypto/cms/libcrypto-shlib-cms_enc.o",
            "crypto/cms/libcrypto-shlib-cms_env.o",
            "crypto/cms/libcrypto-shlib-cms_err.o",
            "crypto/cms/libcrypto-shlib-cms_ess.o",
            "crypto/cms/libcrypto-shlib-cms_io.o",
            "crypto/cms/libcrypto-shlib-cms_kari.o",
            "crypto/cms/libcrypto-shlib-cms_lib.o",
            "crypto/cms/libcrypto-shlib-cms_pwri.o",
            "crypto/cms/libcrypto-shlib-cms_rsa.o",
            "crypto/cms/libcrypto-shlib-cms_sd.o",
            "crypto/cms/libcrypto-shlib-cms_smime.o",
            "crypto/comp/libcrypto-shlib-c_brotli.o",
            "crypto/comp/libcrypto-shlib-c_zlib.o",
            "crypto/comp/libcrypto-shlib-c_zstd.o",
            "crypto/comp/libcrypto-shlib-comp_err.o",
            "crypto/comp/libcrypto-shlib-comp_lib.o",
            "crypto/conf/libcrypto-shlib-conf_api.o",
            "crypto/conf/libcrypto-shlib-conf_def.o",
            "crypto/conf/libcrypto-shlib-conf_err.o",
            "crypto/conf/libcrypto-shlib-conf_lib.o",
            "crypto/conf/libcrypto-shlib-conf_mall.o",
            "crypto/conf/libcrypto-shlib-conf_mod.o",
            "crypto/conf/libcrypto-shlib-conf_sap.o",
            "crypto/conf/libcrypto-shlib-conf_ssl.o",
            "crypto/crmf/libcrypto-shlib-crmf_asn.o",
            "crypto/crmf/libcrypto-shlib-crmf_err.o",
            "crypto/crmf/libcrypto-shlib-crmf_lib.o",
            "crypto/crmf/libcrypto-shlib-crmf_pbm.o",
            "crypto/ct/libcrypto-shlib-ct_b64.o",
            "crypto/ct/libcrypto-shlib-ct_err.o",
            "crypto/ct/libcrypto-shlib-ct_log.o",
            "crypto/ct/libcrypto-shlib-ct_oct.o",
            "crypto/ct/libcrypto-shlib-ct_policy.o",
            "crypto/ct/libcrypto-shlib-ct_prn.o",
            "crypto/ct/libcrypto-shlib-ct_sct.o",
            "crypto/ct/libcrypto-shlib-ct_sct_ctx.o",
            "crypto/ct/libcrypto-shlib-ct_vfy.o",
            "crypto/ct/libcrypto-shlib-ct_x509v3.o",
            "crypto/des/libcrypto-shlib-cbc_cksm.o",
            "crypto/des/libcrypto-shlib-cbc_enc.o",
            "crypto/des/libcrypto-shlib-cfb64ede.o",
            "crypto/des/libcrypto-shlib-cfb64enc.o",
            "crypto/des/libcrypto-shlib-cfb_enc.o",
            "crypto/des/libcrypto-shlib-des_enc.o",
            "crypto/des/libcrypto-shlib-ecb3_enc.o",
            "crypto/des/libcrypto-shlib-ecb_enc.o",
            "crypto/des/libcrypto-shlib-fcrypt.o",
            "crypto/des/libcrypto-shlib-fcrypt_b.o",
            "crypto/des/libcrypto-shlib-ofb64ede.o",
            "crypto/des/libcrypto-shlib-ofb64enc.o",
            "crypto/des/libcrypto-shlib-ofb_enc.o",
            "crypto/des/libcrypto-shlib-pcbc_enc.o",
            "crypto/des/libcrypto-shlib-qud_cksm.o",
            "crypto/des/libcrypto-shlib-rand_key.o",
            "crypto/des/libcrypto-shlib-set_key.o",
            "crypto/des/libcrypto-shlib-str2key.o",
            "crypto/des/libcrypto-shlib-xcbc_enc.o",
            "crypto/dh/libcrypto-shlib-dh_ameth.o",
            "crypto/dh/libcrypto-shlib-dh_asn1.o",
            "crypto/dh/libcrypto-shlib-dh_backend.o",
            "crypto/dh/libcrypto-shlib-dh_check.o",
            "crypto/dh/libcrypto-shlib-dh_depr.o",
            "crypto/dh/libcrypto-shlib-dh_err.o",
            "crypto/dh/libcrypto-shlib-dh_gen.o",
            "crypto/dh/libcrypto-shlib-dh_group_params.o",
            "crypto/dh/libcrypto-shlib-dh_kdf.o",
            "crypto/dh/libcrypto-shlib-dh_key.o",
            "crypto/dh/libcrypto-shlib-dh_lib.o",
            "crypto/dh/libcrypto-shlib-dh_meth.o",
            "crypto/dh/libcrypto-shlib-dh_pmeth.o",
            "crypto/dh/libcrypto-shlib-dh_prn.o",
            "crypto/dh/libcrypto-shlib-dh_rfc5114.o",
            "crypto/dsa/libcrypto-shlib-dsa_ameth.o",
            "crypto/dsa/libcrypto-shlib-dsa_asn1.o",
            "crypto/dsa/libcrypto-shlib-dsa_backend.o",
            "crypto/dsa/libcrypto-shlib-dsa_check.o",
            "crypto/dsa/libcrypto-shlib-dsa_depr.o",
            "crypto/dsa/libcrypto-shlib-dsa_err.o",
            "crypto/dsa/libcrypto-shlib-dsa_gen.o",
            "crypto/dsa/libcrypto-shlib-dsa_key.o",
            "crypto/dsa/libcrypto-shlib-dsa_lib.o",
            "crypto/dsa/libcrypto-shlib-dsa_meth.o",
            "crypto/dsa/libcrypto-shlib-dsa_ossl.o",
            "crypto/dsa/libcrypto-shlib-dsa_pmeth.o",
            "crypto/dsa/libcrypto-shlib-dsa_prn.o",
            "crypto/dsa/libcrypto-shlib-dsa_sign.o",
            "crypto/dsa/libcrypto-shlib-dsa_vrf.o",
            "crypto/dso/libcrypto-shlib-dso_dl.o",
            "crypto/dso/libcrypto-shlib-dso_dlfcn.o",
            "crypto/dso/libcrypto-shlib-dso_err.o",
            "crypto/dso/libcrypto-shlib-dso_lib.o",
            "crypto/dso/libcrypto-shlib-dso_openssl.o",
            "crypto/dso/libcrypto-shlib-dso_win32.o",
            "crypto/ec/curve448/arch_32/libcrypto-shlib-f_impl32.o",
            "crypto/ec/curve448/arch_64/libcrypto-shlib-f_impl64.o",
            "crypto/ec/curve448/libcrypto-shlib-curve448.o",
            "crypto/ec/curve448/libcrypto-shlib-curve448_tables.o",
            "crypto/ec/curve448/libcrypto-shlib-eddsa.o",
            "crypto/ec/curve448/libcrypto-shlib-f_generic.o",
            "crypto/ec/curve448/libcrypto-shlib-scalar.o",
            "crypto/ec/libcrypto-shlib-curve25519.o",
            "crypto/ec/libcrypto-shlib-ec2_oct.o",
            "crypto/ec/libcrypto-shlib-ec2_smpl.o",
            "crypto/ec/libcrypto-shlib-ec_ameth.o",
            "crypto/ec/libcrypto-shlib-ec_asn1.o",
            "crypto/ec/libcrypto-shlib-ec_backend.o",
            "crypto/ec/libcrypto-shlib-ec_check.o",
            "crypto/ec/libcrypto-shlib-ec_curve.o",
            "crypto/ec/libcrypto-shlib-ec_cvt.o",
            "crypto/ec/libcrypto-shlib-ec_deprecated.o",
            "crypto/ec/libcrypto-shlib-ec_err.o",
            "crypto/ec/libcrypto-shlib-ec_key.o",
            "crypto/ec/libcrypto-shlib-ec_kmeth.o",
            "crypto/ec/libcrypto-shlib-ec_lib.o",
            "crypto/ec/libcrypto-shlib-ec_mult.o",
            "crypto/ec/libcrypto-shlib-ec_oct.o",
            "crypto/ec/libcrypto-shlib-ec_pmeth.o",
            "crypto/ec/libcrypto-shlib-ec_print.o",
            "crypto/ec/libcrypto-shlib-ecdh_kdf.o",
            "crypto/ec/libcrypto-shlib-ecdh_ossl.o",
            "crypto/ec/libcrypto-shlib-ecdsa_ossl.o",
            "crypto/ec/libcrypto-shlib-ecdsa_sign.o",
            "crypto/ec/libcrypto-shlib-ecdsa_vrf.o",
            "crypto/ec/libcrypto-shlib-eck_prn.o",
            "crypto/ec/libcrypto-shlib-ecp_meth.o",
            "crypto/ec/libcrypto-shlib-ecp_mont.o",
            "crypto/ec/libcrypto-shlib-ecp_nist.o",
            "crypto/ec/libcrypto-shlib-ecp_nistz256-x86_64.o",
            "crypto/ec/libcrypto-shlib-ecp_nistz256.o",
            "crypto/ec/libcrypto-shlib-ecp_oct.o",
            "crypto/ec/libcrypto-shlib-ecp_smpl.o",
            "crypto/ec/libcrypto-shlib-ecx_backend.o",
            "crypto/ec/libcrypto-shlib-ecx_key.o",
            "crypto/ec/libcrypto-shlib-ecx_meth.o",
            "crypto/ec/libcrypto-shlib-x25519-x86_64.o",
            "crypto/eia3/libcrypto-shlib-eia3.o",
            "crypto/encode_decode/libcrypto-shlib-decoder_err.o",
            "crypto/encode_decode/libcrypto-shlib-decoder_lib.o",
            "crypto/encode_decode/libcrypto-shlib-decoder_meth.o",
            "crypto/encode_decode/libcrypto-shlib-decoder_pkey.o",
            "crypto/encode_decode/libcrypto-shlib-encoder_err.o",
            "crypto/encode_decode/libcrypto-shlib-encoder_lib.o",
            "crypto/encode_decode/libcrypto-shlib-encoder_meth.o",
            "crypto/encode_decode/libcrypto-shlib-encoder_pkey.o",
            "crypto/engine/libcrypto-shlib-eng_all.o",
            "crypto/engine/libcrypto-shlib-eng_cnf.o",
            "crypto/engine/libcrypto-shlib-eng_ctrl.o",
            "crypto/engine/libcrypto-shlib-eng_dyn.o",
            "crypto/engine/libcrypto-shlib-eng_err.o",
            "crypto/engine/libcrypto-shlib-eng_fat.o",
            "crypto/engine/libcrypto-shlib-eng_init.o",
            "crypto/engine/libcrypto-shlib-eng_lib.o",
            "crypto/engine/libcrypto-shlib-eng_list.o",
            "crypto/engine/libcrypto-shlib-eng_openssl.o",
            "crypto/engine/libcrypto-shlib-eng_pkey.o",
            "crypto/engine/libcrypto-shlib-eng_rdrand.o",
            "crypto/engine/libcrypto-shlib-eng_table.o",
            "crypto/engine/libcrypto-shlib-tb_asnmth.o",
            "crypto/engine/libcrypto-shlib-tb_cipher.o",
            "crypto/engine/libcrypto-shlib-tb_dh.o",
            "crypto/engine/libcrypto-shlib-tb_digest.o",
            "crypto/engine/libcrypto-shlib-tb_dsa.o",
            "crypto/engine/libcrypto-shlib-tb_eckey.o",
            "crypto/engine/libcrypto-shlib-tb_ecpmeth.o",
            "crypto/engine/libcrypto-shlib-tb_pkmeth.o",
            "crypto/engine/libcrypto-shlib-tb_rand.o",
            "crypto/engine/libcrypto-shlib-tb_rsa.o",
            "crypto/err/libcrypto-shlib-err.o",
            "crypto/err/libcrypto-shlib-err_all.o",
            "crypto/err/libcrypto-shlib-err_all_legacy.o",
            "crypto/err/libcrypto-shlib-err_blocks.o",
            "crypto/err/libcrypto-shlib-err_mark.o",
            "crypto/err/libcrypto-shlib-err_prn.o",
            "crypto/err/libcrypto-shlib-err_save.o",
            "crypto/ess/libcrypto-shlib-ess_asn1.o",
            "crypto/ess/libcrypto-shlib-ess_err.o",
            "crypto/ess/libcrypto-shlib-ess_lib.o",
            "crypto/evp/libcrypto-shlib-asymcipher.o",
            "crypto/evp/libcrypto-shlib-bio_b64.o",
            "crypto/evp/libcrypto-shlib-bio_enc.o",
            "crypto/evp/libcrypto-shlib-bio_md.o",
            "crypto/evp/libcrypto-shlib-bio_ok.o",
            "crypto/evp/libcrypto-shlib-c_allc.o",
            "crypto/evp/libcrypto-shlib-c_alld.o",
            "crypto/evp/libcrypto-shlib-cmeth_lib.o",
            "crypto/evp/libcrypto-shlib-ctrl_params_translate.o",
            "crypto/evp/libcrypto-shlib-dh_ctrl.o",
            "crypto/evp/libcrypto-shlib-dh_support.o",
            "crypto/evp/libcrypto-shlib-digest.o",
            "crypto/evp/libcrypto-shlib-dsa_ctrl.o",
            "crypto/evp/libcrypto-shlib-e_aes.o",
            "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha1.o",
            "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha256.o",
            "crypto/evp/libcrypto-shlib-e_chacha20_poly1305.o",
            "crypto/evp/libcrypto-shlib-e_des.o",
            "crypto/evp/libcrypto-shlib-e_des3.o",
            "crypto/evp/libcrypto-shlib-e_eea3.o",
            "crypto/evp/libcrypto-shlib-e_null.o",
            "crypto/evp/libcrypto-shlib-e_old.o",
            "crypto/evp/libcrypto-shlib-e_rc4.o",
            "crypto/evp/libcrypto-shlib-e_rc4_hmac_md5.o",
            "crypto/evp/libcrypto-shlib-e_rc5.o",
            "crypto/evp/libcrypto-shlib-e_sm4.o",
            "crypto/evp/libcrypto-shlib-e_wbsm4_baiwu.o",
            "crypto/evp/libcrypto-shlib-e_wbsm4_wsise.o",
            "crypto/evp/libcrypto-shlib-e_wbsm4_xiaolai.o",
            "crypto/evp/libcrypto-shlib-e_xcbc_d.o",
            "crypto/evp/libcrypto-shlib-ec_ctrl.o",
            "crypto/evp/libcrypto-shlib-ec_support.o",
            "crypto/evp/libcrypto-shlib-encode.o",
            "crypto/evp/libcrypto-shlib-evp_cnf.o",
            "crypto/evp/libcrypto-shlib-evp_enc.o",
            "crypto/evp/libcrypto-shlib-evp_err.o",
            "crypto/evp/libcrypto-shlib-evp_fetch.o",
            "crypto/evp/libcrypto-shlib-evp_key.o",
            "crypto/evp/libcrypto-shlib-evp_lib.o",
            "crypto/evp/libcrypto-shlib-evp_pbe.o",
            "crypto/evp/libcrypto-shlib-evp_pkey.o",
            "crypto/evp/libcrypto-shlib-evp_rand.o",
            "crypto/evp/libcrypto-shlib-evp_utils.o",
            "crypto/evp/libcrypto-shlib-exchange.o",
            "crypto/evp/libcrypto-shlib-kdf_lib.o",
            "crypto/evp/libcrypto-shlib-kdf_meth.o",
            "crypto/evp/libcrypto-shlib-kem.o",
            "crypto/evp/libcrypto-shlib-keymgmt_lib.o",
            "crypto/evp/libcrypto-shlib-keymgmt_meth.o",
            "crypto/evp/libcrypto-shlib-legacy_md5.o",
            "crypto/evp/libcrypto-shlib-legacy_md5_sha1.o",
            "crypto/evp/libcrypto-shlib-legacy_sha.o",
            "crypto/evp/libcrypto-shlib-m_null.o",
            "crypto/evp/libcrypto-shlib-m_sigver.o",
            "crypto/evp/libcrypto-shlib-mac_lib.o",
            "crypto/evp/libcrypto-shlib-mac_meth.o",
            "crypto/evp/libcrypto-shlib-names.o",
            "crypto/evp/libcrypto-shlib-p5_crpt.o",
            "crypto/evp/libcrypto-shlib-p5_crpt2.o",
            "crypto/evp/libcrypto-shlib-p_dec.o",
            "crypto/evp/libcrypto-shlib-p_enc.o",
            "crypto/evp/libcrypto-shlib-p_legacy.o",
            "crypto/evp/libcrypto-shlib-p_lib.o",
            "crypto/evp/libcrypto-shlib-p_open.o",
            "crypto/evp/libcrypto-shlib-p_seal.o",
            "crypto/evp/libcrypto-shlib-p_sign.o",
            "crypto/evp/libcrypto-shlib-p_verify.o",
            "crypto/evp/libcrypto-shlib-pbe_scrypt.o",
            "crypto/evp/libcrypto-shlib-pmeth_check.o",
            "crypto/evp/libcrypto-shlib-pmeth_gn.o",
            "crypto/evp/libcrypto-shlib-pmeth_lib.o",
            "crypto/evp/libcrypto-shlib-s_lib.o",
            "crypto/evp/libcrypto-shlib-signature.o",
            "crypto/evp/libcrypto-shlib-skeymgmt_meth.o",
            "crypto/ffc/libcrypto-shlib-ffc_backend.o",
            "crypto/ffc/libcrypto-shlib-ffc_dh.o",
            "crypto/ffc/libcrypto-shlib-ffc_key_generate.o",
            "crypto/ffc/libcrypto-shlib-ffc_key_validate.o",
            "crypto/ffc/libcrypto-shlib-ffc_params.o",
            "crypto/ffc/libcrypto-shlib-ffc_params_generate.o",
            "crypto/ffc/libcrypto-shlib-ffc_params_validate.o",
            "crypto/hashtable/libcrypto-shlib-hashfunc.o",
            "crypto/hashtable/libcrypto-shlib-hashtable.o",
            "crypto/hmac/libcrypto-shlib-hmac.o",
            "crypto/hpke/libcrypto-shlib-hpke.o",
            "crypto/hpke/libcrypto-shlib-hpke_util.o",
            "crypto/http/libcrypto-shlib-http_client.o",
            "crypto/http/libcrypto-shlib-http_err.o",
            "crypto/http/libcrypto-shlib-http_lib.o",
            "crypto/kdf/libcrypto-shlib-kdf_err.o",
            "crypto/lhash/libcrypto-shlib-lh_stats.o",
            "crypto/lhash/libcrypto-shlib-lhash.o",
            "crypto/libcrypto-shlib-asn1_dsa.o",
            "crypto/libcrypto-shlib-bsearch.o",
            "crypto/libcrypto-shlib-comp_methods.o",
            "crypto/libcrypto-shlib-context.o",
            "crypto/libcrypto-shlib-core_algorithm.o",
            "crypto/libcrypto-shlib-core_fetch.o",
            "crypto/libcrypto-shlib-core_namemap.o",
            "crypto/libcrypto-shlib-cpt_err.o",
            "crypto/libcrypto-shlib-cpuid.o",
            "crypto/libcrypto-shlib-cryptlib.o",
            "crypto/libcrypto-shlib-ctype.o",
            "crypto/libcrypto-shlib-cversion.o",
            "crypto/libcrypto-shlib-defaults.o",
            "crypto/libcrypto-shlib-der_writer.o",
            "crypto/libcrypto-shlib-deterministic_nonce.o",
            "crypto/libcrypto-shlib-ebcdic.o",
            "crypto/libcrypto-shlib-ex_data.o",
            "crypto/libcrypto-shlib-getenv.o",
            "crypto/libcrypto-shlib-indicator_core.o",
            "crypto/libcrypto-shlib-info.o",
            "crypto/libcrypto-shlib-init.o",
            "crypto/libcrypto-shlib-initthread.o",
            "crypto/libcrypto-shlib-mem.o",
            "crypto/libcrypto-shlib-mem_sec.o",
            "crypto/libcrypto-shlib-o_dir.o",
            "crypto/libcrypto-shlib-o_fopen.o",
            "crypto/libcrypto-shlib-o_init.o",
            "crypto/libcrypto-shlib-o_str.o",
            "crypto/libcrypto-shlib-o_syslog.o",
            "crypto/libcrypto-shlib-o_time.o",
            "crypto/libcrypto-shlib-packet.o",
            "crypto/libcrypto-shlib-param_build.o",
            "crypto/libcrypto-shlib-param_build_set.o",
            "crypto/libcrypto-shlib-params.o",
            "crypto/libcrypto-shlib-params_dup.o",
            "crypto/libcrypto-shlib-params_from_text.o",
            "crypto/libcrypto-shlib-params_idx.o",
            "crypto/libcrypto-shlib-passphrase.o",
            "crypto/libcrypto-shlib-provider.o",
            "crypto/libcrypto-shlib-provider_child.o",
            "crypto/libcrypto-shlib-provider_conf.o",
            "crypto/libcrypto-shlib-provider_core.o",
            "crypto/libcrypto-shlib-provider_predefined.o",
            "crypto/libcrypto-shlib-punycode.o",
            "crypto/libcrypto-shlib-quic_vlint.o",
            "crypto/libcrypto-shlib-self_test_core.o",
            "crypto/libcrypto-shlib-sleep.o",
            "crypto/libcrypto-shlib-sparse_array.o",
            "crypto/libcrypto-shlib-ssl_err.o",
            "crypto/libcrypto-shlib-threads_lib.o",
            "crypto/libcrypto-shlib-threads_none.o",
            "crypto/libcrypto-shlib-threads_pthread.o",
            "crypto/libcrypto-shlib-threads_win.o",
            "crypto/libcrypto-shlib-time.o",
            "crypto/libcrypto-shlib-trace.o",
            "crypto/libcrypto-shlib-uid.o",
            "crypto/libcrypto-shlib-x86_64cpuid.o",
            "crypto/md5/libcrypto-shlib-md5-x86_64.o",
            "crypto/md5/libcrypto-shlib-md5_dgst.o",
            "crypto/md5/libcrypto-shlib-md5_one.o",
            "crypto/md5/libcrypto-shlib-md5_sha1.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_encoders.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key_compress.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_matrix.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_ntt.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_params.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sample.o",
            "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sign.o",
            "crypto/ml_kem/libcrypto-shlib-ml_kem.o",
            "crypto/modes/libcrypto-shlib-aes-gcm-avx512.o",
            "crypto/modes/libcrypto-shlib-aesni-gcm-x86_64.o",
            "crypto/modes/libcrypto-shlib-cbc128.o",
            "crypto/modes/libcrypto-shlib-ccm128.o",
            "crypto/modes/libcrypto-shlib-cfb128.o",
            "crypto/modes/libcrypto-shlib-ctr128.o",
            "crypto/modes/libcrypto-shlib-cts128.o",
            "crypto/modes/libcrypto-shlib-gcm128.o",
            "crypto/modes/libcrypto-shlib-ghash-x86_64.o",
            "crypto/modes/libcrypto-shlib-ocb128.o",
            "crypto/modes/libcrypto-shlib-ofb128.o",
            "crypto/modes/libcrypto-shlib-siv128.o",
            "crypto/modes/libcrypto-shlib-wrap128.o",
            "crypto/modes/libcrypto-shlib-xts128.o",
            "crypto/modes/libcrypto-shlib-xts128gb.o",
            "crypto/objects/libcrypto-shlib-o_names.o",
            "crypto/objects/libcrypto-shlib-obj_dat.o",
            "crypto/objects/libcrypto-shlib-obj_err.o",
            "crypto/objects/libcrypto-shlib-obj_lib.o",
            "crypto/objects/libcrypto-shlib-obj_xref.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_asn.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_cl.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_err.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_ext.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_http.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_lib.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_prn.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_srv.o",
            "crypto/ocsp/libcrypto-shlib-ocsp_vfy.o",
            "crypto/ocsp/libcrypto-shlib-v3_ocsp.o",
            "crypto/pem/libcrypto-shlib-pem_all.o",
            "crypto/pem/libcrypto-shlib-pem_err.o",
            "crypto/pem/libcrypto-shlib-pem_info.o",
            "crypto/pem/libcrypto-shlib-pem_lib.o",
            "crypto/pem/libcrypto-shlib-pem_oth.o",
            "crypto/pem/libcrypto-shlib-pem_pk8.o",
            "crypto/pem/libcrypto-shlib-pem_pkey.o",
            "crypto/pem/libcrypto-shlib-pem_sign.o",
            "crypto/pem/libcrypto-shlib-pem_x509.o",
            "crypto/pem/libcrypto-shlib-pem_xaux.o",
            "crypto/pem/libcrypto-shlib-pvkfmt.o",
            "crypto/pkcs12/libcrypto-shlib-p12_add.o",
            "crypto/pkcs12/libcrypto-shlib-p12_asn.o",
            "crypto/pkcs12/libcrypto-shlib-p12_attr.o",
            "crypto/pkcs12/libcrypto-shlib-p12_crpt.o",
            "crypto/pkcs12/libcrypto-shlib-p12_crt.o",
            "crypto/pkcs12/libcrypto-shlib-p12_decr.o",
            "crypto/pkcs12/libcrypto-shlib-p12_init.o",
            "crypto/pkcs12/libcrypto-shlib-p12_key.o",
            "crypto/pkcs12/libcrypto-shlib-p12_kiss.o",
            "crypto/pkcs12/libcrypto-shlib-p12_mutl.o",
            "crypto/pkcs12/libcrypto-shlib-p12_npas.o",
            "crypto/pkcs12/libcrypto-shlib-p12_p8d.o",
            "crypto/pkcs12/libcrypto-shlib-p12_p8e.o",
            "crypto/pkcs12/libcrypto-shlib-p12_sbag.o",
            "crypto/pkcs12/libcrypto-shlib-p12_utl.o",
            "crypto/pkcs12/libcrypto-shlib-pk12err.o",
            "crypto/pkcs7/libcrypto-shlib-bio_pk7.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_asn1.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_attr.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_doit.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_lib.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_mime.o",
            "crypto/pkcs7/libcrypto-shlib-pk7_smime.o",
            "crypto/pkcs7/libcrypto-shlib-pkcs7err.o",
            "crypto/poly1305/libcrypto-shlib-poly1305-x86_64.o",
            "crypto/poly1305/libcrypto-shlib-poly1305.o",
            "crypto/property/libcrypto-shlib-defn_cache.o",
            "crypto/property/libcrypto-shlib-property.o",
            "crypto/property/libcrypto-shlib-property_err.o",
            "crypto/property/libcrypto-shlib-property_parse.o",
            "crypto/property/libcrypto-shlib-property_query.o",
            "crypto/property/libcrypto-shlib-property_string.o",
            "crypto/rand/libcrypto-shlib-prov_seed.o",
            "crypto/rand/libcrypto-shlib-rand_deprecated.o",
            "crypto/rand/libcrypto-shlib-rand_err.o",
            "crypto/rand/libcrypto-shlib-rand_lib.o",
            "crypto/rand/libcrypto-shlib-rand_meth.o",
            "crypto/rand/libcrypto-shlib-rand_pool.o",
            "crypto/rand/libcrypto-shlib-rand_uniform.o",
            "crypto/rand/libcrypto-shlib-randfile.o",
            "crypto/rc4/libcrypto-shlib-rc4-md5-x86_64.o",
            "crypto/rc4/libcrypto-shlib-rc4-x86_64.o",
            "crypto/rsa/libcrypto-shlib-rsa_ameth.o",
            "crypto/rsa/libcrypto-shlib-rsa_asn1.o",
            "crypto/rsa/libcrypto-shlib-rsa_backend.o",
            "crypto/rsa/libcrypto-shlib-rsa_chk.o",
            "crypto/rsa/libcrypto-shlib-rsa_crpt.o",
            "crypto/rsa/libcrypto-shlib-rsa_depr.o",
            "crypto/rsa/libcrypto-shlib-rsa_err.o",
            "crypto/rsa/libcrypto-shlib-rsa_gen.o",
            "crypto/rsa/libcrypto-shlib-rsa_lib.o",
            "crypto/rsa/libcrypto-shlib-rsa_meth.o",
            "crypto/rsa/libcrypto-shlib-rsa_mp.o",
            "crypto/rsa/libcrypto-shlib-rsa_mp_names.o",
            "crypto/rsa/libcrypto-shlib-rsa_none.o",
            "crypto/rsa/libcrypto-shlib-rsa_oaep.o",
            "crypto/rsa/libcrypto-shlib-rsa_ossl.o",
            "crypto/rsa/libcrypto-shlib-rsa_pk1.o",
            "crypto/rsa/libcrypto-shlib-rsa_pmeth.o",
            "crypto/rsa/libcrypto-shlib-rsa_prn.o",
            "crypto/rsa/libcrypto-shlib-rsa_pss.o",
            "crypto/rsa/libcrypto-shlib-rsa_saos.o",
            "crypto/rsa/libcrypto-shlib-rsa_schemes.o",
            "crypto/rsa/libcrypto-shlib-rsa_sign.o",
            "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_check.o",
            "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_gen.o",
            "crypto/rsa/libcrypto-shlib-rsa_x931.o",
            "crypto/rsa/libcrypto-shlib-rsa_x931g.o",
            "crypto/sdf/libcrypto-shlib-sdf_lib.o",
            "crypto/sdf/libcrypto-shlib-sdf_meth.o",
            "crypto/sha/libcrypto-shlib-keccak1600-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha1-mb-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha1-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha1_one.o",
            "crypto/sha/libcrypto-shlib-sha1dgst.o",
            "crypto/sha/libcrypto-shlib-sha256-mb-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha256-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha256.o",
            "crypto/sha/libcrypto-shlib-sha3.o",
            "crypto/sha/libcrypto-shlib-sha512-x86_64.o",
            "crypto/sha/libcrypto-shlib-sha512.o",
            "crypto/siphash/libcrypto-shlib-siphash.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_adrs.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_dsa.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_dsa_hash_ctx.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_dsa_key.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_fors.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_hash.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_hypertree.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_params.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_wots.o",
            "crypto/slh_dsa/libcrypto-shlib-slh_xmss.o",
            "crypto/sm2/libcrypto-shlib-sm2_crypt.o",
            "crypto/sm2/libcrypto-shlib-sm2_err.o",
            "crypto/sm2/libcrypto-shlib-sm2_key.o",
            "crypto/sm2/libcrypto-shlib-sm2_kmeth.o",
            "crypto/sm2/libcrypto-shlib-sm2_sign.o",
            "crypto/sm3/libcrypto-shlib-legacy_sm3.o",
            "crypto/sm3/libcrypto-shlib-sm3.o",
            "crypto/sm4/libcrypto-shlib-sm4.o",
            "crypto/srp/libcrypto-shlib-srp_lib.o",
            "crypto/srp/libcrypto-shlib-srp_vfy.o",
            "crypto/stack/libcrypto-shlib-stack.o",
            "crypto/store/libcrypto-shlib-store_err.o",
            "crypto/store/libcrypto-shlib-store_init.o",
            "crypto/store/libcrypto-shlib-store_lib.o",
            "crypto/store/libcrypto-shlib-store_meth.o",
            "crypto/store/libcrypto-shlib-store_register.o",
            "crypto/store/libcrypto-shlib-store_result.o",
            "crypto/store/libcrypto-shlib-store_strings.o",
            "crypto/thread/arch/libcrypto-shlib-thread_none.o",
            "crypto/thread/arch/libcrypto-shlib-thread_posix.o",
            "crypto/thread/arch/libcrypto-shlib-thread_win.o",
            "crypto/thread/libcrypto-shlib-api.o",
            "crypto/thread/libcrypto-shlib-arch.o",
            "crypto/thread/libcrypto-shlib-internal.o",
            "crypto/ts/libcrypto-shlib-ts_asn1.o",
            "crypto/ts/libcrypto-shlib-ts_conf.o",
            "crypto/ts/libcrypto-shlib-ts_err.o",
            "crypto/ts/libcrypto-shlib-ts_lib.o",
            "crypto/ts/libcrypto-shlib-ts_req_print.o",
            "crypto/ts/libcrypto-shlib-ts_req_utils.o",
            "crypto/ts/libcrypto-shlib-ts_rsp_print.o",
            "crypto/ts/libcrypto-shlib-ts_rsp_sign.o",
            "crypto/ts/libcrypto-shlib-ts_rsp_utils.o",
            "crypto/ts/libcrypto-shlib-ts_rsp_verify.o",
            "crypto/ts/libcrypto-shlib-ts_verify_ctx.o",
            "crypto/tsapi/libcrypto-shlib-tsapi_lib.o",
            "crypto/txt_db/libcrypto-shlib-txt_db.o",
            "crypto/ui/libcrypto-shlib-ui_err.o",
            "crypto/ui/libcrypto-shlib-ui_lib.o",
            "crypto/ui/libcrypto-shlib-ui_null.o",
            "crypto/ui/libcrypto-shlib-ui_openssl.o",
            "crypto/ui/libcrypto-shlib-ui_util.o",
            "crypto/x509/libcrypto-shlib-by_dir.o",
            "crypto/x509/libcrypto-shlib-by_file.o",
            "crypto/x509/libcrypto-shlib-by_store.o",
            "crypto/x509/libcrypto-shlib-pcy_cache.o",
            "crypto/x509/libcrypto-shlib-pcy_data.o",
            "crypto/x509/libcrypto-shlib-pcy_lib.o",
            "crypto/x509/libcrypto-shlib-pcy_map.o",
            "crypto/x509/libcrypto-shlib-pcy_node.o",
            "crypto/x509/libcrypto-shlib-pcy_tree.o",
            "crypto/x509/libcrypto-shlib-t_acert.o",
            "crypto/x509/libcrypto-shlib-t_crl.o",
            "crypto/x509/libcrypto-shlib-t_req.o",
            "crypto/x509/libcrypto-shlib-t_x509.o",
            "crypto/x509/libcrypto-shlib-v3_aaa.o",
            "crypto/x509/libcrypto-shlib-v3_ac_tgt.o",
            "crypto/x509/libcrypto-shlib-v3_addr.o",
            "crypto/x509/libcrypto-shlib-v3_admis.o",
            "crypto/x509/libcrypto-shlib-v3_akeya.o",
            "crypto/x509/libcrypto-shlib-v3_akid.o",
            "crypto/x509/libcrypto-shlib-v3_asid.o",
            "crypto/x509/libcrypto-shlib-v3_attrdesc.o",
            "crypto/x509/libcrypto-shlib-v3_attrmap.o",
            "crypto/x509/libcrypto-shlib-v3_audit_id.o",
            "crypto/x509/libcrypto-shlib-v3_authattid.o",
            "crypto/x509/libcrypto-shlib-v3_battcons.o",
            "crypto/x509/libcrypto-shlib-v3_bcons.o",
            "crypto/x509/libcrypto-shlib-v3_bitst.o",
            "crypto/x509/libcrypto-shlib-v3_conf.o",
            "crypto/x509/libcrypto-shlib-v3_cpols.o",
            "crypto/x509/libcrypto-shlib-v3_crld.o",
            "crypto/x509/libcrypto-shlib-v3_enum.o",
            "crypto/x509/libcrypto-shlib-v3_extku.o",
            "crypto/x509/libcrypto-shlib-v3_genn.o",
            "crypto/x509/libcrypto-shlib-v3_group_ac.o",
            "crypto/x509/libcrypto-shlib-v3_ia5.o",
            "crypto/x509/libcrypto-shlib-v3_ind_iss.o",
            "crypto/x509/libcrypto-shlib-v3_info.o",
            "crypto/x509/libcrypto-shlib-v3_int.o",
            "crypto/x509/libcrypto-shlib-v3_iobo.o",
            "crypto/x509/libcrypto-shlib-v3_lib.o",
            "crypto/x509/libcrypto-shlib-v3_ncons.o",
            "crypto/x509/libcrypto-shlib-v3_no_ass.o",
            "crypto/x509/libcrypto-shlib-v3_no_rev_avail.o",
            "crypto/x509/libcrypto-shlib-v3_pci.o",
            "crypto/x509/libcrypto-shlib-v3_pcia.o",
            "crypto/x509/libcrypto-shlib-v3_pcons.o",
            "crypto/x509/libcrypto-shlib-v3_pku.o",
            "crypto/x509/libcrypto-shlib-v3_pmaps.o",
            "crypto/x509/libcrypto-shlib-v3_prn.o",
            "crypto/x509/libcrypto-shlib-v3_purp.o",
            "crypto/x509/libcrypto-shlib-v3_rolespec.o",
            "crypto/x509/libcrypto-shlib-v3_san.o",
            "crypto/x509/libcrypto-shlib-v3_sda.o",
            "crypto/x509/libcrypto-shlib-v3_single_use.o",
            "crypto/x509/libcrypto-shlib-v3_skid.o",
            "crypto/x509/libcrypto-shlib-v3_soa_id.o",
            "crypto/x509/libcrypto-shlib-v3_sxnet.o",
            "crypto/x509/libcrypto-shlib-v3_timespec.o",
            "crypto/x509/libcrypto-shlib-v3_tlsf.o",
            "crypto/x509/libcrypto-shlib-v3_usernotice.o",
            "crypto/x509/libcrypto-shlib-v3_utl.o",
            "crypto/x509/libcrypto-shlib-v3err.o",
            "crypto/x509/libcrypto-shlib-x509_acert.o",
            "crypto/x509/libcrypto-shlib-x509_att.o",
            "crypto/x509/libcrypto-shlib-x509_cmp.o",
            "crypto/x509/libcrypto-shlib-x509_d2.o",
            "crypto/x509/libcrypto-shlib-x509_def.o",
            "crypto/x509/libcrypto-shlib-x509_err.o",
            "crypto/x509/libcrypto-shlib-x509_ext.o",
            "crypto/x509/libcrypto-shlib-x509_lu.o",
            "crypto/x509/libcrypto-shlib-x509_meth.o",
            "crypto/x509/libcrypto-shlib-x509_obj.o",
            "crypto/x509/libcrypto-shlib-x509_r2x.o",
            "crypto/x509/libcrypto-shlib-x509_req.o",
            "crypto/x509/libcrypto-shlib-x509_set.o",
            "crypto/x509/libcrypto-shlib-x509_trust.o",
            "crypto/x509/libcrypto-shlib-x509_txt.o",
            "crypto/x509/libcrypto-shlib-x509_v3.o",
            "crypto/x509/libcrypto-shlib-x509_vfy.o",
            "crypto/x509/libcrypto-shlib-x509_vpm.o",
            "crypto/x509/libcrypto-shlib-x509aset.o",
            "crypto/x509/libcrypto-shlib-x509cset.o",
            "crypto/x509/libcrypto-shlib-x509name.o",
            "crypto/x509/libcrypto-shlib-x509rset.o",
            "crypto/x509/libcrypto-shlib-x509spki.o",
            "crypto/x509/libcrypto-shlib-x509type.o",
            "crypto/x509/libcrypto-shlib-x_all.o",
            "crypto/x509/libcrypto-shlib-x_attrib.o",
            "crypto/x509/libcrypto-shlib-x_crl.o",
            "crypto/x509/libcrypto-shlib-x_exten.o",
            "crypto/x509/libcrypto-shlib-x_ietfatt.o",
            "crypto/x509/libcrypto-shlib-x_name.o",
            "crypto/x509/libcrypto-shlib-x_pubkey.o",
            "crypto/x509/libcrypto-shlib-x_req.o",
            "crypto/x509/libcrypto-shlib-x_x509.o",
            "crypto/x509/libcrypto-shlib-x_x509a.o",
            "crypto/zuc/libcrypto-shlib-zuc.o",
            "libcrypto.ld",
            "providers/libcrypto-shlib-baseprov.o",
            "providers/libcrypto-shlib-defltprov.o",
            "providers/libcrypto-shlib-nullprov.o",
            "providers/libcrypto-shlib-prov_running.o",
            "providers/libdefault.a"
        ],
        "libssl" => [
            "crypto/hashtable/libssl-shlib-hashfunc.o",
            "crypto/libssl-shlib-ctype.o",
            "crypto/libssl-shlib-getenv.o",
            "crypto/libssl-shlib-packet.o",
            "crypto/libssl-shlib-quic_vlint.o",
            "crypto/libssl-shlib-time.o",
            "crypto/siphash/libssl-shlib-siphash.o",
            "crypto/thread/arch/libssl-shlib-thread_none.o",
            "crypto/thread/arch/libssl-shlib-thread_posix.o",
            "crypto/thread/arch/libssl-shlib-thread_win.o",
            "crypto/thread/libssl-shlib-arch.o",
            "libssl.ld",
            "ssl/libssl-shlib-bio_ssl.o",
            "ssl/libssl-shlib-d1_lib.o",
            "ssl/libssl-shlib-d1_msg.o",
            "ssl/libssl-shlib-d1_srtp.o",
            "ssl/libssl-shlib-methods.o",
            "ssl/libssl-shlib-pqueue.o",
            "ssl/libssl-shlib-priority_queue.o",
            "ssl/libssl-shlib-s3_enc.o",
            "ssl/libssl-shlib-s3_lib.o",
            "ssl/libssl-shlib-s3_msg.o",
            "ssl/libssl-shlib-ssl_asn1.o",
            "ssl/libssl-shlib-ssl_cert.o",
            "ssl/libssl-shlib-ssl_cert_comp.o",
            "ssl/libssl-shlib-ssl_ciph.o",
            "ssl/libssl-shlib-ssl_conf.o",
            "ssl/libssl-shlib-ssl_err_legacy.o",
            "ssl/libssl-shlib-ssl_init.o",
            "ssl/libssl-shlib-ssl_lib.o",
            "ssl/libssl-shlib-ssl_mcnf.o",
            "ssl/libssl-shlib-ssl_rsa.o",
            "ssl/libssl-shlib-ssl_rsa_legacy.o",
            "ssl/libssl-shlib-ssl_sess.o",
            "ssl/libssl-shlib-ssl_stat.o",
            "ssl/libssl-shlib-ssl_txt.o",
            "ssl/libssl-shlib-ssl_utst.o",
            "ssl/libssl-shlib-t1_enc.o",
            "ssl/libssl-shlib-t1_lib.o",
            "ssl/libssl-shlib-t1_trce.o",
            "ssl/libssl-shlib-tls13_enc.o",
            "ssl/libssl-shlib-tls_depr.o",
            "ssl/libssl-shlib-tls_srp.o",
            "ssl/quic/libssl-shlib-cc_newreno.o",
            "ssl/quic/libssl-shlib-json_enc.o",
            "ssl/quic/libssl-shlib-qlog.o",
            "ssl/quic/libssl-shlib-qlog_event_helpers.o",
            "ssl/quic/libssl-shlib-quic_ackm.o",
            "ssl/quic/libssl-shlib-quic_cfq.o",
            "ssl/quic/libssl-shlib-quic_channel.o",
            "ssl/quic/libssl-shlib-quic_demux.o",
            "ssl/quic/libssl-shlib-quic_engine.o",
            "ssl/quic/libssl-shlib-quic_fc.o",
            "ssl/quic/libssl-shlib-quic_fifd.o",
            "ssl/quic/libssl-shlib-quic_impl.o",
            "ssl/quic/libssl-shlib-quic_lcidm.o",
            "ssl/quic/libssl-shlib-quic_method.o",
            "ssl/quic/libssl-shlib-quic_obj.o",
            "ssl/quic/libssl-shlib-quic_port.o",
            "ssl/quic/libssl-shlib-quic_rcidm.o",
            "ssl/quic/libssl-shlib-quic_reactor.o",
            "ssl/quic/libssl-shlib-quic_reactor_wait_ctx.o",
            "ssl/quic/libssl-shlib-quic_record_rx.o",
            "ssl/quic/libssl-shlib-quic_record_shared.o",
            "ssl/quic/libssl-shlib-quic_record_tx.o",
            "ssl/quic/libssl-shlib-quic_record_util.o",
            "ssl/quic/libssl-shlib-quic_rstream.o",
            "ssl/quic/libssl-shlib-quic_rx_depack.o",
            "ssl/quic/libssl-shlib-quic_sf_list.o",
            "ssl/quic/libssl-shlib-quic_srt_gen.o",
            "ssl/quic/libssl-shlib-quic_srtm.o",
            "ssl/quic/libssl-shlib-quic_sstream.o",
            "ssl/quic/libssl-shlib-quic_statm.o",
            "ssl/quic/libssl-shlib-quic_stream_map.o",
            "ssl/quic/libssl-shlib-quic_thread_assist.o",
            "ssl/quic/libssl-shlib-quic_tls.o",
            "ssl/quic/libssl-shlib-quic_tls_api.o",
            "ssl/quic/libssl-shlib-quic_tls_api_method.o",
            "ssl/quic/libssl-shlib-quic_trace.o",
            "ssl/quic/libssl-shlib-quic_tserver.o",
            "ssl/quic/libssl-shlib-quic_txp.o",
            "ssl/quic/libssl-shlib-quic_txpim.o",
            "ssl/quic/libssl-shlib-quic_types.o",
            "ssl/quic/libssl-shlib-quic_wire.o",
            "ssl/quic/libssl-shlib-quic_wire_pkt.o",
            "ssl/quic/libssl-shlib-uint_set.o",
            "ssl/record/libssl-shlib-rec_layer_d1.o",
            "ssl/record/libssl-shlib-rec_layer_s3.o",
            "ssl/record/methods/libssl-shlib-dtls_meth.o",
            "ssl/record/methods/libssl-shlib-ssl3_cbc.o",
            "ssl/record/methods/libssl-shlib-ssl3_meth.o",
            "ssl/record/methods/libssl-shlib-tls13_meth.o",
            "ssl/record/methods/libssl-shlib-tls1_meth.o",
            "ssl/record/methods/libssl-shlib-tls_common.o",
            "ssl/record/methods/libssl-shlib-tls_multib.o",
            "ssl/record/methods/libssl-shlib-tls_pad.o",
            "ssl/record/methods/libssl-shlib-tlsany_meth.o",
            "ssl/rio/libssl-shlib-poll_builder.o",
            "ssl/rio/libssl-shlib-poll_immediate.o",
            "ssl/rio/libssl-shlib-rio_notifier.o",
            "ssl/statem/libssl-shlib-extensions.o",
            "ssl/statem/libssl-shlib-extensions_clnt.o",
            "ssl/statem/libssl-shlib-extensions_cust.o",
            "ssl/statem/libssl-shlib-extensions_srvr.o",
            "ssl/statem/libssl-shlib-statem.o",
            "ssl/statem/libssl-shlib-statem_clnt.o",
            "ssl/statem/libssl-shlib-statem_dtls.o",
            "ssl/statem/libssl-shlib-statem_lib.o",
            "ssl/statem/libssl-shlib-statem_srvr.o",
            "ssl/statem_ntls/libssl-shlib-ntls_extensions.o",
            "ssl/statem_ntls/libssl-shlib-ntls_extensions_clnt.o",
            "ssl/statem_ntls/libssl-shlib-ntls_extensions_cust.o",
            "ssl/statem_ntls/libssl-shlib-ntls_extensions_srvr.o",
            "ssl/statem_ntls/libssl-shlib-ntls_statem.o",
            "ssl/statem_ntls/libssl-shlib-ntls_statem_clnt.o",
            "ssl/statem_ntls/libssl-shlib-ntls_statem_lib.o",
            "ssl/statem_ntls/libssl-shlib-ntls_statem_srvr.o"
        ]
    },
    "sources" => {
        "apps/CA.pl" => [
            "apps/CA.pl.in"
        ],
        "apps/lib/libapps-lib-app_libctx.o" => [
            "apps/lib/app_libctx.c"
        ],
        "apps/lib/libapps-lib-app_params.o" => [
            "apps/lib/app_params.c"
        ],
        "apps/lib/libapps-lib-app_provider.o" => [
            "apps/lib/app_provider.c"
        ],
        "apps/lib/libapps-lib-app_rand.o" => [
            "apps/lib/app_rand.c"
        ],
        "apps/lib/libapps-lib-apps.o" => [
            "apps/lib/apps.c"
        ],
        "apps/lib/libapps-lib-apps_opt_printf.o" => [
            "apps/lib/apps_opt_printf.c"
        ],
        "apps/lib/libapps-lib-apps_ui.o" => [
            "apps/lib/apps_ui.c"
        ],
        "apps/lib/libapps-lib-columns.o" => [
            "apps/lib/columns.c"
        ],
        "apps/lib/libapps-lib-engine.o" => [
            "apps/lib/engine.c"
        ],
        "apps/lib/libapps-lib-engine_loader.o" => [
            "apps/lib/engine_loader.c"
        ],
        "apps/lib/libapps-lib-fmt.o" => [
            "apps/lib/fmt.c"
        ],
        "apps/lib/libapps-lib-http_server.o" => [
            "apps/lib/http_server.c"
        ],
        "apps/lib/libapps-lib-log.o" => [
            "apps/lib/log.c"
        ],
        "apps/lib/libapps-lib-names.o" => [
            "apps/lib/names.c"
        ],
        "apps/lib/libapps-lib-opt.o" => [
            "apps/lib/opt.c"
        ],
        "apps/lib/libapps-lib-s_cb.o" => [
            "apps/lib/s_cb.c"
        ],
        "apps/lib/libapps-lib-s_socket.o" => [
            "apps/lib/s_socket.c"
        ],
        "apps/lib/libapps-lib-tlssrp_depr.o" => [
            "apps/lib/tlssrp_depr.c"
        ],
        "apps/lib/openssl-bin-cmp_mock_srv.o" => [
            "apps/lib/cmp_mock_srv.c"
        ],
        "apps/libapps.a" => [
            "apps/lib/libapps-lib-app_libctx.o",
            "apps/lib/libapps-lib-app_params.o",
            "apps/lib/libapps-lib-app_provider.o",
            "apps/lib/libapps-lib-app_rand.o",
            "apps/lib/libapps-lib-apps.o",
            "apps/lib/libapps-lib-apps_opt_printf.o",
            "apps/lib/libapps-lib-apps_ui.o",
            "apps/lib/libapps-lib-columns.o",
            "apps/lib/libapps-lib-engine.o",
            "apps/lib/libapps-lib-engine_loader.o",
            "apps/lib/libapps-lib-fmt.o",
            "apps/lib/libapps-lib-http_server.o",
            "apps/lib/libapps-lib-log.o",
            "apps/lib/libapps-lib-names.o",
            "apps/lib/libapps-lib-opt.o",
            "apps/lib/libapps-lib-s_cb.o",
            "apps/lib/libapps-lib-s_socket.o",
            "apps/lib/libapps-lib-tlssrp_depr.o"
        ],
        "apps/openssl" => [
            "apps/lib/openssl-bin-cmp_mock_srv.o",
            "apps/openssl-bin-asn1parse.o",
            "apps/openssl-bin-ca.o",
            "apps/openssl-bin-ciphers.o",
            "apps/openssl-bin-cmp.o",
            "apps/openssl-bin-cms.o",
            "apps/openssl-bin-crl.o",
            "apps/openssl-bin-crl2pkcs7.o",
            "apps/openssl-bin-dgst.o",
            "apps/openssl-bin-dhparam.o",
            "apps/openssl-bin-dsa.o",
            "apps/openssl-bin-dsaparam.o",
            "apps/openssl-bin-ec.o",
            "apps/openssl-bin-ecparam.o",
            "apps/openssl-bin-enc.o",
            "apps/openssl-bin-engine.o",
            "apps/openssl-bin-errstr.o",
            "apps/openssl-bin-fipsinstall.o",
            "apps/openssl-bin-gendsa.o",
            "apps/openssl-bin-genpkey.o",
            "apps/openssl-bin-genrsa.o",
            "apps/openssl-bin-info.o",
            "apps/openssl-bin-kdf.o",
            "apps/openssl-bin-list.o",
            "apps/openssl-bin-mac.o",
            "apps/openssl-bin-nseq.o",
            "apps/openssl-bin-ocsp.o",
            "apps/openssl-bin-openssl.o",
            "apps/openssl-bin-passwd.o",
            "apps/openssl-bin-pkcs12.o",
            "apps/openssl-bin-pkcs7.o",
            "apps/openssl-bin-pkcs8.o",
            "apps/openssl-bin-pkey.o",
            "apps/openssl-bin-pkeyparam.o",
            "apps/openssl-bin-pkeyutl.o",
            "apps/openssl-bin-prime.o",
            "apps/openssl-bin-progs.o",
            "apps/openssl-bin-rand.o",
            "apps/openssl-bin-rehash.o",
            "apps/openssl-bin-req.o",
            "apps/openssl-bin-rsa.o",
            "apps/openssl-bin-rsautl.o",
            "apps/openssl-bin-s_client.o",
            "apps/openssl-bin-s_server.o",
            "apps/openssl-bin-s_time.o",
            "apps/openssl-bin-sess_id.o",
            "apps/openssl-bin-skeyutl.o",
            "apps/openssl-bin-smime.o",
            "apps/openssl-bin-speed.o",
            "apps/openssl-bin-spkac.o",
            "apps/openssl-bin-srp.o",
            "apps/openssl-bin-storeutl.o",
            "apps/openssl-bin-ts.o",
            "apps/openssl-bin-verify.o",
            "apps/openssl-bin-version.o",
            "apps/openssl-bin-x509.o"
        ],
        "apps/openssl-bin-asn1parse.o" => [
            "apps/asn1parse.c"
        ],
        "apps/openssl-bin-ca.o" => [
            "apps/ca.c"
        ],
        "apps/openssl-bin-ciphers.o" => [
            "apps/ciphers.c"
        ],
        "apps/openssl-bin-cmp.o" => [
            "apps/cmp.c"
        ],
        "apps/openssl-bin-cms.o" => [
            "apps/cms.c"
        ],
        "apps/openssl-bin-crl.o" => [
            "apps/crl.c"
        ],
        "apps/openssl-bin-crl2pkcs7.o" => [
            "apps/crl2pkcs7.c"
        ],
        "apps/openssl-bin-dgst.o" => [
            "apps/dgst.c"
        ],
        "apps/openssl-bin-dhparam.o" => [
            "apps/dhparam.c"
        ],
        "apps/openssl-bin-dsa.o" => [
            "apps/dsa.c"
        ],
        "apps/openssl-bin-dsaparam.o" => [
            "apps/dsaparam.c"
        ],
        "apps/openssl-bin-ec.o" => [
            "apps/ec.c"
        ],
        "apps/openssl-bin-ecparam.o" => [
            "apps/ecparam.c"
        ],
        "apps/openssl-bin-enc.o" => [
            "apps/enc.c"
        ],
        "apps/openssl-bin-engine.o" => [
            "apps/engine.c"
        ],
        "apps/openssl-bin-errstr.o" => [
            "apps/errstr.c"
        ],
        "apps/openssl-bin-fipsinstall.o" => [
            "apps/fipsinstall.c"
        ],
        "apps/openssl-bin-gendsa.o" => [
            "apps/gendsa.c"
        ],
        "apps/openssl-bin-genpkey.o" => [
            "apps/genpkey.c"
        ],
        "apps/openssl-bin-genrsa.o" => [
            "apps/genrsa.c"
        ],
        "apps/openssl-bin-info.o" => [
            "apps/info.c"
        ],
        "apps/openssl-bin-kdf.o" => [
            "apps/kdf.c"
        ],
        "apps/openssl-bin-list.o" => [
            "apps/list.c"
        ],
        "apps/openssl-bin-mac.o" => [
            "apps/mac.c"
        ],
        "apps/openssl-bin-nseq.o" => [
            "apps/nseq.c"
        ],
        "apps/openssl-bin-ocsp.o" => [
            "apps/ocsp.c"
        ],
        "apps/openssl-bin-openssl.o" => [
            "apps/openssl.c"
        ],
        "apps/openssl-bin-passwd.o" => [
            "apps/passwd.c"
        ],
        "apps/openssl-bin-pkcs12.o" => [
            "apps/pkcs12.c"
        ],
        "apps/openssl-bin-pkcs7.o" => [
            "apps/pkcs7.c"
        ],
        "apps/openssl-bin-pkcs8.o" => [
            "apps/pkcs8.c"
        ],
        "apps/openssl-bin-pkey.o" => [
            "apps/pkey.c"
        ],
        "apps/openssl-bin-pkeyparam.o" => [
            "apps/pkeyparam.c"
        ],
        "apps/openssl-bin-pkeyutl.o" => [
            "apps/pkeyutl.c"
        ],
        "apps/openssl-bin-prime.o" => [
            "apps/prime.c"
        ],
        "apps/openssl-bin-progs.o" => [
            "apps/progs.c"
        ],
        "apps/openssl-bin-rand.o" => [
            "apps/rand.c"
        ],
        "apps/openssl-bin-rehash.o" => [
            "apps/rehash.c"
        ],
        "apps/openssl-bin-req.o" => [
            "apps/req.c"
        ],
        "apps/openssl-bin-rsa.o" => [
            "apps/rsa.c"
        ],
        "apps/openssl-bin-rsautl.o" => [
            "apps/rsautl.c"
        ],
        "apps/openssl-bin-s_client.o" => [
            "apps/s_client.c"
        ],
        "apps/openssl-bin-s_server.o" => [
            "apps/s_server.c"
        ],
        "apps/openssl-bin-s_time.o" => [
            "apps/s_time.c"
        ],
        "apps/openssl-bin-sess_id.o" => [
            "apps/sess_id.c"
        ],
        "apps/openssl-bin-skeyutl.o" => [
            "apps/skeyutl.c"
        ],
        "apps/openssl-bin-smime.o" => [
            "apps/smime.c"
        ],
        "apps/openssl-bin-speed.o" => [
            "apps/speed.c"
        ],
        "apps/openssl-bin-spkac.o" => [
            "apps/spkac.c"
        ],
        "apps/openssl-bin-srp.o" => [
            "apps/srp.c"
        ],
        "apps/openssl-bin-storeutl.o" => [
            "apps/storeutl.c"
        ],
        "apps/openssl-bin-ts.o" => [
            "apps/ts.c"
        ],
        "apps/openssl-bin-verify.o" => [
            "apps/verify.c"
        ],
        "apps/openssl-bin-version.o" => [
            "apps/version.c"
        ],
        "apps/openssl-bin-x509.o" => [
            "apps/x509.c"
        ],
        "apps/tsget.pl" => [
            "apps/tsget.in"
        ],
        "crypto/aes/libcrypto-lib-aes-x86_64.o" => [
            "crypto/aes/aes-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-aes_cfb.o" => [
            "crypto/aes/aes_cfb.c"
        ],
        "crypto/aes/libcrypto-lib-aes_ecb.o" => [
            "crypto/aes/aes_ecb.c"
        ],
        "crypto/aes/libcrypto-lib-aes_ige.o" => [
            "crypto/aes/aes_ige.c"
        ],
        "crypto/aes/libcrypto-lib-aes_misc.o" => [
            "crypto/aes/aes_misc.c"
        ],
        "crypto/aes/libcrypto-lib-aes_ofb.o" => [
            "crypto/aes/aes_ofb.c"
        ],
        "crypto/aes/libcrypto-lib-aes_wrap.o" => [
            "crypto/aes/aes_wrap.c"
        ],
        "crypto/aes/libcrypto-lib-aesni-mb-x86_64.o" => [
            "crypto/aes/aesni-mb-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-aesni-sha1-x86_64.o" => [
            "crypto/aes/aesni-sha1-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-aesni-sha256-x86_64.o" => [
            "crypto/aes/aesni-sha256-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-aesni-x86_64.o" => [
            "crypto/aes/aesni-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-aesni-xts-avx512.o" => [
            "crypto/aes/aesni-xts-avx512.s"
        ],
        "crypto/aes/libcrypto-lib-bsaes-x86_64.o" => [
            "crypto/aes/bsaes-x86_64.s"
        ],
        "crypto/aes/libcrypto-lib-vpaes-x86_64.o" => [
            "crypto/aes/vpaes-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aes-x86_64.o" => [
            "crypto/aes/aes-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aes_cfb.o" => [
            "crypto/aes/aes_cfb.c"
        ],
        "crypto/aes/libcrypto-shlib-aes_ecb.o" => [
            "crypto/aes/aes_ecb.c"
        ],
        "crypto/aes/libcrypto-shlib-aes_ige.o" => [
            "crypto/aes/aes_ige.c"
        ],
        "crypto/aes/libcrypto-shlib-aes_misc.o" => [
            "crypto/aes/aes_misc.c"
        ],
        "crypto/aes/libcrypto-shlib-aes_ofb.o" => [
            "crypto/aes/aes_ofb.c"
        ],
        "crypto/aes/libcrypto-shlib-aes_wrap.o" => [
            "crypto/aes/aes_wrap.c"
        ],
        "crypto/aes/libcrypto-shlib-aesni-mb-x86_64.o" => [
            "crypto/aes/aesni-mb-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aesni-sha1-x86_64.o" => [
            "crypto/aes/aesni-sha1-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aesni-sha256-x86_64.o" => [
            "crypto/aes/aesni-sha256-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aesni-x86_64.o" => [
            "crypto/aes/aesni-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-aesni-xts-avx512.o" => [
            "crypto/aes/aesni-xts-avx512.s"
        ],
        "crypto/aes/libcrypto-shlib-bsaes-x86_64.o" => [
            "crypto/aes/bsaes-x86_64.s"
        ],
        "crypto/aes/libcrypto-shlib-vpaes-x86_64.o" => [
            "crypto/aes/vpaes-x86_64.s"
        ],
        "crypto/asn1/libcrypto-lib-a_bitstr.o" => [
            "crypto/asn1/a_bitstr.c"
        ],
        "crypto/asn1/libcrypto-lib-a_d2i_fp.o" => [
            "crypto/asn1/a_d2i_fp.c"
        ],
        "crypto/asn1/libcrypto-lib-a_digest.o" => [
            "crypto/asn1/a_digest.c"
        ],
        "crypto/asn1/libcrypto-lib-a_dup.o" => [
            "crypto/asn1/a_dup.c"
        ],
        "crypto/asn1/libcrypto-lib-a_gentm.o" => [
            "crypto/asn1/a_gentm.c"
        ],
        "crypto/asn1/libcrypto-lib-a_i2d_fp.o" => [
            "crypto/asn1/a_i2d_fp.c"
        ],
        "crypto/asn1/libcrypto-lib-a_int.o" => [
            "crypto/asn1/a_int.c"
        ],
        "crypto/asn1/libcrypto-lib-a_mbstr.o" => [
            "crypto/asn1/a_mbstr.c"
        ],
        "crypto/asn1/libcrypto-lib-a_object.o" => [
            "crypto/asn1/a_object.c"
        ],
        "crypto/asn1/libcrypto-lib-a_octet.o" => [
            "crypto/asn1/a_octet.c"
        ],
        "crypto/asn1/libcrypto-lib-a_print.o" => [
            "crypto/asn1/a_print.c"
        ],
        "crypto/asn1/libcrypto-lib-a_sign.o" => [
            "crypto/asn1/a_sign.c"
        ],
        "crypto/asn1/libcrypto-lib-a_strex.o" => [
            "crypto/asn1/a_strex.c"
        ],
        "crypto/asn1/libcrypto-lib-a_strnid.o" => [
            "crypto/asn1/a_strnid.c"
        ],
        "crypto/asn1/libcrypto-lib-a_time.o" => [
            "crypto/asn1/a_time.c"
        ],
        "crypto/asn1/libcrypto-lib-a_type.o" => [
            "crypto/asn1/a_type.c"
        ],
        "crypto/asn1/libcrypto-lib-a_utctm.o" => [
            "crypto/asn1/a_utctm.c"
        ],
        "crypto/asn1/libcrypto-lib-a_utf8.o" => [
            "crypto/asn1/a_utf8.c"
        ],
        "crypto/asn1/libcrypto-lib-a_verify.o" => [
            "crypto/asn1/a_verify.c"
        ],
        "crypto/asn1/libcrypto-lib-ameth_lib.o" => [
            "crypto/asn1/ameth_lib.c"
        ],
        "crypto/asn1/libcrypto-lib-asn1_err.o" => [
            "crypto/asn1/asn1_err.c"
        ],
        "crypto/asn1/libcrypto-lib-asn1_gen.o" => [
            "crypto/asn1/asn1_gen.c"
        ],
        "crypto/asn1/libcrypto-lib-asn1_item_list.o" => [
            "crypto/asn1/asn1_item_list.c"
        ],
        "crypto/asn1/libcrypto-lib-asn1_lib.o" => [
            "crypto/asn1/asn1_lib.c"
        ],
        "crypto/asn1/libcrypto-lib-asn1_parse.o" => [
            "crypto/asn1/asn1_parse.c"
        ],
        "crypto/asn1/libcrypto-lib-asn_mime.o" => [
            "crypto/asn1/asn_mime.c"
        ],
        "crypto/asn1/libcrypto-lib-asn_moid.o" => [
            "crypto/asn1/asn_moid.c"
        ],
        "crypto/asn1/libcrypto-lib-asn_mstbl.o" => [
            "crypto/asn1/asn_mstbl.c"
        ],
        "crypto/asn1/libcrypto-lib-asn_pack.o" => [
            "crypto/asn1/asn_pack.c"
        ],
        "crypto/asn1/libcrypto-lib-bio_asn1.o" => [
            "crypto/asn1/bio_asn1.c"
        ],
        "crypto/asn1/libcrypto-lib-bio_ndef.o" => [
            "crypto/asn1/bio_ndef.c"
        ],
        "crypto/asn1/libcrypto-lib-d2i_param.o" => [
            "crypto/asn1/d2i_param.c"
        ],
        "crypto/asn1/libcrypto-lib-d2i_pr.o" => [
            "crypto/asn1/d2i_pr.c"
        ],
        "crypto/asn1/libcrypto-lib-d2i_pu.o" => [
            "crypto/asn1/d2i_pu.c"
        ],
        "crypto/asn1/libcrypto-lib-evp_asn1.o" => [
            "crypto/asn1/evp_asn1.c"
        ],
        "crypto/asn1/libcrypto-lib-f_int.o" => [
            "crypto/asn1/f_int.c"
        ],
        "crypto/asn1/libcrypto-lib-f_string.o" => [
            "crypto/asn1/f_string.c"
        ],
        "crypto/asn1/libcrypto-lib-i2d_evp.o" => [
            "crypto/asn1/i2d_evp.c"
        ],
        "crypto/asn1/libcrypto-lib-n_pkey.o" => [
            "crypto/asn1/n_pkey.c"
        ],
        "crypto/asn1/libcrypto-lib-nsseq.o" => [
            "crypto/asn1/nsseq.c"
        ],
        "crypto/asn1/libcrypto-lib-p5_pbe.o" => [
            "crypto/asn1/p5_pbe.c"
        ],
        "crypto/asn1/libcrypto-lib-p5_pbev2.o" => [
            "crypto/asn1/p5_pbev2.c"
        ],
        "crypto/asn1/libcrypto-lib-p5_scrypt.o" => [
            "crypto/asn1/p5_scrypt.c"
        ],
        "crypto/asn1/libcrypto-lib-p8_pkey.o" => [
            "crypto/asn1/p8_pkey.c"
        ],
        "crypto/asn1/libcrypto-lib-t_bitst.o" => [
            "crypto/asn1/t_bitst.c"
        ],
        "crypto/asn1/libcrypto-lib-t_pkey.o" => [
            "crypto/asn1/t_pkey.c"
        ],
        "crypto/asn1/libcrypto-lib-t_spki.o" => [
            "crypto/asn1/t_spki.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_dec.o" => [
            "crypto/asn1/tasn_dec.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_enc.o" => [
            "crypto/asn1/tasn_enc.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_fre.o" => [
            "crypto/asn1/tasn_fre.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_new.o" => [
            "crypto/asn1/tasn_new.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_prn.o" => [
            "crypto/asn1/tasn_prn.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_scn.o" => [
            "crypto/asn1/tasn_scn.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_typ.o" => [
            "crypto/asn1/tasn_typ.c"
        ],
        "crypto/asn1/libcrypto-lib-tasn_utl.o" => [
            "crypto/asn1/tasn_utl.c"
        ],
        "crypto/asn1/libcrypto-lib-x_algor.o" => [
            "crypto/asn1/x_algor.c"
        ],
        "crypto/asn1/libcrypto-lib-x_bignum.o" => [
            "crypto/asn1/x_bignum.c"
        ],
        "crypto/asn1/libcrypto-lib-x_info.o" => [
            "crypto/asn1/x_info.c"
        ],
        "crypto/asn1/libcrypto-lib-x_int64.o" => [
            "crypto/asn1/x_int64.c"
        ],
        "crypto/asn1/libcrypto-lib-x_long.o" => [
            "crypto/asn1/x_long.c"
        ],
        "crypto/asn1/libcrypto-lib-x_pkey.o" => [
            "crypto/asn1/x_pkey.c"
        ],
        "crypto/asn1/libcrypto-lib-x_sig.o" => [
            "crypto/asn1/x_sig.c"
        ],
        "crypto/asn1/libcrypto-lib-x_spki.o" => [
            "crypto/asn1/x_spki.c"
        ],
        "crypto/asn1/libcrypto-lib-x_val.o" => [
            "crypto/asn1/x_val.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_bitstr.o" => [
            "crypto/asn1/a_bitstr.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_d2i_fp.o" => [
            "crypto/asn1/a_d2i_fp.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_digest.o" => [
            "crypto/asn1/a_digest.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_dup.o" => [
            "crypto/asn1/a_dup.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_gentm.o" => [
            "crypto/asn1/a_gentm.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_i2d_fp.o" => [
            "crypto/asn1/a_i2d_fp.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_int.o" => [
            "crypto/asn1/a_int.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_mbstr.o" => [
            "crypto/asn1/a_mbstr.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_object.o" => [
            "crypto/asn1/a_object.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_octet.o" => [
            "crypto/asn1/a_octet.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_print.o" => [
            "crypto/asn1/a_print.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_sign.o" => [
            "crypto/asn1/a_sign.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_strex.o" => [
            "crypto/asn1/a_strex.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_strnid.o" => [
            "crypto/asn1/a_strnid.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_time.o" => [
            "crypto/asn1/a_time.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_type.o" => [
            "crypto/asn1/a_type.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_utctm.o" => [
            "crypto/asn1/a_utctm.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_utf8.o" => [
            "crypto/asn1/a_utf8.c"
        ],
        "crypto/asn1/libcrypto-shlib-a_verify.o" => [
            "crypto/asn1/a_verify.c"
        ],
        "crypto/asn1/libcrypto-shlib-ameth_lib.o" => [
            "crypto/asn1/ameth_lib.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn1_err.o" => [
            "crypto/asn1/asn1_err.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn1_gen.o" => [
            "crypto/asn1/asn1_gen.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn1_item_list.o" => [
            "crypto/asn1/asn1_item_list.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn1_lib.o" => [
            "crypto/asn1/asn1_lib.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn1_parse.o" => [
            "crypto/asn1/asn1_parse.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn_mime.o" => [
            "crypto/asn1/asn_mime.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn_moid.o" => [
            "crypto/asn1/asn_moid.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn_mstbl.o" => [
            "crypto/asn1/asn_mstbl.c"
        ],
        "crypto/asn1/libcrypto-shlib-asn_pack.o" => [
            "crypto/asn1/asn_pack.c"
        ],
        "crypto/asn1/libcrypto-shlib-bio_asn1.o" => [
            "crypto/asn1/bio_asn1.c"
        ],
        "crypto/asn1/libcrypto-shlib-bio_ndef.o" => [
            "crypto/asn1/bio_ndef.c"
        ],
        "crypto/asn1/libcrypto-shlib-d2i_param.o" => [
            "crypto/asn1/d2i_param.c"
        ],
        "crypto/asn1/libcrypto-shlib-d2i_pr.o" => [
            "crypto/asn1/d2i_pr.c"
        ],
        "crypto/asn1/libcrypto-shlib-d2i_pu.o" => [
            "crypto/asn1/d2i_pu.c"
        ],
        "crypto/asn1/libcrypto-shlib-evp_asn1.o" => [
            "crypto/asn1/evp_asn1.c"
        ],
        "crypto/asn1/libcrypto-shlib-f_int.o" => [
            "crypto/asn1/f_int.c"
        ],
        "crypto/asn1/libcrypto-shlib-f_string.o" => [
            "crypto/asn1/f_string.c"
        ],
        "crypto/asn1/libcrypto-shlib-i2d_evp.o" => [
            "crypto/asn1/i2d_evp.c"
        ],
        "crypto/asn1/libcrypto-shlib-n_pkey.o" => [
            "crypto/asn1/n_pkey.c"
        ],
        "crypto/asn1/libcrypto-shlib-nsseq.o" => [
            "crypto/asn1/nsseq.c"
        ],
        "crypto/asn1/libcrypto-shlib-p5_pbe.o" => [
            "crypto/asn1/p5_pbe.c"
        ],
        "crypto/asn1/libcrypto-shlib-p5_pbev2.o" => [
            "crypto/asn1/p5_pbev2.c"
        ],
        "crypto/asn1/libcrypto-shlib-p5_scrypt.o" => [
            "crypto/asn1/p5_scrypt.c"
        ],
        "crypto/asn1/libcrypto-shlib-p8_pkey.o" => [
            "crypto/asn1/p8_pkey.c"
        ],
        "crypto/asn1/libcrypto-shlib-t_bitst.o" => [
            "crypto/asn1/t_bitst.c"
        ],
        "crypto/asn1/libcrypto-shlib-t_pkey.o" => [
            "crypto/asn1/t_pkey.c"
        ],
        "crypto/asn1/libcrypto-shlib-t_spki.o" => [
            "crypto/asn1/t_spki.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_dec.o" => [
            "crypto/asn1/tasn_dec.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_enc.o" => [
            "crypto/asn1/tasn_enc.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_fre.o" => [
            "crypto/asn1/tasn_fre.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_new.o" => [
            "crypto/asn1/tasn_new.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_prn.o" => [
            "crypto/asn1/tasn_prn.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_scn.o" => [
            "crypto/asn1/tasn_scn.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_typ.o" => [
            "crypto/asn1/tasn_typ.c"
        ],
        "crypto/asn1/libcrypto-shlib-tasn_utl.o" => [
            "crypto/asn1/tasn_utl.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_algor.o" => [
            "crypto/asn1/x_algor.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_bignum.o" => [
            "crypto/asn1/x_bignum.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_info.o" => [
            "crypto/asn1/x_info.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_int64.o" => [
            "crypto/asn1/x_int64.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_long.o" => [
            "crypto/asn1/x_long.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_pkey.o" => [
            "crypto/asn1/x_pkey.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_sig.o" => [
            "crypto/asn1/x_sig.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_spki.o" => [
            "crypto/asn1/x_spki.c"
        ],
        "crypto/asn1/libcrypto-shlib-x_val.o" => [
            "crypto/asn1/x_val.c"
        ],
        "crypto/async/arch/libcrypto-lib-async_null.o" => [
            "crypto/async/arch/async_null.c"
        ],
        "crypto/async/arch/libcrypto-lib-async_posix.o" => [
            "crypto/async/arch/async_posix.c"
        ],
        "crypto/async/arch/libcrypto-lib-async_win.o" => [
            "crypto/async/arch/async_win.c"
        ],
        "crypto/async/arch/libcrypto-shlib-async_null.o" => [
            "crypto/async/arch/async_null.c"
        ],
        "crypto/async/arch/libcrypto-shlib-async_posix.o" => [
            "crypto/async/arch/async_posix.c"
        ],
        "crypto/async/arch/libcrypto-shlib-async_win.o" => [
            "crypto/async/arch/async_win.c"
        ],
        "crypto/async/libcrypto-lib-async.o" => [
            "crypto/async/async.c"
        ],
        "crypto/async/libcrypto-lib-async_err.o" => [
            "crypto/async/async_err.c"
        ],
        "crypto/async/libcrypto-lib-async_wait.o" => [
            "crypto/async/async_wait.c"
        ],
        "crypto/async/libcrypto-shlib-async.o" => [
            "crypto/async/async.c"
        ],
        "crypto/async/libcrypto-shlib-async_err.o" => [
            "crypto/async/async_err.c"
        ],
        "crypto/async/libcrypto-shlib-async_wait.o" => [
            "crypto/async/async_wait.c"
        ],
        "crypto/bio/libcrypto-lib-bf_buff.o" => [
            "crypto/bio/bf_buff.c"
        ],
        "crypto/bio/libcrypto-lib-bf_lbuf.o" => [
            "crypto/bio/bf_lbuf.c"
        ],
        "crypto/bio/libcrypto-lib-bf_nbio.o" => [
            "crypto/bio/bf_nbio.c"
        ],
        "crypto/bio/libcrypto-lib-bf_null.o" => [
            "crypto/bio/bf_null.c"
        ],
        "crypto/bio/libcrypto-lib-bf_prefix.o" => [
            "crypto/bio/bf_prefix.c"
        ],
        "crypto/bio/libcrypto-lib-bf_readbuff.o" => [
            "crypto/bio/bf_readbuff.c"
        ],
        "crypto/bio/libcrypto-lib-bio_addr.o" => [
            "crypto/bio/bio_addr.c"
        ],
        "crypto/bio/libcrypto-lib-bio_cb.o" => [
            "crypto/bio/bio_cb.c"
        ],
        "crypto/bio/libcrypto-lib-bio_dump.o" => [
            "crypto/bio/bio_dump.c"
        ],
        "crypto/bio/libcrypto-lib-bio_err.o" => [
            "crypto/bio/bio_err.c"
        ],
        "crypto/bio/libcrypto-lib-bio_lib.o" => [
            "crypto/bio/bio_lib.c"
        ],
        "crypto/bio/libcrypto-lib-bio_meth.o" => [
            "crypto/bio/bio_meth.c"
        ],
        "crypto/bio/libcrypto-lib-bio_print.o" => [
            "crypto/bio/bio_print.c"
        ],
        "crypto/bio/libcrypto-lib-bio_sock.o" => [
            "crypto/bio/bio_sock.c"
        ],
        "crypto/bio/libcrypto-lib-bio_sock2.o" => [
            "crypto/bio/bio_sock2.c"
        ],
        "crypto/bio/libcrypto-lib-bss_acpt.o" => [
            "crypto/bio/bss_acpt.c"
        ],
        "crypto/bio/libcrypto-lib-bss_bio.o" => [
            "crypto/bio/bss_bio.c"
        ],
        "crypto/bio/libcrypto-lib-bss_conn.o" => [
            "crypto/bio/bss_conn.c"
        ],
        "crypto/bio/libcrypto-lib-bss_core.o" => [
            "crypto/bio/bss_core.c"
        ],
        "crypto/bio/libcrypto-lib-bss_dgram.o" => [
            "crypto/bio/bss_dgram.c"
        ],
        "crypto/bio/libcrypto-lib-bss_dgram_pair.o" => [
            "crypto/bio/bss_dgram_pair.c"
        ],
        "crypto/bio/libcrypto-lib-bss_fd.o" => [
            "crypto/bio/bss_fd.c"
        ],
        "crypto/bio/libcrypto-lib-bss_file.o" => [
            "crypto/bio/bss_file.c"
        ],
        "crypto/bio/libcrypto-lib-bss_log.o" => [
            "crypto/bio/bss_log.c"
        ],
        "crypto/bio/libcrypto-lib-bss_mem.o" => [
            "crypto/bio/bss_mem.c"
        ],
        "crypto/bio/libcrypto-lib-bss_null.o" => [
            "crypto/bio/bss_null.c"
        ],
        "crypto/bio/libcrypto-lib-bss_sock.o" => [
            "crypto/bio/bss_sock.c"
        ],
        "crypto/bio/libcrypto-lib-ossl_core_bio.o" => [
            "crypto/bio/ossl_core_bio.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_buff.o" => [
            "crypto/bio/bf_buff.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_lbuf.o" => [
            "crypto/bio/bf_lbuf.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_nbio.o" => [
            "crypto/bio/bf_nbio.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_null.o" => [
            "crypto/bio/bf_null.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_prefix.o" => [
            "crypto/bio/bf_prefix.c"
        ],
        "crypto/bio/libcrypto-shlib-bf_readbuff.o" => [
            "crypto/bio/bf_readbuff.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_addr.o" => [
            "crypto/bio/bio_addr.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_cb.o" => [
            "crypto/bio/bio_cb.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_dump.o" => [
            "crypto/bio/bio_dump.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_err.o" => [
            "crypto/bio/bio_err.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_lib.o" => [
            "crypto/bio/bio_lib.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_meth.o" => [
            "crypto/bio/bio_meth.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_print.o" => [
            "crypto/bio/bio_print.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_sock.o" => [
            "crypto/bio/bio_sock.c"
        ],
        "crypto/bio/libcrypto-shlib-bio_sock2.o" => [
            "crypto/bio/bio_sock2.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_acpt.o" => [
            "crypto/bio/bss_acpt.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_bio.o" => [
            "crypto/bio/bss_bio.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_conn.o" => [
            "crypto/bio/bss_conn.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_core.o" => [
            "crypto/bio/bss_core.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_dgram.o" => [
            "crypto/bio/bss_dgram.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_dgram_pair.o" => [
            "crypto/bio/bss_dgram_pair.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_fd.o" => [
            "crypto/bio/bss_fd.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_file.o" => [
            "crypto/bio/bss_file.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_log.o" => [
            "crypto/bio/bss_log.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_mem.o" => [
            "crypto/bio/bss_mem.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_null.o" => [
            "crypto/bio/bss_null.c"
        ],
        "crypto/bio/libcrypto-shlib-bss_sock.o" => [
            "crypto/bio/bss_sock.c"
        ],
        "crypto/bio/libcrypto-shlib-ossl_core_bio.o" => [
            "crypto/bio/ossl_core_bio.c"
        ],
        "crypto/bn/asm/libcrypto-lib-x86_64-gcc.o" => [
            "crypto/bn/asm/x86_64-gcc.c"
        ],
        "crypto/bn/asm/libcrypto-shlib-x86_64-gcc.o" => [
            "crypto/bn/asm/x86_64-gcc.c"
        ],
        "crypto/bn/libcrypto-lib-bn_add.o" => [
            "crypto/bn/bn_add.c"
        ],
        "crypto/bn/libcrypto-lib-bn_blind.o" => [
            "crypto/bn/bn_blind.c"
        ],
        "crypto/bn/libcrypto-lib-bn_const.o" => [
            "crypto/bn/bn_const.c"
        ],
        "crypto/bn/libcrypto-lib-bn_conv.o" => [
            "crypto/bn/bn_conv.c"
        ],
        "crypto/bn/libcrypto-lib-bn_ctx.o" => [
            "crypto/bn/bn_ctx.c"
        ],
        "crypto/bn/libcrypto-lib-bn_depr.o" => [
            "crypto/bn/bn_depr.c"
        ],
        "crypto/bn/libcrypto-lib-bn_dh.o" => [
            "crypto/bn/bn_dh.c"
        ],
        "crypto/bn/libcrypto-lib-bn_div.o" => [
            "crypto/bn/bn_div.c"
        ],
        "crypto/bn/libcrypto-lib-bn_err.o" => [
            "crypto/bn/bn_err.c"
        ],
        "crypto/bn/libcrypto-lib-bn_exp.o" => [
            "crypto/bn/bn_exp.c"
        ],
        "crypto/bn/libcrypto-lib-bn_exp2.o" => [
            "crypto/bn/bn_exp2.c"
        ],
        "crypto/bn/libcrypto-lib-bn_gcd.o" => [
            "crypto/bn/bn_gcd.c"
        ],
        "crypto/bn/libcrypto-lib-bn_gf2m.o" => [
            "crypto/bn/bn_gf2m.c"
        ],
        "crypto/bn/libcrypto-lib-bn_intern.o" => [
            "crypto/bn/bn_intern.c"
        ],
        "crypto/bn/libcrypto-lib-bn_kron.o" => [
            "crypto/bn/bn_kron.c"
        ],
        "crypto/bn/libcrypto-lib-bn_lib.o" => [
            "crypto/bn/bn_lib.c"
        ],
        "crypto/bn/libcrypto-lib-bn_mod.o" => [
            "crypto/bn/bn_mod.c"
        ],
        "crypto/bn/libcrypto-lib-bn_mont.o" => [
            "crypto/bn/bn_mont.c"
        ],
        "crypto/bn/libcrypto-lib-bn_mpi.o" => [
            "crypto/bn/bn_mpi.c"
        ],
        "crypto/bn/libcrypto-lib-bn_mul.o" => [
            "crypto/bn/bn_mul.c"
        ],
        "crypto/bn/libcrypto-lib-bn_nist.o" => [
            "crypto/bn/bn_nist.c"
        ],
        "crypto/bn/libcrypto-lib-bn_prime.o" => [
            "crypto/bn/bn_prime.c"
        ],
        "crypto/bn/libcrypto-lib-bn_print.o" => [
            "crypto/bn/bn_print.c"
        ],
        "crypto/bn/libcrypto-lib-bn_rand.o" => [
            "crypto/bn/bn_rand.c"
        ],
        "crypto/bn/libcrypto-lib-bn_recp.o" => [
            "crypto/bn/bn_recp.c"
        ],
        "crypto/bn/libcrypto-lib-bn_rsa_fips186_4.o" => [
            "crypto/bn/bn_rsa_fips186_4.c"
        ],
        "crypto/bn/libcrypto-lib-bn_shift.o" => [
            "crypto/bn/bn_shift.c"
        ],
        "crypto/bn/libcrypto-lib-bn_sm2.o" => [
            "crypto/bn/bn_sm2.c"
        ],
        "crypto/bn/libcrypto-lib-bn_sqr.o" => [
            "crypto/bn/bn_sqr.c"
        ],
        "crypto/bn/libcrypto-lib-bn_sqrt.o" => [
            "crypto/bn/bn_sqrt.c"
        ],
        "crypto/bn/libcrypto-lib-bn_srp.o" => [
            "crypto/bn/bn_srp.c"
        ],
        "crypto/bn/libcrypto-lib-bn_word.o" => [
            "crypto/bn/bn_word.c"
        ],
        "crypto/bn/libcrypto-lib-bn_x931p.o" => [
            "crypto/bn/bn_x931p.c"
        ],
        "crypto/bn/libcrypto-lib-rsaz-2k-avx512.o" => [
            "crypto/bn/rsaz-2k-avx512.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-2k-avxifma.o" => [
            "crypto/bn/rsaz-2k-avxifma.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-3k-avx512.o" => [
            "crypto/bn/rsaz-3k-avx512.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-3k-avxifma.o" => [
            "crypto/bn/rsaz-3k-avxifma.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-4k-avx512.o" => [
            "crypto/bn/rsaz-4k-avx512.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-4k-avxifma.o" => [
            "crypto/bn/rsaz-4k-avxifma.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-avx2.o" => [
            "crypto/bn/rsaz-avx2.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz-x86_64.o" => [
            "crypto/bn/rsaz-x86_64.s"
        ],
        "crypto/bn/libcrypto-lib-rsaz_exp.o" => [
            "crypto/bn/rsaz_exp.c"
        ],
        "crypto/bn/libcrypto-lib-rsaz_exp_x2.o" => [
            "crypto/bn/rsaz_exp_x2.c"
        ],
        "crypto/bn/libcrypto-lib-x86_64-gf2m.o" => [
            "crypto/bn/x86_64-gf2m.s"
        ],
        "crypto/bn/libcrypto-lib-x86_64-mont.o" => [
            "crypto/bn/x86_64-mont.s"
        ],
        "crypto/bn/libcrypto-lib-x86_64-mont5.o" => [
            "crypto/bn/x86_64-mont5.s"
        ],
        "crypto/bn/libcrypto-shlib-bn_add.o" => [
            "crypto/bn/bn_add.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_blind.o" => [
            "crypto/bn/bn_blind.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_const.o" => [
            "crypto/bn/bn_const.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_conv.o" => [
            "crypto/bn/bn_conv.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_ctx.o" => [
            "crypto/bn/bn_ctx.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_depr.o" => [
            "crypto/bn/bn_depr.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_dh.o" => [
            "crypto/bn/bn_dh.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_div.o" => [
            "crypto/bn/bn_div.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_err.o" => [
            "crypto/bn/bn_err.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_exp.o" => [
            "crypto/bn/bn_exp.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_exp2.o" => [
            "crypto/bn/bn_exp2.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_gcd.o" => [
            "crypto/bn/bn_gcd.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_gf2m.o" => [
            "crypto/bn/bn_gf2m.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_intern.o" => [
            "crypto/bn/bn_intern.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_kron.o" => [
            "crypto/bn/bn_kron.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_lib.o" => [
            "crypto/bn/bn_lib.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_mod.o" => [
            "crypto/bn/bn_mod.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_mont.o" => [
            "crypto/bn/bn_mont.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_mpi.o" => [
            "crypto/bn/bn_mpi.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_mul.o" => [
            "crypto/bn/bn_mul.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_nist.o" => [
            "crypto/bn/bn_nist.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_prime.o" => [
            "crypto/bn/bn_prime.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_print.o" => [
            "crypto/bn/bn_print.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_rand.o" => [
            "crypto/bn/bn_rand.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_recp.o" => [
            "crypto/bn/bn_recp.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_rsa_fips186_4.o" => [
            "crypto/bn/bn_rsa_fips186_4.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_shift.o" => [
            "crypto/bn/bn_shift.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_sm2.o" => [
            "crypto/bn/bn_sm2.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_sqr.o" => [
            "crypto/bn/bn_sqr.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_sqrt.o" => [
            "crypto/bn/bn_sqrt.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_srp.o" => [
            "crypto/bn/bn_srp.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_word.o" => [
            "crypto/bn/bn_word.c"
        ],
        "crypto/bn/libcrypto-shlib-bn_x931p.o" => [
            "crypto/bn/bn_x931p.c"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-2k-avx512.o" => [
            "crypto/bn/rsaz-2k-avx512.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-2k-avxifma.o" => [
            "crypto/bn/rsaz-2k-avxifma.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-3k-avx512.o" => [
            "crypto/bn/rsaz-3k-avx512.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-3k-avxifma.o" => [
            "crypto/bn/rsaz-3k-avxifma.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-4k-avx512.o" => [
            "crypto/bn/rsaz-4k-avx512.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-4k-avxifma.o" => [
            "crypto/bn/rsaz-4k-avxifma.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-avx2.o" => [
            "crypto/bn/rsaz-avx2.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz-x86_64.o" => [
            "crypto/bn/rsaz-x86_64.s"
        ],
        "crypto/bn/libcrypto-shlib-rsaz_exp.o" => [
            "crypto/bn/rsaz_exp.c"
        ],
        "crypto/bn/libcrypto-shlib-rsaz_exp_x2.o" => [
            "crypto/bn/rsaz_exp_x2.c"
        ],
        "crypto/bn/libcrypto-shlib-x86_64-gf2m.o" => [
            "crypto/bn/x86_64-gf2m.s"
        ],
        "crypto/bn/libcrypto-shlib-x86_64-mont.o" => [
            "crypto/bn/x86_64-mont.s"
        ],
        "crypto/bn/libcrypto-shlib-x86_64-mont5.o" => [
            "crypto/bn/x86_64-mont5.s"
        ],
        "crypto/buffer/libcrypto-lib-buf_err.o" => [
            "crypto/buffer/buf_err.c"
        ],
        "crypto/buffer/libcrypto-lib-buffer.o" => [
            "crypto/buffer/buffer.c"
        ],
        "crypto/buffer/libcrypto-shlib-buf_err.o" => [
            "crypto/buffer/buf_err.c"
        ],
        "crypto/buffer/libcrypto-shlib-buffer.o" => [
            "crypto/buffer/buffer.c"
        ],
        "crypto/chacha/libcrypto-lib-chacha-x86_64.o" => [
            "crypto/chacha/chacha-x86_64.s"
        ],
        "crypto/chacha/libcrypto-shlib-chacha-x86_64.o" => [
            "crypto/chacha/chacha-x86_64.s"
        ],
        "crypto/cmac/libcrypto-lib-cmac.o" => [
            "crypto/cmac/cmac.c"
        ],
        "crypto/cmac/libcrypto-shlib-cmac.o" => [
            "crypto/cmac/cmac.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_asn.o" => [
            "crypto/cmp/cmp_asn.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_client.o" => [
            "crypto/cmp/cmp_client.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_ctx.o" => [
            "crypto/cmp/cmp_ctx.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_err.o" => [
            "crypto/cmp/cmp_err.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_genm.o" => [
            "crypto/cmp/cmp_genm.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_hdr.o" => [
            "crypto/cmp/cmp_hdr.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_http.o" => [
            "crypto/cmp/cmp_http.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_msg.o" => [
            "crypto/cmp/cmp_msg.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_protect.o" => [
            "crypto/cmp/cmp_protect.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_server.o" => [
            "crypto/cmp/cmp_server.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_status.o" => [
            "crypto/cmp/cmp_status.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_util.o" => [
            "crypto/cmp/cmp_util.c"
        ],
        "crypto/cmp/libcrypto-lib-cmp_vfy.o" => [
            "crypto/cmp/cmp_vfy.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_asn.o" => [
            "crypto/cmp/cmp_asn.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_client.o" => [
            "crypto/cmp/cmp_client.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_ctx.o" => [
            "crypto/cmp/cmp_ctx.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_err.o" => [
            "crypto/cmp/cmp_err.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_genm.o" => [
            "crypto/cmp/cmp_genm.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_hdr.o" => [
            "crypto/cmp/cmp_hdr.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_http.o" => [
            "crypto/cmp/cmp_http.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_msg.o" => [
            "crypto/cmp/cmp_msg.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_protect.o" => [
            "crypto/cmp/cmp_protect.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_server.o" => [
            "crypto/cmp/cmp_server.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_status.o" => [
            "crypto/cmp/cmp_status.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_util.o" => [
            "crypto/cmp/cmp_util.c"
        ],
        "crypto/cmp/libcrypto-shlib-cmp_vfy.o" => [
            "crypto/cmp/cmp_vfy.c"
        ],
        "crypto/cms/libcrypto-lib-cms_asn1.o" => [
            "crypto/cms/cms_asn1.c"
        ],
        "crypto/cms/libcrypto-lib-cms_att.o" => [
            "crypto/cms/cms_att.c"
        ],
        "crypto/cms/libcrypto-lib-cms_cd.o" => [
            "crypto/cms/cms_cd.c"
        ],
        "crypto/cms/libcrypto-lib-cms_dd.o" => [
            "crypto/cms/cms_dd.c"
        ],
        "crypto/cms/libcrypto-lib-cms_dh.o" => [
            "crypto/cms/cms_dh.c"
        ],
        "crypto/cms/libcrypto-lib-cms_ec.o" => [
            "crypto/cms/cms_ec.c"
        ],
        "crypto/cms/libcrypto-lib-cms_enc.o" => [
            "crypto/cms/cms_enc.c"
        ],
        "crypto/cms/libcrypto-lib-cms_env.o" => [
            "crypto/cms/cms_env.c"
        ],
        "crypto/cms/libcrypto-lib-cms_err.o" => [
            "crypto/cms/cms_err.c"
        ],
        "crypto/cms/libcrypto-lib-cms_ess.o" => [
            "crypto/cms/cms_ess.c"
        ],
        "crypto/cms/libcrypto-lib-cms_io.o" => [
            "crypto/cms/cms_io.c"
        ],
        "crypto/cms/libcrypto-lib-cms_kari.o" => [
            "crypto/cms/cms_kari.c"
        ],
        "crypto/cms/libcrypto-lib-cms_lib.o" => [
            "crypto/cms/cms_lib.c"
        ],
        "crypto/cms/libcrypto-lib-cms_pwri.o" => [
            "crypto/cms/cms_pwri.c"
        ],
        "crypto/cms/libcrypto-lib-cms_rsa.o" => [
            "crypto/cms/cms_rsa.c"
        ],
        "crypto/cms/libcrypto-lib-cms_sd.o" => [
            "crypto/cms/cms_sd.c"
        ],
        "crypto/cms/libcrypto-lib-cms_smime.o" => [
            "crypto/cms/cms_smime.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_asn1.o" => [
            "crypto/cms/cms_asn1.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_att.o" => [
            "crypto/cms/cms_att.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_cd.o" => [
            "crypto/cms/cms_cd.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_dd.o" => [
            "crypto/cms/cms_dd.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_dh.o" => [
            "crypto/cms/cms_dh.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_ec.o" => [
            "crypto/cms/cms_ec.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_enc.o" => [
            "crypto/cms/cms_enc.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_env.o" => [
            "crypto/cms/cms_env.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_err.o" => [
            "crypto/cms/cms_err.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_ess.o" => [
            "crypto/cms/cms_ess.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_io.o" => [
            "crypto/cms/cms_io.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_kari.o" => [
            "crypto/cms/cms_kari.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_lib.o" => [
            "crypto/cms/cms_lib.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_pwri.o" => [
            "crypto/cms/cms_pwri.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_rsa.o" => [
            "crypto/cms/cms_rsa.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_sd.o" => [
            "crypto/cms/cms_sd.c"
        ],
        "crypto/cms/libcrypto-shlib-cms_smime.o" => [
            "crypto/cms/cms_smime.c"
        ],
        "crypto/comp/libcrypto-lib-c_brotli.o" => [
            "crypto/comp/c_brotli.c"
        ],
        "crypto/comp/libcrypto-lib-c_zlib.o" => [
            "crypto/comp/c_zlib.c"
        ],
        "crypto/comp/libcrypto-lib-c_zstd.o" => [
            "crypto/comp/c_zstd.c"
        ],
        "crypto/comp/libcrypto-lib-comp_err.o" => [
            "crypto/comp/comp_err.c"
        ],
        "crypto/comp/libcrypto-lib-comp_lib.o" => [
            "crypto/comp/comp_lib.c"
        ],
        "crypto/comp/libcrypto-shlib-c_brotli.o" => [
            "crypto/comp/c_brotli.c"
        ],
        "crypto/comp/libcrypto-shlib-c_zlib.o" => [
            "crypto/comp/c_zlib.c"
        ],
        "crypto/comp/libcrypto-shlib-c_zstd.o" => [
            "crypto/comp/c_zstd.c"
        ],
        "crypto/comp/libcrypto-shlib-comp_err.o" => [
            "crypto/comp/comp_err.c"
        ],
        "crypto/comp/libcrypto-shlib-comp_lib.o" => [
            "crypto/comp/comp_lib.c"
        ],
        "crypto/conf/libcrypto-lib-conf_api.o" => [
            "crypto/conf/conf_api.c"
        ],
        "crypto/conf/libcrypto-lib-conf_def.o" => [
            "crypto/conf/conf_def.c"
        ],
        "crypto/conf/libcrypto-lib-conf_err.o" => [
            "crypto/conf/conf_err.c"
        ],
        "crypto/conf/libcrypto-lib-conf_lib.o" => [
            "crypto/conf/conf_lib.c"
        ],
        "crypto/conf/libcrypto-lib-conf_mall.o" => [
            "crypto/conf/conf_mall.c"
        ],
        "crypto/conf/libcrypto-lib-conf_mod.o" => [
            "crypto/conf/conf_mod.c"
        ],
        "crypto/conf/libcrypto-lib-conf_sap.o" => [
            "crypto/conf/conf_sap.c"
        ],
        "crypto/conf/libcrypto-lib-conf_ssl.o" => [
            "crypto/conf/conf_ssl.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_api.o" => [
            "crypto/conf/conf_api.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_def.o" => [
            "crypto/conf/conf_def.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_err.o" => [
            "crypto/conf/conf_err.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_lib.o" => [
            "crypto/conf/conf_lib.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_mall.o" => [
            "crypto/conf/conf_mall.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_mod.o" => [
            "crypto/conf/conf_mod.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_sap.o" => [
            "crypto/conf/conf_sap.c"
        ],
        "crypto/conf/libcrypto-shlib-conf_ssl.o" => [
            "crypto/conf/conf_ssl.c"
        ],
        "crypto/crmf/libcrypto-lib-crmf_asn.o" => [
            "crypto/crmf/crmf_asn.c"
        ],
        "crypto/crmf/libcrypto-lib-crmf_err.o" => [
            "crypto/crmf/crmf_err.c"
        ],
        "crypto/crmf/libcrypto-lib-crmf_lib.o" => [
            "crypto/crmf/crmf_lib.c"
        ],
        "crypto/crmf/libcrypto-lib-crmf_pbm.o" => [
            "crypto/crmf/crmf_pbm.c"
        ],
        "crypto/crmf/libcrypto-shlib-crmf_asn.o" => [
            "crypto/crmf/crmf_asn.c"
        ],
        "crypto/crmf/libcrypto-shlib-crmf_err.o" => [
            "crypto/crmf/crmf_err.c"
        ],
        "crypto/crmf/libcrypto-shlib-crmf_lib.o" => [
            "crypto/crmf/crmf_lib.c"
        ],
        "crypto/crmf/libcrypto-shlib-crmf_pbm.o" => [
            "crypto/crmf/crmf_pbm.c"
        ],
        "crypto/ct/libcrypto-lib-ct_b64.o" => [
            "crypto/ct/ct_b64.c"
        ],
        "crypto/ct/libcrypto-lib-ct_err.o" => [
            "crypto/ct/ct_err.c"
        ],
        "crypto/ct/libcrypto-lib-ct_log.o" => [
            "crypto/ct/ct_log.c"
        ],
        "crypto/ct/libcrypto-lib-ct_oct.o" => [
            "crypto/ct/ct_oct.c"
        ],
        "crypto/ct/libcrypto-lib-ct_policy.o" => [
            "crypto/ct/ct_policy.c"
        ],
        "crypto/ct/libcrypto-lib-ct_prn.o" => [
            "crypto/ct/ct_prn.c"
        ],
        "crypto/ct/libcrypto-lib-ct_sct.o" => [
            "crypto/ct/ct_sct.c"
        ],
        "crypto/ct/libcrypto-lib-ct_sct_ctx.o" => [
            "crypto/ct/ct_sct_ctx.c"
        ],
        "crypto/ct/libcrypto-lib-ct_vfy.o" => [
            "crypto/ct/ct_vfy.c"
        ],
        "crypto/ct/libcrypto-lib-ct_x509v3.o" => [
            "crypto/ct/ct_x509v3.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_b64.o" => [
            "crypto/ct/ct_b64.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_err.o" => [
            "crypto/ct/ct_err.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_log.o" => [
            "crypto/ct/ct_log.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_oct.o" => [
            "crypto/ct/ct_oct.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_policy.o" => [
            "crypto/ct/ct_policy.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_prn.o" => [
            "crypto/ct/ct_prn.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_sct.o" => [
            "crypto/ct/ct_sct.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_sct_ctx.o" => [
            "crypto/ct/ct_sct_ctx.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_vfy.o" => [
            "crypto/ct/ct_vfy.c"
        ],
        "crypto/ct/libcrypto-shlib-ct_x509v3.o" => [
            "crypto/ct/ct_x509v3.c"
        ],
        "crypto/des/libcrypto-lib-cbc_cksm.o" => [
            "crypto/des/cbc_cksm.c"
        ],
        "crypto/des/libcrypto-lib-cbc_enc.o" => [
            "crypto/des/cbc_enc.c"
        ],
        "crypto/des/libcrypto-lib-cfb64ede.o" => [
            "crypto/des/cfb64ede.c"
        ],
        "crypto/des/libcrypto-lib-cfb64enc.o" => [
            "crypto/des/cfb64enc.c"
        ],
        "crypto/des/libcrypto-lib-cfb_enc.o" => [
            "crypto/des/cfb_enc.c"
        ],
        "crypto/des/libcrypto-lib-des_enc.o" => [
            "crypto/des/des_enc.c"
        ],
        "crypto/des/libcrypto-lib-ecb3_enc.o" => [
            "crypto/des/ecb3_enc.c"
        ],
        "crypto/des/libcrypto-lib-ecb_enc.o" => [
            "crypto/des/ecb_enc.c"
        ],
        "crypto/des/libcrypto-lib-fcrypt.o" => [
            "crypto/des/fcrypt.c"
        ],
        "crypto/des/libcrypto-lib-fcrypt_b.o" => [
            "crypto/des/fcrypt_b.c"
        ],
        "crypto/des/libcrypto-lib-ofb64ede.o" => [
            "crypto/des/ofb64ede.c"
        ],
        "crypto/des/libcrypto-lib-ofb64enc.o" => [
            "crypto/des/ofb64enc.c"
        ],
        "crypto/des/libcrypto-lib-ofb_enc.o" => [
            "crypto/des/ofb_enc.c"
        ],
        "crypto/des/libcrypto-lib-pcbc_enc.o" => [
            "crypto/des/pcbc_enc.c"
        ],
        "crypto/des/libcrypto-lib-qud_cksm.o" => [
            "crypto/des/qud_cksm.c"
        ],
        "crypto/des/libcrypto-lib-rand_key.o" => [
            "crypto/des/rand_key.c"
        ],
        "crypto/des/libcrypto-lib-set_key.o" => [
            "crypto/des/set_key.c"
        ],
        "crypto/des/libcrypto-lib-str2key.o" => [
            "crypto/des/str2key.c"
        ],
        "crypto/des/libcrypto-lib-xcbc_enc.o" => [
            "crypto/des/xcbc_enc.c"
        ],
        "crypto/des/libcrypto-shlib-cbc_cksm.o" => [
            "crypto/des/cbc_cksm.c"
        ],
        "crypto/des/libcrypto-shlib-cbc_enc.o" => [
            "crypto/des/cbc_enc.c"
        ],
        "crypto/des/libcrypto-shlib-cfb64ede.o" => [
            "crypto/des/cfb64ede.c"
        ],
        "crypto/des/libcrypto-shlib-cfb64enc.o" => [
            "crypto/des/cfb64enc.c"
        ],
        "crypto/des/libcrypto-shlib-cfb_enc.o" => [
            "crypto/des/cfb_enc.c"
        ],
        "crypto/des/libcrypto-shlib-des_enc.o" => [
            "crypto/des/des_enc.c"
        ],
        "crypto/des/libcrypto-shlib-ecb3_enc.o" => [
            "crypto/des/ecb3_enc.c"
        ],
        "crypto/des/libcrypto-shlib-ecb_enc.o" => [
            "crypto/des/ecb_enc.c"
        ],
        "crypto/des/libcrypto-shlib-fcrypt.o" => [
            "crypto/des/fcrypt.c"
        ],
        "crypto/des/libcrypto-shlib-fcrypt_b.o" => [
            "crypto/des/fcrypt_b.c"
        ],
        "crypto/des/libcrypto-shlib-ofb64ede.o" => [
            "crypto/des/ofb64ede.c"
        ],
        "crypto/des/libcrypto-shlib-ofb64enc.o" => [
            "crypto/des/ofb64enc.c"
        ],
        "crypto/des/libcrypto-shlib-ofb_enc.o" => [
            "crypto/des/ofb_enc.c"
        ],
        "crypto/des/libcrypto-shlib-pcbc_enc.o" => [
            "crypto/des/pcbc_enc.c"
        ],
        "crypto/des/libcrypto-shlib-qud_cksm.o" => [
            "crypto/des/qud_cksm.c"
        ],
        "crypto/des/libcrypto-shlib-rand_key.o" => [
            "crypto/des/rand_key.c"
        ],
        "crypto/des/libcrypto-shlib-set_key.o" => [
            "crypto/des/set_key.c"
        ],
        "crypto/des/libcrypto-shlib-str2key.o" => [
            "crypto/des/str2key.c"
        ],
        "crypto/des/libcrypto-shlib-xcbc_enc.o" => [
            "crypto/des/xcbc_enc.c"
        ],
        "crypto/des/liblegacy-lib-des_enc.o" => [
            "crypto/des/des_enc.c"
        ],
        "crypto/des/liblegacy-lib-fcrypt_b.o" => [
            "crypto/des/fcrypt_b.c"
        ],
        "crypto/dh/libcrypto-lib-dh_ameth.o" => [
            "crypto/dh/dh_ameth.c"
        ],
        "crypto/dh/libcrypto-lib-dh_asn1.o" => [
            "crypto/dh/dh_asn1.c"
        ],
        "crypto/dh/libcrypto-lib-dh_backend.o" => [
            "crypto/dh/dh_backend.c"
        ],
        "crypto/dh/libcrypto-lib-dh_check.o" => [
            "crypto/dh/dh_check.c"
        ],
        "crypto/dh/libcrypto-lib-dh_depr.o" => [
            "crypto/dh/dh_depr.c"
        ],
        "crypto/dh/libcrypto-lib-dh_err.o" => [
            "crypto/dh/dh_err.c"
        ],
        "crypto/dh/libcrypto-lib-dh_gen.o" => [
            "crypto/dh/dh_gen.c"
        ],
        "crypto/dh/libcrypto-lib-dh_group_params.o" => [
            "crypto/dh/dh_group_params.c"
        ],
        "crypto/dh/libcrypto-lib-dh_kdf.o" => [
            "crypto/dh/dh_kdf.c"
        ],
        "crypto/dh/libcrypto-lib-dh_key.o" => [
            "crypto/dh/dh_key.c"
        ],
        "crypto/dh/libcrypto-lib-dh_lib.o" => [
            "crypto/dh/dh_lib.c"
        ],
        "crypto/dh/libcrypto-lib-dh_meth.o" => [
            "crypto/dh/dh_meth.c"
        ],
        "crypto/dh/libcrypto-lib-dh_pmeth.o" => [
            "crypto/dh/dh_pmeth.c"
        ],
        "crypto/dh/libcrypto-lib-dh_prn.o" => [
            "crypto/dh/dh_prn.c"
        ],
        "crypto/dh/libcrypto-lib-dh_rfc5114.o" => [
            "crypto/dh/dh_rfc5114.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_ameth.o" => [
            "crypto/dh/dh_ameth.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_asn1.o" => [
            "crypto/dh/dh_asn1.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_backend.o" => [
            "crypto/dh/dh_backend.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_check.o" => [
            "crypto/dh/dh_check.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_depr.o" => [
            "crypto/dh/dh_depr.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_err.o" => [
            "crypto/dh/dh_err.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_gen.o" => [
            "crypto/dh/dh_gen.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_group_params.o" => [
            "crypto/dh/dh_group_params.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_kdf.o" => [
            "crypto/dh/dh_kdf.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_key.o" => [
            "crypto/dh/dh_key.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_lib.o" => [
            "crypto/dh/dh_lib.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_meth.o" => [
            "crypto/dh/dh_meth.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_pmeth.o" => [
            "crypto/dh/dh_pmeth.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_prn.o" => [
            "crypto/dh/dh_prn.c"
        ],
        "crypto/dh/libcrypto-shlib-dh_rfc5114.o" => [
            "crypto/dh/dh_rfc5114.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_ameth.o" => [
            "crypto/dsa/dsa_ameth.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_asn1.o" => [
            "crypto/dsa/dsa_asn1.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_backend.o" => [
            "crypto/dsa/dsa_backend.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_check.o" => [
            "crypto/dsa/dsa_check.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_depr.o" => [
            "crypto/dsa/dsa_depr.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_err.o" => [
            "crypto/dsa/dsa_err.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_gen.o" => [
            "crypto/dsa/dsa_gen.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_key.o" => [
            "crypto/dsa/dsa_key.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_lib.o" => [
            "crypto/dsa/dsa_lib.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_meth.o" => [
            "crypto/dsa/dsa_meth.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_ossl.o" => [
            "crypto/dsa/dsa_ossl.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_pmeth.o" => [
            "crypto/dsa/dsa_pmeth.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_prn.o" => [
            "crypto/dsa/dsa_prn.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_sign.o" => [
            "crypto/dsa/dsa_sign.c"
        ],
        "crypto/dsa/libcrypto-lib-dsa_vrf.o" => [
            "crypto/dsa/dsa_vrf.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_ameth.o" => [
            "crypto/dsa/dsa_ameth.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_asn1.o" => [
            "crypto/dsa/dsa_asn1.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_backend.o" => [
            "crypto/dsa/dsa_backend.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_check.o" => [
            "crypto/dsa/dsa_check.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_depr.o" => [
            "crypto/dsa/dsa_depr.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_err.o" => [
            "crypto/dsa/dsa_err.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_gen.o" => [
            "crypto/dsa/dsa_gen.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_key.o" => [
            "crypto/dsa/dsa_key.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_lib.o" => [
            "crypto/dsa/dsa_lib.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_meth.o" => [
            "crypto/dsa/dsa_meth.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_ossl.o" => [
            "crypto/dsa/dsa_ossl.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_pmeth.o" => [
            "crypto/dsa/dsa_pmeth.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_prn.o" => [
            "crypto/dsa/dsa_prn.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_sign.o" => [
            "crypto/dsa/dsa_sign.c"
        ],
        "crypto/dsa/libcrypto-shlib-dsa_vrf.o" => [
            "crypto/dsa/dsa_vrf.c"
        ],
        "crypto/dso/libcrypto-lib-dso_dl.o" => [
            "crypto/dso/dso_dl.c"
        ],
        "crypto/dso/libcrypto-lib-dso_dlfcn.o" => [
            "crypto/dso/dso_dlfcn.c"
        ],
        "crypto/dso/libcrypto-lib-dso_err.o" => [
            "crypto/dso/dso_err.c"
        ],
        "crypto/dso/libcrypto-lib-dso_lib.o" => [
            "crypto/dso/dso_lib.c"
        ],
        "crypto/dso/libcrypto-lib-dso_openssl.o" => [
            "crypto/dso/dso_openssl.c"
        ],
        "crypto/dso/libcrypto-lib-dso_win32.o" => [
            "crypto/dso/dso_win32.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_dl.o" => [
            "crypto/dso/dso_dl.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_dlfcn.o" => [
            "crypto/dso/dso_dlfcn.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_err.o" => [
            "crypto/dso/dso_err.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_lib.o" => [
            "crypto/dso/dso_lib.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_openssl.o" => [
            "crypto/dso/dso_openssl.c"
        ],
        "crypto/dso/libcrypto-shlib-dso_win32.o" => [
            "crypto/dso/dso_win32.c"
        ],
        "crypto/ec/curve448/arch_32/libcrypto-lib-f_impl32.o" => [
            "crypto/ec/curve448/arch_32/f_impl32.c"
        ],
        "crypto/ec/curve448/arch_32/libcrypto-shlib-f_impl32.o" => [
            "crypto/ec/curve448/arch_32/f_impl32.c"
        ],
        "crypto/ec/curve448/arch_64/libcrypto-lib-f_impl64.o" => [
            "crypto/ec/curve448/arch_64/f_impl64.c"
        ],
        "crypto/ec/curve448/arch_64/libcrypto-shlib-f_impl64.o" => [
            "crypto/ec/curve448/arch_64/f_impl64.c"
        ],
        "crypto/ec/curve448/libcrypto-lib-curve448.o" => [
            "crypto/ec/curve448/curve448.c"
        ],
        "crypto/ec/curve448/libcrypto-lib-curve448_tables.o" => [
            "crypto/ec/curve448/curve448_tables.c"
        ],
        "crypto/ec/curve448/libcrypto-lib-eddsa.o" => [
            "crypto/ec/curve448/eddsa.c"
        ],
        "crypto/ec/curve448/libcrypto-lib-f_generic.o" => [
            "crypto/ec/curve448/f_generic.c"
        ],
        "crypto/ec/curve448/libcrypto-lib-scalar.o" => [
            "crypto/ec/curve448/scalar.c"
        ],
        "crypto/ec/curve448/libcrypto-shlib-curve448.o" => [
            "crypto/ec/curve448/curve448.c"
        ],
        "crypto/ec/curve448/libcrypto-shlib-curve448_tables.o" => [
            "crypto/ec/curve448/curve448_tables.c"
        ],
        "crypto/ec/curve448/libcrypto-shlib-eddsa.o" => [
            "crypto/ec/curve448/eddsa.c"
        ],
        "crypto/ec/curve448/libcrypto-shlib-f_generic.o" => [
            "crypto/ec/curve448/f_generic.c"
        ],
        "crypto/ec/curve448/libcrypto-shlib-scalar.o" => [
            "crypto/ec/curve448/scalar.c"
        ],
        "crypto/ec/libcrypto-lib-curve25519.o" => [
            "crypto/ec/curve25519.c"
        ],
        "crypto/ec/libcrypto-lib-ec2_oct.o" => [
            "crypto/ec/ec2_oct.c"
        ],
        "crypto/ec/libcrypto-lib-ec2_smpl.o" => [
            "crypto/ec/ec2_smpl.c"
        ],
        "crypto/ec/libcrypto-lib-ec_ameth.o" => [
            "crypto/ec/ec_ameth.c"
        ],
        "crypto/ec/libcrypto-lib-ec_asn1.o" => [
            "crypto/ec/ec_asn1.c"
        ],
        "crypto/ec/libcrypto-lib-ec_backend.o" => [
            "crypto/ec/ec_backend.c"
        ],
        "crypto/ec/libcrypto-lib-ec_check.o" => [
            "crypto/ec/ec_check.c"
        ],
        "crypto/ec/libcrypto-lib-ec_curve.o" => [
            "crypto/ec/ec_curve.c"
        ],
        "crypto/ec/libcrypto-lib-ec_cvt.o" => [
            "crypto/ec/ec_cvt.c"
        ],
        "crypto/ec/libcrypto-lib-ec_deprecated.o" => [
            "crypto/ec/ec_deprecated.c"
        ],
        "crypto/ec/libcrypto-lib-ec_err.o" => [
            "crypto/ec/ec_err.c"
        ],
        "crypto/ec/libcrypto-lib-ec_key.o" => [
            "crypto/ec/ec_key.c"
        ],
        "crypto/ec/libcrypto-lib-ec_kmeth.o" => [
            "crypto/ec/ec_kmeth.c"
        ],
        "crypto/ec/libcrypto-lib-ec_lib.o" => [
            "crypto/ec/ec_lib.c"
        ],
        "crypto/ec/libcrypto-lib-ec_mult.o" => [
            "crypto/ec/ec_mult.c"
        ],
        "crypto/ec/libcrypto-lib-ec_oct.o" => [
            "crypto/ec/ec_oct.c"
        ],
        "crypto/ec/libcrypto-lib-ec_pmeth.o" => [
            "crypto/ec/ec_pmeth.c"
        ],
        "crypto/ec/libcrypto-lib-ec_print.o" => [
            "crypto/ec/ec_print.c"
        ],
        "crypto/ec/libcrypto-lib-ecdh_kdf.o" => [
            "crypto/ec/ecdh_kdf.c"
        ],
        "crypto/ec/libcrypto-lib-ecdh_ossl.o" => [
            "crypto/ec/ecdh_ossl.c"
        ],
        "crypto/ec/libcrypto-lib-ecdsa_ossl.o" => [
            "crypto/ec/ecdsa_ossl.c"
        ],
        "crypto/ec/libcrypto-lib-ecdsa_sign.o" => [
            "crypto/ec/ecdsa_sign.c"
        ],
        "crypto/ec/libcrypto-lib-ecdsa_vrf.o" => [
            "crypto/ec/ecdsa_vrf.c"
        ],
        "crypto/ec/libcrypto-lib-eck_prn.o" => [
            "crypto/ec/eck_prn.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_meth.o" => [
            "crypto/ec/ecp_meth.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_mont.o" => [
            "crypto/ec/ecp_mont.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_nist.o" => [
            "crypto/ec/ecp_nist.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_nistz256-x86_64.o" => [
            "crypto/ec/ecp_nistz256-x86_64.s"
        ],
        "crypto/ec/libcrypto-lib-ecp_nistz256.o" => [
            "crypto/ec/ecp_nistz256.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_oct.o" => [
            "crypto/ec/ecp_oct.c"
        ],
        "crypto/ec/libcrypto-lib-ecp_smpl.o" => [
            "crypto/ec/ecp_smpl.c"
        ],
        "crypto/ec/libcrypto-lib-ecx_backend.o" => [
            "crypto/ec/ecx_backend.c"
        ],
        "crypto/ec/libcrypto-lib-ecx_key.o" => [
            "crypto/ec/ecx_key.c"
        ],
        "crypto/ec/libcrypto-lib-ecx_meth.o" => [
            "crypto/ec/ecx_meth.c"
        ],
        "crypto/ec/libcrypto-lib-x25519-x86_64.o" => [
            "crypto/ec/x25519-x86_64.s"
        ],
        "crypto/ec/libcrypto-shlib-curve25519.o" => [
            "crypto/ec/curve25519.c"
        ],
        "crypto/ec/libcrypto-shlib-ec2_oct.o" => [
            "crypto/ec/ec2_oct.c"
        ],
        "crypto/ec/libcrypto-shlib-ec2_smpl.o" => [
            "crypto/ec/ec2_smpl.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_ameth.o" => [
            "crypto/ec/ec_ameth.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_asn1.o" => [
            "crypto/ec/ec_asn1.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_backend.o" => [
            "crypto/ec/ec_backend.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_check.o" => [
            "crypto/ec/ec_check.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_curve.o" => [
            "crypto/ec/ec_curve.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_cvt.o" => [
            "crypto/ec/ec_cvt.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_deprecated.o" => [
            "crypto/ec/ec_deprecated.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_err.o" => [
            "crypto/ec/ec_err.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_key.o" => [
            "crypto/ec/ec_key.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_kmeth.o" => [
            "crypto/ec/ec_kmeth.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_lib.o" => [
            "crypto/ec/ec_lib.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_mult.o" => [
            "crypto/ec/ec_mult.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_oct.o" => [
            "crypto/ec/ec_oct.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_pmeth.o" => [
            "crypto/ec/ec_pmeth.c"
        ],
        "crypto/ec/libcrypto-shlib-ec_print.o" => [
            "crypto/ec/ec_print.c"
        ],
        "crypto/ec/libcrypto-shlib-ecdh_kdf.o" => [
            "crypto/ec/ecdh_kdf.c"
        ],
        "crypto/ec/libcrypto-shlib-ecdh_ossl.o" => [
            "crypto/ec/ecdh_ossl.c"
        ],
        "crypto/ec/libcrypto-shlib-ecdsa_ossl.o" => [
            "crypto/ec/ecdsa_ossl.c"
        ],
        "crypto/ec/libcrypto-shlib-ecdsa_sign.o" => [
            "crypto/ec/ecdsa_sign.c"
        ],
        "crypto/ec/libcrypto-shlib-ecdsa_vrf.o" => [
            "crypto/ec/ecdsa_vrf.c"
        ],
        "crypto/ec/libcrypto-shlib-eck_prn.o" => [
            "crypto/ec/eck_prn.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_meth.o" => [
            "crypto/ec/ecp_meth.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_mont.o" => [
            "crypto/ec/ecp_mont.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_nist.o" => [
            "crypto/ec/ecp_nist.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_nistz256-x86_64.o" => [
            "crypto/ec/ecp_nistz256-x86_64.s"
        ],
        "crypto/ec/libcrypto-shlib-ecp_nistz256.o" => [
            "crypto/ec/ecp_nistz256.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_oct.o" => [
            "crypto/ec/ecp_oct.c"
        ],
        "crypto/ec/libcrypto-shlib-ecp_smpl.o" => [
            "crypto/ec/ecp_smpl.c"
        ],
        "crypto/ec/libcrypto-shlib-ecx_backend.o" => [
            "crypto/ec/ecx_backend.c"
        ],
        "crypto/ec/libcrypto-shlib-ecx_key.o" => [
            "crypto/ec/ecx_key.c"
        ],
        "crypto/ec/libcrypto-shlib-ecx_meth.o" => [
            "crypto/ec/ecx_meth.c"
        ],
        "crypto/ec/libcrypto-shlib-x25519-x86_64.o" => [
            "crypto/ec/x25519-x86_64.s"
        ],
        "crypto/eia3/libcrypto-lib-eia3.o" => [
            "crypto/eia3/eia3.c"
        ],
        "crypto/eia3/libcrypto-shlib-eia3.o" => [
            "crypto/eia3/eia3.c"
        ],
        "crypto/encode_decode/libcrypto-lib-decoder_err.o" => [
            "crypto/encode_decode/decoder_err.c"
        ],
        "crypto/encode_decode/libcrypto-lib-decoder_lib.o" => [
            "crypto/encode_decode/decoder_lib.c"
        ],
        "crypto/encode_decode/libcrypto-lib-decoder_meth.o" => [
            "crypto/encode_decode/decoder_meth.c"
        ],
        "crypto/encode_decode/libcrypto-lib-decoder_pkey.o" => [
            "crypto/encode_decode/decoder_pkey.c"
        ],
        "crypto/encode_decode/libcrypto-lib-encoder_err.o" => [
            "crypto/encode_decode/encoder_err.c"
        ],
        "crypto/encode_decode/libcrypto-lib-encoder_lib.o" => [
            "crypto/encode_decode/encoder_lib.c"
        ],
        "crypto/encode_decode/libcrypto-lib-encoder_meth.o" => [
            "crypto/encode_decode/encoder_meth.c"
        ],
        "crypto/encode_decode/libcrypto-lib-encoder_pkey.o" => [
            "crypto/encode_decode/encoder_pkey.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-decoder_err.o" => [
            "crypto/encode_decode/decoder_err.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-decoder_lib.o" => [
            "crypto/encode_decode/decoder_lib.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-decoder_meth.o" => [
            "crypto/encode_decode/decoder_meth.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-decoder_pkey.o" => [
            "crypto/encode_decode/decoder_pkey.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-encoder_err.o" => [
            "crypto/encode_decode/encoder_err.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-encoder_lib.o" => [
            "crypto/encode_decode/encoder_lib.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-encoder_meth.o" => [
            "crypto/encode_decode/encoder_meth.c"
        ],
        "crypto/encode_decode/libcrypto-shlib-encoder_pkey.o" => [
            "crypto/encode_decode/encoder_pkey.c"
        ],
        "crypto/engine/libcrypto-lib-eng_all.o" => [
            "crypto/engine/eng_all.c"
        ],
        "crypto/engine/libcrypto-lib-eng_cnf.o" => [
            "crypto/engine/eng_cnf.c"
        ],
        "crypto/engine/libcrypto-lib-eng_ctrl.o" => [
            "crypto/engine/eng_ctrl.c"
        ],
        "crypto/engine/libcrypto-lib-eng_dyn.o" => [
            "crypto/engine/eng_dyn.c"
        ],
        "crypto/engine/libcrypto-lib-eng_err.o" => [
            "crypto/engine/eng_err.c"
        ],
        "crypto/engine/libcrypto-lib-eng_fat.o" => [
            "crypto/engine/eng_fat.c"
        ],
        "crypto/engine/libcrypto-lib-eng_init.o" => [
            "crypto/engine/eng_init.c"
        ],
        "crypto/engine/libcrypto-lib-eng_lib.o" => [
            "crypto/engine/eng_lib.c"
        ],
        "crypto/engine/libcrypto-lib-eng_list.o" => [
            "crypto/engine/eng_list.c"
        ],
        "crypto/engine/libcrypto-lib-eng_openssl.o" => [
            "crypto/engine/eng_openssl.c"
        ],
        "crypto/engine/libcrypto-lib-eng_pkey.o" => [
            "crypto/engine/eng_pkey.c"
        ],
        "crypto/engine/libcrypto-lib-eng_rdrand.o" => [
            "crypto/engine/eng_rdrand.c"
        ],
        "crypto/engine/libcrypto-lib-eng_table.o" => [
            "crypto/engine/eng_table.c"
        ],
        "crypto/engine/libcrypto-lib-tb_asnmth.o" => [
            "crypto/engine/tb_asnmth.c"
        ],
        "crypto/engine/libcrypto-lib-tb_cipher.o" => [
            "crypto/engine/tb_cipher.c"
        ],
        "crypto/engine/libcrypto-lib-tb_dh.o" => [
            "crypto/engine/tb_dh.c"
        ],
        "crypto/engine/libcrypto-lib-tb_digest.o" => [
            "crypto/engine/tb_digest.c"
        ],
        "crypto/engine/libcrypto-lib-tb_dsa.o" => [
            "crypto/engine/tb_dsa.c"
        ],
        "crypto/engine/libcrypto-lib-tb_eckey.o" => [
            "crypto/engine/tb_eckey.c"
        ],
        "crypto/engine/libcrypto-lib-tb_ecpmeth.o" => [
            "crypto/engine/tb_ecpmeth.c"
        ],
        "crypto/engine/libcrypto-lib-tb_pkmeth.o" => [
            "crypto/engine/tb_pkmeth.c"
        ],
        "crypto/engine/libcrypto-lib-tb_rand.o" => [
            "crypto/engine/tb_rand.c"
        ],
        "crypto/engine/libcrypto-lib-tb_rsa.o" => [
            "crypto/engine/tb_rsa.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_all.o" => [
            "crypto/engine/eng_all.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_cnf.o" => [
            "crypto/engine/eng_cnf.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_ctrl.o" => [
            "crypto/engine/eng_ctrl.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_dyn.o" => [
            "crypto/engine/eng_dyn.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_err.o" => [
            "crypto/engine/eng_err.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_fat.o" => [
            "crypto/engine/eng_fat.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_init.o" => [
            "crypto/engine/eng_init.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_lib.o" => [
            "crypto/engine/eng_lib.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_list.o" => [
            "crypto/engine/eng_list.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_openssl.o" => [
            "crypto/engine/eng_openssl.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_pkey.o" => [
            "crypto/engine/eng_pkey.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_rdrand.o" => [
            "crypto/engine/eng_rdrand.c"
        ],
        "crypto/engine/libcrypto-shlib-eng_table.o" => [
            "crypto/engine/eng_table.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_asnmth.o" => [
            "crypto/engine/tb_asnmth.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_cipher.o" => [
            "crypto/engine/tb_cipher.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_dh.o" => [
            "crypto/engine/tb_dh.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_digest.o" => [
            "crypto/engine/tb_digest.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_dsa.o" => [
            "crypto/engine/tb_dsa.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_eckey.o" => [
            "crypto/engine/tb_eckey.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_ecpmeth.o" => [
            "crypto/engine/tb_ecpmeth.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_pkmeth.o" => [
            "crypto/engine/tb_pkmeth.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_rand.o" => [
            "crypto/engine/tb_rand.c"
        ],
        "crypto/engine/libcrypto-shlib-tb_rsa.o" => [
            "crypto/engine/tb_rsa.c"
        ],
        "crypto/err/libcrypto-lib-err.o" => [
            "crypto/err/err.c"
        ],
        "crypto/err/libcrypto-lib-err_all.o" => [
            "crypto/err/err_all.c"
        ],
        "crypto/err/libcrypto-lib-err_all_legacy.o" => [
            "crypto/err/err_all_legacy.c"
        ],
        "crypto/err/libcrypto-lib-err_blocks.o" => [
            "crypto/err/err_blocks.c"
        ],
        "crypto/err/libcrypto-lib-err_mark.o" => [
            "crypto/err/err_mark.c"
        ],
        "crypto/err/libcrypto-lib-err_prn.o" => [
            "crypto/err/err_prn.c"
        ],
        "crypto/err/libcrypto-lib-err_save.o" => [
            "crypto/err/err_save.c"
        ],
        "crypto/err/libcrypto-shlib-err.o" => [
            "crypto/err/err.c"
        ],
        "crypto/err/libcrypto-shlib-err_all.o" => [
            "crypto/err/err_all.c"
        ],
        "crypto/err/libcrypto-shlib-err_all_legacy.o" => [
            "crypto/err/err_all_legacy.c"
        ],
        "crypto/err/libcrypto-shlib-err_blocks.o" => [
            "crypto/err/err_blocks.c"
        ],
        "crypto/err/libcrypto-shlib-err_mark.o" => [
            "crypto/err/err_mark.c"
        ],
        "crypto/err/libcrypto-shlib-err_prn.o" => [
            "crypto/err/err_prn.c"
        ],
        "crypto/err/libcrypto-shlib-err_save.o" => [
            "crypto/err/err_save.c"
        ],
        "crypto/ess/libcrypto-lib-ess_asn1.o" => [
            "crypto/ess/ess_asn1.c"
        ],
        "crypto/ess/libcrypto-lib-ess_err.o" => [
            "crypto/ess/ess_err.c"
        ],
        "crypto/ess/libcrypto-lib-ess_lib.o" => [
            "crypto/ess/ess_lib.c"
        ],
        "crypto/ess/libcrypto-shlib-ess_asn1.o" => [
            "crypto/ess/ess_asn1.c"
        ],
        "crypto/ess/libcrypto-shlib-ess_err.o" => [
            "crypto/ess/ess_err.c"
        ],
        "crypto/ess/libcrypto-shlib-ess_lib.o" => [
            "crypto/ess/ess_lib.c"
        ],
        "crypto/evp/libcrypto-lib-asymcipher.o" => [
            "crypto/evp/asymcipher.c"
        ],
        "crypto/evp/libcrypto-lib-bio_b64.o" => [
            "crypto/evp/bio_b64.c"
        ],
        "crypto/evp/libcrypto-lib-bio_enc.o" => [
            "crypto/evp/bio_enc.c"
        ],
        "crypto/evp/libcrypto-lib-bio_md.o" => [
            "crypto/evp/bio_md.c"
        ],
        "crypto/evp/libcrypto-lib-bio_ok.o" => [
            "crypto/evp/bio_ok.c"
        ],
        "crypto/evp/libcrypto-lib-c_allc.o" => [
            "crypto/evp/c_allc.c"
        ],
        "crypto/evp/libcrypto-lib-c_alld.o" => [
            "crypto/evp/c_alld.c"
        ],
        "crypto/evp/libcrypto-lib-cmeth_lib.o" => [
            "crypto/evp/cmeth_lib.c"
        ],
        "crypto/evp/libcrypto-lib-ctrl_params_translate.o" => [
            "crypto/evp/ctrl_params_translate.c"
        ],
        "crypto/evp/libcrypto-lib-dh_ctrl.o" => [
            "crypto/evp/dh_ctrl.c"
        ],
        "crypto/evp/libcrypto-lib-dh_support.o" => [
            "crypto/evp/dh_support.c"
        ],
        "crypto/evp/libcrypto-lib-digest.o" => [
            "crypto/evp/digest.c"
        ],
        "crypto/evp/libcrypto-lib-dsa_ctrl.o" => [
            "crypto/evp/dsa_ctrl.c"
        ],
        "crypto/evp/libcrypto-lib-e_aes.o" => [
            "crypto/evp/e_aes.c"
        ],
        "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha1.o" => [
            "crypto/evp/e_aes_cbc_hmac_sha1.c"
        ],
        "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha256.o" => [
            "crypto/evp/e_aes_cbc_hmac_sha256.c"
        ],
        "crypto/evp/libcrypto-lib-e_chacha20_poly1305.o" => [
            "crypto/evp/e_chacha20_poly1305.c"
        ],
        "crypto/evp/libcrypto-lib-e_des.o" => [
            "crypto/evp/e_des.c"
        ],
        "crypto/evp/libcrypto-lib-e_des3.o" => [
            "crypto/evp/e_des3.c"
        ],
        "crypto/evp/libcrypto-lib-e_eea3.o" => [
            "crypto/evp/e_eea3.c"
        ],
        "crypto/evp/libcrypto-lib-e_null.o" => [
            "crypto/evp/e_null.c"
        ],
        "crypto/evp/libcrypto-lib-e_old.o" => [
            "crypto/evp/e_old.c"
        ],
        "crypto/evp/libcrypto-lib-e_rc4.o" => [
            "crypto/evp/e_rc4.c"
        ],
        "crypto/evp/libcrypto-lib-e_rc4_hmac_md5.o" => [
            "crypto/evp/e_rc4_hmac_md5.c"
        ],
        "crypto/evp/libcrypto-lib-e_rc5.o" => [
            "crypto/evp/e_rc5.c"
        ],
        "crypto/evp/libcrypto-lib-e_sm4.o" => [
            "crypto/evp/e_sm4.c"
        ],
        "crypto/evp/libcrypto-lib-e_wbsm4_baiwu.o" => [
            "crypto/evp/e_wbsm4_baiwu.c"
        ],
        "crypto/evp/libcrypto-lib-e_wbsm4_wsise.o" => [
            "crypto/evp/e_wbsm4_wsise.c"
        ],
        "crypto/evp/libcrypto-lib-e_wbsm4_xiaolai.o" => [
            "crypto/evp/e_wbsm4_xiaolai.c"
        ],
        "crypto/evp/libcrypto-lib-e_xcbc_d.o" => [
            "crypto/evp/e_xcbc_d.c"
        ],
        "crypto/evp/libcrypto-lib-ec_ctrl.o" => [
            "crypto/evp/ec_ctrl.c"
        ],
        "crypto/evp/libcrypto-lib-ec_support.o" => [
            "crypto/evp/ec_support.c"
        ],
        "crypto/evp/libcrypto-lib-encode.o" => [
            "crypto/evp/encode.c"
        ],
        "crypto/evp/libcrypto-lib-evp_cnf.o" => [
            "crypto/evp/evp_cnf.c"
        ],
        "crypto/evp/libcrypto-lib-evp_enc.o" => [
            "crypto/evp/evp_enc.c"
        ],
        "crypto/evp/libcrypto-lib-evp_err.o" => [
            "crypto/evp/evp_err.c"
        ],
        "crypto/evp/libcrypto-lib-evp_fetch.o" => [
            "crypto/evp/evp_fetch.c"
        ],
        "crypto/evp/libcrypto-lib-evp_key.o" => [
            "crypto/evp/evp_key.c"
        ],
        "crypto/evp/libcrypto-lib-evp_lib.o" => [
            "crypto/evp/evp_lib.c"
        ],
        "crypto/evp/libcrypto-lib-evp_pbe.o" => [
            "crypto/evp/evp_pbe.c"
        ],
        "crypto/evp/libcrypto-lib-evp_pkey.o" => [
            "crypto/evp/evp_pkey.c"
        ],
        "crypto/evp/libcrypto-lib-evp_rand.o" => [
            "crypto/evp/evp_rand.c"
        ],
        "crypto/evp/libcrypto-lib-evp_utils.o" => [
            "crypto/evp/evp_utils.c"
        ],
        "crypto/evp/libcrypto-lib-exchange.o" => [
            "crypto/evp/exchange.c"
        ],
        "crypto/evp/libcrypto-lib-kdf_lib.o" => [
            "crypto/evp/kdf_lib.c"
        ],
        "crypto/evp/libcrypto-lib-kdf_meth.o" => [
            "crypto/evp/kdf_meth.c"
        ],
        "crypto/evp/libcrypto-lib-kem.o" => [
            "crypto/evp/kem.c"
        ],
        "crypto/evp/libcrypto-lib-keymgmt_lib.o" => [
            "crypto/evp/keymgmt_lib.c"
        ],
        "crypto/evp/libcrypto-lib-keymgmt_meth.o" => [
            "crypto/evp/keymgmt_meth.c"
        ],
        "crypto/evp/libcrypto-lib-legacy_md5.o" => [
            "crypto/evp/legacy_md5.c"
        ],
        "crypto/evp/libcrypto-lib-legacy_md5_sha1.o" => [
            "crypto/evp/legacy_md5_sha1.c"
        ],
        "crypto/evp/libcrypto-lib-legacy_sha.o" => [
            "crypto/evp/legacy_sha.c"
        ],
        "crypto/evp/libcrypto-lib-m_null.o" => [
            "crypto/evp/m_null.c"
        ],
        "crypto/evp/libcrypto-lib-m_sigver.o" => [
            "crypto/evp/m_sigver.c"
        ],
        "crypto/evp/libcrypto-lib-mac_lib.o" => [
            "crypto/evp/mac_lib.c"
        ],
        "crypto/evp/libcrypto-lib-mac_meth.o" => [
            "crypto/evp/mac_meth.c"
        ],
        "crypto/evp/libcrypto-lib-names.o" => [
            "crypto/evp/names.c"
        ],
        "crypto/evp/libcrypto-lib-p5_crpt.o" => [
            "crypto/evp/p5_crpt.c"
        ],
        "crypto/evp/libcrypto-lib-p5_crpt2.o" => [
            "crypto/evp/p5_crpt2.c"
        ],
        "crypto/evp/libcrypto-lib-p_dec.o" => [
            "crypto/evp/p_dec.c"
        ],
        "crypto/evp/libcrypto-lib-p_enc.o" => [
            "crypto/evp/p_enc.c"
        ],
        "crypto/evp/libcrypto-lib-p_legacy.o" => [
            "crypto/evp/p_legacy.c"
        ],
        "crypto/evp/libcrypto-lib-p_lib.o" => [
            "crypto/evp/p_lib.c"
        ],
        "crypto/evp/libcrypto-lib-p_open.o" => [
            "crypto/evp/p_open.c"
        ],
        "crypto/evp/libcrypto-lib-p_seal.o" => [
            "crypto/evp/p_seal.c"
        ],
        "crypto/evp/libcrypto-lib-p_sign.o" => [
            "crypto/evp/p_sign.c"
        ],
        "crypto/evp/libcrypto-lib-p_verify.o" => [
            "crypto/evp/p_verify.c"
        ],
        "crypto/evp/libcrypto-lib-pbe_scrypt.o" => [
            "crypto/evp/pbe_scrypt.c"
        ],
        "crypto/evp/libcrypto-lib-pmeth_check.o" => [
            "crypto/evp/pmeth_check.c"
        ],
        "crypto/evp/libcrypto-lib-pmeth_gn.o" => [
            "crypto/evp/pmeth_gn.c"
        ],
        "crypto/evp/libcrypto-lib-pmeth_lib.o" => [
            "crypto/evp/pmeth_lib.c"
        ],
        "crypto/evp/libcrypto-lib-s_lib.o" => [
            "crypto/evp/s_lib.c"
        ],
        "crypto/evp/libcrypto-lib-signature.o" => [
            "crypto/evp/signature.c"
        ],
        "crypto/evp/libcrypto-lib-skeymgmt_meth.o" => [
            "crypto/evp/skeymgmt_meth.c"
        ],
        "crypto/evp/libcrypto-shlib-asymcipher.o" => [
            "crypto/evp/asymcipher.c"
        ],
        "crypto/evp/libcrypto-shlib-bio_b64.o" => [
            "crypto/evp/bio_b64.c"
        ],
        "crypto/evp/libcrypto-shlib-bio_enc.o" => [
            "crypto/evp/bio_enc.c"
        ],
        "crypto/evp/libcrypto-shlib-bio_md.o" => [
            "crypto/evp/bio_md.c"
        ],
        "crypto/evp/libcrypto-shlib-bio_ok.o" => [
            "crypto/evp/bio_ok.c"
        ],
        "crypto/evp/libcrypto-shlib-c_allc.o" => [
            "crypto/evp/c_allc.c"
        ],
        "crypto/evp/libcrypto-shlib-c_alld.o" => [
            "crypto/evp/c_alld.c"
        ],
        "crypto/evp/libcrypto-shlib-cmeth_lib.o" => [
            "crypto/evp/cmeth_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-ctrl_params_translate.o" => [
            "crypto/evp/ctrl_params_translate.c"
        ],
        "crypto/evp/libcrypto-shlib-dh_ctrl.o" => [
            "crypto/evp/dh_ctrl.c"
        ],
        "crypto/evp/libcrypto-shlib-dh_support.o" => [
            "crypto/evp/dh_support.c"
        ],
        "crypto/evp/libcrypto-shlib-digest.o" => [
            "crypto/evp/digest.c"
        ],
        "crypto/evp/libcrypto-shlib-dsa_ctrl.o" => [
            "crypto/evp/dsa_ctrl.c"
        ],
        "crypto/evp/libcrypto-shlib-e_aes.o" => [
            "crypto/evp/e_aes.c"
        ],
        "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha1.o" => [
            "crypto/evp/e_aes_cbc_hmac_sha1.c"
        ],
        "crypto/evp/libcrypto-shlib-e_aes_cbc_hmac_sha256.o" => [
            "crypto/evp/e_aes_cbc_hmac_sha256.c"
        ],
        "crypto/evp/libcrypto-shlib-e_chacha20_poly1305.o" => [
            "crypto/evp/e_chacha20_poly1305.c"
        ],
        "crypto/evp/libcrypto-shlib-e_des.o" => [
            "crypto/evp/e_des.c"
        ],
        "crypto/evp/libcrypto-shlib-e_des3.o" => [
            "crypto/evp/e_des3.c"
        ],
        "crypto/evp/libcrypto-shlib-e_eea3.o" => [
            "crypto/evp/e_eea3.c"
        ],
        "crypto/evp/libcrypto-shlib-e_null.o" => [
            "crypto/evp/e_null.c"
        ],
        "crypto/evp/libcrypto-shlib-e_old.o" => [
            "crypto/evp/e_old.c"
        ],
        "crypto/evp/libcrypto-shlib-e_rc4.o" => [
            "crypto/evp/e_rc4.c"
        ],
        "crypto/evp/libcrypto-shlib-e_rc4_hmac_md5.o" => [
            "crypto/evp/e_rc4_hmac_md5.c"
        ],
        "crypto/evp/libcrypto-shlib-e_rc5.o" => [
            "crypto/evp/e_rc5.c"
        ],
        "crypto/evp/libcrypto-shlib-e_sm4.o" => [
            "crypto/evp/e_sm4.c"
        ],
        "crypto/evp/libcrypto-shlib-e_wbsm4_baiwu.o" => [
            "crypto/evp/e_wbsm4_baiwu.c"
        ],
        "crypto/evp/libcrypto-shlib-e_wbsm4_wsise.o" => [
            "crypto/evp/e_wbsm4_wsise.c"
        ],
        "crypto/evp/libcrypto-shlib-e_wbsm4_xiaolai.o" => [
            "crypto/evp/e_wbsm4_xiaolai.c"
        ],
        "crypto/evp/libcrypto-shlib-e_xcbc_d.o" => [
            "crypto/evp/e_xcbc_d.c"
        ],
        "crypto/evp/libcrypto-shlib-ec_ctrl.o" => [
            "crypto/evp/ec_ctrl.c"
        ],
        "crypto/evp/libcrypto-shlib-ec_support.o" => [
            "crypto/evp/ec_support.c"
        ],
        "crypto/evp/libcrypto-shlib-encode.o" => [
            "crypto/evp/encode.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_cnf.o" => [
            "crypto/evp/evp_cnf.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_enc.o" => [
            "crypto/evp/evp_enc.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_err.o" => [
            "crypto/evp/evp_err.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_fetch.o" => [
            "crypto/evp/evp_fetch.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_key.o" => [
            "crypto/evp/evp_key.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_lib.o" => [
            "crypto/evp/evp_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_pbe.o" => [
            "crypto/evp/evp_pbe.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_pkey.o" => [
            "crypto/evp/evp_pkey.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_rand.o" => [
            "crypto/evp/evp_rand.c"
        ],
        "crypto/evp/libcrypto-shlib-evp_utils.o" => [
            "crypto/evp/evp_utils.c"
        ],
        "crypto/evp/libcrypto-shlib-exchange.o" => [
            "crypto/evp/exchange.c"
        ],
        "crypto/evp/libcrypto-shlib-kdf_lib.o" => [
            "crypto/evp/kdf_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-kdf_meth.o" => [
            "crypto/evp/kdf_meth.c"
        ],
        "crypto/evp/libcrypto-shlib-kem.o" => [
            "crypto/evp/kem.c"
        ],
        "crypto/evp/libcrypto-shlib-keymgmt_lib.o" => [
            "crypto/evp/keymgmt_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-keymgmt_meth.o" => [
            "crypto/evp/keymgmt_meth.c"
        ],
        "crypto/evp/libcrypto-shlib-legacy_md5.o" => [
            "crypto/evp/legacy_md5.c"
        ],
        "crypto/evp/libcrypto-shlib-legacy_md5_sha1.o" => [
            "crypto/evp/legacy_md5_sha1.c"
        ],
        "crypto/evp/libcrypto-shlib-legacy_sha.o" => [
            "crypto/evp/legacy_sha.c"
        ],
        "crypto/evp/libcrypto-shlib-m_null.o" => [
            "crypto/evp/m_null.c"
        ],
        "crypto/evp/libcrypto-shlib-m_sigver.o" => [
            "crypto/evp/m_sigver.c"
        ],
        "crypto/evp/libcrypto-shlib-mac_lib.o" => [
            "crypto/evp/mac_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-mac_meth.o" => [
            "crypto/evp/mac_meth.c"
        ],
        "crypto/evp/libcrypto-shlib-names.o" => [
            "crypto/evp/names.c"
        ],
        "crypto/evp/libcrypto-shlib-p5_crpt.o" => [
            "crypto/evp/p5_crpt.c"
        ],
        "crypto/evp/libcrypto-shlib-p5_crpt2.o" => [
            "crypto/evp/p5_crpt2.c"
        ],
        "crypto/evp/libcrypto-shlib-p_dec.o" => [
            "crypto/evp/p_dec.c"
        ],
        "crypto/evp/libcrypto-shlib-p_enc.o" => [
            "crypto/evp/p_enc.c"
        ],
        "crypto/evp/libcrypto-shlib-p_legacy.o" => [
            "crypto/evp/p_legacy.c"
        ],
        "crypto/evp/libcrypto-shlib-p_lib.o" => [
            "crypto/evp/p_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-p_open.o" => [
            "crypto/evp/p_open.c"
        ],
        "crypto/evp/libcrypto-shlib-p_seal.o" => [
            "crypto/evp/p_seal.c"
        ],
        "crypto/evp/libcrypto-shlib-p_sign.o" => [
            "crypto/evp/p_sign.c"
        ],
        "crypto/evp/libcrypto-shlib-p_verify.o" => [
            "crypto/evp/p_verify.c"
        ],
        "crypto/evp/libcrypto-shlib-pbe_scrypt.o" => [
            "crypto/evp/pbe_scrypt.c"
        ],
        "crypto/evp/libcrypto-shlib-pmeth_check.o" => [
            "crypto/evp/pmeth_check.c"
        ],
        "crypto/evp/libcrypto-shlib-pmeth_gn.o" => [
            "crypto/evp/pmeth_gn.c"
        ],
        "crypto/evp/libcrypto-shlib-pmeth_lib.o" => [
            "crypto/evp/pmeth_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-s_lib.o" => [
            "crypto/evp/s_lib.c"
        ],
        "crypto/evp/libcrypto-shlib-signature.o" => [
            "crypto/evp/signature.c"
        ],
        "crypto/evp/libcrypto-shlib-skeymgmt_meth.o" => [
            "crypto/evp/skeymgmt_meth.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_backend.o" => [
            "crypto/ffc/ffc_backend.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_dh.o" => [
            "crypto/ffc/ffc_dh.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_key_generate.o" => [
            "crypto/ffc/ffc_key_generate.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_key_validate.o" => [
            "crypto/ffc/ffc_key_validate.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_params.o" => [
            "crypto/ffc/ffc_params.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_params_generate.o" => [
            "crypto/ffc/ffc_params_generate.c"
        ],
        "crypto/ffc/libcrypto-lib-ffc_params_validate.o" => [
            "crypto/ffc/ffc_params_validate.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_backend.o" => [
            "crypto/ffc/ffc_backend.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_dh.o" => [
            "crypto/ffc/ffc_dh.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_key_generate.o" => [
            "crypto/ffc/ffc_key_generate.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_key_validate.o" => [
            "crypto/ffc/ffc_key_validate.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_params.o" => [
            "crypto/ffc/ffc_params.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_params_generate.o" => [
            "crypto/ffc/ffc_params_generate.c"
        ],
        "crypto/ffc/libcrypto-shlib-ffc_params_validate.o" => [
            "crypto/ffc/ffc_params_validate.c"
        ],
        "crypto/hashtable/libcrypto-lib-hashfunc.o" => [
            "crypto/hashtable/hashfunc.c"
        ],
        "crypto/hashtable/libcrypto-lib-hashtable.o" => [
            "crypto/hashtable/hashtable.c"
        ],
        "crypto/hashtable/libcrypto-shlib-hashfunc.o" => [
            "crypto/hashtable/hashfunc.c"
        ],
        "crypto/hashtable/libcrypto-shlib-hashtable.o" => [
            "crypto/hashtable/hashtable.c"
        ],
        "crypto/hashtable/libssl-shlib-hashfunc.o" => [
            "crypto/hashtable/hashfunc.c"
        ],
        "crypto/hmac/libcrypto-lib-hmac.o" => [
            "crypto/hmac/hmac.c"
        ],
        "crypto/hmac/libcrypto-shlib-hmac.o" => [
            "crypto/hmac/hmac.c"
        ],
        "crypto/hpke/libcrypto-lib-hpke.o" => [
            "crypto/hpke/hpke.c"
        ],
        "crypto/hpke/libcrypto-lib-hpke_util.o" => [
            "crypto/hpke/hpke_util.c"
        ],
        "crypto/hpke/libcrypto-shlib-hpke.o" => [
            "crypto/hpke/hpke.c"
        ],
        "crypto/hpke/libcrypto-shlib-hpke_util.o" => [
            "crypto/hpke/hpke_util.c"
        ],
        "crypto/http/libcrypto-lib-http_client.o" => [
            "crypto/http/http_client.c"
        ],
        "crypto/http/libcrypto-lib-http_err.o" => [
            "crypto/http/http_err.c"
        ],
        "crypto/http/libcrypto-lib-http_lib.o" => [
            "crypto/http/http_lib.c"
        ],
        "crypto/http/libcrypto-shlib-http_client.o" => [
            "crypto/http/http_client.c"
        ],
        "crypto/http/libcrypto-shlib-http_err.o" => [
            "crypto/http/http_err.c"
        ],
        "crypto/http/libcrypto-shlib-http_lib.o" => [
            "crypto/http/http_lib.c"
        ],
        "crypto/kdf/libcrypto-lib-kdf_err.o" => [
            "crypto/kdf/kdf_err.c"
        ],
        "crypto/kdf/libcrypto-shlib-kdf_err.o" => [
            "crypto/kdf/kdf_err.c"
        ],
        "crypto/legacy-dso-cpuid.o" => [
            "crypto/cpuid.c"
        ],
        "crypto/legacy-dso-ctype.o" => [
            "crypto/ctype.c"
        ],
        "crypto/legacy-dso-x86_64cpuid.o" => [
            "crypto/x86_64cpuid.s"
        ],
        "crypto/lhash/libcrypto-lib-lh_stats.o" => [
            "crypto/lhash/lh_stats.c"
        ],
        "crypto/lhash/libcrypto-lib-lhash.o" => [
            "crypto/lhash/lhash.c"
        ],
        "crypto/lhash/libcrypto-shlib-lh_stats.o" => [
            "crypto/lhash/lh_stats.c"
        ],
        "crypto/lhash/libcrypto-shlib-lhash.o" => [
            "crypto/lhash/lhash.c"
        ],
        "crypto/libcrypto-lib-asn1_dsa.o" => [
            "crypto/asn1_dsa.c"
        ],
        "crypto/libcrypto-lib-bsearch.o" => [
            "crypto/bsearch.c"
        ],
        "crypto/libcrypto-lib-comp_methods.o" => [
            "crypto/comp_methods.c"
        ],
        "crypto/libcrypto-lib-context.o" => [
            "crypto/context.c"
        ],
        "crypto/libcrypto-lib-core_algorithm.o" => [
            "crypto/core_algorithm.c"
        ],
        "crypto/libcrypto-lib-core_fetch.o" => [
            "crypto/core_fetch.c"
        ],
        "crypto/libcrypto-lib-core_namemap.o" => [
            "crypto/core_namemap.c"
        ],
        "crypto/libcrypto-lib-cpt_err.o" => [
            "crypto/cpt_err.c"
        ],
        "crypto/libcrypto-lib-cpuid.o" => [
            "crypto/cpuid.c"
        ],
        "crypto/libcrypto-lib-cryptlib.o" => [
            "crypto/cryptlib.c"
        ],
        "crypto/libcrypto-lib-ctype.o" => [
            "crypto/ctype.c"
        ],
        "crypto/libcrypto-lib-cversion.o" => [
            "crypto/cversion.c"
        ],
        "crypto/libcrypto-lib-defaults.o" => [
            "crypto/defaults.c"
        ],
        "crypto/libcrypto-lib-der_writer.o" => [
            "crypto/der_writer.c"
        ],
        "crypto/libcrypto-lib-deterministic_nonce.o" => [
            "crypto/deterministic_nonce.c"
        ],
        "crypto/libcrypto-lib-ebcdic.o" => [
            "crypto/ebcdic.c"
        ],
        "crypto/libcrypto-lib-ex_data.o" => [
            "crypto/ex_data.c"
        ],
        "crypto/libcrypto-lib-getenv.o" => [
            "crypto/getenv.c"
        ],
        "crypto/libcrypto-lib-indicator_core.o" => [
            "crypto/indicator_core.c"
        ],
        "crypto/libcrypto-lib-info.o" => [
            "crypto/info.c"
        ],
        "crypto/libcrypto-lib-init.o" => [
            "crypto/init.c"
        ],
        "crypto/libcrypto-lib-initthread.o" => [
            "crypto/initthread.c"
        ],
        "crypto/libcrypto-lib-mem.o" => [
            "crypto/mem.c"
        ],
        "crypto/libcrypto-lib-mem_sec.o" => [
            "crypto/mem_sec.c"
        ],
        "crypto/libcrypto-lib-o_dir.o" => [
            "crypto/o_dir.c"
        ],
        "crypto/libcrypto-lib-o_fopen.o" => [
            "crypto/o_fopen.c"
        ],
        "crypto/libcrypto-lib-o_init.o" => [
            "crypto/o_init.c"
        ],
        "crypto/libcrypto-lib-o_str.o" => [
            "crypto/o_str.c"
        ],
        "crypto/libcrypto-lib-o_syslog.o" => [
            "crypto/o_syslog.c"
        ],
        "crypto/libcrypto-lib-o_time.o" => [
            "crypto/o_time.c"
        ],
        "crypto/libcrypto-lib-packet.o" => [
            "crypto/packet.c"
        ],
        "crypto/libcrypto-lib-param_build.o" => [
            "crypto/param_build.c"
        ],
        "crypto/libcrypto-lib-param_build_set.o" => [
            "crypto/param_build_set.c"
        ],
        "crypto/libcrypto-lib-params.o" => [
            "crypto/params.c"
        ],
        "crypto/libcrypto-lib-params_dup.o" => [
            "crypto/params_dup.c"
        ],
        "crypto/libcrypto-lib-params_from_text.o" => [
            "crypto/params_from_text.c"
        ],
        "crypto/libcrypto-lib-params_idx.o" => [
            "crypto/params_idx.c"
        ],
        "crypto/libcrypto-lib-passphrase.o" => [
            "crypto/passphrase.c"
        ],
        "crypto/libcrypto-lib-provider.o" => [
            "crypto/provider.c"
        ],
        "crypto/libcrypto-lib-provider_child.o" => [
            "crypto/provider_child.c"
        ],
        "crypto/libcrypto-lib-provider_conf.o" => [
            "crypto/provider_conf.c"
        ],
        "crypto/libcrypto-lib-provider_core.o" => [
            "crypto/provider_core.c"
        ],
        "crypto/libcrypto-lib-provider_predefined.o" => [
            "crypto/provider_predefined.c"
        ],
        "crypto/libcrypto-lib-punycode.o" => [
            "crypto/punycode.c"
        ],
        "crypto/libcrypto-lib-quic_vlint.o" => [
            "crypto/quic_vlint.c"
        ],
        "crypto/libcrypto-lib-self_test_core.o" => [
            "crypto/self_test_core.c"
        ],
        "crypto/libcrypto-lib-sleep.o" => [
            "crypto/sleep.c"
        ],
        "crypto/libcrypto-lib-sparse_array.o" => [
            "crypto/sparse_array.c"
        ],
        "crypto/libcrypto-lib-ssl_err.o" => [
            "crypto/ssl_err.c"
        ],
        "crypto/libcrypto-lib-threads_lib.o" => [
            "crypto/threads_lib.c"
        ],
        "crypto/libcrypto-lib-threads_none.o" => [
            "crypto/threads_none.c"
        ],
        "crypto/libcrypto-lib-threads_pthread.o" => [
            "crypto/threads_pthread.c"
        ],
        "crypto/libcrypto-lib-threads_win.o" => [
            "crypto/threads_win.c"
        ],
        "crypto/libcrypto-lib-time.o" => [
            "crypto/time.c"
        ],
        "crypto/libcrypto-lib-trace.o" => [
            "crypto/trace.c"
        ],
        "crypto/libcrypto-lib-uid.o" => [
            "crypto/uid.c"
        ],
        "crypto/libcrypto-lib-x86_64cpuid.o" => [
            "crypto/x86_64cpuid.s"
        ],
        "crypto/libcrypto-shlib-asn1_dsa.o" => [
            "crypto/asn1_dsa.c"
        ],
        "crypto/libcrypto-shlib-bsearch.o" => [
            "crypto/bsearch.c"
        ],
        "crypto/libcrypto-shlib-comp_methods.o" => [
            "crypto/comp_methods.c"
        ],
        "crypto/libcrypto-shlib-context.o" => [
            "crypto/context.c"
        ],
        "crypto/libcrypto-shlib-core_algorithm.o" => [
            "crypto/core_algorithm.c"
        ],
        "crypto/libcrypto-shlib-core_fetch.o" => [
            "crypto/core_fetch.c"
        ],
        "crypto/libcrypto-shlib-core_namemap.o" => [
            "crypto/core_namemap.c"
        ],
        "crypto/libcrypto-shlib-cpt_err.o" => [
            "crypto/cpt_err.c"
        ],
        "crypto/libcrypto-shlib-cpuid.o" => [
            "crypto/cpuid.c"
        ],
        "crypto/libcrypto-shlib-cryptlib.o" => [
            "crypto/cryptlib.c"
        ],
        "crypto/libcrypto-shlib-ctype.o" => [
            "crypto/ctype.c"
        ],
        "crypto/libcrypto-shlib-cversion.o" => [
            "crypto/cversion.c"
        ],
        "crypto/libcrypto-shlib-defaults.o" => [
            "crypto/defaults.c"
        ],
        "crypto/libcrypto-shlib-der_writer.o" => [
            "crypto/der_writer.c"
        ],
        "crypto/libcrypto-shlib-deterministic_nonce.o" => [
            "crypto/deterministic_nonce.c"
        ],
        "crypto/libcrypto-shlib-ebcdic.o" => [
            "crypto/ebcdic.c"
        ],
        "crypto/libcrypto-shlib-ex_data.o" => [
            "crypto/ex_data.c"
        ],
        "crypto/libcrypto-shlib-getenv.o" => [
            "crypto/getenv.c"
        ],
        "crypto/libcrypto-shlib-indicator_core.o" => [
            "crypto/indicator_core.c"
        ],
        "crypto/libcrypto-shlib-info.o" => [
            "crypto/info.c"
        ],
        "crypto/libcrypto-shlib-init.o" => [
            "crypto/init.c"
        ],
        "crypto/libcrypto-shlib-initthread.o" => [
            "crypto/initthread.c"
        ],
        "crypto/libcrypto-shlib-mem.o" => [
            "crypto/mem.c"
        ],
        "crypto/libcrypto-shlib-mem_sec.o" => [
            "crypto/mem_sec.c"
        ],
        "crypto/libcrypto-shlib-o_dir.o" => [
            "crypto/o_dir.c"
        ],
        "crypto/libcrypto-shlib-o_fopen.o" => [
            "crypto/o_fopen.c"
        ],
        "crypto/libcrypto-shlib-o_init.o" => [
            "crypto/o_init.c"
        ],
        "crypto/libcrypto-shlib-o_str.o" => [
            "crypto/o_str.c"
        ],
        "crypto/libcrypto-shlib-o_syslog.o" => [
            "crypto/o_syslog.c"
        ],
        "crypto/libcrypto-shlib-o_time.o" => [
            "crypto/o_time.c"
        ],
        "crypto/libcrypto-shlib-packet.o" => [
            "crypto/packet.c"
        ],
        "crypto/libcrypto-shlib-param_build.o" => [
            "crypto/param_build.c"
        ],
        "crypto/libcrypto-shlib-param_build_set.o" => [
            "crypto/param_build_set.c"
        ],
        "crypto/libcrypto-shlib-params.o" => [
            "crypto/params.c"
        ],
        "crypto/libcrypto-shlib-params_dup.o" => [
            "crypto/params_dup.c"
        ],
        "crypto/libcrypto-shlib-params_from_text.o" => [
            "crypto/params_from_text.c"
        ],
        "crypto/libcrypto-shlib-params_idx.o" => [
            "crypto/params_idx.c"
        ],
        "crypto/libcrypto-shlib-passphrase.o" => [
            "crypto/passphrase.c"
        ],
        "crypto/libcrypto-shlib-provider.o" => [
            "crypto/provider.c"
        ],
        "crypto/libcrypto-shlib-provider_child.o" => [
            "crypto/provider_child.c"
        ],
        "crypto/libcrypto-shlib-provider_conf.o" => [
            "crypto/provider_conf.c"
        ],
        "crypto/libcrypto-shlib-provider_core.o" => [
            "crypto/provider_core.c"
        ],
        "crypto/libcrypto-shlib-provider_predefined.o" => [
            "crypto/provider_predefined.c"
        ],
        "crypto/libcrypto-shlib-punycode.o" => [
            "crypto/punycode.c"
        ],
        "crypto/libcrypto-shlib-quic_vlint.o" => [
            "crypto/quic_vlint.c"
        ],
        "crypto/libcrypto-shlib-self_test_core.o" => [
            "crypto/self_test_core.c"
        ],
        "crypto/libcrypto-shlib-sleep.o" => [
            "crypto/sleep.c"
        ],
        "crypto/libcrypto-shlib-sparse_array.o" => [
            "crypto/sparse_array.c"
        ],
        "crypto/libcrypto-shlib-ssl_err.o" => [
            "crypto/ssl_err.c"
        ],
        "crypto/libcrypto-shlib-threads_lib.o" => [
            "crypto/threads_lib.c"
        ],
        "crypto/libcrypto-shlib-threads_none.o" => [
            "crypto/threads_none.c"
        ],
        "crypto/libcrypto-shlib-threads_pthread.o" => [
            "crypto/threads_pthread.c"
        ],
        "crypto/libcrypto-shlib-threads_win.o" => [
            "crypto/threads_win.c"
        ],
        "crypto/libcrypto-shlib-time.o" => [
            "crypto/time.c"
        ],
        "crypto/libcrypto-shlib-trace.o" => [
            "crypto/trace.c"
        ],
        "crypto/libcrypto-shlib-uid.o" => [
            "crypto/uid.c"
        ],
        "crypto/libcrypto-shlib-x86_64cpuid.o" => [
            "crypto/x86_64cpuid.s"
        ],
        "crypto/libssl-shlib-ctype.o" => [
            "crypto/ctype.c"
        ],
        "crypto/libssl-shlib-getenv.o" => [
            "crypto/getenv.c"
        ],
        "crypto/libssl-shlib-packet.o" => [
            "crypto/packet.c"
        ],
        "crypto/libssl-shlib-quic_vlint.o" => [
            "crypto/quic_vlint.c"
        ],
        "crypto/libssl-shlib-time.o" => [
            "crypto/time.c"
        ],
        "crypto/md5/libcrypto-lib-md5-x86_64.o" => [
            "crypto/md5/md5-x86_64.s"
        ],
        "crypto/md5/libcrypto-lib-md5_dgst.o" => [
            "crypto/md5/md5_dgst.c"
        ],
        "crypto/md5/libcrypto-lib-md5_one.o" => [
            "crypto/md5/md5_one.c"
        ],
        "crypto/md5/libcrypto-lib-md5_sha1.o" => [
            "crypto/md5/md5_sha1.c"
        ],
        "crypto/md5/libcrypto-shlib-md5-x86_64.o" => [
            "crypto/md5/md5-x86_64.s"
        ],
        "crypto/md5/libcrypto-shlib-md5_dgst.o" => [
            "crypto/md5/md5_dgst.c"
        ],
        "crypto/md5/libcrypto-shlib-md5_one.o" => [
            "crypto/md5/md5_one.c"
        ],
        "crypto/md5/libcrypto-shlib-md5_sha1.o" => [
            "crypto/md5/md5_sha1.c"
        ],
        "crypto/md5/liblegacy-lib-md5-x86_64.o" => [
            "crypto/md5/md5-x86_64.s"
        ],
        "crypto/md5/liblegacy-lib-md5_dgst.o" => [
            "crypto/md5/md5_dgst.c"
        ],
        "crypto/md5/liblegacy-lib-md5_one.o" => [
            "crypto/md5/md5_one.c"
        ],
        "crypto/md5/liblegacy-lib-md5_sha1.o" => [
            "crypto/md5/md5_sha1.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_encoders.o" => [
            "crypto/ml_dsa/ml_dsa_encoders.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_key.o" => [
            "crypto/ml_dsa/ml_dsa_key.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_key_compress.o" => [
            "crypto/ml_dsa/ml_dsa_key_compress.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_matrix.o" => [
            "crypto/ml_dsa/ml_dsa_matrix.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_ntt.o" => [
            "crypto/ml_dsa/ml_dsa_ntt.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_params.o" => [
            "crypto/ml_dsa/ml_dsa_params.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_sample.o" => [
            "crypto/ml_dsa/ml_dsa_sample.c"
        ],
        "crypto/ml_dsa/libcrypto-lib-ml_dsa_sign.o" => [
            "crypto/ml_dsa/ml_dsa_sign.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_encoders.o" => [
            "crypto/ml_dsa/ml_dsa_encoders.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key.o" => [
            "crypto/ml_dsa/ml_dsa_key.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_key_compress.o" => [
            "crypto/ml_dsa/ml_dsa_key_compress.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_matrix.o" => [
            "crypto/ml_dsa/ml_dsa_matrix.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_ntt.o" => [
            "crypto/ml_dsa/ml_dsa_ntt.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_params.o" => [
            "crypto/ml_dsa/ml_dsa_params.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sample.o" => [
            "crypto/ml_dsa/ml_dsa_sample.c"
        ],
        "crypto/ml_dsa/libcrypto-shlib-ml_dsa_sign.o" => [
            "crypto/ml_dsa/ml_dsa_sign.c"
        ],
        "crypto/ml_kem/libcrypto-lib-ml_kem.o" => [
            "crypto/ml_kem/ml_kem.c"
        ],
        "crypto/ml_kem/libcrypto-shlib-ml_kem.o" => [
            "crypto/ml_kem/ml_kem.c"
        ],
        "crypto/modes/libcrypto-lib-aes-gcm-avx512.o" => [
            "crypto/modes/aes-gcm-avx512.s"
        ],
        "crypto/modes/libcrypto-lib-aesni-gcm-x86_64.o" => [
            "crypto/modes/aesni-gcm-x86_64.s"
        ],
        "crypto/modes/libcrypto-lib-cbc128.o" => [
            "crypto/modes/cbc128.c"
        ],
        "crypto/modes/libcrypto-lib-ccm128.o" => [
            "crypto/modes/ccm128.c"
        ],
        "crypto/modes/libcrypto-lib-cfb128.o" => [
            "crypto/modes/cfb128.c"
        ],
        "crypto/modes/libcrypto-lib-ctr128.o" => [
            "crypto/modes/ctr128.c"
        ],
        "crypto/modes/libcrypto-lib-cts128.o" => [
            "crypto/modes/cts128.c"
        ],
        "crypto/modes/libcrypto-lib-gcm128.o" => [
            "crypto/modes/gcm128.c"
        ],
        "crypto/modes/libcrypto-lib-ghash-x86_64.o" => [
            "crypto/modes/ghash-x86_64.s"
        ],
        "crypto/modes/libcrypto-lib-ocb128.o" => [
            "crypto/modes/ocb128.c"
        ],
        "crypto/modes/libcrypto-lib-ofb128.o" => [
            "crypto/modes/ofb128.c"
        ],
        "crypto/modes/libcrypto-lib-siv128.o" => [
            "crypto/modes/siv128.c"
        ],
        "crypto/modes/libcrypto-lib-wrap128.o" => [
            "crypto/modes/wrap128.c"
        ],
        "crypto/modes/libcrypto-lib-xts128.o" => [
            "crypto/modes/xts128.c"
        ],
        "crypto/modes/libcrypto-lib-xts128gb.o" => [
            "crypto/modes/xts128gb.c"
        ],
        "crypto/modes/libcrypto-shlib-aes-gcm-avx512.o" => [
            "crypto/modes/aes-gcm-avx512.s"
        ],
        "crypto/modes/libcrypto-shlib-aesni-gcm-x86_64.o" => [
            "crypto/modes/aesni-gcm-x86_64.s"
        ],
        "crypto/modes/libcrypto-shlib-cbc128.o" => [
            "crypto/modes/cbc128.c"
        ],
        "crypto/modes/libcrypto-shlib-ccm128.o" => [
            "crypto/modes/ccm128.c"
        ],
        "crypto/modes/libcrypto-shlib-cfb128.o" => [
            "crypto/modes/cfb128.c"
        ],
        "crypto/modes/libcrypto-shlib-ctr128.o" => [
            "crypto/modes/ctr128.c"
        ],
        "crypto/modes/libcrypto-shlib-cts128.o" => [
            "crypto/modes/cts128.c"
        ],
        "crypto/modes/libcrypto-shlib-gcm128.o" => [
            "crypto/modes/gcm128.c"
        ],
        "crypto/modes/libcrypto-shlib-ghash-x86_64.o" => [
            "crypto/modes/ghash-x86_64.s"
        ],
        "crypto/modes/libcrypto-shlib-ocb128.o" => [
            "crypto/modes/ocb128.c"
        ],
        "crypto/modes/libcrypto-shlib-ofb128.o" => [
            "crypto/modes/ofb128.c"
        ],
        "crypto/modes/libcrypto-shlib-siv128.o" => [
            "crypto/modes/siv128.c"
        ],
        "crypto/modes/libcrypto-shlib-wrap128.o" => [
            "crypto/modes/wrap128.c"
        ],
        "crypto/modes/libcrypto-shlib-xts128.o" => [
            "crypto/modes/xts128.c"
        ],
        "crypto/modes/libcrypto-shlib-xts128gb.o" => [
            "crypto/modes/xts128gb.c"
        ],
        "crypto/objects/libcrypto-lib-o_names.o" => [
            "crypto/objects/o_names.c"
        ],
        "crypto/objects/libcrypto-lib-obj_dat.o" => [
            "crypto/objects/obj_dat.c"
        ],
        "crypto/objects/libcrypto-lib-obj_err.o" => [
            "crypto/objects/obj_err.c"
        ],
        "crypto/objects/libcrypto-lib-obj_lib.o" => [
            "crypto/objects/obj_lib.c"
        ],
        "crypto/objects/libcrypto-lib-obj_xref.o" => [
            "crypto/objects/obj_xref.c"
        ],
        "crypto/objects/libcrypto-shlib-o_names.o" => [
            "crypto/objects/o_names.c"
        ],
        "crypto/objects/libcrypto-shlib-obj_dat.o" => [
            "crypto/objects/obj_dat.c"
        ],
        "crypto/objects/libcrypto-shlib-obj_err.o" => [
            "crypto/objects/obj_err.c"
        ],
        "crypto/objects/libcrypto-shlib-obj_lib.o" => [
            "crypto/objects/obj_lib.c"
        ],
        "crypto/objects/libcrypto-shlib-obj_xref.o" => [
            "crypto/objects/obj_xref.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_asn.o" => [
            "crypto/ocsp/ocsp_asn.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_cl.o" => [
            "crypto/ocsp/ocsp_cl.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_err.o" => [
            "crypto/ocsp/ocsp_err.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_ext.o" => [
            "crypto/ocsp/ocsp_ext.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_http.o" => [
            "crypto/ocsp/ocsp_http.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_lib.o" => [
            "crypto/ocsp/ocsp_lib.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_prn.o" => [
            "crypto/ocsp/ocsp_prn.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_srv.o" => [
            "crypto/ocsp/ocsp_srv.c"
        ],
        "crypto/ocsp/libcrypto-lib-ocsp_vfy.o" => [
            "crypto/ocsp/ocsp_vfy.c"
        ],
        "crypto/ocsp/libcrypto-lib-v3_ocsp.o" => [
            "crypto/ocsp/v3_ocsp.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_asn.o" => [
            "crypto/ocsp/ocsp_asn.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_cl.o" => [
            "crypto/ocsp/ocsp_cl.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_err.o" => [
            "crypto/ocsp/ocsp_err.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_ext.o" => [
            "crypto/ocsp/ocsp_ext.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_http.o" => [
            "crypto/ocsp/ocsp_http.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_lib.o" => [
            "crypto/ocsp/ocsp_lib.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_prn.o" => [
            "crypto/ocsp/ocsp_prn.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_srv.o" => [
            "crypto/ocsp/ocsp_srv.c"
        ],
        "crypto/ocsp/libcrypto-shlib-ocsp_vfy.o" => [
            "crypto/ocsp/ocsp_vfy.c"
        ],
        "crypto/ocsp/libcrypto-shlib-v3_ocsp.o" => [
            "crypto/ocsp/v3_ocsp.c"
        ],
        "crypto/pem/libcrypto-lib-pem_all.o" => [
            "crypto/pem/pem_all.c"
        ],
        "crypto/pem/libcrypto-lib-pem_err.o" => [
            "crypto/pem/pem_err.c"
        ],
        "crypto/pem/libcrypto-lib-pem_info.o" => [
            "crypto/pem/pem_info.c"
        ],
        "crypto/pem/libcrypto-lib-pem_lib.o" => [
            "crypto/pem/pem_lib.c"
        ],
        "crypto/pem/libcrypto-lib-pem_oth.o" => [
            "crypto/pem/pem_oth.c"
        ],
        "crypto/pem/libcrypto-lib-pem_pk8.o" => [
            "crypto/pem/pem_pk8.c"
        ],
        "crypto/pem/libcrypto-lib-pem_pkey.o" => [
            "crypto/pem/pem_pkey.c"
        ],
        "crypto/pem/libcrypto-lib-pem_sign.o" => [
            "crypto/pem/pem_sign.c"
        ],
        "crypto/pem/libcrypto-lib-pem_x509.o" => [
            "crypto/pem/pem_x509.c"
        ],
        "crypto/pem/libcrypto-lib-pem_xaux.o" => [
            "crypto/pem/pem_xaux.c"
        ],
        "crypto/pem/libcrypto-lib-pvkfmt.o" => [
            "crypto/pem/pvkfmt.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_all.o" => [
            "crypto/pem/pem_all.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_err.o" => [
            "crypto/pem/pem_err.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_info.o" => [
            "crypto/pem/pem_info.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_lib.o" => [
            "crypto/pem/pem_lib.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_oth.o" => [
            "crypto/pem/pem_oth.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_pk8.o" => [
            "crypto/pem/pem_pk8.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_pkey.o" => [
            "crypto/pem/pem_pkey.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_sign.o" => [
            "crypto/pem/pem_sign.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_x509.o" => [
            "crypto/pem/pem_x509.c"
        ],
        "crypto/pem/libcrypto-shlib-pem_xaux.o" => [
            "crypto/pem/pem_xaux.c"
        ],
        "crypto/pem/libcrypto-shlib-pvkfmt.o" => [
            "crypto/pem/pvkfmt.c"
        ],
        "crypto/pem/loader_attic-dso-pvkfmt.o" => [
            "crypto/pem/pvkfmt.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_add.o" => [
            "crypto/pkcs12/p12_add.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_asn.o" => [
            "crypto/pkcs12/p12_asn.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_attr.o" => [
            "crypto/pkcs12/p12_attr.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_crpt.o" => [
            "crypto/pkcs12/p12_crpt.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_crt.o" => [
            "crypto/pkcs12/p12_crt.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_decr.o" => [
            "crypto/pkcs12/p12_decr.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_init.o" => [
            "crypto/pkcs12/p12_init.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_key.o" => [
            "crypto/pkcs12/p12_key.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_kiss.o" => [
            "crypto/pkcs12/p12_kiss.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_mutl.o" => [
            "crypto/pkcs12/p12_mutl.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_npas.o" => [
            "crypto/pkcs12/p12_npas.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_p8d.o" => [
            "crypto/pkcs12/p12_p8d.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_p8e.o" => [
            "crypto/pkcs12/p12_p8e.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_sbag.o" => [
            "crypto/pkcs12/p12_sbag.c"
        ],
        "crypto/pkcs12/libcrypto-lib-p12_utl.o" => [
            "crypto/pkcs12/p12_utl.c"
        ],
        "crypto/pkcs12/libcrypto-lib-pk12err.o" => [
            "crypto/pkcs12/pk12err.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_add.o" => [
            "crypto/pkcs12/p12_add.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_asn.o" => [
            "crypto/pkcs12/p12_asn.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_attr.o" => [
            "crypto/pkcs12/p12_attr.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_crpt.o" => [
            "crypto/pkcs12/p12_crpt.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_crt.o" => [
            "crypto/pkcs12/p12_crt.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_decr.o" => [
            "crypto/pkcs12/p12_decr.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_init.o" => [
            "crypto/pkcs12/p12_init.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_key.o" => [
            "crypto/pkcs12/p12_key.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_kiss.o" => [
            "crypto/pkcs12/p12_kiss.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_mutl.o" => [
            "crypto/pkcs12/p12_mutl.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_npas.o" => [
            "crypto/pkcs12/p12_npas.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_p8d.o" => [
            "crypto/pkcs12/p12_p8d.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_p8e.o" => [
            "crypto/pkcs12/p12_p8e.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_sbag.o" => [
            "crypto/pkcs12/p12_sbag.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-p12_utl.o" => [
            "crypto/pkcs12/p12_utl.c"
        ],
        "crypto/pkcs12/libcrypto-shlib-pk12err.o" => [
            "crypto/pkcs12/pk12err.c"
        ],
        "crypto/pkcs7/libcrypto-lib-bio_pk7.o" => [
            "crypto/pkcs7/bio_pk7.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_asn1.o" => [
            "crypto/pkcs7/pk7_asn1.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_attr.o" => [
            "crypto/pkcs7/pk7_attr.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_doit.o" => [
            "crypto/pkcs7/pk7_doit.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_lib.o" => [
            "crypto/pkcs7/pk7_lib.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_mime.o" => [
            "crypto/pkcs7/pk7_mime.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pk7_smime.o" => [
            "crypto/pkcs7/pk7_smime.c"
        ],
        "crypto/pkcs7/libcrypto-lib-pkcs7err.o" => [
            "crypto/pkcs7/pkcs7err.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-bio_pk7.o" => [
            "crypto/pkcs7/bio_pk7.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_asn1.o" => [
            "crypto/pkcs7/pk7_asn1.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_attr.o" => [
            "crypto/pkcs7/pk7_attr.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_doit.o" => [
            "crypto/pkcs7/pk7_doit.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_lib.o" => [
            "crypto/pkcs7/pk7_lib.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_mime.o" => [
            "crypto/pkcs7/pk7_mime.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pk7_smime.o" => [
            "crypto/pkcs7/pk7_smime.c"
        ],
        "crypto/pkcs7/libcrypto-shlib-pkcs7err.o" => [
            "crypto/pkcs7/pkcs7err.c"
        ],
        "crypto/poly1305/libcrypto-lib-poly1305-x86_64.o" => [
            "crypto/poly1305/poly1305-x86_64.s"
        ],
        "crypto/poly1305/libcrypto-lib-poly1305.o" => [
            "crypto/poly1305/poly1305.c"
        ],
        "crypto/poly1305/libcrypto-shlib-poly1305-x86_64.o" => [
            "crypto/poly1305/poly1305-x86_64.s"
        ],
        "crypto/poly1305/libcrypto-shlib-poly1305.o" => [
            "crypto/poly1305/poly1305.c"
        ],
        "crypto/property/libcrypto-lib-defn_cache.o" => [
            "crypto/property/defn_cache.c"
        ],
        "crypto/property/libcrypto-lib-property.o" => [
            "crypto/property/property.c"
        ],
        "crypto/property/libcrypto-lib-property_err.o" => [
            "crypto/property/property_err.c"
        ],
        "crypto/property/libcrypto-lib-property_parse.o" => [
            "crypto/property/property_parse.c"
        ],
        "crypto/property/libcrypto-lib-property_query.o" => [
            "crypto/property/property_query.c"
        ],
        "crypto/property/libcrypto-lib-property_string.o" => [
            "crypto/property/property_string.c"
        ],
        "crypto/property/libcrypto-shlib-defn_cache.o" => [
            "crypto/property/defn_cache.c"
        ],
        "crypto/property/libcrypto-shlib-property.o" => [
            "crypto/property/property.c"
        ],
        "crypto/property/libcrypto-shlib-property_err.o" => [
            "crypto/property/property_err.c"
        ],
        "crypto/property/libcrypto-shlib-property_parse.o" => [
            "crypto/property/property_parse.c"
        ],
        "crypto/property/libcrypto-shlib-property_query.o" => [
            "crypto/property/property_query.c"
        ],
        "crypto/property/libcrypto-shlib-property_string.o" => [
            "crypto/property/property_string.c"
        ],
        "crypto/rand/libcrypto-lib-prov_seed.o" => [
            "crypto/rand/prov_seed.c"
        ],
        "crypto/rand/libcrypto-lib-rand_deprecated.o" => [
            "crypto/rand/rand_deprecated.c"
        ],
        "crypto/rand/libcrypto-lib-rand_err.o" => [
            "crypto/rand/rand_err.c"
        ],
        "crypto/rand/libcrypto-lib-rand_lib.o" => [
            "crypto/rand/rand_lib.c"
        ],
        "crypto/rand/libcrypto-lib-rand_meth.o" => [
            "crypto/rand/rand_meth.c"
        ],
        "crypto/rand/libcrypto-lib-rand_pool.o" => [
            "crypto/rand/rand_pool.c"
        ],
        "crypto/rand/libcrypto-lib-rand_uniform.o" => [
            "crypto/rand/rand_uniform.c"
        ],
        "crypto/rand/libcrypto-lib-randfile.o" => [
            "crypto/rand/randfile.c"
        ],
        "crypto/rand/libcrypto-shlib-prov_seed.o" => [
            "crypto/rand/prov_seed.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_deprecated.o" => [
            "crypto/rand/rand_deprecated.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_err.o" => [
            "crypto/rand/rand_err.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_lib.o" => [
            "crypto/rand/rand_lib.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_meth.o" => [
            "crypto/rand/rand_meth.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_pool.o" => [
            "crypto/rand/rand_pool.c"
        ],
        "crypto/rand/libcrypto-shlib-rand_uniform.o" => [
            "crypto/rand/rand_uniform.c"
        ],
        "crypto/rand/libcrypto-shlib-randfile.o" => [
            "crypto/rand/randfile.c"
        ],
        "crypto/rc4/libcrypto-lib-rc4-md5-x86_64.o" => [
            "crypto/rc4/rc4-md5-x86_64.s"
        ],
        "crypto/rc4/libcrypto-lib-rc4-x86_64.o" => [
            "crypto/rc4/rc4-x86_64.s"
        ],
        "crypto/rc4/libcrypto-shlib-rc4-md5-x86_64.o" => [
            "crypto/rc4/rc4-md5-x86_64.s"
        ],
        "crypto/rc4/libcrypto-shlib-rc4-x86_64.o" => [
            "crypto/rc4/rc4-x86_64.s"
        ],
        "crypto/rc4/liblegacy-lib-rc4-md5-x86_64.o" => [
            "crypto/rc4/rc4-md5-x86_64.s"
        ],
        "crypto/rc4/liblegacy-lib-rc4-x86_64.o" => [
            "crypto/rc4/rc4-x86_64.s"
        ],
        "crypto/rsa/libcrypto-lib-rsa_ameth.o" => [
            "crypto/rsa/rsa_ameth.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_asn1.o" => [
            "crypto/rsa/rsa_asn1.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_backend.o" => [
            "crypto/rsa/rsa_backend.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_chk.o" => [
            "crypto/rsa/rsa_chk.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_crpt.o" => [
            "crypto/rsa/rsa_crpt.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_depr.o" => [
            "crypto/rsa/rsa_depr.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_err.o" => [
            "crypto/rsa/rsa_err.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_gen.o" => [
            "crypto/rsa/rsa_gen.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_lib.o" => [
            "crypto/rsa/rsa_lib.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_meth.o" => [
            "crypto/rsa/rsa_meth.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_mp.o" => [
            "crypto/rsa/rsa_mp.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_mp_names.o" => [
            "crypto/rsa/rsa_mp_names.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_none.o" => [
            "crypto/rsa/rsa_none.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_oaep.o" => [
            "crypto/rsa/rsa_oaep.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_ossl.o" => [
            "crypto/rsa/rsa_ossl.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_pk1.o" => [
            "crypto/rsa/rsa_pk1.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_pmeth.o" => [
            "crypto/rsa/rsa_pmeth.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_prn.o" => [
            "crypto/rsa/rsa_prn.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_pss.o" => [
            "crypto/rsa/rsa_pss.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_saos.o" => [
            "crypto/rsa/rsa_saos.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_schemes.o" => [
            "crypto/rsa/rsa_schemes.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_sign.o" => [
            "crypto/rsa/rsa_sign.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_sp800_56b_check.o" => [
            "crypto/rsa/rsa_sp800_56b_check.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_sp800_56b_gen.o" => [
            "crypto/rsa/rsa_sp800_56b_gen.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_x931.o" => [
            "crypto/rsa/rsa_x931.c"
        ],
        "crypto/rsa/libcrypto-lib-rsa_x931g.o" => [
            "crypto/rsa/rsa_x931g.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_ameth.o" => [
            "crypto/rsa/rsa_ameth.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_asn1.o" => [
            "crypto/rsa/rsa_asn1.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_backend.o" => [
            "crypto/rsa/rsa_backend.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_chk.o" => [
            "crypto/rsa/rsa_chk.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_crpt.o" => [
            "crypto/rsa/rsa_crpt.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_depr.o" => [
            "crypto/rsa/rsa_depr.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_err.o" => [
            "crypto/rsa/rsa_err.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_gen.o" => [
            "crypto/rsa/rsa_gen.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_lib.o" => [
            "crypto/rsa/rsa_lib.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_meth.o" => [
            "crypto/rsa/rsa_meth.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_mp.o" => [
            "crypto/rsa/rsa_mp.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_mp_names.o" => [
            "crypto/rsa/rsa_mp_names.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_none.o" => [
            "crypto/rsa/rsa_none.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_oaep.o" => [
            "crypto/rsa/rsa_oaep.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_ossl.o" => [
            "crypto/rsa/rsa_ossl.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_pk1.o" => [
            "crypto/rsa/rsa_pk1.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_pmeth.o" => [
            "crypto/rsa/rsa_pmeth.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_prn.o" => [
            "crypto/rsa/rsa_prn.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_pss.o" => [
            "crypto/rsa/rsa_pss.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_saos.o" => [
            "crypto/rsa/rsa_saos.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_schemes.o" => [
            "crypto/rsa/rsa_schemes.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_sign.o" => [
            "crypto/rsa/rsa_sign.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_check.o" => [
            "crypto/rsa/rsa_sp800_56b_check.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_sp800_56b_gen.o" => [
            "crypto/rsa/rsa_sp800_56b_gen.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_x931.o" => [
            "crypto/rsa/rsa_x931.c"
        ],
        "crypto/rsa/libcrypto-shlib-rsa_x931g.o" => [
            "crypto/rsa/rsa_x931g.c"
        ],
        "crypto/sdf/libcrypto-lib-sdf_lib.o" => [
            "crypto/sdf/sdf_lib.c"
        ],
        "crypto/sdf/libcrypto-lib-sdf_meth.o" => [
            "crypto/sdf/sdf_meth.c"
        ],
        "crypto/sdf/libcrypto-shlib-sdf_lib.o" => [
            "crypto/sdf/sdf_lib.c"
        ],
        "crypto/sdf/libcrypto-shlib-sdf_meth.o" => [
            "crypto/sdf/sdf_meth.c"
        ],
        "crypto/sha/libcrypto-lib-keccak1600-x86_64.o" => [
            "crypto/sha/keccak1600-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha1-mb-x86_64.o" => [
            "crypto/sha/sha1-mb-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha1-x86_64.o" => [
            "crypto/sha/sha1-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha1_one.o" => [
            "crypto/sha/sha1_one.c"
        ],
        "crypto/sha/libcrypto-lib-sha1dgst.o" => [
            "crypto/sha/sha1dgst.c"
        ],
        "crypto/sha/libcrypto-lib-sha256-mb-x86_64.o" => [
            "crypto/sha/sha256-mb-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha256-x86_64.o" => [
            "crypto/sha/sha256-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha256.o" => [
            "crypto/sha/sha256.c"
        ],
        "crypto/sha/libcrypto-lib-sha3.o" => [
            "crypto/sha/sha3.c"
        ],
        "crypto/sha/libcrypto-lib-sha512-x86_64.o" => [
            "crypto/sha/sha512-x86_64.s"
        ],
        "crypto/sha/libcrypto-lib-sha512.o" => [
            "crypto/sha/sha512.c"
        ],
        "crypto/sha/libcrypto-shlib-keccak1600-x86_64.o" => [
            "crypto/sha/keccak1600-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha1-mb-x86_64.o" => [
            "crypto/sha/sha1-mb-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha1-x86_64.o" => [
            "crypto/sha/sha1-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha1_one.o" => [
            "crypto/sha/sha1_one.c"
        ],
        "crypto/sha/libcrypto-shlib-sha1dgst.o" => [
            "crypto/sha/sha1dgst.c"
        ],
        "crypto/sha/libcrypto-shlib-sha256-mb-x86_64.o" => [
            "crypto/sha/sha256-mb-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha256-x86_64.o" => [
            "crypto/sha/sha256-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha256.o" => [
            "crypto/sha/sha256.c"
        ],
        "crypto/sha/libcrypto-shlib-sha3.o" => [
            "crypto/sha/sha3.c"
        ],
        "crypto/sha/libcrypto-shlib-sha512-x86_64.o" => [
            "crypto/sha/sha512-x86_64.s"
        ],
        "crypto/sha/libcrypto-shlib-sha512.o" => [
            "crypto/sha/sha512.c"
        ],
        "crypto/siphash/libcrypto-lib-siphash.o" => [
            "crypto/siphash/siphash.c"
        ],
        "crypto/siphash/libcrypto-shlib-siphash.o" => [
            "crypto/siphash/siphash.c"
        ],
        "crypto/siphash/libssl-shlib-siphash.o" => [
            "crypto/siphash/siphash.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_adrs.o" => [
            "crypto/slh_dsa/slh_adrs.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_dsa.o" => [
            "crypto/slh_dsa/slh_dsa.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_dsa_hash_ctx.o" => [
            "crypto/slh_dsa/slh_dsa_hash_ctx.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_dsa_key.o" => [
            "crypto/slh_dsa/slh_dsa_key.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_fors.o" => [
            "crypto/slh_dsa/slh_fors.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_hash.o" => [
            "crypto/slh_dsa/slh_hash.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_hypertree.o" => [
            "crypto/slh_dsa/slh_hypertree.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_params.o" => [
            "crypto/slh_dsa/slh_params.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_wots.o" => [
            "crypto/slh_dsa/slh_wots.c"
        ],
        "crypto/slh_dsa/libcrypto-lib-slh_xmss.o" => [
            "crypto/slh_dsa/slh_xmss.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_adrs.o" => [
            "crypto/slh_dsa/slh_adrs.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_dsa.o" => [
            "crypto/slh_dsa/slh_dsa.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_dsa_hash_ctx.o" => [
            "crypto/slh_dsa/slh_dsa_hash_ctx.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_dsa_key.o" => [
            "crypto/slh_dsa/slh_dsa_key.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_fors.o" => [
            "crypto/slh_dsa/slh_fors.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_hash.o" => [
            "crypto/slh_dsa/slh_hash.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_hypertree.o" => [
            "crypto/slh_dsa/slh_hypertree.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_params.o" => [
            "crypto/slh_dsa/slh_params.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_wots.o" => [
            "crypto/slh_dsa/slh_wots.c"
        ],
        "crypto/slh_dsa/libcrypto-shlib-slh_xmss.o" => [
            "crypto/slh_dsa/slh_xmss.c"
        ],
        "crypto/sm2/libcrypto-lib-sm2_crypt.o" => [
            "crypto/sm2/sm2_crypt.c"
        ],
        "crypto/sm2/libcrypto-lib-sm2_err.o" => [
            "crypto/sm2/sm2_err.c"
        ],
        "crypto/sm2/libcrypto-lib-sm2_key.o" => [
            "crypto/sm2/sm2_key.c"
        ],
        "crypto/sm2/libcrypto-lib-sm2_kmeth.o" => [
            "crypto/sm2/sm2_kmeth.c"
        ],
        "crypto/sm2/libcrypto-lib-sm2_sign.o" => [
            "crypto/sm2/sm2_sign.c"
        ],
        "crypto/sm2/libcrypto-shlib-sm2_crypt.o" => [
            "crypto/sm2/sm2_crypt.c"
        ],
        "crypto/sm2/libcrypto-shlib-sm2_err.o" => [
            "crypto/sm2/sm2_err.c"
        ],
        "crypto/sm2/libcrypto-shlib-sm2_key.o" => [
            "crypto/sm2/sm2_key.c"
        ],
        "crypto/sm2/libcrypto-shlib-sm2_kmeth.o" => [
            "crypto/sm2/sm2_kmeth.c"
        ],
        "crypto/sm2/libcrypto-shlib-sm2_sign.o" => [
            "crypto/sm2/sm2_sign.c"
        ],
        "crypto/sm3/libcrypto-lib-legacy_sm3.o" => [
            "crypto/sm3/legacy_sm3.c"
        ],
        "crypto/sm3/libcrypto-lib-sm3.o" => [
            "crypto/sm3/sm3.c"
        ],
        "crypto/sm3/libcrypto-shlib-legacy_sm3.o" => [
            "crypto/sm3/legacy_sm3.c"
        ],
        "crypto/sm3/libcrypto-shlib-sm3.o" => [
            "crypto/sm3/sm3.c"
        ],
        "crypto/sm4/libcrypto-lib-sm4.o" => [
            "crypto/sm4/sm4.c"
        ],
        "crypto/sm4/libcrypto-shlib-sm4.o" => [
            "crypto/sm4/sm4.c"
        ],
        "crypto/srp/libcrypto-lib-srp_lib.o" => [
            "crypto/srp/srp_lib.c"
        ],
        "crypto/srp/libcrypto-lib-srp_vfy.o" => [
            "crypto/srp/srp_vfy.c"
        ],
        "crypto/srp/libcrypto-shlib-srp_lib.o" => [
            "crypto/srp/srp_lib.c"
        ],
        "crypto/srp/libcrypto-shlib-srp_vfy.o" => [
            "crypto/srp/srp_vfy.c"
        ],
        "crypto/stack/libcrypto-lib-stack.o" => [
            "crypto/stack/stack.c"
        ],
        "crypto/stack/libcrypto-shlib-stack.o" => [
            "crypto/stack/stack.c"
        ],
        "crypto/store/libcrypto-lib-store_err.o" => [
            "crypto/store/store_err.c"
        ],
        "crypto/store/libcrypto-lib-store_init.o" => [
            "crypto/store/store_init.c"
        ],
        "crypto/store/libcrypto-lib-store_lib.o" => [
            "crypto/store/store_lib.c"
        ],
        "crypto/store/libcrypto-lib-store_meth.o" => [
            "crypto/store/store_meth.c"
        ],
        "crypto/store/libcrypto-lib-store_register.o" => [
            "crypto/store/store_register.c"
        ],
        "crypto/store/libcrypto-lib-store_result.o" => [
            "crypto/store/store_result.c"
        ],
        "crypto/store/libcrypto-lib-store_strings.o" => [
            "crypto/store/store_strings.c"
        ],
        "crypto/store/libcrypto-shlib-store_err.o" => [
            "crypto/store/store_err.c"
        ],
        "crypto/store/libcrypto-shlib-store_init.o" => [
            "crypto/store/store_init.c"
        ],
        "crypto/store/libcrypto-shlib-store_lib.o" => [
            "crypto/store/store_lib.c"
        ],
        "crypto/store/libcrypto-shlib-store_meth.o" => [
            "crypto/store/store_meth.c"
        ],
        "crypto/store/libcrypto-shlib-store_register.o" => [
            "crypto/store/store_register.c"
        ],
        "crypto/store/libcrypto-shlib-store_result.o" => [
            "crypto/store/store_result.c"
        ],
        "crypto/store/libcrypto-shlib-store_strings.o" => [
            "crypto/store/store_strings.c"
        ],
        "crypto/thread/arch/libcrypto-lib-thread_none.o" => [
            "crypto/thread/arch/thread_none.c"
        ],
        "crypto/thread/arch/libcrypto-lib-thread_posix.o" => [
            "crypto/thread/arch/thread_posix.c"
        ],
        "crypto/thread/arch/libcrypto-lib-thread_win.o" => [
            "crypto/thread/arch/thread_win.c"
        ],
        "crypto/thread/arch/libcrypto-shlib-thread_none.o" => [
            "crypto/thread/arch/thread_none.c"
        ],
        "crypto/thread/arch/libcrypto-shlib-thread_posix.o" => [
            "crypto/thread/arch/thread_posix.c"
        ],
        "crypto/thread/arch/libcrypto-shlib-thread_win.o" => [
            "crypto/thread/arch/thread_win.c"
        ],
        "crypto/thread/arch/libssl-shlib-thread_none.o" => [
            "crypto/thread/arch/thread_none.c"
        ],
        "crypto/thread/arch/libssl-shlib-thread_posix.o" => [
            "crypto/thread/arch/thread_posix.c"
        ],
        "crypto/thread/arch/libssl-shlib-thread_win.o" => [
            "crypto/thread/arch/thread_win.c"
        ],
        "crypto/thread/libcrypto-lib-api.o" => [
            "crypto/thread/api.c"
        ],
        "crypto/thread/libcrypto-lib-arch.o" => [
            "crypto/thread/arch.c"
        ],
        "crypto/thread/libcrypto-lib-internal.o" => [
            "crypto/thread/internal.c"
        ],
        "crypto/thread/libcrypto-shlib-api.o" => [
            "crypto/thread/api.c"
        ],
        "crypto/thread/libcrypto-shlib-arch.o" => [
            "crypto/thread/arch.c"
        ],
        "crypto/thread/libcrypto-shlib-internal.o" => [
            "crypto/thread/internal.c"
        ],
        "crypto/thread/libssl-shlib-arch.o" => [
            "crypto/thread/arch.c"
        ],
        "crypto/ts/libcrypto-lib-ts_asn1.o" => [
            "crypto/ts/ts_asn1.c"
        ],
        "crypto/ts/libcrypto-lib-ts_conf.o" => [
            "crypto/ts/ts_conf.c"
        ],
        "crypto/ts/libcrypto-lib-ts_err.o" => [
            "crypto/ts/ts_err.c"
        ],
        "crypto/ts/libcrypto-lib-ts_lib.o" => [
            "crypto/ts/ts_lib.c"
        ],
        "crypto/ts/libcrypto-lib-ts_req_print.o" => [
            "crypto/ts/ts_req_print.c"
        ],
        "crypto/ts/libcrypto-lib-ts_req_utils.o" => [
            "crypto/ts/ts_req_utils.c"
        ],
        "crypto/ts/libcrypto-lib-ts_rsp_print.o" => [
            "crypto/ts/ts_rsp_print.c"
        ],
        "crypto/ts/libcrypto-lib-ts_rsp_sign.o" => [
            "crypto/ts/ts_rsp_sign.c"
        ],
        "crypto/ts/libcrypto-lib-ts_rsp_utils.o" => [
            "crypto/ts/ts_rsp_utils.c"
        ],
        "crypto/ts/libcrypto-lib-ts_rsp_verify.o" => [
            "crypto/ts/ts_rsp_verify.c"
        ],
        "crypto/ts/libcrypto-lib-ts_verify_ctx.o" => [
            "crypto/ts/ts_verify_ctx.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_asn1.o" => [
            "crypto/ts/ts_asn1.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_conf.o" => [
            "crypto/ts/ts_conf.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_err.o" => [
            "crypto/ts/ts_err.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_lib.o" => [
            "crypto/ts/ts_lib.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_req_print.o" => [
            "crypto/ts/ts_req_print.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_req_utils.o" => [
            "crypto/ts/ts_req_utils.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_rsp_print.o" => [
            "crypto/ts/ts_rsp_print.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_rsp_sign.o" => [
            "crypto/ts/ts_rsp_sign.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_rsp_utils.o" => [
            "crypto/ts/ts_rsp_utils.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_rsp_verify.o" => [
            "crypto/ts/ts_rsp_verify.c"
        ],
        "crypto/ts/libcrypto-shlib-ts_verify_ctx.o" => [
            "crypto/ts/ts_verify_ctx.c"
        ],
        "crypto/tsapi/libcrypto-lib-tsapi_lib.o" => [
            "crypto/tsapi/tsapi_lib.c"
        ],
        "crypto/tsapi/libcrypto-shlib-tsapi_lib.o" => [
            "crypto/tsapi/tsapi_lib.c"
        ],
        "crypto/txt_db/libcrypto-lib-txt_db.o" => [
            "crypto/txt_db/txt_db.c"
        ],
        "crypto/txt_db/libcrypto-shlib-txt_db.o" => [
            "crypto/txt_db/txt_db.c"
        ],
        "crypto/ui/libcrypto-lib-ui_err.o" => [
            "crypto/ui/ui_err.c"
        ],
        "crypto/ui/libcrypto-lib-ui_lib.o" => [
            "crypto/ui/ui_lib.c"
        ],
        "crypto/ui/libcrypto-lib-ui_null.o" => [
            "crypto/ui/ui_null.c"
        ],
        "crypto/ui/libcrypto-lib-ui_openssl.o" => [
            "crypto/ui/ui_openssl.c"
        ],
        "crypto/ui/libcrypto-lib-ui_util.o" => [
            "crypto/ui/ui_util.c"
        ],
        "crypto/ui/libcrypto-shlib-ui_err.o" => [
            "crypto/ui/ui_err.c"
        ],
        "crypto/ui/libcrypto-shlib-ui_lib.o" => [
            "crypto/ui/ui_lib.c"
        ],
        "crypto/ui/libcrypto-shlib-ui_null.o" => [
            "crypto/ui/ui_null.c"
        ],
        "crypto/ui/libcrypto-shlib-ui_openssl.o" => [
            "crypto/ui/ui_openssl.c"
        ],
        "crypto/ui/libcrypto-shlib-ui_util.o" => [
            "crypto/ui/ui_util.c"
        ],
        "crypto/x509/libcrypto-lib-by_dir.o" => [
            "crypto/x509/by_dir.c"
        ],
        "crypto/x509/libcrypto-lib-by_file.o" => [
            "crypto/x509/by_file.c"
        ],
        "crypto/x509/libcrypto-lib-by_store.o" => [
            "crypto/x509/by_store.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_cache.o" => [
            "crypto/x509/pcy_cache.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_data.o" => [
            "crypto/x509/pcy_data.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_lib.o" => [
            "crypto/x509/pcy_lib.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_map.o" => [
            "crypto/x509/pcy_map.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_node.o" => [
            "crypto/x509/pcy_node.c"
        ],
        "crypto/x509/libcrypto-lib-pcy_tree.o" => [
            "crypto/x509/pcy_tree.c"
        ],
        "crypto/x509/libcrypto-lib-t_acert.o" => [
            "crypto/x509/t_acert.c"
        ],
        "crypto/x509/libcrypto-lib-t_crl.o" => [
            "crypto/x509/t_crl.c"
        ],
        "crypto/x509/libcrypto-lib-t_req.o" => [
            "crypto/x509/t_req.c"
        ],
        "crypto/x509/libcrypto-lib-t_x509.o" => [
            "crypto/x509/t_x509.c"
        ],
        "crypto/x509/libcrypto-lib-v3_aaa.o" => [
            "crypto/x509/v3_aaa.c"
        ],
        "crypto/x509/libcrypto-lib-v3_ac_tgt.o" => [
            "crypto/x509/v3_ac_tgt.c"
        ],
        "crypto/x509/libcrypto-lib-v3_addr.o" => [
            "crypto/x509/v3_addr.c"
        ],
        "crypto/x509/libcrypto-lib-v3_admis.o" => [
            "crypto/x509/v3_admis.c"
        ],
        "crypto/x509/libcrypto-lib-v3_akeya.o" => [
            "crypto/x509/v3_akeya.c"
        ],
        "crypto/x509/libcrypto-lib-v3_akid.o" => [
            "crypto/x509/v3_akid.c"
        ],
        "crypto/x509/libcrypto-lib-v3_asid.o" => [
            "crypto/x509/v3_asid.c"
        ],
        "crypto/x509/libcrypto-lib-v3_attrdesc.o" => [
            "crypto/x509/v3_attrdesc.c"
        ],
        "crypto/x509/libcrypto-lib-v3_attrmap.o" => [
            "crypto/x509/v3_attrmap.c"
        ],
        "crypto/x509/libcrypto-lib-v3_audit_id.o" => [
            "crypto/x509/v3_audit_id.c"
        ],
        "crypto/x509/libcrypto-lib-v3_authattid.o" => [
            "crypto/x509/v3_authattid.c"
        ],
        "crypto/x509/libcrypto-lib-v3_battcons.o" => [
            "crypto/x509/v3_battcons.c"
        ],
        "crypto/x509/libcrypto-lib-v3_bcons.o" => [
            "crypto/x509/v3_bcons.c"
        ],
        "crypto/x509/libcrypto-lib-v3_bitst.o" => [
            "crypto/x509/v3_bitst.c"
        ],
        "crypto/x509/libcrypto-lib-v3_conf.o" => [
            "crypto/x509/v3_conf.c"
        ],
        "crypto/x509/libcrypto-lib-v3_cpols.o" => [
            "crypto/x509/v3_cpols.c"
        ],
        "crypto/x509/libcrypto-lib-v3_crld.o" => [
            "crypto/x509/v3_crld.c"
        ],
        "crypto/x509/libcrypto-lib-v3_enum.o" => [
            "crypto/x509/v3_enum.c"
        ],
        "crypto/x509/libcrypto-lib-v3_extku.o" => [
            "crypto/x509/v3_extku.c"
        ],
        "crypto/x509/libcrypto-lib-v3_genn.o" => [
            "crypto/x509/v3_genn.c"
        ],
        "crypto/x509/libcrypto-lib-v3_group_ac.o" => [
            "crypto/x509/v3_group_ac.c"
        ],
        "crypto/x509/libcrypto-lib-v3_ia5.o" => [
            "crypto/x509/v3_ia5.c"
        ],
        "crypto/x509/libcrypto-lib-v3_ind_iss.o" => [
            "crypto/x509/v3_ind_iss.c"
        ],
        "crypto/x509/libcrypto-lib-v3_info.o" => [
            "crypto/x509/v3_info.c"
        ],
        "crypto/x509/libcrypto-lib-v3_int.o" => [
            "crypto/x509/v3_int.c"
        ],
        "crypto/x509/libcrypto-lib-v3_iobo.o" => [
            "crypto/x509/v3_iobo.c"
        ],
        "crypto/x509/libcrypto-lib-v3_lib.o" => [
            "crypto/x509/v3_lib.c"
        ],
        "crypto/x509/libcrypto-lib-v3_ncons.o" => [
            "crypto/x509/v3_ncons.c"
        ],
        "crypto/x509/libcrypto-lib-v3_no_ass.o" => [
            "crypto/x509/v3_no_ass.c"
        ],
        "crypto/x509/libcrypto-lib-v3_no_rev_avail.o" => [
            "crypto/x509/v3_no_rev_avail.c"
        ],
        "crypto/x509/libcrypto-lib-v3_pci.o" => [
            "crypto/x509/v3_pci.c"
        ],
        "crypto/x509/libcrypto-lib-v3_pcia.o" => [
            "crypto/x509/v3_pcia.c"
        ],
        "crypto/x509/libcrypto-lib-v3_pcons.o" => [
            "crypto/x509/v3_pcons.c"
        ],
        "crypto/x509/libcrypto-lib-v3_pku.o" => [
            "crypto/x509/v3_pku.c"
        ],
        "crypto/x509/libcrypto-lib-v3_pmaps.o" => [
            "crypto/x509/v3_pmaps.c"
        ],
        "crypto/x509/libcrypto-lib-v3_prn.o" => [
            "crypto/x509/v3_prn.c"
        ],
        "crypto/x509/libcrypto-lib-v3_purp.o" => [
            "crypto/x509/v3_purp.c"
        ],
        "crypto/x509/libcrypto-lib-v3_rolespec.o" => [
            "crypto/x509/v3_rolespec.c"
        ],
        "crypto/x509/libcrypto-lib-v3_san.o" => [
            "crypto/x509/v3_san.c"
        ],
        "crypto/x509/libcrypto-lib-v3_sda.o" => [
            "crypto/x509/v3_sda.c"
        ],
        "crypto/x509/libcrypto-lib-v3_single_use.o" => [
            "crypto/x509/v3_single_use.c"
        ],
        "crypto/x509/libcrypto-lib-v3_skid.o" => [
            "crypto/x509/v3_skid.c"
        ],
        "crypto/x509/libcrypto-lib-v3_soa_id.o" => [
            "crypto/x509/v3_soa_id.c"
        ],
        "crypto/x509/libcrypto-lib-v3_sxnet.o" => [
            "crypto/x509/v3_sxnet.c"
        ],
        "crypto/x509/libcrypto-lib-v3_timespec.o" => [
            "crypto/x509/v3_timespec.c"
        ],
        "crypto/x509/libcrypto-lib-v3_tlsf.o" => [
            "crypto/x509/v3_tlsf.c"
        ],
        "crypto/x509/libcrypto-lib-v3_usernotice.o" => [
            "crypto/x509/v3_usernotice.c"
        ],
        "crypto/x509/libcrypto-lib-v3_utl.o" => [
            "crypto/x509/v3_utl.c"
        ],
        "crypto/x509/libcrypto-lib-v3err.o" => [
            "crypto/x509/v3err.c"
        ],
        "crypto/x509/libcrypto-lib-x509_acert.o" => [
            "crypto/x509/x509_acert.c"
        ],
        "crypto/x509/libcrypto-lib-x509_att.o" => [
            "crypto/x509/x509_att.c"
        ],
        "crypto/x509/libcrypto-lib-x509_cmp.o" => [
            "crypto/x509/x509_cmp.c"
        ],
        "crypto/x509/libcrypto-lib-x509_d2.o" => [
            "crypto/x509/x509_d2.c"
        ],
        "crypto/x509/libcrypto-lib-x509_def.o" => [
            "crypto/x509/x509_def.c"
        ],
        "crypto/x509/libcrypto-lib-x509_err.o" => [
            "crypto/x509/x509_err.c"
        ],
        "crypto/x509/libcrypto-lib-x509_ext.o" => [
            "crypto/x509/x509_ext.c"
        ],
        "crypto/x509/libcrypto-lib-x509_lu.o" => [
            "crypto/x509/x509_lu.c"
        ],
        "crypto/x509/libcrypto-lib-x509_meth.o" => [
            "crypto/x509/x509_meth.c"
        ],
        "crypto/x509/libcrypto-lib-x509_obj.o" => [
            "crypto/x509/x509_obj.c"
        ],
        "crypto/x509/libcrypto-lib-x509_r2x.o" => [
            "crypto/x509/x509_r2x.c"
        ],
        "crypto/x509/libcrypto-lib-x509_req.o" => [
            "crypto/x509/x509_req.c"
        ],
        "crypto/x509/libcrypto-lib-x509_set.o" => [
            "crypto/x509/x509_set.c"
        ],
        "crypto/x509/libcrypto-lib-x509_trust.o" => [
            "crypto/x509/x509_trust.c"
        ],
        "crypto/x509/libcrypto-lib-x509_txt.o" => [
            "crypto/x509/x509_txt.c"
        ],
        "crypto/x509/libcrypto-lib-x509_v3.o" => [
            "crypto/x509/x509_v3.c"
        ],
        "crypto/x509/libcrypto-lib-x509_vfy.o" => [
            "crypto/x509/x509_vfy.c"
        ],
        "crypto/x509/libcrypto-lib-x509_vpm.o" => [
            "crypto/x509/x509_vpm.c"
        ],
        "crypto/x509/libcrypto-lib-x509aset.o" => [
            "crypto/x509/x509aset.c"
        ],
        "crypto/x509/libcrypto-lib-x509cset.o" => [
            "crypto/x509/x509cset.c"
        ],
        "crypto/x509/libcrypto-lib-x509name.o" => [
            "crypto/x509/x509name.c"
        ],
        "crypto/x509/libcrypto-lib-x509rset.o" => [
            "crypto/x509/x509rset.c"
        ],
        "crypto/x509/libcrypto-lib-x509spki.o" => [
            "crypto/x509/x509spki.c"
        ],
        "crypto/x509/libcrypto-lib-x509type.o" => [
            "crypto/x509/x509type.c"
        ],
        "crypto/x509/libcrypto-lib-x_all.o" => [
            "crypto/x509/x_all.c"
        ],
        "crypto/x509/libcrypto-lib-x_attrib.o" => [
            "crypto/x509/x_attrib.c"
        ],
        "crypto/x509/libcrypto-lib-x_crl.o" => [
            "crypto/x509/x_crl.c"
        ],
        "crypto/x509/libcrypto-lib-x_exten.o" => [
            "crypto/x509/x_exten.c"
        ],
        "crypto/x509/libcrypto-lib-x_ietfatt.o" => [
            "crypto/x509/x_ietfatt.c"
        ],
        "crypto/x509/libcrypto-lib-x_name.o" => [
            "crypto/x509/x_name.c"
        ],
        "crypto/x509/libcrypto-lib-x_pubkey.o" => [
            "crypto/x509/x_pubkey.c"
        ],
        "crypto/x509/libcrypto-lib-x_req.o" => [
            "crypto/x509/x_req.c"
        ],
        "crypto/x509/libcrypto-lib-x_x509.o" => [
            "crypto/x509/x_x509.c"
        ],
        "crypto/x509/libcrypto-lib-x_x509a.o" => [
            "crypto/x509/x_x509a.c"
        ],
        "crypto/x509/libcrypto-shlib-by_dir.o" => [
            "crypto/x509/by_dir.c"
        ],
        "crypto/x509/libcrypto-shlib-by_file.o" => [
            "crypto/x509/by_file.c"
        ],
        "crypto/x509/libcrypto-shlib-by_store.o" => [
            "crypto/x509/by_store.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_cache.o" => [
            "crypto/x509/pcy_cache.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_data.o" => [
            "crypto/x509/pcy_data.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_lib.o" => [
            "crypto/x509/pcy_lib.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_map.o" => [
            "crypto/x509/pcy_map.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_node.o" => [
            "crypto/x509/pcy_node.c"
        ],
        "crypto/x509/libcrypto-shlib-pcy_tree.o" => [
            "crypto/x509/pcy_tree.c"
        ],
        "crypto/x509/libcrypto-shlib-t_acert.o" => [
            "crypto/x509/t_acert.c"
        ],
        "crypto/x509/libcrypto-shlib-t_crl.o" => [
            "crypto/x509/t_crl.c"
        ],
        "crypto/x509/libcrypto-shlib-t_req.o" => [
            "crypto/x509/t_req.c"
        ],
        "crypto/x509/libcrypto-shlib-t_x509.o" => [
            "crypto/x509/t_x509.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_aaa.o" => [
            "crypto/x509/v3_aaa.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_ac_tgt.o" => [
            "crypto/x509/v3_ac_tgt.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_addr.o" => [
            "crypto/x509/v3_addr.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_admis.o" => [
            "crypto/x509/v3_admis.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_akeya.o" => [
            "crypto/x509/v3_akeya.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_akid.o" => [
            "crypto/x509/v3_akid.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_asid.o" => [
            "crypto/x509/v3_asid.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_attrdesc.o" => [
            "crypto/x509/v3_attrdesc.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_attrmap.o" => [
            "crypto/x509/v3_attrmap.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_audit_id.o" => [
            "crypto/x509/v3_audit_id.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_authattid.o" => [
            "crypto/x509/v3_authattid.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_battcons.o" => [
            "crypto/x509/v3_battcons.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_bcons.o" => [
            "crypto/x509/v3_bcons.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_bitst.o" => [
            "crypto/x509/v3_bitst.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_conf.o" => [
            "crypto/x509/v3_conf.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_cpols.o" => [
            "crypto/x509/v3_cpols.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_crld.o" => [
            "crypto/x509/v3_crld.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_enum.o" => [
            "crypto/x509/v3_enum.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_extku.o" => [
            "crypto/x509/v3_extku.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_genn.o" => [
            "crypto/x509/v3_genn.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_group_ac.o" => [
            "crypto/x509/v3_group_ac.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_ia5.o" => [
            "crypto/x509/v3_ia5.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_ind_iss.o" => [
            "crypto/x509/v3_ind_iss.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_info.o" => [
            "crypto/x509/v3_info.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_int.o" => [
            "crypto/x509/v3_int.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_iobo.o" => [
            "crypto/x509/v3_iobo.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_lib.o" => [
            "crypto/x509/v3_lib.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_ncons.o" => [
            "crypto/x509/v3_ncons.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_no_ass.o" => [
            "crypto/x509/v3_no_ass.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_no_rev_avail.o" => [
            "crypto/x509/v3_no_rev_avail.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_pci.o" => [
            "crypto/x509/v3_pci.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_pcia.o" => [
            "crypto/x509/v3_pcia.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_pcons.o" => [
            "crypto/x509/v3_pcons.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_pku.o" => [
            "crypto/x509/v3_pku.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_pmaps.o" => [
            "crypto/x509/v3_pmaps.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_prn.o" => [
            "crypto/x509/v3_prn.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_purp.o" => [
            "crypto/x509/v3_purp.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_rolespec.o" => [
            "crypto/x509/v3_rolespec.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_san.o" => [
            "crypto/x509/v3_san.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_sda.o" => [
            "crypto/x509/v3_sda.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_single_use.o" => [
            "crypto/x509/v3_single_use.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_skid.o" => [
            "crypto/x509/v3_skid.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_soa_id.o" => [
            "crypto/x509/v3_soa_id.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_sxnet.o" => [
            "crypto/x509/v3_sxnet.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_timespec.o" => [
            "crypto/x509/v3_timespec.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_tlsf.o" => [
            "crypto/x509/v3_tlsf.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_usernotice.o" => [
            "crypto/x509/v3_usernotice.c"
        ],
        "crypto/x509/libcrypto-shlib-v3_utl.o" => [
            "crypto/x509/v3_utl.c"
        ],
        "crypto/x509/libcrypto-shlib-v3err.o" => [
            "crypto/x509/v3err.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_acert.o" => [
            "crypto/x509/x509_acert.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_att.o" => [
            "crypto/x509/x509_att.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_cmp.o" => [
            "crypto/x509/x509_cmp.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_d2.o" => [
            "crypto/x509/x509_d2.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_def.o" => [
            "crypto/x509/x509_def.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_err.o" => [
            "crypto/x509/x509_err.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_ext.o" => [
            "crypto/x509/x509_ext.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_lu.o" => [
            "crypto/x509/x509_lu.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_meth.o" => [
            "crypto/x509/x509_meth.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_obj.o" => [
            "crypto/x509/x509_obj.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_r2x.o" => [
            "crypto/x509/x509_r2x.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_req.o" => [
            "crypto/x509/x509_req.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_set.o" => [
            "crypto/x509/x509_set.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_trust.o" => [
            "crypto/x509/x509_trust.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_txt.o" => [
            "crypto/x509/x509_txt.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_v3.o" => [
            "crypto/x509/x509_v3.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_vfy.o" => [
            "crypto/x509/x509_vfy.c"
        ],
        "crypto/x509/libcrypto-shlib-x509_vpm.o" => [
            "crypto/x509/x509_vpm.c"
        ],
        "crypto/x509/libcrypto-shlib-x509aset.o" => [
            "crypto/x509/x509aset.c"
        ],
        "crypto/x509/libcrypto-shlib-x509cset.o" => [
            "crypto/x509/x509cset.c"
        ],
        "crypto/x509/libcrypto-shlib-x509name.o" => [
            "crypto/x509/x509name.c"
        ],
        "crypto/x509/libcrypto-shlib-x509rset.o" => [
            "crypto/x509/x509rset.c"
        ],
        "crypto/x509/libcrypto-shlib-x509spki.o" => [
            "crypto/x509/x509spki.c"
        ],
        "crypto/x509/libcrypto-shlib-x509type.o" => [
            "crypto/x509/x509type.c"
        ],
        "crypto/x509/libcrypto-shlib-x_all.o" => [
            "crypto/x509/x_all.c"
        ],
        "crypto/x509/libcrypto-shlib-x_attrib.o" => [
            "crypto/x509/x_attrib.c"
        ],
        "crypto/x509/libcrypto-shlib-x_crl.o" => [
            "crypto/x509/x_crl.c"
        ],
        "crypto/x509/libcrypto-shlib-x_exten.o" => [
            "crypto/x509/x_exten.c"
        ],
        "crypto/x509/libcrypto-shlib-x_ietfatt.o" => [
            "crypto/x509/x_ietfatt.c"
        ],
        "crypto/x509/libcrypto-shlib-x_name.o" => [
            "crypto/x509/x_name.c"
        ],
        "crypto/x509/libcrypto-shlib-x_pubkey.o" => [
            "crypto/x509/x_pubkey.c"
        ],
        "crypto/x509/libcrypto-shlib-x_req.o" => [
            "crypto/x509/x_req.c"
        ],
        "crypto/x509/libcrypto-shlib-x_x509.o" => [
            "crypto/x509/x_x509.c"
        ],
        "crypto/x509/libcrypto-shlib-x_x509a.o" => [
            "crypto/x509/x_x509a.c"
        ],
        "crypto/zuc/libcrypto-lib-zuc.o" => [
            "crypto/zuc/zuc.c"
        ],
        "crypto/zuc/libcrypto-shlib-zuc.o" => [
            "crypto/zuc/zuc.c"
        ],
        "engines/afalg" => [
            "engines/afalg-dso-e_afalg.o",
            "engines/afalg.ld"
        ],
        "engines/afalg-dso-e_afalg.o" => [
            "engines/e_afalg.c"
        ],
        "engines/capi" => [
            "engines/capi-dso-e_capi.o",
            "engines/capi.ld"
        ],
        "engines/capi-dso-e_capi.o" => [
            "engines/e_capi.c"
        ],
        "engines/dasync" => [
            "engines/dasync-dso-e_dasync.o",
            "engines/dasync.ld"
        ],
        "engines/dasync-dso-e_dasync.o" => [
            "engines/e_dasync.c"
        ],
        "engines/ecptest" => [
            "engines/ecptest-dso-e_ecptest.o",
            "engines/ecptest.ld"
        ],
        "engines/ecptest-dso-e_ecptest.o" => [
            "engines/e_ecptest.c"
        ],
        "engines/loader_attic" => [
            "crypto/pem/loader_attic-dso-pvkfmt.o",
            "engines/loader_attic-dso-e_loader_attic.o",
            "engines/loader_attic.ld"
        ],
        "engines/loader_attic-dso-e_loader_attic.o" => [
            "engines/e_loader_attic.c"
        ],
        "engines/ossltest" => [
            "engines/ossltest-dso-e_ossltest.o",
            "engines/ossltest.ld"
        ],
        "engines/ossltest-dso-e_ossltest.o" => [
            "engines/e_ossltest.c"
        ],
        "engines/padlock" => [
            "engines/padlock-dso-e_padlock-x86_64.o",
            "engines/padlock-dso-e_padlock.o",
            "engines/padlock.ld"
        ],
        "engines/padlock-dso-e_padlock-x86_64.o" => [
            "engines/e_padlock-x86_64.s"
        ],
        "engines/padlock-dso-e_padlock.o" => [
            "engines/e_padlock.c"
        ],
        "libcrypto" => [
            "crypto/aes/libcrypto-lib-aes-x86_64.o",
            "crypto/aes/libcrypto-lib-aes_cfb.o",
            "crypto/aes/libcrypto-lib-aes_ecb.o",
            "crypto/aes/libcrypto-lib-aes_ige.o",
            "crypto/aes/libcrypto-lib-aes_misc.o",
            "crypto/aes/libcrypto-lib-aes_ofb.o",
            "crypto/aes/libcrypto-lib-aes_wrap.o",
            "crypto/aes/libcrypto-lib-aesni-mb-x86_64.o",
            "crypto/aes/libcrypto-lib-aesni-sha1-x86_64.o",
            "crypto/aes/libcrypto-lib-aesni-sha256-x86_64.o",
            "crypto/aes/libcrypto-lib-aesni-x86_64.o",
            "crypto/aes/libcrypto-lib-aesni-xts-avx512.o",
            "crypto/aes/libcrypto-lib-bsaes-x86_64.o",
            "crypto/aes/libcrypto-lib-vpaes-x86_64.o",
            "crypto/asn1/libcrypto-lib-a_bitstr.o",
            "crypto/asn1/libcrypto-lib-a_d2i_fp.o",
            "crypto/asn1/libcrypto-lib-a_digest.o",
            "crypto/asn1/libcrypto-lib-a_dup.o",
            "crypto/asn1/libcrypto-lib-a_gentm.o",
            "crypto/asn1/libcrypto-lib-a_i2d_fp.o",
            "crypto/asn1/libcrypto-lib-a_int.o",
            "crypto/asn1/libcrypto-lib-a_mbstr.o",
            "crypto/asn1/libcrypto-lib-a_object.o",
            "crypto/asn1/libcrypto-lib-a_octet.o",
            "crypto/asn1/libcrypto-lib-a_print.o",
            "crypto/asn1/libcrypto-lib-a_sign.o",
            "crypto/asn1/libcrypto-lib-a_strex.o",
            "crypto/asn1/libcrypto-lib-a_strnid.o",
            "crypto/asn1/libcrypto-lib-a_time.o",
            "crypto/asn1/libcrypto-lib-a_type.o",
            "crypto/asn1/libcrypto-lib-a_utctm.o",
            "crypto/asn1/libcrypto-lib-a_utf8.o",
            "crypto/asn1/libcrypto-lib-a_verify.o",
            "crypto/asn1/libcrypto-lib-ameth_lib.o",
            "crypto/asn1/libcrypto-lib-asn1_err.o",
            "crypto/asn1/libcrypto-lib-asn1_gen.o",
            "crypto/asn1/libcrypto-lib-asn1_item_list.o",
            "crypto/asn1/libcrypto-lib-asn1_lib.o",
            "crypto/asn1/libcrypto-lib-asn1_parse.o",
            "crypto/asn1/libcrypto-lib-asn_mime.o",
            "crypto/asn1/libcrypto-lib-asn_moid.o",
            "crypto/asn1/libcrypto-lib-asn_mstbl.o",
            "crypto/asn1/libcrypto-lib-asn_pack.o",
            "crypto/asn1/libcrypto-lib-bio_asn1.o",
            "crypto/asn1/libcrypto-lib-bio_ndef.o",
            "crypto/asn1/libcrypto-lib-d2i_param.o",
            "crypto/asn1/libcrypto-lib-d2i_pr.o",
            "crypto/asn1/libcrypto-lib-d2i_pu.o",
            "crypto/asn1/libcrypto-lib-evp_asn1.o",
            "crypto/asn1/libcrypto-lib-f_int.o",
            "crypto/asn1/libcrypto-lib-f_string.o",
            "crypto/asn1/libcrypto-lib-i2d_evp.o",
            "crypto/asn1/libcrypto-lib-n_pkey.o",
            "crypto/asn1/libcrypto-lib-nsseq.o",
            "crypto/asn1/libcrypto-lib-p5_pbe.o",
            "crypto/asn1/libcrypto-lib-p5_pbev2.o",
            "crypto/asn1/libcrypto-lib-p5_scrypt.o",
            "crypto/asn1/libcrypto-lib-p8_pkey.o",
            "crypto/asn1/libcrypto-lib-t_bitst.o",
            "crypto/asn1/libcrypto-lib-t_pkey.o",
            "crypto/asn1/libcrypto-lib-t_spki.o",
            "crypto/asn1/libcrypto-lib-tasn_dec.o",
            "crypto/asn1/libcrypto-lib-tasn_enc.o",
            "crypto/asn1/libcrypto-lib-tasn_fre.o",
            "crypto/asn1/libcrypto-lib-tasn_new.o",
            "crypto/asn1/libcrypto-lib-tasn_prn.o",
            "crypto/asn1/libcrypto-lib-tasn_scn.o",
            "crypto/asn1/libcrypto-lib-tasn_typ.o",
            "crypto/asn1/libcrypto-lib-tasn_utl.o",
            "crypto/asn1/libcrypto-lib-x_algor.o",
            "crypto/asn1/libcrypto-lib-x_bignum.o",
            "crypto/asn1/libcrypto-lib-x_info.o",
            "crypto/asn1/libcrypto-lib-x_int64.o",
            "crypto/asn1/libcrypto-lib-x_long.o",
            "crypto/asn1/libcrypto-lib-x_pkey.o",
            "crypto/asn1/libcrypto-lib-x_sig.o",
            "crypto/asn1/libcrypto-lib-x_spki.o",
            "crypto/asn1/libcrypto-lib-x_val.o",
            "crypto/async/arch/libcrypto-lib-async_null.o",
            "crypto/async/arch/libcrypto-lib-async_posix.o",
            "crypto/async/arch/libcrypto-lib-async_win.o",
            "crypto/async/libcrypto-lib-async.o",
            "crypto/async/libcrypto-lib-async_err.o",
            "crypto/async/libcrypto-lib-async_wait.o",
            "crypto/bio/libcrypto-lib-bf_buff.o",
            "crypto/bio/libcrypto-lib-bf_lbuf.o",
            "crypto/bio/libcrypto-lib-bf_nbio.o",
            "crypto/bio/libcrypto-lib-bf_null.o",
            "crypto/bio/libcrypto-lib-bf_prefix.o",
            "crypto/bio/libcrypto-lib-bf_readbuff.o",
            "crypto/bio/libcrypto-lib-bio_addr.o",
            "crypto/bio/libcrypto-lib-bio_cb.o",
            "crypto/bio/libcrypto-lib-bio_dump.o",
            "crypto/bio/libcrypto-lib-bio_err.o",
            "crypto/bio/libcrypto-lib-bio_lib.o",
            "crypto/bio/libcrypto-lib-bio_meth.o",
            "crypto/bio/libcrypto-lib-bio_print.o",
            "crypto/bio/libcrypto-lib-bio_sock.o",
            "crypto/bio/libcrypto-lib-bio_sock2.o",
            "crypto/bio/libcrypto-lib-bss_acpt.o",
            "crypto/bio/libcrypto-lib-bss_bio.o",
            "crypto/bio/libcrypto-lib-bss_conn.o",
            "crypto/bio/libcrypto-lib-bss_core.o",
            "crypto/bio/libcrypto-lib-bss_dgram.o",
            "crypto/bio/libcrypto-lib-bss_dgram_pair.o",
            "crypto/bio/libcrypto-lib-bss_fd.o",
            "crypto/bio/libcrypto-lib-bss_file.o",
            "crypto/bio/libcrypto-lib-bss_log.o",
            "crypto/bio/libcrypto-lib-bss_mem.o",
            "crypto/bio/libcrypto-lib-bss_null.o",
            "crypto/bio/libcrypto-lib-bss_sock.o",
            "crypto/bio/libcrypto-lib-ossl_core_bio.o",
            "crypto/bn/asm/libcrypto-lib-x86_64-gcc.o",
            "crypto/bn/libcrypto-lib-bn_add.o",
            "crypto/bn/libcrypto-lib-bn_blind.o",
            "crypto/bn/libcrypto-lib-bn_const.o",
            "crypto/bn/libcrypto-lib-bn_conv.o",
            "crypto/bn/libcrypto-lib-bn_ctx.o",
            "crypto/bn/libcrypto-lib-bn_depr.o",
            "crypto/bn/libcrypto-lib-bn_dh.o",
            "crypto/bn/libcrypto-lib-bn_div.o",
            "crypto/bn/libcrypto-lib-bn_err.o",
            "crypto/bn/libcrypto-lib-bn_exp.o",
            "crypto/bn/libcrypto-lib-bn_exp2.o",
            "crypto/bn/libcrypto-lib-bn_gcd.o",
            "crypto/bn/libcrypto-lib-bn_gf2m.o",
            "crypto/bn/libcrypto-lib-bn_intern.o",
            "crypto/bn/libcrypto-lib-bn_kron.o",
            "crypto/bn/libcrypto-lib-bn_lib.o",
            "crypto/bn/libcrypto-lib-bn_mod.o",
            "crypto/bn/libcrypto-lib-bn_mont.o",
            "crypto/bn/libcrypto-lib-bn_mpi.o",
            "crypto/bn/libcrypto-lib-bn_mul.o",
            "crypto/bn/libcrypto-lib-bn_nist.o",
            "crypto/bn/libcrypto-lib-bn_prime.o",
            "crypto/bn/libcrypto-lib-bn_print.o",
            "crypto/bn/libcrypto-lib-bn_rand.o",
            "crypto/bn/libcrypto-lib-bn_recp.o",
            "crypto/bn/libcrypto-lib-bn_rsa_fips186_4.o",
            "crypto/bn/libcrypto-lib-bn_shift.o",
            "crypto/bn/libcrypto-lib-bn_sm2.o",
            "crypto/bn/libcrypto-lib-bn_sqr.o",
            "crypto/bn/libcrypto-lib-bn_sqrt.o",
            "crypto/bn/libcrypto-lib-bn_srp.o",
            "crypto/bn/libcrypto-lib-bn_word.o",
            "crypto/bn/libcrypto-lib-bn_x931p.o",
            "crypto/bn/libcrypto-lib-rsaz-2k-avx512.o",
            "crypto/bn/libcrypto-lib-rsaz-2k-avxifma.o",
            "crypto/bn/libcrypto-lib-rsaz-3k-avx512.o",
            "crypto/bn/libcrypto-lib-rsaz-3k-avxifma.o",
            "crypto/bn/libcrypto-lib-rsaz-4k-avx512.o",
            "crypto/bn/libcrypto-lib-rsaz-4k-avxifma.o",
            "crypto/bn/libcrypto-lib-rsaz-avx2.o",
            "crypto/bn/libcrypto-lib-rsaz-x86_64.o",
            "crypto/bn/libcrypto-lib-rsaz_exp.o",
            "crypto/bn/libcrypto-lib-rsaz_exp_x2.o",
            "crypto/bn/libcrypto-lib-x86_64-gf2m.o",
            "crypto/bn/libcrypto-lib-x86_64-mont.o",
            "crypto/bn/libcrypto-lib-x86_64-mont5.o",
            "crypto/buffer/libcrypto-lib-buf_err.o",
            "crypto/buffer/libcrypto-lib-buffer.o",
            "crypto/chacha/libcrypto-lib-chacha-x86_64.o",
            "crypto/cmac/libcrypto-lib-cmac.o",
            "crypto/cmp/libcrypto-lib-cmp_asn.o",
            "crypto/cmp/libcrypto-lib-cmp_client.o",
            "crypto/cmp/libcrypto-lib-cmp_ctx.o",
            "crypto/cmp/libcrypto-lib-cmp_err.o",
            "crypto/cmp/libcrypto-lib-cmp_genm.o",
            "crypto/cmp/libcrypto-lib-cmp_hdr.o",
            "crypto/cmp/libcrypto-lib-cmp_http.o",
            "crypto/cmp/libcrypto-lib-cmp_msg.o",
            "crypto/cmp/libcrypto-lib-cmp_protect.o",
            "crypto/cmp/libcrypto-lib-cmp_server.o",
            "crypto/cmp/libcrypto-lib-cmp_status.o",
            "crypto/cmp/libcrypto-lib-cmp_util.o",
            "crypto/cmp/libcrypto-lib-cmp_vfy.o",
            "crypto/cms/libcrypto-lib-cms_asn1.o",
            "crypto/cms/libcrypto-lib-cms_att.o",
            "crypto/cms/libcrypto-lib-cms_cd.o",
            "crypto/cms/libcrypto-lib-cms_dd.o",
            "crypto/cms/libcrypto-lib-cms_dh.o",
            "crypto/cms/libcrypto-lib-cms_ec.o",
            "crypto/cms/libcrypto-lib-cms_enc.o",
            "crypto/cms/libcrypto-lib-cms_env.o",
            "crypto/cms/libcrypto-lib-cms_err.o",
            "crypto/cms/libcrypto-lib-cms_ess.o",
            "crypto/cms/libcrypto-lib-cms_io.o",
            "crypto/cms/libcrypto-lib-cms_kari.o",
            "crypto/cms/libcrypto-lib-cms_lib.o",
            "crypto/cms/libcrypto-lib-cms_pwri.o",
            "crypto/cms/libcrypto-lib-cms_rsa.o",
            "crypto/cms/libcrypto-lib-cms_sd.o",
            "crypto/cms/libcrypto-lib-cms_smime.o",
            "crypto/comp/libcrypto-lib-c_brotli.o",
            "crypto/comp/libcrypto-lib-c_zlib.o",
            "crypto/comp/libcrypto-lib-c_zstd.o",
            "crypto/comp/libcrypto-lib-comp_err.o",
            "crypto/comp/libcrypto-lib-comp_lib.o",
            "crypto/conf/libcrypto-lib-conf_api.o",
            "crypto/conf/libcrypto-lib-conf_def.o",
            "crypto/conf/libcrypto-lib-conf_err.o",
            "crypto/conf/libcrypto-lib-conf_lib.o",
            "crypto/conf/libcrypto-lib-conf_mall.o",
            "crypto/conf/libcrypto-lib-conf_mod.o",
            "crypto/conf/libcrypto-lib-conf_sap.o",
            "crypto/conf/libcrypto-lib-conf_ssl.o",
            "crypto/crmf/libcrypto-lib-crmf_asn.o",
            "crypto/crmf/libcrypto-lib-crmf_err.o",
            "crypto/crmf/libcrypto-lib-crmf_lib.o",
            "crypto/crmf/libcrypto-lib-crmf_pbm.o",
            "crypto/ct/libcrypto-lib-ct_b64.o",
            "crypto/ct/libcrypto-lib-ct_err.o",
            "crypto/ct/libcrypto-lib-ct_log.o",
            "crypto/ct/libcrypto-lib-ct_oct.o",
            "crypto/ct/libcrypto-lib-ct_policy.o",
            "crypto/ct/libcrypto-lib-ct_prn.o",
            "crypto/ct/libcrypto-lib-ct_sct.o",
            "crypto/ct/libcrypto-lib-ct_sct_ctx.o",
            "crypto/ct/libcrypto-lib-ct_vfy.o",
            "crypto/ct/libcrypto-lib-ct_x509v3.o",
            "crypto/des/libcrypto-lib-cbc_cksm.o",
            "crypto/des/libcrypto-lib-cbc_enc.o",
            "crypto/des/libcrypto-lib-cfb64ede.o",
            "crypto/des/libcrypto-lib-cfb64enc.o",
            "crypto/des/libcrypto-lib-cfb_enc.o",
            "crypto/des/libcrypto-lib-des_enc.o",
            "crypto/des/libcrypto-lib-ecb3_enc.o",
            "crypto/des/libcrypto-lib-ecb_enc.o",
            "crypto/des/libcrypto-lib-fcrypt.o",
            "crypto/des/libcrypto-lib-fcrypt_b.o",
            "crypto/des/libcrypto-lib-ofb64ede.o",
            "crypto/des/libcrypto-lib-ofb64enc.o",
            "crypto/des/libcrypto-lib-ofb_enc.o",
            "crypto/des/libcrypto-lib-pcbc_enc.o",
            "crypto/des/libcrypto-lib-qud_cksm.o",
            "crypto/des/libcrypto-lib-rand_key.o",
            "crypto/des/libcrypto-lib-set_key.o",
            "crypto/des/libcrypto-lib-str2key.o",
            "crypto/des/libcrypto-lib-xcbc_enc.o",
            "crypto/dh/libcrypto-lib-dh_ameth.o",
            "crypto/dh/libcrypto-lib-dh_asn1.o",
            "crypto/dh/libcrypto-lib-dh_backend.o",
            "crypto/dh/libcrypto-lib-dh_check.o",
            "crypto/dh/libcrypto-lib-dh_depr.o",
            "crypto/dh/libcrypto-lib-dh_err.o",
            "crypto/dh/libcrypto-lib-dh_gen.o",
            "crypto/dh/libcrypto-lib-dh_group_params.o",
            "crypto/dh/libcrypto-lib-dh_kdf.o",
            "crypto/dh/libcrypto-lib-dh_key.o",
            "crypto/dh/libcrypto-lib-dh_lib.o",
            "crypto/dh/libcrypto-lib-dh_meth.o",
            "crypto/dh/libcrypto-lib-dh_pmeth.o",
            "crypto/dh/libcrypto-lib-dh_prn.o",
            "crypto/dh/libcrypto-lib-dh_rfc5114.o",
            "crypto/dsa/libcrypto-lib-dsa_ameth.o",
            "crypto/dsa/libcrypto-lib-dsa_asn1.o",
            "crypto/dsa/libcrypto-lib-dsa_backend.o",
            "crypto/dsa/libcrypto-lib-dsa_check.o",
            "crypto/dsa/libcrypto-lib-dsa_depr.o",
            "crypto/dsa/libcrypto-lib-dsa_err.o",
            "crypto/dsa/libcrypto-lib-dsa_gen.o",
            "crypto/dsa/libcrypto-lib-dsa_key.o",
            "crypto/dsa/libcrypto-lib-dsa_lib.o",
            "crypto/dsa/libcrypto-lib-dsa_meth.o",
            "crypto/dsa/libcrypto-lib-dsa_ossl.o",
            "crypto/dsa/libcrypto-lib-dsa_pmeth.o",
            "crypto/dsa/libcrypto-lib-dsa_prn.o",
            "crypto/dsa/libcrypto-lib-dsa_sign.o",
            "crypto/dsa/libcrypto-lib-dsa_vrf.o",
            "crypto/dso/libcrypto-lib-dso_dl.o",
            "crypto/dso/libcrypto-lib-dso_dlfcn.o",
            "crypto/dso/libcrypto-lib-dso_err.o",
            "crypto/dso/libcrypto-lib-dso_lib.o",
            "crypto/dso/libcrypto-lib-dso_openssl.o",
            "crypto/dso/libcrypto-lib-dso_win32.o",
            "crypto/ec/curve448/arch_32/libcrypto-lib-f_impl32.o",
            "crypto/ec/curve448/arch_64/libcrypto-lib-f_impl64.o",
            "crypto/ec/curve448/libcrypto-lib-curve448.o",
            "crypto/ec/curve448/libcrypto-lib-curve448_tables.o",
            "crypto/ec/curve448/libcrypto-lib-eddsa.o",
            "crypto/ec/curve448/libcrypto-lib-f_generic.o",
            "crypto/ec/curve448/libcrypto-lib-scalar.o",
            "crypto/ec/libcrypto-lib-curve25519.o",
            "crypto/ec/libcrypto-lib-ec2_oct.o",
            "crypto/ec/libcrypto-lib-ec2_smpl.o",
            "crypto/ec/libcrypto-lib-ec_ameth.o",
            "crypto/ec/libcrypto-lib-ec_asn1.o",
            "crypto/ec/libcrypto-lib-ec_backend.o",
            "crypto/ec/libcrypto-lib-ec_check.o",
            "crypto/ec/libcrypto-lib-ec_curve.o",
            "crypto/ec/libcrypto-lib-ec_cvt.o",
            "crypto/ec/libcrypto-lib-ec_deprecated.o",
            "crypto/ec/libcrypto-lib-ec_err.o",
            "crypto/ec/libcrypto-lib-ec_key.o",
            "crypto/ec/libcrypto-lib-ec_kmeth.o",
            "crypto/ec/libcrypto-lib-ec_lib.o",
            "crypto/ec/libcrypto-lib-ec_mult.o",
            "crypto/ec/libcrypto-lib-ec_oct.o",
            "crypto/ec/libcrypto-lib-ec_pmeth.o",
            "crypto/ec/libcrypto-lib-ec_print.o",
            "crypto/ec/libcrypto-lib-ecdh_kdf.o",
            "crypto/ec/libcrypto-lib-ecdh_ossl.o",
            "crypto/ec/libcrypto-lib-ecdsa_ossl.o",
            "crypto/ec/libcrypto-lib-ecdsa_sign.o",
            "crypto/ec/libcrypto-lib-ecdsa_vrf.o",
            "crypto/ec/libcrypto-lib-eck_prn.o",
            "crypto/ec/libcrypto-lib-ecp_meth.o",
            "crypto/ec/libcrypto-lib-ecp_mont.o",
            "crypto/ec/libcrypto-lib-ecp_nist.o",
            "crypto/ec/libcrypto-lib-ecp_nistz256-x86_64.o",
            "crypto/ec/libcrypto-lib-ecp_nistz256.o",
            "crypto/ec/libcrypto-lib-ecp_oct.o",
            "crypto/ec/libcrypto-lib-ecp_smpl.o",
            "crypto/ec/libcrypto-lib-ecx_backend.o",
            "crypto/ec/libcrypto-lib-ecx_key.o",
            "crypto/ec/libcrypto-lib-ecx_meth.o",
            "crypto/ec/libcrypto-lib-x25519-x86_64.o",
            "crypto/eia3/libcrypto-lib-eia3.o",
            "crypto/encode_decode/libcrypto-lib-decoder_err.o",
            "crypto/encode_decode/libcrypto-lib-decoder_lib.o",
            "crypto/encode_decode/libcrypto-lib-decoder_meth.o",
            "crypto/encode_decode/libcrypto-lib-decoder_pkey.o",
            "crypto/encode_decode/libcrypto-lib-encoder_err.o",
            "crypto/encode_decode/libcrypto-lib-encoder_lib.o",
            "crypto/encode_decode/libcrypto-lib-encoder_meth.o",
            "crypto/encode_decode/libcrypto-lib-encoder_pkey.o",
            "crypto/engine/libcrypto-lib-eng_all.o",
            "crypto/engine/libcrypto-lib-eng_cnf.o",
            "crypto/engine/libcrypto-lib-eng_ctrl.o",
            "crypto/engine/libcrypto-lib-eng_dyn.o",
            "crypto/engine/libcrypto-lib-eng_err.o",
            "crypto/engine/libcrypto-lib-eng_fat.o",
            "crypto/engine/libcrypto-lib-eng_init.o",
            "crypto/engine/libcrypto-lib-eng_lib.o",
            "crypto/engine/libcrypto-lib-eng_list.o",
            "crypto/engine/libcrypto-lib-eng_openssl.o",
            "crypto/engine/libcrypto-lib-eng_pkey.o",
            "crypto/engine/libcrypto-lib-eng_rdrand.o",
            "crypto/engine/libcrypto-lib-eng_table.o",
            "crypto/engine/libcrypto-lib-tb_asnmth.o",
            "crypto/engine/libcrypto-lib-tb_cipher.o",
            "crypto/engine/libcrypto-lib-tb_dh.o",
            "crypto/engine/libcrypto-lib-tb_digest.o",
            "crypto/engine/libcrypto-lib-tb_dsa.o",
            "crypto/engine/libcrypto-lib-tb_eckey.o",
            "crypto/engine/libcrypto-lib-tb_ecpmeth.o",
            "crypto/engine/libcrypto-lib-tb_pkmeth.o",
            "crypto/engine/libcrypto-lib-tb_rand.o",
            "crypto/engine/libcrypto-lib-tb_rsa.o",
            "crypto/err/libcrypto-lib-err.o",
            "crypto/err/libcrypto-lib-err_all.o",
            "crypto/err/libcrypto-lib-err_all_legacy.o",
            "crypto/err/libcrypto-lib-err_blocks.o",
            "crypto/err/libcrypto-lib-err_mark.o",
            "crypto/err/libcrypto-lib-err_prn.o",
            "crypto/err/libcrypto-lib-err_save.o",
            "crypto/ess/libcrypto-lib-ess_asn1.o",
            "crypto/ess/libcrypto-lib-ess_err.o",
            "crypto/ess/libcrypto-lib-ess_lib.o",
            "crypto/evp/libcrypto-lib-asymcipher.o",
            "crypto/evp/libcrypto-lib-bio_b64.o",
            "crypto/evp/libcrypto-lib-bio_enc.o",
            "crypto/evp/libcrypto-lib-bio_md.o",
            "crypto/evp/libcrypto-lib-bio_ok.o",
            "crypto/evp/libcrypto-lib-c_allc.o",
            "crypto/evp/libcrypto-lib-c_alld.o",
            "crypto/evp/libcrypto-lib-cmeth_lib.o",
            "crypto/evp/libcrypto-lib-ctrl_params_translate.o",
            "crypto/evp/libcrypto-lib-dh_ctrl.o",
            "crypto/evp/libcrypto-lib-dh_support.o",
            "crypto/evp/libcrypto-lib-digest.o",
            "crypto/evp/libcrypto-lib-dsa_ctrl.o",
            "crypto/evp/libcrypto-lib-e_aes.o",
            "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha1.o",
            "crypto/evp/libcrypto-lib-e_aes_cbc_hmac_sha256.o",
            "crypto/evp/libcrypto-lib-e_chacha20_poly1305.o",
            "crypto/evp/libcrypto-lib-e_des.o",
            "crypto/evp/libcrypto-lib-e_des3.o",
            "crypto/evp/libcrypto-lib-e_eea3.o",
            "crypto/evp/libcrypto-lib-e_null.o",
            "crypto/evp/libcrypto-lib-e_old.o",
            "crypto/evp/libcrypto-lib-e_rc4.o",
            "crypto/evp/libcrypto-lib-e_rc4_hmac_md5.o",
            "crypto/evp/libcrypto-lib-e_rc5.o",
            "crypto/evp/libcrypto-lib-e_sm4.o",
            "crypto/evp/libcrypto-lib-e_wbsm4_baiwu.o",
            "crypto/evp/libcrypto-lib-e_wbsm4_wsise.o",
            "crypto/evp/libcrypto-lib-e_wbsm4_xiaolai.o",
            "crypto/evp/libcrypto-lib-e_xcbc_d.o",
            "crypto/evp/libcrypto-lib-ec_ctrl.o",
            "crypto/evp/libcrypto-lib-ec_support.o",
            "crypto/evp/libcrypto-lib-encode.o",
            "crypto/evp/libcrypto-lib-evp_cnf.o",
            "crypto/evp/libcrypto-lib-evp_enc.o",
            "crypto/evp/libcrypto-lib-evp_err.o",
            "crypto/evp/libcrypto-lib-evp_fetch.o",
            "crypto/evp/libcrypto-lib-evp_key.o",
            "crypto/evp/libcrypto-lib-evp_lib.o",
            "crypto/evp/libcrypto-lib-evp_pbe.o",
            "crypto/evp/libcrypto-lib-evp_pkey.o",
            "crypto/evp/libcrypto-lib-evp_rand.o",
            "crypto/evp/libcrypto-lib-evp_utils.o",
            "crypto/evp/libcrypto-lib-exchange.o",
            "crypto/evp/libcrypto-lib-kdf_lib.o",
            "crypto/evp/libcrypto-lib-kdf_meth.o",
            "crypto/evp/libcrypto-lib-kem.o",
            "crypto/evp/libcrypto-lib-keymgmt_lib.o",
            "crypto/evp/libcrypto-lib-keymgmt_meth.o",
            "crypto/evp/libcrypto-lib-legacy_md5.o",
            "crypto/evp/libcrypto-lib-legacy_md5_sha1.o",
            "crypto/evp/libcrypto-lib-legacy_sha.o",
            "crypto/evp/libcrypto-lib-m_null.o",
            "crypto/evp/libcrypto-lib-m_sigver.o",
            "crypto/evp/libcrypto-lib-mac_lib.o",
            "crypto/evp/libcrypto-lib-mac_meth.o",
            "crypto/evp/libcrypto-lib-names.o",
            "crypto/evp/libcrypto-lib-p5_crpt.o",
            "crypto/evp/libcrypto-lib-p5_crpt2.o",
            "crypto/evp/libcrypto-lib-p_dec.o",
            "crypto/evp/libcrypto-lib-p_enc.o",
            "crypto/evp/libcrypto-lib-p_legacy.o",
            "crypto/evp/libcrypto-lib-p_lib.o",
            "crypto/evp/libcrypto-lib-p_open.o",
            "crypto/evp/libcrypto-lib-p_seal.o",
            "crypto/evp/libcrypto-lib-p_sign.o",
            "crypto/evp/libcrypto-lib-p_verify.o",
            "crypto/evp/libcrypto-lib-pbe_scrypt.o",
            "crypto/evp/libcrypto-lib-pmeth_check.o",
            "crypto/evp/libcrypto-lib-pmeth_gn.o",
            "crypto/evp/libcrypto-lib-pmeth_lib.o",
            "crypto/evp/libcrypto-lib-s_lib.o",
            "crypto/evp/libcrypto-lib-signature.o",
            "crypto/evp/libcrypto-lib-skeymgmt_meth.o",
            "crypto/ffc/libcrypto-lib-ffc_backend.o",
            "crypto/ffc/libcrypto-lib-ffc_dh.o",
            "crypto/ffc/libcrypto-lib-ffc_key_generate.o",
            "crypto/ffc/libcrypto-lib-ffc_key_validate.o",
            "crypto/ffc/libcrypto-lib-ffc_params.o",
            "crypto/ffc/libcrypto-lib-ffc_params_generate.o",
            "crypto/ffc/libcrypto-lib-ffc_params_validate.o",
            "crypto/hashtable/libcrypto-lib-hashfunc.o",
            "crypto/hashtable/libcrypto-lib-hashtable.o",
            "crypto/hmac/libcrypto-lib-hmac.o",
            "crypto/hpke/libcrypto-lib-hpke.o",
            "crypto/hpke/libcrypto-lib-hpke_util.o",
            "crypto/http/libcrypto-lib-http_client.o",
            "crypto/http/libcrypto-lib-http_err.o",
            "crypto/http/libcrypto-lib-http_lib.o",
            "crypto/kdf/libcrypto-lib-kdf_err.o",
            "crypto/lhash/libcrypto-lib-lh_stats.o",
            "crypto/lhash/libcrypto-lib-lhash.o",
            "crypto/libcrypto-lib-asn1_dsa.o",
            "crypto/libcrypto-lib-bsearch.o",
            "crypto/libcrypto-lib-comp_methods.o",
            "crypto/libcrypto-lib-context.o",
            "crypto/libcrypto-lib-core_algorithm.o",
            "crypto/libcrypto-lib-core_fetch.o",
            "crypto/libcrypto-lib-core_namemap.o",
            "crypto/libcrypto-lib-cpt_err.o",
            "crypto/libcrypto-lib-cpuid.o",
            "crypto/libcrypto-lib-cryptlib.o",
            "crypto/libcrypto-lib-ctype.o",
            "crypto/libcrypto-lib-cversion.o",
            "crypto/libcrypto-lib-defaults.o",
            "crypto/libcrypto-lib-der_writer.o",
            "crypto/libcrypto-lib-deterministic_nonce.o",
            "crypto/libcrypto-lib-ebcdic.o",
            "crypto/libcrypto-lib-ex_data.o",
            "crypto/libcrypto-lib-getenv.o",
            "crypto/libcrypto-lib-indicator_core.o",
            "crypto/libcrypto-lib-info.o",
            "crypto/libcrypto-lib-init.o",
            "crypto/libcrypto-lib-initthread.o",
            "crypto/libcrypto-lib-mem.o",
            "crypto/libcrypto-lib-mem_sec.o",
            "crypto/libcrypto-lib-o_dir.o",
            "crypto/libcrypto-lib-o_fopen.o",
            "crypto/libcrypto-lib-o_init.o",
            "crypto/libcrypto-lib-o_str.o",
            "crypto/libcrypto-lib-o_syslog.o",
            "crypto/libcrypto-lib-o_time.o",
            "crypto/libcrypto-lib-packet.o",
            "crypto/libcrypto-lib-param_build.o",
            "crypto/libcrypto-lib-param_build_set.o",
            "crypto/libcrypto-lib-params.o",
            "crypto/libcrypto-lib-params_dup.o",
            "crypto/libcrypto-lib-params_from_text.o",
            "crypto/libcrypto-lib-params_idx.o",
            "crypto/libcrypto-lib-passphrase.o",
            "crypto/libcrypto-lib-provider.o",
            "crypto/libcrypto-lib-provider_child.o",
            "crypto/libcrypto-lib-provider_conf.o",
            "crypto/libcrypto-lib-provider_core.o",
            "crypto/libcrypto-lib-provider_predefined.o",
            "crypto/libcrypto-lib-punycode.o",
            "crypto/libcrypto-lib-quic_vlint.o",
            "crypto/libcrypto-lib-self_test_core.o",
            "crypto/libcrypto-lib-sleep.o",
            "crypto/libcrypto-lib-sparse_array.o",
            "crypto/libcrypto-lib-ssl_err.o",
            "crypto/libcrypto-lib-threads_lib.o",
            "crypto/libcrypto-lib-threads_none.o",
            "crypto/libcrypto-lib-threads_pthread.o",
            "crypto/libcrypto-lib-threads_win.o",
            "crypto/libcrypto-lib-time.o",
            "crypto/libcrypto-lib-trace.o",
            "crypto/libcrypto-lib-uid.o",
            "crypto/libcrypto-lib-x86_64cpuid.o",
            "crypto/md5/libcrypto-lib-md5-x86_64.o",
            "crypto/md5/libcrypto-lib-md5_dgst.o",
            "crypto/md5/libcrypto-lib-md5_one.o",
            "crypto/md5/libcrypto-lib-md5_sha1.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_encoders.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_key.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_key_compress.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_matrix.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_ntt.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_params.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_sample.o",
            "crypto/ml_dsa/libcrypto-lib-ml_dsa_sign.o",
            "crypto/ml_kem/libcrypto-lib-ml_kem.o",
            "crypto/modes/libcrypto-lib-aes-gcm-avx512.o",
            "crypto/modes/libcrypto-lib-aesni-gcm-x86_64.o",
            "crypto/modes/libcrypto-lib-cbc128.o",
            "crypto/modes/libcrypto-lib-ccm128.o",
            "crypto/modes/libcrypto-lib-cfb128.o",
            "crypto/modes/libcrypto-lib-ctr128.o",
            "crypto/modes/libcrypto-lib-cts128.o",
            "crypto/modes/libcrypto-lib-gcm128.o",
            "crypto/modes/libcrypto-lib-ghash-x86_64.o",
            "crypto/modes/libcrypto-lib-ocb128.o",
            "crypto/modes/libcrypto-lib-ofb128.o",
            "crypto/modes/libcrypto-lib-siv128.o",
            "crypto/modes/libcrypto-lib-wrap128.o",
            "crypto/modes/libcrypto-lib-xts128.o",
            "crypto/modes/libcrypto-lib-xts128gb.o",
            "crypto/objects/libcrypto-lib-o_names.o",
            "crypto/objects/libcrypto-lib-obj_dat.o",
            "crypto/objects/libcrypto-lib-obj_err.o",
            "crypto/objects/libcrypto-lib-obj_lib.o",
            "crypto/objects/libcrypto-lib-obj_xref.o",
            "crypto/ocsp/libcrypto-lib-ocsp_asn.o",
            "crypto/ocsp/libcrypto-lib-ocsp_cl.o",
            "crypto/ocsp/libcrypto-lib-ocsp_err.o",
            "crypto/ocsp/libcrypto-lib-ocsp_ext.o",
            "crypto/ocsp/libcrypto-lib-ocsp_http.o",
            "crypto/ocsp/libcrypto-lib-ocsp_lib.o",
            "crypto/ocsp/libcrypto-lib-ocsp_prn.o",
            "crypto/ocsp/libcrypto-lib-ocsp_srv.o",
            "crypto/ocsp/libcrypto-lib-ocsp_vfy.o",
            "crypto/ocsp/libcrypto-lib-v3_ocsp.o",
            "crypto/pem/libcrypto-lib-pem_all.o",
            "crypto/pem/libcrypto-lib-pem_err.o",
            "crypto/pem/libcrypto-lib-pem_info.o",
            "crypto/pem/libcrypto-lib-pem_lib.o",
            "crypto/pem/libcrypto-lib-pem_oth.o",
            "crypto/pem/libcrypto-lib-pem_pk8.o",
            "crypto/pem/libcrypto-lib-pem_pkey.o",
            "crypto/pem/libcrypto-lib-pem_sign.o",
            "crypto/pem/libcrypto-lib-pem_x509.o",
            "crypto/pem/libcrypto-lib-pem_xaux.o",
            "crypto/pem/libcrypto-lib-pvkfmt.o",
            "crypto/pkcs12/libcrypto-lib-p12_add.o",
            "crypto/pkcs12/libcrypto-lib-p12_asn.o",
            "crypto/pkcs12/libcrypto-lib-p12_attr.o",
            "crypto/pkcs12/libcrypto-lib-p12_crpt.o",
            "crypto/pkcs12/libcrypto-lib-p12_crt.o",
            "crypto/pkcs12/libcrypto-lib-p12_decr.o",
            "crypto/pkcs12/libcrypto-lib-p12_init.o",
            "crypto/pkcs12/libcrypto-lib-p12_key.o",
            "crypto/pkcs12/libcrypto-lib-p12_kiss.o",
            "crypto/pkcs12/libcrypto-lib-p12_mutl.o",
            "crypto/pkcs12/libcrypto-lib-p12_npas.o",
            "crypto/pkcs12/libcrypto-lib-p12_p8d.o",
            "crypto/pkcs12/libcrypto-lib-p12_p8e.o",
            "crypto/pkcs12/libcrypto-lib-p12_sbag.o",
            "crypto/pkcs12/libcrypto-lib-p12_utl.o",
            "crypto/pkcs12/libcrypto-lib-pk12err.o",
            "crypto/pkcs7/libcrypto-lib-bio_pk7.o",
            "crypto/pkcs7/libcrypto-lib-pk7_asn1.o",
            "crypto/pkcs7/libcrypto-lib-pk7_attr.o",
            "crypto/pkcs7/libcrypto-lib-pk7_doit.o",
            "crypto/pkcs7/libcrypto-lib-pk7_lib.o",
            "crypto/pkcs7/libcrypto-lib-pk7_mime.o",
            "crypto/pkcs7/libcrypto-lib-pk7_smime.o",
            "crypto/pkcs7/libcrypto-lib-pkcs7err.o",
            "crypto/poly1305/libcrypto-lib-poly1305-x86_64.o",
            "crypto/poly1305/libcrypto-lib-poly1305.o",
            "crypto/property/libcrypto-lib-defn_cache.o",
            "crypto/property/libcrypto-lib-property.o",
            "crypto/property/libcrypto-lib-property_err.o",
            "crypto/property/libcrypto-lib-property_parse.o",
            "crypto/property/libcrypto-lib-property_query.o",
            "crypto/property/libcrypto-lib-property_string.o",
            "crypto/rand/libcrypto-lib-prov_seed.o",
            "crypto/rand/libcrypto-lib-rand_deprecated.o",
            "crypto/rand/libcrypto-lib-rand_err.o",
            "crypto/rand/libcrypto-lib-rand_lib.o",
            "crypto/rand/libcrypto-lib-rand_meth.o",
            "crypto/rand/libcrypto-lib-rand_pool.o",
            "crypto/rand/libcrypto-lib-rand_uniform.o",
            "crypto/rand/libcrypto-lib-randfile.o",
            "crypto/rc4/libcrypto-lib-rc4-md5-x86_64.o",
            "crypto/rc4/libcrypto-lib-rc4-x86_64.o",
            "crypto/rsa/libcrypto-lib-rsa_ameth.o",
            "crypto/rsa/libcrypto-lib-rsa_asn1.o",
            "crypto/rsa/libcrypto-lib-rsa_backend.o",
            "crypto/rsa/libcrypto-lib-rsa_chk.o",
            "crypto/rsa/libcrypto-lib-rsa_crpt.o",
            "crypto/rsa/libcrypto-lib-rsa_depr.o",
            "crypto/rsa/libcrypto-lib-rsa_err.o",
            "crypto/rsa/libcrypto-lib-rsa_gen.o",
            "crypto/rsa/libcrypto-lib-rsa_lib.o",
            "crypto/rsa/libcrypto-lib-rsa_meth.o",
            "crypto/rsa/libcrypto-lib-rsa_mp.o",
            "crypto/rsa/libcrypto-lib-rsa_mp_names.o",
            "crypto/rsa/libcrypto-lib-rsa_none.o",
            "crypto/rsa/libcrypto-lib-rsa_oaep.o",
            "crypto/rsa/libcrypto-lib-rsa_ossl.o",
            "crypto/rsa/libcrypto-lib-rsa_pk1.o",
            "crypto/rsa/libcrypto-lib-rsa_pmeth.o",
            "crypto/rsa/libcrypto-lib-rsa_prn.o",
            "crypto/rsa/libcrypto-lib-rsa_pss.o",
            "crypto/rsa/libcrypto-lib-rsa_saos.o",
            "crypto/rsa/libcrypto-lib-rsa_schemes.o",
            "crypto/rsa/libcrypto-lib-rsa_sign.o",
            "crypto/rsa/libcrypto-lib-rsa_sp800_56b_check.o",
            "crypto/rsa/libcrypto-lib-rsa_sp800_56b_gen.o",
            "crypto/rsa/libcrypto-lib-rsa_x931.o",
            "crypto/rsa/libcrypto-lib-rsa_x931g.o",
            "crypto/sdf/libcrypto-lib-sdf_lib.o",
            "crypto/sdf/libcrypto-lib-sdf_meth.o",
            "crypto/sha/libcrypto-lib-keccak1600-x86_64.o",
            "crypto/sha/libcrypto-lib-sha1-mb-x86_64.o",
            "crypto/sha/libcrypto-lib-sha1-x86_64.o",
            "crypto/sha/libcrypto-lib-sha1_one.o",
            "crypto/sha/libcrypto-lib-sha1dgst.o",
            "crypto/sha/libcrypto-lib-sha256-mb-x86_64.o",
            "crypto/sha/libcrypto-lib-sha256-x86_64.o",
            "crypto/sha/libcrypto-lib-sha256.o",
            "crypto/sha/libcrypto-lib-sha3.o",
            "crypto/sha/libcrypto-lib-sha512-x86_64.o",
            "crypto/sha/libcrypto-lib-sha512.o",
            "crypto/siphash/libcrypto-lib-siphash.o",
            "crypto/slh_dsa/libcrypto-lib-slh_adrs.o",
            "crypto/slh_dsa/libcrypto-lib-slh_dsa.o",
            "crypto/slh_dsa/libcrypto-lib-slh_dsa_hash_ctx.o",
            "crypto/slh_dsa/libcrypto-lib-slh_dsa_key.o",
            "crypto/slh_dsa/libcrypto-lib-slh_fors.o",
            "crypto/slh_dsa/libcrypto-lib-slh_hash.o",
            "crypto/slh_dsa/libcrypto-lib-slh_hypertree.o",
            "crypto/slh_dsa/libcrypto-lib-slh_params.o",
            "crypto/slh_dsa/libcrypto-lib-slh_wots.o",
            "crypto/slh_dsa/libcrypto-lib-slh_xmss.o",
            "crypto/sm2/libcrypto-lib-sm2_crypt.o",
            "crypto/sm2/libcrypto-lib-sm2_err.o",
            "crypto/sm2/libcrypto-lib-sm2_key.o",
            "crypto/sm2/libcrypto-lib-sm2_kmeth.o",
            "crypto/sm2/libcrypto-lib-sm2_sign.o",
            "crypto/sm3/libcrypto-lib-legacy_sm3.o",
            "crypto/sm3/libcrypto-lib-sm3.o",
            "crypto/sm4/libcrypto-lib-sm4.o",
            "crypto/srp/libcrypto-lib-srp_lib.o",
            "crypto/srp/libcrypto-lib-srp_vfy.o",
            "crypto/stack/libcrypto-lib-stack.o",
            "crypto/store/libcrypto-lib-store_err.o",
            "crypto/store/libcrypto-lib-store_init.o",
            "crypto/store/libcrypto-lib-store_lib.o",
            "crypto/store/libcrypto-lib-store_meth.o",
            "crypto/store/libcrypto-lib-store_register.o",
            "crypto/store/libcrypto-lib-store_result.o",
            "crypto/store/libcrypto-lib-store_strings.o",
            "crypto/thread/arch/libcrypto-lib-thread_none.o",
            "crypto/thread/arch/libcrypto-lib-thread_posix.o",
            "crypto/thread/arch/libcrypto-lib-thread_win.o",
            "crypto/thread/libcrypto-lib-api.o",
            "crypto/thread/libcrypto-lib-arch.o",
            "crypto/thread/libcrypto-lib-internal.o",
            "crypto/ts/libcrypto-lib-ts_asn1.o",
            "crypto/ts/libcrypto-lib-ts_conf.o",
            "crypto/ts/libcrypto-lib-ts_err.o",
            "crypto/ts/libcrypto-lib-ts_lib.o",
            "crypto/ts/libcrypto-lib-ts_req_print.o",
            "crypto/ts/libcrypto-lib-ts_req_utils.o",
            "crypto/ts/libcrypto-lib-ts_rsp_print.o",
            "crypto/ts/libcrypto-lib-ts_rsp_sign.o",
            "crypto/ts/libcrypto-lib-ts_rsp_utils.o",
            "crypto/ts/libcrypto-lib-ts_rsp_verify.o",
            "crypto/ts/libcrypto-lib-ts_verify_ctx.o",
            "crypto/tsapi/libcrypto-lib-tsapi_lib.o",
            "crypto/txt_db/libcrypto-lib-txt_db.o",
            "crypto/ui/libcrypto-lib-ui_err.o",
            "crypto/ui/libcrypto-lib-ui_lib.o",
            "crypto/ui/libcrypto-lib-ui_null.o",
            "crypto/ui/libcrypto-lib-ui_openssl.o",
            "crypto/ui/libcrypto-lib-ui_util.o",
            "crypto/x509/libcrypto-lib-by_dir.o",
            "crypto/x509/libcrypto-lib-by_file.o",
            "crypto/x509/libcrypto-lib-by_store.o",
            "crypto/x509/libcrypto-lib-pcy_cache.o",
            "crypto/x509/libcrypto-lib-pcy_data.o",
            "crypto/x509/libcrypto-lib-pcy_lib.o",
            "crypto/x509/libcrypto-lib-pcy_map.o",
            "crypto/x509/libcrypto-lib-pcy_node.o",
            "crypto/x509/libcrypto-lib-pcy_tree.o",
            "crypto/x509/libcrypto-lib-t_acert.o",
            "crypto/x509/libcrypto-lib-t_crl.o",
            "crypto/x509/libcrypto-lib-t_req.o",
            "crypto/x509/libcrypto-lib-t_x509.o",
            "crypto/x509/libcrypto-lib-v3_aaa.o",
            "crypto/x509/libcrypto-lib-v3_ac_tgt.o",
            "crypto/x509/libcrypto-lib-v3_addr.o",
            "crypto/x509/libcrypto-lib-v3_admis.o",
            "crypto/x509/libcrypto-lib-v3_akeya.o",
            "crypto/x509/libcrypto-lib-v3_akid.o",
            "crypto/x509/libcrypto-lib-v3_asid.o",
            "crypto/x509/libcrypto-lib-v3_attrdesc.o",
            "crypto/x509/libcrypto-lib-v3_attrmap.o",
            "crypto/x509/libcrypto-lib-v3_audit_id.o",
            "crypto/x509/libcrypto-lib-v3_authattid.o",
            "crypto/x509/libcrypto-lib-v3_battcons.o",
            "crypto/x509/libcrypto-lib-v3_bcons.o",
            "crypto/x509/libcrypto-lib-v3_bitst.o",
            "crypto/x509/libcrypto-lib-v3_conf.o",
            "crypto/x509/libcrypto-lib-v3_cpols.o",
            "crypto/x509/libcrypto-lib-v3_crld.o",
            "crypto/x509/libcrypto-lib-v3_enum.o",
            "crypto/x509/libcrypto-lib-v3_extku.o",
            "crypto/x509/libcrypto-lib-v3_genn.o",
            "crypto/x509/libcrypto-lib-v3_group_ac.o",
            "crypto/x509/libcrypto-lib-v3_ia5.o",
            "crypto/x509/libcrypto-lib-v3_ind_iss.o",
            "crypto/x509/libcrypto-lib-v3_info.o",
            "crypto/x509/libcrypto-lib-v3_int.o",
            "crypto/x509/libcrypto-lib-v3_iobo.o",
            "crypto/x509/libcrypto-lib-v3_lib.o",
            "crypto/x509/libcrypto-lib-v3_ncons.o",
            "crypto/x509/libcrypto-lib-v3_no_ass.o",
            "crypto/x509/libcrypto-lib-v3_no_rev_avail.o",
            "crypto/x509/libcrypto-lib-v3_pci.o",
            "crypto/x509/libcrypto-lib-v3_pcia.o",
            "crypto/x509/libcrypto-lib-v3_pcons.o",
            "crypto/x509/libcrypto-lib-v3_pku.o",
            "crypto/x509/libcrypto-lib-v3_pmaps.o",
            "crypto/x509/libcrypto-lib-v3_prn.o",
            "crypto/x509/libcrypto-lib-v3_purp.o",
            "crypto/x509/libcrypto-lib-v3_rolespec.o",
            "crypto/x509/libcrypto-lib-v3_san.o",
            "crypto/x509/libcrypto-lib-v3_sda.o",
            "crypto/x509/libcrypto-lib-v3_single_use.o",
            "crypto/x509/libcrypto-lib-v3_skid.o",
            "crypto/x509/libcrypto-lib-v3_soa_id.o",
            "crypto/x509/libcrypto-lib-v3_sxnet.o",
            "crypto/x509/libcrypto-lib-v3_timespec.o",
            "crypto/x509/libcrypto-lib-v3_tlsf.o",
            "crypto/x509/libcrypto-lib-v3_usernotice.o",
            "crypto/x509/libcrypto-lib-v3_utl.o",
            "crypto/x509/libcrypto-lib-v3err.o",
            "crypto/x509/libcrypto-lib-x509_acert.o",
            "crypto/x509/libcrypto-lib-x509_att.o",
            "crypto/x509/libcrypto-lib-x509_cmp.o",
            "crypto/x509/libcrypto-lib-x509_d2.o",
            "crypto/x509/libcrypto-lib-x509_def.o",
            "crypto/x509/libcrypto-lib-x509_err.o",
            "crypto/x509/libcrypto-lib-x509_ext.o",
            "crypto/x509/libcrypto-lib-x509_lu.o",
            "crypto/x509/libcrypto-lib-x509_meth.o",
            "crypto/x509/libcrypto-lib-x509_obj.o",
            "crypto/x509/libcrypto-lib-x509_r2x.o",
            "crypto/x509/libcrypto-lib-x509_req.o",
            "crypto/x509/libcrypto-lib-x509_set.o",
            "crypto/x509/libcrypto-lib-x509_trust.o",
            "crypto/x509/libcrypto-lib-x509_txt.o",
            "crypto/x509/libcrypto-lib-x509_v3.o",
            "crypto/x509/libcrypto-lib-x509_vfy.o",
            "crypto/x509/libcrypto-lib-x509_vpm.o",
            "crypto/x509/libcrypto-lib-x509aset.o",
            "crypto/x509/libcrypto-lib-x509cset.o",
            "crypto/x509/libcrypto-lib-x509name.o",
            "crypto/x509/libcrypto-lib-x509rset.o",
            "crypto/x509/libcrypto-lib-x509spki.o",
            "crypto/x509/libcrypto-lib-x509type.o",
            "crypto/x509/libcrypto-lib-x_all.o",
            "crypto/x509/libcrypto-lib-x_attrib.o",
            "crypto/x509/libcrypto-lib-x_crl.o",
            "crypto/x509/libcrypto-lib-x_exten.o",
            "crypto/x509/libcrypto-lib-x_ietfatt.o",
            "crypto/x509/libcrypto-lib-x_name.o",
            "crypto/x509/libcrypto-lib-x_pubkey.o",
            "crypto/x509/libcrypto-lib-x_req.o",
            "crypto/x509/libcrypto-lib-x_x509.o",
            "crypto/x509/libcrypto-lib-x_x509a.o",
            "crypto/zuc/libcrypto-lib-zuc.o",
            "providers/libcrypto-lib-baseprov.o",
            "providers/libcrypto-lib-defltprov.o",
            "providers/libcrypto-lib-nullprov.o",
            "providers/libcrypto-lib-prov_running.o",
            "providers/libdefault.a"
        ],
        "libssl" => [
            "ssl/libssl-lib-bio_ssl.o",
            "ssl/libssl-lib-d1_lib.o",
            "ssl/libssl-lib-d1_msg.o",
            "ssl/libssl-lib-d1_srtp.o",
            "ssl/libssl-lib-methods.o",
            "ssl/libssl-lib-pqueue.o",
            "ssl/libssl-lib-priority_queue.o",
            "ssl/libssl-lib-s3_enc.o",
            "ssl/libssl-lib-s3_lib.o",
            "ssl/libssl-lib-s3_msg.o",
            "ssl/libssl-lib-ssl_asn1.o",
            "ssl/libssl-lib-ssl_cert.o",
            "ssl/libssl-lib-ssl_cert_comp.o",
            "ssl/libssl-lib-ssl_ciph.o",
            "ssl/libssl-lib-ssl_conf.o",
            "ssl/libssl-lib-ssl_err_legacy.o",
            "ssl/libssl-lib-ssl_init.o",
            "ssl/libssl-lib-ssl_lib.o",
            "ssl/libssl-lib-ssl_mcnf.o",
            "ssl/libssl-lib-ssl_rsa.o",
            "ssl/libssl-lib-ssl_rsa_legacy.o",
            "ssl/libssl-lib-ssl_sess.o",
            "ssl/libssl-lib-ssl_stat.o",
            "ssl/libssl-lib-ssl_txt.o",
            "ssl/libssl-lib-ssl_utst.o",
            "ssl/libssl-lib-t1_enc.o",
            "ssl/libssl-lib-t1_lib.o",
            "ssl/libssl-lib-t1_trce.o",
            "ssl/libssl-lib-tls13_enc.o",
            "ssl/libssl-lib-tls_depr.o",
            "ssl/libssl-lib-tls_srp.o",
            "ssl/quic/libssl-lib-cc_newreno.o",
            "ssl/quic/libssl-lib-json_enc.o",
            "ssl/quic/libssl-lib-qlog.o",
            "ssl/quic/libssl-lib-qlog_event_helpers.o",
            "ssl/quic/libssl-lib-quic_ackm.o",
            "ssl/quic/libssl-lib-quic_cfq.o",
            "ssl/quic/libssl-lib-quic_channel.o",
            "ssl/quic/libssl-lib-quic_demux.o",
            "ssl/quic/libssl-lib-quic_engine.o",
            "ssl/quic/libssl-lib-quic_fc.o",
            "ssl/quic/libssl-lib-quic_fifd.o",
            "ssl/quic/libssl-lib-quic_impl.o",
            "ssl/quic/libssl-lib-quic_lcidm.o",
            "ssl/quic/libssl-lib-quic_method.o",
            "ssl/quic/libssl-lib-quic_obj.o",
            "ssl/quic/libssl-lib-quic_port.o",
            "ssl/quic/libssl-lib-quic_rcidm.o",
            "ssl/quic/libssl-lib-quic_reactor.o",
            "ssl/quic/libssl-lib-quic_reactor_wait_ctx.o",
            "ssl/quic/libssl-lib-quic_record_rx.o",
            "ssl/quic/libssl-lib-quic_record_shared.o",
            "ssl/quic/libssl-lib-quic_record_tx.o",
            "ssl/quic/libssl-lib-quic_record_util.o",
            "ssl/quic/libssl-lib-quic_rstream.o",
            "ssl/quic/libssl-lib-quic_rx_depack.o",
            "ssl/quic/libssl-lib-quic_sf_list.o",
            "ssl/quic/libssl-lib-quic_srt_gen.o",
            "ssl/quic/libssl-lib-quic_srtm.o",
            "ssl/quic/libssl-lib-quic_sstream.o",
            "ssl/quic/libssl-lib-quic_statm.o",
            "ssl/quic/libssl-lib-quic_stream_map.o",
            "ssl/quic/libssl-lib-quic_thread_assist.o",
            "ssl/quic/libssl-lib-quic_tls.o",
            "ssl/quic/libssl-lib-quic_tls_api.o",
            "ssl/quic/libssl-lib-quic_tls_api_method.o",
            "ssl/quic/libssl-lib-quic_trace.o",
            "ssl/quic/libssl-lib-quic_tserver.o",
            "ssl/quic/libssl-lib-quic_txp.o",
            "ssl/quic/libssl-lib-quic_txpim.o",
            "ssl/quic/libssl-lib-quic_types.o",
            "ssl/quic/libssl-lib-quic_wire.o",
            "ssl/quic/libssl-lib-quic_wire_pkt.o",
            "ssl/quic/libssl-lib-uint_set.o",
            "ssl/record/libssl-lib-rec_layer_d1.o",
            "ssl/record/libssl-lib-rec_layer_s3.o",
            "ssl/record/methods/libssl-lib-dtls_meth.o",
            "ssl/record/methods/libssl-lib-ssl3_meth.o",
            "ssl/record/methods/libssl-lib-tls13_meth.o",
            "ssl/record/methods/libssl-lib-tls1_meth.o",
            "ssl/record/methods/libssl-lib-tls_common.o",
            "ssl/record/methods/libssl-lib-tls_multib.o",
            "ssl/record/methods/libssl-lib-tlsany_meth.o",
            "ssl/rio/libssl-lib-poll_builder.o",
            "ssl/rio/libssl-lib-poll_immediate.o",
            "ssl/rio/libssl-lib-rio_notifier.o",
            "ssl/statem/libssl-lib-extensions.o",
            "ssl/statem/libssl-lib-extensions_clnt.o",
            "ssl/statem/libssl-lib-extensions_cust.o",
            "ssl/statem/libssl-lib-extensions_srvr.o",
            "ssl/statem/libssl-lib-statem.o",
            "ssl/statem/libssl-lib-statem_clnt.o",
            "ssl/statem/libssl-lib-statem_dtls.o",
            "ssl/statem/libssl-lib-statem_lib.o",
            "ssl/statem/libssl-lib-statem_srvr.o",
            "ssl/statem_ntls/libssl-lib-ntls_extensions.o",
            "ssl/statem_ntls/libssl-lib-ntls_extensions_clnt.o",
            "ssl/statem_ntls/libssl-lib-ntls_extensions_cust.o",
            "ssl/statem_ntls/libssl-lib-ntls_extensions_srvr.o",
            "ssl/statem_ntls/libssl-lib-ntls_statem.o",
            "ssl/statem_ntls/libssl-lib-ntls_statem_clnt.o",
            "ssl/statem_ntls/libssl-lib-ntls_statem_lib.o",
            "ssl/statem_ntls/libssl-lib-ntls_statem_srvr.o"
        ],
        "providers/common/der/libcommon-lib-der_digests_gen.o" => [
            "providers/common/der/der_digests_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_dsa_gen.o" => [
            "providers/common/der/der_dsa_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_dsa_key.o" => [
            "providers/common/der/der_dsa_key.c"
        ],
        "providers/common/der/libcommon-lib-der_dsa_sig.o" => [
            "providers/common/der/der_dsa_sig.c"
        ],
        "providers/common/der/libcommon-lib-der_ec_gen.o" => [
            "providers/common/der/der_ec_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_ec_key.o" => [
            "providers/common/der/der_ec_key.c"
        ],
        "providers/common/der/libcommon-lib-der_ec_sig.o" => [
            "providers/common/der/der_ec_sig.c"
        ],
        "providers/common/der/libcommon-lib-der_ecx_gen.o" => [
            "providers/common/der/der_ecx_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_ecx_key.o" => [
            "providers/common/der/der_ecx_key.c"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_gen.o" => [
            "providers/common/der/der_ml_dsa_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_ml_dsa_key.o" => [
            "providers/common/der/der_ml_dsa_key.c"
        ],
        "providers/common/der/libcommon-lib-der_rsa_gen.o" => [
            "providers/common/der/der_rsa_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_rsa_key.o" => [
            "providers/common/der/der_rsa_key.c"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_gen.o" => [
            "providers/common/der/der_slh_dsa_gen.c"
        ],
        "providers/common/der/libcommon-lib-der_slh_dsa_key.o" => [
            "providers/common/der/der_slh_dsa_key.c"
        ],
        "providers/common/der/libcommon-lib-der_wrap_gen.o" => [
            "providers/common/der/der_wrap_gen.c"
        ],
        "providers/common/der/libdefault-lib-der_rsa_sig.o" => [
            "providers/common/der/der_rsa_sig.c"
        ],
        "providers/common/der/libdefault-lib-der_sm2_gen.o" => [
            "providers/common/der/der_sm2_gen.c"
        ],
        "providers/common/der/libdefault-lib-der_sm2_key.o" => [
            "providers/common/der/der_sm2_key.c"
        ],
        "providers/common/der/libdefault-lib-der_sm2_sig.o" => [
            "providers/common/der/der_sm2_sig.c"
        ],
        "providers/common/libcommon-lib-provider_ctx.o" => [
            "providers/common/provider_ctx.c"
        ],
        "providers/common/libcommon-lib-provider_err.o" => [
            "providers/common/provider_err.c"
        ],
        "providers/common/libdefault-lib-bio_prov.o" => [
            "providers/common/bio_prov.c"
        ],
        "providers/common/libdefault-lib-capabilities.o" => [
            "providers/common/capabilities.c"
        ],
        "providers/common/libdefault-lib-digest_to_nid.o" => [
            "providers/common/digest_to_nid.c"
        ],
        "providers/common/libdefault-lib-provider_seeding.o" => [
            "providers/common/provider_seeding.c"
        ],
        "providers/common/libdefault-lib-provider_util.o" => [
            "providers/common/provider_util.c"
        ],
        "providers/common/libdefault-lib-securitycheck.o" => [
            "providers/common/securitycheck.c"
        ],
        "providers/common/libdefault-lib-securitycheck_default.o" => [
            "providers/common/securitycheck_default.c"
        ],
        "providers/common/liblegacy-lib-provider_util.o" => [
            "providers/common/provider_util.c"
        ],
        "providers/implementations/asymciphers/libdefault-lib-rsa_enc.o" => [
            "providers/implementations/asymciphers/rsa_enc.c"
        ],
        "providers/implementations/asymciphers/libdefault-lib-sm2_enc.o" => [
            "providers/implementations/asymciphers/sm2_enc.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon.o" => [
            "providers/implementations/ciphers/ciphercommon.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_block.o" => [
            "providers/implementations/ciphers/ciphercommon_block.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm.o" => [
            "providers/implementations/ciphers/ciphercommon_ccm.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm_hw.o" => [
            "providers/implementations/ciphers/ciphercommon_ccm_hw.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm.o" => [
            "providers/implementations/ciphers/ciphercommon_gcm.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm_hw.o" => [
            "providers/implementations/ciphers/ciphercommon_gcm_hw.c"
        ],
        "providers/implementations/ciphers/libcommon-lib-ciphercommon_hw.o" => [
            "providers/implementations/ciphers/ciphercommon_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes.o" => [
            "providers/implementations/ciphers/cipher_aes.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha.o" => [
            "providers/implementations/ciphers/cipher_aes_cbc_hmac_sha.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha1_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_cbc_hmac_sha1_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha256_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_cbc_hmac_sha256_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm.o" => [
            "providers/implementations/ciphers/cipher_aes_ccm.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_ccm_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm.o" => [
            "providers/implementations/ciphers/cipher_aes_gcm.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_gcm_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv.o" => [
            "providers/implementations/ciphers/cipher_aes_gcm_siv.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_gcm_siv_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_polyval.o" => [
            "providers/implementations/ciphers/cipher_aes_gcm_siv_polyval.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb.o" => [
            "providers/implementations/ciphers/cipher_aes_ocb.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_ocb_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv.o" => [
            "providers/implementations/ciphers/cipher_aes_siv.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_siv_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_wrp.o" => [
            "providers/implementations/ciphers/cipher_aes_wrp.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts.o" => [
            "providers/implementations/ciphers/cipher_aes_xts.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_fips.o" => [
            "providers/implementations/ciphers/cipher_aes_xts_fips.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_hw.o" => [
            "providers/implementations/ciphers/cipher_aes_xts_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_chacha20.o" => [
            "providers/implementations/ciphers/cipher_chacha20.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_hw.o" => [
            "providers/implementations/ciphers/cipher_chacha20_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305.o" => [
            "providers/implementations/ciphers/cipher_chacha20_poly1305.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305_hw.o" => [
            "providers/implementations/ciphers/cipher_chacha20_poly1305_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_cts.o" => [
            "providers/implementations/ciphers/cipher_cts.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_null.o" => [
            "providers/implementations/ciphers/cipher_null.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4.o" => [
            "providers/implementations/ciphers/cipher_sm4.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm.o" => [
            "providers/implementations/ciphers/cipher_sm4_ccm.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm_hw.o" => [
            "providers/implementations/ciphers/cipher_sm4_ccm_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm.o" => [
            "providers/implementations/ciphers/cipher_sm4_gcm.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm_hw.o" => [
            "providers/implementations/ciphers/cipher_sm4_gcm_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_hw.o" => [
            "providers/implementations/ciphers/cipher_sm4_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts.o" => [
            "providers/implementations/ciphers/cipher_sm4_xts.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts_hw.o" => [
            "providers/implementations/ciphers/cipher_sm4_xts_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes.o" => [
            "providers/implementations/ciphers/cipher_tdes.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_common.o" => [
            "providers/implementations/ciphers/cipher_tdes_common.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default.o" => [
            "providers/implementations/ciphers/cipher_tdes_default.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default_hw.o" => [
            "providers/implementations/ciphers/cipher_tdes_default_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_hw.o" => [
            "providers/implementations/ciphers/cipher_tdes_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap.o" => [
            "providers/implementations/ciphers/cipher_tdes_wrap.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap_hw.o" => [
            "providers/implementations/ciphers/cipher_tdes_wrap_hw.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3.o" => [
            "providers/implementations/ciphers/cipher_zuc_eea3.c"
        ],
        "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3_hw.o" => [
            "providers/implementations/ciphers/cipher_zuc_eea3_hw.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_des.o" => [
            "providers/implementations/ciphers/cipher_des.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_des_hw.o" => [
            "providers/implementations/ciphers/cipher_des_hw.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_desx.o" => [
            "providers/implementations/ciphers/cipher_desx.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_desx_hw.o" => [
            "providers/implementations/ciphers/cipher_desx_hw.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_rc4.o" => [
            "providers/implementations/ciphers/cipher_rc4.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5.o" => [
            "providers/implementations/ciphers/cipher_rc4_hmac_md5.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5_hw.o" => [
            "providers/implementations/ciphers/cipher_rc4_hmac_md5_hw.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hw.o" => [
            "providers/implementations/ciphers/cipher_rc4_hw.c"
        ],
        "providers/implementations/ciphers/liblegacy-lib-cipher_tdes_common.o" => [
            "providers/implementations/ciphers/cipher_tdes_common.c"
        ],
        "providers/implementations/digests/libcommon-lib-digestcommon.o" => [
            "providers/implementations/digests/digestcommon.c"
        ],
        "providers/implementations/digests/libdefault-lib-md5_prov.o" => [
            "providers/implementations/digests/md5_prov.c"
        ],
        "providers/implementations/digests/libdefault-lib-md5_sha1_prov.o" => [
            "providers/implementations/digests/md5_sha1_prov.c"
        ],
        "providers/implementations/digests/libdefault-lib-null_prov.o" => [
            "providers/implementations/digests/null_prov.c"
        ],
        "providers/implementations/digests/libdefault-lib-sha2_prov.o" => [
            "providers/implementations/digests/sha2_prov.c"
        ],
        "providers/implementations/digests/libdefault-lib-sha3_prov.o" => [
            "providers/implementations/digests/sha3_prov.c"
        ],
        "providers/implementations/digests/libdefault-lib-sm3_prov.o" => [
            "providers/implementations/digests/sm3_prov.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_der2key.o" => [
            "providers/implementations/encode_decode/decode_der2key.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_epki2pki.o" => [
            "providers/implementations/encode_decode/decode_epki2pki.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_msblob2key.o" => [
            "providers/implementations/encode_decode/decode_msblob2key.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_pem2der.o" => [
            "providers/implementations/encode_decode/decode_pem2der.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_pvk2key.o" => [
            "providers/implementations/encode_decode/decode_pvk2key.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-decode_spki2typespki.o" => [
            "providers/implementations/encode_decode/decode_spki2typespki.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2any.o" => [
            "providers/implementations/encode_decode/encode_key2any.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2blob.o" => [
            "providers/implementations/encode_decode/encode_key2blob.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2ms.o" => [
            "providers/implementations/encode_decode/encode_key2ms.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-encode_key2text.o" => [
            "providers/implementations/encode_decode/encode_key2text.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-endecoder_common.o" => [
            "providers/implementations/encode_decode/endecoder_common.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-ml_common_codecs.o" => [
            "providers/implementations/encode_decode/ml_common_codecs.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-ml_dsa_codecs.o" => [
            "providers/implementations/encode_decode/ml_dsa_codecs.c"
        ],
        "providers/implementations/encode_decode/libdefault-lib-ml_kem_codecs.o" => [
            "providers/implementations/encode_decode/ml_kem_codecs.c"
        ],
        "providers/implementations/exchange/libdefault-lib-dh_exch.o" => [
            "providers/implementations/exchange/dh_exch.c"
        ],
        "providers/implementations/exchange/libdefault-lib-ecdh_exch.o" => [
            "providers/implementations/exchange/ecdh_exch.c"
        ],
        "providers/implementations/exchange/libdefault-lib-ecx_exch.o" => [
            "providers/implementations/exchange/ecx_exch.c"
        ],
        "providers/implementations/exchange/libdefault-lib-kdf_exch.o" => [
            "providers/implementations/exchange/kdf_exch.c"
        ],
        "providers/implementations/exchange/libdefault-lib-sm2dh_exch.o" => [
            "providers/implementations/exchange/sm2dh_exch.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-hkdf.o" => [
            "providers/implementations/kdfs/hkdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-hmacdrbg_kdf.o" => [
            "providers/implementations/kdfs/hmacdrbg_kdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-kbkdf.o" => [
            "providers/implementations/kdfs/kbkdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-krb5kdf.o" => [
            "providers/implementations/kdfs/krb5kdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-pbkdf2.o" => [
            "providers/implementations/kdfs/pbkdf2.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-pbkdf2_fips.o" => [
            "providers/implementations/kdfs/pbkdf2_fips.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-pkcs12kdf.o" => [
            "providers/implementations/kdfs/pkcs12kdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-scrypt.o" => [
            "providers/implementations/kdfs/scrypt.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-sshkdf.o" => [
            "providers/implementations/kdfs/sshkdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-sskdf.o" => [
            "providers/implementations/kdfs/sskdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-tls1_prf.o" => [
            "providers/implementations/kdfs/tls1_prf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-wbsm4kdf.o" => [
            "providers/implementations/kdfs/wbsm4kdf.c"
        ],
        "providers/implementations/kdfs/libdefault-lib-x942kdf.o" => [
            "providers/implementations/kdfs/x942kdf.c"
        ],
        "providers/implementations/kdfs/liblegacy-lib-pbkdf1.o" => [
            "providers/implementations/kdfs/pbkdf1.c"
        ],
        "providers/implementations/kdfs/liblegacy-lib-pvkkdf.o" => [
            "providers/implementations/kdfs/pvkkdf.c"
        ],
        "providers/implementations/kem/libdefault-lib-ec_kem.o" => [
            "providers/implementations/kem/ec_kem.c"
        ],
        "providers/implementations/kem/libdefault-lib-ecx_kem.o" => [
            "providers/implementations/kem/ecx_kem.c"
        ],
        "providers/implementations/kem/libdefault-lib-kem_util.o" => [
            "providers/implementations/kem/kem_util.c"
        ],
        "providers/implementations/kem/libdefault-lib-ml_kem_kem.o" => [
            "providers/implementations/kem/ml_kem_kem.c"
        ],
        "providers/implementations/kem/libdefault-lib-mlx_kem.o" => [
            "providers/implementations/kem/mlx_kem.c"
        ],
        "providers/implementations/kem/libdefault-lib-rsa_kem.o" => [
            "providers/implementations/kem/rsa_kem.c"
        ],
        "providers/implementations/kem/libtemplate-lib-template_kem.o" => [
            "providers/implementations/kem/template_kem.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-dh_kmgmt.o" => [
            "providers/implementations/keymgmt/dh_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-dsa_kmgmt.o" => [
            "providers/implementations/keymgmt/dsa_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-ec_kmgmt.o" => [
            "providers/implementations/keymgmt/ec_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-ecx_kmgmt.o" => [
            "providers/implementations/keymgmt/ecx_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-kdf_legacy_kmgmt.o" => [
            "providers/implementations/keymgmt/kdf_legacy_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-mac_legacy_kmgmt.o" => [
            "providers/implementations/keymgmt/mac_legacy_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-ml_dsa_kmgmt.o" => [
            "providers/implementations/keymgmt/ml_dsa_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-ml_kem_kmgmt.o" => [
            "providers/implementations/keymgmt/ml_kem_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-mlx_kmgmt.o" => [
            "providers/implementations/keymgmt/mlx_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-rsa_kmgmt.o" => [
            "providers/implementations/keymgmt/rsa_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libdefault-lib-slh_dsa_kmgmt.o" => [
            "providers/implementations/keymgmt/slh_dsa_kmgmt.c"
        ],
        "providers/implementations/keymgmt/libtemplate-lib-template_kmgmt.o" => [
            "providers/implementations/keymgmt/template_kmgmt.c"
        ],
        "providers/implementations/macs/libdefault-lib-cmac_prov.o" => [
            "providers/implementations/macs/cmac_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-eia3_prov.o" => [
            "providers/implementations/macs/eia3_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-gmac_prov.o" => [
            "providers/implementations/macs/gmac_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-hmac_prov.o" => [
            "providers/implementations/macs/hmac_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-kmac_prov.o" => [
            "providers/implementations/macs/kmac_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-poly1305_prov.o" => [
            "providers/implementations/macs/poly1305_prov.c"
        ],
        "providers/implementations/macs/libdefault-lib-siphash_prov.o" => [
            "providers/implementations/macs/siphash_prov.c"
        ],
        "providers/implementations/rands/libdefault-lib-drbg.o" => [
            "providers/implementations/rands/drbg.c"
        ],
        "providers/implementations/rands/libdefault-lib-drbg_ctr.o" => [
            "providers/implementations/rands/drbg_ctr.c"
        ],
        "providers/implementations/rands/libdefault-lib-drbg_hash.o" => [
            "providers/implementations/rands/drbg_hash.c"
        ],
        "providers/implementations/rands/libdefault-lib-drbg_hmac.o" => [
            "providers/implementations/rands/drbg_hmac.c"
        ],
        "providers/implementations/rands/libdefault-lib-seed_src.o" => [
            "providers/implementations/rands/seed_src.c"
        ],
        "providers/implementations/rands/libdefault-lib-seed_src_jitter.o" => [
            "providers/implementations/rands/seed_src_jitter.c"
        ],
        "providers/implementations/rands/libdefault-lib-test_rng.o" => [
            "providers/implementations/rands/test_rng.c"
        ],
        "providers/implementations/rands/seeding/libdefault-lib-rand_cpu_x86.o" => [
            "providers/implementations/rands/seeding/rand_cpu_x86.c"
        ],
        "providers/implementations/rands/seeding/libdefault-lib-rand_tsc.o" => [
            "providers/implementations/rands/seeding/rand_tsc.c"
        ],
        "providers/implementations/rands/seeding/libdefault-lib-rand_unix.o" => [
            "providers/implementations/rands/seeding/rand_unix.c"
        ],
        "providers/implementations/rands/seeding/libdefault-lib-rand_win.o" => [
            "providers/implementations/rands/seeding/rand_win.c"
        ],
        "providers/implementations/signature/libdefault-lib-dsa_sig.o" => [
            "providers/implementations/signature/dsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-ecdsa_sig.o" => [
            "providers/implementations/signature/ecdsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-eddsa_sig.o" => [
            "providers/implementations/signature/eddsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-mac_legacy_sig.o" => [
            "providers/implementations/signature/mac_legacy_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-ml_dsa_sig.o" => [
            "providers/implementations/signature/ml_dsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-rsa_sig.o" => [
            "providers/implementations/signature/rsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-slh_dsa_sig.o" => [
            "providers/implementations/signature/slh_dsa_sig.c"
        ],
        "providers/implementations/signature/libdefault-lib-sm2_sig.o" => [
            "providers/implementations/signature/sm2_sig.c"
        ],
        "providers/implementations/skeymgmt/libdefault-lib-aes_skmgmt.o" => [
            "providers/implementations/skeymgmt/aes_skmgmt.c"
        ],
        "providers/implementations/skeymgmt/libdefault-lib-generic.o" => [
            "providers/implementations/skeymgmt/generic.c"
        ],
        "providers/implementations/storemgmt/libdefault-lib-file_store.o" => [
            "providers/implementations/storemgmt/file_store.c"
        ],
        "providers/implementations/storemgmt/libdefault-lib-file_store_any2obj.o" => [
            "providers/implementations/storemgmt/file_store_any2obj.c"
        ],
        "providers/legacy" => [
            "crypto/legacy-dso-cpuid.o",
            "crypto/legacy-dso-ctype.o",
            "crypto/legacy-dso-x86_64cpuid.o",
            "providers/legacy-dso-legacyprov.o",
            "providers/legacy.ld"
        ],
        "providers/legacy-dso-legacyprov.o" => [
            "providers/legacyprov.c"
        ],
        "providers/libcommon.a" => [
            "providers/common/der/libcommon-lib-der_digests_gen.o",
            "providers/common/der/libcommon-lib-der_dsa_gen.o",
            "providers/common/der/libcommon-lib-der_dsa_key.o",
            "providers/common/der/libcommon-lib-der_dsa_sig.o",
            "providers/common/der/libcommon-lib-der_ec_gen.o",
            "providers/common/der/libcommon-lib-der_ec_key.o",
            "providers/common/der/libcommon-lib-der_ec_sig.o",
            "providers/common/der/libcommon-lib-der_ecx_gen.o",
            "providers/common/der/libcommon-lib-der_ecx_key.o",
            "providers/common/der/libcommon-lib-der_ml_dsa_gen.o",
            "providers/common/der/libcommon-lib-der_ml_dsa_key.o",
            "providers/common/der/libcommon-lib-der_rsa_gen.o",
            "providers/common/der/libcommon-lib-der_rsa_key.o",
            "providers/common/der/libcommon-lib-der_slh_dsa_gen.o",
            "providers/common/der/libcommon-lib-der_slh_dsa_key.o",
            "providers/common/der/libcommon-lib-der_wrap_gen.o",
            "providers/common/libcommon-lib-provider_ctx.o",
            "providers/common/libcommon-lib-provider_err.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_block.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_ccm_hw.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_gcm_hw.o",
            "providers/implementations/ciphers/libcommon-lib-ciphercommon_hw.o",
            "providers/implementations/digests/libcommon-lib-digestcommon.o",
            "ssl/record/methods/libcommon-lib-tls_pad.o"
        ],
        "providers/libcrypto-lib-baseprov.o" => [
            "providers/baseprov.c"
        ],
        "providers/libcrypto-lib-defltprov.o" => [
            "providers/defltprov.c"
        ],
        "providers/libcrypto-lib-nullprov.o" => [
            "providers/nullprov.c"
        ],
        "providers/libcrypto-lib-prov_running.o" => [
            "providers/prov_running.c"
        ],
        "providers/libcrypto-shlib-baseprov.o" => [
            "providers/baseprov.c"
        ],
        "providers/libcrypto-shlib-defltprov.o" => [
            "providers/defltprov.c"
        ],
        "providers/libcrypto-shlib-nullprov.o" => [
            "providers/nullprov.c"
        ],
        "providers/libcrypto-shlib-prov_running.o" => [
            "providers/prov_running.c"
        ],
        "providers/libdefault.a" => [
            "providers/common/der/libdefault-lib-der_rsa_sig.o",
            "providers/common/der/libdefault-lib-der_sm2_gen.o",
            "providers/common/der/libdefault-lib-der_sm2_key.o",
            "providers/common/der/libdefault-lib-der_sm2_sig.o",
            "providers/common/libdefault-lib-bio_prov.o",
            "providers/common/libdefault-lib-capabilities.o",
            "providers/common/libdefault-lib-digest_to_nid.o",
            "providers/common/libdefault-lib-provider_seeding.o",
            "providers/common/libdefault-lib-provider_util.o",
            "providers/common/libdefault-lib-securitycheck.o",
            "providers/common/libdefault-lib-securitycheck_default.o",
            "providers/implementations/asymciphers/libdefault-lib-rsa_enc.o",
            "providers/implementations/asymciphers/libdefault-lib-sm2_enc.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha1_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_cbc_hmac_sha256_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_ccm_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_gcm_siv_polyval.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_ocb_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_siv_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_wrp.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_fips.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_aes_xts_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_chacha20.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_chacha20_poly1305_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_cts.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_null.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_ccm_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_gcm_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_sm4_xts_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_common.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_default_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_tdes_wrap_hw.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3.o",
            "providers/implementations/ciphers/libdefault-lib-cipher_zuc_eea3_hw.o",
            "providers/implementations/digests/libdefault-lib-md5_prov.o",
            "providers/implementations/digests/libdefault-lib-md5_sha1_prov.o",
            "providers/implementations/digests/libdefault-lib-null_prov.o",
            "providers/implementations/digests/libdefault-lib-sha2_prov.o",
            "providers/implementations/digests/libdefault-lib-sha3_prov.o",
            "providers/implementations/digests/libdefault-lib-sm3_prov.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_der2key.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_epki2pki.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_msblob2key.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_pem2der.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_pvk2key.o",
            "providers/implementations/encode_decode/libdefault-lib-decode_spki2typespki.o",
            "providers/implementations/encode_decode/libdefault-lib-encode_key2any.o",
            "providers/implementations/encode_decode/libdefault-lib-encode_key2blob.o",
            "providers/implementations/encode_decode/libdefault-lib-encode_key2ms.o",
            "providers/implementations/encode_decode/libdefault-lib-encode_key2text.o",
            "providers/implementations/encode_decode/libdefault-lib-endecoder_common.o",
            "providers/implementations/encode_decode/libdefault-lib-ml_common_codecs.o",
            "providers/implementations/encode_decode/libdefault-lib-ml_dsa_codecs.o",
            "providers/implementations/encode_decode/libdefault-lib-ml_kem_codecs.o",
            "providers/implementations/exchange/libdefault-lib-dh_exch.o",
            "providers/implementations/exchange/libdefault-lib-ecdh_exch.o",
            "providers/implementations/exchange/libdefault-lib-ecx_exch.o",
            "providers/implementations/exchange/libdefault-lib-kdf_exch.o",
            "providers/implementations/exchange/libdefault-lib-sm2dh_exch.o",
            "providers/implementations/kdfs/libdefault-lib-hkdf.o",
            "providers/implementations/kdfs/libdefault-lib-hmacdrbg_kdf.o",
            "providers/implementations/kdfs/libdefault-lib-kbkdf.o",
            "providers/implementations/kdfs/libdefault-lib-krb5kdf.o",
            "providers/implementations/kdfs/libdefault-lib-pbkdf2.o",
            "providers/implementations/kdfs/libdefault-lib-pbkdf2_fips.o",
            "providers/implementations/kdfs/libdefault-lib-pkcs12kdf.o",
            "providers/implementations/kdfs/libdefault-lib-scrypt.o",
            "providers/implementations/kdfs/libdefault-lib-sshkdf.o",
            "providers/implementations/kdfs/libdefault-lib-sskdf.o",
            "providers/implementations/kdfs/libdefault-lib-tls1_prf.o",
            "providers/implementations/kdfs/libdefault-lib-wbsm4kdf.o",
            "providers/implementations/kdfs/libdefault-lib-x942kdf.o",
            "providers/implementations/kem/libdefault-lib-ec_kem.o",
            "providers/implementations/kem/libdefault-lib-ecx_kem.o",
            "providers/implementations/kem/libdefault-lib-kem_util.o",
            "providers/implementations/kem/libdefault-lib-ml_kem_kem.o",
            "providers/implementations/kem/libdefault-lib-mlx_kem.o",
            "providers/implementations/kem/libdefault-lib-rsa_kem.o",
            "providers/implementations/keymgmt/libdefault-lib-dh_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-dsa_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-ec_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-ecx_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-kdf_legacy_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-mac_legacy_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-ml_dsa_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-ml_kem_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-mlx_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-rsa_kmgmt.o",
            "providers/implementations/keymgmt/libdefault-lib-slh_dsa_kmgmt.o",
            "providers/implementations/macs/libdefault-lib-cmac_prov.o",
            "providers/implementations/macs/libdefault-lib-eia3_prov.o",
            "providers/implementations/macs/libdefault-lib-gmac_prov.o",
            "providers/implementations/macs/libdefault-lib-hmac_prov.o",
            "providers/implementations/macs/libdefault-lib-kmac_prov.o",
            "providers/implementations/macs/libdefault-lib-poly1305_prov.o",
            "providers/implementations/macs/libdefault-lib-siphash_prov.o",
            "providers/implementations/rands/libdefault-lib-drbg.o",
            "providers/implementations/rands/libdefault-lib-drbg_ctr.o",
            "providers/implementations/rands/libdefault-lib-drbg_hash.o",
            "providers/implementations/rands/libdefault-lib-drbg_hmac.o",
            "providers/implementations/rands/libdefault-lib-seed_src.o",
            "providers/implementations/rands/libdefault-lib-seed_src_jitter.o",
            "providers/implementations/rands/libdefault-lib-test_rng.o",
            "providers/implementations/rands/seeding/libdefault-lib-rand_cpu_x86.o",
            "providers/implementations/rands/seeding/libdefault-lib-rand_tsc.o",
            "providers/implementations/rands/seeding/libdefault-lib-rand_unix.o",
            "providers/implementations/rands/seeding/libdefault-lib-rand_win.o",
            "providers/implementations/signature/libdefault-lib-dsa_sig.o",
            "providers/implementations/signature/libdefault-lib-ecdsa_sig.o",
            "providers/implementations/signature/libdefault-lib-eddsa_sig.o",
            "providers/implementations/signature/libdefault-lib-mac_legacy_sig.o",
            "providers/implementations/signature/libdefault-lib-ml_dsa_sig.o",
            "providers/implementations/signature/libdefault-lib-rsa_sig.o",
            "providers/implementations/signature/libdefault-lib-slh_dsa_sig.o",
            "providers/implementations/signature/libdefault-lib-sm2_sig.o",
            "providers/implementations/skeymgmt/libdefault-lib-aes_skmgmt.o",
            "providers/implementations/skeymgmt/libdefault-lib-generic.o",
            "providers/implementations/storemgmt/libdefault-lib-file_store.o",
            "providers/implementations/storemgmt/libdefault-lib-file_store_any2obj.o",
            "ssl/record/methods/libdefault-lib-ssl3_cbc.o"
        ],
        "providers/liblegacy-lib-prov_running.o" => [
            "providers/prov_running.c"
        ],
        "providers/liblegacy.a" => [
            "crypto/des/liblegacy-lib-des_enc.o",
            "crypto/des/liblegacy-lib-fcrypt_b.o",
            "crypto/md5/liblegacy-lib-md5-x86_64.o",
            "crypto/md5/liblegacy-lib-md5_dgst.o",
            "crypto/md5/liblegacy-lib-md5_one.o",
            "crypto/md5/liblegacy-lib-md5_sha1.o",
            "crypto/rc4/liblegacy-lib-rc4-md5-x86_64.o",
            "crypto/rc4/liblegacy-lib-rc4-x86_64.o",
            "providers/common/liblegacy-lib-provider_util.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_des.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_des_hw.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_desx.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_desx_hw.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_rc4.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hmac_md5_hw.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_rc4_hw.o",
            "providers/implementations/ciphers/liblegacy-lib-cipher_tdes_common.o",
            "providers/implementations/kdfs/liblegacy-lib-pbkdf1.o",
            "providers/implementations/kdfs/liblegacy-lib-pvkkdf.o",
            "providers/liblegacy-lib-prov_running.o"
        ],
        "providers/libtemplate.a" => [
            "providers/implementations/kem/libtemplate-lib-template_kem.o",
            "providers/implementations/keymgmt/libtemplate-lib-template_kmgmt.o"
        ],
        "ssl/libssl-lib-bio_ssl.o" => [
            "ssl/bio_ssl.c"
        ],
        "ssl/libssl-lib-d1_lib.o" => [
            "ssl/d1_lib.c"
        ],
        "ssl/libssl-lib-d1_msg.o" => [
            "ssl/d1_msg.c"
        ],
        "ssl/libssl-lib-d1_srtp.o" => [
            "ssl/d1_srtp.c"
        ],
        "ssl/libssl-lib-methods.o" => [
            "ssl/methods.c"
        ],
        "ssl/libssl-lib-pqueue.o" => [
            "ssl/pqueue.c"
        ],
        "ssl/libssl-lib-priority_queue.o" => [
            "ssl/priority_queue.c"
        ],
        "ssl/libssl-lib-s3_enc.o" => [
            "ssl/s3_enc.c"
        ],
        "ssl/libssl-lib-s3_lib.o" => [
            "ssl/s3_lib.c"
        ],
        "ssl/libssl-lib-s3_msg.o" => [
            "ssl/s3_msg.c"
        ],
        "ssl/libssl-lib-ssl_asn1.o" => [
            "ssl/ssl_asn1.c"
        ],
        "ssl/libssl-lib-ssl_cert.o" => [
            "ssl/ssl_cert.c"
        ],
        "ssl/libssl-lib-ssl_cert_comp.o" => [
            "ssl/ssl_cert_comp.c"
        ],
        "ssl/libssl-lib-ssl_ciph.o" => [
            "ssl/ssl_ciph.c"
        ],
        "ssl/libssl-lib-ssl_conf.o" => [
            "ssl/ssl_conf.c"
        ],
        "ssl/libssl-lib-ssl_err_legacy.o" => [
            "ssl/ssl_err_legacy.c"
        ],
        "ssl/libssl-lib-ssl_init.o" => [
            "ssl/ssl_init.c"
        ],
        "ssl/libssl-lib-ssl_lib.o" => [
            "ssl/ssl_lib.c"
        ],
        "ssl/libssl-lib-ssl_mcnf.o" => [
            "ssl/ssl_mcnf.c"
        ],
        "ssl/libssl-lib-ssl_rsa.o" => [
            "ssl/ssl_rsa.c"
        ],
        "ssl/libssl-lib-ssl_rsa_legacy.o" => [
            "ssl/ssl_rsa_legacy.c"
        ],
        "ssl/libssl-lib-ssl_sess.o" => [
            "ssl/ssl_sess.c"
        ],
        "ssl/libssl-lib-ssl_stat.o" => [
            "ssl/ssl_stat.c"
        ],
        "ssl/libssl-lib-ssl_txt.o" => [
            "ssl/ssl_txt.c"
        ],
        "ssl/libssl-lib-ssl_utst.o" => [
            "ssl/ssl_utst.c"
        ],
        "ssl/libssl-lib-t1_enc.o" => [
            "ssl/t1_enc.c"
        ],
        "ssl/libssl-lib-t1_lib.o" => [
            "ssl/t1_lib.c"
        ],
        "ssl/libssl-lib-t1_trce.o" => [
            "ssl/t1_trce.c"
        ],
        "ssl/libssl-lib-tls13_enc.o" => [
            "ssl/tls13_enc.c"
        ],
        "ssl/libssl-lib-tls_depr.o" => [
            "ssl/tls_depr.c"
        ],
        "ssl/libssl-lib-tls_srp.o" => [
            "ssl/tls_srp.c"
        ],
        "ssl/libssl-shlib-bio_ssl.o" => [
            "ssl/bio_ssl.c"
        ],
        "ssl/libssl-shlib-d1_lib.o" => [
            "ssl/d1_lib.c"
        ],
        "ssl/libssl-shlib-d1_msg.o" => [
            "ssl/d1_msg.c"
        ],
        "ssl/libssl-shlib-d1_srtp.o" => [
            "ssl/d1_srtp.c"
        ],
        "ssl/libssl-shlib-methods.o" => [
            "ssl/methods.c"
        ],
        "ssl/libssl-shlib-pqueue.o" => [
            "ssl/pqueue.c"
        ],
        "ssl/libssl-shlib-priority_queue.o" => [
            "ssl/priority_queue.c"
        ],
        "ssl/libssl-shlib-s3_enc.o" => [
            "ssl/s3_enc.c"
        ],
        "ssl/libssl-shlib-s3_lib.o" => [
            "ssl/s3_lib.c"
        ],
        "ssl/libssl-shlib-s3_msg.o" => [
            "ssl/s3_msg.c"
        ],
        "ssl/libssl-shlib-ssl_asn1.o" => [
            "ssl/ssl_asn1.c"
        ],
        "ssl/libssl-shlib-ssl_cert.o" => [
            "ssl/ssl_cert.c"
        ],
        "ssl/libssl-shlib-ssl_cert_comp.o" => [
            "ssl/ssl_cert_comp.c"
        ],
        "ssl/libssl-shlib-ssl_ciph.o" => [
            "ssl/ssl_ciph.c"
        ],
        "ssl/libssl-shlib-ssl_conf.o" => [
            "ssl/ssl_conf.c"
        ],
        "ssl/libssl-shlib-ssl_err_legacy.o" => [
            "ssl/ssl_err_legacy.c"
        ],
        "ssl/libssl-shlib-ssl_init.o" => [
            "ssl/ssl_init.c"
        ],
        "ssl/libssl-shlib-ssl_lib.o" => [
            "ssl/ssl_lib.c"
        ],
        "ssl/libssl-shlib-ssl_mcnf.o" => [
            "ssl/ssl_mcnf.c"
        ],
        "ssl/libssl-shlib-ssl_rsa.o" => [
            "ssl/ssl_rsa.c"
        ],
        "ssl/libssl-shlib-ssl_rsa_legacy.o" => [
            "ssl/ssl_rsa_legacy.c"
        ],
        "ssl/libssl-shlib-ssl_sess.o" => [
            "ssl/ssl_sess.c"
        ],
        "ssl/libssl-shlib-ssl_stat.o" => [
            "ssl/ssl_stat.c"
        ],
        "ssl/libssl-shlib-ssl_txt.o" => [
            "ssl/ssl_txt.c"
        ],
        "ssl/libssl-shlib-ssl_utst.o" => [
            "ssl/ssl_utst.c"
        ],
        "ssl/libssl-shlib-t1_enc.o" => [
            "ssl/t1_enc.c"
        ],
        "ssl/libssl-shlib-t1_lib.o" => [
            "ssl/t1_lib.c"
        ],
        "ssl/libssl-shlib-t1_trce.o" => [
            "ssl/t1_trce.c"
        ],
        "ssl/libssl-shlib-tls13_enc.o" => [
            "ssl/tls13_enc.c"
        ],
        "ssl/libssl-shlib-tls_depr.o" => [
            "ssl/tls_depr.c"
        ],
        "ssl/libssl-shlib-tls_srp.o" => [
            "ssl/tls_srp.c"
        ],
        "ssl/quic/libssl-lib-cc_newreno.o" => [
            "ssl/quic/cc_newreno.c"
        ],
        "ssl/quic/libssl-lib-json_enc.o" => [
            "ssl/quic/json_enc.c"
        ],
        "ssl/quic/libssl-lib-qlog.o" => [
            "ssl/quic/qlog.c"
        ],
        "ssl/quic/libssl-lib-qlog_event_helpers.o" => [
            "ssl/quic/qlog_event_helpers.c"
        ],
        "ssl/quic/libssl-lib-quic_ackm.o" => [
            "ssl/quic/quic_ackm.c"
        ],
        "ssl/quic/libssl-lib-quic_cfq.o" => [
            "ssl/quic/quic_cfq.c"
        ],
        "ssl/quic/libssl-lib-quic_channel.o" => [
            "ssl/quic/quic_channel.c"
        ],
        "ssl/quic/libssl-lib-quic_demux.o" => [
            "ssl/quic/quic_demux.c"
        ],
        "ssl/quic/libssl-lib-quic_engine.o" => [
            "ssl/quic/quic_engine.c"
        ],
        "ssl/quic/libssl-lib-quic_fc.o" => [
            "ssl/quic/quic_fc.c"
        ],
        "ssl/quic/libssl-lib-quic_fifd.o" => [
            "ssl/quic/quic_fifd.c"
        ],
        "ssl/quic/libssl-lib-quic_impl.o" => [
            "ssl/quic/quic_impl.c"
        ],
        "ssl/quic/libssl-lib-quic_lcidm.o" => [
            "ssl/quic/quic_lcidm.c"
        ],
        "ssl/quic/libssl-lib-quic_method.o" => [
            "ssl/quic/quic_method.c"
        ],
        "ssl/quic/libssl-lib-quic_obj.o" => [
            "ssl/quic/quic_obj.c"
        ],
        "ssl/quic/libssl-lib-quic_port.o" => [
            "ssl/quic/quic_port.c"
        ],
        "ssl/quic/libssl-lib-quic_rcidm.o" => [
            "ssl/quic/quic_rcidm.c"
        ],
        "ssl/quic/libssl-lib-quic_reactor.o" => [
            "ssl/quic/quic_reactor.c"
        ],
        "ssl/quic/libssl-lib-quic_reactor_wait_ctx.o" => [
            "ssl/quic/quic_reactor_wait_ctx.c"
        ],
        "ssl/quic/libssl-lib-quic_record_rx.o" => [
            "ssl/quic/quic_record_rx.c"
        ],
        "ssl/quic/libssl-lib-quic_record_shared.o" => [
            "ssl/quic/quic_record_shared.c"
        ],
        "ssl/quic/libssl-lib-quic_record_tx.o" => [
            "ssl/quic/quic_record_tx.c"
        ],
        "ssl/quic/libssl-lib-quic_record_util.o" => [
            "ssl/quic/quic_record_util.c"
        ],
        "ssl/quic/libssl-lib-quic_rstream.o" => [
            "ssl/quic/quic_rstream.c"
        ],
        "ssl/quic/libssl-lib-quic_rx_depack.o" => [
            "ssl/quic/quic_rx_depack.c"
        ],
        "ssl/quic/libssl-lib-quic_sf_list.o" => [
            "ssl/quic/quic_sf_list.c"
        ],
        "ssl/quic/libssl-lib-quic_srt_gen.o" => [
            "ssl/quic/quic_srt_gen.c"
        ],
        "ssl/quic/libssl-lib-quic_srtm.o" => [
            "ssl/quic/quic_srtm.c"
        ],
        "ssl/quic/libssl-lib-quic_sstream.o" => [
            "ssl/quic/quic_sstream.c"
        ],
        "ssl/quic/libssl-lib-quic_statm.o" => [
            "ssl/quic/quic_statm.c"
        ],
        "ssl/quic/libssl-lib-quic_stream_map.o" => [
            "ssl/quic/quic_stream_map.c"
        ],
        "ssl/quic/libssl-lib-quic_thread_assist.o" => [
            "ssl/quic/quic_thread_assist.c"
        ],
        "ssl/quic/libssl-lib-quic_tls.o" => [
            "ssl/quic/quic_tls.c"
        ],
        "ssl/quic/libssl-lib-quic_tls_api.o" => [
            "ssl/quic/quic_tls_api.c"
        ],
        "ssl/quic/libssl-lib-quic_tls_api_method.o" => [
            "ssl/quic/quic_tls_api_method.c"
        ],
        "ssl/quic/libssl-lib-quic_trace.o" => [
            "ssl/quic/quic_trace.c"
        ],
        "ssl/quic/libssl-lib-quic_tserver.o" => [
            "ssl/quic/quic_tserver.c"
        ],
        "ssl/quic/libssl-lib-quic_txp.o" => [
            "ssl/quic/quic_txp.c"
        ],
        "ssl/quic/libssl-lib-quic_txpim.o" => [
            "ssl/quic/quic_txpim.c"
        ],
        "ssl/quic/libssl-lib-quic_types.o" => [
            "ssl/quic/quic_types.c"
        ],
        "ssl/quic/libssl-lib-quic_wire.o" => [
            "ssl/quic/quic_wire.c"
        ],
        "ssl/quic/libssl-lib-quic_wire_pkt.o" => [
            "ssl/quic/quic_wire_pkt.c"
        ],
        "ssl/quic/libssl-lib-uint_set.o" => [
            "ssl/quic/uint_set.c"
        ],
        "ssl/quic/libssl-shlib-cc_newreno.o" => [
            "ssl/quic/cc_newreno.c"
        ],
        "ssl/quic/libssl-shlib-json_enc.o" => [
            "ssl/quic/json_enc.c"
        ],
        "ssl/quic/libssl-shlib-qlog.o" => [
            "ssl/quic/qlog.c"
        ],
        "ssl/quic/libssl-shlib-qlog_event_helpers.o" => [
            "ssl/quic/qlog_event_helpers.c"
        ],
        "ssl/quic/libssl-shlib-quic_ackm.o" => [
            "ssl/quic/quic_ackm.c"
        ],
        "ssl/quic/libssl-shlib-quic_cfq.o" => [
            "ssl/quic/quic_cfq.c"
        ],
        "ssl/quic/libssl-shlib-quic_channel.o" => [
            "ssl/quic/quic_channel.c"
        ],
        "ssl/quic/libssl-shlib-quic_demux.o" => [
            "ssl/quic/quic_demux.c"
        ],
        "ssl/quic/libssl-shlib-quic_engine.o" => [
            "ssl/quic/quic_engine.c"
        ],
        "ssl/quic/libssl-shlib-quic_fc.o" => [
            "ssl/quic/quic_fc.c"
        ],
        "ssl/quic/libssl-shlib-quic_fifd.o" => [
            "ssl/quic/quic_fifd.c"
        ],
        "ssl/quic/libssl-shlib-quic_impl.o" => [
            "ssl/quic/quic_impl.c"
        ],
        "ssl/quic/libssl-shlib-quic_lcidm.o" => [
            "ssl/quic/quic_lcidm.c"
        ],
        "ssl/quic/libssl-shlib-quic_method.o" => [
            "ssl/quic/quic_method.c"
        ],
        "ssl/quic/libssl-shlib-quic_obj.o" => [
            "ssl/quic/quic_obj.c"
        ],
        "ssl/quic/libssl-shlib-quic_port.o" => [
            "ssl/quic/quic_port.c"
        ],
        "ssl/quic/libssl-shlib-quic_rcidm.o" => [
            "ssl/quic/quic_rcidm.c"
        ],
        "ssl/quic/libssl-shlib-quic_reactor.o" => [
            "ssl/quic/quic_reactor.c"
        ],
        "ssl/quic/libssl-shlib-quic_reactor_wait_ctx.o" => [
            "ssl/quic/quic_reactor_wait_ctx.c"
        ],
        "ssl/quic/libssl-shlib-quic_record_rx.o" => [
            "ssl/quic/quic_record_rx.c"
        ],
        "ssl/quic/libssl-shlib-quic_record_shared.o" => [
            "ssl/quic/quic_record_shared.c"
        ],
        "ssl/quic/libssl-shlib-quic_record_tx.o" => [
            "ssl/quic/quic_record_tx.c"
        ],
        "ssl/quic/libssl-shlib-quic_record_util.o" => [
            "ssl/quic/quic_record_util.c"
        ],
        "ssl/quic/libssl-shlib-quic_rstream.o" => [
            "ssl/quic/quic_rstream.c"
        ],
        "ssl/quic/libssl-shlib-quic_rx_depack.o" => [
            "ssl/quic/quic_rx_depack.c"
        ],
        "ssl/quic/libssl-shlib-quic_sf_list.o" => [
            "ssl/quic/quic_sf_list.c"
        ],
        "ssl/quic/libssl-shlib-quic_srt_gen.o" => [
            "ssl/quic/quic_srt_gen.c"
        ],
        "ssl/quic/libssl-shlib-quic_srtm.o" => [
            "ssl/quic/quic_srtm.c"
        ],
        "ssl/quic/libssl-shlib-quic_sstream.o" => [
            "ssl/quic/quic_sstream.c"
        ],
        "ssl/quic/libssl-shlib-quic_statm.o" => [
            "ssl/quic/quic_statm.c"
        ],
        "ssl/quic/libssl-shlib-quic_stream_map.o" => [
            "ssl/quic/quic_stream_map.c"
        ],
        "ssl/quic/libssl-shlib-quic_thread_assist.o" => [
            "ssl/quic/quic_thread_assist.c"
        ],
        "ssl/quic/libssl-shlib-quic_tls.o" => [
            "ssl/quic/quic_tls.c"
        ],
        "ssl/quic/libssl-shlib-quic_tls_api.o" => [
            "ssl/quic/quic_tls_api.c"
        ],
        "ssl/quic/libssl-shlib-quic_tls_api_method.o" => [
            "ssl/quic/quic_tls_api_method.c"
        ],
        "ssl/quic/libssl-shlib-quic_trace.o" => [
            "ssl/quic/quic_trace.c"
        ],
        "ssl/quic/libssl-shlib-quic_tserver.o" => [
            "ssl/quic/quic_tserver.c"
        ],
        "ssl/quic/libssl-shlib-quic_txp.o" => [
            "ssl/quic/quic_txp.c"
        ],
        "ssl/quic/libssl-shlib-quic_txpim.o" => [
            "ssl/quic/quic_txpim.c"
        ],
        "ssl/quic/libssl-shlib-quic_types.o" => [
            "ssl/quic/quic_types.c"
        ],
        "ssl/quic/libssl-shlib-quic_wire.o" => [
            "ssl/quic/quic_wire.c"
        ],
        "ssl/quic/libssl-shlib-quic_wire_pkt.o" => [
            "ssl/quic/quic_wire_pkt.c"
        ],
        "ssl/quic/libssl-shlib-uint_set.o" => [
            "ssl/quic/uint_set.c"
        ],
        "ssl/record/libssl-lib-rec_layer_d1.o" => [
            "ssl/record/rec_layer_d1.c"
        ],
        "ssl/record/libssl-lib-rec_layer_s3.o" => [
            "ssl/record/rec_layer_s3.c"
        ],
        "ssl/record/libssl-shlib-rec_layer_d1.o" => [
            "ssl/record/rec_layer_d1.c"
        ],
        "ssl/record/libssl-shlib-rec_layer_s3.o" => [
            "ssl/record/rec_layer_s3.c"
        ],
        "ssl/record/methods/libcommon-lib-tls_pad.o" => [
            "ssl/record/methods/tls_pad.c"
        ],
        "ssl/record/methods/libdefault-lib-ssl3_cbc.o" => [
            "ssl/record/methods/ssl3_cbc.c"
        ],
        "ssl/record/methods/libssl-lib-dtls_meth.o" => [
            "ssl/record/methods/dtls_meth.c"
        ],
        "ssl/record/methods/libssl-lib-ssl3_meth.o" => [
            "ssl/record/methods/ssl3_meth.c"
        ],
        "ssl/record/methods/libssl-lib-tls13_meth.o" => [
            "ssl/record/methods/tls13_meth.c"
        ],
        "ssl/record/methods/libssl-lib-tls1_meth.o" => [
            "ssl/record/methods/tls1_meth.c"
        ],
        "ssl/record/methods/libssl-lib-tls_common.o" => [
            "ssl/record/methods/tls_common.c"
        ],
        "ssl/record/methods/libssl-lib-tls_multib.o" => [
            "ssl/record/methods/tls_multib.c"
        ],
        "ssl/record/methods/libssl-lib-tlsany_meth.o" => [
            "ssl/record/methods/tlsany_meth.c"
        ],
        "ssl/record/methods/libssl-shlib-dtls_meth.o" => [
            "ssl/record/methods/dtls_meth.c"
        ],
        "ssl/record/methods/libssl-shlib-ssl3_cbc.o" => [
            "ssl/record/methods/ssl3_cbc.c"
        ],
        "ssl/record/methods/libssl-shlib-ssl3_meth.o" => [
            "ssl/record/methods/ssl3_meth.c"
        ],
        "ssl/record/methods/libssl-shlib-tls13_meth.o" => [
            "ssl/record/methods/tls13_meth.c"
        ],
        "ssl/record/methods/libssl-shlib-tls1_meth.o" => [
            "ssl/record/methods/tls1_meth.c"
        ],
        "ssl/record/methods/libssl-shlib-tls_common.o" => [
            "ssl/record/methods/tls_common.c"
        ],
        "ssl/record/methods/libssl-shlib-tls_multib.o" => [
            "ssl/record/methods/tls_multib.c"
        ],
        "ssl/record/methods/libssl-shlib-tls_pad.o" => [
            "ssl/record/methods/tls_pad.c"
        ],
        "ssl/record/methods/libssl-shlib-tlsany_meth.o" => [
            "ssl/record/methods/tlsany_meth.c"
        ],
        "ssl/rio/libssl-lib-poll_builder.o" => [
            "ssl/rio/poll_builder.c"
        ],
        "ssl/rio/libssl-lib-poll_immediate.o" => [
            "ssl/rio/poll_immediate.c"
        ],
        "ssl/rio/libssl-lib-rio_notifier.o" => [
            "ssl/rio/rio_notifier.c"
        ],
        "ssl/rio/libssl-shlib-poll_builder.o" => [
            "ssl/rio/poll_builder.c"
        ],
        "ssl/rio/libssl-shlib-poll_immediate.o" => [
            "ssl/rio/poll_immediate.c"
        ],
        "ssl/rio/libssl-shlib-rio_notifier.o" => [
            "ssl/rio/rio_notifier.c"
        ],
        "ssl/statem/libssl-lib-extensions.o" => [
            "ssl/statem/extensions.c"
        ],
        "ssl/statem/libssl-lib-extensions_clnt.o" => [
            "ssl/statem/extensions_clnt.c"
        ],
        "ssl/statem/libssl-lib-extensions_cust.o" => [
            "ssl/statem/extensions_cust.c"
        ],
        "ssl/statem/libssl-lib-extensions_srvr.o" => [
            "ssl/statem/extensions_srvr.c"
        ],
        "ssl/statem/libssl-lib-statem.o" => [
            "ssl/statem/statem.c"
        ],
        "ssl/statem/libssl-lib-statem_clnt.o" => [
            "ssl/statem/statem_clnt.c"
        ],
        "ssl/statem/libssl-lib-statem_dtls.o" => [
            "ssl/statem/statem_dtls.c"
        ],
        "ssl/statem/libssl-lib-statem_lib.o" => [
            "ssl/statem/statem_lib.c"
        ],
        "ssl/statem/libssl-lib-statem_srvr.o" => [
            "ssl/statem/statem_srvr.c"
        ],
        "ssl/statem/libssl-shlib-extensions.o" => [
            "ssl/statem/extensions.c"
        ],
        "ssl/statem/libssl-shlib-extensions_clnt.o" => [
            "ssl/statem/extensions_clnt.c"
        ],
        "ssl/statem/libssl-shlib-extensions_cust.o" => [
            "ssl/statem/extensions_cust.c"
        ],
        "ssl/statem/libssl-shlib-extensions_srvr.o" => [
            "ssl/statem/extensions_srvr.c"
        ],
        "ssl/statem/libssl-shlib-statem.o" => [
            "ssl/statem/statem.c"
        ],
        "ssl/statem/libssl-shlib-statem_clnt.o" => [
            "ssl/statem/statem_clnt.c"
        ],
        "ssl/statem/libssl-shlib-statem_dtls.o" => [
            "ssl/statem/statem_dtls.c"
        ],
        "ssl/statem/libssl-shlib-statem_lib.o" => [
            "ssl/statem/statem_lib.c"
        ],
        "ssl/statem/libssl-shlib-statem_srvr.o" => [
            "ssl/statem/statem_srvr.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_extensions.o" => [
            "ssl/statem_ntls/ntls_extensions.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_extensions_clnt.o" => [
            "ssl/statem_ntls/ntls_extensions_clnt.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_extensions_cust.o" => [
            "ssl/statem_ntls/ntls_extensions_cust.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_extensions_srvr.o" => [
            "ssl/statem_ntls/ntls_extensions_srvr.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_statem.o" => [
            "ssl/statem_ntls/ntls_statem.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_statem_clnt.o" => [
            "ssl/statem_ntls/ntls_statem_clnt.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_statem_lib.o" => [
            "ssl/statem_ntls/ntls_statem_lib.c"
        ],
        "ssl/statem_ntls/libssl-lib-ntls_statem_srvr.o" => [
            "ssl/statem_ntls/ntls_statem_srvr.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_extensions.o" => [
            "ssl/statem_ntls/ntls_extensions.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_extensions_clnt.o" => [
            "ssl/statem_ntls/ntls_extensions_clnt.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_extensions_cust.o" => [
            "ssl/statem_ntls/ntls_extensions_cust.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_extensions_srvr.o" => [
            "ssl/statem_ntls/ntls_extensions_srvr.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_statem.o" => [
            "ssl/statem_ntls/ntls_statem.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_statem_clnt.o" => [
            "ssl/statem_ntls/ntls_statem_clnt.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_statem_lib.o" => [
            "ssl/statem_ntls/ntls_statem_lib.c"
        ],
        "ssl/statem_ntls/libssl-shlib-ntls_statem_srvr.o" => [
            "ssl/statem_ntls/ntls_statem_srvr.c"
        ],
        "util/shlib_wrap.sh" => [
            "util/shlib_wrap.sh.in"
        ],
        "util/wrap.pl" => [
            "util/wrap.pl.in"
        ]
    },
    "targets" => []
);

# Unexported, only used by OpenSSL::Test::Utils::available_protocols()
our %available_protocols = (
    tls  => [
    "ssl3",
    "tls1",
    "tls1_1",
    "tls1_2",
    "tls1_3"
],
    dtls => [
    "dtls1",
    "dtls1_2"
],
);

# The following data is only used when this files is use as a script
my @makevars = (
    "AR",
    "ARFLAGS",
    "AS",
    "ASFLAGS",
    "CC",
    "CFLAGS",
    "CPP",
    "CPPDEFINES",
    "CPPFLAGS",
    "CPPINCLUDES",
    "CROSS_COMPILE",
    "CXX",
    "CXXFLAGS",
    "HASHBANGPERL",
    "LD",
    "LDFLAGS",
    "LDLIBS",
    "MT",
    "MTFLAGS",
    "OBJCOPY",
    "PERL",
    "RANLIB",
    "RC",
    "RCFLAGS",
    "RM"
);
my %disabled_info = (
    "acvp-tests" => {
        "macro" => "OPENSSL_NO_ACVP_TESTS"
    },
    "argon2" => {
        "macro" => "OPENSSL_NO_ARGON2"
    },
    "aria" => {
        "macro" => "OPENSSL_NO_ARIA"
    },
    "asan" => {
        "macro" => "OPENSSL_NO_ASAN"
    },
    "atf_slibce" => {
        "macro" => "OPENSSL_NO_ATF_SLIBCE"
    },
    "bf" => {
        "macro" => "OPENSSL_NO_BF"
    },
    "blake2" => {
        "macro" => "OPENSSL_NO_BLAKE2"
    },
    "bn-method" => {
        "macro" => "OPENSSL_NO_BN_METHOD"
    },
    "brotli" => {
        "macro" => "OPENSSL_NO_BROTLI"
    },
    "brotli-dynamic" => {
        "macro" => "OPENSSL_NO_BROTLI_DYNAMIC"
    },
    "bulletproofs" => {
        "macro" => "OPENSSL_NO_BULLETPROOFS"
    },
    "camellia" => {
        "macro" => "OPENSSL_NO_CAMELLIA"
    },
    "cast" => {
        "macro" => "OPENSSL_NO_CAST"
    },
    "crypto-mdebug" => {
        "macro" => "OPENSSL_NO_CRYPTO_MDEBUG"
    },
    "crypto-mdebug-backtrace" => {
        "macro" => "OPENSSL_NO_CRYPTO_MDEBUG_BACKTRACE"
    },
    "crypto-mdebug-count" => {
        "macro" => "OPENSSL_NO_CRYPTO_MDEBUG_COUNT"
    },
    "delegated-credential" => {
        "macro" => "OPENSSL_NO_DELEGATED_CREDENTIAL"
    },
    "demos" => {
        "macro" => "OPENSSL_NO_DEMOS"
    },
    "devcryptoeng" => {
        "macro" => "OPENSSL_NO_DEVCRYPTOENG"
    },
    "ec_elgamal" => {
        "macro" => "OPENSSL_NO_EC_ELGAMAL"
    },
    "ec_nistp_64_gcc_128" => {
        "macro" => "OPENSSL_NO_EC_NISTP_64_GCC_128"
    },
    "ec_sm2p_64_gcc_128" => {
        "macro" => "OPENSSL_NO_EC_SM2P_64_GCC_128"
    },
    "egd" => {
        "macro" => "OPENSSL_NO_EGD"
    },
    "evp-cipher-api-compat" => {
        "macro" => "OPENSSL_NO_EVP_CIPHER_API_COMPAT"
    },
    "external-tests" => {
        "macro" => "OPENSSL_NO_EXTERNAL_TESTS"
    },
    "fips-jitter" => {
        "macro" => "OPENSSL_NO_FIPS_JITTER"
    },
    "fips-post" => {
        "macro" => "OPENSSL_NO_FIPS_POST"
    },
    "fips-securitychecks" => {
        "macro" => "OPENSSL_NO_FIPS_SECURITYCHECKS"
    },
    "fuzz-afl" => {
        "macro" => "OPENSSL_NO_FUZZ_AFL"
    },
    "fuzz-libfuzzer" => {
        "macro" => "OPENSSL_NO_FUZZ_LIBFUZZER"
    },
    "gost" => {
        "macro" => "OPENSSL_NO_GOST"
    },
    "h3demo" => {
        "macro" => "OPENSSL_NO_H3DEMO"
    },
    "hqinterop" => {
        "macro" => "OPENSSL_NO_HQINTEROP"
    },
    "idea" => {
        "macro" => "OPENSSL_NO_IDEA"
    },
    "jitter" => {
        "macro" => "OPENSSL_NO_JITTER"
    },
    "ktls" => {
        "macro" => "OPENSSL_NO_KTLS"
    },
    "md2" => {
        "macro" => "OPENSSL_NO_MD2"
    },
    "md4" => {
        "macro" => "OPENSSL_NO_MD4"
    },
    "mdc2" => {
        "macro" => "OPENSSL_NO_MDC2"
    },
    "msan" => {
        "macro" => "OPENSSL_NO_MSAN"
    },
    "nizk" => {
        "macro" => "OPENSSL_NO_NIZK"
    },
    "optimize-chacha-choose" => {
        "macro" => "OPENSSL_NO_OPTIMIZE_CHACHA_CHOOSE"
    },
    "paillier" => {
        "macro" => "OPENSSL_NO_PAILLIER",
        "skipped" => [
            "crypto/paillier"
        ]
    },
    "pie" => {
        "macro" => "OPENSSL_NO_PIE"
    },
    "rc2" => {
        "macro" => "OPENSSL_NO_RC2"
    },
    "rc5" => {
        "macro" => "OPENSSL_NO_RC5",
        "skipped" => [
            "crypto/rc5"
        ]
    },
    "ripemd" => {
        "macro" => "OPENSSL_NO_RIPEMD"
    },
    "rmd160" => {
        "macro" => "OPENSSL_NO_RMD160"
    },
    "sctp" => {
        "macro" => "OPENSSL_NO_SCTP"
    },
    "sdf-lib" => {
        "macro" => "OPENSSL_NO_SDF_LIB"
    },
    "sdf-lib-dynamic" => {
        "macro" => "OPENSSL_NO_SDF_LIB_DYNAMIC"
    },
    "seed" => {
        "macro" => "OPENSSL_NO_SEED"
    },
    "sm2_threshold" => {
        "macro" => "OPENSSL_NO_SM2_THRESHOLD"
    },
    "smtc" => {
        "macro" => "OPENSSL_NO_SMTC"
    },
    "smtc-debug" => {
        "macro" => "OPENSSL_NO_SMTC_DEBUG"
    },
    "ssl3" => {
        "macro" => "OPENSSL_NO_SSL3"
    },
    "ssl3-method" => {
        "macro" => "OPENSSL_NO_SSL3_METHOD"
    },
    "sslkeylog" => {
        "macro" => "OPENSSL_NO_SSLKEYLOG"
    },
    "status" => {
        "macro" => "OPENSSL_NO_STATUS"
    },
    "tests" => {
        "macro" => "OPENSSL_NO_TESTS"
    },
    "tfo" => {
        "macro" => "OPENSSL_NO_TFO"
    },
    "trace" => {
        "macro" => "OPENSSL_NO_TRACE"
    },
    "twisted_ec_elgamal" => {
        "macro" => "OPENSSL_NO_TWISTED_EC_ELGAMAL"
    },
    "ubsan" => {
        "macro" => "OPENSSL_NO_UBSAN"
    },
    "unit-test" => {
        "macro" => "OPENSSL_NO_UNIT_TEST"
    },
    "uplink" => {
        "macro" => "OPENSSL_NO_UPLINK"
    },
    "wbsm4-baiwu" => {
        "macro" => "OPENSSL_NO_WBSM4_BAIWU"
    },
    "wbsm4-wsise" => {
        "macro" => "OPENSSL_NO_WBSM4_WSISE"
    },
    "wbsm4-xiaolai" => {
        "macro" => "OPENSSL_NO_WBSM4_XIAOLAI"
    },
    "whirlpool" => {
        "macro" => "OPENSSL_NO_WHIRLPOOL"
    },
    "winstore" => {
        "macro" => "OPENSSL_NO_WINSTORE"
    },
    "zkp-gadget" => {
        "macro" => "OPENSSL_NO_ZKP_GADGET"
    },
    "zkp-transcript" => {
        "macro" => "OPENSSL_NO_ZKP_TRANSCRIPT"
    },
    "zlib" => {
        "macro" => "OPENSSL_NO_ZLIB"
    },
    "zlib-dynamic" => {
        "macro" => "OPENSSL_NO_ZLIB_DYNAMIC"
    },
    "zstd" => {
        "macro" => "OPENSSL_NO_ZSTD"
    },
    "zstd-dynamic" => {
        "macro" => "OPENSSL_NO_ZSTD_DYNAMIC"
    }
);
my @user_crossable = qw( AR AS CC CXX CPP LD MT RANLIB RC );

# If run directly, we can give some answers, and even reconfigure
unless (caller) {
    use Getopt::Long;
    use File::Spec::Functions;
    use File::Basename;
    use File::Compare qw(compare_text);
    use File::Copy;
    use Pod::Usage;

    use lib '/code/platform_all_use/core/libs/src/tongsuo/util/perl';
    use OpenSSL::fallback '/code/platform_all_use/core/libs/src/tongsuo/external/perl/MODULES.txt';

    my $here = dirname($0);

    if (scalar @ARGV == 0) {
        # With no arguments, re-create the build file
        # We do that in two steps, where the first step emits perl
        # snippets.

        my $buildfile = $config{build_file};
        my $buildfile_template = "$buildfile.in";
        my @autowarntext = (
            'WARNING: do not edit!',
            "Generated by configdata.pm from "
            .join(", ", @{$config{build_file_templates}}),
            "via $buildfile_template"
        );
        my %gendata = (
            config => \%config,
            target => \%target,
            disabled => \%disabled,
            withargs => \%withargs,
            unified_info => \%unified_info,
            autowarntext => \@autowarntext,
            );

        use lib '.';
        use lib '/code/platform_all_use/core/libs/src/tongsuo/Configurations';
        use gentemplate;

        open my $buildfile_template_fh, ">$buildfile_template"
            or die "Trying to create $buildfile_template: $!";
        foreach (@{$config{build_file_templates}}) {
            copy($_, $buildfile_template_fh)
                or die "Trying to copy $_ into $buildfile_template: $!";
        }
        gentemplate(output => $buildfile_template_fh, %gendata);
        close $buildfile_template_fh;
        print 'Created ',$buildfile_template,"\n";

        use OpenSSL::Template;

        my $prepend = <<'_____';
use File::Spec::Functions;
use lib '/code/platform_all_use/core/libs/src/tongsuo/util/perl';
use lib '/code/platform_all_use/core/libs/src/tongsuo/Configurations';
use lib '.';
use platform;
_____

        my $tmpl;
        open BUILDFILE, ">$buildfile.new"
            or die "Trying to create $buildfile.new: $!";
        $tmpl = OpenSSL::Template->new(TYPE => 'FILE',
                                       SOURCE => $buildfile_template);
        $tmpl->fill_in(FILENAME => $_,
                       OUTPUT => \*BUILDFILE,
                       HASH => \%gendata,
                       PREPEND => $prepend,
                       # To ensure that global variables and functions
                       # defined in one template stick around for the
                       # next, making them combinable
                       PACKAGE => 'OpenSSL::safe')
            or die $OpenSSL::Template::ERROR;
        close BUILDFILE;
        rename("$buildfile.new", $buildfile)
            or die "Trying to rename $buildfile.new to $buildfile: $!";
        print 'Created ',$buildfile,"\n";

        my $configuration_h =
            catfile('include', 'openssl', 'configuration.h');
        my $configuration_h_in =
            catfile($config{sourcedir}, 'include', 'openssl', 'configuration.h.in');
        open CONFIGURATION_H, ">${configuration_h}.new"
            or die "Trying to create ${configuration_h}.new: $!";
        $tmpl = OpenSSL::Template->new(TYPE => 'FILE',
                                       SOURCE => $configuration_h_in);
        $tmpl->fill_in(FILENAME => $_,
                       OUTPUT => \*CONFIGURATION_H,
                       HASH => \%gendata,
                       PREPEND => $prepend,
                       # To ensure that global variables and functions
                       # defined in one template stick around for the
                       # next, making them combinable
                       PACKAGE => 'OpenSSL::safe')
            or die $OpenSSL::Template::ERROR;
        close CONFIGURATION_H;

        # When using stat() on Windows, we can get it to perform better by
        # avoid some data.  This doesn't affect the mtime field, so we're not
        # losing anything...
        ${^WIN32_SLOPPY_STAT} = 1;

        my $update_configuration_h = 0;
        if (-f $configuration_h) {
            my $configuration_h_mtime = (stat($configuration_h))[9];
            my $configuration_h_in_mtime = (stat($configuration_h_in))[9];

            # If configuration.h.in was updated after the last configuration.h,
            # or if configuration.h.new differs configuration.h, we update
            # configuration.h
            if ($configuration_h_mtime < $configuration_h_in_mtime
                || compare_text("${configuration_h}.new", $configuration_h) != 0) {
                $update_configuration_h = 1;
            } else {
                # If nothing has changed, let's just drop the new one and
                # pretend like nothing happened
                unlink "${configuration_h}.new"
            }
        } else {
            $update_configuration_h = 1;
        }

        if ($update_configuration_h) {
            rename("${configuration_h}.new", $configuration_h)
                or die "Trying to rename ${configuration_h}.new to $configuration_h: $!";
            print 'Created ',$configuration_h,"\n";
        }

        exit(0);
    }

    my $dump = undef;
    my $cmdline = undef;
    my $options = undef;
    my $target = undef;
    my $envvars = undef;
    my $makevars = undef;
    my $buildparams = undef;
    my $reconf = undef;
    my $verbose = undef;
    my $query = undef;
    my $help = undef;
    my $man = undef;
    GetOptions('dump|d'                 => \$dump,
               'command-line|c'         => \$cmdline,
               'options|o'              => \$options,
               'target|t'               => \$target,
               'environment|e'          => \$envvars,
               'make-variables|m'       => \$makevars,
               'build-parameters|b'     => \$buildparams,
               'reconfigure|reconf|r'   => \$reconf,
               'verbose|v'              => \$verbose,
               'query|q=s'              => \$query,
               'help'                   => \$help,
               'man'                    => \$man)
        or die "Errors in command line arguments\n";

    # We allow extra arguments with --query.  That allows constructs like
    # this:
    # ./configdata.pm --query 'get_sources(@ARGV)' file1 file2 file3
    if (!$query && scalar @ARGV > 0) {
        print STDERR <<"_____";
Unrecognised arguments.
For more information, do '$0 --help'
_____
        exit(2);
    }

    if ($help) {
        pod2usage(-exitval => 0,
                  -verbose => 1);
    }
    if ($man) {
        pod2usage(-exitval => 0,
                  -verbose => 2);
    }
    if ($dump || $cmdline) {
        print "\nCommand line (with current working directory = $here):\n\n";
        print '    ',join(' ',
                          $config{PERL},
                          catfile($config{sourcedir}, 'Configure'),
                          @{$config{perlargv}}), "\n";
        print "\nPerl information:\n\n";
        print '    ',$config{perl_cmd},"\n";
        print '    ',$config{perl_version},' for ',$config{perl_archname},"\n";
    }
    if ($dump || $options) {
        my $longest = 0;
        my $longest2 = 0;
        foreach my $what (@disablables) {
            $longest = length($what) if $longest < length($what);
            $longest2 = length($disabled{$what})
                if $disabled{$what} && $longest2 < length($disabled{$what});
        }
        print "\nEnabled features:\n\n";
        foreach my $what (@disablables) {
            print "    $what\n" unless $disabled{$what};
        }
        print "\nDisabled features:\n\n";
        foreach my $what (@disablables) {
            if ($disabled{$what}) {
                print "    $what", ' ' x ($longest - length($what) + 1),
                    "[$disabled{$what}]", ' ' x ($longest2 - length($disabled{$what}) + 1);
                print $disabled_info{$what}->{macro}
                    if $disabled_info{$what}->{macro};
                print ' (skip ',
                    join(', ', @{$disabled_info{$what}->{skipped}}),
                    ')'
                    if $disabled_info{$what}->{skipped};
                print "\n";
            }
        }
    }
    if ($dump || $target) {
        print "\nConfig target attributes:\n\n";
        foreach (sort keys %target) {
            next if $_ =~ m|^_| || $_ eq 'template';
            my $quotify = sub {
                map {
                    if (defined $_) {
                        (my $x = $_) =~ s|([\\\$\@"])|\\$1|g; "\"$x\""
                    } else {
                        "undef";
                    }
                } @_;
            };
            print '    ', $_, ' => ';
            if (ref($target{$_}) eq "ARRAY") {
                print '[ ', join(', ', $quotify->(@{$target{$_}})), " ],\n";
            } else {
                print $quotify->($target{$_}), ",\n"
            }
        }
    }
    if ($dump || $envvars) {
        print "\nRecorded environment:\n\n";
        foreach (sort keys %{$config{perlenv}}) {
            print '    ',$_,' = ',($config{perlenv}->{$_} || ''),"\n";
        }
    }
    if ($dump || $makevars) {
        print "\nMakevars:\n\n";
        foreach my $var (@makevars) {
            my $prefix = '';
            $prefix = $config{CROSS_COMPILE}
                if grep { $var eq $_ } @user_crossable;
            $prefix //= '';
            print '    ',$var,' ' x (16 - length $var),'= ',
                (ref $config{$var} eq 'ARRAY'
                 ? join(' ', @{$config{$var}})
                 : $prefix.$config{$var}),
                "\n"
                if defined $config{$var};
        }

        my @buildfile = ($config{builddir}, $config{build_file});
        unshift @buildfile, $here
            unless file_name_is_absolute($config{builddir});
        my $buildfile = canonpath(catdir(@buildfile));
        print <<"_____";

NOTE: These variables only represent the configuration view.  The build file
template may have processed these variables further, please have a look at the
build file for more exact data:
    $buildfile
_____
    }
    if ($dump || $buildparams) {
        my @buildfile = ($config{builddir}, $config{build_file});
        unshift @buildfile, $here
            unless file_name_is_absolute($config{builddir});
        print "\nbuild file:\n\n";
        print "    ", canonpath(catfile(@buildfile)),"\n";

        print "\nbuild file templates:\n\n";
        foreach (@{$config{build_file_templates}}) {
            my @tmpl = ($_);
            unshift @tmpl, $here
                unless file_name_is_absolute($config{sourcedir});
            print '    ',canonpath(catfile(@tmpl)),"\n";
        }
    }
    if ($reconf) {
        if ($verbose) {
            print 'Reconfiguring with: ', join(' ',@{$config{perlargv}}), "\n";
            foreach (sort keys %{$config{perlenv}}) {
                print '    ',$_,' = ',($config{perlenv}->{$_} || ""),"\n";
            }
        }

        chdir $here;
        exec $^X,catfile($config{sourcedir}, 'Configure'),'reconf';
    }
    if ($query) {
        use OpenSSL::Config::Query;

        my $confquery = OpenSSL::Config::Query->new(info => \%unified_info,
                                                    config => \%config);
        my $result = eval "\$confquery->$query";

        # We may need a result class with a printing function at some point.
        # Until then, we assume that we get a scalar, or a list or a hash table
        # with scalar values and simply print them in some orderly fashion.
        if (ref $result eq 'ARRAY') {
            print "$_\n" foreach @$result;
        } elsif (ref $result eq 'HASH') {
            print "$_ : \\\n  ", join(" \\\n  ", @{$result->{$_}}), "\n"
                foreach sort keys %$result;
        } elsif (ref $result eq 'SCALAR') {
            print "$$result\n";
        }
    }
}

1;

__END__

=head1 NAME

configdata.pm - configuration data for OpenSSL builds

=head1 SYNOPSIS

Interactive:

  perl configdata.pm [options]

As data bank module:

  use configdata;

=head1 DESCRIPTION

This module can be used in two modes, interactively and as a module containing
all the data recorded by OpenSSL's Configure script.

When used interactively, simply run it as any perl script.
If run with no arguments, it will rebuild the build file (Makefile or
corresponding).
With at least one option, it will instead get the information you ask for, or
re-run the configuration process.
See L</OPTIONS> below for more information.

When loaded as a module, you get a few databanks with useful information to
perform build related tasks.  The databanks are:

    %config             Configured things.
    %target             The OpenSSL config target with all inheritances
                        resolved.
    %disabled           The features that are disabled.
    @disablables        The list of features that can be disabled.
    %withargs           All data given through --with-THING options.
    %unified_info       All information that was computed from the build.info
                        files.

=head1 OPTIONS

=over 4

=item B<--help>

Print a brief help message and exit.

=item B<--man>

Print the manual page and exit.

=item B<--dump> | B<-d>

Print all relevant configuration data.  This is equivalent to B<--command-line>
B<--options> B<--target> B<--environment> B<--make-variables>
B<--build-parameters>.

=item B<--command-line> | B<-c>

Print the current configuration command line.

=item B<--options> | B<-o>

Print the features, both enabled and disabled, and display defined macro and
skipped directories where applicable.

=item B<--target> | B<-t>

Print the config attributes for this config target.

=item B<--environment> | B<-e>

Print the environment variables and their values at the time of configuration.

=item B<--make-variables> | B<-m>

Print the main make variables generated in the current configuration

=item B<--build-parameters> | B<-b>

Print the build parameters, i.e. build file and build file templates.

=item B<--reconfigure> | B<--reconf> | B<-r>

Re-run the configuration process.

=item B<--verbose> | B<-v>

Verbose output.

=back

=cut

EOF
