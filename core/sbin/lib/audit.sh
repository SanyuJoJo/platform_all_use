#!/bin/sh
# core/sbin/lib/audit.sh
# 结构化审计与 stderr 日志。所有字符串字段经 JSON 转义后输出。

_CORE_AUDIT_LOADED="${_CORE_AUDIT_LOADED:-0}"
if [ "$_CORE_AUDIT_LOADED" = "1" ]; then
  return 0 2>/dev/null || true
fi
_CORE_AUDIT_LOADED=1

# JSON 字符串转义：\ " \t \r \n
json_escape() {
  printf '%s' "${1:-}" | sed \
    -e 's/\\/\\\\/g' \
    -e 's/"/\\"/g' \
    -e 's/\t/\\t/g' \
    -e 's/\r/\\r/g' \
    -e ':a;N;$!ba;s/\n/\\n/g'
}

core_audit_log_file() {
  if [ -n "${CORE_AUDIT_LOG_FILE:-}" ]; then
    printf '%s\n' "$CORE_AUDIT_LOG_FILE"
  else
    printf '%s\n' "${CORE_ROOT:-.}/logs/audit.jsonl"
  fi
}

audit_write() {
  request_id="${1:-}"
  operation_id="${2:-}"
  result="${3:-}"
  duration_ms="${4:-0}"
  error_code="${5:-}"
  task_id="${6:-}"
  algorithm="${7:-}"
  params_digest="${8:-}"
  whitelist_version="${9:-}"

  log_file="$(core_audit_log_file)"
  log_dir="$(dirname "$log_file")"
  [ -d "$log_dir" ] || mkdir -p "$log_dir"

  audit_id="audit-$(date +%s)-$$"
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  case "$duration_ms" in
    ''|*[!0-9]*) duration_ms=0 ;;
  esac

  printf '{"audit_id":"%s","request_id":"%s","task_id":"%s","operation_id":"%s","actor_type":"%s","actor_id":"%s","algorithm":"%s","params_digest":"%s","whitelist_version":"%s","result":"%s","duration_ms":%s,"ts":"%s","error_code":"%s"}\n' \
    "$(json_escape "$audit_id")" \
    "$(json_escape "$request_id")" \
    "$(json_escape "$task_id")" \
    "$(json_escape "$operation_id")" \
    "$(json_escape "${ACTOR_TYPE:-platform-backend}")" \
    "$(json_escape "${ACTOR_ID:-}")" \
    "$(json_escape "$algorithm")" \
    "$(json_escape "$params_digest")" \
    "$(json_escape "$whitelist_version")" \
    "$(json_escape "$result")" \
    "$duration_ms" \
    "$(json_escape "$ts")" \
    "$(json_escape "$error_code")" \
    >> "$log_file"
}

core_log() {
  level="${1:-INFO}"; stage="${2:-}"; message="${3:-}"; req_id="${4:-}"
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '{"ts":"%s","level":"%s","request_id":"%s","stage":"%s","message":"%s"}\n' \
    "$(json_escape "$ts")" \
    "$(json_escape "$level")" \
    "$(json_escape "$req_id")" \
    "$(json_escape "$stage")" \
    "$(json_escape "$message")" >&2
}
