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

func TestApplyDomainRewrite(t *testing.T) {
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
			r := Rewrite{Type: "domain", Find: tt.find, Replace: tt.replace}
			got := applyRewrite(tt.url, r)
			if got != tt.want {
				t.Errorf("applyRewrite() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyURLRewrite(t *testing.T) {
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
			r := Rewrite{Type: "url", Find: tt.find, Replace: tt.replace}
			got := applyRewrite(tt.url, r)
			if got != tt.want {
				t.Errorf("applyRewrite() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyRewriteDefaultType(t *testing.T) {
	// Empty type should default to domain
	r := Rewrite{Find: "reddit.com", Replace: "old.reddit.com"}
	got := applyRewrite("https://reddit.com/r/test", r)
	want := "https://old.reddit.com/r/test"
	if got != want {
		t.Errorf("applyRewrite() with empty type = %q, want %q", got, want)
	}

	// Should not match subdomain with default type
	got = applyRewrite("https://old.reddit.com/r/test", r)
	want = "https://old.reddit.com/r/test"
	if got != want {
		t.Errorf("applyRewrite() should not match subdomain = %q, want %q", got, want)
	}
}

func TestApplyRewrites(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		rewrites []Rewrite
		want     string
	}{
		{
			name: "domain then url rewrite",
			url:  "https://twitter.com/user?utm_source=share",
			rewrites: []Rewrite{
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
				{Type: "url", Find: "?utm_source=*", Replace: ""},
			},
			want: "https://nitter.net/user",
		},
		{
			name:     "empty rewrites list",
			url:      "https://example.com",
			rewrites: []Rewrite{},
			want:     "https://example.com",
		},
		{
			name: "chained domain rewrites",
			url:  "https://x.com/user",
			rewrites: []Rewrite{
				{Type: "domain", Find: "x.com", Replace: "twitter.com"},
				{Type: "domain", Find: "twitter.com", Replace: "nitter.net"},
			},
			want: "https://nitter.net/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyRewrites(tt.url, tt.rewrites)
			if got != tt.want {
				t.Errorf("applyRewrites() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateRewrite(t *testing.T) {
	tests := []struct {
		name    string
		rewrite Rewrite
		wantErr bool
	}{
		{
			name:    "valid domain rewrite",
			rewrite: Rewrite{Type: "domain", Find: "reddit.com", Replace: "old.reddit.com"},
			wantErr: false,
		},
		{
			name:    "valid domain rewrite default type",
			rewrite: Rewrite{Find: "twitter.com", Replace: "nitter.net"},
			wantErr: false,
		},
		{
			name:    "valid url rewrite with wildcard",
			rewrite: Rewrite{Type: "url", Find: "utm_*", Replace: ""},
			wantErr: false,
		},
		{
			name:    "domain rewrite with wildcard invalid",
			rewrite: Rewrite{Type: "domain", Find: "*.reddit.com", Replace: "old.reddit.com"},
			wantErr: true,
		},
		{
			name:    "empty find pattern",
			rewrite: Rewrite{Type: "domain", Find: "", Replace: "something"},
			wantErr: true,
		},
		{
			name:    "empty replace is valid",
			rewrite: Rewrite{Type: "url", Find: "tracking", Replace: ""},
			wantErr: false,
		},
		{
			name:    "invalid rewrite type",
			rewrite: Rewrite{Type: "invalid", Find: "test", Replace: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRewrite(tt.rewrite)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRewrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
