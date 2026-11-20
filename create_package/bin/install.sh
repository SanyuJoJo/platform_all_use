#!/usr/bin/env bash
# ============================================================
# 产品安装脚本（Go 版 v1.4）
#
# v1.1 变更：
#   - P2-03：SQLite 路径解析兼容单引号 / 双引号 / 无引号
# v1.3 变更：
#   - P1-FE-02：INSTALL_DIR 强制转为绝对路径
#   - P1-FE-03：前端载荷校验增强（main-app/assets + 每个子应用）
#   - P2-FE-09：安装摘要增加前端信息
# v1.4 变更（安装包自带 core，单一安装根 /opt/plats_tool）：
#   - 默认安装目录改为 /opt/plats_tool（backend/frontend/core/user_data 全部在其下）
#   - 安装 core 载荷（core.tar.gz：core-go + Tengine + 铜锁 + oqs-provider）
#   - 升级时保留 core/data、core/logs、core/pid、core/tmp、conf/proxy-*、conf/tcps
#   - .env 写入 PLATS_INSTALL_ROOT / PLATS_DEPLOY_ROOT / PLATS_CORE_GO_BIN / CORE_ROOT
# v1.5 变更（安装包自带 ZLMediaKit 音视频）：
#   - 安装 video 载荷（video.tar.gz：MediaServer + libmk_api.so + config.ini）
#   - .env 写入 ZLM_HTTP_PORT=8080 / ZLM_RTMP_PORT=1935 / ZLM_RTC_PORT=8000
#   - 后端启动时 ZLM_AUTO_START=true 自动托管拉起 MediaServer
# v1.6 变更（公网部署音视频）：
#   - 新增 --zlm-host IP：写入 .env 的 ZLM_HTTP_HOST 与 ZLM config.ini 的 [rtc] externIP
#   - 公网部署必须指定公网 IP，否则自动探测到内网 IP，浏览器无法推流/拉流
#   - 需开放端口：8080/tcp（HTTP 信令/拉流）、1935/tcp（RTMP）、18000/tcp+udp（WebRTC 媒体）、
#     49152-65535/udp（TURN 中继，若开启 enableTurn=1）
# ============================================================
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PAYLOAD_DIR="$SCRIPT_DIR/bin"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
VERSION_FILE="$SCRIPT_DIR/VERSION"
# ★ v1.4：单一安装根目录，默认 /opt/plats_tool（backend/frontend/core/user_data 统一）
INSTALL_DIR="${INSTALL_DIR:-/opt/plats_tool}"
SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
SERVER_PORT="${SERVER_PORT:-8000}"
WORKERS="${WORKERS:-1}"
CORS_ORIGINS="${CORS_ORIGINS:-*}"
ZLM_HOST="${ZLM_HOST:-}"
ZLM_SECRET="${ZLM_SECRET:-}"
START_SERVICE=1
USE_SYSTEMD=0
FORCE=0
log()  { echo -e "[install] $*"; }
warn() { echo -e "[install][WARN] $*" >&2; }
fail() { echo -e "[install][ERROR] $*" >&2; exit 1; }
usage() {
  cat <<EOF
用法：$0 [选项]
  --install-dir PATH   安装目录，默认 /opt/plats_tool
  --port PORT          后端端口，默认 8000
  --host HOST          监听地址，默认 0.0.0.0
  --workers N          worker 数（Go 版保留兼容，不生效），默认 1
  --cors ORIGINS       CORS，默认 *
  --zlm-host IP        音视频对外可达的公网 IP（浏览器访问地址，如 203.0.113.10）
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
    --zlm-host) ZLM_HOST="$2"; shift 2 ;;
    --no-start) START_SERVICE=0; shift ;;
    --systemd) USE_SYSTEMD=1; shift ;;
    --force) FORCE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) fail "未知参数：$1" ;;
  esac
done
# ---------- 非 root 回退 ----------
if [ "$(id -u)" != "0" ] && [ "$INSTALL_DIR" = "/opt/plats_tool" ]; then
  INSTALL_DIR="$HOME/plats_tool"
  log "非 root 用户，安装目录回退为：$INSTALL_DIR"
