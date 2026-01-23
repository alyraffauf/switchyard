// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

func showPickerWindow(app *adw.Application, url string, browsers []*Browser, cfg *Config) {

	// Filter hidden_browsers from the list
	hiddenSet := make(map[string]bool)
	for _, id := range cfg.HiddenBrowsers {
		hiddenSet[id] = true
	}

	filteredBrowsers := make([]*Browser, 0, len(browsers))
	for _, browser := range browsers {
		if !hiddenSet[browser.ID] {
			filteredBrowsers = append(filteredBrowsers, browser)
		}
	}

	// Move favorite to front
	if cfg.FavoriteBrowser != "" {
		for i, browser := range filteredBrowsers {
			if browser.ID == cfg.FavoriteBrowser {
				favorite := browser
				filteredBrowsers = append([]*Browser{favorite}, append(filteredBrowsers[:i], filteredBrowsers[i+1:]...)...)
				break
			}
		}
	}

	win := adw.NewWindow()
	win.SetTitle("Switchyard")
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
	flowBox.SetSelectionMode(gtk.SelectionSingle)
	flowBox.SetActivateOnSingleClick(false)
	flowBox.SetColumnSpacing(16)
	flowBox.SetRowSpacing(16)
	flowBox.SetMaxChildrenPerLine(6)
	flowBox.SetMinChildrenPerLine(2)
	flowBox.SetHAlign(gtk.AlignCenter)
	flowBox.SetVAlign(gtk.AlignStart)

	// Add breakpoint for narrow windows (mobile/small screens)
	narrowBreakpoint := adw.NewBreakpoint(adw.NewBreakpointConditionLength(adw.BreakpointConditionMaxWidth, 400, adw.LengthUnitPx))
	narrowBreakpoint.AddSetter(flowBox, "max-children-per-line", uint(2))
	win.AddBreakpoint(narrowBreakpoint)

	// Add breakpoint for medium windows
	mediumBreakpoint := adw.NewBreakpoint(adw.NewBreakpointConditionLength(adw.BreakpointConditionMaxWidth, 600, adw.LengthUnitPx))
	mediumBreakpoint.AddSetter(flowBox, "max-children-per-line", uint(3))
	win.AddBreakpoint(mediumBreakpoint)

	for _, browser := range filteredBrowsers {
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

		// Set accessible label for screen readers (always, regardless of visual label)
		labelValue := coreglib.NewValue(b.Name)
		btn.UpdateProperty([]gtk.AccessibleProperty{gtk.AccessiblePropertyLabel}, []coreglib.Value{*labelValue})

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

	// Handle Enter/Space activation on selected FlowBox child
	flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		idx := child.Index()
		if idx >= 0 && idx < len(filteredBrowsers) {
			currentURL := urlEntry.Text()
			launchBrowser(filteredBrowsers[idx], currentURL)
			win.Close()
		}
	})

	// Select first browser by default for keyboard navigation
	if first := flowBox.ChildAtIndex(0); first != nil {
		flowBox.SelectChild(first)
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
			if idx < len(filteredBrowsers) {
				currentURL := urlEntry.Text()
				launchBrowser(filteredBrowsers[idx], currentURL)
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

	// Set up action handlers
	actionGroup := setupPickerActions(win, app, filteredBrowsers, url, func() { win.Close() })
	win.InsertActionGroup("win", actionGroup)

	win.Present()
}
