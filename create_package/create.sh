#!/usr/bin/env bash
# ============================================================
# 产品安装包创建脚本（Go 版 v1.1）
# 位置：create_package/create.sh
# 用法：./create.sh   或   bash /abs/path/create.sh
#
# v1.1 修复：
#   - P1-01：增加 bin/goose、DEPLOY.txt 缺失警告（不阻断）。
#   - P2-02：bin/server 改用 -x（可执行）校验，缺失立即 fail。
#   - P2-05：若 file 命令可用，打印 bin/server 架构信息（仅提示）。
#
# 校验 backend.tar.gz 内部结构（Go 版）：
#   - bin/server（必须可执行）
#   - migrations/
#   - .env.production.example
#   - bin/goose（推荐，缺失时警告）
#   - DEPLOY.txt（推荐，缺失时警告）
# ============================================================
set -euo pipefail
_SELF="${BASH_SOURCE[0]}"
if [ -L "$_SELF" ]; then
  _SELF="$(readlink -f "$_SELF" 2>/dev/null || echo "$_SELF")"
fi
if command -v realpath >/dev/null 2>&1; then
  _SELF="$(realpath "$_SELF")"
fi
SCRIPT_DIR="$(cd "$(dirname "$_SELF")" && pwd)"
unset _SELF
BIN_DIR="$SCRIPT_DIR/bin"
PAYLOAD_DIR="$BIN_DIR/bin"
TEMPLATES_DIR="$BIN_DIR/templates"
OUT_DIR="$SCRIPT_DIR/out"
STAGING_ROOT="$SCRIPT_DIR/.staging"
VERSION_FILE="$SCRIPT_DIR/VERSION"
log()  { echo -e "[create] $*"; }
warn() { echo -e "[create][WARN] $*" >&2; }
fail() { echo -e "[create][ERROR] $*" >&2; exit 1; }
# ---------- 版本解析 ----------
if [ -n "${VERSION:-}" ]; then
  VERSION_SOURCE="环境变量 VERSION"
elif [ -f "$VERSION_FILE" ]; then
  VERSION="$(cat "$VERSION_FILE" | tr -d '[:space:]')"
  VERSION_SOURCE="$VERSION_FILE"
  [ -n "$VERSION" ] || fail "VERSION 文件为空：$VERSION_FILE"
else
  VERSION="1.0.0"
  VERSION_SOURCE="默认值"
fi
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.]+)?$'; then
  fail "VERSION 格式无效：$VERSION（应为 x.y.z 或 x.y.z-suffix）"
fi
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
PACKAGE_NAME="product-installer-${VERSION}-${TIMESTAMP}"
STAGING_DIR="$STAGING_ROOT/$PACKAGE_NAME"
ARCHIVE_PATH="$OUT_DIR/${PACKAGE_NAME}.tar.gz"
SHA256_PATH="$OUT_DIR/${PACKAGE_NAME}.tar.gz.sha256"
command -v tar >/dev/null 2>&1 || fail "缺少命令：tar"
log "开始创建产品安装包：$PACKAGE_NAME"
log "版本来源：$VERSION_SOURCE（version=$VERSION）"
log "脚本目录：$SCRIPT_DIR"
[ -f "$BIN_DIR/install.sh" ] || fail "缺少安装脚本：$BIN_DIR/install.sh"
[ -f "$PAYLOAD_DIR/backend.tar.gz" ] || fail "缺少后端包：$PAYLOAD_DIR/backend.tar.gz"
[ -f "$PAYLOAD_DIR/dist.war" ] || fail "缺少前端包：$PAYLOAD_DIR/dist.war"
[ -d "$TEMPLATES_DIR" ] || fail "缺少模板目录：$TEMPLATES_DIR"
# ============================================================================
# 后端载荷结构校验（Go 版，v1.1 增强）
# ============================================================================
log "校验 backend.tar.gz 内部结构..."
TMP_BACKEND="$(mktemp -d)"
TMP_FRONTEND="$(mktemp -d)"
cleanup_tmp() { rm -rf "$TMP_BACKEND" "$TMP_FRONTEND"; }
trap cleanup_tmp EXIT
if ! tar -xzf "$PAYLOAD_DIR/backend.tar.gz" -C "$TMP_BACKEND" 2>/dev/null; then
  fail "backend.tar.gz 不是合法 tar.gz"
fi
# --- 必须项：bin/server（必须可执行，P2-02）---
if [ ! -x "$TMP_BACKEND/bin/server" ]; then
  if [ ! -f "$TMP_BACKEND/bin/server" ]; then
    fail "backend.tar.gz 缺少 bin/server（请检查构建脚本）"
  else
    fail "backend.tar.gz 中 bin/server 不可执行（请检查构建脚本的 chmod +x）"
  fi
fi
# --- 必须项：migrations/ ---
[ -d "$TMP_BACKEND/migrations" ] || {
  echo "[create] backend.tar.gz 顶层内容如下："
  (cd "$TMP_BACKEND" && ls -A | head -20)
  fail "backend.tar.gz 顶层结构异常（应直接包含 bin/、migrations/）"
}
# --- 必须项：.env.production.example ---
[ -f "$TMP_BACKEND/.env.production.example" ] \
  || fail "backend.tar.gz 缺少 .env.production.example"
