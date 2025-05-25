package structure

import (
	"container/heap"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crumbyte/noxdir/drive"
)

// TreeOpt defines a custom type for configuring a *Tree instance.
type TreeOpt func(*Tree)

// WithExclude allows setting a list of directory names that must be excluded
// from the traversal during the tree build-up process. The directory name can
// represent an absolute path or just a part of the name. In the last case, all
// directories that contain this name will be excluded. For example, the following
// path "dir/sub_dir/inner/other" and adding the name "sub" for exclusion will
// completely remove the "dir/sub_dir" directory from traversal. To avoid that,
// use a more specific path, e.g., "dir/sub/".
func WithExclude(exclude []string) TreeOpt {
	return func(t *Tree) {
		for i := range exclude {
			exclude[i] = strings.ToLower(strings.TrimSpace(exclude[i]))
		}

		t.exclude = exclude
	}
}

// Tree provides a set of method for building and traversing the *Entry tree.
type Tree struct {
	root             *Entry
	exclude          []string
	calculateSizeSem uint32
}

func NewTree(root *Entry, opts ...TreeOpt) *Tree {
	t := &Tree{root: root}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Root returns a root *Entry node for the current tree.
func (t *Tree) Root() *Entry {
	return t.root
}

// SetRoot changes the current root of the tree instance.
func (t *Tree) SetRoot(root *Entry) {
	t.root = root
}

// CalculateSize calculates the total number of directories and files, including
// ones within child entries, and the total tail of the current entry instance.
// This function call will recursively calculate the sizes of child entries. The
// final [Entry.Size] field will be a sum of all nested files sizes. If the
// current entry represents a file, only its own tail will be returned.
func (t *Tree) CalculateSize() {
	if t.root == nil || !t.root.IsDir {
		return
	}

	if atomic.SwapUint32(&t.calculateSizeSem, 1) == 1 {
		return
	}

	defer atomic.SwapUint32(&t.calculateSizeSem, 0)

	var calculate func(e *Entry) int64
	calculate = func(e *Entry) int64 {
		if !e.IsDir {
			return e.Size
		}

		e.TotalDirs, e.Size, e.TotalFiles = 0, 0, 0

		for _, child := range e.Child {
			e.Size += calculate(child)

			if child.IsDir {
				e.TotalDirs++
			} else {
				e.TotalFiles++
			}

			e.TotalDirs += child.TotalDirs
			e.TotalFiles += child.TotalFiles
		}

		return e.Size
	}

	calculate(t.root)
}

// Traverse traverses the current root entry instance for all internal files, and
// directories and builds the corresponding tree using a BFS approach. The total
// traverse duration depends on the directory's structure depth.
//
// The traverse process only builds the tree structure of child entries and does
// not calculate the final values for total tail and number of child directories
// and files. To do this, the Tree.CalculateSize must be called during or
// after the traverse finishes the execution. In the first case, the numbers
// will not be accurate but can be used to display the progress of the traversing
// process gradually.
func (t *Tree) Traverse() error {
	var (
		errList     []error
		currentNode *Entry
	)

	if t.root == nil || !t.root.IsDir {
		return nil
	}

	queue := []*Entry{t.root}

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		t.handleEntry(
			currentNode,
			func(newDir *Entry) { queue = append(queue, newDir) },
			func(err error) { errList = append(errList, err) },
		)
	}

	return errors.Join(errList...)
}

func (t *Tree) TraverseAsync() (chan struct{}, chan error) {
	var wg sync.WaitGroup

	drive.InoFilterInstance.Reset()
	TopFilesInstance.Reset()
	heap.Init(&TopFilesInstance)

	if t.root == nil || !t.root.IsDir {
		return nil, nil
	}

	queue := make(chan *Entry, bfsQueueSize)
	done, errChan := make(chan struct{}), make(chan error, 1)

	queue <- t.root

	worker := func() {
		timeoutTimer := time.NewTimer(workerTimeout)

		defer func() {
			wg.Done()
			timeoutTimer.Stop()
		}()

		for {
			select {
			case entry, ok := <-queue:
				if !ok {
					return
				}

				t.handleEntry(
					entry,
					func(newDir *Entry) { go func() { queue <- newDir }() },
					func(err error) { errChan <- err },
				)

				timeoutTimer.Reset(workerReset)
			case <-timeoutTimer.C:
				return
			}
		}
	}

	for range runtime.GOMAXPROCS(-1) * 2 {
		wg.Add(1)
		go worker()
	}

	go func() {
		wg.Wait()

		close(done)
		close(queue)
		close(errChan)
	}()

	return done, errChan
}

var childPathBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, childPathBufSize)

		return &b
	},
}

func (t *Tree) handleEntry(e *Entry, onNewDir func(*Entry), onErr func(error)) {
	if !e.IsDir || t.excludeEntry(e) {
		return
	}

	nodeEntries, err := drive.ReadDir(e.Path)
	if err != nil {
		onErr(err)

		return
	}

	nameBuf, ok := childPathBufPool.Get().(*[]byte)
	if !ok {
		return
	}

	defer childPathBufPool.Put(nameBuf)

	for _, child := range nodeEntries {
		*nameBuf = append(*nameBuf, e.Path...)

		if e.Path[len(e.Path)-1] != filepath.Separator {
			*nameBuf = append(*nameBuf, filepath.Separator)
		}

		*nameBuf = append(*nameBuf, child.Name()...)

		childPath := string(*nameBuf)
		*nameBuf = (*nameBuf)[:0]

		if child.IsDir() {
			newDir := NewDirEntry(childPath, child.ModTime())

			e.AddChild(newDir)
			onNewDir(newDir)

			continue
		}

		fe := NewFileEntry(childPath, child.Size(), child.ModTime())

		TopFilesInstance.PushSafe(fe)
		e.AddChild(fe)
	}
}

func (t *Tree) excludeEntry(e *Entry) bool {
	for _, exclude := range t.exclude {
		if strings.Contains(strings.ToLower(e.Path), exclude) {
			return true
		}
	}

	return false
}
