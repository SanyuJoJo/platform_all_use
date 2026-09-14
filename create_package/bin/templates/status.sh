#!/usr/bin/env bash
# 状态查询脚本（v1.1，v1.2/v1.3 未变更）
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
PORT="@PORT@"
echo "安装目录：$INSTALL_DIR"
echo "监听端口：$PORT"
PORT_LINE=""
if command -v ss >/dev/null 2>&1; then
  PORT_LINE="$(ss -ltnp "sport = :$PORT" 2>/dev/null || true)"
  if [ -z "$PORT_LINE" ]; then
    PORT_LINE="$(ss -ltnp 2>/dev/null | grep -E "[:.]${PORT}\b" || true)"
  fi
elif command -v netstat >/dev/null 2>&1; then
  PORT_LINE="$(netstat -ltnp 2>/dev/null | grep -E "[:.]${PORT}\b" || true)"
fi
if [ -n "$PORT_LINE" ]; then
  echo "端口状态：已监听"
  echo "$PORT_LINE"
else
  echo "端口状态：未监听"
fi
if [ ! -f "$PID_FILE" ]; then
  echo "服务状态：未运行（无 PID 文件）"
  exit 1
fi
PID="$(cat "$PID_FILE" 2>/dev/null || true)"
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "服务状态：未运行（PID 无效）"
  exit 1
fi
echo "服务状态：运行中，PID=$PID"
if "$INSTALL_DIR/bin/healthcheck.sh" >/dev/null 2>&1; then
  echo "健康检查：通过"
  exit 0
else
  echo "健康检查：失败"
  exit 1
fi
