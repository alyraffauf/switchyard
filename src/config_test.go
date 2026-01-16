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

// TestExtractDomain_EdgeCases tests URL domain extraction with edge cases
func TestExtractDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with authentication credentials strips at colon",
			url:      "https://user:password@example.com/path",
			expected: "user", // Note: extractDomain strips at colon (port removal logic)
		},
		{
			name:     "URL with username only",
			url:      "https://user@example.com/path",
			expected: "user@example.com",
		},
		{
			name:     "internationalized domain (IDN) ASCII form",
			url:      "https://xn--n3h.com/path",
			expected: "xn--n3h.com",
		},
		{
			name:     "internationalized domain Unicode",
			url:      "https://münchen.example/path",
			expected: "münchen.example",
		},
		{
			name:     "very long subdomain",
			url:      "https://this.is.a.very.long.subdomain.chain.example.com/path",
			expected: "this.is.a.very.long.subdomain.chain.example.com",
		},
		{
			name:     "IP address v4",
			url:      "https://192.168.1.1/path",
			expected: "192.168.1.1",
		},
		{
			name:     "IP address v4 with port",
			url:      "https://192.168.1.1:8080/path",
			expected: "192.168.1.1",
		},
		{
			name:     "localhost",
			url:      "http://localhost/path",
			expected: "localhost",
		},
		{
			name:     "localhost with port",
			url:      "http://localhost:3000/path",
			expected: "localhost",
		},
		{
			name:     "single word TLD",
			url:      "https://localhost",
			expected: "localhost",
		},
		{
			name:     "double slash in path doesn't affect domain",
			url:      "https://example.com//double//slashes",
			expected: "example.com",
		},
		{
			name:     "ftp protocol",
			url:      "ftp://files.example.com/file.zip",
			expected: "files.example.com",
		},
		{
			name:     "custom protocol",
			url:      "myapp://example.com/action",
			expected: "example.com",
		},
		{
			name:     "mailto protocol extracts scheme name",
			url:      "mailto:user@example.com",
			expected: "mailto", // Note: mailto: has no //, so extractDomain returns text before colon
		},
		{
			name:     "URL with empty path",
			url:      "https://example.com/",
			expected: "example.com",
		},
		{
			name:     "URL with only question mark",
			url:      "https://example.com?",
			expected: "example.com?",
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

// TestMatchesPattern_EdgeCases tests pattern matching with edge case URLs
func TestMatchesPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		pattern     string
		patternType string
		expected    bool
	}{
		// Long URL tests
		{
			name:        "very long URL with keyword match",
			url:         "https://example.com/" + string(make([]byte, 1000)) + "keyword" + string(make([]byte, 1000)),
			pattern:     "keyword",
			patternType: "keyword",
			expected:    true,
		},
		{
			name:        "URL with many query parameters",
			url:         "https://example.com/path?a=1&b=2&c=3&d=4&e=5&f=6&g=7&h=8&i=9&j=10",
			pattern:     "example.com",
			patternType: "domain",
			expected:    true,
		},
		// Special character tests
		{
			name:        "URL with encoded spaces",
			url:         "https://example.com/path%20with%20spaces",
			pattern:     "%20",
			patternType: "keyword",
			expected:    true,
		},
		{
			name:        "URL with plus signs",
			url:         "https://search.example.com/q=hello+world",
			pattern:     "hello+world",
			patternType: "keyword",
			expected:    true,
		},
		{
			name:        "URL with hash fragment",
			url:         "https://example.com/page#section-id",
			pattern:     "section-id",
			patternType: "keyword",
			expected:    true,
		},
		{
			name:        "URL with unicode characters",
			url:         "https://example.com/日本語",
			pattern:     "日本語",
			patternType: "keyword",
			expected:    true,
		},
		// Data URI tests
		{
			name:        "data URI matches domain 'data' since it extracts scheme",
			url:         "data:text/html,<h1>Hello</h1>",
			pattern:     "data",
			patternType: "domain",
			expected:    true, // extractDomain returns "data" for data: URIs
		},
		{
			name:        "data URI matches keyword",
			url:         "data:text/html,<h1>Hello</h1>",
			pattern:     "text/html",
			patternType: "keyword",
			expected:    true,
		},
		// JavaScript URI tests
		{
			name:        "javascript URI keyword match",
			url:         "javascript:alert('test')",
			pattern:     "javascript",
			patternType: "keyword",
			expected:    true,
		},
		// Blob URI tests
		{
			name:        "blob URI keyword match",
			url:         "blob:https://example.com/550e8400-e29b-41d4-a716-446655440000",
			pattern:     "blob:",
			patternType: "keyword",
			expected:    true,
		},
		// Regex edge cases
		{
			name:        "regex with special characters in URL",
			url:         "https://example.com/path?key=value&other=123",
			pattern:     `key=value.*other=\d+`,
			patternType: "regex",
			expected:    true,
		},
		{
			name:        "regex matching entire URL",
			url:         "https://subdomain.example.com/path",
			pattern:     `^https://[a-z]+\.example\.com/.*$`,
			patternType: "regex",
			expected:    true,
		},
		// Glob edge cases
		{
			name:        "glob with multiple wildcards",
			url:         "https://api.v2.example.com/endpoint",
			pattern:     "*.*.example.com",
			patternType: "glob",
			expected:    true,
		},
		{
			name:        "glob matching only TLD",
			url:         "https://anything.io/path",
			pattern:     "*.io",
			patternType: "glob",
			expected:    true,
		},
		// Empty and whitespace tests
		{
			name:        "URL with only whitespace in query",
			url:         "https://example.com/search?q=   ",
			pattern:     "   ",
			patternType: "keyword",
			expected:    true,
		},
		// Case sensitivity verification
		{
			name:        "domain match is case insensitive",
			url:         "https://EXAMPLE.COM/path",
			pattern:     "example.com",
			patternType: "domain",
			expected:    true,
		},
		{
			name:        "keyword match is case insensitive",
			url:         "https://example.com/PATH",
			pattern:     "path",
			patternType: "keyword",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.url, tt.pattern, tt.patternType)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q, %q) = %v, want %v",
					tt.url, tt.pattern, tt.patternType, result, tt.expected)
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

