// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browserscan

import "testing"

func TestDesktopLabelsDecodeStringEscapes(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "fr_CA.UTF-8")
	home, _ := isolatedDirs(t)
	const id = "org.example.Browser.desktop"
	writeDesktop(t, home, id, `[Desktop Entry]
Type=Application
Name=Test\sBrowser
Name[fr]=Café\sWeb
Exec=fixture-browser %u
MimeType=x-scheme-handler/http;
Actions=Profile;

[Desktop Action Profile]
Name=Work\sProfile\\sync
Exec=fixture-browser --profile=work %u
`)
	found, ok := Find(id)
	if !ok || found.Name != "Café Web" {
		t.Errorf("localized browser name = %q, found=%t, want Café Web", found.Name, ok)
	}
	actions := ListDesktopActions(id)
	if len(actions) != 1 || actions[0].Name != `Work Profile\sync` {
		t.Errorf("action labels = %v, want Work Profile\\sync", actions)
	}
}
