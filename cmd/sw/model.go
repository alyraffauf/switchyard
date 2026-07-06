package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alyraffauf/switchyard/internal/browser"
	"github.com/alyraffauf/switchyard/internal/host"
)

type model struct {
	browserList    list.Model
	actionList     list.Model
	delegate       list.DefaultDelegate
	actionDelegate list.DefaultDelegate
	url            string
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

	var cmd tea.Cmd
	if m.pickingAction {
		m.actionList, cmd = m.actionList.Update(msg)
	} else {
		m.browserList, cmd = m.browserList.Update(msg)
	}
	return m, cmd
}

func launchAndQuit(m model, name, exec string) (model, tea.Cmd, bool) {
	m.choice = name
	if err := browser.Launch(exec, m.url, "", host.InFlatpak()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", name, err)
	}
	m.quitting = true
	return m, tea.Quit, true
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
		actions := browser.ListDesktopActions(selected.id)
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

	var content string
	if m.pickingAction {
		content = sideBySideView(m)
	} else {
		content = m.browserList.View()
	}

	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	v := tea.NewView(m.styles.app.Render(content))
	v.AltScreen = true
	v.WindowTitle = "Switchyard"
	return v
}

func sideBySideView(m model) string {
	browserView := m.browserList.View()

	// Mirror the list's title bar geometry (Padding(0, 0, 1, 2)) so the first
	// action row lines up with the first browser row.
	header := lipgloss.NewStyle().
		Padding(0, 0, 1, 2).
		Render(m.styles.title.Render("Actions"))
	actionView := lipgloss.JoinVertical(lipgloss.Left, header, m.actionList.View())

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
