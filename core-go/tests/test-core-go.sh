#!/usr/bin/env bash
# =============================================================================
# core-go 完整测试脚本（v1.2）
#
# 用法：
#   cd /code/platform_all_use/core-go
#   chmod +x tests/test-core-go.sh
#   ./tests/test-core-go.sh
#
# 环境变量：
#   CORE_ROOT          core-go 根目录（默认脚本上级）
#   OPENSSL_BIN        铜锁 openssl 路径（优先 conf/core.yaml，其次自动探测）
#   BACKEND_URL        后端地址（默认 http://localhost:8000）
#   BACKEND_USER       后端用户名（默认 admin）
#   BACKEND_PASS       后端密码（默认 123456）
#   RUN_HTTP_TESTS     1/0/auto，默认 auto
#   KEEP_TMP           1 保留临时文件（默认 0）
#
# 退出码：0 全部通过；1 有失败
#
# v1.2 变更：
#   - libs/ 目录改为条件检查：conf 指定 tongsuo_openssl_bin 时非必需
#   - openssl version 多行输出只显示首行，并提示完整输出位置
# =============================================================================
set -uo pipefail

# -----------------------------------------------------------------------------
# 基础路径
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_ROOT="${CORE_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
BIN="$CORE_ROOT/bin/core"
CONF="$CORE_ROOT/conf/core.yaml"
WL="$CORE_ROOT/conf/algorithm-whitelist.yaml"
AUDIT_LOG="$CORE_ROOT/logs/audit.jsonl"
KEYS_DIR="$CORE_ROOT/data/keys"

BACKEND_URL="${BACKEND_URL:-http://localhost:8000}"
BACKEND_USER="${BACKEND_USER:-admin}"
BACKEND_PASS="${BACKEND_PASS:-123456}"
RUN_HTTP_TESTS="${RUN_HTTP_TESTS:-auto}"
KEEP_TMP="${KEEP_TMP:-0}"

# -----------------------------------------------------------------------------
# 颜色
# -----------------------------------------------------------------------------
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
    CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; CYAN=''; BOLD=''; NC=''
fi

PASS=0; FAIL=0; SKIP=0

pass()    { PASS=$((PASS+1)); printf "  ${GREEN}✅ PASS${NC} %s\n" "$1"; }
fail()    { FAIL=$((FAIL+1)); printf "  ${RED}❌ FAIL${NC} %s\n" "$1"
            if [ -n "${2:-}" ]; then printf "         ${RED}→${NC} %s\n" "$2"; fi; }
skip()    { SKIP=$((SKIP+1)); printf "  ${YELLOW}⏭  SKIP${NC} %s\n" "$1"; }
info()    { printf "  ${CYAN}ℹ${NC}  %s\n" "$1"; }
warn()    { printf "  ${YELLOW}⚠${NC}  %s\n" "$1"; }
section() { printf "\n${BOLD}${CYAN}=== %s ===${NC}\n" "$1"; }

# -----------------------------------------------------------------------------
# 工具
# -----------------------------------------------------------------------------
have() { command -v "$1" >/dev/null 2>&1; }

# 毫秒时间戳（Linux 用 GNU date，macOS 用 Python）
if date +%s%3N 2>/dev/null | grep -qE '^[0-9]{13}$'; then
    _ms() { date +%s%3N; }
elif have python3; then
    _ms() { python3 -c 'import time; print(int(time.time()*1000))'; }
else
    _ms() { echo "$(($(date +%s) * 1000))"; }
fi

# JSON 合法性
json_ok() {
    local f="$1"
    [ -f "$f" ] || return 1
    if have jq; then
        jq -e . "$f" >/dev/null 2>&1
    elif have python3; then
        python3 -c "import json,sys; json.load(open(sys.argv[1]))" "$f" >/dev/null 2>&1
    else
        grep -q '^{' "$f" 2>/dev/null
    fi
}

# 取 JSON 字段（点分路径）
json_get() {
    local f="$1" path="$2"
    [ -f "$f" ] || { echo ""; return; }
    if have jq; then
        jq -r "$path // empty" "$f" 2>/dev/null
    elif have python3; then
        python3 -c "
import json,sys
try:
    d=json.load(open(sys.argv[1]))
    for k in sys.argv[2].strip('.').split('.'):
        if k=='': continue
        if isinstance(d,dict): d=d.get(k,'')
        else: d=''
    print(d if d is not None else '')
except Exception: print('')
" "$f" "$path"
    fi
}

