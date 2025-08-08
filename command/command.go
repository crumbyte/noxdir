package command

import (
	"io"
	"strings"

	"github.com/crumbyte/noxdir/command/archive"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		DisableFlagParsing: true,
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootCmd.SetHelpCommand(NewHelpCmd())
	rootCmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {})

	rootCmd.AddCommand(archive.NewPackCmd(), archive.NewUnpackCmd())

	for _, cmd := range rootCmd.Commands() {
		cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
			_, _ = io.WriteString(cmd.OutOrStdout(), "Usage: "+ViewHelp(cmd))
		})
	}

	return rootCmd
}

// Execute executes the root command. It requires the input and output sources,
// as otherwise it'll try to read the STDIN and STDOUT streams. The "args" value
// contains a complete command expression split into a slice of strings, and the
// "out" represents io.Writer instance for writing the command execution result.
func Execute(root *cobra.Command, args []string, out io.Writer) error {
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs(args)

	return root.Execute()
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
