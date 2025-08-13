package render

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	InfoTitle  = "Info"
	ErrorTitle = "Error"
)

const (
	popupWidth = 40
	queueLimit = 5
)

// PopupStyles contains the styles for the pop-up window elements.
type PopupStyles struct {
	// Title contains style settings for the title which is shown at the top of
	// the pop-up window.
	Title lipgloss.Style

	// Message contains the style setting for the pop-up text message.
	Message lipgloss.Style

	// Box contains the style setting for the pop-up dialog box including border
	// settings, background, etc.
	Box lipgloss.Style
}

// PopupModel is responsible for rendering the pop-up window message. The rendered
// window will be shown for a period defined during the creation of a new model
// instance. The messages will be pushed into a limited-size queue and displayed
// in a sequence, where each message has its visibility duration.
type PopupModel struct {
	mx            sync.RWMutex
	messageQueue  []string
	duration      time.Duration
	ttl           time.Time
	styles        *PopupStyles
	title         string
	queueLimit    int
	hideCountdown bool
	visible       bool
}

// NewPopupModel creates a new *PopupModel instance with the provided title and
// visibility duration. The *PopupStyles are optional, and if the value is empty,
// the DefaultPopupStyle will be used.
func NewPopupModel(title string, d time.Duration, ps *PopupStyles) *PopupModel {
	if len(title) == 0 {
		title = InfoTitle
	}

	if ps == nil {
		ps = DefaultPopupStyle()
	}

	return &PopupModel{
		title:        title,
		duration:     d,
		styles:       ps,
		messageQueue: make([]string, 0, queueLimit),
		queueLimit:   queueLimit,
	}
}

func DefaultPopupStyle() *PopupStyles {
	defaultBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	if style != nil {
		defaultBoxStyle = *style.DialogBox()
	}

	return &PopupStyles{
		Title:   lipgloss.NewStyle().Bold(true).PaddingBottom(2),
		Box:     defaultBoxStyle,
		Message: lipgloss.NewStyle(),
	}
}

func PopupDefaultErrorStyle() *PopupStyles {
	ps := DefaultPopupStyle()

	ps.Title = ps.Title.Foreground(lipgloss.Color("#FF303E"))
	ps.Message = lipgloss.NewStyle().Align(lipgloss.Center)
	ps.Box = style.DialogBox().
		Padding(0, 1, 0, 1).
		Width(popupWidth).
		BorderForeground(lipgloss.Color("#FF303E"))

	return ps
}

func WithPopupStyle(title, message, box lipgloss.Style) *PopupStyles {
	return &PopupStyles{
		Title:   title,
		Message: message,
		Box:     box,
	}
}

func (pm *PopupModel) Init() tea.Cmd {
	return nil
}

func (pm *PopupModel) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	if !pm.visible || !time.Now().After(pm.ttl) {
		return pm, nil
	}

	pm.mx.Lock()
	defer pm.mx.Unlock()

	if len(pm.messageQueue) > 0 {
		pm.ttl = time.Now().Add(pm.duration)
		pm.messageQueue = pm.messageQueue[1:]

		return pm, nil
	}

	pm.visible = false

	return pm, nil
}

func (pm *PopupModel) View() string {
	nextMessage, ok := pm.nextMessage()

	if !pm.visible || !ok {
		return ""
	}

	messages := []string{nextMessage}

	if !pm.hideCountdown {
		closeIn := max(time.Until(pm.ttl).Seconds(), 0)

		messages = append(
			messages, fmt.Sprintf("close in %.0f seconds", closeIn),
		)
	}

	return pm.styles.Box.Align(lipgloss.Center).Render(
		pm.styles.Title.Render(pm.title),
		lipgloss.JoinVertical(lipgloss.Center, messages...),
	)
}

// Style returns the current pop-up window elements styles.
func (pm *PopupModel) Style() *PopupStyles {
	return pm.styles
}

// Visible checks whether the pop-up window is visible and contains any messages
// in its queue.
func (pm *PopupModel) Visible() bool {
	return pm.visible && len(pm.messageQueue) != 0
}

// Show triggers the pop-up window to be shown with the provided message. If the
// message queue is already full, then the most recent message will be replaced
// with the new one.
func (pm *PopupModel) Show(message string) {
	pm.mx.Lock()
	defer pm.mx.Unlock()

	if message = strings.TrimSpace(message); len(message) == 0 {
		return
	}

	if len(pm.messageQueue) >= pm.queueLimit {
		pm.messageQueue[len(pm.messageQueue)-1] = message
	} else {
		pm.messageQueue = append(pm.messageQueue, message)
	}

	pm.visible, pm.ttl = true, time.Now().Add(pm.duration)
}

// Hide hides the pop-up window and clears the message queue.
func (pm *PopupModel) Hide() {
	pm.mx.Lock()
	defer pm.mx.Unlock()

	pm.messageQueue = make([]string, 0, len(pm.messageQueue))
	pm.visible = false
}

func (pm *PopupModel) MessageQueueSnapshot() []string {
	pm.mx.RLock()
	defer pm.mx.RUnlock()

	return append([]string(nil), pm.messageQueue...)
}

func (pm *PopupModel) nextMessage() (string, bool) {
	pm.mx.RLock()
	defer pm.mx.RUnlock()

	if len(pm.messageQueue) == 0 {
		return "", false
	}

	return pm.messageQueue[0], true
}
