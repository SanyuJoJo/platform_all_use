#!/usr/bin/env bash
# =============================================================================
# core 临时部署脚本（含 SSL 代理 + oqs-provider + ZLMediaKit + libsrtp）
#
# 说明：
#   ZLMediaKit 通过 CMAKE_INSTALL_RPATH 已将 Tongsuo 部署路径写入 MediaServer，
#   故本脚本不复制 Tongsuo 到 VideoRoot，直接沿用 CoreRoot 的部署。
#
# 用法：
#   sudo ./tests/deploy_tmp.sh
#   sudo ./tests/deploy_tmp.sh --mode copy
#   sudo ./tests/deploy_tmp.sh --no-build
#   sudo ./tests/deploy_tmp.sh --force
#   sudo ./tests/deploy_tmp.sh --uninstall
# =============================================================================
set -euo pipefail

if [ -t 1 ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
else
  RED=''; GREEN=''; YELLOW=''; CYAN=''; NC=''
fi
log()  { echo -e "${GREEN}[deploy]${NC} $*"; }
warn() { echo -e "${YELLOW}[deploy][WARN]${NC} $*" >&2; }
err()  { echo -e "${RED}[deploy][ERROR]${NC} $*" >&2; }
info() { echo -e "${CYAN}[deploy]${NC} $*"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_SOURCE="$(cd "$SCRIPT_DIR/.." && pwd)"
SOURCE="$DEFAULT_SOURCE"
TARGET="/opt/plats_tool"
MODE="link"
FORCE=0
UNINSTALL=0
DO_BUILD=1

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [ $# -gt 0 ]; do
  case "$1" in
    --source)    SOURCE="${2:-}"; shift 2 ;;
    --target)    TARGET="${2:-}"; shift 2 ;;
    --mode)      MODE="${2:-}"; shift 2 ;;
    --no-build)  DO_BUILD=0; shift ;;
    --force)     FORCE=1; shift ;;
    --uninstall) UNINSTALL=1; shift ;;
    -h|--help)   usage ;;
    *) err "未知参数：$1"; usage ;;
  esac
done

if [ ! -d "$SOURCE" ]; then
  err "源目录不存在：$SOURCE"; exit 1
fi
if [ ! -d "$SOURCE/src/tool" ]; then
  err "源目录缺少 src/tool：$SOURCE/src/tool"; exit 1
fi
case "$MODE" in
  link|copy) ;;
  *) err "--mode 必须是 link 或 copy，实际：$MODE"; exit 1 ;;
esac

