package render

import (
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"time"

	"github.com/crumbyte/noxdir/command"
	"github.com/crumbyte/noxdir/drive"
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

	// CMD mode represents the model state when the application awaits for the
	// internal command.Model to be executed.
	CMD Mode = "CMD"
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
	columns         Columns
	mode            Mode
	dirsTable       *table.Model
	topEntries      *TopEntries
	deleteDialog    *DeleteDialogModel
	diff            *DiffModel
	usagePG         *PG
	nav             *Navigation
	scanPG          *PG
	filters         filter.FiltersList
	cmd             *command.Model
	errPopup        *PopupModel
	topStatusBar    *StatusBar
	bottomStatusBar *StatusBar
	summaryInfo     *summaryInfo
	sortState       SortState
	height          int
	width           int
	fullHelp        bool
	showCart        bool
}

func NewDirModel(nav *Navigation, filters ...filter.EntryFilter) *DirModel {
	defaultFilters := append(
		[]filter.EntryFilter{
			filter.NewNameFilter(
				"Filter... \\ at the beginning for a negative search",
				style.CS().FilterText,
			),
			&filter.DirsFilter{},
			&filter.FilesFilter{},
		},
		filters...,
	)

	usagePG := style.CS().UsageProgressBar
	usagePG.EmptyChar = " "

	dm := &DirModel{
		columns: Columns{
			{Title: "", Width: 5, Fixed: true},
			{Title: "", Hidden: func(_ int) bool { return true }},
			{Title: "Name", SortKey: structure.SortPath, WidthRatio: 0.25},
			{
				Title:      "Size",
				SortKey:    structure.SortSize,
				MinWidth:   12,
				WidthRatio: DefaultColWidthRatio,
			},
			{
				Title:      "Total Dirs",
				SortKey:    structure.SortTotalDirs,
				WidthRatio: DefaultColWidthRatio,
				Hidden:     func(fw int) bool { return fw < 100 },
			},
			{
				Title:      "Total Files",
				SortKey:    structure.SortTotalFiles,
				WidthRatio: DefaultColWidthRatio,
				Hidden:     func(fw int) bool { return fw < 100 },
			},
			{
				Title:      "Last Change",
				WidthRatio: DefaultColWidthRatio,
				Hidden:     func(fw int) bool { return fw < 150 },
			},
			{
				Title:      "Parent Usage",
				WidthRatio: DefaultColWidthRatio,
				MinWidth:   12,
			},
			{Title: "", Full: true},
		},
		filters:         filter.NewFiltersList(defaultFilters...),
		dirsTable:       buildTable(),
		topEntries:      NewTopEntries(),
		diff:            NewDiffModel(nav),
		topStatusBar:    NewStatusBar(),
		bottomStatusBar: NewStatusBar(),
		cmd: command.NewModel(
			func() { go teaProg.Send(EnqueueRefresh{Mode: CMD}) },
		),
		errPopup: NewPopupModel(
			ErrorTitle, time.Second*10, PopupDefaultErrorStyle(),
		),
		summaryInfo: &summaryInfo{},
		sortState:   SortState{Key: structure.SortSize, Desc: true},
		mode:        PENDING,
		nav:         nav,
		scanPG:      &style.CS().ScanProgressBar,
		usagePG:     &usagePG,
	}

	cmdDefaultStyle := command.DefaultStyles

	cmdDefaultStyle.InputTextStyle = *style.CmdInputText()
	cmdDefaultStyle.InputBarStyle = *style.CmdBarBorder()
	cmdDefaultStyle.OutputStyle = *style.CmdBarBorder()

	dm.cmd.SetStyles(cmdDefaultStyle)

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

		if msg.Err != nil {
			dm.errPopup.Show(msg.Err.Error())

			break
		}

		if msg.Deleted {
			go func() {
				teaProg.Send(EnqueueRefresh{Mode: READY})
			}()
		}
	case UpdateDirState:
		dm.mode = PENDING
		runtime.GC()
		dm.nav.tree.CalculateSize()

		dm.updateTableData()
	case ScanFinished:
		dm.mode = msg.Mode

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
		dm.cmd.Update(msg)
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

	_, _ = dm.errPopup.Update(msg)

	t, _ := dm.dirsTable.Update(msg)
	dm.dirsTable = &t

	return dm, tea.Batch(cmd)
}

