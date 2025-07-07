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
		targetKeyMap:    input.TargetKeyMap,
		connectedKeyMap: input.ConnectedKeyMap,
		favoriteKeyMap:  input.FavoriteKeyMap,
	}
	return m
}

func Tui(ctx context.Context, targetListResult *targets.TargetListResult, boundaryClient *api.Client, boundaryToken *authtokens.AuthToken) {
	tuiTargets := make([]list.Item, 0)

	targetList, targetKeyMap := NewList(targetsTabName, targetsView, tuiTargets, map[string]key.Binding{
		bindingShell.name:    bindingShell.binding,
		bindingConnect.name:  bindingConnect.binding,
		bindingFavorite.name: bindingFavorite.binding,
		bindingInfo.name:     bindingInfo.binding,
		bindingRefresh.name:  bindingRefresh.binding,
	})

	connectedList, connectedKeyMap := NewList(connectedTabName, connectedView, []list.Item{}, map[string]key.Binding{
		bindingDisconnect.name: bindingDisconnect.binding,
		bindingReconnect.name:  bindingReconnect.binding,
		bindingInfo.name:       bindingInfo.binding,
		bindingFavorite.name:   bindingFavorite.binding,
	})

	favoriteList, favoriteKeyMap := NewList(favoritesTabName, favoriteView, []list.Item{}, map[string]key.Binding{
		bindingShell.name:        bindingShell.binding,
		bindingDelete.name:       bindingDelete.binding,
		bindingConnect.name:      bindingConnect.binding,
		bindingFavoriteUp.name:   bindingFavoriteUp.binding,
		bindingFavoriteDown.name: bindingFavoriteDown.binding,
		bindingInfo.name:         bindingInfo.binding,
	})

	t := newTui(ctx, TuiInput{
		BoundaryClient: boundaryClient,
		BoundaryToken:  boundaryToken,

		Tabs:            []*list.Model{&targetList, &connectedList, &favoriteList},
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

func NewList(name string, view sessionState, items []list.Item, bindings map[string]key.Binding) (list.Model, *DelegateKeyMap) {
	delegate, keyMap := NewDelegate(bindings, view)
	customList := list.New(items, delegate, 0, 0)
	customList.Title = name
	customList.AdditionalShortHelpKeys = keyMap.ShortHelp
	customList.AdditionalFullHelpKeys = keyMap.ShortHelp
	customList.SetShowTitle(false)
	customList.DisableQuitKeybindings()

	return customList, keyMap
}
