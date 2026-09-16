#!/usr/bin/env bash
# ============================================================
# 用户角色权限模块（RBAC）功能测试脚本
# 版本：v1.3
# 用法：./scripts/test_rbac.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 覆盖内容：
#   1. 健康检查
#   2. admin / guest 登录
#   3. 用户 CRUD（增删改查 + 启用禁用 + 重置密码）
#   4. 角色 CRUD（增删改查 + 权限分配）
#   5. 权限查询（列表 + 按模块）
#   6. 权限校验（guest 无法增删改）
#   7. 错误码验证（10000/10005/10006/10007/10010/10011/20001/20003/20004/20051）
#   8. status 字段边界（默认值 / 非法值）
#   9. 查询参数边界（分页默认值 / type_error）
#
# 依赖：curl、jq
#
# v1.3.1 修复：
#   - TEST_USERNAME 从 rbac_test_${RUN_TAG}（20 字符）缩短为
#     rt_${RUN_TAG}（13 字符），避免拼接后缀 _sp / _bad / _bt 后
#     超过 username 3-20 位上限，导致触发 10006 而非预期错误码。
#   - 4.6 密码过短测试：用户名改为独立的 sp_${RUN_TAG}，
#     确保不因用户名长度污染密码长度校验的断言。
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
pass()    { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS+1)); }
fail()    { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
info()    { echo -e "  ${YELLOW}ℹ${NC}  $1"; }
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
    local expected="$1"; local msg="$2"
    if [ "$HTTP_CODE" = "$expected" ]; then
        pass "$msg (HTTP $HTTP_CODE)"
    else
        fail "$msg (期望 HTTP $expected，实际 $HTTP_CODE，body=$HTTP_BODY)"
    fi
}

assert_code() {
    local expected="$1"; local msg="$2"
    local actual
    actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$expected" ]; then
        pass "$msg (code=$actual)"
    else
        fail "$msg (期望 code=$expected，实际 $actual，body=$HTTP_BODY)"
    fi
}

assert_not_null() {
    local field="$1"; local msg="$2"
    local value
    value=$(echo "$HTTP_BODY" | jq -r "${field} // \"null\"")
    if [ "$value" != "null" ] && [ -n "$value" ]; then
        pass "$msg"
    else
        fail "$msg (字段 $field 为 null)"
    fi
}

assert_eq() {
    local actual="$1"; local expected="$2"; local msg="$3"
    if [ "$actual" = "$expected" ]; then
        pass "$msg ($actual)"
    else
        fail "$msg (期望 $expected，实际 $actual)"
    fi
}

# ============================================================
# 生成唯一标识（避免测试数据冲突）
#
# 长度约束说明：
#   - username：3-20 位，使用 rt_${RUN_TAG}（13 位），
#     拼接后缀 _sp / _bad / _bt 后仍 ≤ 17 位，安全。
#   - role.code：1-50 位，使用 rbac_role_${RUN_TAG}（20 位），安全。
# ============================================================
RUN_TAG=$(date +%s)
TEST_USERNAME="rt_${RUN_TAG}"
TEST_ROLECODE="rbac_role_${RUN_TAG}"

cleanup() {
    if [ -n "${ADMIN_TOKEN:-}" ]; then
        # 清理测试用户
        if [ -n "${TEST_USER_ID:-}" ] && [ "$TEST_USER_ID" != "null" ]; then
            curl -sS -X DELETE "${API}/auth/users/${TEST_USER_ID}" \
                -H "Authorization: Bearer $ADMIN_TOKEN" >/dev/null 2>&1 || true
        fi
        # 清理测试角色
        if [ -n "${TEST_ROLE_ID:-}" ] && [ "$TEST_ROLE_ID" != "null" ]; then
            curl -sS -X DELETE "${API}/auth/roles/${TEST_ROLE_ID}" \
                -H "Authorization: Bearer $ADMIN_TOKEN" >/dev/null 2>&1 || true
        fi
    fi
}
trap cleanup EXIT