# 调用 core-go，回显退出码
run_core() {
    local op="$1" in="$2" out="$3"
    "$BIN" --op "$op" --in "$in" --out "$out" \
        >"$out.stdout" 2>"$out.stderr"
    echo $?
}

# -----------------------------------------------------------------------------
# 准备临时目录
# -----------------------------------------------------------------------------
TMP="$(mktemp -d "${TMPDIR:-/tmp}/core-go-test.XXXXXX")"
cleanup() {
    if [ "$KEEP_TMP" = "1" ]; then
        info "保留临时目录：$TMP"
    else
        rm -rf "$TMP"
    fi
}
trap cleanup EXIT

# =============================================================================
# 探测 OS / Arch / openssl 路径
# =============================================================================

# OS
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS_RAW" in
    linux*)                 PLAT_OS="linux" ;;
    darwin*)                PLAT_OS="darwin" ;;
    mingw*|msys*|cygwin*)   PLAT_OS="windows" ;;
    *)                      PLAT_OS="$OS_RAW" ;;
esac

# Arch
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64)   PLAT_ARCH="amd64" ;;
    aarch64|arm64)  PLAT_ARCH="arm64" ;;
    i386|i686)      PLAT_ARCH="386" ;;
    *)              PLAT_ARCH="$ARCH_RAW" ;;
esac

# 修复点：只在 Windows 上加 .exe
EXE_SUFFIX=""
[ "$PLAT_OS" = "windows" ] && EXE_SUFFIX=".exe"

# -----------------------------------------------------------------------------
# 修复点：读取 conf/core.yaml 的 tongsuo_openssl_bin
#
# 这个变量有两个作用：
#   1. 优先用于 resolve_openssl（与 core-go 内部优先级一致）
#   2. 决定 libs/ 目录是否必需（有明确配置时非必需）
# -----------------------------------------------------------------------------
CONFIGURED_OPENSSL=""
if [ -f "$CONF" ]; then
    CONFIGURED_OPENSSL=$(grep -E '^[[:space:]]*tongsuo_openssl_bin[[:space:]]*:' "$CONF" \
        | head -1 \
        | sed -E 's/^[^:]+:[[:space:]]*//' \
        | tr -d '"' | tr -d "'" \
        | sed -E 's/[[:space:]]+$//')
fi

# -----------------------------------------------------------------------------
# openssl 路径解析，与 core-go 内部逻辑一致
#
# 优先级：
#   1. 环境变量 OPENSSL_BIN
#   2. conf/core.yaml 的 tongsuo_openssl_bin
#   3. libs/openssl/<os>-<arch>/openssl(.exe)
#   4. 系统 PATH（fallback，提示不是铜锁）
# -----------------------------------------------------------------------------
resolve_openssl() {
    # 1. 环境变量
    if [ -n "${OPENSSL_BIN:-}" ] && [ -x "$OPENSSL_BIN" ]; then
        echo "$OPENSSL_BIN"
        return 0
    fi

    # 2. conf/core.yaml 的 tongsuo_openssl_bin
    if [ -n "$CONFIGURED_OPENSSL" ] && [ -x "$CONFIGURED_OPENSSL" ]; then
        echo "$CONFIGURED_OPENSSL"
        return 0
    fi

    # 3. 平台目录
    local p="$CORE_ROOT/libs/openssl/$PLAT_OS-$PLAT_ARCH/openssl$EXE_SUFFIX"
    if [ -x "$p" ]; then
        echo "$p"
        return 0
    fi
    # 兼容旧布局
    p="$CORE_ROOT/libs/bin/tongsuo/bin/openssl$EXE_SUFFIX"
    if [ -x "$p" ]; then
        echo "$p"
        return 0
    fi
    p="$CORE_ROOT/run/bin/openssl$EXE_SUFFIX"
    if [ -x "$p" ]; then
        echo "$p"
        return 0
    fi

    # 4. 系统 PATH 兜底
    if have openssl; then
        echo "$(command -v openssl)"
        return 0
    fi

    echo ""
    return 1
}

