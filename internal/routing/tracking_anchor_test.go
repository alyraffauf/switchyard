// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import (
	"net/url"
	"strings"
	"testing"
)

func TestTrackingRuleHonorsRegexEndAnchor(t *testing.T) {
	rules := parseRules(strings.NewReader("$removeparam=/^sample_[12]$/"))
	const input = "https://example.com/?sample_2=yes&sample_12=no"
	const want = "https://example.com/?sample_12=no"
	parsed, err := url.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := removeTrackingParametersWithRules(input, parsed, rules); got != want {
		t.Fatalf("removeTrackingParametersWithRules(%q) = %q, want %q", input, got, want)
	}
}
