// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"github.com/alyraffauf/switchyard/internal/routing"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func buildRuleDialogContent(
	initialRule *Rule,
	browsers []*Browser,
	actionButton *gtk.Button,
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

	var conditionsSlice []Condition
	if initialRule != nil && len(initialRule.Conditions) > 0 {
		conditionsSlice = make([]Condition, len(initialRule.Conditions))
		copy(conditionsSlice, initialRule.Conditions)
	} else {
		conditionsSlice = []Condition{{Type: "domain", Pattern: ""}}
	}
	conditions = &conditionsSlice

	conditionsGroup := adw.NewPreferencesGroup()
	conditionsGroup.SetTitle("Conditions")
	conditionsGroup.SetDescription("Define conditions to match URLs")

	conditionsListBox := createBoxedListBox()

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
	var addConditionRow *adw.ActionRow
	var rebuildConditions func()

	rebuildConditions = func() {
		for _, row := range conditionRows {
			conditionsListBox.Remove(row)
		}
		conditionRows = nil

		if addConditionRow != nil {
			conditionsListBox.Remove(addConditionRow)
		}

		for i := range *conditions {
			conditionIndex := i
			row := createConditionRow(
				conditions,
				conditionIndex,
				actionButton,
				rebuildConditions,
			)
			conditionsListBox.Append(row)
			conditionRows = append(conditionRows, row)
		}

		addConditionRow = adw.NewActionRow()
		addConditionRow.SetTitle("Add Condition")
		addConditionRow.AddPrefix(gtk.NewImageFromIconName("list-add-symbolic"))
		addConditionRow.SetActivatable(true)
		addConditionRow.ConnectActivated(func() {
			*conditions = append(*conditions, Condition{Type: "domain", Pattern: ""})
			rebuildConditions()
		})
		conditionsListBox.Append(addConditionRow)

		actionButton.SetSensitive(routing.AreAllConditionsValid(*conditions))
	}

	rebuildConditions()
	conditionsGroup.Add(conditionsListBox)
	content.Append(conditionsGroup)

	actionGroup := adw.NewPreferencesGroup()
	actionGroup.SetTitle("Browser Action")
	actionGroup.SetDescription("Select which browser opens matching URLs")

	alwaysAskRow = adw.NewSwitchRow()
	alwaysAskRow.SetTitle("Always show launcher")
	alwaysAskRow.SetSubtitle("Ask which browser to use each time")
	if initialRule != nil {
		alwaysAskRow.SetActive(initialRule.AlwaysAsk)
	}
	actionGroup.Add(alwaysAskRow)

	browserNames := make([]string, len(browsers))
	selectedIndex := uint(0)
	for i, browser := range browsers {
		browserNames[i] = browser.Name
		if initialRule != nil && browser.ID == initialRule.Browser {
			selectedIndex = uint(i)
		}
	}

	browserRow = adw.NewComboRow()
	browserRow.SetTitle("Browser")
	browserRow.SetModel(gtk.NewStringList(browserNames))
	browserRow.SetSelected(selectedIndex)
	if initialRule != nil {
		browserRow.SetSensitive(!initialRule.AlwaysAsk)
	}
	actionGroup.Add(browserRow)

	alwaysAskRow.Connect("notify::active", func() {
		browserRow.SetSensitive(!alwaysAskRow.Active())
	})

	content.Append(actionGroup)

	return
}

func createConditionRow(
	conditions *[]Condition,
	conditionIndex int,
	actionButton *gtk.Button,
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

	typeDropdown := gtk.NewDropDown(
		gtk.NewStringList(routing.ConditionTypeLabels()),
		nil,
	)
	typeDropdown.SetSelected(routing.ConditionTypeToIndex((*conditions)[conditionIndex].Type))
	typeDropdown.SetVAlign(gtk.AlignCenter)
	typeDropdown.SetSizeRequest(150, -1)
	conditionContainer.Append(typeDropdown)

	negateDropdown := gtk.NewDropDown(
		gtk.NewStringList([]string{"is", "is not"}),
		nil,
	)
	if (*conditions)[conditionIndex].Negate {
		negateDropdown.SetSelected(1)
	} else {
		negateDropdown.SetSelected(0)
	}
	negateDropdown.SetVAlign(gtk.AlignCenter)
	negateDropdown.Connect("notify::selected", func() {
		(*conditions)[conditionIndex].Negate = negateDropdown.Selected() == 1
	})
	conditionContainer.Append(negateDropdown)

	patternEntry := gtk.NewEntry()
	patternEntry.SetText((*conditions)[conditionIndex].Pattern)
	patternEntry.SetHExpand(true)
	patternEntry.SetPlaceholderText("Pattern")
	conditionContainer.Append(patternEntry)

	typeDropdown.Connect("notify::selected", func() {
		(*conditions)[conditionIndex].Type = routing.IndexToConditionType(typeDropdown.Selected())
		validateConditionEntry(conditions, conditionIndex, typeDropdown, patternEntry, actionButton)
	})

	patternEntry.Connect("changed", func() {
		(*conditions)[conditionIndex].Pattern = patternEntry.Text()
		validateConditionEntry(conditions, conditionIndex, typeDropdown, patternEntry, actionButton)
	})

	deleteButton := gtk.NewButton()
	deleteButton.SetIconName("edit-delete-symbolic")
	deleteButton.SetTooltipText("Delete this condition")
	deleteButton.AddCSSClass("flat")
	deleteButton.AddCSSClass("circular")
	deleteButton.AddCSSClass("destructive-action")
	deleteButton.SetVAlign(gtk.AlignCenter)
	deleteButton.SetSensitive(len(*conditions) > 1)
	deleteButton.ConnectClicked(func() {
		if len(*conditions) > 1 && conditionIndex < len(*conditions) {
			*conditions = append((*conditions)[:conditionIndex], (*conditions)[conditionIndex+1:]...)
			rebuildConditions()
		}
	})
	conditionContainer.Append(deleteButton)

	conditionRow.SetChild(conditionContainer)
	return conditionRow
}

func validateConditionEntry(
	conditions *[]Condition,
	conditionIndex int,
	typeDropdown *gtk.DropDown,
	patternEntry *gtk.Entry,
	actionButton *gtk.Button,
) {
	pattern := patternEntry.Text()
	conditionType := routing.IndexToConditionType(typeDropdown.Selected())

	err := routing.ValidateConditionPattern(conditionType, pattern)
	if err != nil {
		patternEntry.AddCSSClass("error")
	} else {
		patternEntry.RemoveCSSClass("error")
	}

	actionButton.SetSensitive(routing.AreAllConditionsValid(*conditions))
}
