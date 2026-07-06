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
