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
LEAF="$(json_get "$IN" '.params.leaf_path')"
CHAIN="$(json_get_def "$IN" '.params.chain_path' '')"
CA="$(json_get_def "$IN" '.params.ca_path' '')"
pki_require_param "$IN" '.params.leaf_path' 'leaf_path' "$OUT" 2
ABS_LEAF="$(pki_ext_require_under_core "$LEAF" "leaf_path" "$OUT")"
[ -f "$ABS_LEAF" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "leaf cert not found" '{}' 7
ARGS="-CAfile"
if [ -n "$CA" ]; then
  ABS_CA="$(pki_ext_require_under_core "$CA" "ca_path" "$OUT")"
  [ -f "$ABS_CA" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "ca cert not found" '{}' 7
else
  ABS_CA="$ABS_LEAF"
fi
UNTRUSTED=""
if [ -n "$CHAIN" ]; then
  ABS_CHAIN="$(pki_ext_require_under_core "$CHAIN" "chain_path" "$OUT")"
  [ -f "$ABS_CHAIN" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "chain file not found" '{}' 7
  UNTRUSTED="$ABS_CHAIN"
fi
set +e
if [ -n "$UNTRUSTED" ]; then
  VERIFY_OUT="$("$OPENSSL_BIN" verify -CAfile "$ABS_CA" -untrusted "$UNTRUSTED" "$ABS_LEAF" 2>&1)"
else
  VERIFY_OUT="$("$OPENSSL_BIN" verify -CAfile "$ABS_CA" "$ABS_LEAF" 2>&1)"
fi
RC=$?
set -e
if [ "$RC" -eq 0 ]; then
  VALID=true; REASON=""
else
  VALID=false; REASON="$VERIFY_OUT"
fi
DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "" "SUCCESS" "$DUR" "" "$IN"
DATA="$(jq -n --argjson valid "$VALID" --arg reason "$REASON" --arg leaf "$LEAF" \
  '{valid:$valid,reason:$reason,verified_chain:$leaf}')"
AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" --arg params_digest "$(pki_calc_params_digest "$IN")" --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"
pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
