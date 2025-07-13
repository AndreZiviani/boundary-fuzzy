package tui

import (
	"fmt"
	"io"
	"maps"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type delegateUpdateFunc = func(t targetDelegate, msg tea.Msg, m *list.Model) tea.Cmd
type delegateRenderFunc = func(t targetDelegate, w io.Writer, m list.Model, index int, item list.Item)

type DelegateKeyMap struct {
	binding map[string]key.Binding
	keys    []string
	values  []key.Binding
}

func (c DelegateKeyMap) ShortHelp() []key.Binding {
	return c.values
}

func (c DelegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(),
	}
}

func NewDelegate(keyMap map[string]key.Binding, tab sessionState, updater delegateUpdateFunc, renderer delegateRenderFunc) (*targetDelegate, *DelegateKeyMap) {
	keys := make([]string, 0, len(keyMap))
	values := make([]key.Binding, 0, len(keyMap))
	for k, v := range keyMap {
		keys = append(keys, k)
		values = append(values, v)
	}

	km := &DelegateKeyMap{
		binding: maps.Clone(keyMap),
		keys:    keys,
		values:  values,
	}

	d := &targetDelegate{
		ShowDescription: true,
		Styles:          list.NewDefaultItemStyles(),
		ShortHelpFunc:   func() []key.Binding { return values },
		FullHelpFunc:    func() [][]key.Binding { return [][]key.Binding{values} },
		height:          2,
		spacing:         1,
		UpdateFunc:      updater,
		RenderFunc:      renderer,
		keyMap:          km,

		tab: tab,
	}

	return d, km
}

type targetDelegate struct {
	ShowDescription bool
	Styles          list.DefaultItemStyles
	UpdateFunc      delegateUpdateFunc
	RenderFunc      delegateRenderFunc
	ShortHelpFunc   func() []key.Binding
	FullHelpFunc    func() [][]key.Binding
	height          int
	spacing         int

	tab    sessionState
	keyMap *DelegateKeyMap
}

func (td targetDelegate) Height() int {
	if td.ShowDescription {
		return td.height
	}
	return 1
}

func (td targetDelegate) Spacing() int {
	return td.spacing
}

func (td targetDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	if td.UpdateFunc != nil {
		return td.UpdateFunc(td, msg, m)
	}
	return nil
}

func (td targetDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if td.RenderFunc != nil {
		td.RenderFunc(td, w, m, index, item)
		return
	}

	var (
		s = &td.Styles
	)

	target, ok := item.(*Target)
	if !ok {
		return
	}

	title, rtitle := target.Title(td.tab)
	desc, rdesc := target.Description(td.tab)

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()

	if rtitle != "" {
		titleWidth := textwidth - len(rtitle)
		if len(title) > titleWidth {
			// Truncate title and add ellipsis
			titleWidth = titleWidth - len(ellipsis)
		}

		tmpTitle := ansi.Truncate(title, titleWidth, ellipsis)
		padding := textwidth - len(tmpTitle) - len(rtitle)
		title = fmt.Sprintf("%s%s%s", tmpTitle, strings.Repeat(" ", padding), rtitle)
	} else {
		title = ansi.Truncate(title, textwidth, ellipsis)
	}

	if td.ShowDescription {
		descWidth := textwidth
		if rdesc != "" {
			descWidth = descWidth - len(rdesc)
		}

		if len(desc) > descWidth {
			// Truncate description and add ellipsis
			descWidth = descWidth - len(ellipsis)
		}

		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= td.height-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, descWidth, ellipsis))
		}
		nlines := len(lines)
		lastLine := lines[nlines-1]

		padding := textwidth - len(lastLine) - len(rdesc)
		lines[nlines-1] = fmt.Sprintf("%s%s%s", lastLine, strings.Repeat(" ", padding), rdesc)
		desc = strings.Join(lines, "\n")
	}

	title, desc = renderCustomItem(m, index, item, title, desc, s)

	if td.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc) //nolint: errcheck
		return
	}
	fmt.Fprintf(w, "%s", title) //nolint: errcheck
}

func renderCustomItem(m list.Model, index int, item list.Item, title, desc string, s *list.DefaultItemStyles) (string, string) {
	// Conditions
	var (
		matchedRunes []int
		isSelected   = index == m.Index()
		emptyFilter  = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered   = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied

		titleStyle lipgloss.Style
		descStyle  lipgloss.Style
		active     = true
	)

	if target, ok := item.(*Target); ok {
		active = target.IsConnected()
	}

	if isFiltered && index < len(m.VisibleItems()) {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}
	if emptyFilter {
		titleStyle = s.DimmedTitle
		descStyle = s.DimmedDesc
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			// Highlight matches
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		titleStyle = s.SelectedTitle
		descStyle = s.SelectedDesc
	} else {
		if isFiltered {
			// Highlight matches
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		titleStyle = s.NormalTitle
		descStyle = s.NormalDesc
	}

	// If the title is "connected", apply strikethrough if not active
	if m.Title == connectedTabName {
		if !active {
			titleStyle = titleStyle.Strikethrough(true).Foreground(s.DimmedTitle.GetForeground())
		}
	}

	return titleStyle.Render(title), descStyle.Render(desc)
}