fi
# ---------- v1.3 P1-FE-02：强制绝对路径 ----------
mkdir -p "$INSTALL_DIR"
if ! INSTALL_DIR="$(cd "$INSTALL_DIR" && pwd)"; then
  fail "无法解析 INSTALL_DIR 绝对路径：$INSTALL_DIR"
fi
log "安装目录（绝对路径）：$INSTALL_DIR"
BACKEND_DIR="$INSTALL_DIR/backend"
FRONTEND_DIR="$INSTALL_DIR/frontend"
FRONTEND_DEPLOY_DIR="$FRONTEND_DIR/deploy"
DATA_DIR="$INSTALL_DIR/data"
LOG_DIR="$INSTALL_DIR/logs"
UPLOAD_DIR="$INSTALL_DIR/uploads"
RUN_DIR="$INSTALL_DIR/run"
BIN_DIR="$INSTALL_DIR/bin"
CORE_DIR="$INSTALL_DIR/core"
CORE_ROOT="$CORE_DIR"
VIDEO_DIR="$INSTALL_DIR/video"
ENV_FILE="$BACKEND_DIR/.env"
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
[ -f "$PAYLOAD_DIR/core.tar.gz" ] || fail "缺少 core 包：$PAYLOAD_DIR/core.tar.gz"
[ -f "$PAYLOAD_DIR/video.tar.gz" ] || fail "缺少音视频包：$PAYLOAD_DIR/video.tar.gz"
[ -f "$PAYLOAD_DIR/nshttps_gm.tar.gz" ] || fail "缺少统一入口代理包：$PAYLOAD_DIR/nshttps_gm.tar.gz"
[ -d "$TEMPLATES_DIR" ] || fail "缺少模板目录：$TEMPLATES_DIR"
# ---------- 创建目录 ----------
mkdir -p "$BACKEND_DIR" "$FRONTEND_DEPLOY_DIR" "$DATA_DIR" "$LOG_DIR" \
         "$UPLOAD_DIR/modules" "$RUN_DIR" "$BIN_DIR"
# ★ v1.4：core 与 user_data（证书/日志/媒体）骨架
mkdir -p "$CORE_DIR"
for d in data logs tmp pid run conf client_body_temp fastcgi_temp proxy_temp scgi_temp uwsgi_temp; do
  mkdir -p "$CORE_DIR/$d"
done
mkdir -p "$INSTALL_DIR/user_data"/{cert,ca,server,conf,logs,media,run}
chmod 700 "$INSTALL_DIR/user_data/cert" 2>/dev/null || true
# ★ v1.5：音视频（ZLMediaKit）骨架
mkdir -p "$VIDEO_DIR"/{run/bin,run/log,run/www,libs/ZLMediaKit,conf/zlmediakit,logs}
# ---------- 解压后端 + 结构校验 ----------
log "解压后端..."
tar -xzf "$PAYLOAD_DIR/backend.tar.gz" -C "$BACKEND_DIR"
[ -x "$BACKEND_DIR/bin/server" ] || fail "后端包缺少可执行文件 bin/server"
[ -d "$BACKEND_DIR/migrations" ] || fail "后端包缺少 migrations/"
[ -f "$BACKEND_DIR/.env.production.example" ] || fail "后端包缺少 .env.production.example"
# ---------- 解压前端 + 结构校验（v1.3 P1-FE-03 增强） ----------
log "解压前端..."
rm -rf "$FRONTEND_DEPLOY_DIR"
mkdir -p "$FRONTEND_DEPLOY_DIR"
if command -v unzip >/dev/null 2>&1 && unzip -tq "$PAYLOAD_DIR/dist.war" >/dev/null 2>&1; then
  unzip -q "$PAYLOAD_DIR/dist.war" -d "$FRONTEND_DEPLOY_DIR"
else
  tar -xzf "$PAYLOAD_DIR/dist.war" -C "$FRONTEND_DEPLOY_DIR"
