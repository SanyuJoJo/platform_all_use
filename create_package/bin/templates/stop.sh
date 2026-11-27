#!/usr/bin/env bash
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
SERVICE_NAME="platform-backend"

# 停止服务。
#
# 顺序（v1.9）：
#   1. 优先走 /etc/init.d/platform-backend（默认自启方式，无 systemd 依赖）；
#   2. 退化为直接按 PID 停止（无 sysv 兼容层时也能用）；
#   3. 若检测到历史 systemd unit 但仍被安装且未迁移，仅告警并改由 PID 停止，
#      绝不再执行 systemctl（避免双启与 systemd 残留冲突）。
INITD="/etc/init.d/$SERVICE_NAME"

if [ -x "$INITD" ]; then
  echo "通过 $INITD 停止服务..."
  "$INITD" stop
  exit $?
fi

if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
  echo "⚠️  检测到历史 systemd unit，但已改用 init.d 自启，跳过 systemctl。"
  echo "    建议：systemctl disable $SERVICE_NAME 后删除 /etc/systemd/system/${SERVICE_NAME}.service"
fi

if [ ! -f "$PID_FILE" ]; then
  echo "未找到 PID 文件，服务可能未运行"
  exit 0
fi

PID="$(cat "$PID_FILE" 2>/dev/null || true)"
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "进程不存在，清理 PID 文件"
  rm -f "$PID_FILE"
  exit 0
fi

if command -v ps >/dev/null 2>&1; then
  CMD="$(ps -p "$PID" -o command= 2>/dev/null || true)"
  if [ -n "$CMD" ] && ! echo "$CMD" | grep -q "bin/server"; then
    echo "PID $PID 不是后端进程，跳过 kill：$CMD"
    rm -f "$PID_FILE"
    exit 0
  fi
fi

echo "停止服务 PID=$PID ..."
kill -TERM "$PID"
for i in $(seq 1 30); do
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "服务已停止"
    rm -f "$PID_FILE"
    exit 0
  fi
  sleep 1
done
echo "优雅停止超时，强制终止"
kill -KILL "$PID" 2>/dev/null || true
sleep 1
if kill -0 "$PID" 2>/dev/null; then
  echo "❌ 无法终止进程 PID=$PID，请手动处理"
  exit 1
fi
rm -f "$PID_FILE"
echo "服务已强制停止"