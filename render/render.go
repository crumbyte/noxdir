package render

import (
	"time"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render/table"
	"github.com/crumbyte/noxdir/structure"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	Version = "v0.10.0"

	updateTickerInterval = time.Millisecond * 500
)

type (
	Tick           struct{}
	UpdateDirState struct{}
	ScanFinished   struct{ Mode Mode }
	EnqueueRefresh struct{ Mode Mode }
)

var teaProg *tea.Program

type ViewModel struct {
	driveModel *DriveModel
	dirModel   *DirModel
	nav        *Navigation
	lastErr    []error
}

func NewViewModel(n *Navigation, driveModel *DriveModel, dirMode *DirModel) *ViewModel {
	return &ViewModel{
		lastErr:    make([]error, 0),
		nav:        n,
		driveModel: driveModel,
		dirModel:   dirMode,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(_ time.Time) tea.Msg {
		return Tick{}
	})
}

func (vm *ViewModel) Init() tea.Cmd {
	return tea.Batch(tea.DisableMouse, tickCmd())
}

func (vm *ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case Tick:
		vm.driveModel.Update(msg)
		vm.dirModel.Update(msg)

		return vm, tea.Batch(tickCmd(), cmd)
	case EnqueueRefresh:
		vm.refresh(msg.Mode)
	case tea.KeyMsg:
		if vm.dirModel.mode == INPUT || vm.dirModel.mode == CMD {
			break
		}

		switch {
		case key.Matches(msg, Bindings.Config):
			if err := drive.Explore(vm.nav.Settings().ConfigPath()); err != nil {
				return vm, nil
			}
		case key.Matches(msg, Bindings.Refresh):
			vm.refresh(READY)
		case key.Matches(msg, Bindings.Quit):
			return vm, tea.Quit
		case key.Matches(msg, Bindings.Drive.LevelDown):
			if vm.dirModel.mode == READY || vm.nav.OnDrives() {
				vm.levelDown()
			}
		case key.Matches(msg, Bindings.Dirs.LevelUp):
			if vm.dirModel.mode == READY || vm.nav.OnDrives() {
				vm.levelUp()
			}
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
		vm.driveModel.drivesTable.ResetMarked()
		sr = vm.driveModel.drivesTable.SelectedRow()
	}

	if len(sr) < 2 {
		return
	}

	done, errChan := vm.nav.Down(
		sr[1],
		cursor,
		func(_ *structure.Entry, _ State) {
			vm.dirModel.dirsTable.ResetMarked()
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
				teaProg.Send(ScanFinished{Mode: READY})

				return
			}
		}
	}()
}

func (vm *ViewModel) levelUp() {
	vm.nav.Up(func(_ *structure.Entry, _ State) {
		if vm.nav.OnDrives() {
			vm.driveModel.drivesTable.ResetMarked()
			vm.driveModel.resetSort()
			vm.driveModel.updateTableData(drive.TotalUsedP, true)

			return
		}

		vm.dirModel.dirsTable.ResetMarked()
		vm.dirModel.filters.Reset()
		vm.dirModel.updateTableData()
	})
}

func (vm *ViewModel) refresh(mode Mode) {
	if vm.nav.OnDrives() {
		vm.nav.RefreshDrives()

		vm.driveModel.drivesTable.ResetMarked()
		vm.driveModel.Update(RefreshDrives{})

		return
	}

	done, errChan, err := vm.nav.RefreshEntry()
	if err != nil {
		// TODO: the error might occur only if there were no directories in stack
		return
	}

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
			case err = <-errChan:
				if err != nil {
					vm.lastErr = append(vm.lastErr, err)
				}
			case <-ticker.C:
				teaProg.Send(UpdateDirState{})
			case <-done:
				teaProg.Send(ScanFinished{Mode: mode})

				return
			}
		}
	}()
}

func SetTeaProgram(tp *tea.Program) {
	teaProg = tp
}

func buildTable() *table.Model {
	tbl := table.New()

	s := table.DefaultStyles()
	s.Header = *style.TableHeader()
	s.Selected = *style.SelectedRow()
	s.Marked = *style.MarkedRow()
	s.Cell = lipgloss.NewStyle().Foreground(lipgloss.Color(style.CS().CellText))

	tbl.SetStyles(s)

	tbl.Help = help.New()

	return &tbl
}
