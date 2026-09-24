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
ITEMS="$(jq -c '.params.items // []' "$IN")"
COUNT="$(printf '%s' "$ITEMS" | jq 'length')"
[ "$COUNT" -gt 0 ] || pki_fail "$OUT" "INVALID_PARAM" "items required" '{}' 2
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/batch.XXXXXX")"
export TMP_DIR
trap pki_cleanup_tmp EXIT HUP INT TERM
RESULTS="[]"
SUCCESS=0
FAILED=0
i=0
while [ "$i" -lt "$COUNT" ]; do
  ITEM="$(printf '%s' "$ITEMS" | jq -c ".[$i]")"
  OP_ID="$(printf '%s' "$ITEM" | jq -r '.operation_id // empty')"
  [ "$OP_ID" != "batch.execute" ] || pki_fail "$OUT" "INVALID_PARAM" "recursive batch.execute not allowed" '{}' 2
  SUB_REQ="$(printf '%s' "$ITEM" | jq -r '.request_id // empty')"
  [ -n "$SUB_REQ" ] || SUB_REQ="$(pki_new_id req)"
  SUB_IN="$TMP_DIR/sub-$i.json"
  SUB_OUT="$TMP_DIR/sub-$i.out.json"
  PARAMS="$(printf '%s' "$ITEM" | jq -c '.params // {}')"
  jq -n --arg op "$OP_ID" --arg req "$SUB_REQ" --argjson params "$PARAMS" \
    '{schema_version:"1.0",operation_id:$op,request_id:$req,actor:{type:"platform-backend",id:"batch"},params:$params}' > "$SUB_IN"
  set +e
  "$CORE_ROOT/sbin/dispatch.sh" --op "$OP_ID" --in "$SUB_IN" --out "$SUB_OUT" >/dev/null 2>&1
  RC=$?
  set -e
  if [ "$RC" -eq 0 ]; then SUCCESS=$((SUCCESS+1)); else FAILED=$((FAILED+1)); fi
  CODE="$(jq -r '.code // "UNKNOWN"' "$SUB_OUT" 2>/dev/null || echo UNKNOWN)"
  RESP="$(cat "$SUB_OUT" 2>/dev/null || echo '{}')"
  RESULTS="$(printf '%s' "$RESULTS" | jq --arg op "$OP_ID" --arg req "$SUB_REQ" --arg code "$CODE" --argjson rc "$RC" --argjson resp "$RESP" \
    '. + [{operation_id:$op,request_id:$req,exit_code:$rc,code:$code,response:$resp}]')"
  i=$((i+1))
done
DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "" "SUCCESS" "$DUR" "" "$IN"
DATA="$(jq -n --argjson total "$COUNT" --argjson success "$SUCCESS" --argjson failed "$FAILED" --argjson results "$RESULTS" \
  '{total:$total,success_count:$success,failed_count:$failed,results:$results}')"
AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" --arg params_digest "$(pki_calc_params_digest "$IN")" --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"
pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
