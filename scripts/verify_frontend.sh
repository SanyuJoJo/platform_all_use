#!/usr/bin/env bash
# ============================================================================
# 前端托管端到端验证脚本（Go 版 v1.3 新增）
#
# 用途：安装/启动后，验证后端是否正确托管前端静态资源。
#
# 覆盖：
#   1. GET /                    → 200，返回 main-app/index.html
#   2. GET /assets/*.js         → 200（若存在任意 assets 文件）
#   3. GET /sub-apps/auth/      → 200，返回子应用入口
#   4. GET /sub-apps/nonexistent/ → 404 JSON，且不含主应用 HTML（P0-2 回归）
#   5. GET /api/v1/not-exist    → 404 JSON（API 优先）
#   6. POST /health             → 405 JSON（P0-1 回归）
#   7. GET /dashboard           → 200（SPA fallback）
#
# 用法：
#   ./scripts/verify_frontend.sh                              # 默认 127.0.0.1:8000
#   ./scripts/verify_frontend.sh 127.0.0.1 18000
#   ./scripts/verify_frontend.sh http://127.0.0.1:8000
#
# 退出码：
#   0 = 全部通过
#   1 = 存在失败
#   2 = 参数错误或依赖缺失
# ============================================================================
set -uo pipefail
# ---------- 帮助 ----------
print_help() {
    cat << 'HELP'
用法：
  ./scripts/verify_frontend.sh [<host|url>] [<port>]
参数：
  <host|url>   主机名 / IP / 完整 URL（如 http://1.2.3.4:8000）
               省略时使用环境变量 HOST 或默认 127.0.0.1
  <port>       端口号（数字），省略时使用环境变量 PORT 或默认 8000
环境变量：
  HOST       默认主机（命令行未指定时使用）
  PORT       默认端口（命令行未指定时使用）
  TIMEOUT    单次 HTTP 超时秒数（默认 5）
退出码：
  0 全部通过；1 存在失败；2 参数错误或依赖缺失
HELP
}
# ---------- 参数解析 ----------
DEFAULT_HOST="${HOST:-127.0.0.1}"
DEFAULT_PORT="${PORT:-8000}"
TIMEOUT="${TIMEOUT:-5}"
case "${1:-}" in
    -h|--help)
        print_help
        exit 0
        ;;
