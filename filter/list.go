package filter

import (
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
)

// ID defines a custom type for the filter identifier.
type ID string

// Reset resets the filter's state and disables it.
type Reset interface {
	Reset()
}

// Toggler enables or disables the filter.
type Toggler interface {
	Toggle()
}

// EntryFilter defines a contract for filtering a single *structure.Entry instance.
type EntryFilter interface {
	// ID returns a filter identifier allowing to uniquely identify the filter.
	ID() ID

	// Filter contains the filtering logic and returns a bool value on whether
	// the *Entry instance passed the filtration.
	Filter(e *structure.Entry) bool
}

// Updater defines an interface for filters that require their state to be updated
// during application rendering, e.g., a name filter.
type Updater interface {
	Update(msg tea.Msg)
}

// Viewer defines an interface for filters that can be rendered or require the
// user's input to interact.
type Viewer interface {
	View() string
}

// FiltersList aggregates a list of multiple filters.
type FiltersList map[ID]EntryFilter

// NewFiltersList creates a new list of filters based on the provided argument.
// All filters should be disabled/inactive by default.
func NewFiltersList(ef ...EntryFilter) FiltersList {
	fl := make(FiltersList, len(ef))

	for _, e := range ef {
		fl[e.ID()] = e
	}

	return fl
}

// Valid checks the provided *Entry instance against all filters and returns a
// bool value. The "true" value will be returned if the entry passes all defined
// filters.
func (fl *FiltersList) Valid(e *structure.Entry) bool {
	for _, filter := range *fl {
		if !filter.Filter(e) {
			return false
		}
	}

	return true
}

// ToggleFilter enables or disables the provided filter by the provided unique
// filter name. Unknown filters that were not added to the list will be ignored.
func (fl *FiltersList) ToggleFilter(id ID) {
	if _, ok := (*fl)[id]; !ok {
		return
	}

	if t, ok := (*fl)[id].(Toggler); ok {
		t.Toggle()
	}
}

// Reset traverses all filters in the list and calls EntryFilter.Reset method for
// each of them.
func (fl *FiltersList) Reset() {
	for _, filter := range *fl {
		if r, ok := filter.(Reset); ok {
			r.Reset()
		}
	}
}

// Update traverses all filters implementing the Updater interface in order to
// update the filters' state based on incoming input.
func (fl *FiltersList) Update(msg tea.Msg) {
	for _, f := range *fl {
		if u, ok := f.(Updater); ok {
			u.Update(msg)
		}
	}
}
