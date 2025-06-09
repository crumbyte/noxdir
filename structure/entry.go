package structure

import (
	"bytes"
	"cmp"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const maxTopFiles = 16

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

func (e *Entry) HasChild() bool {
	return len(e.Child) != 0
}

func (e *Entry) SortChild() *Entry {
	slices.SortFunc(e.Child, func(a, b *Entry) int {
		return cmp.Compare(b.Size, a.Size)
	})

	return e
}
