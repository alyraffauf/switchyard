// SPDX-License-Identifier: GPL-3.0-or-later

// Package host runs commands on the host system and detects the Flatpak
// sandbox. It is toolkit-agnostic so backend packages can reason about the
// runtime environment without depending on the GTK layer.
package host

import (
	"os"
	"os/exec"
	"strings"
)

const desktopSuffix = ".desktop"

// InFlatpak reports whether the process is running inside a Flatpak sandbox.
func InFlatpak() bool {
	return os.Getenv("FLATPAK_ID") != ""
}

// HostCommand builds a command that runs on the host, hopping out of the
// Flatpak sandbox via flatpak-spawn --host when one is active.
func HostCommand(name string, args ...string) *exec.Cmd {
	if !InFlatpak() {
		return exec.Command(name, args...)
	}
	hostArgs := append([]string{"--host", name}, args...)
	return exec.Command("flatpak-spawn", hostArgs...)
}

// IsDefaultBrowser reports whether appID's desktop entry is the system default
// web browser, per xdg-settings.
func IsDefaultBrowser(appID string) bool {
	output, err := HostCommand("xdg-settings", "get", "default-web-browser").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == appID+desktopSuffix
}

// SetDefaultBrowser registers appID's desktop entry as the system default web
// browser, per xdg-settings.
func SetDefaultBrowser(appID string) error {
	return HostCommand("xdg-settings", "set", "default-web-browser", appID+desktopSuffix).Run()
}
