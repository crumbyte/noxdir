package render

import (
	"container/heap"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/crumbyte/noxdir/filter"
	"github.com/crumbyte/noxdir/render/table"
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	entrySizeWidth      = 10
	topFilesTableHeight = 16
	colWidthRatio       = 0.13
)

type Mode string

const (
	PENDING Mode = "PENDING"
	READY   Mode = "READY"
	INPUT   Mode = "INPUT"
	DELETE  Mode = "DELETE"
)

type DirModel struct {
	columns       []Column
	dirsTable     *table.Model
	topFilesTable *table.Model
	topDirsTable  *table.Model
	deleteDialog  *DeleteDialogModel
	nav           *Navigation
	scanPG        *PG
	usagePG       *PG
	filters       filter.FiltersList
	mode          Mode
	lastErr       []error
	height        int
	width         int
	showTopFiles  bool
	showTopDirs   bool
	fullHelp      bool
	showCart      bool
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
		filters:       filter.NewFiltersList(defaultFilters...),
		dirsTable:     buildTable(),
		topFilesTable: buildTable(),
		topDirsTable:  buildTable(),
		mode:          PENDING,
		nav:           nav,
		scanPG:        &style.CS().ScanProgressBar,
		usagePG:       &usagePG,
	}

	s := table.DefaultStyles()
	s.Header = *style.TopTableHeader()
	s.Cell = lipgloss.NewStyle()
	s.Selected = lipgloss.NewStyle()

	dm.topFilesTable.SetStyles(s)
	dm.topFilesTable.SetHeight(topFilesTableHeight)

	dm.topDirsTable.SetStyles(s)
	dm.topDirsTable.SetHeight(topFilesTableHeight)

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

		dm.topFilesTable.SetRows(nil)
		dm.topDirsTable.SetRows(nil)

		structure.TopEntriesInstance.ScanFiles(dm.nav.Entry())
		structure.TopEntriesInstance.ScanDirs(dm.nav.Entry())

		dm.fillTopEntries(structure.TopEntriesInstance.Files(), dm.topFilesTable)
		dm.fillTopEntries(structure.TopEntriesInstance.Dirs(), dm.topDirsTable)
	case tea.WindowSizeMsg:
		dm.updateSize(msg.Width, msg.Height)
		dm.filters.Update(msg)
	case tea.KeyMsg:
		if dm.nav.OnDrives() || dm.handleKeyBindings(msg) {
			return dm, nil
		}
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
	keyBindings := dm.dirsTable.Help.ShortHelpView(ShortHelp())

	if dm.fullHelp {
		keyBindings = dm.dirsTable.Help.FullHelpView(
			append(NavigateKeyMap(), DirsKeyMap()...),
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

	if topContent, render := dm.viewTop(); render {
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

	if dm.mode == DELETE {
		return OverlayCenter(
			dm.width,
			dm.height,
			bg,
			dm.deleteDialog.View(),
		)
	}

	return bg
}

func (dm *DirModel) viewTop() (string, bool) {
	if !dm.showTopDirs && !dm.showTopFiles {
		return "", false
	}

	topTable := dm.topFilesTable
	if dm.showTopDirs {
		topTable = dm.topDirsTable
	}

	return topTable.View(), true
}

func (dm *DirModel) handleKeyBindings(msg tea.KeyMsg) bool {
	if dm.mode == PENDING {
		return false
	}

	bk := bindingKey(strings.ToLower(msg.String()))
	if bk == toggleNameFilter {
		if dm.mode == READY {
			dm.mode = INPUT
		} else {
			dm.mode = READY
		}

		dm.filters.ToggleFilter(filter.NameFilterID)
	}

	if dm.handleDeletion(bk, msg) {
		return true
	}

	if dm.mode == INPUT {
		dm.filters.Update(msg)
		dm.updateTableData()

		return true
	}

	switch bk {
	case toggleChart:
		dm.showCart = !dm.showCart
		dm.updateTableData()
	case toggleHelp:
		dm.fullHelp = !dm.fullHelp
	case explore:
		if dm.handleExploreKey() {
			return true
		}
	case toggleTopFiles:
		dm.showTopFiles = !dm.showTopFiles && !dm.showTopDirs
		dm.updateSize(dm.width, dm.height)
	case toggleTopDirs:
		dm.showTopDirs = !dm.showTopDirs && !dm.showTopFiles
		dm.updateSize(dm.width, dm.height)
	case toggleDirsFilter:
		dm.filters.ToggleFilter(filter.DirsOnlyFilterID)
		dm.updateTableData()
	case toggleFilesFilter:
		dm.filters.ToggleFilter(filter.FilesOnlyFilterID)
		dm.updateTableData()
	}

	return false
}

func (dm *DirModel) viewChart() string {
	chartSectors := make([]RawChartSector, 0, len(dm.nav.entry.Child))

	for _, child := range dm.nav.entry.Child {
		chartSectors = append(chartSectors, RawChartSector{
			Label: child.Name(),
			Size:  child.Size,
		})
	}

	return style.ChartBox().Render(
		Chart(
			dm.width/2,
			dm.height/2,
			dm.height/2,
			dm.nav.entry.Size,
			chartSectors,
			style.ChartColors(),
		),
	)
}

func (dm *DirModel) handleExploreKey() bool {
	sr := dm.dirsTable.SelectedRow()
	if len(sr) < 2 {
		return true
	}

	return dm.nav.Explore(sr[1]) != nil
}

func (dm *DirModel) handleDeletion(bk bindingKey, msg tea.Msg) bool {
	if bk == remove && dm.mode == READY {
		sr := dm.dirsTable.SelectedRow()

		dm.mode = DELETE
		dm.deleteDialog = NewDeleteDialogModel(dm.nav, sr[1])

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

	rows := make([]table.Row, 0, len(dm.nav.Entry().Child))
	dm.nav.Entry().SortChild()

	for _, child := range dm.nav.Entry().Child {
		if !dm.filters.Valid(child) {
			continue
		}

		totalDirs, totalFiles := "", ""

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
				FmtName(child.Name(), nameWidth),
				FmtSize(child.Size, entrySizeWidth),
				totalDirs,
				totalFiles,
				time.Unix(child.ModTime, 0).Format("2006-01-02 15:04"),
				FmtUsage(parentUsage),
				pgBar,
			},
		)
	}

	dm.dirsTable.SetRows(rows)
	dm.dirsTable.SetCursor(dm.nav.cursor)
}

func (dm *DirModel) dirsSummary() string {
	items := []*BarItem{
		NewBarItem(Version, style.cs.StatusBar.VersionBG, 0),
		NewBarItem("PATH", style.cs.StatusBar.Dirs.PathBG, 0),
		NewBarItem(dm.nav.Entry().Path, style.cs.StatusBar.BG, -1),
		NewBarItem(string(dm.mode), style.cs.StatusBar.Dirs.ModeBG, 0),
		NewBarItem("SIZE", style.cs.StatusBar.Dirs.SizeBG, 0),
		NewBarItem(FmtSize(dm.nav.Entry().Size, 0), style.cs.StatusBar.BG, 0),
		NewBarItem("DIRS", style.cs.StatusBar.Dirs.DirsBG, 0),
		NewBarItem(unitFmt(dm.nav.Entry().LocalDirs), style.cs.StatusBar.BG, 0),
		NewBarItem("FILES", style.cs.StatusBar.Dirs.FilesBG, 0),
		NewBarItem(unitFmt(dm.nav.Entry().LocalFiles), style.cs.StatusBar.BG, 0),
		NewBarItem("ERRORS", style.cs.StatusBar.Dirs.ErrorBG, 0),
		NewBarItem(unitFmt(uint64(len(dm.lastErr))), style.cs.StatusBar.BG, 0),
	}

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		NewStatusBar(items, dm.width),
	)
}

