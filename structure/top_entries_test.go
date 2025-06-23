package structure_test

import (
	"slices"
	"testing"

	"github.com/crumbyte/noxdir/structure"

	"github.com/stretchr/testify/require"
)

func TestTopEntries_ScanFiles(t *testing.T) {
	te := structure.NewTopEntries(5)

	tfEntry := &structure.Entry{
		Path: "root",
		Child: []*structure.Entry{
			{Path: "root_file_1", Size: 100},
			{Path: "root_file_2", Size: 150},
			{Path: "root_file_3", Size: 200},
			{Path: "root_file_4", Size: 650},
			{
				Path: "level1",
				Child: []*structure.Entry{
					{Path: "level1_file_1", Size: 250},
					{Path: "level1_file_2", Size: 300},
					{Path: "level1_file_3", Size: 700},
					{Path: "level1_file_4", Size: 400},
					{
						Path: "level2",
						Child: []*structure.Entry{
							{Path: "level2_file_1", Size: 450},
							{Path: "level2_file_2", Size: 500},
							{Path: "level2_file_3", Size: 550},
							{Path: "level2_file_4", Size: 600},
						},
						IsDir: true,
					},
				},
				IsDir: true,
			},
		},
		IsDir: true,
	}

	te.ScanFiles(tfEntry)

	expected := []string{
		"root_file_4",
		"level2_file_4",
		"level2_file_3",
		"level2_file_2",
		"level1_file_3",
	}

	require.Equal(t, len(expected), te.Files().Len())

	for te.Files().Len() > 0 {
		tf, ok := te.Files().Pop().(*structure.Entry)
		require.True(t, ok)

		require.True(t, slices.Contains(expected, tf.Name()))
	}
}

func TestTopEntries_ScanDirs(t *testing.T) {
	te := structure.NewTopEntries(3)

	tfEntry := &structure.Entry{
		Path: "root",
		Child: []*structure.Entry{
			{
				Path: "level1_1",
				Child: []*structure.Entry{
					{
						Path: "level1_1_file_1", Size: 100,
					},
				},
				IsDir: true,
				Size:  100,
			},
			{
				Path: "level1_2",
				Child: []*structure.Entry{
					{
						Path: "level1_2_file_1", Size: 150,
					},
				},
				IsDir: true,
				Size:  150,
			},
			{
				Path: "level1_3",
				Child: []*structure.Entry{
					{
						Path: "level1_3_file_1", Size: 200,
					},
					{
						Path: "level1_3_dir_1",
						Child: []*structure.Entry{
							{
								Path: "level1_2_file_1", Size: 250,
							},
							{
								Path: "level1_2_file_2", Size: 300,
							},
						},
						Size:  550,
						IsDir: true,
					},
				},
				Size:  750,
				IsDir: true,
			},
		},
		IsDir: true,
	}

	te.ScanDirs(tfEntry)

	expected := []string{
		"level1_2",
		"level1_3_dir_1",
		"level1_1",
	}

	require.Equal(t, len(expected), te.Dirs().Len())

	for te.Dirs().Len() > 0 {
		tf, ok := te.Dirs().Pop().(*structure.Entry)
		require.True(t, ok)

		require.True(t, slices.Contains(expected, tf.Name()))
	}
}
