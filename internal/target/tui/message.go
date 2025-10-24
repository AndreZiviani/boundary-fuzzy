package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type msgError struct {
	err error
}

type msgConnect struct {
	target *Target
}

type msgFavorite struct {
	target *Target
}

type msgInfo struct {
	target *Target
}

type msgRefresh struct {
}

func (t tui) messageUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		t.message = ""
		t.state = t.previousState
	}
	return t, nil
}

func (t tui) HandleMessageView() string {
	text := alertViewStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			fmt.Sprintf("%s\n\nPress any key to return", messageStyle(t.message)),
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
