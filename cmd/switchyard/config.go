// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alyraffauf/switchyard/internal/routing"
	"github.com/pelletier/go-toml/v2"
)

type Config = routing.Config
type Redirection = routing.Redirection
type Condition = routing.Condition
type Rule = routing.Rule

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "switchyard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "switchyard")
}

func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func newDefaultConfig() *Config {
	return routing.NewDefaultConfig()
}

func loadConfig() *Config {
	cfg := newDefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration\n")
		return cfg
	}

	var compat struct {
		RemoveTrackingParameters *bool `toml:"remove_tracking_parameters"`
		SanitizeLinks            *bool `toml:"sanitize_links"`
	}
	if err := toml.Unmarshal(data, &compat); err == nil {
		switch {
		case compat.RemoveTrackingParameters != nil:
			cfg.RemoveTrackingParameters = *compat.RemoveTrackingParameters
		case compat.SanitizeLinks != nil:
			cfg.RemoveTrackingParameters = *compat.SanitizeLinks
		}
	}

	return cfg
}

func saveConfig(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0644)
}

// hostCommand runs commands on the host when Switchyard is sandboxed by Flatpak.
func hostCommand(name string, args ...string) *exec.Cmd {
	if os.Getenv("FLATPAK_ID") != "" {
		hostArgs := append([]string{"--host", name}, args...)
		return exec.Command("flatpak-spawn", hostArgs...)
	}
	return exec.Command(name, args...)
}

func isDefaultBrowser() bool {
	cmd := hostCommand("xdg-settings", "get", "default-web-browser")

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	defaultBrowser := strings.TrimSpace(string(output))
	desktopFile := getAppID() + ".desktop"
	return defaultBrowser == desktopFile
}

func setAsDefaultBrowser() error {
	desktopFile := getAppID() + ".desktop"
	cmd := hostCommand("xdg-settings", "set", "default-web-browser", desktopFile)
	return cmd.Run()
}

// save the current cfg to the specified path
func exportConfig(cfg *Config, path string) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// load cfg from the specified path and replace current config
func importConfig(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newCfg := newDefaultConfig()
	if err := toml.Unmarshal(data, newCfg); err != nil {
		return err
	}

	*cfg = *newCfg
	return saveConfig(cfg)
}
