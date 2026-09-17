#!/usr/bin/env bash
# 状态查询脚本模板（Go 版 v1.2）
# v1.2 修复：
#   - N-01：从 @ENV_FILE@ 提取端口，兼容 SERVER_PORT / PORT 别名
#   - N-04：仅提取端口变量，避免 set -a 全量导出
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
ENV_FILE="@ENV_FILE@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
# ★ v1.2（N-01 / N-04）：从 .env 提取端口
# 优先级：SERVER_PORT > PORT > @PORT@（install.sh 参数值）
_resolve_port() {
  local f="$1"
  [ -f "$f" ] || { echo "@PORT@"; return; }
  local p
  p="$(grep -E '^SERVER_PORT=' "$f" | head -1 \
       | sed -E "s/^SERVER_PORT=[\"']?//; s/[\"']$//" || true)"
  if [ -z "$p" ]; then
    p="$(grep -E '^PORT=' "$f" | head -1 \
         | sed -E "s/^PORT=[\"']?//; s/[\"']$//" || true)"
  fi
  echo "${p:-@PORT@}"
}
PORT="$(_resolve_port "$ENV_FILE")"
echo "安装目录：$INSTALL_DIR"
echo "监听端口：$PORT"
PORT_LINE=""
if command -v ss >/dev/null 2>&1; then
  PORT_LINE="$(ss -ltnp "sport = :$PORT" 2>/dev/null || true)"
  if [ -z "$PORT_LINE" ]; then
    PORT_LINE="$(ss -ltnp 2>/dev/null | grep -E "[:.]${PORT}\b" || true)"
  fi
elif command -v netstat >/dev/null 2>&1; then
  PORT_LINE="$(netstat -ltnp 2>/dev/null | grep -E "[:.]${PORT}\b" || true)"
fi
if [ -n "$PORT_LINE" ]; then
  echo "端口状态：已监听"
  echo "$PORT_LINE"
else
  echo "端口状态：未监听"
fi
if [ ! -f "$PID_FILE" ]; then
  echo "服务状态：未运行（无 PID 文件）"
  exit 1
fi
PID="$(cat "$PID_FILE" 2>/dev/null || true)"
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "服务状态：未运行（PID 无效）"
  exit 1
fi
echo "服务状态：运行中，PID=$PID"
if "$INSTALL_DIR/bin/healthcheck.sh" >/dev/null 2>&1; then
  echo "健康检查：通过"
  exit 0
else
  echo "健康检查：失败"
  exit 1
fi
