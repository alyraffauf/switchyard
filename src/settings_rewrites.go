// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func formatRewriteSubtitle(r *Rewrite) string {
	rwType := r.Type
	if rwType == "" {
		rwType = "domain"
	}

	var typeLabel string
	if rwType == "domain" {
		typeLabel = "Domain"
	} else {
		typeLabel = "URL"
	}

	if r.Replace == "" {
		return fmt.Sprintf("%s · Remove", typeLabel)
	}
	return fmt.Sprintf("%s · %s", typeLabel, r.Replace)
}

func createRewritesPage(win *adw.Window, cfg *Config) gtk.Widgetter {
	toolbarView := adw.NewToolbarView()

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(true)
	titleLabel := gtk.NewLabel("URL Rewrites")
	titleLabel.AddCSSClass("title")
	header.SetTitleWidget(titleLabel)

	// Add Rewrite button in header
	addButton := gtk.NewButton()
	addButton.SetIconName("list-add-symbolic")
	addButton.SetTooltipText("Add URL Rewrite")
	addButton.SetHasFrame(false)
	header.PackEnd(addButton)

	toolbarView.AddTopBar(header)

	// Scrolled window for rewrites list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)
	content.SetMarginTop(12)
	content.SetMarginBottom(12)


	rewritesListBox := gtk.NewListBox()
	rewritesListBox.SetSelectionMode(gtk.SelectionNone)
	rewritesListBox.AddCSSClass("boxed-list")

	emptyState := adw.NewStatusPage()
	emptyState.SetIconName("edit-find-replace-symbolic")
	emptyState.SetTitle("No Rewrites")
	emptyState.SetDescription("Modify URLs before launching a browser")
	emptyState.SetVExpand(true)

	var rebuildRewritesList func()

	createRewriteRow := func(rewriteIndex int) *adw.ActionRow {
		rewrite := &cfg.Rewrites[rewriteIndex]

		row := adw.NewActionRow()
		row.SetTitle(rewrite.Find)
		row.SetSubtitle(formatRewriteSubtitle(rewrite))
		row.SetActivatable(true)

		icon := gtk.NewImageFromIconName("edit-find-replace-symbolic")
		icon.SetPixelSize(24)
		row.AddPrefix(icon)

		reorderBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
		reorderBox.SetVAlign(gtk.AlignCenter)

		upBtn := gtk.NewButton()
		upBtn.SetIconName("go-up-symbolic")
		upBtn.AddCSSClass("flat")
		upBtn.SetSensitive(rewriteIndex > 0)
		upBtn.SetTooltipText("Move up")
		upBtn.ConnectClicked(func() {
			if rewriteIndex > 0 {
				cfg.Rewrites[rewriteIndex], cfg.Rewrites[rewriteIndex-1] = cfg.Rewrites[rewriteIndex-1], cfg.Rewrites[rewriteIndex]
				saveConfig(cfg)
				rebuildRewritesList()
			}
		})
		reorderBox.Append(upBtn)

		downBtn := gtk.NewButton()
		downBtn.SetIconName("go-down-symbolic")
		downBtn.AddCSSClass("flat")
		downBtn.SetSensitive(rewriteIndex < len(cfg.Rewrites)-1)
		downBtn.SetTooltipText("Move down")
		downBtn.ConnectClicked(func() {
			if rewriteIndex < len(cfg.Rewrites)-1 {
				cfg.Rewrites[rewriteIndex], cfg.Rewrites[rewriteIndex+1] = cfg.Rewrites[rewriteIndex+1], cfg.Rewrites[rewriteIndex]
				saveConfig(cfg)
				rebuildRewritesList()
			}
		})
		reorderBox.Append(downBtn)

		row.AddSuffix(reorderBox)

		// Delete button
		deleteBtn := gtk.NewButton()
		deleteBtn.SetIconName("edit-delete-symbolic")
		deleteBtn.AddCSSClass("flat")
		deleteBtn.AddCSSClass("destructive-action")
		deleteBtn.SetTooltipText("Delete rewrite")
		deleteBtn.ConnectClicked(func() {
			cfg.Rewrites = append(cfg.Rewrites[:rewriteIndex], cfg.Rewrites[rewriteIndex+1:]...)
			saveConfig(cfg)
			rebuildRewritesList()
		})
		row.AddSuffix(deleteBtn)

		// Edit on click
		row.ConnectActivated(func() {
			showEditRewriteDialog(win, cfg, rewrite, rebuildRewritesList)
		})

		return row
	}

	rebuildRewritesList = func() {
		for {
			child := rewritesListBox.FirstChild()
			if child == nil {
				break
			}
			rewritesListBox.Remove(child)
		}

		// handle empty state
		if len(cfg.Rewrites) == 0 {
			rewritesListBox.SetVisible(false)
			emptyState.SetVisible(true)
		} else {
			rewritesListBox.SetVisible(true)
			emptyState.SetVisible(false)

			for i := range cfg.Rewrites {
				row := createRewriteRow(i)
				rewritesListBox.Append(row)
			}
		}
	}

	rebuildRewritesList()

	content.Append(rewritesListBox)
	content.Append(emptyState)
	scrolled.SetChild(content)
	toolbarView.SetContent(scrolled)

	addButton.ConnectClicked(func() {
		showAddRewriteDialog(win, cfg, rebuildRewritesList)
	})

	return toolbarView
}
