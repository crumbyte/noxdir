package filter

import (
	"regexp"
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

func (edf *EmptyDirFilter) Filter(e *structure.Entry) bool {
	return !e.IsDir || e.TotalFiles > 0
}

// NameFilterType represents a filter type that will be applied during the
// filtering process.
type NameFilterType int

const (
	// RegularNameFilter represents a default filter type where the filter value
	// must be a substring of the original text.
	RegularNameFilter NameFilterType = iota

	// NegativeNameFilter represents a filter type with the behavior opposite to
	// RegularNameFilter. It's enabled by the "\" backslash prefix at the beginning
	// of the filter input.
	NegativeNameFilter

	// RegexNameFilter represents a regular expression filter type where the
	// filter input will be treated as a valid regular expression. It's enabled
	// by the ":" colon prefix at the beginning of the filter input.
	RegexNameFilter
)

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

func NewNameFilter(textColor string) *NameFilter {
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))
	ti := textinput.New()

	ti.Placeholder = `Filter… Examples: "mp4" (match), "\mp4" (exclude), ":regex" ":^.+?\.mp4" (regular expression)`
	ti.Focus()
	ti.Width = lipgloss.Width(ti.Placeholder)
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
	filterValue, filtered := nf.input.Value(), false

	filterType := nf.resolveFilterType(filterValue)

	switch filterType {
	case RegularNameFilter:
		filtered = strings.Contains(
			strings.ToLower(e.Name()),
			strings.ToLower(filterValue),
		)
	case NegativeNameFilter:
		filtered = !strings.Contains(
			strings.ToLower(e.Name()),
			strings.ToLower(filterValue[1:]),
		)
	case RegexNameFilter:
		regexFilter, err := regexp.Compile(filterValue[1:])
		if err != nil {
			return false
		}

		filtered = regexFilter.MatchString(e.Name())
	}

	return filtered
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

func (nf *NameFilter) resolveFilterType(filterInput string) NameFilterType {
	resolvedType := RegularNameFilter

	if len(filterInput) < 1 {
		return resolvedType
	}

	switch filterInput[0] {
	case '\\':
		resolvedType = NegativeNameFilter
	case ':':
		resolvedType = RegexNameFilter
	}

	return resolvedType
}
