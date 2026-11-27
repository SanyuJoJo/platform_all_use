#!/usr/bin/env bash
# 启动脚本模板（Go 版 v1.1）
# v1.1 修复 P1-03：兼容 Python 版 HOST / PORT 别名
set -euo pipefail
umask 077
INSTALL_DIR="@INSTALL_DIR@"
BACKEND_DIR="@BACKEND_DIR@"
ENV_FILE="@ENV_FILE@"
PID_FILE="$INSTALL_DIR/run/backend.pid"
LOCK_DIR="$INSTALL_DIR/run/.start.lock"
LOG_DIR="$INSTALL_DIR/logs"
LOG_FILE="$LOG_DIR/backend.log"
mkdir -p "$LOG_DIR" "$INSTALL_DIR/run"
if ! mkdir "$LOCK_DIR" 2>/dev/null; then
  echo "另一个 start.sh 正在运行（锁目录：$LOCK_DIR）"
  exit 1
fi
cleanup_lock() { rmdir "$LOCK_DIR" 2>/dev/null || true; }
trap cleanup_lock EXIT
if [ -f "$PID_FILE" ]; then
  OLD_PID="$(cat "$PID_FILE" 2>/dev/null || true)"
  if [ -n "$OLD_PID" ] && kill -0 "$OLD_PID" 2>/dev/null; then
    echo "服务已在运行，PID=$OLD_PID"
    exit 0
  fi
  rm -f "$PID_FILE"
fi
# 配置文件（.env）为可选：生产环境由 start.sh 注入为环境变量，
# 测试环境可直接用外部环境变量，无需 .env。
# 若存在则 source（ENV_FILE 已是 install.sh 生成的绝对路径，直接 source 即可，
# 与当前工作目录无关）。
if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi
# ★ v1.1（P1-03）：兼容 Python 版 HOST / PORT 别名
SERVER_HOST="${SERVER_HOST:-${HOST:-@HOST@}}"
SERVER_PORT="${SERVER_PORT:-${PORT:-@PORT@}}"
# ★ v1.4：core 根目录兜底 + 导出铜锁动态库路径
#   CORE_ROOT 供 core 静态二进制识别 coreRoot；LD_LIBRARY_PATH 供
#   nginx / 铜锁 openssl 子进程在自定义安装目录下找到动态库。
CORE_ROOT="${CORE_ROOT:-$INSTALL_DIR/core}"
export CORE_ROOT
# ★ v1.9：无论是否有 .env，都把本次安装根作为环境变量注入，
#   让后端在不依赖 .env 时也能读到正确的路径，避免静默回退到默认根。
export PLATS_INSTALL_ROOT="${PLATS_INSTALL_ROOT:-$INSTALL_DIR}"
export PLATS_DEPLOY_ROOT="${PLATS_DEPLOY_ROOT:-$INSTALL_DIR}"
export PLATS_CORE_GO_BIN="${PLATS_CORE_GO_BIN:-$INSTALL_DIR/core/bin/core}"
export FRONTEND_DEPLOY_DIR="${FRONTEND_DEPLOY_DIR:-$INSTALL_DIR/frontend/deploy}"
export DATABASE_URL="${DATABASE_URL:-sqlite+aiosqlite:///$INSTALL_DIR/data/app.db}"
export MODULE_UPLOAD_DIR="${MODULE_UPLOAD_DIR:-$INSTALL_DIR/uploads/modules}"
if [ -d "$CORE_ROOT/libs/tongsuo/libs" ]; then
  if [ -n "${LD_LIBRARY_PATH:-}" ]; then
    export LD_LIBRARY_PATH="$CORE_ROOT/libs/tongsuo/libs:$LD_LIBRARY_PATH"
  else
    export LD_LIBRARY_PATH="$CORE_ROOT/libs/tongsuo/libs"
  fi
fi
# 端口占用检查
port_in_use=0
if command -v ss >/dev/null 2>&1; then
  if ss -ltn "sport = :$SERVER_PORT" 2>/dev/null | grep -q LISTEN; then
    port_in_use=1
  elif ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]${SERVER_PORT}\$"; then
    port_in_use=1
  fi
elif command -v netstat >/dev/null 2>&1; then
  if netstat -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]${SERVER_PORT}\$"; then
    port_in_use=1
  fi
fi
if [ "$port_in_use" -eq 1 ]; then
  echo "端口 $SERVER_PORT 已被占用："
  if command -v ss >/dev/null 2>&1; then
    ss -ltnp 2>/dev/null | grep -E "[:.]${SERVER_PORT}\b" || true
  elif command -v netstat >/dev/null 2>&1; then
    netstat -ltnp 2>/dev/null | grep -E "[:.]${SERVER_PORT}\b" || true
  fi
  exit 1
fi
cd "$BACKEND_DIR"
nohup "$BACKEND_DIR/bin/server" >> "$LOG_FILE" 2>&1 &
NEW_PID=$!
echo "$NEW_PID" > "$PID_FILE"
echo "服务已启动，PID=$NEW_PID，日志=$LOG_FILE"
for i in $(seq 1 30); do
  if "$INSTALL_DIR/bin/healthcheck.sh" >/dev/null 2>&1; then
    echo "服务健康检查通过：http://127.0.0.1:${SERVER_PORT}/health"
    exit 0
  fi
  if ! kill -0 "$NEW_PID" 2>/dev/null; then
    echo "进程已退出，请查看日志：$LOG_FILE"
    tail -n 50 "$LOG_FILE" || true
    exit 1
  fi
  sleep 1
done
echo "服务 30 秒内未通过健康检查，请查看日志：$LOG_FILE"
tail -n 50 "$LOG_FILE" || true
exit 1
