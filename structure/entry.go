package structure

import (
	"bytes"
	"cmp"
	"container/heap"
	"errors"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crumbyte/noxdir/drive"
)

const (
	workerTimeout    = time.Second * 2
	workerReset      = time.Second
	maxTopFiles      = 16
	childPathBufSize = 512
	bfsQueueSize     = 64
)

var (
	TopFilesInstance = EntrySizeHeap{size: maxTopFiles}
	TopDirsInstance  = TopDirs{EntrySizeHeap: EntrySizeHeap{size: maxTopFiles}}
)

// Entry contains the information about a single directory or a file instance
// within the file system. If the entry represents a directory instance, it has
// access to its child elements.
type Entry struct {
	// Path contains the full path to the file or directory represented as an
	// instance of the entry.
	Path string

	// Child contains a list of all child instances including both files and
	// directories. If the current Entry instance represents a file, this
	// property will always be nil.
	Child []*Entry

	mx sync.RWMutex

	// ModTime contains the last modification time of the entry.
	ModTime int64

	// Size contains a total tail in bytes including sizes of all child entries.
	Size int64

	// LocalDirs contain the number of directories within the current entry. This
	// property will always be zero if the current instance represents a file.
	LocalDirs uint64

	// LocalFiles contain the number of files within the current entry. This property
	// will always be zero if the current instance represents a file.
	LocalFiles uint64

	// TotalDirs contains the total number of directories within the current
	// entry, including directories within the child entries. This property will
	// always be zero if the current instance represents a file.
	TotalDirs uint64

	// TotalFiles contains the total number of files within the current entry,
	// including files within the child entries. This property will always be
	// zero if the current instance represents a file.
	TotalFiles uint64

	// IsDir defines whether the current instance represents a dir or a file.
	IsDir bool

	calculateSizeSem uint32
}

func NewDirEntry(path string, modTime int64) *Entry {
	return &Entry{
		Path:    path,
		Child:   make([]*Entry, 0),
		IsDir:   true,
		ModTime: modTime,
	}
}

func NewFileEntry(path string, size int64, modTime int64) *Entry {
	return &Entry{
		Path:    path,
		Size:    size,
		ModTime: modTime,
	}
}

func (e *Entry) Name() string {
	li := bytes.LastIndex([]byte(e.Path), []byte{os.PathSeparator})
	if li == -1 {
		return e.Path
	}

	return e.Path[li+1:]
}

func (e *Entry) Ext() string {
	li := bytes.LastIndex([]byte(e.Path), []byte{'.'})
	if li == -1 {
		return e.Path
	}

	return strings.ToLower(e.Path[li+1:])
}

// Entries returns an iterator for the current node's child elements. Depending
// on the provided argument, the iterator yields either directories or files.
func (e *Entry) Entries(dirs bool) iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		for i := range e.Child {
			if e.Child[i].IsDir == dirs && !yield(e.Child[i]) {
				break
			}
		}
	}
}

// GetChild tries to find a child element by its name. The search will be done
// only on the first level of the child entries. If such an entry was not found,
// a nil value will be returned.
func (e *Entry) GetChild(name string) *Entry {
	e.mx.RLock()
	defer e.mx.RUnlock()

	path := filepath.Join(e.Path, name)

	for _, child := range e.Child {
		if child.Path == path {
			return child
		}
	}

	return nil
}

// AddChild adds the provided [*Entry] instance to a list of child entries. The
// counters will be updated respectively depending on the type of child entry.
func (e *Entry) AddChild(child *Entry) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.Child == nil {
		e.Child = make([]*Entry, 0, 10)
	}

	e.Child = append(e.Child, child)

	if child.IsDir {
		e.TotalDirs, e.LocalDirs = e.TotalDirs+1, e.LocalDirs+1

		return
	}

	e.TotalFiles, e.LocalFiles = e.TotalFiles+1, e.LocalFiles+1
}

// CalculateSize calculates the total number of directories and files, including
// ones within child entries, and the total tail of the current entry instance.
// This function call will recursively calculate the sizes of child entries. The
// final [Entry.Size] field will be a sum of all nested files sizes. If the
// current entry represents a file, only its own tail will be returned.
func (e *Entry) CalculateSize() int64 {
	if atomic.SwapUint32(&e.calculateSizeSem, 1) == 1 || !e.IsDir {
		return e.Size
	}

	defer atomic.SwapUint32(&e.calculateSizeSem, 0)

	e.TotalDirs, e.Size, e.TotalFiles = 0, 0, 0

	for _, child := range e.Child {
		e.Size += child.CalculateSize()

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

func (e *Entry) HasChild() bool {
	return len(e.Child) != 0
}

func (e *Entry) SortChild() *Entry {
	slices.SortFunc(e.Child, func(a, b *Entry) int {
		return cmp.Compare(b.Size, a.Size)
	})

	return e
}

// Traverse traverses the current directory entry instance for all internal files
// and directories and builds the corresponding tree using a BFS approach. The
// total traverse duration depends on the directory's structure depth.
//
// The traverse process only builds the tree structure of child entries and does
// not calculate the final values for total tail and number of child directories
// and files. To do this, the [Entry.CalculateSize] must be called during or
// after the traverse finishes the execution. In the first case, the numbers
// will not be accurate but can be used to display the progress of the traversing
// process gradually.
func (e *Entry) Traverse() error {
	var (
		errList     []error
		currentNode *Entry
	)

	if !e.IsDir {
		return nil
	}

	queue := []*Entry{e}

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		handleEntry(
			currentNode,
			func(newDir *Entry) { queue = append(queue, newDir) },
			func(err error) { errList = append(errList, err) },
		)
	}

	return errors.Join(errList...)
}

func (e *Entry) TraverseAsync() (chan struct{}, chan error) {
	var wg sync.WaitGroup

	drive.InoFilterInstance.Reset()
	TopFilesInstance.Reset()
	heap.Init(&TopFilesInstance)

	if !e.IsDir {
		return nil, nil
	}

	queue := make(chan *Entry, bfsQueueSize)
	done, errChan := make(chan struct{}), make(chan error, 1)

	queue <- e

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

				handleEntry(
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

func handleEntry(e *Entry, onNewDir func(*Entry), onErr func(error)) {
	if !e.IsDir {
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
