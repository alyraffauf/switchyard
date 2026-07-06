// SPDX-License-Identifier: GPL-3.0-or-later

package browser

// Action is an app-specific launch action, e.g. "new-private-window".
type Action struct {
	ID   string
	Name string
	Exec string
}
