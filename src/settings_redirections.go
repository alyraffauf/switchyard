// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func formatRedirectionSubtitle(r *Redirection) string {
	rwType := r.Type
	if rwType == "" {
		rwType = "domain"
	}

	var typeLabel string
	if rwType == "domain" {
		typeLabel = "Domain"
	} else {
		typeLabel = "Pattern"
	}

	if r.Replace == "" {
		return fmt.Sprintf("%s · Removes match", typeLabel)
	}
	return fmt.Sprintf("%s · Replaces with %s", typeLabel, r.Replace)
}

func createRedirectionsPage(win *adw.Window, cfg *Config) gtk.Widgetter {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("Link Redirections")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)

	// Add button in header
	addButton := gtk.NewButton()
	addButton.SetIconName("list-add-symbolic")
	addButton.SetTooltipText("Add Redirection")
	addButton.SetHasFrame(false)
	header.PackEnd(addButton)

	toolbarView.AddTopBar(header)

	// Scrolled window for redirections list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)
	content.SetMarginTop(12)
	content.SetMarginBottom(12)

	// Info banner
	infoLabel := gtk.NewLabel("Redirections modify links before rules are applied.")
	infoLabel.SetWrap(true)
	infoLabel.SetXAlign(0)
	infoLabel.AddCSSClass("dim-label")
	infoLabel.SetMarginStart(12)
	infoLabel.SetMarginEnd(12)
	infoLabel.SetMarginBottom(6)
	content.Append(infoLabel)

	redirectionsListBox := gtk.NewListBox()
	redirectionsListBox.SetSelectionMode(gtk.SelectionNone)
	redirectionsListBox.AddCSSClass("boxed-list")

	emptyState := adw.NewStatusPage()
	emptyState.SetIconName("edit-find-replace-symbolic")
	emptyState.SetTitle("No Redirections")
	emptyState.SetDescription("Change links before they open in a browser")
	emptyState.SetVExpand(true)

	var rebuildRedirectionsList func()

	createRedirectionRow := func(redirectionIndex int) *adw.ActionRow {
		redirection := &cfg.Redirections[redirectionIndex]

		row := adw.NewActionRow()
		row.SetTitle(redirection.Find)
		row.SetSubtitle(formatRedirectionSubtitle(redirection))
		row.SetActivatable(true)

		icon := gtk.NewImageFromIconName("edit-find-replace-symbolic")
		icon.SetPixelSize(24)
		row.AddPrefix(icon)

		reorderBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
		reorderBox.SetVAlign(gtk.AlignCenter)

		upBtn := gtk.NewButton()
		upBtn.SetIconName("go-up-symbolic")
		upBtn.AddCSSClass("flat")
		upBtn.SetSensitive(redirectionIndex > 0)
		upBtn.SetTooltipText("Move up")
		upBtn.ConnectClicked(func() {
			if redirectionIndex > 0 {
				cfg.Redirections[redirectionIndex], cfg.Redirections[redirectionIndex-1] = cfg.Redirections[redirectionIndex-1], cfg.Redirections[redirectionIndex]
				saveConfig(cfg)
				rebuildRedirectionsList()
			}
		})
		reorderBox.Append(upBtn)

		downBtn := gtk.NewButton()
		downBtn.SetIconName("go-down-symbolic")
		downBtn.AddCSSClass("flat")
		downBtn.SetSensitive(redirectionIndex < len(cfg.Redirections)-1)
		downBtn.SetTooltipText("Move down")
		downBtn.ConnectClicked(func() {
			if redirectionIndex < len(cfg.Redirections)-1 {
				cfg.Redirections[redirectionIndex], cfg.Redirections[redirectionIndex+1] = cfg.Redirections[redirectionIndex+1], cfg.Redirections[redirectionIndex]
				saveConfig(cfg)
				rebuildRedirectionsList()
			}
		})
		reorderBox.Append(downBtn)

		row.AddSuffix(reorderBox)

		// Delete button
		deleteBtn := gtk.NewButton()
		deleteBtn.SetIconName("edit-delete-symbolic")
		deleteBtn.AddCSSClass("flat")
		deleteBtn.AddCSSClass("destructive-action")
		deleteBtn.SetTooltipText("Remove")
		deleteBtn.ConnectClicked(func() {
			cfg.Redirections = append(cfg.Redirections[:redirectionIndex], cfg.Redirections[redirectionIndex+1:]...)
			saveConfig(cfg)
			rebuildRedirectionsList()
		})
		row.AddSuffix(deleteBtn)

		// Edit on click
		row.ConnectActivated(func() {
			showEditRedirectionDialog(win, cfg, redirection, rebuildRedirectionsList)
		})

		return row
	}

	rebuildRedirectionsList = func() {
		for {
			child := redirectionsListBox.FirstChild()
			if child == nil {
				break
			}
			redirectionsListBox.Remove(child)
		}

		// handle empty state
		if len(cfg.Redirections) == 0 {
			infoLabel.SetVisible(false)
			redirectionsListBox.SetVisible(false)
			emptyState.SetVisible(true)
		} else {
			infoLabel.SetVisible(true)
			redirectionsListBox.SetVisible(true)
			emptyState.SetVisible(false)

			for i := range cfg.Redirections {
				row := createRedirectionRow(i)
				redirectionsListBox.Append(row)
			}
		}
	}

	rebuildRedirectionsList()

	content.Append(redirectionsListBox)
	content.Append(emptyState)
	scrolled.SetChild(content)
	toolbarView.SetContent(scrolled)

	addButton.ConnectClicked(func() {
		showAddRedirectionDialog(win, cfg, rebuildRedirectionsList)
	})

	return toolbarView
}
