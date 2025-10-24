package tui

import (
	"github.com/AndreZiviani/boundary-fuzzy/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	favoritesTabName = "Favorites"
)

func FavoritesUpdate(t targetDelegate, msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keyMap.binding["delete"]):
			if _, ok := m.SelectedItem().(*Target); ok {
				m.RemoveItem(m.Index())
				m.CursorUp()
				saveFavoriteList(*m)
				return nil
			}

		case key.Matches(msg, t.keyMap.binding["shell"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				cmd, err := i.Shell(func(err error) tea.Msg {
					return tea.Sequence(func() tea.Msg { return msgError{err: err} })
				})

				if err != nil {
					return tea.Sequence(cmd, func() tea.Msg { return msgError{err: err} })
				}

				return tea.Sequence(cmd)
			}

		case key.Matches(msg, t.keyMap.binding["connect"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				_, err := i.Connect()
				if err != nil {
					return tea.Sequence(func() tea.Msg { return msgError{err: err} })
				}

				return tea.Sequence(func() tea.Msg { return msgConnect{target: i} })
			}

		case key.Matches(msg, t.keyMap.binding["up"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				var cmds []tea.Cmd

				currentIdx := m.Index()
				if currentIdx == 0 {
					return nil
				}

				previous := m.Items()[currentIdx-1].(*Target)

				cmd := m.SetItem(currentIdx-1, i)
				cmds = append(cmds, cmd)

				cmd = m.SetItem(currentIdx, previous)
				cmds = append(cmds, cmd)

				saveFavoriteList(*m)
				return tea.Sequence(cmds...)
			}

		case key.Matches(msg, t.keyMap.binding["down"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				var cmds []tea.Cmd

				currentIdx := m.Index()
				size := len(m.Items())
				if currentIdx == size-1 {
					return nil
				}

				next := m.Items()[currentIdx+1].(*Target)

				cmd := m.SetItem(currentIdx+1, i)
				cmds = append(cmds, cmd)

				cmd = m.SetItem(currentIdx, next)
				cmds = append(cmds, cmd)

				saveFavoriteList(*m)
				return tea.Sequence(cmds...)
			}

		case key.Matches(msg, t.keyMap.binding["info"]):
			if i, ok := m.SelectedItem().(*Target); ok {
				return tea.Sequence(func() tea.Msg { return msgInfo{target: i} })
			}
		}

	case msgFavorite:
		cmd := m.InsertItem(len(m.Items()), msg.target)
		saveFavoriteList(*m)
		return tea.Sequence(cmd)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return nil
	}

	return nil
}

func getTarget(id string, targets []list.Item) list.Item {
	for _, v := range targets {
		if v.(*Target).target.Id == id {
			return v
		}
	}
	return nil
}

func (t *tui) refreshFavoriteList() error {
	config, err := config.NewConfig()
	if err != nil {
		return err
	}

	err = config.LoadFavorites()
	if err != nil {
		return err
	}

	favorites := make([]list.Item, 0, len(config.Favorites))
	for _, v := range config.Favorites {
		target := getTarget(v, t.tabs[targetsView].Items())
		if target != nil {
			favorites = append(favorites, target)
		}
	}

	t.tabs[favoriteView].SetItems(favorites)

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
