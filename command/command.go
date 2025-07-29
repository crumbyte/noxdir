package command

import (
	"io"
	"strings"

	"github.com/crumbyte/noxdir/command/archive"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RootCmd represents a root command for all subcommands. This command does not
// contain any behavior but only proxies the execution to the subcommands.
var RootCmd = &cobra.Command{
	SilenceUsage:       true,
	DisableSuggestions: true,
}

func init() {
	RootCmd.AddCommand(archive.PackCmd, archive.UnpackCmd)

	archive.PackCmd.SetHelpFunc(
		func(cmd *cobra.Command, _ []string) {
			_, _ = io.WriteString(cmd.OutOrStdout(), "Usage: "+ViewHelp(cmd))
		},
	)

	archive.UnpackCmd.SetHelpFunc(
		func(cmd *cobra.Command, _ []string) {
			_, _ = io.WriteString(cmd.OutOrStdout(), "Usage: "+ViewHelp(cmd))
		},
	)
}

// Execute executes the root command. It requires the input and output sources,
// as otherwise it'll try to read the STDIN and STDOUT streams. The "args" value
// contains a complete command expression split into a slice of strings, and the
// "out" represents io.Writer instance for writing the command execution result.
func Execute(args []string, out io.Writer) error {
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs(args)

	return RootCmd.Execute()
}

func ViewHelp(cmd *cobra.Command) string {
	parts := []string{cmd.Use}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		name := "--" + f.Name

		if f.Shorthand != "" {
			name = "-" + f.Shorthand + " | " + name
		}

		parts = append(parts, name, "<"+f.Value.Type()+">")
	})

	return strings.Join(parts, " ")
}
