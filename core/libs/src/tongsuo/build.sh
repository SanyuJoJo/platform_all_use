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
