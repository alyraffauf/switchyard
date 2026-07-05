// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"

	"github.com/alyraffauf/switchyard/internal/routing"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PromptOnClick            bool                  `toml:"prompt_on_click"`
	FavoriteBrowser          string                `toml:"favorite_browser"`
	HiddenBrowsers           []string              `toml:"hidden_browsers"`
	CheckDefaultBrowser      bool                  `toml:"check_default_browser"`
	ShowAppNames             bool                  `toml:"show_app_names"`
	ForceDarkMode            bool                  `toml:"force_dark_mode"`
	StayAlive                bool                  `toml:"stay_alive"`
	RemoveTrackingParameters bool                  `toml:"remove_tracking_parameters"`
	Redirections             []routing.Redirection `toml:"redirections,omitempty"`
	Rules                    []routing.Rule        `toml:"rules"`
}

func Dir() string {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "switchyard")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "switchyard")
}

func Path() string {
	return filepath.Join(Dir(), "config.toml")
}

func NewDefault() *Config {
	return &Config{
		PromptOnClick:            true,
		CheckDefaultBrowser:      true,
		ShowAppNames:             false,
		ForceDarkMode:            true,
		StayAlive:                true,
		RemoveTrackingParameters: false,
		Rules:                    []routing.Rule{},
	}
}

func Load(path string) (*Config, error) {
	settings := NewDefault()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}

	if err := toml.Unmarshal(data, settings); err != nil {
		return settings, err
	}

	applyTrackingParameterCompatibility(settings, data)
	return settings, nil
}

func Save(path string, settings *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Export(path string, settings *Config) error {
	data, err := toml.Marshal(settings)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Import(path string) (*Config, error) {
	settings := NewDefault()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	applyTrackingParameterCompatibility(settings, data)
	return settings, nil
}

func (settings *Config) MatchRule(url string) (browserID string, alwaysAsk bool, matched bool) {
	for _, rule := range settings.Rules {
		if rule.MatchesConditions(url) {
			return rule.Browser, rule.AlwaysAsk, true
		}
	}
	return "", false, false
}

func applyTrackingParameterCompatibility(settings *Config, data []byte) {
	var compatibility struct {
		RemoveTrackingParameters *bool `toml:"remove_tracking_parameters"`
		SanitizeLinks            *bool `toml:"sanitize_links"`
	}
	if err := toml.Unmarshal(data, &compatibility); err != nil {
		return
	}

	switch {
	case compatibility.RemoveTrackingParameters != nil:
		settings.RemoveTrackingParameters = *compatibility.RemoveTrackingParameters
	case compatibility.SanitizeLinks != nil:
		settings.RemoveTrackingParameters = *compatibility.SanitizeLinks
	}
}
