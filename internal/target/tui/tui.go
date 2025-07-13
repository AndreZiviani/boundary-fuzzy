package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/api/sessions"
	"github.com/hashicorp/boundary/api/targets"
)

// sessionState is used to track which model is focused
type sessionState uint

type tui struct {
	ctx             context.Context
	state           sessionState
	previousState   sessionState
	tabs            []*list.Model
	targetKeyMap    *DelegateKeyMap
	connectedKeyMap *DelegateKeyMap
	favoriteKeyMap  *DelegateKeyMap

	boundaryClient *api.Client
	targetsClient  *targets.Client
	sessionsClient *sessions.Client
	boundaryToken  *authtokens.AuthToken

	width      int
	height     int
	shouldQuit bool
	message    string
}

const (
	targetsView sessionState = iota
	connectedView
	favoriteView
	messageView
	errorView
	quittingView
)

func (t tui) Init() tea.Cmd {
	return nil
}

func (t tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t.state {
	case errorView, messageView:
		return t.messageUpdate(msg)
	case quittingView:
		return t.quittingUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

		// remove tabs and borders from the window size message
		msg.Height -= 4
		msg.Width -= 2

		return t, t.UpdateTabs(msg)
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if t.InFilterState() {
			break
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return t.gracefullyQuit(msg)

		case "tab":
			t.GoNextTab()
			return t, func() tea.Msg { return tea.ClearScreen() }
		default:
			// only send custom messages to the current tab
			m, cmd := t.CurrentTab().Update(msg)
			t.tabs[t.state] = &m
			return t, cmd
		}

	case msgError:
		t.SetStateAndMessage(errorView, msg.err.Error())
		return t, nil

	case msgInfo:
		t.SetStateAndMessage(messageView, msg.target.Info())
		return t, nil

	default:
		// propagate everything else to all tabs
		cmd := t.UpdateTabs(msg)
		return t, cmd
	}

	return t, nil
}

func (t tui) View() string {
	switch t.state {
	case messageView:
		return t.HandleMessageView()

	case errorView:
		return t.HandleErrorView()

	case quittingView:
		return t.HandleQuittingView()

	default:
		return t.HandleDefaultView()

	}
}

func (t *tui) HandleDefaultView() string {
	if !t.isInitialized() {
		return ""
	}

	renderedTabs := t.RenderTabs()

	style := lipgloss.NewStyle().
		MaxHeight(t.height).
		MaxWidth(t.width)

	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			renderedTabs,
			windowStyle.
				// force width here to make sure border is rendered correctly
				Width(t.width-2).
				Render(
					t.CurrentTab().View(),
				),
		),
	)
}

func (t *tui) RenderTabs() string {
	out := []string{}
	for i := range t.tabs {
		isFirst := i == 0
		isLast := i == len(t.tabs)-1
		isActive := i == int(t.state)

		style := tabStyle
		if isActive {
			style = activeTab
		}

		border := style.GetBorderStyle()
		switch {
		case isFirst && isActive:
			border.BottomLeft = "│"
		case isFirst && !isActive:
			border.BottomLeft = "├"
		case isLast && !isActive:
			border.BottomRight = "┴"
		}

		out = append(out, style.Border(border).Render(
			fmt.Sprintf("%s (%d)", t.tabs[i].Title, len(t.tabs[i].Items())),
		))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)

	gap := lipgloss.NewStyle().Foreground(highlight).Render(
		// Create a gap with the same width as the row, but with the tab border on the right
		strings.Repeat(tabBorder.Top, max(0, t.width-lipgloss.Width(row)-1)) + tabBorder.TopRight)

	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	return row
}

func (t *tui) isInitialized() bool {
	return t.width != 0 && t.height != 0
}
