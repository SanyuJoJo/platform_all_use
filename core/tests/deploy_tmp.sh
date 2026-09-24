#!/usr/bin/env bash
# =============================================================================
# core 临时部署脚本
#
# 用途：
#   将开发目录（默认 /code/platform_all_use/core）临时部署到默认安装目录
#   （默认 /opt/core），使平台后端无需修改 .env 即可直接调用：
#       /opt/core/sbin/dispatch.sh --op ... --in ... --out ...
#
# 模式：
#   - link（默认）：/opt/core → 源目录 的软链接，零拷贝、实时同步；
#   - copy：rsync 复制到 /opt/core，独立副本，改动需重新执行本脚本。
#
# 用法：
#   sudo ./scripts/deploy-tmp.sh                 # 默认 link 模式
#   sudo ./scripts/deploy-tmp.sh --mode copy     # 使用 rsync 复制
#   sudo ./scripts/deploy-tmp.sh --force         # 目标已存在时强制覆盖
#   sudo ./scripts/deploy-tmp.sh --target /srv/core  # 自定义目标
#   sudo ./scripts/deploy-tmp.sh --uninstall     # 卸载（删除 /opt/core）
#
# 参数：
#   --source <path>   源目录，默认脚本所在目录的上级（core 根）
#   --target <path>   目标目录，默认 /opt/core
#   --mode <link|copy>  部署模式，默认 link
#   --force           目标已存在时强制覆盖
#   --uninstall       卸载（删除目标，仅当目标是软链接或本脚本创建时）
#   -h, --help        显示帮助
#
# 退出码：
#   0 成功；1 参数错误；2 环境问题；3 部署失败
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
DEFAULT_SOURCE="$(cd "$SCRIPT_DIR/.." && pwd)"   # scripts/ 的上级即 core 根
SOURCE="$DEFAULT_SOURCE"
TARGET="/opt/core"
MODE="link"
FORCE=0
UNINSTALL=0

# -----------------------------------------------------------------------------
# 参数解析
# -----------------------------------------------------------------------------
usage() {
  sed -n '2,40p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [ $# -gt 0 ]; do
  case "$1" in
    --source)    SOURCE="${2:-}"; shift 2 ;;
    --target)    TARGET="${2:-}"; shift 2 ;;
    --mode)      MODE="${2:-}"; shift 2 ;;
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
  err "源目录不存在：$SOURCE"
  exit 1
fi
if [ ! -f "$SOURCE/sbin/dispatch.sh" ]; then
  err "源目录不是有效的 core 根（缺少 sbin/dispatch.sh）：$SOURCE"
  err "提示：--source 应指向包含 sbin/、libs/、conf/ 的 core 根目录"
  exit 1
fi

case "$MODE" in
  link|copy) ;;
  *) err "--mode 必须是 link 或 copy，实际：$MODE"; exit 1 ;;
esac

# 绝对化
SOURCE="$(cd "$SOURCE" && pwd)"

# -----------------------------------------------------------------------------
# 卸载分支
# -----------------------------------------------------------------------------
if [ "$UNINSTALL" -eq 1 ]; then
  if [ ! -e "$TARGET" ]; then
    log "目标不存在，无需卸载：$TARGET"
    exit 0
  fi
  if [ -L "$TARGET" ]; then
    log "删除软链接：$TARGET"
    rm -f "$TARGET"
    log "✅ 卸载完成"
    exit 0
  fi
  if [ -d "$TARGET" ]; then
    warn "目标是真实目录：$TARGET"
    warn "为避免误删，本脚本不会自动删除真实目录。"
    warn "如确认要删除，请手动执行：sudo rm -rf $TARGET"
    exit 0
  fi
  err "目标既不是软链接也不是目录：$TARGET"
  exit 2
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
# 处理已存在的目标
# -----------------------------------------------------------------------------
if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then
  if [ "$FORCE" -ne 1 ]; then
    err "目标已存在：$TARGET"
    err "如需覆盖，请加 --force（谨慎：link 模式会替换原目录/链接；copy 模式会清空目标后重写）"
    exit 3
  fi
  if [ -L "$TARGET" ]; then
    log "移除已有软链接：$TARGET"
    rm -f "$TARGET"
  elif [ -d "$TARGET" ]; then
    warn "目标是真实目录，--force 将删除整个目录：$TARGET"
    rm -rf "$TARGET"
  else
    err "目标既不是软链接也不是目录，拒绝删除：$TARGET"
    exit 3
  fi
fi

# -----------------------------------------------------------------------------
# 部署
# -----------------------------------------------------------------------------
TARGET_PARENT="$(dirname "$TARGET")"
mkdir -p "$TARGET_PARENT"

