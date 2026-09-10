#!/usr/bin/env bash
# ============================================================
# 权限管理模块 API 测试脚本（v1.0，v1.1 / v1.2 未修改）
# 用法：./scripts/test_permission.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移
#   2. 服务已启动
#
# 依赖：curl、jq
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
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'
PASS=0
FAIL=0
pass() { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS + 1)); }
fail() { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL + 1)); }
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
    if [ "$HTTP_CODE" = "$1" ]; then
        pass "$2 (HTTP $HTTP_CODE)"
    else
        fail "$2 (expected HTTP $1, got $HTTP_CODE)"
    fi
}
assert_code() {
    local actual
    actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$1" ]; then
        pass "$2 (code=$actual)"
    else
        fail "$2 (expected code=$1, got $actual)"
    fi
}
echo -e "${YELLOW}权限管理模块 API 测试（v1.0）${NC}"
echo "BASE_URL = $BASE_URL"
# ---------- 登录 ----------
section "0. 登录获取 Token"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
assert_code 0 "登录 code=0"
TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ admin 登录失败，无法继续测试${NC}"
    exit 1
fi
# ---------- 查询权限列表 ----------
section "1. 查询权限列表"
http_request GET "$API/auth/permissions" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "GET /auth/permissions 返回 200"
assert_code 0 "code=0"
COUNT=$(echo "$HTTP_BODY" | jq '.data | length')
pass "权限总数：$COUNT"
# ---------- 按模块筛选 ----------
section "2. 按模块筛选权限"
http_request GET "$API/auth/permissions?module_id=auth" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按 module_id=auth 查询成功"
assert_code 0 "code=0"
AUTH_COUNT=$(echo "$HTTP_BODY" | jq '.data | length')
pass "auth 模块权限数：$AUTH_COUNT"
# ---------- 按资源筛选 ----------
section "3. 按资源筛选权限"
http_request GET "$API/auth/permissions?module_id=auth&resource=user" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "按 resource=user 查询成功"
assert_code 0 "code=0"
RES_COUNT=$(echo "$HTTP_BODY" | jq '.data | length')
pass "auth:user 权限数：$RES_COUNT"
# ---------- 非法 module_id（v1.2 新增） ----------
section "3.1 非法 module_id 查询（v1.2 新增）"
http_request GET "$API/auth/permissions?module_id=BadModule" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "非法 module_id → 400"
assert_code 90001 "code=90001"
# ---------- 模块权限 ----------
section "4. 获取模块全部权限"
http_request GET "$API/auth/permissions/modules/auth" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "GET /auth/permissions/modules/auth 成功"
assert_code 0 "code=0"
MOD_COUNT=$(echo "$HTTP_BODY" | jq '.data | length')
pass "auth 模块权限数：$MOD_COUNT"
# ---------- 未认证 ----------
section "5. 未认证访问"
http_request GET "$API/auth/permissions"
assert_http 401 "未携带 Token 返回 401"
assert_code 10001 "code=10001"
# ---------- 权限校验 ----------
section "6. 权限校验（guest）"
# 重置 guest 密码，避免被其他脚本污染
http_request GET "$API/auth/users?keyword=guest" \
    -H "Authorization: Bearer $TOKEN"
GUEST_ID=$(echo "$HTTP_BODY" | jq -r '.data.items[]? | select(.username=="guest") | .id' | head -1)
if [ -n "$GUEST_ID" ] && [ "$GUEST_ID" != "null" ]; then
    http_request PATCH "$API/auth/users/$GUEST_ID/password" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"new_password":"123456"}'
    if [ "$HTTP_CODE" = "200" ]; then
        pass "guest 密码已重置为 123456"
    else
        fail "guest 密码重置失败 (HTTP $HTTP_CODE)"
    fi
else
    fail "未找到 guest 用户"
fi
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token // empty')
if [ -n "$GUEST_TOKEN" ] && [ "$GUEST_TOKEN" != "null" ]; then
    pass "guest 登录成功"
    http_request GET "$API/auth/permissions" \
        -H "Authorization: Bearer $GUEST_TOKEN"
    assert_http 403 "guest 查询权限被拒绝"
    assert_code 20051 "code=20051"
else
    fail "guest 登录失败，无法继续权限校验"
fi
# ---------- 汇总 ----------
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS + FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"
if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 权限管理模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
