// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"regexp"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

// validateConditions checks if all conditions have non-empty patterns and valid types
func validateConditions(conditions []Condition) bool {
	validTypes := map[string]bool{
		"domain":  true,
		"keyword": true,
		"glob":    true,
		"regex":   true,
	}

	for _, c := range conditions {
		if c.Pattern == "" {
			return false
		}
		if !validTypes[c.Type] {
			return false
		}
		// For regex type, validate the pattern compiles
		if c.Type == "regex" {
			if _, err := regexp.Compile(c.Pattern); err != nil {
				return false
			}
		}
	}
	return true
}

// getLogicFromComboRow extracts the logic string from a combo row selection
func getLogicFromComboRow(logicRow *adw.ComboRow) string {
	if logicRow.Selected() == 1 {
		return "any"
	}
	return "all"
}

// saveConfigWithFlag saves config while setting the global saving flag to prevent file watcher loops
func saveConfigWithFlag(cfg *Config) {
	savingMux.Lock()
	isSaving = true
	savingMux.Unlock()
	saveConfig(cfg)
	glib.TimeoutAdd(100, func() bool {
		savingMux.Lock()
		isSaving = false
		savingMux.Unlock()
		return false
	})
}

// findBrowserByID finds a browser by its desktop file ID
func findBrowserByID(browsers []*Browser, id string) *Browser {
	for _, b := range browsers {
		if b.ID == id {
			return b
		}
	}
	return nil
}
