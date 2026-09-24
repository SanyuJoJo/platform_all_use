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
CERT_PATH="$(json_get "$IN" '.params.cert_path')"

pki_require_param "$IN" '.params.cert_path' 'cert_path' "$OUT" 2

ABS_CERT="$(pki_ext_require_under_core "$CERT_PATH" "cert_path" "$OUT")"
[ -f "$ABS_CERT" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "cert not found" \
  "$(jq -n --arg p "$CERT_PATH" '{path:$p}')" 7

# ---------- 关键修复：先做一次严格解析，失败即返回 CERT_PARSE_FAILED ----------
if ! "$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout 2>/dev/null; then
  pki_fail "$OUT" "CERT_PARSE_FAILED" "cert parse failed" \
    "$(jq -n --arg p "$CERT_PATH" '{path:$p}')" 8
fi
# ---------------------------------------------------------------------------

SUBJECT="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/^subject=//')"
ISSUER="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/^issuer=//')"
SERIAL="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -serial 2>/dev/null | cut -d= -f2)"
NOT_BEFORE="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -startdate 2>/dev/null | cut -d= -f2)"
NOT_AFTER="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -enddate 2>/dev/null | cut -d= -f2)"
FP="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -fingerprint -sha256 2>/dev/null | cut -d= -f2)"
SAN="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -ext subjectAltName 2>/dev/null | tail -n +2 | tr '\n' ',' | sed 's/,$//')"
KU="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -ext keyUsage 2>/dev/null | tail -n +2 | tr '\n' ',' | sed 's/,$//')"
EKU="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -ext extendedKeyUsage 2>/dev/null | tail -n +2 | tr '\n' ',' | sed 's/,$//')"
PUBALG="$("$OPENSSL_BIN" x509 -in "$ABS_CERT" -noout -text 2>/dev/null | sed -n 's/.*Public Key Algorithm: //p' | head -n1)"

DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg subject "$SUBJECT" \
  --arg issuer "$ISSUER" \
  --arg serial "$SERIAL" \
  --arg not_before "$NOT_BEFORE" \
  --arg not_after "$NOT_AFTER" \
  --arg fp "$FP" \
  --arg san "$SAN" \
  --arg ku "$KU" \
  --arg eku "$EKU" \
  --arg pubalg "$PUBALG" \
  '{subject:$subject,issuer:$issuer,serial:$serial,not_before:$not_before,not_after:$not_after,fingerprint_sha256:$fp,san:$san,key_usage:$ku,eku:$eku,public_key_algorithm:$pubalg}')"

AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
