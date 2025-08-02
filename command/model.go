package command

import (
	"bytes"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	input         textinput.Model
	entries       []string
	onStateChange func()
	history       *History
	path          string
	locked        uint32
	enabled       bool
}

func NewModel(onStateChange func()) *Model {
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#80ed99"))

	ti := textinput.New()
	ti.Focus()
	ti.Prompt = "$ "
	ti.PromptStyle, ti.TextStyle = textStyle, textStyle

	return &Model{
		input:         ti,
		onStateChange: onStateChange,
		history:       NewHistory(50),
		enabled:       false,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetPathContext(path string, entries []string) {
	m.path, m.entries = path, entries
}

func (m *Model) Enable() {
	m.input.Reset()
	m.enabled = true
}

func (m *Model) Enabled() bool {
	return m.enabled
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.input.Width = msg.Width

		return m, nil
	case tea.KeyMsg:
		if !m.enabled {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			m.executeCmd()
		case "ctrl+c":
			m.input.Reset()
			m.history.ResetCursor()
		case "up":
			prevCmd, ok := m.history.Prev()
			if ok {
				m.input.SetValue(prevCmd)
				m.input.SetCursor(len(prevCmd))
			}
		case "down":
			nextCmd, ok := m.history.Next()
			if ok {
				m.input.SetValue(nextCmd)
				m.input.SetCursor(len(nextCmd))
			}
		case "esc":
			m.enabled = false
			m.input.Reset()
			m.history.ResetCursor()

			return m, nil
		}
	}

	m.input, _ = m.input.Update(msg)

	return m, nil
}

func (m *Model) View() string {
	if !m.enabled {
		return ""
	}

	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true)

	return s.Render(m.input.View())
}

func (m *Model) executeCmd() {
	if atomic.SwapUint32(&m.locked, 1) == 1 {
		return
	}

	defer atomic.SwapUint32(&m.locked, 0)

	if m.entries == nil {
		m.input.Placeholder = "no entries selected"

		return
	}

	outBuffer := bytes.NewBuffer(nil)
	input := m.input.Value()

	defer m.history.Push(input)

	args := strings.Fields(input)
	args = append(args, "--entries", strings.Join(m.entries, ","))
	args = append(args, "--ctx-path", m.path)

	m.input.Reset()

	if err := Execute(NewRootCmd(), args, outBuffer); err != nil {
		m.input.Placeholder = err.Error()

		return
	}

	m.input.Placeholder = outBuffer.String()
	m.onStateChange()
}
