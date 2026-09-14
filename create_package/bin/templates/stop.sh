#!/usr/bin/env bash
# 停止脚本（v1.1，v1.2/v1.3 未变更）
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
SERVICE_NAME="platform-backend"
if command -v systemctl >/dev/null 2>&1 \
   && systemctl list-unit-files 2>/dev/null | grep -q "^${SERVICE_NAME}\.service"; then
  if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "通过 systemd 停止服务..."
    systemctl stop "$SERVICE_NAME"
    echo "已停止"
  else
    echo "systemd 服务未运行"
  fi
  exit 0
fi
if [ ! -f "$PID_FILE" ]; then
  echo "未找到 PID 文件，服务可能未运行"
  exit 0
fi
PID="$(cat "$PID_FILE" 2>/dev/null || true)"
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "进程不存在，清理 PID 文件"
  rm -f "$PID_FILE"
  exit 0
fi
if command -v ps >/dev/null 2>&1; then
  CMD="$(ps -p "$PID" -o command= 2>/dev/null || true)"
  if [ -n "$CMD" ] && ! echo "$CMD" | grep -q "uvicorn"; then
    echo "PID $PID 不是 uvicorn 进程，跳过 kill：$CMD"
    rm -f "$PID_FILE"
    exit 0
  fi
fi
echo "停止服务 PID=$PID ..."
kill -TERM "$PID"
for i in $(seq 1 30); do
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "服务已停止"
    rm -f "$PID_FILE"
    exit 0
  fi
  sleep 1
done
echo "优雅停止超时，强制终止"
kill -KILL "$PID" 2>/dev/null || true
sleep 1
if kill -0 "$PID" 2>/dev/null; then
  echo "❌ 无法终止进程 PID=$PID，请手动处理"
  exit 1
fi
rm -f "$PID_FILE"
echo "服务已强制停止"