OPENSSL_BIN="$(resolve_openssl || true)"
OPENSSL_IS_TONGSUO=0
OPENSSL_VERSION_STR=""

if [ -n "$OPENSSL_BIN" ] && [ -x "$OPENSSL_BIN" ]; then
    OPENSSL_VERSION_STR="$("$OPENSSL_BIN" version 2>/dev/null || echo unknown)"
    case "$OPENSSL_VERSION_STR" in
        *Tongsuo*|*tongsuo*) OPENSSL_IS_TONGSUO=1 ;;
    esac
fi

# -----------------------------------------------------------------------------
# 头部
# -----------------------------------------------------------------------------
echo "=================================================="
echo "  core-go 完整测试"
echo "  CORE_ROOT = $CORE_ROOT"
echo "  BIN       = $BIN"
echo "  PLATFORM  = $PLAT_OS-$PLAT_ARCH"
echo "  OPENSSL   = ${OPENSSL_BIN:-(未找到)}"
echo "  TMP       = $TMP"
echo "=================================================="

# =============================================================================
# 测试组 1：环境与文件检查
# =============================================================================
section "测试组 1：环境与文件检查"

# 1.1 二进制
if [ -x "$BIN" ]; then
    pass "bin/core 存在且可执行"
else
    fail "bin/core 不存在或不可执行" "路径：$BIN"
fi

# 1.2 配置
[ -f "$CONF" ] && pass "conf/core.yaml 存在" || fail "conf/core.yaml 缺失"
[ -f "$WL" ]   && pass "conf/algorithm-whitelist.yaml 存在" || fail "conf/algorithm-whitelist.yaml 缺失"

# 1.3 目录
# libs/ 仅在未通过 conf 指定 openssl 时才必需
REQUIRED_DIRS="bin conf data logs tmp"
if [ -z "$CONFIGURED_OPENSSL" ]; then
    REQUIRED_DIRS="$REQUIRED_DIRS libs"
fi

for d in $REQUIRED_DIRS; do
    [ -d "$CORE_ROOT/$d" ] && pass "目录存在：$d/" || fail "目录缺失：$d/"
done

if [ -n "$CONFIGURED_OPENSSL" ]; then
    info "libs/ 非必需（openssl 已由 conf/core.yaml 指定）"
fi

# 1.4 铜锁 openssl
if [ -n "$OPENSSL_BIN" ] && [ -x "$OPENSSL_BIN" ]; then
    pass "openssl 可执行：$OPENSSL_BIN"

    # 修复点：多行输出只显示首行
    VERSION_ONE_LINE="$(printf '%s' "$OPENSSL_VERSION_STR" | head -1)"
    VERSION_LINES="$(printf '%s\n' "$OPENSSL_VERSION_STR" | wc -l | tr -d ' ')"
    info "版本：$VERSION_ONE_LINE"
    if [ "$VERSION_LINES" -gt 1 ]; then
        info "  （多行输出，已截取首行；完整输出见 \`openssl version\`）"
    fi

    if [ "$OPENSSL_IS_TONGSUO" = "1" ]; then
        pass "确认使用铜锁（Tongsuo）"
    else
        warn "未识别为铜锁（Tongsuo），当前为系统 openssl"
        warn "  → 若 conf/core.yaml 的 tongsuo_min_version 高于该版本，后续测试会因 VERSION_UNSUPPORTED 失败"
        warn "  → 生产环境请替换为铜锁 8.5.0+"
    fi
else
    fail "openssl 未找到" \
        "请在 conf/core.yaml 配置 tongsuo_openssl_bin，或放置到 libs/openssl/$PLAT_OS-$PLAT_ARCH/"
fi

# 1.5 依赖工具
for tool in jq python3; do
    have "$tool" && info "已安装：$tool" || info "未安装：$tool（部分检查会降级）"
done

# =============================================================================
# 测试组 2：CLI 基础
# =============================================================================
section "测试组 2：CLI 基础"

if "$BIN" --help >/dev/null 2>&1; then
    pass "core --help 可执行"
else
    info "core --help 未实现（不影响主流程）"
fi

"$BIN" --in /tmp/x --out /tmp/y >/dev/null 2>&1
RC=$?
[ "$RC" -ne 0 ] && pass "缺 --op 返回非零退出码（rc=$RC）" || fail "缺 --op 未返回错误"

