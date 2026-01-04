package archive

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewPackCmd(onStateChange func()) *cobra.Command {
	var (
		entries     []string
		compression string
		output      string
		ctxPath     string
		extract     bool

		packCmd = &cobra.Command{
			Short: "archive dirs/files",
			Use:   "pack",
			RunE: func(_ *cobra.Command, _ []string) error {
				return packRun(
					entries, compression, output, ctxPath, extract, onStateChange,
				)
			},
		}
	)

	packCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	packCmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "zst,gz")
	packCmd.PersistentFlags().BoolVarP(&extract, "extract", "x", false, "extract from archive")
	packCmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	packCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	_ = packCmd.MarkPersistentFlagRequired("ctx-path")
	_ = packCmd.MarkPersistentFlagRequired("entries")
	_ = packCmd.MarkPersistentFlagRequired("output")

	packCmd.Flag("entries").Hidden = true
	packCmd.Flag("ctx-path").Hidden = true

	return packCmd
}

func packRun(entries []string, compression, output, ctxPath string, extract bool, onStateChange func()) error {
	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	archiver := NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	)

	var err error

	if extract {
		err = archiver.UnpackFromFile(entries[0], output)
	} else {
		err = archiver.PackToFile(entries, output)
	}

	if err != nil {
		return err
	}

	onStateChange()

	return nil
}
