package tui

import (
	"reflect"

	"github.com/AndreZiviani/boundary-fuzzy/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func listKeyMap(keymap list.KeyMap) []key.Binding {
	var bindings []key.Binding
	v := reflect.ValueOf(keymap)
	for i := 0; i < v.NumField(); i++ {
		bindings = append(bindings, v.Field(i).Interface().(key.Binding))
	}

	return bindings
}

func (t *tui) InFilterState() bool {
	return t.CurrentTab().FilterState() == list.Filtering
}

func (t *tui) CurrentTab() *list.Model {
	return t.tabs[t.state]
}

func (t *tui) UpdateCurrentTab(msg tea.Msg) tea.Cmd {
	new, cmd := t.CurrentTab().Update(msg)
	t.tabs[t.state] = &new
	return cmd
}

func (t *tui) SetState(state sessionState) {
	t.previousState = t.state
	t.state = state
}

func (t *tui) SetStateAndMessage(state sessionState, msg string) {
	t.SetState(state)
	t.message = msg
}

func (t *tui) GoNextTab() {
	switch t.state {
	case targetsView:
		t.state = connectedView
	case connectedView:
		t.state = favoriteView
	case favoriteView:
		t.state = targetsView
	}
}

func (t *tui) UpdateTabs(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, 0)

	for _, tab := range []sessionState{targetsView, connectedView, favoriteView} {
		upd, cmd := t.tabs[tab].Update(msg)
		t.tabs[tab] = &upd
		cmds = append(cmds, cmd)
	}

	return cmds
}
func (t *tui) terminateAllSessions() {
	for _, item := range t.tabs[connectedView].Items() {
		target := item.(*Target)
		if target.session == nil {
			target.session.Terminate(t.ctx, target.task)
		}
	}
}

func (t *tui) saveFavoriteList(list *list.Model) error {
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

func filter(teaModel tea.Model, msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.QuitMsg); !ok {
		return msg
	}

	m := teaModel.(tui)
	if !m.shouldQuit {
		return nil
	}

	return msg
}