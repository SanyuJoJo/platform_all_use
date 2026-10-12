# backend/Makefile
#
# 目录职责：
#   bin/   —— 开发调试 + 运行环境（可执行文件、插件、日志、PID、备份、配置）
#   build/ —— 交叉编译产物，仅用于发布打包（make cross）
#
# 常用命令：
#   make             # 编译当前平台到 bin/ + 准备运行目录
#   make dev         # 同 make
#   make run         # 编译 + 前台启动 host（Ctrl+C 退出）
#   make run-bg      # 编译 + 后台启动 host
#   make stop        # 停止 host 及插件
#   make status      # 查看运行状态
#   make logs        # 实时查看所有日志
#   make cross       # 交叉编译全平台到 build/
#   make clean       # 清理 bin/ 运行时数据与 build/

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BIN_DIR   := bin
BUILD_DIR := build
GO_FLAGS  := -ldflags "-s -w -X main.version=$(VERSION)"
ARGS      ?=

# 当前主机平台
GOOS_LOCAL   := $(shell go env GOOS)
GOARCH_LOCAL := $(shell go env GOARCH)
ifeq ($(GOOS_LOCAL),windows)
    LOCAL_EXT := .exe
else
    LOCAL_EXT :=
endif

# 交叉编译默认平台列表（仅 make cross 使用）
PLATFORMS ?= linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

.PHONY: all dev host plugins plugins-go plugins-python prepare \
        run run-bg stop status logs clean clean-all cross \
        host-cross plugins-go-cross

# —— 默认目标 ——
all: host plugins prepare

dev: all
	@echo ">> bin/ 已就绪，可执行 make run 启动（前台）"

# —— 主应用：编译到 bin/ ——
host:
	@mkdir -p $(BIN_DIR)
	@echo ">> build $(BIN_DIR)/host$(LOCAL_EXT)"
	@cd host && CGO_ENABLED=0 \
		go build $(GO_FLAGS) -o ../$(BIN_DIR)/host$(LOCAL_EXT) ./cmd/host

# —— 插件 ——
plugins: plugins-go plugins-python

plugins-go:
	@mkdir -p $(BIN_DIR)/plugins/echo
	@echo ">> build $(BIN_DIR)/plugins/echo/echo$(LOCAL_EXT)"
	@cd plugin-echo && CGO_ENABLED=0 \
		go build -o ../$(BIN_DIR)/plugins/echo/echo$(LOCAL_EXT) .
	@cp plugin-echo/manifest.json $(BIN_DIR)/plugins/echo/manifest.json

plugins-python:
	@mkdir -p $(BIN_DIR)/plugins/python-hello
	@echo ">> copy plugin-python-hello"
	@cp plugin-python-hello/main.py          $(BIN_DIR)/plugins/python-hello/
	@cp plugin-python-hello/requirements.txt $(BIN_DIR)/plugins/python-hello/
	@cp plugin-python-hello/manifest.json    $(BIN_DIR)/plugins/python-hello/

# —— 准备运行目录 ——
prepare:
	@mkdir -p $(BIN_DIR)/data/backups
	@mkdir -p $(BIN_DIR)/data/logs
	@mkdir -p $(BIN_DIR)/data/pids
	@mkdir -p $(BIN_DIR)/scripts
	@cp scripts/* $(BIN_DIR)/scripts/ 2>/dev/null || true
	@chmod +x $(BIN_DIR)/scripts/*.sh 2>/dev/null || true
	@if [ -f config.json ]; then \
		cp config.json $(BIN_DIR)/config.json; \
		echo ">> copy config.json -> $(BIN_DIR)/config.json"; \
	fi

# —— 前台运行（默认，先 stop 再 all） ——
ifeq ($(GOOS_LOCAL),windows)
run: stop all
	@cmd /C "$(BIN_DIR)\\scripts\\start.bat" fg $(ARGS)

run-bg: stop all
	@cmd /C "$(BIN_DIR)\\scripts\\start.bat" start $(ARGS)
else
run: stop all
	@bash $(BIN_DIR)/scripts/start.sh fg $(ARGS)

run-bg: stop all
	@bash $(BIN_DIR)/scripts/start.sh start $(ARGS)
endif

# —— 停止（按进程名精确匹配，避免误杀 make/sh） ——
ifeq ($(GOOS_LOCAL),windows)
stop:
	@if exist "$(BIN_DIR)\\scripts\\stop.bat" cmd /C "$(BIN_DIR)\\scripts\\stop.bat"
status:
	@if exist "$(BIN_DIR)\\scripts\\status.bat" cmd /C "$(BIN_DIR)\\scripts\\status.bat"
else
stop:
	@if [ -f "$(BIN_DIR)/scripts/stop.sh" ]; then \
		bash $(BIN_DIR)/scripts/stop.sh || true; \
	fi
	@-pkill -x host 2>/dev/null || true
	@-pkill -x echo 2>/dev/null || true
	@-pkill -f "plugin-python-hello/main.py" 2>/dev/null || true
	@echo "[stop] 完成"
status:
	@if [ -f "$(BIN_DIR)/scripts/status.sh" ]; then \
		bash $(BIN_DIR)/scripts/status.sh; \
	else \
		pgrep -af host 2>/dev/null || echo "无进程在运行"; \
	fi
endif

# —— 实时查看日志 ——
logs:
	@tail -f $(BIN_DIR)/data/logs/*.log

# —— 交叉编译到 build/ ——
cross: host-cross plugins-go-cross prepare
	@echo ">> 交叉编译完成，产物在 $(BUILD_DIR)/"

host-cross:
	@mkdir -p $(BUILD_DIR)
	@for p in $(PLATFORMS); do \
		os=$$(echo $$p | cut -d/ -f1); \
		arch=$$(echo $$p | cut -d/ -f2); \
		out=$(BUILD_DIR)/host-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then out=$$out.exe; fi; \
		echo ">> build $$out"; \
		( cd host && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build $(GO_FLAGS) -o ../$$out ./cmd/host ) || exit 1; \
	done

plugins-go-cross:
	@for p in $(PLATFORMS); do \
		os=$$(echo $$p | cut -d/ -f1); \
		arch=$$(echo $$p | cut -d/ -f2); \
		out=$(BUILD_DIR)/plugins/echo/echo-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then out=$$out.exe; fi; \
		mkdir -p $$(dirname $$out); \
		echo ">> build $$out"; \
		( cd plugin-echo && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -o ../$$out . ) || exit 1; \
	done

# —— 清理 ——
clean:
	@if [ -f "$(BIN_DIR)/scripts/stop.sh" ]; then \
		bash $(BIN_DIR)/scripts/stop.sh 2>/dev/null || true; \
	fi
	@-pkill -x host 2>/dev/null || true
	@-pkill -x echo 2>/dev/null || true
	rm -rf $(BIN_DIR)/host$(LOCAL_EXT)
	rm -rf $(BIN_DIR)/plugins
	rm -rf $(BIN_DIR)/data/logs/*
	rm -rf $(BIN_DIR)/data/pids/*
	rm -rf $(BUILD_DIR)

clean-all:
	@if [ -f "$(BIN_DIR)/scripts/stop.sh" ]; then \
		bash $(BIN_DIR)/scripts/stop.sh 2>/dev/null || true; \
	fi
	@-pkill -x host 2>/dev/null || true
	@-pkill -x echo 2>/dev/null || true
	rm -rf $(BIN_DIR)
	rm -rf $(BUILD_DIR)
