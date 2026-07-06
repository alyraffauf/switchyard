// SPDX-License-Identifier: GPL-3.0-or-later

//go:build darwin

package browserscan

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alyraffauf/switchyard/internal/browser"
)

// browserPlist is a minimal Info.plist for an app that handles HTTP(S).
const browserPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.example.testbrowser</string>
	<key>CFBundleName</key>
	<string>Test Browser</string>
	<key>CFBundleIconFile</key>
	<string>test-browser</string>
	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLSchemes</key>
			<array>
				<string>http</string>
				<string>https</string>
			</array>
		</dict>
	</array>
</dict>
</plist>
`

// writeAppBundle writes an Info.plist at dir/<name>.app/Contents/Info.plist.
func writeAppBundle(t *testing.T, dir, name, contents string) string {
	t.Helper()
	appPath := filepath.Join(dir, name+".app")
	contentsDir := filepath.Join(appPath, "Contents")
	if err := os.MkdirAll(contentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentsDir, "Info.plist"), []byte(contents), 0o644); err != nil {
		t.Fatalf("write Info.plist: %v", err)
	}
	return appPath
}

// isolatedPaths overrides app discovery to return exactly paths, in order,
// fully isolating the scan from the host's real installed apps.
func isolatedPaths(t *testing.T, paths ...string) {
	t.Helper()
	originalAppPaths := appPaths
	appPaths = func() []string { return paths }
	t.Cleanup(func() { appPaths = originalAppPaths })
}

func browserIDs(browsers []browser.Browser) []string {
	browserIdentifiers := make([]string, len(browsers))
	for i, installedBrowser := range browsers {
		browserIdentifiers[i] = installedBrowser.ID
	}
	return browserIdentifiers
}

func TestInstalledIncludesBrowserAndFields(t *testing.T) {
	dir := t.TempDir()
	appPath := writeAppBundle(t, dir, "Test Browser", browserPlist)
	isolatedPaths(t, appPath)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), browserIDs(got))
	}
	installed := got[0]
	if installed.ID != "com.example.testbrowser" || installed.Name != "Test Browser" {
		t.Errorf("unexpected fields: %+v", installed)
	}
	wantIcon := filepath.Join(dir, "Test Browser.app", "Contents", "Resources", "test-browser.icns")
	if installed.Icon != wantIcon {
		t.Errorf("icon: got %q, want %q", installed.Icon, wantIcon)
	}
	wantExec := filepath.Join(dir, "Test Browser.app")
	if installed.Exec != wantExec {
		t.Errorf("exec: got %q, want %q", installed.Exec, wantExec)
	}
	if len(installed.Actions) != 0 {
		t.Errorf("expected no actions, got %v", installed.Actions)
	}
}

func TestInstalledFiltering(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{
			name: "http only, no https",
			contents: plistWithURLTypes(`
				<dict>
					<key>CFBundleURLSchemes</key>
					<array><string>http</string></array>
				</dict>`),
		},
		{
			name:     "no url types at all",
			contents: plistWithURLTypes(""),
		},
		{
			name: "handler rank none",
			contents: plistWithURLTypes(`
				<dict>
					<key>CFBundleURLSchemes</key>
					<array><string>http</string><string>https</string></array>
					<key>LSHandlerRank</key>
					<string>None</string>
				</dict>`),
		},
		{
			name: "type role none",
			contents: plistWithURLTypes(`
				<dict>
					<key>CFBundleURLSchemes</key>
					<array><string>http</string><string>https</string></array>
					<key>CFBundleTypeRole</key>
					<string>None</string>
				</dict>`),
		},
		{
			name: "ui element agent",
			contents: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.example.testbrowser</string>
	<key>LSUIElement</key>
	<true/>
	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLSchemes</key>
			<array><string>http</string><string>https</string></array>
		</dict>
	</array>
</dict>
</plist>
`,
		},
		{
			name: "missing bundle identifier",
			contents: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLSchemes</key>
			<array><string>http</string><string>https</string></array>
		</dict>
	</array>
</dict>
</plist>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			appPath := writeAppBundle(t, dir, "Candidate", tt.contents)
			isolatedPaths(t, appPath)

			if got := Installed(); len(got) != 0 {
				t.Errorf("expected entry excluded, got %v", browserIDs(got))
			}
		})
	}
}

