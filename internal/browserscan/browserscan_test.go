// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browserscan

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alyraffauf/switchyard/internal/browser"
)

// browserEntry is a minimal desktop file that handles HTTP(S).
const browserEntry = `[Desktop Entry]
Type=Application
Name=Test Browser
Icon=test-browser
Exec=test-browser %u
MimeType=text/html;x-scheme-handler/http;x-scheme-handler/https;
`

// writeDesktop writes a .desktop file at dir/<rel> (rel may contain slashes for
// nested subdirs), creating parent dirs as needed.
func writeDesktop(t *testing.T, dir, rel, contents string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// isolatedDirs overrides the applications search path with two temp dirs (a
// higher-priority "home" dir and a lower-priority "system" dir), fully isolating
// the scan from the host's real desktop files. It returns those dirs.
func isolatedDirs(t *testing.T) (home, system string) {
	t.Helper()
	home = filepath.Join(t.TempDir(), "applications")
	system = filepath.Join(t.TempDir(), "applications")
	orig := applicationsDirs
	applicationsDirs = func() []string { return []string{home, system} }
	t.Cleanup(func() { applicationsDirs = orig })
	return home, system
}

func ids(browsers []browser.Browser) []string {
	out := make([]string, len(browsers))
	for i, b := range browsers {
		out[i] = b.ID
	}
	return out
}

func clearLocaleEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		t.Setenv(env, "")
	}
}

func TestInstalledIncludesBrowserAndFields(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "test-browser.desktop", browserEntry)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), ids(got))
	}
	b := got[0]
	if b.ID != "test-browser.desktop" || b.Name != "Test Browser" ||
		b.Icon != "test-browser" || b.Exec != "test-browser %u" {
		t.Errorf("unexpected fields: %+v", b)
	}
	if len(b.Actions) != 0 {
		t.Errorf("expected no actions, got %v", b.Actions)
	}
}

func TestInstalledLocalizedNameAndActions(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "fr_FR.UTF-8")

	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "loc.desktop", `[Desktop Entry]
Type=Application
Name=Web Browser
Name[fr]=Navigateur Web
Exec=browser %u
MimeType=x-scheme-handler/http;

[Desktop Action new-window]
Name=New Window
Name[fr]=Nouvelle fenêtre
Exec=browser --new-window %u
`)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1", len(got))
	}
	if got[0].Name != "Navigateur Web" {
		t.Errorf("browser name: got %q, want %q", got[0].Name, "Navigateur Web")
	}
	if len(got[0].Actions) != 1 || got[0].Actions[0].Name != "Nouvelle fenêtre" {
		t.Errorf("actions: got %+v", got[0].Actions)
	}
}

func TestInstalledFiltering(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{
			name: "no http mimetype",
			contents: `[Desktop Entry]
Type=Application
Name=Not A Browser
MimeType=text/html;
`,
		},
		{
			name: "no mimetype at all",
			contents: `[Desktop Entry]
Type=Application
Name=Plain App
`,
		},
		{
			name: "hidden",
			contents: `[Desktop Entry]
Type=Application
Name=Hidden Browser
Hidden=TRUE
MimeType=x-scheme-handler/http;
`,
		},
		{
			name: "wrong type",
			contents: `[Desktop Entry]
Type=Link
Name=Link Entry
URL=https://example.com
MimeType=x-scheme-handler/http;
`,
		},
		{
			name: "missing type",
			contents: `[Desktop Entry]
Name=No Type
MimeType=x-scheme-handler/http;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, _ := isolatedDirs(t)
			writeDesktop(t, home, "candidate.desktop", tt.contents)

			if got := Installed(); len(got) != 0 {
				t.Errorf("expected entry excluded, got %v", ids(got))
			}
		})
	}
}

func TestInstalledExcludesSwitchyardFamily(t *testing.T) {
	home, _ := isolatedDirs(t)
	// Both the stable and Devel variants are installed; neither may appear.
	writeDesktop(t, home, "io.github.alyraffauf.Switchyard.desktop", browserEntry)
	writeDesktop(t, home, "io.github.alyraffauf.Switchyard.Devel.desktop", browserEntry)
	writeDesktop(t, home, "real-browser.desktop", browserEntry)

	got := ids(Installed())
	want := []string{"real-browser.desktop"}
	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestInstalledShadowingHomeWins(t *testing.T) {
	home, system := isolatedDirs(t)
	// Same ID in both dirs with different names; the data-home copy must win.
	writeDesktop(t, home, "browser.desktop", `[Desktop Entry]
Type=Application
Name=Home Browser
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, system, "browser.desktop", `[Desktop Entry]
Type=Application
Name=System Browser
MimeType=x-scheme-handler/http;
`)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), ids(got))
	}
	if got[0].Name != "Home Browser" {
		t.Errorf("got %q, want %q", got[0].Name, "Home Browser")
	}
}

// TestInstalledIncludesNoDisplay guards against regressing the fix for #10:
// users deliberately set NoDisplay=true on custom per-profile browser entries
// to hide them from launchers while still wanting Switchyard to detect them.
func TestInstalledIncludesNoDisplay(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "browser.desktop", `[Desktop Entry]
Type=Application
Name=NoDisplay Browser
NoDisplay=true
MimeType=x-scheme-handler/http;
`)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), ids(got))
	}
	if got[0].Name != "NoDisplay Browser" {
		t.Errorf("got %q, want %q", got[0].Name, "NoDisplay Browser")
	}
}

