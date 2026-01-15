// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showAddRuleDialog displays the add rule dialog.
func showAddRuleDialog(parent *adw.Window, cfg *Config, browsers []*Browser, rebuildRulesList func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Add Rule")
	dialog.SetContentWidth(600)
	dialog.SetContentHeight(650)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel("Cancel")
	cancelBtn.ConnectClicked(func() { dialog.Close() })
	header.PackStart(cancelBtn)

	addBtn := gtk.NewButton()
	addBtn.SetLabel("Add")
	addBtn.AddCSSClass("suggested-action")
	addBtn.SetSensitive(false) // Insensitive until at least one valid condition is added
	header.PackEnd(addBtn)

	toolbarView.AddTopBar(header)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolledWindow.SetVExpand(true)

	nameEntry, conditions, logicRow, alwaysAskRow, browserRow, content := buildRuleDialogContent(nil, browsers, addBtn)

	scrolledWindow.SetChild(content)
	toolbarView.SetContent(scrolledWindow)
	dialog.SetChild(toolbarView)

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