# =============================================================================
# 测试组 3：ca.create
# =============================================================================
section "测试组 3：ca.create"

CA_REQ="$TMP/ca-req.json"
CA_RESP="$TMP/ca-resp.json"
cat > "$CA_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "ca.create",
  "request_id": "test-ca-create-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {
    "algorithm": "SM2",
    "subject": {"CN": "Test Root CA", "O": "Test Org", "C": "CN"},
    "validity_days": 365
  }
}
EOF

TS0=$(_ms)
RC=$(run_core "ca.create" "$CA_REQ" "$CA_RESP")
TS1=$(_ms)
DUR=$((TS1 - TS0))

if [ "$RC" -eq 0 ]; then
    pass "ca.create 退出码 0（耗时 ${DUR}ms）"
else
    ERR_LINE="$(head -1 "$CA_RESP.stderr" 2>/dev/null || true)"
    fail "ca.create 退出码=$RC" "$ERR_LINE"
fi

json_ok "$CA_RESP" && pass "ca.create 响应 JSON 合法" || fail "ca.create 响应非 JSON"

CODE=$(json_get "$CA_RESP" ".code")
[ "$CODE" = "OK" ] && pass "ca.create code=OK" || fail "ca.create code=$CODE"

CA_ID=$(json_get "$CA_RESP" ".data.ca_id")
CA_CERT=$(json_get "$CA_RESP" ".data.cert_path")
CA_KEY=$(json_get "$CA_RESP" ".data.key_ref")

[ -n "$CA_ID" ]   && pass "返回 ca_id=$CA_ID" || fail "未返回 ca_id"
[ -n "$CA_CERT" ] && pass "返回 cert_path=$CA_CERT" || fail "未返回 cert_path"
[ -n "$CA_KEY" ]  && pass "返回 key_ref=$CA_KEY" || fail "未返回 key_ref"

if [ -n "$CA_CERT" ] && [ -f "$CORE_ROOT/$CA_CERT" ]; then
    pass "证书文件已生成"
else
    fail "证书文件不存在" "$CORE_ROOT/$CA_CERT"
fi

if [ -n "$CA_KEY" ] && [ -f "$KEYS_DIR/${CA_KEY}.key.enc" ]; then
    pass "加密私钥已生成：${CA_KEY}.key.enc"
    if [ "$PLAT_OS" != "windows" ]; then
        MODE=$(stat -c '%a' "$KEYS_DIR/${CA_KEY}.key.enc" 2>/dev/null \
            || stat -f '%Lp' "$KEYS_DIR/${CA_KEY}.key.enc" 2>/dev/null)
        [ "$MODE" = "600" ] && pass "加密私钥权限 600" \
            || fail "加密私钥权限=$MODE（应为 600）"
    fi
else
    fail "加密私钥不存在：${CA_KEY}.key.enc"
fi

if [ -n "$OPENSSL_BIN" ] && [ -f "$CORE_ROOT/$CA_CERT" ]; then
    if "$OPENSSL_BIN" x509 -in "$CORE_ROOT/$CA_CERT" -noout -text >/dev/null 2>&1; then
        pass "证书可被 openssl 解析"
    else
        fail "证书无法解析"
    fi
fi

[ "$DUR" -le 3000 ] && pass "ca.create 耗时 ${DUR}ms ≤ 3000ms" \
    || fail "ca.create 耗时 ${DUR}ms > 3000ms"

# =============================================================================
# 测试组 4：cert.parse
# =============================================================================
section "测试组 4：cert.parse"

