package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type TargetKeyMap struct {
	Shell    key.Binding
	Connect  key.Binding
	Favorite key.Binding
	Info     key.Binding
}

func (c TargetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.Shell,
		c.Connect,
		c.Favorite,
		c.Info,
	}
}

func (c TargetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(),
	}
}

func NewTargetKeyMap() *TargetKeyMap {
	return &TargetKeyMap{
		Shell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		Connect: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to target"),
		),
		Favorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
		Info: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
	}
}

func NewTargetDelegate() (list.DefaultDelegate, *TargetKeyMap) {
	d := list.NewDefaultDelegate()

	keys := NewTargetKeyMap()

	help := []key.Binding{keys.Shell, keys.Connect, keys.Favorite}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}
