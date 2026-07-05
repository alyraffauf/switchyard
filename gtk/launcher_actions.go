// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"context"
	"fmt"
	"strings"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showBrowserActionsMenu(btn *gtk.Button, browser *Browser, url string) {
	actions := appbrowser.ListDesktopActions(browser.ID)
	if len(actions) == 0 {
		return
	}

	menu := gio.NewMenu()
	for _, action := range actions {
		menu.Append(action.Name, fmt.Sprintf("win.launch-action::%s:%s", browser.ID, action.ID))
	}

	popover := gtk.NewPopoverMenuFromModel(menu)
	popover.SetParent(btn)
	popover.Popup()
}

// showShortcutsDialog displays available keyboard shortcuts
func showShortcutsDialog(parent *adw.Window) {
	dialog := adw.NewAlertDialog(
		"Keyboard Shortcuts",
		"Ctrl+1 through Ctrl+9: Select browser 1-9\nEsc: Close launcher",
	)

	dialog.AddResponse("ok", "OK")
	dialog.SetDefaultResponse("ok")
	dialog.SetCloseResponse("ok")

	dialog.Present(parent)
}

func setupLauncherActions(win *adw.Window, app *adw.Application, browsers []*Browser, url string, onClose func()) *gio.SimpleActionGroup {
	actionGroup := gio.NewSimpleActionGroup()

	settingsAction := gio.NewSimpleAction("settings", nil)
	settingsAction.ConnectActivate(func(p *glib.Variant) {
		cfg, _ := appconfig.Load(appconfig.Path())
		showSettingsWindow(app, browsers, cfg)
	})
	actionGroup.AddAction(settingsAction)

	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(p *glib.Variant) {
		showAboutDialog(win)
	})
	actionGroup.AddAction(aboutAction)

	donateAction := gio.NewSimpleAction("donate", nil)
	donateAction.ConnectActivate(func(p *glib.Variant) {
		launcher := gtk.NewURILauncher(DonateURL)
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	actionGroup.AddAction(donateAction)

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		onClose()
	})
	actionGroup.AddAction(quitAction)

	shortcutsAction := gio.NewSimpleAction("shortcuts", nil)
	shortcutsAction.ConnectActivate(func(p *glib.Variant) {
		showShortcutsDialog(win)
	})
	actionGroup.AddAction(shortcutsAction)

	launchActionAction := gio.NewSimpleAction("launch-action", glib.NewVariantType("s"))
	launchActionAction.ConnectActivate(func(param *glib.Variant) {
		if param == nil {
			return
		}

		actionSpec := param.String()
		parts := strings.Split(actionSpec, ":")
		if len(parts) != 2 {
			return
		}

		browserID := parts[0]
		actionID := parts[1]

		var selectedBrowser *Browser
		for _, b := range browsers {
			if b.ID == browserID {
				selectedBrowser = b
				break
			}
		}

		if selectedBrowser == nil {
			return
		}

		actions := appbrowser.ListDesktopActions(selectedBrowser.ID)
		for _, action := range actions {
			if action.ID == actionID {
				launchBrowserAction(selectedBrowser, action, url)
				onClose()
				return
			}
		}
	})
	actionGroup.AddAction(launchActionAction)

	return actionGroup
}