func (dm *DirModel) View() string {
	h := lipgloss.Height

	tsb := dm.viewTopStatusBar()
	bsb := dm.viewBottomStatusBar()
	keyBindings := dm.dirsTable.Help.ShortHelpView(
		Bindings.ShortBindings(),
	)

	if dm.fullHelp {
		keyBindings = dm.dirsTable.Help.FullHelpView(
			Bindings.DirBindings(),
		)
	}

	pgBar := tsb

	if dm.mode == PENDING {
		pgBar = dm.viewProgress()
	}

	dirsTableHeight := dm.height - h(keyBindings) - h(bsb) - h(pgBar)

	rows := []string{keyBindings, bsb}

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

	if cmdInput := dm.cmd.View(); cmdInput != "" {
		rows = append(rows, cmdInput)

		dirsTableHeight -= h(cmdInput)
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
			h(bg)-h(keyBindings)-h(bsb)-h(chart),
			dm.width-lipgloss.Width(chart),
		)
	}

	if popupMessage := dm.errPopup.View(); len(popupMessage) != 0 {
		bg = Overlay(
			dm.width,
			bg,
			popupMessage,
			0,
			dm.width-lipgloss.Width(popupMessage),
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

	handlers := []func(tea.KeyMsg) bool{
		dm.handleFilter, dm.handleDiff, dm.handleDeletion, dm.handleCmd,
	}

	for _, handler := range handlers {
		if handler(msg) {
			return true
		}
	}

	switch {
	case key.Matches(msg, Bindings.Dirs.SortKeys):
		dm.sortEntries(drive.SortKey(msg.String()))

		return true
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
	case key.Matches(msg, Bindings.Dirs.ToggleSelectAll):
		dm.dirsTable.ToggleMarkAll()
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
		max(int(float64(dm.width)*0.45), 100),
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

func (dm *DirModel) handleCmd(msg tea.KeyMsg) bool {
	if key.Matches(msg, Bindings.Dirs.Command) && dm.mode == READY {
		var entries []string

		for _, r := range dm.dirsTable.MarkedRows() {
			entries = append(entries, r[1])
		}

		if len(entries) == 0 {
			entries = append(entries, dm.dirsTable.SelectedRow()[1])
		}

		dm.cmd.SetPathContext(dm.nav.entry.Path, entries)
		dm.cmd.Enable()
		dm.mode = CMD

		dm.updateTableData()

		return true
	}

	if dm.mode == CMD {
		dm.cmd.Update(msg)

		if !dm.cmd.Enabled() {
			dm.mode = READY
		}

		dm.updateTableData()

		return true
	}

	return false
}

func (dm *DirModel) handleDeletion(msg tea.KeyMsg) bool {
	if key.Matches(msg, Bindings.Dirs.Delete) && dm.mode == READY {
		dm.mode = DELETE

		target := make(map[string]int64)

		for _, r := range dm.dirsTable.MarkedRows() {
			childEntry := dm.nav.Entry().GetChild(r[1])

			if childEntry != nil {
				target[r[1]] = childEntry.Size
			}
		}

		if len(target) == 0 {
			childEntry := dm.nav.Entry().GetChild(dm.dirsTable.SelectedRow()[1])

			if childEntry != nil {
				target[dm.dirsTable.SelectedRow()[1]] = childEntry.Size
			}
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

	nameCol, _ := dm.columns.Get(2)
	sizeCol, _ := dm.columns.Get(3)
	usageCol, _ := dm.columns.Get(7)
	pgCol, _ := dm.columns.Get(8)

	dm.dirsTable.SetColumns(dm.columns.TableColumns(dm.width, dm.sortState))
	fillProgress := dm.usagePG.New(pgCol.Width)

	dm.summaryInfo.clear()

	rows := make([]table.Row, 0, len(dm.nav.Entry().Child))
	dm.nav.Entry().SortedChild(dm.sortState.Key, dm.sortState.Desc)

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
				WrapString(child.Name(), nameCol.Width),
				FmtSizeColor(child.Size, entrySizeWidth, sizeCol.Width),
				totalDirs,
				totalFiles,
				time.Unix(child.ModTime, 0).Format("02 Jan 2006"),
				FmtUsage(parentUsage, 20, usageCol.Width),
				pgBar,
			},
		)
	}

	dm.dirsTable.SetRows(rows)
	dm.dirsTable.SetCursor(dm.nav.cursor)
}

func (dm *DirModel) viewTopStatusBar() string {
	if dm.topStatusBar == nil {
		return ""
	}

	dm.topStatusBar.Clear()

	sbStyle := style.CS().StatusBar

	var (
		fullEntryName string
		selectedSize  int64
		isDir         bool
	)

	for _, selected := range dm.dirsTable.MarkedRows() {
		if entry := dm.nav.entry.GetChild(selected[1]); entry != nil {
			selectedSize += entry.Size
		}
	}

	if len(dm.dirsTable.SelectedRow()) != 0 {
		fullEntryName = dm.dirsTable.SelectedRow()[1]

		entry := dm.nav.entry.GetChild(fullEntryName)
		if entry != nil && selectedSize == 0 {
			selectedSize = entry.Size
			isDir = entry.IsDir
		}
	}

	barItems := []*BarItem{
		{Content: "ENTRY", BGColor: sbStyle.VersionBG},
		{Content: "NAME", BGColor: sbStyle.Dirs.PathBG},
		{
			Content: fullEntryName,
			BGColor: sbStyle.BG,
			Wrapper: WrapPath,
			Width:   -1,
		},
	}

	entryType := "DIR"
	if !isDir {
		entryType = "FILE"
	}

	barItems = append(
		barItems,
		[]*BarItem{
			{Content: entryType, BGColor: sbStyle.Dirs.ModeBG},
			{Content: "PICKED", BGColor: sbStyle.Dirs.SizeBG},
			{
				Content: unitFmt(max(uint64(len(dm.dirsTable.MarkedRows())), 1)),
				BGColor: sbStyle.BG,
			},
			{Content: "TO FREE", BGColor: style.cs.StatusBar.Dirs.RowsCounter},
			{Content: FmtSizeColor(selectedSize, 0, 0), BGColor: sbStyle.BG},
		}...,
	)

	dm.topStatusBar.Add(barItems)

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.topStatusBar.Render(dm.width),
	)
}

func (dm *DirModel) viewBottomStatusBar() string {
	if dm.bottomStatusBar == nil {
		return ""
	}

	dm.bottomStatusBar.Clear()

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
			{Content: FmtSizeColor(dm.summaryInfo.size, 0, 0), BGColor: sbStyle.BG},
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

	dm.bottomStatusBar.Add(barItems)

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.bottomStatusBar.Render(dm.width),
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

func (dm *DirModel) sortEntries(sortKey drive.SortKey) {
	if dm.nav.OnDrives() {
		return
	}

	if dm.sortState.Key == sortKey {
		dm.sortState.Desc = !dm.sortState.Desc
	} else {
		dm.sortState = SortState{Key: sortKey}
	}

	dm.updateTableData()
}
