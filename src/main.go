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

const appID = "io.github.alyraffauf.Switchyard"

// Global flag to track if we're currently saving config to avoid file watcher race conditions
var (
	isSaving  bool
	savingMux sync.Mutex
)

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

// setupIconPaths adds host system icon paths when running in Flatpak
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
	if !cfg.PromptOnClick && cfg.FallbackBrowser != "" {
		if browser := findBrowserByID(browsers, cfg.FallbackBrowser); browser != nil {
			launchBrowser(browser, url)
			return
		}
	}

	// Show picker
	showPickerWindow(app, url, browsers)
}
