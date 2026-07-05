// SPDX-License-Identifier: GPL-3.0-or-later

package routing

type Redirection struct {
	Name    string `toml:"name,omitempty"`
	Type    string `toml:"type,omitempty"`
	Find    string `toml:"find"`
	Replace string `toml:"replace"`
}

type Condition struct {
	Type    string `toml:"type"`
	Pattern string `toml:"pattern"`
	Negate  bool   `toml:"negate,omitempty"`
}

type Rule struct {
	Name       string      `toml:"name"`
	Conditions []Condition `toml:"conditions"`
	Logic      string      `toml:"logic,omitempty"`
	Browser    string      `toml:"browser"`
	AlwaysAsk  bool        `toml:"always_ask"`
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
