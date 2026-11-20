#!/usr/bin/env bash
# ============================================================================
# core 载荷构建脚本（安装包含 core 用）
# 位置：create_package/build-core-payload.sh
#
# 用途：从构建树编译 core-go，并组装出**可分发**的 core.tar.gz 载荷：
#   <staging>/bin/core                  ← core-go 静态二进制（重新编译）
#   <staging>/bin/{keycrypt,whitebox_sm4,sm2_decrypt_helper,cfca_envelope}
#   <staging>/sbin/nginx                ← Tengine（x86_64）
#   <staging>/conf/core.yaml            ← 相对路径配置（任意目录可用）
#   <staging>/conf/algorithm-whitelist.yaml
#   <staging>/conf/mime.types           ← nginx mime 映射
#   <staging>/conf/template/*.tmpl      ← SSL 代理模板
#   <staging>/libs/tongsuo/bin/openssl
#   <staging>/libs/tongsuo/libs/{libcrypto.so*,libssl.so*,engines-3,ossl-modules}
#   <staging>/libs/tongsuo/ssl/         ← openssl OPENSSLDIR（cnf 等）
#   <staging>/data,logs,tmp,pid,run     ← 空骨架
#   <staging>/VERSION
#
# 输出：create_package/bin/bin/core.tar.gz
#
# ★ 裁剪原则：
#   - 不含任何运行数据（core/data/keys、core/data/ca 等一律不打包）；
#   - 铜锁丢弃 .a / include / cmake / pkgconfig，只留运行必需物；
#   - oqsprovider.so 从部署目录或源树查找（源树无则警告跳过）。
#
# 用法：
#   bash create_package/build-core-payload.sh
#   环境变量：
#     CORE_ROOT_SRC  构建树根（默认 <repo>/core）
#     CORE_GO_DIR    core-go 源码目录（默认 <repo>/core-go）
#     CORE_GO_SKIP_BUILD=1  跳过 go build（复用已有 bin/core）
# ============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_SRC="${CORE_ROOT_SRC:-$ROOT/core}"
CORE_GO_DIR="${CORE_GO_DIR:-$ROOT/core-go}"
PAYLOAD_DIR="$ROOT/create_package/bin/bin"
OUT_TGZ="$PAYLOAD_DIR/core.tar.gz"
TONGSUO_SRC="$CORE_SRC/libs/bin/tongsuo"
STAGING="$(mktemp -d)"
trap 'rm -rf "$STAGING"' EXIT

log() { echo -e "\n===== [core-payload] $* ====="; }
die() { echo -e "\n❌ [core-payload] $*" >&2; exit 1; }

mkdir -p "$PAYLOAD_DIR"

# ---------------------------------------------------------------------------
# 1) core-go 编译
# ---------------------------------------------------------------------------
log "[1/5] 编译 core-go（静态）"
CORE_BIN_SRC=""
if [ "${CORE_GO_SKIP_BUILD:-0}" = "1" ]; then
  CORE_BIN_SRC="$CORE_GO_DIR/bin/core"
  [ -x "$CORE_BIN_SRC" ] || die "CORE_GO_SKIP_BUILD=1 但 $CORE_BIN_SRC 不可执行"
  log "跳过编译，复用：$CORE_BIN_SRC"
else
  if ! command -v go >/dev/null 2>&1; then
    die "缺少 go 工具链（可用 CORE_GO_SKIP_BUILD=1 复用已有二进制）"
  fi
  GO_BIN_DIR="$(mktemp -d)"
  ( cd "$CORE_GO_DIR" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -trimpath -ldflags="-s -w" \
      -o "$GO_BIN_DIR/core" ./cmd/core )
  CORE_BIN_SRC="$GO_BIN_DIR/core"
  chmod +x "$CORE_BIN_SRC"
fi
"$CORE_BIN_SRC" -h >/dev/null 2>&1 || true
log "core-go 二进制就绪：$CORE_BIN_SRC"

