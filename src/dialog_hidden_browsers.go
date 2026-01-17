// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showHiddenBrowsersDialog displays a dialog for selecting which browsers to hide from the picker.
func showHiddenBrowsersDialog(parent *adw.Window, cfg *Config, browsers []*Browser) {
	dialog := adw.NewAlertDialog(
		"Hidden Browsers",
		"Select browsers to hide from the picker window. Hidden browsers won't appear in the picker, but can still be used in rules and settings.",
	)

	// Create a scrolled window for the browser list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetSizeRequest(400, 300)

	// Create a list box for browsers
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")

	// Create a map for quick lookup of hidden browsers
	hiddenSet := make(map[string]bool)
	for _, id := range cfg.HiddenBrowsers {
		hiddenSet[id] = true
	}

	// Add a row for each browser
	for _, browser := range browsers {
		b := browser // capture for closure

		row := gtk.NewListBoxRow()
		row.SetActivatable(false)

		rowBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
		rowBox.SetMarginStart(12)
		rowBox.SetMarginEnd(12)
		rowBox.SetMarginTop(8)
		rowBox.SetMarginBottom(8)

		// Browser icon
		icon := loadBrowserIcon(b, 24)
		rowBox.Append(icon)

		// Browser name
		nameLabel := gtk.NewLabel(b.Name)
		nameLabel.SetXAlign(0)
		nameLabel.SetHExpand(true)
		rowBox.Append(nameLabel)

		// Checkbox
		checkBox := gtk.NewCheckButton()
		checkBox.SetActive(hiddenSet[b.ID])
		checkBox.SetVAlign(gtk.AlignCenter)

		// Connect handler to update config
		checkBox.ConnectToggled(func() {
			isHidden := checkBox.Active()

			// Update the hidden browsers list
			if isHidden {
				// Add to hidden list if not already present
				found := false
				for _, id := range cfg.HiddenBrowsers {
					if id == b.ID {
						found = true
						break
					}
				}
				if !found {
					cfg.HiddenBrowsers = append(cfg.HiddenBrowsers, b.ID)
				}
			} else {
				// Remove from hidden list
				newHidden := make([]string, 0)
				for _, id := range cfg.HiddenBrowsers {
					if id != b.ID {
						newHidden = append(newHidden, id)
					}
				}
				cfg.HiddenBrowsers = newHidden
			}

			// Save config
			saveConfigWithFlag(cfg)
		})

		rowBox.Append(checkBox)
		row.SetChild(rowBox)
		listBox.Append(row)
	}

	scrolled.SetChild(listBox)
	dialog.SetExtraChild(scrolled)

	dialog.AddResponse("close", "Close")
	dialog.SetDefaultResponse("close")
	dialog.SetCloseResponse("close")

	dialog.Present(parent)
}
