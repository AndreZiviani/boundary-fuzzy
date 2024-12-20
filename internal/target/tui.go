package target

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/AndreZiviani/boundary-fuzzy/internal/client"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/api/targets"
)

// sessionState is used to track which model is focused
type sessionState uint

const (
	targetsView sessionState = iota
	connectedView
	favoriteView
	messageView
	errorView
	quittingView
)

var (
	alertViewStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
	choiceStyle    = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	messageStyle   = lipgloss.NewStyle().
			Render
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#cc0000"}).
			Render

	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	docStyle          = lipgloss.NewStyle()
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
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
	boundaryClient, boundaryToken, err := client.NewBoundaryClient(ctx)
	c.boundaryClient = boundaryClient
	if err != nil {
		return errors.Join(err, fmt.Errorf("\nReauthenticate or provide a token via BOUNDARY_TOKEN env var or -token flag. Reading the token can also be disabled via -keyring-type=none."))
	}

	c.targetClient = targets.NewClient(c.boundaryClient)

	result, err := c.targetClient.List(context.TODO(), "global", targets.WithRecursive(true))
	if err != nil {
		return err
	}

	Tui(result, c.boundaryClient, boundaryToken)
	return nil
}

type mainModel struct {
	state           sessionState
	previousState   sessionState
	index           int
	tabs            []list.Model
	targetKeyMap    *targetKeyMap
	connectedKeyMap *connectedKeyMap
	favoriteKeyMap  *favoriteKeyMap

	tabsName       []string
	boundaryClient *api.Client
	boundaryToken  *authtokens.AuthToken

	width      int
	height     int
	shouldQuit bool
	message    string
}

