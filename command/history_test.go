package command_test

import (
	"testing"

	"github.com/crumbyte/noxdir/command"

	"github.com/stretchr/testify/require"
)

func TestHistory_Prev(t *testing.T) {
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

	for i := 15 - 1; i >= 0; i-- {
		val, exists := h.Prev()

		require.True(t, exists)
		require.Equal(t, entries[(i%5)+5], val)
	}
}

func TestHistory_Next(t *testing.T) {
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

	for i := range 15 {
		val, exists := h.Next()

		require.True(t, exists)
		require.Equal(t, entries[(i%5)+5], val)
	}
}
