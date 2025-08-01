package render

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/crumbyte/noxdir/config"
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

// entryStack represents a stack of the visited entry. The last stack element
// will always contain the previous entry.
type entryStack []*stackItem

func (e *entryStack) len() int {
	return len(*e)
}

func (e *entryStack) push(si *stackItem) {
	*e = append(*e, si)
}

func (e *entryStack) pop() *stackItem {
	if len(*e) == 0 {
		return nil
	}

	item := (*e)[len(*e)-1]
	*e = (*e)[:len(*e)-1]

	return item
}

// Navigation defines the behavior for traversing the file system tree structure
// and handles the changes of state. It contains the current active drive/volume,
// traversing history, and handles the race condition cases.
type Navigation struct {
	tree         *structure.Tree
	entry        *structure.Entry
	drives       *drive.List
	currentDrive *drive.Info
	entryStack   *entryStack
	settings     config.Settings
	state        State
	cursor       int
	locked       atomic.Bool
	cacheEnabled bool
}

func NewNavigation(t *structure.Tree, s config.Settings) *Navigation {
	n := &Navigation{
		tree:         t,
		state:        Drives,
		entryStack:   &entryStack{},
		settings:     s,
		cacheEnabled: s.UseCache,
	}

	n.RefreshDrives()

	return n
}

// NewRootNavigation creates navigation for a predefined root directory entry.
// It starts the blocking traversal immediately rather than in interactive mode.
// Therefore, a root with a wide subdirectory structure might cause a delay.
func NewRootNavigation(t *structure.Tree, s config.Settings) (*Navigation, error) {
	if t.Root() == nil {
		return nil, errors.New("root is nil")
	}

	done, errChan := t.TraverseAsync(true)
	if done == nil {
		return nil, errors.New("root is nil")
	}

wait:
	for {
		select {
		case <-errChan:
			// ignore permission related errors for now
		case <-done:
			break wait
		}
	}

	<-done
	t.CalculateSize()

	n := NewNavigation(t, s)
	n.cacheEnabled = false

	n.state = Dirs
	n.entry = t.Root()

	return n, nil
}

// Settings returns the current application's settings.
func (n *Navigation) Settings() config.Settings {
	return n.settings
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
	minSize := int64(1)

	if n.entry == nil {
		//nolint:gosec // why not, let's overflow
		return max(minSize, int64(n.currentDrive.UsedBytes))
	}

	return max(minSize, n.entry.Size)
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

	if n.entryStack.len() == 0 {
		if err := n.tree.PersistCache(); err != nil {
			// ignore caching error
			_ = err
		}

		n.state, n.cursor = Drives, 0

		return
	}

	defer func() {
		ocl(n.entry, n.state)
	}()

	lastItem := n.entryStack.pop()
	n.entry, n.cursor = lastItem.entry, lastItem.cursor
}

// SetCursor preserves the current position of the table's cursor. The cursor
// position should be updated on each action and used during rendering.
func (n *Navigation) SetCursor(cursor int) {
	n.cursor = cursor
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

	// handle the scenario when the drive was selected
	if n.OnDrives() {
		n.state = Dirs

		n.entry = structure.NewDirEntry(path, 0)
		n.currentDrive = n.drives.DriveInfo(path)
		n.tree.SetRoot(n.entry)

		doneChan, errChan := n.tree.TraverseAsync(false)

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

	n.entryStack.push(&stackItem{entry: n.entry, cursor: cursor})
	n.entry, n.cursor = entry, 0

	return nil, nil
}

// RefreshDrives refreshes the list of the available drives and their memory
// usage data.
func (n *Navigation) RefreshDrives() {
	dl, err := drive.NewList()
	if err != nil {
		panic(err)
	}

	n.drives = dl
}

// RefreshEntry refreshes the current *Entry root by scanning its structure and
// updating the navigation state. The function will check the case when the
// current root has been deleted and tries to fall back to the previous *Entry
// in the stack. If all entries in the stack do not exist anymore, the navigation
// will fall back to the drive list.
//
// The navigation will be locked until the scanning is complete and the "done"
// channel is closed.
func (n *Navigation) RefreshEntry() (chan struct{}, chan error, error) {
	if n.OnDrives() || !n.lock() || n.entry == nil {
		return nil, nil, nil
	}

	for n.entryStack.len() > 0 || n.entry != nil {
		_, err := os.Lstat(n.entry.Path)
		if err == nil {
			break
		}

		if errors.Is(err, os.ErrNotExist) {
			lastItem := n.entryStack.pop()

			if lastItem != nil {
				n.entry, n.cursor = lastItem.entry, lastItem.cursor

				continue
			}

			n.entry, n.cursor = nil, 0
		}

		n.unlock()

		return nil, nil, fmt.Errorf("lstat: %w", err)
	}

	if n.entry == nil {
		n.state, n.cursor = Drives, 0
		n.unlock()

		return nil, nil, nil
	}

	n.entry.Child = nil

	t := n.tree.Clone(n.entry, structure.WithPartialRoot())

	doneChan, errChan := t.TraverseAsync(true)

	go func() {
		<-doneChan
		n.unlock()
	}()

	return doneChan, errChan, nil
}

func (n *Navigation) Explore(name string) error {
	if len(name) == 0 {
		return nil
	}

	var fullPath string

	if n.OnDrives() {
		d := n.drives.DriveInfo(name)
		if d == nil {
			return nil
		}

		fullPath = d.Path
	} else {
		entry := n.entry.GetChild(name)
		if entry == nil {
			return nil
		}

		fullPath = entry.Path
	}

	return drive.Explore(fullPath)
}

// Delete deletes the file or directory, including all internal content, from the
// file system by the provided base path value. The entry lookup will be done
// within the current active *Entry instance, limiting the deletion scope.
//
// If the entry was not found in the current active *Entry instance no error will
// be returned.
//
// TODO: add soft delete
func (n *Navigation) Delete(path string) error {
	entry := n.entry.GetChild(path)
	if entry == nil {
		return nil
	}

	if err := os.RemoveAll(entry.Path); err != nil {
		return fmt.Errorf("delete: path: %s: %w", path, err)
	}

	return nil
}

func (n *Navigation) Diff() (*structure.Tree, chan *structure.Diff, chan error) {
	if n.OnDrives() || !n.lock() || n.entry == nil {
		return nil, nil, nil
	}

	diffChan := make(chan *structure.Diff)
	entry := n.entry.Copy()
	tree := structure.NewTree(entry, structure.WithPartialRoot())

	doneChan, errChan := tree.TraverseAsync(true)

	go func() {
		<-doneChan

		diffChan <- n.entry.Diff(entry)
		close(diffChan)

		n.unlock()
	}()

	return tree, diffChan, errChan
}

func (n *Navigation) lock() bool {
	return !n.locked.Swap(true)
}

func (n *Navigation) unlock() {
	n.locked.Swap(false)
}
