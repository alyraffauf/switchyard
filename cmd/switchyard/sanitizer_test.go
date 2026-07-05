// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestRemoveTrackingParametersWithRules(t *testing.T) {
	rules := parseRules(strings.NewReader(strings.Join([]string{
		"$removeparam=utm_source",
		"$removeparam=/^utm_/",
	}, "\n")))

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "removes named tracking parameter",
			url:  "https://example.com/article?id=1&utm_source=newsletter",
			want: "https://example.com/article?id=1",
		},
		{
			name: "removes regex-matched tracking parameter",
			url:  "https://example.com/article?id=1&utm_campaign=spring",
			want: "https://example.com/article?id=1",
		},
		{
			name: "keeps untouched url when no rule matches",
			url:  "https://example.com/article?id=1&ref=homepage",
			want: "https://example.com/article?id=1&ref=homepage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}

			got := removeTrackingParametersWithRules(tt.url, parsed, rules)
			if got != tt.want {
				t.Fatalf("removeTrackingParametersWithRules() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemoveTrackingParametersUsesBundledRules(t *testing.T) {
	got := removeTrackingParameters("https://example.com/article?id=1&utm_source=newsletter")
	want := "https://example.com/article?id=1"

	if got != want {
		t.Fatalf("removeTrackingParameters() = %q, want %q", got, want)
	}
}
