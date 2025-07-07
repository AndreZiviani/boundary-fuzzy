package tui

import "github.com/charmbracelet/bubbles/key"

var (
	bindingShell = binding{
		name: "shell",
		binding: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
	}
	bindingConnect = binding{
		name: "connect",
		binding: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to target"),
		),
	}
	bindingFavorite = binding{
		name: "favorite",
		binding: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
	}
	bindingInfo = binding{
		name: "info",
		binding: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
	}
	bindingRefresh = binding{
		name: "refresh",
		binding: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh target list"),
		),
	}
	bindingDisconnect = binding{
		name: "disconnect",
		binding: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "disconnect from target"),
		),
	}
	bindingReconnect = binding{
		name: "reconnect",
		binding: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reconnect to target"),
		),
	}
	bindingDelete = binding{
		name: "delete",
		binding: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete target from favorites"),
		),
	}
	bindingFavoriteUp = binding{
		name: "up",
		binding: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "move target up on list"),
		),
	}
	bindingFavoriteDown = binding{
		name: "down",
		binding: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "move target down on list"),
		),
	}
)

type binding struct {
	name        string
	binding 	 key.Binding
}