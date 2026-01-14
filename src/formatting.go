// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "fmt"

// formatRuleSubtitle formats a subtitle for a rule row with pattern included
func formatRuleSubtitle(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, true)
}

// formatRuleSubtitleNoPattern formats a subtitle for a rule row without pattern
func formatRuleSubtitleNoPattern(rule *Rule, browserName string) string {
	return formatRuleSubtitleInternal(rule, browserName, false)
}

// formatRuleSubtitleInternal is the internal implementation for formatting rule subtitles
func formatRuleSubtitleInternal(rule *Rule, browserName string, includePattern bool) string {
	// Handle multi-condition format
	if len(rule.Conditions) > 0 {
		condCount := len(rule.Conditions)
		var logicText string
		if rule.Logic == "any" {
			logicText = "Any match"
		} else {
			logicText = "All match"
		}

		if rule.AlwaysAsk {
			if condCount == 1 && includePattern {
				return fmt.Sprintf("%s: %s · Always ask", getTypeLabel(rule.Conditions[0].Type), rule.Conditions[0].Pattern)
			}
			return fmt.Sprintf("%d conditions (%s) · Always ask", condCount, logicText)
		}
		if condCount == 1 && includePattern {
			return fmt.Sprintf("%s: %s · Opens in %s", getTypeLabel(rule.Conditions[0].Type), rule.Conditions[0].Pattern, browserName)
		}
		return fmt.Sprintf("%d conditions (%s) · Opens in %s", condCount, logicText, browserName)
	}

	// No conditions (should not happen in normal use)
	return "No conditions"
}

// getTypeLabel returns a human-readable label for a condition type
func getTypeLabel(patternType string) string {
	switch patternType {
	case "domain":
		return "Exact domain"
	case "keyword":
		return "URL contains"
	case "glob":
		return "Wildcard"
	case "regex":
		return "Regex"
	default:
		return patternType
	}
}
