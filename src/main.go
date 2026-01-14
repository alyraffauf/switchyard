// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const appID = "io.github.alyraffauf.Switchyard"

func main() {
	app := adw.NewApplication(appID, gio.ApplicationHandlesOpen)

	app.ConnectActivate(func() {
		// Add host icon paths for Flatpak compatibility
		setupIconPaths()
		showSettingsWindow(app)
	})

	app.ConnectOpen(func(files []gio.Filer, hint string) {
		// Add host icon paths for Flatpak compatibility
		setupIconPaths()

		if len(files) == 0 {
			showSettingsWindow(app)
			return
		}

		url := files[0].URI()
		url = sanitizeURL(url)
		handleURL(app, url)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func setupIconPaths() {
	// Add host system icon paths when running in Flatpak
	if os.Getenv("FLATPAK_ID") != "" {
		iconTheme := gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault())
		if iconTheme != nil {
			// Add Flatpak export paths for system and user flatpaks
			iconTheme.AddSearchPath("/var/lib/flatpak/exports/share/icons")
			home, _ := os.UserHomeDir()
			if home != "" {
				iconTheme.AddSearchPath(home + "/.local/share/flatpak/exports/share/icons")
			}
		}
	}
}

func loadBrowserIcon(iconName string, size int) *gtk.Image {
	if iconName == "" {
		iconName = "web-browser-symbolic"
	}

	// Try to load from icon theme first
	icon := gtk.NewImageFromIconName(iconName)
	icon.SetPixelSize(size)

	// Check if we need to try loading from file (for Flatpak apps)
	if os.Getenv("FLATPAK_ID") != "" && iconName != "web-browser-symbolic" {
		// Try to find the icon file directly
		iconPaths := []string{
			"/var/lib/flatpak/exports/share/icons/hicolor/64x64/apps/" + iconName + ".png",
			"/var/lib/flatpak/exports/share/icons/hicolor/128x128/apps/" + iconName + ".png",
			"/var/lib/flatpak/exports/share/icons/hicolor/scalable/apps/" + iconName + ".svg",
		}

		home, _ := os.UserHomeDir()
		if home != "" {
			iconPaths = append(iconPaths,
				home+"/.local/share/flatpak/exports/share/icons/hicolor/64x64/apps/"+iconName+".png",
				home+"/.local/share/flatpak/exports/share/icons/hicolor/128x128/apps/"+iconName+".png",
				home+"/.local/share/flatpak/exports/share/icons/hicolor/scalable/apps/"+iconName+".svg",
			)
		}

		// Try each path
		for _, path := range iconPaths {
			if _, err := os.Stat(path); err == nil {
				icon = gtk.NewImageFromFile(path)
				icon.SetPixelSize(size)
				break
			}
		}
	}

	return icon
}

func handleURL(app *adw.Application, url string) {
	continueHandlingURL(app, url)
}

func continueHandlingURL(app *adw.Application, url string) {
	cfg := loadConfig()
	browsers := detectBrowsers()

	// Try to match a rule
	matchedBrowserID := cfg.matchRuleID(url)
	if matchedBrowserID != "" {
		// Check if rule has AlwaysAsk enabled
		if cfg.matchRuleShouldAsk(url) {
			showPickerWindow(app, url, browsers)
			return
		}

		// Find the browser and launch it
		for _, b := range browsers {
			if b.ID == matchedBrowserID {
				launchBrowser(b, url)
				return
			}
		}
	}

	// No rule matched
	if !cfg.PromptOnClick && cfg.FallbackBrowser != "" {
		for _, b := range browsers {
			if b.ID == cfg.FallbackBrowser {
				launchBrowser(b, url)
				return
			}
		}
	}

	// Show picker
	showPickerWindow(app, url, browsers)
}

func showDefaultBrowserPrompt(parent gtk.Widgetter, cfg *Config, updateUI func()) {
	dialog := adw.NewAlertDialog(
		"Set as Default Browser?",
		"Switchyard is not your default browser. Would you like to set it as the default browser now?",
	)

	dialog.AddResponse("no", "Don't Ask Again")
	dialog.AddResponse("later", "Not Now")
	dialog.AddResponse("yes", "Set as Default")

	dialog.SetResponseAppearance("yes", adw.ResponseSuggested)
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

func showPickerWindow(app *adw.Application, url string, browsers []*Browser) {
	// Sort browsers alphabetically by name
	sortedBrowsers := make([]*Browser, len(browsers))
	copy(sortedBrowsers, browsers)
	for i := 0; i < len(sortedBrowsers); i++ {
		for j := i + 1; j < len(sortedBrowsers); j++ {
			if sortedBrowsers[i].Name > sortedBrowsers[j].Name {
				sortedBrowsers[i], sortedBrowsers[j] = sortedBrowsers[j], sortedBrowsers[i]
			}
		}
	}

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

	// Track if we're currently saving to avoid file watcher race conditions
	var isSaving bool

	// Wrap saveConfig to set the saving flag
	saveConfigSafe := func(c *Config) error {
		isSaving = true
		err := saveConfig(c)
		// Add small delay to ensure file is flushed before file watcher reads it
		glib.TimeoutAdd(100, func() bool {
			isSaving = false
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
		for _, b := range browsers {
			if b.ID == id {
				return b.Name
			}
		}
		return id
	}

	// Helper to get browser icon from ID
	getBrowserIcon := func(id string) string {
		for _, b := range browsers {
			if b.ID == id {
				return b.Icon
			}
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
		// Show name as title if set, otherwise show pattern
		if rule.Name != "" {
			row.SetTitle(rule.Name)
			row.SetSubtitle(formatRuleSubtitle(rule.PatternType, rule.Pattern, getBrowserName(rule.Browser), rule.AlwaysAsk))
		} else {
			row.SetTitle(rule.Pattern)
			row.SetSubtitle(formatRuleSubtitleNoPattern(rule.PatternType, getBrowserName(rule.Browser), rule.AlwaysAsk))
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
					if isSaving {
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
		gtk.ShowURI(&parent.Window, "https://github.com/alyraffauf/switchyard", 0)
	})
	linksGroup.Add(websiteRow)

	issueRow := adw.NewActionRow()
	issueRow.SetTitle("Report an Issue")
	issueRow.SetActivatable(true)
	issueRow.AddSuffix(gtk.NewImageFromIconName("external-link-symbolic"))
	issueRow.ConnectActivated(func() {
		gtk.ShowURI(&parent.Window, "https://github.com/alyraffauf/switchyard/issues", 0)
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

func showAddRuleDialog(parent *adw.Window, cfg *Config, browsers []*Browser, getBrowserName func(string) string, getBrowserIcon func(string) string, rebuildRulesList func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Add Rule")
	dialog.SetContentWidth(500)
	dialog.SetContentHeight(450)
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
	addBtn.SetSensitive(false) // Insensitive until pattern is filled
	addBtn.SetTooltipText("Add this rule")
	header.PackEnd(addBtn)

	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	// Name section
	nameGroup := adw.NewPreferencesGroup()
	nameGroup.SetTitle("Rule Name")
	nameGroup.SetDescription("Optional friendly name for this rule")

	nameEntry := adw.NewEntryRow()
	nameEntry.SetTitle("Name")
	nameGroup.Add(nameEntry)

	content.Append(nameGroup)

	// Rule types with descriptions
	ruleTypes := []string{
		"Exact domain",
		"URL contains",
		"Wildcard pattern",
		"Regular expression",
	}

	ruleTypeDescriptions := []string{
		"Matches only this exact domain",
		"Matches if the URL contains this text",
		"Matches using wildcards (* = anything)",
		"Matches using a regular expression",
	}

	ruleTypeExamples := []string{
		"github.com",
		"youtube.com/watch",
		"*.github.com",
		"^https://.*\\.example\\.(com|org)",
	}

	// Match condition section
	matchGroup := adw.NewPreferencesGroup()
	matchGroup.SetTitle("Match Condition")

	matchTypeRow := adw.NewComboRow()
	matchTypeRow.SetTitle("Rule type")
	matchTypeRow.SetModel(gtk.NewStringList(ruleTypes))
	matchGroup.Add(matchTypeRow)

	patternEntry := adw.NewEntryRow()
	patternEntry.SetTitle("Pattern")
	patternEntry.Connect("changed", func() {
		addBtn.SetSensitive(patternEntry.Text() != "")
	})
	matchGroup.Add(patternEntry)

	// Help row showing description and example
	helpRow := adw.NewActionRow()
	helpRow.SetTitle(ruleTypeDescriptions[0])
	helpRow.SetSubtitle("Example: " + ruleTypeExamples[0])
	helpRow.AddCSSClass("dim-label")
	matchGroup.Add(helpRow)

	// Update help when match type changes
	matchTypeRow.Connect("notify::selected", func() {
		idx := matchTypeRow.Selected()
		if int(idx) < len(ruleTypeDescriptions) {
			helpRow.SetTitle(ruleTypeDescriptions[idx])
			helpRow.SetSubtitle("Example: " + ruleTypeExamples[idx])
		}
	})

	content.Append(matchGroup)

	// Action section
	actionGroup := adw.NewPreferencesGroup()
	actionGroup.SetTitle("Open With")

	// Always Ask toggle
	alwaysAskRow := adw.NewSwitchRow()
	alwaysAskRow.SetTitle("Always ask")
	alwaysAskRow.SetSubtitle("Show browser picker for this rule")
	actionGroup.Add(alwaysAskRow)

	// Browser dropdown
	browserNames := make([]string, len(browsers))
	for i, b := range browsers {
		browserNames[i] = b.Name
	}

	browserRow := adw.NewComboRow()
	browserRow.SetTitle("Browser")
	browserRow.SetModel(gtk.NewStringList(browserNames))
	actionGroup.Add(browserRow)

	// Make browser row sensitive based on always ask toggle
	alwaysAskRow.Connect("notify::active", func() {
		browserRow.SetSensitive(!alwaysAskRow.Active())
	})

	content.Append(actionGroup)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)

	addBtn.ConnectClicked(func() {
		pattern := patternEntry.Text()
		browserIdx := browserRow.Selected()
		matchType := matchTypeRow.Selected()

		if pattern != "" && int(browserIdx) < len(browsers) {
			var patternType string

			switch matchType {
			case 0: // Exact domain
				patternType = "domain"
			case 1: // Domain contains
				patternType = "keyword"
			case 2: // Wildcard pattern
				patternType = "glob"
			case 3: // Regular expression
				patternType = "regex"
			}

			rule := Rule{
				Name:        nameEntry.Text(),
				Pattern:     pattern,
				PatternType: patternType,
				Browser:     browsers[browserIdx].ID,
				AlwaysAsk:   alwaysAskRow.Active(),
			}
			cfg.Rules = append(cfg.Rules, rule)
			saveConfig(cfg)
			rebuildRulesList()
			dialog.Close()
		}
	})

	dialog.Present(parent)
}

func showEditRuleDialog(parent *adw.Window, cfg *Config, rule *Rule, row *adw.ActionRow, browsers []*Browser, getBrowserName func(string) string, getBrowserIcon func(string) string, rebuildRulesList func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Edit Rule")
	dialog.SetContentWidth(500)
	dialog.SetContentHeight(450)
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

	saveBtn := gtk.NewButton()
	saveBtn.SetLabel("Save")
	saveBtn.AddCSSClass("suggested-action")
	saveBtn.SetTooltipText("Save changes to this rule")
	header.PackEnd(saveBtn)

	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	// Name section
	nameGroup := adw.NewPreferencesGroup()
	nameGroup.SetTitle("Rule Name")
	nameGroup.SetDescription("Optional friendly name for this rule")

	nameEntry := adw.NewEntryRow()
	nameEntry.SetTitle("Name")
	nameEntry.SetText(rule.Name)
	nameGroup.Add(nameEntry)

	content.Append(nameGroup)

	// Rule types with descriptions
	ruleTypes := []string{
		"Exact domain",
		"URL contains",
		"Wildcard pattern",
		"Regular expression",
	}

	ruleTypeDescriptions := []string{
		"Matches only this exact domain",
		"Matches if the URL contains this text",
		"Matches using wildcards (* = anything)",
		"Matches using a regular expression",
	}

	ruleTypeExamples := []string{
		"github.com",
		"youtube.com/watch",
		"*.github.com",
		"^https://.*\\.example\\.(com|org)",
	}

	// Determine current match type index
	var currentMatchType uint
	switch rule.PatternType {
	case "domain":
		currentMatchType = 0
	case "keyword":
		currentMatchType = 1
	case "glob":
		currentMatchType = 2
	case "regex":
		currentMatchType = 3
	default:
		currentMatchType = 1
	}

	// Match condition section
	matchGroup := adw.NewPreferencesGroup()
	matchGroup.SetTitle("Match Condition")

	matchTypeRow := adw.NewComboRow()
	matchTypeRow.SetTitle("Rule type")
	matchTypeRow.SetModel(gtk.NewStringList(ruleTypes))
	matchTypeRow.SetSelected(currentMatchType)
	matchGroup.Add(matchTypeRow)

	patternEntry := adw.NewEntryRow()
	patternEntry.SetTitle("Pattern")
	patternEntry.SetText(rule.Pattern)
	matchGroup.Add(patternEntry)

	// Help row showing description and example
	helpRow := adw.NewActionRow()
	helpRow.SetTitle(ruleTypeDescriptions[currentMatchType])
	helpRow.SetSubtitle("Example: " + ruleTypeExamples[currentMatchType])
	helpRow.AddCSSClass("dim-label")
	matchGroup.Add(helpRow)

	// Update help when match type changes
	matchTypeRow.Connect("notify::selected", func() {
		idx := matchTypeRow.Selected()
		if int(idx) < len(ruleTypeDescriptions) {
			helpRow.SetTitle(ruleTypeDescriptions[idx])
			helpRow.SetSubtitle("Example: " + ruleTypeExamples[idx])
		}
	})

	content.Append(matchGroup)

	// Action section
	actionGroup := adw.NewPreferencesGroup()
	actionGroup.SetTitle("Open With")

	// Always Ask toggle
	alwaysAskRow := adw.NewSwitchRow()
	alwaysAskRow.SetTitle("Always ask")
	alwaysAskRow.SetSubtitle("Show browser picker for this rule")
	alwaysAskRow.SetActive(rule.AlwaysAsk)
	actionGroup.Add(alwaysAskRow)

	// Browser dropdown
	browserNames := make([]string, len(browsers))
	selectedIdx := uint(0)

	for i, b := range browsers {
		browserNames[i] = b.Name
		if b.ID == rule.Browser {
			selectedIdx = uint(i)
		}
	}

	browserRow := adw.NewComboRow()
	browserRow.SetTitle("Browser")
	browserRow.SetModel(gtk.NewStringList(browserNames))
	browserRow.SetSelected(selectedIdx)
	browserRow.SetSensitive(!rule.AlwaysAsk) // Grey out if AlwaysAsk is enabled
	actionGroup.Add(browserRow)

	// Make browser row sensitive based on always ask toggle
	alwaysAskRow.Connect("notify::active", func() {
		browserRow.SetSensitive(!alwaysAskRow.Active())
	})

	content.Append(actionGroup)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)

	saveBtn.ConnectClicked(func() {
		pattern := patternEntry.Text()
		browserIdx := browserRow.Selected()
		matchType := matchTypeRow.Selected()

		if pattern != "" && int(browserIdx) < len(browsers) {
			rule.Name = nameEntry.Text()
			rule.Pattern = pattern
			rule.Browser = browsers[browserIdx].ID
			rule.AlwaysAsk = alwaysAskRow.Active()

			switch matchType {
			case 0:
				rule.PatternType = "domain"
			case 1:
				rule.PatternType = "keyword"
			case 2:
				rule.PatternType = "glob"
			case 3:
				rule.PatternType = "regex"
			}

			saveConfig(cfg)
			rebuildRulesList()
			dialog.Close()
		}
	})

	dialog.Present(parent)
}

func formatRuleSubtitle(patternType, pattern, browserName string, alwaysAsk bool) string {
	var typeLabel string
	switch patternType {
	case "domain":
		typeLabel = "Exact domain"
	case "keyword":
		typeLabel = "URL contains"
	case "glob":
		typeLabel = "Wildcard"
	case "regex":
		typeLabel = "Regex"
	default:
		typeLabel = patternType
	}

	if alwaysAsk {
		return fmt.Sprintf("%s: %s · Always ask", typeLabel, pattern)
	}
	return fmt.Sprintf("%s: %s · Opens in %s", typeLabel, pattern, browserName)
}

func formatRuleSubtitleNoPattern(patternType, browserName string, alwaysAsk bool) string {
	var typeLabel string
	switch patternType {
	case "domain":
		typeLabel = "Exact domain"
	case "keyword":
		typeLabel = "URL contains"
	case "glob":
		typeLabel = "Wildcard"
	case "regex":
		typeLabel = "Regex"
	default:
		typeLabel = patternType
	}

	if alwaysAsk {
		return fmt.Sprintf("%s · Always ask", typeLabel)
	}
	return fmt.Sprintf("%s · Opens in %s", typeLabel, browserName)
}

func createBrowserButton(b *Browser) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.SetHAlign(gtk.AlignCenter)

	if b.Icon != "" {
		icon := gtk.NewImageFromIconName(b.Icon)
		icon.SetPixelSize(48)
		box.Append(icon)
	}

	label := gtk.NewLabel(b.Name)
	label.SetEllipsize(pango.EllipsizeEnd)
	label.SetMaxWidthChars(12)
	box.Append(label)

	return box
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
