package render

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultTitle      = "Info"
	defaultPopupWidth = 40
)

type PopupStyles struct {
	Title   lipgloss.Style
	Message lipgloss.Style
	Box     lipgloss.Style
}

type PopupModel struct {
	title        string
	messageQueue []string
	duration     time.Duration
	ttl          time.Time
	styles       PopupStyles
	visible      bool
}

func NewPopupModel(duration time.Duration) *PopupModel {
	return &PopupModel{
		title:    defaultTitle,
		duration: duration,
		styles: PopupStyles{
			Title:   lipgloss.NewStyle().Bold(true).PaddingBottom(2),
			Box:     *style.DialogBox(),
			Message: lipgloss.NewStyle(),
		},
	}
}

func NewErrorPopupModel(duration time.Duration) *PopupModel {
	pm := NewPopupModel(duration)

	pm.title = "Error"

	pm.styles.Title = pm.styles.Title.Foreground(lipgloss.Color("#FF303E"))
	pm.styles.Message = lipgloss.NewStyle().Align(lipgloss.Center)

	pm.styles.Box = style.DialogBox().
		Padding(0, 1, 0, 1).
		Width(defaultPopupWidth).
		BorderForeground(lipgloss.Color("#FF303E"))

	return pm
}

func (pm *PopupModel) Init() tea.Cmd {
	return nil
}

func (pm *PopupModel) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	if !pm.visible {
		return pm, nil
	}

	if time.Now().After(pm.ttl) {
		if len(pm.messageQueue) > 0 {
			pm.messageQueue = pm.messageQueue[1:]
			pm.ttl = time.Now().Add(pm.duration)

			return pm, nil
		}

		pm.visible = false
	}

	return pm, nil
}

func (pm *PopupModel) View() string {
	if !pm.visible {
		return ""
	}

	if len(pm.messageQueue) == 0 {
		return ""
	}

	closeIn := max(time.Until(pm.ttl).Seconds(), 0)

	return pm.styles.Box.Align(lipgloss.Center).Render(
		pm.styles.Title.Render(pm.title),
		lipgloss.JoinVertical(
			lipgloss.Center,
			pm.styles.Message.Render(pm.messageQueue[0]),
			fmt.Sprintf("close in %0.f seconds", closeIn),
		),
	)
}

func (pm *PopupModel) Show(message string) {
	pm.messageQueue = append(pm.messageQueue, message)
	pm.ttl = time.Now().Add(pm.duration)
	pm.visible = true
}
