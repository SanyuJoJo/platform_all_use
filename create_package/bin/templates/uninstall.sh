#!/usr/bin/env bash
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PURGE=0
SERVICE_NAME="platform-backend"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
for arg in "$@"; do
  case "$arg" in
    --purge) PURGE=1 ;;
    -h|--help)
      echo "用法：$0 [--purge]"
      echo "  --purge  同时删除安装目录"
      exit 0
      ;;
    *) echo "未知参数：$arg"; exit 1 ;;
  esac
done
if [ -f "$SERVICE_FILE" ]; then
  if [ "$(id -u)" -ne 0 ]; then
    echo "⚠️  检测到 systemd 服务，但当前用户非 root，跳过清理"
    echo "    如需清理，请执行：sudo $0 $*"
  else
    echo "停止并禁用 systemd 服务..."
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload 2>/dev/null || true
    echo "已清理 systemd 服务"
  fi
else
  if [ -x "$INSTALL_DIR/bin/stop.sh" ]; then
    "$INSTALL_DIR/bin/stop.sh" || true
  fi
fi
if [ "$PURGE" -eq 1 ]; then
  case "$INSTALL_DIR" in
    ""|"/"|"/opt"|"/usr"|"/etc"|"/var"|"/root"|"$HOME")
      echo "❌ 拒绝删除危险路径：$INSTALL_DIR"
      exit 1
      ;;
  esac
  echo "删除安装目录：$INSTALL_DIR"
  rm -rf "$INSTALL_DIR"
  echo "✅ 已卸载并删除数据"
else
  echo "已停止服务，安装目录保留：$INSTALL_DIR"
  echo "如需删除，请执行：rm -rf $INSTALL_DIR"
fi
