#!/usr/bin/env bash
# ============================================================
# 认证模块 API 测试脚本（Go 版 v1.1）
# 用法：./scripts/test_auth.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（make migrate-up）
#   2. 服务已启动（make run）
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
    if [ "$HTTP_CODE" = "$1" ]; then pass "$2 (HTTP $HTTP_CODE)"; else fail "$2 (expected HTTP $1, got $HTTP_CODE, body=$HTTP_BODY)"; fi
}
assert_code() {
    local actual
    actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$1" ]; then pass "$2 (code=$actual)"; else fail "$2 (expected code=$1, got $actual, body=$HTTP_BODY)"; fi
}
assert_not_null() {
    local value
    value=$(echo "$HTTP_BODY" | jq -r "$1 // \"null\"")
    if [ "$value" != "null" ] && [ -n "$value" ]; then pass "$2"; else fail "$2 (field $1 is null)"; fi
}
echo -e "${YELLOW}认证模块 API 测试（Go 版 v1.1）${NC}"
echo "BASE_URL = $BASE_URL"
# ============================================================
section "1. 健康检查"
# ============================================================
http_request GET "$BASE_URL/health"
assert_http 200 "GET /health 返回 200"
assert_code 0 "健康检查 code=0"
# ============================================================
section "2. 登录"
# ============================================================
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
assert_code 0 "登录 code=0"
assert_not_null ".data.access_token" "返回 access_token"
assert_not_null ".data.refresh_token" "返回 refresh_token"
assert_not_null ".data.user.permissions" "返回 user.permissions"
ADMIN_ACCESS_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
ADMIN_REFRESH_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.refresh_token')
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"wrong-password"}'
assert_http 401 "密码错误返回 401"
assert_code 10001 "密码错误 code=10001"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"x"}'
assert_http 401 "短密码返回 401"
assert_code 10001 "短密码 code=10001"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"no_such_user","password":"123456"}'
assert_http 401 "用户不存在返回 401"
assert_code 10001 "用户不存在 code=10001"
# v1.1 P1-01：参数校验失败返回 422/90004
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{}'
assert_http 422 "空登录请求体 → 422"
assert_code 90004 "code=90004"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin"}'
assert_http 422 "缺少 password → 422"
assert_code 90004 "code=90004"
# ============================================================
section "3. 当前用户信息 /auth/me（v1.1 P0-01 修复）"
# ============================================================
http_request GET "$API/auth/me" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
assert_http 200 "携带合法 Token 返回 200"
assert_code 0 "code=0"
assert_not_null ".data.username" "返回 username"
assert_not_null ".data.permissions" "返回 permissions"
http_request GET "$API/auth/me"
assert_http 401 "无 Token 返回 401"
assert_code 10001 "无 Token code=10001"
http_request GET "$API/auth/me" \
    -H "Authorization: Bearer invalid.token.value"
assert_http 401 "无效 Token 返回 401"
assert_code 10001 "无效 Token code=10001"
# ============================================================
section "4. 刷新 Token（轮换）"
# ============================================================
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$ADMIN_REFRESH_TOKEN\"}"
assert_http 200 "刷新成功"
assert_code 0 "code=0"
assert_not_null ".data.access_token" "返回新 access_token"
assert_not_null ".data.refresh_token" "返回新 refresh_token（轮换）"
NEW_REFRESH_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.refresh_token')
if [ "$NEW_REFRESH_TOKEN" != "$ADMIN_REFRESH_TOKEN" ]; then
    pass "refresh_token 已轮换（新旧不同）"
else
    fail "refresh_token 未轮换（新旧相同）"
fi
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$ADMIN_REFRESH_TOKEN\"}"
assert_http 401 "旧 refresh_token 复用被拒绝"
assert_code 10001 "旧 refresh_token code=10001"
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$NEW_REFRESH_TOKEN\"}"
assert_http 200 "新 refresh_token 可继续刷新"
assert_code 0 "code=0"
ADMIN_ACCESS_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
ADMIN_REFRESH_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.refresh_token')
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$ADMIN_ACCESS_TOKEN\"}"
assert_http 401 "使用 access_token 刷新被拒绝"
assert_code 10001 "code=10001"
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d 'not-a-json'
assert_http 422 "无效 JSON body → 422"
assert_code 90004 "code=90004"
# ============================================================
section "5. 登出（v1.1 P0-01 修复：需认证）"
# ============================================================
http_request POST "$API/auth/logout" \
    -H "Content-Type: application/json" \
    -d '{}'
