// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"bufio"
	_ "embed"
	"io"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Bundled snapshot of AdGuard's tracking-parameter filter list.
//
//go:embed embedded/adguard_filter.txt
var bundledAdGuardFilter string

var (
	trackingParameterRules []adGuardRule
	loadTrackingRulesOnce  sync.Once
)

type adGuardRule struct {
	isException    bool
	urlRegex       *regexp.Regexp
	parameterRegex *regexp.Regexp
	parameterName  string
}

func getTrackingParameterRules() []adGuardRule {
	loadTrackingRulesOnce.Do(func() {
		trackingParameterRules = parseRules(strings.NewReader(bundledAdGuardFilter))
	})
	return trackingParameterRules
}

func parseRules(reader io.Reader) []adGuardRule {
	var rules []adGuardRule

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if rule, ok := parseRule(scanner.Text()); ok {
			rules = append(rules, rule)
		}
	}

	return rules
}

func parseRule(line string) (adGuardRule, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "!") {
		return adGuardRule{}, false
	}

	rule := adGuardRule{}
	if rest, ok := strings.CutPrefix(line, "@@"); ok {
		rule.isException = true
		line = rest
	}

	urlPattern, options, hasOptions := strings.Cut(line, "$")

	if urlPattern != "" {
		pattern := regexp.QuoteMeta(urlPattern)
		pattern = strings.ReplaceAll(pattern, `\|\|`, `(^|\.)`)
		pattern = strings.ReplaceAll(pattern, `\^`, `([^a-zA-Z0-9.\-%]|$)`)
		pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
		if re, err := regexp.Compile("(?i)" + pattern); err == nil {
			rule.urlRegex = re
		}
	}

	if hasOptions {
		for opt := range strings.SplitSeq(options, ",") {
			if paramValue, ok := strings.CutPrefix(opt, "removeparam="); ok {
				if strings.HasPrefix(paramValue, "/") && strings.HasSuffix(paramValue, "/") {
					if re, err := regexp.Compile("(?i)" + paramValue[1:len(paramValue)-1]); err == nil {
						rule.parameterRegex = re
					}
				} else {
					rule.parameterName = paramValue
				}
				return rule, true
			} else if opt == "removeparam" {
				rule.parameterName = "*"
				return rule, true
			}
		}
	}
	return adGuardRule{}, false
}

func RemoveTrackingParameters(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.RawQuery == "" {
		return rawURL
	}

	return removeTrackingParametersWithRules(rawURL, u, getTrackingParameterRules())
}

func removeTrackingParametersWithRules(rawURL string, u *url.URL, rules []adGuardRule) string {
	// Filter the raw fields so the surviving values keep their order and encoding.
	fields := strings.Split(u.RawQuery, "&")
	fieldCount := len(fields)
	fields = slices.DeleteFunc(fields, func(field string) bool {
		if field == "" || strings.Contains(field, ";") {
			return false
		}
		key, value, _ := strings.Cut(field, "=")
		parameter, err := url.QueryUnescape(key)
		if err != nil {
			return false
		}
		if _, err := url.QueryUnescape(value); err != nil {
			return false
		}
		return shouldRemoveTrackingParameter(parameter, rawURL, rules)
	})
	if len(fields) == fieldCount {
		return rawURL
	}
	u.RawQuery = strings.Join(fields, "&")
	return u.String()
}

func shouldRemoveTrackingParameter(parameter, rawURL string, rules []adGuardRule) bool {
	shouldRemove := false
	for _, rule := range rules {
		if rule.urlRegex != nil && !rule.urlRegex.MatchString(rawURL) {
			continue
		}
		matchesParameter := rule.parameterName == "*" || rule.parameterName == parameter ||
			rule.parameterRegex != nil && rule.parameterRegex.MatchString(parameter)
		if matchesParameter {
			if rule.isException {
				return false
			}
			shouldRemove = true
		}
	}
	return shouldRemove
}
