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

		expected := []byte{
			32, 49, 32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 50,
			32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 51, 32,
		}

		require.Equal(t, expected, []byte(result))
	})

	t.Run("one item full width", func(t *testing.T) {
		sb.Clear()

		sb.Add([]*render.BarItem{
			{Content: "1", BGColor: "#000"},
			{Content: "2", BGColor: "#000", Width: -1},
			{Content: "3", BGColor: "#000"},
		})

		result := sb.Render(100)

		expected := []byte{
			32, 49, 32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 50,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109,
			32, 51, 32,
		}

		require.Equal(t, expected, []byte(result))
	})

	t.Run("all items full width", func(t *testing.T) {
		sb.Clear()

		sb.Add([]*render.BarItem{
			{Content: "1", BGColor: "#000", Width: -1},
			{Content: "2", BGColor: "#000", Width: -1},
			{Content: "3", BGColor: "#000", Width: -1},
		})

		result := sb.Render(100)

		expected := []byte{
			32, 49, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 27,
			91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 50, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 27, 91, 59, 109, 238, 130,
			176, 27, 91, 48, 109, 32, 51, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
			32, 32, 32,
		}

		require.Equal(t, expected, []byte(result))
	})
}
