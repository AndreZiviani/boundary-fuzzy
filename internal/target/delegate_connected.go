package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type connectedKeyMap struct {
	disconnect key.Binding
	reconnect  key.Binding
	info       key.Binding
}

func (c connectedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.disconnect,
		c.reconnect,
		c.info,
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
			key.WithKeys("d"),
			key.WithHelp("d", "disconnect from target"),
		),
		reconnect: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reconnect to target"),
		),
		info: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
	}
}

func newConnectedDelegate(model *mainModel) (list.DefaultDelegate, *connectedKeyMap) {
	d := list.NewDefaultDelegate()

	keys := newConnectedKeyMap()

	help := []key.Binding{keys.disconnect, keys.reconnect, keys.info}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}
