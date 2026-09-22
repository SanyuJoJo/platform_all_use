#!/bin/sh
set -eu
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cat > "$TMP_DIR/openssl" <<'EOF'
#!/bin/sh
echo "Tongsuo 8.5.0"
EOF
chmod +x "$TMP_DIR/openssl"
"$CORE_ROOT/sbin/lib/version_check.sh" --openssl "$TMP_DIR/openssl" --min 8.5.0 --json
cat > "$TMP_DIR/openssl-low" <<'EOF'
#!/bin/sh
echo "Tongsuo 8.4.9"
EOF
chmod +x "$TMP_DIR/openssl-low"
if "$CORE_ROOT/sbin/lib/version_check.sh" --openssl "$TMP_DIR/openssl-low" --min 8.5.0 --json; then
  echo "expected VERSION_UNSUPPORTED" >&2
  exit 1
else
  code=$?
  [ "$code" -eq 12 ] || (echo "expected exit 12, got $code" >&2; exit 1)
fi
echo "test_version_check: OK"
