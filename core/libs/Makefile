SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR)/..)
TONSUO_SRC := $(CORE_ROOT)/libs/src/tongsuo
TONSUO_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_BUILD_OPENSSL := $(TONSUO_BUILD_PREFIX)/bin/openssl
TONSUO_BUILD_SH := $(TONSUO_SRC)/build.sh

TONSUO_MIN_VERSION ?= 8.5.0
TONSUO_CONFIGURE_OPTS ?=
JOBS ?= $(shell nproc 2>/dev/null || echo 2)

# 默认部署目录
DEPLOY_ROOT ?= /opt/core
TONSUO_DEPLOY_LIB_DIR := $(DEPLOY_ROOT)/libs/tongsuo
TONSUO_DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin
TONSUO_DEPLOY_OPENSSL := $(TONSUO_DEPLOY_BIN_DIR)/openssl

.PHONY: all tongsuo build deploy check-version clean distclean

all: tongsuo deploy

tongsuo: $(TONSUO_BUILD_OPENSSL)

$(TONSUO_BUILD_OPENSSL): $(TONSUO_BUILD_SH)
	@test -d "$(TONSUO_SRC)" || (echo "missing tongsuo source: $(TONSUO_SRC)" >&2; exit 2)
	@test -x "$(TONSUO_BUILD_SH)" || chmod +x "$(TONSUO_BUILD_SH)"
	JOBS="$(JOBS)" \
	TONSUO_BUILD_PREFIX="$(TONSUO_BUILD_PREFIX)" \
	TONSUO_RUNTIME_LIB_DIR="$(TONSUO_DEPLOY_LIB_DIR)" \
	TONSUO_CONFIGURE_OPTS="$(TONSUO_CONFIGURE_OPTS)" \
		"$(TONSUO_BUILD_SH)"
	@test -x "$(TONSUO_BUILD_OPENSSL)" || (echo "openssl not built: $(TONSUO_BUILD_OPENSSL)" >&2; exit 9)
	@$(MAKE) check-version TONSUO_OPENSSL_BIN="$(TONSUO_BUILD_OPENSSL)"

deploy: $(TONSUO_BUILD_OPENSSL)
	@install -d -m 0755 "$(TONSUO_DEPLOY_LIB_DIR)" "$(TONSUO_DEPLOY_BIN_DIR)"
	@cp -a "$(TONSUO_BUILD_PREFIX)/libs/." "$(TONSUO_DEPLOY_LIB_DIR)/"
	@cp -a "$(TONSUO_BUILD_PREFIX)/bin/." "$(TONSUO_DEPLOY_BIN_DIR)/"
	@chmod 0755 "$(TONSUO_DEPLOY_BIN_DIR)/openssl"
	@test -x "$(TONSUO_DEPLOY_OPENSSL)" || (echo "deploy openssl missing: $(TONSUO_DEPLOY_OPENSSL)" >&2; exit 9)
	@$(MAKE) check-version TONSUO_OPENSSL_BIN="$(TONSUO_DEPLOY_OPENSSL)"

build: tongsuo

check-version:
	@TONSUO_OPENSSL_BIN="$(or $(TONSUO_OPENSSL_BIN),$(TONSUO_DEPLOY_OPENSSL))" \
	 TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		"$(CORE_ROOT)/sbin/lib/version_check.sh" --json

clean:
	@if [ -d "$(TONSUO_SRC)" ]; then $(MAKE) -C "$(TONSUO_SRC)" clean || true; fi

distclean: clean
	rm -rf "$(TONSUO_BUILD_PREFIX)"
