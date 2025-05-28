package render

import (
	"strings"
	"time"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/muesli/termenv"
)

const updateTickerInterval = time.Millisecond * 500

type (
	UpdateDirState struct{}
	ScanFinished   struct{}
)

var teaProg *tea.Program

type ViewModel struct {
	driveModel *DriveModel
	dirModel   *DirModel
	nav        *Navigation
	lastErr    []error
}

func NewViewModel(n *Navigation) *ViewModel {
	return &ViewModel{
		lastErr:    make([]error, 0),
		nav:        n,
		driveModel: NewDriveModel(n),
		dirModel:   NewDirModel(n),
	}
}

func (vm *ViewModel) Init() tea.Cmd {
	return tea.Batch(tea.DisableMouse)
}

func (vm *ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		bk := bindingKey(strings.ToLower(msg.String()))

		if vm.dirModel.mode == INPUT {
			break
		}

		switch bk {
		case quit, cancel:
			return vm, tea.Quit
		case enter:
			vm.levelDown()
		case backspace:
			vm.levelUp()
		}
	}

	vm.driveModel.Update(msg)
	vm.dirModel.Update(msg)

	return vm, tea.Batch(cmd)
}

func (vm *ViewModel) View() string {
	if vm.nav.OnDrives() {
		return vm.driveModel.View()
	}

	return vm.dirModel.View()
}

func (vm *ViewModel) levelDown() {
	sr := vm.dirModel.dirsTable.SelectedRow()
	cursor := vm.dirModel.dirsTable.Cursor()

	if vm.nav.OnDrives() {
		sr = vm.driveModel.drivesTable.SelectedRow()
	}

	done, errChan := vm.nav.Down(
		sr[1],
		cursor,
		func(_ *structure.Entry, _ State) {
			vm.dirModel.filters.Reset()
			vm.dirModel.updateTableData()
		},
	)

	if done == nil {
		return
	}

	go func() {
		vm.lastErr = []error{}

		ticker := time.NewTicker(updateTickerInterval)
		defer func() {
			ticker.Stop()
		}()

		teaProg.Send(UpdateDirState{})

		for {
			select {
			case err := <-errChan:
				if err != nil {
					vm.lastErr = append(vm.lastErr, err)
				}
			case <-ticker.C:
				teaProg.Send(UpdateDirState{})
			case <-done:
				teaProg.Send(ScanFinished{})

				return
			}
		}
	}()
}

func (vm *ViewModel) levelUp() {
	vm.nav.Up(func(_ *structure.Entry, _ State) {
		if vm.nav.OnDrives() {
			vm.driveModel.resetSort()
			vm.driveModel.updateTableData(drive.TotalUsedP, true)

			return
		}

		vm.dirModel.filters.Reset()
		vm.dirModel.updateTableData()
	})
}

func NewProgressBar(width int, full, empty rune) progress.Model {
	maxCharLen := max(
		lipgloss.Width(string(full)),
		lipgloss.Width(string(empty)),
	)

	return progress.New(
		progress.WithColorProfile(termenv.Ascii),
		progress.WithWidth(width/maxCharLen),
		progress.WithFillCharacters(full, empty),
		progress.WithoutPercentage(),
	)
}

func SetTeaProgram(tp *tea.Program) {
	teaProg = tp
}

func buildTable() *table.Model {
	tbl := table.New(table.WithFocused(true))

	style := table.DefaultStyles()
	style.Header = TableHeaderStyle
	style.Cell = lipgloss.NewStyle()
	style.Selected = SelectedRowStyle

	tbl.SetStyles(style)

	tbl.Help = help.New()

	return &tbl
}
