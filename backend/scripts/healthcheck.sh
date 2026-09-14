#!/usr/bin/env bash
# ============================================================================
# 健康检查脚本
#
# v1.1 状态：保持 v1.0 逻辑（评审认为正确）。
#
# 用法：
#   ./scripts/healthcheck.sh              # 默认 127.0.0.1:8000
#   HOST=localhost PORT=8080 ./scripts/healthcheck.sh
#
# 退出码：
#   0 = 健康
#   1 = 不健康
#
# 用于：
#   - Docker HEALTHCHECK
#   - Kubernetes liveness/readiness probe
#   - 运维监控
#   - start.sh 启动就绪轮询（v1.1 起）
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
# 校验 JSON（无 jq 时降级为字符串匹配）
if command -v jq >/dev/null 2>&1; then
    STATUS=$(echo "$RESPONSE" | jq -r '.data.status // empty')
    DB=$(echo "$RESPONSE" | jq -r '.data.database // empty')
    CODE=$(echo "$RESPONSE" | jq -r '.code // empty')
    if [ "$CODE" != "0" ] || [ "$STATUS" != "ok" ] || [ "$DB" != "connected" ]; then
        echo "UNHEALTHY: code=$CODE status=$STATUS database=$DB" >&2
        exit 1
    fi
else
    if ! echo "$RESPONSE" | grep -q '"status":"ok"' \
       || ! echo "$RESPONSE" | grep -q '"database":"connected"'; then
        echo "UNHEALTHY: 响应内容异常：$RESPONSE" >&2
        exit 1
    fi
fi
echo "HEALTHY: $URL"
exit 0
