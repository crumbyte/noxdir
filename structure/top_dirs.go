package structure

import (
	"container/heap"
	"sync"
)

type TopDirs struct {
	dirs []*Entry
	mx   sync.RWMutex
	size int
}

func (td *TopDirs) PushSafe(e *Entry) {
	td.mx.Lock()
	defer td.mx.Unlock()

	heap.Push(td, e)

	if td.Len() > td.size {
		heap.Pop(td)
	}
}

func (td *TopDirs) Reset() {
	td.dirs = make([]*Entry, 0)
}

func (td *TopDirs) Less(i, j int) bool {
	return td.dirs[i].Size < td.dirs[j].Size
}

func (td *TopDirs) Swap(i, j int) {
	td.dirs[i], td.dirs[j] = td.dirs[j], td.dirs[i]
}

func (td *TopDirs) Len() int {
	return len(td.dirs)
}

func (td *TopDirs) Pop() (v any) {
	v, td.dirs = td.dirs[td.Len()-1], td.dirs[:td.Len()-1]

	return
}

func (td *TopDirs) Push(v any) {
	entry, ok := v.(*Entry)
	if !ok {
		return
	}

	td.dirs = append(td.dirs, entry)
}

func (td *TopDirs) Scan(root *Entry) {
	if !root.IsDir {
		return
	}

	var currentNode *Entry

	queue := []*Entry{root}

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		totalSize := currentNode.Size

		for _, child := range currentNode.Child {
			if !child.IsDir {
				totalSize -= child.Size
			}
		}

		if totalSize < currentNode.Size/2 {
			td.PushSafe(currentNode)

			continue
		}

		for _, child := range currentNode.Child {
			if child.IsDir {
				queue = append(queue, child)
			}
		}
	}
}
