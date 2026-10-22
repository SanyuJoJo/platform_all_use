#!/usr/bin/env bash
# =============================================================================
# deploy.sh —— 将前端构建产物复制到指定部署目录
#
# 用法:
#   ./scripts/deploy.sh [选项]
#
# 选项:
#   -d, --dest <dir>     部署目录（默认: ./deploy）
#   -s, --skip-build     跳过 pnpm build:all，直接复制现有 dist
#   -c, --clean          复制前清空 deploy/main-app 与 deploy/sub-apps
#   -m, --mode <mode>    构建命令: all|main|sub（默认 all，仅 --skip-build 未指定时生效）
#   -h, --help           显示帮助
#
# 示例:
#   ./scripts/deploy.sh
#   ./scripts/deploy.sh -d /var/www/platform -c
#   ./scripts/deploy.sh -s -d ./deploy
# =============================================================================

set -euo pipefail

# ---------- 路径 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
MAIN_APP_DIR="${FRONTEND_DIR}/main-app"
SUB_APPS_DIR="${FRONTEND_DIR}/sub-apps"

# ---------- 默认参数 ----------
DEST_DIR="${FRONTEND_DIR}/deploy"
SKIP_BUILD=0
CLEAN=0
BUILD_MODE="all"

# ---------- 颜色 ----------
if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'; C_INFO=$'\033[36m'; C_OK=$'\033[32m'
  C_WARN=$'\033[33m'; C_ERR=$'\033[31m'; C_BOLD=$'\033[1m'
else
  C_RESET=""; C_INFO=""; C_OK=""; C_WARN=""; C_ERR=""; C_BOLD=""
fi

log()   { printf '%s[deploy]%s %s\n' "${C_INFO}" "${C_RESET}" "$*"; }
ok()    { printf '%s[deploy]%s %s\n' "${C_OK}"   "${C_RESET}" "$*"; }
warn()  { printf '%s[deploy]%s %s\n' "${C_WARN}" "${C_RESET}" "$*" >&2; }
err()   { printf '%s[deploy]%s %s\n' "${C_ERR}"  "${C_RESET}" "$*" >&2; }

usage() {
  sed -n '2,30p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit 0
}

# ---------- 解析参数 ----------
while [[ $# -gt 0 ]]; do
  case "$1" in
    -d|--dest)        DEST_DIR="$2"; shift 2 ;;
    -s|--skip-build)  SKIP_BUILD=1;   shift ;;
    -c|--clean)       CLEAN=1;        shift ;;
    -m|--mode)        BUILD_MODE="$2"; shift 2 ;;
    -h|--help)        usage ;;
    *) err "未知参数: $1"; usage ;;
  esac
done

# 转绝对路径（dest 可能不存在，用 -m）
DEST_DIR="$(cd "$(dirname "${DEST_DIR}")" 2>/dev/null && pwd || echo "$(dirname "${DEST_DIR}")")/$(basename "${DEST_DIR}")"

# ---------- 工具函数 ----------
need_cmd() { command -v "$1" >/dev/null 2>&1 || { err "缺少命令: $1"; exit 1; }; }

