package render

import (
	"fmt"
	"slices"

	"github.com/crumbyte/noxdir/structure"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DeleteHandler interface {
	Delete(entry *structure.Entry) error
}

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
	deleteHandler DeleteHandler
	toDelete      []*structure.Entry
	choice        DeleteChoice
}

func NewDeleteDialogModel(dh DeleteHandler, toDelete []*structure.Entry) *DeleteDialogModel {
	return &DeleteDialogModel{
		deleteHandler: dh,
		choice:        CancelChoice,
		toDelete:      toDelete,
	}
}

func (ddm *DeleteDialogModel) Init() tea.Cmd {
	return nil
}

func (ddm *DeleteDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
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
			for _, entry := range ddm.toDelete {
				err = ddm.deleteHandler.Delete(entry)
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

func (ddm *DeleteDialogModel) View() tea.View {
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
	namesList := make([]string, 0, len(ddm.toDelete))

	for _, entry := range ddm.toDelete {
		name := entry.Name()

		if len(name) > maxDeleteEntryLength {
			name = name[:maxDeleteEntryLength] + "..."
		}

		namesList = append(namesList, name)

		totalReclaimed += entry.Size
	}

	slices.Sort(namesList)

	if len(namesList) >= maxEntriesDisplay {
		namesList = namesList[:maxEntriesDisplay]

		namesList = append(
			namesList,
			fmt.Sprintf("%d more...", len(ddm.toDelete)-maxEntriesDisplay),
		)
	}

	target := textStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center, namesList...),
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

	return tea.NewView(
		style.DialogBox().BorderForeground(lipgloss.Color("#FF303E")).Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.JoinVertical(lipgloss.Top, confirm, target, reclaimed),
				buttons,
			),
		),
	)
}
