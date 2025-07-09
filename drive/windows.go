//go:build windows

package drive

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	winapi "golang.org/x/sys/windows"
)

const windowDisplayMode = uintptr(1)

var (
	systemDLL = winapi.NewLazySystemDLL("kernel32.dll")
	shellDLL  = winapi.NewLazySystemDLL("Shell32.dll")
)

var (
	procGetDiskFreeSpaceEx   = systemDLL.NewProc("GetDiskFreeSpaceExW")
	procFindFirstFile        = systemDLL.NewProc("FindFirstFileW")
	procFindNextFile         = systemDLL.NewProc("FindNextFileW")
	procGetLogicalDrives     = systemDLL.NewProc("GetLogicalDrives")
	procGetVolumeInformation = systemDLL.NewProc("GetVolumeInformationA")
	procShellExecute         = shellDLL.NewProc("ShellExecuteW")
)

type driveSpace struct {
	totalBytes  uint64
	freeBytes   uint64
	usedBytes   uint64
	usedPercent float64
}

type win32finddata1 struct {
	FileAttributes    uint32
	CreationTime      syscall.Filetime
	LastAccessTime    syscall.Filetime
	LastWriteTime     syscall.Filetime
	FileSizeHigh      uint32
	FileSizeLow       uint32
	Reserved0         uint32
	Reserved1         uint32
	FileName          [syscall.MAX_PATH]uint16
	AlternateFileName [14]uint16
}

func NewFileInfo(data *win32finddata1) FileInfo {
	return FileInfo{
		name:    syscall.UTF16ToString(data.FileName[:]),
		isDir:   data.FileAttributes&16 != 0,
		size:    int64(data.FileSizeHigh)<<32 + int64(data.FileSizeLow),
		modTime: time.Unix(0, data.LastWriteTime.Nanoseconds()).Unix(),
	}
}

type handleWrapper struct {
	handle syscall.Handle
}

func newHandleWrapper(h uintptr) *handleWrapper {
	hw := &handleWrapper{
		handle: syscall.Handle(h),
	}

	runtime.SetFinalizer(hw, (*handleWrapper).Close)

	return hw
}

// Close closes the resource handle created by the FindFirstFileW. This function
// will be set as a finalizer for the handleWrapper instance instead of closing
// the handle explicitly. The handle closing heavily relies on the GC call since
// a large number of open resources eventually leads to the "too many open files"
// error. ~10000 is a limit per process.
func (hw *handleWrapper) Close() error {
	defer runtime.SetFinalizer(hw, nil)

	if hw.handle != syscall.InvalidHandle {
		return nil
	}

	if err := syscall.CloseHandle(hw.handle); err != nil {
		return fmt.Errorf("drive: CloseHandle: %w", err)
	}

	hw.handle = syscall.InvalidHandle

	return nil
}

func NewList() (*List, error) {
	disks, err := logicalDrives()
	if err != nil {
		return nil, err
	}

	l := &List{
		pathInfoMap: make(map[string]*Info, len(disks)),
	}

	for _, disk := range disks {
		path := string(disk) + ":\\"

		var di Info

		if di, err = NewInfo(path); err != nil {
			return nil, err
		}

		l.pathInfoMap[path] = &di

		l.TotalCapacity += di.TotalBytes
		l.TotalUsed += di.UsedBytes
		l.TotalFree += di.FreeBytes
	}

	return l, nil
}

func NewInfo(path string) (Info, error) {
	ds, err := spaceUsage(path)
	if err != nil {
		return Info{}, err
	}

	vName, fsName, err := volumeInfo([]byte(path))
	if err != nil {
		return Info{}, err
	}

	return Info{
		Path:        path,
		Volume:      vName,
		FSName:      fsName,
		TotalBytes:  ds.totalBytes,
		FreeBytes:   ds.freeBytes,
		UsedBytes:   ds.usedBytes,
		UsedPercent: ds.usedPercent,
	}, nil
}

// logicalDrives retrieves a list of available logical drives on the system. It
// uses the Windows API to get the bitmask representing the available drives,
// over each possible drive letter (A-Z), and checks if it's present in the
// bitmask.
//
// It returns a byte slice containing the ASCII names of the available drives,
// or an error if no drives were found.
func logicalDrives() ([]byte, error) {
	// Get the bitmask representing the available logical drives.
	disksBitmask, _, err := procGetLogicalDrives.Call()

	if disksBitmask == 0 {
		return nil, fmt.Errorf("drive: GetDisksBitmask: %w", err)
	}

	var disks []byte

	for diskIdx := 65; diskIdx <= 90 && disksBitmask > 0; diskIdx++ {
		if disksBitmask&1 == 1 {
			disks = append(disks, byte(diskIdx))
		}

		disksBitmask >>= 1
	}

	return disks, nil
}

