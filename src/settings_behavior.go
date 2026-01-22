// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func createBehaviorPage(win *adw.Window, cfg *Config, browsers []*Browser) gtk.Widgetter {
	toolbarView, content, _ := settingsPageLayout("Behavior")

	// General Behavior section
	behaviorGroup := adw.NewPreferencesGroup()
	behaviorGroup.SetTitle("General")

	checkDefaultRow := adw.NewSwitchRow()
	checkDefaultRow.SetTitle("Prompt to set as default browser")
	checkDefaultRow.SetSubtitle("Show prompt on startup if Switchyard is not the default browser")
	checkDefaultRow.SetActive(cfg.CheckDefaultBrowser)
	behaviorGroup.Add(checkDefaultRow)

	promptRow := adw.NewSwitchRow()
	promptRow.SetTitle("Show picker when no rule matches")
	promptRow.SetSubtitle("Let you choose a browser for unmatched URLs")
	promptRow.SetActive(cfg.PromptOnClick)
	behaviorGroup.Add(promptRow)

	// Favorite browser dropdown
	browserNames := make([]string, len(browsers)+1)
	browserNames[0] = "None"
	for i, b := range browsers {
		browserNames[i+1] = b.Name
	}
	browserList := gtk.NewStringList(browserNames)

	defaultRow := adw.NewComboRow()
	defaultRow.SetTitle("Favorite browser")
	defaultRow.SetSubtitle("Appears first in picker and opens when picker is disabled")
	defaultRow.SetModel(browserList)

	// Set initial selection
	selectedIndex := uint(0)
	for i, b := range browsers {
		if b.ID == cfg.FavoriteBrowser {
			selectedIndex = uint(i + 1)
			break
		}
	}
	defaultRow.SetSelected(selectedIndex)

	behaviorGroup.Add(defaultRow)
	content.Append(behaviorGroup)

	// Connect change handlers
	checkDefaultRow.Connect("notify::active", func() {
		cfg.CheckDefaultBrowser = checkDefaultRow.Active()
		saveConfigWithFlag(cfg)
	})

	promptRow.Connect("notify::active", func() {
		cfg.PromptOnClick = promptRow.Active()
		saveConfigWithFlag(cfg)
	})

	defaultRow.Connect("notify::selected", func() {
		idx := defaultRow.Selected()
		if idx == 0 {
			cfg.FavoriteBrowser = ""
		} else if idx > 0 && int(idx) <= len(browsers) {
			cfg.FavoriteBrowser = browsers[idx-1].ID
		}
		saveConfigWithFlag(cfg)
	})

	return toolbarView
}
