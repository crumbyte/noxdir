package render_test

import (
	"testing"

	"github.com/crumbyte/noxdir/render"

	"github.com/stretchr/testify/require"
)

func TestNewStatusBar(t *testing.T) {
	render.InitStyle(render.DefaultColorSchema())

	sb := render.NewStatusBar()

	t.Run("default", func(t *testing.T) {
		sb.Clear()

		sb.Add([]*render.BarItem{
			{Content: "1", BGColor: "#000"},
			{Content: "2", BGColor: "#000"},
			{Content: "3", BGColor: "#000"},
		})

		result := sb.Render(100)
		expected := "\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m1\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m2\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m3\x1b[m\x1b[48;2;0;0;0m \x1b[m"

		require.Equal(t, expected, result)
	})

	t.Run("one item full width", func(t *testing.T) {
		sb.Clear()

		sb.Add([]*render.BarItem{
			{Content: "1", BGColor: "#000"},
			{Content: "2", BGColor: "#000", Width: -1},
			{Content: "3", BGColor: "#000"},
		})

		result := sb.Render(50)
		expected := "\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m1\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m2\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[48;2;0;0;0m                                      \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m3\x1b[m\x1b[48;2;0;0;0m \x1b[m"

		require.Equal(t, expected, result)
	})

	t.Run("all items full width", func(t *testing.T) {
		sb.Clear()

		sb.Add([]*render.BarItem{
			{Content: "1", BGColor: "#000", Width: -1},
			{Content: "2", BGColor: "#000", Width: -1},
			{Content: "3", BGColor: "#000", Width: -1},
		})

		result := sb.Render(40)
		expected := "\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m1\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[48;2;0;0;0m         \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m2\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[48;2;0;0;0m         \x1b[m\x1b[38;2;0;0;0;48;2;0;0;0m\ue0b0\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[38;2;255;253;245;48;2;0;0;0m3\x1b[m\x1b[48;2;0;0;0m \x1b[m\x1b[48;2;0;0;0m        \x1b[m"

		require.Equal(t, expected, result)
	})
}
