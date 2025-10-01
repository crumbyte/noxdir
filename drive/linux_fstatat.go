//go:build linux && !arm64

package drive

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const SysNewFstatat = 262

func fstatat(alloc Allocator, dirFD int, path string, stat *unix.Stat_t, flags int) (err error) {
	var _p0 *byte

	_p0, err = bytePtrFromString(alloc, path)
	if err != nil {
		return
	}

	_, _, e1 := syscall.Syscall6(
		SysNewFstatat,
		uintptr(dirFD),
		uintptr(unsafe.Pointer(_p0)),
		uintptr(unsafe.Pointer(stat)),
		uintptr(flags),
		0,
		0,
	)
	if e1 != 0 {
		return e1
	}

	return
}

func bytePtrFromString(alloc Allocator, s string) (*byte, error) {
	if strings.IndexByte(s, 0) != -1 {
		return nil, syscall.EINVAL
	}

	//nolint:gosec
	buf, err := alloc.Alloc(uint32(len(s) + 1))
	if err != nil {
		return nil, err
	}

	copy(buf, s)

	return &buf[0], nil
}
