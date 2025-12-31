package checksum

import (
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewFileHashCmd() *cobra.Command {
	var (
		entries    []string
		hashType   string
		encodeType string
		ctxPath    string

		hashCmd = &cobra.Command{
			Short: "calculate file hash/checksum",
			Use:   "hash",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return run(cmd, entries, hashType, encodeType, ctxPath)
			},
		}
	)

	hashCmd.PersistentFlags().StringSliceVarP(&entries, "entries", "", nil, "")
	hashCmd.PersistentFlags().StringVarP(&hashType, "type", "t", "", "md5,sha1,sha224,sha256,sha384,sha512")
	hashCmd.PersistentFlags().StringVarP(&encodeType, "encode", "e", "hex", "hex,hex-up,base64")
	hashCmd.PersistentFlags().StringVarP(&ctxPath, "ctx-path", "", "", "")

	_ = hashCmd.MarkPersistentFlagRequired("ctx-path")
	_ = hashCmd.MarkPersistentFlagRequired("entries")
	_ = hashCmd.MarkPersistentFlagRequired("type")

	hashCmd.Flag("entries").Hidden = true
	hashCmd.Flag("ctx-path").Hidden = true

	return hashCmd
}

func run(cmd *cobra.Command, entries []string, hashType, encodeType, ctxPath string) error {
	if len(entries) == 0 {
		return nil
	}

	filePath := filepath.Join(ctxPath, filepath.Base(entries[0]))
	fh := NewFileHash()

	rawHash, err := fh.Calculate(hashType, filePath)
	if err != nil {
		return err
	}

	fmtHash, err := fh.Format(rawHash, Encoding(encodeType))
	if err != nil {
		return err
	}

	_, err = io.WriteString(cmd.OutOrStdout(), fmtHash)

	return err
}
