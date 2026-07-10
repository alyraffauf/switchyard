// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

// Package browserscan discovers installed HTTP(S) browsers by parsing the
// .desktop files on the XDG data search path. It has no GTK/GIO dependency.
package browserscan

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alyraffauf/goxdgdesktop/desktopfile"
	"github.com/alyraffauf/goxdgdesktop/xdg"
	"github.com/alyraffauf/switchyard/internal/browser"
)

var httpSchemeHandlers = []string{
	"x-scheme-handler/http",
	"x-scheme-handler/https",
}

// applicationsDirs is a variable so tests can override the search path.
var applicationsDirs = xdg.ApplicationsDirs

// Installed returns the installed HTTP(S) browsers, sorted by Name. Switchyard's
// own entries are excluded; filtering config-hidden browsers is the caller's job.
// By default NoDisplay=true browsers are included; pass IncludeNoDisplay(false) to
// hide them.
func Installed(opts ...Option) []browser.Browser {
	options := newScanOptions(opts)

	var browsers []browser.Browser
	seen := map[string]bool{}

	for _, dir := range applicationsDirs() {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Skip missing/unreadable dirs rather than aborting the walk.
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".desktop") {
				return nil
			}

			id := desktopID(dir, path)
			// First dir to define an ID wins. Mark seen even on rejection, so a
			// shadowed copy in a lower-priority dir can't resurface.
			if seen[id] {
				return nil
			}
			seen[id] = true

			if isSelf(id) {
				return nil
			}

			if parsed, ok := parseBrowser(id, path, options); ok {
				browsers = append(browsers, parsed)
			}
			return nil
		})
	}

	sortByName(browsers)

	return browsers
}

// Find returns a displayable HTTP(S) browser by desktop file ID. By default a
// NoDisplay=true entry still matches; pass IncludeNoDisplay(false) to reject it.
func Find(id string, opts ...Option) (browser.Browser, bool) {
	if isSelf(id) {
		return browser.Browser{}, false
	}

	options := newScanOptions(opts)

	for _, dir := range applicationsDirs() {
		path, ok := desktopFilePath(dir, id)
		if !ok {
			continue
		}
		return parseBrowser(id, path, options)
	}

	return browser.Browser{}, false
}

func findDesktopFile(id string) string {
	for _, dir := range applicationsDirs() {
		if path, ok := desktopFilePath(dir, id); ok {
			return path
		}
	}
	return ""
}

func desktopFilePath(dir, id string) (string, bool) {
	path := filepath.Join(dir, id)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}

	var match string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".desktop") {
			return nil
		}
		if desktopID(dir, path) != id {
			return nil
		}
		match = path
		return fs.SkipAll
	})
	return match, match != ""
}

// desktopID is path's desktop-file ID relative to dir, with separators replaced
// by "-" per the spec (sub/foo.desktop -> sub-foo.desktop).
func desktopID(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	return strings.ReplaceAll(filepath.ToSlash(rel), "/", "-")
}

// parseBrowser returns the Browser at path, or ok=false if it isn't an HTTP(S)
// browser the given options accept.
func parseBrowser(id, path string, options scanOptions) (browser.Browser, bool) {
	file, err := desktopfile.Read(path)
	if err != nil {
		return browser.Browser{}, false
	}

	if typ, _ := file.Get(desktopfile.EntrySection, "Type"); typ != "Application" {
		return browser.Browser{}, false
	}
	if isTrue(file, "Hidden") {
		return browser.Browser{}, false
	}
	if !options.includeNoDisplay && isTrue(file, "NoDisplay") {
		return browser.Browser{}, false
	}
	if !handlesHTTP(file) {
		return browser.Browser{}, false
	}

	icon, _ := file.Get(desktopfile.EntrySection, "Icon")
	exec, _ := file.Get(desktopfile.EntrySection, "Exec")

	return browser.Browser{
		ID:      id,
		Name:    localizedString(file, desktopfile.EntrySection, "Name"),
		Icon:    icon,
		Exec:    exec,
		Actions: localizedActions(file),
	}, true
}

func isTrue(file *desktopfile.File, key string) bool {
	value, _ := file.Get(desktopfile.EntrySection, key)
	return strings.EqualFold(value, "true")
}

func handlesHTTP(file *desktopfile.File) bool {
	mimeType, ok := file.Get(desktopfile.EntrySection, "MimeType")
	if !ok {
		return false
	}
	for mime := range strings.SplitSeq(mimeType, ";") {
		if slices.Contains(httpSchemeHandlers, strings.TrimSpace(mime)) {
			return true
		}
	}
	return false
}