SOURCE="$(cd "$SOURCE" && pwd)"
case "$TARGET" in
  /*) : ;;
  *)  TARGET="$(cd "$(dirname "$TARGET")" 2>/dev/null && pwd)/$(basename "$TARGET")" ;;
esac
TARGET="${TARGET%/}"

TARGET_CORE_ROOT="$TARGET/core"
TARGET_CORE_BIN="$TARGET_CORE_ROOT/bin"
TARGET_CORE_SBIN="$TARGET_CORE_ROOT/sbin"
TARGET_CORE_CONF_DIR="$TARGET_CORE_ROOT/conf"
TARGET_CORE_CONF="$TARGET_CORE_ROOT/conf/template"
TARGET_CORE_LIBS="$TARGET_CORE_ROOT/libs"
TARGET_TONGSUO="$TARGET_CORE_LIBS/tongsuo"
TARGET_OPENSSL="$TARGET_TONGSUO/bin/openssl"
TARGET_NGINX="$TARGET_CORE_SBIN/nginx"
TARGET_MIME="$TARGET_CORE_CONF_DIR/mime.types"
TARGET_OQSPROVIDER="$TARGET_TONGSUO/libs/ossl-modules/oqsprovider.so"
# ZLM 运行时通过 rpath 找 Tongsuo 的路径
TARGET_TONGSUO_RUNTIME_LIB_DIR="$TARGET_TONGSUO/libs"

TARGET_VIDEO_ROOT="$TARGET/video"
TARGET_VIDEO_LIBS="$TARGET_VIDEO_ROOT/libs/ZLMediaKit"
TARGET_VIDEO_BIN="$TARGET_VIDEO_ROOT/run/bin"
TARGET_VIDEO_CONF="$TARGET_VIDEO_ROOT/conf/zlmediakit"

SOURCE_TOOL_DIR="$SOURCE/src/tool"
SOURCE_TOOL_BIN="$SOURCE_TOOL_DIR/bin"
SOURCE_TONGSUO="$SOURCE/libs/bin/tongsuo"
SOURCE_NGINX="$SOURCE/src/tengine-3.1.0/bin/sbin/nginx"
SOURCE_MIME="$SOURCE/src/tengine-3.1.0/bin/conf/mime.types"
SOURCE_TEMPLATES="$SOURCE/src/tengine-3.1.0/conf/template"
SOURCE_OQSPROVIDER="$SOURCE_TONGSUO/libs/ossl-modules/oqsprovider.so"

SOURCE_ZLM_BUILD="$SOURCE/libs/bin/zlmediakit"
SOURCE_ZLM_MEDIASERVER="$SOURCE_ZLM_BUILD/MediaServer"
SOURCE_ZLM_CONFIG="$SOURCE_ZLM_BUILD/config.ini"
SOURCE_ZLM_LIB_DIR="$SOURCE_ZLM_BUILD/lib"

if [ "$UNINSTALL" -eq 1 ]; then
  log "开始卸载：$TARGET"
  removed=0
  if [ -d "$TARGET_CORE_BIN" ]; then
    shopt -s nullglob
    for f in "$TARGET_CORE_BIN"/*; do
      base="$(basename "$f")"
      case "$base" in core|core.exe) continue ;; esac
      log "  移除工具：$f"; rm -f "$f"; removed=$((removed+1))
    done
    shopt -u nullglob
  fi
  [ -f "$TARGET_NGINX" ] && { log "  移除 Tengine：$TARGET_NGINX"; rm -f "$TARGET_NGINX"; removed=$((removed+1)); }
  [ -f "$TARGET_MIME" ]  && { log "  移除 mime.types：$TARGET_MIME"; rm -f "$TARGET_MIME"; removed=$((removed+1)); }
  if [ -d "$TARGET_CORE_CONF" ]; then
    shopt -s nullglob
    for f in "$TARGET_CORE_CONF"/*.tmpl; do
      log "  移除模板：$f"; rm -f "$f"; removed=$((removed+1))
    done
    shopt -u nullglob
  fi
  if [ -L "$TARGET_TONGSUO" ]; then
    log "  移除铜锁链接：$TARGET_TONGSUO"; rm -f "$TARGET_TONGSUO"; removed=$((removed+1))
  elif [ -d "$TARGET_TONGSUO" ]; then
    warn "  铜锁是真实目录，未自动删除：$TARGET_TONGSUO"
    [ -f "$TARGET_OQSPROVIDER" ] && warn "  其中包含 oqsprovider.so，未单独删除：$TARGET_OQSPROVIDER"
  fi
  [ -f "$TARGET_VIDEO_BIN/MediaServer" ] && { log "  移除 ZLMediaKit 二进制"; rm -f "$TARGET_VIDEO_BIN/MediaServer"; removed=$((removed+1)); }
  [ -f "$TARGET_VIDEO_CONF/config.ini" ] && { log "  移除 ZLMediaKit 配置"; rm -f "$TARGET_VIDEO_CONF/config.ini"; removed=$((removed+1)); }
  if [ -d "$TARGET_VIDEO_LIBS" ]; then
    shopt -s nullglob
    for f in "$TARGET_VIDEO_LIBS"/*.so; do
      log "  移除 ZLMediaKit 库：$f"; rm -f "$f"; removed=$((removed+1))
    done
    shopt -u nullglob
  fi
  log "✅ 卸载完成（共移除 $removed 项）"
  exit 0
fi

need_root=0
case "$TARGET" in /opt/*|/usr/*|/var/*|/etc/*|/srv/*) need_root=1 ;; esac
if [ "$need_root" -eq 1 ] && [ "$(id -u)" != "0" ]; then
  err "部署到 $TARGET 需要 root 权限"; err "请使用：sudo $0 $*"; exit 2
fi

if [ "$DO_BUILD" -eq 1 ]; then
  info "编译 C 工具（$SOURCE_TOOL_DIR）..."
  if make -C "$SOURCE_TOOL_DIR" all >/dev/null 2>&1; then
    log "  ✅ C 工具编译完成"
  else
    warn "  make 失败（或未安装编译器），尝试使用已有 bin/"
  fi
else
  info "跳过 C 工具编译（--no-build）"
fi

if [ ! -d "$SOURCE_TOOL_BIN" ]; then
  err "未找到 C 工具产物目录：$SOURCE_TOOL_BIN"; exit 3
fi

TOOLS=()
shopt -s nullglob
for f in "$SOURCE_TOOL_BIN"/*; do
  base="$(basename "$f")"
  case "$base" in .*) continue ;; core|core.exe) continue ;; esac
  [ -f "$f" ] && TOOLS+=("$base")
done
shopt -u nullglob
if [ "${#TOOLS[@]}" -eq 0 ]; then err "C 工具目录为空：$SOURCE_TOOL_BIN"; exit 3; fi

info "准备目标目录..."
mkdir -p "$TARGET_CORE_BIN" "$TARGET_CORE_SBIN" "$TARGET_CORE_CONF_DIR" "$TARGET_CORE_CONF" \
         "$TARGET_CORE_LIBS" "$TARGET_VIDEO_LIBS" "$TARGET_VIDEO_BIN" "$TARGET_VIDEO_CONF"

info "部署 C 工具 → $TARGET_CORE_BIN/"
for name in "${TOOLS[@]}"; do
  src="$SOURCE_TOOL_BIN/$name"; dst="$TARGET_CORE_BIN/$name"
  if [ -e "$dst" ] || [ -L "$dst" ]; then
    if [ "$FORCE" -ne 1 ]; then warn "  已存在，跳过：$dst（加 --force 覆盖）"; continue; fi
    rm -f "$dst"
  fi
  install -m 0755 "$src" "$dst"; log "  install: $dst"
done

if [ -x "$SOURCE_NGINX" ]; then
  info "部署 Tengine → $TARGET_NGINX"
  if [ -e "$TARGET_NGINX" ] && [ "$FORCE" -ne 1 ]; then
    warn "  已存在，跳过：$TARGET_NGINX（加 --force 覆盖）"
  else
    rm -f "$TARGET_NGINX"
    install -m 0755 "$SOURCE_NGINX" "$TARGET_NGINX"; log "  install: $TARGET_NGINX"
  fi
else
  warn "Tengine 二进制不存在：$SOURCE_NGINX"; warn "  请先执行：cd $SOURCE && make tengine"
fi

if [ -f "$SOURCE_MIME" ]; then
  info "部署 mime.types → $TARGET_MIME"
  if [ -e "$TARGET_MIME" ] && [ "$FORCE" -ne 1 ]; then
    warn "  已存在，跳过：$TARGET_MIME（加 --force 覆盖）"
  else
    rm -f "$TARGET_MIME"
    install -m 0644 "$SOURCE_MIME" "$TARGET_MIME"; log "  install: $TARGET_MIME"
  fi
else
  warn "mime.types 不存在：$SOURCE_MIME"
fi

if [ -d "$SOURCE_TEMPLATES" ]; then
  info "部署模板 → $TARGET_CORE_CONF"
  shopt -s nullglob
  tmpls=("$SOURCE_TEMPLATES"/*.tmpl)
  shopt -u nullglob
  if [ "${#tmpls[@]}" -eq 0 ]; then
    warn "  模板目录为空：$SOURCE_TEMPLATES"
  else
    for t in "${tmpls[@]}"; do
      bn="$(basename "$t")"; dst="$TARGET_CORE_CONF/$bn"
      if [ -e "$dst" ] && [ "$FORCE" -ne 1 ]; then
        warn "  已存在，跳过：$dst（加 --force 覆盖）"; continue
      fi
      install -m 0644 "$t" "$dst"; log "  install: $dst"
    done
  fi
else
  warn "模板目录不存在：$SOURCE_TEMPLATES"
fi

if [ -d "$SOURCE_TONGSUO" ]; then
  info "部署铜锁 openssl → $TARGET_TONGSUO"
  if [ "$MODE" = "link" ]; then
    if [ -e "$TARGET_TONGSUO" ] || [ -L "$TARGET_TONGSUO" ]; then
      if [ "$FORCE" -ne 1 ]; then
        warn "  已存在，跳过：$TARGET_TONGSUO（加 --force 覆盖）"
      else
        [ -L "$TARGET_TONGSUO" ] && rm -f "$TARGET_TONGSUO"
        [ -d "$TARGET_TONGSUO" ] && rm -rf "$TARGET_TONGSUO"
      fi
    fi
    if [ ! -e "$TARGET_TONGSUO" ] && [ ! -L "$TARGET_TONGSUO" ]; then
      ln -s "$SOURCE_TONGSUO" "$TARGET_TONGSUO"
      log "  link: $TARGET_TONGSUO → $SOURCE_TONGSUO"
    fi
  else
    if ! command -v rsync >/dev/null 2>&1; then
      err "copy 模式需要 rsync，请先安装：apt install rsync"; exit 2
    fi
    [ "$FORCE" -eq 1 ] && [ -d "$TARGET_TONGSUO" ] && rm -rf "$TARGET_TONGSUO"
    mkdir -p "$TARGET_TONGSUO"
    rsync -a --delete "$SOURCE_TONGSUO/" "$TARGET_TONGSUO/"
    log "  copy: $TARGET_TONGSUO"
  fi
else
  err "铜锁目录不存在：$SOURCE_TONGSUO"; exit 3
fi

if [ -f "$SOURCE_OQSPROVIDER" ]; then
  info "oqs-provider: 源存在 $SOURCE_OQSPROVIDER"
  if [ -f "$TARGET_OQSPROVIDER" ]; then
    log "  oqsprovider.so 已就位：$TARGET_OQSPROVIDER"
  else
    warn "  oqsprovider.so 未就位：$TARGET_OQSPROVIDER"
    warn "  请确认铜锁部署是否完整，或重新执行：make -C $SOURCE/libs oqs-provider"
  fi
else
  warn "oqs-provider: 源不存在 $SOURCE_OQSPROVIDER"
fi

info "部署 ZLMediaKit → $TARGET_VIDEO_ROOT"
# 注意：MediaServer 已通过 CMAKE_INSTALL_RPATH 将 Tongsuo 部署路径写入，
#       此处不再复制 Tongsuo，直接依赖 $TARGET_TONGSUO_RUNTIME_LIB_DIR。
if [ -x "$SOURCE_ZLM_MEDIASERVER" ]; then
  if [ -e "$TARGET_VIDEO_BIN/MediaServer" ] && [ "$FORCE" -ne 1 ]; then
    warn "  已存在，跳过：$TARGET_VIDEO_BIN/MediaServer（加 --force 覆盖）"
  else
    rm -f "$TARGET_VIDEO_BIN/MediaServer"
    install -m 0755 "$SOURCE_ZLM_MEDIASERVER" "$TARGET_VIDEO_BIN/MediaServer"
    log "  install: $TARGET_VIDEO_BIN/MediaServer"
  fi
else
  warn "ZLMediaKit MediaServer 不存在：$SOURCE_ZLM_MEDIASERVER"
  warn "  请先执行：cd $SOURCE && make zlmediakit"
fi

if [ -f "$SOURCE_ZLM_CONFIG" ]; then
  if [ -e "$TARGET_VIDEO_CONF/config.ini" ] && [ "$FORCE" -ne 1 ]; then
    warn "  已存在，跳过：$TARGET_VIDEO_CONF/config.ini（加 --force 覆盖）"
  else
    rm -f "$TARGET_VIDEO_CONF/config.ini"
    install -m 0644 "$SOURCE_ZLM_CONFIG" "$TARGET_VIDEO_CONF/config.ini"
    log "  install: $TARGET_VIDEO_CONF/config.ini"
  fi
else
  warn "ZLMediaKit config.ini 不存在：$SOURCE_ZLM_CONFIG"
fi

if [ -d "$SOURCE_ZLM_LIB_DIR" ]; then
  shopt -s nullglob
  zlm_libs=("$SOURCE_ZLM_LIB_DIR"/*.so)
  shopt -u nullglob
  if [ "${#zlm_libs[@]}" -eq 0 ]; then
    warn "  ZLMediaKit 库目录为空：$SOURCE_ZLM_LIB_DIR"
  else
    find "$SOURCE_ZLM_LIB_DIR" -maxdepth 1 -type f \
      -name '*.so' -not -name '*.debug' \
      -exec install -m 0755 {} "$TARGET_VIDEO_LIBS/" \;
    log "  install libs → $TARGET_VIDEO_LIBS/"
  fi
else
  warn "ZLMediaKit 库目录不存在：$SOURCE_ZLM_LIB_DIR"
fi

for name in "${TOOLS[@]}"; do
  [ -e "$TARGET_CORE_BIN/$name" ] && chmod 0755 "$TARGET_CORE_BIN/$name" 2>/dev/null || true
done
[ -f "$TARGET_OPENSSL" ] && chmod 0755 "$TARGET_OPENSSL" 2>/dev/null || true
[ -f "$TARGET_OQSPROVIDER" ] && chmod 0755 "$TARGET_OQSPROVIDER" 2>/dev/null || true
[ -f "$TARGET_NGINX" ]   && chmod 0755 "$TARGET_NGINX"   2>/dev/null || true
[ -f "$TARGET_MIME" ]    && chmod 0644 "$TARGET_MIME"    2>/dev/null || true
[ -f "$TARGET_VIDEO_BIN/MediaServer" ] && chmod 0755 "$TARGET_VIDEO_BIN/MediaServer" 2>/dev/null || true
[ -f "$TARGET_VIDEO_CONF/config.ini" ] && chmod 0644 "$TARGET_VIDEO_CONF/config.ini" 2>/dev/null || true

info "校验文件..."
FAIL=0
for name in "${TOOLS[@]}"; do
  bin="$TARGET_CORE_BIN/$name"
  if [ -x "$bin" ]; then log "  可执行：$bin"; else err "  不可执行或缺失：$bin"; FAIL=1; fi
done
if [ -x "$TARGET_NGINX" ]; then
  log "  Tengine：$TARGET_NGINX"
  log "  版本：$("$TARGET_NGINX" -v 2>&1 | head -1 || echo unknown)"
else
  err "  Tengine 缺失或不可执行：$TARGET_NGINX"; FAIL=1
fi
[ -f "$TARGET_MIME" ] && log "  mime.types：$TARGET_MIME" || { err "  mime.types 缺失：$TARGET_MIME"; FAIL=1; }
if [ -d "$TARGET_CORE_CONF" ]; then
  shopt -s nullglob
  tmpl_files=("$TARGET_CORE_CONF"/*.tmpl)
  shopt -u nullglob
  tmpl_count=${#tmpl_files[@]}
  [ "$tmpl_count" -ge 2 ] && log "  模板：$tmpl_count 个" || warn "  模板数不足：$tmpl_count"
fi
if [ -x "$TARGET_OPENSSL" ]; then
  log "  铜锁 openssl：$TARGET_OPENSSL"
  log "  版本：$("$TARGET_OPENSSL" version 2>/dev/null | head -1 || echo unknown)"
else
  warn "  铜锁 openssl 未找到：$TARGET_OPENSSL"
fi
if [ -f "$TARGET_OQSPROVIDER" ]; then
  log "  oqsprovider.so：$TARGET_OQSPROVIDER"
fi
if [ -x "$TARGET_VIDEO_BIN/MediaServer" ]; then
  log "  ZLMediaKit MediaServer：$TARGET_VIDEO_BIN/MediaServer"
else
  err "  ZLMediaKit MediaServer 缺失或不可执行：$TARGET_VIDEO_BIN/MediaServer"; FAIL=1
fi
[ -f "$TARGET_VIDEO_CONF/config.ini" ] && log "  ZLMediaKit config.ini：$TARGET_VIDEO_CONF/config.ini" \
  || warn "  ZLMediaKit config.ini 缺失：$TARGET_VIDEO_CONF/config.ini"
if [ -d "$TARGET_VIDEO_LIBS" ]; then
  shopt -s nullglob
  deployed_zlm_libs=("$TARGET_VIDEO_LIBS"/*.so)
  shopt -u nullglob
  log "  ZLMediaKit 库：${#deployed_zlm_libs[@]} 个"
fi

[ "$FAIL" -ne 0 ] && { err "部署校验失败（文件缺失）"; exit 3; }

# -----------------------------------------------------------------------------
# 校验：ZLMediaKit MediaServer 的 rpath 是否指向部署后的 Tongsuo 目录
# -----------------------------------------------------------------------------
info "校验 MediaServer rpath → $TARGET_TONGSUO_RUNTIME_LIB_DIR"
if [ -x "$TARGET_VIDEO_BIN/MediaServer" ]; then
  rpath=""
  if command -v patchelf >/dev/null 2>&1; then
    rpath="$(patchelf --print-rpath "$TARGET_VIDEO_BIN/MediaServer" 2>/dev/null || true)"
  elif command -v readelf >/dev/null 2>&1; then
    rpath="$(readelf -d "$TARGET_VIDEO_BIN/MediaServer" 2>/dev/null \
             | sed -n 's/.*\(RUNPATH\|RPATH\).*\[\(.*\)\]/\2/p' || true)"
  fi
  if [ -z "$rpath" ]; then
    warn "  未能读取 MediaServer 的 rpath（缺少 patchelf/readelf），跳过检查"
  else
    log "  当前 rpath: $rpath"
    if echo "$rpath" | tr ':' '\n' | grep -qx "$TARGET_TONGSUO_RUNTIME_LIB_DIR"; then
      log "  ✅ rpath 已包含 $TARGET_TONGSUO_RUNTIME_LIB_DIR"
    else
      warn "  rpath 不包含 $TARGET_TONGSUO_RUNTIME_LIB_DIR"
      warn "  建议：(1) 重新编译 ZLM，确认 CMAKE_INSTALL_RPATH 生效"
      warn "         (2) 或临时使用: export LD_LIBRARY_PATH=$TARGET_TONGSUO_RUNTIME_LIB_DIR:\$LD_LIBRARY_PATH"
    fi
  fi
fi

info "校验动态库可达性..."
if command -v ldd >/dev/null 2>&1; then
  LIBFAIL=0
  for name in "${TOOLS[@]}"; do
    bin="$TARGET_CORE_BIN/$name"; [ -x "$bin" ] || continue
    missing="$(ldd "$bin" 2>/dev/null | grep 'not found' || true)"
    [ -n "$missing" ] && { err "  $name 动态库缺失："; echo "$missing" | sed 's/^/    /' >&2; LIBFAIL=1; } \
                       || log "  动态库可达：$name"
  done
  if [ -x "$TARGET_NGINX" ]; then
    missing="$(ldd "$TARGET_NGINX" 2>/dev/null | grep 'not found' || true)"
    [ -n "$missing" ] && { err "  nginx 动态库缺失："; echo "$missing" | sed 's/^/    /' >&2; LIBFAIL=1; } \
                       || log "  动态库可达：nginx"
  fi
  if [ -f "$TARGET_OQSPROVIDER" ]; then
    missing="$(ldd "$TARGET_OQSPROVIDER" 2>/dev/null | grep 'not found' || true)"
    [ -n "$missing" ] && { err "  oqsprovider.so 动态库缺失："; echo "$missing" | sed 's/^/    /' >&2; LIBFAIL=1; } \
                       || log "  动态库可达：oqsprovider.so"
  fi
  if [ -x "$TARGET_VIDEO_BIN/MediaServer" ]; then
    missing="$(ldd "$TARGET_VIDEO_BIN/MediaServer" 2>/dev/null | grep 'not found' || true)"
    if [ -n "$missing" ]; then
      warn "  MediaServer 动态库存在未找到项："
      echo "$missing" | sed 's/^/    /' >&2
      warn "  若为 Tongsuo 的 libcrypto/libssl，请确认 rpath 或设置："
      warn "    export LD_LIBRARY_PATH=$TARGET_TONGSUO_RUNTIME_LIB_DIR:\$LD_LIBRARY_PATH"
    else
      log "  动态库可达：MediaServer"
    fi
  fi
  [ "$LIBFAIL" -ne 0 ] && { err "动态库校验失败"; exit 3; }
else
  warn "系统无 ldd，跳过动态库校验"
fi

echo
echo "============================================================"
echo -e "${GREEN}✅ core 库与工具临时部署完成${NC}"
echo "============================================================"
echo "部署模式：     $MODE"
echo "源目录：       $SOURCE"
echo "目标根：       $TARGET"
echo "CoreRoot：     $TARGET_CORE_ROOT"
echo "Tengine：      $TARGET_NGINX"
echo "mime.types：   $TARGET_MIME"
echo "模板目标：     $TARGET_CORE_CONF"
echo "C 工具目录：   $TARGET_CORE_BIN"
echo "铜锁目录：     $TARGET_TONGSUO"
echo "oqs-provider： $TARGET_OQSPROVIDER"
echo "VideoRoot：    $TARGET_VIDEO_ROOT"
echo "ZLM 二进制：   $TARGET_VIDEO_BIN/MediaServer"
echo "ZLM 库目录：   $TARGET_VIDEO_LIBS"
echo "ZLM 配置：     $TARGET_VIDEO_CONF/config.ini"
echo "ZLM 运行时找 Tongsuo：$TARGET_TONGSUO_RUNTIME_LIB_DIR（通过 rpath，不复制）"
echo
echo "验证命令："
echo "  $TARGET_NGINX -V 2>&1 | head -2"
echo "  ldd $TARGET_VIDEO_BIN/MediaServer | grep -E 'ssl|crypto|srtp|not found'"
echo "  patchelf --print-rpath $TARGET_VIDEO_BIN/MediaServer"
echo "  $TARGET_VIDEO_BIN/MediaServer -h 2>&1 | head -5"
echo
echo "卸载："
echo "  sudo $0 --uninstall"
echo "============================================================"

exit 0
