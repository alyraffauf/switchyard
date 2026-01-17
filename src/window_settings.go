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
	win.SetDefaultSize(900, 600)
	win.SetResizable(true)

	// Set minimum size to prevent too-small window
	win.SetSizeRequest(700, 500)

	cfg := loadConfig()
	browsers := detectBrowsers()

	// Setup app-level actions
	setupAppActions(app, win)

	// Navigation split view (sidebar + content)
	splitView := adw.NewNavigationSplitView()
	splitView.SetShowContent(true)
	splitView.SetMinSidebarWidth(200)
	splitView.SetMaxSidebarWidth(200)

	// Sidebar
	sidebarPage := adw.NewNavigationPage(createSidebar(win, cfg, browsers, splitView), "Switchyard")
	splitView.SetSidebar(sidebarPage)

	// Initial content - show Appearance page by default
	contentPage := adw.NewNavigationPage(createAppearancePage(win, cfg), "Appearance")
	splitView.SetContent(contentPage)

	win.SetContent(splitView)

	// Check if we should prompt to set as default browser
	if cfg.CheckDefaultBrowser && !isDefaultBrowser() {
		showDefaultBrowserPrompt(win, cfg, func() {
			// Reload the behavior page after setting as default
			newPage := adw.NewNavigationPage(createBehaviorPage(win, cfg, browsers), "Behavior")
			splitView.SetContent(newPage)
		})
	}

	// Watch config file for external changes
	watchConfigFile(cfg, func() {
		// Refresh current view
		// For now, we'll just note that external changes happened
		// A full implementation would reload the current page
	})

	win.Present()
}

func setupAppActions(app *adw.Application, win *adw.Window) {
	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(p *glib.Variant) {
		showAboutDialog(win)
	})
	app.AddAction(aboutAction)

	donateAction := gio.NewSimpleAction("donate", nil)
	donateAction.ConnectActivate(func(p *glib.Variant) {
		launcher := gtk.NewURILauncher(DonateURL)
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	app.AddAction(donateAction)

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		win.Close()
	})
	app.AddAction(quitAction)

	// Keyboard shortcuts
	app.SetAccelsForAction("app.quit", []string{"<Ctrl>q"})
}

func createSidebar(win *adw.Window, cfg *Config, browsers []*Browser, splitView *adw.NavigationSplitView) gtk.Widgetter {
	// Use AdwToolbarView for proper sidebar architecture
	toolbarView := adw.NewToolbarView()

	// Sidebar header with menu
	sidebarHeader := adw.NewHeaderBar()
	sidebarHeader.SetShowEndTitleButtons(false)

	titleLabel := gtk.NewLabel("Switchyard")
	titleLabel.AddCSSClass("title")
	sidebarHeader.SetTitleWidget(titleLabel)

	// Hamburger menu
	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetTooltipText("Main Menu")

	menu := gio.NewMenu()
	menu.Append("Donate ❤️", "app.donate")
	menu.Append("About", "app.about")

	quitSection := gio.NewMenu()
	quitSection.Append("Quit", "app.quit")
	menu.AppendSection("", quitSection)

	menuBtn.SetMenuModel(menu)
	sidebarHeader.PackEnd(menuBtn)

	toolbarView.AddTopBar(sidebarHeader)

	// Scrolled window for sidebar content
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)

	// List box for navigation items
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	// Appearance row
	appearanceRow := adw.NewActionRow()
	appearanceRow.SetTitle("Appearance")
	appearanceRow.AddPrefix(gtk.NewImageFromIconName("applications-graphics-symbolic"))
	listBox.Append(appearanceRow)

	// Behavior row
	behaviorRow := adw.NewActionRow()
	behaviorRow.SetTitle("Behavior")
	behaviorRow.AddPrefix(gtk.NewImageFromIconName("preferences-system-symbolic"))
	listBox.Append(behaviorRow)

	// Rules row
	rulesRow := adw.NewActionRow()
	rulesRow.SetTitle("Rules")
	rulesRow.AddPrefix(gtk.NewImageFromIconName("view-list-symbolic"))
	listBox.Append(rulesRow)

	// Advanced row
	advancedRow := adw.NewActionRow()
	advancedRow.SetTitle("Advanced")
	advancedRow.AddPrefix(gtk.NewImageFromIconName("preferences-other-symbolic"))
	listBox.Append(advancedRow)

	scrolled.SetChild(listBox)
	toolbarView.SetContent(scrolled)

	// Select first item by default
	listBox.SelectRow(listBox.RowAtIndex(0))

	// Handle navigation - need to connect to both activated and selected
	navigateToPage := func(index int) {
		var page gtk.Widgetter
		var title string

		switch index {
		case 0: // Appearance
			page = createAppearancePage(win, cfg)
			title = "Appearance"
		case 1: // Behavior
			page = createBehaviorPage(win, cfg, browsers)
			title = "Behavior"
		case 2: // Rules
			page = createRulesPage(win, cfg, browsers)
			title = "Rules"
		case 3: // Advanced
			page = createAdvancedPage(cfg)
			title = "Advanced"
		}

		if page != nil {
			contentPage := adw.NewNavigationPage(page, title)
			splitView.SetContent(contentPage)
		}
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		navigateToPage(row.Index())
	})

	listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row != nil {
			navigateToPage(row.Index())
		}
	})

	return toolbarView
}

