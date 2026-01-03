package command

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateSubCommand defines a custom type for a function that creates a new
// command instance. The produced command will be assigned as a sub-command to
// the root command.
type CreateSubCommand func(onStateChange func()) *cobra.Command

func NewRootCmd(onStateChange func(), csc ...CreateSubCommand) *cobra.Command {
	rootCmd := &cobra.Command{
		DisableFlagParsing: true,
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootCmd.SetHelpCommand(NewHelpCmd())
	rootCmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {})

	for _, subCommand := range csc {
		rootCmd.AddCommand(subCommand(onStateChange))
	}

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
			name = "-" + f.Shorthand + "|" + name
		}

		usage := f.Value.Type()

		if len(f.Usage) != 0 {
			usage = f.Usage
		}

		parts = append(parts, "["+name+"]", "<"+usage+">")
	})

	return strings.Join(parts, " ")
}
