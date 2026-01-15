// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "os"

const defaultAppID = "io.github.alyraffauf.Switchyard"

// getAppID returns the application ID.
// When running in Flatpak, it uses the FLATPAK_ID environment variable
// to support both production and development (.Devel) builds.
func getAppID() string {
	if flatpakID := os.Getenv("FLATPAK_ID"); flatpakID != "" {
		return flatpakID
	}
	return defaultAppID
}
