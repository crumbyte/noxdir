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
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#353533")).
		Padding(0, 3).
		Margin(1, 3)

	activeButtonStyle := buttonStyle.
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Underline(true)

	cancelBtn := activeButtonStyle.Render("No")
	confirmBtn := buttonStyle.Render("Yes")

	if ddm.choice == ConfirmChoice {
		confirmBtn = activeButtonStyle.Render("Yes")
		cancelBtn = buttonStyle.Render("No")
	}

	question := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Render("Confirm the deletion of: \n " + fmtName(ddm.targetPath, 40))

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, cancelBtn, confirmBtn)

	return dialogBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center, question, buttons),
	)
}
