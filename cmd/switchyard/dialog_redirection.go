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
	dialog, content, _ := dialogWithToolbar(title, 450, 450, header)

	// Name section
	nameGroup := adw.NewPreferencesGroup()
	nameGroup.SetTitle("Name")
	nameGroup.SetDescription("Give this redirection a descriptive name (optional)")

	nameRow := adw.NewEntryRow()
	nameRow.SetTitle("Name")
	if !isNew {
		nameRow.SetText(redirection.Name)
	}
	nameGroup.Add(nameRow)
	content.Append(nameGroup)

	// Redirection section
	group := adw.NewPreferencesGroup()
	group.SetTitle("Redirection")
	group.SetDescription("Define how links are modified")

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

	content.Append(group)

	validateInputs := func() {
		find := findRow.Text()
		rwType := indexToRedirectionType(typeRow.Selected())
		r := Redirection{Type: rwType, Find: find, Replace: replaceRow.Text()}
		err := validateRedirection(r)

		if find != "" && err != nil {
			findRow.AddCSSClass("error")
		} else {
			findRow.RemoveCSSClass("error")
		}
		saveBtn.SetSensitive(err == nil)
	}

	nameRow.Connect("changed", validateInputs)
	findRow.Connect("changed", validateInputs)
	replaceRow.Connect("changed", validateInputs)
	typeRow.Connect("notify::selected", validateInputs)
	validateInputs()

	saveBtn.ConnectClicked(func() {
		name := nameRow.Text()
		find := findRow.Text()
		rwType := indexToRedirectionType(typeRow.Selected())

		if isNew {
			cfg.Redirections = append(cfg.Redirections, Redirection{
				Name:    name,
				Type:    rwType,
				Find:    find,
				Replace: replaceRow.Text(),
			})
		} else {
			redirection.Name = name
			redirection.Type = rwType
			redirection.Find = find
			redirection.Replace = replaceRow.Text()
		}

		saveConfigWithFlag(cfg)
		onSave()
		dialog.Close()
	})

	dialog.Present(parent)
}
