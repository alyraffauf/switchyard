// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"regexp"
	"strings"
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
		return fmt.Errorf("Domain cannot be empty")
	}

	// Check for wildcard characters (not allowed in domain type)
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		return fmt.Errorf("Wildcards not allowed in domain patterns (use Wildcard type instead)")
	}

	// Check for spaces
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("Domain cannot contain spaces")
	}

	// Check for invalid starting/ending characters
	if strings.HasPrefix(pattern, ".") || strings.HasSuffix(pattern, ".") {
		return fmt.Errorf("Domain cannot start or end with a dot")
	}
	if strings.HasPrefix(pattern, "-") || strings.HasSuffix(pattern, "-") {
		return fmt.Errorf("Domain cannot start or end with a hyphen")
	}

	// Validate characters: only alphanumeric, dots, hyphens, underscores
	for _, ch := range pattern {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("Domain contains invalid character: %c", ch)
		}
	}

	return nil
}

// validateGlobPattern checks if a glob (wildcard) pattern is valid.
// Glob patterns can contain * wildcards for matching multiple characters.
func validateGlobPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("Wildcard pattern cannot be empty")
	}

	// Check for spaces
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("Wildcard pattern cannot contain spaces")
	}

	// Remove wildcards temporarily for validation
	temp := strings.ReplaceAll(pattern, "*", "X")

	// Check if remaining pattern (without wildcards) has valid characters
	for _, ch := range temp {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("Wildcard pattern contains invalid character: %c", ch)
		}
	}

	// Check for problematic dot placement (unless adjacent to wildcard)
	if strings.HasPrefix(pattern, ".") && !strings.HasPrefix(pattern, ".*") {
		return fmt.Errorf("Wildcard pattern cannot start with a dot")
	}
	if strings.HasSuffix(pattern, ".") && !strings.HasSuffix(pattern, "*.") {
		return fmt.Errorf("Wildcard pattern cannot end with a dot")
	}

	return nil
}

// validateConditionPattern checks if a condition's pattern is valid for its type.
// Returns an error with a descriptive message if invalid, nil if valid.
func validateConditionPattern(condType, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("Pattern cannot be empty")
	}

	switch condType {
	case "domain":
		return validateDomainPattern(pattern)
	case "glob":
		return validateGlobPattern(pattern)
	case "regex":
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("Invalid regex: %w", err)
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
