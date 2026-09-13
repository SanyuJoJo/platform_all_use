#!/usr/bin/env bash
# ============================================================
# 日志审计模块 API 测试脚本（v1.1）
# 用法：./scripts/test_audit_log.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（uv run alembic upgrade head）
#   2. 服务已启动
#
# 依赖：curl、jq、（可选）sqlite3
#
# v1.1 变更：
#   - P0-2：增加 CSV 数据行转义验证（含逗号 / 引号 / 换行）。
#   - 增加 P1-1 验证：单段路径 /api/v1/modules 的 resource 为 "modules"。
#   - 增加 P1-5 验证：/api/v1/audit-logs/export 的 action 为 "export"。
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
echo -e "${YELLOW}日志审计模块 API 测试（v1.1）${NC}"
echo "BASE_URL = $BASE_URL"
# ---------- 登录 ----------
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
# ---------- 列表 ----------
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
http_request GET "$API/audit-logs?status=fail" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按状态筛选成功"
http_request GET "$API/audit-logs?keyword=login" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按关键字筛选成功"
# ---------- P1-1 验证：单段路径 resource 解析 ----------
section "2. P1-1：单段路径 resource 解析"
http_request GET "$API/audit-logs?module_id=module_manager&action=view" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "module_manager 日志查询成功"
RESOURCE_MODULE=$(echo "$HTTP_BODY" | jq -r '.data.items[]? | select(.module_id=="module_manager") | .resource' | head -1)
if [ "$RESOURCE_MODULE" = "modules" ]; then
    pass "P1-1：/api/v1/modules 的 resource=$RESOURCE_MODULE（期望 modules）"
elif [ "$RESOURCE_MODULE" = "null" ] || [ -z "$RESOURCE_MODULE" ]; then
    # 可能还没有 module_manager 的日志，跳过
    echo -e "  ${YELLOW}⚠${NC}  未找到 module_manager 的日志，跳过 P1-1 验证"
else
    fail "P1-1：/api/v1/modules 的 resource=$RESOURCE_MODULE（期望 modules）"
fi
# ---------- P1-5 验证：/export 的 action ----------
section "3. P1-5：/export 的 action 语义"
http_request GET "$API/audit-logs/export?format=csv" \
    -H "Authorization: Bearer $TOKEN" >/dev/null
http_request GET "$API/audit-logs?module_id=audit_log&action=export" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按 action=export 查询成功"
EXPORT_ACTION_COUNT=$(echo "$HTTP_BODY" | jq -r '.data.total // 0')
if [ "$EXPORT_ACTION_COUNT" -ge 1 ]; then
    pass "P1-5：/export 被记录为 action=export（共 $EXPORT_ACTION_COUNT 条）"
else
    fail "P1-5：/export 未被记录为 action=export（total=$EXPORT_ACTION_COUNT）"
fi
# ---------- 时间范围非法 ----------
section "4. 非法时间范围"
http_request GET "$API/audit-logs?start_time=2026-09-11T23:59:59&end_time=2026-09-01T00:00:00" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "start_time > end_time → 400"
assert_code 40001 "code=40001"
# ---------- 详情 ----------
section "5. 查询日志详情"
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
# ---------- 导出 ----------
section "6. 导出审计日志"
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
http_request GET "$API/audit-logs/export?format=xml" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "不支持的格式 → 400"
assert_code 40004 "code=40004"
# ---------- P0-2 验证：CSV 数据行转义 ----------
section "7. P0-2：CSV 数据行转义"
# 通过 SQL 直接插入含特殊字符的日志（若 sqlite3 可用）
if command -v sqlite3 >/dev/null 2>&1 && [ -f app.db ]; then
    SQLITE_OUT=$(sqlite3 app.db <<'SQL' 2>&1
INSERT INTO audit_log_operation
    (user_id, username, module_id, action, resource, resource_id,
     detail, ip, user_agent, status, error_code, request_id, created_at)
VALUES
    (NULL, 'admin', 'test_p0', 'create', 'demo', NULL,
     '包含,逗号 "引号" 和换行', '127.0.0.1', 'pytest', 'success', NULL,
     'p0-2-marker', CURRENT_TIMESTAMP);
SQL
    )
    if [ -z "$SQLITE_OUT" ]; then
        CSV_FILE=$(mktemp)
        curl -sS "$API/audit-logs/export?format=csv" \
            -H "Authorization: Bearer $TOKEN" -o "$CSV_FILE"
        if grep -q '包含,逗号' "$CSV_FILE"; then
            pass "CSV 包含含逗号字段"
        else
            fail "CSV 未包含含逗号字段"
        fi
        if grep -q '""引号""' "$CSV_FILE"; then
            pass "CSV 引号被正确转义为双引号"
        else
            fail "CSV 引号未被正确转义"
        fi
        if grep -q '"""' "$CSV_FILE"; then
            fail "CSV 出现三重引号（P0-1 双重转义残留）"
        else
            pass "CSV 未出现三重引号"
        fi
        rm -f "$CSV_FILE"
        # 清理
        sqlite3 app.db "DELETE FROM audit_log_operation WHERE request_id='p0-2-marker';" 2>/dev/null || true
    else
        echo -e "  ${YELLOW}⚠${NC}  sqlite3 插入失败：$SQLITE_OUT"
        echo -e "  ${YELLOW}⚠${NC}  跳过 P0-2 数据行转义验证"
    fi
else
    echo -e "  ${YELLOW}⚠${NC}  未找到 sqlite3 或 app.db，跳过 P0-2 数据行转义验证"
fi
# ---------- 权限校验 ----------
section "8. 权限校验"
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
# ---------- 汇总 ----------
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
