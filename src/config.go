// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PromptOnClick  bool   `toml:"prompt_on_click"`
	DefaultBrowser string `toml:"default_browser"`
	Rules          []Rule `toml:"rules"`
}

type Rule struct {
	Name        string `toml:"name"`
	Pattern     string `toml:"pattern"`
	PatternType string `toml:"pattern_type"` // glob, regex, keyword
	Browser     string `toml:"browser"`
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
		PromptOnClick: true,
		Rules:         []Rule{},
	}

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	toml.Unmarshal(data, cfg)
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

func (cfg *Config) matchRule(url string) *Browser {
	browsers := detectBrowsers()

	for _, rule := range cfg.Rules {
		if matchesPattern(url, rule.Pattern, rule.PatternType) {
			for _, b := range browsers {
				if b.ID == rule.Browser {
					return b
				}
			}
		}
	}
	return nil
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
