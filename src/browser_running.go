// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// getRunningBrowsers returns a filtered list containing only the browsers
// from allBrowsers that are currently running on the system.
// This is a really nasty hack, and will probably never be merged.
func getRunningBrowsers(allBrowsers []*Browser) []*Browser {
	// Get process list - use flatpak-spawn if we're in a Flatpak
	var cmd *exec.Cmd
	if os.Getenv("FLATPAK_ID") != "" {
		cmd = exec.Command("flatpak-spawn", "--host", "ps", "aux")
	} else {
		cmd = exec.Command("ps", "aux")
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting running browsers: %v\n", err)
		return nil
	}

	processOutput := string(output)
	runningMap := make(map[string]*Browser)

	for _, browser := range allBrowsers {
		if isRunning(browser, processOutput) {
			runningMap[browser.ID] = browser
		}
	}

	// Convert map to slice
	running := make([]*Browser, 0, len(runningMap))
	for _, browser := range runningMap {
		running = append(running, browser)
	}

	return running
}

// isRunning checks if a specific browser is currently running
func isRunning(browser *Browser, processOutput string) bool {
	exe := browser.AppInfo.Executable()
	lines := strings.Split(processOutput, "\n")

	// Check if this is a Flatpak app (executable is /usr/bin/flatpak)
	if exe == "/usr/bin/flatpak" {
		appID := strings.TrimSuffix(browser.ID, ".desktop")

		// Strategy 1: Look for "flatpak run ... appID" pattern
		flatpakRunPattern := "flatpak run"

		// Strategy 2: Look for "--name appID" (used by Firefox)
		namePattern := "--name " + appID

		// Strategy 3: Look for app data directory in process arguments
		// Flatpak apps store data in ~/.var/app/APP_ID/
		appDataPattern := ".var/app/" + appID

		for _, line := range lines {
			// Check flatpak run command
			if strings.Contains(line, flatpakRunPattern) && strings.Contains(line, appID) {
				runIndex := strings.Index(line, flatpakRunPattern)
				appIndex := strings.Index(line, appID)
				if runIndex >= 0 && appIndex > runIndex {
					return true
				}
			}

			// Check --name flag (Firefox style)
			if strings.Contains(line, namePattern) {
				return true
			}

			// Check app data directory (Brave, Edge, etc.)
			if strings.Contains(line, appDataPattern) {
				return true
			}
		}

		return false
	}

	// For system packages and other installations, check the executable path
	if exe != "" {
		baseName := filepath.Base(exe)
		isSystemPath := strings.Contains(exe, "/usr/") || strings.Contains(exe, "/opt/")
		isFlatpakPath := strings.Contains(exe, "/app/")

		for _, line := range lines {
			// Skip if the line doesn't contain our executable name
			if !strings.Contains(line, baseName) {
				continue
			}

			// For system packages: must contain the full system path
			if isSystemPath && strings.Contains(line, exe) {
				return true
			}

			// For Flatpak: must contain /app/ path structure
			if isFlatpakPath && strings.Contains(line, "/app/") && strings.Contains(line, baseName) {
				return true
			}

			// Unknown installation - match on executable path
			if !isSystemPath && !isFlatpakPath && strings.Contains(line, exe) {
				return true
			}
		}
	}

	return false
}
