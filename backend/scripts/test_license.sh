#!/usr/bin/env bash
# ============================================================
# License 管理模块 API 测试脚本（v1.2.2）
# 用法：./scripts/test_license.sh [BASE_URL]
# 默认：http://localhost:8000
#
# 前置条件：
#   1. 已执行数据库迁移（uv run alembic upgrade head）
#   2. 服务已启动
#   3. 环境变量 LICENSE_SECRET_KEY 已设置，且与服务端一致
#      （若未设置，脚本将提示并继续使用默认密钥）
#
# 依赖：curl、jq、python3
#
# v1.2.2 Bug 修复：
#   - 修复 6.4 "过期 License" 场景误报 50001 的问题。
#     根因：generate_license 生成过期 License 时 issued_at=now、
#           expires_at=now-1d，触发后端"issued_at 必须早于 expires_at"
#           校验（50001），永远走不到"过期"检查（50002）。
#     修复：6.4 场景 issued_offset_days 从 0 改为 -365，
#           使 issued_at < expires_at < now。
#     附加：generate_license 增加时间一致性守卫，检测到
#           issued_at >= expires_at 时打印 WARNING（不阻断），
#           便于提前发现同类错误。
#
# v1.2.1 Bug 修复：
#   - 修复 `YELLOW: unbound variable` 报错。
#     原因：脚本使用 `set -u`，但颜色变量在第一次使用之后才定义。
#     修复：将颜色变量定义移至 `set -uo pipefail` 之后，脚本顶部。
#
# v1.2 变更（R-2）：
#   - 拆分"未配置激活服务"测试为 6.6（50005）和 6.7（50006）两个独立场景。
#   - 6.6 需 LICENSE_MACHINE_CODE_OVERRIDE 与请求 machine_code 一致，
#     才触发 50005；未设置该环境变量则跳过并提示。
#
# v1.1 变更：
#   - P1-8：从环境变量 LICENSE_SECRET_KEY 读取签名密钥。
#   - 新增 P0-5 场景验证（max_users=True / issued_at >= expires_at）。
# ============================================================
set -uo pipefail

# ============================================================
# 颜色变量定义前移，确保在任何使用之前已初始化（v1.2.1 修复）
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

BASE_URL="${1:-http://localhost:8000}"
API="${BASE_URL}/api/v1"

for cmd in curl jq python3; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "❌ 需要 $cmd，请先安装"
        exit 1
    fi
done

# ---------- P1-8：从环境变量读取密钥 ----------
if [ -z "${LICENSE_SECRET_KEY:-}" ]; then
    echo -e "${YELLOW}⚠  环境变量 LICENSE_SECRET_KEY 未设置。${NC}"
    echo -e "    脚本将使用默认密钥 'change-this-license-secret-in-production'。"
    echo -e "    若服务端已配置其他密钥，请先执行："
    echo -e "        export LICENSE_SECRET_KEY=<你的密钥>"
    echo ""
    LICENSE_SECRET_KEY="change-this-license-secret-in-production"
fi
echo "使用 LICENSE_SECRET_KEY = ${LICENSE_SECRET_KEY:0:8}...（前 8 位）"

PASS=0
FAIL=0
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
    if [ "$HTTP_CODE" = "$1" ]; then pass "$2 (HTTP $HTTP_CODE)"; else fail "$2 (expected HTTP $1, got $HTTP_CODE, body=$HTTP_BODY)"; fi
}
assert_code() {
    local actual; actual=$(echo "$HTTP_BODY" | jq -r '.code // "null"')
    if [ "$actual" = "$1" ]; then pass "$2 (code=$actual)"; else fail "$2 (expected code=$1, got $actual, body=$HTTP_BODY)"; fi
}

