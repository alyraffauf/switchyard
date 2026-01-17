// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// buildRuleDialogContent creates the shared content for add/edit rule dialogs
func buildRuleDialogContent(
	initialRule *Rule,
	browsers []*Browser,
	actionBtn *gtk.Button,
) (
	nameEntry *adw.EntryRow,
	conditions *[]Condition,
	logicRow *adw.ComboRow,
	alwaysAskRow *adw.SwitchRow,
	browserRow *adw.ComboRow,
	content *gtk.Box,
) {
	content = gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	// Name section
	nameGroup := adw.NewPreferencesGroup()
	nameGroup.SetTitle("Rule Name")
	nameGroup.SetDescription("Give this rule a descriptive name (optional)")

	nameEntry = adw.NewEntryRow()
	nameEntry.SetTitle("Name")
	if initialRule != nil {
		nameEntry.SetText(initialRule.Name)
	}
	nameGroup.Add(nameEntry)
	content.Append(nameGroup)

	// Initialize conditions
	var conditionsSlice []Condition
	if initialRule != nil && len(initialRule.Conditions) > 0 {
		conditionsSlice = make([]Condition, len(initialRule.Conditions))
		copy(conditionsSlice, initialRule.Conditions)
	} else {
		conditionsSlice = []Condition{{Type: "domain", Pattern: ""}}
	}
	conditions = &conditionsSlice

	// Conditions section
	conditionsGroup := adw.NewPreferencesGroup()
	conditionsGroup.SetTitle("Conditions")
	conditionsGroup.SetDescription("Define conditions to match URLs")

	conditionsListBox := gtk.NewListBox()
	conditionsListBox.SetSelectionMode(gtk.SelectionNone)
	conditionsListBox.AddCSSClass("boxed-list")

	// Logic selector row
	logicRow = adw.NewComboRow()
	logicRow.SetTitle("Match Logic")
	logicRow.SetModel(gtk.NewStringList([]string{"All conditions", "Any condition"}))
	if initialRule != nil && initialRule.Logic == "any" {
		logicRow.SetSelected(1)
	} else {
		logicRow.SetSelected(0)
	}
	conditionsListBox.Append(logicRow)

	var conditionRows []*gtk.ListBoxRow
	var rebuildConditions func()

	rebuildConditions = func() {
		// Clear existing condition rows
		for _, row := range conditionRows {
			conditionsListBox.Remove(row)
		}
		conditionRows = nil

		// Build condition rows
		for i := range *conditions {
			condIdx := i
			row := createConditionRow(
				conditions,
				condIdx,
				actionBtn,
				rebuildConditions,
			)
			conditionsListBox.Append(row)
			conditionRows = append(conditionRows, row)
		}

		// Add "Add Condition" row at the end of the list
		addConditionRow := adw.NewActionRow()
		addConditionRow.SetTitle("Add Condition")
		addConditionRow.AddPrefix(gtk.NewImageFromIconName("list-add-symbolic"))
		addConditionRow.SetActivatable(true)
		addConditionRow.ConnectActivated(func() {
			*conditions = append(*conditions, Condition{Type: "domain", Pattern: ""})
			rebuildConditions()
		})
		conditionsListBox.Append(addConditionRow)

		// Update action button state
		actionBtn.SetSensitive(areAllConditionsValid(*conditions))
	}

	rebuildConditions()
	conditionsGroup.Add(conditionsListBox)
	content.Append(conditionsGroup)

	// Action section
	actionGroup := adw.NewPreferencesGroup()
	actionGroup.SetTitle("Browser Action")
	actionGroup.SetDescription("Select which browser opens matching URLs")

	alwaysAskRow = adw.NewSwitchRow()
	alwaysAskRow.SetTitle("Always show picker")
	alwaysAskRow.SetSubtitle("Ask which browser to use each time")
	if initialRule != nil {
		alwaysAskRow.SetActive(initialRule.AlwaysAsk)
	}
	actionGroup.Add(alwaysAskRow)

	// Browser dropdown
	browserNames := make([]string, len(browsers))
	selectedIdx := uint(0)
	for i, b := range browsers {
		browserNames[i] = b.Name
		if initialRule != nil && b.ID == initialRule.Browser {
			selectedIdx = uint(i)
		}
	}

	browserRow = adw.NewComboRow()
	browserRow.SetTitle("Browser")
	browserRow.SetModel(gtk.NewStringList(browserNames))
	browserRow.SetSelected(selectedIdx)
	if initialRule != nil {
		browserRow.SetSensitive(!initialRule.AlwaysAsk)
	}
	actionGroup.Add(browserRow)

	// Link always ask toggle to browser row sensitivity
	alwaysAskRow.Connect("notify::active", func() {
		browserRow.SetSensitive(!alwaysAskRow.Active())
	})

	content.Append(actionGroup)

	return
}

