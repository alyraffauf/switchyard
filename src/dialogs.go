// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"regexp"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showDefaultBrowserPrompt displays a dialog asking to set Switchyard as default browser
func showDefaultBrowserPrompt(parent gtk.Widgetter, cfg *Config, updateUI func()) {
	dialog := adw.NewAlertDialog(
		"Set as Default Browser?",
		"Switchyard is not your default browser. Would you like to set it as the default browser now?",
	)

	dialog.AddResponse("no", "Don't Ask Again")
	dialog.AddResponse("later", "Not Now")
	dialog.AddResponse("yes", "Set as Default")

	dialog.SetDefaultResponse("yes")
	dialog.SetCloseResponse("later")

	dialog.ConnectResponse(func(response string) {
		if response == "yes" {
			setAsDefaultBrowser()
			cfg.CheckDefaultBrowser = false
			saveConfig(cfg)
			updateUI()
		} else if response == "no" {
			cfg.CheckDefaultBrowser = false
			saveConfig(cfg)
			updateUI()
		}
	})

	dialog.Present(parent)
}

// showAboutDialog displays the about dialog
func showAboutDialog(parent *adw.Window) {
	dialog := adw.NewDialog()
	dialog.SetTitle("")
	dialog.SetContentWidth(350)
	dialog.SetContentHeight(450)

	toolbarView := adw.NewToolbarView()

	// Header with close button
	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(true)
	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(12)
	content.SetMarginBottom(24)
	content.SetVAlign(gtk.AlignCenter)
	content.SetHAlign(gtk.AlignCenter)

	// App icon
	icon := gtk.NewImageFromIconName(appID)
	icon.SetPixelSize(96)
	content.Append(icon)

	// App name
	nameLabel := gtk.NewLabel("Switchyard")
	nameLabel.AddCSSClass("title-1")
	content.Append(nameLabel)

	// Developer
	devLabel := gtk.NewLabel("Aly Raffauf")
	devLabel.AddCSSClass("dim-label")
	content.Append(devLabel)

	// Version badge
	versionBtn := gtk.NewButton()
	versionBtn.SetLabel("0.1.0")
	versionBtn.AddCSSClass("pill")
	versionBtn.AddCSSClass("suggested-action")
	versionBtn.SetCanFocus(false)
	versionBtn.SetSensitive(false)
	content.Append(versionBtn)

	// Spacer
	spacer := gtk.NewBox(gtk.OrientationVertical, 0)
	spacer.SetVExpand(true)
	content.Append(spacer)

	// Links group
	linksGroup := adw.NewPreferencesGroup()

	websiteRow := adw.NewActionRow()
	websiteRow.SetTitle("Website")
	websiteRow.SetActivatable(true)
	websiteRow.AddSuffix(gtk.NewImageFromIconName("external-link-symbolic"))
	websiteRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher("https://github.com/alyraffauf/switchyard")
		launcher.Launch(context.Background(), &parent.Window, nil)
	})
	linksGroup.Add(websiteRow)

	issueRow := adw.NewActionRow()
	issueRow.SetTitle("Report an Issue")
	issueRow.SetActivatable(true)
	issueRow.AddSuffix(gtk.NewImageFromIconName("external-link-symbolic"))
	issueRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher("https://github.com/alyraffauf/switchyard/issues")
		launcher.Launch(context.Background(), &parent.Window, nil)
	})
	linksGroup.Add(issueRow)

	content.Append(linksGroup)

	// Copyright and disclaimer
	copyrightLabel := gtk.NewLabel("© 2026 Aly Raffauf")
	copyrightLabel.AddCSSClass("dim-label")
	content.Append(copyrightLabel)

	disclaimerLabel := gtk.NewLabel("This application comes with absolutely no warranty.\nSee the GNU General Public License, version 3 or later for details.")
	disclaimerLabel.AddCSSClass("dim-label")
	disclaimerLabel.AddCSSClass("caption")
	disclaimerLabel.SetJustify(gtk.JustifyCenter)
	disclaimerLabel.SetWrap(true)
	content.Append(disclaimerLabel)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)
	dialog.Present(parent)
}

// showAddRuleDialog displays the add rule dialog
func showAddRuleDialog(parent *adw.Window, cfg *Config, browsers []*Browser, getBrowserName func(string) string, getBrowserIcon func(string) string, rebuildRulesList func()) {
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
	cancelBtn.SetTooltipText("Cancel and close")
	cancelBtn.ConnectClicked(func() { dialog.Close() })
	header.PackStart(cancelBtn)

	addBtn := gtk.NewButton()
	addBtn.SetLabel("Add")
	addBtn.AddCSSClass("suggested-action")
	addBtn.SetSensitive(false) // Insensitive until at least one condition is added
	addBtn.SetTooltipText("Add this rule")
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

// showEditRuleDialog displays the edit rule dialog
func showEditRuleDialog(parent *adw.Window, cfg *Config, rule *Rule, row *adw.ActionRow, browsers []*Browser, getBrowserName func(string) string, getBrowserIcon func(string) string, rebuildRulesList func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Edit Rule")
	dialog.SetContentWidth(600)
	dialog.SetContentHeight(650)
	dialog.SetCanClose(true)

	// Ensure rules have at least one condition
	if len(rule.Conditions) == 0 {
		rule.Conditions = []Condition{{
			Type:    "domain",
			Pattern: "",
		}}
		rule.Logic = "all"
	}

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

// showEditConditionDialog displays a dialog to edit a single condition
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
	switch cond.Type {
	case "domain":
		typeRow.SetSelected(0)
	case "keyword":
		typeRow.SetSelected(1)
	case "glob":
		typeRow.SetSelected(2)
	case "regex":
		typeRow.SetSelected(3)
	}
	group.Add(typeRow)

	// Pattern row
	patternRow := adw.NewEntryRow()
	patternRow.SetTitle("Pattern")
	patternRow.SetText(cond.Pattern)
	group.Add(patternRow)

	// Error label for regex validation
	errorLabel := gtk.NewLabel("")
	errorLabel.SetWrap(true)
	errorLabel.AddCSSClass("error")
	errorLabel.SetVisible(false)

	// Function to validate and update error display
	updateValidation := func() {
		var condType string
		switch typeRow.Selected() {
		case 0:
			condType = "domain"
		case 1:
			condType = "keyword"
		case 2:
			condType = "glob"
		case 3:
			condType = "regex"
		}

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
		switch selectedType {
		case 0:
			cond.Type = "domain"
		case 1:
			cond.Type = "keyword"
		case 2:
			cond.Type = "glob"
		case 3:
			cond.Type = "regex"
		}
		cond.Pattern = pattern

		onSave()
		dialog.Close()
	})

	dialog.Present(parent)
}

