package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (t tui) HandleErrorView() string {
	text := alertViewStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			fmt.Sprintf("Failed to open shell: \n%s\n\nPress any key to return", errorStyle(t.message)),
		),
	)

	paddingHeight := (t.height - lipgloss.Height(text)) / 2
	paddingWidth := (t.width - lipgloss.Width(text)) / 2

	return lipgloss.NewStyle().Padding(
		paddingHeight-1,
		paddingWidth,
		0,
	).Render(text)
}
