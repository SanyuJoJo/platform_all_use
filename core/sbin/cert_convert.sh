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
SRC_FMT="$(json_get_def "$IN" '.params.source_format' 'PEM')"
DST_FMT="$(json_get_def "$IN" '.params.target_format' 'PEM')"
SRC_PATH="$(json_get "$IN" '.params.source_path')"
DST_PATH="$(json_get_def "$IN" '.params.target_path' '')"
PASS_FILE="$(json_get_def "$IN" '.params.password_file' '')"
KEY_REF="$(json_get_def "$IN" '.params.key_ref' '')"

pki_require_param "$IN" '.params.source_format' 'source_format' "$OUT" 2
pki_require_param "$IN" '.params.target_format' 'target_format' "$OUT" 2
pki_require_param "$IN" '.params.source_path' 'source_path' "$OUT" 2

case "$SRC_FMT" in PEM|DER|PKCS12) ;; *) pki_fail "$OUT" "INVALID_PARAM" "invalid source_format" "$(jq -n --arg v "$SRC_FMT" '{field:"source_format",value:$v}')" 2 ;; esac
case "$DST_FMT" in PEM|DER|PKCS12) ;; *) pki_fail "$OUT" "INVALID_PARAM" "invalid target_format" "$(jq -n --arg v "$DST_FMT" '{field:"target_format",value:$v}')" 2 ;; esac

ABS_SRC="$(pki_ext_require_under_core "$SRC_PATH" "source_path" "$OUT")"
[ -f "$ABS_SRC" ] || pki_fail "$OUT" "INVALID_PARAM" "source file not found" "$(jq -n --arg p "$SRC_PATH" '{path:$p}')" 2

CONV_ID="$(pki_new_id convert)"
if [ -z "$DST_PATH" ]; then
  case "$DST_FMT" in
    PEM)    DST_REL="data/convert/${CONV_ID}.pem" ;;
    DER)    DST_REL="data/convert/${CONV_ID}.der" ;;
    PKCS12) DST_REL="data/convert/${CONV_ID}.p12" ;;
  esac
else
  DST_REL="$DST_PATH"
fi
ABS_DST="$(pki_ext_require_under_core "$DST_REL" "target_path" "$OUT")"
pki_mkdir_for "$ABS_DST"

TMP_DIR="$(mktemp -d "$CORE_ROOT/tmp/convert.XXXXXX")"
export TMP_DIR
trap pki_cleanup_tmp EXIT HUP INT TERM

KEY_OUT_REF=""
HAS_KEY="false"

# ---- 统一解析 password_file 为绝对路径（关键修复） ----
# pki_ext_require_password_file 返回绝对路径，这里接住，后续一律用该变量。
ABS_PASS_FILE=""
if [ -n "$PASS_FILE" ]; then
  ABS_PASS_FILE="$(pki_ext_require_password_file "$PASS_FILE" "$OUT")"
fi

if [ "$SRC_FMT" = "PKCS12" ]; then
  [ -n "$ABS_PASS_FILE" ] || pki_fail "$OUT" "INVALID_PARAM" "password_file required for PKCS12" '{}' 2

  "$OPENSSL_BIN" pkcs12 -in "$ABS_SRC" -clcerts -nokeys \
    -passin "file:$ABS_PASS_FILE" -out "$TMP_DIR/cert.pem" 2>"$TMP_DIR/pkcs12-in.err" || {
      cat "$TMP_DIR/pkcs12-in.err" >&2
      pki_fail "$OUT" "CERT_PARSE_FAILED" "pkcs12 cert extract failed" '{}' 8
    }

  "$OPENSSL_BIN" pkcs12 -in "$ABS_SRC" -nocerts -nodes \
    -passin "file:$ABS_PASS_FILE" -out "$TMP_DIR/key.pem" 2>/dev/null || true

  if [ -s "$TMP_DIR/key.pem" ]; then
    KEY_OUT_REF="$(pki_new_key_ref)"
    pki_encrypt_key "$TMP_DIR/key.pem" "$KEY_OUT_REF" >/dev/null
    HAS_KEY="true"
  fi

  case "$DST_FMT" in
    PEM)
      cp "$TMP_DIR/cert.pem" "$ABS_DST"
      ;;
    DER)
      "$OPENSSL_BIN" x509 -in "$TMP_DIR/cert.pem" -outform DER -out "$ABS_DST" 2>"$TMP_DIR/der.err" || {
        cat "$TMP_DIR/der.err" >&2
        pki_fail "$OUT" "CORE_EXEC_FAILED" "convert to DER failed" '{}' 9
      }
      ;;
    PKCS12)
      [ -n "$ABS_PASS_FILE" ] || pki_fail "$OUT" "INVALID_PARAM" "password_file required for PKCS12 export" '{}' 2
      if [ -n "$KEY_OUT_REF" ]; then
        pki_decrypt_key_to_tmp "$KEY_OUT_REF" "$TMP_DIR/key.out" || \
          pki_fail "$OUT" "KEY_NOT_FOUND" "key decrypt failed" '{}' 6
        "$OPENSSL_BIN" pkcs12 -export \
          -inkey "$TMP_DIR/key.out" -in "$TMP_DIR/cert.pem" -out "$ABS_DST" \
          -passout "file:$ABS_PASS_FILE" 2>"$TMP_DIR/pkcs12-out.err" || {
            cat "$TMP_DIR/pkcs12-out.err" >&2
            pki_fail "$OUT" "CORE_EXEC_FAILED" "pkcs12 export failed" '{}' 9
          }
      else
        "$OPENSSL_BIN" pkcs12 -export -nokeys \
          -in "$TMP_DIR/cert.pem" -out "$ABS_DST" \
          -passout "file:$ABS_PASS_FILE" 2>"$TMP_DIR/pkcs12-out.err" || {
            cat "$TMP_DIR/pkcs12-out.err" >&2
            pki_fail "$OUT" "CORE_EXEC_FAILED" "pkcs12 export cert-only failed" '{}' 9
          }
      fi
      ;;
  esac

