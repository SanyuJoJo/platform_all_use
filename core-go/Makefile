SHELL := /bin/bash
CORE_ROOT := $(shell pwd)
BIN_DIR   := $(CORE_ROOT)/bin
GO        ?= go
.PHONY: all build build-linux build-darwin build-windows \
        clean fmt vet test
all: build
# 本平台构建
build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/core$(shell go env GOEXE) ./cmd/core
	@echo "✅ $(BIN_DIR)/core$(shell go env GOEXE)"
# 交叉编译
build-linux:
	@mkdir -p $(BIN_DIR)/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		$(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/linux/core ./cmd/core
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		$(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/linux/core-arm64 ./cmd/core
build-darwin:
	@mkdir -p $(BIN_DIR)/darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
		$(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/darwin/core ./cmd/core
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		$(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/darwin/core-arm64 ./cmd/core
build-windows:
	@mkdir -p $(BIN_DIR)/windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		$(GO) build -trimpath -ldflags="-s -w" \
		-o $(BIN_DIR)/windows/core.exe ./cmd/core
clean:
	rm -rf $(BIN_DIR)
fmt:
	$(GO) fmt ./...
vet:
	$(GO) vet ./...
test:
	$(GO) test ./... -v
