// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

// showPickerWindow displays the browser picker window
func showPickerWindow(app *adw.Application, url string, browsers []*Browser) {
	// Sort browsers alphabetically by name
	sortedBrowsers := make([]*Browser, len(browsers))
	copy(sortedBrowsers, browsers)
	sort.Slice(sortedBrowsers, func(i, j int) bool {
		return sortedBrowsers[i].Name < sortedBrowsers[j].Name
	})

	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetDefaultSize(650, -1)
	win.SetResizable(false)
	win.SetApplication(&app.Application)

	// Main layout
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	toolbarView.AddTopBar(header)

	// Content box with margins
	contentBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentBox.SetMarginStart(24)
	contentBox.SetMarginEnd(24)
	contentBox.SetMarginTop(24)
	contentBox.SetMarginBottom(24)

	// FlowBox for browser buttons - wraps to multiple rows
	flowBox := gtk.NewFlowBox()
	flowBox.SetSelectionMode(gtk.SelectionNone)
	flowBox.SetHomogeneous(true)
	flowBox.SetColumnSpacing(16)
	flowBox.SetRowSpacing(16)
	flowBox.SetMaxChildrenPerLine(4)
	flowBox.SetHAlign(gtk.AlignCenter)
	flowBox.SetVAlign(gtk.AlignStart)

	for i, browser := range sortedBrowsers {
		b := browser // capture
		idx := i

		// Button for each browser
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		btn.SetSizeRequest(140, -1)

		// Container inside button - icon above, name and number below
		btnBox := gtk.NewBox(gtk.OrientationVertical, 8)
		btnBox.SetHAlign(gtk.AlignCenter)
		btnBox.SetVAlign(gtk.AlignCenter)

		// Fixed-size container for icon to ensure uniform sizing
		iconBox := gtk.NewBox(gtk.OrientationVertical, 0)
		iconBox.SetSizeRequest(64, 64)
		iconBox.SetHAlign(gtk.AlignCenter)
		iconBox.SetVAlign(gtk.AlignCenter)

		// Large browser icon - use helper to load with fallback
		icon := loadBrowserIcon(b.Icon, 64)
		icon.SetHAlign(gtk.AlignCenter)
		icon.SetVAlign(gtk.AlignCenter)
		iconBox.Append(icon)

		btnBox.Append(iconBox)

		// Browser name - single line with ellipsis
		label := gtk.NewLabel(b.Name)
		label.SetEllipsize(pango.EllipsizeEnd)
		label.SetMaxWidthChars(18)
		label.SetJustify(gtk.JustifyCenter)
		label.SetLines(1)
		btnBox.Append(label)

		// Number shortcut (1-9)
		if idx < 9 {
			shortcutLabel := gtk.NewLabel(fmt.Sprintf("%d", idx+1))
			shortcutLabel.AddCSSClass("dim-label")
			shortcutLabel.AddCSSClass("caption")
			btnBox.Append(shortcutLabel)
		}

		btn.SetChild(btnBox)

		btn.ConnectClicked(func() {
			launchBrowser(b, url)
			win.Close()
		})

		flowBox.Insert(btn, -1)
	}

	contentBox.Append(flowBox)

	// Add URL display at bottom
	urlBox := gtk.NewBox(gtk.OrientationVertical, 0)
	urlBox.SetMarginTop(16)

	// URL entry (read-only, centered text)
	urlEntry := gtk.NewEntry()
	urlEntry.SetText(url)
	urlEntry.SetEditable(false)
	urlEntry.SetCanFocus(false)
	urlEntry.SetHAlign(gtk.AlignCenter)
	urlEntry.SetAlignment(0.5) // Center text within the entry
	urlEntry.SetMaxWidthChars(60)
	urlBox.Append(urlEntry)

	contentBox.Append(urlBox)
	toolbarView.SetContent(contentBox)
	win.SetContent(toolbarView)

	// Keyboard shortcuts
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		// Number keys 1-9 for quick selection
		if keyval >= gdk.KEY_1 && keyval <= gdk.KEY_9 {
			idx := int(keyval - gdk.KEY_1)
			if idx < len(sortedBrowsers) {
				launchBrowser(sortedBrowsers[idx], url)
				win.Close()
				return true
			}
		}
		// Escape to close
		if keyval == gdk.KEY_Escape {
			win.Close()
			return true
		}
		return false
	})
	win.AddController(keyController)

	win.Present()
}