echo "=============================================="
echo "  RBAC 功能测试（v1.3.1）"
echo "  BASE_URL = $BASE_URL"
echo "=============================================="

# ============================================================
# 1. 健康检查
# ============================================================
section "1. 健康检查"
http_request GET "${BASE_URL}/health"
assert_http 200 "GET /health 返回 200"
assert_code 0 "健康检查 code=0"

# ============================================================
# 2. 登录
# ============================================================
section "2. 登录获取 Token"

http_request POST "${API}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
assert_code 0 "admin code=0"
ADMIN_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo -e "${RED}❌ admin 登录失败，无法继续${NC}"
    exit 1
fi

http_request POST "${API}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
assert_http 200 "guest 登录成功"
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -z "$GUEST_TOKEN" ] || [ "$GUEST_TOKEN" = "null" ]; then
    echo -e "${RED}❌ guest 登录失败，无法继续${NC}"
    exit 1
fi

# ============================================================
# 3. 用户管理 - 列表
# ============================================================
section "3. 用户管理 - 列表"

http_request GET "${API}/auth/users?page=1&page_size=10" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "GET /auth/users 返回 200"
assert_code 0 "code=0"
TOTAL=$(echo "$HTTP_BODY" | jq -r '.data.total')
info "当前用户总数：$TOTAL"

# 分页默认值（v1.2 修复 P1-4）
http_request GET "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "未传分页参数返回 200"
PAGE_DEFAULT=$(echo "$HTTP_BODY" | jq -r '.data.page')
assert_eq "$PAGE_DEFAULT" "1" "分页默认 page=1（v1.2 P1-4）"

# 关键字筛选
http_request GET "${API}/auth/users?keyword=admin" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "关键字筛选返回 200"

# 未认证
http_request GET "${API}/auth/users"
assert_http 401 "未认证 → 401"
assert_code 10001 "code=10001"

# ============================================================
# 4. 用户管理 - 创建
# ============================================================
section "4. 用户管理 - 创建"

# 4.1 正常创建（省略 status，应默认启用）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"123456\",\"nickname\":\"RBAC测试用户\",\"email\":\"${TEST_USERNAME}@example.com\"}"
assert_http 200 "创建用户成功"
assert_code 0 "code=0"
TEST_USER_ID=$(echo "$HTTP_BODY" | jq -r '.data.id')
DEFAULT_STATUS=$(echo "$HTTP_BODY" | jq -r '.data.status')
assert_eq "$DEFAULT_STATUS" "1" "省略 status 默认启用（v1.2 P1-A）"

# 4.2 缺失必填字段（v1.1 P1-3）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{}'
assert_http 422 "空 body → 422"
assert_code 90004 "code=90004"

# 4.3 非法 status（v1.2 P1-B）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}_bad\",\"password\":\"123456\",\"nickname\":\"x\",\"status\":5}"
assert_http 422 "status=5 → 422"
assert_code 90004 "code=90004"

# 4.4 用户名重复（10000）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"123456\",\"nickname\":\"重复\"}"
assert_http 400 "用户名重复 → 400"
assert_code 10000 "code=10000"

# 4.5 用户名格式非法（10006）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"ab\",\"password\":\"123456\",\"nickname\":\"短名\"}"
assert_http 400 "用户名格式非法 → 400"
assert_code 10006 "code=10006"

# 4.6 密码过短（10007）
# 注意：使用独立的短用户名 sp_${RUN_TAG}（13 位），
# 避免继承 TEST_USERNAME 拼接后缀后超长，触发 10006 而非 10007。
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"sp_${RUN_TAG}\",\"password\":\"x\",\"nickname\":\"短密码\"}"
assert_http 400 "密码过短 → 400"
assert_code 10007 "code=10007"

# ============================================================
# 5. 用户管理 - 详情 / 更新 / 删除
# ============================================================
section "5. 用户管理 - 详情/更新"

http_request GET "${API}/auth/users/${TEST_USER_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "获取用户详情成功"
assert_code 0 "code=0"

http_request GET "${API}/auth/users/999999" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 404 "用户不存在 → 404"
assert_code 10005 "code=10005"

