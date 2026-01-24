// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		pattern     string
		patternType string
		negate      bool
		want        bool
	}{
		// Domain matching - exact match only
		{"domain match", "https://github.com/user/repo", "github.com", "domain", false, true},
		{"domain case insensitive", "https://GitHub.COM/user", "github.com", "domain", false, true},
		{"domain no match", "https://gitlab.com", "github.com", "domain", false, false},
		{"domain subdomain no match", "https://api.github.com", "github.com", "domain", false, false},

		// Keyword matching - anywhere in URL
		{"keyword in domain", "https://github.com", "github", "keyword", false, true},
		{"keyword in path", "https://example.com/github/repo", "github", "keyword", false, true},
		{"keyword in query", "https://example.com?repo=github", "github", "keyword", false, true},
		{"keyword case insensitive", "https://GITHUB.com", "github", "keyword", false, true},
		{"keyword no match", "https://gitlab.com", "github", "keyword", false, false},

		// Glob matching - wildcards for subdomains
		{"glob wildcard subdomain", "https://api.github.com", "*.github.com", "glob", false, true},
		{"glob exact", "https://github.com", "github.com", "glob", false, true},
		{"glob no match", "https://github.com", "*.gitlab.com", "glob", false, false},
		{"glob multiple wildcards", "https://api.v2.example.com", "*.*.example.com", "glob", false, true},

		// Regex matching - full control
		{"regex simple", "https://github.com/user/repo", `github\.com`, "regex", false, true},
		{"regex path pattern", "https://github.com/user123/repo", `github\.com/user\d+`, "regex", false, true},
		{"regex no match", "https://github.com", `gitlab\.com`, "regex", false, false},
		{"regex invalid pattern", "https://example.com", "[invalid", "regex", false, false},

		// Unknown pattern type
		{"unknown type", "https://example.com", "example.com", "invalid", false, false},

		// Negation tests
		{"domain match negated", "https://github.com/user", "github.com", "domain", true, false},
		{"domain no match negated", "https://gitlab.com", "github.com", "domain", true, true},
		{"keyword match negated", "https://github.com", "github", "keyword", true, false},
		{"keyword no match negated", "https://example.com", "github", "keyword", true, true},
		{"glob match negated", "https://api.github.com", "*.github.com", "glob", true, false},
		{"glob no match negated", "https://example.com", "*.github.com", "glob", true, true},
		{"regex match negated", "https://github.com", `github\.com`, "regex", true, false},
		{"regex no match negated", "https://example.com", `github\.com`, "regex", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.url, tt.pattern, tt.patternType, tt.negate)
			if result != tt.want {
				t.Errorf("matchesPattern(%q, %q, %q, %v) = %v, want %v",
					tt.url, tt.pattern, tt.patternType, tt.negate, result, tt.want)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		pattern string
		want    bool
	}{
		{"wildcard subdomain", "https://api.example.com", "*.example.com", true},
		{"wildcard any subdomain depth", "https://deep.sub.example.com", "*.example.com", true},
		{"exact match", "https://example.com", "example.com", true},
		{"no match different domain", "https://other.com", "*.example.com", false},
		{"wildcard TLD", "https://example.io", "*.io", true},
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
