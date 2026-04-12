// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showSettingsWindow(app *adw.Application, browsers []*Browser, cfg *Config) {
	win := adw.NewWindow()
	win.SetTitle("Switchyard")
	win.SetApplication(&app.Application)
	win.SetDefaultSize(900, 600)
	win.SetResizable(true)
	win.SetSizeRequest(700, 500)

	setupAppActions(app, win)

	// Navigation split view
	splitView := adw.NewNavigationSplitView()
	splitView.SetShowContent(true)
	splitView.SetMinSidebarWidth(200)
	splitView.SetMaxSidebarWidth(200)

	// Sidebar
	sidebarPage := adw.NewNavigationPage(createSidebar(win, cfg, browsers, splitView), "Switchyard")
	splitView.SetSidebar(sidebarPage)

	// Initial content
	contentPage := adw.NewNavigationPage(createAppearancePage(win, browsers, cfg), "Appearance")
	splitView.SetContent(contentPage)

	win.SetContent(splitView)

	if cfg.CheckDefaultBrowser && !isDefaultBrowser() {
		showDefaultBrowserPrompt(win, cfg, func() {})
	}

	watchConfigFile(cfg, func() {})

	win.Present()
}

func setupAppActions(app *adw.Application, win *adw.Window) {
	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(p *glib.Variant) {
		showAboutDialog(win)
	})
	app.AddAction(aboutAction)

	donateAction := gio.NewSimpleAction("donate", nil)
	donateAction.ConnectActivate(func(p *glib.Variant) {
		launcher := gtk.NewURILauncher(DonateURL)
		launcher.Launch(context.Background(), &win.Window, nil)
	})
	app.AddAction(donateAction)

	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(p *glib.Variant) {
		win.Close()
	})
	app.AddAction(quitAction)

	app.SetAccelsForAction("app.quit", []string{"<Ctrl>q"})
}

func createSidebar(win *adw.Window, cfg *Config, browsers []*Browser, splitView *adw.NavigationSplitView) gtk.Widgetter {
	toolbarView := adw.NewToolbarView()

	// Header with menu
	sidebarHeader := adw.NewHeaderBar()
	sidebarHeader.SetShowEndTitleButtons(false)

	titleLabel := gtk.NewLabel("Switchyard")
	titleLabel.AddCSSClass("title")
	sidebarHeader.SetTitleWidget(titleLabel)

	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetTooltipText("Main Menu")

	menu := gio.NewMenu()
	menu.Append("Donate ❤️", "app.donate")
	menu.Append("About", "app.about")

	quitSection := gio.NewMenu()
	quitSection.Append("Quit", "app.quit")
	menu.AppendSection("", quitSection)

	menuBtn.SetMenuModel(menu)
	sidebarHeader.PackEnd(menuBtn)

	toolbarView.AddTopBar(sidebarHeader)

	// Navigation list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)

	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	// Navigation rows
	appearanceRow := adw.NewActionRow()
	appearanceRow.SetTitle("Appearance")
	appearanceRow.AddPrefix(gtk.NewImageFromIconName("applications-graphics-symbolic"))
	listBox.Append(appearanceRow)

	behaviorRow := adw.NewActionRow()
	behaviorRow.SetTitle("Behavior")
	behaviorRow.AddPrefix(gtk.NewImageFromIconName("preferences-system-symbolic"))
	listBox.Append(behaviorRow)

	redirectionsRow := adw.NewActionRow()
	redirectionsRow.SetTitle("Link Redirections")
	redirectionsRow.AddPrefix(gtk.NewImageFromIconName("edit-find-replace-symbolic"))
	listBox.Append(redirectionsRow)

	rulesRow := adw.NewActionRow()
	rulesRow.SetTitle("Browser Rules")
	rulesRow.AddPrefix(gtk.NewImageFromIconName("view-list-symbolic"))
	listBox.Append(rulesRow)

	advancedRow := adw.NewActionRow()
	advancedRow.SetTitle("Advanced")
	advancedRow.AddPrefix(gtk.NewImageFromIconName("preferences-other-symbolic"))
	listBox.Append(advancedRow)

	listBox.SetHeaderFunc(func(row, before *gtk.ListBoxRow) {
		if row.Index() == 2 || row.Index() == 4 {
			row.SetHeader(gtk.NewSeparator(gtk.OrientationHorizontal))
		}
	})

	scrolled.SetChild(listBox)
	toolbarView.SetContent(scrolled)

	listBox.SelectRow(listBox.RowAtIndex(0))

	// Navigation handler
	navigateToPage := func(index int) {
		var page gtk.Widgetter
		var title string

		switch index {
		case 0:
			page = createAppearancePage(win, browsers, cfg)
			title = "Appearance"
		case 1:
			page = createBehaviorPage(win, browsers, cfg)
			title = "Behavior"
		case 2:
			page = createRedirectionsPage(win, cfg)
			title = "Link Redirections"
		case 3:
			page = createRulesPage(win, browsers, cfg)
			title = "Browser Rules"
		case 4:
			page = createAdvancedPage(win, cfg)
			title = "Advanced"
		}

		if page != nil {
			contentPage := adw.NewNavigationPage(page, title)
			splitView.SetContent(contentPage)
		}
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		navigateToPage(row.Index())
	})

	listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row != nil {
			navigateToPage(row.Index())
		}
	})

	return toolbarView
}

var (
	configMonitor *gio.FileMonitor
)

func watchConfigFile(cfg *Config, onChange func()) {
	configFile := gio.NewFileForPath(configPath())
	monitorIface, err := configFile.Monitor(context.Background(), gio.FileMonitorNone)
	if err == nil && monitorIface != nil {
		configMonitor = gio.BaseFileMonitor(monitorIface)
		if configMonitor != nil {
			configMonitor.ConnectChanged(func(file, otherFile gio.Filer, eventType gio.FileMonitorEvent) {
				if eventType == gio.FileMonitorEventChanged || eventType == gio.FileMonitorEventCreated {
					savingMux.Lock()
					saving := isSaving
					savingMux.Unlock()
					if saving {
						return
					}

					newCfg := loadConfig()
					cfg.PromptOnClick = newCfg.PromptOnClick
					cfg.FavoriteBrowser = newCfg.FavoriteBrowser
					cfg.CheckDefaultBrowser = newCfg.CheckDefaultBrowser
					cfg.ShowAppNames = newCfg.ShowAppNames
					cfg.ForceDarkMode = newCfg.ForceDarkMode
					cfg.StayAlive = newCfg.StayAlive
					cfg.SanitizeLinks = newCfg.SanitizeLinks
					cfg.HiddenBrowsers = newCfg.HiddenBrowsers
					cfg.Redirections = newCfg.Redirections
					cfg.Rules = newCfg.Rules

					if onChange != nil {
						onChange()
					}
				}
			})
		}
	}
}
