#!/usr/bin/env bash
# ============================================================
# nshttps_gm 统一入口代理载荷构建脚本
#
# 构建 nshttps_gm 自包含包（Tengine + 内部 Tongsuo + 脚本），
# 复制为 create_package/bin/bin/nshttps_gm.tar.gz。
#
# 用法：
#   bash build-nshttps-payload.sh
#   环境变量：
#     NSG_VERSION  - nshttps_gm 版本（默认 1.0.0）
#     TONGSUO_RUNTIME_PREFIX - 部署 rpath 前缀（默认 /opt/plats_tool/nshttps_gm/tongsuo-bin）
#
# 依赖：
#   - 已有 nshttps_gm/ 目录（内部集成 tongsuo 源码 + 脚本）
#   - 已编译 nshttps_gm（若未编译，本脚本调用 build.sh --force 编译）
# ============================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
NSG_DIR="$ROOT_DIR/nshttps_gm"
PAYLOAD_DIR="$SCRIPT_DIR/bin/bin"
NSG_VERSION="${NSG_VERSION:-1.0.0}"
OUT_DIR="$NSG_DIR/out"
TARGET="$PAYLOAD_DIR/nshttps_gm.tar.gz"

log() { echo -e "[nshttps-payload] $*"; }
die() { echo -e "[nshttps-payload][ERROR] $*" >&2; exit 1; }

# ---------- 检查 nshttps_gm 目录 ----------
[ -d "$NSG_DIR" ] || die "缺 nshttps_gm 目录：$NSG_DIR"
[ -x "$NSG_DIR/build-package.sh" ] || die "缺 nshttps_gm/build-package.sh"
[ -x "$NSG_DIR/build.sh" ] || die "缺 nshttps_gm/build.sh"

mkdir -p "$PAYLOAD_DIR" "$OUT_DIR"

# ---------- 构建 nshttps_gm 部署包 ----------
log "调用 nshttps_gm/build-package.sh 构建自包含包..."
NSG_VERSION="$NSG_VERSION" bash "$NSG_DIR/build-package.sh" \
  || die "nshttps_gm 打包失败"

# ---------- 定位最新部署包 ----------
LATEST_PKG="$(ls -t "$OUT_DIR"/nshttps-gm-*.tar.gz 2>/dev/null | head -1)"
[ -n "$LATEST_PKG" ] || die "nshttps_gm 未生成部署包：$OUT_DIR/nshttps-gm-*.tar.gz"
log "最新部署包：$LATEST_PKG"

# ---------- 复制为统一载荷名 ----------
cp -f "$LATEST_PKG" "$TARGET"
log "已复制到：$TARGET（$(du -h "$TARGET" | cut -f1)）"

# ---------- 结构校验 ----------
TMP_CHK="$(mktemp -d)"
tar -xzf "$TARGET" -C "$TMP_CHK" 2>/dev/null || die "nshttps_gm.tar.gz 不是合法 tar.gz"
for f in \
  "runtime/tengine-3.1.0/bin/sbin/nginx" \
  "runtime/tongsuo-bin/bin/openssl" \
  "runtime/tongsuo-bin/libs/libcrypto.so.3" \
  "install.sh" ; do
  [ -e "$TMP_CHK/$f" ] || die "nshttps_gm.tar.gz 缺少：$f"
done
[ -x "$TMP_CHK/runtime/tengine-3.1.0/bin/sbin/nginx" ] || die "nshttps_gm.tar.gz nginx 不可执行"
rm -rf "$TMP_CHK"
log "✅ nshttps_gm.tar.gz 校验通过"

log "[nshttps-payload] 完成：$TARGET"