// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showEditRuleDialog displays the edit rule dialog.
func showEditRuleDialog(parent *adw.Window, cfg *Config, rule *Rule, browsers []*Browser, rebuildRulesList func()) {
	// Ensure rules have at least one condition
	if len(rule.Conditions) == 0 {
		rule.Conditions = []Condition{{
			Type:    "domain",
			Pattern: "",
		}}
		rule.Logic = "all"
	}

	dialog := adw.NewDialog()
	dialog.SetTitle("Edit Rule")
	dialog.SetContentWidth(600)
	dialog.SetContentHeight(650)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel("Cancel")
	cancelBtn.SetTooltipText("Cancel and close")
	cancelBtn.ConnectClicked(func() { dialog.Close() })
	header.PackStart(cancelBtn)

	saveBtn := gtk.NewButton()
	saveBtn.SetLabel("Save")
	saveBtn.AddCSSClass("suggested-action")
	saveBtn.SetTooltipText("Save changes to this rule")
	header.PackEnd(saveBtn)

	toolbarView.AddTopBar(header)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolledWindow.SetVExpand(true)

	nameEntry, conditions, logicRow, alwaysAskRow, browserRow, content := buildRuleDialogContent(rule, browsers, saveBtn)

	scrolledWindow.SetChild(content)
	toolbarView.SetContent(scrolledWindow)
	dialog.SetChild(toolbarView)

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
