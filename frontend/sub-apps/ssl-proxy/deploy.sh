#!/usr/bin/env bash
# ============================================================
# ssl-proxy 子应用一键编译部署脚本
# 路径：frontend/sub-apps/ssl-proxy/deploy.sh
#
# 用法：
#   ./deploy.sh            # 正式构建（production）
#   ./deploy.sh --clean    # 先清缓存再构建
#   ./deploy.sh --dev      # 构建 development 模式（一般不用于部署）
#
# 产物：
#   frontend/deploy/sub-apps/ssl-proxy/
# ============================================================
set -euo pipefail

APP_NAME="ssl-proxy"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(cd "$SRC_DIR/../../deploy/sub-apps" && pwd)/$APP_NAME"

CLEAN=0
BUILD_SCRIPT="build:prod"

for arg in "$@"; do
  case "$arg" in
    --clean) CLEAN=1 ;;
    --dev)   BUILD_SCRIPT="build" ;;
    -h|--help)
      grep '^#' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
  esac
done

echo "==> 环境检查"
if ! command -v pnpm >/dev/null 2>&1; then
  echo "[deploy.sh] ERROR: 未找到 pnpm，请先安装 pnpm" >&2
  exit 2
fi
if [ ! -f "$SRC_DIR/package.json" ]; then
  echo "[deploy.sh] ERROR: 未找到 $SRC_DIR/package.json" >&2
  exit 2
fi

if [ "$CLEAN" -eq 1 ]; then
  echo "==> 清理旧产物与缓存"
  rm -rf "$SRC_DIR/dist" "$SRC_DIR/node_modules/.vite"
fi

if [ ! -d "$SRC_DIR/node_modules" ]; then
  echo "==> 安装依赖（首次）"
  cd "$SRC_DIR"
  pnpm install
fi

echo "==> 构建 $APP_NAME ($BUILD_SCRIPT)"
cd "$SRC_DIR"
pnpm "$BUILD_SCRIPT"

if [ ! -d "$SRC_DIR/dist" ]; then
  echo "[deploy.sh] ERROR: 构建产物不存在：$SRC_DIR/dist" >&2
  exit 9
fi

echo "==> 复制到 $DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR"
# 清空目标目录但保留目录本身
find "$DEPLOY_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
cp -r "$SRC_DIR/dist/." "$DEPLOY_DIR/"

echo "==> 完成"
ls -la "$DEPLOY_DIR"
echo ""
echo "生产切换（后端模块记录）："
echo "  UPDATE module_manager_module"
echo "  SET entry_frontend = '/sub-apps/$APP_NAME/'"
echo "  WHERE id = 'ssl_proxy';"
