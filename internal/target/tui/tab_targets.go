package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (t *tui) HandleTargetsUpdate(ctx context.Context, msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.targetKeyMap.Shell):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				cmd, err := i.Shell(ctx, func(err error) tea.Msg {
					t.SetStateAndMessage(errorView, err.Error())
					return nil
				})
				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
					return true, nil
				}

				return true, tea.Sequence(cmd)
			}

		case key.Matches(msg, t.targetKeyMap.Connect):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				// send connect event upstream
				_, err := i.Connect()
				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
					return true, nil
				}

				t.tabs[connectedView].InsertItem(len(t.tabs[connectedView].Items()), i)
				return true, nil
			}

		case key.Matches(msg, t.targetKeyMap.Favorite):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.tabs[favoriteView].InsertItem(len(t.tabs[favoriteView].Items()), i)
				return true, nil
			}

		case key.Matches(msg, t.targetKeyMap.Info):
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
