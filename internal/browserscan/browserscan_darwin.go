// SPDX-License-Identifier: GPL-3.0-or-later

//go:build darwin

// Package browserscan discovers installed HTTP(S) browsers via NSWorkspace
// (Launch Services), then reads each candidate's Info.plist for display
// metadata.
package browserscan

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

char *browserscan_nsworkspace_app_paths(void);
*/
import "C"

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/alyraffauf/switchyard/internal/browser"
	"howett.net/plist"
)

// appPaths is a variable so tests can stub app discovery without touching
// the real filesystem or Launch Services.
var appPaths = nsWorkspaceAppPaths

// nsWorkspaceAppPaths lists every app registered to open https:// URLs, via
// NSWorkspace/Launch Services (see nsworkspace_darwin.m).
func nsWorkspaceAppPaths() []string {
	rawPaths := C.browserscan_nsworkspace_app_paths()
	if rawPaths == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(rawPaths))

	return strings.Split(C.GoString(rawPaths), "\n")
}

type bundleInfo struct {
	BundleIdentifier  string          `plist:"CFBundleIdentifier"`
	BundleName        string          `plist:"CFBundleName"`
	BundleDisplayName string          `plist:"CFBundleDisplayName"`
	BundleIconFile    string          `plist:"CFBundleIconFile"`
	BundleIconName    string          `plist:"CFBundleIconName"`
	URLTypes          []bundleURLType `plist:"CFBundleURLTypes"`
	UIElement         bool            `plist:"LSUIElement"`
	BackgroundOnly    bool            `plist:"LSBackgroundOnly"`
}

type bundleURLType struct {
	Schemes []string `plist:"CFBundleURLSchemes"`
	Role    string   `plist:"CFBundleTypeRole"`
	Rank    string   `plist:"LSHandlerRank"`
}

type parsedBrowserResult struct {
	browser browser.Browser
	ok      bool
}

// lsHandlerRankNone marks a URL type as opted out of being a handler candidate.
const lsHandlerRankNone = "None"

// Installed returns the installed HTTP(S) browsers, sorted by Name. Switchyard's
// own entries are excluded; filtering config-hidden browsers is the caller's job.
func Installed() []browser.Browser {
	browsers := scanBrowsers()
	sortByName(browsers)

	return browsers
}

// Find returns a displayable HTTP(S) browser by bundle identifier.
func Find(id string) (browser.Browser, bool) {
	if isSelf(id) {
		return browser.Browser{}, false
	}

	for _, installedBrowser := range scanBrowsers() {
		if installedBrowser.ID == id {
			return installedBrowser, true
		}
	}

	return browser.Browser{}, false
}

func scanBrowsers() []browser.Browser {
	discoveredAppPaths := appPaths()
	results := make([]parsedBrowserResult, len(discoveredAppPaths))

	var waitGroup sync.WaitGroup
	for index, appPath := range discoveredAppPaths {
		waitGroup.Add(1)

		go func(index int, appPath string) {
			defer waitGroup.Done()

			parsedBrowser, ok := parseBrowser(appPath)
			results[index] = parsedBrowserResult{
				browser: parsedBrowser,
				ok:      ok,
			}
		}(index, appPath)
	}
	waitGroup.Wait()

	browsers := make([]browser.Browser, 0, len(results))
	seen := map[string]bool{}

	for _, result := range results {
		if !result.ok || seen[result.browser.ID] {
			continue
		}
		seen[result.browser.ID] = true
		browsers = append(browsers, result.browser)
	}

	return browsers
}

// ListDesktopActions returns app-specific launch actions for appID. macOS app
// bundles have no generic mechanism analogous to XDG desktop actions, so this
// always returns nil.
func ListDesktopActions(appID string) []browser.Action {
	return nil
}

// parseBrowser returns the Browser at appPath, or ok=false if it isn't a
// displayable HTTP(S) browser.
func parseBrowser(appPath string) (browser.Browser, bool) {
	data, err := os.ReadFile(filepath.Join(appPath, "Contents", "Info.plist"))
	if err != nil {
		return browser.Browser{}, false
	}

	var info bundleInfo
	if _, err := plist.Unmarshal(data, &info); err != nil {
		return browser.Browser{}, false
	}

	if isSelf(info.BundleIdentifier) {
		return browser.Browser{}, false
	}
	if info.UIElement || info.BackgroundOnly {
		return browser.Browser{}, false
	}
	if !handlesHTTP(info.URLTypes) {
		return browser.Browser{}, false
	}

	return browser.Browser{
		ID:   info.BundleIdentifier,
		Name: displayName(appPath, info),
		Icon: iconPath(appPath, info),
		Exec: appPath, // a bundle path here, not a shell command line
	}, true
}

// handlesHTTP requires both http and https, with neither opted out via lsHandlerRankNone
func handlesHTTP(types []bundleURLType) bool {
	var hasHTTP, hasHTTPS bool
	for _, urlType := range types {
		if strings.EqualFold(urlType.Role, lsHandlerRankNone) || strings.EqualFold(urlType.Rank, lsHandlerRankNone) {
			continue
		}
		if slices.Contains(urlType.Schemes, "http") {
			hasHTTP = true
		}
		if slices.Contains(urlType.Schemes, "https") {
			hasHTTPS = true
		}
	}
	return hasHTTP && hasHTTPS
}

func displayName(appPath string, info bundleInfo) string {
	if info.BundleDisplayName != "" {
		return info.BundleDisplayName
	}
	if info.BundleName != "" {
		return info.BundleName
	}
	return strings.TrimSuffix(filepath.Base(appPath), ".app")
}

func iconPath(appPath string, info bundleInfo) string {
	icon := info.BundleIconFile
	if icon == "" {
		icon = info.BundleIconName
	}
	if icon == "" {
		return ""
	}
	if filepath.Ext(icon) == "" {
		// CFBundleIconFile conventionally omits the extension.
		icon += ".icns"
	}
	return filepath.Join(appPath, "Contents", "Resources", icon)
}
