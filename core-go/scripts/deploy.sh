#!/usr/bin/env bash
# =============================================================================
# core-go 部署脚本
#
# 用途：将开发目录的 core-go 安装到 /opt/core（生产路径）
#
# 用法：
#   sudo ./scripts/deploy.sh                    # 完整部署
#   sudo ./scripts/deploy.sh --upgrade          # 升级（保留 data/keys/、conf/）
#   sudo ./scripts/deploy.sh --target /srv/core # 自定义目标
#   sudo ./scripts/deploy.sh --check            # 只检查环境，不部署
#
# 部署内容：
#   - bin/core                ← 从 ./bin/core 拷贝
#   - libs/openssl/...        ← 从 ./libs/openssl/... 拷贝
#   - conf/core.yaml          ← 从 ./conf/core.yaml 拷贝（升级时保留旧的）
#   - conf/algorithm-whitelist.yaml
#   - data/keys/              ← 空目录（首次）/ 保留（升级）
#   - logs/、tmp/             ← 空目录
# =============================================================================
set -euo pipefail

# -----------------------------------------------------------------------------
# 基础路径
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TARGET="/opt/core"
MODE="install"   # install | upgrade | check
KEEP_DATA=0
KEEP_CONF=0

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
  --target PATH       自定义安装目录，默认 /opt/core
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
# 源文件检查
# -----------------------------------------------------------------------------
check_source() {
    local fail=0
    [ -x "$SRC_ROOT/bin/core" ] || { err "缺少 $SRC_ROOT/bin/core（先 make）"; fail=1; }
    [ -f "$SRC_ROOT/conf/core.yaml" ] || { err "缺少 $SRC_ROOT/conf/core.yaml"; fail=1; }
    [ -f "$SRC_ROOT/conf/algorithm-whitelist.yaml" ] || { err "缺少 $SRC_ROOT/conf/algorithm-whitelist.yaml"; fail=1; }

    local openssl_src="$SRC_ROOT/libs/openssl/$PLATFORM/openssl"
    if [ ! -x "$openssl_src" ]; then
        # 也接受 conf 指定的
        local cfg_openssl=""
        cfg_openssl=$(grep -E '^[[:space:]]*tongsuo_openssl_bin[[:space:]]*:' \
            "$SRC_ROOT/conf/core.yaml" 2>/dev/null \
            | head -1 | sed -E 's/^[^:]+:[[:space:]]*//' | tr -d '"' | tr -d "'")
        if [ -n "$cfg_openssl" ] && [ -x "$cfg_openssl" ]; then
            info "openssl 使用 conf 指定路径：$cfg_openssl"
        else
            err "缺少 $openssl_src，且 conf/core.yaml 未指定有效 openssl"
            fail=1
        fi
    else
        info "openssl 使用平台路径：$openssl_src"
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
    BACKUP_DIR="/opt/core-backup-$(date +%Y%m%d-%H%M%S)"
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

# 权限
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
# 拷贝铜锁
# -----------------------------------------------------------------------------
echo "== 拷贝铜锁 openssl =="
OPENSSL_SRC="$SRC_ROOT/libs/openssl/$PLATFORM"
if [ -d "$OPENSSL_SRC" ]; then
    mkdir -p "$TARGET/libs/openssl/$PLATFORM"
    # 拷贝 openssl 可执行
    if [ -x "$OPENSSL_SRC/openssl" ]; then
        install -m 0755 "$OPENSSL_SRC/openssl" \
            "$TARGET/libs/openssl/$PLATFORM/openssl"
        log "✅ $TARGET/libs/openssl/$PLATFORM/openssl"
    fi
    # 拷贝动态库（若存在）
    if [ -d "$OPENSSL_SRC/libs" ]; then
        mkdir -p "$TARGET/libs/openssl/$PLATFORM/libs"
        cp -a "$OPENSSL_SRC/libs/." "$TARGET/libs/openssl/$PLATFORM/libs/"
        chmod -R 0755 "$TARGET/libs/openssl/$PLATFORM/libs"
        log "✅ $TARGET/libs/openssl/$PLATFORM/libs/"
    fi
else
    warn "源目录无 $OPENSSL_SRC，将依赖 conf/core.yaml 指定 openssl"
fi

# -----------------------------------------------------------------------------
# 拷贝配置（升级模式保留现有配置）
# -----------------------------------------------------------------------------
echo "== 拷贝配置 =="
if [ "$MODE" = "upgrade" ] && [ -f "$TARGET/conf/core.yaml" ]; then
    info "保留现有 conf/core.yaml"
else
    install -m 0640 "$SRC_ROOT/conf/core.yaml" "$TARGET/conf/core.yaml"
    # 修正路径：把 dev 环境的绝对路径替换为 /opt/core
    if [ "$TARGET" != "/opt/core" ]; then
        sed -i "s#/code/platform_all_use/core-go#$TARGET#g" "$TARGET/conf/core.yaml" 2>/dev/null || true
    fi
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
# 主密钥处理
# -----------------------------------------------------------------------------
if [ ! -f "$TARGET/data/keys/master.key" ]; then
    echo "== 生成主密钥 =="
    umask 077
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 32 > "$TARGET/data/keys/master.key"
    elif [ -x "$TARGET/libs/openssl/$PLATFORM/openssl" ]; then
        "$TARGET/libs/openssl/$PLATFORM/openssl" rand -hex 32 \
            > "$TARGET/data/keys/master.key"
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
# 首次运行：初始化目录 + 生成 master.key
# -----------------------------------------------------------------------------
echo "== 首次运行校验 =="
if ! CORE_ROOT="$TARGET" "$TARGET/bin/core" --help >/dev/null 2>&1; then
    # --help 可能未实现，尝试一次无害操作（cert.parse 不存在的证书）
    cat > /tmp/_core_init_req.json <<'EOF'
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"init",
 "actor":{"type":"init","id":"deploy"},"params":{"cert_path":"data/ca/__init__.pem"}}
