// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"regexp"
	"strings"
)

func extractDomain(rawURL string) string {
	// Add scheme if missing so url.Parse works correctly
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func sanitizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	// Reject local file paths
	if strings.HasPrefix(rawURL, "/") || strings.HasPrefix(rawURL, ".") {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Add https scheme if missing
	if u.Scheme == "" {
		rawURL = "https://" + rawURL
		u, _ = url.Parse(rawURL)
	}

	// Only allow browser-routable schemes
	switch u.Scheme {
	case "http", "https", "file", "ftp":
		return u.String()
	default:
		return ""
	}
}

func matchesPattern(url, pattern, patternType string) bool {
	domain := extractDomain(url)

	switch patternType {
	case "domain":
		// Exact domain match
		return strings.EqualFold(domain, pattern)
	case "keyword":
		// URL contains text
		return strings.Contains(strings.ToLower(url), strings.ToLower(pattern))
	case "regex":
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(url)
	case "glob":
		return matchGlob(url, pattern)
	default:
		return false
	}
}

func matchGlob(url, pattern string) bool {
	// Extract domain from URL for matching
	domain := extractDomain(url)

	// Simple glob matching: * matches any characters
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	// Match against domain or full URL
	return re.MatchString(domain) || re.MatchString(url)
}
