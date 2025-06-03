package filter

import (
	"strings"

	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	DirsOnlyFilterID  ID = "DirsOnly"
	FilesOnlyFilterID ID = "FilesOnly"
	NameFilterID      ID = "NameFilter"
	EmptyDirFilterID  ID = "EmptyDirFilter"
)

// DirsFilter filters *Entry by its type and allows directories only.
type DirsFilter struct {
	enabled bool
}

func (df *DirsFilter) ID() ID {
	return DirsOnlyFilterID
}

func (df *DirsFilter) Toggle() {
	df.enabled = !df.enabled
}

func (df *DirsFilter) Filter(e *structure.Entry) bool {
	return !df.enabled || e.IsDir
}

func (df *DirsFilter) Reset() {
	df.enabled = false
}

// FilesFilter filters *Entry by its type and allows files only.
type FilesFilter struct {
	enabled bool
}

func (df *FilesFilter) ID() ID {
	return FilesOnlyFilterID
}

func (df *FilesFilter) Toggle() {
	df.enabled = !df.enabled
}

func (df *FilesFilter) Reset() {
	df.enabled = false
}

func (df *FilesFilter) Filter(e *structure.Entry) bool {
	return !df.enabled || !e.IsDir
}

// EmptyDirFilter filters empty directories. It checks the total number of files,
// including those in subdirectories, and discards it if it does not have any.
//
// The filter does not affect file *Entry instances.
type EmptyDirFilter struct{}

func (edf *EmptyDirFilter) ID() ID {
	return EmptyDirFilterID
}

func (edf *EmptyDirFilter) Toggle() {}

// Reset ...
// TODO: interface segregation required
func (edf *EmptyDirFilter) Reset() {}

func (edf *EmptyDirFilter) Filter(e *structure.Entry) bool {
	return !e.IsDir || e.TotalFiles > 0
}

// NameFilter filters a single instance of the *structure.Entry by its path value.
// If the entry's path value does not contain the user's input, it will not be
// filtered/discarded.
//
// The user's input is handled by the textinput.Model instance, therefore the
// filter must update internal state by providing the corresponding Updater
// implementation.
type NameFilter struct {
	input   textinput.Model
	enabled bool
}

func NewNameFilter(placeholder string) *NameFilter {
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbd34"))
	ti := textinput.New()

	ti.Placeholder = placeholder
	ti.Focus()
	ti.Width = lipgloss.Width(placeholder)
	ti.Prompt = "\uE68F  "
	ti.PromptStyle, ti.TextStyle = textStyle, textStyle

	return &NameFilter{input: ti, enabled: false}
}

func (nf *NameFilter) ID() ID {
	return NameFilterID
}

func (nf *NameFilter) Toggle() {
	nf.enabled = !nf.enabled
}

// Filter filters an instance of *structure.Entry by checking if its path value
// contains the current filter input.
func (nf *NameFilter) Filter(e *structure.Entry) bool {
	return strings.Contains(
		strings.ToLower(e.Name()),
		strings.ToLower(nf.input.Value()),
	)
}

func (nf *NameFilter) Update(msg tea.Msg) {
	resizeMsg, ok := msg.(tea.WindowSizeMsg)
	if ok {
		nf.input.Width = resizeMsg.Width
	}

	if !nf.enabled {
		return
	}

	nf.input, _ = nf.input.Update(msg)
}

func (nf *NameFilter) Reset() {
	nf.enabled = false
	nf.input.Reset()
}

func (nf *NameFilter) View() string {
	if !nf.enabled {
		return ""
	}

	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true)

	return s.Render(nf.input.View())
}