# ---------------------------------------------------------------------------
# 生成合法 License 文件
#
# 参数：
#   $1 out_file              输出文件路径
#   $2 license_key           License 密钥（默认自动生成）
#   $3 expires_days          过期时间相对 now 的天数偏移（默认 365）
#   $4 max_users             最大用户数（默认 100）
#   $5 issued_offset_days    issued_at 相对 now 的天数偏移（默认 0）
#
# v1.2.2：增加时间一致性守卫。
#   若 issued_at >= expires_at，打印 WARNING 提示调用方检查参数。
#   不阻断执行，因为 7.2 场景需要故意构造 issued_at >= expires_at 的非法数据。
# ---------------------------------------------------------------------------
generate_license() {
    local out_file="$1"
    local license_key="${2:-LIC-SCRIPT-$(date +%s)}"
    local expires_days="${3:-365}"
    local max_users="${4:-100}"
    local issued_offset_days="${5:-0}"

    # v1.2.2：时间一致性守卫
    if [ "$issued_offset_days" -ge "$expires_days" ]; then
        echo -e "  ${YELLOW}⚠${NC}  generate_license: issued_offset_days=$issued_offset_days >= expires_days=$expires_days" >&2
        echo -e "      生成的 License 将触发 issued_at >= expires_at 校验（50001）。" >&2
        echo -e "      若非故意构造该场景，请调整参数。" >&2
    fi

    LICENSE_SECRET_KEY="$LICENSE_SECRET_KEY" python3 - \
        "$out_file" "$license_key" "$expires_days" "$max_users" "$issued_offset_days" <<'PYEOF'
import hashlib
import hmac
import json
import os
import sys
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

generate_license_with_signature() {
    local out_file="$1"
    local signature="$2"
    local license_key="${3:-LIC-SIG-$(date +%s)}"
    python3 - "$out_file" "$license_key" "$signature" <<'PYEOF'
import json, sys
from datetime import datetime, timedelta, timezone

out_file, license_key, signature = sys.argv[1:4]
now = datetime.now(timezone.utc).replace(tzinfo=None)
payload = {
    "license_key": license_key,
    "license_type": "enterprise",
    "max_users": 100,
    "authorized_modules": ["auth"],
    "machine_code": None,
    "issued_at": now.isoformat(),
    "expires_at": (now + timedelta(days=365)).isoformat(),
}
with open(out_file, "w", encoding="utf-8") as f:
    json.dump({"payload": payload, "signature": signature}, f, ensure_ascii=False)
PYEOF
}

generate_license_with_max_users_bool() {
    local out_file="$1"
    local license_key="${2:-LIC-BOOL-$(date +%s)}"
    LICENSE_SECRET_KEY="$LICENSE_SECRET_KEY" python3 - \
        "$out_file" "$license_key" <<'PYEOF'
import hashlib, hmac, json, os, sys
from datetime import datetime, timedelta, timezone

out_file, license_key = sys.argv[1:3]
secret = os.environ["LICENSE_SECRET_KEY"]
def canonical(payload):
    return json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
now = datetime.now(timezone.utc).replace(tzinfo=None)
payload = {
    "license_key": license_key,
    "license_type": "enterprise",
    "max_users": True,
    "authorized_modules": ["auth"],
    "machine_code": None,
    "issued_at": now.isoformat(),
    "expires_at": (now + timedelta(days=365)).isoformat(),
}
signature = hmac.new(secret.encode(), canonical(payload).encode(), hashlib.sha256).hexdigest()
with open(out_file, "w", encoding="utf-8") as f:
    json.dump({"payload": payload, "signature": signature}, f, ensure_ascii=False)
PYEOF
}

echo -e "${YELLOW}License 管理模块 API 测试（v1.2.2）${NC}"
echo "BASE_URL = $BASE_URL"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

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

http_request POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"guest","password":"123456"}'
GUEST_TOKEN=$(echo "$HTTP_BODY" | jq -r '.data.access_token')
http_request GET "$API/license/status" \
    -H "Authorization: Bearer $GUEST_TOKEN"
assert_http 403 "guest 查询状态 → 403"
assert_code 20051 "code=20051"

# ---------- 初始状态 ----------
section "2. 初始状态（未导入 License）"
http_request GET "$API/license/status" \
    -H "Authorization: Bearer $TOKEN"
if [ "$HTTP_CODE" = "404" ]; then
    assert_code 50009 "无 License → code=50009"
elif [ "$HTTP_CODE" = "200" ]; then
    pass "已存在 License，跳过初始状态验证"
fi

# ---------- 导入合法 License ----------
section "3. 导入合法 License"
LICENSE_FILE="$TMP_DIR/valid.lic"
# expires=365 天后，issued=now，max_users=100，issued_offset=0
generate_license "$LICENSE_FILE" "LIC-SCRIPT-VALID-$(date +%s)" 365 100 0

http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$LICENSE_FILE;type=application/octet-stream"
assert_http 200 "导入成功"
assert_code 0 "code=0"
EXPIRES=$(echo "$HTTP_BODY" | jq -r '.data.expires_at')
pass "License 过期时间：$EXPIRES"

