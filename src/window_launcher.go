// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showLauncherWindow(app *adw.Application, url string, browsers []*Browser, cfg *Config) {
	filteredBrowsers := filterAndSortBrowsers(browsers, cfg)

	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetApplication(&app.Application)

	// Main layout
	mainBox := gtk.NewBox(gtk.OrientationVertical, 0)

	// create early so button handlers can reference it
	urlEntry := createURLEntry(url)

	// Content area with browser buttons
	contentBox := createLauncherContentBox()
	launcherFlow := createLauncherFlowBox()
	flowBox := launcherFlow.FlowBox

	// Add responsive breakpoints to window
	win.AddBreakpoint(launcherFlow.NarrowBreakpoint)
	win.AddBreakpoint(launcherFlow.MediumBreakpoint)

	// Create browser buttons
	callbacks := BrowserButtonCallbacks{
		OnClick: func(browser *Browser) {
			launchBrowser(browser, urlEntry.Text())
			win.Close()
		},
		OnRightClick: func(btn *gtk.Button, browser *Browser) {
			showBrowserActionsMenu(btn, browser, urlEntry.Text())
		},
	}

	for _, browser := range filteredBrowsers {
		btn := createBrowserButton(browser, cfg.ShowAppNames, callbacks)
		flowBox.Insert(btn, -1)
	}

	// Handle Enter/Space activation on selected FlowBox child
	flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		idx := child.Index()
		if idx >= 0 && idx < len(filteredBrowsers) {
			launchBrowser(filteredBrowsers[idx], urlEntry.Text())
			win.Close()
		}
	})

	// Select first browser by default for keyboard navigation
	if first := flowBox.ChildAtIndex(0); first != nil {
		flowBox.SelectChild(first)
	}

	// URL entry activation launches selected browser
	urlEntry.ConnectActivate(func() {
		selected := flowBox.SelectedChildren()
		if len(selected) > 0 {
			idx := selected[0].Index()
			if idx >= 0 && idx < len(filteredBrowsers) {
				launchBrowser(filteredBrowsers[idx], urlEntry.Text())
				win.Close()
			}
		} else if len(filteredBrowsers) > 0 {
			launchBrowser(filteredBrowsers[0], urlEntry.Text())
			win.Close()
		}
	})

	contentBox.Append(flowBox)
	mainBox.Append(contentBox)

	// Bottom bar
	bottomBar := createLauncherBottomBar(urlEntry, func() { win.Close() })
	mainBox.Append(bottomBar)

	win.SetContent(mainBox)

	// Keyboard shortcuts
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		// Ctrl+[1-9] for quick selection
		if keyval >= gdk.KEY_1 && keyval <= gdk.KEY_9 && state&gdk.ControlMask != 0 {
			idx := int(keyval - gdk.KEY_1)
			if idx < len(filteredBrowsers) {
				launchBrowser(filteredBrowsers[idx], urlEntry.Text())
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
	actionGroup := setupLauncherActions(win, app, filteredBrowsers, url, func() { win.Close() })
	win.InsertActionGroup("win", actionGroup)

	win.Present()
}

// filter hidden browsers and moves favorite to front.
func filterAndSortBrowsers(browsers []*Browser, cfg *Config) []*Browser {
	// Filter hidden browsers
	hiddenSet := make(map[string]bool)
	for _, id := range cfg.HiddenBrowsers {
		hiddenSet[id] = true
	}

	filtered := make([]*Browser, 0, len(browsers))
	for _, browser := range browsers {
		if !hiddenSet[browser.ID] {
			filtered = append(filtered, browser)
		}
	}

	// Move favorite to front
	if cfg.FavoriteBrowser != "" {
		for i, browser := range filtered {
			if browser.ID == cfg.FavoriteBrowser {
				favorite := browser
				filtered = append([]*Browser{favorite}, append(filtered[:i], filtered[i+1:]...)...)
				break
			}
		}
	}

	return filtered
}