if [ -n "$CA_CERT" ]; then
    PARSE_REQ="$TMP/parse-req.json"
    PARSE_RESP="$TMP/parse-resp.json"
    cat > "$PARSE_REQ" <<EOF
{
  "schema_version": "1.0",
  "operation_id": "cert.parse",
  "request_id": "test-parse-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {"cert_path": "$CA_CERT"}
}
EOF
    RC=$(run_core "cert.parse" "$PARSE_REQ" "$PARSE_RESP")
    [ "$RC" -eq 0 ] && pass "cert.parse rc=0" || fail "cert.parse rc=$RC"
    json_ok "$PARSE_RESP" && pass "cert.parse 响应 JSON 合法" || fail "cert.parse 响应非 JSON"

    SUBJECT=$(json_get "$PARSE_RESP" ".data.subject")
    ISSUER=$(json_get "$PARSE_RESP" ".data.issuer")
    SERIAL=$(json_get "$PARSE_RESP" ".data.serial")
    PUBALG=$(json_get "$PARSE_RESP" ".data.public_key_algorithm")

    [ -n "$SUBJECT" ] && pass "subject=$SUBJECT" || fail "subject 为空"
    [ -n "$ISSUER" ]  && pass "issuer=$ISSUER" || fail "issuer 为空"
    [ -n "$SERIAL" ]  && pass "serial=$SERIAL" || fail "serial 为空"
    [ -n "$PUBALG" ]  && pass "public_key_algorithm=$PUBALG" || fail "pubkey alg 为空"
else
    skip "cert.parse 跳过（无可用证书）"
fi

# =============================================================================
# 测试组 5：key.manage
# =============================================================================
section "测试组 5：key.manage"

KGEN_REQ="$TMP/kg-req.json"
KGEN_RESP="$TMP/kg-resp.json"
cat > "$KGEN_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "key.manage",
  "request_id": "test-kgen-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {"action": "generate", "algorithm": "SM2"}
}
EOF
RC=$(run_core "key.manage" "$KGEN_REQ" "$KGEN_RESP")
[ "$RC" -eq 0 ] && pass "key.manage generate rc=0" || fail "key.manage generate rc=$RC"

GEN_KEYREF=$(json_get "$KGEN_RESP" ".data.key_ref")
[ -n "$GEN_KEYREF" ] && pass "生成 key_ref=$GEN_KEYREF" || fail "未返回 key_ref"

# export 无授权
KEXP_REQ="$TMP/ke-req.json"
KEXP_RESP="$TMP/ke-resp.json"
cat > "$KEXP_REQ" <<EOF
{
  "schema_version": "1.0",
  "operation_id": "key.manage",
  "request_id": "test-kexp-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {"action": "export", "key_ref": "$GEN_KEYREF"}
}
EOF
RC=$(run_core "key.manage" "$KEXP_REQ" "$KEXP_RESP")
CODE=$(json_get "$KEXP_RESP" ".code")
[ "$RC" -ne 0 ] && [ "$CODE" = "PERMISSION_DENIED" ] \
    && pass "key.manage export 无授权被拒绝（code=$CODE）" \
    || fail "key.manage export 无授权未被拒绝（rc=$RC, code=$CODE）"

# export 授权
cat > "$KEXP_REQ" <<EOF
{
  "schema_version": "1.0",
  "operation_id": "key.manage",
  "request_id": "test-kexp-002",
  "actor": {"type": "test", "id": "cli"},
  "params": {
    "action": "export",
    "key_ref": "$GEN_KEYREF",
    "export_path": "data/export/test-export.key",
    "allow_plain_export": true
  }
}
EOF
RC=$(run_core "key.manage" "$KEXP_REQ" "$KEXP_RESP")
[ "$RC" -eq 0 ] && pass "key.manage export 授权导出成功" || fail "key.manage export rc=$RC"

# delete
KDEL_REQ="$TMP/kd-req.json"
KDEL_RESP="$TMP/kd-resp.json"
cat > "$KDEL_REQ" <<EOF
{
  "schema_version": "1.0",
  "operation_id": "key.manage",
  "request_id": "test-kdel-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {"action": "delete", "key_ref": "$GEN_KEYREF"}
}
EOF
RC=$(run_core "key.manage" "$KDEL_REQ" "$KDEL_RESP")
[ "$RC" -eq 0 ] && pass "key.manage delete 成功" || fail "key.manage delete rc=$RC"

# =============================================================================
# 测试组 6：cert.convert
# =============================================================================
section "测试组 6：cert.convert"

