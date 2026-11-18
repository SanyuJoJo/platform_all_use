#!/usr/bin/env bash
# ============================================================
# media 子应用一键编译部署脚本
# 路径：frontend/sub-apps/media/deploy.sh
#
# 用法：
#   ./deploy.sh            # 正式构建（production）
#   ./deploy.sh --clean    # 先清缓存再构建
#   ./deploy.sh --dev      # 构建 development 模式（一般不用于部署）
#
# 产物：
#   frontend/deploy/sub-apps/media/
#
# 说明：
#   - 依赖 ZLMediaKit 已由 core/libs/Makefile 部署（不在本脚本职责内）
#   - 生产 publicPath 由 .env.production 中的 VITE_PUBLIC_PATH 控制
#     （默认回退为 /sub-apps/media/）
# ============================================================
set -euo pipefail

APP_NAME="media"
MODULE_ID="media"
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
    *) echo "[deploy.sh] WARN: 忽略未知参数 $arg" >&2 ;;
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
if [ ! -f "$SRC_DIR/.env.production" ]; then
  echo "[deploy.sh] WARN: 未找到 $SRC_DIR/.env.production，将使用 vite.config.ts 默认值" >&2
  echo "          建议：cp .env.example .env.production 并调整 VITE_API_BASE_URL / VITE_PUBLIC_PATH" >&2
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
if [ ! -f "$SRC_DIR/dist/index.html" ]; then
  echo "[deploy.sh] ERROR: dist/index.html 缺失，构建可能失败" >&2
  exit 9
fi

echo "==> 复制到 $DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR"
# 清空目标目录但保留目录本身
find "$DEPLOY_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
cp -r "$SRC_DIR/dist/." "$DEPLOY_DIR/"

# 权限修正（nginx 需要可读）
chmod -R 0755 "$DEPLOY_DIR"

echo "==> 完成"
ls -la "$DEPLOY_DIR"

echo ""
echo "生产切换（后端模块记录）："
echo "  UPDATE module_manager_module"
echo "  SET entry_frontend = '/sub-apps/$APP_NAME/'"
echo "  WHERE id = '$MODULE_ID';"
echo ""
echo "验证清单："
echo "  1) 子应用 HTTP 可达："
echo "       curl -I http://<host>/sub-apps/$APP_NAME/index.html"
echo "  2) 模块注册状态（后端接口）："
echo "       curl -s http://<host>/api/v1/modules/active | jq '.[] | select(.id==\"$MODULE_ID\")'"
echo "  3) 主应用菜单可见路径："
echo "       /$MODULE_ID/media   （媒体库）"
echo "       /$MODULE_ID/live    （直播管理）"
echo "  4) ZLMediaKit 存活："
echo "       curl -s 'http://<zlm-host>:8080/index/api/getServerConfig?secret=<secret>' | head"
echo ""
echo "提示："
echo "  - 主应用 REGISTRY 的 SUB_APP_DIR_MAP 必须包含 media: 'media'"
echo "  - 若部署到非默认前缀，请同步修改 .env.production 的 VITE_PUBLIC_PATH"
echo "  - ZLMediaKit 独立部署根：/opt/plats_tool/video（由 core/libs/Makefile 完成）"
