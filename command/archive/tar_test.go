package archive_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crumbyte/noxdir/command/archive"
)

func TestArchive(t *testing.T) {
	tarArchiver := archive.NewTar(5<<20, archive.Gzip)

	archiveBuffer := bytes.NewBuffer(nil)

	filesPathContentMap := map[string]string{
		"file1.txt": "file 1 content",
		"file2.txt": "file 2 content",
		"file3.txt": "file 3 content",
	}

	t.Run("pack", func(t *testing.T) {
		inputPath := filepath.Join(t.TempDir(), "noxdir_tar_test")

		require.NoError(t, os.MkdirAll(inputPath, 0750))

		filesPath := make([]string, 0, len(filesPathContentMap))

		for path, content := range filesPathContentMap {
			file, err := os.Create(filepath.Join(inputPath, path))
			require.NoError(t, err)

			_, err = io.Copy(file, strings.NewReader(content))

			require.NoError(t, err)
			require.NoError(t, file.Close())

			filesPath = append(filesPath, filepath.Join(inputPath, path))
		}

		require.NoError(t, tarArchiver.Pack(filesPath, true, archiveBuffer))
		require.NoError(t, os.RemoveAll(inputPath))
	})

	t.Run("unpack", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "noxdir_untar_test")

		require.NoError(t, os.MkdirAll(outputPath, 0750))

		require.NoError(t, tarArchiver.Unpack(archiveBuffer, true, outputPath))

		for path, content := range filesPathContentMap {
			file, err := os.Open(filepath.Join(outputPath, path))
			require.NoError(t, err)

			fileContent, err := io.ReadAll(file)
			require.NoError(t, err)

			require.Equal(t, content, string(fileContent))
			require.NoError(t, file.Close())
		}

		require.NoError(t, os.RemoveAll(outputPath))
	})
}
