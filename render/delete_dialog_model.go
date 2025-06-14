package render

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DeleteChoice int

const (
	CancelChoice DeleteChoice = iota
	ConfirmChoice

	deleteDialogWidth = 50
	deleteNameWidth   = 40
)

type EntryDeleted struct {
	Err     error
	Deleted bool
}

type DeleteDialogModel struct {
	nav        *Navigation
	targetPath string
	choice     DeleteChoice
}

func NewDeleteDialogModel(nav *Navigation, targetPath string) *DeleteDialogModel {
	return &DeleteDialogModel{
		choice:     CancelChoice,
		targetPath: targetPath,
		nav:        nav,
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

	bk := bindingKey(strings.ToLower(keyMsg.String()))
	switch bk {
	case enter:
		var (
			err     error
			deleted bool
		)

		if ddm.choice == ConfirmChoice {
			deleted, err = true, ddm.nav.Delete(ddm.targetPath)
		}

		go func() {
			teaProg.Send(EntryDeleted{Err: err, Deleted: deleted})
		}()
	case left:
		ddm.choice = CancelChoice
	case right:
		ddm.choice = ConfirmChoice
	}

	return ddm, nil
}

func (ddm *DeleteDialogModel) View() string {
	cancelBtn, confirmBtn := activeButtonStyle, confirmButtonStyle

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

	target := textStyle.Render(ddm.targetPath)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		cancelBtn.Render("No"),
		confirmBtn.Render("Yes"),
	)

	return dialogBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Top, confirm, target),
			buttons,
		),
	)
}
