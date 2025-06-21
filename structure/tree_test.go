package structure_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crumbyte/noxdir/structure"

	"github.com/stretchr/testify/require"
)

type testEntry struct {
	name  string
	files []string
	dirs  []testEntry
}

var testEntryInstance = testEntry{
	name:  "noxdir_root_test_entry",
	files: []string{"root_file_1", "root_file_2", "root_file_3"},
	dirs: []testEntry{
		{
			name:  "level_1_1",
			files: []string{"level_1_file_1", "level_1_file_2"},
		},
		{
			name:  "level_1_2",
			files: []string{"level_1_file_1", "level_1_file_2"},
		},
		{
			name:  "level_1_3",
			files: []string{"level_1_file_1", "level_1_file_2"},
			dirs: []testEntry{
				{
					name:  "level_2_1",
					files: []string{"level_2_file_1", "level_2_file_2"},
				},
				{
					name:  "level_2_2",
					files: []string{"level_2_file_1", "level_2_file_2"},
					dirs: []testEntry{
						{
							name:  "level_3_1",
							files: []string{"level_3_file_1", "level_3_file_2"},
						},
						{
							name:  "level_3_2",
							files: []string{"level_3_file_1", "level_3_file_2"},
						},
					},
				},
				{
					name:  "level_2_3",
					files: []string{"level_2_file_1", "level_2_file_2"},
				},
			},
		},
		{
			name:  "level_1_4",
			files: []string{"level_1_file_1", "level_1_file_2"},
		},
	},
}

func TestTree_Traverse(t *testing.T) {
	root, err := filepath.Abs(".")
	require.NoError(t, err)

	entryRoot := initTmpEntry(t, &testEntryInstance, root)

	e := structure.NewDirEntry(entryRoot, 0)
	tree := structure.NewTree(e)

	require.NoError(t, tree.Traverse(true))
	tree.CalculateSize()

	require.Equal(t, uint64(4), e.LocalDirs)
	require.Equal(t, uint64(3), e.LocalFiles)

	require.Equal(t, uint64(21), e.TotalFiles)
	require.Equal(t, uint64(9), e.TotalDirs)

	verifyEntryStructure(t, e, &testEntryInstance)

	require.NoError(t, os.RemoveAll(entryRoot))
}

func TestTree_TraverseExclude(t *testing.T) {
	root, err := filepath.Abs(".")
	require.NoError(t, err)

	entryRoot := initTmpEntry(t, &testEntryInstance, root)

	tableData := []struct {
		exclude          []string
		expectedDirsCnt  uint64
		expectedFilesCnt uint64
	}{
		{
			exclude:          []string{"noxdir_root_test_entry"},
			expectedDirsCnt:  0,
			expectedFilesCnt: 0,
		},
		{
			exclude:          []string{"level_1_1"},
			expectedDirsCnt:  9,
			expectedFilesCnt: 19,
		},
		{
			exclude:          []string{"level_2"},
			expectedDirsCnt:  7,
			expectedFilesCnt: 11,
		},
		{
			exclude:          []string{"level_3", "level_1_4"},
			expectedDirsCnt:  9,
			expectedFilesCnt: 15,
		},
	}

	for _, data := range tableData {
		t.Run(
			"exclude: "+strings.Join(data.exclude, ","),
			func(t *testing.T) {
				e := structure.NewDirEntry(entryRoot, 0)
				tree := structure.NewTree(
					e, structure.WithExclude(data.exclude),
				)

				require.NoError(t, tree.Traverse(true))
				tree.CalculateSize()

				require.Equal(t, data.expectedDirsCnt, e.TotalDirs)
				require.Equal(t, data.expectedFilesCnt, e.TotalFiles)
			},
		)
	}

	require.NoError(t, os.RemoveAll(entryRoot))
}

func TestTree_TraverseAsync(t *testing.T) {
	root, err := filepath.Abs(".")
	require.NoError(t, err)

	entryRoot := initTmpEntry(t, &testEntryInstance, root)

	e := structure.NewDirEntry(entryRoot, 0)
	tree := structure.NewTree(e)

	done, errCh := tree.TraverseAsync(true)

	select {
	case err = <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second * 3):
		t.Fatalf("traverse async failed on timeout")
	case <-done:
		break
	}

	tree.CalculateSize()

	require.Equal(t, uint64(4), e.LocalDirs)
	require.Equal(t, uint64(3), e.LocalFiles)

	require.Equal(t, uint64(21), e.TotalFiles)
	require.Equal(t, uint64(9), e.TotalDirs)

	verifyEntryStructure(t, e, &testEntryInstance)

	require.NoError(t, os.RemoveAll(entryRoot))
}

func TestEntry_AddChild(t *testing.T) {
	e := structure.NewDirEntry("root", 0)

	require.False(t, e.HasChild())

	type tableItem struct {
		name  string
		isDir bool
	}

	tableData := []tableItem{
		{"root_file_1", false},
		{"root_file_2", false},
		{"root_file_3", false},
		{"root_dir_1", true},
		{"root_dir_2", true},
		{"root_dir_3", true},
	}

	for i := range tableData {
		path := "root" + string(os.PathSeparator) + tableData[i].name

		childEntry := structure.NewFileEntry(path, 1, 0)

		if tableData[i].isDir {
			childEntry = structure.NewDirEntry(path, 0)
		}

		e.AddChild(childEntry)

		require.NotNil(t, e.GetChild(tableData[i].name))
	}

	require.Len(t, e.Child, 6)
	require.EqualValues(t, 3, e.LocalDirs)
	require.EqualValues(t, 3, e.LocalFiles)
	require.EqualValues(t, 3, e.TotalFiles)
	require.EqualValues(t, 3, e.TotalDirs)
}

func verifyEntryStructure(t *testing.T, e *structure.Entry, te *testEntry) {
	t.Helper()

	require.Equal(t, te.name, e.Name())

	for i := range te.files {
		c := e.GetChild(te.files[i])

		require.NotNil(t, c, "child %s not found", te.files[i])
		require.Equal(t, te.files[i], c.Name())
	}

	for i := range te.dirs {
		d := e.GetChild(te.dirs[i].name)

		require.NotNil(t, d)
		verifyEntryStructure(t, d, &te.dirs[i])
	}
}

func initTmpEntry(t *testing.T, et *testEntry, parent string) string {
	t.Helper()

	root := parent + string(os.PathSeparator) + et.name

	absRootPath, err := filepath.Abs(root)
	require.NoError(t, err)

	if _, err = os.Stat(absRootPath); !errors.Is(err, os.ErrNotExist) {
		return absRootPath
	}

	require.NoError(t, os.Mkdir(root, 0777))

	for i := range et.files {
		var f *os.File

		f, err = os.Create(root + string(os.PathSeparator) + et.files[i])
		require.NoError(t, err)
		_ = f.Close()
	}

	for i := range et.dirs {
		initTmpEntry(t, &et.dirs[i], root)
	}

	return absRootPath
}