# --- 推荐项：bin/goose（P1-01，缺失时警告不阻断）---
if [ ! -x "$TMP_BACKEND/bin/goose" ]; then
  warn "backend.tar.gz 未包含 bin/goose（或不可执行）"
  warn "  → install.sh 将尝试从 PATH 中查找 goose；若 PATH 中也无，数据库迁移会失败"
  warn "  → 建议构建时使用：./scripts/build.sh --flat --with-goose"
else
  log "✅ 已包含 bin/goose"
fi
# --- 推荐项：DEPLOY.txt（P1-01，缺失时警告不阻断）---
if [ ! -f "$TMP_BACKEND/DEPLOY.txt" ]; then
  warn "backend.tar.gz 未包含 DEPLOY.txt（建议补充，用于目标机部署说明）"
else
  log "✅ 已包含 DEPLOY.txt"
fi
# --- 平台架构提示（P2-05，仅提示不阻断）---
if command -v file >/dev/null 2>&1; then
  FILE_OUT="$(file "$TMP_BACKEND/bin/server" 2>/dev/null || true)"
  if [ -n "$FILE_OUT" ]; then
    log "后端二进制：$FILE_OUT"
    log "  → 请确认目标机与二进制架构一致（例如 Linux x86_64）"
  fi
fi
# ============================================================================
# 前端载荷结构校验
# ============================================================================
log "校验 dist.war 内部结构..."
if command -v unzip >/dev/null 2>&1 && unzip -tq "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  unzip -q "$PAYLOAD_DIR/dist.war" -d "$TMP_FRONTEND" 2>/dev/null \
    || fail "dist.war 解压失败"
elif tar -tzf "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  log "dist.war 非 zip 格式，尝试以 tar.gz 解压"
  tar -xzf "$PAYLOAD_DIR/dist.war" -C "$TMP_FRONTEND" 2>/dev/null \
    || fail "dist.war 解压失败"
else
  fail "dist.war 既不是合法 zip 也不是合法 tar.gz"
fi
[ -f "$TMP_FRONTEND/main-app/index.html" ] \
  || fail "dist.war 缺少 main-app/index.html"
[ -d "$TMP_FRONTEND/sub-apps" ] \
  || fail "dist.war 缺少 sub-apps/ 目录"
# ============================================================================
# 复制到 staging
# ============================================================================
rm -rf "$STAGING_DIR"
mkdir -p "$STAGING_DIR/bin" "$STAGING_DIR/templates" "$OUT_DIR"
log "复制安装脚本与载荷..."
cp "$BIN_DIR/install.sh" "$STAGING_DIR/install.sh"
cp "$PAYLOAD_DIR/backend.tar.gz" "$STAGING_DIR/bin/backend.tar.gz"
cp "$PAYLOAD_DIR/dist.war" "$STAGING_DIR/bin/dist.war"
cp -R "$TEMPLATES_DIR/." "$STAGING_DIR/templates/"
chmod +x "$STAGING_DIR/install.sh"
find "$STAGING_DIR/templates" -type f -name "*.sh" -exec chmod +x {} \;
echo "$VERSION" > "$STAGING_DIR/VERSION"
# ============================================================================
# SHA256（打包前先算，写入 manifest）
# ============================================================================
if command -v sha256sum >/dev/null 2>&1; then
  BACKEND_SHA="$(sha256sum "$PAYLOAD_DIR/backend.tar.gz" | awk '{print $1}')"
  FRONTEND_SHA="$(sha256sum "$PAYLOAD_DIR/dist.war" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  BACKEND_SHA="$(shasum -a 256 "$PAYLOAD_DIR/backend.tar.gz" | awk '{print $1}')"
  FRONTEND_SHA="$(shasum -a 256 "$PAYLOAD_DIR/dist.war" | awk '{print $1}')"
else
  fail "缺少 sha256sum 或 shasum"
fi
BACKEND_SIZE="$(stat -c%s "$PAYLOAD_DIR/backend.tar.gz" 2>/dev/null || stat -f%z "$PAYLOAD_DIR/backend.tar.gz")"
FRONTEND_SIZE="$(stat -c%s "$PAYLOAD_DIR/dist.war" 2>/dev/null || stat -f%z "$PAYLOAD_DIR/dist.war")"
cat > "$STAGING_DIR/manifest.json" <<EOF
{
  "name": "platform-product-installer",
  "version": "${VERSION}",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "files": {
    "backend":  { "path": "bin/backend.tar.gz", "size": ${BACKEND_SIZE},  "sha256": "${BACKEND_SHA}" },
    "frontend": { "path": "bin/dist.war",       "size": ${FRONTEND_SIZE}, "sha256": "${FRONTEND_SHA}" }
  },
  "default_install_dir": "/opt/platform",
  "default_port": 8000
}
EOF
# ============================================================================
# 打包
# ============================================================================
log "打包安装包..."
tar -czf "$ARCHIVE_PATH" -C "$STAGING_ROOT" "$PACKAGE_NAME"
if command -v sha256sum >/dev/null 2>&1; then
  SHA256="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
else
  fail "缺少 sha256sum 或 shasum"
fi
echo "$SHA256  $(basename "$ARCHIVE_PATH")" > "$SHA256_PATH"
rm -rf "$STAGING_DIR"
log "✅ 安装包创建完成"
log "   安装包：$ARCHIVE_PATH"
log "   SHA256：$SHA256_PATH"
log "   大小：$(du -h "$ARCHIVE_PATH" | cut -f1)"
