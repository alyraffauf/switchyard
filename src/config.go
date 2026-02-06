// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PromptOnClick       bool          `toml:"prompt_on_click"`
	FavoriteBrowser     string        `toml:"favorite_browser"`
	HiddenBrowsers      []string      `toml:"hidden_browsers"`
	CheckDefaultBrowser bool          `toml:"check_default_browser"`
	ShowAppNames        bool          `toml:"show_app_names"`
	ForceDarkMode       bool          `toml:"force_dark_mode"`
	StayAlive           bool          `toml:"stay_alive"`
	Redirections        []Redirection `toml:"redirections,omitempty"`
	Rules               []Rule        `toml:"rules"`
}

type Redirection struct {
	Name    string `toml:"name,omitempty"`
	Type    string `toml:"type,omitempty"` // "domain", "wildcard", or "regex", defaults to "domain"
	Find    string `toml:"find"`
	Replace string `toml:"replace"`
}

type Condition struct {
	Type    string `toml:"type"` // "domain", "keyword", "glob", "regex"
	Pattern string `toml:"pattern"`
	Negate  bool   `toml:"negate,omitempty"`
}

type Rule struct {
	Name       string      `toml:"name"`
	Conditions []Condition `toml:"conditions"`
	Logic      string      `toml:"logic,omitempty"` // "all" or "any"
	Browser    string      `toml:"browser"`
	AlwaysAsk  bool        `toml:"always_ask"`
}

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

func loadConfig() *Config {
	cfg := &Config{
		PromptOnClick:       true,
		CheckDefaultBrowser: true,
		ShowAppNames:        false, // Default: hide app names, show tooltips
		ForceDarkMode:       true,  // Default: force dark mode
		Rules:               []Rule{},
	}

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration\n")
		return &Config{
			PromptOnClick:       true,
			CheckDefaultBrowser: true,
			Rules:               []Rule{},
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

func (cfg *Config) matchRule(url string) (browserID string, alwaysAsk bool, matched bool) {
	for _, rule := range cfg.Rules {
		if rule.matchesConditions(url) {
			return rule.Browser, rule.AlwaysAsk, true
		}
	}
	return "", false, false
}

func (r *Rule) matchesConditions(url string) bool {
	if len(r.Conditions) == 0 {
		return false
	}

	logic := r.Logic
	if logic == "" {
		logic = "all" // Default to AND logic
	}

	if logic == "all" {
		// AND: All conditions must match
		for _, cond := range r.Conditions {
			if !matchesPattern(url, cond.Pattern, cond.Type, cond.Negate) {
				return false
			}
		}
		return true
	} else {
		// OR: Any condition must match
		for _, cond := range r.Conditions {
			if matchesPattern(url, cond.Pattern, cond.Type, cond.Negate) {
				return true
			}
		}
		return false
	}
}

// hostCommand creates a command that runs on the host system when in flatpak,
// or directly otherwise
func hostCommand(name string, args ...string) *exec.Cmd {
	if os.Getenv("FLATPAK_ID") != "" {
		hostArgs := append([]string{"--host", name}, args...)
		return exec.Command("flatpak-spawn", hostArgs...)
	}
	return exec.Command(name, args...)
}

func isDefaultBrowser() bool {
	// Check if Switchyard is the default browser using xdg-settings
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
	// Set Switchyard as the default browser using xdg-settings
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

	newCfg := &Config{}
	if err := toml.Unmarshal(data, newCfg); err != nil {
		return err
	}

	*cfg = *newCfg
	return saveConfig(cfg)
}
