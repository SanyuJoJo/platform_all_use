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

if ! pki_check_algorithm "ca.create" "$ALG"; then
  DETAIL="$(jq -n --arg a "$ALG" '{algorithm:$a}')"
  pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "algorithm not allowed" "$DETAIL" 3
fi

pki_require_param "$IN" '.params.subject.CN' 'subject.CN' "$OUT" 2
pki_require_param "$IN" '.params.validity_days' 'validity_days' "$OUT" 2

CA_ID="$(pki_new_id ca)"
KEY_REF="$(pki_new_key_ref)"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/ca-create.XXXXXX")"
TMP_KEY="$TMP_DIR/key.pem"
export TMP_DIR TMP_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

if ! pki_gen_key "$ALG" "$TMP_KEY" "$IN"; then
  pki_fail "$OUT" "INVALID_PARAM" "key generate failed" "{}" 2
fi

pki_encrypt_key "$TMP_KEY" "$KEY_REF" >/dev/null

CERT_REL="data/ca/${CA_ID}.pem"
CERT_ABS="$CORE_ROOT/$CERT_REL"
pki_mkdir_for "$CERT_ABS"
SUBJ="$(pki_subject_arg "$IN")"
DAYS="$(pki_days "$IN")"

EXT="$TMP_DIR/ca.ext"
{
  printf 'basicConstraints=critical,CA:TRUE\n'
  printf 'keyUsage=critical,keyCertSign,cRLSign\n'
  printf 'subjectKeyIdentifier=hash\n'
} > "$EXT"

case "$ALG" in
  SM2) DIGEST="-sm3" ;;
  RSA|ECC) DIGEST="-sha256" ;;
  ML-DSA) DIGEST="" ;;
  *) pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "unsupported algorithm" '{}' 3 ;;
esac

if [ -n "$DIGEST" ]; then
  "$OPENSSL_BIN" req -new -x509 -key "$TMP_KEY" -out "$CERT_ABS" -days "$DAYS" \
    -subj "$SUBJ" "$DIGEST" -extensions v3_ca -config "$EXT" 2>"$TMP_DIR/openssl.err" || {
      # 部分铜锁版本不支持 -extensions，回退到 -addext
      "$OPENSSL_BIN" req -new -x509 -key "$TMP_KEY" -out "$CERT_ABS" -days "$DAYS" \
        -subj "$SUBJ" "$DIGEST" \
        -addext "basicConstraints=critical,CA:TRUE" \
        -addext "keyUsage=critical,keyCertSign,cRLSign" \
        2>>"$TMP_DIR/openssl.err" || {
          cat "$TMP_DIR/openssl.err" >&2
          pki_fail "$OUT" "CORE_EXEC_FAILED" "openssl req failed" '{}' 9
        }
    }
else
  "$OPENSSL_BIN" req -new -x509 -key "$TMP_KEY" -out "$CERT_ABS" -days "$DAYS" \
    -subj "$SUBJ" \
    -addext "basicConstraints=critical,CA:TRUE" \
    -addext "keyUsage=critical,keyCertSign,cRLSign" \
    2>"$TMP_DIR/openssl.err" || {
      cat "$TMP_DIR/openssl.err" >&2
      pki_fail "$OUT" "CORE_EXEC_FAILED" "openssl req failed" '{}' 9
    }
fi

[ -f "$CERT_ABS" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "cert not generated" '{}' 9

SERIAL="$("$OPENSSL_BIN" x509 -in "$CERT_ABS" -noout -serial 2>/dev/null | cut -d= -f2)"
DUR=$(( $(date +%s) - START ))

pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg ca_id "$CA_ID" \
  --arg cert_path "$CERT_REL" \
  --arg key_ref "$KEY_REF" \
  --arg serial "$SERIAL" \
  '{ca_id:$ca_id,cert_path:$cert_path,key_ref:$key_ref,serial:$serial}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
