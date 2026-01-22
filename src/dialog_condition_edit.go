// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"regexp"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showEditConditionDialog displays a dialog for editing a single condition.
func showEditConditionDialog(parent *adw.Window, cond *Condition, onSave func()) {
	var dialog *adw.Dialog
	header, saveBtn := dialogHeader("Cancel", "Save", func() { dialog.Close() }, nil)
	dialog, content := simpleDialogWithToolbar("Edit Condition", 400, 300, header)

	// Preferences group
	group := adw.NewPreferencesGroup()

	// Type row
	typeRow := adw.NewComboRow()
	typeRow.SetTitle("Match Type")
	typeRow.SetModel(gtk.NewStringList([]string{"Exact Domain", "URL Contains", "Wildcard", "Regex"}))
	typeRow.SetSelected(conditionTypeToIndex(cond.Type))
	group.Add(typeRow)

	// Pattern row
	patternRow := adw.NewEntryRow()
	patternRow.SetTitle("Pattern")
	patternRow.SetText(cond.Pattern)
	group.Add(patternRow)

	// Error label for validation
	errorLabel := gtk.NewLabel("")
	errorLabel.SetWrap(true)
	errorLabel.AddCSSClass("error")
	errorLabel.SetVisible(false)

	// Function to validate and update error display
	updateValidation := func() {
		condType := indexToConditionType(typeRow.Selected())
		pattern := patternRow.Text()

		err := validateConditionPattern(condType, pattern)

		if err != nil {
			errorLabel.SetLabel(err.Error())
			errorLabel.SetVisible(true)
			patternRow.AddCSSClass("error")
			saveBtn.SetSensitive(false)
		} else {
			errorLabel.SetVisible(false)
			patternRow.RemoveCSSClass("error")
			saveBtn.SetSensitive(pattern != "")
		}
	}

	patternRow.Connect("changed", func() {
		updateValidation()
	})

	typeRow.Connect("notify::selected", func() {
		updateValidation()
	})

	content.Append(group)
	content.Append(errorLabel)

	// Initial validation check
	updateValidation()

	saveBtn.ConnectClicked(func() {
		pattern := patternRow.Text()
		if pattern == "" {
			return
		}

		// Final validation
		selectedType := typeRow.Selected()
		if selectedType == 3 { // Regex
			if _, err := regexp.Compile(pattern); err != nil {
				return
			}
		}

		// Update condition
		cond.Type = indexToConditionType(selectedType)
		cond.Pattern = pattern

		onSave()
		dialog.Close()
	})

	dialog.Present(parent)
}
