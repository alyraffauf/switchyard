// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

// launchCommand executes a desktop file command line with URL substitution
// and proper activation token handling for window raising on Wayland.
func launchCommand(cmdline, url string, appInfo *gio.AppInfo) error {
	// Replace %u, %U, %f, %F with URL
	cmdline = strings.ReplaceAll(cmdline, "%u", url)
	cmdline = strings.ReplaceAll(cmdline, "%U", url)
	cmdline = strings.ReplaceAll(cmdline, "%f", url)
	cmdline = strings.ReplaceAll(cmdline, "%F", url)

	// Remove other field codes
	for _, code := range []string{"%i", "%c", "%k"} {
		cmdline = strings.ReplaceAll(cmdline, code, "")
	}

	// Parse the command line
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return nil
	}

	// Get activation token from GDK launch context for window raising on Wayland
	var activationToken string
	if display := gdk.DisplayGetDefault(); display != nil {
		activationToken = display.AppLaunchContext().StartupNotifyID(appInfo, nil)
	}

	// When running in Flatpak, wrap with flatpak-spawn --host
	// and pass activation token via --env flag
	if os.Getenv("FLATPAK_ID") != "" && !strings.HasPrefix(parts[0], "flatpak-spawn") {
		parts = append([]string{"flatpak-spawn", "--host", "--env=XDG_ACTIVATION_TOKEN=" + activationToken}, parts...)
		activationToken = "" // Already handled via flatpak-spawn
	}

	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	if activationToken != "" {
		cmd.Env = append(os.Environ(), "XDG_ACTIVATION_TOKEN="+activationToken)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go cmd.Wait()
	return nil
}
