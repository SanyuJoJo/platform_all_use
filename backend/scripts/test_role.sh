#!/usr/bin/env bash
# ============================================================
# 角色管理模块 API 测试脚本（v1.3.2）
# 用法：./scripts/test_role.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（./scripts/migrate.sh upgrade head）
#   2. 服务已启动（默认端口 8000）
#
# 依赖：curl、jq
#
# v1.3.2 修复：
#   - 步骤 7 前重置 guest 密码为 123456，避免其他测试脚本
#     （test_auth.sh / test_user.sh）中途失败污染 guest 密码
#     导致 guest 登录返回 401 / 10001，进而角色模块步骤 7 失败。
#   - 增加 guest 用户不存在的明确诊断提示。
#
# v1.3 变更：
#   - P2-NEW-E：脚本头部版本标记统一为 v1.3，与文档 § 2.2 / § 五.6 一致。
#
# v1.2 变更：
#   - P1-NEW-3：新增 description 超长用例（90001）。
#
# v1.1 新增：
#   - P1-8：提供端到端测试脚本
#   - 覆盖创建/列表/详情/更新/删除/权限保护/错误码
#   - 覆盖 P1-1：路由无尾斜杠
#   - 覆盖 P0-3：系统内置角色基于 is_system 判断
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

echo -e "${YELLOW}角色管理模块 API 测试（v1.3.2）${NC}"
echo "BASE_URL = $BASE_URL"

# ---------- 登录 ----------
section "0. 登录获取 Token"
http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ admin 登录失败，无法继续测试，请确认服务与数据库迁移正常${NC}"
    exit 1
fi

# ---------- 列表 ----------
section "1. 查询角色列表"
# P1-1：无尾斜杠路径
http_request GET "$API/auth/roles?page=1&page_size=10" -H "Authorization: Bearer $TOKEN"
assert_http 200 "GET /auth/roles 返回 200"
assert_code 0 "code=0"
TOTAL=$(echo "$HTTP_BODY" | jq -r '.data.total')
pass "当前角色总数：$TOTAL"

# ---------- 创建 ----------
section "2. 创建角色"
ROLE_CODE="test_role_$(date +%s)"
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"测试角色_$ROLE_CODE\",\"code\":\"$ROLE_CODE\",\"description\":\"脚本测试\",\"permission_codes\":[\"auth:user:view\"]}"
assert_http 200 "创建成功"
assert_code 0 "code=0"
ROLE_ID=$(echo "$HTTP_BODY" | jq -r '.data.id')
pass "新角色 ID = $ROLE_ID"

# 重复编码
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"重复编码","code":"admin","permission_codes":[]}'
assert_http 400 "重复编码 → 400"
assert_code 20001 "code=20001"

# 重复名称
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"管理员\",\"code\":\"dup_$ROLE_CODE\",\"permission_codes\":[]}"
assert_http 400 "重复名称 → 400"
assert_code 20004 "code=20004"

# 编码格式非法
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"格式错误_$ROLE_CODE\",\"code\":\"BadCode\",\"permission_codes\":[]}"
assert_http 400 "编码格式非法 → 400"
assert_code 90001 "code=90001"

# 描述超长（v1.2 新增用例）
DESC_LONG=$(printf 'x%.0s' {1..300})
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"描述超长_$ROLE_CODE\",\"code\":\"desclong_$ROLE_CODE\",\"description\":\"$DESC_LONG\",\"permission_codes\":[]}"
assert_http 400 "描述超长 → 400"
assert_code 90001 "code=90001"

# 权限格式非法
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"权限格式_$ROLE_CODE\",\"code\":\"fmt_$ROLE_CODE\",\"permission_codes\":[\"badformat\"]}"
assert_http 400 "权限格式非法 → 400"
assert_code 20054 "code=20054"

# 权限不存在
http_request POST "$API/auth/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"权限不存在_$ROLE_CODE\",\"code\":\"noperm_$ROLE_CODE\",\"permission_codes\":[\"no:such:perm\"]}"
assert_http 404 "权限不存在 → 404"
assert_code 20052 "code=20052"

# ---------- 详情 ----------
section "3. 获取角色详情"
http_request GET "$API/auth/roles/$ROLE_ID" -H "Authorization: Bearer $TOKEN"
assert_http 200 "详情请求成功"
assert_code 0 "code=0"
IS_SYSTEM=$(echo "$HTTP_BODY" | jq -r '.data.is_system')
pass "is_system = $IS_SYSTEM（自定义角色应为 0）"