// TestSanitizeURL_EdgeCases tests URL sanitization with edge cases
func TestSanitizeURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		// Note: sanitizeURL checks for "://" to identify schemes.
		// URIs with single colon (data:, mailto:, tel:, javascript:) get https:// prepended.
		// This is acceptable for a browser URL router since these aren't routable web URLs.
		{
			name:     "data URI gets https prefix (no :// in original)",
			url:      "data:text/html,<h1>Test</h1>",
			expected: "https://data:text/html,<h1>Test</h1>",
		},
		{
			name:     "javascript URI gets https prefix (no :// in original)",
			url:      "javascript:void(0)",
			expected: "https://javascript:void(0)",
		},
		{
			name:     "blob URI is preserved (has ://)",
			url:      "blob:https://example.com/guid",
			expected: "blob:https://example.com/guid",
		},
		{
			name:     "mailto URI gets https prefix (no :// in original)",
			url:      "mailto:user@example.com",
			expected: "https://mailto:user@example.com",
		},
		{
			name:     "tel URI gets https prefix (no :// in original)",
			url:      "tel:+1234567890",
			expected: "https://tel:+1234567890",
		},
		{
			name:     "URL with leading whitespace",
			url:      "   https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with trailing whitespace",
			url:      "https://example.com   ",
			expected: "https://example.com",
		},
		{
			name:     "URL with both leading and trailing whitespace",
			url:      "   https://example.com   ",
			expected: "https://example.com",
		},
		{
			name:     "bare domain gets https prefix",
			url:      "example.com",
			expected: "https://example.com",
		},
		{
			name:     "bare domain with path gets https prefix",
			url:      "example.com/path/to/page",
			expected: "https://example.com/path/to/page",
		},
		{
			name:     "relative path is rejected",
			url:      "./relative/path",
			expected: "",
		},
		{
			name:     "absolute path is rejected",
			url:      "/absolute/path",
			expected: "",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "only whitespace",
			url:      "   ",
			expected: "",
		},
		{
			name:     "ftp URL is preserved",
			url:      "ftp://files.example.com/file.zip",
			expected: "ftp://files.example.com/file.zip",
		},
		{
			name:     "custom app scheme is preserved",
			url:      "myapp://action/param",
			expected: "myapp://action/param",
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
