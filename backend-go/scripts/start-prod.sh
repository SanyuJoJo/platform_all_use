#!/usr/bin/env bash
# ============================================================================
# 启动脚本（生产模式）
#
# 使用 bin/server 静态二进制启动。
# 前置条件：已执行 ./scripts/build.sh 完成构建。
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."
if [ ! -f .env ]; then
    echo "❌ 缺少 .env，请先 cp .env.example .env 并修改关键配置"
    exit 1
fi
if [ ! -x bin/server ]; then
    echo "❌ 未找到 bin/server，请先执行 ./scripts/build.sh"
    exit 1
fi
mkdir -p data logs
exec ./bin/server
