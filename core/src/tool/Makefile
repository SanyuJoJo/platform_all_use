# core/src/tool/Makefile
#
# C 工具编译框架。
#
# 源文件发现（两处，自动去重）：
#   - $(CURDIR)/*.c       白盒工具（whitebox_sm4 等）
#   - $(CURDIR)/src/*.c   keycrypt、sm2_decrypt_helper 等
#
# 产物：bin/<name>
# 部署：$(DEPLOY_BIN_DIR)/
#
# 铜锁（Tongsuo）：
#   TONSUO_PREFIX 默认取本仓 $(CORE_ROOT)/libs/bin/tongsuo（编译产物），
#   可由顶层 Makefile 通过 TONSUO_PREFIX=... 覆盖。
#
# ★ 运行时 RPATH（相对路径）：
#   二进制 <root>/bin/<tool>            →  库 <root>/libs/tongsuo/libs/
#   → $ORIGIN/../libs/tongsuo/libs
#
# 用法：
#   make                构建全部（含编译期 RPATH 检查）
#   make whitebox       仅 whitebox_sm4
#   make keycrypt       仅 keycrypt
#   make sm2_decrypt_helper  仅 sm2_decrypt_helper
#   make check          编译期：校验 RUNPATH 字符串
#   make check-deploy   运行期：校验部署后二进制的动态库可达性
#   make install        部署二进制到 DEPLOY_BIN_DIR
#   make list           列出源与输出
#   make clean          清理

SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR)/../..)

# ---------- 铜锁路径 ----------
# 默认取本仓已编译好的产物，可由外部覆盖
TONSUO_PREFIX ?= $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_INCLUDE := $(TONSUO_PREFIX)/include

# 编译期 .so 路径（优先 libs/ 子目录，回退到前缀根）
TONSUO_LIB := $(TONSUO_PREFIX)/libs
ifeq ($(wildcard $(TONSUO_LIB)/libcrypto.so*),)
  TONSUO_LIB := $(TONSUO_PREFIX)
endif

# ---------- 编译器 ----------
CC      ?= gcc
CFLAGS  := -Wall -Wextra -O2 -I$(TONSUO_INCLUDE)

# ★ 相对 RPATH（部署后自解析）
#   对应部署布局：<root>/bin/<tool>  →  <root>/libs/tongsuo/libs/
TONGSUO_RPATH := $$ORIGIN/../libs/tongsuo/libs

LDFLAGS := -L$(TONSUO_LIB) -lcrypto -lssl \
           -Wl,-rpath,'$(TONGSUO_RPATH)' \
           -Wl,--enable-new-dtags

# ---------- 源发现 ----------
SRCS := $(wildcard $(CURDIR)/*.c) \
        $(wildcard $(CURDIR)/src/*.c)
NAMES := $(sort $(notdir $(patsubst %.c,%,$(SRCS))))
BINS  := $(addprefix bin/,$(NAMES))
VPATH := $(CURDIR):$(CURDIR)/src

# ---------- 部署 ----------
INSTALL_ROOT   ?= /opt/plats_tool/core
DEPLOY_BIN_DIR ?= $(INSTALL_ROOT)/bin

# ---------- 目标 ----------
.PHONY: all whitebox keycrypt sm2_decrypt_helper \
        check check-deploy install list clean

all: $(BINS)
	@$(MAKE) --no-print-directory check

bin/%: %.c
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $@ $< $(LDFLAGS)

whitebox: bin/whitebox_sm4
keycrypt: bin/keycrypt
sm2_decrypt_helper: bin/sm2_decrypt_helper

# ---------- 编译期检查：只看 RPATH 字符串 ----------
check: $(BINS)
	@echo "== 校验 RUNPATH（编译产物）=="
	@for b in $(BINS); do \
	  [ -f "$$b" ] || continue; \
	  echo "--- $$b"; \
	  readelf -d "$$b" 2>/dev/null | grep -iE 'runpath|rpath' || true; \
	done
	@echo ""
	@echo "开发机库路径：$(TONSUO_LIB)"
	@echo "部署后库路径：$(DEPLOY_BIN_DIR)/../libs/tongsuo/libs/"
	@echo ""
	@echo "提示：运行期可达性请对部署后的二进制执行："
	@echo "  make install && make check-deploy"

# ---------- 运行期检查：对部署后的二进制做 ldd ----------
check-deploy:
	@echo "== 校验部署后动态库可达性 =="
	@for b in $(BINS); do \
	  t="$(DEPLOY_BIN_DIR)/$$(basename $$b)"; \
	  if [ ! -f "$$t" ]; then \
	    echo "--- $$t (未部署，跳过)"; continue; \
	  fi; \
	  echo "--- $$t"; \
	  if ldd "$$t" 2>/dev/null | grep -q 'not found'; then \
	    ldd "$$t" 2>/dev/null | grep 'not found' | sed 's/^/    ✗ /'; \
	  else \
	    echo "    ✓ 动态库可达"; \
	  fi; \
	done

# ---------- 部署 ----------
install: $(BINS)
	@install -d -m 0755 "$(DEPLOY_BIN_DIR)"
	@for b in $(BINS); do \
	  install -m 0755 "$$b" "$(DEPLOY_BIN_DIR)/"; \
	  echo "installed: $(DEPLOY_BIN_DIR)/$$(basename $$b)"; \
	done

list:
	@echo "CORE_ROOT      = $(CORE_ROOT)"
	@echo "TONSUO_PREFIX  = $(TONSUO_PREFIX)"
	@echo "TONSUO_LIB     = $(TONSUO_LIB)"
	@echo "TONGSUO_RPATH  = $(TONGSUO_RPATH)"
	@echo "DEPLOY_BIN_DIR = $(DEPLOY_BIN_DIR)"
	@echo "SRCS:"; @for s in $(SRCS); do echo "  $$s"; done
	@echo "BINS:"; @for b in $(BINS); do echo "  $$b"; done

clean:
	rm -rf bin