EOF
    CORE_ROOT="$TARGET" "$TARGET/bin/core" \
        --op cert.parse \
        --in /tmp/_core_init_req.json \
        --out /tmp/_core_init_resp.json \
        >/dev/null 2>&1 || true
    rm -f /tmp/_core_init_req.json /tmp/_core_init_resp.json
fi

# 重新确保权限（core 可能在初始化时创建了文件）
chmod 0700 "$TARGET/data/keys" "$TARGET/tmp" 2>/dev/null || true
chmod 0750 "$TARGET/conf" "$TARGET/logs" 2>/dev/null || true
chmod 0750 "$TARGET/data" 2>/dev/null || true
[ -f "$TARGET/data/keys/master.key" ] && chmod 0600 "$TARGET/data/keys/master.key"
[ -f "$TARGET/conf/core.yaml" ] && chmod 0640 "$TARGET/conf/core.yaml"
[ -f "$TARGET/conf/algorithm-whitelist.yaml" ] && chmod 0640 "$TARGET/conf/algorithm-whitelist.yaml"

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
echo "  铜锁：      $TARGET/libs/openssl/$PLATFORM/openssl"
echo "  主密钥：    $TARGET/data/keys/master.key"
echo "  审计日志：  $TARGET/logs/audit.jsonl"
echo ""
if [ -n "$BACKUP_DIR" ]; then
    echo "  升级备份：  $BACKUP_DIR"
    echo ""
fi
echo "  平台后端配置（.env 或 systemd）："
echo "    CORE_DISPATCH_PATH=$TARGET/bin/core"
echo "    CORE_TIMEOUT_MS=5000"
echo "    CORE_MAX_CONCURRENCY=8"
echo ""
echo "  验证命令："
echo "    $TARGET/bin/core --help"
echo "    echo '{\"schema_version\":\"1.0\",\"operation_id\":\"cert.parse\",\"request_id\":\"t\",\"actor\":{\"type\":\"t\",\"id\":\"t\"},\"params\":{\"cert_path\":\"data/ca/test.pem\"}}' \\"
echo "      > /tmp/req.json"
echo "    $TARGET/bin/core --op cert.parse --in /tmp/req.json --out /tmp/resp.json"
echo "=================================================="
