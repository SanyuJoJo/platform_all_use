#!/bin/sh
set -eu
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export CORE_ROOT
. "$CORE_ROOT/sbin/lib/common.sh"
. "$CORE_ROOT/sbin/lib/audit.sh"
. "$CORE_ROOT/sbin/lib/key_store.sh"
. "$CORE_ROOT/sbin/lib/pki_common.sh"
. "$CORE_ROOT/sbin/lib/pki_ext.sh"
OP=""; IN=""; OUT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --op)  OP="$2";  shift 2 ;;
    --in)  IN="$2";  shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    *)     pki_fail "$OUT" "INVALID_PARAM" "unknown argument: $1" '{}' 2 ;;
  esac
done
[ -n "$OP" ] && [ -n "$IN" ] && [ -n "$OUT" ] || pki_fail "$OUT" "INVALID_PARAM" "missing args" '{}' 2
[ -f "$IN" ] || pki_fail "$OUT" "INVALID_PARAM" "input json not found" '{}' 2
pki_require_jq
START=$(date +%s)
REQ="$(json_get "$IN" '.request_id')"
TASK="$(json_get_def "$IN" '.task_id' '')"
ACTION="$(json_get "$IN" '.params.action')"
ALG="$(json_get "$IN" '.params.algorithm')"
KEY_REF="$(json_get "$IN" '.params.key_ref')"
INPUT="$(json_get "$IN" '.params.input_path')"
OUTPUT="$(json_get_def "$IN" '.params.output_path' '')"
case "$ACTION" in sign|verify|encrypt|decrypt|hmac|kdf) ;; *) pki_fail "$OUT" "INVALID_PARAM" "invalid action" '{}' 2 ;; esac
pki_require_param "$IN" '.params.action' 'action' "$OUT" 2
pki_require_param "$IN" '.params.algorithm' 'algorithm' "$OUT" 2
pki_require_param "$IN" '.params.key_ref' 'key_ref' "$OUT" 2
pki_require_param "$IN" '.params.input_path' 'input_path' "$OUT" 2
ABS_IN="$(pki_ext_require_under_core "$INPUT" "input_path" "$OUT")"
[ -f "$ABS_IN" ] || pki_fail "$OUT" "INVALID_PARAM" "input file not found" '{}' 2
if [ -z "$OUTPUT" ]; then
  OUTPUT="data/crypto/$(pki_new_id crypto).out"
fi
ABS_OUT="$(pki_ext_require_under_core "$OUTPUT" "output_path" "$OUT")"
pki_mkdir_for "$ABS_OUT"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/crypto.XXXXXX")"
TMP_KEY="$TMP_DIR/key.pem"
export TMP_DIR TMP_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM
pki_decrypt_key_to_tmp "$KEY_REF" "$TMP_KEY" || pki_fail "$OUT" "KEY_NOT_FOUND" "key not found" '{}' 6
case "$ACTION" in
  sign)
    "$OPENSSL_BIN" dgst -sign "$TMP_KEY" -out "$ABS_OUT" "$ABS_IN" 2>/dev/null || pki_fail "$OUT" "CORE_EXEC_FAILED" "sign failed" '{}' 9
    ;;
  verify)
    PUB="$(json_get_def "$IN" '.params.public_key_path' '')"
    SIG="$(json_get_def "$IN" '.params.signature_path' '')"
    [ -n "$PUB" ] && [ -n "$SIG" ] || pki_fail "$OUT" "INVALID_PARAM" "verify requires public_key_path and signature_path" '{}' 2
    ABS_PUB="$(pki_ext_require_under_core "$PUB" "public_key_path" "$OUT")"
    ABS_SIG="$(pki_ext_require_under_core "$SIG" "signature_path" "$OUT")"
    "$OPENSSL_BIN" dgst -verify "$ABS_PUB" -signature "$ABS_SIG" "$ABS_IN" > "$ABS_OUT" 2>/dev/null || true
    ;;
  encrypt|decrypt)
    "$OPENSSL_BIN" pkeyutl -$ACTION -inkey "$TMP_KEY" -in "$ABS_IN" -out "$ABS_OUT" 2>/dev/null || pki_fail "$OUT" "CORE_EXEC_FAILED" "$ACTION failed" '{}' 9
    ;;
  hmac)
    SECRET="$(json_get_def "$IN" '.params.secret_file' '')"
    [ -n "$SECRET" ] || pki_fail "$OUT" "INVALID_PARAM" "hmac requires secret_file" '{}' 2
    ABS_SECRET="$(pki_ext_require_under_core "$SECRET" "secret_file" "$OUT")"
    "$OPENSSL_BIN" dgst -hmac "$(cat "$ABS_SECRET")" -out "$ABS_OUT" "$ABS_IN" 2>/dev/null || pki_fail "$OUT" "CORE_EXEC_FAILED" "hmac failed" '{}' 9
    ;;
  kdf)
    pki_fail "$OUT" "INTERNAL_ERROR" "kdf not implemented in P2 prototype" '{}' 99
    ;;
esac
chmod 0600 "$ABS_OUT" 2>/dev/null || true
DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"
DATA="$(jq -n --arg output "$OUTPUT" --arg action "$ACTION" --arg alg "$ALG" \
  '{output_path:$output,action:$action,algorithm:$alg}')"
AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" --arg params_digest "$(pki_calc_params_digest "$IN")" --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"
pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
