#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
# Switchyard desktop integration installer — https://github.com/alyraffauf/Switchyard
#
# Usage:  curl -fsSL <url> | bash -s -- [--install | --uninstall]
#
# This is a thin entrypoint that downloads and runs the Python installer.
# Arguments are forwarded directly to install-desktop-integration.py.

set -euo pipefail

if ! command -v python3 &>/dev/null; then
  echo "error: python3 is required but not installed" >&2
  exit 1
fi

PY_URL="https://raw.githubusercontent.com/alyraffauf/Switchyard/refs/heads/master/scripts/install-desktop-integration.py"

curl -fsSL "$PY_URL" | python3 - "$@"