else
  case "$SRC_FMT:$DST_FMT" in
    PEM:PEM)
      cp "$ABS_SRC" "$ABS_DST"
      ;;
    PEM:DER)
      "$OPENSSL_BIN" x509 -in "$ABS_SRC" -outform DER -out "$ABS_DST" 2>"$TMP_DIR/der.err" || {
        cat "$TMP_DIR/der.err" >&2
        pki_fail "$OUT" "CERT_PARSE_FAILED" "PEM to DER failed" '{}' 8
      }
      ;;
    DER:PEM)
      "$OPENSSL_BIN" x509 -in "$ABS_SRC" -inform DER -out "$ABS_DST" 2>"$TMP_DIR/pem.err" || {
        cat "$TMP_DIR/pem.err" >&2
        pki_fail "$OUT" "CERT_PARSE_FAILED" "DER to PEM failed" '{}' 8
      }
      ;;
    DER:DER)
      cp "$ABS_SRC" "$ABS_DST"
      ;;
    *:PKCS12)
      [ -n "$ABS_PASS_FILE" ] || pki_fail "$OUT" "INVALID_PARAM" "password_file required for PKCS12 export" '{}' 2
      if [ -n "$KEY_REF" ]; then
        pki_decrypt_key_to_tmp "$KEY_REF" "$TMP_DIR/key.out" || \
          pki_fail "$OUT" "KEY_NOT_FOUND" "key not found" '{}' 6
        "$OPENSSL_BIN" pkcs12 -export \
          -inkey "$TMP_DIR/key.out" -in "$ABS_SRC" -out "$ABS_DST" \
          -passout "file:$ABS_PASS_FILE" 2>"$TMP_DIR/pkcs12-out.err" || {
            cat "$TMP_DIR/pkcs12-out.err" >&2
            pki_fail "$OUT" "CORE_EXEC_FAILED" "pkcs12 export failed" '{}' 9
          }
        HAS_KEY="true"
      else
        "$OPENSSL_BIN" pkcs12 -export -nokeys \
          -in "$ABS_SRC" -out "$ABS_DST" \
          -passout "file:$ABS_PASS_FILE" 2>"$TMP_DIR/pkcs12-out.err" || {
            cat "$TMP_DIR/pkcs12-out.err" >&2
            pki_fail "$OUT" "CORE_EXEC_FAILED" "pkcs12 export cert-only failed" '{}' 9
          }
      fi
      ;;
    *)
      pki_fail "$OUT" "INVALID_PARAM" "unsupported conversion" '{}' 2
      ;;
  esac
fi

[ -f "$ABS_DST" ] || pki_fail "$OUT" "CORE_EXEC_FAILED" "converted file not generated" '{}' 9
chmod 0600 "$ABS_DST" 2>/dev/null || true

DUR=$(( $(date +%s) - START ))
pki_write_audit "$REQ" "$TASK" "$OP" "$SRC_FMT" "SUCCESS" "$DUR" "" "$IN"

DATA="$(jq -n \
  --arg path "$DST_REL" \
  --arg fmt "$DST_FMT" \
  --arg key_ref "$KEY_OUT_REF" \
  --argjson has_key "$HAS_KEY" \
  '{converted_path:$path,format:$fmt,cert_count:1,has_key:$has_key,key_ref:(if $key_ref=="" then null else $key_ref end)}')"

AUDIT="$(jq -n \
  --arg audit_id "audit-$(date +%s)-$$" \
  --arg params_digest "$(pki_calc_params_digest "$IN")" \
  --argjson duration_ms "$DUR" \
  '{audit_id:$audit_id,params_digest:$params_digest,result:"SUCCESS",duration_ms:$duration_ms}')"

pki_success "$IN" "$OUT" "$OP" "$REQ" "$TASK" "$DATA" "$AUDIT"
exit 0
