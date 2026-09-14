#!/usr/bin/env bash
# ============================================================================
# 构建后端部署包（tar.gz）
#
# v1.1 修复：
#   - P1-3：兼容 macOS，`sha256sum` 不存在时回退 `shasum -a 256`。
#   - P2-3：对关键文件（src/migrations/alembic.ini/pyproject.toml/uv.lock）
#           增加存在性校验，缺失则报错退出，避免静默产出残缺包。
#
# 用法：
#   ./scripts/build.sh                # 完整构建（含 ruff 检查）
#   ./scripts/build.sh --skip-lint    # 跳过 ruff 检查
#   ./scripts/build.sh --include-frontend  # 若前端产物存在，一并打包
#
# 输出：
#   dist/backend-deploy-YYYYMMDD-HHMMSS.tar.gz
#   dist/backend-deploy-YYYYMMDD-HHMMSS.manifest.json
# ============================================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$BACKEND_DIR/.." && pwd)"
cd "$BACKEND_DIR"
SKIP_LINT=0
INCLUDE_FRONTEND=0
for arg in "$@"; do
    case "$arg" in
        --skip-lint) SKIP_LINT=1 ;;
        --include-frontend) INCLUDE_FRONTEND=1 ;;
        -h|--help)
            echo "用法：$0 [--skip-lint] [--include-frontend]"
            exit 0
            ;;
        *)
            echo "❌ 未知参数：$arg"
            exit 1
            ;;
    esac
done
# ---------- 准备 ----------
DIST_DIR="$BACKEND_DIR/dist"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
PACKAGE_NAME="backend-deploy-${TIMESTAMP}"
STAGING_DIR="$DIST_DIR/$PACKAGE_NAME"
ARCHIVE_PATH="$DIST_DIR/${PACKAGE_NAME}.tar.gz"
MANIFEST_PATH="$DIST_DIR/${PACKAGE_NAME}.manifest.json"
rm -rf "$STAGING_DIR"
mkdir -p "$STAGING_DIR" "$DIST_DIR"
# ---------- P2-3：关键文件存在性校验 ----------
echo "🔍 校验关键文件..."
REQUIRED_ITEMS=("src" "migrations" "scripts" "alembic.ini" "pyproject.toml" "uv.lock")
MISSING=()
for item in "${REQUIRED_ITEMS[@]}"; do
    if [ ! -e "$BACKEND_DIR/$item" ]; then
        MISSING+=("$item")
    fi
done
if [ "${#MISSING[@]}" -gt 0 ]; then
    echo "❌ 缺少关键文件/目录：${MISSING[*]}"
    exit 1
fi
echo "✅ 关键文件齐全"
# ---------- 代码风格检查 ----------
if [ "$SKIP_LINT" -eq 0 ] && command -v ruff >/dev/null 2>&1; then
    echo "🔍 执行 ruff check（FastAPI 惯例 B008 / 复杂度 C901 已忽略）..."
    if ! ruff check src/ tests/; then
        echo ""
        echo "❌ ruff check 失败。建议修复步骤："
        echo "   1. 自动修复可修复项："
        echo "        ruff check --fix src/ tests/"
        echo "   2. 查看剩余问题："
        echo "        ruff check src/ tests/"
        echo "   3. 如确认某项为误报，可在 pyproject.toml 的"
        echo "        [tool.ruff.lint] ignore 中追加规则代码。"
        echo "   4. 临时跳过 lint 构建："
        echo "        $0 --skip-lint"
        exit 1
    fi
    echo "✅ ruff check 通过"
elif [ "$SKIP_LINT" -eq 1 ]; then
    echo "⏭  已跳过 ruff check（--skip-lint）"
