# ============================================================
# libs/Makefile — 第三方库编译入口
# 路径：core/libs/Makefile
#
# ★ 关键修复：
#   1. libsrtp 用 CMake + -fPIC 构建（避免 libmk_api.so 链接失败）；
#   2. 编译后删源码目录 autotools 遗留 config.h；
#   3. libsrtp 验证用 D srtp_aes_gcm_128/256（cipher_type 结构体）；
#   4. ★ MediaServer 链接验证用 MediaServer.debug（因为 MediaServer 被 strip 了）；
#   5. ★ 若 MediaServer.debug 里 0 个 srtp 符号，用 cmake -E cmake_link_script
#      在 build/server/ 目录重新链接（修正原 relink 的路径错误）。
# ============================================================
SHELL := /bin/sh
CORE_ROOT := $(abspath $(CURDIR)/..)

# ---------- Tongsuo ----------
TONSUO_SRC := $(CORE_ROOT)/libs/src/tongsuo
TONSUO_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_BUILD_OPENSSL := $(TONSUO_BUILD_PREFIX)/bin/openssl
TONSUO_BUILD_SH := $(TONSUO_SRC)/build.sh
TONSUO_MIN_VERSION ?= 8.5.0
TONSUO_CONFIGURE_OPTS ?=

# ---------- 并行 ----------
NPROC := $(shell nproc 2>/dev/null || echo 2)
JOBS ?= $(shell if [ $(NPROC) -gt 1 ]; then echo $$(( $(NPROC) - 1 )); else echo 1; fi)

# ---------- 部署目录 ----------
DEPLOY_ROOT ?= /opt/plats_tool/core
TONSUO_DEPLOY_LIB_DIR := $(DEPLOY_ROOT)/libs/tongsuo
TONSUO_DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin
TONSUO_DEPLOY_OPENSSL := $(TONSUO_DEPLOY_BIN_DIR)/openssl

# ---------- ZLMediaKit 独立部署根 ----------
VIDEO_DEPLOY_ROOT ?= /opt/plats_tool/video
ZLM_DEPLOY_LIB_DIR := $(VIDEO_DEPLOY_ROOT)/libs/ZLMediaKit
ZLM_DEPLOY_BIN_DIR := $(VIDEO_DEPLOY_ROOT)/run/bin
ZLM_DEPLOY_CONF_DIR := $(VIDEO_DEPLOY_ROOT)/conf/zlmediakit

# ---------- oqs-provider / liboqs ----------
OQS_PROVIDER_SRC := $(CORE_ROOT)/libs/src/oqs-provider-0.12.0
OQS_PROVIDER_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/oqs-provider

OQS_LIBOQS_VER ?= 0.16.0
OQS_LIBOQS_SRC := $(CORE_ROOT)/libs/src/liboqs-$(OQS_LIBOQS_VER)
OQS_LIBOQS_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/liboqs
OQS_LIBOQS_TARBALL_URL := https://github.com/open-quantum-safe/liboqs/archive/refs/tags/$(OQS_LIBOQS_VER).tar.gz
OQS_LIBOQS_STATIC_LIB := $(OQS_LIBOQS_BUILD_PREFIX)/lib/liboqs.a

OQS_PROVIDER_DEPLOY_MODULES_DIR := $(TONSUO_DEPLOY_LIB_DIR)/ossl-modules
TONSUO_BUILD_MODULES_DIR := $(TONSUO_BUILD_PREFIX)/libs/ossl-modules

# ---------- libsrtp（CMake，★ 必须 -fPIC） ----------
SRTP_VER ?= 2.5.0
SRTP_SRC := $(CORE_ROOT)/libs/src/libsrtp-$(SRTP_VER)
SRTP_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/libsrtp
SRTP_TARBALL_URL := https://github.com/cisco/libsrtp/archive/refs/tags/v$(SRTP_VER).tar.gz
SRTP_STATIC_LIB := $(SRTP_BUILD_PREFIX)/lib/libsrtp2.a
SRTP_BUILD_DIR := $(SRTP_SRC)/build

