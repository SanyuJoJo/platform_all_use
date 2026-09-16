# ============================================================================
# 后端研发环境 Makefile
#
# 快速上手：
#   make bootstrap       # 一键初始化（生成 .env、装依赖、装 goose、建目录）
#   make run             # 开发模式启动（go run）
#   make smoke           # 校验 /health
#   make doctor          # 环境自检
#   make build           # 生成静态二进制 bin/server
#   make start-prod      # 生产模式启动（bin/server）
#   make reset           # 重置本地运行数据
#
# V11-P1-02 修复：所有 goose 调用使用绝对路径 $(GOOSE_BIN)，
#                 避免依赖用户 PATH 中是否包含 $GOPATH/bin。
# V11-P2-02 修复：删除与 bootstrap 重叠的 init 目标。
# ============================================================================
GO              ?= go
GOOSE_VERSION   ?= v3.22.1
GOOSE_BIN       ?= $(shell $(GO) env GOPATH)/bin/goose
SQLITE_PATH     ?= ./app.db
BIN_DIR         ?= bin
.PHONY: bootstrap doctor smoke tidy run build test lint \
        start-prod migrate-up migrate-create clean reset help
# ---------------------------------------------------------------------------
# 一键初始化
# ---------------------------------------------------------------------------
bootstrap:
	@echo "== 创建目录 =="
	@mkdir -p data logs $(BIN_DIR)
	@echo "== 生成 .env =="
	@if [ ! -f .env ]; then cp .env.example .env; echo "已生成 .env"; else echo ".env 已存在，跳过"; fi
	@echo "== 拉取依赖 =="
	$(GO) mod tidy
	@echo "== 检查 goose =="
	@if [ ! -x "$(GOOSE_BIN)" ]; then \
		echo "安装 goose $(GOOSE_VERSION) 到 $(GOOSE_BIN)..."; \
		$(GO) install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION); \
	else \
		echo "goose 已安装：$(GOOSE_BIN)"; \
	fi
	@echo ""
	@echo "✅ bootstrap 完成"
	@echo "下一步：make run"
# ---------------------------------------------------------------------------
# 环境自检
# ---------------------------------------------------------------------------
doctor:
	@echo "== Go 版本 =="
	@$(GO) version 2>/dev/null || echo "❌ 未安装 Go"
	@echo ""
	@echo "== .env 文件 =="
	@if [ -f .env ]; then echo "✅ .env 存在"; else echo "❌ 缺少 .env（执行 make bootstrap）"; fi
	@echo ""
	@echo "== 关键目录 =="
	@for d in data logs $(BIN_DIR); do \
		if [ -d "$$d" ]; then echo "✅ $$d"; else echo "⚠️  $$d 不存在（执行 make bootstrap）"; fi; \
	done
	@echo ""
	@echo "== goose =="
	@if [ -x "$(GOOSE_BIN)" ]; then \
		echo "✅ $(GOOSE_BIN)"; \
		$(GOOSE_BIN) -version; \
	else \
		echo "❌ 未安装 goose（执行 make bootstrap）"; \
	fi
	@echo ""
	@echo "== jq =="
	@if command -v jq >/dev/null 2>&1; then jq --version; else echo "⚠️  未安装 jq（smoke / healthcheck 会降级）"; fi
	@echo ""
	@echo "== 端口 8000 =="
	@if command -v lsof >/dev/null 2>&1; then \
		if lsof -i :8000 >/dev/null 2>&1; then echo "⚠️  端口 8000 被占用"; else echo "✅ 端口 8000 空闲"; fi; \
	else \
		echo "（未安装 lsof，跳过端口检查）"; \
	fi
# ---------------------------------------------------------------------------
# 冒烟测试：校验 /health 中的 code/status/database
# ---------------------------------------------------------------------------
smoke:
	@echo "== 校验 /health =="
	@RESP=$$(curl -fsS http://localhost:8000/health 2>/dev/null) || { echo "❌ 无法访问 /health"; exit 1; }; \
	if command -v jq >/dev/null 2>&1; then \
		echo "$$RESP" | jq -e '.code == 0 and .data.status == "ok" and .data.database == "connected"' >/dev/null \
			&& echo "✅ smoke 通过" \
			|| { echo "❌ smoke 失败: $$RESP"; exit 1; }; \
	else \
		echo "$$RESP" | grep -q '"code":0' \
			&& echo "$$RESP" | grep -q '"status":"ok"' \
			&& echo "$$RESP" | grep -q '"database":"connected"' \
			&& echo "✅ smoke 通过" \
			|| { echo "❌ smoke 失败: $$RESP"; exit 1; }; \
	fi
# ---------------------------------------------------------------------------
# 常用命令
# ---------------------------------------------------------------------------
tidy:
	$(GO) mod tidy
run:
	@mkdir -p data logs
	$(GO) run ./cmd/server
build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/server ./cmd/server
	@echo "✅ 构建完成: $(BIN_DIR)/server"
	@ls -lh $(BIN_DIR)/server
start-prod:
	@if [ ! -x $(BIN_DIR)/server ]; then \
		echo "❌ 未找到 $(BIN_DIR)/server，请先执行 make build"; \
		exit 1; \
	fi
	@mkdir -p data logs
	exec $(BIN_DIR)/server
test:
	$(GO) test ./... -v
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "未安装 golangci-lint，跳过"; \
	fi
# P0 阶段仅支持 SQLite 迁移；PostgreSQL 迁移在 P1 认证模块启动后补齐。
migrate-up:
	@if [ ! -x "$(GOOSE_BIN)" ]; then \
		echo "❌ 未安装 goose（执行 make bootstrap）"; \
		exit 1; \
	fi
	@echo "执行 SQLite 迁移: $(SQLITE_PATH)"
	$(GOOSE_BIN) -dir migrations sqlite3 $(SQLITE_PATH) up
migrate-create:
	@if [ -z "$(name)" ]; then echo "用法：make migrate-create name=<migration_name>"; exit 1; fi
	@if [ ! -x "$(GOOSE_BIN)" ]; then \
		echo "❌ 未安装 goose（执行 make bootstrap）"; \
		exit 1; \
	fi
	$(GOOSE_BIN) -dir migrations create $(name) sql
clean:
	rm -rf $(BIN_DIR)
	@rm -f data/*.db logs/*.log
# 一键重置本地运行数据（评审第四节建议）
reset:
	@rm -rf $(BIN_DIR) data/*.db logs/*.log
	@echo "✅ 已重置本地运行数据（bin/、data/*.db、logs/*.log）"
help:
	@echo "可用目标："
	@echo "  bootstrap      一键初始化（生成 .env、装依赖、装 goose、建目录）"
	@echo "  doctor         环境自检"
	@echo "  smoke          冒烟测试 /health"
	@echo "  run            开发模式启动"
	@echo "  build          生成静态二进制 bin/server"
	@echo "  start-prod     生产模式启动"
	@echo "  test           运行单元测试"
	@echo "  lint           运行 golangci-lint"
	@echo "  migrate-up     执行 SQLite 迁移"
	@echo "  migrate-create 创建新迁移（name=<name>）"
	@echo "  reset          重置本地运行数据"
	@echo "  clean          清理 bin/、data/*.db、logs/*.log"
