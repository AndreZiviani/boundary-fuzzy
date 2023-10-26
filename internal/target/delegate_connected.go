package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type connectedKeyMap struct {
	disconnect key.Binding
}

func (c connectedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.disconnect,
	}
}

func (c connectedKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.disconnect,
		},
	}
}

func newConnectedKeyMap() *connectedKeyMap {
	return &connectedKeyMap{
		disconnect: key.NewBinding(
			key.WithKeys("d", "enter"),
			key.WithHelp("d/enter", "disconnect from target"),
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
			case key.Matches(msg, keys.disconnect):
				if i, ok := m.SelectedItem().(Target); ok {
					// send disconnect event upstream
					m.RemoveItem(m.Index())
					return tea.Sequence(
						func() tea.Msg { return terminateSessionMsg{i} },
						func() tea.Msg { return tea.ClearScreen() },
					)
				}
			}
		case connectMsg:
			m.InsertItem(len(m.Items()), msg.Target)
		}

		return nil
	}

	help := []key.Binding{keys.disconnect}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
