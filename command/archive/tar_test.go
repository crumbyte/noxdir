package archive_test

import (
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crumbyte/noxdir/command/archive"
)

func TestArchive(t *testing.T) {
	archiveName := "testArchive"
	archiveFileName := archiveName + ".tar.gz"

	t.Run("pack", func(t *testing.T) {
		filesPathContentMap := map[string]string{
			"file1.txt": "file 1 content",
			"file2.txt": "file 2 content",
			"file3.txt": "file 3 content",
		}

		for path, content := range filesPathContentMap {
			file, err := os.Create(path)
			require.NoError(t, err)

			_, err = io.Copy(file, strings.NewReader(content))

			require.NoError(t, err)
			require.NoError(t, file.Close())
		}

		err := archive.Pack(
			slices.Collect(maps.Keys(filesPathContentMap)),
			true,
			archiveName,
		)

		require.NoError(t, err)

		for _, path := range slices.Collect(maps.Keys(filesPathContentMap)) {
			require.NoError(t, os.Remove(path))
		}
	})

	t.Run("unpack", func(t *testing.T) {
		outputPath := filepath.Join(os.TempDir(), "noxdir_tar_test")

		err := archive.Unpack(archiveFileName, outputPath)
		require.NoError(t, err)

		require.NoError(t, os.RemoveAll(outputPath))
	})

	require.NoError(t, os.Remove(archiveFileName))
}
