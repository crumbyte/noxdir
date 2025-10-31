//go:build darwin

package drive

/*
#include "readdir.h"
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

var excludedFlags uint32 = unix.MNT_RDONLY | unix.MNT_SNAPSHOT | unix.MNT_ROOTFS | unix.MNT_AUTOMOUNTED

var excludedMounts = []string{
	"/dev",
	"/System/Volumes/VM",
	"/System/Volumes/Preboot",
	"/System/Volumes/Update",
	"/System/Volumes/xarts",
	"/System/Volumes/iSCPreboot",
	"/System/Volumes/Hardware",
	"/System/Volumes/Data/home",
}

var excludedPaths = map[string]map[string]struct{}{
	"/System/Volumes/Data": {
		"Volumes": {},
	},
}

func NewList() (*List, error) {
	mounts, err := mntList()
	if err != nil {
		return nil, err
	}

	list := &List{pathInfoMap: make(map[string]*Info, len(mounts))}

mounts:
	for _, mount := range mounts {
		if mount.Flags&excludedFlags != 0 {
			continue
		}

		for _, mntName := range excludedMounts {
			if bytes.HasPrefix(mount.Mntonname[:], []byte(mntName)) {
				continue mounts
			}
		}

		info := statFSToInfo(&mount)

		list.pathInfoMap[info.Path] = info
		list.TotalCapacity += info.TotalBytes
		list.TotalFree += info.FreeBytes
		list.TotalUsed += info.UsedBytes
	}

	return list, nil
}

func mntList() ([]unix.Statfs_t, error) {
	count, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return nil, fmt.Errorf("error getting fsstat: %w", err)
	}

	fs := make([]unix.Statfs_t, count)

	if _, err = unix.Getfsstat(fs, unix.MNT_NOWAIT); err != nil {
		return nil, fmt.Errorf("error getting fsstat: %w", err)
	}

	return fs, nil
}

func statFSToInfo(stat *unix.Statfs_t) *Info {
	usedBlocks := stat.Blocks - stat.Bfree

	blockSize := uint64(stat.Bsize)

	return &Info{
		Path:        byteToString(stat.Mntonname[:]),
		FSName:      byteToString(stat.Fstypename[:]),
		TotalBytes:  stat.Blocks * blockSize,
		FreeBytes:   stat.Bfree * blockSize,
		UsedBytes:   usedBlocks * blockSize,
		UsedPercent: (float64(usedBlocks) / float64(stat.Blocks)) * 100,
	}
}

func NewFileInfo(name string, data *unix.Stat_t) FileInfo {
	return FileInfo{
		name:    name,
		isDir:   data.Mode&unix.S_IFMT == unix.S_IFDIR,
		size:    data.Size,
		modTime: time.Unix(int64(data.Mtim.Sec), int64(data.Mtim.Nsec)).Unix(),
	}
}

var direntBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 1024*64)

		return &b
	},
}

func ReadDir(_ Allocator, path string) ([]FileInfo, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath)) //nolint:nlreturn

	var (
		arr   *C.FileInfoC
		count C.int
	)

	//nolint:gocritic,nlreturn
	if errno := C.read_dir(cPath, &arr, &count); errno != 0 {
		return nil, fmt.Errorf("readdir failed: %d", int(errno))
	}

	defer C.free_result(arr)

	fis := make([]FileInfo, 0, int(count))
	slice := unsafe.Slice(arr, int(count))

	for _, fi := range slice {
		name := C.GoString(&fi.name[0])

		if pathExcluded(path, name) || !InoFilterInstance.Add(uint64(fi.ino)) {
			continue
		}

		fis = append(
			fis,
			FileInfo{
				name:    C.GoString(&fi.name[0]),
				isDir:   fi.isDir != 0,
				size:    int64(fi.size),
				modTime: time.Unix(int64(fi.modSec), int64(fi.modNSec)).Unix(),
			},
		)
	}

	return fis, nil
}

func ReadDirFallback(a Allocator, path string) ([]FileInfo, error) {
	var rootStat unix.Stat_t

	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	defer func(fd int) {
		_ = unix.Close(fd)
	}(fd)

	if err = unix.Fstat(fd, &rootStat); err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	buf, ok := direntBufPool.Get().(*[]byte)
	if !ok {
		return nil, errors.New("get dirent buffer")
	}

	defer direntBufPool.Put(buf)

	fis := make([]FileInfo, 0, 32)

	var n int

	for {
		n, err = unix.ReadDirent(fd, *buf)
		if err != nil {
			return nil, fmt.Errorf("getdents error: %w", err)
		}

		if n == 0 {
			break
		}

		offset := 0

		for offset < n {
			dirent := (*unix.Dirent)(unsafe.Pointer(&(*buf)[offset]))

			nameBytes := (*[256]byte)(unsafe.Pointer(&dirent.Name[0]))
			name := bytePtrToString(a, nameBytes[:])

			if pathExcluded(path, name) {
				offset += int(dirent.Reclen)

				continue
			}

			var stat unix.Stat_t

			err = unix.Fstatat(fd, name, &stat, unix.AT_SYMLINK_NOFOLLOW)
			// TODO: consider making device check optional
			if err == nil && InoFilterInstance.Add(stat.Ino) && rootStat.Dev == stat.Dev {
				fis = append(fis, NewFileInfo(name, &stat))
			}

			offset += int(dirent.Reclen)
		}
	}

	return fis, nil
}

func pathExcluded(path, name string) bool {
	fsMetaData := strings.HasPrefix(name, "\u2400") || strings.HasPrefix(name, ".HFS+")

	if name == "." || name == ".." || fsMetaData {
		return true
	}

	if excludedChild, excluded := excludedPaths[path]; excluded {
		_, childExcluded := excludedChild[name]

		return childExcluded
	}

	return false
}

func Explore(path string) error {
	if len(path) == 0 {
		return nil
	}

	cmd := exec.CommandContext(context.Background(), "open", path)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting open: %w", err)
	}

	go func() {
		_ = cmd.Wait()
	}()

	return nil
}

func bytePtrToString(alloc Allocator, bytes []byte) string {
	var nameLen uint32

	for i := range bytes {
		if bytes[i] == 0 {
			//nolint:gosec
			nameLen = uint32(i)

			break
		}
	}

	nameBuf, _ := alloc.Alloc(nameLen)
	copy(nameBuf, bytes[:nameLen])

	return unsafe.String(unsafe.SliceData(nameBuf), nameLen)
}

func byteToString(b []byte) string {
	for n := range b {
		if b[n] == 0 {
			return string(b[:n])
		}
	}

	return ""
}
