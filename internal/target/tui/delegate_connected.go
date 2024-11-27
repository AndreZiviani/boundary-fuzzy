package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ConnectedKeyMap struct {
	Disconnect key.Binding
	Reconnect  key.Binding
	Info       key.Binding
	Favorite   key.Binding
}

func (c ConnectedKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		c.Disconnect,
		c.Reconnect,
		c.Info,
		c.Favorite,
	}
}

func (c ConnectedKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(),
	}
}

type connectedItemDelegate struct {
	ShowDescription bool
	Styles          list.DefaultItemStyles
	UpdateFunc      func(tea.Msg, *tea.Model) tea.Cmd
	ShortHelpFunc   func() []key.Binding
	FullHelpFunc    func() [][]key.Binding
	height          int
	spacing         int
}

func (cid connectedItemDelegate) Height() int                             { return 2 }
func (cid connectedItemDelegate) Spacing() int                            { return 1 }
func (cid connectedItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (cid connectedItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc  string
		port         int
		matchedRunes []int
		s            = &cid.Styles
	)

	if i, ok := item.(*Target); ok {
		title = i.Title()
		desc = i.Description()
		if i.session != nil {
			port = i.session.Port
		}
	} else {
		return
	}

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()

	if port != 0 {
		portStr := fmt.Sprintf("(%d)", port)
		titleWidth := textwidth - len(portStr)
		if len(title) > titleWidth {
			// Truncate title and add ellipsis
			titleWidth = titleWidth - len(ellipsis)
		}

		tmpTitle := ansi.Truncate(title, titleWidth, ellipsis)
		padding := textwidth - len(tmpTitle) - len(portStr)
		title = fmt.Sprintf("%s%s%s", tmpTitle, strings.Repeat(" ", padding), portStr)
	} else {
		title = ansi.Truncate(title, textwidth, ellipsis)
	}
	if cid.ShowDescription {
		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= cid.height-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, textwidth, ellipsis))
		}
		desc = strings.Join(lines, "\n")
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	if isFiltered && index < len(m.VisibleItems()) {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}

	if emptyFilter {
		title = s.DimmedTitle.Render(title)
		desc = s.DimmedDesc.Render(desc)
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			// Highlight matches
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		if isFiltered {
			// Highlight matches
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}

	if cid.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc) //nolint: errcheck
		return
	}
	fmt.Fprintf(w, "%s", title) //nolint: errcheck
}

func NewConnectedKeyMap() *ConnectedKeyMap {
	return &ConnectedKeyMap{
		Disconnect: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "disconnect from target"),
		),
		Reconnect: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reconnect to target"),
		),
		Info: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show session info"),
		),
		Favorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "add target to favorites"),
		),
	}
}

func NewConnectedDelegate() (list.ItemDelegate, *ConnectedKeyMap) {
	keys := NewConnectedKeyMap()

	help := []key.Binding{keys.Disconnect, keys.Reconnect, keys.Info, keys.Favorite}

	d := connectedItemDelegate{
		ShowDescription: true,
		Styles:          list.NewDefaultItemStyles(),
		ShortHelpFunc:   func() []key.Binding { return help },
		FullHelpFunc:    func() [][]key.Binding { return [][]key.Binding{help} },
		height:          2,
		spacing:         1,
	}
	return d, keys
}
