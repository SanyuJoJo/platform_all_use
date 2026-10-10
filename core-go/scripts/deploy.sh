#!/usr/bin/env bash
# =============================================================================
# core-go 部署脚本
#
# 用途：将开发目录的 core-go 安装到 /opt/plats_tool/core（生产路径）
#
# 部署内容：
#   - bin/core               ← 从 ./bin/core 拷贝
#   - libs/tongsuo/          ← 从探测到的铜锁目录拷贝
#                              （统一布局：bin/openssl + *.so + include/）
#   - conf/core.yaml         ← 从 ./conf/core.yaml 拷贝
#                              ★ 部署后自动改写 tongsuo_openssl_bin
#                                 为 $TARGET/libs/tongsuo/bin/openssl
#   - conf/algorithm-whitelist.yaml
#   - data/keys/             ← 空目录（首次）/ 保留（升级）
#   - logs/、tmp/            ← 空目录
#
# 铜锁源探测顺序：
#   1. ./libs/openssl/<os>-<arch>/openssl      （平台规范，CI 产物常见）
#   2. ../core/libs/bin/tongsuo/bin/openssl    （core 源码，开发常用）
#
# ★ 关键约定：
#   core 二进制不读 CORE_ROOT 推导 openssl，
#   只认 conf/core.yaml 的 tongsuo_openssl_bin。
#   因此部署时必须把该字段写为「部署后绝对路径」。
#
# 用法：
#   sudo ./scripts/deploy.sh                    # 完整部署
#   sudo ./scripts/deploy.sh --upgrade          # 升级（保留 data/keys/、conf/）
#   sudo ./scripts/deploy.sh --target /srv/core # 自定义目标
#   sudo ./scripts/deploy.sh --check            # 只检查环境，不部署
# =============================================================================
set -euo pipefail

# -----------------------------------------------------------------------------
# 基础路径
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TARGET="/opt/plats_tool/core"
MODE="install"   # install | upgrade | check

# 铜锁源目录（由 detect_openssl 填充）
OPENSSL_SRC_DIR=""

# -----------------------------------------------------------------------------
# 颜色
# -----------------------------------------------------------------------------
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; CYAN=''; NC=''
fi
log()  { printf "${GREEN}[deploy]${NC} %s\n" "$*"; }
warn() { printf "${YELLOW}[deploy][WARN]${NC} %s\n" "$*" >&2; }
err()  { printf "${RED}[deploy][ERROR]${NC} %s\n" "$*" >&2; }
info() { printf "${CYAN}[deploy]${NC} %s\n" "$*"; }

# -----------------------------------------------------------------------------
# 参数解析
# -----------------------------------------------------------------------------
while [ $# -gt 0 ]; do
    case "$1" in
        --upgrade) MODE="upgrade"; shift ;;
        --check)   MODE="check"; shift ;;
        --target)  TARGET="$2"; shift 2 ;;
        -h|--help)
            cat <<'EOF'
用法：sudo ./scripts/deploy.sh [选项]
  --upgrade           升级模式（保留现有 conf/ 与 data/keys/）
  --check             只检查环境，不部署
  --target PATH       自定义安装目录，默认 /opt/plats_tool/core
  -h, --help          显示帮助
EOF
            exit 0 ;;
        *) err "未知参数：$1"; exit 1 ;;
    esac
done

# -----------------------------------------------------------------------------
# 平台探测
# -----------------------------------------------------------------------------
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS_RAW" in
    linux*)   PLAT_OS="linux" ;;
    darwin*)  PLAT_OS="darwin" ;;
    *)        err "不支持的平台：$OS_RAW"; exit 1 ;;
esac

ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64)  PLAT_ARCH="amd64" ;;
    aarch64|arm64) PLAT_ARCH="arm64" ;;
    *)             err "不支持的架构：$ARCH_RAW"; exit 1 ;;
esac

