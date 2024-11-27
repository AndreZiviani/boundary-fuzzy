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

	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	docStyle          = lipgloss.NewStyle()
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true).Bold(true).Faint(false)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1).Faint(true)
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
