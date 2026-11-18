// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import "testing"

func TestRemoveTrackingParametersPreservesOtherQueryFields(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{
			name:  "keeps an unrelated field with a literal semicolon",
			input: "https://example.com/?colours=red;blue&utm_source=demo",
			want:  "https://example.com/?colours=red;blue",
		},
		{
			name:  "keeps the order spelling and multiplicity of other fields",
			input: "https://example.com/a%2Fb?z=one%20two&utm_source=first&flag&a=1&z=three&utm%5Fsource=last#part%2f2",
			want:  "https://example.com/a%2Fb?z=one%20two&flag&a=1&z=three#part%2f2",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := RemoveTrackingParameters(test.input); got != test.want {
				t.Errorf("RemoveTrackingParameters(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