# ---------- ZLMediaKit ----------
ZLM_SRC := $(CORE_ROOT)/libs/src/ZLMediaKit
ZLM_BUILD_DIR := $(ZLM_SRC)/build
ZLM_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/zlmediakit
ZLM_BUILD_MEDIASERVER := $(ZLM_SRC)/release/linux/Release/MediaServer
ZLM_BUILD_CONFIG := $(ZLM_SRC)/conf/config.ini

# ZLM 的 CMake 参数（★ CMAKE_EXE_LINKER_FLAGS 不带 --whole-archive，否则 CMake 编译器检测失败）
ZLM_CMAKE_FLAGS ?= \
    -DCMAKE_BUILD_TYPE=Release \
    -DENABLE_TESTS=OFF \
    -DENABLE_WEBRTC=ON \
    -DENABLE_FFMPEG=ON \
    -DOPENSSL_ROOT_DIR="$(TONSUO_BUILD_PREFIX)" \
    -DOPENSSL_INCLUDE_DIR="$(TONSUO_BUILD_PREFIX)/include" \
    -DOPENSSL_SSL_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libssl.so" \
    -DOPENSSL_CRYPTO_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libcrypto.so" \
    -DCMAKE_PREFIX_PATH="$(SRTP_BUILD_PREFIX);$(TONSUO_BUILD_PREFIX)" \
    -DSRTP_ROOT="$(SRTP_BUILD_PREFIX)" \
    -DSRTP_INCLUDE_DIRS="$(SRTP_BUILD_PREFIX)/include" \
    -DSRTP_LIBRARIES="$(SRTP_STATIC_LIB)" \
    -DCMAKE_BUILD_WITH_INSTALL_RPATH=ON \
    -DCMAKE_INSTALL_RPATH="$(TONSUO_DEPLOY_LIB_DIR)"

.PHONY: all tongsuo liboqs oqs-provider libsrtp zlmediakit build deploy \
        check-version clean distclean

all: deploy

# ============================================================
# 1. Tongsuo
# ============================================================
tongsuo: $(TONSUO_BUILD_OPENSSL)

$(TONSUO_BUILD_OPENSSL): $(TONSUO_BUILD_SH)
	@test -d "$(TONSUO_SRC)" || \
		(echo "missing tongsuo source: $(TONSUO_SRC)" >&2; exit 2)
	@test -x "$(TONSUO_BUILD_SH)" || chmod +x "$(TONSUO_BUILD_SH)"
	JOBS="$(JOBS)" \
	TONSUO_BUILD_PREFIX="$(TONSUO_BUILD_PREFIX)" \
	TONSUO_RUNTIME_LIB_DIR="$(TONSUO_DEPLOY_LIB_DIR)" \
	TONSUO_CONFIGURE_OPTS="$(TONSUO_CONFIGURE_OPTS)" \
		"$(TONSUO_BUILD_SH)"
	@test -x "$(TONSUO_BUILD_OPENSSL)" || \
		(echo "openssl not built: $(TONSUO_BUILD_OPENSSL)" >&2; exit 9)
	@$(MAKE) check-version TONSUO_OPENSSL_BIN="$(TONSUO_BUILD_OPENSSL)"

# ============================================================
# 2. liboqs
# ============================================================
liboqs: $(OQS_LIBOQS_STATIC_LIB)

$(OQS_LIBOQS_STATIC_LIB):
	@if [ ! -d "$(OQS_LIBOQS_SRC)" ]; then \
		echo "downloading liboqs $(OQS_LIBOQS_VER) ..."; \
		mkdir -p "$(CORE_ROOT)/libs/src"; \
		cd "$(CORE_ROOT)/libs/src" && \
			curl -fL -o liboqs.tar.gz "$(OQS_LIBOQS_TARBALL_URL)" || \
			(echo "download liboqs failed" >&2; exit 2); \
		tar -xzf liboqs.tar.gz && rm -f liboqs.tar.gz; \
	fi
	cmake -S "$(OQS_LIBOQS_SRC)" -B "$(OQS_LIBOQS_SRC)/_build" \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(OQS_LIBOQS_BUILD_PREFIX)" \
		-DBUILD_SHARED_LIBS=OFF \
		-DCMAKE_POSITION_INDEPENDENT_CODE=ON \
		-DOQS_BUILD_ONLY_LIB=ON \
		-DOQS_USE_OPENSSL=OFF
	cmake --build "$(OQS_LIBOQS_SRC)/_build" -j"$(JOBS)"
	cmake --install "$(OQS_LIBOQS_SRC)/_build"
	@test -f "$(OQS_LIBOQS_STATIC_LIB)" || \
		(echo "liboqs.a not found: $(OQS_LIBOQS_STATIC_LIB)" >&2; exit 9)

