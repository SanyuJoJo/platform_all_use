#!/usr/bin/env bash
# 健康检查脚本（v1.2，v1.3 未变更）
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
PORT="@PORT@"
BACKEND_DIR="@BACKEND_DIR@"
URL="http://127.0.0.1:${PORT}/health"
RESP="$(curl -fsS --max-time 5 "$URL" 2>/dev/null)" || {
  echo "UNHEALTHY: 无法访问 $URL"
  exit 1
}
PY="$BACKEND_DIR/.venv/bin/python"
RC=1
if [ -x "$PY" ]; then
  if echo "$RESP" | "$PY" -c '
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
if [ "$RC" -eq 0 ]; then
  echo "HEALTHY: $URL"
  exit 0
else
  echo "UNHEALTHY: 响应异常：$RESP"
  exit 1
fi
