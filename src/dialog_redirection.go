// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showAddRedirectionDialog(parent *adw.Window, cfg *Config, onSave func()) {
	showRedirectionDialog(parent, cfg, nil, onSave)
}

func showEditRedirectionDialog(parent *adw.Window, cfg *Config, redirection *Redirection, onSave func()) {
	showRedirectionDialog(parent, cfg, redirection, onSave)
}

func showRedirectionDialog(parent *adw.Window, cfg *Config, redirection *Redirection, onSave func()) {
	isNew := redirection == nil

	var title, actionLabel string
	if isNew {
		title = "Add Redirection"
		actionLabel = "Add"
	} else {
		title = "Edit Redirection"
		actionLabel = "Save"
	}

	var dialog *adw.Dialog
	header, saveBtn := dialogHeader("Cancel", actionLabel, func() { dialog.Close() }, nil)
	dialog, content, _ := dialogWithToolbar(title, 450, 350, header)

	group := adw.NewPreferencesGroup()

	typeRow := adw.NewComboRow()
	typeRow.SetTitle("Type")
	typeRow.SetModel(gtk.NewStringList(getRedirectionTypeLabels()))
	if !isNew {
		typeRow.SetSelected(redirectionTypeToIndex(redirection.Type))
	}
	group.Add(typeRow)

	findRow := adw.NewEntryRow()
	findRow.SetTitle("Match")
	if !isNew {
		findRow.SetText(redirection.Find)
	}
	group.Add(findRow)

	replaceRow := adw.NewEntryRow()
	replaceRow.SetTitle("Replace With")
	if !isNew {
		replaceRow.SetText(redirection.Replace)
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
		} else { // Pattern
			helpLabel.SetLabel("Use * as a wildcard to match any text")
		}
	}
	updateHelpText()

	typeRow.Connect("notify::selected", updateHelpText)

	content.Append(group)
	content.Append(helpLabel)

	validateInputs := func() {
		find := findRow.Text()
		rwType := indexToRedirectionType(typeRow.Selected())
		r := Redirection{Type: rwType, Find: find, Replace: replaceRow.Text()}
		err := validateRedirection(r)
		saveBtn.SetSensitive(err == nil)
	}

	findRow.Connect("changed", validateInputs)
	typeRow.Connect("notify::selected", validateInputs)
	validateInputs()

	saveBtn.ConnectClicked(func() {
		find := findRow.Text()
		rwType := indexToRedirectionType(typeRow.Selected())

		if isNew {
			cfg.Redirections = append(cfg.Redirections, Redirection{
				Type:    rwType,
				Find:    find,
				Replace: replaceRow.Text(),
			})
		} else {
			redirection.Type = rwType
			redirection.Find = find
			redirection.Replace = replaceRow.Text()
		}

		saveConfig(cfg)
		onSave()
		dialog.Close()
	})

	dialog.Present(parent)
}
