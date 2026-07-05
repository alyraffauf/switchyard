// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"testing"
)

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
				Logic: "",
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
		{
			name: "negated condition excludes match",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "glob", Pattern: "*.github.com"},
					{Type: "domain", Pattern: "gist.github.com", Negate: true},
				},
			},
			url:  "https://gist.github.com",
			want: false,
		},
		{
			name: "negated condition allows non-match",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "glob", Pattern: "*.github.com"},
					{Type: "domain", Pattern: "gist.github.com", Negate: true},
				},
			},
			url:  "https://api.github.com",
			want: true,
		},
		{
			name: "all negated conditions",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "example.com", Negate: true},
					{Type: "keyword", Pattern: "admin", Negate: true},
				},
			},
			url:  "https://other.com/user",
			want: true,
		},
		{
			name: "negated condition fails when pattern not matched",
			rule: Rule{
				Logic: "all",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com"},
					{Type: "keyword", Pattern: "secret", Negate: true},
				},
			},
			url:  "https://github.com/user/repo",
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.rule.MatchesConditions(test.url)
			if result != test.want {
				t.Errorf("Rule.MatchesConditions(%q) = %v, want %v", test.url, result, test.want)
			}
		})
	}
}

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
		{
			name: "OR with negated condition matches",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com", Negate: true},
					{Type: "domain", Pattern: "gitlab.com"},
				},
			},
			url:  "https://example.com",
			want: true,
		},
		{
			name: "OR with all negated conditions none match",
			rule: Rule{
				Logic: "any",
				Conditions: []Condition{
					{Type: "domain", Pattern: "github.com", Negate: true},
					{Type: "domain", Pattern: "gitlab.com", Negate: true},
				},
			},
			url:  "https://github.com",
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.rule.MatchesConditions(test.url)
			if result != test.want {
				t.Errorf("Rule.MatchesConditions(%q) = %v, want %v", test.url, result, test.want)
			}
		})
	}
}
