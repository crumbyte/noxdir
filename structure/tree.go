package structure

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/pkg/arena"
	"github.com/crumbyte/noxdir/pkg/cache"
)

const (
	workerTimeout    = time.Second
	childPathBufSize = 512
	bfsQueueSize     = 1024
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

// WithFileInfoFilter allows setting a list of filters for a drive.FileInfo
// instances. The filters will be applied during the tree traversal and discard
// nodes that do not meet the specific filter's specification.
//
// The Tree instance does not dictate the filter behavior; hence, the entire
// filtration logic is defined within each drive.FileInfoFilter filter.
func WithFileInfoFilter(fl []drive.FileInfoFilter) TreeOpt {
	return func(t *Tree) {
		if len(fl) != 0 {
			t.fiFilters = fl
		}
	}
}

func WithCache(c *cache.Cache) TreeOpt {
	return func(t *Tree) {
		t.cache = c
	}
}

func WithUseCache() TreeOpt {
	return func(t *Tree) {
		t.useCache = true
	}
}

func WithPartialRoot() TreeOpt {
	return func(t *Tree) {
		t.partialRoot = true
	}
}

// Tree provides a set of method for building and traversing the *Entry tree.
type Tree struct {
	root             *Entry
	cache            *cache.Cache
	exclude          []string
	fiFilters        []drive.FileInfoFilter
	calculateSizeSem uint32
	partialRoot      bool
	useCache         bool
	dirty            bool
}

func NewTree(root *Entry, opts ...TreeOpt) *Tree {
	t := &Tree{root: root}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Clone clones the existing *Tree instance. It creates a new Tree instance and
// copies all the predefined settings, except for the root Entry. The root still
// must be specified explicitly, and the copied settings can be overwritten with
// the optional set of TreeOpt options.
func (t *Tree) Clone(root *Entry, opts ...TreeOpt) *Tree {
	clonedTree := &Tree{
		root:        root,
		cache:       t.cache,
		exclude:     t.exclude,
		fiFilters:   t.fiFilters,
		partialRoot: t.partialRoot,
	}

	for _, opt := range opts {
		opt(clonedTree)
	}

	return clonedTree
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
func (t *Tree) Traverse(skipCache bool) error {
	var (
		errList     []error
		currentNode *Entry
	)

	if !skipCache && t.cachingEnabled() {
		if err := t.cache.Get(t.root.Path, t.root); err == nil {
			return nil
		}
	}

	t.dirty = true

	drive.InoFilterInstance.Reset()

	if t.root == nil || !t.root.IsDir {
		return nil
	}

	queue := []*Entry{t.root}

	ba := arena.NewBytes(1024*1024, true)

	for len(queue) > 0 {
		currentNode, queue = queue[0], queue[1:]

		t.handleEntry(
			ba,
			currentNode,
			func(newDir *Entry) { queue = append(queue, newDir) },
			func(err error) { errList = append(errList, err) },
		)
	}

	return errors.Join(errList...)
}

func (t *Tree) Cached() (*Tree, error) {
	if t.cachingEnabled() {
		return t, nil
	}

	if !t.cache.Has(t.root.Path) {
		return nil, nil
	}

	tree := NewTree(NewDirEntry(t.root.Path, time.Now().Unix()))
	if err := t.cache.Get(tree.root.Path, tree.root); err != nil {
		return nil, err
	}

	return tree, nil
}

func (t *Tree) PersistCache() (chan struct{}, error) {
	if t.cache == nil || t.partialRoot || t.root == nil || !t.dirty {
		done := make(chan struct{})
		close(done)

		return done, nil
	}

	return t.cache.SetAsync(t.root.Path, t.root)
}

// scanQueue represents a queue for *Entry instances scheduled for traversal.
type scanQueue struct {
	entries []*Entry
	mx      sync.RWMutex
}

func (sq *scanQueue) Push(val *Entry) {
	if val == nil {
		return
	}

	sq.mx.Lock()
	defer sq.mx.Unlock()

	sq.entries = append(sq.entries, val)
}

func (sq *scanQueue) Get() (*Entry, bool) {
	sq.mx.Lock()
	defer sq.mx.Unlock()

	if len(sq.entries) == 0 {
		return nil, false
	}

	entry := sq.entries[0]
	sq.entries = sq.entries[1:]

	return entry, true
}

func (t *Tree) TraverseNodeAsync(node *Entry) (chan struct{}, chan error) {
	t.dirty, node.Child = true, nil

	return t.Clone(node, WithPartialRoot()).TraverseAsync(true)
}

func (t *Tree) TraverseAsync(skipCache bool) (chan struct{}, chan error) {
	drive.InoFilterInstance.Reset()

	if t.root == nil || !t.root.IsDir {
		return nil, nil
	}

	done, errChan := make(chan struct{}), make(chan error, 1)

	if !skipCache && t.cachingEnabled() && t.cache.Has(t.root.Path) {
		go func() {
			if err := t.cache.Get(t.root.Path, t.root); err == nil {
				close(done)
			}
		}()

		return done, errChan
	}

	var wg sync.WaitGroup

	t.dirty = true

	queue := scanQueue{entries: make([]*Entry, 0, bfsQueueSize)}
	queue.Push(t.root)

	worker := func() {
		timeoutTimer := time.NewTimer(workerTimeout)
		ba := arena.NewBytes(1024*1024, true)

		defer func() {
			wg.Done()
			timeoutTimer.Stop()
			ba.Reset()
		}()

		for {
			select {
			case <-timeoutTimer.C:
				return
			default:
				item, ok := queue.Get()
				if !ok {
					continue
				}

				t.handleEntry(
					ba,
					item,
					func(newDir *Entry) { queue.Push(newDir) },
					func(err error) { errChan <- err },
				)

				timeoutTimer.Reset(workerTimeout)
			}
		}
	}

	for range runtime.NumCPU() * 2 {
		wg.Add(1)
		go worker()
	}

	go func() {
		wg.Wait()

		close(done)
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

func (t *Tree) handleEntry(ba *arena.Bytes, e *Entry, onNewDir func(*Entry), onErr func(error)) {
	if !e.IsDir || t.excludeEntry(e) {
		return
	}

	nodeEntries, err := drive.ReadDir(ba, e.Path)
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
		if !t.filterFileInfo(child) {
			continue
		}

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

		e.AddChild(NewFileEntry(childPath, child.Size(), child.ModTime()))
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

func (t *Tree) filterFileInfo(fi drive.FileInfo) bool {
	for i := range t.fiFilters {
		if !t.fiFilters[i](fi) {
			return false
		}
	}

	return true
}

func (t *Tree) cachingEnabled() bool {
	return t.useCache && t.cache != nil
}
