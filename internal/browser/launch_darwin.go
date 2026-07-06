// SPDX-License-Identifier: GPL-3.0-or-later

//go:build darwin

package browser

import "os/exec"

// buildCommand assembles the macOS open invocation for bundlePath and url.
// Linux-only launch options are ignored. Returns nil when bundlePath is empty.
func buildCommand(bundlePath, url, _ string, _ bool) *exec.Cmd {
	if bundlePath == "" {
		return nil
	}

	args := []string{"-a", bundlePath}
	if url != "" {
		args = append(args, url)
	}
	return exec.Command("open", args...)
}
