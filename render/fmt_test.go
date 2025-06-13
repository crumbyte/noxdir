package render_test

import (
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