func newModel(boundaryClient *api.Client, boundaryToken *authtokens.AuthToken) mainModel {
	m := mainModel{
		state:          targetsView,
		previousState:  targetsView,
		boundaryClient: boundaryClient,
		boundaryToken:  boundaryToken,
	}
	return m
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.state {
	case errorView, messageView:
		return m.messageUpdate(msg)
	case quittingView:
		return m.quittingUpdate(msg)
	case targetsView:
		// Don't match any of the keys below if we're actively filtering.
		if m.tabs[m.state].FilterState() == list.Filtering {
			break
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.targetKeyMap.shell):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					cmd, err := m.Shell(i)
					if err != nil {
						m.previousState = m.state
						m.state = errorView
						m.message = err.Error()
						return m, nil
					}

					return m, tea.Sequence(cmd)
				}
			case key.Matches(msg, m.targetKeyMap.connect):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					// send connect event upstream
					task, _, session, err := ConnectToTarget(i)
					if err != nil {
						m.previousState = m.state
						m.state = errorView
						m.message = err.Error()
						return m, nil
					}

					i.session = session
					i.task = task
					m.tabs[connectedView].InsertItem(len(m.tabs[connectedView].Items()), i)
					return m, nil
				}
			case key.Matches(msg, m.targetKeyMap.favorite):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					m.tabs[favoriteView].InsertItem(len(m.tabs[favoriteView].Items()), i)
					return m, nil
				}
			// Prioritize our keybinding instead of default
			case key.Matches(msg, listKeyMap(m.tabs[m.state].KeyMap)...):
				m.tabs[m.state], cmd = m.tabs[m.state].Update(msg)
				return m, cmd
			}
		}
	case connectedView:
		// Don't match any of the keys below if we're actively filtering.
		if m.tabs[m.state].FilterState() == list.Filtering {
			break
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.connectedKeyMap.reconnect):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					TerminateSession(m.boundaryClient, i.session, i.task)
					task, _, session, err := ConnectToTarget(i)
					if err != nil {
						m.previousState = m.state
						m.state = errorView
						m.message = err.Error()
						return m, nil
					}
					i.session = session
					i.task = task
					return m, nil
				}
			case key.Matches(msg, m.connectedKeyMap.disconnect):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					TerminateSession(m.boundaryClient, i.session, i.task)
					m.tabs[m.state].RemoveItem(m.tabs[m.state].Index())
					return m, nil
				}
			case key.Matches(msg, m.connectedKeyMap.info):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					m.previousState = m.state
					m.state = messageView
					m.message = fmt.Sprintf(
						"Scope: %s\n"+
							"Scope Description: %s\n"+
							"Name: %s\n"+
							"Description: %s\n"+
							"\n"+
							"Port: %d\n"+
							"Expiration: %s\n"+
							"Session Id: %s\n",
						i.target.Scope.Name, i.target.Scope.Description, i.target.Name, i.target.Description,
						i.session.Port, i.session.Expiration, i.session.SessionId,
					)
					if len(i.session.Credentials) > 0 {
						m.message = fmt.Sprintf(
							"%s"+
								"Dynamic Credentials:\n"+
								"  Username: %s\n"+
								"  Password: %s\n",
							m.message, i.session.Credentials[0].Secret.Decoded["username"], i.session.Credentials[0].Secret.Decoded["password"],
						)
					}
					return m, nil
				}
			case key.Matches(msg, m.connectedKeyMap.favorite):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					m.tabs[favoriteView].InsertItem(len(m.tabs[favoriteView].Items()), i)
					return m, nil
				}
			// Prioritize our keybinding instead of default
			case key.Matches(msg, listKeyMap(m.tabs[m.state].KeyMap)...):
				m.tabs[m.state], cmd = m.tabs[m.state].Update(msg)
				return m, cmd
			}
		}
	case favoriteView:
		// Don't match any of the keys below if we're actively filtering.
		if m.tabs[m.state].FilterState() == list.Filtering {
			break
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.favoriteKeyMap.delete):
				if _, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					m.tabs[m.state].RemoveItem(m.tabs[m.state].Index())
					return m, nil
				}
			case key.Matches(msg, m.favoriteKeyMap.shell):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					cmd, err := m.Shell(i)
					if err != nil {
						m.previousState = m.state
						m.state = errorView
						m.message = err.Error()
						return m, nil
					}

					return m, tea.Sequence(cmd)
				}
			case key.Matches(msg, m.favoriteKeyMap.connect):
				if i, ok := m.tabs[m.state].SelectedItem().(*Target); ok {
					// send connect event upstream
					task, _, session, err := ConnectToTarget(i)
					if err != nil {
						m.previousState = m.state
						m.state = errorView
						m.message = err.Error()
						return m, nil
					}

					i.session = session
					i.task = task
					m.tabs[connectedView].InsertItem(len(m.tabs[connectedView].Items()), i)
					return m, nil
				}

			// Prioritize our keybinding instead of default
			case key.Matches(msg, listKeyMap(m.tabs[m.state].KeyMap)...):
				m.tabs[m.state], cmd = m.tabs[m.state].Update(msg)
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.tabs[m.state].FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.previousState = m.state
			m.state = quittingView
			return m, tea.Quit
		case "tab":
			switch m.state {
			case targetsView:
				m.state = connectedView
			case connectedView:
				m.state = favoriteView
			case favoriteView:
				m.state = targetsView
			}
			cmds = append(cmds, func() tea.Msg { return tea.ClearScreen() })
		}
	}

	m.tabs[targetsView], cmd = m.tabs[targetsView].Update(msg)
	cmds = append(cmds, cmd)
	m.tabs[connectedView], cmd = m.tabs[connectedView].Update(msg)
	cmds = append(cmds, cmd)
	m.tabs[favoriteView], cmd = m.tabs[favoriteView].Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	switch m.state {
	case messageView:
		text := alertViewStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("%s\n\nPress any key to return", messageStyle(m.message))))
		return lipgloss.NewStyle().Padding(((m.height-lipgloss.Height(text))/2)-1, (m.width-lipgloss.Width(text))/2, 0).Render(text)
	case errorView:
		text := alertViewStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("Failed to open shell: \n%s\n\nPress any key to return", errorStyle(m.message))))
		return lipgloss.NewStyle().Padding(((m.height-lipgloss.Height(text))/2)-1, (m.width-lipgloss.Width(text))/2, 0).Render(text)
	case quittingView:
		if len(m.tabs[connectedView].Items()) > 0 {
			text := alertViewStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("You have %d active session(s), terminate every session and quit?", len(m.tabs[connectedView].Items())), choiceStyle.Render("[y/N]")))
			return lipgloss.NewStyle().Padding(((m.height-lipgloss.Height(text))/2)-1, (m.width-lipgloss.Width(text))/2, 0).Render(text)
		} else {
			return ""
		}
	default:

		var renderedTabs []string

		for i := range m.tabs {
			var style lipgloss.Style
			isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == int(m.state)
			if isActive {
				style = activeTabStyle
			} else {
				style = inactiveTabStyle
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

			remainingTopBorder = lipgloss.NewStyle().Foreground(highlightColor).Render(strings.Repeat(borderStyle.Top, m.width-rowLength-1) + borderStyle.TopRight)
		}

		expires := lipgloss.NewStyle().Render("Session expires at " + m.boundaryToken.ExpirationTime.Local().Format("2006/01/02 15:04:05 MST") + "\n")

		header := expires + row + remainingTopBorder
		headerHeight := lipgloss.Height(header) - 1 // ignore last \n
		contentHeight := m.height - windowStyle.GetVerticalFrameSize() - headerHeight
		contentWidth := m.width - windowStyle.GetHorizontalFrameSize()

		m.tabs[targetsView].SetSize(contentWidth, contentHeight)
		m.tabs[connectedView].SetSize(contentWidth, contentHeight)
		m.tabs[favoriteView].SetSize(contentWidth, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Top,
			header,
			windowStyle.
				Width(contentWidth).
				Height(contentHeight).
				Render(
					m.tabs[m.state].View(),
				),
		)
	}
}

