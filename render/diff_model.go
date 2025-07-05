package render

import (
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render/table"
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	UpdateDiffState  struct{}
	DiffScanFinished struct{}
)

type DiffModel struct {
	nav          *Navigation
	table        *table.Model
	lastRootPath string
	pg           *PG
	targetTree   *structure.Tree
	diff         *structure.Diff
	columns      []table.Column
	height       int
	width        int
	ready        bool
}

func NewDiffModel(n *Navigation) *DiffModel {
	return &DiffModel{
		nav:   n,
		table: buildTable(),
		pg:    &style.CS().ScanProgressBar,
		columns: []table.Column{
			{Title: ""},
			{Title: ""},
			{Title: ""},
			{Title: "Name"},
			{Title: "Size"},
		},
	}
}

func (dm *DiffModel) Init() tea.Cmd {
	return nil
}

func (dm *DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dm.height = int(float64(msg.Height) * 0.7)
		dm.width = int(float64(msg.Width) * 0.7)

		dm.table.SetWidth(dm.width)
		dm.table.SetHeight(dm.height)

		dm.updateTableData()

		return dm, nil
	case UpdateDiffState:
		dm.ready = false
		runtime.GC()

		dm.updateTableData()
	case DiffScanFinished:
		dm.ready = true

		runtime.GC()
		dm.updateTableData()
	case tea.KeyMsg:
		bk := bindingKey(strings.ToLower(msg.String()))

		if bk == explore {
			dm.handleExploreKey()
		}
	}

	t, _ := dm.table.Update(msg)
	dm.table = &t

	return dm, nil
}

func (dm *DiffModel) View() string {
	hasDiff := dm.diff != nil && !dm.diff.Empty()

	rows, message := make([]string, 0, 3), "No delta found for: "+dm.nav.Entry().Path

	messageStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(dm.width).
		Bold(true)

	if !dm.ready {
		rows = append(
			rows,
			dm.viewProgress(),
			messageStyle.Render("Scanning the delta for: "+dm.nav.Entry().Path),
		)
	}

	if dm.ready && hasDiff {
		total := dm.viewStats()
		dm.table.SetHeight(dm.height - lipgloss.Height(total))

		rows = append(
			rows,
			messageStyle.Render("Current delta for: "+dm.nav.Entry().Path),
			dm.table.View(),
			total,
		)
	}

	if dm.ready && !hasDiff {
		rows = append(rows, messageStyle.Render(message))
	}

	return style.DialogBox().Render(
		lipgloss.NewStyle().Padding(0, 1, 0, 1).Render(
			lipgloss.JoinVertical(lipgloss.Top, rows...),
		),
	)
}

func (dm *DiffModel) Run(width, height int) {
	dm.height = int(float64(height) * 0.7)
	dm.width = int(float64(width) * 0.7)

	dm.table.SetWidth(dm.width)
	dm.table.SetHeight(dm.height)

	dm.diff, dm.targetTree = nil, nil

	tree, done, errChan := dm.nav.Diff()

	dm.targetTree = tree

	go func() {
		ticker := time.NewTicker(updateTickerInterval)
		defer func() {
			ticker.Stop()
		}()

		teaProg.Send(UpdateDiffState{})

		for {
			select {
			case <-errChan:
			case <-ticker.C:
				teaProg.Send(UpdateDiffState{})

				dm.targetTree.CalculateSize()
			case dm.diff = <-done:
				teaProg.Send(DiffScanFinished{})

				dm.targetTree.CalculateSize()

				return
			}
		}
	}()
}

func (dm *DiffModel) handleExploreKey() bool {
	sr := dm.table.SelectedRow()
	if len(sr) < 2 {
		return true
	}

	return drive.Explore(sr[2]) != nil
}

func (dm *DiffModel) updateTableData() {
	if dm.targetTree == nil || dm.diff == nil {
		return
	}

	removedIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF303E")).
		Render("---  ")

	addedIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06923E")).
		Render("+++  ")

	iconWidth := 5
	signWidth := 5
	nameWidth := dm.width - iconWidth - signWidth - 20

	dm.columns[0].Width = signWidth
	dm.columns[1].Width = iconWidth
	dm.columns[2].Width = 0
	dm.columns[3].Width = nameWidth
	dm.columns[4].Width = 20

	dm.table.SetColumns(dm.columns)

	if len(dm.table.Rows()) > 0 && dm.lastRootPath == dm.targetTree.Root().Path {
		return
	}

	rows := make([]table.Row, 0, len(dm.nav.Entry().Child))
	dm.nav.Entry().SortChild()

	for _, child := range dm.diff.Added {
		rows = append(
			rows,
			table.Row{
				addedIcon,
				EntryIcon(child),
				child.Path,
				FmtName(child.Path, nameWidth),
				FmtSize(child.Size, entrySizeWidth),
			},
		)
	}

	for _, child := range dm.diff.Removed {
		rows = append(
			rows,
			table.Row{
				removedIcon,
				EntryIcon(child),
				child.Path,
				FmtName(child.Path, nameWidth),
				FmtSize(child.Size, entrySizeWidth),
			},
		)
	}

	dm.table.SetRows(rows)
	dm.table.SetCursor(0)

	dm.lastRootPath = dm.targetTree.Root().Path
}

func (dm *DiffModel) viewStats() string {
	if dm.diff == nil {
		return ""
	}

	addedDirs, addedFiles, addedSize := structure.DiffStats(dm.diff.Added)
	remDirs, remFiles, remSize := structure.DiffStats(dm.diff.Removed)

	statStyle := lipgloss.NewStyle().Bold(true).Underline(true)

	addedStats := lipgloss.JoinHorizontal(
		lipgloss.Center,
		"ADDED: ",
		statStyle.Render(FmtSize(addedSize, 0)),
		", directories - ",
		statStyle.Render(strconv.Itoa(addedDirs)),
		", files - ",
		statStyle.Render(strconv.Itoa(addedFiles)),
	)

	removedStats := lipgloss.JoinHorizontal(
		lipgloss.Center,
		"REMOVED: ",
		statStyle.Render(FmtSize(remSize, 0)),
		", directories - ",
		statStyle.Render(strconv.Itoa(remDirs)),
		", files - ",
		statStyle.Render(strconv.Itoa(remFiles)),
	)

	return lipgloss.NewStyle().Width(dm.width).
		MarginTop(0).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true).
		Align(lipgloss.Center).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, addedStats, " | ", removedStats))
}

func (dm *DiffModel) viewProgress() string {
	completed := (float64(dm.targetTree.Root().Size) / float64(dm.nav.entry.Size)) - 0.01

	return style.StatusBar().Margin(1, 0, 1, 0).Render(
		dm.pg.New(dm.width).ViewAs(completed),
	)
}
