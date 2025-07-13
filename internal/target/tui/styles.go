package tui

import "github.com/charmbracelet/lipgloss"

const (
	bullet   = "•"
	ellipsis = "…"
)

var (
	alertViewStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
	choiceStyle    = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	messageStyle   = lipgloss.NewStyle().
			Render
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#cc0000"}).
			Render

	windowStyle = lipgloss.NewStyle().BorderForeground(highlight).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()

	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	activeTab = tabStyle.Border(activeTabBorder, true)
)