PLATFORM="$PLAT_OS-$PLAT_ARCH"
info "平台：$PLATFORM"

# -----------------------------------------------------------------------------
# 权限检查
# -----------------------------------------------------------------------------
if [ "$MODE" != "check" ] && [ "$(id -u)" != "0" ]; then
    err "部署到 $TARGET 需要 root 权限"
    err "请使用：sudo $0 $*"
    exit 2
fi

# -----------------------------------------------------------------------------
# 铜锁源探测（2 级顺序）
# -----------------------------------------------------------------------------
detect_openssl() {
    # 1. 平台规范路径（CI 产物）
    local p="$SRC_ROOT/libs/openssl/$PLATFORM"
    if [ -x "$p/openssl" ]; then
        OPENSSL_SRC_DIR="$p"
        info "openssl 源：平台规范路径 $p"
        return 0
    fi

    # 2. core 源码目录的铜锁（开发常用）
    local c
    c="$(cd "$SRC_ROOT/.." 2>/dev/null && pwd)/core/libs/bin/tongsuo"
    if [ -x "$c/bin/openssl" ]; then
        OPENSSL_SRC_DIR="$c"
        info "openssl 源：core 源码路径 $c"
        return 0
    fi

    return 1
}

# -----------------------------------------------------------------------------
# 源文件检查
# -----------------------------------------------------------------------------
check_source() {
    local fail=0
    [ -x "$SRC_ROOT/bin/core" ] || { err "缺少 $SRC_ROOT/bin/core（先 make）"; fail=1; }
    [ -f "$SRC_ROOT/conf/core.yaml" ] || { err "缺少 $SRC_ROOT/conf/core.yaml"; fail=1; }
    [ -f "$SRC_ROOT/conf/algorithm-whitelist.yaml" ] || { err "缺少 $SRC_ROOT/conf/algorithm-whitelist.yaml"; fail=1; }

    if ! detect_openssl; then
        err "未找到铜锁 openssl，已尝试："
        err "  1. $SRC_ROOT/libs/openssl/$PLATFORM/openssl"
        err "  2. $SRC_ROOT/../core/libs/bin/tongsuo/bin/openssl"
        fail=1
    fi
    return $fail
}

# -----------------------------------------------------------------------------
# --check 模式
# -----------------------------------------------------------------------------
if [ "$MODE" = "check" ]; then
    echo "== 源文件检查 =="
    check_source && log "✅ 源文件检查通过" || { err "源文件检查失败"; exit 1; }

    echo ""
    echo "== 目标路径检查 =="
    if [ -d "$TARGET" ]; then
        info "目标目录已存在：$TARGET"
        if [ -f "$TARGET/conf/core.yaml" ]; then
            info "  发现现有配置，--upgrade 会保留"
        fi
        if [ -d "$TARGET/data/keys" ]; then
            info "  发现现有密钥，--upgrade 会保留"
        fi
    else
        info "目标目录不存在，将新建：$TARGET"
    fi
    exit 0
fi

# -----------------------------------------------------------------------------
# 源文件检查（install/upgrade 模式）
# -----------------------------------------------------------------------------
echo "== 检查源文件 =="
check_source || { err "源文件检查失败"; exit 1; }
log "✅ 源文件检查通过"

# -----------------------------------------------------------------------------
# 升级模式：备份关键文件
# -----------------------------------------------------------------------------
BACKUP_DIR=""
if [ "$MODE" = "upgrade" ] && [ -d "$TARGET" ]; then
    BACKUP_DIR="/opt/plats_tool/core-backup-$(date +%Y%m%d-%H%M%S)"
    log "升级模式：备份关键数据到 $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    [ -f "$TARGET/conf/core.yaml" ] && cp "$TARGET/conf/core.yaml" "$BACKUP_DIR/"
    [ -f "$TARGET/conf/algorithm-whitelist.yaml" ] && cp "$TARGET/conf/algorithm-whitelist.yaml" "$BACKUP_DIR/"
    if [ -d "$TARGET/data/keys" ]; then
        mkdir -p "$BACKUP_DIR/data"
        cp -a "$TARGET/data/keys" "$BACKUP_DIR/data/" 2>/dev/null || true
    fi
    log "✅ 备份完成"
