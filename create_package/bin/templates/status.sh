#!/usr/bin/env bash
# 状态查询脚本模板（Go 版 v1.3）
# v1.2 修复：
#   - N-01：从 @ENV_FILE@ 提取端口，兼容 SERVER_PORT / PORT 别名
#   - N-04：仅提取端口变量，避免 set -a 全量导出
# v1.3 修复：
#   - P2-FE-07：增加前端托管状态（前端目录、main-app 入口、子应用数量）
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
# ★ v1.3（P2-FE-07）：从 .env 提取 FRONTEND_DEPLOY_DIR
_resolve_frontend_dir() {
  local f="$1"
  [ -f "$f" ] || { echo ""; return; }
  grep -E '^FRONTEND_DEPLOY_DIR=' "$f" | head -1 \
    | sed -E "s/^FRONTEND_DEPLOY_DIR=[\"']?//; s/[\"']$//" || true
}
PORT="$(_resolve_port "$ENV_FILE")"
FRONTEND_DIR="$(_resolve_frontend_dir "$ENV_FILE")"
[ -z "$FRONTEND_DIR" ] && FRONTEND_DIR="$INSTALL_DIR/frontend/deploy"
echo "安装目录：$INSTALL_DIR"
echo "监听端口：$PORT"
# ---------- 端口状态 ----------
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
# ---------- PID 状态 ----------
if [ ! -f "$PID_FILE" ]; then
  echo "服务状态：未运行（无 PID 文件）"
  # 仍输出前端状态，便于排障
  _print_frontend_status() {
    echo "前端目录：$FRONTEND_DIR"
    if [ -f "$FRONTEND_DIR/main-app/index.html" ]; then
      echo "前端主应用：✅ 存在"
    else
      echo "前端主应用：❌ 缺失（$FRONTEND_DIR/main-app/index.html）"
    fi
    if [ -d "$FRONTEND_DIR/sub-apps" ]; then
      local n
      n="$(find "$FRONTEND_DIR/sub-apps" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')"
      echo "前端子应用数：$n"
    else
      echo "前端子应用数：0（sub-apps 缺失）"
    fi
  }
  _print_frontend_status
  exit 1
fi
PID="$(cat "$PID_FILE" 2>/dev/null || true)"
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "服务状态：未运行（PID 无效）"
  _print_frontend_status() {
    echo "前端目录：$FRONTEND_DIR"
    if [ -f "$FRONTEND_DIR/main-app/index.html" ]; then
      echo "前端主应用：✅ 存在"
    else
      echo "前端主应用：❌ 缺失"
    fi
    if [ -d "$FRONTEND_DIR/sub-apps" ]; then
      local n
      n="$(find "$FRONTEND_DIR/sub-apps" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')"
      echo "前端子应用数：$n"
    else
      echo "前端子应用数：0"
    fi
  }
  _print_frontend_status
  exit 1
fi
echo "服务状态：运行中，PID=$PID"
# ---------- 前端状态（v1.3 P2-FE-07） ----------
echo "前端目录：$FRONTEND_DIR"
if [ -f "$FRONTEND_DIR/main-app/index.html" ]; then
  echo "前端主应用：✅ 存在"
else
  echo "前端主应用：❌ 缺失（$FRONTEND_DIR/main-app/index.html）"
fi
if [ -d "$FRONTEND_DIR/sub-apps" ]; then
  SUB_APP_COUNT="$(find "$FRONTEND_DIR/sub-apps" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')"
  echo "前端子应用数：$SUB_APP_COUNT"
  if [ "$SUB_APP_COUNT" -gt 0 ]; then
    find "$FRONTEND_DIR/sub-apps" -mindepth 1 -maxdepth 1 -type d 2>/dev/null \
      | sort | while read -r d; do
        local_name="$(basename "$d")"
        if [ -f "$d/index.html" ]; then
          echo "  ✅ sub-apps/$local_name/index.html"
        else
          echo "  ❌ sub-apps/$local_name/index.html（缺失）"
        fi
      done
  fi
else
  echo "前端子应用数：0（sub-apps 缺失）"
fi
# ---------- 健康检查 ----------
if "$INSTALL_DIR/bin/healthcheck.sh" >/dev/null 2>&1; then
  echo "健康检查：通过"
  exit 0
else
  echo "健康检查：失败"
  exit 1
fi
