#!/usr/bin/env bash
# ============================================================================
# 一键启动后端服务（非容器部署）
#
# v1.1 修复：
#   - P1-2：健康检查改为调用 scripts/healthcheck.sh，
#           校验 `data.database == "connected"`，避免 DB 不可用时误判就绪。
#
# 用法：
#   ./scripts/start.sh                    # 使用默认配置
#   ./scripts/start.sh --skip-migrate     # 跳过数据库迁移
#
# 行为：
#   1. 检查 .env 与虚拟环境
#   2. 若已运行则拒绝启动
#   3. 执行 alembic upgrade head（除非 --skip-migrate）
#   4. nohup 后台启动 uvicorn（多 worker）
#   5. 调用 healthcheck.sh 轮询至多 30 秒
#   6. 写入 backend.pid
# ============================================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$BACKEND_DIR"
PID_FILE="$BACKEND_DIR/backend.pid"
LOG_DIR="$BACKEND_DIR/logs"
LOG_FILE="$LOG_DIR/backend.log"
HEALTHCHECK_SCRIPT="$SCRIPT_DIR/healthcheck.sh"
SKIP_MIGRATE=0
for arg in "$@"; do
    case "$arg" in
        --skip-migrate) SKIP_MIGRATE=1 ;;
        -h|--help)
            echo "用法：$0 [--skip-migrate]"
            exit 0
            ;;
        *)
            echo "❌ 未知参数：$arg"
            exit 1
            ;;
    esac
done
# ---------- 依赖检查 ----------
if [ ! -f ".env" ]; then
    echo "❌ 未找到 .env 文件。请先执行："
    echo "   cp .env.production.example .env"
    echo "   然后编辑 .env 修改 SECRET_KEY / LICENSE_SECRET_KEY 等关键配置"
    exit 1
fi
if [ ! -d ".venv" ]; then
    echo "❌ 未找到虚拟环境。请先执行："
    echo "   uv venv --python 3.12 && uv sync --all-groups"
    exit 1
fi
# shellcheck disable=SC1091
source .venv/bin/activate
if ! command -v uvicorn >/dev/null 2>&1; then
    echo "❌ 未找到 uvicorn，请确认已执行 uv sync"
    exit 1
fi
if [ ! -x "$HEALTHCHECK_SCRIPT" ]; then
    echo "❌ 健康检查脚本不存在或不可执行：$HEALTHCHECK_SCRIPT"
    exit 1
fi
# ---------- 检查是否已运行 ----------
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if kill -0 "$OLD_PID" 2>/dev/null; then
        echo "⚠️  服务已在运行（PID=$OLD_PID）。请先执行 ./scripts/stop.sh"
        exit 1
    else
        echo "⚠️  发现残留 PID 文件（$PID_FILE），已清理"
        rm -f "$PID_FILE"
    fi
fi
# ---------- 加载环境变量 ----------
set -a
# shellcheck disable=SC1091
source .env
set +a
HOST="${HOST:-0.0.0.0}"
PORT="${PORT:-8000}"
WORKERS="${WORKERS:-1}"
# ---------- 创建运行目录 ----------
mkdir -p "$LOG_DIR" "$BACKEND_DIR/uploads/modules"
# ---------- 数据库迁移 ----------
if [ "$SKIP_MIGRATE" -eq 0 ]; then
    echo "🚀 执行数据库迁移（alembic upgrade head）..."
    if ! alembic upgrade head; then
        echo "❌ 数据库迁移失败，启动中止"
        exit 1
    fi
    echo "✅ 数据库迁移完成"
else
    echo "⏭  已跳过数据库迁移（--skip-migrate）"
fi
# ---------- 后台启动 ----------
echo "🚀 启动后端服务（host=$HOST port=$PORT workers=$WORKERS）..."
nohup uvicorn src.main:app \
    --host "$HOST" \
    --port "$PORT" \
    --workers "$WORKERS" \
    > "$LOG_FILE" 2>&1 &
NEW_PID=$!
echo "$NEW_PID" > "$PID_FILE"
echo "   进程 PID：$NEW_PID"
echo "   日志文件：$LOG_FILE"
# ---------- 健康检查轮询（P1-2 修复：调用 healthcheck.sh 校验数据库）----------
echo "⏳ 等待服务就绪（最多 30 秒）..."
for i in $(seq 1 30); do
    if ! kill -0 "$NEW_PID" 2>/dev/null; then
        echo "❌ 进程已退出，请检查日志：$LOG_FILE"
        tail -n 40 "$LOG_FILE" || true
        rm -f "$PID_FILE"
        exit 1
    fi
    if HOST="127.0.0.1" PORT="$PORT" "$HEALTHCHECK_SCRIPT" > /dev/null 2>&1; then
        echo "✅ 服务已就绪：http://127.0.0.1:${PORT}/health"
        echo ""
        echo "   API 文档：http://127.0.0.1:${PORT}/docs"
        echo "   停止服务：./scripts/stop.sh"
        exit 0
    fi
    sleep 1
done
echo "❌ 服务 30 秒内未就绪，请检查日志：$LOG_FILE"
tail -n 40 "$LOG_FILE" || true
exit 1
