package command

import "sync"

type History struct {
	mx       sync.RWMutex
	entries  []string
	head     int
	cursor   int
	capacity int
	size     int
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

	h.cursor, h.entries[h.head] = h.head, entry
	h.head = (h.head + 1) % h.capacity

	if h.size < h.capacity {
		h.size++
	}
}

func (h *History) Pop() (string, bool) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if h.size == 0 {
		return "", false
	}

	item := h.entries[h.cursor]

	h.cursor--

	if h.cursor < 0 {
		h.cursor = h.size - 1
	}

	return item, true
}

func (h *History) Peek() (string, bool) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if h.size == 0 {
		return "", false
	}

	return h.entries[h.cursor], true
}

func (h *History) Size() int {
	h.mx.RLock()
	defer h.mx.RUnlock()

	return h.size
}
