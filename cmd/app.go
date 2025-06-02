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
	ErrUnknown = errors.New("unknown error")

	exclude     []string
	root        string
	sizeLimit   string
	noEmptyDirs bool

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

// TODO: add flag for skipping empty folders
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
case, a "backspace" still can be used to return to the drives list.
Also, all trailing slash characters will be removed from the provided
path.

Example: --root="C:\Program Files (x86)"`)

	appCmd.PersistentFlags().StringVarP(
		&sizeLimit,
		"size-limit",
		"l",
		"",
		`Define size limits/boundaries for files that should be shown in the
scanner output. Files that do not fit in the provided limits will be
skipped.

The size limits can be defined using format "<size><unit>:<size><unit>
where "unit" value can be: KB, MB, GB, TB, PB, and "size" is a positive 
numeric value. For example: "1GB:5GB".

Both values are optional. Therefore, it can also be an upper bound only
or a lower bound only. These are the valid flag values: "1GB:", ":10GB"

NOTE: providing this flag will lead to inaccurate sizes of the
directories, since the calculation process will include only files
that meet the boundaries. Also, this flag cannot be applied to the
directories but only to files within.

Example:
	--size-limit="3GB:20GB"
	--size-limit="3MB:"
	--size-limit=":1TB"
`,
	)

	appCmd.PersistentFlags().BoolVarP(
		&noEmptyDirs,
		"no-empty-dirs",
		"d",
		false,
		`Excludes all empty directories from the output. The directory is
considered empty if it or its subdirectories do not contain any files.

Even if the specific directory represents the entire tree structure of 
subdirectories, without a single file, it will be completely skipped.

Default value is "false".

Example: --no-empty-dirs (provide a flag)
`,
	)
}

func Execute() {
	if err := appCmd.Execute(); err != nil {
		var cliErr *CLIError

		if errors.As(err, &cliErr) {
			printError(cliErr.Error())
		} else {
			printError(render.ReportError(err, debug.Stack()))
		}

		os.Exit(1)
	}
}

func runApp(_ *cobra.Command, _ []string) error {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = ErrUnknown
			}

			printError(render.ReportError(err, debug.Stack()))
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

	var opts []structure.TreeOpt

	if len(exclude) > 0 {
		opts = append(opts, structure.WithExclude(exclude))
	}

	sizeLimitFilter, err := parseSizeLimit()
	if err != nil {
		return nil, NewCLIError(
			fmt.Errorf("invalid value for size-limit flag: %s", err.Error()),
		)
	}

	if sizeLimitFilter != nil {
		opts = append(
			opts,
			structure.WithFileInfoFilter(
				[]drive.FileInfoFilter{sizeLimitFilter},
			),
		)
	}

	if root != "" {
		root = strings.TrimSuffix(root, string(os.PathSeparator))

		return render.NewRootNavigation(
			drivesList,
			structure.NewTree(
				structure.NewDirEntry(root, time.Now().Unix()),
				opts...,
			),
		)
	}

	return render.NewNavigation(
		drivesList,
		structure.NewTree(nil, opts...),
	), nil
}

func printError(errMsg string) {
	if _, err := os.Stdout.WriteString(errMsg + "\n"); err != nil {
		return
	}
}
