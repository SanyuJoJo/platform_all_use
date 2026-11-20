#!/bin/sh
# ============================================================
# Tongsuo 编译脚本
# 路径：core/libs/src/tongsuo/build.sh
# 用法：由 core/libs/Makefile 自动调用
# 环境变量：
#   TONSUO_BUILD_PREFIX    - 编译产物安装前缀
#   TONSUO_RUNTIME_LIB_DIR - 运行时库目录（rpath 第二路径）
#   JOBS                   - 并行编译任务数
#   TONSUO_CONFIGURE_OPTS  - 额外的 Configure 选项
# ============================================================
set -eu
# ---------- 必需环境变量检查 ----------
: "${TONSUO_BUILD_PREFIX:?TONSUO_BUILD_PREFIX is required}"
: "${TONSUO_RUNTIME_LIB_DIR:?TONSUO_RUNTIME_LIB_DIR is required}"
: "${JOBS:=$(nproc 2>/dev/null || echo 2)}"
: "${TONSUO_CONFIGURE_OPTS:=}"
# ---------- 切换到脚本所在目录 ----------
cd "$(dirname "$0")"
# ---------- 默认国密编译选项 ----------
# 根据实际需求可调整，Tongsuo 的编译选项采用 enable-xx 格式
TONGSUO_NTLS_OPTS="enable-ntls enable-sm2 enable-sm3 enable-sm4"
# ---------- Configure ----------
./Configure \
  --prefix="$TONSUO_BUILD_PREFIX" \
  --openssldir="$TONSUO_BUILD_PREFIX/ssl" \
  --libdir=libs \
  -Wl,-rpath,"$TONSUO_BUILD_PREFIX/libs" \
  -Wl,-rpath,"$TONSUO_RUNTIME_LIB_DIR" \
  shared \
  $TONGSUO_NTLS_OPTS \
  $TONSUO_CONFIGURE_OPTS
# ---------- 编译 ----------
make -j"$JOBS"
# ---------- 安装软件（不含文档） ----------
make install_sw
# ---------- 安装 SSL 目录内容（openssl.cnf 等） ----------
make install_ssldirs
# ---------- 兜底：确保 openssl.cnf 存在 ----------
if [ ! -f "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf" ] && [ -f "apps/openssl.cnf" ]; then
  mkdir -p "$TONSUO_BUILD_PREFIX/ssl"
  cp apps/openssl.cnf "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf"
  chmod 0644 "$TONSUO_BUILD_PREFIX/ssl/openssl.cnf"
fi
# ---------- 验证产物 ----------
if [ ! -x "$TONSUO_BUILD_PREFIX/bin/openssl" ]; then
  echo "[build.sh] ERROR: openssl binary not found at $TONSUO_BUILD_PREFIX/bin/openssl" >&2
  exit 9
fi
echo "[build.sh] Tongsuo build completed: $TONSUO_BUILD_PREFIX"
