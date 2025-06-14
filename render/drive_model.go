package render

import (
	"strconv"
	"strings"

	"github.com/crumbyte/noxdir/drive"

	"git.sr.ht/~mpldr/go-indicators/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const driveSizeWidth = 10

type DriveModel struct {
	driveColumns []Column
	drivesTable  *table.Model
	nav          *Navigation
	sortState    SortState
	height       int
	width        int
	fullHelp     bool
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
		dm.height, dm.width = msg.Height, msg.Width

		dm.drivesTable.SetHeight(msg.Height)
		dm.drivesTable.SetWidth(msg.Width)

		dm.updateTableData(dm.sortState.Key, dm.sortState.Desc)

		return dm, nil
	case tea.KeyMsg:
		bk := bindingKey(strings.ToLower(msg.String()))

		switch bk {
		case toggleHelp:
			dm.fullHelp = !dm.fullHelp
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
	h := lipgloss.Height
	summary := dm.drivesSummary()
	keyBindings := dm.drivesTable.Help.ShortHelpView(shortHelp)

	if dm.fullHelp {
		keyBindings = dm.drivesTable.Help.FullHelpView(
			append(navigateKeyMap, drivesKeyMap...),
		)
	}

	dm.drivesTable.SetHeight(dm.height - h(keyBindings) - h(summary)*2)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		summary,
		dm.drivesTable.View(),
		summary,
		keyBindings,
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

	progBar.Width = progressWidth - 2
	if progBar.Width < 0 {
		progBar.Width = 0
	}

	allDrives := dm.nav.DrivesList().Sort(key, sortDesc)
	rows := make([]table.Row, 0, len(allDrives))

	for _, d := range allDrives {
		pgBar := progBar.GetBar(d.UsedPercent, 100.0)
		rows = append(rows, table.Row{
			"⛃",
			d.Path,
			d.Volume,
			d.FSName,
			FmtSize(d.TotalBytes, driveSizeWidth),
			FmtSize(d.UsedBytes, driveSizeWidth),
			FmtSize(d.FreeBytes, driveSizeWidth),
			strconv.FormatFloat(d.UsedPercent, 'f', 2, 64) + " %",
			"▕" + pgBar + "▏",
		})
	}

	dm.drivesTable.SetRows(rows)
	dm.drivesTable.SetCursor(0)
}

func (dm *DriveModel) drivesSummary() string {
	dl := dm.nav.DrivesList()

	items := []*BarItem{
		NewBarItem(Version, "#8338ec", 0),
		NewBarItem("MODE", "#FF5F87", 0),
		NewBarItem("Drives List", "", -1),
		NewBarItem("CAPACITY", "#FF5F87", 0),
		DefaultBarItem(FmtSize(dl.TotalCapacity, 0)),
		NewBarItem("FREE", "#FF5F87", 0),
		DefaultBarItem(FmtSize(dl.TotalFree, 0)),
		NewBarItem("USED", "#FF5F87", 0),
		DefaultBarItem(FmtSize(dl.TotalUsed, 0)),
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

var (
	progBar = &progress.Progress{
		Width: 0,
	}
	barStyle = lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(9)).
			Background(lipgloss.ANSIColor(10))
)

func init() {
	progBar.SetStyle("block")
}