fi

# -----------------------------------------------------------------------------
# 创建目标目录
# -----------------------------------------------------------------------------
echo "== 创建目录 =="
mkdir -p "$TARGET"/{bin,conf,libs,data/{keys,ca,certs,csr,crl,export},logs,tmp}

chmod 0755 "$TARGET" "$TARGET/bin" "$TARGET/libs"
chmod 0750 "$TARGET/conf" "$TARGET/logs"
chmod 0700 "$TARGET/data/keys" "$TARGET/tmp"
chmod 0750 "$TARGET/data" \
           "$TARGET/data/ca" "$TARGET/data/certs" \
           "$TARGET/data/csr" "$TARGET/data/crl" "$TARGET/data/export"

# -----------------------------------------------------------------------------
# 拷贝二进制
# -----------------------------------------------------------------------------
echo "== 拷贝 bin/core =="
install -m 0755 "$SRC_ROOT/bin/core" "$TARGET/bin/core"
log "✅ $TARGET/bin/core"

# -----------------------------------------------------------------------------
# 拷贝铜锁 openssl
# -----------------------------------------------------------------------------
echo "== 拷贝铜锁 openssl =="
TARGET_TONGSUO="$TARGET/libs/tongsuo"
TARGET_OPENSSL="$TARGET_TONGSUO/bin/openssl"

if [ -n "$OPENSSL_SRC_DIR" ] && [ -d "$OPENSSL_SRC_DIR" ]; then
    # 清理旧目标（软链或目录）
    if [ -L "$TARGET_TONGSUO" ]; then
        rm -f "$TARGET_TONGSUO"
    elif [ -d "$TARGET_TONGSUO" ]; then
        rm -rf "$TARGET_TONGSUO"
    fi
    mkdir -p "$TARGET_TONGSUO"

    # 兼容两种源布局：
    #   A) $SRC/libs/openssl/<platform>/{openssl, libs/}
    #   B) $SRC/{bin/openssl, libcrypto.so.*, libssl.so.*, include/}
    if [ -x "$OPENSSL_SRC_DIR/bin/openssl" ]; then
        # 布局 B：直接整目录拷
        cp -a "$OPENSSL_SRC_DIR/." "$TARGET_TONGSUO/"
    else
        # 布局 A：规范化为 bin/openssl + libs/
        mkdir -p "$TARGET_TONGSUO/bin"
        if [ -x "$OPENSSL_SRC_DIR/openssl" ]; then
            install -m 0755 "$OPENSSL_SRC_DIR/openssl" \
                "$TARGET_TONGSUO/bin/openssl"
        fi
        if [ -d "$OPENSSL_SRC_DIR/libs" ]; then
            mkdir -p "$TARGET_TONGSUO/libs"
            cp -a "$OPENSSL_SRC_DIR/libs/." "$TARGET_TONGSUO/libs/"
        fi
    fi

    chmod -R u+rwX,go+rX "$TARGET_TONGSUO" 2>/dev/null || true
    [ -x "$TARGET_OPENSSL" ] && chmod 0755 "$TARGET_OPENSSL"

    log "✅ $TARGET_TONGSUO"
    if [ -x "$TARGET_OPENSSL" ]; then
        info "  openssl 版本：$("$TARGET_OPENSSL" version 2>/dev/null || echo unknown)"
    fi
else
    err "铜锁源目录为空，无法部署"
    exit 1
fi

# -----------------------------------------------------------------------------
# 拷贝配置
# -----------------------------------------------------------------------------
echo "== 拷贝配置 =="
if [ "$MODE" = "upgrade" ] && [ -f "$TARGET/conf/core.yaml" ]; then
    info "保留现有 conf/core.yaml（不覆盖）"
