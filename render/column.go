package render

import (
	"iter"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render/table"
)

// DefaultColWidthRatio defines a default ration between full window width and
// a single column.
const DefaultColWidthRatio = 0.1

// Columns defines a custom type for a set of table columns where each column
// represented as an instance *Column.
type Columns []*Column

// Get returns an instance of *Column by its index in the set. The function
// returns two values, where the first value is the column instance or nil,
// and the second denotes whether the value was found.
func (c *Columns) Get(idx int) (*Column, bool) {
	if c == nil || idx < 0 || idx >= len(*c) {
		return nil, false
	}

	return (*c)[idx], true
}

// Visible returns a non-hidden subset of columns.
func (c *Columns) Visible(fullWith int) iter.Seq[*Column] {
	return func(yield func(*Column) bool) {
		for _, column := range *c {
			visible := column.Hidden == nil || !column.Hidden(fullWith)

			if visible && !yield(column) {
				break
			}
		}
	}
}

// TableColumns converts Column instances into a set of [table.Column] prepared
// for rendering. It calculates the width of each column based on its configuration.
//
// This conversion must be made on each window size change since some column
// widths are defined as a ratio between the full window width and the column
// width.
func (c *Columns) TableColumns(fullWidth int, sortState SortState) []table.Column {
	if c == nil || fullWidth == 0 {
		return nil
	}

	fullColCnt := 0

	// handle columns with fixed size and adjust the available window width for
	// the rest of the columns.
	for column := range c.Visible(fullWidth) {
		if column.Full {
			fullColCnt++
		}

		if column.Fixed && column.Width != 0 {
			fullWidth -= column.Width
		}
	}

	resultWidth := fullWidth

	// handle columns which size was defined as a ratio and adjust the available
	// window width for the rest of the columns.
	for column := range c.Visible(fullWidth) {
		if column.WidthRatio != 0 && !column.Fixed && !column.Full {
			column.Width = max(
				int(float64(fullWidth)*column.WidthRatio),
				column.MinWidth,
			)

			resultWidth -= column.Width
		}
	}

	// handle the columns that do not have a specific width, but will take all
	// the available space that was left. If there are multiple "full" columns
	// the available space will be evenly distributed between them.
	for column := range c.Visible(fullWidth) {
		if column.Full {
			column.Width = resultWidth / fullColCnt
		}
	}

	columns := make([]table.Column, len(*c))

	for i, column := range *c {
		if column.Hidden != nil && column.Hidden(fullWidth) {
			column.Width = 0
		}

		columns[i] = table.Column{
			Title: column.FmtName(sortState),
			Width: column.Width,
		}
	}

	return columns
}

// Column represents a single table column. It contains the column's title and
// configuration options used during the width calculation.
type Column struct {
	// Title contains the title that will be displayed in the table's header.
	Title string

	// SortKey contains the unique column's sorting key used during the sorting.
	SortKey drive.SortKey

	// Width contains predefined width value of column. If this value should not
	// change during the rendering the Fixed option must be set to true.
	Width int

	// MinWidth defines lower boundary fot the Width.
	MinWidth int

	// WidthRatio defines a ratio between the full table width and the column
	// width. The resulting Width property will be calculated based on it.
	WidthRatio float64

	// Fixed checks whether the column width can be changed during re-rendering.
	Fixed bool

	// Full defines whether the column width should occupy all available table
	// width. Such columns have the least priority, and their width will be
	// calculated after all others.
	Full bool

	// Hidden contains a condition that checks if the column should be visible.
	Hidden func(fullWidth int) bool
}

func (c *Column) FmtName(sortState SortState) string {
	var order string

	if len(sortState.Key) > 0 && sortState.Key == c.SortKey {
		order = " ▲"

		if sortState.Desc {
			order = " ▼"
		}
	}

	return c.Title + order
}

type SortState struct {
	Key  drive.SortKey
	Desc bool
}
