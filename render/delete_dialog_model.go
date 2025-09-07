package render

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DeleteChoice int

const (
	CancelChoice DeleteChoice = iota
	ConfirmChoice

	deleteDialogWidth    = 50
	maxDeleteEntryLength = 40
	maxEntriesDisplay    = 20
)

type EntryDeleted struct {
	Err     error
	Deleted bool
}

type DeleteDialogModel struct {
	nav     *Navigation
	pathMap map[string]int64
	choice  DeleteChoice
}

func NewDeleteDialogModel(nav *Navigation, pathMap map[string]int64) *DeleteDialogModel {
	return &DeleteDialogModel{
		choice:  CancelChoice,
		pathMap: pathMap,
		nav:     nav,
	}
}

func (ddm *DeleteDialogModel) Init() tea.Cmd {
	return nil
}

func (ddm *DeleteDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return ddm, nil
	}

	switch {
	case keyMsg.String() == "enter":
		var (
			err     error
			deleted bool
		)

		if ddm.choice == ConfirmChoice {
			for path := range ddm.pathMap {
				err = ddm.nav.Delete(path)
			}

			deleted = true
		}

		go func() {
			teaProg.Send(EntryDeleted{Err: err, Deleted: deleted})
		}()
	case keyMsg.String() == "left":
		ddm.choice = CancelChoice
	case keyMsg.String() == "right":
		ddm.choice = ConfirmChoice
	}

	return ddm, nil
}

func (ddm *DeleteDialogModel) View() string {
	cancelBtn, confirmBtn := style.ActiveButton(), style.ConfirmButton()

	if ddm.choice == ConfirmChoice {
		cancelBtn, confirmBtn = confirmBtn, cancelBtn
	}

	textStyle := lipgloss.NewStyle().
		Width(deleteDialogWidth).
		Align(lipgloss.Center).
		Bold(true)

	confirm := textStyle.
		Foreground(lipgloss.Color("#FF303E")).
		Render("Confirm Deletion\n")

	totalReclaimed := int64(0)
	pathList := make([]string, 0, len(ddm.pathMap))

	for path, size := range ddm.pathMap {
		if len(path) > maxDeleteEntryLength {
			path = path[:maxDeleteEntryLength] + "..."
		}

		pathList = append(pathList, path)

		totalReclaimed += size
	}

	slices.Sort(pathList)

	if len(pathList) >= maxEntriesDisplay {
		pathList = pathList[:maxEntriesDisplay]

		pathList = append(
			pathList,
			fmt.Sprintf("%d more...", len(ddm.pathMap)-maxEntriesDisplay),
		)
	}

	target := textStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center, pathList...),
	)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		cancelBtn.Render("No"),
		confirmBtn.Render("Yes"),
	)

	reclaimed := textStyle.
		Foreground(lipgloss.Color("#80ed99")).
		MarginTop(1).
		Render(FmtSize(totalReclaimed, 0), "to be reclaimed")

	return style.DialogBox().BorderForeground(
		lipgloss.Color("#FF303E"),
	).Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Top, confirm, target, reclaimed),
			buttons,
		),
	)
}
