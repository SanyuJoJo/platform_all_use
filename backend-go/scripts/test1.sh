#!/usr/bin/env bash
# ============================================================================
# 后端服务功能验证脚本（v1.2）
#
# 覆盖内容：
#   0. 服务连通性
#   1. /health 契约（与 Python 版一致）
#   2. timestamp 格式（Python isoformat 兼容）
#   3. 404 未注册路径
#   4. 405 方法不允许
#   5. X-Request-ID 透传与生成
#   6. CORS（预检 + 实际响应头）
#   7. Content-Type
#   8. /health 稳定性
#   9. 静态二进制（可选）
#  10. 统一响应字段完整性
#
# 用法：
#   ./scripts/test1.sh                               # 默认 127.0.0.1:8000
#   ./scripts/test1.sh 192.168.1.100                 # 指定 IP，端口默认 8000
#   ./scripts/test1.sh 192.168.1.100 18000           # 指定 IP 和端口
#   ./scripts/test1.sh http://192.168.1.100:18000    # 完整 URL
#   ./scripts/test1.sh --help                        # 查看帮助
#
# 环境变量（命令行参数优先）：
#   HOST=127.0.0.1
#   PORT=8000
#   TIMEOUT=5
#
# 依赖：
#   - curl（必须）
#   - jq（建议；缺失时自动降级为字符串匹配）
#
# 退出码：
#   0 = 全部通过
#   1 = 存在失败
# ============================================================================
set -uo pipefail

# ---------- 帮助 ----------
print_help() {
    cat << 'HELP'
用法：
  ./scripts/test1.sh [<host|url>] [<port>]

参数：
  <host|url>   主机名 / IP / 完整 URL（如 http://1.2.3.4:8000）
               省略时使用环境变量 HOST 或默认 127.0.0.1
  <port>       端口号（数字），省略时使用环境变量 PORT 或默认 8000

示例：
  ./scripts/test1.sh
  ./scripts/test1.sh 192.168.1.100
  ./scripts/test1.sh 192.168.1.100 18000
  ./scripts/test1.sh http://192.168.1.100:18000

环境变量：
  HOST       默认主机（命令行未指定时使用）
  PORT       默认端口（命令行未指定时使用）
  TIMEOUT    单次 HTTP 超时秒数（默认 5）

退出码：
  0 全部通过；1 存在失败；2 参数错误或依赖缺失
HELP
}

# ---------- 解析参数 ----------
DEFAULT_HOST="${HOST:-127.0.0.1}"
DEFAULT_PORT="${PORT:-8000}"
TIMEOUT="${TIMEOUT:-5}"

ARG_HOST=""
ARG_PORT=""

case "${1:-}" in
    -h|--help)
        print_help
        exit 0
        ;;
esac

