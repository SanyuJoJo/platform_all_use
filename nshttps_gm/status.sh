#!/usr/bin/env bash
# ============================================================
# nshttps_gm 状态脚本
# ============================================================
set -uo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
PID_FILE="$NSG_HOME/pid/nginx.pid"

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  PID="$(cat "$PID_FILE")"
  echo "[nshttps_gm] RUNNING  PID=$PID"
  ps -o pid,lstart,cmd -p "$PID" 2>/dev/null | tail -1
  exit 0
else
  echo "[nshttps_gm] STOPPED"
  exit 1
fi