// volumeInfo retrieves information about the volume using its specified path.
// It uses the Windows API to get the volume name and the file system type. If an
// error occurs, it returns an error with empty strings as names.
func volumeInfo(path []byte) (string, string, error) {
	var (
		volumeNameBuffer     = make([]uint8, 32)
		fsNameBuffer         = make([]byte, 8)
		volumeNameBufferSize = uint32(32)
		fsNameBufferSize     = uint32(8)
	)

	vi, _, err := procGetVolumeInformation.Call(
		toUintptr(&path[0]),
		toUintptr(&volumeNameBuffer[0]),
		toUintptr(&volumeNameBufferSize),
		0, 0, 0,
		toUintptr(&fsNameBuffer[0]),
		toUintptr(&fsNameBufferSize),
	)

	if vi == 0 {
		return "", "", fmt.Errorf("drive: GetVolumeInformationA: %w", err)
	}

	return string(volumeNameBuffer), string(fsNameBuffer), nil
}

// spaceUsage retrieves the volume's space usage information using Windows API
// and returns it as a driveSpace instance. The retrieved data contains the total
// bytes and free bytes. Using these values, it calculates the used space in bytes
// and percentage usage.
func spaceUsage(path string) (driveSpace, error) {
	var ds driveSpace

	pathPtr, err := winapi.UTF16PtrFromString(path)
	if err != nil {
		return ds, fmt.Errorf("drive: UTF16PtrFromString: %w", err)
	}

	diskFreeSpace, _, err := procGetDiskFreeSpaceEx.Call(
		toUintptr(pathPtr),
		0,
		toUintptr(&ds.totalBytes),
		toUintptr(&ds.freeBytes),
	)

	if diskFreeSpace == 0 {
		return ds, fmt.Errorf("drive: GetDiskFreeSpaceExW: %w", err)
	}

	ds.usedBytes = ds.totalBytes - ds.freeBytes
	ds.usedPercent = float64(ds.usedBytes) / float64(ds.totalBytes) * 100

	return ds, nil
}

// ExploreFallback explores the directory using Windows API. If the provided path is a
// valid directory path, it will open it in the file explorer in a new window.
// Otherwise, an error will be returned.
func ExploreFallback(path string) error {
	pathPtr, err := winapi.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("drive: UTF16PtrFromString: %w", err)
	}

	action, err := winapi.UTF16PtrFromString("open")
	if err != nil {
		return fmt.Errorf("drive: UTF16PtrFromString: %w", err)
	}

	code, _, err := procShellExecute.Call(
		0, toUintptr(action), toUintptr(pathPtr), 0, 0, windowDisplayMode,
	)
	if err != nil || *(*int)(unsafe.Pointer(&code)) < 32 {
		return fmt.Errorf("drive: ShellExecuteW: %w", err)
	}

	return nil
}

// Explore explores the directory using the PowerShell command execution. If the
// provided path is a valid directory path, it will open it in the file explorer
// in a new window. Otherwise, an error will be returned.
func Explore(path string) error {
	//nolint:gosec // because of the one text editor that shouldn't exist
	cmd := exec.Command(
		"powershell",
		"-Command",
		fmt.Sprintf("Start-Process -FilePath %q", strings.TrimSpace(path)),
	)

	// Detach stdio to prevent terminal pollution
	cmd.Stdout, cmd.Stderr, cmd.Stdin = nil, nil, nil

	return cmd.Start()
}

// ReadDir reads the provided directory and returns its entries as a slice of
// instances [FileInfoAccess] instances.
func ReadDir(path string) ([]FileInfo, error) {
	var data win32finddata1

	fis := make([]FileInfo, 0, 32)

	pathPtr, err := syscall.UTF16PtrFromString(path + "\\*")
	if err != nil {
		return nil, fmt.Errorf("drive: UTF16PtrFromString: %w", err)
	}

	handle, _, errno := syscall.SyscallN(
		procFindFirstFile.Addr(),
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&data)),
	)

	hw := newHandleWrapper(handle)

	if hw.handle == syscall.InvalidHandle && errno != 0 {
		return nil, fmt.Errorf("drive: FindFirstFile: %w", errno)
	}

	for {
		name := data.FileName

		//nolint:staticcheck // the alternative is even worse
		if !(name[0] == '.' && (name[1] == 0 || (name[1] == '.' && name[2] == 0))) {
			fis = append(fis, NewFileInfo(&data))
		}

		result, _, _ := syscall.SyscallN(
			procFindNextFile.Addr(),
			handle,
			uintptr(unsafe.Pointer(&data)),
		)
		if result == 0 {
			break
		}
	}

	return fis, nil
}

func toUintptr[T any](val *T) uintptr {
	return uintptr(unsafe.Pointer(val))
}
