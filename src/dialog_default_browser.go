// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showDefaultBrowserPrompt displays a dialog asking to set Switchyard as default browser.
func showDefaultBrowserPrompt(parent gtk.Widgetter, cfg *Config, updateUI func()) {
	dialog := adw.NewAlertDialog(
		"Set as Default Browser?",
		"Switchyard works best when set as your default browser. This allows it to intercept links and route them to the appropriate browser based on your rules.",
	)

	dialog.AddResponse("no", "Don't Ask Again")
	dialog.AddResponse("later", "Not Now")
	dialog.AddResponse("yes", "Set as Default")

	dialog.SetDefaultResponse("later")
	dialog.SetCloseResponse("later")

	// Make "Set as Default" the suggested action
	dialog.SetResponseAppearance("yes", adw.ResponseSuggested)

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
