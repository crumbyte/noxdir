package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render"
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	exclude []string

	appCmd = &cobra.Command{
		Use:   "noxdir",
		Short: "Start a terminal utility for visualizing file system usage.",
		Long: `
📊 NoxDir is a terminal-based user interface for visualizing and analyzing disk
space usage across drives and volumes. It scans all directories and files on the
selected drive and presents the space consumption in a clear, user-friendly layout.

🔗 Learn more: https://github.com/crumbyte/noxdir`,
		RunE: runApp,
	}
)

func init() {
	appCmd.PersistentFlags().StringSliceVarP(
		&exclude,
		"exclude",
		"x",
		nil,
		`Exclude specific directories from scanning. Useful for directories 
with many subdirectories but minimal disk usage (e.g., node_modules). 

NOTE: The check targets any string occurrence. The excluded directory
name can be either an absolute path or only part of it. In the last case, 
all directories whose name contains that string will be excluded from
scanning.

Example: --exclude="node_modules,Steam\appcache"
(first rule will exclude all existing "node_modules" directories)`)
}

func Execute() {
	logger := slog.Default()

	if err := appCmd.Execute(); err != nil {
		// TODO: display error properly formatted
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runApp(_ *cobra.Command, _ []string) error {
	drivesList, err := drive.NewList()
	if err != nil {
		return fmt.Errorf("drive.NewList: %w", err)
	}

	var opts []structure.TreeOpt

	if len(exclude) > 0 {
		opts = append(opts, structure.WithExclude(exclude))
	}

	teaProg := tea.NewProgram(
		render.NewViewModel(
			render.NewNavigation(
				drivesList,
				structure.NewTree(nil, opts...),
			),
		),
		tea.WithAltScreen(),
	)

	render.SetTeaProgram(teaProg)

	if _, err = teaProg.Run(); err != nil {
		return fmt.Errorf("tea.Run: %w", err)
	}

	return nil
}
