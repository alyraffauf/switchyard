// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PromptOnClick       bool   `toml:"prompt_on_click"`
	FallbackBrowser     string `toml:"fallback_browser"`
	CheckDefaultBrowser bool   `toml:"check_default_browser"`
	Rules               []Rule `toml:"rules"`
}

type Condition struct {
	Type    string `toml:"type"`    // "domain", "keyword", "glob", "regex"
	Pattern string `toml:"pattern"`
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
		PromptOnClick:      true,
		CheckDefaultBrowser: true,
		Rules:              []Rule{},
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
			if !matchesPattern(url, cond.Pattern, cond.Type) {
				return false
			}
		}
		return true
	} else {
		// OR: Any condition must match
		for _, cond := range r.Conditions {
			if matchesPattern(url, cond.Pattern, cond.Type) {
				return true
			}
		}
		return false
	}
}



func matchesPattern(url, pattern, patternType string) bool {
	domain := extractDomain(url)

	switch patternType {
	case "domain":
		// Exact domain match
		return strings.EqualFold(domain, pattern)
	case "keyword":
		// URL contains text
		return strings.Contains(strings.ToLower(url), strings.ToLower(pattern))
	case "regex":
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(url)
	case "glob":
		return matchGlob(url, pattern)
	default:
		return false
	}
}

func matchGlob(url, pattern string) bool {
	// Extract domain from URL for matching
	domain := extractDomain(url)

	// Simple glob matching: * matches any characters
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	// Match against domain or full URL
	return re.MatchString(domain) || re.MatchString(url)
}

func extractDomain(url string) string {
	// Remove protocol
	u := url
	if idx := strings.Index(u, "://"); idx != -1 {
		u = u[idx+3:]
	}
	// Remove path
	if idx := strings.Index(u, "/"); idx != -1 {
		u = u[:idx]
	}
	// Remove port
	if idx := strings.Index(u, ":"); idx != -1 {
		u = u[:idx]
	}
	return u
}

func sanitizeURL(url string) string {
	// Trim whitespace
	url = strings.TrimSpace(url)

	if url == "" {
		return ""
	}

	// Handle file:// URIs that GIO creates from bare domains
	// If it's a file:// URI but the path doesn't exist and looks like a domain, convert it
	if strings.HasPrefix(url, "file://") {
		filePath := strings.TrimPrefix(url, "file://")

		// Check if file actually exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist - might be a bare domain that GIO converted
			// Extract just the filename (last component)
			lastSlash := strings.LastIndex(filePath, "/")
			if lastSlash != -1 {
				possibleDomain := filePath[lastSlash+1:]
				// If it looks like a domain (contains dots, no spaces), treat it as one
				if strings.Contains(possibleDomain, ".") && !strings.Contains(possibleDomain, " ") {
					return "https://" + possibleDomain
				}
			}
		}
		// Otherwise, it's a real file path, return as-is
		return url
	}

	// If it already has another scheme, return as-is
	if strings.Contains(url, "://") {
		return url
	}

	// No scheme - check if it looks like a file path (starts with / or .)
	if strings.HasPrefix(url, "/") || strings.HasPrefix(url, ".") {
		// Looks like a file path, don't modify
		return url
	}

	// Add https:// prefix for bare domains/URLs
	return "https://" + url
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
	return defaultBrowser == "io.github.alyraffauf.Switchyard.desktop"
}

func setAsDefaultBrowser() error {
	// Set Switchyard as the default browser using xdg-settings
	cmd := hostCommand("xdg-settings", "set", "default-web-browser", "io.github.alyraffauf.Switchyard.desktop")
	return cmd.Run()
}
