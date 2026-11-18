// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux

package gtk

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	appbrowser "github.com/alyraffauf/switchyard/internal/browser"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	widget "github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var launcherActionTestID atomic.Uint64

func TestLauncherBrowserActionUsesEditedURL(t *testing.T) {
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

	// Record the argument locally without starting a browser or opening a URL.
	resultPath := filepath.Join(t.TempDir(), "launched-url")
	if err := syscall.Mkfifo(resultPath, 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.OpenFile(resultPath, os.O_RDWR|syscall.O_NONBLOCK, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	t.Setenv("SWITCHYARD_TEST_LAUNCHED_URL", resultPath)

	app := adw.NewApplication(fmt.Sprintf("io.github.alyraffauf.Switchyard.ActionTest.run%d", launcherActionTestID.Add(1)), gio.ApplicationNonUnique)
	if err := app.Register(context.Background()); err != nil {
		t.Fatal(err)
	}
	browsers := []*Browser{{Browser: &appbrowser.Browser{
		ID: "alpha.desktop", Name: "Alpha",
		Actions: []appbrowser.Action{{
			ID: "new-private-window", Name: "New Private Window",
			Exec: `printf '%s\n' %u > "$SWITCHYARD_TEST_LAUNCHED_URL"`,
		}},
	}}}
	showLauncherWindow(app, "https://example.com/original", browsers, &Config{})
	window := app.Windows()[0]
	defer window.Destroy()

	entry := findLauncherURLEntry(window)
	if entry == nil {
		t.Fatal("could not find launcher URL entry")
	}
	edited := "https://example.com/edited"
	entry.SetText(edited)

	target := glib.NewVariantString("alpha.desktop:new-private-window")
	if !window.ActivateAction("win.launch-action", target) {
		t.Fatal("could not activate the browser action")
	}
	if err := reader.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	got, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if got != edited+"\n" {
		t.Fatalf("browser action received %q, want %q", got, edited+"\n")
	}
}

func findLauncherURLEntry(root widget.Widgetter) *widget.Entry {
	if entry, ok := root.(*widget.Entry); ok {
		return entry
	}
	for child := widget.BaseWidget(root).FirstChild(); child != nil; child = widget.BaseWidget(child).NextSibling() {
		if entry := findLauncherURLEntry(child); entry != nil {
			return entry
		}
	}
	return nil
}
