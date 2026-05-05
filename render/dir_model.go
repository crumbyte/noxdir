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

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/key"
)

const (
	entrySizeWidth      = 10
	topFilesTableHeight = 16
	dirsTableRatio      = 0.7
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
	previewTable    *table.Model
	topEntries      *TopEntries
	deleteDialog    *DeleteDialogModel
	diff            *DiffModel
	nav             *Navigation
	scanPG          *PG
	filters         filter.FiltersList
	cmd             *command.Model
	errPopup        *PopupModel
	topStatusBar    *StatusBar
	bottomStatusBar *StatusBar
	summaryInfo     *summaryInfo
	sortState       SortState
	view            tea.View
	height          int
	width           int
	fullHelp        bool
	showChart       bool
}

func NewDirModel(nav *Navigation, filters ...filter.EntryFilter) *DirModel {
	defaultFilters := append(
		[]filter.EntryFilter{
			filter.NewNameFilter(style.CS().FilterText),
			&filter.DirsFilter{},
			&filter.FilesFilter{},
		},
		filters...,
	)

	dm := &DirModel{
		columns: Columns{
			{Title: "", Width: 5, Fixed: true},
			{Title: "", Hidden: func(_ int) bool { return true }},
			{Title: "Name", SortKey: structure.SortPath, Full: true},
			{
				Title:      "Size",
				SortKey:    structure.SortSize,
				MinWidth:   15,
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
				Hidden:     func(fw int) bool { return fw < 125 },
			},
			{
				Title:      "Parent Usage",
				WidthRatio: DefaultColWidthRatio,
				MinWidth:   15,
			},
		},
		filters:         filter.NewFiltersList(defaultFilters...),
		dirsTable:       buildTable(),
		previewTable:    buildTable(),
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
	}

	s := dm.dirsTable.Styles()
	s.Selected = lipgloss.NewStyle()

	dm.previewTable.SetStyles(s)

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

	switch msg := msg.(type) {
	case PopupMsgTick:
		dm.updateTableData()

		return dm, cmd
	case EntryDeleted:
		dm.mode, dm.deleteDialog = READY, nil

		if msg.Err != nil {
			dm.errPopup.Show(msg.Err.Error())

			break
		}

		if msg.Deleted {
			dm.dirsTable.ResetMarked()

			dm.updateTableData()
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
		dm.updateTableSize(msg)
	case tea.KeyPressMsg:
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

	dm.nav.SetCursor(dm.dirsTable.Cursor())

	dm.updatePreviewTable()

	return dm, tea.Batch(cmd)
}

func (dm *DirModel) View() tea.View {
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

		rendered := v.View().Content

		if len(rendered) > 0 {
			dirsTableHeight -= h(rendered)

			rows = append(rows, rendered)
		}
	}

	if cmdInput := dm.cmd.View(); len(cmdInput.Content) != 0 {
		rows = append(rows, cmdInput.Content)

		dirsTableHeight -= h(cmdInput.Content)
	}

	if topContent := dm.topEntries.View(); len(topContent.Content) != 0 {
		dirsTableHeight -= h(topContent.Content)
		rows = append(rows, topContent.Content)
	}

	dm.dirsTable.SetHeight(dirsTableHeight)
	dm.previewTable.SetHeight(dirsTableHeight)

	tablesLayout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		style.DirTable(dm.dirTableWidth()).Render(dm.dirsTable.View().Content),
		style.TableSeparator(dirsTableHeight).Render(),
		style.DirTable(dm.previewTableWidth()).Render(dm.previewTable.View().Content),
	)

	rows = append(rows, tablesLayout, pgBar)
	slices.Reverse(rows)

	bg := lipgloss.JoinVertical(lipgloss.Top, rows...)

	return dm.renderOverlay(&bg, h(bg)-h(keyBindings)-h(bsb))
}

func (dm *DirModel) Clear() {
	dm.dirsTable.SetRows(nil)
}

func (dm *DirModel) renderOverlay(layout *string, layoutHeight int) tea.View {
	if dm.showChart {
		chart := dm.viewChart()

		*layout = Overlay(
			dm.width,
			*layout,
			chart,
			layoutHeight-lipgloss.Height(chart),
			dm.width-lipgloss.Width(chart),
		)
	}

	if errPopup := dm.errPopup.View(); len(errPopup.Content) != 0 {
		*layout = Overlay(
			dm.width,
			*layout,
			errPopup.Content,
			0,
			dm.width-lipgloss.Width(errPopup.Content),
		)
	}

	if dm.mode == DIFF {
		dm.view.SetContent(OverlayCenter(
			dm.width, dm.height, *layout, dm.diff.View().Content,
		))

		return dm.view
	}

	if dm.mode == DELETE {
		dm.view.SetContent(OverlayCenter(
			dm.width, dm.height, *layout, dm.deleteDialog.View().Content,
		))

		return dm.view
	}

	dm.view.SetContent(*layout)

	return dm.view
}

