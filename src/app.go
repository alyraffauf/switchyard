// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "os"

const defaultAppID = "io.github.alyraffauf.Switchyard"

// Application metadata
const (
	AppName       = "Switchyard"
	DeveloperName = "Aly Raffauf"
	Copyright     = "© 2026 Aly Raffauf"
	Version       = "0.7.0"

	// Links
	WebsiteURL = "https://github.com/alyraffauf/switchyard"
	IssueURL   = "https://github.com/alyraffauf/switchyard/issues"
	DonateURL  = "https://ko-fi.com/alyraffauf"
)

// Contributor represents a person who contributed to the project
type Contributor struct {
	Name string
	URL  string
}

// Contributors is a list of people who contributed to the project
var Contributors = []Contributor{
	{Name: "Aly Raffauf", URL: "https://github.com/alyraffauf"},
}

// getAppID returns the application ID.
// When running in Flatpak, it uses the FLATPAK_ID environment variable
// to support both production and development (.Devel) builds.
func getAppID() string {
	if flatpakID := os.Getenv("FLATPAK_ID"); flatpakID != "" {
		return flatpakID
	}
	return defaultAppID
}
