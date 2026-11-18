// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browserscan

import "testing"

func TestFindLocalizesDesktopIcon(t *testing.T) {
	clearLocaleEnv(t)
	home, _ := isolatedDirs(t)
	const id = "org.example.Browser.desktop"
	writeDesktop(t, home, id, `[Desktop Entry]
Type=Application
Name=Test Browser
Exec=fixture-browser %u
MimeType=x-scheme-handler/http;
Icon=fixture-browser
Icon[fr]=fixture-browser-fr
`)

	t.Setenv("LANG", "fr_CA.UTF-8")
	if found, ok := Find(id); !ok || found.Icon != "fixture-browser-fr" {
		t.Errorf("French Canadian icon = %q, found=%t, want fixture-browser-fr", found.Icon, ok)
	}
	t.Setenv("LANG", "de_DE.UTF-8")
	if found, ok := Find(id); !ok || found.Icon != "fixture-browser" {
		t.Errorf("German icon = %q, found=%t, want fixture-browser", found.Icon, ok)
	}
}
