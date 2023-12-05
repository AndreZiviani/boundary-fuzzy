package target

import (
	"github.com/AndreZiviani/boundary-fuzzy/internal/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type favoriteKeyMap struct {
	shell   key.Binding
	delete  key.Binding
	connect key.Binding
}

func (c favoriteKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.shell,
		c.delete,
		c.connect,
	}
}

func (c favoriteKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.shell,
			c.delete,
			c.connect,
		},
	}
}

func newFavoriteKeyMap() *favoriteKeyMap {
	return &favoriteKeyMap{
		shell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete target from favorites"),
		),
		connect: key.NewBinding(
			key.WithKeys("c", "enter"),
			key.WithHelp("c/enter", "connect to target"),
		),
	}
}

func newFavoriteDelegate(model *mainModel) (list.DefaultDelegate, *favoriteKeyMap) {
	d := list.NewDefaultDelegate()

	keys := newFavoriteKeyMap()

	help := []key.Binding{keys.shell, keys.delete, keys.connect}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d, keys
}

func getTarget(id string, targets []list.Item) list.Item {
	for _, v := range targets {
		if v.(*Target).target.Id == id {
			return v
		}
	}
	return nil
}

func loadFavoriteList(list *list.Model, targets []list.Item) error {
	config, err := config.NewConfig()
	if err != nil {
		return err
	}

	err = config.LoadFavorites()
	if err != nil {
		return err
	}

	for _, v := range config.Favorites {
		target := getTarget(v, targets)
		if target != nil {
			list.InsertItem(len(list.Items()), target)
		}
	}

	return nil
}

func saveFavoriteList(list list.Model) error {
	config, err := config.NewConfig()
	if err != nil {
		return err
	}

	favorites := make([]string, len(list.Items()))
	for i, v := range list.Items() {
		favorites[i] = v.(*Target).target.Id
	}
	config.Favorites = favorites

	err = config.SaveFavorites()
	if err != nil {
		return err
	}

	return nil
}
