// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showDefaultBrowserPrompt displays a dialog asking to set Switchyard as default browser.
func showDefaultBrowserPrompt(parent gtk.Widgetter, cfg *Config, updateUI func()) {
	dialog := adw.NewAlertDialog(
		"Set as Default Browser",
		"Switchyard needs to be your default browser to intercept and route URLs to the correct browser.",
	)

	dialog.AddResponse("no", "Skip")
	dialog.AddResponse("later", "Later")
	dialog.AddResponse("yes", "Set as Default")

	dialog.SetDefaultResponse("yes")
	dialog.SetCloseResponse("later")

	dialog.ConnectResponse(func(response string) {
		if response == "yes" {
			setAsDefaultBrowser()
			cfg.CheckDefaultBrowser = false
			saveConfig(cfg)
			updateUI()
		} else if response == "no" {
			cfg.CheckDefaultBrowser = false
			saveConfig(cfg)
			updateUI()
		}
	})

	dialog.Present(parent)
}
