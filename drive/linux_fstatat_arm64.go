//go:build linux && arm64

package drive

import "golang.org/x/sys/unix"

func fstatat(fd int, name string, stat *unix.Stat_t, flags int) error {
	return unix.Fstatat(fd, name, stat, flags)
}
