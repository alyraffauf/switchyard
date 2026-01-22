// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

// TestRuleMatchesConditions_AND tests AND logic (all conditions must match)
func TestRuleMatchesConditions_AND(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
		url  string
		want bool
	}{
		{
			name: "single condition matches",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "single condition no match",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "gitlab.com"},
				},
			},
			url:  "https://github.com/user/repo",
			want: false,
		},
		{
			name: "multiple conditions all match",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "keyword", Pattern: "user"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "multiple conditions one fails",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "keyword", Pattern: "nonexistent"},
				},
			},
			url:  "https://github.com/user/repo",
			want: false,
		},
		{
			name: "default logic (empty string) defaults to all",
			rule: Rule{
				Logic: "", // Should default to "all"
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "keyword", Pattern: "user"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "no conditions returns false",
			rule: Rule{
				Logic:      "all",
				Conditions: []Condition{},
			},
			url:  "https://github.com/user/repo",
			want: false,
		},
		{
			name: "three conditions all match",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "docs.github.com"},
					{Type: "keyword", Pattern: "api"},
					{Type: "keyword", Pattern: "reference"},
				},
			},
			url:  "https://docs.github.com/api/reference",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.matchesConditions(tt.url)
			if result != tt.want {
				t.Errorf("Rule.matchesConditions(%q) = %v, want %v", tt.url, result, tt.want)
			}
		})
	}
}

// TestRuleMatchesConditions_OR tests OR logic (any condition can match)
func TestRuleMatchesConditions_OR(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
		url  string
		want bool
	}{
		{
			name: "first condition matches",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "domain", Pattern: "gitlab.com"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "second condition matches",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "gitlab.com"},
					{Type: "keyword", Pattern: "github"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "all conditions match",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "keyword", Pattern: "github"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "no conditions match",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "gitlab.com"},
					{Type: "keyword", Pattern: "bitbucket"},
				},
			},
			url:  "https://github.com/user/repo",
			want: false,
		},
		{
			name: "last of many conditions matches",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "gitlab.com"},
					{Type: "domain", Pattern: "bitbucket.com"},
					{Type: "domain", Pattern: "github.com"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "mixed condition types one matches",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "gitlab.com"},
					{Type: "regex", Pattern: "github\\.com/[a-z]+/"},
					{Type: "glob", Pattern: "*.bitbucket.com"},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.matchesConditions(tt.url)
			if result != tt.want {
				t.Errorf("Rule.matchesConditions(%q) = %v, want %v", tt.url, result, tt.want)
			}
		})
	}
}

// TestConfigMatchRule tests the full rule matching from Config
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
				Rules: []Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
					{
						Name:      "GitLab",
						Browser:   "chrome.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
							{Type: "domain", Pattern: "github.com"},
						},
					},
					{
						Name:      "GitLab",
						Browser:   "chrome.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:      "GitHub",
						Browser:   "firefox.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:      "Work Sites",
						Browser:   "",
						AlwaysAsk: true,
						Conditions: []Condition{
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
				Rules: []Rule{},
			},
			url:           "https://example.com",
			wantBrowserID: "",
			wantAlwaysAsk: false,
			wantMatched:   false,
		},
		{
			name: "first matching rule wins",
			config: Config{
				Rules: []Rule{
					{
						Name:      "First Match",
						Browser:   "first.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
							{Type: "keyword", Pattern: "github"},
						},
					},
					{
						Name:      "Second Match",
						Browser:   "second.desktop",
						AlwaysAsk: false,
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:      "GitHub API Docs",
						Browser:   "work-browser.desktop",
						AlwaysAsk: false,
						Logic:     "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:      "Git Platforms",
						Browser:   "dev-browser.desktop",
						AlwaysAsk: false,
						Logic:     "any",
						Conditions: []Condition{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			browserID, alwaysAsk, matched := tt.config.matchRule(tt.url)
			if browserID != tt.wantBrowserID {
				t.Errorf("Config.matchRule(%q) browserID = %q, want %q", tt.url, browserID, tt.wantBrowserID)
			}
			if alwaysAsk != tt.wantAlwaysAsk {
				t.Errorf("Config.matchRule(%q) alwaysAsk = %v, want %v", tt.url, alwaysAsk, tt.wantAlwaysAsk)
			}
			if matched != tt.wantMatched {
				t.Errorf("Config.matchRule(%q) matched = %v, want %v", tt.url, matched, tt.wantMatched)
			}
		})
	}
}

// TestConfigMatchRule_RuleOrdering tests that rules are matched in order (first match wins)
func TestConfigMatchRule_RuleOrdering(t *testing.T) {
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
				Rules: []Rule{
					{
						Name:    "First Rule",
						Browser: "first-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "example.com"},
						},
					},
					{
						Name:    "Second Rule",
						Browser: "second-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:    "Specific Rule",
						Browser: "specific-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "docs.example.com"},
						},
					},
					{
						Name:    "General Rule",
						Browser: "general-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:    "Specific Rule",
						Browser: "specific-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "docs.example.com"},
						},
					},
					{
						Name:    "General Rule",
						Browser: "general-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:    "Non-matching Rule",
						Browser: "wrong-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "other.com"},
						},
					},
					{
						Name:    "Matching Rule",
						Browser: "correct-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{},
			},
			url:           "https://example.com/path",
			wantBrowserID: "",
			wantMatched:   false,
		},
		{
			name: "all rules fail to match",
			config: Config{
				Rules: []Rule{
					{
						Name:    "Rule 1",
						Browser: "browser1.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "other1.com"},
						},
					},
					{
						Name:    "Rule 2",
						Browser: "browser2.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:       "Empty Rule",
						Browser:    "empty-browser.desktop",
						Logic:      "all",
						Conditions: []Condition{},
					},
					{
						Name:    "Valid Rule",
						Browser: "valid-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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
				Rules: []Rule{
					{
						Name:    "AND Rule",
						Browser: "and-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
							{Type: "domain", Pattern: "example.com"},
							{Type: "keyword", Pattern: "admin"}, // won't match
						},
					},
					{
						Name:    "Fallback Rule",
						Browser: "fallback-browser.desktop",
						Logic:   "all",
						Conditions: []Condition{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			browserID, _, matched := tt.config.matchRule(tt.url)
			if browserID != tt.wantBrowserID {
				t.Errorf("Config.matchRule(%q) browserID = %q, want %q",
					tt.url, browserID, tt.wantBrowserID)
			}
			if matched != tt.wantMatched {
				t.Errorf("Config.matchRule(%q) matched = %v, want %v",
					tt.url, matched, tt.wantMatched)
			}
		})
	}
}
