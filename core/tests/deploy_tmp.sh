#!/usr/bin/env bash
# =============================================================================
# core 临时部署脚本（Go 化改造后 + SSL 代理扩展）
#
# 用途：
#   把开发目录的「铜锁 openssl + Tengine + 模板 + mime.types + C 工具」
#   部署到 CoreRoot：
#     <root>/core/sbin/nginx                       ← Tengine 二进制
#     <root>/core/conf/mime.types                  ← nginx mime 映射（模板引用绝对路径）
#     <root>/core/conf/template/*.tmpl             ← SSL 代理模板
#     <root>/core/bin/keycrypt                     ← C 工具
#     <root>/core/bin/whitebox_sm4
#     <root>/core/libs/tongsuo/bin/openssl         ← 铜锁
#
# 源路径（固定）：
#   Tengine     : <core>/src/tengine/bin/sbin/nginx
#   mime.types  : <core>/src/tengine/bin/conf/mime.types
#   模板        : <core>/src/tengine/conf/template/*.tmpl
#   C 工具      : <core>/src/tool/bin/
#   铜锁        : <core>/libs/bin/tongsuo/
#
# 用法：
#   sudo ./tests/deploy_tmp.sh
#   sudo ./tests/deploy_tmp.sh --mode copy
#   sudo ./tests/deploy_tmp.sh --no-build
#   sudo ./tests/deploy_tmp.sh --force
#   sudo ./tests/deploy_tmp.sh --uninstall
# =============================================================================
set -euo pipefail

