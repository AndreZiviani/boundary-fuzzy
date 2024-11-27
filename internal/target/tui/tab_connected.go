package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (t *tui) HandleConnectedUpdate(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.connectedKeyMap.Reconnect):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				i.session.Terminate(t.ctx, i.task)
				i.session = nil
				_, err := i.Connect()
				if err != nil {
					t.SetStateAndMessage(errorView, err.Error())
				}
				return true, nil
			}

		case key.Matches(msg, t.connectedKeyMap.Disconnect):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				i.session.Terminate(t.ctx, i.task)
				i.session = nil
				t.CurrentTab().RemoveItem(t.CurrentTab().Index())
				t.CurrentTab().CursorUp()

				return true, nil
			}

		case key.Matches(msg, t.connectedKeyMap.Info):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.SetStateAndMessage(messageView, i.Info())
				return true, nil
			}

		case key.Matches(msg, t.connectedKeyMap.Favorite):
			if i, ok := t.CurrentTab().SelectedItem().(*Target); ok {
				t.tabs[favoriteView].InsertItem(len(t.tabs[favoriteView].Items()), i)
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
