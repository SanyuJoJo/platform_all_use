#!/bin/sh
# ============================================================
# Tengine 编译脚本（集成 Tongsuo 铜锁 + NTLS 支持）
# 路径：core/src/tengine/build.sh
# 用法：由 core/Makefile 自动调用
#
# 环境变量：
#   TENGINE_BUILD_PREFIX   - Tengine 安装前缀
#   TONSUO_BUILD_PREFIX    - Tongsuo 源码目录（仅供 sanity 检查，不传给 configure）
#   TONSUO_BUILD_INC_DIR   - Tongsuo 编译产物的 include 目录（含 NTLS 声明）
#   TONSUO_BUILD_LIB_DIR   - Tongsuo 编译产物的 lib 目录（共享库所在）
#   TONSUO_RUNTIME_LIB_DIR - Tongsuo 运行时库目录（部署目录）
#   JOBS                   - 并行编译任务数
#   TENGINE_CONFIGURE_OPTS - 额外的 configure 选项
# ============================================================
set -eu

# ---------- 必需环境变量检查 ----------
: "${TENGINE_BUILD_PREFIX:?TENGINE_BUILD_PREFIX is required}"
: "${TONSUO_BUILD_PREFIX:?TONSUO_BUILD_PREFIX is required}"
: "${TONSUO_BUILD_INC_DIR:?TONSUO_BUILD_INC_DIR is required}"
: "${TONSUO_BUILD_LIB_DIR:?TONSUO_BUILD_LIB_DIR is required}"
: "${TONSUO_RUNTIME_LIB_DIR:?TONSUO_RUNTIME_LIB_DIR is required}"
JOBS=${JOBS:-$(nproc 2>/dev/null || echo 2)}
TENGINE_CONFIGURE_OPTS=${TENGINE_CONFIGURE_OPTS:-}

# ---------- 切换到脚本所在目录（Tengine 源码根目录） ----------
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

echo "[build.sh] src dir        : $SCRIPT_DIR"
echo "[build.sh] install prefix : $TENGINE_BUILD_PREFIX"
echo "[build.sh] tongsuo src    : $TONSUO_BUILD_PREFIX"
echo "[build.sh] tongsuo incdir : $TONSUO_BUILD_INC_DIR"
echo "[build.sh] tongsuo libdir : $TONSUO_BUILD_LIB_DIR"
echo "[build.sh] tongsuo rtdir  : $TONSUO_RUNTIME_LIB_DIR"
echo "[build.sh] jobs           : $JOBS"

# ---------- 检查 Tongsuo 源码是否存在（仅作 sanity 检查） ----------
if [ ! -d "$TONSUO_BUILD_PREFIX" ]; then
    echo "[build.sh] ERROR: Tongsuo source not found at $TONSUO_BUILD_PREFIX" >&2
    exit 2
fi
if [ ! -f "$TONSUO_BUILD_PREFIX/config" ]; then
    echo "[build.sh] ERROR: $TONSUO_BUILD_PREFIX does not look like OpenSSL/Tongsuo source (no ./config)" >&2
    exit 2
fi

# ---------- 检查 Tongsuo 编译产物的 include 目录 ----------
if [ ! -d "$TONSUO_BUILD_INC_DIR" ]; then
    echo "[build.sh] ERROR: Tongsuo build include dir not found at $TONSUO_BUILD_INC_DIR" >&2
    echo "[build.sh]        Did 'make -C libs all' finish successfully?" >&2
    exit 2
fi
if [ ! -f "$TONSUO_BUILD_INC_DIR/openssl/ssl.h" ]; then
    echo "[build.sh] ERROR: $TONSUO_BUILD_INC_DIR/openssl/ssl.h not found" >&2
    exit 2
fi

# ---------- 检查 Tongsuo 是否启用了 NTLS ----------
if ! grep -q "OPENSSL_NTLS" "$TONSUO_BUILD_INC_DIR/openssl/ssl.h"; then
    echo "[build.sh] WARNING: Tongsuo does not appear to have NTLS enabled (OPENSSL_NTLS not found in ssl.h)." >&2
    echo "[build.sh]          Please rebuild Tongsuo with './config enable-ntls' and reinstall." >&2
    echo "[build.sh]          Continuing, but NTLS support may not work correctly." >&2
