package main

import (
	"bufio"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestPreferredBrowserReceivesRedirectedURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("XDG_DATA_DIRS", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("FLATPAK_ID", "")
	t.Setenv("PATH", dir)

	write := func(path, contents string, mode os.FileMode) {
		t.Helper()
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("applications/switchyard-test-preferred.desktop", `[Desktop Entry]
Name=Test Browser
Type=Application
MimeType=x-scheme-handler/https;
Exec=test-browser %u
`, 0o600)
	write("switchyard/config.toml", `[[redirections]]
find = 'old.example'
replace = 'new.example'
`, 0o600)

	// Intercept the shell boundary and record its command without running it.
	result := filepath.Join(dir, "launch")
	if err := syscall.Mkfifo(result, 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.OpenFile(result, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := reader.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SWITCHYARD_TEST_RESULT", result)
	write("sh", "#!/bin/sh\nprintf '%s\\n' \"$2\" > \"$SWITCHYARD_TEST_RESULT\"\n", 0o700)

	args := os.Args
	t.Cleanup(func() { os.Args = args })
	os.Args = []string{"sw", "switchyard://open?url=https%3A%2F%2Fold.example%2Farticle&browser=switchyard-test-preferred"}
	main()

	command, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	const want = "test-browser 'https://new.example/article'\n"
	if command != want {
		t.Fatalf("launch command = %q, want %q", command, want)
	}
}
