package render

import (
	"runtime"
	"strings"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render/table"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const driveSizeWidth = 10

type RefreshDrives struct{}

type DriveModel struct {
	columns     Columns
	drivesTable *table.Model
	nav         *Navigation
	usagePG     *PG
	statusBar   *StatusBar
	sortState   SortState
	height      int
	width       int
	fullHelp    bool
}

func NewDriveModel(n *Navigation) *DriveModel {
	pathRatio := DefaultColWidthRatio

	if runtime.GOOS == "darwin" {
		pathRatio = 0.17
	}

	dc := Columns{
		{Width: 5, Fixed: true},
		{Hidden: func(_ int) bool { return true }},
		{
			Title:      "Path",
			WidthRatio: pathRatio,
			MinWidth:   20,
		},
		{
			Title:      "Volume",
			WidthRatio: pathRatio,
			MinWidth:   20,
			Hidden:     func(_ int) bool { return runtime.GOOS == "darwin" },
		},
		{
			Title:      "File System",
			Hidden:     func(fw int) bool { return fw < 150 },
			WidthRatio: DefaultColWidthRatio,
		},
		{
			Title:      "Total",
			WidthRatio: DefaultColWidthRatio,
			SortKey:    drive.TotalCap,
			MinWidth:   12,
		},
		{
			Title:      "Used",
			WidthRatio: DefaultColWidthRatio,
			SortKey:    drive.TotalUsed,
			MinWidth:   12,
		},
		{
			Title:      "Free",
			WidthRatio: DefaultColWidthRatio,
			SortKey:    drive.TotalFree,
			MinWidth:   12,
		},
		{
			Title:      "Usage",
			WidthRatio: DefaultColWidthRatio,
			SortKey:    drive.TotalUsedP,
			MinWidth:   12,
		},
		{Full: true},
	}

	if runtime.GOOS == "linux" {
		dc[2].Title = "Device"
		dc[3].Title = "Mount"
	}

	return &DriveModel{
		nav:         n,
		columns:     dc,
		sortState:   SortState{Key: drive.TotalUsedP, Desc: true},
		drivesTable: buildTable(),
		statusBar:   NewStatusBar(),
		usagePG:     &style.CS().UsageProgressBar,
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
		switch {
		case key.Matches(msg, Bindings.Help):
			dm.fullHelp = !dm.fullHelp
		case key.Matches(msg, Bindings.Drive.SortKeys):
			dm.sortDrives(drive.SortKey(msg.String()))

			return dm, nil
		case key.Matches(msg, Bindings.Explore):
			sr := dm.drivesTable.SelectedRow()
			if len(sr) < 2 {
				return dm, nil
			}

			if err := dm.nav.Explore(sr[1]); err != nil {
				return dm, nil
			}
		}
	case RefreshDrives:
		dm.updateTableData(dm.sortState.Key, dm.sortState.Desc)
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
	keyBindings := dm.drivesTable.Help.ShortHelpView(
		Bindings.ShortBindings(),
	)

	if dm.fullHelp {
		keyBindings = dm.drivesTable.Help.FullHelpView(
			Bindings.DriveBindings(),
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
	pathCol, _ := dm.columns.Get(2)
	usageCol, _ := dm.columns.Get(8)
	pgCol, _ := dm.columns.Get(9)

	dm.drivesTable.SetColumns(dm.columns.TableColumns(dm.width, dm.sortState))
	dm.drivesTable.SetCursor(0)

	diskFillProgress := dm.usagePG.New(pgCol.Width)

	drivesList := dm.nav.DrivesList()
	sortedDrives := drivesList.Sort(key, sortDesc)

	rows := make([]table.Row, 0, len(sortedDrives))

	for _, d := range sortedDrives {
		pgBar := diskFillProgress.ViewAs(d.UsedPercent / 100)
		r := table.Row{
			"⛃",
			d.Path,
			WrapString(d.Path, pathCol.Width),
			d.Volume,
			d.FSName,
			FmtSize(d.TotalBytes, driveSizeWidth),
			FmtSize(d.UsedBytes, driveSizeWidth),
			FmtSize(d.FreeBytes, driveSizeWidth),
			FmtUsage(d.UsedPercent/100, 80, usageCol.Width),
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				strings.Repeat(" ", max(0, pgCol.Width-lipgloss.Width(pgBar))),
				pgBar,
			),
		}

		if !drivesList.MountsLayout {
			rows = append(rows, r)

			continue
		}

		// update the row layout for the Linux-based systems. Each device and
		// all the corresponding mounts will be rendered according to the
		// specified sorting rule.
		if d.IsDev != 0 {
			r[1], r[2], r[3], r[4] = "", d.Device, "", "-"
		} else {
			r = table.Row{
				"⤷",
				d.Path,
				"",
				WrapString(d.Path, pathCol.Width),
				d.FSName,
				"-", "-", "-", "-", "",
			}
		}

		rows = append(rows, r)
	}

	dm.drivesTable.SetRows(rows)
	dm.drivesTable.SetCursor(0)
}

func (dm *DriveModel) drivesSummary() string {
	if dm.statusBar == nil {
		return ""
	}

	dm.statusBar.Clear()

	driveTitle := "No Drives Selected"
	sbStyle := style.CS().StatusBar

	if len(dm.drivesTable.Rows()) != 0 && dm.drivesTable.SelectedRow()[1] != "" {
		driveTitle = dm.drivesTable.SelectedRow()[1]
	}

	barItems := []*BarItem{
		{Content: Version, BGColor: sbStyle.VersionBG},
		{Content: "DRIVES", BGColor: sbStyle.Drives.ModeBG},
		{Content: driveTitle, BGColor: sbStyle.BG, Width: -1},
	}

	if dm.nav.cacheEnabled {
		barItems = append(
			barItems,
			&BarItem{
				Content: "CACHED",
				BGColor: sbStyle.VersionBG,
			},
		)
	}

	dl := dm.nav.DrivesList()

	barItems = append(
		barItems,
		[]*BarItem{
			{Content: "CAPACITY", BGColor: sbStyle.Drives.CapacityBG},
			{Content: FmtSize(dl.TotalCapacity, 0), BGColor: sbStyle.BG},
			{Content: "FREE", BGColor: sbStyle.Drives.FreeBG},
			{Content: FmtSize(dl.TotalFree, 0), BGColor: sbStyle.BG},
			{Content: "USED", BGColor: sbStyle.Drives.UsedBG},
			{Content: FmtSize(dl.TotalUsed, 0), BGColor: sbStyle.BG},
		}...,
	)

	dm.statusBar.Add(barItems)

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.statusBar.Render(dm.width),
	)
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