func (dm *DirModel) fillTopEntries(entries heap.Interface, tm *table.Model) {
	iconWidth := 5
	colSize := int(float64(dm.width-iconWidth) * colWidthRatio)
	nameWidth := dm.width - (colSize * 2) - iconWidth

	columns := []table.Column{
		{Title: "", Width: iconWidth},
		{Title: "", Width: 0},
		{Title: "Name", Width: nameWidth},
		{Title: "Size", Width: colSize},
		{Title: "Last Change", Width: colSize},
	}

	tm.SetColumns(columns)
	tm.SetCursor(0)

	if entries.Len() == 0 || tm.Rows() != nil {
		return
	}

	rows := make([]table.Row, 15)
	heap.Pop(entries)

	for i := len(rows) - 1; i >= 0; i-- {
		file, ok := heap.Pop(entries).(*structure.Entry)
		if !ok {
			continue
		}

		rootPath := dm.nav.Entry().Path + string(os.PathSeparator)

		if dm.nav.currentDrive != nil {
			rootPath = dm.nav.currentDrive.Path
		}

		path := strings.TrimSuffix(
			strings.TrimPrefix(file.Path, rootPath),
			file.Name(),
		)

		rows[i] = table.Row{
			EntryIcon(file),
			file.Path,
			path + style.TopFiles().Render(file.Name()),
			FmtSize(file.Size, entrySizeWidth),
			time.Unix(file.ModTime, 0).Format("2006-01-02 15:04"),
		}
	}

	tm.SetRows(rows)
	tm.SetCursor(0)
}

func (dm *DirModel) updateSize(width, height int) {
	dm.width, dm.height = width, height

	dm.dirsTable.SetWidth(width)
	dm.topFilesTable.SetWidth(width)
	dm.topDirsTable.SetWidth(width)

	dm.updateTableData()

	dm.fillTopEntries(structure.TopEntriesInstance.Files(), dm.topFilesTable)
	dm.fillTopEntries(structure.TopEntriesInstance.Dirs(), dm.topDirsTable)
}

func (dm *DirModel) viewProgress() string {
	completed := (float64(dm.nav.Entry().Size) / float64(dm.nav.currentDrive.UsedBytes)) - 0.01

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.scanPG.New(dm.width).ViewAs(completed),
	)
}
