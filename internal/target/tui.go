package target

import (
	"context"
	"fmt"
	"os"

	"github.com/AndreZiviani/boundary-fuzzy/internal/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/targets"
)

// sessionState is used to track which model is focused
type sessionState uint

const (
	targetsView sessionState = iota
	connectedView
)

var (
	modelStyle         = lipgloss.NewStyle().BorderStyle(lipgloss.HiddenBorder())
	targetListStyle    = modelStyle
	connectedListStyle = modelStyle
	focusedModelStyle  = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

func NewConnect() *Connect {
	connect := Connect{}

	return &connect
}

func (c *Connect) Execute(ctx context.Context) error {
	var err error
	c.boundaryClient, err = client.NewBoundaryClient()
	if err != nil {
		return err
	}

	c.targetClient = targets.NewClient(c.boundaryClient)

	result, err := c.targetClient.List(context.TODO(), "global", targets.WithRecursive(true))
	if err != nil {
		return err
	}

	Tui(result, c.boundaryClient)
	return nil
}

type mainModel struct {
	state          sessionState
	index          int
	targets        list.Model
	connected      list.Model
	boundaryClient *api.Client
}

func newModel(boundaryClient *api.Client) mainModel {
	m := mainModel{state: targetsView, boundaryClient: boundaryClient}
	return m
}

func (m mainModel) currentFocusedModel() sessionState {
	switch m.state {
	case targetsView:
		return targetsView
	default:
		return connectedView
	}
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.state == targetsView {
				m.state = connectedView
			} else {
				m.state = targetsView
			}
		}
	case connectMsg:
		res, cmd := m.connected.Update(msg)
		m.connected = res
		return m, cmd
	case terminateSessionMsg:
		res, cmd := m.targets.Update(msg)
		m.targets = res
		cmds = append(cmds, cmd)

		res, cmd = m.connected.Update(msg)
		m.connected = res
		cmds = append(cmds, cmd)

		TerminateSession(m.boundaryClient, msg.sessionID, msg.task)
		return m, cmd
	case tea.WindowSizeMsg:
		//m.targets.SetSize(msg.Width-3, msg.Height-2)
		m.targets.SetSize((msg.Width/2)-3, msg.Height-2)
		m.connected.SetSize((msg.Width/2)-3, msg.Height-2)
	}

	switch m.state {
	// update whichever model is focused
	case connectedView:
		m.connected, cmd = m.connected.Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.targets, cmd = m.targets.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	switch m.currentFocusedModel() {
	case targetsView:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			focusedModelStyle.Render(m.targets.View()),
			connectedListStyle.Render(m.connected.View()),
		)
	default:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			targetListStyle.Render(m.targets.View()),
			focusedModelStyle.Render(m.connected.View()),
		)
	}
}

func Tui(targets *targets.TargetListResult, boundaryClient *api.Client) {
	tuiTargets := make([]list.Item, 0)

	for _, t := range targets.Items {
		tuiTargets = append(tuiTargets, Target{title: fmt.Sprintf("%s (%s)", t.Name, t.Scope.Name), description: t.Description, target: t})
	}

	m := newModel(boundaryClient)

	targetList := list.New(tuiTargets, newTargetDelegate(&m), 0, 0)
	targetList.Title = "Targets"
	connectedList := list.New([]list.Item{}, newConnectedDelegate(&m), 0, 0)
	connectedList.Title = "Connected"

	m.targets = targetList
	m.connected = connectedList
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
