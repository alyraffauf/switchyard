// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

// Desktop file actions have IDs, names, and exec commands.
type DesktopAction struct {
	ID   string // action ID (e.g., "new-private-window")
	Name string // display name (e.g., "New Private Window")
	Exec string // Exec command line for this action
}

// findDesktopFile locates a desktop file by ID using XDG Base Directory specification.
// It searches in XDG_DATA_HOME and XDG_DATA_DIRS, plus Flatpak-specific locations.
func findDesktopFile(appID string) string {
	// Build list of data directories to search, following XDG spec
	var dataDirs []string

	// XDG_DATA_HOME (default: ~/.local/share)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = os.Getenv("HOME") + "/.local/share"
	}
	dataDirs = append(dataDirs, dataHome)

	// XDG_DATA_DIRS (default: /usr/local/share:/usr/share)
	dataDirsEnv := os.Getenv("XDG_DATA_DIRS")
	if dataDirsEnv == "" {
		dataDirsEnv = "/usr/local/share:/usr/share"
	}
	for _, dir := range strings.Split(dataDirsEnv, ":") {
		if dir != "" {
			dataDirs = append(dataDirs, dir)
		}
	}

	// Add Flatpak-specific locations
	dataDirs = append(dataDirs, "/var/lib/flatpak/exports/share")
	if home := os.Getenv("HOME"); home != "" {
		dataDirs = append(dataDirs, home+"/.local/share/flatpak/exports/share")
	}

	// Search for appID in each data directory's applications subdirectory
	for _, dataDir := range dataDirs {
		path := dataDir + "/applications/" + appID
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

/** ListDesktopActions returns available actions for an AppInfo by parsing its desktop file directly. For some reason, AppInfo does not expose actions. So we either call g_desktop_app_info_list_actions, which is not bound to Go, or parse it ourselves. I'd prefer to not use Cgo. **/
func ListDesktopActions(appInfo *gio.AppInfo) []DesktopAction {
	if appInfo == nil {
		return nil
	}

	// Get the desktop file path
	// For GIO AppInfo, we can use the ID to find the desktop file
	appID := appInfo.ID()
	if appID == "" {
		return nil
	}

	// Find the desktop file using XDG Base Directory specification
	desktopFilePath := findDesktopFile(appID)
	if desktopFilePath == "" {
		return nil
	}

	// Parse the desktop file
	return parseDesktopFileActions(desktopFilePath)
}

// parseDesktopFileActions parses a desktop file and extracts all actions.
// Desktop files are INI-format with sections like:
//
//	[Desktop Entry]
//	Actions=new-window;new-private-window;
//
//	[Desktop Action new-window]
//	Name=New Window
//	Exec=firefox --new-window %u
//
//	[Desktop Action new-private-window]
//	Name=New Private Window
//	Exec=firefox --private-window %u
func parseDesktopFileActions(path string) []DesktopAction {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var actions []DesktopAction
	var currentSection string
	var currentActionID string
	var currentActionName string
	var currentActionExec string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous action if we were in one
			if currentSection != "" && strings.HasPrefix(currentSection, "Desktop Action ") {
				if currentActionID != "" && currentActionName != "" && currentActionExec != "" {
					/** Only add if exec contains %u or %U (accepts URLs)
					This is commented because it breaks man Chromium-based browers completely, and even some Firefox-based brwosers allow passing URLs to actions unequiped to deal with them (e.g. LibreWolf). Looks less polished for the user, but works in more cases. **/

					// if strings.Contains(currentActionExec, "%u") || strings.Contains(currentActionExec, "%U") {
					actions = append(actions, DesktopAction{
						ID:   currentActionID,
						Name: currentActionName,
						Exec: currentActionExec,
					})
					// }
				}
				currentActionID = ""
				currentActionName = ""
				currentActionExec = ""
			}

			currentSection = strings.TrimSpace(line[1 : len(line)-1])

			// Extract action ID from section name
			if strings.HasPrefix(currentSection, "Desktop Action ") {
				currentActionID = strings.TrimPrefix(currentSection, "Desktop Action ")
			}
			continue
		}

		// Key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// In [Desktop Action ...], look for Name and Exec
			if strings.HasPrefix(currentSection, "Desktop Action ") {
				if key == "Name" {
					currentActionName = value
				} else if key == "Exec" {
					currentActionExec = value
				}
			}
		}
	}

	// Don't forget to save the last action if file ends while in an action section
	if currentSection != "" && strings.HasPrefix(currentSection, "Desktop Action ") {
		if currentActionID != "" && currentActionName != "" && currentActionExec != "" {
			actions = append(actions, DesktopAction{
				ID:   currentActionID,
				Name: currentActionName,
				Exec: currentActionExec,
			})
		}
	}

	return actions
}
