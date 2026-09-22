#!/bin/sh
# core/sbin/lib/key_store.sh
# 私钥 AES-256-GCM 加密存储。AEAD 由 keycrypt C 工具实现，不用 openssl enc。

CORE_ROOT="${CORE_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)}"

key_store_keycrypt_bin() {
  printf '%s\n' "${KEYCRYPT_BIN:-$CORE_ROOT/src/tool/bin/keycrypt}"
}

key_store_master_key_path() {
  printf '%s\n' "${MASTER_KEY_PATH:-$CORE_ROOT/data/keys/master.key}"
}

key_store_tongsuo_openssl() {
  printf '%s\n' "${TONSUO_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
}

key_store_init_master_key() {
  master_key_path="$(key_store_master_key_path)"
  if [ -f "$master_key_path" ]; then
    return 0
  fi
  master_dir="$(dirname "$master_key_path")"
  mkdir -p "$master_dir"
  chmod 0700 "$master_dir" 2>/dev/null || true
  umask 077
  openssl_bin="$(key_store_tongsuo_openssl)"
  [ -x "$openssl_bin" ] || { echo "key_store: openssl not found: $openssl_bin" >&2; return 9; }
  "$openssl_bin" rand -hex 32 > "$master_key_path"
  chmod 0600 "$master_key_path"
}

key_store_require_keycrypt() {
  bin="$(key_store_keycrypt_bin)"
  if [ ! -x "$bin" ]; then
    echo "key_store: keycrypt not found or not executable: $bin" >&2
    return 9
  fi
  printf '%s\n' "$bin"
}

key_store_encrypt() {
  plain="$1"; out="$2"
  bin="$(key_store_require_keycrypt)" || return $?
  master="$(key_store_master_key_path)"
  key_store_init_master_key || return $?
  "$bin" encrypt --master "$master" --in "$plain" --out "$out" || return $?
  chmod 0600 "$out"
}

key_store_decrypt() {
  enc="$1"; out="$2"
  bin="$(key_store_require_keycrypt)" || return $?
  master="$(key_store_master_key_path)"
  [ -f "$master" ] || { echo "key_store: master key missing: $master" >&2; return 9; }
  "$bin" decrypt --master "$master" --in "$enc" --out "$out" || return $?
  chmod 0600 "$out"
}
