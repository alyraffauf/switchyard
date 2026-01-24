// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"regexp"
	"strings"
)

func applyRewrites(rawURL string, rewrites []Rewrite) string {
	for _, r := range rewrites {
		rawURL = applyRewrite(rawURL, r)
	}
	return rawURL
}

func applyRewrite(rawURL string, r Rewrite) string {
	rwType := r.Type
	if rwType == "" {
		rwType = "domain"
	}

	switch rwType {
	case "domain":
		return applyDomainRewrite(rawURL, r)
	case "url":
		return applyURLRewrite(rawURL, r)
	default:
		return rawURL
	}
}

func applyDomainRewrite(rawURL string, r Rewrite) string {
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

func applyURLRewrite(rawURL string, r Rewrite) string {
	pattern := wildcardToRegex(r.Find)
	re, ok := getCompiledRegex("(?i)" + pattern) // case-insensitive
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
