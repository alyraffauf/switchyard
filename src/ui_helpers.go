// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

// validateConditions checks if all conditions have non-empty patterns and valid types
func validateConditions(conditions []Condition) bool {
	validTypes := map[string]bool{
		"domain":  true,
		"keyword": true,
		"glob":    true,
		"regex":   true,
	}

	for _, c := range conditions {
		if c.Pattern == "" {
			return false
		}
		if !validTypes[c.Type] {
			return false
		}
		// For regex type, validate the pattern compiles
		if c.Type == "regex" {
			if _, err := regexp.Compile(c.Pattern); err != nil {
				return false
			}
		}
	}
	return true
}

// validateDomainPattern checks if a domain pattern is valid (no wildcards allowed).
// Domain patterns should be exact hostnames like "example.com" or "api.github.com".
func validateDomainPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check for wildcard characters (not allowed in domain type)
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		return fmt.Errorf("wildcards not allowed in domain patterns (use Wildcard type instead)")
	}

	// Check for spaces
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("domain cannot contain spaces")
	}

	// Check for invalid starting/ending characters
	if strings.HasPrefix(pattern, ".") || strings.HasSuffix(pattern, ".") {
		return fmt.Errorf("domain cannot start or end with a dot")
	}
	if strings.HasPrefix(pattern, "-") || strings.HasSuffix(pattern, "-") {
		return fmt.Errorf("domain cannot start or end with a hyphen")
	}

	// Validate characters: only alphanumeric, dots, hyphens, underscores
	for _, ch := range pattern {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("domain contains invalid character: %c", ch)
		}
	}

	return nil
}

// validateGlobPattern checks if a glob (wildcard) pattern is valid.
// Glob patterns can contain * wildcards for matching multiple characters.
func validateGlobPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("wildcard pattern cannot be empty")
	}

	// Check for spaces
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("wildcard pattern cannot contain spaces")
	}

	// Remove wildcards temporarily for validation
	temp := strings.ReplaceAll(pattern, "*", "X")

	// Check if remaining pattern (without wildcards) has valid characters
	for _, ch := range temp {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("wildcard pattern contains invalid character: %c", ch)
		}
	}

	// Check for problematic dot placement (unless adjacent to wildcard)
	if strings.HasPrefix(pattern, ".") && !strings.HasPrefix(pattern, ".*") {
		return fmt.Errorf("wildcard pattern cannot start with a dot")
	}
	if strings.HasSuffix(pattern, ".") && !strings.HasSuffix(pattern, "*.") {
		return fmt.Errorf("wildcard pattern cannot end with a dot")
	}

	return nil
}

// validateConditionPattern checks if a condition's pattern is valid for its type.
// Returns an error with a descriptive message if invalid, nil if valid.
func validateConditionPattern(condType, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	switch condType {
	case "domain":
		return validateDomainPattern(pattern)
	case "glob":
		return validateGlobPattern(pattern)
	case "regex":
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	case "keyword":
		// Keywords can be any non-empty string
		return nil
	}

	return nil
}

// isConditionValid checks if a single condition is valid.
func isConditionValid(c Condition) bool {
	return validateConditionPattern(c.Type, c.Pattern) == nil
}

// areAllConditionsValid checks if all conditions in a slice are valid.
func areAllConditionsValid(conditions []Condition) bool {
	if len(conditions) == 0 {
		return false
	}
	for _, c := range conditions {
		if !isConditionValid(c) {
			return false
		}
	}
	return true
}

// getLogicFromComboRow extracts the logic string from a combo row selection
func getLogicFromComboRow(logicRow *adw.ComboRow) string {
	if logicRow.Selected() == 1 {
		return "any"
	}
	return "all"
}

// saveConfigWithFlag saves config while setting the global saving flag to prevent file watcher loops
func saveConfigWithFlag(cfg *Config) {
	savingMux.Lock()
	isSaving = true
	savingMux.Unlock()
	saveConfig(cfg)
	glib.TimeoutAdd(100, func() bool {
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
		return false
	})
}

// findBrowserByID finds a browser by its desktop file ID
func findBrowserByID(browsers []*Browser, id string) *Browser {
	for _, b := range browsers {
		if b.ID == id {
			return b
		}
	}
	return nil
}
