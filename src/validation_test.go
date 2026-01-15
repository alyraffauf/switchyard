// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"
)

// TestValidateDomainPattern tests domain pattern validation
func TestValidateDomainPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple domain",
			pattern: "example.com",
			wantErr: false,
		},
		{
			name:    "valid subdomain",
			pattern: "api.example.com",
			wantErr: false,
		},
		{
			name:    "valid multiple subdomains",
			pattern: "deep.api.example.com",
			wantErr: false,
		},
		{
			name:    "valid domain with hyphen",
			pattern: "my-site.example.com",
			wantErr: false,
		},
		{
			name:    "valid domain with numbers",
			pattern: "site123.example.com",
			wantErr: false,
		},
		{
			name:    "valid domain with underscore",
			pattern: "my_site.example.com",
			wantErr: false,
		},
		{
			name:    "empty domain",
			pattern: "",
			wantErr: true,
			errMsg:  "Domain cannot be empty",
		},
		{
			name:    "domain with wildcard asterisk",
			pattern: "*.example.com",
			wantErr: true,
			errMsg:  "Wildcards not allowed in domain patterns (use Wildcard type instead)",
		},
		{
			name:    "domain with wildcard question mark",
			pattern: "example?.com",
			wantErr: true,
			errMsg:  "Wildcards not allowed in domain patterns (use Wildcard type instead)",
		},
		{
			name:    "domain with space",
			pattern: "example .com",
			wantErr: true,
			errMsg:  "Domain cannot contain spaces",
		},
		{
			name:    "domain starting with dot",
			pattern: ".example.com",
			wantErr: true,
			errMsg:  "Domain cannot start or end with a dot",
		},
		{
			name:    "domain ending with dot",
			pattern: "example.com.",
			wantErr: true,
			errMsg:  "Domain cannot start or end with a dot",
		},
		{
			name:    "domain starting with hyphen",
			pattern: "-example.com",
			wantErr: true,
			errMsg:  "Domain cannot start or end with a hyphen",
		},
		{
			name:    "domain ending with hyphen",
			pattern: "example.com-",
			wantErr: true,
			errMsg:  "Domain cannot start or end with a hyphen",
		},
		{
			name:    "domain with slash",
			pattern: "example.com/path",
			wantErr: true,
			errMsg:  "Domain contains invalid character: /",
		},
		{
			name:    "domain with special character",
			pattern: "example@test.com",
			wantErr: true,
			errMsg:  "Domain contains invalid character: @",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomainPattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDomainPattern(%q) expected error but got nil", tt.pattern)
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateDomainPattern(%q) error = %q, want %q", tt.pattern, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateDomainPattern(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

// TestValidateGlobPattern tests wildcard pattern validation
func TestValidateGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid wildcard at start",
			pattern: "*.example.com",
			wantErr: false,
		},
		{
			name:    "valid wildcard at end",
			pattern: "example.*",
			wantErr: false,
		},
		{
			name:    "valid wildcard in middle",
			pattern: "api.*.example.com",
			wantErr: false,
		},
		{
			name:    "valid multiple wildcards",
			pattern: "*.*.example.com",
			wantErr: false,
		},
		{
			name:    "valid domain with hyphen and wildcard",
			pattern: "my-*.example.com",
			wantErr: false,
		},
		{
			name:    "valid domain with numbers and wildcard",
			pattern: "site*.example.com",
			wantErr: false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			wantErr: true,
			errMsg:  "Wildcard pattern cannot be empty",
		},
		{
			name:    "pattern with space",
			pattern: "* .example.com",
			wantErr: true,
			errMsg:  "Wildcard pattern cannot contain spaces",
		},
		{
			name:    "pattern starting with dot (not wildcard)",
			pattern: ".example.com",
			wantErr: true,
			errMsg:  "Wildcard pattern cannot start with a dot",
		},
		{
			name:    "pattern ending with dot (not wildcard)",
			pattern: "example.com.",
			wantErr: true,
			errMsg:  "Wildcard pattern cannot end with a dot",
		},
		{
			name:    "pattern with special character",
			pattern: "example@*.com",
			wantErr: true,
			errMsg:  "Wildcard pattern contains invalid character: @",
		},
		{
			name:    "pattern with slash",
			pattern: "example.com/*",
			wantErr: true,
			errMsg:  "Wildcard pattern contains invalid character: /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGlobPattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateGlobPattern(%q) expected error but got nil", tt.pattern)
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateGlobPattern(%q) error = %q, want %q", tt.pattern, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateGlobPattern(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