// TestInstalledIncludeNoDisplayFalse covers the opt-in launcher-style behavior
// for other consumers of the parser: NoDisplay=true entries are excluded while
// ordinary browsers still come through.
func TestInstalledIncludeNoDisplayFalse(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "hidden.desktop", `[Desktop Entry]
Type=Application
Name=NoDisplay Browser
NoDisplay=true
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, home, "visible.desktop", `[Desktop Entry]
Type=Application
Name=Visible Browser
MimeType=x-scheme-handler/http;
`)

	got := ids(Installed(IncludeNoDisplay(false)))
	want := []string{"visible.desktop"}
	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestFindIncludeNoDisplayFalse rejects a NoDisplay entry only when the caller
// opts out; the default still matches it.
func TestFindIncludeNoDisplayFalse(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "hidden.desktop", `[Desktop Entry]
Type=Application
Name=NoDisplay Browser
NoDisplay=true
MimeType=x-scheme-handler/http;
`)

	if _, ok := Find("hidden.desktop"); !ok {
		t.Error("default Find should match a NoDisplay browser")
	}
	if got, ok := Find("hidden.desktop", IncludeNoDisplay(false)); ok {
		t.Errorf("Find with IncludeNoDisplay(false) returned %+v, want false", got)
	}
}

func TestInstalledShadowingHiddenHomeSuppressesSystem(t *testing.T) {
	home, system := isolatedDirs(t)
	// A hidden copy in data-home shadows a valid copy in the system dir.
	writeDesktop(t, home, "browser.desktop", `[Desktop Entry]
Type=Application
Name=Home Browser
Hidden=true
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, system, "browser.desktop", browserEntry)

	if got := Installed(); len(got) != 0 {
		t.Errorf("expected shadowed system copy suppressed, got %v", ids(got))
	}
}

func TestInstalledNestedSubdirID(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "sub/foo.desktop", browserEntry)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), ids(got))
	}
	if got[0].ID != "sub-foo.desktop" {
		t.Errorf("got ID %q, want %q", got[0].ID, "sub-foo.desktop")
	}
}

func TestInstalledSortedByName(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "charlie.desktop", `[Desktop Entry]
Type=Application
Name=Charlie
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, home, "alice.desktop", `[Desktop Entry]
Type=Application
Name=Alice
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, home, "bob.desktop", `[Desktop Entry]
Type=Application
Name=Bob
MimeType=x-scheme-handler/https;
`)

	got := ids(Installed())
	want := []string{"alice.desktop", "bob.desktop", "charlie.desktop"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestFindBrowser(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "test-browser.desktop", browserEntry)

	got, ok := Find("test-browser.desktop")
	if !ok {
		t.Fatal("Find returned false")
	}
	if got.ID != "test-browser.desktop" || got.Name != "Test Browser" ||
		got.Icon != "test-browser" || got.Exec != "test-browser %u" {
		t.Errorf("unexpected browser: %+v", got)
	}
}

func TestFindMissingOrRejected(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		contents string
	}{
		{
			name: "missing",
			id:   "missing.desktop",
		},
		{
			name: "hidden",
			id:   "hidden.desktop",
			contents: `[Desktop Entry]
Type=Application
Name=Hidden
Hidden=true
MimeType=x-scheme-handler/http;
`,
		},
		{
			name: "not browser",
			id:   "app.desktop",
			contents: `[Desktop Entry]
Type=Application
Name=App
MimeType=text/plain;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, _ := isolatedDirs(t)
			if tt.contents != "" {
				writeDesktop(t, home, tt.id, tt.contents)
			}

			if got, ok := Find(tt.id); ok {
				t.Errorf("Find returned %+v, want false", got)
			}
		})
	}
}

func TestFindShadowingRejectedHomeSuppressesSystem(t *testing.T) {
	home, system := isolatedDirs(t)
	writeDesktop(t, home, "browser.desktop", `[Desktop Entry]
Type=Application
Name=Hidden Home Browser
Hidden=true
MimeType=x-scheme-handler/http;
`)
	writeDesktop(t, system, "browser.desktop", browserEntry)

	if got, ok := Find("browser.desktop"); ok {
		t.Errorf("Find returned %+v, want false", got)
	}
}

func TestFindNestedSubdirID(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "sub/foo.desktop", browserEntry)

	got, ok := Find("sub-foo.desktop")
	if !ok {
		t.Fatal("Find returned false")
	}
	if got.ID != "sub-foo.desktop" {
		t.Errorf("got ID %q, want %q", got.ID, "sub-foo.desktop")
	}
}

func TestFindExcludesSwitchyardFamily(t *testing.T) {
	home, _ := isolatedDirs(t)
	writeDesktop(t, home, "io.github.alyraffauf.Switchyard.desktop", browserEntry)

	if got, ok := Find("io.github.alyraffauf.Switchyard.desktop"); ok {
		t.Errorf("Find returned %+v, want false", got)
	}
}

func BenchmarkInstalledManyDesktopFiles(b *testing.B) {
	home := filepath.Join(b.TempDir(), "applications")
	system := filepath.Join(b.TempDir(), "applications")
	orig := applicationsDirs
	applicationsDirs = func() []string { return []string{home, system} }
	b.Cleanup(func() { applicationsDirs = orig })

	for i := range 500 {
		writeBenchmarkDesktop(b, home, fmt.Sprintf("browser-%03d.desktop", i), fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Browser %03d
Icon=browser
Exec=browser-%03d %%u
MimeType=x-scheme-handler/http;x-scheme-handler/https;
`, i, i))
		writeBenchmarkDesktop(b, home, fmt.Sprintf("app-%03d.desktop", i), fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=App %03d
Exec=app-%03d
MimeType=text/plain;
`, i, i))
	}

	b.ResetTimer()
	for b.Loop() {
		got := Installed()
		if len(got) != 500 {
			b.Fatalf("got %d browsers, want 500", len(got))
		}
	}
}

func writeBenchmarkDesktop(b *testing.B, dir, rel, contents string) {
	b.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		b.Fatalf("write %s: %v", rel, err)
	}
}
