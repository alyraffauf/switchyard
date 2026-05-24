// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed embedded/native-messaging-host-chromium.json
var chromiumManifestTemplate string

//go:embed embedded/native-messaging-host-firefox.json
var firefoxManifestTemplate string

// Must match HOST in webextension/popup.js.
const nativeHostName = "io.github.alyraffauf.switchyard"

var chromiumConfigSubdirs = []string{
	"net.imput.helium",
	"google-chrome",
	"chromium",
	"BraveSoftware/Brave-Browser",
	"microsoft-edge",
	"vivaldi",
}

var firefoxHomeSubdirs = []string{
	".mozilla",
	".librewolf",
	".zen",
}

func installNativeHost() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	wrapperPath, err := ensureNativeHostWrapper(home)
	if err != nil {
		return err
	}
	fmt.Println("Wrapper:", wrapperPath)

	chromiumJSON := []byte(strings.ReplaceAll(chromiumManifestTemplate, "@SWITCHYARD_NM@", wrapperPath))
	firefoxJSON := []byte(strings.ReplaceAll(firefoxManifestTemplate, "@SWITCHYARD_NM@", wrapperPath))

	installed := 0
	for _, sub := range chromiumConfigSubdirs {
		if installManifestTo(filepath.Join(home, ".config", sub), "NativeMessagingHosts", chromiumJSON) {
			installed++
		}
	}
	for _, sub := range firefoxHomeSubdirs {
		if installManifestTo(filepath.Join(home, sub), "native-messaging-hosts", firefoxJSON) {
			installed++
		}
	}

	if installed == 0 {
		fmt.Println("\nNo supported browsers found. Install a Chromium- or Firefox-family browser first.")
		return nil
	}
	fmt.Printf("\nInstalled %d manifest(s). Restart your browser to detect the native host.\n", installed)
	return nil
}

func uninstallNativeHost() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	removed := 0
	for _, sub := range chromiumConfigSubdirs {
		if uninstallManifestFrom(filepath.Join(home, ".config", sub), "NativeMessagingHosts") {
			removed++
		}
	}
	for _, sub := range firefoxHomeSubdirs {
		if uninstallManifestFrom(filepath.Join(home, sub), "native-messaging-hosts") {
			removed++
		}
	}

	wrappers := []string{
		nativeHostWrapperPath(home),
		// Legacy locations from earlier installs.
		filepath.Join(home, ".local/share/flatpak/exports/bin/io.github.alyraffauf.Switchyard-native-host"),
		filepath.Join(home, ".local/share/flatpak/exports/bin/io.github.alyraffauf.Switchyard.Devel-native-host"),
	}
	for _, p := range wrappers {
		if removeFileHost(p) {
			fmt.Println("  removed " + p)
		}
	}

	fmt.Printf("\nRemoved %d manifest(s).\n", removed)
	return nil
}

func installManifestTo(profileDir, manifestSubdir string, content []byte) bool {
	if _, err := os.Stat(profileDir); err != nil {
		return false
	}
	target := filepath.Join(profileDir, manifestSubdir, nativeHostName+".json")
	if err := writeFileHost(target, content, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "warning:", target+":", err)
		return false
	}
	fmt.Println("  " + target)
	return true
}

func uninstallManifestFrom(profileDir, manifestSubdir string) bool {
	target := filepath.Join(profileDir, manifestSubdir, nativeHostName+".json")
	if !removeFileHost(target) {
		return false
	}
	fmt.Println("  removed " + target)
	return true
}

func nativeHostWrapperPath(home string) string {
	return filepath.Join(home, ".local/share/switchyard/native-host-wrapper.sh")
}

// ensureNativeHostWrapper writes the script the browser will exec as the
// manifest's "path" field. The manifest format forbids inline args, so we
// need a shim to add --native-host before re-entering switchyard.
func ensureNativeHostWrapper(home string) (string, error) {
	wrapperPath := nativeHostWrapperPath(home)
	var script string
	if id := os.Getenv("FLATPAK_ID"); id != "" {
		script = fmt.Sprintf("#!/bin/sh\nexec /usr/bin/flatpak run --command=switchyard %s --native-host \"$@\"\n", id)
	} else {
		self, err := os.Executable()
		if err != nil {
			return "", err
		}
		script = fmt.Sprintf("#!/bin/sh\nexec %s --native-host \"$@\"\n", shellQuote(self))
	}
	if err := writeFileHost(wrapperPath, []byte(script), 0o755); err != nil {
		return "", fmt.Errorf("create wrapper at %s: %w", wrapperPath, err)
	}
	return wrapperPath, nil
}

// writeFileHost writes to a path on the host filesystem. Inside Flatpak the
// sandbox can read but not write most host paths, so it shells out via
// flatpak-spawn instead of writing directly.
func writeFileHost(path string, content []byte, mode os.FileMode) error {
	if os.Getenv("FLATPAK_ID") != "" {
		script := fmt.Sprintf("mkdir -p %s && cat > %s && chmod %o %s",
			shellQuote(filepath.Dir(path)), shellQuote(path), mode, shellQuote(path))
		cmd := exec.Command("flatpak-spawn", "--host", "sh", "-c", script)
		cmd.Stdin = bytes.NewReader(content)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, mode)
}

func removeFileHost(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	if os.Getenv("FLATPAK_ID") != "" {
		return exec.Command("flatpak-spawn", "--host", "rm", "-f", path).Run() == nil
	}
	return os.Remove(path) == nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
