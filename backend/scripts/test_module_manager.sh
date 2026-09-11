#!/usr/bin/env bash
# ============================================================
# 模块管理模块 API 测试脚本（v1.1）
# 用法：./scripts/test_module_manager.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移
#   2. 服务已启动
#   3. 已安装 curl、jq、zip
# ============================================================
set -uo pipefail

BASE_URL="${1:-http://localhost:8000}"
API="${BASE_URL}/api/v1"

for cmd in curl jq zip; do
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
    if [ "$actual" = "$1" ]; then pass "$2 (code=$actual)"; else fail "$2 (expected code=$1, got $actual, body=$HTTP_BODY)"; fi
}

echo -e "${YELLOW}模块管理模块 API 测试（v1.1）${NC}"
echo "BASE_URL = $BASE_URL"

section "0. 登录"
http_request POST "$API/auth/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"123456"}'
assert_http 200 "admin 登录成功"
TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ admin 登录失败，无法继续${NC}"
    exit 1
fi

section "1. 查询模块列表"
http_request GET "$API/modules?page=1&page_size=10" -H "Authorization: Bearer $TOKEN"
assert_http 200 "GET /modules 返回 200"
assert_code 0 "code=0"
TOTAL=$(echo "$HTTP_BODY" | jq -r '.data.total')
pass "当前模块总数：$TOTAL"

section "2. 核心模块保护"
http_request POST "$API/modules/auth/disable" -H "Authorization: Bearer $TOKEN"
assert_http 403 "停用核心模块 → 403"
assert_code 30013 "code=30013"

http_request POST "$API/modules/auth/enable" -H "Authorization: Bearer $TOKEN"
assert_http 403 "启用核心模块 → 403"
assert_code 30013 "code=30013"

http_request DELETE "$API/modules/auth" -H "Authorization: Bearer $TOKEN"
assert_http 403 "卸载核心模块 → 403"
assert_code 30013 "code=30013"

section "3. 配置查询与更新"
http_request GET "$API/modules/module_manager/config" -H "Authorization: Bearer $TOKEN"
assert_http 200 "获取 module_manager 配置成功"
assert_code 0 "code=0"

http_request PUT "$API/modules/module_manager/config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"config":{"enable_auto_install":false}}'
assert_http 200 "更新配置成功"
assert_code 0 "code=0"

section "4. 卸载不存在模块"
http_request DELETE "$API/modules/not_exist_module" -H "Authorization: Bearer $TOKEN"
assert_http 404 "不存在模块 → 404"
assert_code 30003 "code=30003"

section "5. ZIP 上传安装"
MODULE_ID="e2e_$(date +%s)"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/$MODULE_ID"
cat > "$TMP_DIR/$MODULE_ID/manifest.json" <<EOF
{
  "id": "$MODULE_ID",
  "name": "E2E Test Module",
  "version": "1.0.0",
  "description": "E2E test",
  "permissions": [
    {"code":"$MODULE_ID:demo:view","name":"查看","resource":"demo","action":"view"}
  ],
  "menus": [],
  "entry_backend": "router:router",
  "entry_frontend": null,
  "database_tables": []
}
EOF
cat > "$TMP_DIR/$MODULE_ID/router.py" <<'EOF'
from fastapi import APIRouter
router = APIRouter(tags=['e2e'])

@router.get('/ping')
async def ping():
    return {'pong': True}
EOF
touch "$TMP_DIR/$MODULE_ID/__init__.py"

ZIP_PATH="$TMP_DIR/$MODULE_ID.zip"
(cd "$TMP_DIR" && zip -qr "$ZIP_PATH" "$MODULE_ID")

http_request POST "$API/modules/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@$ZIP_PATH;type=application/zip"
assert_http 200 "上传安装成功"
assert_code 0 "code=0"

section "6. 启用模块并访问 API（热加载）"
http_request POST "$API/modules/$MODULE_ID/enable" -H "Authorization: Bearer $TOKEN"
assert_http 200 "启用成功"
assert_code 0 "code=0"

http_request GET "$API/$MODULE_ID/ping" -H "Authorization: Bearer $TOKEN"
assert_http 200 "启用后模块 API 可访问"
PONG=$(echo "$HTTP_BODY" | jq -r '.pong')
if [ "$PONG" = "true" ]; then
    pass "模块 API 返回 pong=true"
else
    fail "模块 API 返回异常：$HTTP_BODY"
fi

section "7. 停用模块"
http_request POST "$API/modules/$MODULE_ID/disable" -H "Authorization: Bearer $TOKEN"
assert_http 200 "停用成功"
assert_code 0 "code=0"

section "8. 卸载模块"
http_request DELETE "$API/modules/$MODULE_ID?force=true" -H "Authorization: Bearer $TOKEN"
assert_http 200 "卸载成功"
assert_code 0 "code=0"

section "9. 权限校验"
http_request POST "$API/auth/login" -H "Content-Type: application/json" -d '{"username":"guest","password":"123456"}'
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
http_request POST "$API/modules/upload" \
    -H "Authorization: Bearer $GUEST_TOKEN" \
    -F "file=@$ZIP_PATH;type=application/zip"
assert_http 403 "guest 上传安装被拒绝"
assert_code 20051 "code=20051"

echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS+FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"

if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 模块管理模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
