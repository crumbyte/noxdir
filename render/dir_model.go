package render

import (
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"time"

	"github.com/crumbyte/noxdir/filter"
	"github.com/crumbyte/noxdir/render/table"
	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	entrySizeWidth      = 10
	topFilesTableHeight = 16
	colWidthRatio       = 0.13
)

// Mode defines a custom type that represents the current view mode. Depending
// on the current Mode value, the UI behavior can vary.
type Mode string

const (
	// PENDING mode represents the locked model state. This state is enabled
	// while waiting for task completion to prevent UI state changes.
	PENDING Mode = "PENDING"

	// READY mode represents the normal model state when there are no pending
	// tasks or user inputs.
	READY Mode = "READY"

	// INPUT mode represents the model state when the application awaits the
	// user's input. In that state, any key binding will be processed as plain
	// text.
	INPUT Mode = "INPUT"

	// DELETE mode represents the model state when the application awaits the
	// deletion confirmation. The UI behavior is limited in this mode.
	DELETE Mode = "DELETE"

	// DIFF mode represents the model state while showing the file system state
	// changes from the previous session. The UI behavior is limited in this mode.
	DIFF Mode = "DIFF"
)

type summaryInfo struct {
	size  int64
	dirs  uint64
	files uint64
}

func (si *summaryInfo) add(e *structure.Entry) {
	si.size += e.Size

	if e.IsDir {
		si.dirs++

		return
	}

	si.files++
}

func (si *summaryInfo) clear() {
	si.size, si.dirs, si.files = 0, 0, 0
}

type DirModel struct {
	columns      []Column
	mode         Mode
	dirsTable    *table.Model
	topEntries   *TopEntries
	deleteDialog *DeleteDialogModel
	diff         *DiffModel
	usagePG      *PG
	nav          *Navigation
	scanPG       *PG
	filters      filter.FiltersList
	statusBar    *StatusBar
	summaryInfo  *summaryInfo
	height       int
	width        int
	fullHelp     bool
	showCart     bool
}

func NewDirModel(nav *Navigation, filters ...filter.EntryFilter) *DirModel {
	defaultFilters := append(
		[]filter.EntryFilter{
			filter.NewNameFilter("Filter..."),
			&filter.DirsFilter{},
			&filter.FilesFilter{},
		},
		filters...,
	)

	usagePG := style.CS().UsageProgressBar
	usagePG.EmptyChar = " "

	dm := &DirModel{
		columns: []Column{
			{Title: ""},
			{Title: ""},
			{Title: "Name"},
			{Title: "Size"},
			{Title: "Total Dirs"},
			{Title: "Total Files"},
			{Title: "Last Change"},
			{Title: "Parent usage"},
			{Title: ""},
		},
		filters:     filter.NewFiltersList(defaultFilters...),
		dirsTable:   buildTable(),
		topEntries:  NewTopEntries(),
		diff:        NewDiffModel(nav),
		statusBar:   NewStatusBar(),
		summaryInfo: &summaryInfo{},
		mode:        PENDING,
		nav:         nav,
		scanPG:      &style.CS().ScanProgressBar,
		usagePG:     &usagePG,
	}

	return dm
}

func (dm *DirModel) Init() tea.Cmd {
	return nil
}

func (dm *DirModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	dm.nav.SetCursor(dm.dirsTable.Cursor())

	switch msg := msg.(type) {
	case EntryDeleted:
		dm.mode, dm.deleteDialog = READY, nil

		if msg.Deleted {
			go func() {
				teaProg.Send(EnqueueRefresh{})
			}()
		}
	case UpdateDirState:
		dm.mode = PENDING
		runtime.GC()
		dm.nav.tree.CalculateSize()

		dm.updateTableData()
	case ScanFinished:
		dm.mode = READY

		runtime.GC()
		dm.nav.tree.CalculateSize()
		dm.updateTableData()

		dm.dirsTable.ResetMarked()
		dm.topEntries.Clear()

		structure.TopEntriesInstance.ScanFiles(dm.nav.Entry())
		structure.TopEntriesInstance.ScanDirs(dm.nav.Entry())

		dm.topEntries.UpdateTopEntries()
	case tea.WindowSizeMsg:
		dm.updateSize(msg.Width, msg.Height)

		dm.diff.Update(msg)
		dm.filters.Update(msg)
		dm.topEntries.Update(msg)
	case tea.KeyMsg:
		if dm.nav.OnDrives() || dm.handleKeyBindings(msg) {
			return dm, nil
		}
	}

	if dm.mode == DIFF {
		dm.diff.Update(msg)
	}

	if dm.nav.OnDrives() {
		return dm, nil
	}

	t, _ := dm.dirsTable.Update(msg)
	dm.dirsTable = &t

	return dm, tea.Batch(cmd)
}

