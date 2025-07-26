package archive

import (
	"github.com/spf13/cobra"
)

var (
	entries     []string
	compression string
	archiveType string
	output      string

	Cmd = &cobra.Command{
		Use:  "pack",
		RunE: archiveCmdRun,
	}
)

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&entries, "entries", "e", nil, "")
	Cmd.PersistentFlags().StringVarP(&compression, "compression", "c", "", "")
	Cmd.PersistentFlags().StringVarP(&archiveType, "type", "t", "tar", "")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "")
}

func archiveCmdRun(_ *cobra.Command, _ []string) error {
	ct := CompressionType(0).FromString(compression)

	return NewTar(DefaultBufferSize, ct).PackToFile(entries, output)
}
