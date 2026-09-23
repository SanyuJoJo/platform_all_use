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
DIGEST="$(json_get_def "$IN" '.params.digest_algorithm' 'SM3')"

case "$DIGEST" in
  SM3|SHA256|SHA384|SHA512) ;;
  *)
    DETAIL="$(jq -n --arg d "$DIGEST" '{field:"digest_algorithm",value:$d}')"
    pki_fail "$OUT" "INVALID_PARAM" "invalid digest_algorithm" "$DETAIL" 2
    ;;
esac
pki_require_param "$IN" '.params.ca_id' 'ca_id' "$OUT" 2
pki_require_param "$IN" '.params.ca_key_ref' 'ca_key_ref' "$OUT" 2

CRL_ID="$(pki_new_id crl)"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/crl.XXXXXX")"
TMP_CA_KEY="$TMP_DIR/ca.key"
export TMP_DIR TMP_CA_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

if ! pki_decrypt_key_to_tmp "$CA_KEY" "$TMP_CA_KEY"; then
  pki_fail "$OUT" "KEY_NOT_FOUND" "ca key not found" '{}' 6
fi

CA_CERT="$CORE_ROOT/data/ca/${CA_ID}.pem"
[ -f "$CA_CERT" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "ca cert not found" '{}' 7

INDEX="$TMP_DIR/index.txt"
: > "$INDEX"
COUNT=0
if jq -e '.params.revoked_serials | type == "array"' "$IN" >/dev/null 2>&1; then
  jq -r '.params.revoked_serials[]?' "$IN" > "$TMP_DIR/serials.txt"
  while IFS= read -r serial; do
    [ -n "$serial" ] || continue
    printf 'R\t99991231235959Z\t260101000000Z\t%s\tunknown\t/CN=revoked\n' "$serial" >> "$INDEX"
    COUNT=$((COUNT + 1))
  done < "$TMP_DIR/serials.txt"
fi

CRL_REL="data/crl/${CRL_ID}.crl.pem"
CRL_ABS="$CORE_ROOT/$CRL_REL"
pki_mkdir_for "$CRL_ABS"

CONF="$TMP_DIR/openssl.cnf"
CRLNUM="$TMP_DIR/crlnumber"
echo "1000" > "$CRLNUM"

cat > "$CONF" <<EOF
[ ca ]
default_ca = CA_default

[ CA_default ]
database = $INDEX
crlnumber = $CRLNUM
default_md = $DIGEST
default_crl_days = 30
crl_extensions = crl_ext
unique_subject = no

[ crl_ext ]
authorityKeyIdentifier = keyid:always
EOF

"$OPENSSL_BIN" ca -gencrl -config "$CONF" \
  -cert "$CA_CERT" -keyfile "$TMP_CA_KEY" -out "$CRL_ABS" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "openssl ca -gencrl failed" '{}' 9
  }

[ -f "$CRL_ABS" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "crl not generated" '{}' 9

NEXT="$(json_get_def "$IN" '.params.next_update' '')"
DUR=$(( $(date +%s) - START ))

pki_write_audit "$REQ" "$TASK" "$OP" "$DIGEST" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg crl_id "$CRL_ID" \
  --arg crl_path "$CRL_REL" \
  --argjson revoked_count "$COUNT" \
  --arg next_update "$NEXT" \
  '{crl_id:$crl_id,crl_path:$crl_path,revoked_count:$revoked_count,next_update:$next_update}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
