# 🧹 NoxDir

[![Build](https://github.com/crumbyte/noxdir/actions/workflows/build.yml/badge.svg)](https://github.com/crumbyte/noxdir/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/crumbyte/noxdir)](https://goreportcard.com/report/github.com/crumbyte/noxdir)

**NoxDir** is a high-performance, cross-platform command-line tool for
visualizing and exploring your file system usage. It detects mounted drives or
volumes and presents disk usage metrics through a responsive, keyboard-driven
terminal UI. Designed to help you quickly locate space hogs and streamline your
cleanup workflow.

## 🚀 Features

- ✅ Cross-platform drive and mount point detection (**Windows**, **macOS**, **Linux**)
- 📊 Real-time disk usage insights: used, free, total capacity, and utilization
  percentage
- 🖥️ Interactive and intuitive terminal interface with keyboard navigation
- ⚡ Built for speed — uses native system calls for maximum performance
- 🔒 Fully local and privacy-respecting — **no telemetry**, ever
- 🧰 Single binary, portable

![full-preview!](/img/full-preview.png "full preview")

![two panes!](/img/two-panes.png "two panes")

## 📦 Installation

### Pre-compiled Binaries

Obtain the latest optimized binary from
the [Releases](https://github.com/crumbyte/noxdir/releases) page. The
application is self-contained and requires no installation process.

### Build from source (Go 1.24+)

```bash
git clone https://github.com/crumbyte/noxdir.git
cd noxdir
make build

./bin/noxdir
```

## 🛠 Usage

Just run in the terminal:

```bash
noxdir
```

The interactive interface initializes immediately without configuration
requirements.

## ⚙️ How It Works

It identifies all available partitions for Windows, or volumes in the case of
macOS and Linux. It'll immediately show the capacity info for all drives,
including file system type, total capacity, free space, and usage data. All
drives will be sorted (by default) by the free space left.

Press `Enter` to explore a particular drive and check what files or directories
occupy the most space. Wait while the scan is finished, and the status will
update in the status bar.
Now you have the full view of the files and directories, including the space
usage info by each entry. Use `ctrl+q`
to immediately see the biggest files on the drive, or `ctrl+e` to
see the biggest directories. Use `ctrl+f` to filter entries by their names or
`,` and `.` to show only files or directories.

Also, NoxDir accepts flags on a startup. Here's a list of currently available
CLI flags:

```
Usage:
  noxdir [flags]

Flags:
      --color-schema string   Set the color schema configuration file. The file contains a custom
                              color settings for the UI elements.

  -x, --exclude strings       Exclude specific directories from scanning. Useful for directories
                              with many subdirectories but minimal disk usage (e.g., node_modules).

                              NOTE: The check targets any string occurrence. The excluded directory
                              name can be either an absolute path or only part of it. In the last case,
                              all directories whose name contains that string will be excluded from
                              scanning.

                              Example: --exclude="node_modules,Steam\appcache"
                              (first rule will exclude all existing "node_modules" directories)
  -h, --help                  help for noxdir
  -d, --no-empty-dirs         Excludes all empty directories from the output. The directory is
                              considered empty if it or its subdirectories do not contain any files.

                              Even if the specific directory represents the entire tree structure of
                              subdirectories, without a single file, it will be completely skipped.

                              Default value is "false".

                              Example: --no-empty-dirs (provide a flag)

      --no-hidden             Excludes all hidden files and directories from the output. The entry is
                              considered hidden if its name starts with a dot, e.g., ".git".

                              Default value is "false".

                              Example: --no-hidden (provide a flag)

  -r, --root string           Start from a predefined root directory. Instead of selecting the target
                              drive and scanning all folders within, a root directory can be provided.
                              In this case, the scanning will be performed exclusively for the specified
                              directory, drastically reducing the scanning time.

                              Providing an invalid path results in a blank application output. In this
                              case, a "backspace" still can be used to return to the drives list.
                              Also, all trailing slash characters will be removed from the provided
                              path.

                              Example: --root="C:\Program Files (x86)"
  -l, --size-limit string     Define size limits/boundaries for files that should be shown in the
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

  -c, --use-cache             Force the application to cache the data. With cache enabled, the full
                              file system scan will be performed only once. After that, the cache will be
                              used as long as the flag is provided.

                              The cache will always store the last session data. In order to update the
                              cache and the application's state, use the "r" (refresh) command on a
                              target directory.

                              Default value is "false".

                              Example: -c|--use-cache (provide a flag)
```

## ⚠️ Known Issues

- The scan process on macOS might be slow sometimes. If it is an issue, consider
  using `--exclude` argument.
- In some cases, the volumes might duplicate on macOS and Linux. This issue will
  be fixed in the next releases.

## 🧩 Planned Features

- [ ] Real-time filesystem event monitoring and interface updates
- [ ] Exportable reports in various formats (JSON, CSV, HTML)
- [ ] Sort directories by usage, free space, etc. (already done for
  drives)

## 🙋 FAQ

- **Q:** Can I use this in scripts or headless environments?
- **A:** Not yet — it's designed for interactive use.
  <br><br>
- **Q:** What are the security implications of running NoxDir?
- **A:** NoxDir operates in a strictly read-only capacity, with no file
  modification capabilities except for deletion, which requires confirmation.
  <br><br>
- **Q:** The interface appears to have rendering issues with icons or
  formatting, and there are no multiple panes like in the screenshots.
- **A:** Visual presentation depends on terminal capabilities and font
  configuration. For optimal experience, a terminal with Unicode and glyph
  support is recommended. The screenshots were made in `WezTerm` using `MesloLGM Nerd Font` font. 

## 🧪 Contributing

Pull requests are welcome! If you’d like to add features or report bugs, please
open an issue first to discuss.

## 📝 License

MIT © [crumbyte](https://github.com/crumbyte)

---
