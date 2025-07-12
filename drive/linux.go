//go:build linux

package drive

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const NTFSSbMagic = 0x5346544e

const mountInfoPath = "/proc/self/mounts"

var excludedFSTypes = map[int64]struct{}{
	unix.CGROUP_SUPER_MAGIC:    {},
	unix.CGROUP2_SUPER_MAGIC:   {},
	unix.SYSFS_MAGIC:           {},
	unix.OVERLAYFS_SUPER_MAGIC: {},
	unix.TMPFS_MAGIC:           {},
	unix.DEBUGFS_MAGIC:         {},
	unix.SQUASHFS_MAGIC:        {},
	unix.PROC_SUPER_MAGIC:      {},
	unix.SECURITYFS_MAGIC:      {},
}

var excludedPaths = map[string]map[string]struct{}{
	"/": {
		"mnt":        {},
		"sys":        {},
		"lost+found": {},
		"boot":       {},
		"proc":       {},
	},
}

var fsTypesMap = map[int64]string{
	unix.EXT4_SUPER_MAGIC:  "ext4",
	unix.XFS_SUPER_MAGIC:   "xfs",
	unix.BTRFS_SUPER_MAGIC: "btrfs",
	unix.NFS_SUPER_MAGIC:   "nfs",
	unix.MSDOS_SUPER_MAGIC: "msdos",
	unix.V9FS_MAGIC:        "v9",
	NTFSSbMagic:            "ntfs",
}

func NewList() (*List, error) {
	mntList, err := mounts()
	if err != nil {
		return nil, err
	}

	list := &List{pathInfoMap: make(map[string]*Info, len(mntList))}

	for i := range mntList {
		info, excluded, mntErr := mntInfo(mntList[i])
		// suppress an error mostly related to the permissions, and requires
		// root access.
		if excluded || mntErr != nil {
			continue
		}

		list.pathInfoMap[mntList[i]] = info
		list.TotalCapacity += info.TotalBytes
		list.TotalFree += info.FreeBytes
		list.TotalUsed += info.UsedBytes
	}

	return list, nil
}

func NewFileInfo(name string, data *unix.Stat_t) FileInfo {
	return FileInfo{
		name:    name,
		isDir:   data.Mode&unix.S_IFMT == unix.S_IFDIR,
		size:    data.Size,
		modTime: time.Unix(int64(data.Mtim.Sec), int64(data.Mtim.Nsec)).Unix(),
	}
}

func mntInfo(path string) (*Info, bool, error) {
	var stat unix.Statfs_t

	if err := unix.Statfs(path, &stat); err != nil {
		return nil, false, fmt.Errorf("failed to statfs: %w", err)
	}

	// use an implicitly defined list of excluded FS types rather than names map
	if _, ok := excludedFSTypes[int64(stat.Type)]; ok || stat.Blocks == 0 {
		return nil, true, nil
	}

	//nolint:gosec // try guessing
	blockSize := uint64(stat.Bsize)

	usedBlocks := stat.Blocks - stat.Bfree

	info := &Info{
		Path:        path,
		FSName:      fsTypesMap[int64(stat.Type)],
		TotalBytes:  stat.Blocks * blockSize,
		FreeBytes:   stat.Bfree * blockSize,
		UsedBytes:   usedBlocks * blockSize,
		UsedPercent: (float64(usedBlocks) / float64(stat.Blocks)) * 100,
	}

	return info, false, nil
}

func mounts() ([]string, error) {
	mountsFile, err := os.Open(mountInfoPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", mountInfoPath, err)
	}

	defer func() {
		_ = mountsFile.Close()
	}()

	scanner, devMntMap := bufio.NewScanner(mountsFile), make(map[string][]string)

	for scanner.Scan() {
		segments := strings.Split(scanner.Text(), " ")
		if len(segments) < 2 {
			// if we dont have at least two fields, we can't really
			// identify anything.
			continue
		}

		if !strings.HasPrefix(segments[0], "/") {
			// special filesystem
			continue
		}

		var stat unix.Stat_t

		if err = unix.Stat(segments[1], &stat); err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", segments[1], err)
		}

		devKey := segments[0] + strconv.FormatUint(uint64(stat.Ino), 10)

		devMntMap[devKey] = append(devMntMap[devKey], segments[1])
	}

	// only return the shortest mount
	mntList := make([]string, 0, len(devMntMap))

	for _, devMounts := range devMntMap {
		shortest := devMounts[0]

		for _, mnt := range devMounts {
			if len(mnt) < len(shortest) {
				shortest = mnt
			}
		}

		mntList = append(mntList, shortest)
	}

	return mntList, nil
}

var direntBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 1024*64)

		return &b
	},
}

func ReadDir(path string) ([]FileInfo, error) {
	var rootStat unix.Stat_t

	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	defer func(fd int) {
		_ = unix.Close(fd)
	}(fd)

	buf, ok := direntBufPool.Get().(*[]byte)
	if !ok {
		return nil, errors.New("get dirent buffer")
	}

	defer direntBufPool.Put(buf)

	fis := make([]FileInfo, 0, 32)

	if err = unix.Fstat(fd, &rootStat); err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	var n int

	for {
		n, err = unix.Getdents(fd, *buf)
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
			name := bytePtrToString(nameBytes[:])

			if pathExcluded(path, name) {
				offset += int(dirent.Reclen)

				continue
			}

			var stat unix.Stat_t

			err = unix.Fstatat(fd, name, &stat, unix.AT_SYMLINK_NOFOLLOW)
			if err == nil && InoFilterInstance.Add(stat.Ino) && stat.Dev == rootStat.Dev {
				fis = append(fis, NewFileInfo(name, &stat))
			}

			offset += int(dirent.Reclen)
		}
	}

	return fis, nil
}

func Explore(path string) error {
	if len(path) == 0 {
		return nil
	}

	cmd := exec.Command("xdg-open", path)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting open: %w", err)
	}

	go func() {
		_ = cmd.Wait()
	}()

	return nil
}

func pathExcluded(path, name string) bool {
	if name == "." || name == ".." {
		return true
	}

	if excludedChild, excluded := excludedPaths[path]; excluded {
		_, childExcluded := excludedChild[name]

		return childExcluded
	}

	return false
}

func bytePtrToString(b []byte) string {
	for n := range b {
		if b[n] == 0 {
			return string(b[:n])
		}
	}

	return ""
}
