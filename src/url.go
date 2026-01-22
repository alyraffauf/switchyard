// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"net/url"
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

func parseSwitchyardURL(rawURL string) (targetURL string, browserPrefs []string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}

	if u.Scheme != "switchyard" || u.Host != "open" {
		return "", nil, fmt.Errorf("invalid switchyard URL")
	}

	query := u.Query()
	targetURL = query.Get("url")
	if targetURL == "" {
		return "", nil, fmt.Errorf("missing url parameter")
	}

	if browser := query.Get("browser"); browser != "" {
		browserPrefs = strings.Split(browser, ",")
	}

	return targetURL, browserPrefs, nil
}
