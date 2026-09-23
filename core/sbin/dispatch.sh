#!/bin/sh
# core/sbin/dispatch.sh
# 统一调度入口：dispatch.sh --op <operation_id> --in <json> --out <json>

set -eu

CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export CORE_ROOT

TMP="$(mktemp -d "${TMPDIR:-/tmp}/dispatch.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT HUP INT TERM

. "$CORE_ROOT/sbin/lib/common.sh"
. "$CORE_ROOT/sbin/lib/audit.sh"

OP=""
IN=""
OUT=""

usage() {
  cat >&2 <<'EOF'
Usage: dispatch.sh --op <operation_id> --in <json> --out <json>
EOF
}

json_escape() {
  printf '%s' "${1:-}" | sed \
    -e 's/\\/\\\\/g' \
    -e 's/"/\\"/g' \
    -e 's/\t/\\t/g' \
    -e 's/\r/\\r/g' \
    -e ':a;N;$!ba;s/\n/\\n/g'
}

emit_error() {
  out="$1"; code="$2"; msg="$3"; exit_code="$4"
  op="${5:-}"; req="${6:-}"; task="${7:-}"

  req_json="null"; task_json="null"
  [ -n "$req" ]  && req_json="$(printf '"%s"' "$(json_escape "$req")")"
  [ -n "$task" ] && task_json="$(printf '"%s"' "$(json_escape "$task")")"

  json="$(printf '{"schema_version":"1.0","code":"%s","message":"%s","request_id":%s,"operation_id":"%s","task_id":%s,"data":null,"error":{"code":"%s","message":"%s","detail":{},"retryable":false},"audit":null}\n' \
    "$(json_escape "$code")" "$(json_escape "$msg")" "$req_json" \
    "$(json_escape "$op")" "$task_json" \
    "$(json_escape "$code")" "$(json_escape "$msg")")"

  if [ -z "$out" ]; then
    printf '%s' "$json" >&2
  else
    d="$(dirname "$out")"
    [ -d "$d" ] || mkdir -p "$d"
    printf '%s' "$json" > "$out"
  fi

  exit "$exit_code"
}

while [ $# -gt 0 ]; do
  case "$1" in
    --op)
      [ $# -ge 2 ] || { usage; exit 2; }
      OP="$2"; shift 2 ;;
    --in)
      [ $# -ge 2 ] || { usage; exit 2; }
      IN="$2"; shift 2 ;;
    --out)
      [ $# -ge 2 ] || { usage; exit 2; }
      OUT="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      emit_error "$OUT" "INVALID_PARAM" "unknown argument: $1" 2 "$OP" "" "" ;;
  esac
done

[ -n "$OP" ] || emit_error "$OUT" "INVALID_PARAM" "missing --op" 2 "" "" ""
[ -n "$IN" ] || emit_error "$OUT" "INVALID_PARAM" "missing --in" 2 "$OP" "" ""
[ -n "$OUT" ] || emit_error "" "INVALID_PARAM" "missing --out" 2 "$OP" "" ""
[ -f "$IN" ] || emit_error "$OUT" "INVALID_PARAM" "input json not found: $IN" 2 "$OP" "" ""

REQ_ID=""; TASK_ID=""
if command -v jq >/dev/null 2>&1; then
  REQ_ID="$(jq -r '.request_id // empty' "$IN" 2>/dev/null || true)"
  TASK_ID="$(jq -r '.task_id // empty' "$IN" 2>/dev/null || true)"
fi

# 铜锁版本校验，失败返回 VERSION_UNSUPPORTED，退出码 12
VERSION_JSON=""
if ! VERSION_JSON="$(core_check_tongsuo_version 2>"$TMP/vc.err")"; then
  VERSION_MSG="tongsuo version unsupported"
  if [ -n "$VERSION_JSON" ]; then
    if command -v jq >/dev/null 2>&1; then
      vmsg="$(printf '%s' "$VERSION_JSON" | jq -r '.message // empty' 2>/dev/null || true)"
      [ -n "$vmsg" ] && VERSION_MSG="$vmsg"
    fi
  elif [ -s "$TMP/vc.err" ]; then
    VERSION_MSG="$(head -n1 "$TMP/vc.err" | tr -d '\n' | head -c 200)"
  fi
  emit_error "$OUT" "VERSION_UNSUPPORTED" "$VERSION_MSG" 12 "$OP" "$REQ_ID" "$TASK_ID"
fi

case "$OP" in
  ca.create)                SCRIPT="$CORE_ROOT/sbin/ca_create.sh" ;;
  ca.intermediate.create)   SCRIPT="$CORE_ROOT/sbin/ca_intermediate.sh" ;;
  csr.create)               SCRIPT="$CORE_ROOT/sbin/csr_create.sh" ;;
  cert.sign)                SCRIPT="$CORE_ROOT/sbin/cert_sign.sh" ;;
  dual_cert.create)         SCRIPT="$CORE_ROOT/sbin/dual_cert_create.sh" ;;
  crl.create)               SCRIPT="$CORE_ROOT/sbin/crl_create.sh" ;;
  *)
    emit_error "$OUT" "INVALID_PARAM" "unknown operation_id: $OP" 2 "$OP" "$REQ_ID" "$TASK_ID"
    ;;
esac

[ -f "$SCRIPT" ] || emit_error "$OUT" "INTERNAL_ERROR" "handler script not found: $SCRIPT" 99 "$OP" "$REQ_ID" "$TASK_ID"
[ -x "$SCRIPT" ] || chmod +x "$SCRIPT" 2>/dev/null || true
[ -x "$SCRIPT" ] || emit_error "$OUT" "INTERNAL_ERROR" "handler script not executable: $SCRIPT" 99 "$OP" "$REQ_ID" "$TASK_ID"

core_log "INFO" "dispatch.route" "route operation_id=$OP" "$REQ_ID"

exec "$SCRIPT" --op "$OP" --in "$IN" --out "$OUT"