fi
# validate_frontend_deploy: 校验前端产物完整性
validate_frontend_deploy() {
  local root="$1"
  [ -f "$root/main-app/index.html" ] \
    || fail "前端包缺少 main-app/index.html"
  [ -d "$root/main-app/assets" ] \
    || fail "前端包缺少 main-app/assets/"
  [ -d "$root/sub-apps" ] \
    || fail "前端包缺少 sub-apps/ 目录"
  local found=0
  local d
  for d in "$root"/sub-apps/*/; do
    [ -d "$d" ] || continue
    found=1
    local name
    name="$(basename "$d")"
    [ -f "${d}index.html" ] \
      || fail "前端包子应用缺少 index.html: sub-apps/${name}/index.html"
    [ -d "${d}assets" ] \
      || fail "前端包子应用缺少 assets/: sub-apps/${name}/assets/"
    log "  ✅ sub-apps/${name}/（index.html + assets/）"
  done
  [ "$found" -eq 1 ] \
    || fail "前端包 sub-apps/ 下没有任何子应用"
  if [ -f "$root/deploy-manifest.json" ]; then
    log "  ✅ deploy-manifest.json"
  fi
}
validate_frontend_deploy "$FRONTEND_DEPLOY_DIR"
SUB_APP_COUNT="$(find "$FRONTEND_DEPLOY_DIR/sub-apps" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
log "前端子应用数量：$SUB_APP_COUNT"
# ---------- core 解压 + 合并（v1.4） ----------
log "解压 core 载荷..."
TMP_CORE="$(mktemp -d)"
tar -xzf "$PAYLOAD_DIR/core.tar.gz" -C "$TMP_CORE" \
  || fail "core.tar.gz 不是合法 tar.gz"
[ -x "$TMP_CORE/core/bin/core" ] || fail "core 载荷缺少可执行文件 bin/core"
[ -x "$TMP_CORE/core/sbin/nginx" ] || fail "core 载荷缺少可执行文件 sbin/nginx"
[ -f "$TMP_CORE/core/libs/tongsuo/bin/openssl" ] \
  || fail "core 载荷缺少铜锁 openssl"
# 升级时保留运行时数据与动态生成的配置
for p in "$CORE_DIR"/conf/proxy-*; do
  [ -e "$p" ] || continue
  mkdir -p "$TMP_CORE/core/conf"
  mv "$p" "$TMP_CORE/core/conf/" 2>/dev/null || true
done
if [ -d "$CORE_DIR/conf/tcps" ]; then
  mkdir -p "$TMP_CORE/core/conf"
  mv "$CORE_DIR/conf/tcps" "$TMP_CORE/core/conf/"
fi
# 替换代码 / 静态配置目录
rm -rf "$CORE_DIR/bin" "$CORE_DIR/sbin" "$CORE_DIR/libs"
rm -f "$CORE_DIR/conf/core.yaml" "$CORE_DIR/conf/algorithm-whitelist.yaml" \
      "$CORE_DIR/conf/mime.types"
rm -rf "$CORE_DIR/conf/template"
# 合并（cp -a 保留既有 data/logs/tmp/pid/run 内容）
cp -a "$TMP_CORE/core/." "$CORE_DIR/"
rm -rf "$TMP_CORE"
# core 可执行性复核
[ -x "$CORE_DIR/bin/core" ] || fail "core 安装后 bin/core 不可执行"
log "✅ core 已安装：$CORE_DIR"
# ---------- 重装前清理（v1.8）----------
# 旧 ZLM 进程会用旧 secret/旧 externIP 继续运行，而 config.ini 已被新载荷覆盖，
# 导致 backend 调 ZLM 鉴权失败（getMediaList: Please login first），
# 表现为所有直播间「未推流」、拉流异常。因此重装前必须先停旧服务。
if [ -f "$BIN_DIR/stop.sh" ]; then
  log "停止旧服务（stop.sh）..."
  bash "$BIN_DIR/stop.sh" >/dev/null 2>&1 || true
fi
if command -v pkill >/dev/null 2>&1; then
  pkill -f "MediaServer -c $VIDEO_DIR/conf/zlmediakit/config.ini" 2>/dev/null || true
  sleep 1
fi
# ★ v1.10：未传 --zlm-host 时自动探测公网 IP（阿里云 EIP 元数据优先，
#   其次通用公网回显服务），避免部署到公网机却回落到内网 IP。
detect_public_ip() {
  local ip=""
  if command -v curl >/dev/null 2>&1; then
    ip="$(curl -fsS --max-time 3 http://100.100.100.200/latest/meta-data/eipv4 2>/dev/null | tr -d '[:space:]' || true)"
    if [ -z "$ip" ]; then
      ip="$(curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null | tr -d '[:space:]' || true)"
    fi
    if [ -z "$ip" ]; then
      ip="$(curl -fsS --max-time 5 http://ifconfig.me/ip 2>/dev/null | tr -d '[:space:]' || true)"
    fi
  fi
  case "$ip" in
    ""|10.*|192.168.*|172.1[6-9].*|172.2[0-9].*|172.3[01].*|127.*) echo "" ;;
    *[!0-9.]*) echo "" ;;
    *) echo "$ip" ;;
  esac
}
# 捕获重装前的公网 host，避免本次未传 --zlm-host 时丢失已配置的公网地址
PREV_ZLM_HOST="$(grep -E '^ZLM_HTTP_HOST=' "$ENV_FILE" 2>/dev/null | head -1 \
  | sed -E 's/^ZLM_HTTP_HOST=//; s/^"//; s/"$//' || true)"
PREV_EXTERN_IP="$(grep -E '^externIP=' "$VIDEO_DIR/conf/zlmediakit/config.ini" 2>/dev/null \
  | head -1 | sed -E 's/^externIP=//' || true)"
if [ -z "$ZLM_HOST" ]; then
  if [ -n "$PREV_ZLM_HOST" ]; then
    ZLM_HOST="$PREV_ZLM_HOST"
    log "未指定 --zlm-host，沿用已有 ZLM_HTTP_HOST=$ZLM_HOST"
  elif [ -n "$PREV_EXTERN_IP" ]; then
    ZLM_HOST="$PREV_EXTERN_IP"
    log "未指定 --zlm-host，沿用已有 externIP=$ZLM_HOST"
  else
    DETECTED_IP="$(detect_public_ip || true)"
    if [ -n "$DETECTED_IP" ]; then
      ZLM_HOST="$DETECTED_IP"
      log "未指定 --zlm-host，自动探测公网 IP=$ZLM_HOST"
    else
      log "WARN 未指定 --zlm-host 且自动探测失败，ZLM 将使用内网地址；公网拉流可能失败，请传 --zlm-host <公网IP>"
    fi
  fi
fi
# ---------- 音视频（ZLMediaKit）解压（v1.5） ----------
log "解压音视频载荷..."
TMP_VIDEO="$(mktemp -d)"
tar -xzf "$PAYLOAD_DIR/video.tar.gz" -C "$TMP_VIDEO" \
  || fail "video.tar.gz 不是合法 tar.gz"
[ -x "$TMP_VIDEO/video/run/bin/MediaServer" ] \
  || fail "音视频载荷缺少可执行文件 run/bin/MediaServer"
[ -f "$TMP_VIDEO/video/conf/zlmediakit/config.ini" ] \
  || fail "音视频载荷缺少 conf/zlmediakit/config.ini"
# 保留既有日志 / 运行期数据
rm -rf "$VIDEO_DIR/run/bin" "$VIDEO_DIR/libs" "$VIDEO_DIR/conf"
cp -a "$TMP_VIDEO/video/." "$VIDEO_DIR/"
rm -rf "$TMP_VIDEO"
[ -x "$VIDEO_DIR/run/bin/MediaServer" ] || fail "音视频安装后 MediaServer 不可执行"
# ★ v1.6：--zlm-host 指定的公网 IP 写入 ZLM config.ini 的 [rtc] externIP，
#   使 WebRTC SDP 返回公网 candidate，浏览器才能直连 ZLM 传音视频。
# ★ v1.7：预生成 ZLM secret 并写入 config.ini，避免 backend 启动时
#   检测到官方示例 secret 后写回新值，但 ZLM 已用旧值运行导致鉴权失败
#   （生产复现：zlm getMediaList 错误: Please login first）。
gen_zlm_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 20
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c 'import secrets; print(secrets.token_hex(20))'
  else
    echo ""
  fi
}
ZLM_INI="$VIDEO_DIR/conf/zlmediakit/config.ini"
if [ -f "$ZLM_INI" ]; then
  if [ -z "$ZLM_SECRET" ]; then
    ZLM_SECRET="$(gen_zlm_secret)"
  fi
  if [ -n "$ZLM_SECRET" ]; then
    python3 - "$ZLM_INI" "$ZLM_SECRET" <<'PYEOF'
import re, sys
p, secret = sys.argv[1], sys.argv[2]
nl = chr(10)
lines = open(p, encoding="utf-8").read().split(nl)
in_api = False
found = False
for i, ln in enumerate(lines):
    line = ln.strip()
    if line.startswith("[") and line.endswith("]"):
        in_api = (line == "[api]")
        continue
    if in_api and re.match(r"^secret\s*=\s*.*$", line):
        indent = ln[:len(ln) - len(ln.lstrip())]
        lines[i] = indent + "secret=" + secret
        found = True
        break
open(p, "w", encoding="utf-8").write(nl.join(lines))
if not found:
    raise SystemExit("!! config.ini 中未找到 [api] secret 行")
PYEOF
    log "✅ ZLM [api] secret 已预生成（${ZLM_SECRET:0:8}...）"
  fi
  # v1.6：公网 IP 写入 [rtc] externIP
  if [ -n "$ZLM_HOST" ]; then
    python3 - "$ZLM_INI" "$ZLM_HOST" <<'PYEOF'
import re, sys
p, host = sys.argv[1], sys.argv[2]
nl = chr(10)
lines = open(p, encoding="utf-8").read().split(nl)
in_rtc = False
found = False
for i, ln in enumerate(lines):
    s = ln.strip()
    if s.startswith("[") and s.endswith("]"):
        in_rtc = (s == "[rtc]")
        continue
    if in_rtc and re.match(r"^externIP\s*=\s*.*$", s):
        indent = ln[:len(ln) - len(ln.lstrip())]
        lines[i] = indent + "externIP=" + host
        found = True
        break
open(p, "w", encoding="utf-8").write(nl.join(lines))
if not found:
    raise SystemExit("!! config.ini 中未找到 [rtc] externIP 行")
PYEOF
    log "✅ ZLM [rtc] externIP=${ZLM_HOST}（已写入 config.ini）"
  fi
  # ★ v1.9：[http] allow_ip_range 默认只允许私网 IP，会导致公网浏览器
  #   访问 8080 的 HLS/FLV/WebRTC 信令被拒（-100 Your ip is not allowed），
  #   表现为拉流异常。置空 = 不限制（敏感 API 仍由 secret 保护）。
  if grep -qE '^allow_ip_range=' "$ZLM_INI" 2>/dev/null; then
    sed -i -E 's|^allow_ip_range=.*|allow_ip_range=|' "$ZLM_INI"
    log "✅ ZLM [http] allow_ip_range 已置空（允许公网拉流）"
  fi
fi
log "✅ 音视频已安装：$VIDEO_DIR"
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
# ---------- 安装根 / core（v1.4：安装包自带 core）----------
# 单一安装根：backend / frontend / core / user_data 全部在 $INSTALL_DIR 下
PLATS_INSTALL_ROOT="${INSTALL_DIR}"
PLATS_DEPLOY_ROOT="${INSTALL_DIR}"
PLATS_CORE_GO_BIN="${CORE_DIR}/bin/core"
CORE_ROOT="${CORE_DIR}"
# ---------- 音视频（ZLMediaKit）----------
# 后端启动时若 ZLM 探活失败则自动托管拉起；部署目录为 <PLATS_INSTALL_ROOT>/video/
ZLM_AUTO_START="true"
ZLM_ENGINE_SHUTDOWN_ON_EXIT="true"
# 以下留空则后端自动从 config.ini 读取 http 端口 / api secret，
# 并自动派生 ZLM_BASE_URL=http://127.0.0.1:<port>
ZLM_HTTP_PORT="8080"
ZLM_SECRET="${ZLM_SECRET}"
ZLM_BASE_URL=
# 对外可达 host（拼播放地址用，浏览器必须能访问）。
# 留空时后端自动取本机第一个非回环 IPv4（如 192.0.2.1）；
# 多网卡/域名部署时请显式指定。
ZLM_HTTP_HOST="${ZLM_HOST}"
ZLM_RTMP_PORT="1935"
ZLM_RTC_PORT="8000"
EOF
  chmod 600 "$ENV_FILE"
  log "已生成配置文件：$ENV_FILE"
else
  # ★ v1.4：升级时若旧 .env 缺少安装根/core 键，追加（不覆盖已有值）
  append_env() {
    local key="$1" val="$2"
    if ! grep -qE "^${key}=" "$ENV_FILE" 2>/dev/null; then
      printf '%s="%s"
' "$key" "$val" >> "$ENV_FILE"
      log "已追加 ${key}=${val}"
    fi
  }
  append_env "PLATS_INSTALL_ROOT"  "$INSTALL_DIR"
  append_env "PLATS_DEPLOY_ROOT"   "$INSTALL_DIR"
  append_env "PLATS_CORE_GO_BIN"   "$CORE_DIR/bin/core"
  append_env "CORE_ROOT"           "$CORE_DIR"
  append_env "ZLM_HTTP_PORT"       "8080"
  append_env "ZLM_RTMP_PORT"       "1935"
  append_env "ZLM_RTC_PORT"        "8000"
  if [ -n "$ZLM_HOST" ]; then
    append_env "ZLM_HTTP_HOST"     "$ZLM_HOST"
  fi
  if [ -n "$ZLM_SECRET" ]; then
    append_env "ZLM_SECRET"        "$ZLM_SECRET"
  fi
  chmod 600 "$ENV_FILE" 2>/dev/null || true
fi
# ★ v1.8：强制 .env 的 ZLM_SECRET / ZLM_HTTP_HOST 与本次安装一致。
#   升级安装时 .env 会被保留，若不强制同步，会出现
#   .env 旧 secret != config.ini 新 secret（ZLM 鉴权失败）或公网 host 丢失。
set_env_key() {
  local key="$1" val="$2"
  [ -f "$ENV_FILE" ] || return 0
  if grep -qE "^${key}=" "$ENV_FILE" 2>/dev/null; then
    sed -i -E "s|^${key}=.*|${key}=\"${val}\"|" "$ENV_FILE"
  else
    printf '%s="%s"\n' "$key" "$val" >> "$ENV_FILE"
  fi
}
if [ -n "$ZLM_SECRET" ] && [ -f "$ENV_FILE" ]; then
  set_env_key "ZLM_SECRET" "$ZLM_SECRET"
  log "✅ .env ZLM_SECRET 已与 config.ini 同步"
fi
if [ -n "$ZLM_HOST" ] && [ -f "$ENV_FILE" ]; then
  set_env_key "ZLM_HTTP_HOST" "$ZLM_HOST"
  log "✅ .env ZLM_HTTP_HOST=${ZLM_HOST}"
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
# ---------- 统一入口代理 nshttps_gm（v1.7） ----------
NSG_DIR="$INSTALL_DIR/nshttps_gm"
log "解压统一入口代理 nshttps_gm..."
rm -rf "$NSG_DIR"
mkdir -p "$NSG_DIR"
TMP_NSG="$(mktemp -d)"
tar -xzf "$PAYLOAD_DIR/nshttps_gm.tar.gz" -C "$TMP_NSG" 2>/dev/null \
  || { rm -rf "$TMP_NSG"; fail "nshttps_gm.tar.gz 解压失败"; }
# 部署 runtime 与脚本
cp -a "$TMP_NSG/runtime/." "$NSG_DIR/"
cp -a "$TMP_NSG/scripts/." "$NSG_DIR/"
rm -rf "$TMP_NSG"
mkdir -p "$NSG_DIR/conf" "$NSG_DIR/logs" "$NSG_DIR/pid" "$NSG_DIR/tmp" "$NSG_DIR/run" "$NSG_DIR/certs"
for f in gen-certs.sh gen-config.sh start.sh stop.sh status.sh; do
  chmod +x "$NSG_DIR/$f" 2>/dev/null || true
done
chmod +x "$NSG_DIR/tengine-3.1.0/bin/sbin/nginx" 2>/dev/null || true
NSG_UPSTREAM="127.0.0.1:${SERVER_PORT}"
# 自动生成证书与配置，并启动代理（把业务流量代理到后端，监听 0.0.0.0:8443）
if [ "$START_SERVICE" -eq 1 ]; then
  if [ -x "$NSG_DIR/gen-certs.sh" ]; then
    log "生成 nshttps_gm 国密证书（CN=localhost，10 年）..."
    (cd "$NSG_DIR" && bash gen-certs.sh --cn localhost >/dev/null 2>&1) \
      && log "✅ nshttps_gm 证书已生成" \
      || warn "nshttps_gm 证书生成失败（可稍后手动执行 gen-certs.sh）"
  fi
  if [ -x "$NSG_DIR/gen-config.sh" ] && [ -f "$NSG_DIR/certs/server/server.crt" ]; then
    log "生成 nshttps_gm 代理配置（8443 → ${NSG_UPSTREAM}）..."
    (cd "$NSG_DIR" && bash gen-config.sh --port 8443 --upstream "$NSG_UPSTREAM" >/dev/null 2>&1) \
      && log "✅ nshttps_gm 配置已生成" \
      || warn "nshttps_gm 配置生成失败"
  fi
  if [ -x "$NSG_DIR/start.sh" ] && [ -f "$NSG_DIR/conf/nshttps.conf" ]; then
    log "启动 nshttps_gm 代理..."
    (cd "$NSG_DIR" && bash start.sh >/dev/null 2>&1) \
      && log "✅ nshttps_gm 代理已启动（监听 0.0.0.0:8443，HTTPS 统一入口）" \
      || warn "nshttps_gm 代理启动失败（可稍后手动执行 $NSG_DIR/start.sh）"
  fi
else
  log "已跳过 nshttps_gm 启动（--no-start）"
  log "  代理部署于：$NSG_DIR"
  log "  手动启动：cd $NSG_DIR && bash gen-certs.sh && bash gen-config.sh --upstream $NSG_UPSTREAM && bash start.sh"
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
# ---------- 安装摘要（v1.3 P2-FE-09 增强） ----------
FRONTEND_MAIN_STATUS="❌ 缺失"
if [ -f "$FRONTEND_DEPLOY_DIR/main-app/index.html" ]; then
  FRONTEND_MAIN_STATUS="✅ 存在"
fi
cat <<EOF
============================================================
✅ 安装完成
============================================================
产品版本：        $INSTALLED_VERSION
安装目录：        $INSTALL_DIR
Core 引擎：       $CORE_DIR（core-go / Tengine / 铜锁）
音视频引擎：      $VIDEO_DIR（ZLMediaKit，HTTP=8080${ZLM_HOST:+ / 公网 $ZLM_HOST}）
后端目录：        $BACKEND_DIR
前端目录：        $FRONTEND_DEPLOY_DIR
前端主应用：      $FRONTEND_MAIN_STATUS
前端子应用数：    $SUB_APP_COUNT
统一入口代理：    $NSG_DIR（nshttps_gm，HTTPS 监听 0.0.0.0:8443）
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
★ 统一入口代理端口放行（公网安全组需开放）：
  - 8443/tcp  nshttps_gm HTTPS 统一入口（前端/业务流量走 SSL）
★ 音视频端口放行（公网安全组需开放）：
  - 8080/tcp       ZLM HTTP 信令 + HLS/FLV 拉流
  - 1935/tcp       RTMP 推流/拉流（OBS）
  - 18000/tcp+udp  WebRTC 音视频媒体
  - 49152-65535/udp TURN 中继（若 ZLM enableTurn=1）
首次登录后请立即修改密码。
生产环境建议修改 CORS_ORIGINS、配置 NTP、定期备份 .env 与 data/app.db。
如需修改参数，请编辑 $ENV_FILE 后执行 $BIN_DIR/restart.sh。
============================================================
EOF
