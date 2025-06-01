package drive

import (
	"math"
)

// FileInfoFilter defines a custom function type for filtering *FileInfo
// instances against specific rules. The type implementation only tells if the
// current *FileInfo meets the specification by returning a boolean value.
type FileInfoFilter func(FileInfo) bool

// SizeFilter provided a FileInfoFilter implementation responsible for filtering
// FileInfo file instances by their size. The file's size must be within the
// defined boundaries, otherwise, it must be rejected.
//
// The filter will check files only, and the size boundaries must be defined in
// bytes.
type SizeFilter struct {
	minLimit int64
	maxLimit int64
}

func NewSizeFilter(minLimit, maxLimit int64) *SizeFilter {
	if maxLimit <= 0 {
		maxLimit = math.MaxInt64
	}

	return &SizeFilter{
		minLimit: max(0, minLimit),
		maxLimit: maxLimit,
	}
}

// Filter filters the provided *FileInfo file instance by its size and returns
// a corresponding boolean value. For the directory type, the result will always
// be true.
func (sf *SizeFilter) Filter(fi FileInfo) bool {
	if fi.isDir {
		return true
	}

	return fi.size >= sf.minLimit && fi.size <= sf.maxLimit
}
