// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"testing"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	widget "github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var launcherSettingsTestID atomic.Uint64

func TestLauncherSettingsCanUnhideBrowser(t *testing.T) {
	if os.Getenv("SWITCHYARD_GTK_TEST") != "1" {
		t.Skip("set SWITCHYARD_GTK_TEST=1 in an isolated graphical session")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	if !widget.InitCheck() {
		t.Fatal("could not initialize GTK")
	}
	adw.Init()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Cleanup(func() {
		if configMonitor != nil {
			configMonitor.Cancel()
			configMonitor = nil
		}
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
	})

	app := adw.NewApplication(fmt.Sprintf("io.github.alyraffauf.Switchyard.SettingsTest.run%d", launcherSettingsTestID.Add(1)), gio.ApplicationNonUnique)
	if err := app.Register(context.Background()); err != nil {
		t.Fatal(err)
	}
	browsers := []*Browser{
		{Browser: &appbrowser.Browser{ID: "alpha.desktop", Name: "Alpha"}},
		{Browser: &appbrowser.Browser{ID: "beta.desktop", Name: "Beta"}},
	}
	cfg := &Config{HiddenBrowsers: []string{"beta.desktop"}}
	if err := appconfig.Save(appconfig.Path(), cfg); err != nil {
		t.Fatal(err)
	}

	showLauncherWindow(app, "https://example.com/", browsers, cfg)
	launcher := app.Windows()[0]
	defer launcher.Destroy()
	if !launcher.ActivateAction("win.settings", nil) {
		t.Fatal("could not open Settings from launcher")
	}

	windows := app.Windows()
	if len(windows) != 2 {
		t.Fatalf("open windows = %d, want launcher and Settings", len(windows))
	}
	var settings *adw.Window
	var split *adw.NavigationSplitView
	for _, window := range windows {
		candidate := window.Cast().(*adw.Window)
		if view, ok := candidate.Content().(*adw.NavigationSplitView); ok {
			settings, split = candidate, view
		}
	}
	if settings == nil {
		t.Fatal("launcher did not open the Settings window")
	}
	defer settings.Destroy()
	sidebar := split.Sidebar().Child()
	list := launcherSettingsWidget(sidebar, func(list *widget.ListBox) bool { return true })
	if list == nil {
		t.Fatal("could not find sidebar list")
	}
	list.SelectRow(list.RowAtIndex(0))
	hidden := launcherSettingsWidget(split.Content().Child(), func(row *adw.ActionRow) bool {
		return row.Title() == "Hidden browsers"
	})
	if hidden == nil {
		t.Fatal("could not find Hidden browsers in Appearance")
	}
	hidden.Activate()
	dialog := settings.VisibleDialog()
	if dialog == nil {
		t.Fatal("could not open Hidden browsers")
	}
	check := launcherSettingsWidget(dialog.Child(), func(btn *widget.CheckButton) bool { return btn.Active() })
	if check == nil {
		t.Fatal("the hidden browser has no checked entry in Settings")
	}
	check.SetActive(false)
	reloaded, err := appconfig.Load(appconfig.Path())
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.HiddenBrowsers) != 0 {
		t.Fatalf("hidden browsers after unchecking Beta = %v, want none", reloaded.HiddenBrowsers)
	}
}

func launcherSettingsWidget[T comparable](root widget.Widgetter, matches func(T) bool) T {
	if match, ok := root.(T); ok && matches(match) {
		return match
	}
	for child := widget.BaseWidget(root).FirstChild(); child != nil; child = widget.BaseWidget(child).NextSibling() {
		found := launcherSettingsWidget(child, matches)
		var zero T
		if found != zero {
			return found
		}
	}
	var zero T
	return zero
}
