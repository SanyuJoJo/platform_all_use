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
#   优先 $(CORE_ROOT)/../core-go/libs/tongsuo
#   回退 $(CORE_ROOT)/libs/bin/tongsuo
#
# ★ 运行时 RPATH（相对路径）：
#   二进制 <root>/core-go/bin/<tool>  →  库 <root>/core-go/libs/tongsuo/libs/
#   → $ORIGIN/../libs/tongsuo/libs
#
# 用法：
#   make            构建全部
#   make whitebox   仅 whitebox_sm4
#   make check      校验 RUNPATH 与动态库可达性
#   make install    部署
#   make list       列出源与输出
#   make clean      清理

SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR)/../..)

# ---------- 铜锁路径 ----------
TONSUO_PREFIX ?= $(abspath $(CORE_ROOT)/../core-go/libs/tongsuo)
ifeq ($(wildcard $(TONSUO_PREFIX)/include),)
TONSUO_PREFIX := $(CORE_ROOT)/libs/bin/tongsuo
endif
TONSUO_INCLUDE := $(TONSUO_PREFIX)/include

# 编译期 .so 路径
TONSUO_LIB := $(TONSUO_PREFIX)/libs
ifeq ($(wildcard $(TONSUO_LIB)/libcrypto.so*),)
TONSUO_LIB := $(TONSUO_PREFIX)
endif

# ---------- 编译器 ----------
CC      ?= gcc
CFLAGS  := -Wall -Wextra -O2 -I$(TONSUO_INCLUDE)

# ★ 相对 RPATH（部署后自解析）
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
INSTALL_ROOT   ?= /opt/plats_tool
DEPLOY_BIN_DIR := $(INSTALL_ROOT)/core-go/bin

# ---------- 目标 ----------
.PHONY: all whitebox keycrypt sm2_decrypt_helper check install list clean

all: $(BINS)
	@$(MAKE) --no-print-directory check

bin/%: %.c
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $@ $< $(LDFLAGS)

whitebox: bin/whitebox_sm4
keycrypt: bin/keycrypt
sm2_decrypt_helper: bin/sm2_decrypt_helper

check: $(BINS)
	@echo "== 校验 RUNPATH 与动态库可达性 =="
	@for b in $(BINS); do \
	  [ -f "$$b" ] || continue; \
	  echo "--- $$b"; \
	  rpath="$$(readelf -d "$$b" 2>/dev/null | grep -iE 'runpath|rpath' | sed -E 's/.*\[(.*)\].*/\1/')"; \
	  echo "    RUNPATH = $$rpath"; \
	  if ! echo "$$rpath" | grep -q '\$$ORIGIN'; then \
	    echo "    ⚠  RUNPATH 不含 \$$ORIGIN"; \
	  fi; \
	  missing="$$(ldd "$$b" 2>/dev/null | grep 'not found' || true)"; \
	  if [ -n "$$missing" ]; then \
	    echo "    ✗ 缺失库："; echo "$$missing" | sed 's/^/        /'; \
	  else \
	    echo "    ✓ 动态库可达"; \
	  fi; \
	done
	@echo ""
	@echo "开发机库路径：$(TONSUO_LIB)"
	@echo "部署后库路径：$(DEPLOY_BIN_DIR)/../libs/tongsuo/libs/"

install: all
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
