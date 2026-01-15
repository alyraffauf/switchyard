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
	nameGroup.SetDescription("Optional friendly name for this rule")

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
	conditionsGroup.SetDescription("Define one or more conditions to match URLs")

	conditionsListBox := gtk.NewListBox()
	conditionsListBox.SetSelectionMode(gtk.SelectionNone)
	conditionsListBox.AddCSSClass("boxed-list")

	// Logic selector row
	logicRow = adw.NewComboRow()
	logicRow.SetTitle("Match")
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

		// Update action button state
		actionBtn.SetSensitive(areAllConditionsValid(*conditions))
	}

	rebuildConditions()
	conditionsGroup.Add(conditionsListBox)
	content.Append(conditionsGroup)

	// Add condition button
	addConditionGroup := adw.NewPreferencesGroup()
	addConditionRow := adw.NewButtonRow()
	addConditionRow.SetTitle("Add Condition")
	addConditionRow.SetStartIconName("list-add-symbolic")
	addConditionRow.ConnectActivated(func() {
		*conditions = append(*conditions, Condition{Type: "domain", Pattern: ""})
		rebuildConditions()
	})
	addConditionGroup.Add(addConditionRow)
	content.Append(addConditionGroup)

	// Action section
	actionGroup := adw.NewPreferencesGroup()
	actionGroup.SetTitle("Actions")
	actionGroup.SetDescription("Choose which browser to use")

	alwaysAskRow = adw.NewSwitchRow()
	alwaysAskRow.SetTitle("Always ask")
	alwaysAskRow.SetSubtitle("Show browser picker for this rule")
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

	conditionContainer := gtk.NewBox(gtk.OrientationHorizontal, 6)
	conditionContainer.SetMarginTop(12)
	conditionContainer.SetMarginBottom(12)
	conditionContainer.SetMarginStart(12)
	conditionContainer.SetMarginEnd(12)

	// Inner listbox for type and pattern fields
	innerListBox := gtk.NewListBox()
	innerListBox.SetSelectionMode(gtk.SelectionNone)
	innerListBox.AddCSSClass("boxed-list")
	innerListBox.SetHExpand(true)

	// Type row
	typeRow := adw.NewComboRow()
	typeRow.SetTitle("Match type")
	typeRow.SetModel(gtk.NewStringList([]string{"Domain", "Contains", "Wildcard", "Regex"}))
	typeRow.SetSelected(conditionTypeToIndex((*conditions)[condIdx].Type))
	innerListBox.Append(typeRow)

	// Pattern row
	patternRow := adw.NewEntryRow()
	patternRow.SetTitle("Pattern")
	patternRow.SetText((*conditions)[condIdx].Pattern)
	innerListBox.Append(patternRow)

	// Connect handlers after both widgets are created
	typeRow.Connect("notify::selected", func() {
		(*conditions)[condIdx].Type = indexToConditionType(typeRow.Selected())
		validateAndUpdateCondition(conditions, condIdx, typeRow, patternRow, actionBtn)
	})

	patternRow.Connect("changed", func() {
		(*conditions)[condIdx].Pattern = patternRow.Text()
		validateAndUpdateCondition(conditions, condIdx, typeRow, patternRow, actionBtn)
	})

	conditionContainer.Append(innerListBox)

	// Buttons box
	btnBox := gtk.NewBox(gtk.OrientationVertical, 3)
	btnBox.SetVAlign(gtk.AlignCenter)

	// Move up button
	upBtn := gtk.NewButton()
	upBtn.SetIconName("go-up-symbolic")
	upBtn.SetTooltipText("Move condition up")
	upBtn.AddCSSClass("flat")
	upBtn.AddCSSClass("circular")
	upBtn.SetSensitive(condIdx > 0)
	upBtn.ConnectClicked(func() {
		if condIdx > 0 && condIdx < len(*conditions) {
			(*conditions)[condIdx], (*conditions)[condIdx-1] = (*conditions)[condIdx-1], (*conditions)[condIdx]
			rebuildConditions()
		}
	})
	btnBox.Append(upBtn)

	// Delete button
	deleteBtn := gtk.NewButton()
	deleteBtn.SetIconName("edit-delete-symbolic")
	deleteBtn.SetTooltipText("Delete this condition")
	deleteBtn.AddCSSClass("flat")
	deleteBtn.AddCSSClass("circular")
	deleteBtn.AddCSSClass("destructive-action")
	deleteBtn.SetSensitive(len(*conditions) > 1)
	deleteBtn.ConnectClicked(func() {
		if len(*conditions) > 1 && condIdx < len(*conditions) {
			*conditions = append((*conditions)[:condIdx], (*conditions)[condIdx+1:]...)
			rebuildConditions()
		}
	})
	btnBox.Append(deleteBtn)

	// Move down button
	downBtn := gtk.NewButton()
	downBtn.SetIconName("go-down-symbolic")
	downBtn.SetTooltipText("Move condition down")
	downBtn.AddCSSClass("flat")
	downBtn.AddCSSClass("circular")
	downBtn.SetSensitive(condIdx < len(*conditions)-1)
	downBtn.ConnectClicked(func() {
		if condIdx >= 0 && condIdx < len(*conditions)-1 {
			(*conditions)[condIdx], (*conditions)[condIdx+1] = (*conditions)[condIdx+1], (*conditions)[condIdx]
			rebuildConditions()
		}
	})
	btnBox.Append(downBtn)

	conditionContainer.Append(btnBox)
	conditionRow.SetChild(conditionContainer)

	return conditionRow
}

// validateAndUpdateCondition validates a pattern and updates UI accordingly
func validateAndUpdateCondition(
	conditions *[]Condition,
	condIdx int,
	typeRow *adw.ComboRow,
	patternRow *adw.EntryRow,
	actionBtn *gtk.Button,
) {
	pattern := patternRow.Text()
	condType := indexToConditionType(typeRow.Selected())

	err := validateConditionPattern(condType, pattern)
	if err != nil {
		patternRow.AddCSSClass("error")
	} else {
		patternRow.RemoveCSSClass("error")
	}

	actionBtn.SetSensitive(areAllConditionsValid(*conditions))
}
