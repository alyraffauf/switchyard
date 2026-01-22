// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Common inputs from users and applications
		{"bare domain", "example.com", "https://example.com"},
		{"bare domain with path", "example.com/page", "https://example.com/page"},
		{"bare domain with whitespace", "  example.com  ", "https://example.com"},
		{"https URL", "https://example.com", "https://example.com"},
		{"http URL", "http://example.com", "http://example.com"},
		{"https with path and query", "https://example.com/path?q=1", "https://example.com/path?q=1"},

		// File URLs - pass through for local HTML files
		{"file URL", "file:///home/user/doc.html", "file:///home/user/doc.html"},

		// FTP - still used for some downloads
		{"ftp URL", "ftp://ftp.example.com/file.zip", "ftp://ftp.example.com/file.zip"},

		// Schemes that should go to xdg-open, not browsers
		{"mailto rejected", "mailto:user@example.com", ""},
		{"tel rejected", "tel:+1234567890", ""},
		{"javascript rejected", "javascript:void(0)", ""},
		{"data rejected", "data:text/html,<h1>Hi</h1>", ""},

		// Invalid/unsupported inputs
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"absolute path rejected", "/home/user/file.html", ""},
		{"relative path rejected", "./file.html", ""},
		{"unknown scheme rejected", "myapp://action", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Standard URLs
		{"https URL", "https://example.com", "example.com"},
		{"https with path", "https://example.com/path/page", "example.com"},
		{"https with port", "https://example.com:8080/path", "example.com"},
		{"https with query", "https://example.com?q=test", "example.com"},
		{"http URL", "http://example.com", "example.com"},

		// Subdomains
		{"subdomain", "https://www.example.com", "www.example.com"},
		{"deep subdomain", "https://api.v2.example.com", "api.v2.example.com"},

		// Bare domains (no scheme) - common user input
		{"bare domain", "example.com", "example.com"},
		{"bare domain with path", "example.com/path", "example.com"},

		// Auth in URL (legacy but still seen)
		{"URL with credentials", "https://user:pass@example.com", "example.com"},

		// IP addresses
		{"IPv4", "https://192.168.1.1", "192.168.1.1"},
		{"IPv4 with port", "https://192.168.1.1:8080", "192.168.1.1"},
		{"localhost", "http://localhost:3000", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.input)
			if result != tt.expected {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSwitchyardURL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantURL      string
		wantBrowsers []string
		wantErr      bool
	}{
		// Valid URLs
		{
			name:         "basic with browser",
			input:        "switchyard://open?url=https://example.com&browser=org.mozilla.firefox",
			wantURL:      "https://example.com",
			wantBrowsers: []string{"org.mozilla.firefox"},
		},
		{
			name:         "multiple browsers",
			input:        "switchyard://open?url=https://example.com&browser=org.mozilla.firefox,com.google.Chrome",
			wantURL:      "https://example.com",
			wantBrowsers: []string{"org.mozilla.firefox", "com.google.Chrome"},
		},
		{
			name:         "no browser specified",
			input:        "switchyard://open?url=https://example.com",
			wantURL:      "https://example.com",
			wantBrowsers: nil,
		},
		{
			name:         "encoded URL",
			input:        "switchyard://open?url=https%3A%2F%2Fexample.com%3Ffoo%3Dbar&browser=org.mozilla.firefox",
			wantURL:      "https://example.com?foo=bar",
			wantBrowsers: []string{"org.mozilla.firefox"},
		},

		// Invalid URLs
		{
			name:    "wrong scheme",
			input:   "http://open?url=https://example.com",
			wantErr: true,
		},
		{
			name:    "wrong host",
			input:   "switchyard://launch?url=https://example.com",
			wantErr: true,
		},
		{
			name:    "missing url parameter",
			input:   "switchyard://open?browser=org.mozilla.firefox",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotBrowsers, err := parseSwitchyardURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSwitchyardURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("parseSwitchyardURL() url = %q, want %q", gotURL, tt.wantURL)
			}
			if len(gotBrowsers) != len(tt.wantBrowsers) {
				t.Errorf("parseSwitchyardURL() browsers = %v, want %v", gotBrowsers, tt.wantBrowsers)
				return
			}
			for i, b := range gotBrowsers {
				if b != tt.wantBrowsers[i] {
					t.Errorf("parseSwitchyardURL() browsers[%d] = %q, want %q", i, b, tt.wantBrowsers[i])
				}
			}
		})
	}
}