// TestValidateConditionPattern tests the main condition pattern validator
func TestValidateConditionPattern(t *testing.T) {
	tests := []struct {
		name     string
		condType string
		pattern  string
		wantErr  bool
	}{
		// Domain type
		{
			name:     "valid domain type",
			condType: "domain",
			pattern:  "example.com",
			wantErr:  false,
		},
		{
			name:     "invalid domain type - wildcard",
			condType: "domain",
			pattern:  "*.example.com",
			wantErr:  true,
		},
		// Keyword type
		{
			name:     "valid keyword type",
			condType: "keyword",
			pattern:  "github",
			wantErr:  false,
		},
		{
			name:     "valid keyword with special chars",
			condType: "keyword",
			pattern:  "/api/v2/",
			wantErr:  false,
		},
		{
			name:     "empty keyword",
			condType: "keyword",
			pattern:  "",
			wantErr:  true,
		},
		// Glob type
		{
			name:     "valid glob type",
			condType: "glob",
			pattern:  "*.example.com",
			wantErr:  false,
		},
		{
			name:     "invalid glob type - spaces",
			condType: "glob",
			pattern:  "* .example.com",
			wantErr:  true,
		},
		// Regex type
		{
			name:     "valid regex type",
			condType: "regex",
			pattern:  "^https://.*\\.example\\.com",
			wantErr:  false,
		},
		{
			name:     "invalid regex type - bad syntax",
			condType: "regex",
			pattern:  "[invalid",
			wantErr:  true,
		},
		{
			name:     "empty regex",
			condType: "regex",
			pattern:  "",
			wantErr:  true,
		},
		// Empty pattern universal check
		{
			name:     "empty pattern for domain",
			condType: "domain",
			pattern:  "",
			wantErr:  true,
		},
		{
			name:     "empty pattern for glob",
			condType: "glob",
			pattern:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConditionPattern(tt.condType, tt.pattern)
			if tt.wantErr && err == nil {
				t.Errorf("validateConditionPattern(%q, %q) expected error but got nil", tt.condType, tt.pattern)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateConditionPattern(%q, %q) unexpected error: %v", tt.condType, tt.pattern, err)
			}
		})
	}
}

// TestIsConditionValid tests single condition validation
func TestIsConditionValid(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		want      bool
	}{
		{
			name: "valid domain condition",
			condition: Condition{
				Type:    "domain",
				Pattern: "example.com",
			},
			want: true,
		},
		{
			name: "valid keyword condition",
			condition: Condition{
				Type:    "keyword",
				Pattern: "github",
			},
			want: true,
		},
		{
			name: "valid glob condition",
			condition: Condition{
				Type:    "glob",
				Pattern: "*.example.com",
			},
			want: true,
		},
		{
			name: "valid regex condition",
			condition: Condition{
				Type:    "regex",
				Pattern: "^https://.*",
			},
			want: true,
		},
		{
			name: "invalid domain condition - wildcard",
			condition: Condition{
				Type:    "domain",
				Pattern: "*.example.com",
			},
			want: false,
		},
		{
			name: "invalid condition - empty pattern",
			condition: Condition{
				Type:    "domain",
				Pattern: "",
			},
			want: false,
		},
		{
			name: "invalid regex condition",
			condition: Condition{
				Type:    "regex",
				Pattern: "[invalid",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConditionValid(tt.condition)
			if got != tt.want {
				t.Errorf("isConditionValid(%v) = %v, want %v", tt.condition, got, tt.want)
			}
		})
	}
}

// TestAreAllConditionsValid tests validation of condition slices
func TestAreAllConditionsValid(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		want       bool
	}{
		{
			name: "all valid conditions",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "keyword", Pattern: "github"},
			},
			want: true,
		},
		{
			name: "one invalid condition",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "domain", Pattern: "*.invalid.com"},
			},
			want: false,
		},
		{
			name: "empty pattern in list",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "keyword", Pattern: ""},
			},
			want: false,
		},
		{
			name:       "empty conditions list",
			conditions: []Condition{},
			want:       false,
		},
		{
			name:       "nil conditions list",
			conditions: nil,
			want:       false,
		},
		{
			name: "single valid condition",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
			},
			want: true,
		},
		{
			name: "invalid regex in list",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "regex", Pattern: "[invalid"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := areAllConditionsValid(tt.conditions)
			if got != tt.want {
				t.Errorf("areAllConditionsValid(%v) = %v, want %v", tt.conditions, got, tt.want)
			}
		})
	}
}

// TestValidateConditions tests the main validateConditions function
func TestValidateConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		want       bool
	}{
		{
			name: "valid conditions with all types",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "keyword", Pattern: "test"},
				{Type: "glob", Pattern: "*.example.com"},
				{Type: "regex", Pattern: "^https://.*"},
			},
			want: true,
		},
		{
			name: "invalid type",
			conditions: []Condition{
				{Type: "invalid_type", Pattern: "example.com"},
			},
			want: false,
		},
		{
			name: "empty pattern",
			conditions: []Condition{
				{Type: "domain", Pattern: ""},
			},
			want: false,
		},
		{
			name: "invalid regex pattern",
			conditions: []Condition{
				{Type: "regex", Pattern: "[unclosed"},
			},
			want: false,
		},
		{
			name:       "empty list",
			conditions: []Condition{},
			want:       true,
		},
		{
			name: "mixed valid and invalid",
			conditions: []Condition{
				{Type: "domain", Pattern: "valid.com"},
				{Type: "domain", Pattern: ""},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateConditions(tt.conditions)
			if got != tt.want {
				t.Errorf("validateConditions(%v) = %v, want %v", tt.conditions, got, tt.want)
			}
		})
	}
}