func createAppearancePage(win *adw.Window, cfg *Config) gtk.Widgetter {
	browsers := detectBrowsers()

	// Use AdwToolbarView for proper page architecture
	toolbarView := adw.NewToolbarView()

	// Header for this page
	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Appearance")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)
	toolbarView.AddTopBar(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	// App-wide appearance settings
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("General")

	forceDarkRow := adw.NewSwitchRow()
	forceDarkRow.SetTitle("Force dark mode")
	forceDarkRow.SetSubtitle("Always use dark mode")
	forceDarkRow.SetActive(cfg.ForceDarkMode)
	appearanceGroup.Add(forceDarkRow)

	content.Append(appearanceGroup)

	// Picker Window section
	pickerGroup := adw.NewPreferencesGroup()
	pickerGroup.SetTitle("Picker Window")

	showNamesRow := adw.NewSwitchRow()
	showNamesRow.SetTitle("Show browser names")
	showNamesRow.SetSubtitle("Show browser names below icons")
	showNamesRow.SetActive(cfg.ShowAppNames)
	pickerGroup.Add(showNamesRow)

	// Hidden browsers row
	hiddenBrowsersRow := adw.NewActionRow()
	hiddenBrowsersRow.SetTitle("Hidden browsers")
	hiddenBrowsersRow.SetSubtitle("Choose which browsers to hide from the picker")
	hiddenBrowsersRow.SetActivatable(true)

	// Show arrow icon to indicate it's clickable
	chevron := gtk.NewImageFromIconName("go-next-symbolic")
	hiddenBrowsersRow.AddSuffix(chevron)

	hiddenBrowsersRow.ConnectActivated(func() {
		showHiddenBrowsersDialog(win, cfg, browsers)
	})

	pickerGroup.Add(hiddenBrowsersRow)

	content.Append(pickerGroup)

	// Connect change handlers
	forceDarkRow.Connect("notify::active", func() {
		cfg.ForceDarkMode = forceDarkRow.Active()
		saveConfigWithFlag(cfg)
	})

	showNamesRow.Connect("notify::active", func() {
		cfg.ShowAppNames = showNamesRow.Active()
		saveConfigWithFlag(cfg)
	})

	toolbarView.SetContent(scrolled)
	return toolbarView
}

func createBehaviorPage(win *adw.Window, cfg *Config, browsers []*Browser) gtk.Widgetter {
	// Use AdwToolbarView for proper page architecture
	toolbarView := adw.NewToolbarView()

	// Header for this page
	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Behavior")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)
	toolbarView.AddTopBar(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	// General Behavior section
	behaviorGroup := adw.NewPreferencesGroup()
	behaviorGroup.SetTitle("General")

	checkDefaultRow := adw.NewSwitchRow()
	checkDefaultRow.SetTitle("Prompt to set as default browser")
	checkDefaultRow.SetSubtitle("Show prompt on startup if Switchyard is not the default browser")
	checkDefaultRow.SetActive(cfg.CheckDefaultBrowser)
	behaviorGroup.Add(checkDefaultRow)

	promptRow := adw.NewSwitchRow()
	promptRow.SetTitle("Show picker when no rule matches")
	promptRow.SetSubtitle("Let you choose a browser for unmatched URLs")
	promptRow.SetActive(cfg.PromptOnClick)
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

	// Set initial selection
	selectedIndex := uint(0)
	for i, b := range browsers {
		if b.ID == cfg.FavoriteBrowser {
			selectedIndex = uint(i + 1)
			break
		}
	}
	defaultRow.SetSelected(selectedIndex)

	behaviorGroup.Add(defaultRow)
	content.Append(behaviorGroup)

	// Connect change handlers
	checkDefaultRow.Connect("notify::active", func() {
		cfg.CheckDefaultBrowser = checkDefaultRow.Active()
		saveConfigWithFlag(cfg)
	})

	promptRow.Connect("notify::active", func() {
		cfg.PromptOnClick = promptRow.Active()
		saveConfigWithFlag(cfg)
	})

	defaultRow.Connect("notify::selected", func() {
		idx := defaultRow.Selected()
		if idx == 0 {
			cfg.FavoriteBrowser = ""
		} else if idx > 0 && int(idx) <= len(browsers) {
			cfg.FavoriteBrowser = browsers[idx-1].ID
		}
		saveConfigWithFlag(cfg)
	})

	toolbarView.SetContent(scrolled)
	return toolbarView
}

func createRulesPage(win *adw.Window, cfg *Config, browsers []*Browser) gtk.Widgetter {
	// Use AdwToolbarView for proper page architecture
	toolbarView := adw.NewToolbarView()

	// Header for this page
	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Rules")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)

	// Add Rule button in header
	addButton := gtk.NewButton()
	addButton.SetIconName("list-add-symbolic")
	addButton.SetTooltipText("Add New Rule")
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

	// Info banner (shown when rules exist)
	infoLabel := gtk.NewLabel("Rules are evaluated in order. First match wins.")
	infoLabel.SetWrap(true)
	infoLabel.SetXAlign(0)
	infoLabel.AddCSSClass("dim-label")
	infoLabel.SetMarginStart(12)
	infoLabel.SetMarginEnd(12)
	infoLabel.SetMarginBottom(6)
	content.Append(infoLabel)

	// Rules list with drag-and-drop support
	rulesListBox := gtk.NewListBox()
	rulesListBox.SetSelectionMode(gtk.SelectionNone)
	rulesListBox.AddCSSClass("boxed-list")

	// Empty state (shown when no rules exist)
	emptyState := adw.NewStatusPage()
	emptyState.SetIconName("list-add-symbolic")
	emptyState.SetTitle("No Rules")
	emptyState.SetDescription("Add rules to automatically route URLs to specific browsers")
	emptyState.SetVExpand(true)

	// Helper to get browser name from ID
	getBrowserName := func(id string) string {
		if browser := findBrowserByID(browsers, id); browser != nil {
			return browser.Name
		}
		return id
	}

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

			// Add all rule rows
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

func createAdvancedPage(cfg *Config) gtk.Widgetter {
	// Use AdwToolbarView for proper page architecture
	toolbarView := adw.NewToolbarView()

	// Header for this page
	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Advanced")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)
	toolbarView.AddTopBar(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	// Config file info
	configGroup := adw.NewPreferencesGroup()
	configGroup.SetTitle("Configuration")

	configRow := adw.NewActionRow()
	configRow.SetTitle("Configuration File")
	configRow.SetSubtitle(configPath())
	configRow.SetActivatable(true)
	configRow.AddSuffix(gtk.NewImageFromIconName("document-edit-symbolic"))
	configRow.ConnectActivated(func() {
		// Ensure config file exists
		saveConfig(cfg)
		// Open with xdg-open via flatpak-spawn when in Flatpak
		cmd := hostCommand("xdg-open", configPath())
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to open config file: %v\n", err)
		}
		go cmd.Wait()
	})
	configGroup.Add(configRow)

	content.Append(configGroup)

	toolbarView.SetContent(scrolled)
	return toolbarView
}

func watchConfigFile(cfg *Config, onChange func()) {
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
					cfg.ForceDarkMode = newCfg.ForceDarkMode
					cfg.HiddenBrowsers = newCfg.HiddenBrowsers
					cfg.Rules = newCfg.Rules

					if onChange != nil {
						onChange()
					}
				}
			})
		}
	}
}