# ---------- 更新 ----------
section "4. 更新角色"
http_request PUT "$API/auth/roles/$ROLE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"description":"已更新","permission_codes":["auth:role:view"]}'
assert_http 200 "更新成功"
assert_code 0 "code=0"
PERMS=$(echo "$HTTP_BODY" | jq -r '.data.permission_codes | join(",")')
pass "更新后权限列表：$PERMS"

# v1.3 新增用例 P2-NEW-B：传 code 字段 → 422 / 90004
http_request PUT "$API/auth/roles/$ROLE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"code":"new_code"}'
assert_http 422 "传 code 字段 → 422"
assert_code 90004 "code=90004"

# v1.3 新增用例 P2-NEW-B：传 is_system 字段 → 422 / 90004
http_request PUT "$API/auth/roles/$ROLE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"is_system":1}'
assert_http 422 "传 is_system 字段 → 422"
assert_code 90004 "code=90004"

# ---------- 删除系统角色保护 ----------
section "5. 删除系统内置角色"
http_request GET "$API/auth/roles?keyword=admin" -H "Authorization: Bearer $TOKEN"
ADMIN_ROLE_ID=$(echo "$HTTP_BODY" | jq -r '.data.items[] | select(.code=="admin") | .id' | head -1)
if [ -n "$ADMIN_ROLE_ID" ] && [ "$ADMIN_ROLE_ID" != "null" ]; then
    http_request DELETE "$API/auth/roles/$ADMIN_ROLE_ID" -H "Authorization: Bearer $TOKEN"
    assert_http 403 "删除系统角色 → 403"
    assert_code 20003 "code=20003"
else
    fail "未找到 admin 角色"
fi

# ---------- 删除自定义角色 ----------
section "6. 删除自定义角色"
http_request DELETE "$API/auth/roles/$ROLE_ID" -H "Authorization: Bearer $TOKEN"
assert_http 200 "删除成功"
assert_code 0 "code=0"

# ---------- 权限校验 ----------
section "7. 权限校验"

# ★ v1.3.2 关键修复：
# 确保 guest 用户存在且密码为 123456。
# 原因：guest 是全局共享的种子账号，其他测试脚本
# （test_auth.sh / test_user.sh）可能临时修改过其密码而未恢复，
# 导致本脚本步骤 7 的 guest 登录返回 401 / 10001。
# 方案：用 admin token 通过 PATCH /auth/users/{id}/password 重置。

http_request GET "$API/auth/users?keyword=guest" -H "Authorization: Bearer $TOKEN"
GUEST_ID=$(echo "$HTTP_BODY" | jq -r '.data.items[]? | select(.username=="guest") | .id' | head -1)

if [ -z "$GUEST_ID" ] || [ "$GUEST_ID" = "null" ]; then
    fail "guest 用户不存在，无法测试权限校验；请重启服务以触发 ensure_auth_seed_data 或手动检查数据库"
else
    pass "定位到 guest 用户 ID = $GUEST_ID"

    # 重置 guest 密码为 123456（幂等；即便原本就是 123456 也安全）
    http_request PATCH "$API/auth/users/$GUEST_ID/password" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"new_password":"123456"}'

    if [ "$HTTP_CODE" = "200" ]; then
        pass "guest 密码已重置为 123456"
    else
        fail "guest 密码重置失败 (HTTP $HTTP_CODE, body=$HTTP_BODY)"
    fi

    # guest 登录
    http_request POST "$API/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"guest","password":"123456"}'

    GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token // empty')

    if [ -z "$GUEST_TOKEN" ] || [ "$GUEST_TOKEN" = "null" ]; then
        fail "guest 登录失败，无法继续权限校验测试 (HTTP $HTTP_CODE, body=$HTTP_BODY)"
    else
        pass "guest 登录成功"

        # guest 无 auth:role:create 权限 → 20051
        http_request POST "$API/auth/roles" \
            -H "Authorization: Bearer $GUEST_TOKEN" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"无权限_$ROLE_CODE\",\"code\":\"guest_$ROLE_CODE\",\"permission_codes\":[]}"
        assert_http 403 "guest 创建角色被拒绝"
        assert_code 20051 "code=20051"
    fi
fi

# ---------- 汇总 ----------
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS+FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"

if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 角色管理模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
