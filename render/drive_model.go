package render

import (
	"strconv"
	"strings"

	"github.com/crumbyte/noxdir/drive"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DriveModel struct {
	driveColumns []Column
	drivesTable  *table.Model
	nav          *Navigation
	sortState    SortState
	width        int
}

func NewDriveModel(n *Navigation) *DriveModel {
	dc := []Column{
		{},
		{Title: "Path"},
		{Title: "Volume Name"},
		{Title: "File System"},
		{Title: "Total Space", SortKey: drive.TotalCap},
		{Title: "Used Space", SortKey: drive.TotalUsed},
		{Title: "Free Space", SortKey: drive.TotalFree},
		{Title: "Usage", SortKey: drive.TotalUsedP},
		{},
	}

	return &DriveModel{
		nav:          n,
		driveColumns: dc,
		sortState:    SortState{Key: drive.TotalUsedP, Desc: true},
		drivesTable:  buildTable(),
	}
}

func (dm *DriveModel) Init() tea.Cmd {
	return nil
}

func (dm *DriveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dm.width = msg.Width

		dm.drivesTable.SetHeight(msg.Height - 10)
		dm.drivesTable.SetWidth(msg.Width)

		dm.updateTableData(dm.sortState.Key, dm.sortState.Desc)

		return dm, nil
	case tea.KeyMsg:
		bk := bindingKey(strings.ToLower(msg.String()))

		switch bk {
		case sortTotalCap, sortTotalUsed, sortTotalFree, sortTotalUsedP:
			dm.sortDrives(
				drive.SortKey(strings.TrimPrefix(msg.String(), "alt+")),
			)

			return dm, nil
		case explore:
			sr := dm.drivesTable.SelectedRow()
			if len(sr) < 2 {
				return dm, nil
			}

			if err := dm.nav.Explore(sr[1]); err != nil {
				return dm, nil
			}
		}
	}

	if !dm.nav.OnDrives() {
		return dm, nil
	}

	t, _ := dm.drivesTable.Update(msg)
	dm.drivesTable = &t

	return dm, nil
}

func (dm *DriveModel) View() string {
	summary := dm.drivesSummary()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		summary,
		dm.drivesTable.View(),
		summary,
		dm.drivesTable.Help.FullHelpView(
			append(navigateKeyMap, drivesKeyMap...),
		),
	)
}

func (dm *DriveModel) updateTableData(key drive.SortKey, sortDesc bool) {
	pathWidth, iconWidth := 30, 5
	tableWidth := dm.width

	colWidth := int(float64(tableWidth) * 0.07)
	progressWidth := tableWidth - (colWidth * 6) - iconWidth - pathWidth

	columns := make([]table.Column, len(dm.driveColumns))

	for i, c := range dm.driveColumns {
		columns[i] = table.Column{
			Title: c.FmtName(dm.sortState),
			Width: colWidth,
		}
	}

	columns[0].Width = iconWidth
	columns[1].Width = pathWidth
	columns[len(columns)-1].Width = progressWidth

	dm.drivesTable.SetColumns(columns)
	dm.drivesTable.SetCursor(0)

	diskFillProgress := NewProgressBar(progressWidth, '🟥', '🟩')

	allDrives := dm.nav.DrivesList().Sort(key, sortDesc)
	rows := make([]table.Row, 0, len(allDrives))

	for _, d := range allDrives {
		pgBar := diskFillProgress.ViewAs(d.UsedPercent / 100)
		rows = append(rows, table.Row{
			"⛃",
			d.Path,
			d.Volume,
			d.FSName,
			fmtSize(d.TotalBytes, true),
			fmtSize(d.UsedBytes, true),
			fmtSize(d.FreeBytes, true),
			strconv.FormatFloat(d.UsedPercent, 'f', 2, 64) + " %",
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				strings.Repeat(" ", progressWidth-lipgloss.Width(pgBar)),
				pgBar,
			),
		})
	}

	dm.drivesTable.SetRows(rows)
	dm.drivesTable.SetCursor(0)
}

func (dm *DriveModel) drivesSummary() string {
	dl := dm.nav.DrivesList()

	items := []*BarItem{
		NewBarItem("MODE", "#FF5F87", 0),
		NewBarItem("Drives List", "", -1),
		NewBarItem("CAPACITY", "#FF5F87", 0),
		DefaultBarItem(fmtSize(dl.TotalCapacity, false)),
		NewBarItem("FREE", "#FF5F87", 0),
		DefaultBarItem(fmtSize(dl.TotalFree, false)),
		NewBarItem("USED", "#FF5F87", 0),
		DefaultBarItem(fmtSize(dl.TotalUsed, false)),
	}

	return statusBarStyle.Margin(1, 0, 1, 0).
		Render(NewStatusBar(items, dm.width))
}

func (dm *DriveModel) sortDrives(sortKey drive.SortKey) {
	if !dm.nav.OnDrives() {
		return
	}

	if dm.sortState.Key == sortKey {
		dm.sortState.Desc = !dm.sortState.Desc
	} else {
		dm.sortState = SortState{Key: sortKey}
	}

	dm.updateTableData(
		dm.sortState.Key,
		dm.sortState.Desc,
	)
}

func (dm *DriveModel) resetSort() {
	dm.sortState = SortState{Key: drive.TotalUsedP, Desc: false}
}
