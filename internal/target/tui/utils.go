package tui

import (
	"net"
	"reflect"
	"regexp"
	"strings"

	"github.com/AndreZiviani/boundary-fuzzy/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/boundary/globals"
)

// This regular expression is used to find all instances of square brackets within a string.
// This regular expression is used to remove the square brackets from an IPv6 address.
var squareBrackets = regexp.MustCompile("\\[|\\]")

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

// SplitHostPort splits a network address of the form "host:port", "host%zone:port", "[host]:port" or "[host%zone]:port" into host or host%zone and port.
//
// A literal IPv6 address in hostport must be enclosed in square brackets, as in "[::1]:80", "[::1%lo0]:80".
func SplitHostPort(hostport string) (host string, port string, err error) {
	host, port, err = net.SplitHostPort(hostport)
	// use the hostport value as a backup when we have a missing port error
	if err != nil && strings.Contains(err.Error(), globals.MissingPortErrStr) {
		// incase the hostport value is an ipv6, we must remove the enclosed square
		// brackets to retain the same behavior as the net.SplitHostPort() method
		host = squareBrackets.ReplaceAllString(hostport, "")
		err = nil
	}
	return
}
