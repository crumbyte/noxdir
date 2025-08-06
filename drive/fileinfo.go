package drive

import (
	"os"
)

type Allocator interface {
	Alloc(size uint32) ([]byte, error)
}

// FileInfo defines a custom fs.FileInfo implementation for wrapping the results
// from the file info system calls.
type FileInfo struct {
	name    string
	modTime int64
	size    int64
	isDir   bool
}

func (fi FileInfo) Name() string {
	return fi.name
}

func (fi FileInfo) Size() int64 {
	return fi.size
}

func (fi FileInfo) Mode() os.FileMode {
	// since we are not using the os.FileMode values we can skip the mapping
	// from the Windows API file attributes.
	panic("os.FileMode not supported")
}

func (fi FileInfo) ModTime() int64 {
	return fi.modTime
}

func (fi FileInfo) IsDir() bool {
	return fi.isDir
}

func (fi FileInfo) Sys() any {
	return nil
}