if [ -n "$CA_CERT" ]; then
    CONV_REQ="$TMP/conv-req.json"
    CONV_RESP="$TMP/conv-resp.json"
    cat > "$CONV_REQ" <<EOF
{
  "schema_version": "1.0",
  "operation_id": "cert.convert",
  "request_id": "test-conv-001",
  "actor": {"type": "test", "id": "cli"},
  "params": {
    "source_format": "PEM",
    "target_format": "DER",
    "source_path": "$CA_CERT"
  }
}
EOF
    RC=$(run_core "cert.convert" "$CONV_REQ" "$CONV_RESP")
    [ "$RC" -eq 0 ] && pass "cert.convert PEM→DER rc=0" || fail "cert.convert rc=$RC"

    DST=$(json_get "$CONV_RESP" ".data.converted_path")
    if [ -n "$DST" ] && [ -f "$CORE_ROOT/$DST" ]; then
        pass "转换输出已生成：$DST"
    else
        fail "转换输出不存在" "$DST"
    fi
else
    skip "cert.convert 跳过（无可用证书）"
fi

# =============================================================================
# 测试组 7：错误场景
# =============================================================================
section "测试组 7：错误场景"

# 7.1 未知 operation_id
UNK_REQ="$TMP/unk-req.json"
UNK_RESP="$TMP/unk-resp.json"
cat > "$UNK_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "does.not.exist",
  "request_id": "test-unk",
  "actor": {"type": "test", "id": "cli"},
  "params": {}
}
EOF
RC=$(run_core "does.not.exist" "$UNK_REQ" "$UNK_RESP")
CODE=$(json_get "$UNK_RESP" ".code")
[ "$RC" -ne 0 ] && [ "$CODE" = "INVALID_PARAM" ] \
    && pass "未知 operation_id 被拒绝（rc=$RC, code=$CODE）" \
    || fail "未知 operation_id 未被拒绝（rc=$RC, code=$CODE）"

# 7.2 非法算法
BADALG_REQ="$TMP/badalg-req.json"
BADALG_RESP="$TMP/badalg-resp.json"
cat > "$BADALG_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "ca.create",
  "request_id": "test-badalg",
  "actor": {"type": "test", "id": "cli"},
  "params": {
    "algorithm": "DES",
    "subject": {"CN": "x"},
    "validity_days": 365
  }
}
EOF
RC=$(run_core "ca.create" "$BADALG_REQ" "$BADALG_RESP")
CODE=$(json_get "$BADALG_RESP" ".code")
[ "$CODE" = "ALGORITHM_NOT_ALLOWED" ] \
    && pass "非法算法被拒绝（code=$CODE）" \
    || fail "非法算法未被拒绝（code=$CODE）"

# 7.3 路径越界
BADPATH_REQ="$TMP/badpath-req.json"
BADPATH_RESP="$TMP/badpath-resp.json"
cat > "$BADPATH_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "cert.parse",
  "request_id": "test-badpath",
  "actor": {"type": "test", "id": "cli"},
  "params": {"cert_path": "../../../etc/passwd"}
}
EOF
RC=$(run_core "cert.parse" "$BADPATH_REQ" "$BADPATH_RESP")
CODE=$(json_get "$BADPATH_RESP" ".code")
[ "$CODE" = "PATH_NOT_ALLOWED" ] \
    && pass "路径越界被拒绝（code=$CODE）" \
    || fail "路径越界未被拒绝（code=$CODE）"

# 7.4 缺 subject
NOSUB_REQ="$TMP/nosub-req.json"
NOSUB_RESP="$TMP/nosub-resp.json"
cat > "$NOSUB_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "ca.create",
  "request_id": "test-nosub",
  "actor": {"type": "test", "id": "cli"},
  "params": {"algorithm": "SM2", "validity_days": 365}
}
EOF
RC=$(run_core "ca.create" "$NOSUB_REQ" "$NOSUB_RESP")
CODE=$(json_get "$NOSUB_RESP" ".code")
[ "$RC" -ne 0 ] && pass "缺 subject 被拒绝（rc=$RC, code=$CODE）" \
    || fail "缺 subject 未被拒绝"

# 7.5 RSA 1024 弱密钥
WEAK_REQ="$TMP/weak-req.json"
WEAK_RESP="$TMP/weak-resp.json"
cat > "$WEAK_REQ" <<'EOF'
{
  "schema_version": "1.0",
  "operation_id": "ca.create",
  "request_id": "test-weak",
  "actor": {"type": "test", "id": "cli"},
  "params": {
    "algorithm": "RSA",
    "key_params": {"key_size": 1024},
    "subject": {"CN": "weak"},
    "validity_days": 365
  }
}
EOF
RC=$(run_core "ca.create" "$WEAK_REQ" "$WEAK_RESP")
if [ "$RC" -ne 0 ]; then
    pass "RSA 1024 弱密钥被拒绝"
