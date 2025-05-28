package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/render"
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	exclude []string
	root    string

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

	appCmd.PersistentFlags().StringVarP(
		&root,
		"root",
		"r",
		"",
		`Start from a predefined root directory. Instead of selecting the target
drive and scanning all folders within, a root directory can be provided. 
In this case, the scanning will be performed exclusively for the specified
directory, drastically reducing the scanning time.

Providing an invalid path results in a blank application output. In this 
case, a "backspace" still can be used to return to the drives list. Also, all
trailing slash characters will be removed from the provided path.

Example: --root="C:\Program Files (x86)"
`,
	)
}

func Execute() {
	if err := appCmd.Execute(); err != nil {
		printError(err, debug.Stack())

		os.Exit(1)
	}
}

func runApp(_ *cobra.Command, _ []string) error {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = errors.New("unknown error")
			}

			printError(err, debug.Stack())
		}
	}()

	vm, err := initViewModel()
	if err != nil {
		return err
	}

	teaProg := tea.NewProgram(
		vm,
		tea.WithAltScreen(),
		tea.WithoutCatchPanics(),
	)

	render.SetTeaProgram(teaProg)

	if _, err = teaProg.Run(); err != nil {
		return err
	}

	return nil
}

func printError(err error, stackTrace []byte) {
	report := render.ReportError(err, stackTrace)

	_, err = os.Stdout.WriteString(report)
	if err != nil {
		return
	}
}

func initViewModel() (*render.ViewModel, error) {
	nav, err := resolveNavigation()
	if err != nil {
		return nil, err
	}

	vm := render.NewViewModel(nav)

	if root != "" {
		vm.Update(render.ScanFinished{})
	}

	return vm, nil
}

func resolveNavigation() (*render.Navigation, error) {
	drivesList, err := drive.NewList()
	if err != nil {
		return nil, fmt.Errorf("drive.NewList: %w", err)
	}

	if root != "" {
		root = strings.TrimSuffix(root, string(os.PathSeparator))

		return render.NewRootNavigation(
			drivesList,
			structure.NewTree(
				structure.NewDirEntry(root, time.Now().Unix()),
			),
		)
	}

	var opts []structure.TreeOpt

	if len(exclude) > 0 {
		opts = append(opts, structure.WithExclude(exclude))
	}

	return render.NewNavigation(
		drivesList,
		structure.NewTree(nil, opts...),
	), nil
}
