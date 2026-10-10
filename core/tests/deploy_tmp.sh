#!/usr/bin/env bash
# =============================================================================
# core 临时部署脚本（Go 化改造后）
#
# 用途：
#   把开发目录的「铜锁 openssl + C 工具」部署到 CoreRoot 下对应位置：
#     <root>/core/bin/keycrypt
#     <root>/core/bin/whitebox_sm4
#     <root>/core/libs/tongsuo/bin/openssl
#
# ★ 部署结构说明：
#   CoreRoot = <root>/core  （与 core.yaml 中 core.* 相对路径基准一致）
#   - C 工具 → <CoreRoot>/bin/           （与 core 二进制同目录）
#   - 铜锁   → <CoreRoot>/libs/tongsuo/  （与 core.yaml 中 "libs/tongsuo" 一致）
#
#   C 工具 RPATH = $ORIGIN/../libs/tongsuo/libs
#     $ORIGIN = <CoreRoot>/bin
#     库      = <CoreRoot>/libs/tongsuo/libs/
#
# ★ 关键规则：
#   - C 工具必须真拷贝（不能软链），否则 $ORIGIN 会解析到开发目录
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
  sed -n '2,36p' "$0" | sed 's/^# \{0,1\}//'
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
# ★ 目标路径（与 core.yaml 相对路径基准严格对齐）
# -----------------------------------------------------------------------------
# CoreRoot = <root>/core
TARGET_CORE_ROOT="$TARGET/core"
TARGET_CORE_BIN="$TARGET_CORE_ROOT/bin"
TARGET_CORE_LIBS="$TARGET_CORE_ROOT/libs"
TARGET_TONGSUO="$TARGET_CORE_LIBS/tongsuo"
TARGET_OPENSSL="$TARGET_TONGSUO/bin/openssl"

# 源路径
SOURCE_TOOL_DIR="$SOURCE/src/tool"
SOURCE_TOOL_BIN="$SOURCE_TOOL_DIR/bin"
SOURCE_TONGSUO="$SOURCE/libs/bin/tongsuo"

# -----------------------------------------------------------------------------
# 卸载
# -----------------------------------------------------------------------------
if [ "$UNINSTALL" -eq 1 ]; then
  log "开始卸载：$TARGET"
  removed=0

  if [ -d "$TARGET_CORE_BIN" ]; then
    shopt -s nullglob
    for f in "$TARGET_CORE_BIN"/*; do
      base="$(basename "$f")"
      case "$base" in
        core|core.exe) continue ;;  # 不动 core 二进制
      esac
      log "  移除工具：$f"
      rm -f "$f"
      removed=$((removed+1))
    done
    shopt -u nullglob
  fi

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

# 收集工具列表（排除 core、core.exe）
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
mkdir -p "$TARGET_CORE_LIBS"

# -----------------------------------------------------------------------------
# 部署 C 工具（★ 始终真拷贝，不用软链）
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

# -----------------------------------------------------------------------------
# 校验 1：文件存在
# -----------------------------------------------------------------------------
info "校验文件..."
FAIL=0
for name in "${TOOLS[@]}"; do
  bin="$TARGET_CORE_BIN/$name"
  if [ -x "$bin" ]; then
    log "  可执行：$bin"
  else
    err "  不可执行或缺失：$bin"
    FAIL=1
  fi
done

if [ -x "$TARGET_OPENSSL" ]; then
  log "  铜锁 openssl：$TARGET_OPENSSL"
  log "  版本：$("$TARGET_OPENSSL" version 2>/dev/null | head -1 || echo unknown)"
else
  warn "  铜锁 openssl 未找到：$TARGET_OPENSSL"
fi

[ "$FAIL" -ne 0 ] && { err "部署校验失败（文件缺失）"; exit 3; }

# -----------------------------------------------------------------------------
# 校验 2：动态库可达性
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

  if [ "$LIBFAIL" -ne 0 ]; then
    echo
    err "动态库校验失败"
    for name in "${TOOLS[@]}"; do
      bin="$TARGET_CORE_BIN/$name"
      [ -x "$bin" ] || continue
      err "  readelf -d $bin | grep -iE 'rpath|runpath'"
      err "  ldd $bin"
    done
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
echo "部署模式：     $MODE（C 工具始终真拷贝）"
echo "源目录：       $SOURCE"
echo "目标根：       $TARGET"
echo "CoreRoot：     $TARGET_CORE_ROOT"
echo "C 工具目录：   $TARGET_CORE_BIN"
echo "铜锁目录：     $TARGET_TONGSUO"
echo
echo "已部署的 C 工具："
for name in "${TOOLS[@]}"; do
  echo "  - $name"
done
echo
echo "验证命令："
echo "  readelf -d $TARGET_CORE_BIN/whitebox_sm4 | grep -i runpath"
echo "  ldd $TARGET_CORE_BIN/whitebox_sm4 | grep -E 'libcrypto|not found'"
echo "  $TARGET_CORE_BIN/whitebox_sm4 2>&1 | head -3"
echo
echo "卸载："
echo "  sudo $0 --uninstall"
echo "============================================================"

exit 0
