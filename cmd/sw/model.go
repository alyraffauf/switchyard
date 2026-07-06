package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alyraffauf/switchyard/internal/browser"
	"github.com/alyraffauf/switchyard/internal/browserscan"
	"github.com/alyraffauf/switchyard/internal/host"
)

type model struct {
	browserList    list.Model
	actionList     list.Model
	delegate       list.DefaultDelegate
	actionDelegate list.DefaultDelegate
	urlInput       textinput.Model
	urlFocused     bool
	choice         string
	styles         styles
	width          int
	height         int
	quitting       bool
	pickingAction  bool
}

func (m *model) updateStyles(isDark bool) {
	m.styles = newStyles(isDark)
	m.delegate.Styles = newDelegateStyles(isDark)

	m.browserList.Styles.Title = m.styles.title
	m.browserList.Styles.PaginationStyle = m.styles.pagination
	m.browserList.Styles.HelpStyle = m.styles.help
	m.browserList.SetDelegate(m.delegate)

	m.actionDelegate.Styles = newDelegateStyles(isDark)
	m.actionList.Styles.Title = m.styles.title
	m.actionList.Styles.PaginationStyle = m.styles.pagination
	m.actionList.Styles.HelpStyle = m.styles.help
	m.actionList.SetDelegate(m.actionDelegate)

	m.urlInput.SetStyles(newURLInputStyles(isDark))
}

func (m model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.updateStyles(msg.IsDark())
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if m.urlFocused {
			if next, cmd, handled := handleURLInputKey(msg, m); handled {
				return next, cmd
			}
		} else {
			switch msg.String() {
			case "/":
				m.urlFocused = true
				return m, m.urlInput.Focus()
			}
		}

		if !m.urlFocused {
			if m.pickingAction {
				if next, cmd, handled := handleActionKey(msg, m); handled {
					return next, cmd
				}
			} else {
				if next, cmd, handled := handleBrowserKey(msg, m); handled {
					return next, cmd
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.urlFocused {
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}
	if m.pickingAction {
		m.actionList, cmd = m.actionList.Update(msg)
	} else {
		m.browserList, cmd = m.browserList.Update(msg)
	}
	return m, cmd
}

func launchAndQuit(m model, name, exec string) (model, tea.Cmd, bool) {
	m.choice = name
	url := m.urlInput.Value()
	if err := browser.Launch(exec, url, "", host.InFlatpak()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", name, err)
	}
	m.quitting = true
	return m, tea.Quit, true
}

func handleURLInputKey(msg tea.KeyPressMsg, m model) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc", "tab", "enter":
		m.urlFocused = false
		m.urlInput.Blur()
		return m, nil, true
	}
	return m, nil, false
}

func handleBrowserKey(msg tea.KeyPressMsg, m model) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit, true

	case "enter":
		selected, ok := m.browserList.SelectedItem().(browserItem)
		if !ok {
			return m, nil, true
		}
		return launchAndQuit(m, selected.name, selected.exec)

	case "right":
		selected, ok := m.browserList.SelectedItem().(browserItem)
		if !ok {
			return m, nil, true
		}
		actions := browserscan.ListDesktopActions(selected.id)
		if len(actions) == 0 {
			return m, nil, true
		}
		m.pickingAction = true
		items := make([]list.Item, len(actions))
		for index, action := range actions {
			items[index] = actionItem{name: action.Name, exec: action.Exec}
		}
		return m, m.actionList.SetItems(items), true
	}

	return m, nil, false
}

func handleActionKey(msg tea.KeyPressMsg, m model) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit, true

	case "enter":
		action, ok := m.actionList.SelectedItem().(actionItem)
		if !ok {
			m.quitting = true
			return m, tea.Quit, true
		}
		return launchAndQuit(m, action.name, action.exec)

	case "left", "esc":
		m.pickingAction = false
		return m, nil, true
	}

	return m, nil, false
}

func (m model) View() tea.View {
	if m.quitting {
		return quitView(m)
	}
	if m.choice != "" {
		return launchingView(m)
	}

	var columns string
	if m.pickingAction {
		columns = sideBySideView(m)
	} else {
		columns = browserColumnView(m)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		browserStageRowView(m, stageHeight, columns),
		"",
		statusRowView(m),
	)

	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	v := tea.NewView(m.styles.app.Render(content))
	v.AltScreen = true
	v.WindowTitle = "Switchyard"
	return v
}

func browserStageRowView(m model, height int, content string) string {
	align := lipgloss.Center
	if m.pickingAction {
		align = lipgloss.Left
	}

	positioned := lipgloss.PlaceHorizontal(stageWidth, align, content)
	return lipgloss.NewStyle().
		Width(stageWidth).
		Height(height).
		Render(positioned)
}

func urlBarView(m model) string {
	urlInput := m.urlInput.View()
	return centeredStageRowView(lipgloss.Height(urlInput), m.styles.urlBar.Render(urlInput))
}

func statusRowView(m model) string {
	if m.urlFocused {
		return urlBarView(m)
	}
	return helpView(m)
}

func helpView(m model) string {
	var hints string
	if m.pickingAction {
		hints = "↑↓ navigate  •  ↩ launch  •  ← back  •  / edit URL  •  q quit"
	} else {
		hints = "↑↓ navigate  •  ↩ launch  •  → actions  •  / edit URL  •  q quit"
	}
	return centeredStageRowView(lipgloss.Height(hints), m.styles.helpText.Render(hints))
}

func centeredStageRowView(height int, content string) string {
	positioned := lipgloss.PlaceHorizontal(stageWidth, lipgloss.Center, content)
	return lipgloss.NewStyle().
		Width(stageWidth).
		Height(height).
		Render(positioned)
}

func browserColumnView(m model) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		headerView(m, "Switchyard", browserListWidth),
		m.browserList.View(),
		pagerView(m),
	)
}

func headerView(m model, title string, width int) string {
	header := m.styles.title.
		Width(width).
		Render(title)
	return lipgloss.JoinVertical(lipgloss.Left, header, "")
}

func pagerView(m model) string {
	pager := " "
	if m.browserList.Paginator.TotalPages < 2 {
		return pager
	}
	pager = m.browserList.Paginator.View()
	return lipgloss.PlaceHorizontal(browserListWidth, lipgloss.Center, pager)
}

func sideBySideView(m model) string {
	browserView := browserColumnView(m)

	actionView := lipgloss.JoinVertical(lipgloss.Left,
		headerView(m, "Actions", actionListWidth),
		m.actionList.View(),
	)

	height := max(lipgloss.Height(browserView), lipgloss.Height(actionView))
	separator := m.styles.separator.Render(verticalRule(height))

	return lipgloss.JoinHorizontal(lipgloss.Top, browserView, separator, actionView)
}

func verticalRule(height int) string {
	return strings.Repeat("│\n", height-1) + "│"
}

func quitView(m model) tea.View {
	return tea.NewView(m.styles.quitText.Render("Exiting..."))
}

func launchingView(m model) tea.View {
	return tea.NewView(m.styles.quitText.Render(fmt.Sprintf("Launching %s...", m.choice)))
}
