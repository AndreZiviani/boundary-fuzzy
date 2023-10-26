package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type connectedKeyMap struct {
	disconnect key.Binding
	reconnect  key.Binding
}

func (c connectedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.disconnect,
		c.reconnect,
	}
}

func (c connectedKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.disconnect,
			c.reconnect,
		},
	}
}

func newConnectedKeyMap() *connectedKeyMap {
	return &connectedKeyMap{
		disconnect: key.NewBinding(
			key.WithKeys("d", "enter"),
			key.WithHelp("d/enter", "disconnect from target"),
		),
		reconnect: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reconnect to target"),
		),
	}
}

func newConnectedDelegate(model *mainModel) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	keys := newConnectedKeyMap()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.reconnect):
				if i, ok := m.SelectedItem().(*Target); ok {
					// send reconnect event upstream
					return tea.Sequence(
						func() tea.Msg { return terminateSessionMsg{i} },
						func() tea.Msg { return connectMsg{i} },
						func() tea.Msg { return tea.ClearScreen() },
					)
				}
			case key.Matches(msg, keys.disconnect):
				if i, ok := m.SelectedItem().(*Target); ok {
					// send disconnect event upstream
					//m.RemoveItem(m.Index())
					return tea.Sequence(
						func() tea.Msg { return terminateSessionMsg{i} },
						func() tea.Msg { return tea.ClearScreen() },
					)
				}
			}
		case terminateSessionMsg:
			m.RemoveItem(m.Index())
		case connectMsg:
			tmp := msg.Target
			m.InsertItem(len(m.Items()), tmp)
		}

		return nil
	}

	help := []key.Binding{keys.disconnect, keys.reconnect}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
