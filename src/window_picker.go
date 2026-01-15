// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

// showPickerWindow displays the browser picker window
func showPickerWindow(app *adw.Application, url string, browsers []*Browser) {
	cfg := loadConfig()

	if cfg.ForceDarkMode {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorSchemeForceDark)
	}

	// Sort browsers: favorite first, then alphabetically by name
	sortedBrowsers := make([]*Browser, len(browsers))
	copy(sortedBrowsers, browsers)
	sort.Slice(sortedBrowsers, func(i, j int) bool {
		// Favorite browser always comes first
		isFavoriteI := cfg.FavoriteBrowser != "" && sortedBrowsers[i].ID == cfg.FavoriteBrowser
		isFavoriteJ := cfg.FavoriteBrowser != "" && sortedBrowsers[j].ID == cfg.FavoriteBrowser

		if isFavoriteI && !isFavoriteJ {
			return true
		}
		if !isFavoriteI && isFavoriteJ {
			return false
		}

		// Otherwise sort alphabetically
		return sortedBrowsers[i].Name < sortedBrowsers[j].Name
	})

	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetDefaultSize(700, -1)
	win.SetResizable(false)
	win.SetApplication(&app.Application)

	// Main layout - simple vertical box without title bar
	mainBox := gtk.NewBox(gtk.OrientationVertical, 0)

	// Create URL entry early so it can be referenced in button handlers
	urlEntry := gtk.NewEntry()
	urlEntry.SetText(url)
	urlEntry.SetEditable(true)
	urlEntry.SetCanFocus(true)
	urlEntry.SetAlignment(0.5)
	urlEntry.SetMaxWidthChars(50)
	urlEntry.SetWidthChars(40)

	// Content box with margins
	contentBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentBox.SetMarginStart(12)
	contentBox.SetMarginEnd(12)
	contentBox.SetMarginTop(24)
	contentBox.SetMarginBottom(8)

	// FlowBox for browser buttons - wraps to multiple rows
	flowBox := gtk.NewFlowBox()
	flowBox.SetSelectionMode(gtk.SelectionNone)
	flowBox.SetHomogeneous(true)
	flowBox.SetColumnSpacing(16)
	flowBox.SetRowSpacing(16)
	flowBox.SetMaxChildrenPerLine(4)
	flowBox.SetHAlign(gtk.AlignCenter)
	flowBox.SetVAlign(gtk.AlignStart)

	for _, browser := range sortedBrowsers {
		b := browser // capture

		// Button for each browser
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		btn.SetSizeRequest(134, 134)

		// Container inside button - icon above, name and shortcut below
		btnBox := gtk.NewBox(gtk.OrientationVertical, 8)
		btnBox.SetHAlign(gtk.AlignCenter)
		btnBox.SetVAlign(gtk.AlignCenter)

		// Fixed-size container for icon to ensure uniform sizing
		iconBox := gtk.NewBox(gtk.OrientationVertical, 0)
		iconBox.SetSizeRequest(128, 128)
		iconBox.SetHAlign(gtk.AlignCenter)
		iconBox.SetVAlign(gtk.AlignCenter)

		// Large browser icon - use helper to load with fallback
		icon := loadBrowserIcon(b, 128)
		icon.SetHAlign(gtk.AlignCenter)
		icon.SetVAlign(gtk.AlignCenter)
		iconBox.Append(icon)

		btnBox.Append(iconBox)

		// Show browser name based on config
		if cfg.ShowAppNames {
			// Show as visible label
			label := gtk.NewLabel(b.Name)
			label.SetEllipsize(pango.EllipsizeEnd)
			label.SetMaxWidthChars(18)
			label.SetJustify(gtk.JustifyCenter)
			label.SetLines(1)
			label.SetMarginTop(6)
			btnBox.Append(label)
		} else {
			// Show as tooltip on hover
			btn.SetTooltipText(b.Name)
		}

		btn.SetChild(btnBox)

		btn.ConnectClicked(func() {
			currentURL := urlEntry.Text()
			launchBrowser(b, currentURL)
			win.Close()
		})

		// Add right-click handler for desktop file actions
		gesture := gtk.NewGestureClick()
		gesture.SetButton(gdk.BUTTON_SECONDARY)
		gesture.ConnectPressed(func(nPress int, x, y float64) {
			currentURL := urlEntry.Text()
			showBrowserActionsMenu(btn, b, currentURL)
		})
		btn.AddController(gesture)

		flowBox.Insert(btn, -1)
	}

	contentBox.Append(flowBox)
	mainBox.Append(contentBox)

	// Bottom bar with hamburger menu, URL, and close button
	bottomBar := gtk.NewBox(gtk.OrientationHorizontal, 12)
	bottomBar.SetMarginStart(8)
	bottomBar.SetMarginEnd(8)
	bottomBar.SetMarginTop(8)
	bottomBar.SetMarginBottom(8)

	// Hamburger menu button (left)
	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetTooltipText("Main menu")
	menuBtn.AddCSSClass("flat")

	menu := gio.NewMenu()
	menu.Append("Settings", "win.settings")

	aboutSection := gio.NewMenu()
	aboutSection.Append("Donate ❤️", "win.donate")
	aboutSection.Append("About", "win.about")
	aboutSection.Append("Keyboard Shortcuts", "win.shortcuts")
	menu.AppendSection("", aboutSection)

	quitSection := gio.NewMenu()
	quitSection.Append("Quit", "win.quit")
	menu.AppendSection("", quitSection)

	menuBtn.SetMenuModel(menu)
	bottomBar.Append(menuBtn)

	// Spacer before URL (to center it)
	leftSpacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	leftSpacer.SetHExpand(true)
	bottomBar.Append(leftSpacer)

	// Append the URL entry we created earlier
	bottomBar.Append(urlEntry)

	// Spacer after URL (to center it)
	rightSpacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	rightSpacer.SetHExpand(true)
	bottomBar.Append(rightSpacer)

	// Close button (right, circular like standard GTK close button)
	closeBtn := gtk.NewButton()
	closeBtn.SetIconName("window-close-symbolic")
	closeBtn.SetTooltipText("Close")
	closeBtn.AddCSSClass("circular")
	closeBtn.ConnectClicked(func() {
		win.Close()
	})
	bottomBar.Append(closeBtn)

	mainBox.Append(bottomBar)
	win.SetContent(mainBox)

	// Keyboard shortcuts
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		// Ctrl+[1-9] for quick selection
		if keyval >= gdk.KEY_1 && keyval <= gdk.KEY_9 && state&gdk.ControlMask != 0 {
			idx := int(keyval - gdk.KEY_1)
			if idx < len(sortedBrowsers) {
				currentURL := urlEntry.Text()
				launchBrowser(sortedBrowsers[idx], currentURL)
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

	// Set up action handlers for menu and desktop file actions
	actionGroup := gio.NewSimpleActionGroup()

	// Menu actions
	settingsAction := gio.NewSimpleAction("settings", nil)
	settingsAction.ConnectActivate(func(p *glib.Variant) {
		showSettingsWindow(app)
	})
	actionGroup.AddAction(settingsAction)

	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(p *glib.Variant) {
		showAboutDialog(win)
	})
	actionGroup.AddAction(aboutAction)

	donateAction := gio.NewSimpleAction("donate", nil)
	donateAction.ConnectActivate(func(p *glib.Variant) {
		launcher := gtk.NewURILauncher("https://ko-fi.com/alyraffauf")
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	actionGroup.AddAction(donateAction)

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		win.Close()
	})
	actionGroup.AddAction(quitAction)

	// Keyboard shortcuts action
	shortcutsAction := gio.NewSimpleAction("shortcuts", nil)
	shortcutsAction.ConnectActivate(func(p *glib.Variant) {
		showShortcutsDialog(win)
	})
	actionGroup.AddAction(shortcutsAction)

	// Action to launch browser with a specific action
	// Parameter format: "browserID:actionID"
	launchActionAction := gio.NewSimpleAction("launch-action", glib.NewVariantType("s"))
	launchActionAction.ConnectActivate(func(param *glib.Variant) {
		if param == nil {
			return
		}

		// Parse "browserID:actionID" from the parameter
		actionSpec := param.String()
		parts := strings.Split(actionSpec, ":")
		if len(parts) != 2 {
			return
		}

		browserID := parts[0]
		actionID := parts[1]

		// Find the browser
		var selectedBrowser *Browser
		for _, b := range sortedBrowsers {
			if b.ID == browserID {
				selectedBrowser = b
				break
			}
		}

		if selectedBrowser == nil {
			return
		}

		// Find the action and launch it
		actions := ListDesktopActions(selectedBrowser.AppInfo)
		for _, action := range actions {
			if action.ID == actionID {
				launchBrowserAction(selectedBrowser, action, url)
				win.Close()
				return
			}
		}
	})
	actionGroup.AddAction(launchActionAction)

	win.InsertActionGroup("win", actionGroup)

	win.Present()
}

// showBrowserActionsMenu shows a context menu with desktop file actions
func showBrowserActionsMenu(btn *gtk.Button, browser *Browser, url string) {
	actions := ListDesktopActions(browser.AppInfo)
	if len(actions) == 0 {
		return
	}

	// Build menu model
	menu := gio.NewMenu()

	// Add desktop file actions
	for _, action := range actions {
		// Use the action ID as a unique identifier for the action
		menu.Append(action.Name, fmt.Sprintf("win.launch-action::%s:%s", browser.ID, action.ID))
	}

	// Create and show popover
	popover := gtk.NewPopoverMenuFromModel(menu)
	popover.SetParent(btn)
	popover.Popup()
}

// showShortcutsDialog displays available keyboard shortcuts
func showShortcutsDialog(parent *adw.Window) {
	dialog := adw.NewAlertDialog(
		"Keyboard Shortcuts",
		"Ctrl+1 through Ctrl+9: Select browser 1-9\nEsc: Close picker window",
	)

	dialog.AddResponse("ok", "OK")
	dialog.SetDefaultResponse("ok")
	dialog.SetCloseResponse("ok")

	dialog.Present(parent)
}
