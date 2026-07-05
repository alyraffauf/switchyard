// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

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

	developerStrings := make([]string, len(Contributors))
	for i, contributor := range Contributors {
		developerStrings[i] = contributor.Name + " " + contributor.URL
	}

	about.SetDevelopers(developerStrings)

	about.Present(parent)
}
