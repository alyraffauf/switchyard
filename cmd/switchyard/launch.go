// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/alyraffauf/goxdgdesktop/desktopexec"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

// launchCommand executes a desktop file command line with URL substitution
// and proper activation token handling for window raising on Wayland.
func launchCommand(cmdline, url string, appInfo *gio.AppInfo) error {
	cmdline = desktopexec.SubstituteOrAppendURL(cmdline, url)

	if strings.TrimSpace(cmdline) == "" {
		return nil
	}

	// Get activation token from GDK launch context for window raising on Wayland
	var activationToken string
	if display := gdk.DisplayGetDefault(); display != nil {
		activationToken = display.AppLaunchContext().StartupNotifyID(appInfo, nil)
	}

	var cmd *exec.Cmd
	// When running in Flatpak, wrap with flatpak-spawn --host
	// and pass activation token via --env flag
	if os.Getenv("FLATPAK_ID") != "" && !strings.HasPrefix(cmdline, "flatpak-spawn") {
		cmd = exec.Command("flatpak-spawn", "--host", "--env=XDG_ACTIVATION_TOKEN="+activationToken, "sh", "-c", cmdline)
		activationToken = "" // Already handled via flatpak-spawn
	} else {
		// let the shell handle quote parsing
		// https://github.com/alyraffauf/switchyard/issues/27
		cmd = exec.Command("sh", "-c", cmdline)
		if activationToken != "" {
			cmd.Env = append(os.Environ(), "XDG_ACTIVATION_TOKEN="+activationToken)
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go cmd.Wait()
	return nil
}
