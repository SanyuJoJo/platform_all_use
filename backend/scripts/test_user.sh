#!/usr/bin/env bash
# ============================================================
# 用户管理模块 API 测试脚本（v1.2，v1.3 未变更）
# 用法：./scripts/test_user.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移
#   2. 服务已启动（默认端口 8000）
#
# 依赖：curl、jq、（可选）sqlite3
#
# v1.2 变更：
#   - P2-3：DB_FILE 支持环境变量覆盖；支持从 .env 的 DATABASE_URL 解析
#   - 新增昵称超长用例（P2-4）
# v1.1 变更：
#   - P1-3：动态获取 admin id 与 guest 角色 id
#   - P2-4：增加 trap cleanup EXIT
#   - P2-6：默认端口从 18000 调整为 8000
# ============================================================
set -uo pipefail
BASE_URL="${1:-http://localhost:18000}"
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
# ---------------------------------------------------------------------------
# P2-3：数据库文件路径解析
# 优先级：环境变量 DB_FILE > .env 中 DATABASE_URL > 默认 app.db
# ---------------------------------------------------------------------------
resolve_db_file() {
    if [ -n "${DB_FILE:-}" ]; then
        echo "$DB_FILE"
        return
    fi
    if [ -f .env ]; then
        local url
        url=$(grep -E '^DATABASE_URL=' .env | head -1 | sed -E 's/^DATABASE_URL=//')
        if [[ "$url" =~ sqlite(\+aiosqlite)?:///\.?/?(.+)$ ]]; then
            echo "${BASH_REMATCH[2]}"
            return
        fi
    fi
    echo "app.db"
}
DB_FILE_PATH="$(resolve_db_file)"
TOKEN=""
NEW_ID=""
cleanup() {
    if [ -n "$NEW_ID" ] && [ "$NEW_ID" != "null" ] && [ -n "$TOKEN" ]; then
        curl -sS -X DELETE "${API}/auth/users/${NEW_ID}" \
            -H "Authorization: Bearer $TOKEN" >/dev/null 2>&1 || true
    fi
}
trap cleanup EXIT
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
# ---------------------------------------------------------------------------
# P1-3 / P2-3：动态获取 guest 角色 ID
# ---------------------------------------------------------------------------
get_guest_role_id() {
    local role_id=""
    local roles_resp
    roles_resp=$(curl -sS "${API}/auth/roles?page=1&page_size=100" \
        -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo "")
    if echo "$roles_resp" | jq -e '.data.items // [] | length > 0' >/dev/null 2>&1; then
        role_id=$(echo "$roles_resp" | jq -r '.data.items[]? | select(.code=="guest") | .id' 2>/dev/null | head -1)
    fi
    if { [ -z "$role_id" ] || [ "$role_id" = "null" ]; } \
        && command -v sqlite3 >/dev/null 2>&1 \
        && [ -f "$DB_FILE_PATH" ]; then
        role_id=$(sqlite3 "$DB_FILE_PATH" "SELECT id FROM auth_role WHERE code='guest' LIMIT 1;" 2>/dev/null || echo "")
    fi
    echo "$role_id"
}
echo -e "${YELLOW}用户管理模块 API 测试（v1.2）${NC}"
echo "BASE_URL     = $BASE_URL"
echo "DB_FILE_PATH = $DB_FILE_PATH"
# ---------- 登录 ----------
section "0. 登录获取 Token"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
http_request GET "$API/auth/me" -H "Authorization: Bearer $TOKEN"
ADMIN_ID=$(echo "$HTTP_BODY" | jq -r '.data.id')
if [ -n "$ADMIN_ID" ] && [ "$ADMIN_ID" != "null" ]; then
    pass "动态获取 admin id=$ADMIN_ID"
else
    fail "无法获取 admin id"
fi
GUEST_ROLE_ID=$(get_guest_role_id)
if [ -n "$GUEST_ROLE_ID" ] && [ "$GUEST_ROLE_ID" != "null" ]; then
    pass "动态获取 guest 角色 id=$GUEST_ROLE_ID"
    ROLE_IDS_JSON="[$GUEST_ROLE_ID]"
else
    echo -e "  ${YELLOW}⚠${NC} 未获取到 guest 角色 ID，创建用户将不分配角色"
    ROLE_IDS_JSON="[]"
fi
# ---------- 列表 ----------
section "1. 查询用户列表"
http_request GET "$API/auth/users?page=1&page_size=10" -H "Authorization: Bearer $TOKEN"
assert_http 200 "列表请求成功"
assert_code 0 "code=0"
TOTAL=$(echo "$HTTP_BODY" | jq -r '.data.total')
pass "当前用户总数：$TOTAL"
http_request GET "$API/auth/users?keyword=admin" -H "Authorization: Bearer $TOKEN"
assert_http 200 "关键字筛选成功"
http_request GET "$API/auth/users?status=1" -H "Authorization: Bearer $TOKEN"
assert_http 200 "状态筛选成功"
# ---------- 创建 ----------
section "2. 创建用户"
NEW_USER="curl_$(date +%s)"
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$NEW_USER\",\"password\":\"123456\",\"nickname\":\"测试用户\",\"email\":\"$NEW_USER@example.com\",\"role_ids\":$ROLE_IDS_JSON}"
assert_http 200 "创建成功"
assert_code 0 "code=0"
NEW_ID=$(echo "$HTTP_BODY" | jq -r '.data.id')
UPDATED_AT=$(echo "$HTTP_BODY" | jq -r '.data.updated_at')
pass "新用户 ID = $NEW_ID"
if [ -n "$UPDATED_AT" ] && [ "$UPDATED_AT" != "null" ]; then
    pass "P1-1：updated_at 非空（$UPDATED_AT）"
else
    fail "P1-1：updated_at 为 null"
fi
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456","nickname":"重复","role_ids":[]}'
assert_http 400 "重复用户名 → 400"
assert_code 10000 "code=10000"
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"email_$(date +%s)\",\"password\":\"123456\",\"nickname\":\"重复邮箱\",\"email\":\"admin@example.com\",\"role_ids\":[]}"
assert_http 409 "重复邮箱 → 409"
assert_code 10009 "code=10009"
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"username":"ab","password":"123456","nickname":"短名","role_ids":[]}'
assert_http 400 "用户名格式无效 → 400"
assert_code 10006 "code=10006"
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$(printf 'a%.0s' {1..30})\",\"password\":\"123456\",\"nickname\":\"长名\",\"role_ids\":[]}"
assert_http 400 "P1-4：长用户名 → 400"
assert_code 10006 "code=10006"
NICK_LONG=$(printf 'x%.0s' {1..100})
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"nn_$(date +%s)\",\"password\":\"123456\",\"nickname\":\"$NICK_LONG\",\"role_ids\":[]}"
assert_http 400 "P2-4：昵称超长 → 400"
assert_code 90001 "code=90001"
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"sp_$(date +%s)\",\"password\":\"x\",\"nickname\":\"短密码\",\"role_ids\":[]}"
assert_http 400 "密码过短 → 400"
assert_code 10007 "code=10007"
# ---------- 详情 ----------
section "3. 获取用户详情"
http_request GET "$API/auth/users/$NEW_ID" -H "Authorization: Bearer $TOKEN"
assert_http 200 "详情请求成功"
assert_code 0 "code=0"
http_request GET "$API/auth/users/999999" -H "Authorization: Bearer $TOKEN"
assert_http 404 "用户不存在 → 404"
assert_code 10005 "code=10005"
# ---------- 更新 ----------
section "4. 更新用户"
http_request PUT "$API/auth/users/$NEW_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"nickname\":\"已更新昵称\",\"role_ids\":$ROLE_IDS_JSON}"
assert_http 200 "更新成功"
assert_code 0 "code=0"
http_request PUT "$API/auth/users/$NEW_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"nickname\":\"$NICK_LONG\"}"
assert_http 400 "P2-4：更新昵称超长 → 400"
assert_code 90001 "code=90001"
http_request PUT "$API/auth/users/$ADMIN_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":0}'
assert_http 403 "禁用自己 → 403"
assert_code 10011 "code=10011"
# ---------- 启用/禁用 ----------
section "5. 启用/禁用用户"
http_request PATCH "$API/auth/users/$NEW_ID/status" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":0}'
assert_http 200 "禁用成功"
assert_code 0 "code=0"
http_request PATCH "$API/auth/users/$NEW_ID/status" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":1}'
assert_http 200 "启用成功"
assert_code 0 "code=0"
# ---------- 重置密码 ----------
section "6. 重置用户密码"
http_request PATCH "$API/auth/users/$NEW_ID/password" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"new_password":"abcdef"}'
assert_http 200 "重置成功"
assert_code 0 "code=0"
http_request PATCH "$API/auth/users/$NEW_ID/password" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"new_password":"x"}'
assert_http 400 "密码过短 → 400"
assert_code 10007 "code=10007"
# ---------- 权限 ----------
section "7. 权限校验"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
http_request POST "$API/auth/users" \
    -H "Authorization: Bearer $GUEST_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"gt_$(date +%s)\",\"password\":\"123456\",\"nickname\":\"无权限\",\"role_ids\":[]}"
assert_http 403 "guest 创建用户被拒绝"
assert_code 20051 "code=20051"
# ---------- 删除 ----------
section "8. 删除用户"
http_request DELETE "$API/auth/users/$NEW_ID" -H "Authorization: Bearer $TOKEN"
assert_http 200 "删除成功"
assert_code 0 "code=0"
NEW_ID=""
http_request DELETE "$API/auth/users/$ADMIN_ID" -H "Authorization: Bearer $TOKEN"
assert_http 403 "删除自己 → 403"
assert_code 10010 "code=10010"
# ---------- 汇总 ----------
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS+FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"
if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 用户管理模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
