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
			name:     "file:// uri with existing file passed through",
			url:      "file:///etc/hosts",
			expected: "file:///etc/hosts",
		},
		{
			name:     "file:// uri with nonexistent path passed through",
			url:      "file:///nonexistent/path/example.com",
			expected: "file:///nonexistent/path/example.com",
		},
		{
			name:     "file:// uri with nonexistent file passed through",
			url:      "file:///nonexistent/file.txt",
			expected: "file:///nonexistent/file.txt",
		},
		{
			name:     "ftp protocol",
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

// TestExtractDomain_EdgeCases tests URL domain extraction with edge cases
func TestExtractDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with authentication credentials",
			url:      "https://user:password@example.com/path",
			expected: "example.com",
		},
		{
			name:     "URL with username only",
			url:      "https://user@example.com/path",
			expected: "example.com",
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
			name:     "mailto protocol extracts host",
			url:      "mailto:user@example.com",
			expected: "example.com",
		},
		{
			name:     "URL with empty path",
			url:      "https://example.com/",
			expected: "example.com",
		},
		{
			name:     "URL with only question mark",
			url:      "https://example.com?",
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
			name:        "data URI has no domain",
			url:         "data:text/html,<h1>Hello</h1>",
			pattern:     "data",
			patternType: "domain",
			expected:    false, // data: URIs have no host component
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

// TestSanitizeURL_EdgeCases tests URL sanitization with edge cases
func TestSanitizeURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		// URIs that should be handled by xdg-open, not browsers
		{
			name:     "data URI is rejected",
			url:      "data:text/html,<h1>Test</h1>",
			expected: "",
		},
		{
			name:     "javascript URI is rejected",
			url:      "javascript:void(0)",
			expected: "",
		},
		{
			name:     "blob URI is rejected",
			url:      "blob:https://example.com/guid",
			expected: "",
		},
		{
			name:     "mailto URI is rejected",
			url:      "mailto:user@example.com",
			expected: "",
		},
		{
			name:     "tel URI is rejected",
			url:      "tel:+1234567890",
			expected: "",
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
			name:     "custom app scheme is rejected",
			url:      "myapp://action/param",
			expected: "",
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
