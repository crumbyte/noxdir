package command_test

import (
	"testing"

	"github.com/crumbyte/noxdir/command"

	"github.com/stretchr/testify/require"
)

func TestHistory_Push(t *testing.T) {
	capacity := 5
	h := command.NewHistory(capacity)

	entries := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
	}

	for i := range entries {
		require.Equal(t, min(i, capacity), h.Size())
		h.Push(entries[i])
		require.Equal(t, min(i+1, capacity), h.Size())
	}

	for i := len(entries) - 1; i >= 5; i-- {
		val, exists := h.Pop()
		require.True(t, exists)

		require.Equal(t, entries[i], val)
	}

	for i := range entries[:5] {
		h.Push(entries[i])
	}

	for i := 4; i >= 0; i-- {
		val, exists := h.Pop()
		require.True(t, exists)

		require.Equal(t, entries[i], val)
	}
}
