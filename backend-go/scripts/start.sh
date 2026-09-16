#!/usr/bin/env bash
# ============================================================================
# 启动脚本（开发模式）
#
# 开发：使用 go run 直接启动，便于快速迭代。
# 生产：请使用 bin/server 二进制（scripts/start-prod.sh）。
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "已生成 .env，按需修改后重新启动"
fi
mkdir -p data logs
exec go run ./cmd/server
