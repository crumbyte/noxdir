package archive

import (
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	entries     []string
	compression string
	output      string
	ctxPath     string

	Cmd = &cobra.Command{
		Use:  "pack",
		RunE: archiveCmdRun,
	}
)

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	Cmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	Cmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")
}

func archiveCmdRun(_ *cobra.Command, _ []string) error {
	if len(ctxPath) == 0 {
		return errors.New("base path required")
	}

	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	return NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	).PackToFile(entries, output)
}