assert_http 401 "未认证登出 → 401"
assert_code 10001 "code=10001"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
assert_http 200 "guest 登录成功"
GUEST_ACCESS_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
GUEST_REFRESH_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.refresh_token')
http_request POST "$API/auth/logout" \
    -H "Authorization: Bearer $GUEST_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$GUEST_REFRESH_TOKEN\"}"
assert_http 200 "登出成功"
assert_code 0 "code=0"
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$GUEST_REFRESH_TOKEN\"}"
assert_http 401 "登出后 refresh_token 失效"
assert_code 10001 "code=10001"
http_request POST "$API/auth/logout" \
    -H "Authorization: Bearer $GUEST_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$GUEST_REFRESH_TOKEN\"}"
assert_http 200 "重复登出返回 200（幂等）"
assert_code 0 "code=0"
# ============================================================
section "6. 修改密码 - 错误分支"
# ============================================================
http_request PUT "$API/auth/me/password" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"123456","new_password":"654321","confirm_password":"654321"}'
assert_http 401 "未认证修改密码 → 401"
assert_code 10001 "code=10001"
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"wrong-password","new_password":"654321","confirm_password":"654321"}'
assert_http 400 "旧密码错误返回 400"
assert_code 10003 "旧密码错误 code=10003"
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"123456","new_password":"654321","confirm_password":"abcdef"}'
assert_http 400 "两次密码不一致返回 400"
assert_code 10004 "code=10004"
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"123456","new_password":"x","confirm_password":"x"}'
assert_http 400 "短新密码返回 400"
assert_code 10007 "code=10007"
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"123456"}'
assert_http 422 "缺失字段 → 422"
assert_code 90004 "code=90004"
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"","new_password":"654321","confirm_password":"654321"}'
assert_http 400 "空旧密码进入业务校验 → 400"
assert_code 10003 "空旧密码 code=10003"
# ============================================================
section "7. 修改密码成功 & Refresh Token 撤销（guest 往返）"
# ============================================================
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
assert_http 200 "guest 重新登录成功"
GUEST_ACCESS=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
GUEST_REFRESH=$(echo "$HTTP_BODY" | jq -r '.data.refresh_token')
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $GUEST_ACCESS" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"123456","new_password":"abcdef","confirm_password":"abcdef"}'
assert_http 200 "guest 修改密码成功"
assert_code 0 "code=0"
http_request POST "$API/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$GUEST_REFRESH\"}"
assert_http 401 "改密后原 refresh_token 失效"
assert_code 10001 "code=10001"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"abcdef"}'
assert_http 200 "新密码登录成功"
assert_code 0 "code=0"
GUEST_ACCESS_NEW=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
http_request PUT "$API/auth/me/password" \
    -H "Authorization: Bearer $GUEST_ACCESS_NEW" \
    -H "Content-Type: application/json" \
    -d '{"old_password":"abcdef","new_password":"123456","confirm_password":"123456"}'
assert_http 200 "guest 密码已恢复为 123456"
assert_code 0 "code=0"
# ============================================================
section "8. 权限校验 - 空 Bearer"
# ============================================================
http_request GET "$API/auth/me" -H "Authorization: Bearer "
assert_http 401 "空 Bearer Token 返回 401"
assert_code 10001 "code=10001"
# ============================================================
# 汇总
# ============================================================
TOTAL=$((PASS + FAIL))
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "${YELLOW}测试结果汇总${NC}"
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：${TOTAL}"
echo -e "  通过：${GREEN}${PASS}${NC}"
echo -e "  失败：${RED}${FAIL}${NC}"
if [ "$FAIL" -eq 0 ]; then
    echo -e "\n${GREEN}🎉 认证模块测试全部通过${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
