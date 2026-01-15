// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// showAboutDialog displays the application's about dialog.
func showAboutDialog(parent *adw.Window) {
	dialog := adw.NewDialog()
	dialog.SetTitle("")
	dialog.SetContentWidth(350)
	dialog.SetContentHeight(450)

	toolbarView := adw.NewToolbarView()

	// Header with close button
	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(true)
	toolbarView.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)
	content.SetMarginTop(12)
	content.SetMarginBottom(24)
	content.SetVAlign(gtk.AlignCenter)
	content.SetHAlign(gtk.AlignCenter)

	// App icon
	icon := gtk.NewImageFromIconName(getAppID())
	icon.SetPixelSize(96)
	content.Append(icon)

	// App name
	nameLabel := gtk.NewLabel("Switchyard")
	nameLabel.AddCSSClass("title-1")
	content.Append(nameLabel)

	// Developer
	devLabel := gtk.NewLabel("Aly Raffauf")
	devLabel.AddCSSClass("dim-label")
	content.Append(devLabel)

	// Version badge
	versionBtn := gtk.NewButton()
	versionBtn.SetLabel("0.1.0")
	versionBtn.AddCSSClass("pill")
	versionBtn.AddCSSClass("suggested-action")
	versionBtn.SetCanFocus(false)
	versionBtn.SetSensitive(false)
	content.Append(versionBtn)

	// Spacer
	spacer := gtk.NewBox(gtk.OrientationVertical, 0)
	spacer.SetVExpand(true)
	content.Append(spacer)

	// Links group
	linksGroup := adw.NewPreferencesGroup()

	websiteRow := adw.NewActionRow()
	websiteRow.SetTitle("Website")
	websiteRow.SetActivatable(true)
	websiteRow.AddSuffix(gtk.NewImageFromIconName("external-link-symbolic"))
	websiteRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher("https://github.com/alyraffauf/switchyard")
		launcher.Launch(context.Background(), &parent.Window, nil)
	})
	linksGroup.Add(websiteRow)

	issueRow := adw.NewActionRow()
	issueRow.SetTitle("Report an Issue")
	issueRow.SetActivatable(true)
	issueRow.AddSuffix(gtk.NewImageFromIconName("external-link-symbolic"))
	issueRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher("https://github.com/alyraffauf/switchyard/issues")
		launcher.Launch(context.Background(), &parent.Window, nil)
	})
	linksGroup.Add(issueRow)

	content.Append(linksGroup)

	// Copyright and disclaimer
	copyrightLabel := gtk.NewLabel("© 2026 Aly Raffauf")
	copyrightLabel.AddCSSClass("dim-label")
	content.Append(copyrightLabel)

	disclaimerLabel := gtk.NewLabel("This application comes with absolutely no warranty.\nSee the GNU General Public License, version 3 or later for details.")
	disclaimerLabel.AddCSSClass("dim-label")
	disclaimerLabel.AddCSSClass("caption")
	disclaimerLabel.SetJustify(gtk.JustifyCenter)
	disclaimerLabel.SetWrap(true)
	content.Append(disclaimerLabel)

	toolbarView.SetContent(content)
	dialog.SetChild(toolbarView)
	dialog.Present(parent)
}
