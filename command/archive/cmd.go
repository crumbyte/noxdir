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

		packCmd = &cobra.Command{
			Short: "archive dirs/files",
			Use:   "pack",
			RunE: func(_ *cobra.Command, _ []string) error {
				return packRun(
					entries, compression, output, ctxPath, onStateChange,
				)
			},
		}
	)

	packCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	packCmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "zst,gz")
	packCmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	packCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	_ = packCmd.MarkPersistentFlagRequired("ctx-path")
	_ = packCmd.MarkPersistentFlagRequired("entries")
	_ = packCmd.MarkPersistentFlagRequired("output")

	packCmd.Flag("entries").Hidden = true
	packCmd.Flag("ctx-path").Hidden = true

	return packCmd
}

func NewUnpackCmd(onStateChange func()) *cobra.Command {
	var (
		entries     []string
		compression string
		output      string
		ctxPath     string

		unpackCmd = &cobra.Command{
			Short: "unarchive dirs/files",
			Use:   "unpack",
			RunE: func(_ *cobra.Command, _ []string) error {
				return unpackRun(entries, compression, output, ctxPath, onStateChange)
			},
		}
	)

	unpackCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	unpackCmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "zst,gz")
	unpackCmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
	unpackCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	unpackCmd.Flag("entries").Hidden = true
	unpackCmd.Flag("ctx-path").Hidden = true

	return unpackCmd
}

func packRun(entries []string, compression, output, ctxPath string, onStateChange func()) error {
	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	err := NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	).PackToFile(entries, output)
	if err != nil {
		return err
	}

	onStateChange()

	return nil
}

func unpackRun(entries []string, compression, output, ctxPath string, onStateChange func()) error {
	output = filepath.Join(ctxPath, filepath.Base(filepath.Clean(output)))

	for i := range entries {
		entries[i] = filepath.Join(ctxPath, filepath.Base(entries[i]))
	}

	err := NewTar(
		DefaultBufferSize,
		NoCompression.FromString(compression),
	).UnpackFromFile(entries[0], output)
	if err != nil {
		return err
	}

	onStateChange()

	return nil
}
