#!/bin/sh
set -eu

: "${TONSUO_BUILD_PREFIX:?TONSUO_BUILD_PREFIX is required}"
: "${TONSUO_RUNTIME_LIB_DIR:?TONSUO_RUNTIME_LIB_DIR is required}"
: "${JOBS:=$(nproc 2>/dev/null || echo 2)}"
: "${TONSUO_CONFIGURE_OPTS:=}"

cd "$(dirname "$0")"

./Configure \
  --prefix="$TONSUO_BUILD_PREFIX" \
  --openssldir="$TONSUO_BUILD_PREFIX/ssl" \
  --libdir=libs \
  -Wl,-rpath,"$TONSUO_BUILD_PREFIX/libs" \
  -Wl,-rpath,"$TONSUO_RUNTIME_LIB_DIR" \
  shared \
  $TONSUO_CONFIGURE_OPTS

make -j"$JOBS"
make install_sw

make install_ssldirs   # <<< 新增：安装 ssl/ 目录内容（openssl.cnf 等）
# 兜底：若 install_ssldirs 未生成 openssl.cnf，从源码复制
if [ ! -f "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf" ] && [ -f "apps/openssl.cnf" ]; then
  mkdir -p "$TONSUO_BUILD_PREFIX/ssl"
  cp apps/openssl.cnf "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf"
  chmod 0644 "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf"
fi
