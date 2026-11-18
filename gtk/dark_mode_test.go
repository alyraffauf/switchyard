// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"os"
	"runtime"
	"testing"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	widget "github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func TestSetupAppRestoresSystemColorScheme(t *testing.T) {
	if os.Getenv("SWITCHYARD_GTK_TEST") != "1" {
		t.Skip("set SWITCHYARD_GTK_TEST=1 in an isolated graphical session")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	if !widget.InitCheck() {
		t.Fatal("could not initialize GTK")
	}
	adw.Init()
	t.Setenv("FLATPAK_ID", "")

	manager := adw.StyleManagerGetDefault()
	t.Cleanup(func() { manager.SetColorScheme(adw.ColorSchemeDefault) })
	cfg := &Config{ForceDarkMode: true}
	setupApp(cfg)
	if got := manager.ColorScheme(); got != adw.ColorSchemeForceDark {
		t.Fatalf("initial scheme = %v, want forced dark", got)
	}

	cfg.ForceDarkMode = false
	setupApp(cfg)
	if got := manager.ColorScheme(); got != adw.ColorSchemeDefault {
		t.Fatalf("scheme after disabling the override = %v, want default", got)
	}
}
