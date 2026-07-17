// SPDX-License-Identifier: GPL-3.0-or-later

package startup

import "testing"

func TestLaunchModeBackgroundStart(t *testing.T) {
	mode := LaunchMode{Background: true}

	if !mode.ShouldHold(false) {
		t.Fatal("background start should hold the application even when stay_alive is disabled")
	}
	if mode.ShouldShowWindow() {
		t.Fatal("background start should not show a window")
	}
	if !mode.ShouldHandleLocally(true) {
		t.Fatal("background start should exit locally when another instance is already running")
	}
}

func TestLaunchModeBackgroundOnlyAppliesToFirstActivation(t *testing.T) {
	mode := LaunchMode{Background: true}

	mode.CompleteActivation()

	if !mode.ShouldShowWindow() {
		t.Fatal("a later activation should show a window")
	}
}

func TestLaunchModeNormalStart(t *testing.T) {
	mode := LaunchMode{}

	if mode.ShouldHold(false) {
		t.Fatal("normal start should respect a disabled stay_alive setting")
	}
	if !mode.ShouldHold(true) {
		t.Fatal("normal start should respect an enabled stay_alive setting")
	}
	if !mode.ShouldShowWindow() {
		t.Fatal("normal start should show a window")
	}
	if mode.ShouldHandleLocally(true) {
		t.Fatal("normal start should be forwarded to the primary instance")
	}
}
