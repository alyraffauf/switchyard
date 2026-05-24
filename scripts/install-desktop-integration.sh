#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
# Switchyard desktop integration installer — https://github.com/alyraffauf/Switchyard
#
# Installs the desktop integration that lets the browser extension list your installed browsers.
#
# For Flatpak browsers this grants --talk-name=org.freedesktop.Flatpak, which
# lets code inside the browser sandbox run arbitrary host commands via
# flatpak-spawn --host. This weakens the Flatpak sandbox for those browsers.
# Read and understand this before running it.
#
# Usage:  bash install-native-host.sh [--uninstall]

set -euo pipefail

# ── Find Switchyard ────────────────────────────────────────────────────────────

CMD=""
for id in io.github.alyraffauf.Switchyard io.github.alyraffauf.Switchyard.Devel; do
  if flatpak info "$id" &>/dev/null; then
    CMD="/usr/bin/flatpak run --command=switchyard $id --native-host"
    break
  fi
done
if [[ -z $CMD ]]; then
  bin=$(command -v switchyard 2>/dev/null) || {
    echo "error: Switchyard not found" >&2
    exit 1
  }
  CMD="$bin --native-host"
fi

# ── Config ─────────────────────────────────────────────────────────────────────

NAME="io.github.alyraffauf.switchyard"
WRAPPER="$HOME/.local/share/switchyard/native-host-wrapper.sh"
CHROME_ORIGIN="chrome-extension://gmdmmjfmpfbmddgphjbkbbmdolkifloi/"
FF_EXT="switchyard@alyraffauf.github.io"

NATIVE_CHROMIUM=(net.imput.helium google-chrome chromium BraveSoftware/Brave-Browser microsoft-edge vivaldi)
NATIVE_FIREFOX=(.mozilla .librewolf .zen)

# (appId:configSubdir) — configSubdir maps to ~/.config/<sub>/ inside the sandbox
FLATPAK_CHROMIUM=(
  com.google.Chrome:google-chrome
  com.google.ChromeDev:google-chrome-unstable
  org.chromium.Chromium:chromium
  com.brave.Browser:BraveSoftware/Brave-Browser
  com.microsoft.Edge:microsoft-edge
  com.vivaldi.Vivaldi:vivaldi
  io.github.ungoogled_software.ungoogled_chromium:chromium
  net.imput.helium:net.imput.helium
)

# (appId:profileDir) — profileDir is persisted as ~/<sub>/ inside the sandbox
FLATPAK_FIREFOX=(
  org.mozilla.firefox:.mozilla
  io.gitlab.librewolf-community:.librewolf
  io.github.zen_browser.zen:.zen
)

# ── Helpers ────────────────────────────────────────────────────────────────────

N=0

chromium_manifest() {
  printf '{"name":"%s","description":"Switchyard browser selector","path":"%s","type":"stdio","allowed_origins":["%s"]}\n' \
    "$NAME" "$1" "$CHROME_ORIGIN"
}

firefox_manifest() {
  printf '{"name":"%s","description":"Switchyard browser selector","path":"%s","type":"stdio","allowed_extensions":["%s"]}\n' \
    "$NAME" "$1" "$FF_EXT"
}

# ── Uninstall ──────────────────────────────────────────────────────────────────

