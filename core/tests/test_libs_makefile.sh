#!/bin/sh
set -eu
CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test -f "$CORE_ROOT/Makefile"
test -f "$CORE_ROOT/libs/Makefile"
test -f "$CORE_ROOT/src/tool/Makefile"
test -x "$CORE_ROOT/sbin/lib/version_check.sh"
test -f "$CORE_ROOT/conf/core.conf"
test -f "$CORE_ROOT/sbin/lib/common.sh"
test -f "$CORE_ROOT/sbin/lib/key_store.sh"
test -f "$CORE_ROOT/sbin/lib/audit.sh"
echo "test_libs_makefile: OK"
