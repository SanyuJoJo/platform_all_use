#!/usr/bin/env bash
# ============================================================================
# 一键停止后端服务
#
# v1.1 修复：
#   - P2-6：停止前校验进程 command 是否包含 uvicorn，避免误杀同名进程。
#
# 用法：
#   ./scripts/stop.sh
#
# 行为：
#   1. 读取 backend.pid
#   2. 校验进程 command 是否为 uvicorn（防误杀）
#   3. SIGTERM 优雅停止（等待最多 30 秒）
#   4. 超时则 SIGKILL
#   5. 删除 PID 文件
# ============================================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$BACKEND_DIR/backend.pid"
if [ ! -f "$PID_FILE" ]; then
    echo "⚠️  未找到 PID 文件（$PID_FILE），服务可能未运行"
    exit 0
fi
PID=$(cat "$PID_FILE")
if ! kill -0 "$PID" 2>/dev/null; then
    echo "⚠️  进程不存在（PID=$PID），清理 PID 文件"
    rm -f "$PID_FILE"
    exit 0
fi
# ★ P2-6：校验进程 command 是否为 uvicorn
if command -v ps >/dev/null 2>&1; then
    PROC_CMD=$(ps -p "$PID" -o command= 2>/dev/null || true)
    if [ -n "$PROC_CMD" ]; then
        if ! echo "$PROC_CMD" | grep -q "uvicorn"; then
            echo "⚠️  PID $PID 不是 uvicorn 进程（command：${PROC_CMD:0:120}）"
            echo "   可能为 PID 复用，跳过 kill 并清理 PID 文件"
            rm -f "$PID_FILE"
            exit 0
        fi
    fi
fi
echo "🛑 停止服务（PID=$PID）..."
kill -TERM "$PID"
# 等待至多 30 秒
for i in $(seq 1 30); do
    if ! kill -0 "$PID" 2>/dev/null; then
        echo "✅ 服务已停止"
        rm -f "$PID_FILE"
        exit 0
    fi
    sleep 1
done
echo "⚠️  优雅停止超时，强制终止（SIGKILL）..."
kill -KILL "$PID" 2>/dev/null || true
sleep 1
if kill -0 "$PID" 2>/dev/null; then
    echo "❌ 无法终止进程（PID=$PID），请手动处理"
    exit 1
fi
rm -f "$PID_FILE"
echo "✅ 服务已强制停止"