func (dm *DirModel) handleKeyBindings(msg tea.KeyPressMsg) bool {
	if dm.mode == PENDING {
		return false
	}

	handlers := []func(tea.KeyPressMsg) bool{
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
		dm.showChart = !dm.showChart
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
	if sr != nil && len(sr.Cols) < 2 {
		return true
	}

	return dm.nav.Explore(sr.Cols[1]) != nil
}

func (dm *DirModel) handleFilter(msg tea.KeyPressMsg) bool {
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

		if f, ok := dm.filters[filter.NameFilterID]; ok {
			if !f.Enabled() {
				dm.mode = READY
			}
		}

		return true
	}

	return false
}

func (dm *DirModel) handleCmd(msg tea.KeyPressMsg) bool {
	if key.Matches(msg, Bindings.Dirs.Command) && dm.mode == READY {
		var entries []string

		for _, r := range dm.dirsTable.MarkedRows() {
			entries = append(entries, r.Cols[1])
		}

		if len(entries) == 0 {
			entries = append(entries, dm.dirsTable.SelectedRow().Cols[1])
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

func (dm *DirModel) handleDeletion(msg tea.KeyPressMsg) bool {
	if key.Matches(msg, Bindings.Dirs.Delete) && dm.mode == READY {
		dm.mode = DELETE

		toDelete := make([]*structure.Entry, 0)

		for _, r := range dm.dirsTable.MarkedRows() {
			childEntry := dm.nav.Entry().GetChildByName(r.Cols[1])

			if childEntry != nil {
				toDelete = append(toDelete, childEntry)
			}
		}

		if len(toDelete) == 0 {
			childEntry := dm.nav.Entry().
				GetChildByName(dm.dirsTable.SelectedRow().Cols[1])

			if childEntry != nil {
				toDelete = append(toDelete, childEntry)
			}
		}

		dm.deleteDialog = NewDeleteDialogModel(dm.nav, toDelete)

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

func (dm *DirModel) handleDiff(msg tea.KeyPressMsg) bool {
	isDiffKey := key.Matches(msg, Bindings.Dirs.Diff)

	switch {
	case isDiffKey && dm.mode == READY:
		dm.mode = DIFF
		dm.diff.Run(dm.width, dm.height)
	case isDiffKey && dm.mode == DIFF:
		dm.mode = READY
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

	dm.dirsTable.SetColumns(
		dm.columns.TableColumns(dm.dirTableWidth(), dm.sortState),
	)

	nameCol, ok := dm.dirsTable.Column(2)
	if !ok {
		return
	}

	dm.summaryInfo.clear()

	rows := make([]table.Row, 0, len(dm.nav.Entry().Child))

	dm.nav.Entry().SortedChild(dm.sortState.Key, dm.sortState.Desc)

	for _, child := range dm.nav.Entry().Child {
		if !dm.filters.Valid(child) {
			continue
		}

		totalDirs, totalFiles := "-", "-"

		dm.summaryInfo.add(child)

		if child.IsDir {
			totalDirs = strconv.FormatUint(child.TotalDirs, 10)
			totalFiles = strconv.FormatUint(child.TotalFiles, 10)
		}

		parentUsage := float64(child.Size) / float64(dm.nav.ParentSize())

		rows = append(
			rows,
			table.Row{
				Cols: []string{
					EntryIcon(child),
					child.Name(),
					WrapString(child.Name(), nameCol.Width),
					FmtSizeColor(child.Size, entrySizeWidth),
					Faint(totalDirs),
					Faint(totalFiles),
					Faint(time.Unix(child.ModTime, 0).Format("02 Jan 2006")),
					FmtUsage(parentUsage, 20),
				},
			},
		)
	}

	dm.dirsTable.SetRows(rows)
	dm.dirsTable.MoveCursor(dm.nav.cursor)

	dm.updatePreviewTable()
}

func (dm *DirModel) updatePreviewTable() {
	if dm.nav.Entry() == nil || dm.dirsTable.SelectedRow() == nil {
		return
	}

	parent := dm.nav.Entry().GetChildByName(dm.dirsTable.SelectedRow().Cols[1])
	if parent == nil {
		return
	}

	cols := Columns{
		{Title: "", Width: 5, Fixed: true},
		{Title: "Preview", Full: true},
		{Title: "", Width: 15, Fixed: true},
	}

	dm.previewTable.SetColumns(
		cols.TableColumns(dm.previewTableWidth(), SortState{}),
	)

	nameCol, ok := dm.previewTable.Column(1)
	if !ok {
		return
	}

	rows := make([]table.Row, 0, dm.height)

	if len(parent.Child) == 0 {
		rows = append(rows, table.Row{Cols: []string{"", "No Preview", ""}})

		dm.previewTable.SetRows(rows)

		return
	}

	parent.SortedChild("", true)

	for i := range min(dm.height, len(parent.Child)) {
		child := parent.Child[i]

		rows = append(
			rows,
			table.Row{
				Cols: []string{
					EntryIcon(child),
					WrapString(child.Name(), nameCol.Width),
					FmtSizeColor(child.Size, entrySizeWidth),
				},
			},
		)
	}

	dm.previewTable.SetRows(rows)
}

func (dm *DirModel) viewTopStatusBar() string {
	if dm.topStatusBar == nil {
		return ""
	}

	dm.topStatusBar.Clear()

	statusBarStyle := style.CS().StatusBar

	var (
		fullEntryName string
		selectedSize  int64
		isDir         bool
	)

	for _, selected := range dm.dirsTable.MarkedRows() {
		if entry := dm.nav.entry.GetChildByName(selected.Cols[1]); entry != nil {
			selectedSize += entry.Size
		}
	}

	if dm.dirsTable.SelectedRow() != nil {
		fullEntryName = dm.dirsTable.SelectedRow().Cols[1]

		entry := dm.nav.entry.GetChildByName(fullEntryName)
		if entry != nil && selectedSize == 0 {
			selectedSize = entry.Size
			isDir = entry.IsDir
		}
	}

	barItems := make([]*BarItem, 0, 8)

	barItems = append(barItems,
		&BarItem{Content: "NAME", BGColor: statusBarStyle.Dirs.PathBG},
		&BarItem{
			Content: fullEntryName,
			BGColor: statusBarStyle.BG,
			Wrapper: PrefixWrapString,
			Width:   -1,
		},
	)

	entryType := "DIR"
	if !isDir {
		entryType = "FILE"
	}

	barItems = append(
		barItems,
		&BarItem{Content: entryType, BGColor: statusBarStyle.Dirs.ModeBG},
		&BarItem{Content: "PICKED", BGColor: statusBarStyle.Dirs.SizeBG},
		&BarItem{
			Content: unitFmt(max(uint64(len(dm.dirsTable.MarkedRows())), 1)),
			BGColor: statusBarStyle.BG,
		},
		&BarItem{Content: "TO FREE", BGColor: style.cs.StatusBar.Dirs.RowsCounter},
		&BarItem{Content: FmtSizeColor(selectedSize, 0), BGColor: statusBarStyle.BG},
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

	statusBarStyle := style.CS().StatusBar

	barItems := []*BarItem{
		{Content: "PATH", BGColor: statusBarStyle.Dirs.PathBG},
		{
			Content: dm.nav.Entry().Path,
			BGColor: statusBarStyle.BG,
			Wrapper: PrefixWrapString,
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
			{Content: string(dm.mode), BGColor: statusBarStyle.Dirs.ModeBG},
			{Content: "SIZE", BGColor: statusBarStyle.Dirs.SizeBG},
			{Content: FmtSizeColor(dm.summaryInfo.size, 0), BGColor: statusBarStyle.BG},
			{Content: "DIRS", BGColor: statusBarStyle.Dirs.DirsBG},
			{Content: unitFmt(dm.summaryInfo.dirs), BGColor: statusBarStyle.BG},
			{Content: "FILES", BGColor: statusBarStyle.Dirs.FilesBG},
			{Content: unitFmt(dm.summaryInfo.files), BGColor: statusBarStyle.BG},
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

// updateTableSize accepts the [tea.WindowSizeMsg] message and updates the
// directory entries table according to new width and height values. The
// message will be propagated to all nested models respectively.
func (dm *DirModel) updateTableSize(msg tea.WindowSizeMsg) {
	dm.width, dm.height = msg.Width, msg.Height

	dm.updateTableData()

	dm.diff.Update(msg)
	dm.filters.Update(msg)
	dm.topEntries.Update(msg)
	dm.cmd.Update(msg)
}

// dirTableWidth returns the directory entries table width with respect to the
// terminal window width.
func (dm *DirModel) dirTableWidth() int {
	return int(float64(dm.width) * dirsTableRatio)
}

// previewTableWidth returns the entry preview table width with respect to the
// directory entries table width.
func (dm *DirModel) previewTableWidth() int {
	return dm.width - dm.dirTableWidth()
}

func (dm *DirModel) viewProgress() string {
	// prevents showing 100% progress. Since the scanning progress might be
	// slowed down with a deep folder structure, but most heavy objects are
	// scanned already, we end up with 100% before the scan actually completes.
	completed := 0.99
	scannedSize := float64(dm.nav.Entry().Size)

	if dm.nav.currentDrive != nil {
		driveUsedSize := float64(dm.nav.currentDrive.UsedBytes)

		completed = scannedSize/driveUsedSize - 0.01
	}

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.scanPG.New(dm.width).ViewAs(completed),
	)
}

// sortEntries sorts directory entries based on the provided [drive.SortKey].
// It updates the current sort state and re-renders the directory entries
// according to the new sort key value and order.
func (dm *DirModel) sortEntries(sk drive.SortKey) {
	if dm.nav.OnDrives() {
		return
	}

	defer dm.updateTableData()

	if dm.sortState.Key == sk {
		dm.sortState.Desc = !dm.sortState.Desc

		return
	}

	dm.sortState = SortState{Key: sk}
}