func TestInstalledExcludesSwitchyardFamily(t *testing.T) {
	dir := t.TempDir()
	switchyardPath := writeAppBundle(t, dir, "Switchyard", plistForID("io.github.alyraffauf.Switchyard"))
	switchyardDevelPath := writeAppBundle(t, dir, "Switchyard Devel", plistForID("io.github.alyraffauf.Switchyard.Devel"))
	realBrowserPath := writeAppBundle(t, dir, "Real Browser", plistForID("com.example.realbrowser"))
	isolatedPaths(t, switchyardPath, switchyardDevelPath, realBrowserPath)

	got := browserIDs(Installed())
	want := []string{"com.example.realbrowser"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestInstalledDedupesByBundleID(t *testing.T) {
	dir := t.TempDir()
	// Same bundle ID reported twice; the first one found wins.
	first := writeAppBundle(t, dir, "First Browser", plistForID("com.example.browser"))
	second := writeAppBundle(t, dir, "Second Browser", plistForID("com.example.browser"))
	isolatedPaths(t, first, second)

	got := Installed()
	if len(got) != 1 {
		t.Fatalf("got %d browsers, want 1: %v", len(got), browserIDs(got))
	}
	if got[0].Name != "First Browser" {
		t.Errorf("got %q, want %q", got[0].Name, "First Browser")
	}
}

func TestInstalledSortedByName(t *testing.T) {
	dir := t.TempDir()
	charliePath := writeAppBundle(t, dir, "Charlie", namedPlistForID("com.example.charlie", "Charlie"))
	alicePath := writeAppBundle(t, dir, "Alice", namedPlistForID("com.example.alice", "Alice"))
	bobPath := writeAppBundle(t, dir, "Bob", namedPlistForID("com.example.bob", "Bob"))
	isolatedPaths(t, charliePath, alicePath, bobPath)

	got := []string{}
	for _, installedBrowser := range Installed() {
		got = append(got, installedBrowser.Name)
	}
	want := []string{"Alice", "Bob", "Charlie"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFindBrowser(t *testing.T) {
	dir := t.TempDir()
	appPath := writeAppBundle(t, dir, "Test Browser", browserPlist)
	isolatedPaths(t, appPath)

	got, ok := Find("com.example.testbrowser")
	if !ok {
		t.Fatal("Find returned false")
	}
	if got.Name != "Test Browser" {
		t.Errorf("unexpected browser: %+v", got)
	}
}

func TestFindMissingOrRejected(t *testing.T) {
	dir := t.TempDir()
	appPath := writeAppBundle(t, dir, "Not A Browser", plistWithURLTypes(""))
	isolatedPaths(t, appPath)

	if got, ok := Find("com.example.testbrowser"); ok {
		t.Errorf("Find returned %+v, want false", got)
	}
	if got, ok := Find("missing.id"); ok {
		t.Errorf("Find returned %+v, want false", got)
	}
}

func TestListDesktopActionsIsAlwaysEmpty(t *testing.T) {
	if got := ListDesktopActions("com.example.testbrowser"); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func plistWithURLTypes(urlTypesXML string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.example.testbrowser</string>
	<key>CFBundleURLTypes</key>
	<array>` + urlTypesXML + `</array>
</dict>
</plist>
`
}

func plistForID(id string) string {
	return namedPlistForID(id, "")
}

func namedPlistForID(id, name string) string {
	nameXML := ""
	if name != "" {
		nameXML = "\t<key>CFBundleName</key>\n\t<string>" + name + "</string>\n"
	}
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>` + id + `</string>
` + nameXML + `	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLSchemes</key>
			<array><string>http</string><string>https</string></array>
		</dict>
	</array>
</dict>
</plist>
`
}
