// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"fmt"
	"net/url"
	"strings"
)

func ExtractDomain(rawURL string) string {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

func SanitizeURL(rawURL string) string {
	rawURL = strings.ReplaceAll(rawURL, "\n", "")
	rawURL = strings.ReplaceAll(rawURL, "\r", "")
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	if strings.HasPrefix(rawURL, "/") || strings.HasPrefix(rawURL, ".") {
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	if parsedURL.Scheme == "" {
		rawURL = "https://" + rawURL
		parsedURL, err = url.Parse(rawURL)
		if err != nil {
			return ""
		}
	}

	switch parsedURL.Scheme {
	case "http", "https", "file", "ftp":
		return parsedURL.String()
	default:
		return ""
	}
}

func ParseSwitchyardURL(rawURL string) (targetURL string, browserPrefs []string, err error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}

	if parsedURL.Scheme != "switchyard" || parsedURL.Host != "open" {
		return "", nil, fmt.Errorf("invalid switchyard URL")
	}

	query := parsedURL.Query()
	targetURL = query.Get("url")
	if targetURL == "" {
		return "", nil, fmt.Errorf("missing url parameter")
	}

	if browser := query.Get("browser"); browser != "" {
		browserPrefs = strings.Split(browser, ",")
	}

	return targetURL, browserPrefs, nil
}
