# ============================================================
# libs/Makefile — 第三方库编译入口（Tongsuo + liboqs + oqs-provider）
# 路径：core/libs/Makefile
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

# ---------- 部署目录（默认；可由顶层 Makefile 覆盖） ----------
DEPLOY_ROOT ?= /opt/plats_tool/core
TONSUO_DEPLOY_LIB_DIR := $(DEPLOY_ROOT)/libs/tongsuo
TONSUO_DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin
TONSUO_DEPLOY_OPENSSL := $(TONSUO_DEPLOY_BIN_DIR)/openssl

# ---------- oqs-provider / liboqs ----------
OQS_PROVIDER_SRC := $(CORE_ROOT)/libs/src/oqs-provider-0.12.0
OQS_PROVIDER_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/oqs-provider

OQS_LIBOQS_VER ?= 0.16.0
OQS_LIBOQS_SRC := $(CORE_ROOT)/libs/src/liboqs-$(OQS_LIBOQS_VER)
OQS_LIBOQS_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/liboqs
OQS_LIBOQS_TARBALL_URL := https://github.com/open-quantum-safe/liboqs/archive/refs/tags/$(OQS_LIBOQS_VER).tar.gz
OQS_LIBOQS_STATIC_LIB := $(OQS_LIBOQS_BUILD_PREFIX)/lib/liboqs.a

# oqsprovider.so 的部署目录 = Tongsuo 部署后的 ossl-modules 目录
OQS_PROVIDER_DEPLOY_MODULES_DIR := $(TONSUO_DEPLOY_LIB_DIR)/ossl-modules

# oqsprovider.so 的构建路径模块目录 = Tongsuo 编译时硬编码的默认搜索路径
# 该值由 Tongsuo 的 --libdir 决定，可用 `openssl version -m` 确认
TONSUO_BUILD_MODULES_DIR := $(TONSUO_BUILD_PREFIX)/libs/ossl-modules

.PHONY: all tongsuo liboqs oqs-provider build deploy check-version clean distclean

all: deploy

# ---------- 编译 Tongsuo ----------
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

# ---------- 编译 liboqs（静态库） ----------
liboqs: $(OQS_LIBOQS_STATIC_LIB)

$(OQS_LIBOQS_STATIC_LIB):
	@if [ ! -d "$(OQS_LIBOQS_SRC)" ]; then \
		echo "downloading liboqs $(OQS_LIBOQS_VER) ..."; \
		mkdir -p "$(CORE_ROOT)/libs/src"; \
		cd "$(CORE_ROOT)/libs/src" && \
			curl -fL -o liboqs.tar.gz "$(OQS_LIBOQS_TARBALL_URL)" || \
			(echo "download liboqs failed; please place it at $(OQS_LIBOQS_SRC)" >&2; exit 2); \
		tar -xzf liboqs.tar.gz && rm -f liboqs.tar.gz; \
	fi
	cmake -S "$(OQS_LIBOQS_SRC)" -B "$(OQS_LIBOQS_SRC)/_build" \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(OQS_LIBOQS_BUILD_PREFIX)" \
		-DBUILD_SHARED_LIBS=OFF \
		-DOQS_BUILD_ONLY_LIB=ON \
		-DOQS_USE_OPENSSL=OFF
	cmake --build "$(OQS_LIBOQS_SRC)/_build" -j"$(JOBS)"
	cmake --install "$(OQS_LIBOQS_SRC)/_build"
	@test -f "$(OQS_LIBOQS_STATIC_LIB)" || \
		(echo "liboqs.a not found: $(OQS_LIBOQS_STATIC_LIB)" >&2; exit 9)

# ---------- 编译 oqs-provider ----------
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
	# 复制到 oqs-provider 自己的构建前缀（供 deploy 使用）
	install -d "$(OQS_PROVIDER_BUILD_PREFIX)/lib"
	install -m 0755 "$(OQS_PROVIDER_SRC)/_build/lib/oqsprovider.so" \
		"$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so"
	# ★ 关键：同时复制到 Tongsuo 构建路径的 ossl-modules 目录
	#   这样编译完即可直接测试，无需 deploy
	install -d "$(TONSUO_BUILD_MODULES_DIR)"
	install -m 0755 "$(OQS_PROVIDER_SRC)/_build/lib/oqsprovider.so" \
		"$(TONSUO_BUILD_MODULES_DIR)/oqsprovider.so"
	@test -f "$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so" || \
		(echo "oqsprovider.so not found at $(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so" >&2; exit 9)
	@test -f "$(TONSUO_BUILD_MODULES_DIR)/oqsprovider.so" || \
		(echo "oqsprovider.so not found at $(TONSUO_BUILD_MODULES_DIR)/oqsprovider.so" >&2; exit 9)
	@echo "oqs-provider built: $(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so"
	@echo "oqs-provider installed to build path: $(TONSUO_BUILD_MODULES_DIR)/oqsprovider.so"

# ---------- 部署 ----------
deploy: $(TONSUO_BUILD_OPENSSL) oqs-provider
	@install -d -m 0755 "$(TONSUO_DEPLOY_LIB_DIR)" "$(TONSUO_DEPLOY_BIN_DIR)"
	@install -d -m 0755 "$(OQS_PROVIDER_DEPLOY_MODULES_DIR)"
	@cp -a "$(TONSUO_BUILD_PREFIX)/libs/." "$(TONSUO_DEPLOY_LIB_DIR)/"
	@cp -a "$(TONSUO_BUILD_PREFIX)/bin/." "$(TONSUO_DEPLOY_BIN_DIR)/"
	# 部署 oqsprovider.so 到部署路径的 ossl-modules 目录
	@install -m 0755 "$(OQS_PROVIDER_BUILD_PREFIX)/lib/oqsprovider.so" \
		"$(OQS_PROVIDER_DEPLOY_MODULES_DIR)/oqsprovider.so"
	@echo "oqs-provider deployed to: $(OQS_PROVIDER_DEPLOY_MODULES_DIR)/oqsprovider.so"
	@chmod 0755 "$(TONSUO_DEPLOY_BIN_DIR)/openssl"
	@test -x "$(TONSUO_DEPLOY_OPENSSL)" || \
		(echo "deploy openssl missing: $(TONSUO_DEPLOY_OPENSSL)" >&2; exit 9)
	@$(MAKE) check-version TONSUO_OPENSSL_BIN="$(TONSUO_DEPLOY_OPENSSL)"

# ---------- 仅编译（不部署） ----------
build: tongsuo liboqs oqs-provider

# ---------- 版本检查 ----------
check-version:
	@TONSUO_OPENSSL_BIN="$(or $(TONSUO_OPENSSL_BIN),$(TONSUO_DEPLOY_OPENSSL))" \
	 TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		"$(CORE_ROOT)/sbin/lib/version_check.sh" --json

# ---------- 清理 ----------
clean:
	@if [ -d "$(TONSUO_SRC)" ]; then \
		$(MAKE) -C "$(TONSUO_SRC)" clean || true; \
	fi
	@rm -rf "$(OQS_PROVIDER_SRC)/_build"
	@rm -rf "$(OQS_LIBOQS_SRC)/_build"

distclean: clean
	rm -rf "$(TONSUO_BUILD_PREFIX)"
	rm -rf "$(OQS_PROVIDER_BUILD_PREFIX)"
	rm -rf "$(OQS_LIBOQS_BUILD_PREFIX)"