// showSettingsWindow displays the main settings window
func showSettingsWindow(app *adw.Application) {
	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetDefaultSize(550, 600)
	win.SetApplication(&app.Application)

	cfg := loadConfig()
	browsers := detectBrowsers()

	// Main layout with toolbar
	toolbarView := adw.NewToolbarView()

	// Header bar with hamburger menu
	header := adw.NewHeaderBar()

	// Hamburger menu
	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetTooltipText("Main menu")

	menu := gio.NewMenu()
	menu.Append("About Switchyard", "app.about")

	quitSection := gio.NewMenu()
	quitSection.Append("Quit", "app.quit")
	menu.AppendSection("", quitSection)

	menuBtn.SetMenuModel(menu)
	header.PackEnd(menuBtn)

	// Actions
	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(p *glib.Variant) {
		showAboutDialog(win)
	})
	app.AddAction(aboutAction)

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		win.Close()
	})
	app.AddAction(quitAction)

	// Keyboard shortcut for Quit
	app.SetAccelsForAction("app.quit", []string{"<Ctrl>q"})

	toolbarView.AddTopBar(header)

	// Scrolled window for content
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	// Content box
	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	// Clamp for max width
	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	// Behavior section
	behaviorGroup := adw.NewPreferencesGroup()
	behaviorGroup.SetTitle("Behavior")

	promptRow := adw.NewSwitchRow()
	promptRow.SetTitle("Prompt when no rule matches")
	promptRow.SetSubtitle("Show browser picker for URLs without matching rules")
	behaviorGroup.Add(promptRow)

	// Fallback browser dropdown
	browserNames := make([]string, len(browsers)+1)
	browserNames[0] = "None"
	for i, b := range browsers {
		browserNames[i+1] = b.Name
	}
	browserList := gtk.NewStringList(browserNames)

	defaultRow := adw.NewComboRow()
	defaultRow.SetTitle("Fallback browser")
	defaultRow.SetSubtitle("Browser to open when prompt is disabled and no rule matches")
	defaultRow.SetModel(browserList)
	behaviorGroup.Add(defaultRow)

	// Check default browser toggle
	checkDefaultRow := adw.NewSwitchRow()
	checkDefaultRow.SetTitle("Check if Switchyard is default browser")
	checkDefaultRow.SetSubtitle("Prompt to set Switchyard as system default browser on startup")
	behaviorGroup.Add(checkDefaultRow)

	content.Append(behaviorGroup)

	// Function to update UI from config
	updateUI := func() {
		promptRow.SetActive(cfg.PromptOnClick)
		defaultRow.SetSensitive(!cfg.PromptOnClick)
		checkDefaultRow.SetActive(cfg.CheckDefaultBrowser)

		// Update fallback browser selection
		defaultRow.SetSelected(0)
		for i, b := range browsers {
			if b.ID == cfg.FallbackBrowser {
				defaultRow.SetSelected(uint(i + 1))
				break
			}
		}
	}

	// Wrap saveConfig to set the global saving flag
	saveConfigSafe := func(c *Config) error {
		savingMux.Lock()
		isSaving = true
		savingMux.Unlock()
		err := saveConfig(c)
		// Add small delay to ensure file is flushed before file watcher reads it
		glib.TimeoutAdd(100, func() bool {
			savingMux.Lock()
			isSaving = false
			savingMux.Unlock()
			return false
		})
		return err
	}

	// Function to save config and update UI
	saveAndUpdate := func() {
		saveConfigSafe(cfg)
		updateUI()
	}

	// Initial UI update
	updateUI()

	// Connect change handlers
	promptRow.Connect("notify::active", func() {
		cfg.PromptOnClick = promptRow.Active()
		saveAndUpdate()
	})

	defaultRow.Connect("notify::selected", func() {
		idx := defaultRow.Selected()
		if idx == 0 {
			cfg.FallbackBrowser = ""
		} else if idx > 0 && int(idx) <= len(browsers) {
			cfg.FallbackBrowser = browsers[idx-1].ID
		}
		saveConfigSafe(cfg)
	})

	checkDefaultRow.Connect("notify::active", func() {
		cfg.CheckDefaultBrowser = checkDefaultRow.Active()
		saveConfigSafe(cfg)
	})

	// Check if we should prompt to set as default browser (after UI is created)
	if cfg.CheckDefaultBrowser && !isDefaultBrowser() {
		showDefaultBrowserPrompt(win, cfg, updateUI)
	}

	// Helper to get browser name from ID
	getBrowserName := func(id string) string {
		if browser := findBrowserByID(browsers, id); browser != nil {
			return browser.Name
		}
		return id
	}

	// Helper to get browser icon from ID
	getBrowserIcon := func(id string) string {
		if browser := findBrowserByID(browsers, id); browser != nil {
			return browser.Icon
		}
		return "web-browser-symbolic"
	}

	// Rules section
	rulesGroup := adw.NewPreferencesGroup()
	rulesGroup.SetTitle("Rules")
	rulesGroup.SetDescription("Rules are evaluated in order. First match wins.")

	// Use a ListBox for proper ordering control
	rulesListBox := gtk.NewListBox()
	rulesListBox.SetSelectionMode(gtk.SelectionNone)
	rulesListBox.AddCSSClass("boxed-list")

	// Function to rebuild the rules list UI
	var rebuildRulesList func()

	// Function to create a rule row
	createRuleRow := func(ruleIndex int) *adw.ActionRow {
		rule := &cfg.Rules[ruleIndex]

		row := adw.NewActionRow()
		// Show name as title if set, otherwise show first condition pattern
		if rule.Name != "" {
			row.SetTitle(rule.Name)
			row.SetSubtitle(formatRuleSubtitle(rule, getBrowserName(rule.Browser)))
		} else {
			// For rules without names, show first condition pattern
			if len(rule.Conditions) > 0 {
				row.SetTitle(rule.Conditions[0].Pattern)
			}
			row.SetSubtitle(formatRuleSubtitleNoPattern(rule, getBrowserName(rule.Browser)))
		}
		row.SetActivatable(true)

		// Browser icon - use Switchyard icon if AlwaysAsk is enabled
		var iconName string
		if rule.AlwaysAsk {
			iconName = appID
		} else {
			iconName = getBrowserIcon(rule.Browser)
		}
		icon := gtk.NewImageFromIconName(iconName)
		icon.SetPixelSize(24)
		row.AddPrefix(icon)

		// Reorder buttons box
		reorderBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
		reorderBox.SetVAlign(gtk.AlignCenter)

		// Move up button
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

		// Move down button
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
		deleteBtn.SetIconName("user-trash-symbolic")
		deleteBtn.AddCSSClass("flat")
		deleteBtn.SetTooltipText("Delete rule")
		deleteBtn.ConnectClicked(func() {
			cfg.Rules = append(cfg.Rules[:ruleIndex], cfg.Rules[ruleIndex+1:]...)
			saveConfig(cfg)
			rebuildRulesList()
		})
		row.AddSuffix(deleteBtn)

		// Edit on click
		row.ConnectActivated(func() {
			showEditRuleDialog(win, cfg, rule, row, browsers, getBrowserName, getBrowserIcon, rebuildRulesList)
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

		// Add all rule rows
		for i := range cfg.Rules {
			row := createRuleRow(i)
			rulesListBox.Append(row)
		}
	}

	// Initial build
	rebuildRulesList()

	rulesGroup.Add(rulesListBox)
	content.Append(rulesGroup)

	// Add rule button
	addGroup := adw.NewPreferencesGroup()
	addRow := adw.NewButtonRow()
	addRow.SetTitle("Add Rule")
	addRow.SetStartIconName("list-add-symbolic")
	addRow.ConnectActivated(func() {
		showAddRuleDialog(win, cfg, browsers, getBrowserName, getBrowserIcon, rebuildRulesList)
	})
	addGroup.Add(addRow)
	content.Append(addGroup)

	// Config file info
	infoGroup := adw.NewPreferencesGroup()
	infoGroup.SetTitle("Advanced")
	infoRow := adw.NewActionRow()
	infoRow.SetTitle("Config file")
	infoRow.SetSubtitle(configPath())
	infoRow.SetActivatable(true)
	infoRow.AddSuffix(gtk.NewImageFromIconName("document-edit-symbolic"))
	infoRow.SetTooltipText("Open config file in text editor")
	infoRow.ConnectActivated(func() {
		// Ensure config file exists
		saveConfig(cfg)
		// Open with default text editor
		gtk.ShowURI(&win.Window, "file://"+configPath(), 0)
	})
	infoGroup.Add(infoRow)
	content.Append(infoGroup)

	// Watch config file for external changes
	configFile := gio.NewFileForPath(configPath())
	monitorIface, err := configFile.Monitor(context.Background(), gio.FileMonitorNone)
	if err == nil && monitorIface != nil {
		monitor := gio.BaseFileMonitor(monitorIface)
		if monitor != nil {
			monitor.ConnectChanged(func(file, otherFile gio.Filer, eventType gio.FileMonitorEvent) {
				if eventType == gio.FileMonitorEventChanged || eventType == gio.FileMonitorEventCreated {
					// Ignore file changes while we're saving to avoid race conditions
					savingMux.Lock()
					saving := isSaving
					savingMux.Unlock()
					if saving {
						return
					}

					// Reload config from disk
					newCfg := loadConfig()
					cfg.PromptOnClick = newCfg.PromptOnClick
					cfg.FallbackBrowser = newCfg.FallbackBrowser
					cfg.CheckDefaultBrowser = newCfg.CheckDefaultBrowser
					cfg.Rules = newCfg.Rules

					// Update UI
					updateUI()
					rebuildRulesList()
				}
			})
		}
	}

	toolbarView.SetContent(scrolled)
	win.SetContent(toolbarView)

	win.Present()
}
