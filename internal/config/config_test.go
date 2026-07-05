// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func assertDefaultConfig(t *testing.T, config *Config) {
	t.Helper()

	if !config.PromptOnClick {
		t.Fatal("PromptOnClick default should be true")
	}
	if !config.CheckDefaultBrowser {
		t.Fatal("CheckDefaultBrowser default should be true")
	}
	if !config.ForceDarkMode {
		t.Fatal("ForceDarkMode default should be true")
	}
	if !config.StayAlive {
		t.Fatal("StayAlive default should be true")
	}
	if config.RemoveTrackingParameters {
		t.Fatal("RemoveTrackingParameters default should be false")
	}
	if config.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func writeTestConfig(t *testing.T, configData string) string {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Dir(Path())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	configPath := Path()
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func TestLoadUsesDefaultsWhenConfigIsMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	config, err := Load(Path())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assertDefaultConfig(t, config)
}

func TestLoadUsesDefaultsWhenConfigIsInvalid(t *testing.T) {
	configPath := writeTestConfig(t, "prompt_on_click = [invalid\n")

	config, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() expected invalid TOML error")
	}

	assertDefaultConfig(t, config)
}

func TestLoadReturnsReadError(t *testing.T) {
	configPath := t.TempDir()

	config, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() expected read error")
	}

	assertDefaultConfig(t, config)
}

func TestLoadKeepsDefaultsForOmittedFields(t *testing.T) {
	configPath := writeTestConfig(t, "prompt_on_click = false\n")

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.PromptOnClick {
		t.Fatal("Load() should honor explicit prompt_on_click = false")
	}
	if !config.ForceDarkMode {
		t.Fatal("Load() should keep ForceDarkMode default when omitted")
	}
	if config.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func TestImportKeepsDefaultsForOmittedFields(t *testing.T) {
	tempDir := t.TempDir()
	importPath := filepath.Join(tempDir, "import.toml")
	if err := os.WriteFile(importPath, []byte("prompt_on_click = false\n"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	config, err := Import(importPath)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if config.PromptOnClick {
		t.Fatal("Import() should honor explicit prompt_on_click = false")
	}
	if !config.ForceDarkMode {
		t.Fatal("Import() should keep ForceDarkMode default when omitted")
	}
	if config.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func TestLoadUsesLegacySanitizeLinksKey(t *testing.T) {
	configPath := writeTestConfig(t, "sanitize_links = true\n")

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !config.RemoveTrackingParameters {
		t.Fatal("Load() did not honor legacy sanitize_links key")
	}
}

func TestLoadPrefersNewTrackingParameterKey(t *testing.T) {
	configPath := writeTestConfig(t, "remove_tracking_parameters = false\nsanitize_links = true\n")

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.RemoveTrackingParameters {
		t.Fatal("Load() should prefer remove_tracking_parameters over sanitize_links")
	}
}

func TestImportUsesLegacySanitizeLinksKey(t *testing.T) {
	importPath := filepath.Join(t.TempDir(), "import.toml")
	if err := os.WriteFile(importPath, []byte("sanitize_links = true\n"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	config, err := Import(importPath)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !config.RemoveTrackingParameters {
		t.Fatal("Import() did not honor legacy sanitize_links key")
	}
}
