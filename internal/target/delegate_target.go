package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type connectMsg struct {
	*Target
}

type execErrorMsg struct {
	error
}

type terminateSessionMsg struct {
	*Target
}

type openShellMsg struct {
	Target
}

type targetKeyMap struct {
	shell    key.Binding
	connect  key.Binding
	favorite key.Binding
}

func (c targetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.shell,
		c.connect,
		c.favorite,
	}
}

func (c targetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.shell,
			c.connect,
			c.favorite,
		},
	}
}

func newTargetKeyMap() *targetKeyMap {
	return &targetKeyMap{
		shell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		connect: key.NewBinding(
			key.WithKeys("c", "enter"),
			key.WithHelp("c/enter", "connect to target"),
		),
		favorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
	}
}

func newTargetDelegate() (list.DefaultDelegate, *targetKeyMap) {
	d := list.NewDefaultDelegate()

	keys := newTargetKeyMap()

	help := []key.Binding{keys.shell, keys.connect, keys.favorite}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}
