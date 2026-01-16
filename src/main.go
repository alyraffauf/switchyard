// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"sync"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Global flag to track if we're currently saving config to avoid file watcher race conditions
var (
	isSaving  bool
	savingMux sync.Mutex
)

func main() {
	app := adw.NewApplication(getAppID(), gio.ApplicationHandlesOpen)

	app.ConnectActivate(func() {
		setupApp()
		showSettingsWindow(app)
	})

	app.ConnectOpen(func(files []gio.Filer, hint string) {
		setupApp()

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

// setupApp initializes app-wide settings like dark mode and icon paths
func setupApp() {
	cfg := loadConfig()

	// Apply dark mode app-wide
	if cfg.ForceDarkMode {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorSchemeForceDark)
	}

	// Add host system icon paths when running in Flatpak
	if os.Getenv("FLATPAK_ID") != "" {
		iconTheme := gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault())
		if iconTheme != nil {
			iconTheme.AddSearchPath("/var/lib/flatpak/exports/share/icons")
			home, _ := os.UserHomeDir()
			if home != "" {
				iconTheme.AddSearchPath(home + "/.local/share/flatpak/exports/share/icons")
			}
		}
	}
}

// handleURL routes a URL to the appropriate browser based on rules
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
	if !cfg.PromptOnClick && cfg.FavoriteBrowser != "" {
		if browser := findBrowserByID(browsers, cfg.FavoriteBrowser); browser != nil {
			launchBrowser(browser, url)
			return
		}
	}

	// Show picker
	showPickerWindow(app, url, browsers)
}
