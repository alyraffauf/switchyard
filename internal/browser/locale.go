// SPDX-License-Identifier: GPL-3.0-or-later

package browser

import (
	"os"
	"strings"

	"github.com/alyraffauf/goxdgdesktop/desktopfile"
)

// LocalizedString resolves key for the current Desktop Entry locale.
func LocalizedString(file *desktopfile.File, section, key string) string {
	for _, candidate := range localeKeyCandidates(key) {
		if value, ok := file.Get(section, candidate); ok && value != "" {
			return value
		}
	}
	return ""
}

// LocalizedActions resolves each action name for the current locale.
func LocalizedActions(file *desktopfile.File) []Action {
	actions := file.Actions()
	for i := range actions {
		section := desktopfile.ActionSectionStart + actions[i].ID
		if name := LocalizedString(file, section, "Name"); name != "" {
			actions[i].Name = name
		}
	}
	return actions
}

// For lang_COUNTRY@MODIFIER, the spec order is Key[lang_COUNTRY@MODIFIER],
// Key[lang_COUNTRY], Key[lang@MODIFIER], Key[lang], then Key.
func localeKeyCandidates(key string) []string {
	var candidates []string
	for _, loc := range preferredLocales() {
		lang, country, modifier := splitLocale(loc)
		if lang == "" || lang == "C" || lang == "POSIX" {
			continue
		}
		if country != "" && modifier != "" {
			candidates = append(candidates, key+"["+lang+"_"+country+"@"+modifier+"]")
		}
		if country != "" {
			candidates = append(candidates, key+"["+lang+"_"+country+"]")
		}
		if modifier != "" {
			candidates = append(candidates, key+"["+lang+"@"+modifier+"]")
		}
		candidates = append(candidates, key+"["+lang+"]")
	}
	return append(candidates, key)
}

func preferredLocales() []string {
	if language := os.Getenv("LANGUAGE"); language != "" {
		var locales []string
		for _, loc := range strings.Split(language, ":") {
			if loc = strings.TrimSpace(loc); loc != "" {
				locales = append(locales, loc)
			}
		}
		if len(locales) > 0 {
			return locales
		}
	}
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := os.Getenv(env); value != "" {
			return []string{value}
		}
	}
	return nil
}

func splitLocale(loc string) (lang, country, modifier string) {
	if i := strings.IndexByte(loc, '@'); i >= 0 {
		modifier = loc[i+1:]
		loc = loc[:i]
	}
	if i := strings.IndexByte(loc, '.'); i >= 0 {
		loc = loc[:i]
	}
	if i := strings.IndexByte(loc, '_'); i >= 0 {
		return loc[:i], loc[i+1:], modifier
	}
	return loc, "", modifier
}
