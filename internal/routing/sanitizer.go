// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"bufio"
	_ "embed"
	"io"
	"net/url"
	"regexp"
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
	if strings.HasPrefix(line, "@@") {
		rule.isException = true
		line = line[2:]
	}

	parts := strings.Split(line, "$")
	urlPattern := parts[0]

	if urlPattern != "" {
		pattern := regexp.QuoteMeta(urlPattern)
		pattern = strings.ReplaceAll(pattern, `\|\|`, `(^|\.)`)
		pattern = strings.ReplaceAll(pattern, `\^`, `([^a-zA-Z0-9.\-%]|$)`)
		pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
		if re, err := regexp.Compile("(?i)" + pattern); err == nil {
			rule.urlRegex = re
		}
	}

	if len(parts) > 1 {
		options := strings.Split(parts[1], ",")
		for _, opt := range options {
			if strings.HasPrefix(opt, "removeparam=") {
				paramValue := strings.TrimPrefix(opt, "removeparam=")
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
	query := u.Query()
	changed := false

	for parameter := range query {
		shouldRemove := false
		isWhitelisted := false

		for _, rule := range rules {
			if rule.urlRegex != nil && !rule.urlRegex.MatchString(rawURL) {
				continue
			}

			matchesParameter := false
			if rule.parameterName == "*" || rule.parameterName == parameter {
				matchesParameter = true
			} else if rule.parameterRegex != nil && rule.parameterRegex.MatchString(parameter) {
				matchesParameter = true
			}

			if matchesParameter {
				if rule.isException {
					isWhitelisted = true
					break
				}
				shouldRemove = true
			}
		}

		if shouldRemove && !isWhitelisted {
			query.Del(parameter)
			changed = true
		}
	}

	if !changed {
		return rawURL
	}

	u.RawQuery = query.Encode()
	return u.String()
}
