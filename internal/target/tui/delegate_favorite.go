package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type FavoriteKeyMap struct {
	Shell    key.Binding
	Delete   key.Binding
	Connect  key.Binding
	MoveUp   key.Binding
	MoveDown key.Binding
	Info key.Binding
}

func (c FavoriteKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.Shell,
		c.Delete,
		c.Connect,
		c.MoveUp,
		c.MoveDown,
		c.Info,
	}
}

func (c FavoriteKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(),
	}
}

func NewFavoriteKeyMap() *FavoriteKeyMap {
	return &FavoriteKeyMap{
		Shell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete target from favorites"),
		),
		Connect: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to target"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "move target up on list"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "move target down on list"),
		),
		Info: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
	}
}

func NewFavoriteDelegate() (list.DefaultDelegate, *FavoriteKeyMap) {
	d := list.NewDefaultDelegate()

	keys := NewFavoriteKeyMap()

	help := []key.Binding{keys.Shell, keys.Delete, keys.Connect, keys.MoveUp, keys.MoveDown}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}
