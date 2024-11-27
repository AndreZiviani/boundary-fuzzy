package tui

import (
	"github.com/AndreZiviani/boundary-fuzzy/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (t *tui) HandleFavoritesUpdate(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.favoriteKeyMap.Delete):
			if _, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.CurrentTab().RemoveItem(t.CurrentTab().Index())
				t.CurrentTab().CursorUp()
				saveFavoriteList(*t.CurrentTab())
				return true, nil
			}

		case key.Matches(msg, t.favoriteKeyMap.Shell):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				cmd, err := i.Shell(t.ctx, func(err error) tea.Msg {
					t.SetStateAndMessage(errorView, err.Error())
					return nil
				})

				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
					return true, nil
				}

				return true, tea.Sequence(cmd)
			}

		case key.Matches(msg, t.favoriteKeyMap.Connect):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				_, err := i.Connect()
				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
					return true, nil
				}

				t.tabs[connectedView].InsertItem(len(t.tabs[connectedView].Items()), i)
				return true, nil
			}

		case key.Matches(msg, t.favoriteKeyMap.MoveUp):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				var cmds []tea.Cmd

				currentIdx := t.CurrentTab().Index()
				if currentIdx == 0 {
					return true, nil
				}

				previous := t.CurrentTab().Items()[currentIdx-1].(*Target)

				cmd := t.CurrentTab().SetItem(currentIdx-1, i)
				cmds = append(cmds, cmd)

				cmd = t.CurrentTab().SetItem(currentIdx, previous)
				cmds = append(cmds, cmd)

				saveFavoriteList(*t.CurrentTab())
				return true, tea.Sequence(cmds...)
			}

		case key.Matches(msg, t.favoriteKeyMap.MoveDown):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				var cmds []tea.Cmd

				currentIdx := t.CurrentTab().Index()
				size := len(t.CurrentTab().Items())
				if currentIdx == size-1 {
					return true, nil
				}

				next := t.CurrentTab().Items()[currentIdx+1].(*Target)

				cmd := t.CurrentTab().SetItem(currentIdx+1, i)
				cmds = append(cmds, cmd)

				cmd = t.CurrentTab().SetItem(currentIdx, next)
				cmds = append(cmds, cmd)

				saveFavoriteList(*t.CurrentTab())
				return true, tea.Sequence(cmds...)
			}

		case key.Matches(msg, t.favoriteKeyMap.Info):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.SetStateAndMessage(messageView, i.Info())
				return true, nil
			}

			// Prioritize our keybinding instead of default
		case key.Matches(msg, listKeyMap(t.CurrentTab().KeyMap)...):
			cmd := t.UpdateCurrentTab(msg)
			return true, cmd
		}
	}

	return false, nil
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
