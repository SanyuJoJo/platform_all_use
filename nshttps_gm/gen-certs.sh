#!/usr/bin/env bash
# ============================================================
# nshttps_gm 国密证书自动生成脚本
#
# 功能：
#   1. SM2 根 CA + SM2 签名/加密双证书  → 国密 NTLS 通道
#   2. 独立 ECC(P-256) 根 CA + ECC 服务器证书 → 国际 TLS 通道
#   （SM2 CA 只能以 SM3 签名，无法签 SHA-256 的 ECC 证书，故国际通道用独立 ECC CA）
#
# 有效期：CA 与服务器证书均默认 3650 天（10 年，避免频繁过期）
#
# 用法：
#   bash gen-certs.sh [--days N] [--cn 主机名] [--ca-days N]
#
# 输出目录：$NSG_HOME/certs/
#   ca/sm2-root.key / sm2-root.crt              SM2 根 CA（国密）
#   ca/ecc-root.key / ecc-root.crt              ECC(P-256) 根 CA（国际）
#   server/server.key / server.crt              SM2 服务器签名证书（NTLS 签名）
#   server/server_enc.key / server_enc.crt      SM2 服务器加密证书（NTLS 加密）
#   server/international.key / international.crt  ECC 服务器证书（国际 TLS）
#   server/chain.crt                            SM2 签名 + SM2 根 CA 链
#   server/international_chain.crt              ECC 国际 + ECC 根 CA 链
# ============================================================
set -euo pipefail

# ---------- 路径解析（自包含：使用 nshttps_gm 内部集成的 tongsuo） ----------
NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
TONGSUO_BIN="$NSG_HOME/tongsuo-bin"
TONGSUO_LIB="$TONGSUO_BIN/libs"
OPENSSL_BIN="$TONGSUO_BIN/bin/openssl"
OPENSSL_CONF="$TONGSUO_BIN/ssl/openssl.cnf"

# ---------- 参数 ----------
DAYS=3650
CA_DAYS=3650
CN="localhost"
while [ $# -gt 0 ]; do
  case "$1" in
    --days)    DAYS="$2";    shift 2 ;;
    --ca-days) CA_DAYS="$2"; shift 2 ;;
    --cn)      CN="$2";      shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

# ---------- 依赖检查 ----------
[ -x "$OPENSSL_BIN" ] || { echo "找不到铜锁 openssl：$OPENSSL_BIN（请先运行 build.sh 编译内部 tongsuo）" >&2; exit 2; }
[ -d "$TONGSUO_LIB" ] || { echo "找不到铜锁 lib：$TONGSUO_LIB" >&2; exit 2; }

export LD_LIBRARY_PATH="$TONGSUO_LIB:${LD_LIBRARY_PATH:-}"
if [ -f "$OPENSSL_CONF" ]; then
  export OPENSSL_CONF="$OPENSSL_CONF"
fi

echo "[nshttps_gm] 铜锁 openssl : $OPENSSL_BIN"
echo "[nshttps_gm] 有效期       : CA=${CA_DAYS}天 / 服务器证书=${DAYS}天"
echo "[nshttps_gm] CN           : $CN"

CERT_DIR="$NSG_HOME/certs"
CA_DIR="$CERT_DIR/ca"
SRV_DIR="$CERT_DIR/server"
TMP="$CERT_DIR/tmp"
rm -rf "$CERT_DIR"
mkdir -p "$CA_DIR" "$SRV_DIR" "$TMP"

SUBJ="/CN=$CN/O=nshttps-gm/C=CN"
SAN="DNS:localhost,DNS:$CN,IP:127.0.0.1"

# ============================================================
# A. 国密 NTLS 通道：SM2 CA + SM2 双证书
# ============================================================
echo "==> [国密] 生成 SM2 根 CA ..."
"$OPENSSL_BIN" genpkey -algorithm SM2 -out "$CA_DIR/sm2-root.key" 2>/dev/null
chmod 600 "$CA_DIR/sm2-root.key"
"$OPENSSL_BIN" req -new -x509 -key "$CA_DIR/sm2-root.key" \
  -out "$CA_DIR/sm2-root.crt" -days "$CA_DAYS" -subj "$SUBJ" -sm3 \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -addext "subjectKeyIdentifier=hash" 2>/dev/null

