// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"context"
	"fmt"
	"os"

	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/alyraffauf/switchyard/internal/host"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func createAdvancedPage(win *adw.Window, cfg *Config) gtk.Widgetter {
	toolbarView, content, _ := settingsPageLayout("Advanced")

	configGroup := adw.NewPreferencesGroup()
	configGroup.SetTitle("Configuration")

	configRow := adw.NewActionRow()
	configRow.SetTitle("Configuration File")
	configRow.SetSubtitle(appconfig.Path())
	configRow.SetActivatable(true)
	configRow.AddSuffix(gtk.NewImageFromIconName("document-edit-symbolic"))
	configRow.ConnectActivated(func() {
		appconfig.Save(appconfig.Path(), cfg)
		cmd := host.HostCommand("xdg-open", appconfig.Path())
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to open config file: %v\n", err)
		}
		go cmd.Wait()
	})
	configGroup.Add(configRow)

	exportRow := adw.NewActionRow()
	exportRow.SetTitle("Export Configuration")
	exportRow.SetSubtitle("Save configuration to a file")
	exportRow.SetActivatable(true)
	exportRow.AddSuffix(gtk.NewImageFromIconName("document-save-symbolic"))
	exportRow.ConnectActivated(func() {
		dialog := gtk.NewFileDialog()
		dialog.SetTitle("Export Configuration")
		dialog.SetInitialName("switchyard.toml")
		dialog.SetFilters(configFileFilters())

		dialog.Save(context.Background(), &win.Window, func(res gio.AsyncResulter) {
			file, err := dialog.SaveFinish(res)
			if err != nil || file == nil {
				return
			}
			path := file.Path()
			if path == "" {
				return
			}
			if err := appconfig.Export(path, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to export config: %v\n", err)
			}
		})
	})
	configGroup.Add(exportRow)

	importRow := adw.NewActionRow()
	importRow.SetTitle("Import Configuration")
	importRow.SetSubtitle("Load configuration from a file")
	importRow.SetActivatable(true)
	importRow.AddSuffix(gtk.NewImageFromIconName("document-open-symbolic"))
	importRow.ConnectActivated(func() {
		fileDialog := gtk.NewFileDialog()
		fileDialog.SetTitle("Import Configuration")
		fileDialog.SetFilters(configFileFilters())

		fileDialog.Open(context.Background(), &win.Window, func(res gio.AsyncResulter) {
			file, err := fileDialog.OpenFinish(res)
			if err != nil || file == nil {
				return
			}
			path := file.Path()
			if path == "" {
				return
			}

			warnDialog := adw.NewAlertDialog("Import Configuration?", "This will replace your current settings and rules.")
			warnDialog.AddResponse("cancel", "Cancel")
			warnDialog.AddResponse("import", "Import")
			warnDialog.SetResponseAppearance("import", adw.ResponseDestructive)
			warnDialog.SetDefaultResponse("cancel")
			warnDialog.SetCloseResponse("cancel")

			warnDialog.ConnectResponse(func(response string) {
				if response == "import" {
					imported, err := appconfig.Import(path)
					if err == nil {
						*cfg = *imported
						err = appconfig.Save(appconfig.Path(), cfg)
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to import config: %v\n", err)
					}
				}
			})
			warnDialog.Present(win)
		})
	})
	configGroup.Add(importRow)

	content.Append(configGroup)

	extensionGroup := adw.NewPreferencesGroup()
	extensionGroup.SetTitle("Browser Extension")

	firefoxRow := adw.NewActionRow()
	firefoxRow.SetTitle("Firefox Extension")
	firefoxRow.SetSubtitle("Open links in Switchyard directly from Firefox")
	firefoxRow.SetActivatable(true)
	firefoxRow.AddSuffix(gtk.NewImageFromIconName("adw-external-link-symbolic"))
	firefoxRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher(FirefoxExtensionURL)
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	extensionGroup.Add(firefoxRow)

	chromeRow := adw.NewActionRow()
	chromeRow.SetTitle("Chrome Extension")
	chromeRow.SetSubtitle("Open links in Switchyard directly from Chrome")
	chromeRow.SetActivatable(true)
	chromeRow.AddSuffix(gtk.NewImageFromIconName("adw-external-link-symbolic"))
	chromeRow.ConnectActivated(func() {
		launcher := gtk.NewURILauncher(ChromeExtensionURL)
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	extensionGroup.Add(chromeRow)

	content.Append(extensionGroup)

	return toolbarView
}
