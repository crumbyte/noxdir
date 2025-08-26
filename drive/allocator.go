package drive

// Allocator defines a basic interface for the arena allocator.
type Allocator interface {
	// Alloc allocates the required number of bytes using the specific allocator
	// implementation. If the memory region cannot be allocated an error will be
	// returned.
	Alloc(size uint32) ([]byte, error)
}
