package structure

import (
	"container/heap"
	"sync"
)

// EntrySizeHeap contains a list of the biggest dirs or files found on a specific
// drive/volume. It implements the heap.Interface, hence it persists entries
// sorted by their sizes. The EntrySizeHeap could contain up to n files, where n
// is defined when creating a new heap instance.
type EntrySizeHeap struct {
	files []*Entry
	mx    sync.RWMutex
	size  int
}

// PushSafe provides a thread-safe method for adding elements to the heap. On each
// call, it will check the current number of items and pop the oldest.
func (esh *EntrySizeHeap) PushSafe(e *Entry) {
	esh.mx.Lock()
	defer esh.mx.Unlock()

	heap.Push(esh, e)

	if esh.Len() > esh.size {
		heap.Pop(esh)
	}
}

func (esh *EntrySizeHeap) Reset() {
	esh.files = make([]*Entry, 0)
}

func (esh *EntrySizeHeap) Less(i, j int) bool {
	return esh.files[i].Size < esh.files[j].Size
}

func (esh *EntrySizeHeap) Swap(i, j int) {
	esh.files[i], esh.files[j] = esh.files[j], esh.files[i]
}

func (esh *EntrySizeHeap) Len() int {
	return len(esh.files)
}

func (esh *EntrySizeHeap) Pop() (v any) {
	v, esh.files = esh.files[esh.Len()-1], esh.files[:esh.Len()-1]

	return
}

func (esh *EntrySizeHeap) Push(v any) {
	entry, ok := v.(*Entry)
	if !ok {
		return
	}

	esh.files = append(esh.files, entry)
}
