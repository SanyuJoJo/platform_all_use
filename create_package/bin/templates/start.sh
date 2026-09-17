#!/usr/bin/env bash
# 启动脚本模板（Go 版 v1.1）
# v1.1 修复 P1-03：兼容 Python 版 HOST / PORT 别名
set -euo pipefail
umask 077
INSTALL_DIR="@INSTALL_DIR@"
BACKEND_DIR="@BACKEND_DIR@"
ENV_FILE="@ENV_FILE@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
LOCK_DIR="$INSTALL_DIR/run/.start.lock"
LOG_DIR="$INSTALL_DIR/logs"
LOG_FILE="$LOG_DIR/backend.log"
mkdir -p "$LOG_DIR" "$INSTALL_DIR/run"
if ! mkdir "$LOCK_DIR" 2>/dev/null; then
  echo "另一个 start.sh 正在运行（锁目录：$LOCK_DIR）"
  exit 1
fi
cleanup_lock() { rmdir "$LOCK_DIR" 2>/dev/null || true; }
trap cleanup_lock EXIT
if [ -f "$PID_FILE" ]; then
  OLD_PID="$(cat "$PID_FILE" 2>/dev/null || true)"
  if [ -n "$OLD_PID" ] && kill -0 "$OLD_PID" 2>/dev/null; then
    echo "服务已在运行，PID=$OLD_PID"
    exit 0
  fi
  rm -f "$PID_FILE"
fi
[ -f "$ENV_FILE" ] || { echo "缺少配置文件：$ENV_FILE"; exit 1; }
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
# ★ v1.1（P1-03）：兼容 Python 版 HOST / PORT 别名
SERVER_HOST="${SERVER_HOST:-${HOST:-@HOST@}}"
SERVER_PORT="${SERVER_PORT:-${PORT:-@PORT@}}"
# 端口占用检查
port_in_use=0
if command -v ss >/dev/null 2>&1; then
  if ss -ltn "sport = :$SERVER_PORT" 2>/dev/null | grep -q LISTEN; then
    port_in_use=1
  elif ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]${SERVER_PORT}\$"; then
    port_in_use=1
  fi
elif command -v netstat >/dev/null 2>&1; then
  if netstat -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]${SERVER_PORT}\$"; then
    port_in_use=1
  fi
fi
if [ "$port_in_use" -eq 1 ]; then
  echo "端口 $SERVER_PORT 已被占用："
  if command -v ss >/dev/null 2>&1; then
    ss -ltnp 2>/dev/null | grep -E "[:.]${SERVER_PORT}\b" || true
  elif command -v netstat >/dev/null 2>&1; then
    netstat -ltnp 2>/dev/null | grep -E "[:.]${SERVER_PORT}\b" || true
  fi
  exit 1
fi
cd "$BACKEND_DIR"
nohup "$BACKEND_DIR/bin/server" >> "$LOG_FILE" 2>&1 &
NEW_PID=$!
echo "$NEW_PID" > "$PID_FILE"
echo "服务已启动，PID=$NEW_PID，日志=$LOG_FILE"
for i in $(seq 1 30); do
  if "$INSTALL_DIR/bin/healthcheck.sh" >/dev/null 2>&1; then
    echo "服务健康检查通过：http://127.0.0.1:${SERVER_PORT}/health"
    exit 0
  fi
  if ! kill -0 "$NEW_PID" 2>/dev/null; then
    echo "进程已退出，请查看日志：$LOG_FILE"
    tail -n 50 "$LOG_FILE" || true
    exit 1
  fi
  sleep 1
done
echo "服务 30 秒内未通过健康检查，请查看日志：$LOG_FILE"
tail -n 50 "$LOG_FILE" || true
exit 1