# ---------------------------------------------------------------------------
# 2) 组装目录骨架
# ---------------------------------------------------------------------------
log "[2/5] 组装 core 目录骨架"
CORE="$STAGING/core"
mkdir -p "$CORE"/{bin,sbin,conf/template,libs,data,logs,tmp,pid,run}

# --- bin：core-go + C 工具 ---
cp "$CORE_BIN_SRC" "$CORE/bin/core"
chmod 755 "$CORE/bin/core"
if [ -d "$CORE_SRC/src/tool/bin" ]; then
  cp -f "$CORE_SRC"/src/tool/bin/* "$CORE/bin/" 2>/dev/null || true
fi
for f in keycrypt whitebox_sm4 sm2_decrypt_helper cfca_envelope; do
  [ -x "$CORE/bin/$f" ] || log "警告：C 工具 $f 未找到，跳过（不影响核心 CA 流程）"
done

# --- sbin：Tengine nginx ---
NGINX_SRC="$CORE_SRC/src/tengine-3.1.0/bin/sbin/nginx"
[ -x "$NGINX_SRC" ] || die "Tengine 二进制不存在：$NGINX_SRC"
cp "$NGINX_SRC" "$CORE/sbin/nginx"
chmod 755 "$CORE/sbin/nginx"

# --- conf ---
cp "$CORE_GO_DIR/conf/core.yaml" "$CORE/conf/core.yaml"
if [ -f "$CORE_GO_DIR/conf/algorithm-whitelist.yaml" ]; then
  cp "$CORE_GO_DIR/conf/algorithm-whitelist.yaml" "$CORE/conf/algorithm-whitelist.yaml"
else
  log "警告：core-go/conf/algorithm-whitelist.yaml 不存在，跳过"
fi
cp "$CORE_SRC/src/tengine-3.1.0/bin/conf/mime.types" "$CORE/conf/mime.types"
cp -f "$CORE_SRC/src/tengine-3.1.0/conf/template/"*.tmpl "$CORE/conf/template/" 2>/dev/null || true

# --- libs/tongsuo：裁剪 ---
log "裁剪铜锁（丢弃 .a/include/cmake/pkgconfig）"
mkdir -p "$CORE/libs/tongsuo"/{bin,libs,ssl}
[ -x "$TONGSUO_SRC/bin/openssl" ] || die "铜锁 openssl 不存在：$TONGSUO_SRC/bin/openssl"
cp "$TONGSUO_SRC/bin/openssl" "$CORE/libs/tongsuo/bin/openssl"
chmod 755 "$CORE/libs/tongsuo/bin/openssl"

# libs/*.so*
find "$TONGSUO_SRC/libs" -maxdepth 1 \( -name '*.so' -o -name '*.so.*' \) -exec cp -fL {} "$CORE/libs/tongsuo/libs/" \;
# 库目录下的 ossl-modules + engines-3（递归复制）
for sub in ossl-modules engines-3; do
  if [ -d "$TONGSUO_SRC/libs/$sub" ]; then
    cp -rf "$TONGSUO_SRC/libs/$sub" "$CORE/libs/tongsuo/libs/"
  fi
done
# openssl 运行时配置目录（OPENSSLDIR）
cp -rf "$TONGSUO_SRC/ssl/." "$CORE/libs/tongsuo/ssl/" 2>/dev/null || true

# --- oqs-provider（源树可能缺失，回退到部署目录）---
OQS_FOUND=0
for cand in \
  "$TONGSUO_SRC/libs/ossl-modules/oqsprovider.so" \
  "$TONGSUO_SRC/ossl-modules/oqsprovider.so" \
  "/opt/plats_tool/core/libs/tongsuo/libs/ossl-modules/oqsprovider.so" \
  "/opt/plats_tool/core/libs/tongsuo/ossl-modules/oqsprovider.so"; do
  if [ -f "$cand" ]; then
    mkdir -p "$CORE/libs/tongsuo/libs/ossl-modules"
    cp -fL "$cand" "$CORE/libs/tongsuo/libs/ossl-modules/oqsprovider.so"
    chmod 755 "$CORE/libs/tongsuo/libs/ossl-modules/oqsprovider.so"
    OQS_FOUND=1
    log "已包含 oqs-provider：$cand"
    break
  fi
done
if [ "$OQS_FOUND" -ne 1 ]; then
  log "⚠️ 警告：未找到 oqsprovider.so（Falcon 等 PQC 算法不可用，其余功能正常）"
fi

# 与 core.yaml 约定目录保持一致：libs/tongsuo/libs/ossl-modules 下
# 若只有 oqsprovider.so 而缺少 legacy.so，从 tongsuo 源 libs 复制
if [ -f "$TONGSUO_SRC/libs/ossl-modules/legacy.so" ] && \
   [ ! -f "$CORE/libs/tongsuo/libs/ossl-modules/legacy.so" ]; then
  cp -f "$TONGSUO_SRC/libs/ossl-modules/legacy.so" "$CORE/libs/tongsuo/libs/ossl-modules/legacy.so"
fi

# --- VERSION ---
CORE_VER="1.0.0"
if [ -f "$CORE_SRC/VERSION" ]; then
  CORE_VER="$(cat "$CORE_SRC/VERSION" | tr -d '[:space:]')"
fi
printf '%s\n' "$CORE_VER" > "$CORE/VERSION"

# ---------------------------------------------------------------------------
# 3) 可执行性校验
# ---------------------------------------------------------------------------
log "[3/5] 可执行性校验"
# core 静态 Go：直接运行 usage
if "$CORE/bin/core" -h >/dev/null 2>&1; then
  log "✅ core"
else
  log "✅ core（usage 退出码非 0 属正常）"
fi
# nginx：需要 LD_LIBRARY_PATH 指向铜锁 libs
if LD_LIBRARY_PATH="$CORE/libs/tongsuo/libs" "$CORE/sbin/nginx" -v >/dev/null 2>&1; then
  log "✅ nginx（Tengine）"
else
  die "nginx 校验失败（检查 LD_LIBRARY_PATH 与 Tengine 架构）"
fi
# openssl
if LD_LIBRARY_PATH="$CORE/libs/tongsuo/libs" "$CORE/libs/tongsuo/bin/openssl" version >/dev/null 2>&1; then
  log "✅ openssl（铜锁）"
else
  die "openssl 校验失败"
fi
# oqs-provider 可用性
if [ "$OQS_FOUND" -eq 1 ]; then
  if LD_LIBRARY_PATH="$CORE/libs/tongsuo/libs" OPENSSL_MODULES="$CORE/libs/tongsuo/libs/ossl-modules" \
     "$CORE/libs/tongsuo/bin/openssl" list -providers 2>/dev/null | grep -qi oqsprovider; then
    log "✅ oqs-provider 加载正常"
  else
    log "⚠️  oqs-provider 未在 providers 列表出现（不影响非 PQC 功能）"
  fi
fi

# ---------------------------------------------------------------------------
# 4) 打包
# ---------------------------------------------------------------------------
log "[4/5] 打包 core.tar.gz"
rm -f "$OUT_TGZ"
tar -czf "$OUT_TGZ" -C "$STAGING" core
log "输出：$OUT_TGZ"
du -h "$OUT_TGZ"

# ---------------------------------------------------------------------------
# 5) 包内结构抽检
# ---------------------------------------------------------------------------
log "[5/5] 包内结构抽检"
TMPCHK="$(mktemp -d)"
tar -xzf "$OUT_TGZ" -C "$TMPCHK"
for f in \
  core/bin/core core/sbin/nginx core/conf/core.yaml core/conf/mime.types \
  core/libs/tongsuo/bin/openssl core/libs/tongsuo/libs/libcrypto.so.3; do
  [ -e "$TMPCHK/$f" ] || die "core.tar.gz 缺少：$f"
done
[ -d "$TMPCHK/core/conf/template" ] || die "core.tar.gz 缺少 conf/template"
rm -rf "$TMPCHK"
log "✅ core 载荷构建完成"