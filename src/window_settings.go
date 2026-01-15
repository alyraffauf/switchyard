// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showSettingsWindow(app *adw.Application) {
	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetApplication(&app.Application)
	win.SetDefaultSize(700, 800)
	win.SetResizable(true)

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
	menu.Append("Donate ❤️", "app.donate")
	menu.Append("About", "app.about")

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

	donateAction := gio.NewSimpleAction("donate", nil)
	donateAction.ConnectActivate(func(p *glib.Variant) {
		launcher := gtk.NewURILauncher("https://ko-fi.com/alyraffauf")
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	app.AddAction(donateAction)

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

	// Appearance section
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("Appearance")

	forceDarkRow := adw.NewSwitchRow()
	forceDarkRow.SetTitle("Force dark mode in browser picker")
	forceDarkRow.SetSubtitle("Always use dark mode for the picker window")
	appearanceGroup.Add(forceDarkRow)

	showNamesRow := adw.NewSwitchRow()
	showNamesRow.SetTitle("Show browser names in picker")
	showNamesRow.SetSubtitle("Show browser names below icons")
	appearanceGroup.Add(showNamesRow)

	content.Append(appearanceGroup)

	// Behavior section
	behaviorGroup := adw.NewPreferencesGroup()
	behaviorGroup.SetTitle("Behavior")

	checkDefaultRow := adw.NewSwitchRow()
	checkDefaultRow.SetTitle("Check if Switchyard is default browser")
	checkDefaultRow.SetSubtitle("Ask to set Switchyard as default browser on startup")
	behaviorGroup.Add(checkDefaultRow)

	promptRow := adw.NewSwitchRow()
	promptRow.SetTitle("Show picker when no rule matches")
	promptRow.SetSubtitle("Let you choose a browser for unmatched URLs")
	behaviorGroup.Add(promptRow)

	// Favorite browser dropdown
	browserNames := make([]string, len(browsers)+1)
	browserNames[0] = "None"
	for i, b := range browsers {
		browserNames[i+1] = b.Name
	}
	browserList := gtk.NewStringList(browserNames)

	defaultRow := adw.NewComboRow()
	defaultRow.SetTitle("Favorite browser")
	defaultRow.SetSubtitle("Appears first in picker and opens when picker is disabled")
	defaultRow.SetModel(browserList)
	behaviorGroup.Add(defaultRow)

	content.Append(behaviorGroup)

	// Function to update UI from config
	updateUI := func() {
		promptRow.SetActive(cfg.PromptOnClick)
		checkDefaultRow.SetActive(cfg.CheckDefaultBrowser)
		showNamesRow.SetActive(cfg.ShowAppNames)
		forceDarkRow.SetActive(cfg.ForceDarkMode)

		// Update favorite browser selection only if it differs from current
		var targetSelection uint = 0
		for i, b := range browsers {
			if b.ID == cfg.FavoriteBrowser {
				targetSelection = uint(i + 1)
				break
			}
		}
		if defaultRow.Selected() != targetSelection {
			defaultRow.SetSelected(targetSelection)
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
			cfg.FavoriteBrowser = ""
		} else if idx > 0 && int(idx) <= len(browsers) {
			cfg.FavoriteBrowser = browsers[idx-1].ID
		}
		saveConfigSafe(cfg)
	})

	checkDefaultRow.Connect("notify::active", func() {
		cfg.CheckDefaultBrowser = checkDefaultRow.Active()
		saveConfigSafe(cfg)
	})

	showNamesRow.Connect("notify::active", func() {
		cfg.ShowAppNames = showNamesRow.Active()
		saveConfigSafe(cfg)
	})

	forceDarkRow.Connect("notify::active", func() {
		cfg.ForceDarkMode = forceDarkRow.Active()
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
		var icon *gtk.Image
		if rule.AlwaysAsk {
			// Find app's own browser entry to load its icon
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
				// Fallback icon if browser not found
				icon = gtk.NewImageFromIconName("web-browser-symbolic")
				icon.SetPixelSize(24)
			}
		}
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
		showAddRuleDialog(win, cfg, browsers, rebuildRulesList)
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
	infoRow.ConnectActivated(func() {
		// Ensure config file exists
		saveConfig(cfg)
		// Open with xdg-open via flatpak-spawn when in Flatpak
		cmd := hostCommand("xdg-open", configPath())
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to open config file: %v\n", err)
		}
		go cmd.Wait()
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
					cfg.FavoriteBrowser = newCfg.FavoriteBrowser
					cfg.CheckDefaultBrowser = newCfg.CheckDefaultBrowser
					cfg.ShowAppNames = newCfg.ShowAppNames
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
