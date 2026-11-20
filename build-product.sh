#!/usr/bin/env bash
# ============================================================================
# 一键生成产品安装包：后端 → 前端 → 载荷校验 → 调用 create.sh 打包
# 位置：项目根目录 build-product.sh
#
# v1.2 修复：
#   - N-02：后端校验改用 -x（可执行），与 create.sh 一致；
#           新增 bin/goose、DEPLOY.txt 缺失警告（不阻断）。
# v1.3 修复：
#   - P1-FE-01：后端 [1/4] 阶段前置 go test ./internal/frontend/... + go vet ./...
#   - P2-FE-05：前端 [2/4] 阶段前置检测 frontend/deploy/main-app/index.html
#   - P2-FE-08：dist.war 打入 deploy-manifest.json（若存在）
#   - P1-FE-03：前端 [3/4] 阶段校验增强（main-app/assets + 每个子应用）
# v1.4 变更（安装包含 core）：
#   - 阶段扩展为 [1/5]~[5/5]
#   - 新增 [3/5] 构建 core 载荷（build-core-payload.sh → core.tar.gz）
#   - [4/5] 载荷校验增加 core.tar.gz 结构校验
# v1.5 变更（安装包含音视频）：
#   - 阶段扩展为 [1/6]~[6/6]
#   - 新增 [4/6] 构建 ZLM 载荷（build-zlm-payload.sh → video.tar.gz）
#   - [5/6] 载荷校验增加 video.tar.gz 结构校验
# ============================================================================
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PAYLOAD_DIR="$ROOT/create_package/bin/bin"
CREATE_SH="$ROOT/create_package/create.sh"
log() { echo -e "\n===== $* ====="; }
warn() { echo -e "  ⚠️  $*"; }
die() { echo -e "\n❌ $*" >&2; exit 1; }
# ============================================================================
# [1/5] 后端：前置校验 + 构建
# ============================================================================
log "[1/6] 构建后端 Flat 包（含 goose）"
cd "$ROOT/backend-go"
# ---- P1-FE-01：前置校验前端托管集成 ----
echo "  → 校验 cmd/server/main.go 前端托管集成..."
if ! grep -q "backend-go/internal/frontend" cmd/server/main.go; then
  die "cmd/server/main.go 未导入 internal/frontend（请同步《开发与维护指南 v1.1》）"
fi
if ! grep -q "frontend.NewHandler" cmd/server/main.go; then
  die "cmd/server/main.go 未调用 frontend.NewHandler（请同步《开发与维护指南 v1.1》）"
fi
if ! grep -q "frontendHandler.Register" cmd/server/main.go; then
  die "cmd/server/main.go 未调用 frontendHandler.Register（请同步《开发与维护指南 v1.1》）"
fi
if ! grep -q "r.NoMethod" cmd/server/main.go; then
  die "cmd/server/main.go 未注册 r.NoMethod（P0-1 修复未同步）"
fi
echo "  ✅ main.go 前端托管集成检查通过"
echo "  → 运行 internal/frontend 单元测试..."
if ! go test ./internal/frontend/... >/dev/null 2>&1; then
  echo "  ❌ go test ./internal/frontend/... 失败"
  go test ./internal/frontend/... || true
  die "前端托管单元测试未通过"
fi
echo "  ✅ go test ./internal/frontend/... 通过"
echo "  → go vet ./..."
if ! go vet ./... >/dev/null 2>&1; then
  echo "  ❌ go vet ./... 失败"
  go vet ./... || true
  die "go vet 未通过"
fi
echo "  ✅ go vet 通过"
# ---- 后端构建 ----
rm -f dist/backend-flat-*.tar.gz dist/backend-flat-*.manifest.json
./scripts/build.sh --flat --with-goose
LATEST_FLAT="$(ls -t dist/backend-flat-*.tar.gz 2>/dev/null | head -1 || true)"
[ -n "$LATEST_FLAT" ] || die "未找到 backend-flat-*.tar.gz"
echo "本次 Flat 包：$LATEST_FLAT"
mkdir -p "$PAYLOAD_DIR"
rm -f "$PAYLOAD_DIR/backend.tar.gz"
cp -f "$LATEST_FLAT" "$PAYLOAD_DIR/backend.tar.gz"
echo "已复制到：$PAYLOAD_DIR/backend.tar.gz"
# ============================================================================
# [2/5] 前端：前置检测 + 构建 dist.war
# ============================================================================
log "[2/6] 构建前端 dist.war"
FRONTEND_DEPLOY="$ROOT/frontend/deploy"
# ---- P2-FE-05：前置检测 ----
if [ ! -d "$FRONTEND_DEPLOY/main-app" ]; then
  die "未找到 frontend/deploy/main-app。请先执行：cd frontend && pnpm build:prod && pnpm verify:deploy"
