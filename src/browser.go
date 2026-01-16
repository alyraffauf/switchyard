// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

type Browser struct {
	ID      string // desktop file ID (e.g., "firefox.desktop")
	Name    string
	Icon    string
	AppInfo *gio.AppInfo // Store the GIO AppInfo for launching
}

func detectBrowsers() []*Browser {
	var browsers []*Browser

	// Use GIO to get all applications that handle HTTP URLs
	// This automatically handles system apps, Flatpaks, Snaps, etc.
	appInfos := gio.AppInfoGetRecommendedForType("x-scheme-handler/http")

	for _, appInfo := range appInfos {
		id := appInfo.ID()
		if id == "" {
			continue
		}

		// Skip ourselves
		if id == getAppID()+".desktop" {
			continue
		}

		// Skip apps that shouldn't be shown
		if !appInfo.ShouldShow() {
			continue
		}

		name := appInfo.Name()
		icon := ""
		if gicon := appInfo.Icon(); gicon != nil {
			icon = gicon.String()
		}

		browsers = append(browsers, &Browser{
			ID:      id,
			Name:    name,
			Icon:    icon,
			AppInfo: appInfo,
		})
	}

	// Sort browsers alphabetically by name
	sort.Slice(browsers, func(i, j int) bool {
		return browsers[i].Name < browsers[j].Name
	})

	return browsers
}

func launchBrowser(b *Browser, url string) {
	cmdline := b.AppInfo.Commandline()
	if cmdline == "" {
		fmt.Fprintf(os.Stderr, "Error: No command line for browser %s\n", b.Name)
		return
	}
	if err := launchCommand(cmdline, url, b.AppInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser: %v\n", err)
	}
}

// launchBrowserAction launches a browser with a specific desktop file action (e.g., "new-private-window")
func launchBrowserAction(b *Browser, action DesktopAction, url string) {
	if action.Exec == "" {
		fmt.Fprintf(os.Stderr, "Error: No exec line for action %s\n", action.ID)
		return
	}
	if err := launchCommand(action.Exec, url, b.AppInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser action: %v\n", err)
	}
}
