// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showEditRuleDialog(parent *adw.Window, cfg *Config, rule *Rule, browsers []*Browser, rebuildRulesList func()) {
	// Ensure rules have at least one condition
	if len(rule.Conditions) == 0 {
		rule.Conditions = []Condition{{
			Type:    "domain",
			Pattern: "",
		}}
		rule.Logic = "all"
	}

	var dialog *adw.Dialog
	header, saveBtn := dialogHeader("Cancel", "Save", func() { dialog.Close() }, nil)

	var scrolledWindow *gtk.ScrolledWindow
	dialog, _, scrolledWindow = dialogWithToolbar("Edit Rule", 600, 650, header)

	nameEntry, conditions, logicRow, alwaysAskRow, browserRow, formContent := buildRuleDialogContent(rule, browsers, saveBtn)
	scrolledWindow.SetChild(formContent)

	saveBtn.ConnectClicked(func() {
		browserIdx := browserRow.Selected()

		if len(*conditions) > 0 && int(browserIdx) < len(browsers) {
			if !validateConditions(*conditions) {
				return
			}

			// Update rule
			rule.Name = nameEntry.Text()
			rule.Conditions = *conditions
			rule.Logic = getLogicFromComboRow(logicRow)
			rule.Browser = browsers[browserIdx].ID
			rule.AlwaysAsk = alwaysAskRow.Active()

			saveConfigWithFlag(cfg)
			rebuildRulesList()
			dialog.Close()
		}
	})

	dialog.Present(parent)
}
