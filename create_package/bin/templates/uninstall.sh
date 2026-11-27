#!/usr/bin/env bash
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PURGE=0
SERVICE_NAME="platform-backend"
INITD="/etc/init.d/$SERVICE_NAME"
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

# 停止服务（优先 init.d；保留对历史 systemd unit 的提示，但不再执行 systemctl）。
if [ -x "$INITD" ]; then
  echo "停止服务（$INITD）..."
  "$INITD" stop 2>/dev/null || true
elif [ -x "$INSTALL_DIR/bin/stop.sh" ]; then
  "$INSTALL_DIR/bin/stop.sh" || true
fi
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
  echo "⚠️  检测到历史 systemd unit：$SERVICE_NAME"
  echo "    如需彻底移除，请以 root 执行："
  echo "      systemctl disable $SERVICE_NAME && rm -f /etc/systemd/system/${SERVICE_NAME}.service && systemctl daemon-reload"
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