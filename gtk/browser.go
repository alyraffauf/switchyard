// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"fmt"
	"os"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
	"github.com/alyraffauf/switchyard/internal/browserscan"
	"github.com/alyraffauf/switchyard/internal/host"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

// Browser is a detected browser as the GTK layer sees it: the agnostic model
// plus the live GIO handle needed for sharp icons and Wayland activation
// tokens. The handle rides with each browser so those GTK concerns stay local.
type Browser struct {
	*appbrowser.Browser
	appInfo *gio.AppInfo
}

// DesktopAction is a desktop-entry action such as "new-private-window".
type DesktopAction = appbrowser.Action

// detectBrowsers gets metadata from browserscan and an optional live GIO handle.
func detectBrowsers() []*Browser {
	appInfoByID := make(map[string]*gio.AppInfo)
	for _, appInfo := range gio.AppInfoGetRecommendedForType("x-scheme-handler/http") {
		if id := appInfo.ID(); id != "" {
			appInfoByID[id] = appInfo
		}
	}

	installed := browserscan.Installed()
	browsers := make([]*Browser, 0, len(installed))
	for i := range installed {
		browserModel := installed[i]
		browsers = append(browsers, &Browser{
			Browser: &browserModel,
			appInfo: appInfoByID[browserModel.ID],
		})
	}

	return browsers
}

func launchBrowser(b *Browser, url string) {
	if b.Exec == "" {
		fmt.Fprintf(os.Stderr, "Error: No command line for browser %s\n", b.Name)
		return
	}
	if err := appbrowser.Launch(b.Exec, url, b.activationToken(), host.InFlatpak()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser: %v\n", err)
	}
}

// launchBrowserAction launches a browser using a specific desktop action's
// command line (e.g. "new-private-window").
func launchBrowserAction(b *Browser, action DesktopAction, url string) {
	if action.Exec == "" {
		fmt.Fprintf(os.Stderr, "Error: No exec line for action %s\n", action.ID)
		return
	}
	if err := appbrowser.Launch(action.Exec, url, b.activationToken(), host.InFlatpak()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser action: %v\n", err)
	}
}

// activationToken returns the Wayland/X11 startup token used to raise b's
// window, or "" when the handle or display is unavailable.
func (b *Browser) activationToken() string {
	if b.appInfo == nil {
		return ""
	}
	display := gdk.DisplayGetDefault()
	if display == nil {
		return ""
	}
	return display.AppLaunchContext().StartupNotifyID(b.appInfo, nil)
}
