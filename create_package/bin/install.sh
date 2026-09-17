#!/usr/bin/env bash
# ============================================================
# 产品安装脚本（Go 版 v1.1）
#
# v1.1 变更：
#   - P2-03：SQLite 路径解析兼容单引号 / 双引号 / 无引号
# ============================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PAYLOAD_DIR="$SCRIPT_DIR/bin"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
VERSION_FILE="$SCRIPT_DIR/VERSION"
INSTALL_DIR="${INSTALL_DIR:-/opt/platform}"
SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
SERVER_PORT="${SERVER_PORT:-8000}"
WORKERS="${WORKERS:-1}"
CORS_ORIGINS="${CORS_ORIGINS:-*}"
START_SERVICE=1
USE_SYSTEMD=0
FORCE=0
log()  { echo -e "[install] $*"; }
warn() { echo -e "[install][WARN] $*" >&2; }
fail() { echo -e "[install][ERROR] $*" >&2; exit 1; }
usage() {
  cat <<EOF
用法：$0 [选项]
  --install-dir PATH   安装目录，默认 /opt/platform
  --port PORT          后端端口，默认 8000
  --host HOST          监听地址，默认 0.0.0.0
  --workers N          worker 数（Go 版保留兼容，不生效），默认 1
  --cors ORIGINS       CORS，默认 *
  --no-start           安装后不启动
  --systemd            安装并启用 systemd 服务（需 root）
  --force              覆盖已有 .env（重新生成密钥等）
  -h, --help           显示帮助
EOF
}
while [ $# -gt 0 ]; do
  case "$1" in
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    --port) SERVER_PORT="$2"; shift 2 ;;
    --host) SERVER_HOST="$2"; shift 2 ;;
    --workers) WORKERS="$2"; shift 2 ;;
    --cors) CORS_ORIGINS="$2"; shift 2 ;;
    --no-start) START_SERVICE=0; shift ;;
    --systemd) USE_SYSTEMD=1; shift ;;
    --force) FORCE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) fail "未知参数：$1" ;;
  esac
done
if [ "$(id -u)" != "0" ] && [ "$INSTALL_DIR" = "/opt/platform" ]; then
  INSTALL_DIR="$HOME/platform"
  log "非 root 用户，安装目录回退为：$INSTALL_DIR"
fi
BACKEND_DIR="$INSTALL_DIR/backend"
FRONTEND_DIR="$INSTALL_DIR/frontend"
FRONTEND_DEPLOY_DIR="$FRONTEND_DIR/deploy"
DATA_DIR="$INSTALL_DIR/data"
LOG_DIR="$INSTALL_DIR/logs"
UPLOAD_DIR="$INSTALL_DIR/uploads"
RUN_DIR="$INSTALL_DIR/run"
BIN_DIR="$INSTALL_DIR/bin"
ENV_FILE="$BACKEND_DIR/.env"
log "安装目录：$INSTALL_DIR"
log "后端目录：$BACKEND_DIR"
log "后端端口：$SERVER_PORT"
# ---------- 依赖检查 ----------
for cmd in tar curl; do
  command -v "$cmd" >/dev/null 2>&1 || fail "缺少命令：$cmd"
done
if ! command -v unzip >/dev/null 2>&1; then
  command -v tar >/dev/null 2>&1 || fail "缺少 unzip，且 tar 不可用"
  log "未找到 unzip，dist.war 将尝试以 tar.gz 解压"
fi
if ! command -v openssl >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
  fail "需要 openssl 或 python3 之一用于生成密钥"
fi
[ -f "$PAYLOAD_DIR/backend.tar.gz" ] || fail "缺少后端包：$PAYLOAD_DIR/backend.tar.gz"
[ -f "$PAYLOAD_DIR/dist.war" ] || fail "缺少前端包：$PAYLOAD_DIR/dist.war"
[ -d "$TEMPLATES_DIR" ] || fail "缺少模板目录：$TEMPLATES_DIR"
# ---------- 创建目录 ----------
mkdir -p "$BACKEND_DIR" "$FRONTEND_DEPLOY_DIR" "$DATA_DIR" "$LOG_DIR" \
         "$UPLOAD_DIR/modules" "$RUN_DIR" "$BIN_DIR"
