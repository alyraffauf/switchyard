// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// formatRuleSubtitle formats a subtitle for a rule row with pattern included
func formatRuleSubtitle(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, true)
}

// formatRuleSubtitleNoPattern formats a subtitle for a rule row without pattern
func formatRuleSubtitleNoPattern(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, false)
}

func formatRuleSubtitleInternal(rule *Rule, browserName string, includePattern bool) string {
	if len(rule.Conditions) > 0 {
		condCount := len(rule.Conditions)
		var logicText string
		if rule.Logic == "any" {
			logicText = "Any match"
		} else {
			logicText = "All match"
		}

		formatSingleCondition := func(cond *Condition) string {
			return fmt.Sprintf("%s %s", getConditionLabel(cond.Type, cond.Negate), cond.Pattern)
		}

		if rule.AlwaysAsk {
			if condCount == 1 && includePattern {
				return fmt.Sprintf("%s · Always ask", formatSingleCondition(&rule.Conditions[0]))
			}
			return fmt.Sprintf("%d conditions (%s) · Always ask", condCount, logicText)
		}
		if condCount == 1 && includePattern {
			return fmt.Sprintf("%s · Opens in %s", formatSingleCondition(&rule.Conditions[0]), browserName)
		}
		return fmt.Sprintf("%d conditions (%s) · Opens in %s", condCount, logicText, browserName)
	}

	return "No conditions"
}

func getConditionLabel(patternType string, negate bool) string {
	switch patternType {
	case "domain":
		if negate {
			return "Domain is not"
		}
		return "Domain is"
	case "keyword":
		if negate {
			return "URL does not contain"
		}
		return "URL contains"
	case "glob":
		if negate {
			return "Wildcard is not"
		}
		return "Wildcard is"
	case "regex":
		if negate {
			return "Regex does not match"
		}
		return "Regex matches"
	default:
		return patternType
	}
}

