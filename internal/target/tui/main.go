package tui

import (
	"context"
	"fmt"
	"os"

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
	TargetKeyMap    *TargetKeyMap
	ConnectedKeyMap *ConnectedKeyMap
	FavoriteKeyMap  *FavoriteKeyMap
}

func newTui(ctx context.Context, input TuiInput) tui {
	m := tui{
		ctx:             ctx,
		state:           targetsView,
		previousState:   targetsView,
		boundaryClient:  input.BoundaryClient,
		boundaryToken:   input.BoundaryToken,
		tabs:            input.Tabs,
		tabsName:        input.TabsName,
		targetKeyMap:    input.TargetKeyMap,
		connectedKeyMap: input.ConnectedKeyMap,
		favoriteKeyMap:  input.FavoriteKeyMap,
	}
	return m
}

func Tui(ctx context.Context, targets *targets.TargetListResult, boundaryClient *api.Client, boundaryToken *authtokens.AuthToken) {
	tuiTargets := make([]list.Item, 0)
	sessionClient := sessions.NewClient(boundaryClient)

	for _, t := range targets.Items {
		tuiTargets = append(
			tuiTargets,
			&Target{
				title: fmt.Sprintf("%s (%s)", t.Name, t.Scope.Name),
				description: t.Description,
				target: t,
				sessionClient: sessionClient,

			})
	}

	targetDelegate, targetKeyMap := NewTargetDelegate()
	targetList := list.New(tuiTargets, targetDelegate, 0, 0)
	targetList.SetShowTitle(false)
	targetList.DisableQuitKeybindings()
	connectedDelegate, connectedKeyMap := NewConnectedDelegate()
	connectedList := list.New([]list.Item{}, connectedDelegate, 0, 0)
	connectedList.SetShowTitle(false)
	connectedList.DisableQuitKeybindings()
	favoriteDelegate, favoriteKeyMap := NewFavoriteDelegate()
	favoriteList := list.New([]list.Item{}, favoriteDelegate, 0, 0)
	favoriteList.SetShowTitle(false)
	favoriteList.DisableQuitKeybindings()

	err := loadFavoriteList(&favoriteList, tuiTargets)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	t := newTui(ctx, TuiInput{
		BoundaryClient: boundaryClient,
		BoundaryToken:  boundaryToken,

		Tabs:            []*list.Model{&targetList, &connectedList, &favoriteList},
		TabsName:        []string{"Targets", "Connected", "Favorites"},
		TargetKeyMap:    targetKeyMap,
		ConnectedKeyMap: connectedKeyMap,
		FavoriteKeyMap:  favoriteKeyMap,
	})

	p := tea.NewProgram(t, tea.WithAltScreen(), tea.WithFilter(filter))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