# ---------- 解压后端 + 结构校验 ----------
log "解压后端..."
tar -xzf "$PAYLOAD_DIR/backend.tar.gz" -C "$BACKEND_DIR"
[ -x "$BACKEND_DIR/bin/server" ] || fail "后端包缺少可执行文件 bin/server"
[ -d "$BACKEND_DIR/migrations" ] || fail "后端包缺少 migrations/"
[ -f "$BACKEND_DIR/.env.production.example" ] || fail "后端包缺少 .env.production.example"
# ---------- 解压前端 + 结构校验 ----------
log "解压前端..."
rm -rf "$FRONTEND_DEPLOY_DIR"
mkdir -p "$FRONTEND_DEPLOY_DIR"
if command -v unzip >/dev/null 2>&1 && unzip -tq "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  unzip -q "$PAYLOAD_DIR/dist.war" -d "$FRONTEND_DEPLOY_DIR"
else
  tar -xzf "$PAYLOAD_DIR/dist.war" -C "$FRONTEND_DEPLOY_DIR"
fi
[ -f "$FRONTEND_DEPLOY_DIR/main-app/index.html" ] \
  || fail "前端包缺少 main-app/index.html"
[ -d "$FRONTEND_DEPLOY_DIR/sub-apps" ] \
  || fail "前端包缺少 sub-apps/ 目录"
# ---------- 密钥生成 ----------
gen_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c 'import secrets; print(secrets.token_hex(32))'
  else
    fail "缺少 openssl 和 python3，无法生成密钥"
  fi
}
# ---------- .env 处理 ----------
if [ -f "$ENV_FILE" ] && [ "$FORCE" -ne 1 ]; then
  log "检测到已有配置文件：$ENV_FILE"
  log "未指定 --force，完全保留现有 .env"
  GENERATE_ENV=0
else
  if [ -f "$ENV_FILE" ] && [ "$FORCE" -eq 1 ]; then
    BACKUP="$ENV_FILE.bak.$(date +%Y%m%d%H%M%S)"
    cp "$ENV_FILE" "$BACKUP"
    log "已备份原 .env 到：$BACKUP"
  fi
  GENERATE_ENV=1
fi
if [ "$GENERATE_ENV" -eq 1 ]; then
  SECRET_KEY="$(gen_secret)"
  LICENSE_SECRET_KEY="$(gen_secret)"
  cat > "$ENV_FILE" <<EOF
APP_NAME="Platform Backend"
APP_ENV="production"
DEBUG="false"
LOG_LEVEL="INFO"
SECRET_KEY="${SECRET_KEY}"
ALGORITHM="HS256"
ACCESS_TOKEN_EXPIRE_MINUTES="1440"
LICENSE_SECRET_KEY="${LICENSE_SECRET_KEY}"
LICENSE_ACTIVATION_URL=""
LICENSE_MACHINE_CODE_OVERRIDE=""
DATABASE_URL="sqlite+aiosqlite:///${DATA_DIR}/app.db"
SERVER_HOST="${SERVER_HOST}"
SERVER_PORT="${SERVER_PORT}"
WORKERS="${WORKERS}"
FRONTEND_DEPLOY_DIR="${FRONTEND_DEPLOY_DIR}"
CORS_ORIGINS="${CORS_ORIGINS}"
MODULES_DIR="src/modules"
MODULE_UPLOAD_DIR="${UPLOAD_DIR}/modules"
MODULE_ZIP_MAX_SIZE="52428800"
MODULE_ZIP_MAX_TOTAL="209715200"
MODULE_ZIP_MAX_FILES="2000"
EOF
  chmod 600 "$ENV_FILE"
  log "已生成配置文件：$ENV_FILE"
else
  chmod 600 "$ENV_FILE" 2>/dev/null || true
fi
# ---------- 数据库迁移（goose） ----------
log "执行数据库迁移..."
GOOSE_BIN=""
if [ -x "$BACKEND_DIR/bin/goose" ]; then
  GOOSE_BIN="$BACKEND_DIR/bin/goose"
elif command -v goose >/dev/null 2>&1; then
  GOOSE_BIN="$(command -v goose)"
else
  fail "缺少 goose，无法执行数据库迁移。请将 goose 二进制放入 backend/bin/goose"
fi
# 从 .env 读取 DATABASE_URL，提取 sqlite 文件路径
# ★ v1.1（P2-03）：兼容单引号 / 双引号 / 无引号
SQLITE_PATH="$(grep -E '^DATABASE_URL=' "$ENV_FILE" | head -1 \
  | sed -E "s/^DATABASE_URL=[\"']?//; s/[\"']$//" \
  | sed -E 's#^sqlite\+aiosqlite:///##; s#^sqlite:///##')"
if [ -z "$SQLITE_PATH" ]; then
  fail "无法从 DATABASE_URL 解析 SQLite 路径"
