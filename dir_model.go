package main

import (
	"container/heap"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/crumbyte/noxdir/filter"
	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	topFilesTableHeight = 16
	colWidthRatio       = 0.13
)

type Mode string

const (
	PENDING Mode = "PENDING"
	READY   Mode = "READY"
	INPUT   Mode = "INPUT"
)

type DirModel struct {
	columns       []Column
	dirsTable     *table.Model
	topFilesTable *table.Model
	nav           *Navigation
	filters       filter.FiltersList
	mode          Mode
	lastErr       []error
	width         int
	height        int
	showTopFiles  bool
}

func NewDirModel(nav *Navigation) *DirModel {
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
		filters: filter.NewFiltersList(
			filter.NewNameFilter("type..."),
			&filter.DirsFilter{},
			&filter.FilesFilter{},
		),
		dirsTable:     buildTable(),
		topFilesTable: buildTable(),
		mode:          PENDING,
		nav:           nav,
	}

	style := table.DefaultStyles()
	style.Header = TableHeaderStyle
	style.Cell = lipgloss.NewStyle()
	style.Selected = lipgloss.NewStyle()

	dm.topFilesTable.SetStyles(style)
	dm.topFilesTable.SetHeight(topFilesTableHeight)

	return dm
}

func (dm *DirModel) Init() tea.Cmd {
	return nil
}

func (dm *DirModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case updateDirState:
		runtime.GC()
		dm.nav.Entry().CalculateSize()

		dm.updateTableData()
	case scanFinished:
		dm.mode = READY

		runtime.GC()
		dm.nav.Entry().CalculateSize()
		dm.updateTableData()

		dm.topFilesTable.SetRows(nil)
		dm.fillTopFiles()
	case tea.WindowSizeMsg:
		dm.updateSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if dm.nav.State() == Drives || dm.handleKeyBindings(msg) {
			return dm, nil
		}
	}

	if dm.nav.State() == Drives {
		return dm, nil
	}

	t, _ := dm.dirsTable.Update(msg)
	dm.dirsTable = &t

	return dm, tea.Batch(cmd)
}

func (dm *DirModel) View() string {
	h := lipgloss.Height

	keyBindings := dm.dirsTable.Help.FullHelpView(
		append(navigateKeyMap, dirsKeyMap...),
	)

	summary := dm.dirsSummary()

	dirsTableHeight := dm.height - h(keyBindings) - (h(summary) * 2)
	dm.dirsTable.SetHeight(dirsTableHeight)

	rows := []string{summary}

	if dm.showTopFiles {
		tft := dm.topFilesTable.View()

		dm.dirsTable.SetHeight(dirsTableHeight - h(tft))
		rows = append(rows, dm.dirsTable.View(), tft)
	} else {
		rows = append(rows, dm.dirsTable.View())
	}

	for _, f := range dm.filters {
		v, ok := f.(filter.Viewer)
		if !ok {
			continue
		}

		rendered := v.View()

		if len(rendered) > 0 {
			rows = append(rows, rendered)
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Top, append(rows, summary, keyBindings)...,
	)
}

func (dm *DirModel) handleKeyBindings(msg tea.KeyMsg) bool {
	bk := bindingKey(strings.ToLower(msg.String()))

	if bk == toggleNameFilter {
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

	switch bk {
	case explore:
		sr := dm.dirsTable.SelectedRow()
		if len(sr) < 2 {
			return true
		}

		if err := dm.nav.Explore(sr[1]); err != nil {
			return true
		}
	case toggleTopFiles:
		dm.showTopFiles = !dm.showTopFiles
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

func (dm *DirModel) updateTableData() {
	if dm.nav.State() == Drives || dm.nav.Entry() == nil || !dm.nav.Entry().IsDir {
		return
	}

	iconWidth := 5
	nameWidth := (dm.width - iconWidth) / 4

	colWidth := int(float64(dm.width-iconWidth-nameWidth) * colWidthRatio)
	progressWidth := dm.width - (colWidth * 5) - iconWidth - nameWidth

	columns := make([]table.Column, len(dm.columns))

	for i, c := range dm.columns {
		columns[i] = table.Column{Title: c.Title, Width: colWidth}
	}

	columns[0].Width = iconWidth
	columns[1].Width = 0
	columns[2].Width = nameWidth
	columns[len(columns)-1].Width = progressWidth

	dm.dirsTable.SetColumns(columns)
	dm.dirsTable.SetCursor(0)

	fillProgress := NewProgressBar(progressWidth, '🟥', ' ')

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
		name := child.Name()

		fmtName := lipgloss.NewStyle().MaxWidth(nameWidth - 5).Render(name)
		if lipgloss.Width(name) == nameWidth-5 {
			fmtName += "..."
		}

		rows = append(
			rows,
			table.Row{
				EntryIcon(child),
				name,
				fmtName,
				fmtSize(child.Size, true),
				totalDirs,
				totalFiles,
				time.Unix(child.ModTime, 0).Format("2006-01-02 15:04"),
				strconv.FormatFloat(parentUsage*100, 'f', 2, 64) + " %",
				pgBar,
			},
		)
	}

	dm.dirsTable.SetRows(rows)
	dm.dirsTable.SetCursor(dm.nav.cursor)
}

func (dm *DirModel) dirsSummary() string {
	items := []*BarItem{
		NewBarItem("PATH", "#FF5F87", 0),
		NewBarItem(dm.nav.Entry().Path, "", -1),
		NewBarItem(string(dm.mode), "#FF8531", 0),
		NewBarItem("SIZE", "#FF5F87", 0),
		DefaultBarItem(fmtSize(dm.nav.Entry().Size, false)),
		NewBarItem("DIRS", "#FF5F87", 0),
		DefaultBarItem(unitFmt(dm.nav.Entry().LocalDirs)),
		NewBarItem("FILES", "#FF5F87", 0),
		DefaultBarItem(unitFmt(dm.nav.Entry().LocalFiles)),
		NewBarItem("ERRORS", "#FF303E", 0),
		DefaultBarItem(unitFmt(uint64(len(dm.lastErr)))),
	}

	return statusBarStyle.Margin(1, 0, 1, 0).
		Render(NewStatusBar(items, dm.width))
}

func (dm *DirModel) fillTopFiles() {
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

	dm.topFilesTable.SetColumns(columns)
	dm.topFilesTable.SetCursor(0)

	if structure.TopFilesInstance.Len() == 0 || dm.topFilesTable.Rows() != nil {
		return
	}

	rows := make([]table.Row, 15)
	heap.Pop(&structure.TopFilesInstance)

	for i := len(rows) - 1; i >= 0; i-- {
		file, ok := heap.Pop(&structure.TopFilesInstance).(*structure.Entry)
		if !ok {
			continue
		}

		path := strings.TrimSuffix(
			strings.TrimPrefix(file.Path, dm.nav.currentDrive.Path),
			file.Name(),
		)

		rows[i] = table.Row{
			EntryIcon(file),
			file.Path,
			path + topFileStyle.Render(file.Name()),
			fmtSize(file.Size, true),
			time.Unix(file.ModTime, 0).Format("2006-01-02 15:04"),
		}
	}

	dm.topFilesTable.SetRows(rows)
	dm.topFilesTable.SetCursor(0)
}

func (dm *DirModel) updateSize(width, height int) {
	dm.width, dm.height = width, height

	dm.dirsTable.SetWidth(width)
	dm.topFilesTable.SetWidth(width)

	dm.updateTableData()
	dm.fillTopFiles()
}
