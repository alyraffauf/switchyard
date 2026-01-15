// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

// TestExtractDomain tests URL domain extraction
func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple https url",
			url:      "https://example.com",
			expected: "example.com",
		},
		{
			name:     "https url with path",
			url:      "https://example.com/path/to/page",
			expected: "example.com",
		},
		{
			name:     "https url with port",
			url:      "https://example.com:8080",
			expected: "example.com",
		},
		{
			name:     "https url with port and path",
			url:      "https://example.com:8080/path",
			expected: "example.com",
		},
		{
			name:     "subdomain",
			url:      "https://sub.example.com",
			expected: "sub.example.com",
		},
		{
			name:     "multiple subdomains",
			url:      "https://deep.sub.example.com/path",
			expected: "deep.sub.example.com",
		},
		{
			name:     "http protocol",
			url:      "http://example.com",
			expected: "example.com",
		},
		{
			name:     "no protocol",
			url:      "example.com",
			expected: "example.com",
		},
		{
			name:     "no protocol with path",
			url:      "example.com/path",
			expected: "example.com",
		},
		{
			name:     "url with query params",
			url:      "https://example.com/path?key=value",
			expected: "example.com",
		},
		{
			name:     "url with fragment",
			url:      "https://example.com/path#section",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.url)
			if result != tt.expected {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestSanitizeURL tests URL sanitization logic
func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "already has https",
			url:      "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "already has http",
			url:      "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "bare domain",
			url:      "example.com",
			expected: "https://example.com",
		},
		{
			name:     "bare domain with path",
			url:      "example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "domain with whitespace",
			url:      "  example.com  ",
			expected: "https://example.com",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "whitespace only",
			url:      "   ",
			expected: "",
		},
		{
			name:     "file path with leading slash rejected",
			url:      "/home/user/file.txt",
			expected: "",
		},
		{
			name:     "relative file path rejected",
			url:      "./file.txt",
			expected: "",
		},
		{
			name:     "file:// uri with existing file rejected",
			url:      "file:///etc/hosts",
			expected: "",
		},
		{
			name:     "file:// uri that looks like bare domain converted",
			url:      "file:///nonexistent/path/example.com",
			expected: "https://example.com",
		},
		{
			name:     "file:// uri with nonexistent file rejected",
			url:      "file:///nonexistent/file.txt",
			expected: "",
		},
		{
			name:     "custom protocol",
			url:      "ftp://example.com",
			expected: "ftp://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.url)
			if result != tt.expected {
				t.Errorf("sanitizeURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestMatchGlob tests glob pattern matching
func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		pattern string
		want    bool
	}{
		{
			name:    "wildcard subdomain",
			url:     "https://sub.example.com",
			pattern: "*.example.com",
			want:    true,
		},
		{
			name:    "wildcard subdomain no match",
			url:     "https://different.com",
			pattern: "*.example.com",
			want:    false,
		},
		{
			name:    "exact match",
			url:     "https://example.com",
			pattern: "example.com",
			want:    true,
		},
		{
			name:    "wildcard at end",
			url:     "https://example.com/path",
			pattern: "example.com*",
			want:    true,
		},
		{
			name:    "multiple subdomains",
			url:     "https://deep.sub.example.com",
			pattern: "*.example.com",
			want:    true,
		},
		{
			name:    "wildcard in middle",
			url:     "https://test.example.com",
			pattern: "*.example.*",
			want:    true,
		},
		{
			name:    "invalid pattern causes regex error",
			url:     "https://example.com",
			pattern: "[invalid",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchGlob(tt.url, tt.pattern)
			if result != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.url, tt.pattern, result, tt.want)
			}
		})
	}
}

// TestMatchesPattern tests pattern matching for different types
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		pattern     string
		patternType string
		want        bool
	}{
		// Domain type tests
		{
			name:        "domain exact match",
			url:         "https://github.com",
			pattern:     "github.com",
			patternType: "domain",
			want:        true,
		},
		{
			name:        "domain case insensitive",
			url:         "https://GitHub.COM",
			pattern:     "github.com",
			patternType: "domain",
			want:        true,
		},
		{
			name:        "domain with path still matches",
			url:         "https://github.com/user/repo",
			pattern:     "github.com",
			patternType: "domain",
			want:        true,
		},
		{
			name:        "domain no match different domain",
			url:         "https://gitlab.com",
			pattern:     "github.com",
			patternType: "domain",
			want:        false,
		},
		{
			name:        "domain subdomain no match",
			url:         "https://api.github.com",
			pattern:     "github.com",
			patternType: "domain",
			want:        false,
		},
		// Keyword type tests
		{
			name:        "keyword in domain",
			url:         "https://github.com",
			pattern:     "github",
			patternType: "keyword",
			want:        true,
		},
		{
			name:        "keyword in path",
			url:         "https://example.com/github/repo",
			pattern:     "github",
			patternType: "keyword",
			want:        true,
		},
		{
			name:        "keyword case insensitive",
			url:         "https://GITHUB.com",
			pattern:     "github",
			patternType: "keyword",
			want:        true,
		},
		{
			name:        "keyword no match",
			url:         "https://gitlab.com",
			pattern:     "github",
			patternType: "keyword",
			want:        false,
		},
		// Regex type tests
		{
			name:        "regex simple match",
			url:         "https://github.com/user/repo",
			pattern:     "github\\.com",
			patternType: "regex",
			want:        true,
		},
		{
			name:        "regex with groups",
			url:         "https://github.com/user123/repo",
			pattern:     "github\\.com/user\\d+",
			patternType: "regex",
			want:        true,
		},
		{
			name:        "regex no match",
			url:         "https://github.com",
			pattern:     "gitlab\\.com",
			patternType: "regex",
			want:        false,
		},
		{
			name:        "regex invalid pattern",
			url:         "https://github.com",
			pattern:     "[invalid(regex",
			patternType: "regex",
			want:        false,
		},
		// Glob type tests
		{
			name:        "glob wildcard subdomain",
			url:         "https://api.github.com",
			pattern:     "*.github.com",
			patternType: "glob",
			want:        true,
		},
		{
			name:        "glob exact",
			url:         "https://github.com",
			pattern:     "github.com",
			patternType: "glob",
			want:        true,
		},
		// Unknown type
		{
			name:        "unknown type returns false",
			url:         "https://github.com",
			pattern:     "github.com",
			patternType: "unknown",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.url, tt.pattern, tt.patternType)
			if result != tt.want {
				t.Errorf("matchesPattern(%q, %q, %q) = %v, want %v",
					tt.url, tt.pattern, tt.patternType, result, tt.want)
			}
		})
	}
}

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