# 更新昵称
http_request PUT "${API}/auth/users/${TEST_USER_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"nickname":"更新后的昵称"}'
assert_http 200 "更新用户成功"
assert_code 0 "code=0"

# 禁用自己（10011）
ADMIN_ID=$(curl -sS "${API}/auth/me" \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.data.id')
http_request PUT "${API}/auth/users/${ADMIN_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":0}'
assert_http 403 "禁用自己 → 403"
assert_code 10011 "code=10011"

# 删除自己（10010）
http_request DELETE "${API}/auth/users/${ADMIN_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 403 "删除自己 → 403"
assert_code 10010 "code=10010"

# ============================================================
# 6. 用户管理 - 启用/禁用 + 重置密码
# ============================================================
section "6. 用户管理 - 状态/密码"

http_request PATCH "${API}/auth/users/${TEST_USER_ID}/status" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":0}'
assert_http 200 "禁用用户成功"
assert_code 0 "code=0"

http_request PATCH "${API}/auth/users/${TEST_USER_ID}/status" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":1}'
assert_http 200 "启用用户成功"
assert_code 0 "code=0"

http_request PATCH "${API}/auth/users/${TEST_USER_ID}/password" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"new_password":"abcdef"}'
assert_http 200 "重置密码成功"
assert_code 0 "code=0"

# ============================================================
# 7. 角色管理 - 列表 / 创建
# ============================================================
section "7. 角色管理 - 列表/创建"

http_request GET "${API}/auth/roles?page=1&page_size=10" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "GET /auth/roles 返回 200"
assert_code 0 "code=0"

# 创建角色（带权限）
http_request POST "${API}/auth/roles" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"RBAC测试角色_${RUN_TAG}\",\"code\":\"${TEST_ROLECODE}\",\"description\":\"test\",\"permission_codes\":[\"auth:user:view\"]}"
assert_http 200 "创建角色成功"
assert_code 0 "code=0"
TEST_ROLE_ID=$(echo "$HTTP_BODY" | jq -r '.data.id')
info "测试角色 ID = $TEST_ROLE_ID"

# 缺失必填字段
http_request POST "${API}/auth/roles" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{}'
assert_http 422 "空 body → 422"
assert_code 90004 "code=90004"

# 角色编码重复
http_request POST "${API}/auth/roles" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"重复编码_${RUN_TAG}\",\"code\":\"${TEST_ROLECODE}\"}"
assert_http 400 "角色编码重复 → 400"
assert_code 20001 "code=20001"

# 角色名称重复
http_request POST "${API}/auth/roles" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"RBAC测试角色_${RUN_TAG}\",\"code\":\"dup_${RUN_TAG}\"}"
assert_http 400 "角色名称重复 → 400"
assert_code 20004 "code=20004"

# ============================================================
# 8. 角色管理 - 详情 / 更新 / 删除
# ============================================================
section "8. 角色管理 - 详情/更新"

http_request GET "${API}/auth/roles/${TEST_ROLE_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "获取角色详情成功"
assert_code 0 "code=0"

# 更新角色
http_request PUT "${API}/auth/roles/${TEST_ROLE_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"description":"更新后","permission_codes":["auth:user:view","auth:role:view"]}'
assert_http 200 "更新角色成功"
assert_code 0 "code=0"

# extra="forbid"：传 code 字段被拒绝（v1.1 P0-3）
http_request PUT "${API}/auth/roles/${TEST_ROLE_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"code":"new_code"}'
assert_http 422 "传 code 字段 → 422"
assert_code 90004 "code=90004"

# ============================================================
# 9. 角色管理 - 系统角色保护
# ============================================================
section "9. 角色管理 - 系统角色保护"

# 查询 admin 角色 ID
http_request GET "${API}/auth/roles?keyword=admin" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
ADMIN_ROLE_ID=$(echo "$HTTP_BODY" | jq -r '.data.items[]? | select(.code=="admin") | .id' | head -1)
if [ -n "$ADMIN_ROLE_ID" ] && [ "$ADMIN_ROLE_ID" != "null" ]; then
    http_request DELETE "${API}/auth/roles/${ADMIN_ROLE_ID}" \
        -H "Authorization: Bearer $ADMIN_TOKEN"
    assert_http 403 "删除系统角色 → 403"
    assert_code 20003 "code=20003"
