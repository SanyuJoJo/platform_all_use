#!/usr/bin/env bash
# ============================================================
# 升级回退脚本（可回退）
# 位置：create_package/upgrade/rollback.sh
#
# 从 apply.sh 写入的回滚槽 <INSTALL_DIR 之外>.plat_rollback/<ID> 恢复。
# 只恢复槽位，数据保留（rollback 不删数据）。
#
# 用法：
#   rollback.sh --id <rollback_id> --install-dir <INSTALL_DIR>
# ============================================================
set -euo pipefail

ID=""
INSTALL_DIR=""
while [ $# -gt 0 ]; do
  case "$1" in
    --id) ID="$2"; shift 2 ;;
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    *) echo "未知参数：$1" >&2; exit 2 ;;
  esac
done
[ -n "$ID" ] || { echo "--id 必填" >&2; exit 2; }
[ -n "$INSTALL_DIR" ] || { echo "--install-dir 必填" >&2; exit 2; }

# 防穿越：只允许 rollback id 为安全字符。
case "$ID" in
  *[/\\]*) echo "非法的 id" >&2; exit 2 ;;
esac

ROLLBACK_ROOT="${INSTALL_DIR%/}.plat_rollback"
RB_DIR="$ROLLBACK_ROOT/$ID"
[ -d "$RB_DIR" ] || { echo "回滚目录不存在：$RB_DIR" >&2; exit 1; }

BIN_DIR="$INSTALL_DIR/bin"

# 停服务（尽量小：仅重启后端，不用机器关机）。
if [ -x "$BIN_DIR/stop.sh" ]; then
  "$BIN_DIR/stop.sh" >/dev/null 2>&1 || true
fi

# 恢复各槽位
for comp in backend frontend core; do
  src="$RB_DIR/$comp"
  dst="$INSTALL_DIR/$comp"
  if [ -d "$src" ]; then
    rm -rf "$dst.prev"
    [ -d "$dst" ] && mv "$dst" "$dst.prev"
    mkdir -p "$dst"
    cp -a "$src/." "$dst/"
    echo "已恢复 $comp"
  fi
done

# 前端 deploy 单独处理（若存在于回滚目录）
if [ -d "$RB_DIR/frontend/deploy" ]; then
  DEPLOY="$INSTALL_DIR/frontend/deploy"
  rm -rf "$DEPLOY.prev"
  [ -d "$DEPLOY" ] && mv "$DEPLOY" "$DEPLOY.prev"
  mkdir -p "$DEPLOY"
  cp -a "$RB_DIR/frontend/deploy/." "$DEPLOY/"
  echo "已恢复 frontend/deploy"
fi

# 清理标记
rm -f "$INSTALL_DIR/.last_rollback_id"

# 记录 state 为回退
STATE_FILE="$INSTALL_DIR/state.json"
if [ -f "$STATE_FILE" ]; then
  sed -i -E 's/"last_upgrade_result"[[:space:]]*:[[:space:]]*"[^"]*"/"last_upgrade_result": "rolled_back"/' "$STATE_FILE"
  sed -i -E "s/\"last_upgrade_at\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"last_upgrade_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"/" "$STATE_FILE"
fi

# 重启
if [ -x "$BIN_DIR/restart.sh" ]; then
  "$BIN_DIR/restart.sh"
else
  "$BIN_DIR/start.sh"
fi
echo "回退完成：$ID"