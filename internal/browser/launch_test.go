// SPDX-License-Identifier: GPL-3.0-or-later

package browser

import (
	"slices"
	"testing"

	"github.com/alyraffauf/goxdgdesktop/desktopexec"
)

func TestBuildCommandEmptyCmdlineIsNoOp(t *testing.T) {
	if cmd := buildCommand("", "", "", false); cmd != nil {
		t.Fatalf("empty command line: got %v, want nil", cmd.Args)
	}
}

func TestBuildCommandHostSubstitutesURL(t *testing.T) {
	const url = "https://example.com"
	want := desktopexec.SubstituteOrAppendURL("firefox %u", url)

	cmd := buildCommand("firefox %u", url, "", false)

	wantArgs := []string{"sh", "-c", want}
	if !slices.Equal(cmd.Args, wantArgs) {
		t.Fatalf("args: got %v, want %v", cmd.Args, wantArgs)
	}
	if cmd.Env != nil {
		t.Fatalf("env: got %v, want nil when no activation token", cmd.Env)
	}
}

func TestBuildCommandHostPassesActivationTokenViaEnv(t *testing.T) {
	cmd := buildCommand("firefox %u", "https://example.com", "tok-123", false)

	lastEnv := cmd.Env[len(cmd.Env)-1]
	if want := activationTokenEnv + "=tok-123"; lastEnv != want {
		t.Fatalf("activation token env: got %q, want %q", lastEnv, want)
	}
}

func TestBuildCommandFlatpakWrapsWithHostSpawn(t *testing.T) {
	const url = "https://example.com"
	want := desktopexec.SubstituteOrAppendURL("firefox %u", url)

	cmd := buildCommand("firefox %u", url, "tok-123", true)

	wantArgs := []string{
		"flatpak-spawn", "--host",
		"--env=" + activationTokenEnv + "=tok-123",
		"sh", "-c", want,
	}
	if !slices.Equal(cmd.Args, wantArgs) {
		t.Fatalf("args: got %v, want %v", cmd.Args, wantArgs)
	}
}

// A command line that already escapes the sandbox must not be wrapped twice.
func TestBuildCommandFlatpakDoesNotDoubleWrap(t *testing.T) {
	cmd := buildCommand("flatpak-spawn --host firefox %u", "https://example.com", "tok", true)

	if cmd.Args[0] != "sh" {
		t.Fatalf("expected already-spawned command to run under sh, got %v", cmd.Args)
	}
}
