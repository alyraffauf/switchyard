// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"testing"

	"github.com/alyraffauf/switchyard/internal/routing"
)

func TestConfigMatchRule(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		url           string
		wantBrowserID string
		wantAlwaysAsk bool
		wantMatched   bool
	}{
		{
			name: "matches first rule",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
					{
						Name:      "GitLab",
						Browser:   "chrome.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "gitlab.com"},
						},
					},
				},
			},
			url:           "https://github.com/user/repo",
			wantBrowserID: "firefox.desktop",
			wantAlwaysAsk: false,
			wantMatched:   true,
		},
		{
			name: "matches second rule",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
					{
						Name:      "GitLab",
						Browser:   "chrome.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "gitlab.com"},
						},
					},
				},
			},
			url:           "https://gitlab.com/user/repo",
			wantBrowserID: "chrome.desktop",
			wantAlwaysAsk: false,
			wantMatched:   true,
		},
		{
			name: "no rules match",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
				},
			},
			url:           "https://example.com",
			wantBrowserID: "",
			wantAlwaysAsk: false,
			wantMatched:   false,
		},
		{
			name: "rule with always_ask set",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "Work Sites",
						Browser:   "",
						AlwaysAsk: true,
						Conditions: []routing.Condition{
							{Type: "keyword", Pattern: "work"},
						},
					},
				},
			},
			url:           "https://work.example.com",
			wantBrowserID: "",
			wantAlwaysAsk: true,
			wantMatched:   true,
		},
		{
			name: "empty rules list",
			config: Config{
				Rules: []routing.Rule{},
			},
			url:           "https://example.com",
			wantBrowserID: "",
			wantAlwaysAsk: false,
			wantMatched:   false,
		},
		{
			name: "first matching rule wins",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "First Match",
						Browser:   "first.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "keyword", Pattern: "github"},
						},
					},
					{
						Name:      "Second Match",
						Browser:   "second.desktop",
						AlwaysAsk: false,
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
				},
			},
			url:           "https://github.com",
			wantBrowserID: "first.desktop",
			wantAlwaysAsk: false,
			wantMatched:   true,
		},
		{
			name: "complex rule with AND logic matches",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "GitHub API Docs",
						Browser:   "work-browser.desktop",
						AlwaysAsk: false,
						Logic:     "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "docs.github.com"},
							{Type: "keyword", Pattern: "api"},
						},
					},
				},
			},
			url:           "https://docs.github.com/api/reference",
			wantBrowserID: "work-browser.desktop",
			wantAlwaysAsk: false,
			wantMatched:   true,
		},
		{
			name: "complex rule with OR logic matches",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:      "Git Platforms",
						Browser:   "dev-browser.desktop",
						AlwaysAsk: false,
						Logic:     "any",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "github.com"},
							{Type: "domain", Pattern: "gitlab.com"},
							{Type: "domain", Pattern: "bitbucket.org"},
						},
					},
				},
			},
			url:           "https://bitbucket.org/user/repo",
			wantBrowserID: "dev-browser.desktop",
			wantAlwaysAsk: false,
			wantMatched:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			browserID, alwaysAsk, matched := test.config.MatchRule(test.url)
			if browserID != test.wantBrowserID {
				t.Errorf("Config.MatchRule(%q) browserID = %q, want %q", test.url, browserID, test.wantBrowserID)
			}
			if alwaysAsk != test.wantAlwaysAsk {
				t.Errorf("Config.MatchRule(%q) alwaysAsk = %v, want %v", test.url, alwaysAsk, test.wantAlwaysAsk)
			}
			if matched != test.wantMatched {
				t.Errorf("Config.MatchRule(%q) matched = %v, want %v", test.url, matched, test.wantMatched)
			}
		})
	}
}

func TestConfigMatchRuleOrdering(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		url           string
		wantBrowserID string
		wantMatched   bool
	}{
		{
			name: "first matching rule wins",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "First Rule",
						Browser: "first-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
					{
						Name:    "Second Rule",
						Browser: "second-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
				},
			},
			url:           "https://example.com/path",
			wantBrowserID: "first-browser.desktop",
			wantMatched:   true,
		},
		{
			name: "more specific rule first",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "Specific Rule",
						Browser: "specific-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "docs.example.com"},
						},
					},
					{
						Name:    "General Rule",
						Browser: "general-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "glob", Pattern: "*.example.com"},
						},
					},
				},
			},
			url:           "https://docs.example.com/guide",
			wantBrowserID: "specific-browser.desktop",
			wantMatched:   true,
		},
		{
			name: "general rule matches when specific doesn't",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "Specific Rule",
						Browser: "specific-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "docs.example.com"},
						},
					},
					{
						Name:    "General Rule",
						Browser: "general-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "glob", Pattern: "*.example.com"},
						},
					},
				},
			},
			url:           "https://api.example.com/endpoint",
			wantBrowserID: "general-browser.desktop",
			wantMatched:   true,
		},
		{
			name: "non-matching rules are skipped",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "Non-matching Rule",
						Browser: "wrong-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "other.com"},
						},
					},
					{
						Name:    "Matching Rule",
						Browser: "correct-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
				},
			},
			url:           "https://example.com/path",
			wantBrowserID: "correct-browser.desktop",
			wantMatched:   true,
		},
		{
			name: "empty rules list",
			config: Config{
				Rules: []routing.Rule{},
			},
			url:           "https://example.com/path",
			wantBrowserID: "",
			wantMatched:   false,
		},
		{
			name: "all rules fail to match",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "Rule 1",
						Browser: "browser1.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "other1.com"},
						},
					},
					{
						Name:    "Rule 2",
						Browser: "browser2.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "other2.com"},
						},
					},
				},
			},
			url:           "https://example.com/path",
			wantBrowserID: "",
			wantMatched:   false,
		},
		{
			name: "rule with empty conditions is skipped",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:       "Empty Rule",
						Browser:    "empty-browser.desktop",
						Logic:      "all",
						Conditions: []routing.Condition{},
					},
					{
						Name:    "Valid Rule",
						Browser: "valid-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
				},
			},
			url:           "https://example.com/path",
			wantBrowserID: "valid-browser.desktop",
			wantMatched:   true,
		},
		{
			name: "AND rule fails partial match, next rule matches",
			config: Config{
				Rules: []routing.Rule{
					{
						Name:    "AND Rule",
						Browser: "and-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
							{Type: "keyword", Pattern: "admin"},
						},
					},
					{
						Name:    "Fallback Rule",
						Browser: "fallback-browser.desktop",
						Logic:   "all",
						Conditions: []routing.Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
				},
			},
			url:           "https://example.com/user",
			wantBrowserID: "fallback-browser.desktop",
			wantMatched:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			browserID, _, matched := test.config.MatchRule(test.url)
			if browserID != test.wantBrowserID {
				t.Errorf("Config.MatchRule(%q) browserID = %q, want %q",
					test.url, browserID, test.wantBrowserID)
			}
			if matched != test.wantMatched {
				t.Errorf("Config.MatchRule(%q) matched = %v, want %v",
					test.url, matched, test.wantMatched)
			}
		})
	}
}
