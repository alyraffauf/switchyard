// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"fmt"
	"os"
	"sort"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
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

// detectBrowsers returns the installed browsers that handle HTTP URLs, sorted
// by name. GIO transparently covers system apps, Flatpaks, and Snaps.
func detectBrowsers() []*Browser {
	appInfos := gio.AppInfoGetRecommendedForType("x-scheme-handler/http")

	browsers := make([]*Browser, 0, len(appInfos))
	selfDesktopID := getAppID() + ".desktop"

	for _, appInfo := range appInfos {
		id := appInfo.ID()
		if id == "" || id == selfDesktopID {
			continue
		}

		icon := ""
		if gicon := appInfo.Icon(); gicon != nil {
			icon = gicon.String()
		}

		browsers = append(browsers, &Browser{
			Browser: &appbrowser.Browser{
				ID:   id,
				Name: appInfo.Name(),
				Icon: icon,
				Exec: appInfo.Commandline(),
			},
			appInfo: appInfo,
		})
	}

	sort.Slice(browsers, func(i, j int) bool {
		return browsers[i].Name < browsers[j].Name
	})

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
