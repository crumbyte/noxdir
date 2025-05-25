package render

import "github.com/crumbyte/noxdir/drive"

type Column struct {
	Title   string
	SortKey drive.SortKey
	Width   int
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
