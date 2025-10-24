package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	connectedTabName = "Connected"
)

func ConnectedUpdate(t targetDelegate, msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keyMap.binding["reconnect"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				i.session.Terminate(i.task)
				i.session = nil
				_, err := i.Connect()
				if err != nil {
					return tea.Sequence(func() tea.Msg { return msgError{err: err} })
				}
				return nil
			}

		case key.Matches(msg, t.keyMap.binding["disconnect"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				i.session.Terminate(i.task)
				i.session = nil
				m.RemoveItem(m.Index())
				m.CursorUp()

				return nil
			}

		case key.Matches(msg, t.keyMap.binding["info"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				return tea.Sequence(func() tea.Msg { return msgInfo{target: i} })
			}

		case key.Matches(msg, t.keyMap.binding["favorite"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				return tea.Sequence(func() tea.Msg { return msgFavorite{target: i} })
			}
		}

	case msgConnect:
		return m.InsertItem(len(m.Items()), msg.target)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return nil
	}

	return nil
}
