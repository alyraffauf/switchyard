// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/alyraffauf/goxdgdesktop/desktopfile"
	"github.com/alyraffauf/goxdgdesktop/xdg"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

type DesktopAction = desktopfile.Action

func ListDesktopActions(appInfo *gio.AppInfo) []DesktopAction {
	if appInfo == nil {
		return nil
	}

	appID := appInfo.ID()
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
	return file.Actions()
}
