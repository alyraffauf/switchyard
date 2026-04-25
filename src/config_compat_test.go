// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigUsesLegacySanitizeLinksKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "switchyard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	configData := []byte("sanitize_links = true\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), configData, 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg := loadConfig()
	if !cfg.RemoveTrackingParameters {
		t.Fatal("loadConfig() did not honor legacy sanitize_links key")
	}
}

func TestLoadConfigPrefersNewTrackingParameterKey(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "switchyard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	configData := []byte("remove_tracking_parameters = false\nsanitize_links = true\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), configData, 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg := loadConfig()
	if cfg.RemoveTrackingParameters {
		t.Fatal("loadConfig() should prefer remove_tracking_parameters over sanitize_links")
	}
}
