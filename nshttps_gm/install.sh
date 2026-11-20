#!/usr/bin/env bash
# ============================================================
# nshttps_gm 一键部署脚本（打进安装包，由 build-package.sh 生成）
#
# 用法：
#   bash install.sh [--prefix /opt/nshttps_gm] [--port 8443] [--upstream 127.0.0.1:8000]
#     --prefix    部署目录（默认 /opt/nshttps_gm）
#     --port      HTTPS 监听端口（默认 8443）
#     --upstream  反代上游（默认 127.0.0.1:8000）
#
# 部署目录结构：
#   <prefix>/
#   ├── tengine-3.1.0/      # 已编译 Tengine nginx
#   ├── tongsuo-bin/        # 内部集成的 Tongsuo（铜锁）运行库
#   ├── gen-certs.sh  gen-config.sh  start.sh  stop.sh  status.sh
#   └── conf/  certs/  logs/  pid/  tmp/  run/
# ============================================================
set -euo pipefail

PREFIX="/opt/nshttps_gm"
PORT=8443
UPSTREAM="127.0.0.1:8000"
while [ $# -gt 0 ]; do
  case "$1" in
    --prefix)   PREFIX="$2";   shift 2 ;;
    --port)     PORT="$2";     shift 2 ;;
    --upstream) UPSTREAM="$2"; shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

# ---------- 包内根目录 ----------
PKG_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

# ---------- 停止已存在的实例 ----------
if [ -f "$PREFIX/stop.sh" ]; then
  echo "[install] 停止旧实例..."
  bash "$PREFIX/stop.sh" >/dev/null 2>&1 || true
fi

echo "[install] 部署到：$PREFIX"
mkdir -p "$PREFIX"
# 复制运行产物（tengine + tongsuo-bin）与脚本（到根目录）
cp -a "$PKG_DIR/runtime/." "$PREFIX/"
cp -a "$PKG_DIR/scripts/." "$PREFIX/"

# 生成启动需要的目录
mkdir -p "$PREFIX/conf" "$PREFIX/logs" "$PREFIX/pid" "$PREFIX/tmp" "$PREFIX/run" "$PREFIX/certs"

chmod +x "$PREFIX/gen-certs.sh" "$PREFIX/gen-config.sh" "$PREFIX/start.sh" "$PREFIX/stop.sh" "$PREFIX/status.sh"
chmod +x "$PREFIX/tengine-3.1.0/bin/sbin/nginx" 2>/dev/null || true

echo "[install] 部署完成：$PREFIX"
echo
echo "后续步骤："
echo "  1) cd $PREFIX && bash gen-certs.sh --cn <你的域名>"
echo "  2) bash gen-config.sh --port $PORT --upstream $UPSTREAM"
echo "  3) bash start.sh"
echo
echo "[install] 完成"