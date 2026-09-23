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
PARENT_CA="$(json_get "$IN" '.params.parent_ca_id')"
PARENT_KEY="$(json_get "$IN" '.params.parent_key_ref')"

if ! pki_check_algorithm "ca.intermediate.create" "$ALG"; then
  DETAIL="$(jq -n --arg a "$ALG" '{algorithm:$a}')"
  pki_fail "$OUT" "ALGORITHM_NOT_ALLOWED" "algorithm not allowed" "$DETAIL" 3
fi

# 基础参数校验
pki_require_param "$IN" '.params.parent_ca_id' 'parent_ca_id' "$OUT" 2
pki_require_param "$IN" '.params.parent_key_ref' 'parent_key_ref' "$OUT" 2
pki_require_param "$IN" '.params.subject.CN' 'subject.CN' "$OUT" 2

# 先校验父 CA 证书存在
PARENT_CERT="$CORE_ROOT/data/ca/${PARENT_CA}.pem"
if [ ! -f "$PARENT_CERT" ]; then
  DETAIL="$(jq -n --arg p "$PARENT_CA" '{ca_id:$p}')"
  pki_fail "$OUT" "CERT_NOT_FOUND" "parent ca not found" "$DETAIL" 7
fi

# 再校验父私钥存在且可解密
TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/ca-int.XXXXXX")"
TMP_KEY="$TMP_DIR/key.pem"
TMP_PARENT_KEY="$TMP_DIR/parent.key"
export TMP_DIR TMP_KEY TMP_PARENT_KEY
trap pki_cleanup_tmp EXIT HUP INT TERM

if ! pki_decrypt_key_to_tmp "$PARENT_KEY" "$TMP_PARENT_KEY"; then
  DETAIL="$(jq -n --arg k "$PARENT_KEY" '{key_ref:$k}')"
  pki_fail "$OUT" "KEY_NOT_FOUND" "parent key not found" "$DETAIL" 6
fi

# 生成子 CA 私钥
CA_ID="$(pki_new_id ca)"
KEY_REF="$(pki_new_key_ref)"

if ! pki_gen_key "$ALG" "$TMP_KEY" "$IN"; then
  pki_fail "$OUT" "INVALID_PARAM" "key generate failed" "{}" 2
fi
pki_encrypt_key "$TMP_KEY" "$KEY_REF" >/dev/null

CSR="$TMP_DIR/int.csr"
CERT_REL="data/ca/${CA_ID}.pem"
CERT_ABS="$CORE_ROOT/$CERT_REL"
CHAIN_REL="data/ca/${CA_ID}-chain.pem"
CHAIN_ABS="$CORE_ROOT/$CHAIN_REL"
pki_mkdir_for "$CERT_ABS"

SUBJ="$(pki_subject_arg "$IN")"
DAYS="$(pki_days "$IN")"
PATHLEN="$(json_get_def "$IN" '.params.path_len' '0')"

"$OPENSSL_BIN" req -new -key "$TMP_KEY" -out "$CSR" -subj "$SUBJ" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "csr generate failed" '{}' 9
  }

EXT="$TMP_DIR/int.ext"
{
  printf 'basicConstraints=critical,CA:TRUE,pathlen:%s\n' "$PATHLEN"
  printf 'keyUsage=critical,keyCertSign,cRLSign\n'
} > "$EXT"

# 按父 CA 公钥算法选摘要：SM2 → -sm3，其余 → -sha256
DIGEST="$(pki_digest_for_cert "$PARENT_CERT")"

"$OPENSSL_BIN" x509 -req -in "$CSR" \
  -CA "$PARENT_CERT" -CAkey "$TMP_PARENT_KEY" -CAcreateserial \
  -out "$CERT_ABS" -days "$DAYS" "$DIGEST" -extfile "$EXT" \
  2>"$TMP_DIR/openssl.err" || {
    cat "$TMP_DIR/openssl.err" >&2
    pki_fail "$OUT" "CORE_EXEC_FAILED" "openssl x509 sign failed" '{}' 9
  }

[ -f "$CERT_ABS" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "cert not generated" '{}' 9

cat "$CERT_ABS" "$PARENT_CERT" > "$CHAIN_ABS"

SERIAL="$("$OPENSSL_BIN" x509 -in "$CERT_ABS" -noout -serial 2>/dev/null | cut -d= -f2)"
DUR=$(( $(date +%s) - START ))

pki_write_audit "$REQ" "$TASK" "$OP" "$ALG" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg ca_id "$CA_ID" \
  --arg cert_path "$CERT_REL" \
  --arg key_ref "$KEY_REF" \
  --arg chain_path "$CHAIN_REL" \
  --arg serial "$SERIAL" \
  '{ca_id:$ca_id,cert_path:$cert_path,key_ref:$key_ref,chain_path:$chain_path,serial:$serial}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
