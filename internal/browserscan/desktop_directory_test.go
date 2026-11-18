// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browserscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAndListActionsSkipDirectoriesNamedLikeDesktopFiles(t *testing.T) {
	for _, nested := range []bool{false, true} {
		name := "lower priority desktop file"
		if nested {
			name = "same root nested desktop file"
		}
		t.Run(name, func(t *testing.T) {
			home, system := isolatedDirs(t)
			const id = "vendor-org.example.Browser.desktop"
			if err := os.MkdirAll(filepath.Join(home, id), 0o755); err != nil {
				t.Fatal(err)
			}
			dir, rel := system, id
			if nested {
				dir, rel = home, "vendor/org.example.Browser.desktop"
			}
			writeDesktop(t, dir, rel, `[Desktop Entry]
Type=Application
Name=Test Browser
Exec=fixture-browser %u
MimeType=x-scheme-handler/http;
Actions=NewWindow;

[Desktop Action NewWindow]
Name=New Window
Exec=fixture-browser --new-window %u
`)
			if found, ok := Find(id); !ok || found.Name != "Test Browser" {
				t.Errorf("Find(%q) = %+v, %v, want Test Browser", id, found, ok)
			}
			if actions := ListDesktopActions(id); len(actions) != 1 || actions[0].ID != "NewWindow" {
				t.Errorf("ListDesktopActions(%q) = %v, want NewWindow", id, actions)
			}
		})
	}
}