fi
"$GOOSE_BIN" -dir "$BACKEND_DIR/migrations" sqlite3 "$SQLITE_PATH" up
# ---------- 生成控制脚本 ----------
log "生成控制脚本..."
replace_vars() {
  local template_path="$1"
  local content
  content="$(cat "$template_path")"
  content="${content//@INSTALL_DIR@/$INSTALL_DIR}"
  content="${content//@BACKEND_DIR@/$BACKEND_DIR}"
  content="${content//@ENV_FILE@/$ENV_FILE}"
  content="${content//@PORT@/$SERVER_PORT}"
  content="${content//@HOST@/$SERVER_HOST}"
  content="${content//@WORKERS@/$WORKERS}"
  printf '%s' "$content"
}
for f in start.sh stop.sh restart.sh status.sh healthcheck.sh uninstall.sh; do
  replace_vars "$TEMPLATES_DIR/$f" > "$BIN_DIR/$f"
  chmod +x "$BIN_DIR/$f"
done
# ---------- systemd ----------
RUN_USER=""
RUN_GROUP=""
if [ "$USE_SYSTEMD" -eq 1 ]; then
  if [ "$(id -u)" != "0" ]; then
    fail "--systemd 需要 root 权限"
  fi
  if ! command -v systemctl >/dev/null 2>&1; then
    fail "系统不支持 systemctl"
  fi
  RUN_USER="${SUDO_USER:-root}"
  RUN_GROUP="$(id -gn "$RUN_USER" 2>/dev/null || echo "$RUN_USER")"
  log "调整安装目录属主为 $RUN_USER:$RUN_GROUP ..."
  chown -R "$RUN_USER":"$RUN_GROUP" "$INSTALL_DIR"
  chmod 600 "$ENV_FILE"
  SERVICE_CONTENT="$(replace_vars "$TEMPLATES_DIR/platform-backend.service")"
  SERVICE_CONTENT="${SERVICE_CONTENT//@RUN_USER@/$RUN_USER}"
  SERVICE_CONTENT="${SERVICE_CONTENT//@RUN_GROUP@/$RUN_GROUP}"
  printf '%s' "$SERVICE_CONTENT" > /etc/systemd/system/platform-backend.service
  chmod 644 /etc/systemd/system/platform-backend.service
  systemctl daemon-reload
  systemctl enable platform-backend
  log "已安装 systemd 服务：platform-backend（运行用户：$RUN_USER:$RUN_GROUP）"
fi
# ---------- 启动 ----------
if [ "$START_SERVICE" -eq 1 ]; then
  if [ "$USE_SYSTEMD" -eq 1 ]; then
    systemctl restart platform-backend
    log "已通过 systemd 启动服务"
  else
    "$BIN_DIR/start.sh"
  fi
else
  log "已跳过启动。可执行：$BIN_DIR/start.sh"
fi
# ---------- 版本号 ----------
INSTALLED_VERSION="unknown"
if [ -f "$VERSION_FILE" ]; then
  INSTALLED_VERSION="$(cat "$VERSION_FILE" | tr -d '[:space:]')"
fi
# ---------- 安装摘要 ----------
cat <<EOF
============================================================
✅ 安装完成
============================================================
产品版本：        $INSTALLED_VERSION
安装目录：        $INSTALL_DIR
后端目录：        $BACKEND_DIR
前端目录：        $FRONTEND_DEPLOY_DIR
配置文件：        $ENV_FILE
数据目录：        $DATA_DIR
日志文件：        $LOG_DIR/backend.log
PID 文件：        $RUN_DIR/backend.pid
监听地址：        $SERVER_HOST:$SERVER_PORT
访问地址：        http://127.0.0.1:${SERVER_PORT}/
健康检查：        http://127.0.0.1:${SERVER_PORT}/health
API 文档：        http://127.0.0.1:${SERVER_PORT}/docs
控制命令：
  启动：          $BIN_DIR/start.sh
  停止：          $BIN_DIR/stop.sh
  重启：          $BIN_DIR/restart.sh
  状态：          $BIN_DIR/status.sh
  健康检查：      $BIN_DIR/healthcheck.sh
  卸载：          $BIN_DIR/uninstall.sh [--purge]
默认账号：        admin / 123456
首次登录后请立即修改密码。
生产环境建议修改 CORS_ORIGINS、配置 NTP、定期备份 .env 与 data/app.db。
如需修改参数，请编辑 $ENV_FILE 后执行 $BIN_DIR/restart.sh。
============================================================
EOF