func (dm *DirModel) View() string {
	h := lipgloss.Height

	summary := dm.dirsSummary()
	keyBindings := dm.dirsTable.Help.ShortHelpView(
		Bindings.ShortBindings(),
	)

	if dm.fullHelp {
		keyBindings = dm.dirsTable.Help.FullHelpView(
			Bindings.DirBindings(),
		)
	}

	pgBar := summary

	if dm.mode == PENDING {
		pgBar = dm.viewProgress()
	}

	dirsTableHeight := dm.height - h(keyBindings) - h(summary) - h(pgBar)

	rows := []string{keyBindings, summary}

	for _, f := range dm.filters {
		v, ok := f.(filter.Viewer)
		if !ok {
			continue
		}

		rendered := v.View()

		if len(rendered) > 0 {
			dirsTableHeight -= h(rendered)

			rows = append(rows, rendered)
		}
	}

	if topContent := dm.topEntries.View(); len(topContent) != 0 {
		dirsTableHeight -= h(topContent)
		rows = append(rows, topContent)
	}

	dm.dirsTable.SetHeight(dirsTableHeight)

	rows = append(rows, dm.dirsTable.View(), pgBar)
	slices.Reverse(rows)

	bg := lipgloss.JoinVertical(lipgloss.Top, rows...)

	if dm.showCart {
		chart := dm.viewChart()

		bg = Overlay(
			dm.width,
			bg,
			chart,
			h(bg)-h(keyBindings)-h(summary)-h(chart),
			dm.width-lipgloss.Width(chart),
		)
	}

	if dm.mode == DIFF {
		return OverlayCenter(dm.width, dm.height, bg, dm.diff.View())
	}

	if dm.mode == DELETE {
		return OverlayCenter(dm.width, dm.height, bg, dm.deleteDialog.View())
	}

	return bg
}

func (dm *DirModel) handleKeyBindings(msg tea.KeyMsg) bool {
	if dm.mode == PENDING {
		return false
	}

	if dm.handleFilter(msg) {
		return true
	}

	if dm.handleDiff(msg) {
		return true
	}

	if dm.handleDeletion(msg) {
		return true
	}

	switch {
	case key.Matches(msg, Bindings.Dirs.Chart):
		dm.showCart = !dm.showCart
		dm.updateTableData()
	case key.Matches(msg, Bindings.Help):
		dm.fullHelp = !dm.fullHelp
	case key.Matches(msg, Bindings.Explore):
		if dm.handleExploreKey() {
			return true
		}
	case key.Matches(msg, Bindings.Dirs.DirsOnly):
		dm.filters.ToggleFilter(filter.DirsOnlyFilterID)
		dm.updateTableData()
	case key.Matches(msg, Bindings.Dirs.FilesOnly):
		dm.filters.ToggleFilter(filter.FilesOnlyFilterID)
		dm.updateTableData()
	}

	dm.topEntries.Update(msg)

	return false
}

func (dm *DirModel) viewChart() string {
	si := make([]SectorInfo, 0, len(dm.nav.entry.Child))

	for _, child := range dm.nav.entry.Child {
		si = append(si, SectorInfo{Label: child.Name(), Size: child.Size})
	}

	c := NewChart(
		dm.width/2,
		int(float64(dm.height)*0.43),
		int(float64(dm.height)*0.43),
		style.CS().ChartColors.AspectRatioFix,
		style.ChartColors(),
	)

	return style.ChartBox().Render(c.Render(dm.nav.entry.Size, si))
}

func (dm *DirModel) handleExploreKey() bool {
	sr := dm.dirsTable.SelectedRow()
	if len(sr) < 2 {
		return true
	}

	return dm.nav.Explore(sr[1]) != nil
}

func (dm *DirModel) handleFilter(msg tea.KeyMsg) bool {
	if key.Matches(msg, Bindings.Dirs.NameFilter) {
		if dm.mode == READY {
			dm.mode = INPUT
		} else {
			dm.mode = READY
		}

		dm.filters.ToggleFilter(filter.NameFilterID)
	}

	if dm.mode == INPUT {
		dm.filters.Update(msg)
		dm.updateTableData()

		return true
	}

	return false
}

func (dm *DirModel) handleDeletion(msg tea.KeyMsg) bool {
	if key.Matches(msg, Bindings.Dirs.Delete) && dm.mode == READY {
		dm.mode = DELETE

		var target []string

		for _, r := range dm.dirsTable.MarkedRows() {
			target = append(target, r[1])
		}

		if len(target) == 0 {
			target = append(target, dm.dirsTable.SelectedRow()[1])
		}

		dm.deleteDialog = NewDeleteDialogModel(dm.nav, target)

		dm.updateTableData()

		return true
	}

	if dm.mode == DELETE {
		dm.deleteDialog.Update(msg)
		dm.updateTableData()

		return true
	}

	return false
}

