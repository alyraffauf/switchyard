// SPDX-License-Identifier: GPL-3.0-or-later

package routing

type Config struct {
	PromptOnClick            bool          `toml:"prompt_on_click"`
	FavoriteBrowser          string        `toml:"favorite_browser"`
	HiddenBrowsers           []string      `toml:"hidden_browsers"`
	CheckDefaultBrowser      bool          `toml:"check_default_browser"`
	ShowAppNames             bool          `toml:"show_app_names"`
	ForceDarkMode            bool          `toml:"force_dark_mode"`
	StayAlive                bool          `toml:"stay_alive"`
	RemoveTrackingParameters bool          `toml:"remove_tracking_parameters"`
	Redirections             []Redirection `toml:"redirections,omitempty"`
	Rules                    []Rule        `toml:"rules"`
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

func NewDefaultConfig() *Config {
	return &Config{
		PromptOnClick:            true,
		CheckDefaultBrowser:      true,
		ShowAppNames:             false,
		ForceDarkMode:            true,
		StayAlive:                true,
		RemoveTrackingParameters: false,
		Rules:                    []Rule{},
	}
}

func (cfg *Config) MatchRule(url string) (browserID string, alwaysAsk bool, matched bool) {
	for _, rule := range cfg.Rules {
		if rule.MatchesConditions(url) {
			return rule.Browser, rule.AlwaysAsk, true
		}
	}
	return "", false, false
}

func (rule *Rule) MatchesConditions(url string) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	logic := rule.Logic
	if logic == "" {
		logic = "all"
	}

	if logic == "all" {
		for _, condition := range rule.Conditions {
			if !matchesPattern(url, condition.Pattern, condition.Type, condition.Negate) {
				return false
			}
		}
		return true
	}

	for _, condition := range rule.Conditions {
		if matchesPattern(url, condition.Pattern, condition.Type, condition.Negate) {
			return true
		}
	}
	return false
}
