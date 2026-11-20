#!/usr/bin/env bash
# ============================================================
# nshttps_gm 自包含构建脚本
#
# 内部集成 Tongsuo（铜锁）源码，不依赖外部库。
# 步骤：
#   1. 编译内部 Tongsuo → <nsg>/tongsuo-bin（libs/bin/openssl, libs/libcrypto/libssl）
#   2. 用内部 Tongsuo 编译 Tengine → <nsg>/tengine-3.1.0/bin/sbin/nginx
#
# 产物自带 rpath 指向 <nsg>/tongsuo-bin/libs，可整体打包部署到任何 Linux x86_64。
#
# 用法：
#   bash build.sh [--force] [--prefix /opt/nshttps_gm]
#     --force   强制重新 configure + make（否则复用已有产物）
#     --prefix  部署目标前缀（写入 rpath 第二路径，默认 --prefix 的 <nsg>/tongsuo-bin/libs）
# ============================================================
set -euo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
TONGSUO_SRC="$NSG_HOME/tongsuo"
TONGSUO_PREFIX="$NSG_HOME/tongsuo-bin"
TENGINE="$NSG_HOME/tengine-3.1.0"
TONGSUO_LIB="$TONGSUO_PREFIX/libs"
TONGSUO_INC="$TONGSUO_PREFIX/include"

FORCE=0
DEPLOY_PREFIX="$NSG_HOME/tongsuo-bin"
while [ $# -gt 0 ]; do
  case "$1" in
    --force)  FORCE=1; shift ;;
    --prefix) DEPLOY_PREFIX="$2"; shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done
JOBS=${JOBS:-$(nproc 2>/dev/null || echo 2)}

# ---------- 1. 编译内部 Tongsuo ----------
if [ -x "$TONGSUO_PREFIX/bin/openssl" ] && [ -f "$TONGSUO_LIB/libcrypto.so.3" ] && [ "$FORCE" -eq 0 ]; then
  echo "[nshttps_gm] 内部 Tongsuo 已编译，跳过"
else
  echo "==> 编译内部 Tongsuo（$TONGSUO_SRC）..."
  [ -d "$TONGSUO_SRC" ] || { echo "缺 Tongsuo 源码：$TONGSUO_SRC" >&2; exit 2; }
  mkdir -p "$TONGSUO_PREFIX"
  TONSUO_BUILD_PREFIX="$TONGSUO_PREFIX" \
  TONSUO_RUNTIME_LIB_DIR="$TONGSUO_LIB" \
  JOBS="$JOBS" \
    bash "$TONGSUO_SRC/build.sh"
  [ -x "$TONGSUO_PREFIX/bin/openssl" ] || { echo "Tongsuo 编译失败" >&2; exit 9; }
  echo "[nshttps_gm] Tongsuo 完成：$TONGSUO_PREFIX"
fi

# ---------- 2. 编译 Tengine（链接内部 Tongsuo） ----------
if [ -x "$TENGINE/bin/sbin/nginx" ] && [ "$FORCE" -eq 0 ]; then
  echo "[nshttps_gm] Tengine 已编译，跳过"
  echo "[nshttps_gm] 如需重建（以指向内部 Tongsuo）请加 --force"
  # 检查 rpath 是否已指向内部 tongsuo
  if LD_LIBRARY_PATH="$TONGSUO_LIB" "$TENGINE/bin/sbin/nginx" -V 2>&1 | grep -q "$TONGSUO_LIB"; then
    echo "[nshttps_gm] rpath 已指向内部 Tongsuo ✓"
  else
    echo "[nshttps_gm] WARN: rpath 未指向内部 Tongsuo，请加 --force 重建"
  fi
  exit 0
fi

[ -d "$TENGINE" ] || { echo "缺 Tengine 源码：$TENGINE" >&2; exit 2; }
cd "$TENGINE"

echo "==> 编译 Tengine（链接内部 Tongsuo）..."
./configure \
  --prefix="$TENGINE/bin" \
  --with-http_ssl_module \
  --with-http_v2_module \
  --with-stream \
  --with-stream_ssl_module \
  --with-stream_sni \
  --with-cc-opt="-I$TONGSUO_INC" \
  --with-ld-opt="-L$TONGSUO_LIB -Wl,-rpath,$TONGSUO_LIB -Wl,-rpath,$DEPLOY_PREFIX/libs" \
  --add-module="$TENGINE/modules/ngx_tongsuo_ntls"
make -j"$JOBS"
make install

[ -x "$TENGINE/bin/sbin/nginx" ] || { echo "Tengine 编译失败" >&2; exit 9; }

# ---------- 清理 tongsuo 源码树构建产物（保持源码纯净，利于 git 提交） ----------
cd "$TONGSUO_SRC"
make distclean >/dev/null 2>&1 || true
rm -f libcrypto.a libcrypto.so libcrypto.so.3 libssl.a libssl.so libssl.so.3 \
      configdata.pm builddata.pm installdata.pm config.log config.status 2>/dev/null || true
find . -name '*.o' -delete 2>/dev/null || true
find . -name '*.d' -delete 2>/dev/null || true
find . -name '*.lo' -delete 2>/dev/null || true
find . -name '*.a' -delete 2>/dev/null || true

echo
echo "[nshttps_gm] 自包含构建完成："
echo "  nginx     : $TENGINE/bin/sbin/nginx"
echo "  tongsuo   : $TONGSUO_PREFIX（运行库，保留）"
echo "  tongsuo   : $TONGSUO_SRC（源码，已清理构建产物）"
echo "  rpath     : $TONGSUO_LIB"
echo
echo "[nshttps_gm] 验证版本："
LD_LIBRARY_PATH="$TONGSUO_LIB" "$TENGINE/bin/sbin/nginx" -V 2>&1 | head -2