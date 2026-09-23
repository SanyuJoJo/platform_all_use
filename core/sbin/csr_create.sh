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
ALG="$(json_get "$IN" '.params.algorithm')"

if ! pki_check_algorithm "csr.create" "$ALG"; then
  DETAIL="$(jq -n --arg a "$ALG" '{algorithm:$a}')"
  pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "algorithm not allowed" "$DETAIL" 3
fi
pki_require_param "$IN" '.params.subject.CN' 'subject.CN' "$OUT" 2

CSR_ID="$(pki_new_id csr)"
KEY_REF="$(pki_new_key_ref)"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/csr.XXXXXX")"
TMP_KEY="$TMP_DIR/key.pem"
export TMP_DIR TMP_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

if ! pki_gen_key "$ALG" "$TMP_KEY" "$IN"; then
  pki_fail "$OUT" "INVALID_PARAM" "key generate failed" "{}" 2
fi
pki_encrypt_key "$TMP_KEY" "$KEY_REF" >/dev/null

CSR_REL="data/csr/${CSR_ID}.csr"
CSR_ABS="$CORE_ROOT/$CSR_REL"
pki_mkdir_for "$CSR_ABS"
SUBJ="$(pki_subject_arg "$IN")"

# SAN 校验
SAN_JSON="$(jq -c '.params.san // []' "$IN")"
SAN_COUNT="$(printf '%s' "$SAN_JSON" | jq 'length')"
if [ "$SAN_COUNT" -gt 0 ]; then
  BAD_SAN="$(printf '%s' "$SAN_JSON" | jq -r '.[] | select(test("^[A-Za-z0-9.-]+$")|not)' | head -n1)"
  if [ -n "$BAD_SAN" ]; then
    DETAIL="$(jq -n --arg s "$BAD_SAN" '{san:$s}')"
    pki_fail "$OUT" "INVALID_PARAM" "invalid SAN format" "$DETAIL" 2
  fi
fi

# 写扩展配置文件
EXT="$TMP_DIR/csr.ext"
{
  printf '[ req ]\n'
  printf 'distinguished_name = req_dn\n'
  printf 'req_extensions = v3_req\n'
  printf 'prompt = no\n\n'
  printf '[ req_dn ]\n'
  jq -r '.params.subject // {} | to_entries | .[] | "\(.key) = \(.value)"' "$IN"
  printf '\n[ v3_req ]\n'
  printf 'basicConstraints = CA:FALSE\n'
  printf 'keyUsage = critical,digitalSignature,keyEncipherment\n'
  printf 'extendedKeyUsage = serverAuth,clientAuth\n'
  if [ "$SAN_COUNT" -gt 0 ]; then
    printf 'subjectAltName = '
    printf '%s' "$SAN_JSON" | jq -r 'map("DNS:" + .) | join(",")'
    printf '\n'
  fi
} > "$EXT"

"$OPENSSL_BIN" req -new -key "$TMP_KEY" -out "$CSR_ABS" \
  -config "$EXT" -extensions v3_req \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "csr generate failed" '{}' 9
  }

[ -f "$CSR_ABS" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "csr not generated" '{}' 9

PUB="$("$OPENSSL_BIN" req -in "$CSR_ABS" -noout -pubkey 2>/dev/null \
  | "$OPENSSL_BIN" pkey -pubin -text -noout 2>/dev/null | head -n1)"
[ -n "$PUB" ] || PUB="unknown"

DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg csr_id "$CSR_ID" \
  --arg csr_path "$CSR_REL" \
  --arg key_ref "$KEY_REF" \
  --arg pub "$PUB" \
  '{csr_id:$csr_id,csr_path:$csr_path,key_ref:$key_ref,public_key_info:{summary:$pub}}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