# 使用 rsync（可保留删除语义），没有则降级 cp
copy_dir() {
  local src="$1" dst="$2"
  mkdir -p "${dst}"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete "${src}/" "${dst}/"
  else
    rm -rf "${dst:?}"/*
    cp -a "${src}/." "${dst}/"
  fi
}

# module id -> 目录名（与 main-app/src/micro-frontend/registry.ts 保持一致）
dir_name() {
  case "$1" in
    module_manager) echo "module-manager" ;;
    audit_log)      echo "audit-log"      ;;
    crypto_console) echo "crypto-console" ;;
    ssl_proxy)      echo "ssl-proxy"      ;;
    *)              echo "${1//_/-}"      ;;
  esac
}

# ---------- 前置检查 ----------
log "前端目录: ${FRONTEND_DIR}"
log "部署目录: ${DEST_DIR}"

need_cmd pnpm
need_cmd node

if [[ "${SKIP_BUILD}" -eq 0 ]]; then
  case "${BUILD_MODE}" in
    all)  BUILD_CMD="pnpm build:all" ;;
    main) BUILD_CMD="pnpm --filter main-app build" ;;
    sub)  BUILD_CMD="pnpm --filter './sub-apps/*' --filter '!./sub-apps/_template' build" ;;
    *)    err "非法 --mode: ${BUILD_MODE}（可选 all|main|sub）"; exit 1 ;;
  esac
  log "执行构建: ${BUILD_CMD}"
  ( cd "${FRONTEND_DIR}" && eval "${BUILD_CMD}" )
  ok "构建完成"
else
  warn "已指定 --skip-build，跳过构建"
fi

# ---------- 校验源产物 ----------
if [[ ! -f "${MAIN_APP_DIR}/dist/index.html" ]]; then
  err "主应用产物缺失: ${MAIN_APP_DIR}/dist/index.html"
  err "请先执行 pnpm build:all，或去掉 --skip-build"
  exit 1
fi

# ---------- 准备部署目录 ----------
mkdir -p "${DEST_DIR}"
DEPLOY_MAIN="${DEST_DIR}/main-app"
DEPLOY_SUBS="${DEST_DIR}/sub-apps"

if [[ "${CLEAN}" -eq 1 ]]; then
  log "清理旧产物: ${DEPLOY_MAIN}, ${DEPLOY_SUBS}"
  rm -rf "${DEPLOY_MAIN}" "${DEPLOY_SUBS}"
fi

# ---------- 复制主应用 ----------
log "复制主应用: main-app/dist -> ${DEPLOY_MAIN}"
copy_dir "${MAIN_APP_DIR}/dist" "${DEPLOY_MAIN}"
ok "主应用复制完成"

# ---------- 复制子应用 ----------
if [[ ! -d "${SUB_APPS_DIR}" ]]; then
  err "子应用目录不存在: ${SUB_APPS_DIR}"
  exit 1
fi

COPIED=0
SKIPPED=0
FAILED=0

for app_dir in "${SUB_APPS_DIR}"/*/; do
  app_name="$(basename "${app_dir}")"

  # 跳过模板与隐藏目录
  [[ "${app_name}" == _template ]] && { warn "跳过模板: ${app_name}"; continue; }
  [[ "${app_name}" == .* ]]        && continue
  [[ -d "${app_dir}" ]] || continue

  src_dist="${app_dir}dist"
  dst_dir="${DEPLOY_SUBS}/$(dir_name "${app_name}")"

  if [[ ! -f "${src_dist}/index.html" ]]; then
    warn "跳过 ${app_name}: 缺少 ${src_dist}/index.html（构建失败？）"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  log "复制子应用: ${app_name}/dist -> ${dst_dir}"
  if copy_dir "${src_dist}" "${dst_dir}"; then
    ok "  ${app_name} ✓"
    COPIED=$((COPIED + 1))
  else
    err "  ${app_name} ✗"
    FAILED=$((FAILED + 1))
  fi
done

# ---------- 汇总 ----------
echo
ok "${C_BOLD}部署完成${C_RESET}"
log "  主应用: ${DEPLOY_MAIN}"
log "  子应用: ${DEPLOY_SUBS}（成功 ${COPIED} / 跳过 ${SKIPPED} / 失败 ${FAILED}）"
log "  目录清单:"
( cd "${DEPLOY_SUBS}" && ls -1 | sed 's/^/    - /' ) || true

if [[ "${FAILED}" -gt 0 ]]; then
  err "存在 ${FAILED} 个子应用复制失败"
  exit 1
fi

# ---------- 校验 publicPath 与目录名一致性 ----------
log "校验子应用 index.html 的 publicPath"
MISMATCH=0
for d in "${DEPLOY_SUBS}"/*/; do
  name="$(basename "${d}")"
  if ! grep -q "/sub-apps/${name}/" "${d}/index.html" 2>/dev/null; then
    warn "  ${name}: index.html 中未发现 /sub-apps/${name}/ 前缀，请确认 vite base 配置"
    MISMATCH=$((MISMATCH + 1))
  fi
done
[[ "${MISMATCH}" -eq 0 ]] && ok "publicPath 校验通过"

exit 0