# -----------------------------------------------------------------------------
# 颜色
# -----------------------------------------------------------------------------
if [ -t 1 ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
else
  RED=''; GREEN=''; YELLOW=''; CYAN=''; NC=''
fi
log()  { echo -e "${GREEN}[deploy]${NC} $*"; }
warn() { echo -e "${YELLOW}[deploy][WARN]${NC} $*" >&2; }
err()  { echo -e "${RED}[deploy][ERROR]${NC} $*" >&2; }
info() { echo -e "${CYAN}[deploy]${NC} $*"; }

# -----------------------------------------------------------------------------
# 默认参数
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_SOURCE="$(cd "$SCRIPT_DIR/.." && pwd)"
SOURCE="$DEFAULT_SOURCE"
TARGET="/opt/plats_tool"
MODE="link"
FORCE=0
UNINSTALL=0
DO_BUILD=1

usage() {
  sed -n '2,26p' "$0" | sed 's/^# \{0,1\}//'
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

# -----------------------------------------------------------------------------
# 参数校验
# -----------------------------------------------------------------------------
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

# -----------------------------------------------------------------------------
# 目标路径
# -----------------------------------------------------------------------------
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

# 源路径
SOURCE_TOOL_DIR="$SOURCE/src/tool"
SOURCE_TOOL_BIN="$SOURCE_TOOL_DIR/bin"
SOURCE_TONGSUO="$SOURCE/libs/bin/tongsuo"
SOURCE_NGINX="$SOURCE/src/tengine-3.1.0/bin/sbin/nginx"
SOURCE_MIME="$SOURCE/src/tengine-3.1.0/bin/conf/mime.types"
SOURCE_TEMPLATES="$SOURCE/src/tengine-3.1.0/conf/template"

# -----------------------------------------------------------------------------
# 卸载
# -----------------------------------------------------------------------------
if [ "$UNINSTALL" -eq 1 ]; then
  log "开始卸载：$TARGET"
  removed=0

  # C 工具
  if [ -d "$TARGET_CORE_BIN" ]; then
    shopt -s nullglob
    for f in "$TARGET_CORE_BIN"/*; do
      base="$(basename "$f")"
      case "$base" in
        core|core.exe) continue ;;
      esac
      log "  移除工具：$f"
      rm -f "$f"
      removed=$((removed+1))
    done
    shopt -u nullglob
  fi

  # Tengine
  if [ -f "$TARGET_NGINX" ]; then
    log "  移除 Tengine：$TARGET_NGINX"
    rm -f "$TARGET_NGINX"
    removed=$((removed+1))
  fi

  # mime.types
  if [ -f "$TARGET_MIME" ]; then
    log "  移除 mime.types：$TARGET_MIME"
    rm -f "$TARGET_MIME"
    removed=$((removed+1))
  fi

  # 模板
  if [ -d "$TARGET_CORE_CONF" ]; then
    shopt -s nullglob
    for f in "$TARGET_CORE_CONF"/*.tmpl; do
      log "  移除模板：$f"
      rm -f "$f"
      removed=$((removed+1))
    done
    shopt -u nullglob
  fi

  # 铜锁
  if [ -L "$TARGET_TONGSUO" ]; then
    log "  移除铜锁链接：$TARGET_TONGSUO"
    rm -f "$TARGET_TONGSUO"
    removed=$((removed+1))
  elif [ -d "$TARGET_TONGSUO" ]; then
    warn "  铜锁是真实目录，未自动删除：$TARGET_TONGSUO"
  fi

  log "✅ 卸载完成（共移除 $removed 项）"
  exit 0
fi

# -----------------------------------------------------------------------------
# 权限检查
# -----------------------------------------------------------------------------
need_root=0
case "$TARGET" in
  /opt/*|/usr/*|/var/*|/etc/*|/srv/*) need_root=1 ;;
esac
if [ "$need_root" -eq 1 ] && [ "$(id -u)" != "0" ]; then
  err "部署到 $TARGET 需要 root 权限"
  err "请使用：sudo $0 $*"
  exit 2
fi

# -----------------------------------------------------------------------------
# 编译 C 工具
# -----------------------------------------------------------------------------
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
  err "未找到 C 工具产物目录：$SOURCE_TOOL_BIN"
  err "请先执行：make -C $SOURCE_TOOL_DIR"
  exit 3
fi

TOOLS=()
shopt -s nullglob
for f in "$SOURCE_TOOL_BIN"/*; do
  base="$(basename "$f")"
  case "$base" in
    .*) continue ;;
    core|core.exe) continue ;;
  esac
  [ -f "$f" ] && TOOLS+=("$base")
done
shopt -u nullglob

if [ "${#TOOLS[@]}" -eq 0 ]; then
  err "C 工具目录为空：$SOURCE_TOOL_BIN"
  exit 3
fi

# -----------------------------------------------------------------------------
# 准备目标目录
# -----------------------------------------------------------------------------
info "准备目标目录..."
mkdir -p "$TARGET_CORE_BIN"
mkdir -p "$TARGET_CORE_SBIN"
mkdir -p "$TARGET_CORE_CONF_DIR"
mkdir -p "$TARGET_CORE_CONF"
mkdir -p "$TARGET_CORE_LIBS"

# -----------------------------------------------------------------------------
# 部署 C 工具
# -----------------------------------------------------------------------------
info "部署 C 工具 → $TARGET_CORE_BIN/"
for name in "${TOOLS[@]}"; do
  src="$SOURCE_TOOL_BIN/$name"
  dst="$TARGET_CORE_BIN/$name"

  if [ -e "$dst" ] || [ -L "$dst" ]; then
    if [ "$FORCE" -ne 1 ]; then
      warn "  已存在，跳过：$dst（加 --force 覆盖）"
      continue
    fi
    rm -f "$dst"
  fi

  install -m 0755 "$src" "$dst"
  log "  install: $dst"
done

# -----------------------------------------------------------------------------
# 部署 Tengine
# -----------------------------------------------------------------------------
if [ -x "$SOURCE_NGINX" ]; then
  info "部署 Tengine → $TARGET_NGINX"
  if [ -e "$TARGET_NGINX" ]; then
    if [ "$FORCE" -ne 1 ]; then
      warn "  已存在，跳过：$TARGET_NGINX（加 --force 覆盖）"
    else
      rm -f "$TARGET_NGINX"
    fi
  fi
  if [ ! -e "$TARGET_NGINX" ]; then
    install -m 0755 "$SOURCE_NGINX" "$TARGET_NGINX"
    log "  install: $TARGET_NGINX"
  fi
else
  warn "Tengine 二进制不存在：$SOURCE_NGINX"
  warn "  请先执行：cd $SOURCE && make tengine"
fi

# -----------------------------------------------------------------------------
# 部署 mime.types
# -----------------------------------------------------------------------------
if [ -f "$SOURCE_MIME" ]; then
  info "部署 mime.types → $TARGET_MIME"
  if [ -e "$TARGET_MIME" ]; then
    if [ "$FORCE" -ne 1 ]; then
      warn "  已存在，跳过：$TARGET_MIME（加 --force 覆盖）"
    else
      rm -f "$TARGET_MIME"
    fi
  fi
  if [ ! -e "$TARGET_MIME" ]; then
    install -m 0644 "$SOURCE_MIME" "$TARGET_MIME"
    log "  install: $TARGET_MIME"
  fi
else
  warn "mime.types 不存在：$SOURCE_MIME"
  warn "  请先执行：cd $SOURCE && make tengine"
fi

# -----------------------------------------------------------------------------
# 部署模板
# -----------------------------------------------------------------------------
if [ -d "$SOURCE_TEMPLATES" ]; then
  info "部署模板 → $TARGET_CORE_CONF"
  info "  模板源：$SOURCE_TEMPLATES"

  shopt -s nullglob
  tmpls=("$SOURCE_TEMPLATES"/*.tmpl)
  shopt -u nullglob

  if [ "${#tmpls[@]}" -eq 0 ]; then
    warn "  模板目录为空：$SOURCE_TEMPLATES"
  else
    for t in "${tmpls[@]}"; do
      bn="$(basename "$t")"
      dst="$TARGET_CORE_CONF/$bn"
      if [ -e "$dst" ] && [ "$FORCE" -ne 1 ]; then
        warn "  已存在，跳过：$dst（加 --force 覆盖）"
        continue
      fi
      install -m 0644 "$t" "$dst"
      log "  install: $dst"
    done
  fi
else
  warn "模板目录不存在：$SOURCE_TEMPLATES"
fi

# -----------------------------------------------------------------------------
# 部署铜锁
# -----------------------------------------------------------------------------
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
      err "copy 模式需要 rsync，请先安装：apt install rsync"
      exit 2
    fi
    [ "$FORCE" -eq 1 ] && [ -d "$TARGET_TONGSUO" ] && rm -rf "$TARGET_TONGSUO"
    mkdir -p "$TARGET_TONGSUO"
    rsync -a --delete "$SOURCE_TONGSUO/" "$TARGET_TONGSUO/"
    log "  copy: $TARGET_TONGSUO"
  fi
else
  err "铜锁目录不存在：$SOURCE_TONGSUO"
  exit 3
fi

# -----------------------------------------------------------------------------
# 权限修正
# -----------------------------------------------------------------------------
for name in "${TOOLS[@]}"; do
  [ -e "$TARGET_CORE_BIN/$name" ] && chmod 0755 "$TARGET_CORE_BIN/$name" 2>/dev/null || true
done
[ -f "$TARGET_OPENSSL" ] && chmod 0755 "$TARGET_OPENSSL" 2>/dev/null || true
[ -f "$TARGET_NGINX" ]   && chmod 0755 "$TARGET_NGINX"   2>/dev/null || true
[ -f "$TARGET_MIME" ]    && chmod 0644 "$TARGET_MIME"    2>/dev/null || true

# -----------------------------------------------------------------------------
# 校验 1：文件存在
# -----------------------------------------------------------------------------
info "校验文件..."
FAIL=0

# C 工具
for name in "${TOOLS[@]}"; do
  bin="$TARGET_CORE_BIN/$name"
  if [ -x "$bin" ]; then
    log "  可执行：$bin"
  else
    err "  不可执行或缺失：$bin"
    FAIL=1
  fi
done

# Tengine
if [ -x "$TARGET_NGINX" ]; then
  log "  Tengine：$TARGET_NGINX"
  log "  版本：$("$TARGET_NGINX" -v 2>&1 | head -1 || echo unknown)"
else
  err "  Tengine 缺失或不可执行：$TARGET_NGINX"
  FAIL=1
fi

# mime.types
if [ -f "$TARGET_MIME" ]; then
  log "  mime.types：$TARGET_MIME"
else
  err "  mime.types 缺失：$TARGET_MIME"
  FAIL=1
fi

# 模板
if [ -d "$TARGET_CORE_CONF" ]; then
  shopt -s nullglob
  tmpl_files=("$TARGET_CORE_CONF"/*.tmpl)
  shopt -u nullglob
  tmpl_count=${#tmpl_files[@]}
  if [ "$tmpl_count" -ge 2 ]; then
    log "  模板：$tmpl_count 个（$TARGET_CORE_CONF）"
    for f in "${tmpl_files[@]}"; do
      log "    - $(basename "$f")"
    done
  else
    warn "  模板数不足（期望 2 个）：$tmpl_count"
  fi
fi

# 铜锁
if [ -x "$TARGET_OPENSSL" ]; then
  log "  铜锁 openssl：$TARGET_OPENSSL"
  log "  版本：$("$TARGET_OPENSSL" version 2>/dev/null | head -1 || echo unknown)"
else
  warn "  铜锁 openssl 未找到：$TARGET_OPENSSL"
fi

[ "$FAIL" -ne 0 ] && { err "部署校验失败（文件缺失）"; exit 3; }

# -----------------------------------------------------------------------------
# 校验 2：模板 include mime.types 引用检查
# -----------------------------------------------------------------------------
info "校验模板 include 配置..."
if [ -d "$TARGET_CORE_CONF" ]; then
  shopt -s nullglob
  for t in "$TARGET_CORE_CONF"/*.tmpl; do
    # 检查 include 行是否是绝对路径（避免 -p 相对路径陷阱）
    if grep -qE '^\s*include\s+mime\.types\s*;' "$t"; then
      warn "  $(basename "$t") 使用相对路径 'include mime.types;'"
      warn "    建议改为：include {{mime_types_path}};"
    fi
  done
  shopt -u nullglob
fi

# -----------------------------------------------------------------------------
# 校验 3：动态库可达性
# -----------------------------------------------------------------------------
info "校验动态库可达性..."
if command -v ldd >/dev/null 2>&1; then
  LIBFAIL=0

  for name in "${TOOLS[@]}"; do
    bin="$TARGET_CORE_BIN/$name"
    [ -x "$bin" ] || continue
    missing="$(ldd "$bin" 2>/dev/null | grep 'not found' || true)"
    if [ -n "$missing" ]; then
      err "  $name 动态库缺失："
      echo "$missing" | sed 's/^/    /' >&2
      LIBFAIL=1
    else
      log "  动态库可达：$name"
    fi
  done

  if [ -x "$TARGET_NGINX" ]; then
    missing="$(ldd "$TARGET_NGINX" 2>/dev/null | grep 'not found' || true)"
    if [ -n "$missing" ]; then
      err "  nginx 动态库缺失："
      echo "$missing" | sed 's/^/    /' >&2
      LIBFAIL=1
    else
      log "  动态库可达：nginx"
    fi
  fi

  if [ "$LIBFAIL" -ne 0 ]; then
    echo
    err "动态库校验失败"
    err "  排查："
    err "    readelf -d $TARGET_NGINX | grep -iE 'rpath|runpath'"
    err "    ldd $TARGET_NGINX | grep 'not found'"
    err "  常见修复："
    err "    1) 设置 LD_LIBRARY_PATH=$TARGET_TONGSUO/libs"
    err "    2) 或 patchelf --set-rpath '$TARGET_TONGSUO/libs' $TARGET_NGINX"
    exit 3
  fi
else
  warn "系统无 ldd，跳过动态库校验"
fi

# -----------------------------------------------------------------------------
# 完成
# -----------------------------------------------------------------------------
echo
echo "============================================================"
echo -e "${GREEN}✅ core 库与工具临时部署完成${NC}"
echo "============================================================"
echo "部署模式：     $MODE（C 工具与 Tengine 始终真拷贝）"
echo "源目录：       $SOURCE"
echo "目标根：       $TARGET"
echo "CoreRoot：     $TARGET_CORE_ROOT"
echo "Tengine：      $TARGET_NGINX"
echo "mime.types：   $TARGET_MIME"
echo "模板源：       $SOURCE_TEMPLATES"
echo "模板目标：     $TARGET_CORE_CONF"
echo "C 工具目录：   $TARGET_CORE_BIN"
echo "铜锁目录：     $TARGET_TONGSUO"
echo
echo "已部署的 C 工具："
for name in "${TOOLS[@]}"; do
  echo "  - $name"
done
echo
echo "验证命令："
echo "  $TARGET_NGINX -V 2>&1 | head -2"
echo "  ldd $TARGET_NGINX | grep -iE 'ssl|crypto|not found'"
echo "  ls -la $TARGET_CORE_CONF_DIR/"
echo "  ls -la $TARGET_CORE_CONF/"
echo
echo "卸载："
echo "  sudo $0 --uninstall"
echo "============================================================"

exit 0
