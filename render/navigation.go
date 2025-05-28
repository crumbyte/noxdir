package render

import (
	"errors"
	"sync/atomic"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/structure"
)

// State defines a custom type representing the current GUI state. The application's
// behavior depends on the current state value.
type State int

const (
	// Drives defines an application state while selecting the target drive.
	Drives State = iota

	// Dirs defines an application state while traversing a specific drive for
	// its files and directories.
	Dirs
)

// OnChangeLevel defines a custom type for a function being called on changing
// the current tree level. It accepts the current active entry instance and the
// navigation's state: Drives or Dirs.
type OnChangeLevel func(e *structure.Entry, s State)

type stackItem struct {
	entry  *structure.Entry
	cursor int
}

// Navigation defines the behavior for traversing the file system tree structure
// and handles the changes of state. It contains the current active drive/volume,
// traversing history, and handles the race condition cases.
type Navigation struct {
	tree         *structure.Tree
	entry        *structure.Entry
	drives       *drive.List
	currentDrive *drive.Info
	entryStack   []*stackItem
	state        State
	cursor       int
	locked       atomic.Bool
}

func NewNavigation(l *drive.List, t *structure.Tree) *Navigation {
	return &Navigation{
		tree:   t,
		state:  Drives,
		drives: l,
	}
}

// NewRootNavigation creates navigation for a predefined root directory entry.
// It starts the blocking traversal immediately rather than in interactive mode.
// Therefore, a root with a wide subdirectory structure might cause a delay.
func NewRootNavigation(l *drive.List, t *structure.Tree) (*Navigation, error) {
	if t.Root() == nil {
		return nil, errors.New("root is nil")
	}

	done, _ := t.TraverseAsync()
	<-done
	t.CalculateSize()

	return &Navigation{
		drives: l,
		tree:   t,
		entry:  t.Root(),
		state:  Dirs,
	}, nil
}

// OnDrives checks whether the current navigation state is Drives or not.
func (n *Navigation) OnDrives() bool {
	return n.state == Drives
}

// DrivesList returns the list of drives available on the system. Depending on
// the specific operating system, some volumes may be filtered. Refer to the
// corresponding drive.List implementation for details.
func (n *Navigation) DrivesList() *drive.List {
	return n.drives
}

// Entry returns the current active *structure.Entry instance. It never returns
// an instance of a file, but only a directory.
func (n *Navigation) Entry() *structure.Entry {
	return n.entry
}

func (n *Navigation) ParentSize() int64 {
	if n.entry == nil {
		//nolint:gosec // why not, let's overflow
		return int64(n.currentDrive.UsedBytes)
	}

	return n.entry.Size
}

// Up changes the current tree level up to the previous one. It doesn't accept
// the target level but instead takes it from navigation history. The Up function
// will change the level only if the current state is Dirs and will do nothing in
// case of the Drives state.
//
// If the navigation is currently locked, the function will do nothing and return
// immediately without an error.
func (n *Navigation) Up(ocl OnChangeLevel) {
	if n.OnDrives() || !n.lock() {
		return
	}

	defer n.unlock()

	if len(n.entryStack) == 0 {
		n.state, n.cursor = Drives, 0

		return
	}

	defer func() {
		ocl(n.entry, n.state)
	}()

	lastItem := n.entryStack[len(n.entryStack)-1]

	n.entry, n.cursor = lastItem.entry, lastItem.cursor
	n.entryStack = n.entryStack[:len(n.entryStack)-1]
}

// Down changes the current tree level down to the provided path. The path value
// is a directory relative path within the currently active tree node.
//
// It handles two scenarios. The first is changing the level from drives to dirs,
// and the second is traversing between directories. In the first case, the
// function will not block execution but instead returns channels for "done" and
// "errors" that occurred during the drive scan. The client is responsible for
// listening to the channels and handling the state of scanning. The navigation
// will be locked until the "done" channel is closed. In the second case, both
// channels will be returned as nil values, since the scanning is already done.
func (n *Navigation) Down(path string, cursor int, ocl OnChangeLevel) (chan struct{}, chan error) {
	if len(path) == 0 || !n.lock() {
		return nil, nil
	}

	// handle scenario when the drive was selected
	if n.OnDrives() {
		n.state = Dirs

		n.entry = structure.NewDirEntry(path, 0)
		n.currentDrive = n.drives.DriveInfo(path)
		n.tree.SetRoot(n.entry)

		doneChan, errChan := n.tree.TraverseAsync()

		go func() {
			<-doneChan
			n.unlock()
		}()

		return doneChan, errChan
	}

	// in other cases, it's just a lookup for a child directory
	defer func() {
		ocl(n.entry, n.state)
		n.unlock()
	}()

	entry := n.entry.GetChild(path)
	if entry == nil || !entry.IsDir {
		return nil, nil
	}

	n.entryStack = append(
		n.entryStack,
		&stackItem{entry: n.entry, cursor: cursor},
	)

	n.entry, n.cursor = entry, 0

	return nil, nil
}

func (n *Navigation) Explore(path string) error {
	if len(path) == 0 {
		return nil
	}

	var fullPath string

	if n.OnDrives() {
		d := n.drives.DriveInfo(path)
		if d == nil {
			return nil
		}

		fullPath = d.Path
	} else {
		entry := n.entry.GetChild(path)
		if entry == nil {
			return nil
		}

		fullPath = entry.Path
	}

	return drive.Explore(fullPath)
}

func (n *Navigation) lock() bool {
	return !n.locked.Swap(true)
}

func (n *Navigation) unlock() {
	n.locked.Swap(false)
}
