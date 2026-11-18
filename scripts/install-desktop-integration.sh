#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
# Switchyard desktop integration installer — https://github.com/alyraffauf/Switchyard
#
# Usage:  curl -fsSL <url> | bash -s -- [--install | --uninstall]
#
# This is a thin entrypoint that downloads and runs the Python installer.
# Arguments are forwarded directly to browser-setup/main.py.

set -euo pipefail

if ! command -v python3 &>/dev/null; then
  echo "error: python3 is required but not installed" >&2
  exit 1
fi

PY_URL="https://raw.githubusercontent.com/alyraffauf/Switchyard/refs/heads/master/browser-setup/main.py"

installer=$(mktemp "${TMPDIR:-/tmp}/switchyard-browser-setup.XXXXXXXX")
trap 'rm -f -- "$installer"' EXIT
curl -fsSL "$PY_URL" > "$installer"

# With bash -s, stdin carries this shell script. Read answers from the terminal.
if [[ -z "${BASH_SOURCE[0]-}" ]] && { exec 3</dev/tty; } 2>/dev/null; then
  python3 "$installer" "$@" <&3
else
  python3 "$installer" "$@"
fi