fi
if [ ! -f "$FRONTEND_DEPLOY/main-app/index.html" ]; then
  die "未找到 frontend/deploy/main-app/index.html。请先执行：cd frontend && pnpm build:prod && pnpm verify:deploy"
fi
if [ ! -d "$FRONTEND_DEPLOY/sub-apps" ]; then
  die "未找到 frontend/deploy/sub-apps。请先执行：cd frontend && pnpm build:prod && pnpm verify:deploy"
fi
echo "  ✅ 前端产物目录存在"
rm -f "$PAYLOAD_DIR/dist.war"
# ---- P2-FE-08：打入 deploy-manifest.json（若存在） ----
cd "$FRONTEND_DEPLOY"
ZIP_ARGS=("main-app" "sub-apps")
if [ -f "deploy-manifest.json" ]; then
  ZIP_ARGS+=("deploy-manifest.json")
  echo "  → 包含 deploy-manifest.json"
else
  warn "deploy-manifest.json 不存在（可选，后端不依赖）"
fi
zip -qr "$PAYLOAD_DIR/dist.war" "${ZIP_ARGS[@]}"
echo "已生成：$PAYLOAD_DIR/dist.war"
# ============================================================================
# [3/5] 构建 core 载荷（core-go + Tengine + 铜锁 + oqs-provider）
# ============================================================================
log "[3/6] 构建 core 载荷"
CORE_PAYLOAD_SH="$ROOT/create_package/build-core-payload.sh"
[ -x "$CORE_PAYLOAD_SH" ] || die "缺少或不可执行：$CORE_PAYLOAD_SH"
if ! command -v go >/dev/null 2>&1; then
  if [ -x /usr/local/go/bin/go ]; then
    export PATH="/usr/local/go/bin:$PATH"
  fi
fi
bash "$CORE_PAYLOAD_SH"
[ -f "$PAYLOAD_DIR/core.tar.gz" ] || die "core 载荷构建未生成 core.tar.gz"
echo "  ✅ core.tar.gz 已生成"
# ============================================================================
# [4/6] 构建音视频载荷（ZLMediaKit → video.tar.gz）
# ============================================================================
log "[4/6] 构建音视频载荷"
ZLM_PAYLOAD_SH="$ROOT/create_package/build-zlm-payload.sh"
[ -x "$ZLM_PAYLOAD_SH" ] || die "缺少或不可执行：$ZLM_PAYLOAD_SH"
bash "$ZLM_PAYLOAD_SH"
[ -f "$PAYLOAD_DIR/video.tar.gz" ] || die "音视频载荷构建未生成 video.tar.gz"
echo "  ✅ video.tar.gz 已生成"
# ============================================================================
# [5/6] 载荷校验
# ============================================================================
log "[5/6] 载荷校验"
TMP_BACKEND="$(mktemp -d)"
TMP_FRONTEND="$(mktemp -d)"
cleanup_tmp() { rm -rf "$TMP_BACKEND" "$TMP_FRONTEND"; }
trap cleanup_tmp EXIT
tar -xzf "$PAYLOAD_DIR/backend.tar.gz" -C "$TMP_BACKEND" 2>/dev/null \
  || die "backend.tar.gz 不是合法 tar.gz"
FAIL=0
# --- 必须项：bin/server（必须可执行，v1.2 N-02）---
if [ ! -x "$TMP_BACKEND/bin/server" ]; then
  if [ ! -f "$TMP_BACKEND/bin/server" ]; then
    echo "  ❌ bin/server（缺失，请检查构建脚本）"
  else
    echo "  ❌ bin/server（不可执行，请检查构建脚本的 chmod +x）"
  fi
  FAIL=1
else
  echo "  ✅ bin/server"
fi
# --- 必须项：migrations/ ---
if [ -d "$TMP_BACKEND/migrations" ]; then
  echo "  ✅ migrations/"
else
  echo "  ❌ migrations/"
  FAIL=1
fi
# --- 必须项：.env.production.example ---
if [ -f "$TMP_BACKEND/.env.production.example" ]; then
  echo "  ✅ .env.production.example"
else
  echo "  ❌ .env.production.example"
  FAIL=1
fi
# --- 推荐项：bin/goose（警告不阻断，v1.2 N-02 增强）---
if [ ! -x "$TMP_BACKEND/bin/goose" ]; then
  warn "bin/goose 缺失或不可执行（目标机需自行安装 goose）"
else
  echo "  ✅ bin/goose"
fi
# --- 推荐项：DEPLOY.txt（警告不阻断，v1.2 N-02 增强）---
if [ ! -f "$TMP_BACKEND/DEPLOY.txt" ]; then
  warn "DEPLOY.txt 缺失（建议补充）"
