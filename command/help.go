package command

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func NewHelpCmd() *cobra.Command {
	var (
		packCmd = &cobra.Command{
			Short:              "execute this command",
			Use:                "help",
			DisableFlagParsing: true,
			RunE:               helpCmd,
		}
	)

	return packCmd
}

func helpCmd(cmd *cobra.Command, _ []string) error {
	rootCmd := cmd.Root()

	if len(rootCmd.Commands()) == 0 {
		_, err := io.WriteString(cmd.OutOrStdout(), "no commands found")

		return err
	}

	commandsHelp := make([]string, 0, len(rootCmd.Commands()))

	for _, subCommands := range rootCmd.Commands() {
		commandsHelp = append(commandsHelp, subCommands.Name()+" - "+subCommands.Short)
	}

	_, err := io.WriteString(
		cmd.OutOrStdout(), strings.Join(commandsHelp, " | "),
	)

	return err
}
