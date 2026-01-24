// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"regexp"
	"strings"
)

func applyRedirections(rawURL string, redirections []Redirection) string {
	for _, r := range redirections {
		rawURL = applyRedirection(rawURL, r)
	}
	return rawURL
}

func applyRedirection(rawURL string, r Redirection) string {
	rwType := r.Type
	if rwType == "" {
		rwType = "domain"
	}

	switch rwType {
	case "domain":
		return applyDomainRedirection(rawURL, r)
	case "pattern":
		return applyPatternRedirection(rawURL, r)
	case "regex":
		return applyRegexRedirection(rawURL, r)
	default:
		return rawURL
	}
}

func applyDomainRedirection(rawURL string, r Redirection) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if !strings.EqualFold(u.Hostname(), r.Find) {
		return rawURL
	}

	u.Host = strings.Replace(u.Host, u.Hostname(), r.Replace, 1)
	return u.String()
}

func applyPatternRedirection(rawURL string, r Redirection) string {
	pattern := wildcardToRegex(r.Find)
	re, ok := getCompiledRegex("(?i)" + pattern) // case-insensitive
	if !ok {
		return rawURL
	}
	return re.ReplaceAllString(rawURL, r.Replace)
}

func applyRegexRedirection(rawURL string, r Redirection) string {
	re, ok := getCompiledRegex(r.Find)
	if !ok {
		return rawURL
	}
	return re.ReplaceAllString(rawURL, r.Replace)
}

func wildcardToRegex(pattern string) string {
	// Escape regex special chars except *
	escaped := regexp.QuoteMeta(pattern)
	// Convert \* back to .*
	return strings.ReplaceAll(escaped, `\*`, `.*`)
}
