package command

import (
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/crumbyte/noxdir/command/archive"
	"github.com/crumbyte/noxdir/command/checksum"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	InputTextStyle    lipgloss.Style
	InputBarStyle     lipgloss.Style
	OutputStyle       lipgloss.Style
	ErrTextStyle      lipgloss.Style
	ExecTimeTextStyle lipgloss.Style
}

var DefaultStyles = Styles{
	InputTextStyle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#80ed99")),

	InputBarStyle: lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true),

	OutputStyle: lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true),

	ErrTextStyle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF303E")),

	ExecTimeTextStyle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EBBD34")),
}

type Model struct {
	input         textinput.Model
	viewport      viewport.Model
	styles        Styles
	entries       []string
	messages      []string
	onStateChange func()
	history       *History
	path          string
	locked        uint32
	enabled       bool
}

func NewModel(onStateChange func()) *Model {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = "$ "
	ti.Placeholder = "type the command..."

	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(12))
	vp.VisibleLineCount()

	m := &Model{
		styles:        DefaultStyles,
		input:         ti,
		viewport:      vp,
		onStateChange: onStateChange,
		history:       NewHistory(50),
		enabled:       false,
	}

	m.SetStyles(DefaultStyles)

	return m
}

func (m *Model) SetStyles(s Styles) {
	m.styles = s

	tiStyle := textinput.DefaultStyles(true)
	tiStyle.Focused.Prompt = m.styles.InputTextStyle
	tiStyle.Focused.Text = m.styles.InputTextStyle

	m.input.SetStyles(tiStyle)

	m.viewport.Style = m.styles.OutputStyle
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
		m.input.SetWidth(msg.Width)
		m.viewport.SetWidth(msg.Width)

		m.updateViewportMessages()

		return m, nil
	case tea.KeyPressMsg:
		if !m.enabled {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			m.executeCmd()
			m.updateViewportMessages()
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

func (m *Model) View() tea.View {
	if !m.enabled {
		return tea.View{}
	}

	var viewportContent string

	if len(m.messages) > 0 {
		viewportContent = m.viewport.View()
	}

	return tea.NewView(
		fmt.Sprintf(
			"%s\n%s",
			viewportContent,
			m.styles.InputBarStyle.Render(m.input.View()),
		),
	)
}

func (m *Model) updateViewportMessages() {
	if len(m.messages) == 0 {
		return
	}

	m.viewport.SetContent(
		lipgloss.NewStyle().Width(m.viewport.Width()).Render(
			strings.Join(m.messages, "\n"),
		),
	)

	m.viewport.GotoBottom()
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
	input := strings.TrimSpace(m.input.Value())

	if len(input) == 0 {
		return
	}

	defer m.history.Push(input)

	args := strings.Fields(input)
	args = append(args, "--entries", strings.Join(m.entries, ","))
	args = append(args, "--ctx-path", m.path)

	m.messages = append(m.messages, m.styles.InputTextStyle.Render("$ ", input))
	m.input.Reset()

	beforeExec := time.Now()

	rootCmd := NewRootCmd(
		m.onStateChange,
		archive.NewPackCmd,
		checksum.NewFileHashCmd,
	)

	err := Execute(rootCmd, args, outBuffer)
	if err != nil {
		m.messages = append(
			m.messages, m.styles.ErrTextStyle.Render(err.Error()),
		)

		return
	}

	took := m.styles.ExecTimeTextStyle.Render(
		"\ntook " + time.Since(beforeExec).String(),
	)

	output := outBuffer.String()

	if len(output) > 0 && output[len(output)-1] != '\n' {
		output += "\n"
	}

	m.messages = append(m.messages, output+took)
}
