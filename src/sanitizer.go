// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const adguardURL = "https://raw.githubusercontent.com/AdguardTeam/FiltersRegistry/master/filters/filter_17_TrackParam/filter.txt"

var (
	sanitizerRules []adGuardRule
	ruleMutex      sync.RWMutex
	initOnce       sync.Once
)

type adGuardRule struct {
	isException bool
	urlRegex    *regexp.Regexp // Filter for the URL (e.g., ||domain.com^)
	paramRegex  *regexp.Regexp // Filter for the parameter name
	paramName   string         // Literal parameter name if not regex
}

func InitSanitizer(cfg *Config) {
	if !cfg.SanitizeLinks {
		return
	}
	initOnce.Do(func() {
		go func() {
			path := filepath.Join(configDir(), "adguard_filter.txt")
			if info, err := os.Stat(path); err != nil || time.Since(info.ModTime()) > 24*time.Hour {
				if resp, err := http.Get(adguardURL); err == nil && resp.StatusCode == 200 {
					defer resp.Body.Close()
					os.MkdirAll(filepath.Dir(path), 0755)
					if out, err := os.Create(path); err == nil {
						defer out.Close()
						_, _ = out.ReadFrom(resp.Body)
					}
				}
			}

			if file, err := os.Open(path); err == nil {
				defer file.Close()
				var rules []adGuardRule
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					if rule, ok := parseRule(scanner.Text()); ok {
						rules = append(rules, rule)
					}
				}
				ruleMutex.Lock()
				sanitizerRules = rules
				ruleMutex.Unlock()
			}
		}()
	})
}

func parseRule(line string) (adGuardRule, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "!") {
		return adGuardRule{}, false
	}

	rule := adGuardRule{}
	if strings.HasPrefix(line, "@@") {
		rule.isException = true
		line = line[2:]
	}

	parts := strings.Split(line, "$")
	urlPattern := parts[0]
	
	// Convert AdGuard URL pattern to Regex
	if urlPattern != "" {
		pattern := regexp.QuoteMeta(urlPattern)
		pattern = strings.ReplaceAll(pattern, `\|\|`, `(^|\.)`) // AdGuard || 
		pattern = strings.ReplaceAll(pattern, `\^`, `([^a-zA-Z0-9.\-%]|$)`) // AdGuard ^
		pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
		if re, err := regexp.Compile("(?i)" + pattern); err == nil {
			rule.urlRegex = re
		}
	}

	// Parse $removeparam
	if len(parts) > 1 {
		options := strings.Split(parts[1], ",")
		for _, opt := range options {
			if strings.HasPrefix(opt, "removeparam=") {
				p := strings.TrimPrefix(opt, "removeparam=")
				if strings.HasPrefix(p, "/") && strings.HasSuffix(p, "/") {
					if re, err := regexp.Compile("(?i)" + p[1:len(p)-1]); err == nil {
						rule.paramRegex = re
					}
				} else {
					rule.paramName = p
				}
				return rule, true
			} else if opt == "removeparam" {
				rule.paramName = "*" // Remove all tracking params matches
				return rule, true
			}
		}
	}
	return adGuardRule{}, false
}

func applyTrackingProtection(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.RawQuery == "" {
		return rawURL
	}

	ruleMutex.RLock()
	rules := sanitizerRules
	ruleMutex.RUnlock()

	q := u.Query()
	changed := false

	for param := range q {
		shouldRemove := false
		isWhitelisted := false

		for _, rule := range rules {
			// 1. Does the URL match this rule's domain/path filter?
			if rule.urlRegex != nil && !rule.urlRegex.MatchString(rawURL) {
				continue
			}

			// 2. Does this parameter match the rule's target?
			matchesParam := false
			if rule.paramName == "*" || rule.paramName == param {
				matchesParam = true
			} else if rule.paramRegex != nil && rule.paramRegex.MatchString(param) {
				matchesParam = true
			}

			if matchesParam {
				if rule.isException {
					isWhitelisted = true
					break // If whitelisted, we stop checking for this param
				} else {
					shouldRemove = true
				}
			}
		}

		if shouldRemove && !isWhitelisted {
			q.Del(param)
			changed = true
		}
	}

	if changed {
		u.RawQuery = q.Encode()
		return u.String()
	}
	return rawURL
}
