package target

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type connectMsg struct {
	*Target
}

type execErrorMsg struct {
	error
}

type terminateSessionMsg struct {
	*Target
}

type openShellMsg struct {
	Target
}

type targetKeyMap struct {
	shell   key.Binding
	connect key.Binding
}

func (c targetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.shell,
		c.connect,
	}
}

func (c targetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			c.shell,
			c.connect,
		},
	}
}

func newTargetKeyMap() *targetKeyMap {
	return &targetKeyMap{
		shell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open a shell"),
		),
		connect: key.NewBinding(
			key.WithKeys("c", "enter"),
			key.WithHelp("c/enter", "connect to target"),
		),
	}
}

func newTargetDelegate(model *mainModel) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	keys := newTargetKeyMap()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.shell):
				if i, ok := m.SelectedItem().(*Target); ok {
					task, cmd, session, _ := ConnectToTarget(i)

					i.session = session
					i.task = task
					// send connect event upstream
					if cmd == nil {
						// we are trying to connect to a target that we could not identify its type or does not have a client (e.g. HTTP)
						// just connect to it without opening a shell
						//TODO: show error message
						return tea.Sequence(
							func() tea.Msg { return connectMsg{i} },
						)
					} else {
						return tea.Sequence(
							tea.ExecProcess(
								cmd,
								func(err error) tea.Msg {
									if err != nil {
										return execErrorMsg{err}
									}
									return nil
								},
							),
							func() tea.Msg { return terminateSessionMsg{i} },
						)
					}
				}
			case key.Matches(msg, keys.connect):
				if i, ok := m.SelectedItem().(*Target); ok {
					// send connect event upstream
					return tea.Sequence(
						func() tea.Msg { return connectMsg{i} },
					)
				}
			}
		case connectMsg:
			task, _, session, _ := ConnectToTarget(msg.Target)
			msg.Target.session = session
			msg.Target.task = task
		}

		return nil
	}

	help := []key.Binding{keys.shell, keys.connect}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