if [ $# -ge 1 ]; then
    ARG_HOST="$1"
fi
if [ $# -ge 2 ]; then
    ARG_PORT="$2"
fi
if [ $# -gt 2 ]; then
    echo "❌ 参数过多。使用 --help 查看用法" >&2
    exit 2
fi

# 若第一个参数是完整 URL，直接使用；否则拼接
if [ -n "$ARG_HOST" ] && echo "$ARG_HOST" | grep -qE '^https?://'; then
    BASE_URL="${ARG_HOST%/}"
elif [ -n "$ARG_HOST" ]; then
    P="${ARG_PORT:-$DEFAULT_PORT}"
    BASE_URL="http://${ARG_HOST}:${P}"
else
    BASE_URL="http://${DEFAULT_HOST}:${DEFAULT_PORT}"
fi

# ---------- 颜色 ----------
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; CYAN=''; NC=''
fi

PASS=0; FAIL=0
pass()    { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS+1)); }
fail()    { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
info()    { echo -e "  ${YELLOW}ℹ${NC}  $1"; }
section() { echo -e "\n${CYAN}=== $1 ===${NC}"; }

# ---------- 依赖检查 ----------
if ! command -v curl >/dev/null 2>&1; then
    echo "❌ 需要 curl，请先安装"
    exit 2
fi
HAS_JQ=0
if command -v jq >/dev/null 2>&1; then
    HAS_JQ=1
fi

# ---------- 临时文件 ----------
HEADERS_TMP=$(mktemp -t smoke_headers.XXXXXX)
trap 'rm -f "$HEADERS_TMP"' EXIT

# ---------- HTTP 请求辅助 ----------
# 用法：http_request METHOD URL [curl 额外参数...]
# 结果：HTTP_CODE、HTTP_BODY、HTTP_HEADERS
http_request() {
    local method="$1"; shift
    local url="$1"; shift
    local resp
    resp=$(curl -sS --max-time "$TIMEOUT" -X "$method" "$url" \
        -D "$HEADERS_TMP" \
        -w "\n__HTTP_CODE__:%{http_code}" \
        "$@" 2>/dev/null) || true
    HTTP_CODE=$(echo "$resp" | sed -n 's/^__HTTP_CODE__://p')
    HTTP_BODY=$(echo "$resp" | sed '/^__HTTP_CODE__:/d')
    HTTP_HEADERS=$(cat "$HEADERS_TMP" 2>/dev/null || echo "")
}

# ---------- 断言辅助 ----------
assert_http() {
    local expected="$1"; local msg="$2"
    if [ "$HTTP_CODE" = "$expected" ]; then
        pass "$msg (HTTP $HTTP_CODE)"
    else
        fail "$msg (期望 HTTP $expected，实际 $HTTP_CODE，body=$HTTP_BODY)"
    fi
}

assert_json_field() {
    local field="$1"; local expected="$2"; local msg="$3"
    local actual=""
    if [ "$HAS_JQ" -eq 1 ]; then
        actual=$(echo "$HTTP_BODY" | jq -r "$field // \"__MISSING__\"" 2>/dev/null || echo "__PARSE_ERROR__")
    else
        actual=$(echo "$HTTP_BODY" | grep -oE "\"$(basename "$field")\"\s*:\s*\"?[^,\"}]*\"?" | head -1 | sed -E 's/.*:\s*"?([^,"}]*)"?/\1/')
    fi
    if [ "$actual" = "$expected" ]; then
        pass "$msg ($field=$actual)"
    else
        fail "$msg (期望 $field=$expected，实际 $actual，body=$HTTP_BODY)"
    fi
}

assert_json_field_exists() {
    local field="$1"; local msg="$2"
    local actual=""
    if [ "$HAS_JQ" -eq 1 ]; then
        actual=$(echo "$HTTP_BODY" | jq -r "$field // \"__MISSING__\"" 2>/dev/null || echo "__PARSE_ERROR__")
        if [ "$actual" = "__MISSING__" ] || [ "$actual" = "__PARSE_ERROR__" ] || [ "$actual" = "null" ]; then
            fail "$msg (字段 $field 缺失)"
            return
        fi
    else
        if ! echo "$HTTP_BODY" | grep -q "\"$(basename "$field")\""; then
            fail "$msg (字段 $field 缺失)"
            return
        fi
    fi
    pass "$msg (字段 $field 存在)"
}

assert_header_contains() {
    local header_name="$1"; local pattern="$2"; local msg="$3"
    local line
    line=$(echo "$HTTP_HEADERS" | grep -i "^${header_name}:" | head -1 || true)
    if [ -z "$line" ]; then
        fail "$msg (响应头 $header_name 缺失)"
        return
    fi
    if echo "$line" | grep -qi "$pattern"; then
        pass "$msg"
    else
        fail "$msg (响应头 $header_name 未包含 $pattern，实际：$line)"
    fi
}

assert_header_exists() {
    local header_name="$1"; local msg="$2"
    if echo "$HTTP_HEADERS" | grep -qi "^${header_name}:"; then
        pass "$msg"
    else
        fail "$msg (响应头 $header_name 缺失)"
    fi
}

assert_header_not_contains() {
    local header_name="$1"; local pattern="$2"; local msg="$3"
    local line
    line=$(echo "$HTTP_HEADERS" | grep -i "^${header_name}:" | head -1 || true)
    if [ -z "$line" ]; then
        pass "$msg (响应头 $header_name 不存在，符合预期)"
        return
    fi
    if echo "$line" | grep -qi "$pattern"; then
        fail "$msg (响应头 $header_name 不应包含 $pattern，实际：$line)"
    else
        pass "$msg"
    fi
}

# ============================================================================
# 开始测试
# ============================================================================
echo "=============================================="
echo "  后端服务功能验证（v1.2）"
echo "  BASE_URL = $BASE_URL"
echo "  TIMEOUT  = ${TIMEOUT}s"
echo "  jq 支持  = $([ "$HAS_JQ" -eq 1 ] && echo "是" || echo "否（降级模式）")"
echo "=============================================="

# ---------------------------------------------------------------------------
# 0. 服务连通性
# ---------------------------------------------------------------------------
section "0. 服务连通性"
if ! curl -fsS --max-time 3 "$BASE_URL/health" >/dev/null 2>&1; then
    echo -e "${RED}❌ 无法访问 $BASE_URL/health，请确认服务已启动${NC}"
    echo "   排查建议："
    echo "     1. 确认服务已启动：make run"
    echo "     2. 确认监听端口：curl -v $BASE_URL/health"
    echo "     3. 若服务监听 18000：./scripts/test1.sh 127.0.0.1 18000"
    echo "     4. 若是远程主机：./scripts/test1.sh <ip> <port>"
    exit 1
fi
pass "服务可访问：$BASE_URL"

# ---------------------------------------------------------------------------
# 1. /health 契约（与 Python 版一致）
# ---------------------------------------------------------------------------
section "1. /health 契约（与 Python 版一致）"

http_request GET "$BASE_URL/health"
assert_http 200 "GET /health 返回 200"

if [ "$HAS_JQ" -eq 1 ]; then
    assert_json_field_exists ".code"      "含 code 字段"
    assert_json_field_exists ".message"   "含 message 字段"
    assert_json_field_exists ".data"      "含 data 字段"
    assert_json_field_exists ".timestamp" "含 timestamp 字段"
    assert_json_field_exists ".requestId" "含 requestId 字段"

    assert_json_field ".code"          "0"         "code=0"
    assert_json_field ".message"       "success"   "message=success"
    assert_json_field ".data.status"   "ok"        "data.status=ok"
    assert_json_field ".data.database" "connected" "data.database=connected"

    DATA_KEYS=$(echo "$HTTP_BODY" | jq -r '.data | keys | join(",")')
    if [ "$DATA_KEYS" = "database,status" ]; then
        pass "data 仅含 status/database"
    else
        fail "data 字段不符合预期：$DATA_KEYS（期望 database,status）"
    fi

    TOP_KEYS=$(echo "$HTTP_BODY" | jq -r 'keys | sort | join(",")')
    if [ "$TOP_KEYS" = "code,data,message,requestId,timestamp" ]; then
        pass "顶层仅含 code/data/message/requestId/timestamp"
    else
        fail "顶层字段不符合预期：$TOP_KEYS"
    fi
else
    for pat in '"code":0' '"message":"success"' '"status":"ok"' '"database":"connected"' '"requestId"'; do
        if echo "$HTTP_BODY" | grep -q "$pat"; then
            pass "包含 $pat"
        else
            fail "缺少 $pat"
        fi
    done
fi

# ---------------------------------------------------------------------------
# 2. timestamp 格式（Python isoformat 兼容）
# ---------------------------------------------------------------------------
section "2. timestamp 格式（Python isoformat 兼容）"

TS=$(echo "$HTTP_BODY" | sed -n 's/.*"timestamp":"\([^"]*\)".*/\1/p')
info "收到 timestamp = $TS"

if echo "$TS" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{6})?\+00:00$'; then
    pass "timestamp 格式正确（ISO 8601 UTC）"
else
    fail "timestamp 格式不正确：$TS"
fi

# ---------------------------------------------------------------------------
# 3. 404 未注册路径
# ---------------------------------------------------------------------------
section "3. 404 未注册路径"

http_request GET "$BASE_URL/not-exist-path-xyz"
assert_http 404 "GET /not-exist-path-xyz 返回 404"
assert_json_field ".code"    "90002"      "code=90002"
assert_json_field ".message" "资源不存在" "message=资源不存在"

http_request GET "$BASE_URL/api/v1/not-exist"
assert_http 404 "GET /api/v1/not-exist 返回 404"
assert_json_field ".code" "90002" "code=90002"

# ---------------------------------------------------------------------------
# 4. 405 方法不允许
# ---------------------------------------------------------------------------
section "4. 405 方法不允许"

http_request POST "$BASE_URL/health"
assert_http 405 "POST /health 返回 405"
assert_json_field ".code"    "90001"    "code=90001"
assert_json_field ".message" "请求错误" "message=请求错误"

http_request DELETE "$BASE_URL/health"
assert_http 405 "DELETE /health 返回 405"
assert_json_field ".code" "90001" "code=90001"

http_request PUT "$BASE_URL/health"
assert_http 405 "PUT /health 返回 405"
assert_json_field ".code" "90001" "code=90001"

# ---------------------------------------------------------------------------
# 5. X-Request-ID 透传与生成
# ---------------------------------------------------------------------------
section "5. X-Request-ID 透传与生成"

# 5.1 透传
CUSTOM_ID="smoke-test-$(date +%s)-abcd"
http_request GET "$BASE_URL/health" -H "X-Request-ID: $CUSTOM_ID"
assert_header_contains "X-Request-ID" "$CUSTOM_ID" "响应头回填 X-Request-ID"
assert_json_field ".requestId" "$CUSTOM_ID" "响应体 requestId 与请求头一致"

# 5.2 生成
http_request GET "$BASE_URL/health"
GEN_ID=$(echo "$HTTP_HEADERS" | grep -i "^X-Request-ID:" | head -1 | sed -E 's/^[^:]+:\s*//' | tr -d '\r')
if [ -n "$GEN_ID" ]; then
    pass "未携带时自动生成 X-Request-ID：$GEN_ID"
    if echo "$GEN_ID" | grep -qE '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$'; then
        pass "自动生成的 ID 为合法 UUID"
    else
        info "自动生成的 ID 不是标准 UUID 格式（可能为自定义，非阻塞）：$GEN_ID"
    fi
else
    fail "未携带 X-Request-ID 时未生成响应头"
fi

# 5.3 响应头/体一致性
HEADER_RID=$(echo "$HTTP_HEADERS" | grep -i "^X-Request-ID:" | head -1 | sed -E 's/^[^:]+:\s*//' | tr -d '\r')
BODY_RID=$(echo "$HTTP_BODY" | sed -n 's/.*"requestId":"\([^"]*\)".*/\1/p')
if [ -n "$HEADER_RID" ] && [ "$HEADER_RID" = "$BODY_RID" ]; then
    pass "响应头 X-Request-ID 与响应体 requestId 一致"
else
    fail "响应头 X-Request-ID ($HEADER_RID) 与响应体 requestId ($BODY_RID) 不一致"
fi

# ---------------------------------------------------------------------------
# 6. CORS（预检 + 实际响应头）
# ---------------------------------------------------------------------------
section "6. CORS 预检与实际响应头"

# 6.1 OPTIONS 预检
http_request OPTIONS "$BASE_URL/health" \
    -H "Origin: http://example.com" \
    -H "Access-Control-Request-Method: GET" \
    -H "Access-Control-Request-Headers: content-type,authorization"

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    pass "OPTIONS 预检返回 $HTTP_CODE"
else
    fail "OPTIONS 预检返回异常：$HTTP_CODE"
fi

assert_header_contains "Access-Control-Allow-Origin" "\*" "预检含 Allow-Origin: *"
assert_header_exists   "Access-Control-Allow-Methods"      "预检含 Allow-Methods"
assert_header_exists   "Access-Control-Allow-Headers"      "预检含 Allow-Headers"

# V11-P1-03 修复验证：Allow-Headers 不应含 X-Request-ID
assert_header_not_contains "Access-Control-Allow-Headers" "X-Request-ID" \
    "Allow-Headers 未包含 X-Request-ID（符合 V11-P1-03）"

# 6.2 实际请求（Expose-Headers 属于实际响应头，规范上不出现在预检响应中）
http_request GET "$BASE_URL/health" -H "Origin: http://example.com"
assert_header_exists "Access-Control-Expose-Headers" "实际响应含 Expose-Headers"

if echo "$HTTP_HEADERS" | grep -i "^Access-Control-Expose-Headers:" | grep -qi "X-Request-ID"; then
    pass "Expose-Headers 含 X-Request-ID"
else
    fail "Expose-Headers 缺少 X-Request-ID"
fi

if echo "$HTTP_HEADERS" | grep -i "^Access-Control-Expose-Headers:" | grep -qi "X-Error-Code"; then
    pass "Expose-Headers 含 X-Error-Code"
else
    fail "Expose-Headers 缺少 X-Error-Code"
fi

# ---------------------------------------------------------------------------
# 7. Content-Type
# ---------------------------------------------------------------------------
section "7. Content-Type"

http_request GET "$BASE_URL/health"
assert_header_contains "Content-Type" "application/json" "/health 返回 application/json"

http_request GET "$BASE_URL/not-exist"
assert_header_contains "Content-Type" "application/json" "404 返回 application/json"

# ---------------------------------------------------------------------------
# 8. /health 稳定性（连续 5 次）
# ---------------------------------------------------------------------------
section "8. /health 稳定性（连续 5 次）"

HEALTH_OK=1
for i in 1 2 3 4 5; do
    CODE=$(curl -fsS --max-time "$TIMEOUT" "$BASE_URL/health" 2>/dev/null \
        | sed -n 's/.*"code":\([0-9]*\).*/\1/p')
    if [ "$CODE" != "0" ]; then
        HEALTH_OK=0
        break
    fi
done
if [ "$HEALTH_OK" -eq 1 ]; then
    pass "连续 5 次 /health 均返回 code=0"
else
    fail "连续 5 次 /health 中存在非 0 响应"
fi

# ---------------------------------------------------------------------------
# 9. 静态二进制（可选，仅当 bin/server 存在时验证）
# ---------------------------------------------------------------------------
section "9. 静态二进制（可选）"

if [ -x "./bin/server" ]; then
    if command -v file >/dev/null 2>&1; then
        FILE_OUT=$(file ./bin/server)
        info "file 输出：$FILE_OUT"
        if echo "$FILE_OUT" | grep -qi "statically linked"; then
            pass "bin/server 为静态链接二进制"
        else
            info "file 未明确标注 statically linked（Go 静态编译可能不带该标记）"
        fi
    else
        info "未安装 file 命令，跳过静态链接检查"
    fi

    if command -v ldd >/dev/null 2>&1; then
        LDD_OUT=$(ldd ./bin/server 2>&1 || true)
        if echo "$LDD_OUT" | grep -qi "not a dynamic executable" || echo "$LDD_OUT" | grep -qi "statically linked"; then
            pass "ldd 确认无动态依赖"
        else
            info "ldd 输出：$LDD_OUT"
        fi
    fi
else
    info "未找到 bin/server，跳过（可执行 make build 后再运行本脚本）"
fi

# ---------------------------------------------------------------------------
# 10. 统一响应字段完整性
# ---------------------------------------------------------------------------
section "10. 统一响应字段完整性"

check_response_shape() {
    local label="$1"
    if [ "$HAS_JQ" -eq 1 ]; then
        local keys
        keys=$(echo "$HTTP_BODY" | jq -r 'keys | sort | join(",")' 2>/dev/null || echo "")
        if [ "$keys" = "code,data,message,requestId,timestamp" ]; then
            pass "$label 响应字段完整（5 个）"
        else
            fail "$label 响应字段不完整：$keys"
        fi
    else
        local ok=1
        for f in code message data timestamp requestId; do
            echo "$HTTP_BODY" | grep -q "\"$f\"" || ok=0
        done
        if [ "$ok" -eq 1 ]; then
            pass "$label 响应字段完整（5 个）"
        else
            fail "$label 响应字段不完整"
        fi
    fi
}

http_request GET "$BASE_URL/health"
check_response_shape "200 /health"

http_request GET "$BASE_URL/not-exist"
check_response_shape "404 /not-exist"

http_request POST "$BASE_URL/health"
check_response_shape "405 /health"

# ---------------------------------------------------------------------------
# 汇总
# ---------------------------------------------------------------------------
TOTAL=$((PASS+FAIL))
echo ""
echo "=============================================="
echo "  测试结果汇总"
echo "=============================================="
echo "  BASE_URL = $BASE_URL"
echo -e "  总计：$TOTAL"
echo -e "  通过：${GREEN}${PASS}${NC}"
echo -e "  失败：${RED}${FAIL}${NC}"
echo "=============================================="

if [ "$FAIL" -eq 0 ]; then
    echo -e "\n${GREEN}🎉 所有测试通过，服务功能符合 v1.2 交付标准${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
