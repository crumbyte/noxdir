package command

import (
	"io"

	"github.com/crumbyte/noxdir/command/archive"

	"github.com/spf13/cobra"
)

// RootCmd represents a root command for all subcommands. This command does not
// contain any behavior but only proxies the execution to the subcommands.
var RootCmd = &cobra.Command{
	SilenceUsage:       true,
	DisableSuggestions: true,
}

func init() {
	RootCmd.AddCommand(archive.Cmd)
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
