// Package table is an altered version of the original bubbles package for the
// table rendering https://github.com/charmbracelet/bubbles/tree/master/table.
package table

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Model struct {
	KeyMap   KeyMap
	styles   Styles
	cols     []Column
	Help     help.Model
	rows     []Row
	viewport viewport.Model
	marked   map[int]struct{}
	cursor   int
	start    int
	end      int
}

type Row []string

type Column struct {
	Title string
	Width int
}

type KeyMap struct {
	LineUp     key.Binding
	LineDown   key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
	MarkRow    key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", " "),
			key.WithHelp("f/pgdn", "page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		MarkRow: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "mark row"),
		),
	}
}

type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
	Marked   lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Marked:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

func New() Model {
	m := Model{
		cursor:   0,
		viewport: viewport.New(0, 20),
		KeyMap:   DefaultKeyMap(),
		Help:     help.New(),
		styles:   DefaultStyles(),
		marked:   make(map[int]struct{}),
	}

	m.UpdateViewport()

	return m
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return *m, nil
	}

	switch {
	case key.Matches(keyMsg, m.KeyMap.LineUp):
		m.MoveUp(1)
	case key.Matches(keyMsg, m.KeyMap.LineDown):
		m.MoveDown(1)
	case key.Matches(keyMsg, m.KeyMap.PageUp):
		m.MoveUp(m.viewport.Height)
	case key.Matches(keyMsg, m.KeyMap.PageDown):
		m.MoveDown(m.viewport.Height)
	case key.Matches(keyMsg, m.KeyMap.GotoTop):
		m.GotoTop()
	case key.Matches(keyMsg, m.KeyMap.GotoBottom):
		m.GotoBottom()
	case key.Matches(keyMsg, m.KeyMap.MarkRow):
		m.MarkSelected()
	}

	return *m, nil
}

func (m *Model) View() string {
	return m.headersView() + "\n" + m.viewport.View()
}

func (m *Model) UpdateViewport() {
	renderedRows := make([]string, 0, len(m.rows))

	if m.cursor >= 0 {
		m.start = clamp(m.cursor-m.viewport.Height, 0, m.cursor)
	} else {
		m.start = 0
	}
	m.end = clamp(m.cursor+m.viewport.Height, m.cursor, len(m.rows))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

func (m *Model) SelectedRow() Row {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}

	return m.rows[m.cursor]
}

func (m *Model) Rows() []Row {
	return m.rows
}

func (m *Model) MarkedRows() []Row {
	rows := make([]Row, 0, len(m.marked))

	for i := range m.rows {
		if _, ok := m.marked[i]; ok {
			rows = append(rows, m.rows[i])
		}
	}

	return rows
}

func (m *Model) Columns() []Column {
	return m.cols
}

func (m *Model) SetRows(r []Row) {
	m.rows = r
	m.UpdateViewport()
}

func (m *Model) SetColumns(c []Column) {
	m.cols = c
	m.UpdateViewport()
}

func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.UpdateViewport()
}

func (m *Model) Height() int {
	return m.viewport.Height
}

func (m *Model) Width() int {
	return m.viewport.Width
}

func (m *Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetCursor(n int) {
	m.cursor = clamp(n, 0, len(m.rows)-1)
	m.UpdateViewport()
}

func (m *Model) MarkSelected() {
	cursor := clamp(m.cursor, 0, len(m.rows)-1)

	if _, ok := m.marked[cursor]; ok {
		delete(m.marked, cursor)
	} else {
		m.marked[clamp(m.cursor, 0, len(m.rows)-1)] = struct{}{}
	}

	m.UpdateViewport()
}

func (m *Model) ResetMarked() {
	m.marked = make(map[int]struct{})
}

func (m *Model) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, len(m.rows)-1)

	switch {
	case m.start == 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset, 0, m.cursor))
	case m.start < m.viewport.Height:
		m.viewport.YOffset = clamp(
			clamp(m.viewport.YOffset+n, 0, m.cursor),
			0,
			m.viewport.Height,
		)
	case m.viewport.YOffset >= 1:
		m.viewport.YOffset = clamp(m.viewport.YOffset+n, 1, m.viewport.Height)
	}

	m.UpdateViewport()
}

func (m *Model) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, len(m.rows)-1)
	m.UpdateViewport()

	switch {
	case m.end == len(m.rows) && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.viewport.Height))
	case m.cursor > (m.end-m.start)/2 && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.cursor))
	case m.viewport.YOffset > 1:
	case m.cursor > m.viewport.YOffset+m.viewport.Height-1:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset+1, 0, 1))
	}
}

func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

func (m *Model) GotoBottom() {
	m.MoveDown(len(m.rows))
}

func (m *Model) headersView() string {
	cols := make([]string, 0, len(m.cols))

	style := lipgloss.NewStyle().Inline(true)

	for _, col := range m.cols {
		if col.Width <= 0 {
			continue
		}

		style = style.Width(col.Width).MaxWidth(col.Width)

		renderedCell := style.Render(
			runewidth.Truncate(col.Title, col.Width, "…"),
		)

		cols = append(cols, m.styles.Header.Render(renderedCell))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func (m *Model) renderRow(r int) string {
	cols := make([]string, 0, len(m.cols))

	for i, value := range m.rows[r] {
		if m.cols[i].Width <= 0 {
			continue
		}

		renderer := m.styles.Cell

		if _, ok := m.marked[r]; ok {
			renderer = m.styles.Marked
		}

		if r == m.cursor {
			renderer = m.styles.Selected
		}

		cell := lipgloss.NewStyle().
			Width(m.cols[i].Width).
			MaxWidth(m.cols[i].Width).
			Inline(true).
			Render(value)

		cols = append(cols, renderer.Render(cell))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
