#!/usr/bin/env bash
# 健康检查脚本模板（Go 版 v1.3）
# v1.1 修复 P1-03：从 @ENV_FILE@ 提取端口，兼容 SERVER_PORT / PORT
# v1.2 修复 N-04：仅提取端口变量，避免 set -a 全量导出敏感变量
# v1.3 新增 P2-FE-07：额外输出前端入口存在性（不阻断健康检查结果）
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
ENV_FILE="@ENV_FILE@"
# ★ v1.2（N-04）：仅提取端口，避免全量导出
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
# ★ v1.3（P2-FE-07）：提取前端目录
_resolve_frontend_dir() {
  local f="$1"
  [ -f "$f" ] || { echo ""; return; }
  grep -E '^FRONTEND_DEPLOY_DIR=' "$f" | head -1 \
    | sed -E "s/^FRONTEND_DEPLOY_DIR=[\"']?//; s/[\"']$//" || true
}
PORT="$(_resolve_port "$ENV_FILE")"
FRONTEND_DIR="$(_resolve_frontend_dir "$ENV_FILE")"
[ -z "$FRONTEND_DIR" ] && FRONTEND_DIR="$INSTALL_DIR/frontend/deploy"
URL="http://127.0.0.1:${PORT}/health"
# ---------- /health 校验 ----------
RESP="$(curl -fsS --max-time 5 "$URL" 2>/dev/null)" || {
  echo "UNHEALTHY: 无法访问 $URL"
  exit 1
}
RC=1
if command -v python3 >/dev/null 2>&1; then
  if echo "$RESP" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
    ok = d.get("code") == 0 and d.get("data", {}).get("database") == "connected"
    sys.exit(0 if ok else 1)
except Exception:
    sys.exit(1)
'; then
    RC=0
  fi
fi
if [ "$RC" -ne 0 ]; then
  if echo "$RESP" | grep -Eq '"code"[[:space:]]*:[[:space:]]*0' \
     && echo "$RESP" | grep -Eq '"database"[[:space:]]*:[[:space:]]*"connected"'; then
    RC=0
  fi
fi
if [ "$RC" -ne 0 ]; then
  echo "UNHEALTHY: 响应异常：$RESP"
  exit 1
fi
echo "HEALTHY: $URL"
# ---------- v1.3 P2-FE-07：前端提示（不阻断） ----------
if [ -f "$FRONTEND_DIR/main-app/index.html" ]; then
  echo "  frontend: main-app/index.html 存在"
else
  echo "  frontend: ⚠️  main-app/index.html 缺失（$FRONTEND_DIR/main-app/index.html）"
  echo "            → 前端托管未启用；若预期使用前端托管，请检查 dist.war 解压结果"
fi
exit 0
