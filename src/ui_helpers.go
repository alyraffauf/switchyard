// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// loadBrowserIcon loads a browser icon using GIcon for best quality.
// Using GIcon allows GTK to select the optimal icon size from the theme,
// avoiding blurry scaling that occurs with named icons.
func loadBrowserIcon(browser *Browser, size int) *gtk.Image {
	// Try to use GIcon from AppInfo for best quality
	if browser.AppInfo != nil {
		if gicon := browser.AppInfo.Icon(); gicon != nil {
			image := gtk.NewImageFromGIcon(gicon)
			image.SetPixelSize(size)
			return image
		}
	}

	// Fallback to icon name
	iconName := browser.Icon
	if iconName == "" {
		iconName = "web-browser-symbolic"
	}

	image := gtk.NewImageFromIconName(iconName)
	image.SetPixelSize(size)
	return image
}

// getLogicFromComboRow extracts the logic string from a combo row selection
func getLogicFromComboRow(logicRow *adw.ComboRow) string {
	if logicRow.Selected() == 1 {
		return "any"
	}
	return "all"
}

// saveConfigWithFlag saves config while setting the global saving flag to prevent file watcher loops
func saveConfigWithFlag(cfg *Config) {
	savingMux.Lock()
	isSaving = true
	savingMux.Unlock()
	saveConfig(cfg)
	glib.TimeoutAdd(100, func() bool {
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
		return false
	})
}

// findBrowserByID finds a browser by its desktop file ID
func findBrowserByID(browsers []*Browser, id string) *Browser {
	for _, b := range browsers {
		if b.ID == id {
			return b
		}
	}
	return nil
}
