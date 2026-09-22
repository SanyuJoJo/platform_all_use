#!/bin/sh
# core/tests/test_keycrypt.sh
# keycrypt AES-256-GCM 回环 + 篡改检测 + 错误密钥检测

set -eu

CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
KEYCRYPT_BIN="${KEYCRYPT_BIN:-$CORE_ROOT/src/tool/bin/keycrypt}"
OPENSSL_BIN="${BUILD_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

[ -x "$KEYCRYPT_BIN" ] || { echo "keycrypt missing: $KEYCRYPT_BIN" >&2; exit 1; }

if [ -x "$OPENSSL_BIN" ]; then
  "$OPENSSL_BIN" rand -hex 32 > "$TMP_DIR/master.key"
else
  head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' > "$TMP_DIR/master.key"
fi
chmod 0600 "$TMP_DIR/master.key"

printf 'hello-keycrypt\n' > "$TMP_DIR/plain"
"$KEYCRYPT_BIN" encrypt --master "$TMP_DIR/master.key" \
  --in "$TMP_DIR/plain" --out "$TMP_DIR/enc"
"$KEYCRYPT_BIN" decrypt --master "$TMP_DIR/master.key" \
  --in "$TMP_DIR/enc" --out "$TMP_DIR/dec"
cmp "$TMP_DIR/plain" "$TMP_DIR/dec"

cp "$TMP_DIR/enc" "$TMP_DIR/tamper"
sz=$(wc -c < "$TMP_DIR/tamper")
pos=$((sz - 20))
printf '\xff' | dd of="$TMP_DIR/tamper" bs=1 seek="$pos" count=1 conv=notrunc 2>/dev/null
if "$KEYCRYPT_BIN" decrypt --master "$TMP_DIR/master.key" \
     --in "$TMP_DIR/tamper" --out "$TMP_DIR/bad" 2>/dev/null; then
  echo "expected decrypt failure on tampered ciphertext" >&2
  exit 1
fi

if [ -x "$OPENSSL_BIN" ]; then
  "$OPENSSL_BIN" rand -hex 32 > "$TMP_DIR/other.key"
else
  head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' > "$TMP_DIR/other.key"
fi
chmod 0600 "$TMP_DIR/other.key"
if "$KEYCRYPT_BIN" decrypt --master "$TMP_DIR/other.key" \
     --in "$TMP_DIR/enc" --out "$TMP_DIR/wrong" 2>/dev/null; then
  echo "expected decrypt failure with wrong key" >&2
  exit 1
fi

echo "test_keycrypt: OK"