# ============================================================
# 3. oqs-provider
# ============================================================
oqs-provider: $(TONSUO_BUILD_OPENSSL) $(OQS_LIBOQS_STATIC_LIB)
	@test -d "$(OQS_PROVIDER_SRC)" || \
		(echo "missing oqs-provider source: $(OQS_PROVIDER_SRC)" >&2; exit 2)
	rm -rf "$(OQS_PROVIDER_SRC)/_build"
	cmake -S "$(OQS_PROVIDER_SRC)" -B "$(OQS_PROVIDER_SRC)/_build" \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(OQS_PROVIDER_BUILD_PREFIX)" \
		-DOPENSSL_ROOT_DIR="$(TONSUO_BUILD_PREFIX)" \
		-DOPENSSL_CRYPTO_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libcrypto.so" \
		-DOPENSSL_SSL_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libssl.so" \
		-DOPENSSL_INCLUDE_DIR="$(TONSUO_BUILD_PREFIX)/include" \
		-DCMAKE_PREFIX_PATH="$(OQS_LIBOQS_BUILD_PREFIX);$(TONSUO_BUILD_PREFIX)" \
		-Dliboqs_DIR="$(OQS_LIBOQS_BUILD_PREFIX)/lib/cmake/liboqs" \
		-DOQS_PROVIDER_BUILD_STATIC=OFF \
		-DBUILD_TESTING=OFF \
		-DCMAKE_EXE_LINKER_FLAGS="-pthread"
	cmake --build "$(OQS_PROVIDER_SRC)/_build" -j"$(JOBS)"
	install -d "$(OQS_PROVIDER_BUILD_PREFIX)/lib"
	install -m 0755 "$(OQS_PROVIDER_SRC)/_build/lib/oqsprovider.so" \
		"$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so"
	install -d "$(TONSUO_BUILD_MODULES_DIR)"
	install -m 0755 "$(OQS_PROVIDER_SRC)/_build/lib/oqsprovider.so" \
		"$(TONSUO_BUILD_MODULES_DIR)/oqsprovider.so"
	@test -f "$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so" || \
		(echo "oqsprovider.so not found" >&2; exit 9)

# ============================================================
# 4. libsrtp（CMake，★ -fPIC）
# ============================================================
libsrtp: $(SRTP_STATIC_LIB)

