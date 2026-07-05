// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func dialogHeader(cancelLabel, actionLabel string, onCancel, onAction func()) (*adw.HeaderBar, *gtk.Button) {
	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel(cancelLabel)
	cancelBtn.ConnectClicked(func() { onCancel() })
	header.PackStart(cancelBtn)

	actionButton := gtk.NewButton()
	actionButton.SetLabel(actionLabel)
	actionButton.AddCSSClass("suggested-action")
	if onAction != nil {
		actionButton.ConnectClicked(func() { onAction() })
	}
	header.PackEnd(actionButton)

	return header, actionButton
}

func dialogWithToolbar(title string, width, height int, header *adw.HeaderBar) (*adw.Dialog, *gtk.Box, *gtk.ScrolledWindow) {
	dialog := adw.NewDialog()
	dialog.SetTitle(title)
	dialog.SetContentWidth(width)
	dialog.SetContentHeight(height)
	dialog.SetCanClose(true)

	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(header)

	scrolledWindow := createScrolledWindow()

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

func conditionTypeToIndex(conditionType string) uint {
	switch conditionType {
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

func redirectionTypeToIndex(redirectionType string) uint {
	switch redirectionType {
	case "wildcard":
		return 1
	case "regex":
		return 2
	default:
		return 0
	}
}

func indexToRedirectionType(index uint) string {
	switch index {
	case 1:
		return "wildcard"
	case 2:
		return "regex"
	default:
		return "domain"
	}
}

func getRedirectionTypeLabels() []string {
	return []string{"Domain", "Wildcard", "Regex"}
}
