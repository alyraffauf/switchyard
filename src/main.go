// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"os"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	app := adw.NewApplication(getAppID(), gio.ApplicationHandlesOpen)

	app.AddMainOption("native-host", 0, glib.OptionFlagNone, glib.OptionArgNone,
		"Run as native-messaging host (invoked by browsers)", "")

	app.ConnectHandleLocalOptions(func(options *glib.VariantDict) int {
		if options.Contains("native-host") {
			runNativeMessagingHost()
			return 0
		}
		return -1
	})

	app.ConnectActivate(func() {
		cfg := loadConfig()
		if cfg.StayAlive {
			app.Hold()
		}
		setupApp(cfg)
		browsers := detectBrowsers()
		showSettingsWindow(app, browsers, cfg)
	})

	app.ConnectOpen(func(files []gio.Filer, hint string) {
		cfg := loadConfig()
		if cfg.StayAlive {
			app.Hold()
		}
		setupApp(cfg)
		browsers := detectBrowsers()

		if len(files) == 0 {
			showSettingsWindow(app, browsers, cfg)
			return
		}

		rawURL := files[0].URI()

		// Check if this is a switchyard:// URL
		if u, err := url.Parse(rawURL); err == nil && u.Scheme == "switchyard" {
			handleSwitchyardURL(app, browsers, cfg, rawURL)
			return
		}

		sanitized := prepareURLForRouting(rawURL, cfg)
		if sanitized == "" {
			// URL was rejected (mailto:, tel:, etc.) - pass to xdg-open
			cmd := hostCommand("xdg-open", rawURL)
			cmd.Start()
			return
		}

		handleURL(app, browsers, cfg, sanitized)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

// setupApp initializes app-wide settings like dark mode and icon paths
func setupApp(cfg *Config) {
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

// handleSwitchyardURL processes switchyard:// URLs with browser preferences
func handleSwitchyardURL(app *adw.Application, browsers []*Browser, cfg *Config, rawURL string) {
	targetURL, browserPrefs, err := parseSwitchyardURL(rawURL)
	if err != nil {
		// Invalid switchyard URL - ignore
		return
	}

	sanitized := prepareURLForRouting(targetURL, cfg)
	if sanitized == "" {
		// Pass non-browser URLs to xdg-open
		cmd := hostCommand("xdg-open", targetURL)
		cmd.Start()
		return
	}

	// If browser preferences specified, try each in order
	if len(browserPrefs) > 0 {
		for _, pref := range browserPrefs {
			// Try with and without .desktop suffix
			id := pref
			if !strings.HasSuffix(id, ".desktop") {
				id = id + ".desktop"
			}
			if browser := findBrowserByID(browsers, id); browser != nil {
				launchBrowser(browser, sanitized)
				return
			}
		}
		// No preferred browser found - show launcher
		showLauncherWindow(app, sanitized, browsers, cfg)
		return
	}

	// No browser specified - use standard routing
	handleURL(app, browsers, cfg, sanitized)
}

func prepareURLForRouting(rawURL string, cfg *Config) string {
	sanitized := sanitizeURL(rawURL)
	if sanitized == "" {
		return ""
	}

	if cfg.RemoveTrackingParameters {
		sanitized = removeTrackingParameters(sanitized)
	}

	if len(cfg.Redirections) > 0 {
		sanitized = applyRedirections(sanitized, cfg.Redirections)
	}

	return sanitized
}

// handleURL routes a URL to the appropriate browser based on rules
func handleURL(app *adw.Application, browsers []*Browser, cfg *Config, urlStr string) {
	// Try to match a rule
	browserID, alwaysAsk, matched := cfg.matchRule(urlStr)
	if matched {
		// Check if rule has AlwaysAsk enabled
		if alwaysAsk {
			showLauncherWindow(app, urlStr, browsers, cfg)
			return
		}

		// Find the browser and launch it
		if browser := findBrowserByID(browsers, browserID); browser != nil {
			launchBrowser(browser, urlStr)
			return
		}
	}

	// No rule matched
	if !cfg.PromptOnClick && cfg.FavoriteBrowser != "" {
		if browser := findBrowserByID(browsers, cfg.FavoriteBrowser); browser != nil {
			launchBrowser(browser, urlStr)
			return
		}
	}

	showLauncherWindow(app, urlStr, browsers, cfg)
}
