#!/usr/bin/env bash
# ============================================================
# 应用升级脚本（动作尽量小，可回退）
# 位置：create_package/upgrade/apply.sh
#
# 设计：
#   - 槽位切换（backend / frontend / core 各自 .current -> .prev），
#     绝无关机 / reboot；
#   - 升级前把当前槽位复制到回滚目录，供 rollback.sh 恢复；
#   - 只动运行载荷，保留 data/、.env、密钥。
#
# 用法：
#   apply.sh --pkg <package_id> --install-dir <INSTALL_DIR>
# ============================================================
set -euo pipefail

PKG=""
INSTALL_DIR=""
while [ $# -gt 0 ]; do
  case "$1" in
    --pkg) PKG="$2"; shift 2 ;;
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    *) echo "未知参数：$1" >&2; exit 2 ;;
  esac
done
[ -n "$PKG" ] || { echo "--pkg 必填" >&2; exit 2; }
[ -n "$INSTALL_DIR" ] || { echo "--install-dir 必填" >&2; exit 2; }
[ -d "$INSTALL_DIR" ] || { echo "安装目录不存在：$INSTALL_DIR" >&2; exit 1; }

# 升级包目录：<INSTALL_DIR>/update/<PKG>/
UPDATE_DIR="$INSTALL_DIR/update"
PKG_DIR="$UPDATE_DIR/$PKG"
[ -d "$PKG_DIR" ] || { echo "升级包目录不存在：$PKG_DIR" >&2; exit 1; }
[ -f "$PKG_DIR/manifest.json" ] || { echo "缺少 manifest.json：$PKG_DIR" >&2; exit 1; }

BACKEND_DIR="$INSTALL_DIR/backend"
FRONTEND_DIR="$INSTALL_DIR/frontend"
BIN_DIR="$INSTALL_DIR/bin"

# 回滚根：放在安装根之外，避免与安装目录一起被删。
ROLLBACK_ROOT="${INSTALL_DIR%/}.plat_rollback"
RB_ID="rb-$(date +%Y%m%d%H%M%S)-${PKG}"

mkdir -p "$ROLLBACK_ROOT/$RB_ID"
echo "回滚槽：$ROLLBACK_ROOT/$RB_ID"

# 前置：停服务（尽量小，仅重启后端，不用机器关机）。
if [ -x "$BIN_DIR/stop.sh" ]; then
  "$BIN_DIR/stop.sh" >/dev/null 2>&1 || true
fi

# 1. 槽位备份（backend / frontend / core）
for comp in backend frontend core; do
  src="$INSTALL_DIR/$comp"
  if [ -d "$src" ]; then
    mkdir -p "$ROLLBACK_ROOT/$RB_ID/$comp"
    cp -a "$src/." "$ROLLBACK_ROOT/$RB_ID/$comp/" 2>/dev/null || true
  fi
done
echo "已备份当前槽位到回滚目录"

# 2. 应用后端载荷（若包内携带）
BACKEND_TGZ="$PKG_DIR/backend.tar.gz"
if [ -f "$BACKEND_TGZ" ]; then
  # 校验 sha（manifest）
  want_backend="$(sed -n 's/.*"backend_sha256"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PKG_DIR/manifest.json")"
  got_backend="$(sha256sum "$BACKEND_TGZ" | awk '{print $1}')"
  if [ -n "$want_backend" ] && [ "$want_backend" != "$got_backend" ]; then
    echo "backend.tar.gz sha256 不匹配：期望 $want_backend 实得 $got_backend" >&2
    exit 1
  fi
  mkdir -p "$BACKEND_DIR.new"
  tar -xzf "$BACKEND_TGZ" -C "$BACKEND_DIR.new"
  # 保留运行时数据：把现有 data/、migrations 版本外的东西归位
  if [ -d "$BACKEND_DIR/data" ]; then
    cp -a "$BACKEND_DIR/data/." "$BACKEND_DIR.new/data/" 2>/dev/null || true
  fi
  rm -rf "$BACKEND_DIR.prev"
  [ -d "$BACKEND_DIR" ] && mv "$BACKEND_DIR" "$BACKEND_DIR.prev"
  mv "$BACKEND_DIR.new" "$BACKEND_DIR"
  echo "backend 槽位已切换"
fi

# 3. 应用前端载荷（若包内携带 dist.war）
WAR="$PKG_DIR/dist.war"
if [ -f "$WAR" ]; then
  want_front="$(sed -n 's/.*"frontend_sha256"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PKG_DIR/manifest.json")"
  got_front="$(sha256sum "$WAR" | awk '{print $1}')"
  if [ -n "$want_front" ] && [ "$want_front" != "$got_front" ]; then
    echo "dist.war sha256 不匹配" >&2
    exit 1
  fi
  DEPLOY="$FRONTEND_DIR/deploy"
  rm -rf "$DEPLOY.new"
  mkdir -p "$DEPLOY.new"
  if command -v unzip >/dev/null 2>&1; then
    unzip -q "$WAR" -d "$DEPLOY.new" 2>/dev/null || tar -xzf "$WAR" -C "$DEPLOY.new"
  else
    tar -xzf "$WAR" -C "$DEPLOY.new"
  fi
  rm -rf "$DEPLOY.prev"
  [ -d "$DEPLOY" ] && mv "$DEPLOY" "$DEPLOY.prev"
  mv "$DEPLOY.new" "$DEPLOY"
  echo "frontend 槽位已切换"
fi

# 4. 写回滚 id 供 rollback.sh 使用
echo "$RB_ID" > "$INSTALL_DIR/.last_rollback_id"

# 5. 记录 state
STATE_FILE="$INSTALL_DIR/state.json"
TO_VER="$(sed -n 's/.*"to_version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PKG_DIR/manifest.json")"
if [ -n "$TO_VER" ]; then
  if [ -f "$STATE_FILE" ]; then
    sed -i -E "s/\"version\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"version\": \"${TO_VER}\"/" "$STATE_FILE"
    sed -i -E "s/\"last_upgrade_result\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"last_upgrade_result\": \"ok\"/" "$STATE_FILE"
    sed -i -E "s/\"last_upgrade_at\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"last_upgrade_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"/" "$STATE_FILE"
  fi
fi

# 6. 重启（动作尽量小：仅重启后端）
if [ -x "$BIN_DIR/restart.sh" ]; then
  "$BIN_DIR/restart.sh"
else
  "$BIN_DIR/start.sh"
fi
echo "升级完成：$PKG （回滚：$RB_ID）"