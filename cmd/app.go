package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/crumbyte/noxdir/config"
	"github.com/crumbyte/noxdir/drive"
	"github.com/crumbyte/noxdir/filter"
	"github.com/crumbyte/noxdir/pkg/cache"
	"github.com/crumbyte/noxdir/render"
	"github.com/crumbyte/noxdir/structure"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	ErrUnknown = errors.New("unknown error")

	exclude         []string
	root            string
	sizeLimit       string
	noEmptyDirs     bool
	noHidden        bool
	colorSchemaPath string
	useCache        bool
	clearCache      bool
	simpleCS        bool
	settings        *config.Settings

	tree *structure.Tree

	appCmd = &cobra.Command{
		Use:   "noxdir",
		Short: "Start a terminal utility for visualizing file system usage.",
		Long: `
📊 NoxDir is a terminal-based user interface for visualizing and analyzing disk
space usage across drives and volumes. It scans all directories and files on the
selected drive and presents the space consumption in a clear, user-friendly layout.

🔗 Learn more: https://github.com/crumbyte/noxdir`,
		RunE: runApp,
		PostRun: func(_ *cobra.Command, _ []string) {
			if tree == nil || !settings.UseCache {
				return
			}

			if err := tree.PersistCache(); err != nil {
				printError(err.Error())
			}
		},
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

	appCmd.PersistentFlags().BoolVarP(
		&noHidden,
		"no-hidden",
		"",
		false,
		`Excludes all hidden files and directories from the output. The entry is
considered hidden if its name starts with a dot, e.g., ".git".

Default value is "false".

Example: --no-hidden (provide a flag)
`,
	)

	appCmd.PersistentFlags().StringVarP(
		&colorSchemaPath,
		"color-schema",
		"",
		"",
		`Set the color schema configuration file. The file contains a custom
color settings for the UI elements.
`,
	)

	appCmd.PersistentFlags().BoolVarP(
		&useCache,
		"use-cache",
		"c",
		false,
		`Force the application to cache the data. With cache enabled, the full
file system scan will be performed only once. After that, the cache will be
used as long as the flag is provided.

The cache will always store the last session data. In order to update the
cache and the application's state, use the "r" (refresh) command on a 
target directory.

Default value is "false".

Example: -c|--use-cache (provide a flag)
`,
	)

	appCmd.PersistentFlags().BoolVarP(
		&clearCache,
		"clear-cache",
		"",
		false,
		`Delete all cache files from the application's directory.

Example: --clear-cache (provide a flag)
`,
	)

	appCmd.PersistentFlags().BoolVarP(
		&simpleCS,
		"simple-color",
		"",
		false,
		`Use a simplified color schema without emojis and glyphs.

Example: --simple-color (provide a flag)
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

func initConfig() (*config.Settings, error) {
	var err error

	settings, err = config.LoadSettings()
	if err != nil {
		return nil, err
	}

	if colorSchemaPath != "" {
		settings.ColorSchema = colorSchemaPath
	}

	if simpleCS {
		settings.SimpleColor = simpleCS
	}

	if noHidden {
		settings.NoHidden = true
	}

	if noEmptyDirs {
		settings.NoEmptyDirs = true
	}

	if useCache {
		settings.UseCache = true
	}

	if len(exclude) != 0 {
		settings.Exclude = exclude
	}

	return settings, nil
}

func runApp(_ *cobra.Command, _ []string) error {
	var (
		err   error
		style *render.Style
	)

	settings, err = initConfig()
	if err != nil {
		return err
	}

	if style, err = initStyle(settings); err != nil {
		return err
	}

	render.InitKeyMap(&settings.Bindings, style)

	vm, err := initViewModel(settings)
	if err != nil {
		return err
	}

	teaProg := tea.NewProgram(
		vm,
		tea.WithAltScreen(),
		tea.WithoutCatchPanics(),
	)

	defer func() {
		if r := recover(); r != nil {
			var ok bool

			_ = teaProg.ReleaseTerminal()

			if err, ok = r.(error); !ok {
				err = ErrUnknown
			}

			printError(render.ReportError(err, debug.Stack()))
		}
	}()

	render.SetTeaProgram(teaProg)

	if _, err = teaProg.Run(); err != nil {
		return err
	}

	return nil
}

func initViewModel(s *config.Settings) (*render.ViewModel, error) {
	nav, err := resolveNavigation(s)
	if err != nil {
		return nil, err
	}

	var dirModelFilters []filter.EntryFilter

	if s.NoEmptyDirs {
		dirModelFilters = append(dirModelFilters, &filter.EmptyDirFilter{})
	}

	vm := render.NewViewModel(
		nav,
		render.NewDriveModel(nav),
		render.NewDirModel(nav, dirModelFilters...),
	)

	if root != "" {
		vm.Update(render.ScanFinished{Mode: render.READY})
	}

	return vm, nil
}

func resolveNavigation(s *config.Settings) (*render.Navigation, error) {
	var (
		opts          []structure.TreeOpt
		fif           []drive.FileInfoFilter
		cacheInstance *cache.Cache
	)

	if len(s.Exclude) > 0 {
		opts = append(opts, structure.WithExclude(s.Exclude))
	}

	sizeLimitFilter, err := parseSizeLimit()
	if err != nil {
		return nil, NewCLIError(
			fmt.Errorf("invalid value for size-limit flag: %s", err.Error()),
		)
	}

	if sizeLimitFilter != nil {
		fif = append(fif, sizeLimitFilter)
	}

	if s.NoHidden {
		fif = append(fif, drive.HiddenFilter)
	}

	if s.UseCache || clearCache {
		cacheInstance, err = cache.NewCache(
			func(w io.Writer) cache.Encoder {
				return structure.NewEncoder(w)
			},
			func(r io.Reader) cache.Decoder {
				return structure.NewDecoder(r)
			},
			clearCache,
			s.Path,
			cache.WithCompress(),
		)
		if err != nil {
			return nil, err
		}
	}

	opts = append(
		opts,
		structure.WithFileInfoFilter(fif),
		structure.WithCache(cacheInstance),
	)

	if root != "" {
		root = strings.TrimSuffix(root, string(os.PathSeparator))

		if root, err = filepath.Abs(root); err != nil {
			return nil, fmt.Errorf("resolve absolute root rpath: %s", err.Error())
		}

		tree = structure.NewTree(
			structure.NewDirEntry(root, time.Now().Unix()),
			append(opts, structure.WithPartialRoot())...,
		)

		return render.NewRootNavigation(tree, *settings)
	}

	tree = structure.NewTree(nil, opts...)

	return render.NewNavigation(tree, *settings), nil
}

func printError(errMsg string) {
	if _, err := os.Stdout.WriteString(errMsg + "\n"); err != nil {
		return
	}
}

func initStyle(s *config.Settings) (*render.Style, error) {
	cs := render.DefaultColorSchema()

	if s.SimpleColor {
		cs = render.SimpleColorSchema()
	}

	if len(s.ColorSchema) != 0 {
		if err := render.DecodeColorSchema(s.ColorSchema, &cs); err != nil {
			return nil, err
		}
	}

	return render.InitStyle(cs), nil
}
