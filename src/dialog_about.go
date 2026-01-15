// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showAboutDialog displays the application's about dialog.
func showAboutDialog(parent *adw.Window) {
	about := adw.NewAboutDialog()

	about.SetApplicationName(AppName)
	about.SetApplicationIcon(getAppID())
	about.SetVersion(Version)
	about.SetDeveloperName(DeveloperName)
	about.SetCopyright(Copyright)
	about.SetLicenseType(gtk.LicenseGPL30)
	about.SetWebsite(WebsiteURL)
	about.SetIssueURL(IssueURL)

	// Set developers from contributors list
	developerStrings := make([]string, len(Contributors))
	for i, contributor := range Contributors {
		developerStrings[i] = contributor.Name + " " + contributor.URL
	}

	about.SetDevelopers(developerStrings)

	about.Present(parent)
}