else
    info "RSA 1024 未被拒绝（如无要求可忽略）"
fi

# =============================================================================
# 测试组 8：审计日志
# =============================================================================
section "测试组 8：审计日志"

if [ -f "$AUDIT_LOG" ]; then
    LINES=$(wc -l < "$AUDIT_LOG" | tr -d ' ')
    pass "审计日志存在：$AUDIT_LOG（$LINES 行）"

    if have python3; then
        BAD_LINES=0
        while IFS= read -r line; do
            [ -z "$line" ] && continue
            echo "$line" | python3 -c "import json,sys; json.loads(sys.stdin.read())" \
                2>/dev/null || BAD_LINES=$((BAD_LINES+1))
        done < "$AUDIT_LOG"
        [ "$BAD_LINES" -eq 0 ] && pass "所有审计行均为合法 JSON" \
            || fail "有 $BAD_LINES 行不是合法 JSON"
    fi

    LAST=$(tail -1 "$AUDIT_LOG")
    echo "$LAST" | grep -q '"request_id"'    && pass "审计含 request_id"    || fail "审计缺 request_id"
    echo "$LAST" | grep -q '"operation_id"'  && pass "审计含 operation_id"  || fail "审计缺 operation_id"
    echo "$LAST" | grep -q '"params_digest"' && pass "审计含 params_digest" || fail "审计缺 params_digest"
    echo "$LAST" | grep -q '"result"'        && pass "审计含 result"        || fail "审计缺 result"
else
    fail "审计日志不存在" "$AUDIT_LOG"
fi

# =============================================================================
# 测试组 9：密钥存储
# =============================================================================
section "测试组 9：密钥存储"

if [ -d "$KEYS_DIR" ]; then
    ENC_COUNT=$(find "$KEYS_DIR" -name "*.key.enc" 2>/dev/null | wc -l | tr -d ' ')
    [ "$ENC_COUNT" -gt 0 ] && pass "存在 $ENC_COUNT 个加密私钥" \
        || skip "尚无加密私钥"
else
    fail "密钥目录不存在：$KEYS_DIR"
fi

if [ -f "$KEYS_DIR/master.key" ]; then
    pass "主密钥存在"
    if [ "$PLAT_OS" != "windows" ]; then
        MODE=$(stat -c '%a' "$KEYS_DIR/master.key" 2>/dev/null \
            || stat -f '%Lp' "$KEYS_DIR/master.key" 2>/dev/null)
        [ "$MODE" = "600" ] && pass "主密钥权限 600" \
            || fail "主密钥权限=$MODE（应为 600）"
    fi
    KEY_LEN=$(tr -d '\n\r ' < "$KEYS_DIR/master.key" | wc -c | tr -d ' ')
    [ "$KEY_LEN" -eq 64 ] && pass "主密钥长度 64 位 hex" \
        || fail "主密钥长度=$KEY_LEN（应为 64）"
else
    skip "主密钥尚未生成（首次运行会生成）"
fi

# =============================================================================
# 测试组 10：后端 HTTP 集成
# =============================================================================
section "测试组 10：后端 HTTP 集成"

if [ "$RUN_HTTP_TESTS" = "auto" ]; then
    if curl -sf --max-time 2 "$BACKEND_URL/health" >/dev/null 2>&1; then
        RUN_HTTP_TESTS=1
        info "检测到后端在运行：$BACKEND_URL"
    else
        RUN_HTTP_TESTS=0
        info "后端未运行，跳过 HTTP 测试"
    fi
fi

