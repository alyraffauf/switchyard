// SPDX-License-Identifier: GPL-3.0-or-later

//go:build darwin

package browser

import (
	"slices"
	"testing"
)

func TestBuildCommandEmptyBundlePathIsNoOp(t *testing.T) {
	if cmd := buildCommand("", "https://example.com", "", false); cmd != nil {
		t.Fatalf("empty bundle path: got %v, want nil", cmd.Args)
	}
}

func TestBuildCommandOpensBundleWithURL(t *testing.T) {
	cmd := buildCommand("/Applications/Firefox.app", "https://example.com", "", false)

	want := []string{"open", "-a", "/Applications/Firefox.app", "https://example.com"}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("args: got %v, want %v", cmd.Args, want)
	}
}

func TestBuildCommandOmitsURLWhenEmpty(t *testing.T) {
	cmd := buildCommand("/Applications/Firefox.app", "", "", false)

	want := []string{"open", "-a", "/Applications/Firefox.app"}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("args: got %v, want %v", cmd.Args, want)
	}
}
