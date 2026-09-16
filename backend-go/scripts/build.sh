#!/usr/bin/env bash
# ============================================================================
# 构建脚本（生成静态二进制）
#
# - 显式创建 bin 目录，避免 go build 报错。
# - 关键文件存在性校验。
# - 使用纯 Go SQLite 驱动（github.com/glebarez/sqlite），
#   CGO_ENABLED=0 即可产出静态二进制。
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."
echo "🔍 校验关键文件..."
REQUIRED=("go.mod" "cmd/server" "internal")
MISSING=()
for item in "${REQUIRED[@]}"; do
    if [ ! -e "$item" ]; then
        MISSING+=("$item")
    fi
done
if [ "${#MISSING[@]}" -gt 0 ]; then
    echo "❌ 缺少关键文件/目录：${MISSING[*]}"
    exit 1
fi
echo "✅ 关键文件齐全"
echo "📦 创建输出目录..."
mkdir -p bin
echo "🛠  编译中（CGO_ENABLED=0）..."
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server
echo "✅ 构建完成: bin/server"
ls -lh bin/server
