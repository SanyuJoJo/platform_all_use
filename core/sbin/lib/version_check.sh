#!/bin/sh
set -eu

MIN_VERSION="${TONSUO_MIN_VERSION:-8.5.0}"
OPENSSL_BIN="${TONSUO_OPENSSL_BIN:-}"
OUT_JSON=0

usage() {
  cat >&2 <<'EOF'
Usage: version_check.sh [--openssl <path>] [--min <x.y.z>] [--json]
Env:
  TONSUO_OPENSSL_BIN    openssl binary path
  TONSUO_MIN_VERSION    minimum version, default 8.5.0
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --openssl) OPENSSL_BIN="${2:-}"; shift 2 ;;
    --min)     MIN_VERSION="${2:-}"; shift 2 ;;
    --json)    OUT_JSON=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
CORE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

if [ -z "$OPENSSL_BIN" ]; then
  OPENSSL_BIN="$CORE_ROOT/libs/bin/tongsuo/bin/openssl"
fi

json_escape() {
  printf '%s' "${1:-}" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

emit_ok() {
  version="$1"
  if [ "$OUT_JSON" -eq 1 ]; then
    printf '{"code":"OK","message":"tongsuo version supported","version":"%s","min_version":"%s","openssl_bin":"%s","exit_code":0}\n' \
      "$(json_escape "$version")" \
      "$(json_escape "$MIN_VERSION")" \
      "$(json_escape "$OPENSSL_BIN")"
  else
    printf 'OK: Tongsuo %s >= %s (%s)\n' "$version" "$MIN_VERSION" "$OPENSSL_BIN"
  fi
  exit 0
}

emit_fail() {
  code="$1"; message="$2"; version="${3:-}"; exit_code="${4:-12}"
  if [ "$OUT_JSON" -eq 1 ]; then
    printf '{"code":"%s","message":"%s","version":"%s","min_version":"%s","openssl_bin":"%s","exit_code":%s}\n' \
      "$(json_escape "$code")" \
      "$(json_escape "$message")" \
      "$(json_escape "$version")" \
      "$(json_escape "$MIN_VERSION")" \
      "$(json_escape "$OPENSSL_BIN")" \
      "$exit_code"
  else
    printf '%s: %s (version=%s, min=%s, bin=%s)\n' \
      "$code" "$message" "$version" "$MIN_VERSION" "$OPENSSL_BIN" >&2
  fi
  exit "$exit_code"
}

[ -x "$OPENSSL_BIN" ] || emit_fail "OPENSSL_BIN_NOT_FOUND" "openssl binary not found or not executable" "" 12

RAW="$("$OPENSSL_BIN" version 2>/dev/null || true)"
[ -n "$RAW" ] || emit_fail "VERSION_PARSE_FAILED" "cannot read openssl version" "" 12

# 优先匹配 "Tongsuo x.y.z"，其次任意 x.y.z
VERSION="$(printf '%s\n' "$RAW" \
  | sed -n 's/.*Tongsuo[^0-9]*\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' \
  | head -n 1)"
if [ -z "$VERSION" ]; then
  VERSION="$(printf '%s\n' "$RAW" \
    | sed -n 's/.*[^0-9]\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' \
    | head -n 1)"
fi
[ -n "$VERSION" ] || emit_fail "VERSION_PARSE_FAILED" "cannot parse openssl version: $RAW" "" 12

version_ge() {
  awk -v a="$1" -v b="$2" 'BEGIN {
    split(a, A, ".");
    split(b, B, ".");
    for (i = 1; i <= 3; i++) {
      ai = A[i] + 0; bi = B[i] + 0;
      if (ai > bi) exit 0;
      if (ai < bi) exit 1;
    }
    exit 0;
  }'
}

if version_ge "$VERSION" "$MIN_VERSION"; then
  emit_ok "$VERSION"
else
  emit_fail "VERSION_UNSUPPORTED" "tongsuo version too low: $RAW" "$VERSION" 12
fi
