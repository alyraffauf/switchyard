// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
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

// settingsPageLayout creates the standard layout for settings pages.
// Returns the toolbar view, content box for adding widgets, and the header bar.
func settingsPageLayout(title string) (*adw.ToolbarView, *gtk.Box, *adw.HeaderBar) {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel(title)
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)
	toolbarView.AddTopBar(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(24)
	content.SetMarginBottom(24)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(content)
	scrolled.SetChild(clamp)

	toolbarView.SetContent(scrolled)
	return toolbarView, content, header
}

// createScrolledWindow creates a standard scrolled window for list content.
func createScrolledWindow() *gtk.ScrolledWindow {
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	return scrolled
}

// createBoxedListBox creates a ListBox with standard boxed-list styling.
func createBoxedListBox() *gtk.ListBox {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	return listBox
}

// createEmptyState creates a status page for empty list states.
func createEmptyState(iconName, title, description string) *adw.StatusPage {
	emptyState := adw.NewStatusPage()
	emptyState.SetIconName(iconName)
	emptyState.SetTitle(title)
	emptyState.SetDescription(description)
	emptyState.SetVExpand(true)
	return emptyState
}

// clearListBox removes all children from a ListBox.
func clearListBox(listBox *gtk.ListBox) {
	for {
		child := listBox.FirstChild()
		if child == nil {
			break
		}
		listBox.Remove(child)
	}
}

// configFileFilters creates the standard file filter list for config import/export.
func configFileFilters() *gio.ListStore {
	tomlFilter := gtk.NewFileFilter()
	tomlFilter.SetName("TOML files")
	tomlFilter.AddPattern("*.toml")

	allFilter := gtk.NewFileFilter()
	allFilter.SetName("All files")
	allFilter.AddPattern("*")

	filters := gio.NewListStore(glib.TypeObject)
	filters.Append(tomlFilter.Object)
	filters.Append(allFilter.Object)
	return filters
}
