// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		wildcard string
		want     string
	}{
		{"youtube.com", `youtube\.com`},
		{"*.example.com", `.*\.example\.com`},
		{"utm_*", `utm_.*`},
		{"foo*bar", `foo.*bar`},
		{"a.b.c", `a\.b\.c`},
		{"test?param", `test\?param`}, // ? is escaped, not a wildcard
	}

	for _, tt := range tests {
		t.Run(tt.wildcard, func(t *testing.T) {
			got := wildcardToRegex(tt.wildcard)
			if got != tt.want {
				t.Errorf("wildcardToRegex(%q) = %q, want %q", tt.wildcard, got, tt.want)
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
		{
			name:    "preserves fragment",
			url:     "https://example.com/page#section",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://new.example.com/page#section",
		},
		{
			name:    "preserves username and password",
			url:     "https://user:pass@example.com/path",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://user:pass@new.example.com/path",
		},
		{
			name:    "http scheme preserved",
			url:     "http://example.com/path",
			find:    "example.com",
			replace: "new.example.com",
			want:    "http://new.example.com/path",
		},
		{
			name:    "does not match partial domain",
			url:     "https://notexample.com/path",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://notexample.com/path",
		},
		{
			name:    "does not match domain in path",
			url:     "https://other.com/example.com/path",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://other.com/example.com/path",
		},
		{
			name:    "empty replace removes domain",
			url:     "https://www.example.com/path",
			find:    "www.example.com",
			replace: "example.com",
			want:    "https://example.com/path",
		},
		{
			name:    "handles malformed URL gracefully",
			url:     "not a url",
			find:    "example.com",
			replace: "new.example.com",
			want:    "not a url",
		},
		{
			name:    "preserves complex query string",
			url:     "https://example.com/search?q=test&page=1&sort=date",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://new.example.com/search?q=test&page=1&sort=date",
		},
		{
			name:    "handles encoded characters in path",
			url:     "https://example.com/path%20with%20spaces",
			find:    "example.com",
			replace: "new.example.com",
			want:    "https://new.example.com/path%20with%20spaces",
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

func TestApplyWildcardRedirection(t *testing.T) {
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
		{
			name:    "wildcard at beginning matches greedily",
			url:     "https://www.example.com/path",
			find:    "*www.",
			replace: "",
			want:    "example.com/path", // wildcard matches "https://" too
		},
		{
			name:    "multiple wildcards",
			url:     "https://example.com/a/b/c/d",
			find:    "/a/*/c/*",
			replace: "/x",
			want:    "https://example.com/x",
		},
		{
			name:    "wildcard with special regex chars",
			url:     "https://example.com/path?query=value",
			find:    "?query=*",
			replace: "",
			want:    "https://example.com/path",
		},
		{
			name:    "no match returns original",
			url:     "https://example.com/path",
			find:    "nomatch*",
			replace: "replaced",
			want:    "https://example.com/path",
		},
		{
			name:    "wildcard matches empty string",
			url:     "https://example.com/path",
			find:    "path*",
			replace: "newpath",
			want:    "https://example.com/newpath",
		},
		{
			name:    "replace with literal text",
			url:     "https://old.example.com/path",
			find:    "old.",
			replace: "new.",
			want:    "https://new.example.com/path",
		},
		{
			name:    "handles URL with fragment",
			url:     "https://example.com/path?utm_source=test#section",
			find:    "?utm_source=*#",
			replace: "#",
			want:    "https://example.com/path#section",
		},
		{
			name:    "greedy wildcard behavior",
			url:     "https://example.com/a/b/a/c",
			find:    "/a/*",
			replace: "/x",
			want:    "https://example.com/x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Redirection{Type: "wildcard", Find: tt.find, Replace: tt.replace}
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
		// Edge cases
		{
			name:    "multiple capture groups",
			url:     "https://example.com/user/123/post/456",
			find:    "/user/([0-9]+)/post/([0-9]+)",
			replace: "/u/$1/p/$2",
			want:    "https://example.com/u/123/p/456",
		},
		{
			name:    "named-style capture groups with numbers",
			url:     "https://example.com/2024/01/15/article",
			find:    "/([0-9]{4})/([0-9]{2})/([0-9]{2})/",
			replace: "/archive/$1-$2-$3/",
			want:    "https://example.com/archive/2024-01-15/article",
		},
		{
			name:    "case sensitive by default",
			url:     "https://Example.COM/PATH",
			find:    "example\\.com/path",
			replace: "new.example.com/newpath",
			want:    "https://Example.COM/PATH", // no match - case sensitive
		},
		{
			name:    "case insensitive with flag",
			url:     "https://Example.COM/PATH",
			find:    "(?i)example\\.com/path",
			replace: "new.example.com/newpath",
			want:    "https://new.example.com/newpath",
		},
		{
			name:    "invalid regex returns original",
			url:     "https://example.com/page",
			find:    "[invalid",
			replace: "replaced",
			want:    "https://example.com/page",
		},
		{
			name:    "backreference to non-existent group",
			url:     "https://example.com/test",
			find:    "test",
			replace: "$1", // no capture group
			want:    "https://example.com/",
		},
		{
			name:    "literal dollar sign in replacement",
			url:     "https://example.com/price",
			find:    "price",
			replace: "cost$$100",
			want:    "https://example.com/cost$100",
		},
		{
			name:    "google to kagi search",
			url:     "https://www.google.com/search?q=test+query&sourceid=chrome",
			find:    "google\\.com/search\\?(.*)q=([^&]+)(.*)",
			replace: "kagi.com/search?q=$2",
			want:    "https://www.kagi.com/search?q=test+query",
		},
		{
			name:    "handles special regex chars in URL",
			url:     "https://example.com/path?a=1&b=2",
			find:    "\\?a=1&b=2",
			replace: "?x=y",
			want:    "https://example.com/path?x=y",
		},
		{
			name:    "preserves unmatched parts",
			url:     "https://example.com/prefix/match/suffix",
			find:    "/match/",
			replace: "/replaced/",
			want:    "https://example.com/prefix/replaced/suffix",
		},
		{
			name:    "handles empty capture group",
			url:     "https://example.com/test",
			find:    "/(test)(.*)",
			replace: "/$1-end$2",
			want:    "https://example.com/test-end",
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
			name: "domain then wildcard redirection",
			url:  "https://twitter.com/user?utm_source=share",
			redirections: []Redirection{
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
				{Type: "wildcard", Find: "?utm_source=*", Replace: ""},
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
			name:        "valid wildcard redirection with wildcard",
			redirection: Redirection{Type: "wildcard", Find: "utm_*", Replace: ""},
			wantErr:     false,
		},
		{
			name:        "domain redirection with wildcard invalid",
			redirection: Redirection{Type: "domain", Find: "*.reddit.com", Replace: "old.reddit.com"},
			wantErr:     true,
		},
		{
			name:        "empty find wildcard",
			redirection: Redirection{Type: "domain", Find: "", Replace: "something"},
			wantErr:     true,
		},
		{
			name:        "empty replace is valid",
			redirection: Redirection{Type: "wildcard", Find: "tracking", Replace: ""},
			wantErr:     false,
		},
		{
			name:        "invalid redirection type",
			redirection: Redirection{Type: "invalid", Find: "test", Replace: ""},
			wantErr:     true,
		},
		{
			name:        "valid regex redirection",
			redirection: Redirection{Type: "regex", Find: "test.*pattern", Replace: "replaced"},
			wantErr:     false,
		},
		{
			name:        "invalid regex syntax",
			redirection: Redirection{Type: "regex", Find: "[invalid", Replace: "replaced"},
			wantErr:     true,
		},
		{
			name:        "regex with capture groups valid",
			redirection: Redirection{Type: "regex", Find: "([a-z]+)/([0-9]+)", Replace: "$2/$1"},
			wantErr:     false,
		},
		{
			name:        "wildcard without asterisk valid",
			redirection: Redirection{Type: "wildcard", Find: "exactmatch", Replace: "replaced"},
			wantErr:     false,
		},
		{
			name:        "domain with port invalid",
			redirection: Redirection{Type: "domain", Find: "example.com:8080", Replace: "example.com"},
			wantErr:     true, // colon is invalid in domain pattern
		},
		{
			name:        "domain with path invalid",
			redirection: Redirection{Type: "domain", Find: "example.com/path", Replace: "new.com"},
			wantErr:     true,
		},
		{
			name:        "domain with protocol invalid",
			redirection: Redirection{Type: "domain", Find: "https://example.com", Replace: "new.com"},
			wantErr:     true,
		},
		{
			name:        "whitespace only find invalid",
			redirection: Redirection{Type: "domain", Find: "   ", Replace: "example.com"},
			wantErr:     true,
		},
		{
			name:        "name field does not affect validation",
			redirection: Redirection{Name: "My Rule", Type: "domain", Find: "example.com", Replace: "new.com"},
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

// TestRealWorldRedirections tests common real-world URL transformations
func TestRealWorldRedirections(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		redirections []Redirection
		want         string
	}{
		{
			name: "twitter/x.com to nitter",
			url:  "https://x.com/user/status/123456",
			redirections: []Redirection{
				{Type: "domain", Find: "x.com", Replace: "nitter.net"},
			},
			want: "https://nitter.net/user/status/123456",
		},
		{
			name: "reddit to old reddit",
			url:  "https://reddit.com/r/programming/comments/abc123",
			redirections: []Redirection{
				{Type: "domain", Find: "reddit.com", Replace: "old.reddit.com"},
			},
			want: "https://old.reddit.com/r/programming/comments/abc123",
		},
		{
			name: "remove facebook tracking params",
			url:  "https://example.com/article?fbclid=ABC123&utm_source=facebook",
			redirections: []Redirection{
				// Use regex for more reliable param removal
				{Type: "regex", Find: "[?&]fbclid=[^&]*", Replace: ""},
				{Type: "regex", Find: "[?&]utm_source=[^&]*", Replace: ""},
			},
			want: "https://example.com/article",
		},
		{
			name: "clean amazon product URL",
			url:  "https://www.amazon.com/dp/B0DZD91W4F/?tag=affiliate-20&linkCode=xyz&ref=abc",
			redirections: []Redirection{
				{Type: "regex", Find: "(amazon\\.[a-z.]+/dp/[A-Z0-9]+).*", Replace: "$1"},
			},
			want: "https://www.amazon.com/dp/B0DZD91W4F",
		},
		{
			name: "google to kagi search",
			url:  "https://www.google.com/search?q=test+search&sourceid=chrome&ie=UTF-8",
			redirections: []Redirection{
				{Type: "regex", Find: "google\\.com/search\\?(.*)q=([^&]+)(.*)", Replace: "kagi.com/search?q=$2"},
			},
			want: "https://www.kagi.com/search?q=test+search",
		},
		{
			name: "youtube mobile to desktop",
			url:  "https://m.youtube.com/watch?v=abc123",
			redirections: []Redirection{
				{Type: "domain", Find: "m.youtube.com", Replace: "youtube.com"},
			},
			want: "https://youtube.com/watch?v=abc123",
		},
		{
			name: "instagram to bibliogram",
			url:  "https://www.instagram.com/p/ABC123/",
			redirections: []Redirection{
				{Type: "domain", Find: "www.instagram.com", Replace: "bibliogram.art"},
				{Type: "domain", Find: "instagram.com", Replace: "bibliogram.art"},
			},
			want: "https://bibliogram.art/p/ABC123/",
		},
		{
			name: "medium to scribe",
			url:  "https://medium.com/@user/article-title-123abc",
			redirections: []Redirection{
				{Type: "domain", Find: "medium.com", Replace: "scribe.rip"},
			},
			want: "https://scribe.rip/@user/article-title-123abc",
		},
		{
			name: "multiple tracking params removal",
			url:  "https://example.com/page?id=1&utm_source=twitter&utm_medium=social&utm_campaign=test&ref=abc",
			redirections: []Redirection{
				{Type: "regex", Find: "[?&]utm_[a-z_]+=[^&]*", Replace: ""},
				{Type: "regex", Find: "[?&]ref=[^&]*", Replace: ""},
			},
			want: "https://example.com/page?id=1",
		},
		{
			name: "chained redirections - domain then cleanup",
			url:  "https://twitter.com/user?utm_source=share&s=20",
			redirections: []Redirection{
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
				{Type: "wildcard", Find: "?utm_source=*", Replace: ""},
				{Type: "wildcard", Find: "&s=*", Replace: ""},
			},
			want: "https://nitter.net/user",
		},
		{
			name: "tiktok tracking removal",
			url:  "https://www.tiktok.com/@user/video/123456?is_from_webapp=1&sender_device=pc",
			redirections: []Redirection{
				{Type: "regex", Find: "\\?.*", Replace: ""},
			},
			want: "https://www.tiktok.com/@user/video/123456",
		},
		{
			name: "amp link cleanup",
			url:  "https://www.google.com/amp/s/example.com/article",
			redirections: []Redirection{
				{Type: "regex", Find: "google\\.com/amp/s/", Replace: ""},
			},
			want: "https://www.example.com/article",
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
