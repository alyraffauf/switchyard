// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"regexp"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showEditConditionDialog displays a dialog for editing a single condition.
func showEditConditionDialog(parent *adw.Window, cond *Condition, onSave func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Edit Condition")
	dialog.SetContentWidth(400)
	dialog.SetContentHeight(300)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel("Cancel")
	cancelBtn.ConnectClicked(func() { dialog.Close() })
	header.PackStart(cancelBtn)

	saveBtn := gtk.NewButton()
	saveBtn.SetLabel("Save")
	saveBtn.AddCSSClass("suggested-action")
	header.PackEnd(saveBtn)

	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	// Preferences group
	group := adw.NewPreferencesGroup()

	// Type row
	typeRow := adw.NewComboRow()
	typeRow.SetTitle("Type")
	typeRow.SetModel(gtk.NewStringList([]string{"Domain", "Contains", "Wildcard", "Regex"}))
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

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)

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
