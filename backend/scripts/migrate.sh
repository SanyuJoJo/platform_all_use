bash

#!/usr/bin/env bash
# ============================================================
# 数据库迁移执行脚本 (Linux/macOS)
# 用法: ./scripts/migrate.sh [upgrade|downgrade|revision]
# 默认: upgrade head
# ============================================================
set -e
cd "$(dirname "$0")/.."  # 切换到项目根目录
ACTION=${1:-upgrade}
TARGET=${2:-head}
if [ ! -f ".env" ]; then
    echo "❌ .env 文件不存在，请先创建（参考 .env.example）"
    exit 1
fi
# 加载虚拟环境
if [ -d ".venv" ]; then
    source .venv/bin/activate
else
    echo "❌ 虚拟环境不存在，请先运行 ./scripts/setup_dev.sh"
    exit 1
fi
echo "🚀 执行迁移: alembic $ACTION $TARGET"
uv run alembic "$ACTION" "$TARGET"
if [ $? -eq 0 ]; then
    echo "✅ 迁移执行成功"
else
    echo "❌ 迁移执行失败"
    exit 1
fi
