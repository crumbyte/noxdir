package command

import (
	"sync"
)

type History struct {
	mx         sync.RWMutex
	entries    []string
	head       int
	prevCursor int
	nextCursor int
	capacity   int
	size       int
}

func NewHistory(capacity int) *History {
	return &History{
		entries:  make([]string, capacity),
		capacity: capacity,
	}
}

func (h *History) Push(entry string) {
	h.mx.Lock()
	defer h.mx.Unlock()

	h.prevCursor, h.nextCursor, h.entries[h.head] = h.head, h.head, entry
	h.head = (h.head + 1) % h.capacity

	if h.size < h.capacity {
		h.size++
	}

	h.nextCursor = h.head % h.size
}

func (h *History) Prev() (string, bool) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if h.size == 0 {
		return "", false
	}

	h.nextCursor = (h.prevCursor + 1) % h.size
	item := h.entries[h.prevCursor]

	if h.prevCursor--; h.prevCursor < 0 {
		h.prevCursor = h.size - 1
	}

	return item, true
}

func (h *History) Next() (string, bool) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if h.size == 0 {
		return "", false
	}

	h.prevCursor = h.nextCursor - 1
	if h.prevCursor < 0 {
		h.prevCursor = h.size - 1
	}

	item := h.entries[h.nextCursor]

	h.nextCursor = (h.nextCursor + 1) % h.size

	return item, true
}

func (h *History) Size() int {
	h.mx.RLock()
	defer h.mx.RUnlock()

	return h.size
}

func (h *History) SetCursor(cursor int) {
	h.prevCursor, h.nextCursor = cursor%h.size, cursor%h.size
}

func (h *History) ResetCursor() {
	newCursor := h.head - 1

	if newCursor < 0 {
		newCursor = h.size - 1
	}

	h.prevCursor, h.nextCursor = newCursor, newCursor
}

func (h *History) Clear() {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.size = 0
	h.head = 0
	h.prevCursor = 0
	h.nextCursor = 0
}
