// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin

package browserscan

import (
	"github.com/alyraffauf/goxdgdesktop/desktopfile"
	"github.com/alyraffauf/switchyard/internal/browser"
)

func ListDesktopActions(appID string) []browser.Action {
	if appID == "" {
		return nil
	}

	desktopFilePath := findDesktopFile(appID)
	if desktopFilePath == "" {
		return nil
	}

	file, err := desktopfile.Read(desktopFilePath)
	if err != nil {
		return nil
	}
	return localizedActions(file)
}
