#!/usr/bin/env bash
# ============================================================
# nshttps_gm 启动脚本
# 依赖：已执行 gen-certs.sh 生成证书、gen-config.sh 生成配置
# ============================================================
set -euo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
NGINX_BIN="$NSG_HOME/tengine-3.1.0/bin/sbin/nginx"
CONF="$NSG_HOME/conf/nshttps.conf"
TONGSUO_LIB="$NSG_HOME/tongsuo-bin/libs"

[ -x "$NGINX_BIN" ] || { echo "缺 nginx：$NGINX_BIN（先运行 build.sh）" >&2; exit 2; }
[ -f "$CONF" ] || { echo "缺配置：$CONF（先运行 gen-config.sh）" >&2; exit 2; }
[ -f "$NSG_HOME/certs/server/server.crt" ] || { echo "缺证书（先运行 gen-certs.sh）" >&2; exit 2; }

export LD_LIBRARY_PATH="$TONGSUO_LIB:${LD_LIBRARY_PATH:-}"

if [ -f "$NSG_HOME/pid/nginx.pid" ] && kill -0 "$(cat "$NSG_HOME/pid/nginx.pid")" 2>/dev/null; then
  echo "[nshttps_gm] 已在运行 PID=$(cat "$NSG_HOME/pid/nginx.pid")"
  exit 0
fi

echo "[nshttps_gm] 测试配置语法..."
"$NGINX_BIN" -t -p "$NSG_HOME" -c "$CONF"
echo "[nshttps_gm] 启动..."
"$NGINX_BIN" -p "$NSG_HOME" -c "$CONF"
sleep 1
if [ -f "$NSG_HOME/pid/nginx.pid" ]; then
  echo "[nshttps_gm] 已启动 PID=$(cat "$NSG_HOME/pid/nginx.pid")"
else
  echo "[nshttps_gm] 启动失败，请查看 $NSG_HOME/logs/error.log"
  exit 1
fi