else
    install -m 0640 "$SRC_ROOT/conf/core.yaml" "$TARGET/conf/core.yaml"
    log "✅ $TARGET/conf/core.yaml"
fi

if [ "$MODE" = "upgrade" ] && [ -f "$TARGET/conf/algorithm-whitelist.yaml" ]; then
    info "保留现有 conf/algorithm-whitelist.yaml"
else
    install -m 0640 "$SRC_ROOT/conf/algorithm-whitelist.yaml" \
        "$TARGET/conf/algorithm-whitelist.yaml"
    log "✅ $TARGET/conf/algorithm-whitelist.yaml"
fi

# -----------------------------------------------------------------------------
# ★ 强制写入 tongsuo_openssl_bin 为部署后绝对路径
# -----------------------------------------------------------------------------
#
# 说明：
#   core 二进制不读 CORE_ROOT 推导 openssl，
#   只认 conf/core.yaml 的 tongsuo_openssl_bin。
#   留空会回退到 PATH 里的系统 openssl → 报 tongsuo version too low。
#
# 因此无论 install 还是 upgrade，都必须把该字段写为部署后绝对路径。
echo "== 写入 tongsuo_openssl_bin =="
TARGET_YAML="$TARGET/conf/core.yaml"

if [ ! -f "$TARGET_YAML" ]; then
    err "conf/core.yaml 不存在：$TARGET_YAML"
    exit 1
fi
if [ ! -x "$TARGET_OPENSSL" ]; then
    err "部署后的铜锁 openssl 不可执行：$TARGET_OPENSSL"
    exit 1
fi

if grep -qE '^[[:space:]]*tongsuo_openssl_bin[[:space:]]*:' "$TARGET_YAML"; then
    # 已有字段：就地替换
    sed -i -E \
        "s#^([[:space:]]*tongsuo_openssl_bin[[:space:]]*:).*#\1 \"$TARGET_OPENSSL\"#" \
        "$TARGET_YAML"
else
    # 无字段：在 `core:` 行后插入（保持 2 空格缩进）
    if grep -qE '^core[[:space:]]*:' "$TARGET_YAML"; then
        sed -i "0,/^core[[:space:]]*:/s##core:\n  tongsuo_openssl_bin: \"$TARGET_OPENSSL\"#" \
            "$TARGET_YAML"
    else
        # 极端情况：无 core 段，追加一个
        printf '\ncore:\n  tongsuo_openssl_bin: "%s"\n' "$TARGET_OPENSSL" \
            >> "$TARGET_YAML"
    fi
fi

# 校验
if grep -qF "$TARGET_OPENSSL" "$TARGET_YAML"; then
    log "✅ core.yaml.tongsuo_openssl_bin → $TARGET_OPENSSL"
else
    err "写入 tongsuo_openssl_bin 失败，请手工检查：$TARGET_YAML"
    exit 1
fi

# -----------------------------------------------------------------------------
# 主密钥处理
# -----------------------------------------------------------------------------
if [ ! -f "$TARGET/data/keys/master.key" ]; then
    echo "== 生成主密钥 =="
    umask 077
    if [ -x "$TARGET_OPENSSL" ]; then
        "$TARGET_OPENSSL" rand -hex 32 > "$TARGET/data/keys/master.key"
    elif command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 32 > "$TARGET/data/keys/master.key"
    else
        head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' \
            > "$TARGET/data/keys/master.key"
    fi
    echo "" >> "$TARGET/data/keys/master.key"
    chmod 0600 "$TARGET/data/keys/master.key"
    log "✅ 已生成 $TARGET/data/keys/master.key"
else
    info "保留现有主密钥"
fi

# -----------------------------------------------------------------------------
# VERSION 文件
# -----------------------------------------------------------------------------
if [ -f "$SRC_ROOT/VERSION" ]; then
    install -m 0644 "$SRC_ROOT/VERSION" "$TARGET/VERSION"
