#!/usr/bin/env bash
# ============================================================
# License 管理模块 API 测试脚本（Go 版 v1.0）
# 用法：./scripts/test_license.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（make migrate-up）
#   2. 服务已启动（make run）
#   3. 环境变量 LICENSE_SECRET_KEY 与服务端一致
#
# 依赖：curl、jq、python3
# ============================================================
set -uo pipefail
BASE_URL="${1:-http://localhost:8000}"
API="${BASE_URL}/api/v1"
for cmd in curl jq python3; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "❌ 需要 $cmd，请先安装"
        exit 1
    fi
done
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
PASS=0; FAIL=0
pass()    { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS+1)); }
fail()    { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
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
    local actual; actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$1" ]; then pass "$2 (code=$actual)"; else fail "$2 (expected code=$1, got $actual, body=$HTTP_BODY)"; fi
}
# ---------- 密钥 ----------
if [ -z "${LICENSE_SECRET_KEY:-}" ]; then
    echo -e "${YELLOW}⚠  LICENSE_SECRET_KEY 未设置，使用默认密钥${NC}"
    LICENSE_SECRET_KEY="change-this-license-secret-in-production"
fi
# ---------- License 生成 ----------
generate_license() {
    local out_file="$1"
    local license_key="${2:-LIC-SCRIPT-$(date +%s)}"
    local expires_days="${3:-365}"
    local max_users="${4:-100}"
    local issued_offset_days="${5:-0}"
    LICENSE_SECRET_KEY="$LICENSE_SECRET_KEY" python3 - \
        "$out_file" "$license_key" "$expires_days" "$max_users" "$issued_offset_days" <<'PYEOF'
import hashlib, hmac, json, os, sys
from datetime import datetime, timedelta, timezone
out_file, license_key, expires_days, max_users, issued_offset_days = sys.argv[1:6]
secret = os.environ["LICENSE_SECRET_KEY"]
def canonical(payload):
    return json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
now = datetime.now(timezone.utc).replace(tzinfo=None)
issued = now + timedelta(days=int(issued_offset_days))
expires = now + timedelta(days=int(expires_days))
payload = {
    "license_key": license_key,
    "license_type": "enterprise",
    "max_users": int(max_users),
    "authorized_modules": ["auth", "audit_log", "customer_relation"],
    "machine_code": None,
    "issued_at": issued.isoformat(),
    "expires_at": expires.isoformat(),
}
signature = hmac.new(secret.encode(), canonical(payload).encode(), hashlib.sha256).hexdigest()
with open(out_file, "w", encoding="utf-8") as f:
    json.dump({"payload": payload, "signature": signature}, f, ensure_ascii=False)
PYEOF
}
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
echo "License 管理模块 API 测试（Go 版 v1.0）"
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
# ---------- 权限校验 ----------
section "1. 权限校验"
http_request GET "$API/license/status"
assert_http 401 "未认证 → 401"
assert_code 10001 "code=10001"
# ---------- 初始状态 ----------
section "2. 初始状态（未导入 License）"
http_request GET "$API/license/status" -H "Authorization: Bearer $TOKEN"
if [ "$HTTP_CODE" = "404" ]; then
    assert_code 50009 "无 License → 50009"
elif [ "$HTTP_CODE" = "200" ]; then
    pass "已存在 License，跳过初始状态验证"
fi
# ---------- 导入合法 License ----------
section "3. 导入合法 License"
LICENSE_FILE="$TMP_DIR/valid.lic"
generate_license "$LICENSE_FILE" "LIC-SCRIPT-VALID-$(date +%s)" 365 100 0
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$LICENSE_FILE;type=application/octet-stream"
assert_http 200 "导入成功"
assert_code 0 "code=0"
# ---------- 状态查询 ----------
section "4. 状态查询"
http_request GET "$API/license/status" -H "Authorization: Bearer $TOKEN"
assert_http 200 "状态查询成功"
assert_code 0 "code=0"
IS_VALID=$(echo "$HTTP_BODY" | jq -r '.data.is_valid')
pass "is_valid=$IS_VALID"
# ---------- 模块授权 ----------
section "5. 模块授权状态"
http_request GET "$API/license/modules" -H "Authorization: Bearer $TOKEN"
assert_http 200 "模块授权查询成功"
assert_code 0 "code=0"
# ---------- 错误场景 ----------
section "6. 错误场景"
# 6.1 非法 JSON
BAD_FILE="$TMP_DIR/bad.lic"
echo "not a json" > "$BAD_FILE"
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$BAD_FILE;type=application/octet-stream"
assert_http 400 "非法 JSON → 400"
assert_code 50001 "code=50001"
# 6.2 签名失败
BAD_SIG_FILE="$TMP_DIR/badsig.lic"
python3 - "$BAD_SIG_FILE" <<'PYEOF'
import json, sys
from datetime import datetime, timedelta, timezone
now = datetime.now(timezone.utc).replace(tzinfo=None)
payload = {
    "license_key": "LIC-BAD-SIG",
    "license_type": "enterprise",
    "issued_at": now.isoformat(),
    "expires_at": (now + timedelta(days=365)).isoformat(),
}
with open(sys.argv[1], "w") as f:
    json.dump({"payload": payload, "signature": "0" * 64}, f)
PYEOF
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$BAD_SIG_FILE;type=application/octet-stream"
assert_http 400 "签名失败 → 400"
assert_code 50003 "code=50003"
# 6.3 重复导入
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$LICENSE_FILE;type=application/octet-stream"
assert_http 409 "重复导入 → 409"
assert_code 50004 "code=50004"
# 6.4 过期 License
EXPIRED_FILE="$TMP_DIR/expired.lic"
generate_license "$EXPIRED_FILE" "LIC-SCRIPT-EXP-$(date +%s)" -1 100 -365
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$EXPIRED_FILE;type=application/octet-stream"
assert_http 400 "过期 License → 400"
assert_code 50002 "code=50002"
# 6.5 无文件无激活码
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN"
assert_http 400 "无文件无激活码 → 400"
assert_code 50001 "code=50001"
# 6.6 machine_code 不一致 → 50006
section "6.6 machine_code 不一致 → 50006"
http_request POST "$API/license/activate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"activation_code":"ABCD-1234-EFGH-5678","machine_code":"DUMMY-MC-NOT-MATCH"}'
assert_http 400 "machine_code 不一致 → 400"
assert_code 50006 "code=50006"
# ---------- 汇总 ----------
echo ""
echo -e "${YELLOW}============================================${NC}"
echo -e "  总计：$((PASS+FAIL)) 通过：${GREEN}${PASS}${NC} 失败：${RED}${FAIL}${NC}"
echo -e "${YELLOW}============================================${NC}"
if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 License 管理模块测试全部通过${NC}"
    exit 0
else
    echo -e "${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
