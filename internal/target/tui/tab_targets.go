package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	targetsTabName = "Targets"
)

func TargetsUpdate(t targetDelegate, msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keyMap.binding["shell"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				cmd, err := i.Shell(func(err error) tea.Msg {
					return msgError{err: err}
				})
				if err != nil {
					return tea.Sequence(cmd, func() tea.Msg { return msgError{err: err} })
				}

				return tea.Sequence(cmd)
			}

		case key.Matches(msg, t.keyMap.binding["connect"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				// send connect event upstream
				_, err := i.Connect()
				if err != nil {
					return tea.Sequence(func() tea.Msg { return msgError{err: err} })
				}

				return tea.Sequence(func() tea.Msg { return msgConnect{target: i} })
			}

		case key.Matches(msg, t.keyMap.binding["favorite"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				return tea.Sequence(func() tea.Msg { return msgFavorite{target: i} })
			}

		case key.Matches(msg, t.keyMap.binding["info"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				return tea.Sequence(func() tea.Msg { return msgInfo{target: i} })
			}

		case key.Matches(msg, t.keyMap.binding["refresh"]):
			return tea.Sequence(func() tea.Msg { return msgRefresh{} })

		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return nil
	}

	return nil
}
