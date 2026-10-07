// SPDX-License-Identifier: GPL-3.0-or-later

// Package startup describes how a Switchyard invocation should start.
package startup

// LaunchMode contains invocation-specific startup behavior.
type LaunchMode struct {
	Background bool
}

// ShouldHold reports whether the application should keep running without a
// window.
func (mode LaunchMode) ShouldHold(stayAlive bool) bool {
	return mode.Background || stayAlive
}

// ShouldShowWindow reports whether activation should present the settings
// window.
func (mode LaunchMode) ShouldShowWindow() bool {
	return !mode.Background
}

// CompleteActivation clears invocation-specific behavior so later activations
// forwarded to this process behave normally.
func (mode *LaunchMode) CompleteActivation() {
	mode.Background = false
}

// ShouldHandleLocally reports whether this invocation can exit without
// activating an existing primary instance.
func (mode LaunchMode) ShouldHandleLocally(remote bool) bool {
	return mode.Background && remote
}
