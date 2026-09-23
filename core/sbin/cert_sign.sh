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
CA_ID="$(json_get "$IN" '.params.ca_id')"
CA_KEY="$(json_get "$IN" '.params.ca_key_ref')"
CSR_ID="$(json_get "$IN" '.params.csr_id')"
CERT_TYPE="$(json_get_def "$IN" '.params.cert_type' 'server')"
ALG="$(json_get_def "$IN" '.params.algorithm' 'SM2')"

if ! pki_check_algorithm "cert.sign" "$ALG"; then
  DETAIL="$(jq -n --arg a "$ALG" '{algorithm:$a}')"
  pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "algorithm not allowed" "$DETAIL" 3
fi
case "$CERT_TYPE" in
  server|client) ;;
  *)
    DETAIL="$(jq -n --arg t "$CERT_TYPE" '{field:"cert_type",value:$t}')"
    pki_fail "$OUT" "INVALID_PARAM" "invalid cert_type" "$DETAIL" 2
    ;;
esac
pki_require_param "$IN" '.params.csr_id' 'csr_id' "$OUT" 2
pki_require_param "$IN" '.params.ca_id' 'ca_id' "$OUT" 2
pki_require_param "$IN" '.params.ca_key_ref' 'ca_key_ref' "$OUT" 2

CERT_ID="$(pki_new_id cert)"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/cert-sign.XXXXXX")"
TMP_CA_KEY="$TMP_DIR/ca.key"
export TMP_DIR TMP_CA_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

if ! pki_decrypt_key_to_tmp "$CA_KEY" "$TMP_CA_KEY"; then
  pki_fail "$OUT" "KEY_NOT_FOUND" "ca key not found" '{}' 6
fi

CA_CERT="$CORE_ROOT/data/ca/${CA_ID}.pem"
CSR="$CORE_ROOT/data/csr/${CSR_ID}.csr"
[ -f "$CA_CERT" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "ca cert not found" '{}' 7
[ -f "$CSR" ]     || pki_fail "$OUT" "CERT_NOT_FOUND" "csr not found" '{}' 7

CERT_REL="data/certs/${CERT_ID}.pem"
CHAIN_REL="data/certs/${CERT_ID}-chain.pem"
CERT_ABS="$CORE_ROOT/$CERT_REL"
CHAIN_ABS="$CORE_ROOT/$CHAIN_REL"
pki_mkdir_for "$CERT_ABS"
DAYS="$(pki_days "$IN")"

case "$CERT_TYPE" in
  server) KU="digitalSignature,keyEncipherment"; EKU="serverAuth" ;;
  client) KU="digitalSignature";                  EKU="clientAuth" ;;
esac

EXT="$TMP_DIR/cert.ext"
{
  printf 'basicConstraints=CA:FALSE\n'
  printf 'keyUsage=critical,%s\n' "$KU"
  printf 'extendedKeyUsage=%s\n' "$EKU"
} > "$EXT"

DIGEST="$(pki_digest_for_cert "$CA_CERT")"

"$OPENSSL_BIN" x509 -req -in "$CSR" \
  -CA "$CA_CERT" -CAkey "$TMP_CA_KEY" -CAcreateserial \
  -out "$CERT_ABS" -days "$DAYS" "$DIGEST" -extfile "$EXT" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "openssl x509 sign failed" '{}' 9
  }

[ -f "$CERT_ABS" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "cert not generated" '{}' 9

cat "$CERT_ABS" "$CA_CERT" > "$CHAIN_ABS"
SERIAL="$("$OPENSSL_BIN" x509 -in "$CERT_ABS" -noout -serial 2>/dev/null | cut -d= -f2)"
DUR=$(( $(date +%s) - START ))

pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg cert_id "$CERT_ID" \
  --arg cert_path "$CERT_REL" \
  --arg chain_path "$CHAIN_REL" \
  --arg serial "$SERIAL" \
  '{cert_id:$cert_id,cert_path:$cert_path,chain_path:$chain_path,serial:$serial}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
