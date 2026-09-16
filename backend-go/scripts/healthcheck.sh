#!/usr/bin/env bash
# ============================================================================
# 健康检查脚本
#
# 校验以下三项，全部满足才返回 0：
#   1. /health 返回 HTTP 2xx
#   2. 响应 JSON 中 code == 0
#   3. data.status == "ok"
#   4. data.database == "connected"
#
# 用法：
#   ./scripts/healthcheck.sh                # 默认 127.0.0.1:8000
#   HOST=localhost PORT=8080 ./scripts/healthcheck.sh
#
# 退出码：
#   0 = 健康；1 = 不健康
#
# 用于：
#   - Makefile smoke 目标
#   - CI 环境探测
#   - 运维监控脚本
# ============================================================================
set -euo pipefail
HOST="${HOST:-127.0.0.1}"
PORT="${PORT:-8000}"
TIMEOUT="${TIMEOUT:-5}"
URL="http://${HOST}:${PORT}/health"
RESPONSE=$(curl -fsS --max-time "$TIMEOUT" "$URL" 2>/dev/null) || {
    echo "UNHEALTHY: 无法访问 $URL" >&2
    exit 1
}
if command -v jq >/dev/null 2>&1; then
    CODE=$(echo "$RESPONSE" | jq -r '.code // empty')
    STATUS=$(echo "$RESPONSE" | jq -r '.data.status // empty')
    DB=$(echo "$RESPONSE" | jq -r '.data.database // empty')
    if [ "$CODE" != "0" ] || [ "$STATUS" != "ok" ] || [ "$DB" != "connected" ]; then
        echo "UNHEALTHY: code=$CODE status=$STATUS database=$DB" >&2
        exit 1
    fi
else
    # 无 jq 时降级为字符串匹配
    if ! echo "$RESPONSE" | grep -q '"code":0' \
       || ! echo "$RESPONSE" | grep -q '"status":"ok"' \
       || ! echo "$RESPONSE" | grep -q '"database":"connected"'; then
        echo "UNHEALTHY: 响应内容异常：$RESPONSE" >&2
        exit 1
    fi
fi
echo "HEALTHY: $URL"
exit 0
