// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func createAppearancePage(win *adw.Window, cfg *Config) gtk.Widgetter {
	browsers := detectBrowsers()
	toolbarView, content, _ := settingsPageLayout("Appearance")

	// General section
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("General")

	forceDarkRow := adw.NewSwitchRow()
	forceDarkRow.SetTitle("Force dark mode")
	forceDarkRow.SetSubtitle("Always use dark mode")
	forceDarkRow.SetActive(cfg.ForceDarkMode)
	appearanceGroup.Add(forceDarkRow)

	content.Append(appearanceGroup)

	// Picker Window section
	pickerGroup := adw.NewPreferencesGroup()
	pickerGroup.SetTitle("Picker Window")

	showNamesRow := adw.NewSwitchRow()
	showNamesRow.SetTitle("Show browser names")
	showNamesRow.SetSubtitle("Show browser names below icons")
	showNamesRow.SetActive(cfg.ShowAppNames)
	pickerGroup.Add(showNamesRow)

	// Hidden browsers row
	hiddenBrowsersRow := adw.NewActionRow()
	hiddenBrowsersRow.SetTitle("Hidden browsers")
	hiddenBrowsersRow.SetSubtitle("Choose which browsers to hide from the picker")
	hiddenBrowsersRow.SetActivatable(true)

	chevron := gtk.NewImageFromIconName("go-next-symbolic")
	hiddenBrowsersRow.AddSuffix(chevron)

	hiddenBrowsersRow.ConnectActivated(func() {
		showHiddenBrowsersDialog(win, cfg, browsers)
	})

	pickerGroup.Add(hiddenBrowsersRow)

	content.Append(pickerGroup)

	// Connect change handlers
	forceDarkRow.Connect("notify::active", func() {
		cfg.ForceDarkMode = forceDarkRow.Active()
		saveConfigWithFlag(cfg)
	})

	showNamesRow.Connect("notify::active", func() {
		cfg.ShowAppNames = showNamesRow.Active()
		saveConfigWithFlag(cfg)
	})

	return toolbarView
}
