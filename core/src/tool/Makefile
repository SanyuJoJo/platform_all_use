# core/src/tool/Makefile
#
# C 工具编译框架。自动发现 src/*.c，编译为 bin/<name>。
#
# 部署路径：/opt/core/run/bin/  （与其他二进制统一位置）
#
# 已编译的工具：
#   - keycrypt            AES-256-GCM 私钥加解密
#   - sm2_decrypt_helper  SM2 私钥解密辅助（绕开 pkeyutl 缺陷）

SHELL := /bin/sh

CORE_ROOT := $(abspath $(CURDIR)/../..)
TONSUO_PREFIX  := $(CORE_ROOT)/libs/bin/tongsuo
TONSUO_INCLUDE := $(TONSUO_PREFIX)/include
TONSUO_LIB     := $(TONSUO_PREFIX)/libs

CC ?= gcc
CFLAGS  := -Wall -Wextra -O2 -I$(TONSUO_INCLUDE)
LDFLAGS := -L$(TONSUO_LIB) -lcrypto -lssl -Wl,-rpath,$(TONSUO_LIB)

# 自动发现所有 src/*.c
SRCS := $(wildcard src/*.c)
BINS := $(patsubst src/%.c,bin/%,$(SRCS))

# 部署目录：统一到 /opt/core/run/bin/
DEPLOY_ROOT ?= /opt/core
DEPLOY_BIN_DIR := $(DEPLOY_ROOT)/run/bin

.PHONY: all clean install list

all: $(BINS)

# 通用规则：每个 .c 编译为同名 bin/
bin/%: src/%.c
	@mkdir -p bin
	$(CC) $(CFLAGS) -o $@ $< $(LDFLAGS)

# 部署到 $(DEPLOY_ROOT)/run/bin/
install: all
	@install -d -m 0755 "$(DEPLOY_BIN_DIR)"
	@for b in $(BINS); do \
	  install -m 0755 "$$b" "$(DEPLOY_BIN_DIR)/"; \
	  echo "installed: $(DEPLOY_BIN_DIR)/$$(basename $$b)"; \
	done

# 列出当前将被编译的 C 源
list:
	@echo "SRCS:"; for s in $(SRCS); do echo "  $$s"; done
	@echo "BINS:"; for b in $(BINS); do echo "  $$b"; done

clean:
	rm -rf bin
