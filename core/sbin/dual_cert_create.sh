#!/bin/sh
set -eu
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export CORE_ROOT
. "$CORE_ROOT/sbin/lib/common.sh"
. "$CORE_ROOT/sbin/lib/audit.sh"
. "$CORE_ROOT/sbin/lib/key_store.sh"
. "$CORE_ROOT/sbin/lib/pki_common.sh"

OP=""; IN=""; OUT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --op)  OP="$2";  shift 2 ;;
    --in)  IN="$2";  shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    *)     pki_fail "$OUT" "INVALID_PARAM" "unknown argument: $1" '{}' 2 ;;
  esac
done
[ -n "$OP" ] && [ -n "$IN" ] && [ -n "$OUT" ] \
  || pki_fail "$OUT" "INVALID_PARAM" "missing args" '{}' 2
[ -f "$IN" ] || pki_fail "$OUT" "INVALID_PARAM" "input json not found" '{}' 2
pki_require_jq

START=$(date +%s)
REQ="$(json_get "$IN" '.request_id')"
TASK="$(json_get_def "$IN" '.task_id' '')"
SIGN_ALG="$(json_get_def "$IN" '.params.sign_algorithm' 'SM2')"
ENC_ALG="$(json_get_def "$IN" '.params.enc_algorithm' 'SM2')"
TLCP="$(json_get_def "$IN" '.params.tlcp_profile' 'GB/T 38636-2020')"
CA_ID="$(json_get "$IN" '.params.ca_id')"
CA_KEY="$(json_get "$IN" '.params.ca_key_ref')"

# 双证严格：SM2 + SM2 + GB/T 38636-2020
if [ "$SIGN_ALG" != "SM2" ] || [ "$ENC_ALG" != "SM2" ] || [ "$TLCP" != "GB/T 38636-2020" ]; then
  DETAIL="$(jq -n --arg sa "$SIGN_ALG" --arg ea "$ENC_ALG" --arg t "$TLCP" \
    '{sign_algorithm:$sa,enc_algorithm:$ea,tlcp_profile:$t}')"
  pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "dual cert requires SM2+SM2+GB/T 38636-2020" "$DETAIL" 3
fi
pki_require_param "$IN" '.params.ca_id' 'ca_id' "$OUT" 2
pki_require_param "$IN" '.params.ca_key_ref' 'ca_key_ref' "$OUT" 2
pki_require_param "$IN" '.params.subject.CN' 'subject.CN' "$OUT" 2

TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/dual.XXXXXX")"
TMP_CA_KEY="$TMP_DIR/ca.key"
TMP_SIGN_KEY="$TMP_DIR/sign.key"
TMP_ENC_KEY="$TMP_DIR/enc.key"
export TMP_DIR TMP_CA_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

SIGN_KEY_REF="$(pki_new_key_ref)"
ENC_KEY_REF="$(pki_new_key_ref)"

if ! pki_gen_key "SM2" "$TMP_SIGN_KEY" "$IN"; then
  pki_fail "$OUT" "INVALID_PARAM" "sign key generate failed" '{}' 2
fi
if ! pki_gen_key "SM2" "$TMP_ENC_KEY" "$IN"; then
  pki_fail "$OUT" "INVALID_PARAM" "enc key generate failed" '{}' 2
fi

pki_encrypt_key "$TMP_SIGN_KEY" "$SIGN_KEY_REF" >/dev/null
pki_encrypt_key "$TMP_ENC_KEY"  "$ENC_KEY_REF"  >/dev/null

if ! pki_decrypt_key_to_tmp "$CA_KEY" "$TMP_CA_KEY"; then
  pki_fail "$OUT" "KEY_NOT_FOUND" "ca key not found" '{}' 6
fi

CA_CERT="$CORE_ROOT/data/ca/${CA_ID}.pem"
[ -f "$CA_CERT" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "ca cert not found" '{}' 7

SIGN_CSR="$TMP_DIR/sign.csr"
ENC_CSR="$TMP_DIR/enc.csr"
SUBJ="$(pki_subject_arg "$IN")"
DAYS="$(pki_days "$IN")"

"$OPENSSL_BIN" req -new -key "$TMP_SIGN_KEY" -out "$SIGN_CSR" -subj "$SUBJ" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "sign csr failed" '{}' 9
  }
"$OPENSSL_BIN" req -new -key "$TMP_ENC_KEY" -out "$ENC_CSR" -subj "$SUBJ" \
  2>>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "enc csr failed" '{}' 9
  }

SIGN_CERT_REL="data/dual/${SIGN_KEY_REF}-sign.pem"
ENC_CERT_REL="data/dual/${ENC_KEY_REF}-enc.pem"
CHAIN_REL="data/dual/${SIGN_KEY_REF}-chain.pem"
SIGN_CERT="$CORE_ROOT/$SIGN_CERT_REL"
ENC_CERT="$CORE_ROOT/$ENC_CERT_REL"
CHAIN="$CORE_ROOT/$CHAIN_REL"
pki_mkdir_for "$SIGN_CERT"

SIGN_EXT="$TMP_DIR/sign.ext"
ENC_EXT="$TMP_DIR/enc.ext"
{
  printf 'basicConstraints=CA:FALSE\n'
  printf 'keyUsage=critical,digitalSignature\n'
  printf 'extendedKeyUsage=clientAuth\n'
} > "$SIGN_EXT"
{
  printf 'basicConstraints=CA:FALSE\n'
  printf 'keyUsage=critical,keyEncipherment,keyAgreement\n'
  printf 'extendedKeyUsage=clientAuth\n'
} > "$ENC_EXT"

"$OPENSSL_BIN" x509 -req -in "$SIGN_CSR" \
  -CA "$CA_CERT" -CAkey "$TMP_CA_KEY" -CAcreateserial \
  -out "$SIGN_CERT" -days "$DAYS" -sm3 -extfile "$SIGN_EXT" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "sign cert failed" '{}' 9
  }

"$OPENSSL_BIN" x509 -req -in "$ENC_CSR" \
  -CA "$CA_CERT" -CAkey "$TMP_CA_KEY" -CAcreateserial \
  -out "$ENC_CERT" -days "$DAYS" -sm3 -extfile "$ENC_EXT" \
  2>>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "enc cert failed" '{}' 9
  }

[ -f "$SIGN_CERT" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "sign cert not generated" '{}' 9
[ -f "$ENC_CERT" ]  || pki_fail "$OUT" "CORE_EXEC_FAILED" "enc cert not generated" '{}' 9

cat "$SIGN_CERT" "$CA_CERT" > "$CHAIN"

DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "SM2" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg sign_cert_path "$SIGN_CERT_REL" \
  --arg enc_cert_path "$ENC_CERT_REL" \
  --arg sign_key_ref "$SIGN_KEY_REF" \
  --arg enc_key_ref "$ENC_KEY_REF" \
  --arg chain_path "$CHAIN_REL" \
  '{sign_cert_path:$sign_cert_path,enc_cert_path:$enc_cert_path,sign_key_ref:$sign_key_ref,enc_key_ref:$enc_key_ref,chain_path:$chain_path}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
