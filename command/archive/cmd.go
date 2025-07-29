package archive

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	entries     []string
	compression string
	output      string
	ctxPath     string

	PackCmd = &cobra.Command{
		Use:  "pack",
		RunE: packRun,
	}

	UnpackCmd = &cobra.Command{
		Use:  "unpack",
		RunE: unpackRun,
	}
)

func init() {
	PackCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	PackCmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "")
	PackCmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	PackCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	PackCmd.Flag("entries").Hidden = true
	PackCmd.Flag("ctx-path").Hidden = true

	UnpackCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	UnpackCmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	UnpackCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	UnpackCmd.Flag("entries").Hidden = true
	UnpackCmd.Flag("ctx-path").Hidden = true

	_ = PackCmd.MarkPersistentFlagRequired("ctx-path")
	_ = PackCmd.MarkPersistentFlagRequired("entries")
	_ = PackCmd.MarkPersistentFlagRequired("output")

	_ = UnpackCmd.MarkPersistentFlagRequired("ctx-path")
	_ = UnpackCmd.MarkPersistentFlagRequired("archive")
}

func packRun(_ *cobra.Command, _ []string) error {
	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	return NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	).PackToFile(entries, output)
}

func unpackRun(_ *cobra.Command, _ []string) error {
	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	return NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	).UnpackFromFile(entries[0], output)
}
