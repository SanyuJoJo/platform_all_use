#!/usr/bin/env bash
# ============================================================================
# 构建 Go 后端部署包（tar.gz）
#
# 用法：
#   ./scripts/build.sh                # 默认含顶层目录
#   ./scripts/build.sh --flat         # 无顶层目录（产品部署包推荐）
#   ./scripts/build.sh --with-goose   # 将 goose 二进制打入包内
#   ./scripts/build.sh --flat --with-goose
#
# 输出：
#   dist/backend-deploy-YYYYMMDD-HHMMSS.tar.gz
#   dist/backend-flat-YYYYMMDD-HHMMSS.tar.gz
#   dist/*.manifest.json
#
# 说明：
#   --with-goose 会尝试从 $(go env GOPATH)/bin/goose 复制；
#   若本机未安装 goose，会打印警告并跳过，不阻断构建。
# ============================================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$BACKEND_DIR"
FLAT=0
WITH_GOOSE=0
for arg in "$@"; do
  case "$arg" in
    --flat) FLAT=1 ;;
    --with-goose) WITH_GOOSE=1 ;;
    -h|--help)
      echo "用法：$0 [--flat] [--with-goose]"
      exit 0
      ;;
    *) echo "未知参数：$arg"; exit 1 ;;
  esac
done
DIST_DIR="$BACKEND_DIR/dist"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
if [ "$FLAT" -eq 1 ]; then
  PACKAGE_NAME="backend-flat-${TIMESTAMP}"
else
  PACKAGE_NAME="backend-deploy-${TIMESTAMP}"
fi
STAGING_DIR="$DIST_DIR/$PACKAGE_NAME"
ARCHIVE_PATH="$DIST_DIR/${PACKAGE_NAME}.tar.gz"
MANIFEST_PATH="$DIST_DIR/${PACKAGE_NAME}.manifest.json"
rm -rf "$STAGING_DIR"
mkdir -p "$STAGING_DIR/bin" "$STAGING_DIR/migrations" "$DIST_DIR"
echo "🔍 校验关键文件..."
for item in cmd/server internal migrations go.mod go.sum; do
  [ -e "$BACKEND_DIR/$item" ] || { echo "❌ 缺少 $item"; exit 1; }
done
echo "✅ 关键文件齐全"
echo "🛠  编译 Go 二进制（CGO_ENABLED=0）..."
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$STAGING_DIR/bin/server" ./cmd/server
chmod +x "$STAGING_DIR/bin/server"
echo "📦 复制迁移脚本..."
cp -R "$BACKEND_DIR/migrations/." "$STAGING_DIR/migrations/"
if [ -f "$BACKEND_DIR/.env.production.example" ]; then
  cp "$BACKEND_DIR/.env.production.example" "$STAGING_DIR/.env.production.example"
else
  echo "⚠️  未找到 .env.production.example，跳过"
fi
if [ "$WITH_GOOSE" -eq 1 ]; then
  GOOSE_BIN="$(go env GOPATH)/bin/goose"
  if [ -x "$GOOSE_BIN" ]; then
    echo "📦 复制 goose：$GOOSE_BIN"
    cp "$GOOSE_BIN" "$STAGING_DIR/bin/goose"
    chmod +x "$STAGING_DIR/bin/goose"
  else
    echo "⚠️  未找到 goose（$GOOSE_BIN），跳过。目标机需自行安装 goose。"
  fi
fi
cat > "$STAGING_DIR/DEPLOY.txt" <<'EOF'
============================================================
Go 后端部署包说明
============================================================
1. 解压后直接得到：
   bin/server                 # Go 静态二进制
   bin/goose                  # goose 迁移工具（若构建时包含）
   migrations/                # SQL 迁移脚本
   .env.production.example    # 生产环境变量模板
2. 数据库迁移：
   ./bin/goose -dir migrations sqlite3 <数据库路径> up
3. 启动：
   ./bin/server
4. 健康检查：
   curl http://localhost:8000/health
5. 环境变量：
   复制 .env.production.example 为 .env 并按需修改。
============================================================
EOF
if [ "$FLAT" -eq 1 ]; then
  echo "📦 打包 tar.gz（--flat：无顶层目录）..."
  tar -czf "$ARCHIVE_PATH" -C "$STAGING_DIR" .
else
  echo "📦 打包 tar.gz（含顶层目录）..."
  tar -czf "$ARCHIVE_PATH" -C "$DIST_DIR" "$PACKAGE_NAME"
fi
if command -v sha256sum >/dev/null 2>&1; then
  SHA256="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
else
  SHA256="unavailable"
fi
SIZE="$(stat -c%s "$ARCHIVE_PATH" 2>/dev/null || stat -f%z "$ARCHIVE_PATH")"
cat > "$MANIFEST_PATH" <<EOF
{
  "package": "${PACKAGE_NAME}.tar.gz",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "size_bytes": ${SIZE},
  "sha256": "${SHA256}",
  "flat": $([ "$FLAT" -eq 1 ] && echo "true" || echo "false"),
  "with_goose": $([ "$WITH_GOOSE" -eq 1 ] && echo "true" || echo "false")
}
EOF
rm -rf "$STAGING_DIR"
echo ""
echo "✅ 构建完成"
echo "   部署包：$ARCHIVE_PATH"
echo "   Manifest：$MANIFEST_PATH"
echo "   SHA256：$SHA256"
echo "   模式：$([ "$FLAT" -eq 1 ] && echo "flat（无顶层目录）" || echo "默认（含顶层目录）")"
