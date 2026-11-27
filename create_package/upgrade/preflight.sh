#!/usr/bin/env bash
# ============================================================
# 升级前体检（只读，不修改任何东西）
# 位置：create_package/upgrade/preflight.sh
#
# 用法：
#   preflight.sh --install-dir <INSTALL_DIR> [--pkg <package_id>]
# 退出码：
#   0 = 可升级；1 = 阻断；2 = 参数错误
# ============================================================
set -euo pipefail

INSTALL_DIR=""
PKG=""
while [ $# -gt 0 ]; do
  case "$1" in
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    --pkg) PKG="$2"; shift 2 ;;
    *) echo "未知参数：$1" >&2; exit 2 ;;
  esac
done
[ -n "$INSTALL_DIR" ] || { echo "--install-dir 必填" >&2; exit 2; }
[ -d "$INSTALL_DIR" ] || { echo "安装目录不存在：$INSTALL_DIR" >&2; exit 1; }

FAIL=0
warn() { echo "  ⚠️  $*"; }
ok()   { echo "  ✅ $*"; }
bad()  { echo "  ❌ $*" >&2; FAIL=1; }

echo "升级前体检：$INSTALL_DIR"

# 1. 后端二进制在运行？
BIN_DIR="$INSTALL_DIR/bin"
PID_FILE="$INSTALL_DIR/run/backend.pid"
if [ -f "$PID_FILE" ] && [ -n "$(cat "$PID_FILE")" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  ok "后端运行中（pid $(cat "$PID_FILE")）"
else
  warn "后端未运行（可升级，但应用后需手动启动）"
fi

# 2. 磁盘空间（需要约 3x 载荷）
if command -v df >/dev/null 2>&1; then
  avail_kb="$(df -Pk "$INSTALL_DIR" 2>/dev/null | awk 'NR==2{print $4}')"
  echo "  可用空间：$((avail_kb/1024)) MB"
fi

# 3. .env 权限
ENV_FILE="$INSTALL_DIR/backend/.env"
if [ -f "$ENV_FILE" ]; then
  perm="$(stat -c%a "$ENV_FILE" 2>/dev/null || echo '?')"
  echo "  .env 权限：$perm（建议 600）"
else
  ok ".env 不存在（环境变量模式，可选）"
fi

# 4. 指定升级包时的校验
if [ -n "$PKG" ]; then
  PKG_DIR="$INSTALL_DIR/update/$PKG"
  if [ -d "$PKG_DIR" ] && [ -f "$PKG_DIR/manifest.json" ]; then
    ok "升级包存在：$PKG_DIR"
    # manifest 中的 to_version 必须 > 当前 version（state.json）
    CUR="$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$INSTALL_DIR/state.json" 2>/dev/null || true)"
    TO="$(sed -n 's/.*"to_version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PKG_DIR/manifest.json")"
    echo "  当前版本：$CUR → 升级到：$TO"
  else
    bad "升级包不存在：$PKG_DIR"
  fi
fi

# 5. 数据目录可读
if [ -d "$INSTALL_DIR/user_data/cert/ca" ]; then
  ok "证书目录存在"
else
  warn "证书目录为空（可能尚未生成 CA）"
fi

[ "$FAIL" -eq 0 ] && echo "体检通过" || echo "体检存在阻断项" >&2
exit "$FAIL"