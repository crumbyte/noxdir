package structure

type TopDirs struct {
	EntrySizeHeap
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

		for child := range currentNode.Entries(false) {
			totalSize -= child.Size
		}

		if totalSize < currentNode.Size/2 {
			td.PushSafe(currentNode)

			continue
		}

		for child := range currentNode.Entries(true) {
			queue = append(queue, child)
		}
	}
}