// createConditionRow creates a single condition editing row with all controls
func createConditionRow(
	conditions *[]Condition,
	condIdx int,
	actionBtn *gtk.Button,
	rebuildConditions func(),
) *gtk.ListBoxRow {
	conditionRow := gtk.NewListBoxRow()
	conditionRow.SetActivatable(false)
	conditionRow.SetSelectable(false)

	conditionContainer := gtk.NewBox(gtk.OrientationHorizontal, 8)
	conditionContainer.SetMarginTop(8)
	conditionContainer.SetMarginBottom(8)
	conditionContainer.SetMarginStart(12)
	conditionContainer.SetMarginEnd(12)

	// Match type dropdown
	typeDropdown := gtk.NewDropDown(
		gtk.NewStringList([]string{"Exact Domain", "URL Contains", "Wildcard", "Regex"}),
		nil,
	)
	typeDropdown.SetSelected(conditionTypeToIndex((*conditions)[condIdx].Type))
	typeDropdown.SetVAlign(gtk.AlignCenter)
	typeDropdown.SetSizeRequest(150, -1) // Fixed width for consistent alignment
	conditionContainer.Append(typeDropdown)

	// Pattern entry
	patternEntry := gtk.NewEntry()
	patternEntry.SetText((*conditions)[condIdx].Pattern)
	patternEntry.SetHExpand(true)
	patternEntry.SetPlaceholderText("Pattern")
	conditionContainer.Append(patternEntry)

	// Connect handlers
	typeDropdown.Connect("notify::selected", func() {
		(*conditions)[condIdx].Type = indexToConditionType(typeDropdown.Selected())
		validateConditionEntry(conditions, condIdx, typeDropdown, patternEntry, actionBtn)
	})

	patternEntry.Connect("changed", func() {
		(*conditions)[condIdx].Pattern = patternEntry.Text()
		validateConditionEntry(conditions, condIdx, typeDropdown, patternEntry, actionBtn)
	})

	// Delete button
	deleteBtn := gtk.NewButton()
	deleteBtn.SetIconName("edit-delete-symbolic")
	deleteBtn.SetTooltipText("Delete this condition")
	deleteBtn.AddCSSClass("flat")
	deleteBtn.AddCSSClass("circular")
	deleteBtn.AddCSSClass("destructive-action")
	deleteBtn.SetVAlign(gtk.AlignCenter)
	deleteBtn.SetSensitive(len(*conditions) > 1)
	deleteBtn.ConnectClicked(func() {
		if len(*conditions) > 1 && condIdx < len(*conditions) {
			*conditions = append((*conditions)[:condIdx], (*conditions)[condIdx+1:]...)
			rebuildConditions()
		}
	})
	conditionContainer.Append(deleteBtn)

	conditionRow.SetChild(conditionContainer)
	return conditionRow
}

// validateConditionEntry validates a pattern and updates UI accordingly
func validateConditionEntry(
	conditions *[]Condition,
	condIdx int,
	typeDropdown *gtk.DropDown,
	patternEntry *gtk.Entry,
	actionBtn *gtk.Button,
) {
	pattern := patternEntry.Text()
	condType := indexToConditionType(typeDropdown.Selected())

	err := validateConditionPattern(condType, pattern)
	if err != nil {
		patternEntry.AddCSSClass("error")
	} else {
		patternEntry.RemoveCSSClass("error")
	}

	actionBtn.SetSensitive(areAllConditionsValid(*conditions))
}
