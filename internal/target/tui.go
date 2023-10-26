package target

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	alertViewStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
	choiceStyle    = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	errorStyle     = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#cc0000"}).
			Render

	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	docStyle          = lipgloss.NewStyle().Padding(2, 1, 0, 1)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(1, 1, 1, 1).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true).Bold(true).Faint(false)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1).Faint(true)
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

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
	tabs           []list.Model
	tabsName       []string
	boundaryClient *api.Client

	width      int
	height     int
	quitting   bool
	shouldQuit bool
	message    string
}

func newModel(boundaryClient *api.Client) mainModel {
	m := mainModel{state: targetsView, boundaryClient: boundaryClient}
	return m
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if len(m.message) > 0 {
		return m.messageUpdate(msg)
	}
	if m.quitting {
		return m.quittingUpdate(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.tabs[m.state].FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			if len(m.tabs[connectedView].Items()) == 0 {
				m.shouldQuit = true
			}
			return m, tea.Quit
		case "tab":
			if m.state == targetsView {
				m.state = connectedView
			} else {
				m.state = targetsView
			}
			return m, tea.Sequence(func() tea.Msg { return tea.ClearScreen() })
		}
	case connectMsg:
		res, cmd := m.tabs[connectedView].Update(msg)
		m.tabs[connectedView] = res
		return m, cmd
	case execErrorMsg:
		m.message = msg.Error()
		return m, cmd
	case terminateSessionMsg:
		TerminateSession(m.boundaryClient, msg.sessionID, msg.task)
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tabs[targetsView].SetSize(msg.Width, msg.Height-6)
		m.tabs[connectedView].SetSize(msg.Width, msg.Height-6)
	}

	switch m.state {
	// update whichever model is focused
	case connectedView:
		m.tabs[connectedView], cmd = m.tabs[connectedView].Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.tabs[targetsView], cmd = m.tabs[targetsView].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	if len(m.message) > 0 {
		text := alertViewStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("Failed to open shell: \n%s\n\nPress any key to return", errorStyle(m.message))))
		return lipgloss.NewStyle().Padding((m.height/2)-1, (m.width-lipgloss.Width(text))/2).Render(text)
	}

	if m.quitting {
		if len(m.tabs[connectedView].Items()) > 0 {
			text := alertViewStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("You have %d active session(s), terminate every session and quit?", len(m.tabs[connectedView].Items())), choiceStyle.Render("[y/N]")))
			return lipgloss.NewStyle().Padding((m.height/2)-1, (m.width-lipgloss.Width(text))/2).Render(text)
		} else {
			return ""
		}
	}

	var renderedTabs []string

	for i, _ := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == int(m.state)
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && !isActive {
			border.BottomRight = "┴"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(fmt.Sprintf("%s (%d)", m.tabsName[i], len(m.tabs[i].Items()))))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	remainingTopBorder := ""
	if m.width > 0 {
		rowLength, _ := lipgloss.Size(row)
		borderStyle := windowStyle.GetBorderStyle()

		remainingTopBorder = lipgloss.NewStyle().Foreground(highlightColor).Render(strings.Repeat(borderStyle.Top, m.width-rowLength-3) + borderStyle.TopRight)
	}

	doc := strings.Builder{}
	doc.WriteString(row + remainingTopBorder)
	doc.WriteString("\n")
	doc.WriteString(
		windowStyle.
			Width(
				(m.width - windowStyle.GetHorizontalFrameSize()),
			).
			Height(
				(m.height - windowStyle.GetVerticalFrameSize()),
			).
			Render(
				m.tabs[m.state].View(),
			),
	)
	return docStyle.Render(doc.String())
}

func (m mainModel) messageUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.message = ""
	}
	return m, nil
}

func (m mainModel) quittingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.tabs[connectedView].Items()) == 0 {
		m.shouldQuit = true
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// For simplicity's sake, we'll treat any key besides "y" as "no"
		if msg.String() == "y" {
			m.shouldQuit = true
			return m, tea.Quit
		}
		m.shouldQuit = false
		m.quitting = false
	}
	return m, nil
}

func filter(teaModel tea.Model, msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.QuitMsg); !ok {
		return msg
	}

	m := teaModel.(mainModel)
	if !m.shouldQuit {
		return nil
	}

	return msg
}

func Tui(targets *targets.TargetListResult, boundaryClient *api.Client) {
	tuiTargets := make([]list.Item, 0)

	for _, t := range targets.Items {
		tuiTargets = append(tuiTargets, Target{title: fmt.Sprintf("%s (%s)", t.Name, t.Scope.Name), description: t.Description, target: t})
	}

	m := newModel(boundaryClient)

	targetList := list.New(tuiTargets, newTargetDelegate(&m), 0, 0)
	targetList.SetShowTitle(false)
	connectedList := list.New([]list.Item{}, newConnectedDelegate(&m), 0, 0)
	connectedList.SetShowTitle(false)

	m.tabs = []list.Model{targetList, connectedList}
	m.tabsName = []string{"Targets", "Connected"}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithFilter(filter))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
