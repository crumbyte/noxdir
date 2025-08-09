package structure

import (
	"bytes"
	"cmp"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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

// EntriesByType returns an iterator for the current node's child elements.
// Depending on the provided argument, the iterator yields either directories
// or files.
func (e *Entry) EntriesByType(dirs bool) iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		for i := range e.Child {
			if e.Child[i].IsDir == dirs && !yield(e.Child[i]) {
				break
			}
		}
	}
}

// Entries returns an iterator for all the current node's child elements.
func (e *Entry) Entries() iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		for i := range e.Child {
			if !yield(e.Child[i]) {
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

	path := filepath.Join(e.Path, strings.TrimPrefix(name, e.Path))

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

func (e *Entry) HasChild() bool {
	return len(e.Child) != 0
}

func (e *Entry) SortChild() *Entry {
	slices.SortFunc(e.Child, func(a, b *Entry) int {
		return cmp.Compare(b.Size, a.Size)
	})

	return e
}

func (e *Entry) Copy() *Entry {
	return &Entry{
		Path:       e.Path,
		Child:      make([]*Entry, 0, len(e.Child)),
		IsDir:      e.IsDir,
		ModTime:    e.ModTime,
		Size:       e.Size,
		LocalDirs:  e.LocalDirs,
		LocalFiles: e.LocalFiles,
		TotalDirs:  e.TotalDirs,
		TotalFiles: e.TotalFiles,
	}
}

func (e *Entry) Diff(ne *Entry) *Diff {
	var ep EntryPair

	d := Diff{Added: make([]*Entry, 0), Removed: make([]*Entry, 0)}

	queue := []EntryPair{{e, ne}}

	for len(queue) > 0 {
		ep, queue = queue[0], queue[1:]

		diff := EntryList(ep[0].Child).Diff(ep[1].Child)

		d.Added = append(d.Added, diff.Added...)
		d.Removed = append(d.Removed, diff.Removed...)

		for _, sameEntries := range diff.Same {
			if sameEntries[0].IsDir {
				queue = append(queue, sameEntries)
			}
		}
	}

	return d.Sort()
}

type EntryPair [2]*Entry

type Diff struct {
	Same    []EntryPair
	Added   []*Entry
	Removed []*Entry
}

func (d *Diff) Empty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0
}

func (d *Diff) Sort() *Diff {
	slices.SortFunc(d.Added, func(a, b *Entry) int {
		return cmp.Compare(b.Size, a.Size)
	})

	slices.SortFunc(d.Removed, func(a, b *Entry) int {
		return cmp.Compare(b.Size, a.Size)
	})

	return d
}

func DiffStats(entries []*Entry) (uint64, uint64, int64) {
	dirs, files := uint64(0), uint64(0)
	total := int64(0)

	for _, entry := range entries {
		total += entry.Size

		if entry.IsDir {
			dirs += entry.TotalDirs + 1
			files += entry.TotalFiles

			continue
		}

		files++
	}

	return dirs, files, total
}

// EntryList contains a set of *Entry instances. It can represent a list of child
// entries.
type EntryList []*Entry

// Diff returns the delta between the current and the provided EntryList instances.
// As a result, it returns the Diff instance containing Diff.Added, Diff.Removed,
// and Diff.Same entries.
//
// It uses a straightforward approach by comparing the EntryList items one by one
// for each set. If there is an item in the provided EntryList that does not
// exist in the current EntryList, then it is considered to be "added".
func (el EntryList) Diff(newList EntryList) *Diff {
	d := &Diff{
		Same:    make([]EntryPair, 0),
		Added:   make([]*Entry, 0),
		Removed: make([]*Entry, 0),
	}

	if len(newList) == 0 {
		return d
	}

	elMap := make(map[string]*Entry, len(el))

	for _, entry := range el {
		elMap[entry.Path] = entry
	}

	for _, newChild := range newList {
		oldChild, ok := elMap[newChild.Path]

		if ok && oldChild.IsDir == newChild.IsDir {
			d.Same = append(d.Same, EntryPair{oldChild, newChild})

			delete(elMap, newChild.Path)

			continue
		}

		d.Added = append(d.Added, newChild)
	}

	for removed := range maps.Values(elMap) {
		d.Removed = append(d.Removed, removed)
	}

	return d
}