# ---------- 状态查询 ----------
section "4. 状态查询"
http_request GET "$API/license/status" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "状态查询成功"
assert_code 0 "code=0"
IS_VALID=$(echo "$HTTP_BODY" | jq -r '.data.is_valid')
IS_EXPIRED=$(echo "$HTTP_BODY" | jq -r '.data.is_expired')
IS_EXPIRING=$(echo "$HTTP_BODY" | jq -r '.data.is_expiring_soon')
DAYS=$(echo "$HTTP_BODY" | jq -r '.data.days_remaining')
pass "is_valid=$IS_VALID is_expired=$IS_EXPIRED is_expiring_soon=$IS_EXPIRING days_remaining=$DAYS"

# ---------- 模块授权 ----------
section "5. 模块授权状态"
http_request GET "$API/license/modules" \
    -H "Authorization: Bearer $TOKEN"
assert_http 200 "模块授权查询成功"
assert_code 0 "code=0"
MOD_COUNT=$(echo "$HTTP_BODY" | jq -r '.data | length')
pass "模块总数：$MOD_COUNT"

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
generate_license_with_signature "$BAD_SIG_FILE" "0000000000000000000000000000000000000000000000000000000000000000"
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

# ---------------------------------------------------------------------------
# 6.4 过期 License
#
# v1.2.2 修复：
#   原参数 -1 100 0 导致：
#       issued_at  = now + 0    = now
#       expires_at = now + (-1) = now - 1d
#       → issued_at > expires_at → 触发 50001（时间逻辑错误），而非 50002。
#
#   新参数 -1 100 -365：
#       issued_at  = now - 365d
#       expires_at = now - 1d
#       → issued_at < expires_at < now → 正确触发 50002（已过期）。
# ---------------------------------------------------------------------------
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

# 6.6 未配置激活服务 → 50005
# 需要 LICENSE_MACHINE_CODE_OVERRIDE 与请求 machine_code 一致，
# 才能通过前置校验（P0-3），进而触发"未配置激活服务"。
# 若环境未设置该变量，则跳过并提示。
section "6.6 未配置激活服务（需 LICENSE_MACHINE_CODE_OVERRIDE）"
if [ -n "${LICENSE_MACHINE_CODE_OVERRIDE:-}" ]; then
    http_request POST "$API/license/activate" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"activation_code\":\"ABCD-1234-EFGH-5678\",\"machine_code\":\"$LICENSE_MACHINE_CODE_OVERRIDE\"}"
    assert_http 400 "未配置激活服务 → 400"
    assert_code 50005 "code=50005"
else
    echo -e "  ${YELLOW}⚠${NC}  未设置 LICENSE_MACHINE_CODE_OVERRIDE 环境变量，跳过 50005 验证。"
    echo -e "      若要验证此场景，请："
    echo -e "      1) 服务端设置 LICENSE_MACHINE_CODE_OVERRIDE=<固定值> 并重启服务"
    echo -e "      2) 测试端 export LICENSE_MACHINE_CODE_OVERRIDE=<同一个固定值>"
    echo -e "      3) 重新运行本脚本"
fi

# 6.7 machine_code 不一致 → 50006
# 无论是否设置 LICENSE_MACHINE_CODE_OVERRIDE，请求 machine_code 与当前值不同
# 都会在前置校验阶段（P0-3）直接返回 50006。
section "6.7 machine_code 不一致 → 50006"
http_request POST "$API/license/activate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"activation_code":"ABCD-1234-EFGH-5678","machine_code":"DUMMY-MC-NOT-MATCH"}'
assert_http 400 "machine_code 不一致 → 400"
assert_code 50006 "code=50006"

# ---------- P0-5 数据校验 ----------
section "7. P0-5 数据校验"

# 7.1 max_users=True（bool 被拒绝）
BOOL_FILE="$TMP_DIR/bool.lic"
generate_license_with_max_users_bool "$BOOL_FILE" "LIC-BOOL-$(date +%s)"
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$BOOL_FILE;type=application/octet-stream"
assert_http 400 "max_users=True → 400"
assert_code 50001 "code=50001"

# ---------------------------------------------------------------------------
# 7.2 issued_at >= expires_at（故意构造非法时间逻辑）
# expires = now + 5d，issued = now + 10d → issued_at > expires_at
# generate_license 会打印 WARNING（因 issued_offset_days >= expires_days），
# 这是预期行为。
# ---------------------------------------------------------------------------
REVERSED_FILE="$TMP_DIR/reversed.lic"
generate_license "$REVERSED_FILE" "LIC-REV-$(date +%s)" 5 100 10
http_request POST "$API/license/import" \
    -H "Authorization: Bearer $TOKEN" \
    -F "license_file=@$REVERSED_FILE;type=application/octet-stream"
assert_http 400 "issued_at >= expires_at → 400"
assert_code 50001 "code=50001"

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
