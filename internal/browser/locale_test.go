// SPDX-License-Identifier: GPL-3.0-or-later

package browser

import (
	"testing"

	"github.com/alyraffauf/goxdgdesktop/desktopfile"
)

const localeEntry = `[Desktop Entry]
Type=Application
Name=Web Browser
Name[fr]=Navigateur Web
Name[fr_CA]=Fureteur Web
Exec=browser %u

[Desktop Action new-window]
Name=New Window
Name[fr]=Nouvelle fenêtre
Exec=browser --new-window %u
`

func clearLocaleEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		t.Setenv(env, "")
	}
}

func TestLocalizedStringFallback(t *testing.T) {
	file := desktopfile.Parse([]byte(localeEntry))

	tests := []struct {
		name string
		lang string
		want string
	}{
		{"unset falls back to plain", "", "Web Browser"},
		{"exact language", "fr_FR.UTF-8", "Navigateur Web"},
		{"country-specific wins", "fr_CA.UTF-8", "Fureteur Web"},
		{"unknown locale falls back", "de_DE.UTF-8", "Web Browser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearLocaleEnv(t)
			if tt.lang != "" {
				t.Setenv("LANG", tt.lang)
			}
			if got := LocalizedString(file, desktopfile.EntrySection, "Name"); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocalizedStringLanguagePriority(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "de_DE.UTF-8")
	t.Setenv("LANGUAGE", "de:fr")

	file := desktopfile.Parse([]byte(localeEntry))
	if got := LocalizedString(file, desktopfile.EntrySection, "Name"); got != "Navigateur Web" {
		t.Errorf("got %q, want %q", got, "Navigateur Web")
	}
}

func TestLocalizedActions(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "fr_FR.UTF-8")

	file := desktopfile.Parse([]byte(localeEntry))
	actions := LocalizedActions(file)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
	if actions[0].ID != "new-window" || actions[0].Name != "Nouvelle fenêtre" {
		t.Errorf("unexpected action: %+v", actions[0])
	}
}
