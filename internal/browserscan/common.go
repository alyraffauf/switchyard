// SPDX-License-Identifier: GPL-3.0-or-later

package browserscan

import (
	"cmp"
	"slices"
	"strings"

	"github.com/alyraffauf/switchyard/internal/browser"
)

// A browser switcher must never list itself, across all packaged variants.
const selfIDPrefix = "io.github.alyraffauf.Switchyard"

// An Option configures how the scanner treats .desktop entries. Pass Options to
// Installed and Find to override the defaults.
type Option func(*scanOptions)

// IncludeNoDisplay controls whether entries with NoDisplay=true are returned. It
// defaults to true.
func IncludeNoDisplay(include bool) Option {
	return func(options *scanOptions) { options.includeNoDisplay = include }
}

// scanOptions holds the parser behavior toggles set via Option values.
type scanOptions struct {
	includeNoDisplay bool
}

// this package can opt out with IncludeNoDisplay(false).
func defaultScanOptions() scanOptions {
	return scanOptions{includeNoDisplay: true}
}

func newScanOptions(opts []Option) scanOptions {
	options := defaultScanOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// isSelf reports whether id belongs to Switchyard itself.
// The id is a desktop file ID on Linux and a bundle identifier on macOS.
func isSelf(id string) bool {
	return id == "" || strings.HasPrefix(id, selfIDPrefix)
}

// sortByName sorts browsers in place.
func sortByName(browsers []browser.Browser) {
	slices.SortFunc(browsers, func(first, second browser.Browser) int {
		return cmp.Compare(first.Name, second.Name)
	})
}
