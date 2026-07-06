package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/alyraffauf/switchyard/internal/browser"
	"github.com/alyraffauf/switchyard/internal/browserscan"
	"github.com/alyraffauf/switchyard/internal/config"
	"github.com/alyraffauf/switchyard/internal/host"
	"github.com/alyraffauf/switchyard/internal/routing"
)

const (
	browserListWidth = 30
	actionListWidth  = 25
	listHeight       = 20
	listHeaderHeight = 2
	listPagerHeight  = 1

	browserListHeight = listHeight - listHeaderHeight - listPagerHeight
	actionListHeight  = browserListHeight

	// Everything is laid out inside this fixed-size box and then centered, so
	// opening actions never resizes the outer block. Width has slack for the
	// widest help line, which is wider than the browser+separator+actions columns.
	stageWidth  = 70
	stageHeight = listHeight
)

func initialModel(cfg *config.Config, url string) model {
	hiddenSet := make(map[string]bool, len(cfg.HiddenBrowsers))
	for _, id := range cfg.HiddenBrowsers {
		hiddenSet[id] = true
	}

	browsers := browserscan.Installed()
	items := make([]list.Item, 0, len(browsers))
	for _, installed := range browsers {
		if hiddenSet[installed.ID] {
			continue
		}
		items = append(items, browserItem{name: installed.Name, id: installed.ID, exec: installed.Exec})
	}

	if cfg.FavoriteBrowser != "" {
		for index, item := range items {
			candidate, ok := item.(browserItem)
			if ok && candidate.id == cfg.FavoriteBrowser {
				items[0], items[index] = items[index], items[0]
				break
			}
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles = newDelegateStyles(true)

	browserList := list.New(items, delegate, browserListWidth, listHeight)
	browserList.SetHeight(browserListHeight)
	browserList.SetShowTitle(false)
	browserList.SetShowStatusBar(false)
	browserList.SetShowPagination(false)
	browserList.SetFilteringEnabled(false)
	browserList.SetShowHelp(false)

	actionDelegate := list.NewDefaultDelegate()
	actionDelegate.ShowDescription = false
	actionDelegate.Styles = newDelegateStyles(true)

	actionList := list.New([]list.Item{}, actionDelegate, actionListWidth, actionListHeight)
	actionList.SetShowTitle(false)
	actionList.SetShowStatusBar(false)
	actionList.SetShowPagination(false)
	actionList.SetFilteringEnabled(false)
	actionList.SetShowHelp(false)

	urlInput := textinput.New()
	urlInput.Prompt = "🔗 "
	urlInput.Placeholder = "https://…"
	urlInput.SetValue(url)
	urlInput.CharLimit = 0
	urlInput.SetWidth(browserListWidth)

	m := model{
		browserList:    browserList,
		actionList:     actionList,
		delegate:       delegate,
		actionDelegate: actionDelegate,
		urlInput:       urlInput,
	}
	m.updateStyles(true)
	return m
}

func findBrowser(browsers []browser.Browser, id string) *browser.Browser {
	for i := range browsers {
		if browsers[i].ID == id {
			return &browsers[i]
		}
	}
	return nil
}

func launchOrLog(browsers []browser.Browser, id, url string) bool {
	match := findBrowser(browsers, id)
	if match == nil {
		return false
	}
	if err := browser.Launch(match.Exec, url, "", host.InFlatpak()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", match.Name, err)
	}
	return true
}

func main() {
	rawURL := ""
	if len(os.Args) > 1 {
		rawURL = os.Args[1]
	}

	if u, err := url.Parse(rawURL); err == nil && u.Scheme == "switchyard" {
		targetURL, browserPrefs, err := routing.ParseSwitchyardURL(rawURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid switchyard URL: %v\n", err)
			os.Exit(1)
		}
		allBrowsers := browserscan.Installed()
		for _, browserPreference := range browserPrefs {
			if launchOrLog(allBrowsers, browserPreference, targetURL) {
				return
			}

			if strings.HasSuffix(browserPreference, ".desktop") {
				continue
			}

			if launchOrLog(allBrowsers, browserPreference+".desktop", targetURL) {
				return
			}
		}
		rawURL = targetURL
	}

	cfg, _ := config.Load(config.Path())
	sanitized := routing.PrepareURLForRouting(rawURL, cfg.RemoveTrackingParameters, cfg.Redirections)

	if rawURL != "" && sanitized == "" {
		if err := host.HostCommand("xdg-open", rawURL).Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", rawURL, err)
		}
		os.Exit(0)
	}

	allBrowsers := browserscan.Installed()

	if sanitized != "" {
		if browserID, alwaysAsk, matched := cfg.MatchRule(sanitized); matched {
			if !alwaysAsk && launchOrLog(allBrowsers, browserID, sanitized) {
				return
			}
		} else if !cfg.PromptOnClick && cfg.FavoriteBrowser != "" {
			if launchOrLog(allBrowsers, cfg.FavoriteBrowser, sanitized) {
				return
			}
		}
	}

	if _, err := tea.NewProgram(initialModel(cfg, sanitized)).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
