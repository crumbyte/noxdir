package arena

import (
	"errors"
	"unsafe"
)

var ErrAllocOverflow = errors.New("allocation overflow")

// Bytes defines a simple implementation of a dynamic arena allocator. This
// implementation is not safe for concurrent usage, therefore each concurrent
// unit of work must allocate its own arena instance.
type Bytes struct {
	layout   []byte
	capacity uint32
	offset   uint32
	dynamic  bool
}

func NewBytes(capacity uint32, dynamic bool) *Bytes {
	return &Bytes{
		layout:   make([]byte, capacity),
		capacity: capacity,
		dynamic:  dynamic,
	}
}

func (ba *Bytes) Alloc(size uint32) ([]byte, error) {
	if ba.layout == nil {
		ba.layout = make([]byte, size)
	}

	if ba.offset+size > ba.capacity {
		if !ba.dynamic {
			return nil, ErrAllocOverflow
		}

		ba.capacity *= 2
		newLayout := make([]byte, ba.capacity)

		copy(newLayout, ba.layout)

		ba.layout = newLayout
	}

	start := ba.offset
	ba.offset += size

	return unsafe.Slice(&ba.layout[start], size), nil
}

func (ba *Bytes) Reset() {
	ba.offset = 0
}
