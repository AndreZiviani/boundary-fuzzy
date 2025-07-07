package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	targetsTabName = "Targets"
)

func (t *tui) HandleTargetsUpdate(ctx context.Context, msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.targetKeyMap.binding["shell"]):
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

		case key.Matches(msg, t.targetKeyMap.binding["connect"]):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				// send connect event upstream
				_, err := i.Connect(t.ctx)
				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
					return true, nil
				}

				t.tabs[connectedView].InsertItem(len(t.tabs[connectedView].Items()), i)
				return true, nil
			}

		case key.Matches(msg, t.targetKeyMap.binding["favorite"]):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.tabs[favoriteView].InsertItem(len(t.tabs[favoriteView].Items()), i)
				return true, nil
			}

		case key.Matches(msg, t.targetKeyMap.binding["info"]):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.SetStateAndMessage(messageView, i.Info())
				return true, nil
			}

		case key.Matches(msg, t.targetKeyMap.binding["refresh"]):
			err := t.refreshTargets()
			if err != nil {
				t.SetStateAndMessage(errorView, err.Error())
			}

			return true, nil

		// Prioritize our keybinding instead of default
		case key.Matches(msg, listKeyMap(t.CurrentTab().KeyMap)...):
			cmd := t.UpdateCurrentTab(msg)
			return true, cmd
		}
	}

	return false, nil
}
