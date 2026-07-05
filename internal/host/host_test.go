// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"slices"
	"testing"
)

func TestInFlatpak(t *testing.T) {
	t.Setenv("FLATPAK_ID", "")
	if InFlatpak() {
		t.Fatal("want false when FLATPAK_ID is empty")
	}

	t.Setenv("FLATPAK_ID", "io.example.App")
	if !InFlatpak() {
		t.Fatal("want true when FLATPAK_ID is set")
	}
}

func TestHostCommandRunsDirectlyWhenNotSandboxed(t *testing.T) {
	t.Setenv("FLATPAK_ID", "")

	cmd := HostCommand("xdg-open", "https://example.com")

	want := []string{"xdg-open", "https://example.com"}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("args: got %v, want %v", cmd.Args, want)
	}
}

func TestHostCommandWrapsWithSpawnWhenSandboxed(t *testing.T) {
	t.Setenv("FLATPAK_ID", "io.example.App")

	cmd := HostCommand("xdg-open", "https://example.com")

	want := []string{"flatpak-spawn", "--host", "xdg-open", "https://example.com"}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("args: got %v, want %v", cmd.Args, want)
	}
}