func createRulesPage(win *adw.Window, browsers []*Browser, cfg *Config) gtk.Widgetter {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Browser Rules")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)

	// Add Rule button in header
	addButton := gtk.NewButton()
	addButton.SetIconName("list-add-symbolic")
	addButton.SetTooltipText("Add Rule")
	addButton.SetHasFrame(false)
	header.PackEnd(addButton)

	toolbarView.AddTopBar(header)

	// Scrolled window for rules list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)
	content.SetMarginTop(12)
	content.SetMarginBottom(12)

	// Info banner
	infoLabel := gtk.NewLabel("Rules route links to browsers. First match wins.")
	infoLabel.SetWrap(true)
	infoLabel.SetXAlign(0)
	infoLabel.AddCSSClass("dim-label")
	infoLabel.SetMarginStart(12)
	infoLabel.SetMarginEnd(12)
	infoLabel.SetMarginBottom(6)
	content.Append(infoLabel)

	// Rules list
	rulesListBox := gtk.NewListBox()
	rulesListBox.SetSelectionMode(gtk.SelectionNone)
	rulesListBox.AddCSSClass("boxed-list")

	// Empty state
	emptyState := adw.NewStatusPage()
	emptyState.SetIconName("list-add-symbolic")
	emptyState.SetTitle("No Browser Rules")
	emptyState.SetDescription("Add rules to automatically open links in specific browsers")
	emptyState.SetVExpand(true)

	// Helper to get browser name from ID
	getBrowserName := func(id string) string {
		if browser := findBrowserByID(browsers, id); browser != nil {
			return browser.Name
		}
		return id
	}

	var rebuildRulesList func()

	createRuleRow := func(ruleIndex int) *adw.ActionRow {
		rule := &cfg.Rules[ruleIndex]

		row := adw.NewActionRow()
		if rule.Name != "" {
			row.SetTitle(rule.Name)
			row.SetSubtitle(formatRuleSubtitle(rule, getBrowserName(rule.Browser)))
		} else {
			if len(rule.Conditions) > 0 {
				row.SetTitle(rule.Conditions[0].Pattern)
			}
			row.SetSubtitle(formatRuleSubtitleNoPattern(rule, getBrowserName(rule.Browser)))
		}
		row.SetActivatable(true)

		// Browser icon
		var icon *gtk.Image
		if rule.AlwaysAsk {
			appBrowser := &Browser{
				ID:      getAppID(),
				Icon:    getAppID(),
				AppInfo: nil,
			}
			icon = loadBrowserIcon(appBrowser, 24)
		} else {
			browser := findBrowserByID(browsers, rule.Browser)
			if browser != nil {
				icon = loadBrowserIcon(browser, 24)
			} else {
				icon = gtk.NewImageFromIconName("web-browser-symbolic")
				icon.SetPixelSize(24)
			}
		}
		row.AddPrefix(icon)

		// Reorder buttons
		reorderBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
		reorderBox.SetVAlign(gtk.AlignCenter)

		upBtn := gtk.NewButton()
		upBtn.SetIconName("go-up-symbolic")
		upBtn.AddCSSClass("flat")
		upBtn.SetSensitive(ruleIndex > 0)
		upBtn.SetTooltipText("Move rule up")
		upBtn.ConnectClicked(func() {
			if ruleIndex > 0 {
				cfg.Rules[ruleIndex], cfg.Rules[ruleIndex-1] = cfg.Rules[ruleIndex-1], cfg.Rules[ruleIndex]
				saveConfig(cfg)
				rebuildRulesList()
			}
		})
		reorderBox.Append(upBtn)

		downBtn := gtk.NewButton()
		downBtn.SetIconName("go-down-symbolic")
		downBtn.AddCSSClass("flat")
		downBtn.SetSensitive(ruleIndex < len(cfg.Rules)-1)
		downBtn.SetTooltipText("Move rule down")
		downBtn.ConnectClicked(func() {
			if ruleIndex < len(cfg.Rules)-1 {
				cfg.Rules[ruleIndex], cfg.Rules[ruleIndex+1] = cfg.Rules[ruleIndex+1], cfg.Rules[ruleIndex]
				saveConfig(cfg)
				rebuildRulesList()
			}
		})
		reorderBox.Append(downBtn)

		row.AddSuffix(reorderBox)

		// Delete button
		deleteBtn := gtk.NewButton()
		deleteBtn.SetIconName("edit-delete-symbolic")
		deleteBtn.AddCSSClass("flat")
		deleteBtn.AddCSSClass("destructive-action")
		deleteBtn.SetTooltipText("Delete rule")
		deleteBtn.ConnectClicked(func() {
			cfg.Rules = append(cfg.Rules[:ruleIndex], cfg.Rules[ruleIndex+1:]...)
			saveConfig(cfg)
			rebuildRulesList()
		})
		row.AddSuffix(deleteBtn)

		// Edit on click
		row.ConnectActivated(func() {
			showEditRuleDialog(win, cfg, rule, browsers, rebuildRulesList)
		})

		return row
	}

	rebuildRulesList = func() {
		// Remove all children
		for {
			child := rulesListBox.FirstChild()
			if child == nil {
				break
			}
			rulesListBox.Remove(child)
		}

		// Show/hide empty state vs rules list
		if len(cfg.Rules) == 0 {
			infoLabel.SetVisible(false)
			rulesListBox.SetVisible(false)
			emptyState.SetVisible(true)
		} else {
			infoLabel.SetVisible(true)
			rulesListBox.SetVisible(true)
			emptyState.SetVisible(false)

			for i := range cfg.Rules {
				row := createRuleRow(i)
				rulesListBox.Append(row)
			}
		}
	}

	// Initial build
	rebuildRulesList()

	content.Append(rulesListBox)
	content.Append(emptyState)
	scrolled.SetChild(content)
	toolbarView.SetContent(scrolled)

	// Connect Add Rule button handler
	addButton.ConnectClicked(func() {
		showAddRuleDialog(win, cfg, browsers, rebuildRulesList)
	})

	return toolbarView
}