fi

# ---------- 检查 Tongsuo 编译产物的 lib 目录 ----------
if [ ! -d "$TONSUO_BUILD_LIB_DIR" ]; then
    echo "[build.sh] ERROR: Tongsuo build lib dir not found at $TONSUO_BUILD_LIB_DIR" >&2
    echo "[build.sh]        Did 'make -C libs all' finish successfully?" >&2
    exit 2
fi

# 用 find 检查，避免 glob 里带空格时的分词坑
_found_ssl=$(find "$TONSUO_BUILD_LIB_DIR" -maxdepth 1 \
             \( -name 'libssl.so*' -o -name 'libssl*.dylib' \) -print -quit)
_found_crypto=$(find "$TONSUO_BUILD_LIB_DIR" -maxdepth 1 \
             \( -name 'libcrypto.so*' -o -name 'libcrypto*.dylib' \) -print -quit)

if [ -z "$_found_ssl" ] || [ -z "$_found_crypto" ]; then
    echo "[build.sh] ERROR: shared libssl/libcrypto not found in $TONSUO_BUILD_LIB_DIR" >&2
    echo "[build.sh]        make sure Tongsuo was configured with 'shared'." >&2
    echo "[build.sh]        contents of $TONSUO_BUILD_LIB_DIR:" >&2
    ls -la "$TONSUO_BUILD_LIB_DIR" >&2 || true
    exit 2
fi

echo "[build.sh] libssl found   : $_found_ssl"
echo "[build.sh] libcrypto found: $_found_crypto"

# ---------- 检查 NTLS 模块是否存在 ----------
NTLS_MODULE_DIR=""
for d in modules/ngx_tongsuo_ntls modules/ngx_openssl_ntls; do
    if [ -d "$d" ]; then
        NTLS_MODULE_DIR="$d"
        break
    fi
done
if [ -z "$NTLS_MODULE_DIR" ]; then
    echo "[build.sh] WARNING: NTLS module not found, NTLS support will be disabled" >&2
else
    echo "[build.sh] NTLS module found: $NTLS_MODULE_DIR"
fi

# ---------- 构建 configure 参数列表 ----------
# 关键：不再使用 --with-openssl=，避免 nginx 在 $OPENSSL/.openssl/ 下
#       重编一份 Tongsuo 静态库（浪费几分钟且和 libs 的产物重复）。
#       改为通过 --with-cc-opt / --with-ld-opt 直接使用 libs 编好的共享库。
set -- \
    --prefix="$TENGINE_BUILD_PREFIX" \
    --with-http_ssl_module \
    --with-http_v2_module \
    --with-stream \
    --with-stream_ssl_module \
    --with-stream_sni \
    --with-cc-opt="-I$TONSUO_BUILD_INC_DIR" \
    --with-ld-opt="-L$TONSUO_BUILD_LIB_DIR -Wl,-rpath,$TONSUO_BUILD_LIB_DIR -Wl,-rpath,$TONSUO_RUNTIME_LIB_DIR"

# NTLS 模块存在则加入
if [ -n "$NTLS_MODULE_DIR" ]; then
    set -- "$@" --add-module="$NTLS_MODULE_DIR"
fi

# 追加用户自定义选项：只对这一段做一次 shell 词拆分
if [ -n "$TENGINE_CONFIGURE_OPTS" ]; then
    # shellcheck disable=SC2086
    eval "set -- \"\$@\" $TENGINE_CONFIGURE_OPTS"
fi

# ---------- 执行 configure ----------
echo "[build.sh] ./configure $*"
./configure "$@"

# ---------- 编译 ----------
make -j"$JOBS"

# ---------- 安装 ----------
make install

# ---------- 验证产物 ----------
if [ ! -x "$TENGINE_BUILD_PREFIX/sbin/nginx" ]; then
    echo "[build.sh] ERROR: nginx binary not found at $TENGINE_BUILD_PREFIX/sbin/nginx" >&2
    exit 9
fi

echo "[build.sh] Tengine build completed: $TENGINE_BUILD_PREFIX"
