// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const appID = "io.github.alyraffauf.Switchyard"

// Global flag to track if we're currently saving config to avoid file watcher race conditions
var (
	isSaving  bool
	savingMux sync.Mutex
)

func findBrowserByID(browsers []*Browser, id string) *Browser {
	for _, b := range browsers {
		if b.ID == id {
			return b
		}
	}
	return nil
}

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
	cfg := loadConfig()
	browsers := detectBrowsers()

	// Try to match a rule
	browserID, alwaysAsk, matched := cfg.matchRule(url)
	if matched {
		// Check if rule has AlwaysAsk enabled
		if alwaysAsk {
			showPickerWindow(app, url, browsers)
			return
		}

		// Find the browser and launch it
		if browser := findBrowserByID(browsers, browserID); browser != nil {
			launchBrowser(browser, url)
			return
		}
	}

	// No rule matched
	if !cfg.PromptOnClick && cfg.FallbackBrowser != "" {
		if browser := findBrowserByID(browsers, cfg.FallbackBrowser); browser != nil {
			launchBrowser(browser, url)
			return
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
		// Show name as title if set, otherwise show pattern
		if rule.Name != "" {
			row.SetTitle(rule.Name)
			row.SetSubtitle(formatRuleSubtitle(rule, getBrowserName(rule.Browser)))
		} else {
			// For rules without names, show first pattern or condition
			if len(rule.Conditions) > 0 {
				row.SetTitle(rule.Conditions[0].Pattern)
			} else {
				row.SetTitle(rule.Pattern)
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

// validateConditions checks if all conditions have non-empty patterns
func validateConditions(conditions []Condition) bool {
	for _, c := range conditions {
		if c.Pattern == "" {
			return false
		}
	}
	return true
}

// getLogicFromComboRow extracts the logic string from a combo row selection
func getLogicFromComboRow(logicRow *adw.ComboRow) string {
	if logicRow.Selected() == 1 {
		return "any"
	}
	return "all"
}

// saveConfigWithFlag saves config while setting the global saving flag to prevent file watcher loops
func saveConfigWithFlag(cfg *Config) {
	savingMux.Lock()
	isSaving = true
	savingMux.Unlock()
	saveConfig(cfg)
	glib.TimeoutAdd(100, func() bool {
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
		return false
	})
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
			patternRow.Connect("changed", func() {
				(*conditions)[condIdx].Pattern = patternRow.Text()
				// Enable/disable action button based on whether all conditions have patterns
				allValid := true
				for _, c := range *conditions {
					if c.Pattern == "" {
						allValid = false
						break
					}
				}
				actionBtn.SetSensitive(allValid)
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

func showEditRuleDialog(parent *adw.Window, cfg *Config, rule *Rule, row *adw.ActionRow, browsers []*Browser, getBrowserName func(string) string, getBrowserIcon func(string) string, rebuildRulesList func()) {
	dialog := adw.NewDialog()
	dialog.SetTitle("Edit Rule")
	dialog.SetContentWidth(600)
	dialog.SetContentHeight(650)
	dialog.SetCanClose(true)

	// Auto-migrate legacy rules when opening edit dialog
	if len(rule.Conditions) == 0 && rule.Pattern != "" {
		rule.Conditions = []Condition{{
			Type:    rule.PatternType,
			Pattern: rule.Pattern,
		}}
		rule.Logic = "all"
	}

	// Ensure new rules have at least one condition
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
			// Clear old fields
			rule.Pattern = ""
			rule.PatternType = ""

			saveConfigWithFlag(cfg)
			rebuildRulesList()
			dialog.Close()
		}
	})

	dialog.Present(parent)
}

func formatRuleSubtitle(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, true)
}

func formatRuleSubtitleNoPattern(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, false)
}

func formatRuleSubtitleInternal(rule *Rule, browserName string, includePattern bool) string {
	// Handle new multi-condition format
	if len(rule.Conditions) > 0 {
		condCount := len(rule.Conditions)
		var logicText string
		if rule.Logic == "any" {
			logicText = "Any match"
		} else {
			logicText = "All match"
		}

		if rule.AlwaysAsk {
			if condCount == 1 && includePattern {
				return fmt.Sprintf("%s: %s · Always ask", getTypeLabel(rule.Conditions[0].Type), rule.Conditions[0].Pattern)
			}
			return fmt.Sprintf("%d conditions (%s) · Always ask", condCount, logicText)
		}
		if condCount == 1 && includePattern {
			return fmt.Sprintf("%s: %s · Opens in %s", getTypeLabel(rule.Conditions[0].Type), rule.Conditions[0].Pattern, browserName)
		}
		return fmt.Sprintf("%d conditions (%s) · Opens in %s", condCount, logicText, browserName)
	}

	// Handle legacy single pattern format
	typeLabel := getTypeLabel(rule.PatternType)
	if rule.AlwaysAsk {
		if includePattern {
			return fmt.Sprintf("%s: %s · Always ask", typeLabel, rule.Pattern)
		}
		return fmt.Sprintf("%s · Always ask", typeLabel)
	}
	if includePattern {
		return fmt.Sprintf("%s: %s · Opens in %s", typeLabel, rule.Pattern, browserName)
	}
	return fmt.Sprintf("%s · Opens in %s", typeLabel, browserName)
}

func getTypeLabel(patternType string) string {
	switch patternType {
	case "domain":
		return "Exact domain"
	case "keyword":
		return "URL contains"
	case "glob":
		return "Wildcard"
	case "regex":
		return "Regex"
	default:
		return patternType
	}
}

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

	content := gtk.NewBox(gtk.OrientationVertical, 0)

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

	content.Append(group)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)

	saveBtn.ConnectClicked(func() {
		pattern := patternRow.Text()
		if pattern == "" {
			return
		}

		// Update condition
		switch typeRow.Selected() {
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
