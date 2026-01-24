// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"regexp"
	"strings"
)

var regexCache = make(map[string]*regexp.Regexp)

func getCompiledRegex(pattern string) (*regexp.Regexp, bool) {
	if re, ok := regexCache[pattern]; ok {
		return re, true
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, false
	}
	regexCache[pattern] = re
	return re, true
}

func matchesPattern(url, pattern, patternType string) bool {
	switch patternType {
	case "domain":
		domain := extractDomain(url)
		return strings.EqualFold(domain, pattern)
	case "keyword":
		return strings.Contains(strings.ToLower(url), strings.ToLower(pattern))
	case "regex":
		re, ok := getCompiledRegex(pattern)
		if !ok {
			return false
		}
		return re.MatchString(url)
	case "glob":
		return matchGlob(url, pattern)
	default:
		return false
	}
}

// matchGlob performs glob-style pattern matching against a URL's domain or full URL
func matchGlob(url, pattern string) bool {
	domain := extractDomain(url)

	// Convert glob to regex: escape dots, convert * to .*
	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(pattern, ".", "\\."), "*", ".*") + "$"

	re, ok := getCompiledRegex(regexPattern)
	if !ok {
		return false
	}

	return re.MatchString(domain) || re.MatchString(url)
}
