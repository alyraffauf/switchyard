// SPDX-License-Identifier: GPL-3.0-or-later

// Package browser models installed web browsers and launches them, free of any
// GTK/GIO types. Discovering browsers and fetching activation tokens is the
// caller's job; this package works in plain strings.
package browser

// Browser is an installed web browser that can open URLs.
type Browser struct {
	ID   string // desktop file ID, e.g. "firefox.desktop"
	Name string
	Icon string // themed icon name; may be empty
	Exec string // desktop-entry command line with field codes, e.g. "firefox %u"
}
