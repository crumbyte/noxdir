package render_test

import (
	"path/filepath"
	"testing"

	"github.com/crumbyte/noxdir/render"

	"github.com/stretchr/testify/require"
)

func TestFmtSize(t *testing.T) {
	tableData := []struct {
		expected string
		bytes    uint64
		width    int
	}{
		{"0.00          B", 0, 15},
		{"1.00          B", 1, 15},
		{"1023.00    B", 1023, 12},
		{"1.00      KB", 1024, 12},
		{"1.00 MB", 1024 << 10, 0},
		{"1.00 GB", 1024 << 20, 0},
		{"1.00 TB", 1024 << 30, 0},
		{"1.00 PB", 1024 << 40, 0},
		{"1.00 EB", 1024 << 50, 0},
		{"512.00 KB", 1024 << 10 / 2, 0},
		{"512.00 MB", 1024 << 20 / 2, 0},
		{"512.00 GB", 1024 << 30 / 2, 0},
		{"512.00 TB", 1024 << 40 / 2, 0},
		{"512.00 PB", 1024 << 50 / 2, 0},
	}

	for _, data := range tableData {
		require.Equal(t, data.expected, render.FmtSize(data.bytes, data.width))
	}
}

func TestFmtUsage(t *testing.T) {
	tableData := []struct {
		usage    float64
		expected string
	}{
		{1, "100.00 %"},
		{0.2, "20.00  %"},
		{0.155, "15.50  %"},
		{0.01, "1.00   %"},
		{0, "0.00   %"},
		{-0.2, "0.00   %"},
	}

	for _, data := range tableData {
		require.Equal(t, data.expected, render.FmtUsage(data.usage, 80, 8))
	}
}

func TestWrapPath(t *testing.T) {
	tableData := []struct {
		path     string
		limit    int
		expected string
	}{
		{filepath.Join("a", "b", "c", "d"), 0, "..."},
		{filepath.Join("a", "b", "c", "d"), 1, "...d"},
		{filepath.Join("a", "b", "c", "d"), 2, filepath.Join("...", "d")},
		{filepath.Join("a", "b", "c", "d"), 3, filepath.Join("...", "d")},
		{filepath.Join("a", "b", "c", "d"), 4, filepath.Join("...", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), 5, filepath.Join("...", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), 6, filepath.Join("...", "b", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), 7, filepath.Join("a", "b", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), 8, filepath.Join("a", "b", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), 10, filepath.Join("a", "b", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), -1, filepath.Join("a", "b", "c", "d")},
		{filepath.Join("a", "b", "c", "d"), -10, filepath.Join("a", "b", "c", "d")},
		{"longPathName", 4, "...Name"},
		{"longPathName/subPath", 4, "...Path"},
	}

	for _, data := range tableData {
		require.Equal(t, data.expected, render.WrapPath(data.path, data.limit))
	}
}
