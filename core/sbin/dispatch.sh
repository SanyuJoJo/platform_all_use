#!/bin/sh
set -eu
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$CORE_ROOT/sbin/lib/common.sh"
OP=""
IN=""
OUT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --op) OP="${2:-}"; shift 2 ;;
    --in) IN="${2:-}"; shift 2 ;;
    --out) OUT="${2:-}"; shift 2 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done
[ -n "$OP" ] || exit 2
[ -n "$IN" ] || exit 2
[ -n "$OUT" ] || exit 2
# 铜锁版本校验，失败返回 VERSION_UNSUPPORTED，退出码 12
if ! VERSION_JSON=$(core_check_tongsuo_version); then
  printf '%s\n' "$VERSION_JSON" > "$OUT"
  exit 12
fi
# 后续由密码组接入 OperationRegistry 与具体 Handler
# 本分册只保证版本校验与统一入口骨架
printf '{"schema_version":"1.0","code":"INTERNAL_ERROR","message":"operation handler not implemented","operation_id":"%s","data":null,"error":{"code":"INTERNAL_ERROR","message":"operation handler not implemented","retryable":false}}\n' "$OP" > "$OUT"
exit 99
