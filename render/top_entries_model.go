package render

import (
	"container/heap"
	"path/filepath"
	"time"

	"github.com/crumbyte/noxdir/render/table"
	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	topRowsNumber = 15
	topDirsTitle  = "Top Directories"
	topFilesTitle = "Top Files"
)

type TopEntries struct {
	filesTable   *table.Model
	dirsTable    *table.Model
	height       int
	width        int
	showTopFiles bool
	showTopDirs  bool
}

func NewTopEntries() *TopEntries {
	te := &TopEntries{
		filesTable: buildTable(),
		dirsTable:  buildTable(),
	}

	s := table.DefaultStyles()
	s.Header = *style.TopTableHeader()
	s.Cell = lipgloss.NewStyle()
	s.Selected = lipgloss.NewStyle()

	te.filesTable.SetStyles(s)
	te.filesTable.SetHeight(topFilesTableHeight)

	te.dirsTable.SetStyles(s)
	te.dirsTable.SetHeight(topFilesTableHeight)

	return te
}

func (te *TopEntries) Init() tea.Cmd {
	return nil
}

func (te *TopEntries) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		te.height = msg.Height
		te.width = msg.Width

		te.UpdateTopEntries()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Bindings.Dirs.TopDirs):
			te.showTopDirs = !te.showTopDirs && !te.showTopFiles
		case key.Matches(msg, Bindings.Dirs.TopFiles):
			te.showTopFiles = !te.showTopFiles && !te.showTopDirs
		}
	}

	return te, nil
}

func (te *TopEntries) View() string {
	if !te.showTopDirs && !te.showTopFiles {
		return ""
	}

	topTable := te.filesTable

	if te.showTopDirs {
		topTable = te.dirsTable
	}

	return topTable.View()
}

func (te *TopEntries) UpdateTopEntries() {
	te.setEntries(
		structure.TopEntriesInstance.Files(), te.filesTable, topFilesTitle,
	)

	te.setEntries(
		structure.TopEntriesInstance.Dirs(), te.dirsTable, topDirsTitle,
	)
}

func (te *TopEntries) Clear() {
	te.filesTable.SetRows(nil)
	te.dirsTable.SetRows(nil)
}

func (te *TopEntries) setEntries(entries heap.Interface, tm *table.Model, title string) {
	iconWidth := 5

	colSize := int(float64(te.width-iconWidth) * colWidthRatio)
	nameWidth := te.width - (colSize * 2) - iconWidth

	columns := []table.Column{
		{Title: "", Width: iconWidth},
		{Title: "", Width: 0},
		{Title: title, Width: nameWidth},
		{Title: "Size", Width: colSize},
		{Title: "Last Change", Width: colSize},
	}

	tm.SetColumns(columns)
	tm.SetCursor(0)

	if entries.Len() == 0 && len(tm.Rows()) == 0 {
		return
	}

	if te.rerenderExistingRows(tm, nameWidth) {
		return
	}

	heap.Pop(entries)
	rows := make([]table.Row, topRowsNumber)

	for i := len(rows) - 1; i >= 0; i-- {
		file, ok := heap.Pop(entries).(*structure.Entry)
		if !ok {
			continue
		}

		filePath := WrapPath(file.Path, nameWidth)

		filePath = filepath.Join(
			filepath.Dir(filePath),
			style.TopFiles().Render(filepath.Base(filePath)),
		)

		rows[i] = table.Row{
			EntryIcon(file),
			file.Path,
			filePath,
			FmtSizeColor(file.Size, entrySizeWidth, colSize),
			time.Unix(file.ModTime, 0).Format("Jan 02 15:04"),
		}
	}

	tm.SetRows(rows)
	tm.SetCursor(0)
}

func (te *TopEntries) rerenderExistingRows(tm *table.Model, nameWidth int) bool {
	if tm.Rows() == nil {
		return false
	}

	rows := make([]table.Row, topRowsNumber)

	for i, r := range tm.Rows() {
		if len(r) < 2 {
			continue
		}

		filePath := WrapPath(r[1], nameWidth)

		filePath = filepath.Join(
			filepath.Dir(filePath),
			style.TopFiles().Render(filepath.Base(filePath)),
		)

		rows[i] = table.Row{r[0], r[1], filePath, r[3], r[4]}
	}

	tm.SetRows(rows)
	tm.SetCursor(0)

	return true
}
