#!/bin/sh
# core/sbin/lib/common.sh
# 公共函数：配置读取、路径解析、版本校验集成。

_CORE_COMMON_LOADED="${_CORE_COMMON_LOADED:-0}"
if [ "$_CORE_COMMON_LOADED" = "1" ]; then
  return 0 2>/dev/null || true
fi
_CORE_COMMON_LOADED=1

CORE_ROOT="${CORE_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)}"
CORE_CONF="${CORE_CONF:-$CORE_ROOT/conf/core.conf}"

core_conf_get() {
  key="$1"; default="${2:-}"
  if [ -f "$CORE_CONF" ]; then
    value="$(sed -n "s/^[[:space:]]*$key[[:space:]]*:[[:space:]]*//p" "$CORE_CONF" \
      | head -n 1 | tr -d '"' | tr -d "'")"
    if [ -n "$value" ]; then
      printf '%s\n' "$value"
      return 0
    fi
  fi
  printf '%s\n' "$default"
}

core_resolve_path() {
  p="$1"
  case "$p" in
    /*) printf '%s\n' "$p" ;;
    *)  printf '%s\n' "$CORE_ROOT/$p" ;;
  esac
}

core_check_tongsuo_version() {
  openssl_bin="${TONSUO_OPENSSL_BIN:-$(core_conf_get "tongsuo_openssl_bin" "$CORE_ROOT/libs/bin/tongsuo/bin/openssl")}"
  openssl_bin="$(core_resolve_path "$openssl_bin")"
  min_version="${TONSUO_MIN_VERSION:-$(core_conf_get "tongsuo_min_version" "8.5.0")}"
  "$CORE_ROOT/sbin/lib/version_check.sh" \
    --openssl "$openssl_bin" \
    --min "$min_version" \
    --json
}
