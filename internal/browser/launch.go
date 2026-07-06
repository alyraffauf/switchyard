// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browser

import (
	"os"
	"os/exec"
	"strings"

	"github.com/alyraffauf/goxdgdesktop/desktopexec"
)

const activationTokenEnv = "XDG_ACTIVATION_TOKEN"

// buildCommand assembles the OS command that launches cmdline for url via a
// desktop Exec line, honoring activationToken (the Wayland/X11 startup token
// for raising the window, may be empty) and running on the host via
// flatpak-spawn when inFlatpak is set. Returns nil when the resolved command
// line is empty.
func buildCommand(cmdline, url, activationToken string, inFlatpak bool) *exec.Cmd {
	if url != "" {
		cmdline = desktopexec.SubstituteOrAppendURL(cmdline, url)
	} else {
		// Strip field codes. Shell-quoting an empty URL produces "''",
		// which most browsers open as file:///.
		cmdline = stripDesktopFieldCodes(cmdline)
	}
	if strings.TrimSpace(cmdline) == "" {
		return nil
	}

	// The sandbox drops our environment, so pass the token to the host via --env.
	if inFlatpak && !strings.HasPrefix(cmdline, "flatpak-spawn") {
		return exec.Command("flatpak-spawn", "--host",
			"--env="+activationTokenEnv+"="+activationToken, "sh", "-c", cmdline)
	}

	// Let the shell parse quotes; see switchyard issue #27.
	cmd := exec.Command("sh", "-c", cmdline)
	if activationToken != "" {
		cmd.Env = append(os.Environ(), activationTokenEnv+"="+activationToken)
	}
	return cmd
}

// stripDesktopFieldCodes removes desktop field codes without shell-quoting.
// SubstituteURL would produce "”" for an empty URL, which most browsers
// interpret as file:///.
func stripDesktopFieldCodes(cmdline string) string {
	for _, code := range []string{"%u", "%U", "%f", "%F", "%i", "%c", "%k"} {
		cmdline = strings.ReplaceAll(cmdline, code, "")
	}
	return cmdline
}
