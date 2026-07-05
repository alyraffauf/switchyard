// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showHiddenBrowsersDialog(parent *adw.Window, cfg *Config, browsers []*Browser) {
	dialog := adw.NewAlertDialog(
		"Hidden Browsers",
		"You can still use hidden browsers in rules.",
	)

	scrolled := createScrolledWindow()
	scrolled.SetSizeRequest(400, 300)

	listBox := createBoxedListBox()

	hiddenSet := make(map[string]bool)
	for _, id := range cfg.HiddenBrowsers {
		hiddenSet[id] = true
	}

	for _, browser := range browsers {
		b := browser // capture for closure

		row := gtk.NewListBoxRow()
		row.SetActivatable(false)

		rowBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
		rowBox.SetMarginStart(12)
		rowBox.SetMarginEnd(12)
		rowBox.SetMarginTop(8)
		rowBox.SetMarginBottom(8)

		icon := loadBrowserIcon(b, 24)
		rowBox.Append(icon)

		nameLabel := gtk.NewLabel(b.Name)
		nameLabel.SetXAlign(0)
		nameLabel.SetHExpand(true)
		rowBox.Append(nameLabel)

		checkBox := gtk.NewCheckButton()
		checkBox.SetActive(hiddenSet[b.ID])
		checkBox.SetVAlign(gtk.AlignCenter)

		checkBox.ConnectToggled(func() {
			isHidden := checkBox.Active()

			if isHidden {
				// Append only if not already hidden.
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
				newHidden := make([]string, 0)
				for _, id := range cfg.HiddenBrowsers {
					if id != b.ID {
						newHidden = append(newHidden, id)
					}
				}
				cfg.HiddenBrowsers = newHidden
			}

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
