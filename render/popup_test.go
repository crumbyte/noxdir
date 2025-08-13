package render_test

import (
	"testing"
	"time"

	"github.com/crumbyte/noxdir/render"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestPopupModel_Show_AddsMessage(t *testing.T) {
	pm := render.NewPopupModel("Test", time.Second, nil)

	pm.Show("msg1")

	assert.True(t, pm.Visible())
	assert.Equal(t, []string{"msg1"}, pm.MessageQueueSnapshot())
}

func TestPopupModel_Show_ReplacesWhenFull(t *testing.T) {
	pm := render.NewPopupModel("Test", time.Second, nil)

	for i := 1; i <= 5; i++ {
		pm.Show("msg" + string(rune('0'+i)))
	}

	pm.Show("newest")

	q := pm.MessageQueueSnapshot()

	assert.Len(t, q, 5)
	assert.Equal(t, "newest", q[len(q)-1], "last item should be replaced")
}

func TestPopupModel_Update_RemovesMessagesOnTTL(t *testing.T) {
	pm := render.NewPopupModel("Test", time.Millisecond*50, nil)

	pm.Show("msg1")
	pm.Show("msg2")

	time.Sleep(time.Millisecond * 60)

	pm.Update(nil)

	q := pm.MessageQueueSnapshot()

	assert.Len(t, q, 1, "first message should be removed after TTL")
	assert.Equal(t, "msg2", q[0])
}

func TestPopupModel_View_ReturnsEmptyWhenNotVisible(t *testing.T) {
	pm := render.NewPopupModel("Test", time.Second, nil)

	assert.Empty(t, pm.View())

	pm.Show(" ")
	assert.False(t, pm.Visible())

	pm.Show("msg1")
	assert.True(t, pm.Visible())

	pm.Hide()
	assert.Empty(t, pm.View())
}

func TestPopupModel_View_ShowsCountdown(t *testing.T) {
	style := render.WithPopupStyle(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
	)

	pm := render.NewPopupModel("Test", time.Second, style)
	pm.Show("hello")

	view := pm.View()
	assert.Contains(t, view, "hello")
	assert.Contains(t, view, "close in")
}
