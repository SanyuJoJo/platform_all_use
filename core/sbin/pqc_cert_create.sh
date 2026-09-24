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
ALG="$(json_get "$IN" '.params.algorithm')"
CERT_TYPE="$(json_get_def "$IN" '.params.cert_type' 'server')"
CA_ID="$(json_get_def "$IN" '.params.ca_id' '')"
CA_KEY="$(json_get_def "$IN" '.params.ca_key_ref' '')"

pki_ext_check_algorithm "pqc.cert.create" "$ALG" || pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" \
  "algorithm not allowed" "$(jq -n --arg a "$ALG" '{algorithm:$a}')" 3
pki_require_param "$IN" '.params.algorithm' 'algorithm' "$OUT" 2
pki_require_param "$IN" '.params.subject.CN' 'subject.CN' "$OUT" 2
pki_require_param "$IN" '.params.validity_days' 'validity_days' "$OUT" 2

# ML-KEM 只能用于加密/KEM，不允许签名证书或 CA
if [ "$ALG" = "ML-KEM" ]; then
  case "$CERT_TYPE" in
    enc|kem|client) ;;
    *)
      pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" \
        "ML-KEM cannot be used for signature/CA cert" \
        "$(jq -n --arg t "$CERT_TYPE" '{cert_type:$t}')" 3
      ;;
  esac
fi

CERT_ID="$(pki_new_id pqc)"
KEY_REF="$(pki_new_key_ref)"
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/pqc.XXXXXX")"
TMP_KEY="$TMP_DIR/key.pem"
export TMP_DIR TMP_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

# ---------- PQC 错误统一封装 ----------
pqc_fail_unsupported() {
  phase="$1"; err_file="${2:-}"
  ERR_MSG=""
  if [ -n "$err_file" ] && [ -f "$err_file" ]; then
    ERR_MSG="$(head -n1 "$err_file" 2>/dev/null | tr -d '\n' | head -c 200)"
  fi
  [ -n "$ERR_MSG" ] || ERR_MSG="PQC $phase not supported by current Tongsuo"
  pki_fail "$OUT" "VERSION_UNSUPPORTED" \
    "PQC $phase unsupported: $ALG ($ERR_MSG)" \
    "$(jq -n --arg a "$ALG" --arg p "$phase" '{algorithm:$a,phase:$p}')" 12
}
# --------------------------------------

# 1) keygen
if ! pki_ext_gen_key "$ALG" "$TMP_KEY" "$IN" 2>"$TMP_DIR/keygen.err"; then
  pqc_fail_unsupported "keygen" "$TMP_DIR/keygen.err"
fi

pki_encrypt_key "$TMP_KEY" "$KEY_REF" >/dev/null

CSR="$TMP_DIR/req.csr"
SUBJ="$(pki_subject_arg "$IN")"
DAYS="$(pki_days "$IN")"

# 2) CSR
if ! "$OPENSSL_BIN" req -new -key "$TMP_KEY" -out "$CSR" -subj "$SUBJ" 2>"$TMP_DIR/req.err"; then
  pqc_fail_unsupported "csr" "$TMP_DIR/req.err"
fi

CERT_REL="data/pqc/${CERT_ID}.pem"
CERT_ABS="$CORE_ROOT/$CERT_REL"
pki_mkdir_for "$CERT_ABS"

EXT="$TMP_DIR/pqc.ext"
if [ "$ALG" = "ML-KEM" ]; then
  printf 'basicConstraints=CA:FALSE\nkeyUsage=critical,keyEncipherment,keyAgreement\n' > "$EXT"
else
  printf 'basicConstraints=CA:FALSE\nkeyUsage=critical,digitalSignature\n' > "$EXT"
fi

# 3) 自签 或 由 CA 签发
if [ -n "$CA_ID" ] && [ -n "$CA_KEY" ]; then
  CA_CERT="$CORE_ROOT/data/ca/${CA_ID}.pem"
  TMP_CA_KEY="$TMP_DIR/ca.key"
  [ -f "$CA_CERT" ] || pki_fail "$OUT" "CERT_NOT_FOUND" "ca cert not found" \
    "$(jq -n --arg c "$CA_ID" '{ca_id:$c}')" 7
  pki_decrypt_key_to_tmp "$CA_KEY" "$TMP_CA_KEY" || pki_fail "$OUT" "KEY_NOT_FOUND" \
    "ca key not found" "$(jq -n --arg k "$CA_KEY" '{key_ref:$k}')" 6
  DIGEST="$(pki_digest_for_cert "$CA_CERT")"

  # 关键修复：CA 签发失败统一 VERSION_UNSUPPORTED
  if ! "$OPENSSL_BIN" x509 -req -in "$CSR" \
       -CA "$CA_CERT" -CAkey "$TMP_CA_KEY" -CAcreateserial \
       -out "$CERT_ABS" -days "$DAYS" "$DIGEST" -extfile "$EXT" \
       2>"$TMP_DIR/x509sign.err"; then
    pqc_fail_unsupported "x509-sign" "$TMP_DIR/x509sign.err"
  fi
else
  # 关键修复：自签失败统一 VERSION_UNSUPPORTED
  if ! "$OPENSSL_BIN" req -new -x509 -key "$TMP_KEY" -out "$CERT_ABS" \
       -days "$DAYS" -subj "$SUBJ" -extfile "$EXT" \
       2>"$TMP_DIR/selfsign.err"; then
    pqc_fail_unsupported "self-sign" "$TMP_DIR/selfsign.err"
  fi
fi

[ -f "$CERT_ABS" ] || pqc_fail_unsupported "cert-emit"

SERIAL="$("$OPENSSL_BIN" x509 -in "$CERT_ABS" -noout -serial 2>/dev/null | cut -d= -f2)"
DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n --arg cert_path "$CERT_REL" --arg key_ref "$KEY_REF" --arg alg "$ALG" --arg serial "$SERIAL" \
  '{cert_path:$cert_path,key_ref:$key_ref,algorithm:$alg,serial:$serial}')"

AUDIT="$(jq -n --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