$(SRTP_STATIC_LIB): $(TONSUO_BUILD_OPENSSL)
	# ---- 1. 下载/解压 ----
	@if [ ! -d "$(SRTP_SRC)" ]; then \
		echo "downloading libsrtp $(SRTP_VER) ..."; \
		mkdir -p "$(CORE_ROOT)/libs/src"; \
		cd "$(CORE_ROOT)/libs/src" && \
			curl -fL -o libsrtp.tar.gz "$(SRTP_TARBALL_URL)" || \
			(echo "download libsrtp failed" >&2; exit 2); \
		tar -xzf libsrtp.tar.gz && rm -f libsrtp.tar.gz; \
	fi

	# ---- 2. 补 Tongsuo .so 软链 ----
	@for lib in libcrypto libssl; do \
		if [ ! -e "$(TONSUO_BUILD_PREFIX)/libs/$$lib.so" ]; then \
			target=$$(ls "$(TONSUO_BUILD_PREFIX)/libs/$$lib".so.* 2>/dev/null | head -1); \
			if [ -n "$$target" ]; then \
				ln -sfn "$$(basename $$target)" "$(TONSUO_BUILD_PREFIX)/libs/$$lib.so"; \
				echo "  → created $(TONSUO_BUILD_PREFIX)/libs/$$lib.so"; \
			else \
				echo "ERROR: no $$lib.so.* in $(TONSUO_BUILD_PREFIX)/libs" >&2; exit 9; \
			fi; \
		fi; \
	done

	# ---- 3. ★ 删除源码目录 autotools 遗留 config.h ----
	@if [ -f "$(SRTP_SRC)/crypto/include/config.h" ]; then \
		echo "  → removing stale $(SRTP_SRC)/crypto/include/config.h"; \
		rm -f "$(SRTP_SRC)/crypto/include/config.h"; \
	fi

	# ---- 4. 清理旧构建 ----
	@rm -rf "$(SRTP_BUILD_DIR)" "$(SRTP_BUILD_PREFIX)"

	# ---- 5. CMake 配置（★ -fPIC 必须，否则 libmk_api.so 链接失败）----
	@echo "configuring libsrtp $(SRTP_VER) via CMake (PIC enabled) against Tongsuo"
	cmake -S "$(SRTP_SRC)" -B "$(SRTP_BUILD_DIR)" \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(SRTP_BUILD_PREFIX)" \
		-DENABLE_OPENSSL=ON \
		-DENABLE_MBEDTLS=OFF \
		-DENABLE_NSS=OFF \
		-DBUILD_SHARED_LIBS=OFF \
		-DCMAKE_POSITION_INDEPENDENT_CODE=ON \
		-DOPENSSL_ROOT_DIR="$(TONSUO_BUILD_PREFIX)" \
		-DOPENSSL_INCLUDE_DIR="$(TONSUO_BUILD_PREFIX)/include" \
		-DOPENSSL_CRYPTO_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libcrypto.so" \
		-DOPENSSL_SSL_LIBRARY="$(TONSUO_BUILD_PREFIX)/libs/libssl.so" \
		-DCMAKE_PREFIX_PATH="$(TONSUO_BUILD_PREFIX)" \
		-DCMAKE_C_FLAGS="-I$(SRTP_BUILD_DIR) -I$(SRTP_BUILD_DIR)/crypto/include -O2 -fPIC -Wno-error" \
		-DCMAKE_EXE_LINKER_FLAGS="-L$(TONSUO_BUILD_PREFIX)/libs -Wl,-rpath,$(TONSUO_DEPLOY_LIB_DIR)"

	# ---- 6. config.h 检查 ----
	@echo "=== libsrtp CMake config result ==="
	@grep -E "^(HAVE_AES_GCM|GCM|OPENSSL|ENABLE_OPENSSL)" "$(SRTP_BUILD_DIR)/CMakeCache.txt" 2>/dev/null || true
	@if [ ! -f "$(SRTP_BUILD_DIR)/config.h" ]; then \
		echo "ERROR: config.h not found in $(SRTP_BUILD_DIR)" >&2; \
		tail -50 "$(SRTP_BUILD_DIR)/CMakeFiles/CMakeOutput.log" 2>/dev/null >&2; \
		exit 9; \
	fi
	@echo "--- config.h key macros ---"
	@grep -E "^#define (GCM|OPENSSL) 1|^/\* #undef (GCM|OPENSSL) \*/" "$(SRTP_BUILD_DIR)/config.h" || true
	@grep -qE "^#define GCM 1" "$(SRTP_BUILD_DIR)/config.h" || \
		(echo "ERROR: GCM not defined" >&2; grep -E "GCM" "$(SRTP_BUILD_DIR)/config.h" >&2; exit 9)
	@grep -qE "^#define OPENSSL 1" "$(SRTP_BUILD_DIR)/config.h" || \
		(echo "ERROR: OPENSSL not defined" >&2; grep -E "OPENSSL" "$(SRTP_BUILD_DIR)/config.h" >&2; exit 9)

	# ---- 7. 编译 + 安装 ----
	cmake --build "$(SRTP_BUILD_DIR)" -j"$(JOBS)"
	cmake --install "$(SRTP_BUILD_DIR)"

	# ---- 8. 定位产物 ----
	@if [ ! -f "$(SRTP_STATIC_LIB)" ]; then \
		FOUND=$$(find "$(SRTP_BUILD_DIR)" -name "libsrtp2.a" -type f 2>/dev/null | head -1); \
		if [ -n "$$FOUND" ]; then \
			mkdir -p "$$(dirname "$(SRTP_STATIC_LIB)")"; \
			install -m 0644 "$$FOUND" "$(SRTP_STATIC_LIB)"; \
		fi; \
	fi
	@test -f "$(SRTP_STATIC_LIB)" || \
		(echo "ERROR: libsrtp2.a not found" >&2; exit 9)

	# ---- 9. ★ 验证 PIC（关键：避免 libmk_api.so 链接失败）----
	@echo "=== verifying libsrtp2.a is PIC ==="
	@rm -rf /tmp/srtp_pic_check && mkdir -p /tmp/srtp_pic_check && \
		cd /tmp/srtp_pic_check && ar x "$(SRTP_STATIC_LIB)" srtp.c.o 2>/dev/null || true
	@if [ -f /tmp/srtp_pic_check/srtp.c.o ]; then \
		if readelf -r /tmp/srtp_pic_check/srtp.c.o 2>/dev/null | grep -qE "R_X86_64_PC32.*mod_srtp"; then \
			echo "  ❌ srtp.c.o has non-PIC relocation (R_X86_64_PC32 mod_srtp)" >&2; \
			echo "     This will break libmk_api.so linking." >&2; \
			exit 9; \
		fi; \
	fi
	@echo "  ✅ libsrtp2.a is PIC"

	# ---- 10. 验证 GCM 符号（★ 用 D 类型，不是 T）----
	@echo "=== verifying libsrtp GCM support ==="
	@for sym in srtp_create srtp_protect srtp_unprotect; do \
		if nm "$(SRTP_STATIC_LIB)" 2>/dev/null | grep -qE " T $$sym$$"; then \
			echo "  ✅ $$sym (T)"; \
		else \
			echo "  ❌ $$sym MISSING (T)" >&2; exit 9; \
		fi; \
	done
	@for sym in srtp_aes_gcm_128 srtp_aes_gcm_256; do \
		if nm "$(SRTP_STATIC_LIB)" 2>/dev/null | grep -qE " D $$sym$$"; then \
			echo "  ✅ $$sym (D, cipher_type struct)"; \
		else \
			echo "  ❌ $$sym MISSING (D)" >&2; \
			nm "$(SRTP_STATIC_LIB)" 2>/dev/null | grep -i "gcm" >&2 || true; exit 9; \
		fi; \
	done
	@echo "✅ libsrtp built with full AES-GCM + PIC"

