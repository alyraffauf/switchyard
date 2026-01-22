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
		want        bool
	}{
		// Domain matching - exact match only
		{"domain match", "https://github.com/user/repo", "github.com", "domain", true},
		{"domain case insensitive", "https://GitHub.COM/user", "github.com", "domain", true},
		{"domain no match", "https://gitlab.com", "github.com", "domain", false},
		{"domain subdomain no match", "https://api.github.com", "github.com", "domain", false},

		// Keyword matching - anywhere in URL
		{"keyword in domain", "https://github.com", "github", "keyword", true},
		{"keyword in path", "https://example.com/github/repo", "github", "keyword", true},
		{"keyword in query", "https://example.com?repo=github", "github", "keyword", true},
		{"keyword case insensitive", "https://GITHUB.com", "github", "keyword", true},
		{"keyword no match", "https://gitlab.com", "github", "keyword", false},

		// Glob matching - wildcards for subdomains
		{"glob wildcard subdomain", "https://api.github.com", "*.github.com", "glob", true},
		{"glob exact", "https://github.com", "github.com", "glob", true},
		{"glob no match", "https://github.com", "*.gitlab.com", "glob", false},
		{"glob multiple wildcards", "https://api.v2.example.com", "*.*.example.com", "glob", true},

		// Regex matching - full control
		{"regex simple", "https://github.com/user/repo", `github\.com`, "regex", true},
		{"regex path pattern", "https://github.com/user123/repo", `github\.com/user\d+`, "regex", true},
		{"regex no match", "https://github.com", `gitlab\.com`, "regex", false},
		{"regex invalid pattern", "https://example.com", "[invalid", "regex", false},

		// Unknown pattern type
		{"unknown type", "https://example.com", "example.com", "invalid", false},
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