else
    fail "未找到 admin 角色"
fi

# ============================================================
# 10. 权限管理 - 查询
# ============================================================
section "10. 权限管理 - 查询"

http_request GET "${API}/auth/permissions" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "查询全部权限成功"
assert_code 0 "code=0"
PERM_COUNT=$(echo "$HTTP_BODY" | jq '.data | length')
info "权限总数：$PERM_COUNT"

http_request GET "${API}/auth/permissions?module_id=auth" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "按模块查询成功"
assert_code 0 "code=0"

http_request GET "${API}/auth/permissions/modules/auth" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "获取模块权限成功"
assert_code 0 "code=0"

# 非法 module_id
http_request GET "${API}/auth/permissions?module_id=BadModule" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 400 "非法 module_id → 400"
assert_code 90001 "code=90001"

# ============================================================
# 11. 权限校验 - guest 只读
# ============================================================
section "11. 权限校验 - guest 只读"

# guest 可以查看
http_request GET "${API}/auth/users" \
    -H "Authorization: Bearer $GUEST_TOKEN"
assert_http 200 "guest 查看用户列表 → 200"

# guest 不能创建（P0-3 安全漏洞修复）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $GUEST_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"hacker_${RUN_TAG}\",\"password\":\"123456\",\"nickname\":\"hacker\"}"
assert_http 403 "guest 创建用户 → 403（P0-3）"
assert_code 20051 "code=20051"

# guest 不能删除
http_request DELETE "${API}/auth/users/${TEST_USER_ID}" \
    -H "Authorization: Bearer $GUEST_TOKEN"
assert_http 403 "guest 删除用户 → 403（P0-3）"
assert_code 20051 "code=20051"

# guest 不能创建角色
http_request POST "${API}/auth/roles" \
    -H "Authorization: Bearer $GUEST_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"x\",\"code\":\"g_${RUN_TAG}\"}"
assert_http 403 "guest 创建角色 → 403（P0-3）"
assert_code 20051 "code=20051"

# ============================================================
# 12. 边界场景
# ============================================================
section "12. 边界场景"

# query 参数类型错误（P2-B）
http_request GET "${API}/auth/users?page=abc" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 422 "page=abc → 422"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type')
assert_eq "$ERR_TYPE" "type_error" "query 类型错误 type=type_error（v1.2 P2-B）"

# body 类型错误（P2-C）
http_request POST "${API}/auth/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}_bt\",\"password\":\"123456\",\"nickname\":\"x\",\"status\":\"not-a-number\"}"
assert_http 422 "body status 为字符串 → 422"
assert_code 90004 "code=90004"
ERR_TYPE=$(echo "$HTTP_BODY" | jq -r '.data.errors[0].type')
assert_eq "$ERR_TYPE" "invalid_json" "body 类型错误 type=invalid_json（v1.3 P2-C）"

# ============================================================
# 13. 清理
# ============================================================
section "13. 清理测试数据"

http_request DELETE "${API}/auth/users/${TEST_USER_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "删除测试用户成功"
TEST_USER_ID=""

http_request DELETE "${API}/auth/roles/${TEST_ROLE_ID}" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
assert_http 200 "删除测试角色成功"
TEST_ROLE_ID=""

# ============================================================
# 汇总
# ============================================================
TOTAL=$((PASS + FAIL))
echo ""
echo "=============================================="
echo "  测试结果汇总"
echo "=============================================="
echo "  总计：$TOTAL"
echo -e "  通过：${GREEN}${PASS}${NC}"
echo -e "  失败：${RED}${FAIL}${NC}"
echo "=============================================="

if [ "$FAIL" -eq 0 ]; then
    echo -e "\n${GREEN}🎉 RBAC 功能测试全部通过${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
