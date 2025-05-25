package render_test

import (
	"testing"

	"github.com/crumbyte/noxdir/render"

	"github.com/stretchr/testify/require"
)

func TestNewStatusBar(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		items := []*render.BarItem{
			render.NewBarItem("1", "#000", 0),
			render.NewBarItem("2", "#000", 0),
			render.NewBarItem("3", "#000", 0),
		}

		sb := render.NewStatusBar(items, 100)

		expected := []byte{
			32, 49, 32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 50,
			32, 27, 91, 59, 109, 238, 130, 176, 27, 91, 48, 109, 32, 51, 32,
		}

		require.Equal(t, expected, []byte(sb))
	})

	t.Run("one item full width", func(t *testing.T) {
		items := []*render.BarItem{
			render.NewBarItem("1", "#000", 0),
			render.NewBarItem("2", "#000", -1),
			render.NewBarItem("3", "#000", 0),
		}

		sb := render.NewStatusBar(items, 100)

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

		require.Equal(t, expected, []byte(sb))
	})

	t.Run("all items full width", func(t *testing.T) {
		items := []*render.BarItem{
			render.NewBarItem("1", "#000", -1),
			render.NewBarItem("2", "#000", -1),
			render.NewBarItem("3", "#000", -1),
		}

		sb := render.NewStatusBar(items, 100)

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

		require.Equal(t, expected, []byte(sb))
	})
}
