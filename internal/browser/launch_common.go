// SPDX-License-Identifier: GPL-3.0-or-later

package browser

// Launch starts the platform command for cmdline/url in the background.
// Empty command lines are no-ops.
func Launch(cmdline, url, activationToken string, inFlatpak bool) error {
	cmd := buildCommand(cmdline, url, activationToken, inFlatpak)
	if cmd == nil {
		return nil
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
