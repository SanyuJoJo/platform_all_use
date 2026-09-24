#!/bin/sh
# =============================================================================
# core/sbin/dispatch.sh
# 统一调度入口：dispatch.sh --op <operation_id> --in <json> --out <json>
#
# 约定（C-04）：
#   - stdout：不输出内容（结果写 --out 文件）
#   - stderr：只输出 JSON Lines 日志
#   - 退出码：0 成功，非 0 失败
#   - 平台后端必须以 argv 参数列表调用，禁止 shell 字符串拼接
#
# 退出码与 core 错误码映射（C-04 / C-05）：
#   0  OK
#   2  INVALID_PARAM
#   3  ALGORITHM_NOT_ALLOWED
#   4  PATH_NOT_ALLOWED
#   5  PERMISSION_DENIED
#   6  KEY_NOT_FOUND
#   7  CERT_NOT_FOUND
#   8  CERT_PARSE_FAILED
#   9  CORE_EXEC_FAILED
#   10 CORE_TIMEOUT
#   11 CORE_JSON_INVALID
#   12 VERSION_UNSUPPORTED
#   99 INTERNAL_ERROR
# =============================================================================

set -eu

# -----------------------------------------------------------------------------
# 0. 基础路径与临时目录
# -----------------------------------------------------------------------------
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export CORE_ROOT

TMP="$(mktemp -d "${TMPDIR:-/tmp}/dispatch.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT HUP INT TERM

# -----------------------------------------------------------------------------
# 1. 加载公共库
# -----------------------------------------------------------------------------
. "$CORE_ROOT/sbin/lib/common.sh"
. "$CORE_ROOT/sbin/lib/audit.sh"

# -----------------------------------------------------------------------------
# 2. 参数解析
# -----------------------------------------------------------------------------
OP=""
IN=""
OUT=""

usage() {
  cat >&2 <<'EOF'
Usage: dispatch.sh --op <operation_id> --in <json> --out <json>

Options:
  --op   <operation_id>   operation identifier, e.g. ca.create
  --in   <json>           request JSON file path
  --out  <json>           response JSON file path
  -h, --help              show this help
EOF
}

# JSON 字符串转义：\ " \t \r \n
json_escape() {
  printf '%s' "${1:-}" | sed \
    -e 's/\\/\\\\/g' \
    -e 's/"/\\"/g' \
    -e 's/\t/\\t/g' \
    -e 's/\r/\\r/g' \
    -e ':a;N;$!ba;s/\n/\\n/g'
}

# -----------------------------------------------------------------------------
# 3. 统一错误输出
#    - out 非空：写 --out JSON 文件
#    - out 为空：写 stderr
#    - 无论哪种，都退出指定 exit_code
# -----------------------------------------------------------------------------
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

# -----------------------------------------------------------------------------
# 4. 参数解析循环
# -----------------------------------------------------------------------------
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

# 参数必填性校验
[ -n "$OP" ]  || emit_error "$OUT" "INVALID_PARAM" "missing --op"  2 "" "" ""
[ -n "$IN" ]  || emit_error "$OUT" "INVALID_PARAM" "missing --in"  2 "$OP" "" ""
[ -n "$OUT" ] || emit_error ""    "INVALID_PARAM" "missing --out" 2 "$OP" "" ""
[ -f "$IN" ]  || emit_error "$OUT" "INVALID_PARAM" "input json not found: $IN" 2 "$OP" "" ""

# -----------------------------------------------------------------------------
# 5. 提取 request_id / task_id（用于错误响应回填）
# -----------------------------------------------------------------------------
REQ_ID=""
TASK_ID=""
if command -v jq >/dev/null 2>&1; then
  REQ_ID="$(jq -r '.request_id // empty' "$IN" 2>/dev/null || true)"
  TASK_ID="$(jq -r '.task_id // empty' "$IN" 2>/dev/null || true)"
fi

# -----------------------------------------------------------------------------
# 6. 铜锁版本校验
#    失败返回 VERSION_UNSUPPORTED，退出码 12
# -----------------------------------------------------------------------------
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

# -----------------------------------------------------------------------------
# 7. operation_id 路由表
#    P0：ca.create / ca.intermediate.create / csr.create / cert.sign /
#        dual_cert.create / crl.create
#    P1：cert.convert / cert.parse / key.manage / pqc.cert.create
#    P2：chain.verify / batch.execute / ssl.config.generate / crypto.service
# -----------------------------------------------------------------------------
case "$OP" in
  # -------- P0 --------
  ca.create)                SCRIPT="$CORE_ROOT/sbin/ca_create.sh" ;;
  ca.intermediate.create)   SCRIPT="$CORE_ROOT/sbin/ca_intermediate.sh" ;;
  csr.create)               SCRIPT="$CORE_ROOT/sbin/csr_create.sh" ;;
  cert.sign)                SCRIPT="$CORE_ROOT/sbin/cert_sign.sh" ;;
  dual_cert.create)         SCRIPT="$CORE_ROOT/sbin/dual_cert_create.sh" ;;
  crl.create)               SCRIPT="$CORE_ROOT/sbin/crl_create.sh" ;;

  # -------- P1 --------
  cert.convert)             SCRIPT="$CORE_ROOT/sbin/cert_convert.sh" ;;
  cert.parse)               SCRIPT="$CORE_ROOT/sbin/cert_parse.sh" ;;
  key.manage)               SCRIPT="$CORE_ROOT/sbin/key_manage.sh" ;;
  pqc.cert.create)          SCRIPT="$CORE_ROOT/sbin/pqc_cert_create.sh" ;;

  # -------- P2 --------
  chain.verify)             SCRIPT="$CORE_ROOT/sbin/chain_verify.sh" ;;
  batch.execute)            SCRIPT="$CORE_ROOT/sbin/batch_execute.sh" ;;
  ssl.config.generate)      SCRIPT="$CORE_ROOT/sbin/ssl_config_generate.sh" ;;
  crypto.service)           SCRIPT="$CORE_ROOT/sbin/crypto_service.sh" ;;

  # -------- 未知 operation --------
  *)
    emit_error "$OUT" "INVALID_PARAM" "unknown operation_id: $OP" 2 "$OP" "$REQ_ID" "$TASK_ID"
    ;;
esac

# -----------------------------------------------------------------------------
# 8. Handler 存在性与可执行性校验
# -----------------------------------------------------------------------------
[ -f "$SCRIPT" ] || emit_error "$OUT" "INTERNAL_ERROR" "handler script not found: $SCRIPT" 99 "$OP" "$REQ_ID" "$TASK_ID"
[ -x "$SCRIPT" ] || chmod +x "$SCRIPT" 2>/dev/null || true
[ -x "$SCRIPT" ] || emit_error "$OUT" "INTERNAL_ERROR" "handler script not executable: $SCRIPT" 99 "$OP" "$REQ_ID" "$TASK_ID"

# -----------------------------------------------------------------------------
# 9. 记录路由日志（JSON Lines 到 stderr）
# -----------------------------------------------------------------------------
core_log "INFO" "dispatch.route" "route operation_id=$OP" "$REQ_ID"

# -----------------------------------------------------------------------------
# 10. 执行 Handler
#     Handler 必须：
#       - 从 --in 读请求 JSON
#       - 写 --out 响应 JSON
#       - stderr 输出 JSON Lines
#       - 退出码与 C-04/C-05 一致
# -----------------------------------------------------------------------------
exec "$SCRIPT" --op "$OP" --in "$IN" --out "$OUT"
