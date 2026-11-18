package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alyraffauf/switchyard/internal/config"
)

func TestOpeningDifferentBrowsersActionsSelectsFirstAction(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("XDG_DATA_DIRS", data)
	t.Setenv("HOME", data)
	// An accidental launch must fail before starting a process.
	t.Setenv("FLATPAK_ID", "")
	t.Setenv("PATH", data)
	apps := filepath.Join(data, "applications")
	if err := os.Mkdir(apps, 0o755); err != nil {
		t.Fatal(err)
	}

	const desktopFields = `Type=Application
MimeType=x-scheme-handler/https;
Exec=stub-browser %u
`
	const firstAction = `[Desktop Action window]
Name=New Window
Exec=stub-browser --new-window %u
`
	fixtures := map[string]string{
		"switchyard-test-a.desktop": "[Desktop Entry]\nName=000 Test Browser\n" + desktopFields +
			"Actions=window;private;\n" + firstAction + `[Desktop Action private]
Name=Private Window
Exec=stub-browser --private %u
`,
		"switchyard-test-b.desktop": "[Desktop Entry]\nName=001 Test Browser\n" + desktopFields +
			"Actions=window;\n" + firstAction,
	}
	for id, entry := range fixtures {
		if err := os.WriteFile(filepath.Join(apps, id), []byte(entry), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var current tea.Model = initialModel(&config.Config{FavoriteBrowser: "switchyard-test-a.desktop"}, "")
	for _, key := range []rune{tea.KeyRight, tea.KeyDown, tea.KeyEscape, tea.KeyDown, tea.KeyRight} {
		current, _ = current.Update(tea.KeyPressMsg{Code: key})
	}
	picker := current.(model)
	selected, ok := picker.browserList.SelectedItem().(browserItem)
	if !picker.pickingAction || !ok || selected.id != "switchyard-test-b.desktop" {
		t.Fatalf("expected the second browser's actions, got %v", selected)
	}
	action, ok := picker.actionList.SelectedItem().(actionItem)
	if !ok || action.name != "New Window" {
		t.Fatalf("selected action = %v, want New Window", picker.actionList.SelectedItem())
	}
}