else
  echo "  ✅ DEPLOY.txt"
fi
if [ "$FAIL" -ne 0 ]; then
  echo "--- backend.tar.gz 顶层内容 ---"
  (cd "$TMP_BACKEND" && ls -A | head -20)
  die "backend.tar.gz 校验失败"
fi
# --- 前端载荷校验（v1.3 P1-FE-03 增强）---
if command -v unzip >/dev/null 2>&1 && unzip -tq "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  unzip -q "$PAYLOAD_DIR/dist.war" -d "$TMP_FRONTEND" 2>/dev/null \
    || die "dist.war 解压失败"
else
  die "dist.war 不是合法 zip"
fi
validate_frontend_dir() {
  local root="$1"
  [ -f "$root/main-app/index.html" ] || { echo "  ❌ main-app/index.html"; return 1; }
  echo "  ✅ main-app/index.html"
  [ -d "$root/main-app/assets" ] || { echo "  ❌ main-app/assets/"; return 1; }
  echo "  ✅ main-app/assets/"
  [ -d "$root/sub-apps" ] || { echo "  ❌ sub-apps/"; return 1; }
  local found=0
  local d
  for d in "$root"/sub-apps/*/; do
    [ -d "$d" ] || continue
    found=1
    local name
    name="$(basename "$d")"
    if [ ! -f "${d}index.html" ]; then
      echo "  ❌ sub-apps/${name}/index.html"
      return 1
    fi
    if [ ! -d "${d}assets" ]; then
      echo "  ❌ sub-apps/${name}/assets/"
      return 1
    fi
    echo "  ✅ sub-apps/${name}/"
  done
  [ "$found" -eq 1 ] || { echo "  ❌ sub-apps 下无任何子应用"; return 1; }
  return 0
}
if ! validate_frontend_dir "$TMP_FRONTEND"; then
  die "dist.war 前端载荷校验失败"
fi
if [ -f "$TMP_FRONTEND/deploy-manifest.json" ]; then
  echo "  ✅ deploy-manifest.json"
fi
# --- core 载荷校验（v1.4）---
if [ -f "$PAYLOAD_DIR/core.tar.gz" ]; then
  TMP_CORE="$(mktemp -d)"
  tar -xzf "$PAYLOAD_DIR/core.tar.gz" -C "$TMP_CORE" 2>/dev/null \
    || die "core.tar.gz 不是合法 tar.gz"
  CORE_FAIL=0
  for f in bin/core sbin/nginx conf/core.yaml conf/mime.types \
           libs/tongsuo/bin/openssl libs/tongsuo/libs/libcrypto.so.3; do
    if [ -e "$TMP_CORE/core/$f" ]; then
      echo "  ✅ core/$f"
    else
      echo "  ❌ core/$f"
      CORE_FAIL=1
    fi
  done
  rm -rf "$TMP_CORE"
  if [ "$CORE_FAIL" -ne 0 ]; then
    die "core.tar.gz 校验失败"
  fi
else
  warn "core.tar.gz 缺失（可能未构建 core 载荷）"
fi
# --- 音视频载荷校验（v1.5）---
if [ -f "$PAYLOAD_DIR/video.tar.gz" ]; then
  TMP_VIDEO="$(mktemp -d)"
  tar -xzf "$PAYLOAD_DIR/video.tar.gz" -C "$TMP_VIDEO" 2>/dev/null \
    || die "video.tar.gz 不是合法 tar.gz"
  if [ -x "$TMP_VIDEO/video/run/bin/MediaServer" ]; then
    echo "  ✅ video/run/bin/MediaServer"
  else
    echo "  ❌ video/run/bin/MediaServer"
    rm -rf "$TMP_VIDEO"
    die "video.tar.gz 校验失败"
  fi
  if [ -f "$TMP_VIDEO/video/conf/zlmediakit/config.ini" ]; then
    echo "  ✅ video/conf/zlmediakit/config.ini"
  else
    echo "  ❌ video/conf/zlmediakit/config.ini"
    rm -rf "$TMP_VIDEO"
    die "video.tar.gz 校验失败"
  fi
  rm -rf "$TMP_VIDEO"
else
  warn "video.tar.gz 缺失（可能未构建音视频载荷）"
fi
echo "✅ 载荷校验通过"
# ============================================================================
# [6/6] 生成产品安装包
# ============================================================================
log "[6/6] 生成产品安装包"
[ -x "$CREATE_SH" ] || die "缺少或不可执行：$CREATE_SH"
cd "$ROOT/create_package"
bash "$CREATE_SH"
echo
echo "🎉 完成，输出目录：$ROOT/create_package/out/"
ls -lh "$ROOT/create_package/out/" | tail -5