if [[ ${1-} == "--uninstall" ]]; then
  echo "Removing Switchyard native messaging host..."

  for sub in "${NATIVE_CHROMIUM[@]}"; do
    f="$HOME/.config/$sub/NativeMessagingHosts/$NAME.json"
    [[ -f $f ]] && rm -f "$f" && echo "  removed $f" && N=$((N + 1)) || true
  done

  for sub in "${NATIVE_FIREFOX[@]}"; do
    f="$HOME/$sub/native-messaging-hosts/$NAME.json"
    [[ -f $f ]] && rm -f "$f" && echo "  removed $f" && N=$((N + 1)) || true
  done

  for entry in "${FLATPAK_CHROMIUM[@]}"; do
    appId="${entry%%:*}"
    sub="${entry##*:}"
    dir="$HOME/.var/app/$appId/config/$sub/NativeMessagingHosts"
    f="$dir/$NAME.json"
    [[ -f $f ]] && rm -f "$f" "$dir/switchyard-native-host.sh" && echo "  removed $f" && N=$((N + 1)) || true
  done

  for entry in "${FLATPAK_FIREFOX[@]}"; do
    appId="${entry%%:*}"
    sub="${entry##*:}"
    dir="$HOME/.var/app/$appId/$sub/native-messaging-hosts"
    f="$dir/$NAME.json"
    [[ -f $f ]] && rm -f "$f" "$dir/switchyard-native-host.sh" && echo "  removed $f" && N=$((N + 1)) || true
  done

  [[ -f $WRAPPER ]] && rm -f "$WRAPPER" && echo "  removed $WRAPPER" || true

  echo
  echo "Removed $N manifest(s)."
  exit 0
fi

# ── Install ────────────────────────────────────────────────────────────────────

echo "Installing Switchyard native messaging host..."

mkdir -p "$(dirname "$WRAPPER")"
printf '#!/bin/sh\nexec %s "$@"\n' "$CMD" >"$WRAPPER"
chmod +x "$WRAPPER"
echo "Wrapper: $WRAPPER"

# Flatpak browsers need a different wrapper: they detect the sandbox via
# $container and use flatpak-spawn --host to escape it. This requires the
# org.freedesktop.Flatpak talk-name, granted per-browser below.
FP_WRAPPER="#!/bin/sh
if [ \"\${container-}\" = flatpak ]; then
    exec /usr/bin/flatpak-spawn --host $CMD
else
    exec $CMD
fi"

for sub in "${NATIVE_CHROMIUM[@]}"; do
  [[ -d "$HOME/.config/$sub" ]] || continue
  target="$HOME/.config/$sub/NativeMessagingHosts/$NAME.json"
  mkdir -p "$(dirname "$target")"
  chromium_manifest "$WRAPPER" >"$target"
  echo "  $target"
  N=$((N + 1))
done

for sub in "${NATIVE_FIREFOX[@]}"; do
  [[ -d "$HOME/$sub" ]] || continue
  target="$HOME/$sub/native-messaging-hosts/$NAME.json"
  mkdir -p "$(dirname "$target")"
  firefox_manifest "$WRAPPER" >"$target"
  echo "  $target"
  N=$((N + 1))
done

for entry in "${FLATPAK_CHROMIUM[@]}"; do
  appId="${entry%%:*}"
  sub="${entry##*:}"
  [[ -d "$HOME/.var/app/$appId" ]] || continue
  dir="$HOME/.var/app/$appId/config/$sub/NativeMessagingHosts"
  wp="$dir/switchyard-native-host.sh"
  mkdir -p "$dir"
  printf '%s\n' "$FP_WRAPPER" >"$wp" && chmod +x "$wp"
  chromium_manifest "$wp" >"$dir/$NAME.json"
  flatpak override --user --talk-name=org.freedesktop.Flatpak "$appId"
  echo "  $dir/$NAME.json  [Flatpak]"
  N=$((N + 1))
done

for entry in "${FLATPAK_FIREFOX[@]}"; do
  appId="${entry%%:*}"
  sub="${entry##*:}"
  [[ -d "$HOME/.var/app/$appId" ]] || continue
  dir="$HOME/.var/app/$appId/$sub/native-messaging-hosts"
  wp="$dir/switchyard-native-host.sh"
  mkdir -p "$dir"
  printf '%s\n' "$FP_WRAPPER" >"$wp" && chmod +x "$wp"
  firefox_manifest "$wp" >"$dir/$NAME.json"
  flatpak override --user --talk-name=org.freedesktop.Flatpak "$appId"
  echo "  $dir/$NAME.json  [Flatpak]"
  N=$((N + 1))
done

[[ $N -gt 0 ]] || {
  echo
  echo "No supported browsers found."
  exit 0
}
echo
echo "Installed $N manifest(s). Restart your browser to activate."