fi
# ---------- 复制文件 ----------
echo "📦 复制后端文件..."
copy_item() {
    local src="$1"
    local dst="$2"
    if [ -e "$src" ]; then
        mkdir -p "$(dirname "$dst")"
        cp -R "$src" "$dst"
    fi
}
copy_item "$BACKEND_DIR/src"                    "$STAGING_DIR/src"
copy_item "$BACKEND_DIR/migrations"             "$STAGING_DIR/migrations"
copy_item "$BACKEND_DIR/scripts"                "$STAGING_DIR/scripts"
copy_item "$BACKEND_DIR/alembic.ini"            "$STAGING_DIR/alembic.ini"
copy_item "$BACKEND_DIR/pyproject.toml"         "$STAGING_DIR/pyproject.toml"
copy_item "$BACKEND_DIR/uv.lock"                "$STAGING_DIR/uv.lock"
copy_item "$BACKEND_DIR/.env.production.example" "$STAGING_DIR/.env.production.example"
copy_item "$BACKEND_DIR/Dockerfile"             "$STAGING_DIR/Dockerfile"
copy_item "$BACKEND_DIR/.dockerignore"          "$STAGING_DIR/.dockerignore"
# 移除 __pycache__
find "$STAGING_DIR" -type d -name "__pycache__" -prune -exec rm -rf {} + 2>/dev/null || true
find "$STAGING_DIR" -type f -name "*.pyc" -delete 2>/dev/null || true
# ---------- 可选：包含前端产物 ----------
if [ "$INCLUDE_FRONTEND" -eq 1 ]; then
    if [ -d "$REPO_ROOT/frontend/deploy/main-app" ]; then
        echo "📦 包含前端产物..."
        mkdir -p "$STAGING_DIR/frontend"
        cp -R "$REPO_ROOT/frontend/deploy" "$STAGING_DIR/frontend/deploy"
    else
        echo "⚠️  未找到前端产物（frontend/deploy/main-app），跳过"
    fi
fi
# ---------- 生成部署说明 ----------
cat > "$STAGING_DIR/DEPLOY.txt" <<'EOF'
============================================================
后端部署包说明
============================================================
1. 解压
   tar -xzf backend-deploy-*.tar.gz
   cd backend-deploy-*
2. 准备环境
   - 安装 Python 3.12+ 与 uv（https://astral.sh/uv）
   - 复制 .env.production.example 为 .env 并修改关键配置
       cp .env.production.example .env
       # 必须修改 SECRET_KEY / LICENSE_SECRET_KEY
       # 建议修改 CORS_ORIGINS
       # 若使用容器，注意 FRONTEND_DEPLOY_DIR 与 volume 挂载点保持一致
3. 安装依赖
   uv venv --python 3.12
   source .venv/bin/activate
   uv sync --frozen --no-dev
4. 数据库迁移
   alembic upgrade head
5. 启动服务
   # 前台启动（调试）
   uvicorn src.main:app --host 0.0.0.0 --port 8000
   # 后台启动（生产，推荐）
   ./scripts/start.sh
6. 停止服务
   ./scripts/stop.sh
7. 健康检查
   ./scripts/healthcheck.sh
   # 或
   curl http://localhost:8000/health
8. 容器化部署（可选）
   docker build -t platform-backend .
   docker run -d -p 8000:8000 --env-file .env \
     -v $(pwd)/data:/app/data \
     -v $(pwd)/logs:/app/logs \
     -v $(pwd)/uploads:/app/uploads \
     -v <前端产物路径>:/app/frontend/deploy:ro \
     platform-backend
============================================================
EOF
# ---------- 打包 ----------
echo "📦 打包 tar.gz..."
tar -czf "$ARCHIVE_PATH" -C "$DIST_DIR" "$PACKAGE_NAME"
# ---------- P1-3：SHA256 计算（兼容 macOS）----------
echo "📄 计算 SHA256..."
if command -v sha256sum >/dev/null 2>&1; then
    SHA256=$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    SHA256=$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')
else
    echo "⚠️  未找到 sha256sum 或 shasum，跳过 SHA256 计算"
    SHA256="unavailable"
fi
SIZE=$(stat -c%s "$ARCHIVE_PATH" 2>/dev/null || stat -f%z "$ARCHIVE_PATH")
# ---------- 生成 manifest ----------
cat > "$MANIFEST_PATH" <<EOF
{
  "package": "${PACKAGE_NAME}.tar.gz",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "size_bytes": ${SIZE},
  "sha256": "${SHA256}",
  "include_frontend": $([ "$INCLUDE_FRONTEND" -eq 1 ] && echo "true" || echo "false")
}
EOF
# ---------- 清理 staging ----------
rm -rf "$STAGING_DIR"
# ---------- 输出 ----------
echo ""
echo "✅ 构建完成"
echo "   部署包：$ARCHIVE_PATH"
echo "   Manifest：$MANIFEST_PATH"
echo "   SHA256：$SHA256"
echo "   大小：$(numfmt --to=iec "$SIZE" 2>/dev/null || echo "${SIZE} bytes")"
