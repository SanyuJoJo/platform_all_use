# ============================================================
# core/Makefile — 项目总编译入口
# 路径：core/Makefile
# ============================================================
SHELL := /bin/sh
CORE_ROOT := $(abspath $(CURDIR))

# ---------- 部署路径 ----------
DEPLOY_ROOT ?= /opt/plats_tool/core
# ← 改：加一层 libs，匹配 tool 的 RPATH $ORIGIN/../libs/tongsuo/libs
TONSUO_DEPLOY_LIB_DIR := $(DEPLOY_ROOT)/libs/tongsuo/libs
TONSUO_DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin
TONSUO_DEPLOY_OPENSSL := $(TONSUO_DEPLOY_BIN_DIR)/openssl

# ---------- Tengine 编译路径 ----------
TENGINE_SRC := $(CORE_ROOT)/src/tengine-3.1.0
TENGINE_BUILD_PREFIX := $(CORE_ROOT)/src/tengine-3.1.0/bin
TENGINE_BUILD_NGINX := $(TENGINE_BUILD_PREFIX)/sbin/nginx
TENGINE_BUILD_SH := $(TENGINE_SRC)/build.sh

# ---------- Tongsuo 编译路径 ----------
TONSUO_SRC := $(CORE_ROOT)/libs/src/tongsuo
TONSUO_BUILD_PREFIX := $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_BUILD_OPENSSL := $(TONSUO_BUILD_PREFIX)/bin/openssl

# ---------- 版本与并行 ----------
TONSUO_MIN_VERSION ?= 8.5.0
TENGINE_CONFIGURE_OPTS ?=
NPROC := $(shell nproc 2>/dev/null || echo 2)
# 默认 JOBS = max(1, NPROC-1)
JOBS ?= $(shell if [ $(NPROC) -gt 1 ]; then echo $$(( $(NPROC) - 1 )); else echo 1; fi)

.PHONY: all libs tengine tool sbin check-version \
        init-dirs init-deploy-dirs clean distclean

# ============================================================
# 默认目标：编译全部
# ============================================================
all: libs tengine tool sbin check-version

# ---------- 编译 Tongsuo ----------
libs:
	$(MAKE) -C libs all \
		DEPLOY_ROOT="$(DEPLOY_ROOT)" \
		TONSUO_DEPLOY_LIB_DIR="$(TONSUO_DEPLOY_LIB_DIR)" \
		TONSUO_DEPLOY_BIN_DIR="$(TONSUO_DEPLOY_BIN_DIR)" \
		TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		TONSUO_CONFIGURE_OPTS="$(TONSUO_CONFIGURE_OPTS)" \
		JOBS="$(JOBS)"

# ---------- 编译 Tengine ----------
tengine: $(TENGINE_BUILD_NGINX)

$(TENGINE_BUILD_NGINX): $(TENGINE_BUILD_SH) $(TONSUO_BUILD_OPENSSL)
	@test -d "$(TENGINE_SRC)" || \
		(echo "missing tengine source: $(TENGINE_SRC)" >&2; exit 2)
	@test -x "$(TENGINE_BUILD_SH)" || chmod +x "$(TENGINE_BUILD_SH)"
	JOBS="$(JOBS)" \
	TENGINE_BUILD_PREFIX="$(TENGINE_BUILD_PREFIX)" \
	TONSUO_BUILD_PREFIX="$(TONSUO_SRC)" \
	TONSUO_BUILD_INC_DIR="$(TONSUO_BUILD_PREFIX)/include" \
	TONSUO_BUILD_LIB_DIR="$(TONSUO_BUILD_PREFIX)/libs" \
	TONSUO_RUNTIME_LIB_DIR="$(TONSUO_DEPLOY_LIB_DIR)" \
	TENGINE_CONFIGURE_OPTS="$(TENGINE_CONFIGURE_OPTS)" \
		"$(TENGINE_BUILD_SH)"
	@test -x "$(TENGINE_BUILD_NGINX)" || \
		(echo "tengine not built: $(TENGINE_BUILD_NGINX)" >&2; exit 9)

# ---------- 编译 C 工具 ----------
# ← 改：把部署二进制目录传下去，覆盖 tool/Makefile 里 core-go 的错值
tool:
	$(MAKE) -C src/tool all \
		DEPLOY_BIN_DIR="$(DEPLOY_ROOT)/bin"

# ---------- 脚本权限 ----------
sbin:
	chmod +x sbin/*.sh sbin/lib/*.sh
	-chmod +x tests/*.sh 2>/dev/null || true

# ---------- 版本检查 ----------
check-version:
	TONSUO_OPENSSL_BIN="$(TONSUO_BUILD_OPENSSL)" \
	TONSUO_MIN_VERSION="$(TONSUO_MIN_VERSION)" \
		./sbin/lib/version_check.sh --json

# ---------- 初始化目录 ----------
init-dirs:
	mkdir -p libs/bin/tongsuo src/tool/bin \
		 src/tengine-3.1.0/bin/sbin \
		 conf logs data/keys tmp tests
	chmod 0755 .
	chmod 0750 conf logs
	chmod 0700 data/keys tmp
	@if [ -f conf/core.conf ]; then chmod 0640 conf/core.conf; fi

init-deploy-dirs:
	mkdir -p "$(TONSUO_DEPLOY_LIB_DIR)" "$(TONSUO_DEPLOY_BIN_DIR)" "$(DEPLOY_ROOT)/bin"
	chmod 0755 "$(DEPLOY_ROOT)" \
	  "$(DEPLOY_ROOT)/libs" \
	  "$(DEPLOY_ROOT)/libs/tongsuo" \
	  "$(TONSUO_DEPLOY_LIB_DIR)" \
	  "$(DEPLOY_ROOT)/bin" \
	  "$(DEPLOY_ROOT)/run" \
	  "$(TONSUO_DEPLOY_BIN_DIR)"

# ---------- 清理 ----------
clean:
	$(MAKE) -C libs clean || true
	$(MAKE) -C src/tool clean || true
	@if [ -d "$(TENGINE_SRC)" ]; then \
		$(MAKE) -C "$(TENGINE_SRC)" clean || true; \
	fi
	rm -rf tmp/*

distclean: clean
	rm -rf libs/bin/tongsuo
	rm -rf src/tengine-3.1.0/bin
