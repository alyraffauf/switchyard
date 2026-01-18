#!/usr/bin/env bash
# Wrapper script to set XDG_DATA_DIRS for browser discovery in Flatpak
DATA_HOME="${HOME}/.local/share"
export XDG_DATA_DIRS="${XDG_DATA_DIRS:-/usr/local/share:/usr/share}:/run/host/usr/share:/var/lib/flatpak/exports/share:/var/lib/snapd/desktop:${DATA_HOME}/flatpak/exports/share"
exec /app/bin/switchyard-bin "$@"