// buildRuleDialogContent creates the common UI content for add/edit rule dialogs
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

	// Conditions section
	conditionsGroup := adw.NewPreferencesGroup()
	conditionsGroup.SetTitle("Conditions")
	conditionsGroup.SetDescription("Define one or more conditions to match URLs")

	// Store conditions in a slice
	var conditionsSlice []Condition
	if initialRule != nil && len(initialRule.Conditions) > 0 {
		conditionsSlice = make([]Condition, len(initialRule.Conditions))
		copy(conditionsSlice, initialRule.Conditions)
	} else {
		conditionsSlice = []Condition{{Type: "domain", Pattern: ""}}
	}
	conditions = &conditionsSlice

	// Create a single ListBox to hold Match selector and all conditions
	conditionsListBox := gtk.NewListBox()
	conditionsListBox.SetSelectionMode(gtk.SelectionNone)
	conditionsListBox.AddCSSClass("boxed-list")

	// Logic selector row (All/Any) as a ComboRow
	logicRow = adw.NewComboRow()
	logicRow.SetTitle("Match")
	logicRow.SetModel(gtk.NewStringList([]string{"All conditions", "Any condition"}))
	if initialRule != nil && initialRule.Logic == "any" {
		logicRow.SetSelected(1)
	} else {
		logicRow.SetSelected(0)
	}

	conditionsListBox.Append(logicRow)

	// Track condition rows separately so we can rebuild them
	var conditionRows []*gtk.ListBoxRow

	// Declare function variable first to allow recursive calls
	var rebuildConditions func()

	// Function to rebuild the conditions list UI
	rebuildConditions = func() {
		// Clear all condition rows
		for _, row := range conditionRows {
			conditionsListBox.Remove(row)
		}
		conditionRows = nil

		// Add each condition as a grouped unit
		for i := range *conditions {
			condIdx := i // Capture for closure

			// Create a ListBoxRow to properly handle spacing
			conditionRow := gtk.NewListBoxRow()
			conditionRow.SetActivatable(false)
			conditionRow.SetSelectable(false)

			// Create a box to group this condition's fields + buttons
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
			// Set current value
			switch (*conditions)[condIdx].Type {
			case "domain":
				typeRow.SetSelected(0)
			case "keyword":
				typeRow.SetSelected(1)
			case "glob":
				typeRow.SetSelected(2)
			case "regex":
				typeRow.SetSelected(3)
			}
			typeRow.Connect("notify::selected", func() {
				switch typeRow.Selected() {
				case 0:
					(*conditions)[condIdx].Type = "domain"
				case 1:
					(*conditions)[condIdx].Type = "keyword"
				case 2:
					(*conditions)[condIdx].Type = "glob"
				case 3:
					(*conditions)[condIdx].Type = "regex"
				}
			})
			innerListBox.Append(typeRow)

			// Pattern row
			patternRow := adw.NewEntryRow()
			patternRow.SetTitle("Pattern")
			patternRow.SetText((*conditions)[condIdx].Pattern)

			// Function to validate pattern (especially for regex)
			validatePattern := func() {
				pattern := patternRow.Text()
				condType := (*conditions)[condIdx].Type

				err := validateConditionPattern(condType, pattern)

				if err != nil {
					patternRow.AddCSSClass("error")
				} else {
					patternRow.RemoveCSSClass("error")
				}

				// Enable/disable action button based on all conditions validity
				actionBtn.SetSensitive(areAllConditionsValid(*conditions))
			}

			patternRow.Connect("changed", func() {
				(*conditions)[condIdx].Pattern = patternRow.Text()
				validatePattern()
			})

			typeRow.Connect("notify::selected", func() {
				validatePattern()
			})

			innerListBox.Append(patternRow)

			conditionContainer.Append(innerListBox)

			// Buttons box on the right
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
					// Swap with previous
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
					// Swap with next
					(*conditions)[condIdx], (*conditions)[condIdx+1] = (*conditions)[condIdx+1], (*conditions)[condIdx]
					rebuildConditions()
				}
			})
			btnBox.Append(downBtn)

			conditionContainer.Append(btnBox)
			conditionRow.SetChild(conditionContainer)
			conditionsListBox.Append(conditionRow)
			conditionRows = append(conditionRows, conditionRow)
		}

		// Update action button sensitivity
		allValid := len(*conditions) > 0
		for _, c := range *conditions {
			if c.Pattern == "" {
				allValid = false
				break
			}
		}
		actionBtn.SetSensitive(allValid)
	}

	// Initialize conditions UI
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

	// Always Ask toggle
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

	// Make browser row sensitive based on always ask toggle
	alwaysAskRow.Connect("notify::active", func() {
		browserRow.SetSensitive(!alwaysAskRow.Active())
	})

	content.Append(actionGroup)

	return
}
