// SPDX-License-Identifier: GPL-3.0-or-later

package browser

import (
	"github.com/alyraffauf/goxdgdesktop/desktopfile"
	"github.com/alyraffauf/goxdgdesktop/xdg"
)

// Action is a desktop-entry action, e.g. "new-private-window".
type Action = desktopfile.Action

// ListDesktopActions returns the actions declared in appID's desktop entry.
// It returns nil when the desktop file is missing or cannot be parsed.
func ListDesktopActions(appID string) []Action {
	if appID == "" {
		return nil
	}

	desktopFilePath := xdg.FindDesktopFile(appID)
	if desktopFilePath == "" {
		return nil
	}

	file, err := desktopfile.Read(desktopFilePath)
	if err != nil {
		return nil
	}
	return LocalizedActions(file)
}
