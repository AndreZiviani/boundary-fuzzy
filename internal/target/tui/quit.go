package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (t tui) quittingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(t.tabs[connectedView].Items()) == 0 {
		t.shouldQuit = true
		t.terminateAllSessions()
		t.saveFavoriteList(t.tabs[favoriteView])

		return t, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// For simplicity's sake, we'll treat any key besides "y" as "no"
		if msg.String() == "y" {
			t.shouldQuit = true
			t.terminateAllSessions()
			return t, tea.Quit
		}
		t.shouldQuit = false
		t.state = t.previousState
	}
	return t, nil
}

func (t tui) HandleQuittingView() string {
	if len(t.tabs[connectedView].Items()) != 0 {
		return ""
	}

	text := alertViewStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			fmt.Sprintf("You have %d active session(s), terminate every session and quit?", len(t.tabs[connectedView].Items())),
			choiceStyle.Render("[y/N]"),
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
