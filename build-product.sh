#!/usr/bin/env bash
# ============================================================================
# 一键生成产品安装包：后端 → 前端 → 载荷校验 → 调用 create.sh 打包
# 位置：项目根目录 build-product.sh
#
# v1.2 修复：
#   - N-02：后端校验改用 -x（可执行），与 create.sh 一致；
#           新增 bin/goose、DEPLOY.txt 缺失警告（不阻断）。
# ============================================================================
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PAYLOAD_DIR="$ROOT/create_package/bin/bin"
CREATE_SH="$ROOT/create_package/create.sh"
log() { echo -e "\n===== $* ====="; }
warn() { echo -e "  ⚠️  $*"; }
die() { echo -e "\n❌ $*" >&2; exit 1; }
log "[1/4] 构建后端 Flat 包（含 goose）"
cd "$ROOT/backend-go"
rm -f dist/backend-flat-*.tar.gz dist/backend-flat-*.manifest.json
./scripts/build.sh --flat --with-goose
LATEST_FLAT="$(ls -t dist/backend-flat-*.tar.gz 2>/dev/null | head -1 || true)"
[ -n "$LATEST_FLAT" ] || die "未找到 backend-flat-*.tar.gz"
echo "本次 Flat 包：$LATEST_FLAT"
mkdir -p "$PAYLOAD_DIR"
rm -f "$PAYLOAD_DIR/backend.tar.gz"
cp -f "$LATEST_FLAT" "$PAYLOAD_DIR/backend.tar.gz"
echo "已复制到：$PAYLOAD_DIR/backend.tar.gz"
log "[2/4] 构建前端 dist.war"
[ -d "$ROOT/frontend/deploy/main-app" ] || die "未找到 frontend/deploy/main-app"
[ -d "$ROOT/frontend/deploy/sub-apps" ] || die "未找到 frontend/deploy/sub-apps"
rm -f "$PAYLOAD_DIR/dist.war"
cd "$ROOT/frontend/deploy"
zip -qr "$PAYLOAD_DIR/dist.war" main-app sub-apps
echo "已生成：$PAYLOAD_DIR/dist.war"
log "[3/4] 载荷校验"
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
# --- 前端载荷校验 ---
if command -v unzip >/dev/null 2>&1 && unzip -tq "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  unzip -q "$PAYLOAD_DIR/dist.war" -d "$TMP_FRONTEND" 2>/dev/null \
    || die "dist.war 解压失败"
else
  die "dist.war 不是合法 zip"
fi
[ -f "$TMP_FRONTEND/main-app/index.html" ] || die "dist.war 缺 main-app/index.html"
[ -d "$TMP_FRONTEND/sub-apps" ]           || die "dist.war 缺 sub-apps/"
echo "  ✅ main-app/index.html"
echo "  ✅ sub-apps/"
echo "✅ 载荷校验通过"
log "[4/4] 生成产品安装包"
[ -x "$CREATE_SH" ] || die "缺少或不可执行：$CREATE_SH"
cd "$ROOT/create_package"
bash "$CREATE_SH"
echo
echo "🎉 完成，输出目录：$ROOT/create_package/out/"
ls -lh "$ROOT/create_package/out/" | tail -5
