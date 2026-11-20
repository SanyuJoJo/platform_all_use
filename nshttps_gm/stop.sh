#!/usr/bin/env bash
# ============================================================
# nshttps_gm 停止脚本
# ============================================================
set -euo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
NGINX_BIN="$NSG_HOME/tengine-3.1.0/bin/sbin/nginx"
PID_FILE="$NSG_HOME/pid/nginx.pid"

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  PID="$(cat "$PID_FILE")"
  echo "[nshttps_gm] 停止 PID=$PID ..."
  "$NGINX_BIN" -s quit -p "$NSG_HOME" -c "$NSG_HOME/conf/nshttps.conf" 2>/dev/null || true
  # 兜底：quit 优雅退出，等待后强制
  for i in $(seq 1 10); do
    kill -0 "$PID" 2>/dev/null || break
    sleep 1
  done
  if kill -0 "$PID" 2>/dev/null; then
    echo "[nshttps_gm] 强制结束 PID=$PID"
    kill -9 "$PID" 2>/dev/null || true
  fi
  rm -f "$PID_FILE"
  echo "[nshttps_gm] 已停止"
else
  echo "[nshttps_gm] 未在运行"
fi