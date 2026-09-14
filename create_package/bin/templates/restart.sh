#!/usr/bin/env bash
set -euo pipefail
INSTALL_DIR="@INSTALL_DIR@"
"$INSTALL_DIR/bin/stop.sh" || true
sleep 1
"$INSTALL_DIR/bin/start.sh"
