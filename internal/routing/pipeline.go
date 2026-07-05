// SPDX-License-Identifier: GPL-3.0-or-later

package routing

// PrepareURLForRouting runs the full URL processing pipeline: sanitize,
// optionally strip tracking parameters, and apply redirections.
// Returns "" if the URL was rejected (e.g. mailto:, tel:).
func PrepareURLForRouting(rawURL string, removeTracking bool, redirections []Redirection) string {
	sanitized := SanitizeURL(rawURL)
	if sanitized == "" {
		return ""
	}

	if removeTracking {
		sanitized = RemoveTrackingParameters(sanitized)
	}

	if len(redirections) > 0 {
		sanitized = ApplyRedirections(sanitized, redirections)
	}

	return sanitized
}