else
    echo "core-go-$(date +%Y%m%d)" > "$TARGET/VERSION"
    chmod 0644 "$TARGET/VERSION"
fi

# -----------------------------------------------------------------------------
# 首次运行校验（验证 core 能正确解析铜锁）
# -----------------------------------------------------------------------------
echo "== 首次运行校验 =="
RUNTIME_OK=0

# 1) 尝试 --help
if CORE_ROOT="$TARGET" "$TARGET/bin/core" --help >/dev/null 2>&1; then
    RUNTIME_OK=1
    log "✅ core --help 执行成功"
fi

# 2) 若 --help 不存在，用一次无害操作验证（验证铜锁可用）
if [ "$RUNTIME_OK" -ne 1 ]; then
    cat > /tmp/_core_init_req.json <<'EOF'
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"init",
 "actor":{"type":"init","id":"deploy"},"params":{"cert_path":"data/ca/__init__.pem"}}
EOF
    if CORE_ROOT="$TARGET" "$TARGET/bin/core" \
        --op cert.parse \
        --in /tmp/_core_init_req.json \
        --out /tmp/_core_init_resp.json \
        >/dev/null 2>&1; then
        RUNTIME_OK=1
        log "✅ core 首次运行校验通过（openssl 解析成功）"
    else
        warn "core 首次运行校验失败（不影响部署，但运行期可能报错）"
        warn "  诊断："
        warn "    CORE_ROOT=$TARGET $TARGET/bin/core --help"
        warn "    $TARGET_OPENSSL version"
        warn "    grep tongsuo_openssl_bin $TARGET_YAML"
    fi
    rm -f /tmp/_core_init_req.json /tmp/_core_init_resp.json
fi

# 重新确保权限
chmod 0700 "$TARGET/data/keys" "$TARGET/tmp" 2>/dev/null || true
chmod 0750 "$TARGET/conf" "$TARGET/logs" 2>/dev/null || true
chmod 0750 "$TARGET/data" 2>/dev/null || true
[ -f "$TARGET/data/keys/master.key" ] && chmod 0600 "$TARGET/data/keys/master.key"
[ -f "$TARGET/conf/core.yaml" ] && chmod 0640 "$TARGET/conf/core.yaml"
[ -f "$TARGET/conf/algorithm-whitelist.yaml" ] && \
    chmod 0640 "$TARGET/conf/algorithm-whitelist.yaml"

# -----------------------------------------------------------------------------
# 完成
# -----------------------------------------------------------------------------
echo ""
echo "=================================================="
printf "${GREEN}✅ core-go 部署完成${NC}\n"
echo "=================================================="
echo "  目标目录：  $TARGET"
echo "  二进制：    $TARGET/bin/core"
echo "  配置文件：  $TARGET/conf/core.yaml"
echo "  白名单：    $TARGET/conf/algorithm-whitelist.yaml"
echo "  铜锁：      $TARGET_OPENSSL"
echo "  主密钥：    $TARGET/data/keys/master.key"
echo "  审计日志：  $TARGET/logs/audit.jsonl"
echo ""
if [ -n "$BACKUP_DIR" ]; then
    echo "  升级备份：  $BACKUP_DIR"
    echo ""
fi
echo "  平台后端配置（.env 或 systemd）："
echo "    CORE_DISPATCH_PATH=$TARGET/bin/core"
echo "    CORE_ROOT=$TARGET"
echo "    CORE_TIMEOUT_MS=5000"
echo "    CORE_MAX_CONCURRENCY=8"
echo ""
echo "  验证命令："
echo "    $TARGET_OPENSSL version"
echo "    grep tongsuo_openssl_bin $TARGET/conf/core.yaml"
echo "    CORE_ROOT=$TARGET $TARGET/bin/core --help"
echo "=================================================="
