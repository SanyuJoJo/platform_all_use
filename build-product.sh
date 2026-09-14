#!/usr/bin/env bash
# ============================================================================
# /code/platform_all_use/build-product.sh
# 一键生成产品安装包：后端 → 前端 → 载荷校验 → 调用 create.sh 打包
#
# v1.4 修复：
#   - [3/4] 载荷校验去掉 .venv 相关项（新版 build.sh 默认不打包 .venv，
#     由 install.sh 在目标机重建）。
# v1.3 修复：
#   - [4/4] 段使用绝对路径调用 create.sh，避免工作目录错位。
#   - [3/4] 段用文件系统检查代替 grep 字符串匹配。
# v1.2 修复：
#   - 修复 backend/dist/ 多 flat 包时 cp 通配符展开为多文件的问题。
# ============================================================================
set -euo pipefail

ROOT="/code/platform_all_use"
PAYLOAD_DIR="$ROOT/create_package/bin/bin"
CREATE_SH="$ROOT/create_package/create.sh"

log() { echo -e "\n===== $* ====="; }
die() { echo -e "\n❌ $*" >&2; exit 1; }

# ============================================================================
# [1/4] 构建后端 Flat 包
# ============================================================================
log "[1/4] 构建后端 Flat 包"
cd "$ROOT/backend"
rm -f "$ROOT/backend/dist/backend-flat-"*.tar.gz
rm -f "$ROOT/backend/dist/backend-flat-"*.manifest.json
./scripts/build.sh --skip-lint --flat

LATEST_FLAT="$(ls -t "$ROOT/backend/dist/"backend-flat-*.tar.gz 2>/dev/null | head -1 || true)"
[ -n "$LATEST_FLAT" ] || die "未找到 backend-flat-*.tar.gz"
echo "本次 Flat 包：$LATEST_FLAT"

mkdir -p "$PAYLOAD_DIR"
rm -f "$PAYLOAD_DIR/backend.tar.gz"
cp -f "$LATEST_FLAT" "$PAYLOAD_DIR/backend.tar.gz"
echo "已复制到：$PAYLOAD_DIR/backend.tar.gz"

# ============================================================================
# [2/4] 构建前端 dist.war
# ============================================================================
log "[2/4] 构建前端 dist.war"
[ -d "$ROOT/frontend/deploy/main-app" ] || die "未找到 frontend/deploy/main-app"
[ -d "$ROOT/frontend/deploy/sub-apps" ] || die "未找到 frontend/deploy/sub-apps"
rm -f "$PAYLOAD_DIR/dist.war"
cd "$ROOT/frontend/deploy"
zip -qr "$PAYLOAD_DIR/dist.war" main-app sub-apps
echo "已生成：$PAYLOAD_DIR/dist.war"

# ============================================================================
# [3/4] 载荷校验（文件系统检查；不再校验 .venv）
# ============================================================================
log "[3/4] 载荷校验"
TMP_BACKEND="$(mktemp -d)"
TMP_FRONTEND="$(mktemp -d)"
cleanup_tmp() { rm -rf "$TMP_BACKEND" "$TMP_FRONTEND"; }
trap cleanup_tmp EXIT

tar -xzf "$PAYLOAD_DIR/backend.tar.gz" -C "$TMP_BACKEND" 2>/dev/null \
  || die "backend.tar.gz 不是合法 tar.gz"

FAIL=0
# 说明：新版 build.sh 不再打包 .venv，由 install.sh 在目标机重建。
#       这里只校验源码与配置文件。
for rel in "src/main.py" "alembic.ini" "pyproject.toml" "migrations"; do
  if [ -e "$TMP_BACKEND/$rel" ]; then
    echo "  ✅ $rel"
  else
    echo "  ❌ $rel"
    FAIL=1
  fi
done
if [ "$FAIL" -ne 0 ]; then
  echo "--- backend.tar.gz 顶层内容 ---"
  (cd "$TMP_BACKEND" && ls -A | head -20)
  die "backend.tar.gz 校验失败"
fi

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

# ============================================================================
# [4/4] 生成产品安装包（用绝对路径调用 create.sh）
# ============================================================================
log "[4/4] 生成产品安装包"
[ -x "$CREATE_SH" ] || die "缺少或不可执行：$CREATE_SH"
cd "$ROOT/create_package"
bash "$CREATE_SH"

echo
echo "🎉 完成，输出目录：$ROOT/create_package/out/"
ls -lh "$ROOT/create_package/out/" | tail -5