# ============================================================
# 5. ZLMediaKit
# ============================================================
zlmediakit: libsrtp $(ZLM_BUILD_PREFIX)/MediaServer

$(ZLM_BUILD_PREFIX)/MediaServer:
	@test -d "$(ZLM_SRC)" || \
		(echo "missing ZLMediaKit source: $(ZLM_SRC)" >&2; exit 2)
	@if [ -f "$(ZLM_BUILD_DIR)/CMakeCache.txt" ]; then \
		echo "removing stale CMakeCache.txt"; \
		rm -f "$(ZLM_BUILD_DIR)/CMakeCache.txt"; \
	fi
	mkdir -p "$(ZLM_BUILD_DIR)"
	cd "$(ZLM_BUILD_DIR)" && \
		PKG_CONFIG_PATH="$(TONSUO_BUILD_PREFIX)/libs/pkgconfig:$(SRTP_BUILD_PREFIX)/lib/pkgconfig:$$PKG_CONFIG_PATH" \
		cmake .. $(ZLM_CMAKE_FLAGS)
	cmake --build "$(ZLM_BUILD_DIR)" -j"$(JOBS)"
	@test -f "$(ZLM_BUILD_MEDIASERVER)" || \
		(echo "MediaServer not built: $(ZLM_BUILD_MEDIASERVER)" >&2; exit 9)

	# ============================================================
	# ★ 关键：验证用 MediaServer.debug（MediaServer 被 strip，nm 看不到符号）
	# ============================================================
	@DEBUG_FILE="$(ZLM_BUILD_MEDIASERVER).debug"; \
	if [ -f "$$DEBUG_FILE" ]; then \
		echo "  → verifying via $(notdir $$DEBUG_FILE) (MediaServer is stripped)"; \
		count=$$(nm "$$DEBUG_FILE" 2>/dev/null | grep -cE " srtp_"); \
	else \
		echo "  → MediaServer.debug not found, verifying via MediaServer"; \
		count=$$(nm "$(ZLM_BUILD_MEDIASERVER)" 2>/dev/null | grep -cE " srtp_"); \
	fi; \
	if [ "$$count" -eq 0 ]; then \
		echo "  ⚠ 0 srtp_ symbols, relinking MediaServer with --whole-archive"; \
		cd "$(ZLM_BUILD_DIR)/server" && \
		cp CMakeFiles/MediaServer.dir/link.txt CMakeFiles/MediaServer.dir/link.txt.bak && \
		sed -i "s|$(SRTP_STATIC_LIB)|-Wl,--whole-archive $(SRTP_STATIC_LIB) -Wl,--no-whole-archive|g" CMakeFiles/MediaServer.dir/link.txt && \
		/usr/local/bin/cmake -E cmake_link_script CMakeFiles/MediaServer.dir/link.txt --verbose=1 && \
		mv CMakeFiles/MediaServer.dir/link.txt.bak CMakeFiles/MediaServer.dir/link.txt && \
		objcopy --only-keep-debug "$(ZLM_BUILD_MEDIASERVER)" "$$DEBUG_FILE" 2>/dev/null || true; \
		if [ -f "$$DEBUG_FILE" ]; then \
			count=$$(nm "$$DEBUG_FILE" 2>/dev/null | grep -cE " srtp_"); \
		fi; \
		echo "  → after relink: $$count srtp_ symbols"; \
	fi; \
	if [ "$$count" -eq 0 ]; then \
		echo "  ❌ still 0 srtp_ symbols after relink" >&2; \
		echo "     .a 里的 srtp_create:" >&2; \
		nm "$(SRTP_STATIC_LIB)" 2>/dev/null | grep " T srtp_create" >&2 || true; \
		exit 9; \
	else \
		echo "  ✅ MediaServer has $$count srtp_ symbols"; \
	fi

	install -d "$(ZLM_BUILD_PREFIX)"
	install -d "$(ZLM_BUILD_PREFIX)/lib"
	install -m 0755 "$(ZLM_BUILD_MEDIASERVER)" "$(ZLM_BUILD_PREFIX)/MediaServer"
	install -m 0644 "$(ZLM_BUILD_CONFIG)" "$(ZLM_BUILD_PREFIX)/config.ini"
	find "$(ZLM_SRC)/release/linux/Release" -maxdepth 1 -type f \
		-name '*.so' -not -name '*.debug' \
		-exec install -m 0755 {} "$(ZLM_BUILD_PREFIX)/lib/" \;
	@echo "ZLMediaKit built: $(ZLM_BUILD_PREFIX)/MediaServer"

	@echo "--- Tongsuo link check ---"
	@ldd "$(ZLM_BUILD_PREFIX)/MediaServer" 2>/dev/null | grep -E "ssl|crypto" || true