if [ "$RUN_HTTP_TESTS" = "1" ]; then
    LOGIN_RESP=$(curl -sS --max-time 5 -X POST "$BACKEND_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$BACKEND_USER\",\"password\":\"$BACKEND_PASS\"}")

    TOKEN=$(echo "$LOGIN_RESP" | python3 -c "
import json,sys
try:
    d=json.load(sys.stdin)
    print(d.get('data',{}).get('access_token',''))
except Exception: print('')
" 2>/dev/null)

    if [ -n "$TOKEN" ]; then
        pass "后端登录成功"

        API_CA_REQ="$TMP/api-ca.json"
        cat > "$API_CA_REQ" <<'EOF'
{"params":{"algorithm":"SM2","subject":{"CN":"API Test CA","O":"Test","C":"CN"},"validity_days":365}}
EOF
        API_RESP=$(curl -sS --max-time 15 -X POST \
            "$BACKEND_URL/api/v1/crypto/operations/ca.create" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            --data-binary "@$API_CA_REQ")
        echo "$API_RESP" > "$TMP/api-ca-resp.json"
        API_CODE=$(json_get "$TMP/api-ca-resp.json" ".code")

        [ "$API_CODE" = "SUCCESS" ] \
            && pass "HTTP ca.create 成功" \
            || fail "HTTP ca.create 失败（code=$API_CODE）" \
                    "$(echo "$API_RESP" | head -c 300)"

        if [ -n "$CA_CERT" ]; then
            API_PARSE_REQ="$TMP/api-parse.json"
            cat > "$API_PARSE_REQ" <<EOF
{"params":{"cert_path":"$CA_CERT"}}
EOF
            API_RESP=$(curl -sS --max-time 10 -X POST \
                "$BACKEND_URL/api/v1/crypto/operations/cert.parse" \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json" \
                --data-binary "@$API_PARSE_REQ")
            echo "$API_RESP" > "$TMP/api-parse-resp.json"
            API_CODE=$(json_get "$TMP/api-parse-resp.json" ".code")

            [ "$API_CODE" = "SUCCESS" ] \
                && pass "HTTP cert.parse 成功" \
                || fail "HTTP cert.parse 失败（code=$API_CODE）"
        fi

        HTTP_CODE=$(curl -sS -o /dev/null -w "%{http_code}" --max-time 5 -X POST \
            "$BACKEND_URL/api/v1/crypto/operations/cert.parse" \
            -H "Content-Type: application/json" \
            -d '{"params":{}}')
        [ "$HTTP_CODE" = "401" ] \
            && pass "未认证调用返回 401" \
            || fail "未认证调用返回 $HTTP_CODE（应为 401）"
    else
        fail "后端登录失败" "$(echo "$LOGIN_RESP" | head -c 200)"
    fi
else
    skip "HTTP 测试未启用"
fi

# =============================================================================
# 汇总（含根因提示）
# =============================================================================
section "测试汇总"

TOTAL=$((PASS + FAIL + SKIP))
echo ""
printf "  ${BOLD}总计：%d${NC}\n" "$TOTAL"
printf "  ${GREEN}通过：%d${NC}\n" "$PASS"
printf "  ${RED}失败：%d${NC}\n" "$FAIL"
printf "  ${YELLOW}跳过：%d${NC}\n" "$SKIP"
echo ""

EXIT_CODE=0

if [ "$FAIL" -eq 0 ]; then
    printf "  ${GREEN}${BOLD}🎉 全部通过${NC}\n\n"
else
    printf "  ${RED}${BOLD}❌ 存在 %d 项失败${NC}\n\n" "$FAIL"

    # 根因提示：如果失败集中在 VERSION_UNSUPPORTED
    if [ -f "$CA_RESP" ] && [ "$(json_get "$CA_RESP" ".code")" = "VERSION_UNSUPPORTED" ]; then
        echo -e "  ${YELLOW}${BOLD}根因提示：${NC}"
        echo -e "    当前 openssl：${OPENSSL_BIN}"
        echo -e "    版本：        $(printf '%s' "$OPENSSL_VERSION_STR" | head -1)"
        echo -e "    检测到 Tongsuo：$([ "$OPENSSL_IS_TONGSUO" = "1" ] && echo 是 || echo 否)"
        echo ""
        echo -e "    修复方式（三选一）："
        echo -e "      1. 在 conf/core.yaml 配置 tongsuo_openssl_bin 指向铜锁"
        echo -e "      2. 临时降低 conf/core.yaml 的 tongsuo_min_version"
        echo -e "      3. 编译铜锁 8.5.0+ 后放入 libs/openssl/$PLAT_OS-$PLAT_ARCH/"
        echo ""
    fi

    if [ "$KEEP_TMP" != "1" ]; then
        info "提示：加 KEEP_TMP=1 重跑可保留临时文件"
    fi

    EXIT_CODE=1
fi

exit "$EXIT_CODE"
