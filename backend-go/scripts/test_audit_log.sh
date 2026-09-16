#!/usr/bin/env bash
# ============================================================
# 日志审计模块 API 测试脚本（Go 版 v1.2）
# 用法：./scripts/test_audit_log.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（make migrate-up）
#   2. 服务已启动（make run）
#
# 依赖：curl、jq、（可选）sqlite3
# ============================================================
set -uo pipefail
BASE_URL="${1:-http://localhost:8000}"
API="${BASE_URL}/api/v1"
for cmd in curl jq; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "❌ 需要 $cmd，请先安装"
        exit 1
    fi
done
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
PASS=0; FAIL=0
pass() { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS+1)); }
fail() { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
section() { echo -e "\n${CYAN}=== $1 ===${NC}"; }
http_request() {
    local method="$1"; shift
    local url="$1"; shift
    local resp
    resp=$(curl -sS -X "$method" "$url" -w "\n__HTTP_CODE__:%{http_code}" "$@" 2>/dev/null)
    HTTP_CODE=$(echo "$resp" | sed -n 's/^__HTTP_CODE__://p')
    HTTP_BODY=$(echo "$resp" | sed '/^__HTTP_CODE__:/d')
}
assert_http() {
    if [ "$HTTP_CODE" = "$1" ]; then pass "$2 (HTTP $HTTP_CODE)"; else fail "$2 (expected HTTP $1, got $HTTP_CODE)"; fi
}
assert_code() {
    local actual; actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$1" ]; then pass "$2 (code=$actual)"; else fail "$2 (expected code=$1, got $actual)"; fi
}
echo -e "${YELLOW}日志审计模块 API 测试（Go 版 v1.2）${NC}"
echo "BASE_URL = $BASE_URL"
section "0. 登录获取 Token"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ admin 登录失败，无法继续${NC}"
    exit 1
fi
section "1. 查询审计日志列表"
http_request GET "$API/audit-logs?page=1&page_size=10" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "GET /audit-logs 返回 200"
assert_code 0 "code=0"
TOTAL=$(echo "$HTTP_BODY" | jq -r '.data.total')
pass "当前日志总数：$TOTAL"
http_request GET "$API/audit-logs?module_id=auth" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按模块筛选成功"
http_request GET "$API/audit-logs?action=view" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按操作类型筛选成功"
http_request GET "$API/audit-logs?keyword=login" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按关键字筛选成功"
section "2. 分页参数校验（v1.2 P2-NEW-1：?page=0 返回 422）"
http_request GET "$API/audit-logs?page=0" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "page=0 → 422（与 Python 版契约一致）"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type // "null"')
if [ "$ERR_TYPE" = "min" ]; then
    pass "errors[0].type=min"
else
    fail "expected errors[0].type=min, got $ERR_TYPE"
fi
LOC1=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].loc[1] // "null"')
if [ "$LOC1" = "page" ]; then
    pass "loc[1]=page（form tag 字段名）"
else
    fail "expected loc[1]=page, got $LOC1"
fi
http_request GET "$API/audit-logs?page_size=0" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "page_size=0 → 422（与 Python 版契约一致）"
assert_code 90004 "code=90004"
http_request GET "$API/audit-logs?page=-1" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "page=-1 → 422"
assert_code 90004 "code=90004"
http_request GET "$API/audit-logs?page_size=200" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "page_size=200 → 422"
assert_code 90004 "code=90004"
section "3. query 类型错误"
http_request GET "$API/audit-logs?page=abc" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "page=abc → 422"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type // "null"')
if [ "$ERR_TYPE" = "type_error" ]; then
    pass "errors[0].type=type_error"
else
    fail "expected errors[0].type=type_error, got $ERR_TYPE"
fi
section "4. 时间格式错误"
http_request GET "$API/audit-logs?start_time=invalid" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "start_time=invalid → 422"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type // "null"')
if [ "$ERR_TYPE" = "datetime_parsing" ]; then
    pass "errors[0].type=datetime_parsing"
else
    fail "expected errors[0].type=datetime_parsing, got $ERR_TYPE"
fi
section "5. 时间范围非法"
http_request GET "$API/audit-logs?start_time=2026-09-11T23:59:59&end_time=2026-09-01T00:00:00" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "start_time > end_time → 400"
assert_code 40001 "code=40001"
section "6. 查询日志详情"
LOG_ID=$(curl -sS "$API/audit-logs?page=1&page_size=1" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.data.items[0].id')
if [ -n "$LOG_ID" ] && [ "$LOG_ID" != "null" ]; then
    http_request GET "$API/audit-logs/$LOG_ID" \
        -H "Authorization: Bearer $TOKEN"
    assert_http 200 "详情请求成功"
    assert_code 0 "code=0"
else
    fail "未找到日志，无法测试详情"
fi
http_request GET "$API/audit-logs/99999999" \
    -H "Authorization: Bearer $TOKEN"
assert_http 404 "不存在的日志 → 404"
assert_code 40002 "code=40002"
section "7. 路径参数校验"
http_request GET "$API/audit-logs/abc" \
    -H "Authorization: Bearer $TOKEN"
assert_http 422 "log_id=abc → 422"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type // "null"')
if [ "$ERR_TYPE" = "type_error" ]; then
    pass "errors[0].type=type_error"
else
    fail "expected errors[0].type=type_error, got $ERR_TYPE"
fi
LOC0=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].loc[0] // "null"')
if [ "$LOC0" = "path" ]; then
    pass "loc[0]=path"
else
    fail "expected loc[0]=path, got $LOC0"
fi
section "8. 导出审计日志"
http_request GET "$API/audit-logs/export?format=csv" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "导出 CSV 成功"
if echo "$HTTP_BODY" | head -1 | grep -q "id,user_id"; then
    pass "CSV 包含表头"
else
    fail "CSV 缺少表头"
fi
http_request GET "$API/audit-logs/export?format=json" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "导出 JSON 成功"
JSON_TOTAL=$(echo "$HTTP_BODY" | jq -r '.total // "null"')
if [ "$JSON_TOTAL" != "null" ]; then
    pass "JSON 包含 total=$JSON_TOTAL"
else
    fail "JSON 缺少 total"
fi
section "9. 导出格式大小写不敏感"
http_request GET "$API/audit-logs/export?format=CSV" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "format=CSV 被接受（大小写不敏感）"
http_request GET "$API/audit-logs/export?format=xml" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "不支持的格式 → 400"
assert_code 40004 "code=40004"
section "10. 权限校验"
http_request GET "$API/audit-logs"
assert_http 401 "未认证 → 401"
assert_code 10001 "code=10001"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -n "$GUEST_TOKEN" ] && [ "$GUEST_TOKEN" != "null" ]; then
    http_request GET "$API/audit-logs" \
        -H "Authorization: Bearer $GUEST_TOKEN"
    assert_http 403 "guest 查询日志被拒绝"
    assert_code 20051 "code=20051"
    http_request GET "$API/audit-logs/export" \
        -H "Authorization: Bearer $GUEST_TOKEN"
    assert_http 403 "guest 导出日志被拒绝"
    assert_code 20051 "code=20051"
else
    fail "guest 登录失败"
fi
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS+FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"
if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 日志审计模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