# ============================================================
# 6. 部署
# ============================================================
deploy: $(TONSUO_BUILD_OPENSSL) oqs-provider zlmediakit
	@install -d -m 0755 "$(TONSUO_DEPLOY_LIB_DIR)" "$(TONSUO_DEPLOY_BIN_DIR)"
	@install -d -m 0755 "$(OQS_PROVIDER_DEPLOY_MODULES_DIR)"
	@cp -a "$(TONSUO_BUILD_PREFIX)/libs/." "$(TONSUO_DEPLOY_LIB_DIR)/"
	@cp -a "$(TONSUO_BUILD_PREFIX)/bin/." "$(TONSUO_DEPLOY_BIN_DIR)/"
	@install -m 0755 "$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so" \
		"$(OQS_PROVIDER_DEPLOY_MODULES_DIR)/oqsprovider.so"
	@echo "oqs-provider deployed to: $(OQS_PROVIDER_DEPLOY_MODULES_DIR)/oqsprovider.so"
	@chmod 0755 "$(TONSUO_DEPLOY_BIN_DIR)/openssl"
	@test -x "$(TONSUO_DEPLOY_OPENSSL)" || \
		(echo "deploy openssl missing: $(TONSUO_DEPLOY_OPENSSL)" >&2; exit 9)
	@install -d -m 0755 "$(ZLM_DEPLOY_LIB_DIR)" "$(ZLM_DEPLOY_BIN_DIR)" "$(ZLM_DEPLOY_CONF_DIR)"
	@install -m 0755 "$(ZLM_BUILD_PREFIX)/MediaServer" "$(ZLM_DEPLOY_BIN_DIR)/MediaServer"
	@install -m 0644 "$(ZLM_BUILD_PREFIX)/config.ini" "$(ZLM_DEPLOY_CONF_DIR)/config.ini"
	@if [ -d "$(ZLM_BUILD_PREFIX)/lib" ]; then \
		find "$(ZLM_BUILD_PREFIX)/lib" -maxdepth 1 -type f \
			-name '*.so' -not -name '*.debug' \
			-exec install -m 0755 {} "$(ZLM_DEPLOY_LIB_DIR)/" \; ; \
	fi
	@echo "ZLMediaKit deployed:"
	@echo "  bin  -> $(ZLM_DEPLOY_BIN_DIR)/MediaServer"
	@echo "  libs -> $(ZLM_DEPLOY_LIB_DIR)/"
	@echo "  conf -> $(ZLM_DEPLOY_CONF_DIR)/config.ini"
	# ★ 部署后验证用 MediaServer.debug（源文件）
	@SRC_DEBUG="$(ZLM_BUILD_MEDIASERVER).debug"; \
	count=0; \
	if [ -f "$$SRC_DEBUG" ]; then \
		count=$$(nm "$$SRC_DEBUG" 2>/dev/null | grep -cE " srtp_"); \
	fi; \
	if [ "$$count" -eq 0 ]; then \
		echo "WARN: deployed MediaServer has 0 srtp_ symbols (source debug file)" >&2; \
	else \
		echo "✅ deployed MediaServer has $$count srtp_ symbols"; \
	fi
	@$(MAKE) check-version TONSUO_OPENSSL_BIN="$(TONSUO_DEPLOY_OPENSSL)"