func (dm *DirModel) handleDiff(msg tea.KeyMsg) bool {
	isDiffKey := key.Matches(msg, Bindings.Dirs.Diff)

	switch {
	case isDiffKey && dm.mode == READY:
		dm.mode = DIFF
		dm.diff.Run(dm.width, dm.height)
	case isDiffKey && dm.mode == DIFF:
		dm.mode = READY
		dm.updateTableData()
	case dm.mode == DIFF:
		dm.diff.Update(msg)
	default:
		return false
	}

	return true
}

func (dm *DirModel) updateTableData() {
	if dm.nav.OnDrives() || dm.nav.Entry() == nil || !dm.nav.Entry().IsDir {
		return
	}

	iconWidth := 5
	nameWidth := (dm.width - iconWidth) / 4

	colWidth := int(float64(dm.width-iconWidth-nameWidth) * colWidthRatio)
	progressWidth := dm.width - (colWidth * 5) - iconWidth - nameWidth

	// columns must be re-rendered ech time to support window resize
	columns := make([]table.Column, len(dm.columns))

	for i, c := range dm.columns {
		columns[i] = table.Column{Title: c.Title, Width: colWidth}
	}

	columns[0].Width = iconWidth
	columns[1].Width = 0
	columns[2].Width = nameWidth
	columns[len(columns)-1].Width = progressWidth

	dm.dirsTable.SetColumns(columns)

	fillProgress := dm.usagePG.New(progressWidth)
	dm.summaryInfo.clear()

	rows := make([]table.Row, 0, len(dm.nav.Entry().Child))
	dm.nav.Entry().SortChild()

	for _, child := range dm.nav.Entry().Child {
		if !dm.filters.Valid(child) {
			continue
		}

		totalDirs, totalFiles := "", ""

		dm.summaryInfo.add(child)

		if child.IsDir {
			totalDirs = strconv.FormatUint(child.TotalDirs, 10)
			totalFiles = strconv.FormatUint(child.TotalFiles, 10)
		}

		parentUsage := float64(child.Size) / float64(dm.nav.ParentSize())
		pgBar := fillProgress.ViewAs(parentUsage)

		rows = append(
			rows,
			table.Row{
				EntryIcon(child),
				child.Name(),
				WrapString(child.Name(), nameWidth),
				FmtSizeColor(child.Size, entrySizeWidth, colWidth),
				totalDirs,
				totalFiles,
				time.Unix(child.ModTime, 0).Format("Jan 02 15:04"),
				FmtUsage(parentUsage, 20, colWidth),
				pgBar,
			},
		)
	}

	dm.dirsTable.SetRows(rows)
	dm.dirsTable.SetCursor(dm.nav.cursor)
}

func (dm *DirModel) dirsSummary() string {
	if dm.statusBar == nil {
		return ""
	}

	dm.statusBar.Clear()

	sbStyle := style.CS().StatusBar

	barItems := []*BarItem{
		{Content: Version, BGColor: sbStyle.VersionBG},
		{Content: "PATH", BGColor: sbStyle.Dirs.PathBG},
		{
			Content: dm.nav.Entry().Path,
			BGColor: sbStyle.BG,
			Wrapper: WrapPath,
			Width:   -1,
		},
	}

	if dm.nav.cacheEnabled {
		barItems = append(
			barItems, &BarItem{
				Content: "CACHED",
				BGColor: style.CS().StatusBar.VersionBG,
			},
		)
	}

	barItems = append(
		barItems,
		[]*BarItem{
			{Content: string(dm.mode), BGColor: sbStyle.Dirs.ModeBG},
			{Content: "SIZE", BGColor: sbStyle.Dirs.SizeBG},
			{Content: FmtSize(dm.summaryInfo.size, 0), BGColor: sbStyle.BG},
			{Content: "DIRS", BGColor: sbStyle.Dirs.DirsBG},
			{Content: unitFmt(dm.summaryInfo.dirs), BGColor: sbStyle.BG},
			{Content: "FILES", BGColor: sbStyle.Dirs.FilesBG},
			{Content: unitFmt(dm.summaryInfo.files), BGColor: sbStyle.BG},
			{Content: fmt.Sprintf(
				"%d:%d",
				dm.dirsTable.Cursor()+1,
				len(dm.dirsTable.Rows()),
			), BGColor: style.cs.StatusBar.Dirs.RowsCounter},
		}...,
	)

	dm.statusBar.Add(barItems)

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.statusBar.Render(dm.width),
	)
}

func (dm *DirModel) updateSize(width, height int) {
	dm.width, dm.height = width, height

	dm.dirsTable.SetWidth(width)

	dm.updateTableData()
}

func (dm *DirModel) viewProgress() string {
	completed := 0.99

	if dm.nav.currentDrive != nil {
		completed = float64(dm.nav.Entry().Size)/float64(dm.nav.currentDrive.UsedBytes) - 0.01
	}

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.scanPG.New(dm.width).ViewAs(completed),
	)
}
