// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"net/url"
	"regexp"
	"strings"
)

func ApplyRedirections(rawURL string, redirections []Redirection) string {
	for _, redirection := range redirections {
		rawURL = applyRedirection(rawURL, redirection)
	}
	return rawURL
}

func applyRedirection(rawURL string, redirection Redirection) string {
	redirectionType := NormalizeRedirectionType(redirection.Type)

	switch redirectionType {
	case "domain":
		return applyDomainRedirection(rawURL, redirection)
	case "wildcard":
		return applyWildcardRedirection(rawURL, redirection)
	case "regex":
		return applyRegexRedirection(rawURL, redirection)
	default:
		return rawURL
	}
}

func applyDomainRedirection(rawURL string, redirection Redirection) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if !strings.EqualFold(parsedURL.Hostname(), redirection.Find) {
		return rawURL
	}

	parsedURL.Host = strings.Replace(parsedURL.Host, parsedURL.Hostname(), redirection.Replace, 1)
	return parsedURL.String()
}

func applyWildcardRedirection(rawURL string, redirection Redirection) string {
	pattern := wildcardToRegex(redirection.Find)
	compiledRegex, ok := getCompiledRegex("(?i)" + pattern)
	if !ok {
		return rawURL
	}
	return compiledRegex.ReplaceAllString(rawURL, redirection.Replace)
}

func applyRegexRedirection(rawURL string, redirection Redirection) string {
	compiledRegex, ok := getCompiledRegex(redirection.Find)
	if !ok {
		return rawURL
	}
	return compiledRegex.ReplaceAllString(rawURL, redirection.Replace)
}

func wildcardToRegex(pattern string) string {
	escaped := regexp.QuoteMeta(pattern)
	return strings.ReplaceAll(escaped, `\*`, `.*`)
}
