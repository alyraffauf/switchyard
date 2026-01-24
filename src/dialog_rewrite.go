// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showAddRewriteDialog(parent *adw.Window, cfg *Config, onSave func()) {
	showRewriteDialog(parent, cfg, nil, onSave)
}

func showEditRewriteDialog(parent *adw.Window, cfg *Config, rewrite *Rewrite, onSave func()) {
	showRewriteDialog(parent, cfg, rewrite, onSave)
}

func showRewriteDialog(parent *adw.Window, cfg *Config, rewrite *Rewrite, onSave func()) {
	isNew := rewrite == nil

	var title, actionLabel string
	if isNew {
		title = "Add Rewrite"
		actionLabel = "Add"
	} else {
		title = "Edit Rewrite"
		actionLabel = "Save"
	}

	var dialog *adw.Dialog
	header, saveBtn := dialogHeader("Cancel", actionLabel, func() { dialog.Close() }, nil)
	dialog, content, _ := dialogWithToolbar(title, 450, 350, header)

	group := adw.NewPreferencesGroup()

	typeRow := adw.NewComboRow()
	typeRow.SetTitle("Type")
	typeRow.SetModel(gtk.NewStringList(getRewriteTypeLabels()))
	if !isNew {
		typeRow.SetSelected(rewriteTypeToIndex(rewrite.Type))
	}
	group.Add(typeRow)

	findRow := adw.NewEntryRow()
	findRow.SetTitle("Find")
	if !isNew {
		findRow.SetText(rewrite.Find)
	}
	group.Add(findRow)

	replaceRow := adw.NewEntryRow()
	replaceRow.SetTitle("Replace with")
	if !isNew {
		replaceRow.SetText(rewrite.Replace)
	}
	group.Add(replaceRow)

	helpLabel := gtk.NewLabel("")
	helpLabel.SetWrap(true)
	helpLabel.SetXAlign(0)
	helpLabel.AddCSSClass("dim-label")
	helpLabel.SetMarginTop(12)

	updateHelpText := func() {
		if typeRow.Selected() == 0 { // Domain
			helpLabel.SetLabel("Matches the exact domain name")
		} else { // URL
			helpLabel.SetLabel("Use * to match any text")
		}
	}
	updateHelpText()

	typeRow.Connect("notify::selected", updateHelpText)

	content.Append(group)
	content.Append(helpLabel)

	validateInputs := func() {
		find := findRow.Text()
		rwType := indexToRewriteType(typeRow.Selected())
		r := Rewrite{Type: rwType, Find: find, Replace: replaceRow.Text()}
		err := validateRewrite(r)
		saveBtn.SetSensitive(err == nil)
	}

	findRow.Connect("changed", validateInputs)
	typeRow.Connect("notify::selected", validateInputs)
	validateInputs()

	saveBtn.ConnectClicked(func() {
		find := findRow.Text()
		rwType := indexToRewriteType(typeRow.Selected())

		if isNew {
			cfg.Rewrites = append(cfg.Rewrites, Rewrite{
				Type:    rwType,
				Find:    find,
				Replace: replaceRow.Text(),
			})
		} else {
			rewrite.Type = rwType
			rewrite.Find = find
			rewrite.Replace = replaceRow.Text()
		}

		saveConfig(cfg)
		onSave()
		dialog.Close()
	})

	dialog.Present(parent)
}
