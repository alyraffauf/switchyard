// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showAddRuleDialog displays the add rule dialog.
func showAddRuleDialog(parent *adw.Window, cfg *Config, browsers []*Browser, rebuildRulesList func()) {
	var dialog *adw.Dialog
	header, addBtn := dialogHeader("Cancel", "Add", func() { dialog.Close() }, nil)
	addBtn.SetSensitive(false) // Insensitive until at least one valid condition is added

	var scrolledWindow *gtk.ScrolledWindow
	dialog, _, scrolledWindow = dialogWithToolbar("Add Rule", 600, 650, header)

	nameEntry, conditions, logicRow, alwaysAskRow, browserRow, formContent := buildRuleDialogContent(nil, browsers, addBtn)
	scrolledWindow.SetChild(formContent)

	addBtn.ConnectClicked(func() {
		browserIdx := browserRow.Selected()

		if len(*conditions) > 0 && int(browserIdx) < len(browsers) {
			if !validateConditions(*conditions) {
				return
			}

			rule := Rule{
				Name:       nameEntry.Text(),
				Conditions: *conditions,
				Logic:      getLogicFromComboRow(logicRow),
				Browser:    browsers[browserIdx].ID,
				AlwaysAsk:  alwaysAskRow.Active(),
			}
			cfg.Rules = append(cfg.Rules, rule)
			saveConfigWithFlag(cfg)
			rebuildRulesList()
			dialog.Close()
		}
	})

	dialog.Present(parent)
}