func (m mainModel) messageUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.message = ""
		m.state = m.previousState
	}
	return m, nil
}

func (m *mainModel) terminateAllSessions() {
	for _, item := range m.tabs[connectedView].Items() {
		target := item.(*Target)
		TerminateSession(m.boundaryClient, target.session, target.task)
	}
}

func (m mainModel) quittingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.tabs[connectedView].Items()) == 0 {
		m.shouldQuit = true
		m.terminateAllSessions()
		saveFavoriteList(m.tabs[favoriteView])

		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// For simplicity's sake, we'll treat any key besides "y" as "no"
		if msg.String() == "y" {
			m.shouldQuit = true
			m.terminateAllSessions()
			return m, tea.Quit
		}
		m.shouldQuit = false
		m.state = m.previousState
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

func listKeyMap(keymap list.KeyMap) []key.Binding {
	var bindings []key.Binding
	v := reflect.ValueOf(keymap)
	for i := 0; i < v.NumField(); i++ {
		bindings = append(bindings, v.Field(i).Interface().(key.Binding))
	}

	return bindings
}

func Tui(targets *targets.TargetListResult, boundaryClient *api.Client, boundaryToken *authtokens.AuthToken) {
	tuiTargets := make([]list.Item, 0)

	for _, t := range targets.Items {
		tuiTargets = append(tuiTargets, &Target{title: fmt.Sprintf("%s (%s)", t.Name, t.Scope.Name), description: t.Description, target: t})
	}

	m := newModel(boundaryClient, boundaryToken)

	targetDelegate, targetKeyMap := newTargetDelegate()
	targetList := list.New(tuiTargets, targetDelegate, 0, 0)
	targetList.SetShowTitle(false)
	targetList.DisableQuitKeybindings()
	connectedDelegate, connectedKeyMap := newConnectedDelegate()
	connectedList := list.New([]list.Item{}, connectedDelegate, 0, 0)
	connectedList.SetShowTitle(false)
	connectedList.DisableQuitKeybindings()
	favoriteDelegate, favoriteKeyMap := newFavoriteDelegate()
	favoriteList := list.New([]list.Item{}, favoriteDelegate, 0, 0)
	favoriteList.SetShowTitle(false)
	favoriteList.DisableQuitKeybindings()

	err := loadFavoriteList(&favoriteList, tuiTargets)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	m.tabs = []list.Model{targetList, connectedList, favoriteList}
	m.tabsName = []string{"Targets", "Connected", "Favorites"}
	m.targetKeyMap = targetKeyMap
	m.connectedKeyMap = connectedKeyMap
	m.favoriteKeyMap = favoriteKeyMap

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithFilter(filter))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
