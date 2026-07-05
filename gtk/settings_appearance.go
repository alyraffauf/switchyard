// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func createAppearancePage(win *adw.Window, browsers []*Browser, cfg *Config) gtk.Widgetter {
	toolbarView, content, _ := settingsPageLayout("Appearance")

	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("General")

	forceDarkRow := adw.NewSwitchRow()
	forceDarkRow.SetTitle("Force dark mode")
	forceDarkRow.SetSubtitle("Override system color scheme")
	forceDarkRow.SetActive(cfg.ForceDarkMode)
	appearanceGroup.Add(forceDarkRow)

	content.Append(appearanceGroup)

	launcherGroup := adw.NewPreferencesGroup()
	launcherGroup.SetTitle("Launcher")

	showNamesRow := adw.NewSwitchRow()
	showNamesRow.SetTitle("Show browser names")
	showNamesRow.SetSubtitle("Display names under browser icons")
	showNamesRow.SetActive(cfg.ShowAppNames)
	launcherGroup.Add(showNamesRow)

	hiddenBrowsersRow := adw.NewActionRow()
	hiddenBrowsersRow.SetTitle("Hidden browsers")
	hiddenBrowsersRow.SetSubtitle("Hide browsers you don't use")
	hiddenBrowsersRow.SetActivatable(true)

	chevron := gtk.NewImageFromIconName("go-next-symbolic")
	hiddenBrowsersRow.AddSuffix(chevron)

	hiddenBrowsersRow.ConnectActivated(func() {
		showHiddenBrowsersDialog(win, cfg, browsers)
	})

	launcherGroup.Add(hiddenBrowsersRow)

	content.Append(launcherGroup)

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
