// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func assertDefaultConfig(t *testing.T, cfg *Config) {
	t.Helper()

	if !cfg.PromptOnClick {
		t.Fatal("PromptOnClick default should be true")
	}
	if !cfg.CheckDefaultBrowser {
		t.Fatal("CheckDefaultBrowser default should be true")
	}
	if !cfg.ForceDarkMode {
		t.Fatal("ForceDarkMode default should be true")
	}
	if !cfg.StayAlive {
		t.Fatal("StayAlive default should be true")
	}
	if cfg.RemoveTrackingParameters {
		t.Fatal("RemoveTrackingParameters default should be false")
	}
	if cfg.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func writeTestConfig(t *testing.T, configData string) string {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "switchyard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func TestLoadConfigUsesDefaultsWhenConfigIsMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := loadConfig()

	assertDefaultConfig(t, cfg)
}

func TestLoadConfigUsesDefaultsWhenConfigIsInvalid(t *testing.T) {
	writeTestConfig(t, "prompt_on_click = [invalid\n")

	cfg := loadConfig()

	assertDefaultConfig(t, cfg)
}

func TestLoadConfigKeepsDefaultsForOmittedFields(t *testing.T) {
	writeTestConfig(t, "prompt_on_click = false\n")

	cfg := loadConfig()

	if cfg.PromptOnClick {
		t.Fatal("loadConfig() should honor explicit prompt_on_click = false")
	}
	if !cfg.ForceDarkMode {
		t.Fatal("loadConfig() should keep ForceDarkMode default when omitted")
	}
	if cfg.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func TestImportConfigKeepsDefaultsForOmittedFields(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	importPath := filepath.Join(tempDir, "import.toml")
	if err := os.WriteFile(importPath, []byte("prompt_on_click = false\n"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg := &Config{}
	if err := importConfig(cfg, importPath); err != nil {
		t.Fatalf("importConfig() error = %v", err)
	}

	if cfg.PromptOnClick {
		t.Fatal("importConfig() should honor explicit prompt_on_click = false")
	}
	if !cfg.ForceDarkMode {
		t.Fatal("importConfig() should keep ForceDarkMode default when omitted")
	}
	if cfg.Rules == nil {
		t.Fatal("Rules default should be an empty slice, not nil")
	}
}

func TestLoadConfigUsesLegacySanitizeLinksKey(t *testing.T) {
	writeTestConfig(t, "sanitize_links = true\n")

	cfg := loadConfig()
	if !cfg.RemoveTrackingParameters {
		t.Fatal("loadConfig() did not honor legacy sanitize_links key")
	}
}

func TestLoadConfigPrefersNewTrackingParameterKey(t *testing.T) {
	writeTestConfig(t, "remove_tracking_parameters = false\nsanitize_links = true\n")

	cfg := loadConfig()
	if cfg.RemoveTrackingParameters {
		t.Fatal("loadConfig() should prefer remove_tracking_parameters over sanitize_links")
	}
}
