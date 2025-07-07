package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/api/sessions"
	"github.com/hashicorp/boundary/api/targets"
)

type TuiInput struct {
	BoundaryClient  *api.Client
	BoundaryToken   *authtokens.AuthToken
	Tabs            []*list.Model
	TabsName        []string
	TargetKeyMap    *DelegateKeyMap
	ConnectedKeyMap *DelegateKeyMap
	FavoriteKeyMap  *DelegateKeyMap
}

func newTui(ctx context.Context, input TuiInput) tui {
	m := tui{
		ctx:             ctx,
		state:           targetsView,
		previousState:   targetsView,
		boundaryClient:  input.BoundaryClient,
		targetsClient:   targets.NewClient(input.BoundaryClient),
		sessionsClient:  sessions.NewClient(input.BoundaryClient),
		boundaryToken:   input.BoundaryToken,
		tabs:            input.Tabs,
		tabsName:        input.TabsName,
		targetKeyMap:    input.TargetKeyMap,
		connectedKeyMap: input.ConnectedKeyMap,
		favoriteKeyMap:  input.FavoriteKeyMap,
	}
	return m
}

func Tui(ctx context.Context, targetListResult *targets.TargetListResult, boundaryClient *api.Client, boundaryToken *authtokens.AuthToken) {
	tuiTargets := make([]list.Item, 0)

	targetDelegate, targetKeyMap := NewDelegate(map[string]key.Binding{
		"shell": key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		"connect": key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to target"),
		),
		"favorite": key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
		"info": key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
		"refresh": key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh target list"),
		),
	}, targetsView)
	targetList := list.New(tuiTargets, targetDelegate, 0, 0)
	targetList.SetShowTitle(false)
	targetList.DisableQuitKeybindings()

	// connectedDelegate, connectedKeyMap := NewConnectedDelegate()
	connectedDelegate, connectedKeyMap := NewDelegate(map[string]key.Binding{
		"disconnect": key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "disconnect from target"),
		),
		"reconnect": key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reconnect to target"),
		),
		"info": key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
		"favorite": key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
	}, connectedView)
	connectedList := list.New([]list.Item{}, connectedDelegate, 0, 0)
	connectedList.SetShowTitle(false)
	connectedList.DisableQuitKeybindings()

	favoriteDelegate, favoriteKeyMap := NewDelegate(map[string]key.Binding{
		"shell": key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		"delete": key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete target from favorites"),
		),
		"connect": key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to target"),
		),
		"up": key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "move target up on list"),
		),
		"down": key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "move target down on list"),
		),
		"info": key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
	}, favoriteView)
	favoriteList := list.New([]list.Item{}, favoriteDelegate, 0, 0)
	favoriteList.SetShowTitle(false)
	favoriteList.DisableQuitKeybindings()

	t := newTui(ctx, TuiInput{
		BoundaryClient: boundaryClient,
		BoundaryToken:  boundaryToken,

		Tabs:            []*list.Model{&targetList, &connectedList, &favoriteList},
		TabsName:        []string{"Targets", "Connected", "Favorites"},
		TargetKeyMap:    targetKeyMap,
		ConnectedKeyMap: connectedKeyMap,
		FavoriteKeyMap:  favoriteKeyMap,
	})

	err := t.refreshTargets()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(t, tea.WithAltScreen(), tea.WithFilter(filter))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
