// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

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
		if id == "io.github.alyraffauf.Switchyard.desktop" {
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
	// Get the command line from the AppInfo
	cmdline := b.AppInfo.Commandline()
	if cmdline == "" {
		fmt.Fprintf(os.Stderr, "Error: No command line for browser %s\n", b.Name)
		return
	}

	// Replace %u, %U, %f, %F with URL
	cmdline = strings.ReplaceAll(cmdline, "%u", url)
	cmdline = strings.ReplaceAll(cmdline, "%U", url)
	cmdline = strings.ReplaceAll(cmdline, "%f", url)
	cmdline = strings.ReplaceAll(cmdline, "%F", url)

	// Remove other field codes
	for _, code := range []string{"%i", "%c", "%k"} {
		cmdline = strings.ReplaceAll(cmdline, code, "")
	}

	// Parse the command line
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Empty command line for browser %s\n", b.Name)
		return
	}

	// When running in Flatpak, wrap with flatpak-spawn --host
	if os.Getenv("FLATPAK_ID") != "" && !strings.HasPrefix(parts[0], "flatpak-spawn") {
		parts = append([]string{"flatpak-spawn", "--host"}, parts...)
	}

	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser: %v\n", err)
		return
	}

	// Clean up asynchronously
	go cmd.Wait()
}

// launchBrowserAction launches a browser with a specific desktop file action (e.g., "new-private-window")
func launchBrowserAction(b *Browser, action DesktopAction, url string) {
	cmdline := action.Exec
	if cmdline == "" {
		fmt.Fprintf(os.Stderr, "Error: No exec line for action %s\n", action.ID)
		return
	}

	// Replace %u, %U, %f, %F with URL
	cmdline = strings.ReplaceAll(cmdline, "%u", url)
	cmdline = strings.ReplaceAll(cmdline, "%U", url)
	cmdline = strings.ReplaceAll(cmdline, "%f", url)
	cmdline = strings.ReplaceAll(cmdline, "%F", url)

	// Remove other field codes
	for _, code := range []string{"%i", "%c", "%k"} {
		cmdline = strings.ReplaceAll(cmdline, code, "")
	}

	// Parse the command line
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Empty command line for action %s\n", action.ID)
		return
	}

	// When running in Flatpak, wrap with flatpak-spawn --host
	if os.Getenv("FLATPAK_ID") != "" && !strings.HasPrefix(parts[0], "flatpak-spawn") {
		parts = append([]string{"flatpak-spawn", "--host"}, parts...)
	}

	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser action: %v\n", err)
		return
	}

	// Clean up asynchronously
	go cmd.Wait()
}
