package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type connectedKeyMap struct {
	disconnect key.Binding
	reconnect  key.Binding
	info       key.Binding
	favorite   key.Binding
}

func (c connectedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.disconnect,
		c.reconnect,
		c.info,
		c.favorite,
	}
}

func (c connectedKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.disconnect,
			c.reconnect,
			c.info,
			c.favorite,
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
		favorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
	}
}

func newConnectedDelegate() (list.DefaultDelegate, *connectedKeyMap) {
	d := list.NewDefaultDelegate()

	keys := newConnectedKeyMap()

	help := []key.Binding{keys.disconnect, keys.reconnect, keys.info, keys.favorite}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}
