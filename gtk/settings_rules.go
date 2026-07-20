// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"html"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
	"github.com/alyraffauf/switchyard/internal/routing"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func createRulesPage(win *adw.Window, browsers []*Browser, cfg *Config) gtk.Widgetter {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Browser Rules")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)

	addButton := gtk.NewButton()
	addButton.SetIconName("list-add-symbolic")
	addButton.SetTooltipText("Add Rule")
	addButton.SetHasFrame(false)
	header.PackEnd(addButton)

	toolbarView.AddTopBar(header)

	scrolled := createScrolledWindow()

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)
	content.SetMarginTop(12)
	content.SetMarginBottom(12)

	infoLabel := gtk.NewLabel("Rules route links to browsers. First match wins.")
	infoLabel.SetWrap(true)
	infoLabel.SetXAlign(0)
	infoLabel.AddCSSClass("dim-label")
	infoLabel.SetMarginStart(12)
	infoLabel.SetMarginEnd(12)
	infoLabel.SetMarginBottom(6)
	content.Append(infoLabel)

	rulesListBox := createBoxedListBox()
	emptyState := createEmptyState("view-list-symbolic", "No Browser Rules", "Add rules to automatically open links in specific browsers")

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
			row.SetSubtitle(html.EscapeString(routing.FormatRuleSubtitle(rule, getBrowserName(rule.Browser))))
		} else {
			if len(rule.Conditions) > 0 {
				row.SetTitle(rule.Conditions[0].Pattern)
			}
			row.SetSubtitle(html.EscapeString(routing.FormatRuleSubtitleNoPattern(rule, getBrowserName(rule.Browser))))
		}
		row.SetActivatable(true)

		var icon *gtk.Image
		if rule.AlwaysAsk {
			appBrowser := &Browser{
				Browser: &appbrowser.Browser{
					ID:   getAppID(),
					Icon: getAppID(),
				},
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
				saveConfigWithFlag(cfg)
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
				saveConfigWithFlag(cfg)
				rebuildRulesList()
			}
		})
		reorderBox.Append(downBtn)

		row.AddSuffix(reorderBox)

		deleteBtn := gtk.NewButton()
		deleteBtn.SetIconName("edit-delete-symbolic")
		deleteBtn.AddCSSClass("flat")
		deleteBtn.AddCSSClass("destructive-action")
		deleteBtn.SetTooltipText("Delete rule")
		deleteBtn.ConnectClicked(func() {
			cfg.Rules = append(cfg.Rules[:ruleIndex], cfg.Rules[ruleIndex+1:]...)
			saveConfigWithFlag(cfg)
			rebuildRulesList()
		})
		row.AddSuffix(deleteBtn)

		row.ConnectActivated(func() {
			showEditRuleDialog(win, cfg, rule, browsers, rebuildRulesList)
		})

		return row
	}

	rebuildRulesList = func() {
		clearListBox(rulesListBox)

		// Swap between the empty-state page and the populated list.
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

	rebuildRulesList()

	content.Append(rulesListBox)
	content.Append(emptyState)
	scrolled.SetChild(content)
	toolbarView.SetContent(scrolled)

	addButton.ConnectClicked(func() {
		showAddRuleDialog(win, cfg, browsers, rebuildRulesList)
	})

	return toolbarView
}
