// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"fmt"
	"regexp"
	"strings"
)

func validateDomainPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("Domain cannot be empty")
	}

	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		return fmt.Errorf("Wildcards not allowed in domain patterns (use Wildcard type instead)")
	}

	if strings.Contains(pattern, " ") {
		return fmt.Errorf("Domain cannot contain spaces")
	}

	if strings.HasPrefix(pattern, ".") || strings.HasSuffix(pattern, ".") {
		return fmt.Errorf("Domain cannot start or end with a dot")
	}
	if strings.HasPrefix(pattern, "-") || strings.HasSuffix(pattern, "-") {
		return fmt.Errorf("Domain cannot start or end with a hyphen")
	}

	for _, character := range pattern {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '-' || character == '_') {
			return fmt.Errorf("Domain contains invalid character: %c", character)
		}
	}

	return nil
}

func validateGlobPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("Wildcard pattern cannot be empty")
	}

	if strings.Contains(pattern, " ") {
		return fmt.Errorf("Wildcard pattern cannot contain spaces")
	}

	patternWithoutWildcards := strings.ReplaceAll(pattern, "*", "X")

	for _, character := range patternWithoutWildcards {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '-' || character == '_') {
			return fmt.Errorf("Wildcard pattern contains invalid character: %c", character)
		}
	}

	if strings.HasPrefix(pattern, ".") && !strings.HasPrefix(pattern, ".*") {
		return fmt.Errorf("Wildcard pattern cannot start with a dot")
	}
	if strings.HasSuffix(pattern, ".") && !strings.HasSuffix(pattern, "*.") {
		return fmt.Errorf("Wildcard pattern cannot end with a dot")
	}

	return nil
}

func ValidateConditionPattern(conditionType, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("Pattern cannot be empty")
	}

	switch conditionType {
	case "domain":
		return validateDomainPattern(pattern)
	case "glob":
		return validateGlobPattern(pattern)
	case "regex":
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("Invalid regex: %w", err)
		}
	case "keyword":
		return nil
	default:
		return fmt.Errorf("Invalid condition type: %s", conditionType)
	}

	return nil
}

func isConditionValid(condition Condition) bool {
	return ValidateConditionPattern(condition.Type, condition.Pattern) == nil
}

func AreAllConditionsValid(conditions []Condition) bool {
	if len(conditions) == 0 {
		return false
	}
	for _, condition := range conditions {
		if !isConditionValid(condition) {
			return false
		}
	}
	return true
}

func ValidateRedirection(redirection Redirection) error {
	if redirection.Find == "" {
		return fmt.Errorf("Find pattern cannot be empty")
	}

	redirectionType := redirection.Type
	if redirectionType == "" {
		redirectionType = "domain"
	}

	switch redirectionType {
	case "domain":
		return validateDomainPattern(redirection.Find)
	case "wildcard":
		pattern := wildcardToRegex(redirection.Find)
		if _, err := regexp.Compile("(?i)" + pattern); err != nil {
			return fmt.Errorf("Invalid pattern: %w", err)
		}
	case "regex":
		if _, err := regexp.Compile(redirection.Find); err != nil {
			return fmt.Errorf("Invalid regex: %w", err)
		}
	default:
		return fmt.Errorf("Invalid redirection type: %s", redirectionType)
	}
	return nil
}

func isRedirectionValid(redirection Redirection) bool {
	return ValidateRedirection(redirection) == nil
}

func AreAllRedirectionsValid(redirections []Redirection) bool {
	for _, redirection := range redirections {
		if !isRedirectionValid(redirection) {
			return false
		}
	}
	return true
}
