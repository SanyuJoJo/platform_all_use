SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR))

DEPLOY_ROOT ?= /opt/core
TONSUO_DEPLOY_LIB_DIR := $(DEPLOY_ROOT)/libs/tongsuo
TONSUO_DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin
TONSUO_DEPLOY_OPENSSL := $(TONSUO_DEPLOY_BIN_DIR)/openssl

TONSUO_MIN_VERSION ?= 8.5.0
JOBS ?= $(shell nproc 2>/dev/null || echo 2)

.PHONY: all libs tool sbin check-version init-dirs init-deploy-dirs clean distclean

all: libs tool sbin check-version

libs:
	$(MAKE) -C libs all DEPLOY_ROOT="$(DEPLOY_ROOT)"

tool:
	$(MAKE) -C src/tool all

sbin:
	chmod +x sbin/*.sh sbin/lib/*.sh
	-chmod +x tests/*.sh 2>/dev/null || true

check-version:
	TONSUO_OPENSSL_BIN="$(TONSUO_DEPLOY_OPENSSL)" \
	TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		./sbin/lib/version_check.sh --json

init-dirs:
	mkdir -p libs/bin/tongsuo src/tool/bin conf logs data/keys tmp tests
	chmod 0755 .
	chmod 0750 conf logs
	chmod 0700 data/keys tmp
	@if [ -f conf/core.conf ]; then chmod 0640 conf/core.conf; fi

init-deploy-dirs:
	mkdir -p "$(TONSUO_DEPLOY_LIB_DIR)" "$(TONSUO_DEPLOY_BIN_DIR)"
	chmod 0755 "$(DEPLOY_ROOT)" \
	  "$(DEPLOY_ROOT)/libs" \
	  "$(TONSUO_DEPLOY_LIB_DIR)" \
	  "$(DEPLOY_ROOT)/run" \
	  "$(TONSUO_DEPLOY_BIN_DIR)"

clean:
	$(MAKE) -C libs clean || true
	$(MAKE) -C src/tool clean || true
	rm -rf tmp/*

distclean: clean
	rm -rf libs/bin/tongsuo
