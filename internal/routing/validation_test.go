// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"testing"
)

func TestValidateDomainPattern(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		wantErr      bool
		errorMessage string
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
			name:         "empty domain",
			pattern:      "",
			wantErr:      true,
			errorMessage: "Domain cannot be empty",
		},
		{
			name:         "domain with wildcard asterisk",
			pattern:      "*.example.com",
			wantErr:      true,
			errorMessage: "Wildcards not allowed in domain patterns (use Wildcard type instead)",
		},
		{
			name:         "domain with wildcard question mark",
			pattern:      "example?.com",
			wantErr:      true,
			errorMessage: "Wildcards not allowed in domain patterns (use Wildcard type instead)",
		},
		{
			name:         "domain with space",
			pattern:      "example .com",
			wantErr:      true,
			errorMessage: "Domain cannot contain spaces",
		},
		{
			name:         "domain starting with dot",
			pattern:      ".example.com",
			wantErr:      true,
			errorMessage: "Domain cannot start or end with a dot",
		},
		{
			name:         "domain ending with dot",
			pattern:      "example.com.",
			wantErr:      true,
			errorMessage: "Domain cannot start or end with a dot",
		},
		{
			name:         "domain starting with hyphen",
			pattern:      "-example.com",
			wantErr:      true,
			errorMessage: "Domain cannot start or end with a hyphen",
		},
		{
			name:         "domain ending with hyphen",
			pattern:      "example.com-",
			wantErr:      true,
			errorMessage: "Domain cannot start or end with a hyphen",
		},
		{
			name:         "domain with slash",
			pattern:      "example.com/path",
			wantErr:      true,
			errorMessage: "Domain contains invalid character: /",
		},
		{
			name:         "domain with special character",
			pattern:      "example@test.com",
			wantErr:      true,
			errorMessage: "Domain contains invalid character: @",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomainPattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDomainPattern(%q) expected error but got nil", tt.pattern)
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("validateDomainPattern(%q) error = %q, want %q", tt.pattern, err.Error(), tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("validateDomainPattern(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

func TestValidateGlobPattern(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		wantErr      bool
		errorMessage string
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
			name:         "empty pattern",
			pattern:      "",
			wantErr:      true,
			errorMessage: "Wildcard pattern cannot be empty",
		},
		{
			name:         "pattern with space",
			pattern:      "* .example.com",
			wantErr:      true,
			errorMessage: "Wildcard pattern cannot contain spaces",
		},
		{
			name:         "pattern starting with dot (not wildcard)",
			pattern:      ".example.com",
			wantErr:      true,
			errorMessage: "Wildcard pattern cannot start with a dot",
		},
		{
			name:         "pattern ending with dot (not wildcard)",
			pattern:      "example.com.",
			wantErr:      true,
			errorMessage: "Wildcard pattern cannot end with a dot",
		},
		{
			name:         "pattern with special character",
			pattern:      "example@*.com",
			wantErr:      true,
			errorMessage: "Wildcard pattern contains invalid character: @",
		},
		{
			name:         "pattern with slash",
			pattern:      "example.com/*",
			wantErr:      true,
			errorMessage: "Wildcard pattern contains invalid character: /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGlobPattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateGlobPattern(%q) expected error but got nil", tt.pattern)
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("validateGlobPattern(%q) error = %q, want %q", tt.pattern, err.Error(), tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("validateGlobPattern(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

func TestValidateConditionPattern(t *testing.T) {
	tests := []struct {
		name          string
		conditionType string
		pattern       string
		wantErr       bool
	}{
		{
			name:          "valid domain type",
			conditionType: "domain",
			pattern:       "example.com",
			wantErr:       false,
		},
		{
			name:          "invalid domain type - wildcard",
			conditionType: "domain",
			pattern:       "*.example.com",
			wantErr:       true,
		},
		{
			name:          "valid keyword type",
			conditionType: "keyword",
			pattern:       "github",
			wantErr:       false,
		},
		{
			name:          "valid keyword with special chars",
			conditionType: "keyword",
			pattern:       "/api/v2/",
			wantErr:       false,
		},
		{
			name:          "empty keyword",
			conditionType: "keyword",
			pattern:       "",
			wantErr:       true,
		},
		{
			name:          "valid glob type",
			conditionType: "glob",
			pattern:       "*.example.com",
			wantErr:       false,
		},
		{
			name:          "invalid glob type - spaces",
			conditionType: "glob",
			pattern:       "* .example.com",
			wantErr:       true,
		},
		{
			name:          "valid regex type",
			conditionType: "regex",
			pattern:       "^https://.*\\.example\\.com",
			wantErr:       false,
		},
		{
			name:          "invalid regex type - bad syntax",
			conditionType: "regex",
			pattern:       "[invalid",
			wantErr:       true,
		},
		{
			name:          "empty regex",
			conditionType: "regex",
			pattern:       "",
			wantErr:       true,
		},
		{
			name:          "empty pattern for domain",
			conditionType: "domain",
			pattern:       "",
			wantErr:       true,
		},
		{
			name:          "empty pattern for glob",
			conditionType: "glob",
			pattern:       "",
			wantErr:       true,
		},
		{
			name:          "invalid condition type",
			conditionType: "invalid_type",
			pattern:       "example.com",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConditionPattern(tt.conditionType, tt.pattern)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateConditionPattern(%q, %q) expected error but got nil", tt.conditionType, tt.pattern)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateConditionPattern(%q, %q) unexpected error: %v", tt.conditionType, tt.pattern, err)
			}
		})
	}
}

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
		{
			name: "invalid condition type",
			condition: Condition{
				Type:    "invalid_type",
				Pattern: "example.com",
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
		{
			name: "invalid condition type in list",
			conditions: []Condition{
				{Type: "domain", Pattern: "example.com"},
				{Type: "invalid_type", Pattern: "example.com"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AreAllConditionsValid(tt.conditions)
			if got != tt.want {
				t.Errorf("AreAllConditionsValid(%v) = %v, want %v", tt.conditions, got, tt.want)
			}
		})
	}
}
