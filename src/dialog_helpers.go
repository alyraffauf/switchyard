// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// dialogHeader creates a standard dialog header with cancel and action buttons.
// The action button will have the "suggested-action" CSS class applied.
func dialogHeader(cancelLabel, actionLabel string, onCancel, onAction func()) (*adw.HeaderBar, *gtk.Button) {
	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel(cancelLabel)
	cancelBtn.ConnectClicked(func() { onCancel() })
	header.PackStart(cancelBtn)

	actionBtn := gtk.NewButton()
	actionBtn.SetLabel(actionLabel)
	actionBtn.AddCSSClass("suggested-action")
	if onAction != nil {
		actionBtn.ConnectClicked(func() { onAction() })
	}
	header.PackEnd(actionBtn)

	return header, actionBtn
}

// dialogWithToolbar creates a dialog with a standard toolbar and scrolled content area.
// Returns the dialog and a content box where widgets can be added.
func dialogWithToolbar(title string, width, height int, header *adw.HeaderBar) (*adw.Dialog, *gtk.Box, *gtk.ScrolledWindow) {
	dialog := adw.NewDialog()
	dialog.SetTitle(title)
	dialog.SetContentWidth(width)
	dialog.SetContentHeight(height)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(header)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolledWindow.SetVExpand(true)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	scrolledWindow.SetChild(content)
	toolbarView.SetContent(scrolledWindow)
	dialog.SetChild(toolbarView)

	return dialog, content, scrolledWindow
}

// simpleDialogWithToolbar creates a dialog without a scrolled window (for simple, small dialogs).
func simpleDialogWithToolbar(title string, width, height int, header *adw.HeaderBar) (*adw.Dialog, *gtk.Box) {
	dialog := adw.NewDialog()
	dialog.SetTitle(title)
	dialog.SetContentWidth(width)
	dialog.SetContentHeight(height)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)

	return dialog, content
}

// conditionTypeToIndex converts a condition type string to combo row index.
func conditionTypeToIndex(condType string) uint {
	switch condType {
	case "domain":
		return 0
	case "keyword":
		return 1
	case "glob":
		return 2
	case "regex":
		return 3
	default:
		return 0
	}
}

// indexToConditionType converts a combo row index to condition type string.
func indexToConditionType(index uint) string {
	switch index {
	case 0:
		return "domain"
	case 1:
		return "keyword"
	case 2:
		return "glob"
	case 3:
		return "regex"
	default:
		return "domain"
	}
}

// conditionTypeComboRow creates a standardized combo row for selecting condition types.
func conditionTypeComboRow(title string, initialType string) *adw.ComboRow {
	typeRow := adw.NewComboRow()
	typeRow.SetTitle(title)
	typeRow.SetModel(gtk.NewStringList([]string{"Exact Domain", "URL Contains", "Wildcard", "Regex"}))
	typeRow.SetSelected(conditionTypeToIndex(initialType))
	return typeRow
}

// getConditionTypeLabels returns the display labels for condition types in order.
func getConditionTypeLabels() []string {
	return []string{"Exact Domain", "URL Contains", "Wildcard", "Regex"}
}
