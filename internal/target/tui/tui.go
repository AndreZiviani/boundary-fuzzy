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
)

// sessionState is used to track which model is focused
type sessionState uint

type tui struct {
	ctx             context.Context
	state           sessionState
	previousState   sessionState
	index           int
	tabs            []*list.Model
	targetKeyMap    *TargetKeyMap
	connectedKeyMap *ConnectedKeyMap
	favoriteKeyMap  *FavoriteKeyMap

	tabsName       []string
	boundaryClient *api.Client
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
	var cmds []tea.Cmd
	var cmd tea.Cmd
	matched := false

	switch t.state {
	case errorView, messageView:
		return t.messageUpdate(msg)
	case quittingView:
		return t.quittingUpdate(msg)
	case targetsView:
		// Don't match any of the keys below if we're actively filtering.
		if t.InFilterState() {
			break
		}

		matched, cmd = t.HandleTargetsUpdate(t.ctx, msg)

	case connectedView:
		// Don't match any of the keys below if we're actively filtering.
		if t.InFilterState() {
			break
		}

		matched, cmd = t.HandleConnectedUpdate(msg)

	case favoriteView:
		// Don't match any of the keys below if we're actively filtering.
		if t.InFilterState() {
			break
		}

		matched, cmd = t.HandleFavoritesUpdate(msg)
	}

	if matched {
		return t, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if t.InFilterState() {
			break
		}

		switch msg.String() {
		case "q", "ctrl+c":
			t.SetState(quittingView)
			return t, tea.Quit

		case "tab":
			t.GoNextTab()
			cmds = append(cmds, func() tea.Msg { return tea.ClearScreen() })
		}
	}

	updateCmds := t.UpdateTabs(msg)
	cmds = append(cmds, updateCmds...)

	return t, tea.Batch(cmds...)
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
	renderedTabs := t.RenderTabs()

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	remainingTopBorder := ""
	if t.width > 0 {
		rowLength, _ := lipgloss.Size(row)
		borderStyle := windowStyle.GetBorderStyle()

		remainingTopBorder = lipgloss.NewStyle().Foreground(highlightColor).Render(
			strings.Repeat(borderStyle.Top, t.width-rowLength-1) + borderStyle.TopRight,
		)
	}

	header := ""
	if t.boundaryToken != nil {
		header = header + lipgloss.NewStyle().Render(
			"Session expires at "+
				t.boundaryToken.ExpirationTime.Local().Format("2006/01/02 15:04:05 MST")+
				"\n",
		)
	}

	header = header + row + remainingTopBorder
	headerHeight := lipgloss.Height(header) - 1 // ignore last \n
	contentHeight := t.height - windowStyle.GetVerticalFrameSize() - headerHeight
	contentWidth := t.width - windowStyle.GetHorizontalFrameSize()

	t.tabs[targetsView].SetSize(contentWidth, contentHeight)
	t.tabs[connectedView].SetSize(contentWidth, contentHeight)
	t.tabs[favoriteView].SetSize(contentWidth, contentHeight)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		windowStyle.
			Width(contentWidth).
			Height(contentHeight).
			Render(
				t.CurrentTab().View(),
			),
	)
}

func (t *tui) RenderTabs() []string {
	var renderedTabs []string

	for i := range t.tabs {
		isFirst := i == 0
		isLast := i == len(t.tabs)-1
		isActive := i == int(t.state)

		style := inactiveTabStyle

		if isActive {
			style = activeTabStyle
		}

		border, _, _, _, _ := style.GetBorder()

		switch {
		case isFirst && isActive:
			border.BottomLeft = "│"
		case isFirst && !isActive:
			border.BottomLeft = "├"
		case isLast && !isActive:
			border.BottomRight = "┴"
		}

		style = style.Border(border)
		renderedTabs = append(
			renderedTabs,
			style.Render(
				fmt.Sprintf("%s (%d)", t.tabsName[i], len(t.tabs[i].Items())),
			),
		)
	}

	return renderedTabs
}
