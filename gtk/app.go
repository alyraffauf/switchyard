// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import "os"

const defaultAppID = "io.github.alyraffauf.Switchyard"

// Application metadata
const (
	AppName       = "Switchyard"
	DeveloperName = "Aly Raffauf"
	Copyright     = "© 2026 Aly Raffauf"
	Version       = "0.17.1"

	// Links
	WebsiteURL          = "https://switchyard.aly.codes/"
	IssueURL            = "https://github.com/alyraffauf/switchyard/issues"
	DonateURL           = "https://ko-fi.com/alyraffauf"
	ChromeExtensionURL  = "https://chromewebstore.google.com/detail/switchyard/ncehhpikkabfdcceimdhjjjodogflokc"
	FirefoxExtensionURL = "https://addons.mozilla.org/firefox/addon/switchyard/"
	DocsURL             = "https://switchyard.aly.codes/docs"
)

type Contributor struct {
	Name string
	URL  string
}

var Contributors = []Contributor{
	{Name: "Aly Raffauf", URL: "https://github.com/alyraffauf"},
}

// getAppID returns the application ID.
// When running in Flatpak, it uses the FLATPAK_ID environment variable
func getAppID() string {
	if flatpakID := os.Getenv("FLATPAK_ID"); flatpakID != "" {
		return flatpakID
	}
	return defaultAppID
}
