#!/usr/bin/env bash
# ============================================================
# nshttps_gm 一键部署包构建脚本
#
# 功能：
#   1. 确保内部 Tongsuo 与 Tengine 已构建（调用 build.sh）
#   2. 打包为自包含安装包，含：
#        - runtime/  tengine nginx + tongsuo-bin（运行所需，无源码）
#        - scripts/  gen-certs.sh / gen-config.sh / start.sh / stop.sh / status.sh
#        - install.sh  一键部署脚本
#
# 用法：
#   bash build-package.sh [--out /path/to/out]
#   输出：nshttps_gm/out/nshttps-gm-<version>-<timestamp>.tar.gz
# ============================================================
set -euo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
OUT_DIR="${1:-$NSG_HOME/out}"
VERSION="${NSG_VERSION:-1.0.0}"
TS="$(date +%Y%m%d%H%M%S)"

# ---------- 1. 确保已构建 ----------
bash "$NSG_HOME/build.sh"

# ---------- 2. 校验运行产物 ----------
NGINX="$NSG_HOME/tengine-3.1.0/bin/sbin/nginx"
[ -x "$NGINX" ] || { echo "缺 nginx：$NGINX（先运行 build.sh）" >&2; exit 2; }
[ -x "$NSG_HOME/tongsuo-bin/bin/openssl" ] || { echo "缺 tongsuo：$NSG_HOME/tongsuo-bin" >&2; exit 2; }
[ -f "$NSG_HOME/tongsuo-bin/libs/libcrypto.so.3" ] || { echo "缺 libcrypto.so.3" >&2; exit 2; }

# ---------- 3. 组装暂存区 ----------
STAGE="$NSG_HOME/.pkg-stage"
rm -rf "$STAGE"
mkdir -p "$STAGE/runtime" "$STAGE/scripts"

# runtime：仅打包必要的 nginx 文件（二进制 + mime.types），不打包整个 tengine 源码树
mkdir -p "$STAGE/runtime/tengine-3.1.0/bin/sbin" "$STAGE/runtime/tengine-3.1.0/conf"
cp -f "$NSG_HOME/tengine-3.1.0/bin/sbin/nginx" "$STAGE/runtime/tengine-3.1.0/bin/sbin/nginx"
cp -f "$NSG_HOME/tengine-3.1.0/conf/mime.types" "$STAGE/runtime/tengine-3.1.0/conf/mime.types"
# tongsuo-bin（运行库）
cp -a "$NSG_HOME/tongsuo-bin" "$STAGE/runtime/"

# scripts
for s in gen-certs.sh gen-config.sh start.sh stop.sh status.sh; do
  cp "$NSG_HOME/$s" "$STAGE/scripts/"
  chmod +x "$STAGE/scripts/$s"
done

# install.sh
cp "$NSG_HOME/install.sh" "$STAGE/install.sh"
chmod +x "$STAGE/install.sh"

# ---------- 4. 打包 ----------
mkdir -p "$OUT_DIR"
PKG_NAME="nshttps-gm-${VERSION}-${TS}.tar.gz"
echo "[pkg] 打包 $PKG_NAME ..."
tar -czf "$OUT_DIR/$PKG_NAME" -C "$STAGE" . --exclude='./runtime/tengine-3.1.0/html' 2>/dev/null || \
tar -czf "$OUT_DIR/$PKG_NAME" -C "$STAGE" .

# ---------- 5. sha256 ----------
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$OUT_DIR" && sha256sum "$PKG_NAME") > "$OUT_DIR/$PKG_NAME.sha256"
fi

rm -rf "$STAGE"
echo
echo "[pkg] 完成：$OUT_DIR/$PKG_NAME"
echo "[pkg] 部署到服务器："
echo "  1) tar -xzf $PKG_NAME"
echo "  2) bash install.sh --prefix /opt/nshttps_gm --port 8443 --upstream 127.0.0.1:8000"
echo "  3) cd /opt/nshttps_gm && bash gen-certs.sh --cn <域名> && bash gen-config.sh && bash start.sh"