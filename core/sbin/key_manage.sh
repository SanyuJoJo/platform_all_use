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
ALG="$(json_get_def "$IN" '.params.algorithm' 'SM2')"
KEY_REF="$(json_get_def "$IN" '.params.key_ref' '')"
case "$ACTION" in
  generate|import|export|delete) ;;
  *) pki_fail "$OUT" "INVALID_PARAM" "invalid action" "$(jq -n --arg v "$ACTION" '{field:"action",value:$v}')" 2 ;;
esac
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/key-manage.XXXXXX")"
export TMP_DIR
trap pki_cleanup_tmp EXIT HUP INT TERM
KEY_ID=""
ENCRYPTED_PATH=""
EXPORT_PATH=""
STATE="ACTIVE"
case "$ACTION" in
  generate)
    pki_ext_check_algorithm "key.manage" "$ALG" || pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "algorithm not allowed" "$(jq -n --arg a "$ALG" '{algorithm:$a}')" 3
    KEY_REF="$(pki_new_key_ref)"
    KEY_ID="$(pki_new_id key)"
    TMP_KEY="$TMP_DIR/key.pem"
    pki_ext_gen_key "$ALG" "$TMP_KEY" "$IN" || pki_fail "$OUT" "INVALID_PARAM" "key generate failed" '{}' 2
    ENCRYPTED_PATH="$(pki_encrypt_key "$TMP_KEY" "$KEY_REF")"
    ;;
  import)
    KEY_PATH="$(json_get "$IN" '.params.key_path')"
    pki_require_param "$IN" '.params.key_path' 'key_path' "$OUT" 2
    ABS_KEY="$(pki_ext_require_under_core "$KEY_PATH" "key_path" "$OUT")"
    [ -f "$ABS_KEY" ] || pki_fail "$OUT" "KEY_NOT_FOUND" "key not found" '{}' 6
    "$OPENSSL_BIN" pkey -in "$ABS_KEY" -noout 2>/dev/null || pki_fail "$OUT" "CERT_PARSE_FAILED" "invalid key" '{}' 8
    KEY_REF="$(pki_new_key_ref)"
    KEY_ID="$(pki_new_id key)"
    ENCRYPTED_PATH="$(pki_encrypt_key "$ABS_KEY" "$KEY_REF")"
    ;;
  export)
    [ -n "$KEY_REF" ] || pki_fail "$OUT" "INVALID_PARAM" "key_ref required" '{}' 2
    ALLOW="$(json_get_def "$IN" '.params.allow_plain_export' 'false')"
    [ "$ALLOW" = "true" ] || pki_fail "$OUT" "PERMISSION_DENIED" "plain export disabled" '{}' 5
    EXPORT_PATH="$(json_get_def "$IN" '.params.export_path' "data/export/${KEY_REF}.pem")"
    ABS_EXPORT="$(pki_ext_require_under_core "$EXPORT_PATH" "export_path" "$OUT")"
    pki_mkdir_for "$ABS_EXPORT"
    pki_decrypt_key_to_tmp "$KEY_REF" "$TMP_DIR/key.out" || pki_fail "$OUT" "KEY_NOT_FOUND" "key not found" '{}' 6
    cp "$TMP_DIR/key.out" "$ABS_EXPORT"
    chmod 0600 "$ABS_EXPORT"
    ENCRYPTED_PATH="$CORE_ROOT/data/keys/${KEY_REF}.key.enc"
    ;;
  delete)
    [ -n "$KEY_REF" ] || pki_fail "$OUT" "INVALID_PARAM" "key_ref required" '{}' 2
    ENC="$CORE_ROOT/data/keys/${KEY_REF}.key.enc"
    [ -f "$ENC" ] || pki_fail "$OUT" "KEY_NOT_FOUND" "key not found" '{}' 6
    mv "$ENC" "$ENC.deleted"
    chmod 0600 "$ENC.deleted" 2>/dev/null || true
    STATE="DELETED"
    ENCRYPTED_PATH="$ENC.deleted"
    ;;
esac
DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"
DATA="$(jq -n \
  --arg key_ref "$KEY_REF" \
  --arg key_id "$KEY_ID" \
  --arg alg "$ALG" \
  --arg state "$STATE" \
  --arg enc "$ENCRYPTED_PATH" \
  --arg exp "$EXPORT_PATH" \
  '{key_ref:$key_ref,key_id:$key_id,algorithm:$alg,state:$state,encrypted_path:$enc,export_path:(if $exp=="" then null else $exp end)}')"
AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" --arg params_digest "$(pki_calc_params_digest "$IN")" --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"
pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
