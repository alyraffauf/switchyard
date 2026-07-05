// Switchyard - A configurable default browser for Linux
// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"net/url"
	"os"
	"strings"

	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/alyraffauf/switchyard/internal/host"
	"github.com/alyraffauf/switchyard/internal/routing"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Run starts the Switchyard GTK application and blocks until it exits.
func Run() {
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
		cfg, _ := appconfig.Load(appconfig.Path())
		if cfg.StayAlive {
			app.Hold()
		}
		setupApp(cfg)
		browsers := detectBrowsers()
		showSettingsWindow(app, browsers, cfg)
	})

	app.ConnectOpen(func(files []gio.Filer, hint string) {
		cfg, _ := appconfig.Load(appconfig.Path())
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

		if u, err := url.Parse(rawURL); err == nil && u.Scheme == "switchyard" {
			handleSwitchyardURL(app, browsers, cfg, rawURL)
			return
		}

		sanitized := routing.PrepareURLForRouting(rawURL, cfg.RemoveTrackingParameters, cfg.Redirections)
		if sanitized == "" {
			// URL was rejected (mailto:, tel:, etc.) - pass to xdg-open
			cmd := host.HostCommand("xdg-open", rawURL)
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
	if cfg.ForceDarkMode {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorSchemeForceDark)
	}

	// Add host system icon paths when running in Flatpak
	if host.InFlatpak() {
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
	targetURL, browserPrefs, err := routing.ParseSwitchyardURL(rawURL)
	if err != nil {
		// Invalid switchyard URL - ignore
		return
	}

	sanitized := routing.PrepareURLForRouting(targetURL, cfg.RemoveTrackingParameters, cfg.Redirections)
	if sanitized == "" {
		// Pass non-browser URLs to xdg-open
		cmd := host.HostCommand("xdg-open", targetURL)
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

// handleURL routes a URL to the appropriate browser based on rules
func handleURL(app *adw.Application, browsers []*Browser, cfg *Config, urlStr string) {
	browserID, alwaysAsk, matched := cfg.MatchRule(urlStr)
	if matched {
		if alwaysAsk {
			showLauncherWindow(app, urlStr, browsers, cfg)
			return
		}

		if browser := findBrowserByID(browsers, browserID); browser != nil {
			launchBrowser(browser, urlStr)
			return
		}
	}

	// No rule matched: fall back to the favorite browser, else prompt.
	if !cfg.PromptOnClick && cfg.FavoriteBrowser != "" {
		if browser := findBrowserByID(browsers, cfg.FavoriteBrowser); browser != nil {
			launchBrowser(browser, urlStr)
			return
		}
	}

	showLauncherWindow(app, urlStr, browsers, cfg)
}
