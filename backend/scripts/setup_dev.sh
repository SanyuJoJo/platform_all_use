#!/usr/bin/env bash
set -e
echo "🚀 开始初始化后端开发环境..."
if ! python3 -c "import sys; sys.exit(0 if sys.version_info >= (3,12) else 1)" 2>/dev/null; then
    echo "❌ 需要 Python 3.12 或更高版本，当前版本：$(python3 --version 2>&1)"
    exit 1
fi
if ! command -v openssl &> /dev/null; then
    echo "❌ 未找到 openssl 命令，无法生成随机 SECRET_KEY，请先安装 openssl"
    exit 1
fi
if ! command -v uv &> /dev/null; then
    echo "📦 安装 uv..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    export PATH="$HOME/.cargo/bin:$PATH"
fi
echo "🐍 创建 Python 虚拟环境..."
uv venv --python 3.12
source .venv/bin/activate
echo "📦 安装项目依赖..."
uv sync --all-groups
if [ ! -f .env ]; then
    echo "📄 生成 .env 文件..."
    cp .env.example .env
    SECRET=$(openssl rand -hex 32)
    sed -i.bak "s/change-this-in-production-please/$SECRET/g" .env
    rm -f .env.bak
fi
echo "📁 创建 Python 包标识文件..."
touch src/__init__.py
touch src/core/__init__.py
if [ ! -f migrations/script.py.mako ]; then
    echo "📄 复制 Alembic 迁移模板..."
    TEMPLATE_PATH=$(find .venv -path "*/alembic/templates/async/script.py.mako" 2>/dev/null | head -1)
    if [ -n "$TEMPLATE_PATH" ]; then
        cp "$TEMPLATE_PATH" migrations/
        echo "   ✅ 模板复制成功"
    else
        echo "❌ 未找到 Alembic 迁移模板，请确保 alembic 已正确安装"
        echo "   可尝试手动执行：alembic init migrations -t async"
        exit 1
    fi
fi
echo "🗄️ 初始化数据库..."
if [ -z "$(ls -A migrations/versions/ 2>/dev/null)" ]; then
    echo "生成初始迁移脚本..."
    uv run alembic revision --autogenerate -m "init"
    if grep -q "pass" migrations/versions/*.py 2>/dev/null; then
        echo "⚠️ 警告：生成的迁移脚本可能为空（未检测到模型）。请确保在 migrations/env.py 中导入了所有模型。"
    fi
fi
uv run alembic upgrade head
echo "✅ 环境准备完成！启动服务..."
python -m uvicorn src.main:app --reload --host 0.0.0.0 --port 18000
