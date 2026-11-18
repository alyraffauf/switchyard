// SPDX-License-Identifier: GPL-3.0-or-later

package routing

import "testing"

func TestWildcardRedirectionInsertsLiteralReplacement(t *testing.T) {
	redirect := Redirection{Type: "wildcard", Find: "/old", Replace: "/$1"}
	input := "https://example.com/old"
	want := "https://example.com/$1"
	if got := ApplyRedirections(input, []Redirection{redirect}); got != want {
		t.Fatalf("ApplyRedirections(%q, %#v) = %q, want %q", input, redirect, got, want)
	}
}
