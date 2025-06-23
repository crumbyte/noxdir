package structure

import "container/heap"

const DefaultMaxTopEntries = 16

var TopEntriesInstance = NewTopEntries(DefaultMaxTopEntries)

type TopEntries struct {
	files EntrySizeHeap
	dirs  EntrySizeHeap
}

func NewTopEntries(maxEntries int) *TopEntries {
	return &TopEntries{
		files: EntrySizeHeap{size: maxEntries},
		dirs:  EntrySizeHeap{size: maxEntries},
	}
}

func (te *TopEntries) Files() heap.Interface {
	return &te.files
}

func (te *TopEntries) Dirs() heap.Interface {
	return &te.dirs
}

func (te *TopEntries) ScanFiles(root *Entry) {
	if root == nil || !root.IsDir {
		return
	}

	te.files.Reset()

	var currentNode *Entry

	queue := []*Entry{root}

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		for child := range currentNode.Entries() {
			if child.IsDir {
				queue = append(queue, child)

				continue
			}

			te.files.PushSafe(child)
		}
	}
}

func (te *TopEntries) ScanDirs(root *Entry) {
	if root == nil || !root.IsDir {
		return
	}

	te.dirs.Reset()

	var currentNode *Entry

	queue := []*Entry{root}

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		totalSize := currentNode.Size

		for child := range currentNode.EntriesByType(false) {
			totalSize -= child.Size
		}

		if totalSize < currentNode.Size/2 {
			te.dirs.PushSafe(currentNode)

			continue
		}

		for child := range currentNode.EntriesByType(true) {
			queue = append(queue, child)
		}
	}
}
