package main

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

const (
	green       = "#4f895d"
	greenBright = "#6dd791"
	greenDark   = "#3d6b48"
)

type styles struct {
	app        lipgloss.Style
	title      lipgloss.Style
	pagination lipgloss.Style
	help       lipgloss.Style
	helpText   lipgloss.Style
	quitText   lipgloss.Style
	separator  lipgloss.Style
	urlBar     lipgloss.Style
}

func newStyles(darkBG bool) styles {
	var s styles

	s.app = lipgloss.NewStyle().
		Padding(1, 2)

	s.title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color(green)).
		Align(lipgloss.Center)

	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)

	s.help = list.DefaultStyles(darkBG).HelpStyle.
		PaddingLeft(4).
		PaddingBottom(1)

	s.quitText = lipgloss.NewStyle().
		Margin(1, 0, 2, 4).
		Foreground(lipgloss.Color(green))

	ld := lipgloss.LightDark(darkBG)
	s.separator = lipgloss.NewStyle().
		Foreground(ld(lipgloss.Color("#b0b0b0"), lipgloss.Color("240"))).
		Padding(0, 2)

	s.urlBar = lipgloss.NewStyle().
		MarginBottom(0)

	s.helpText = lipgloss.NewStyle().
		MarginTop(0).
		Foreground(ld(lipgloss.Color("#999999"), lipgloss.Color("#666666")))

	return s
}

func newDelegateStyles(darkBG bool) list.DefaultItemStyles {
	s := list.NewDefaultItemStyles(darkBG)
	ld := lipgloss.LightDark(darkBG)

	s.NormalTitle = s.NormalTitle.Foreground(ld(lipgloss.Color("#1a1a1a"), lipgloss.Color("#dddddd")))
	s.NormalDesc = s.NormalDesc.Foreground(ld(lipgloss.Color("#888888"), lipgloss.Color("241")))

	s.SelectedTitle = s.SelectedTitle.
		Foreground(ld(lipgloss.Color(greenDark), lipgloss.Color(greenBright))).
		BorderForeground(ld(lipgloss.Color(greenDark), lipgloss.Color(greenBright))).
		Bold(true)

	s.SelectedDesc = s.SelectedDesc.
		Foreground(ld(lipgloss.Color("#888888"), lipgloss.Color("241"))).
		BorderForeground(ld(lipgloss.Color(greenDark), lipgloss.Color(greenBright)))

	s.DimmedTitle = s.DimmedTitle.Foreground(ld(lipgloss.Color("#A49FA5"), lipgloss.Color("#777777")))
	s.DimmedDesc = s.DimmedDesc.Foreground(ld(lipgloss.Color("#C2B8C2"), lipgloss.Color("#4D4D4D")))

	return s
}

func newURLInputStyles(darkBG bool) textinput.Styles {
	s := textinput.DefaultStyles(darkBG)
	ld := lipgloss.LightDark(darkBG)

	s.Focused.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(greenBright))
	s.Blurred.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(green))

	s.Focused.Placeholder = lipgloss.NewStyle().
		Foreground(ld(lipgloss.Color("#999999"), lipgloss.Color("#666666")))
	s.Blurred.Placeholder = s.Focused.Placeholder

	return s
}
