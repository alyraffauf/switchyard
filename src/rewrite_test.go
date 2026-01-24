// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		want    string
	}{
		{"youtube.com", `youtube\.com`},
		{"*.example.com", `.*\.example\.com`},
		{"utm_*", `utm_.*`},
		{"foo*bar", `foo.*bar`},
		{"a.b.c", `a\.b\.c`},
		{"test?param", `test\?param`}, // ? is escaped, not a wildcard
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := wildcardToRegex(tt.pattern)
			if got != tt.want {
				t.Errorf("wildcardToRegex(%q) = %q, want %q", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestApplyDomainRedirection(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		find    string
		replace string
		want    string
	}{
		{
			name:    "exact domain match",
			url:     "https://reddit.com/r/test",
			find:    "reddit.com",
			replace: "old.reddit.com",
			want:    "https://old.reddit.com/r/test",
		},
		{
			name:    "no match on subdomain",
			url:     "https://old.reddit.com/r/test",
			find:    "reddit.com",
			replace: "old.reddit.com",
			want:    "https://old.reddit.com/r/test", // unchanged
		},
		{
			name:    "case insensitive",
			url:     "https://Reddit.COM/r/test",
			find:    "reddit.com",
			replace: "old.reddit.com",
			want:    "https://old.reddit.com/r/test",
		},
		{
			name:    "preserves port",
			url:     "https://reddit.com:8080/r/test",
			find:    "reddit.com",
			replace: "old.reddit.com",
			want:    "https://old.reddit.com:8080/r/test",
		},
		{
			name:    "preserves path and query",
			url:     "https://twitter.com/user?tab=posts",
			find:    "twitter.com",
			replace: "nitter.net",
			want:    "https://nitter.net/user?tab=posts",
		},
		{
			name:    "no match different domain",
			url:     "https://github.com/user",
			find:    "twitter.com",
			replace: "nitter.net",
			want:    "https://github.com/user",
		},
		{
			name:    "x.com to twitter.com",
			url:     "https://x.com/user",
			find:    "x.com",
			replace: "twitter.com",
			want:    "https://twitter.com/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Redirection{Type: "domain", Find: tt.find, Replace: tt.replace}
			got := applyRedirection(tt.url, r)
			if got != tt.want {
				t.Errorf("applyRedirection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyPatternRedirection(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		find    string
		replace string
		want    string
	}{
		{
			name:    "wildcard removes utm mid-url",
			url:     "https://example.com?utm_source=twitter&id=1",
			find:    "utm_source=*&",
			replace: "",
			want:    "https://example.com?id=1",
		},
		{
			name:    "wildcard removes fbclid at end",
			url:     "https://example.com?id=1&fbclid=abc123",
			find:    "&fbclid=*",
			replace: "",
			want:    "https://example.com?id=1",
		},
		{
			name:    "wildcard at start of param",
			url:     "https://example.com?fbclid=abc123",
			find:    "fbclid=*",
			replace: "",
			want:    "https://example.com?",
		},
		{
			name:    "multiple occurrences all replaced",
			url:     "https://foo.com/foo/foo",
			find:    "foo",
			replace: "bar",
			want:    "https://bar.com/bar/bar",
		},
		{
			name:    "empty replace removes match",
			url:     "https://example.com/tracking/page",
			find:    "/tracking",
			replace: "",
			want:    "https://example.com/page",
		},
		{
			name:    "wildcard matches anything",
			url:     "https://cdn.example.com/image.png",
			find:    "cdn.*",
			replace: "static.newsite.com",
			want:    "https://static.newsite.com",
		},
		{
			name:    "case insensitive",
			url:     "https://example.com?UTM_SOURCE=twitter",
			find:    "utm_source=*",
			replace: "",
			want:    "https://example.com?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Redirection{Type: "pattern", Find: tt.find, Replace: tt.replace}
			got := applyRedirection(tt.url, r)
			if got != tt.want {
				t.Errorf("applyRedirection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyRegexRedirection(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		find    string
		replace string
		want    string
	}{
		{
			name:    "youtube shorts to watch",
			url:     "https://youtube.com/shorts/abc123xyz",
			find:    "youtube\\.com/shorts/([^?]+)",
			replace: "youtube.com/watch?v=$1",
			want:    "https://youtube.com/watch?v=abc123xyz",
		},
		{
			name:    "capture group replacement",
			url:     "https://foo.example.com/123",
			find:    "([a-z]+)\\.example\\.com/([0-9]+)",
			replace: "new-$1.example.org/id/$2",
			want:    "https://new-foo.example.org/id/123",
		},
		{
			name:    "remove utm parameters",
			url:     "https://example.com?utm_source=twitter&utm_medium=social&id=1",
			find:    "[?&]utm_[a-z_]+=[^&]*",
			replace: "",
			want:    "https://example.com&id=1",
		},
		{
			name:    "strip amazon tracking",
			url:     "https://amazon.com/dp/B001234/ref=sr_1_1?keywords=test",
			find:    "(amazon\\.[^/]+/dp/[^/]+).*",
			replace: "$1",
			want:    "https://amazon.com/dp/B001234",
		},
		{
			name:    "no match returns original",
			url:     "https://example.com/page",
			find:    "nomatch",
			replace: "replaced",
			want:    "https://example.com/page",
		},
		{
			name:    "empty replace removes match",
			url:     "https://example.com/tracking/page",
			find:    "/tracking",
			replace: "",
			want:    "https://example.com/page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Redirection{Type: "regex", Find: tt.find, Replace: tt.replace}
			got := applyRedirection(tt.url, r)
			if got != tt.want {
				t.Errorf("applyRedirection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyRedirectionDefaultType(t *testing.T) {
	// Empty type should default to domain
	r := Redirection{Find: "reddit.com", Replace: "old.reddit.com"}
	got := applyRedirection("https://reddit.com/r/test", r)
	want := "https://old.reddit.com/r/test"
	if got != want {
		t.Errorf("applyRedirection() with empty type = %q, want %q", got, want)
	}

	// Should not match subdomain with default type
	got = applyRedirection("https://old.reddit.com/r/test", r)
	want = "https://old.reddit.com/r/test"
	if got != want {
		t.Errorf("applyRedirection() should not match subdomain = %q, want %q", got, want)
	}
}

func TestApplyRedirections(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		redirections []Redirection
		want         string
	}{
		{
			name: "domain then pattern redirection",
			url:  "https://twitter.com/user?utm_source=share",
			redirections: []Redirection{
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
				{Type: "pattern", Find: "?utm_source=*", Replace: ""},
			},
			want: "https://nitter.net/user",
		},
		{
			name:         "empty redirections list",
			url:          "https://example.com",
			redirections: []Redirection{},
			want:         "https://example.com",
		},
		{
			name: "chained domain redirections",
			url:  "https://x.com/user",
			redirections: []Redirection{
				{Type: "domain", Find: "x.com", Replace: "twitter.com"},
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
			},
			want: "https://nitter.net/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyRedirections(tt.url, tt.redirections)
			if got != tt.want {
				t.Errorf("applyRedirections() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateRedirection(t *testing.T) {
	tests := []struct {
		name        string
		redirection Redirection
		wantErr     bool
	}{
		{
			name:        "valid domain redirection",
			redirection: Redirection{Type: "domain", Find: "reddit.com", Replace: "old.reddit.com"},
			wantErr:     false,
		},
		{
			name:        "valid domain redirection default type",
			redirection: Redirection{Find: "twitter.com", Replace: "nitter.net"},
			wantErr:     false,
		},
		{
			name:        "valid pattern redirection with wildcard",
			redirection: Redirection{Type: "pattern", Find: "utm_*", Replace: ""},
			wantErr:     false,
		},
		{
			name:        "domain redirection with wildcard invalid",
			redirection: Redirection{Type: "domain", Find: "*.reddit.com", Replace: "old.reddit.com"},
			wantErr:     true,
		},
		{
			name:        "empty find pattern",
			redirection: Redirection{Type: "domain", Find: "", Replace: "something"},
			wantErr:     true,
		},
		{
			name:        "empty replace is valid",
			redirection: Redirection{Type: "pattern", Find: "tracking", Replace: ""},
			wantErr:     false,
		},
		{
			name:        "invalid redirection type",
			redirection: Redirection{Type: "invalid", Find: "test", Replace: ""},
			wantErr:     true,
		},
		{
			name:        "valid regex redirection",
			redirection: Redirection{Type: "regex", Find: "youtube\\.com/shorts/([^?]+)", Replace: "youtube.com/watch?v=$1"},
			wantErr:     false,
		},
		{
			name:        "valid regex with capture groups",
			redirection: Redirection{Type: "regex", Find: "([a-z]+)\\.example\\.com", Replace: "$1.newsite.com"},
			wantErr:     false,
		},
		{
			name:        "invalid regex syntax",
			redirection: Redirection{Type: "regex", Find: "[invalid(regex", Replace: ""},
			wantErr:     true,
		},
		{
			name:        "regex with empty replace is valid",
			redirection: Redirection{Type: "regex", Find: "[?&]utm_[^&]*", Replace: ""},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRedirection(tt.redirection)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRedirection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
