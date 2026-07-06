// SPDX-License-Identifier: GPL-3.0-or-later

// Package browserscan discovers installed HTTP(S) browsers by parsing the
// .desktop files on the XDG data search path. It has no GTK/GIO dependency.
package browserscan

import (
	"cmp"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alyraffauf/goxdgdesktop/desktopfile"
	"github.com/alyraffauf/goxdgdesktop/xdg"
	"github.com/alyraffauf/switchyard/internal/browser"
)

// A browser switcher must never list itself, across all packaged variants.
const selfIDPrefix = "io.github.alyraffauf.Switchyard"

var httpSchemeHandlers = []string{
	"x-scheme-handler/http",
	"x-scheme-handler/https",
}

// applicationsDirs is a variable so tests can override the search path.
var applicationsDirs = xdg.ApplicationsDirs

// Installed returns the installed HTTP(S) browsers, sorted by Name. Switchyard's
// own entries are excluded; filtering config-hidden browsers is the caller's job.
func Installed() []browser.Browser {
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

			if strings.HasPrefix(id, selfIDPrefix) {
				return nil
			}

			if parsed, ok := parseBrowser(id, path); ok {
				browsers = append(browsers, parsed)
			}
			return nil
		})
	}

	slices.SortFunc(browsers, func(first, second browser.Browser) int {
		return cmp.Compare(first.Name, second.Name)
	})

	return browsers
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

// parseBrowser returns the Browser at path, or ok=false if it isn't a
// displayable HTTP(S) browser.
func parseBrowser(id, path string) (browser.Browser, bool) {
	file, err := desktopfile.Read(path)
	if err != nil {
		return browser.Browser{}, false
	}

	if typ, _ := file.Get(desktopfile.EntrySection, "Type"); typ != "Application" {
		return browser.Browser{}, false
	}
	if isTrue(file, "Hidden") || isTrue(file, "NoDisplay") {
		return browser.Browser{}, false
	}
	if !handlesHTTP(file) {
		return browser.Browser{}, false
	}

	name, _ := file.Get(desktopfile.EntrySection, "Name")
	icon, _ := file.Get(desktopfile.EntrySection, "Icon")
	exec, _ := file.Get(desktopfile.EntrySection, "Exec")

	return browser.Browser{
		ID:   id,
		Name: name,
		Icon: icon,
		Exec: exec,
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