info "部署模式：$MODE"
info "源目录：  $SOURCE"
info "目标目录：$TARGET"

if [ "$MODE" = "link" ]; then
  ln -s "$SOURCE" "$TARGET"
  log "✅ 已创建软链接：$TARGET → $SOURCE"
else
  if ! command -v rsync >/dev/null 2>&1; then
    err "copy 模式需要 rsync，请先安装：apt install rsync / yum install rsync"
    exit 2
  fi
  mkdir -p "$TARGET"
  rsync -a --delete \
    --exclude='logs/*' \
    --exclude='tmp/*' \
    --exclude='data/keys/*.key.enc' \
    "$SOURCE/" "$TARGET/"
  log "✅ 已复制目录：$SOURCE → $TARGET"
fi

# -----------------------------------------------------------------------------
# 修正关键权限（C-08 要求）
# -----------------------------------------------------------------------------
info "修正目录与文件权限..."

# 目录
for d in \
  "$TARGET" \
  "$TARGET/libs" "$TARGET/libs/src" "$TARGET/libs/bin" \
  "$TARGET/libs/bin/tongsuo" \
  "$TARGET/sbin" "$TARGET/sbin/lib" \
  "$TARGET/src" "$TARGET/src/tool" "$TARGET/src/tool/src" "$TARGET/src/tool/bin"
do
  [ -d "$d" ] && chmod 0755 "$d" 2>/dev/null || true
done

for d in "$TARGET/conf" "$TARGET/logs"; do
  [ -d "$d" ] && chmod 0750 "$d" 2>/dev/null || true
done

for d in "$TARGET/data/keys" "$TARGET/tmp"; do
  [ -d "$d" ] && chmod 0700 "$d" 2>/dev/null || true
done

# 关键文件
[ -f "$TARGET/conf/core.conf" ] && chmod 0640 "$TARGET/conf/core.conf" 2>/dev/null || true
[ -f "$TARGET/conf/algorithm-whitelist.yaml" ] && chmod 0640 "$TARGET/conf/algorithm-whitelist.yaml" 2>/dev/null || true

# 脚本可执行
if [ -d "$TARGET/sbin" ]; then
  find "$TARGET/sbin" -type f -name "*.sh" -exec chmod +x {} \; 2>/dev/null || true
fi
if [ -d "$TARGET/scripts" ]; then
  find "$TARGET/scripts" -type f -name "*.sh" -exec chmod +x {} \; 2>/dev/null || true
fi

log "✅ 权限已修正"

# -----------------------------------------------------------------------------
# 校验
# -----------------------------------------------------------------------------
info "校验关键文件..."

FAIL=0
need_exec() {
  if [ -x "$1" ]; then
    log "  可执行：$1"
  else
    err "  不可执行或缺失：$1"
    FAIL=1
  fi
}
need_file() {
  if [ -f "$1" ]; then
    log "  存在：$1"
  else
    err "  缺失：$1"
    FAIL=1
  fi
}

need_exec "$TARGET/sbin/dispatch.sh"
need_exec "$TARGET/sbin/lib/version_check.sh"
need_file "$TARGET/conf/core.conf"

# 铜锁 openssl 可选
if [ -x "$TARGET/libs/bin/tongsuo/bin/openssl" ]; then
  log "  铜锁 openssl：$TARGET/libs/bin/tongsuo/bin/openssl"
  VER="$("$TARGET/libs/bin/tongsuo/bin/openssl" version 2>/dev/null || true)"
  log "  铜锁版本：$VER"
else
  warn "  铜锁 openssl 未编译：$TARGET/libs/bin/tongsuo/bin/openssl"
  warn "  密码操作将不可用，需要先在 core 中执行 make libs"
fi

if [ "$FAIL" -ne 0 ]; then
  err "部署校验失败"
  exit 3
fi

# -----------------------------------------------------------------------------
# 完成提示
# -----------------------------------------------------------------------------
echo
echo "============================================================"
echo -e "${GREEN}✅ core 临时部署完成${NC}"
echo "============================================================"
echo "部署模式：     $MODE"
echo "源目录：       $SOURCE"
echo "目标目录：     $TARGET"
if [ "$MODE" = "link" ]; then
  echo "类型：         软链接（改动源目录会实时生效）"
else
  echo "类型：         独立副本（改动源目录需重新执行本脚本）"
fi
echo
echo "平台后端无需修改 .env，默认路径即为："
echo "  CORE_DISPATCH_PATH=/opt/core/sbin/dispatch.sh"
echo
echo "验证命令："
echo "  ls -la $TARGET/sbin/dispatch.sh"
echo "  $TARGET/sbin/lib/version_check.sh --json"
echo
echo "卸载："
echo "  sudo $0 --uninstall"
echo "============================================================"

exit 0