echo "==> [国密] SM2 签名服务器证书 ..."
"$OPENSSL_BIN" genpkey -algorithm SM2 -out "$SRV_DIR/server.key" 2>/dev/null
chmod 600 "$SRV_DIR/server.key"
"$OPENSSL_BIN" req -new -key "$SRV_DIR/server.key" \
  -out "$TMP/server.csr" -subj "$SUBJ" 2>/dev/null
cat > "$TMP/server_sign.ext" <<EOF
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=$SAN
EOF
"$OPENSSL_BIN" x509 -req -in "$TMP/server.csr" \
  -CA "$CA_DIR/sm2-root.crt" -CAkey "$CA_DIR/sm2-root.key" -CAcreateserial \
  -out "$SRV_DIR/server.crt" -days "$DAYS" -sm3 \
  -extfile "$TMP/server_sign.ext" 2>/dev/null

echo "==> [国密] SM2 加密服务器证书 ..."
"$OPENSSL_BIN" genpkey -algorithm SM2 -out "$SRV_DIR/server_enc.key" 2>/dev/null
chmod 600 "$SRV_DIR/server_enc.key"
"$OPENSSL_BIN" req -new -key "$SRV_DIR/server_enc.key" \
  -out "$TMP/server_enc.csr" -subj "$SUBJ" 2>/dev/null
cat > "$TMP/server_enc.ext" <<EOF
basicConstraints=critical,CA:FALSE
keyUsage=critical,keyEncipherment,keyAgreement
extendedKeyUsage=serverAuth
subjectAltName=$SAN
EOF
"$OPENSSL_BIN" x509 -req -in "$TMP/server_enc.csr" \
  -CA "$CA_DIR/sm2-root.crt" -CAkey "$CA_DIR/sm2-root.key" -CAcreateserial \
  -out "$SRV_DIR/server_enc.crt" -days "$DAYS" -sm3 \
  -extfile "$TMP/server_enc.ext" 2>/dev/null
cat "$SRV_DIR/server.crt" "$CA_DIR/sm2-root.crt" > "$SRV_DIR/chain.crt"

# ============================================================
# B. 国际 TLS 通道：独立 ECC(P-256) CA + ECC 服务器证书
#    （SM2 CA 只能 SM3 签名，无法签 SHA-256 的 ECC，故用独立 ECC CA）
# ============================================================
echo "==> [国际] 生成 ECC(P-256) 根 CA ..."
"$OPENSSL_BIN" ecparam -name prime256v1 -genkey -noout -out "$CA_DIR/ecc-root.key" 2>/dev/null
chmod 600 "$CA_DIR/ecc-root.key"
"$OPENSSL_BIN" req -new -x509 -key "$CA_DIR/ecc-root.key" \
  -out "$CA_DIR/ecc-root.crt" -days "$CA_DAYS" -subj "$SUBJ" -sha256 \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -addext "subjectKeyIdentifier=hash" 2>/dev/null

echo "==> [国际] ECC(P-256) 服务器证书 ..."
"$OPENSSL_BIN" ecparam -name prime256v1 -genkey -noout -out "$SRV_DIR/international.key" 2>/dev/null
chmod 600 "$SRV_DIR/international.key"
"$OPENSSL_BIN" req -new -key "$SRV_DIR/international.key" \
  -out "$TMP/international.csr" -subj "$SUBJ" 2>/dev/null
cat > "$TMP/international.ext" <<EOF
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=$SAN
EOF
"$OPENSSL_BIN" x509 -req -in "$TMP/international.csr" \
  -CA "$CA_DIR/ecc-root.crt" -CAkey "$CA_DIR/ecc-root.key" -CAcreateserial \
  -out "$SRV_DIR/international.crt" -days "$DAYS" -sha256 \
  -extfile "$TMP/international.ext" 2>/dev/null
cat "$SRV_DIR/international.crt" "$CA_DIR/ecc-root.crt" > "$SRV_DIR/international_chain.crt"

# ---------- 清理临时 ----------
rm -rf "$TMP"

echo
echo "[nshttps_gm] 证书已生成（均 ${DAYS} 天）："
ls -la "$CA_DIR" "$SRV_DIR" | grep -E '\.key|\.crt'
echo
echo "[nshttps_gm] 完成。"
echo "[nshttps_gm] 国密 NTLS：server.crt + server_enc.crt（SM2 CA 签发）"
echo "[nshttps_gm] 国际 TLS ：international.crt（独立 ECC CA 签发，标准浏览器可用）"