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

func NewDelegate(keyMap map[string]key.Binding, tab sessionState) (*targetDelegate, *DelegateKeyMap) {
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

		tab: tab,
	}

	return d, km
}

type targetDelegate struct {
	ShowDescription bool
	Styles          list.DefaultItemStyles
	UpdateFunc      func(tea.Msg, *list.Model) tea.Cmd
	ShortHelpFunc   func() []key.Binding
	FullHelpFunc    func() [][]key.Binding
	height          int
	spacing         int

	tab sessionState
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
	if td.UpdateFunc == nil {
		return nil
	}
	return td.UpdateFunc(msg, m)
}

func (td targetDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		matchedRunes []int
		s            = &td.Styles
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

	if td.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc) //nolint: errcheck
		return
	}
	fmt.Fprintf(w, "%s", title) //nolint: errcheck
}
