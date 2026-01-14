// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Browser struct {
	ID   string // desktop file ID (e.g., "firefox.desktop")
	Name string
	Icon string
	Exec string
}

func detectBrowsers() []*Browser {
	var browsers []*Browser
	seen := make(map[string]bool)

	dirs := getApplicationDirs()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}
			if seen[entry.Name()] {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			b := parseDesktopFile(path, entry.Name())
			if b != nil && isBrowser(b) {
				browsers = append(browsers, b)
				seen[entry.Name()] = true
			}
		}
	}

	return browsers
}

func getApplicationDirs() []string {
	var dirs []string
	seen := make(map[string]bool)

	// Helper to add unique directories
	addDir := func(path string) {
		if !seen[path] {
			dirs = append(dirs, path)
			seen[path] = true
		}
	}

	// Start with XDG_DATA_DIRS if set
	if xdg := os.Getenv("XDG_DATA_DIRS"); xdg != "" {
		for _, d := range strings.Split(xdg, ":") {
			addDir(filepath.Join(d, "applications"))
		}
	}

	// When running in a flatpak, also check host system paths
	// The flatpak manifest grants read access to these via --filesystem
	if os.Getenv("FLATPAK_ID") != "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			addDir(filepath.Join(home, ".local", "share", "applications"))
			addDir(filepath.Join(home, ".local", "share", "flatpak", "exports", "share", "applications"))
		}
		addDir("/usr/share/applications")
		addDir("/var/lib/flatpak/exports/share/applications")
		addDir("/var/lib/snapd/desktop/applications")
	}

	// If nothing was found, use fallback paths
	if len(dirs) == 0 {
		home, _ := os.UserHomeDir()
		if home != "" {
			addDir(filepath.Join(home, ".local", "share", "applications"))
			addDir(filepath.Join(home, ".local", "share", "flatpak", "exports", "share", "applications"))
		}
		addDir("/usr/local/share/applications")
		addDir("/usr/share/applications")
		addDir("/var/lib/flatpak/exports/share/applications")
		addDir("/var/lib/snapd/desktop/applications")
	}

	return dirs
}

func parseDesktopFile(path, id string) *Browser {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	b := &Browser{ID: id}
	inDesktopEntry := false
	hasMimeType := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}

		if !inDesktopEntry {
			continue
		}

		if strings.HasPrefix(line, "Name=") {
			b.Name = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Icon=") {
			b.Icon = strings.TrimPrefix(line, "Icon=")
		} else if strings.HasPrefix(line, "Exec=") {
			b.Exec = strings.TrimPrefix(line, "Exec=")
		} else if strings.HasPrefix(line, "MimeType=") {
			mimeTypes := strings.TrimPrefix(line, "MimeType=")
			if strings.Contains(mimeTypes, "x-scheme-handler/http") {
				hasMimeType = true
			}
		} else if strings.HasPrefix(line, "NoDisplay=true") {
			return nil
		}
	}

	if b.Name == "" || b.Exec == "" || !hasMimeType {
		return nil
	}

	return b
}

func isBrowser(b *Browser) bool {
	// Already filtered by MimeType in parseDesktopFile
	// Skip ourselves
	return b.ID != "io.github.alyraffauf.Switchyard.desktop"
}

func launchBrowser(b *Browser, url string) {
	execLine := b.Exec

	// Replace %u, %U, %f, %F with URL
	execLine = strings.ReplaceAll(execLine, "%u", url)
	execLine = strings.ReplaceAll(execLine, "%U", url)
	execLine = strings.ReplaceAll(execLine, "%f", url)
	execLine = strings.ReplaceAll(execLine, "%F", url)

	// Remove other field codes
	for _, code := range []string{"%i", "%c", "%k"} {
		execLine = strings.ReplaceAll(execLine, code, "")
	}

	// If URL wasn't substituted, append it
	if !strings.Contains(b.Exec, "%u") && !strings.Contains(b.Exec, "%U") {
		execLine = execLine + " " + url
	}

	parts := strings.Fields(execLine)
	if len(parts) == 0 {
		return
	}

	// If running in a flatpak, use flatpak-spawn to launch on the host
	var cmd *exec.Cmd
	if os.Getenv("FLATPAK_ID") != "" {
		// Use flatpak-spawn --host to launch browsers on the host system
		args := append([]string{"--host"}, parts...)
		cmd = exec.Command("flatpak-spawn", args...)
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching browser: %v\n", err)
		return
	}

	// Clean up the process asynchronously to prevent zombie processes
	go cmd.Wait()
}