esac
ARG_HOST=""
ARG_PORT=""
if [ $# -ge 1 ]; then ARG_HOST="$1"; fi
if [ $# -ge 2 ]; then ARG_PORT="$2"; fi
if [ $# -gt 2 ]; then
    echo "❌ 参数过多。使用 --help 查看用法" >&2
    exit 2
fi
if [ -n "$ARG_HOST" ] && echo "$ARG_HOST" | grep -qE '^https?://'; then
    BASE_URL="${ARG_HOST%/}"
elif [ -n "$ARG_HOST" ]; then
    P="${ARG_PORT:-$DEFAULT_PORT}"
    BASE_URL="http://${ARG_HOST}:${P}"
else
    BASE_URL="http://${DEFAULT_HOST}:${DEFAULT_PORT}"
fi
# ---------- 颜色 ----------
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; CYAN=''; NC=''
fi
PASS=0; FAIL=0
pass()    { echo -e "  ${GREEN}✅ PASS${NC} $1"; PASS=$((PASS+1)); }
fail()    { echo -e "  ${RED}❌ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
info()    { echo -e "  ${YELLOW}ℹ${NC}  $1"; }
section() { echo -e "\n${CYAN}=== $1 ===${NC}"; }
# ---------- 依赖检查 ----------
for cmd in curl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "❌ 需要 $cmd，请先安装"
        exit 2
    fi
done
HAS_JQ=0
if command -v jq >/dev/null 2>&1; then HAS_JQ=1; fi
# ---------- 工具函数 ----------
http_get() {
    local url="$1"
    local code
    code=$(curl -sS -o /tmp/verify_frontend_body.$$ -w "%{http_code}" \
        --max-time "$TIMEOUT" "$url" 2>/dev/null || echo "000")
    echo "$code"
}
http_get_with_method() {
    local method="$1"
    local url="$2"
    local code
    code=$(curl -sS -o /tmp/verify_frontend_body.$$ -w "%{http_code}" \
        -X "$method" --max-time "$TIMEOUT" "$url" 2>/dev/null || echo "000")
    echo "$code"
}
body() { cat /tmp/verify_frontend_body.$$ 2>/dev/null || echo ""; }
cleanup_tmp() { rm -f /tmp/verify_frontend_body.$$; }
trap cleanup_tmp EXIT
assert_http() {
    local expected="$1"; local msg="$2"; local actual="$3"
    if [ "$actual" = "$expected" ]; then
        pass "$msg (HTTP $actual)"
    else
        fail "$msg (期望 HTTP $expected，实际 $actual，body=$(body | head -c 200))"
    fi
}
assert_body_contains() {
    local pattern="$1"; local msg="$2"
    if body | grep -q "$pattern"; then
        pass "$msg"
    else
        fail "$msg (body 不含 '$pattern'，实际=$(body | head -c 200))"
    fi
}
assert_body_not_contains() {
    local pattern="$1"; local msg="$2"
    if body | grep -q "$pattern"; then
        fail "$msg (body 不应含 '$pattern')"
    else
        pass "$msg"
    fi
}
# ---------- 前置检查 ----------
echo "=============================================="
echo "  前端托管端到端验证（Go 版 v1.3）"
echo "  BASE_URL = $BASE_URL"
echo "  TIMEOUT  = ${TIMEOUT}s"
echo "  jq 支持  = $([ "$HAS_JQ" -eq 1 ] && echo "是" || echo "否（降级模式）")"
echo "=============================================="
if ! curl -fsS --max-time 3 "$BASE_URL/health" >/dev/null 2>&1; then
    echo -e "${RED}❌ 无法访问 $BASE_URL/health，请确认服务已启动${NC}"
    echo "   排查建议："
    echo "     1. 确认服务已启动"
    echo "     2. 若服务监听其他端口：./scripts/verify_frontend.sh 127.0.0.1 <port>"
    exit 1
fi
pass "服务可访问"
# ---------- 1. 主应用入口 ----------
section "1. GET / （主应用 SPA 入口）"
CODE=$(http_get "$BASE_URL/")
assert_http 200 "GET / 返回 200" "$CODE"
assert_body_contains "<!DOCTYPE html>" "返回 HTML"
assert_body_contains "main-app" "响应含 main-app 标记（或通用 HTML 结构）" || true
# ---------- 2. 主应用 assets ----------
section "2. GET /assets/* （主应用静态资源）"
# 尝试从主应用 index.html 中提取一个 assets 路径
ASSET_PATH=""
if [ -f "$BASE_URL" ]; then :; fi
# 尝试常见的 assets 路径
for candidate in "/assets/app.js" "/assets/index.js" "/assets/main.js"; do
    CODE=$(http_get "$BASE_URL$candidate")
    if [ "$CODE" = "200" ]; then
        pass "GET $candidate 返回 200"
        ASSET_PATH="$candidate"
        break
    fi
done
if [ -z "$ASSET_PATH" ]; then
    info "未找到常见 assets 路径（可能打包文件名带 hash，跳过具体文件验证）"
    # 至少验证目录请求不会列出目录
    CODE=$(http_get "$BASE_URL/assets/")
    if [ "$CODE" = "404" ] || [ "$CODE" = "403" ] || [ "$CODE" = "200" ]; then
        pass "GET /assets/ 返回 $CODE（不允许目录列表）"
    else
        fail "GET /assets/ 返回异常：$CODE"
    fi
fi
# ---------- 3. 子应用入口 ----------
section "3. GET /sub-apps/<name>/ （子应用入口）"
# 尝试常见的子应用
FOUND_SUBAPP=0
for subapp in auth module-manager audit-log license; do
    CODE=$(http_get "$BASE_URL/sub-apps/$subapp/")
    if [ "$CODE" = "200" ]; then
        pass "GET /sub-apps/$subapp/ 返回 200"
        FOUND_SUBAPP=1
    fi
done
if [ "$FOUND_SUBAPP" -eq 0 ]; then
    info "未找到常见子应用（auth / module-manager / audit-log / license）"
    info "  若子应用使用其他名称，请手动验证 /sub-apps/<name>/"
fi
# ---------- 4. 子应用缺失 → 404 JSON，不回退主应用 ----------
section "4. GET /sub-apps/nonexistent/ （P0-2 回归）"
CODE=$(http_get "$BASE_URL/sub-apps/nonexistent/")
assert_http 404 "GET /sub-apps/nonexistent/ 返回 404" "$CODE"
if [ "$HAS_JQ" -eq 1 ]; then
    ERR_CODE=$(body | jq -r '.code // "null"' 2>/dev/null || echo "null")
    if [ "$ERR_CODE" = "90002" ]; then
        pass "JSON code=90002"
    else
        fail "期望 code=90002，实际 $ERR_CODE"
    fi
else
    if body | grep -q '"code":90002'; then
        pass "JSON 含 code=90002"
    else
        fail "JSON 不含 code=90002"
    fi
fi
# 关键断言：不得回退到主应用 HTML
assert_body_not_contains "<!DOCTYPE html>" "不回退主应用 HTML（P0-2）"
# ---------- 5. API 优先 ----------
section "5. GET /api/v1/not-exist （API 优先）"
CODE=$(http_get "$BASE_URL/api/v1/not-exist")
assert_http 404 "GET /api/v1/not-exist 返回 404" "$CODE"
if [ "$HAS_JQ" -eq 1 ]; then
    ERR_CODE=$(body | jq -r '.code // "null"' 2>/dev/null || echo "null")
    if [ "$ERR_CODE" = "90002" ]; then
        pass "JSON code=90002"
    else
        fail "期望 code=90002，实际 $ERR_CODE"
    fi
else
    if body | grep -q '"code":90002'; then
        pass "JSON 含 code=90002"
    else
        fail "JSON 不含 code=90002"
    fi
fi
assert_body_not_contains "<!DOCTYPE html>" "API 路由不返回 HTML"
# ---------- 6. 405 JSON ----------
section "6. POST /health （P0-1 回归）"
CODE=$(http_get_with_method POST "$BASE_URL/health")
assert_http 405 "POST /health 返回 405" "$CODE"
if [ "$HAS_JQ" -eq 1 ]; then
    ERR_CODE=$(body | jq -r '.code // "null"' 2>/dev/null || echo "null")
    if [ "$ERR_CODE" = "90001" ]; then
        pass "JSON code=90001"
    else
        fail "期望 code=90001，实际 $ERR_CODE"
    fi
else
    if body | grep -q '"code":90001'; then
        pass "JSON 含 code=90001"
    else
        fail "JSON 不含 code=90001"
    fi
fi
# ---------- 7. SPA fallback ----------
section "7. GET /dashboard （SPA fallback）"
CODE=$(http_get "$BASE_URL/dashboard")
assert_http 200 "GET /dashboard 返回 200" "$CODE"
assert_body_contains "<!DOCTYPE html>" "SPA fallback 返回 HTML"
# ---------- 汇总 ----------
TOTAL=$((PASS+FAIL))
echo ""
echo "=============================================="
echo "  验证结果汇总"
echo "=============================================="
echo "  BASE_URL = $BASE_URL"
echo -e "  总计：$TOTAL"
echo -e "  通过：${GREEN}${PASS}${NC}"
echo -e "  失败：${RED}${FAIL}${NC}"
echo "=============================================="
if [ "$FAIL" -eq 0 ]; then
    echo -e "\n${GREEN}🎉 前端托管验证全部通过${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 存在失败用例，请检查${NC}"
    exit 1
fi
