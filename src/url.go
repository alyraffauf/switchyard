// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"net/url"
	"os"
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
		u, err = url.Parse(rawURL)
		if err != nil {
			return ""
		}
	}

	// Handle special schemes
	switch u.Scheme {
	case "file":
		// Handle file:// URIs that GIO sometimes creates from bare domains
		if _, err := os.Stat(u.Path); os.IsNotExist(err) {
			// File doesn't exist - might be a bare domain that GIO converted
			// Extract the last path component as a potential domain
			lastSlash := strings.LastIndex(u.Path, "/")
			if lastSlash != -1 {
				possibleDomain := u.Path[lastSlash+1:]
				if looksLikeDomain(possibleDomain) {
					return "https://" + possibleDomain
				}
			}
		}
		// Real file path - pass through for local HTML files
		return u.String()
	case "mailto", "tel", "sms", "data", "javascript":
		// These should be handled by xdg-open, not routed to browsers
		return ""
	}

	return u.String()
}

func looksLikeDomain(s string) bool {
	if strings.Contains(s, " ") || !strings.Contains(s, ".") {
		return false
	}

	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}

	// Reject common file extensions
	fileExtensions := map[string]bool{
		"txt": true, "pdf": true, "doc": true, "docx": true,
		"jpg": true, "jpeg": true, "png": true, "gif": true,
		"zip": true, "tar": true, "gz": true,
	}

	lastPart := strings.ToLower(parts[len(parts)-1])
	return len(lastPart) > 1 && !fileExtensions[lastPart]
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