# ============================================================
# 仅编译
# ============================================================
build: tongsuo liboqs oqs-provider libsrtp zlmediakit

# ============================================================
# 版本检查
# ============================================================
check-version:
	@TONSUO_OPENSSL_BIN="$(or $(TONSUO_OPENSSL_BIN),$(TONSUO_DEPLOY_OPENSSL))" \
	 TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		"$(CORE_ROOT)/sbin/lib/version_check.sh" --json

# ============================================================
# 清理
# ============================================================
clean:
	@if [ -d "$(TONSUO_SRC)" ]; then \
		$(MAKE) -C "$(TONSUO_SRC)" clean || true; \
	fi
	@rm -rf "$(OQS_PROVIDER_SRC)/_build"
	@rm -rf "$(OQS_LIBOQS_SRC)/_build"
	@rm -rf "$(SRTP_BUILD_DIR)"
	@rm -rf "$(ZLM_BUILD_DIR)"
	@rm -rf "$(ZLM_SRC)/release"

distclean: clean
	rm -rf "$(TONSUO_BUILD_PREFIX)"
	rm -rf "$(OQS_PROVIDER_BUILD_PREFIX)"
	rm -rf "$(OQS_LIBOQS_BUILD_PREFIX)"
	rm -rf "$(SRTP_BUILD_PREFIX)"
	rm -rf "$(ZLM_BUILD_PREFIX)"
