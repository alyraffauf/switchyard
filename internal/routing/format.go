// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import "fmt"

// --- Condition types: single source of truth ---

// conditionTypeDef is the single source of truth for a condition type:
// its combo-box label, and the labels used in rule subtitles.
type conditionTypeDef struct {
	Type        string
	Label       string // combo-box dropdown label
	MatchLabel  string // e.g. "Domain is"
	NegateLabel string // e.g. "Domain is not"
}

// conditionTypes is the ordered list of supported condition types.
// Index order is the combo-box order; adding a type here is sufficient
// for all mappers, dropdown models, and condition labels.
var conditionTypes = []conditionTypeDef{
	{Type: "domain", Label: "Exact Domain", MatchLabel: "Domain is", NegateLabel: "Domain is not"},
	{Type: "keyword", Label: "URL Contains", MatchLabel: "URL contains", NegateLabel: "URL does not contain"},
	{Type: "glob", Label: "Wildcard", MatchLabel: "Wildcard is", NegateLabel: "Wildcard is not"},
	{Type: "regex", Label: "Regex", MatchLabel: "Regex matches", NegateLabel: "Regex does not match"},
}

// ConditionTypeToIndex maps a condition type string to its combo-box index.
func ConditionTypeToIndex(conditionType string) uint {
	for i, def := range conditionTypes {
		if def.Type == conditionType {
			return uint(i)
		}
	}
	return 0
}

// IndexToConditionType maps a combo-box index to its condition type string.
func IndexToConditionType(index uint) string {
	if int(index) < len(conditionTypes) {
		return conditionTypes[index].Type
	}
	return conditionTypes[0].Type
}

// ConditionTypeLabels returns the human-readable labels for condition types.
func ConditionTypeLabels() []string {
	labels := make([]string, len(conditionTypes))
	for i, def := range conditionTypes {
		labels[i] = def.Label
	}
	return labels
}

// --- Redirection types ---

// redirectionTypeDef describes a redirection type for UI dropdowns and mappers.
type redirectionTypeDef struct {
	Type  string
	Label string
}

// redirectionTypes is the ordered list of supported redirection types.
var redirectionTypes = []redirectionTypeDef{
	{Type: "domain", Label: "Domain"},
	{Type: "wildcard", Label: "Wildcard"},
	{Type: "regex", Label: "Regex"},
}

// RedirectionTypeToIndex maps a redirection type string to its combo-box index.
func RedirectionTypeToIndex(redirectionType string) uint {
	for i, def := range redirectionTypes {
		if def.Type == redirectionType {
			return uint(i)
		}
	}
	return 0
}

// IndexToRedirectionType maps a combo-box index to its redirection type string.
func IndexToRedirectionType(index uint) string {
	if int(index) < len(redirectionTypes) {
		return redirectionTypes[index].Type
	}
	return redirectionTypes[0].Type
}

// RedirectionTypeLabels returns the human-readable labels for redirection types.
func RedirectionTypeLabels() []string {
	labels := make([]string, len(redirectionTypes))
	for i, def := range redirectionTypes {
		labels[i] = def.Label
	}
	return labels
}

// NormalizeRedirectionType returns the canonical redirection type string,
// defaulting empty to "domain". Unknown non-empty types are returned as-is
// so callers can detect and handle them.
func NormalizeRedirectionType(typ string) string {
	if typ == "" {
		return "domain"
	}
	return typ
}

// RedirectionTypeLabel returns the human-readable label for a redirection type.
// Falls back to the raw type string for unknown types.
func RedirectionTypeLabel(typ string) string {
	typ = NormalizeRedirectionType(typ)
	for _, def := range redirectionTypes {
		if def.Type == typ {
			return def.Label
		}
	}
	return typ
}

// --- Rule / condition formatting (plain text, no markup) ---

// FormatRuleSubtitle returns a plain-text subtitle for a rule row with the
// condition pattern included. Callers must escape the result before passing
// it to a widget that parses markup.
func FormatRuleSubtitle(rule *Rule, browserName string) string {
	return formatRuleSubtitle(rule, browserName, true)
}

// FormatRuleSubtitleNoPattern is like FormatRuleSubtitle but omits the
// condition pattern from the result.
func FormatRuleSubtitleNoPattern(rule *Rule, browserName string) string {
	return formatRuleSubtitle(rule, browserName, false)
}

func formatRuleSubtitle(rule *Rule, browserName string, includePattern bool) string {
	if len(rule.Conditions) == 0 {
		return "No conditions"
	}

	conditionCount := len(rule.Conditions)
	var logicText string
	if rule.Logic == "any" {
		logicText = "Any match"
	} else {
		logicText = "All match"
	}

	formatSingleCondition := func(condition *Condition) string {
		return fmt.Sprintf("%s %s", ConditionLabel(condition.Type, condition.Negate), condition.Pattern)
	}

	if rule.AlwaysAsk {
		if conditionCount == 1 && includePattern {
			return fmt.Sprintf("%s · Always ask", formatSingleCondition(&rule.Conditions[0]))
		}
		return fmt.Sprintf("%d conditions (%s) · Always ask", conditionCount, logicText)
	}
	if conditionCount == 1 && includePattern {
		return fmt.Sprintf("%s · Opens in %s", formatSingleCondition(&rule.Conditions[0]), browserName)
	}
	return fmt.Sprintf("%d conditions (%s) · Opens in %s", conditionCount, logicText, browserName)
}

// ConditionLabel returns a human-readable label for a condition type,
// derived from the conditionTypes table.
func ConditionLabel(patternType string, negate bool) string {
	for _, def := range conditionTypes {
		if def.Type == patternType {
			if negate {
				return def.NegateLabel
			}
			return def.MatchLabel
		}
	}
	return patternType
}

// FormatRedirectionSubtitle returns a plain-text subtitle describing a
// redirection's type and action. Callers must escape the result before
// passing it to a widget that parses markup.
func FormatRedirectionSubtitle(redirection *Redirection) string {
	typeLabel := RedirectionTypeLabel(redirection.Type)

	if redirection.Replace == "" {
		return fmt.Sprintf("%s · Removes match", typeLabel)
	}
	return fmt.Sprintf("%s · Replaces with %s", typeLabel, redirection.Replace)
}
