# 🧹 NoxDir

[![Build](https://github.com/crumbyte/noxdir/actions/workflows/build.yml/badge.svg)](https://github.com/crumbyte/noxdir/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/crumbyte/noxdir)](https://goreportcard.com/report/github.com/crumbyte/noxdir)
![GitHub Release](https://img.shields.io/github/v/release/crumbyte/noxdir)

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

### 🍺 Homebrew

```bash
brew tap crumbyte/noxdir
brew install --cask noxdir
```

### Linux

```bash
curl -s https://crumbyte.github.io/noxdir/scripts/install.sh | bash
```

```bash
curl -s https://crumbyte.github.io/noxdir/scripts/install.sh | bash -s -- v0.3.0
```

### Pre-compiled Binaries

Obtain the latest optimized binary from
the [Releases](https://github.com/crumbyte/noxdir/releases) page. The
application is self-contained and requires no installation process.

### Go install (Go 1.24+)

```bash
go install github.com/crumbyte/noxdir@latest
```

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

## 🔍 Viewing Changes (Delta Mode)

The diff is calculated be performing a scan of the **current** directory and compares it with the cached version you
currently have. If the caching was not enabled for the previous session the diff will show nothing.

NoxDir can display file system changes since your last session. It highlights added or deleted files and directories, as
well as changes in disk space usage.

> Caching must be enabled (`--use-cache` or `-c`) for diffs to work.

To view changes in the current directory, press the `+` key (toggle diff). NoxDir will compare the current state of the
directory with its cached version and display the difference:

![diff!](/img/diff.png "diff")

The diff is calculated by scanning only the current directory and comparing it against the previously cached state. If
no cache exists from the previous session, no differences will be shown.

## 🔧 Configuration File

On first launch, the application automatically generates a simple configuration file. This file allows you to define
default behaviors without needing to pass flags every time.

The configuration file is created at:

* Windows: `%LOCALAPPDATA%\.noxdir\settings.json` (e.g., `C:\Users\{user}\AppData\Local\.noxdir\settings.json`)
* Linux/macOS: `~/.noxdir/settings.json`

The created configurations file already contains all available settings and has the following structure:

```json
{
  "exclude": null,
  "colorSchema": "",
  "noEmptyDirs": false,
  "noHidden": false,
  "simpleColor": false,
  "useCache": false
}
```

Values follow the same format and behavior as CLI flags. For example:

```json
{
  "exclude": "node_modules,Steam\\appcache",
  "colorSchema": "custom_schema.json",
  "noEmptyDirs": true,
  "noHidden": false,
  "simpleColor": true,
  "useCache": false
}
```
👉 If you cannot find the configuration file you can open it right from the application using `%` key binding.

## 🗂️ Caching for Faster Scanning

Scanning can take time, especially on volumes with many small files and directories (e.g., log folders or
`node_modules`). To improve performance in such cases, NoxDir supports caching.

When the `--use-cache (-c)` flag is provided, NoxDir will attempt to use an existing cache file for the selected drive
or volume. If no cache file exists, it performs a full scan and saves the result to a cache file for future use.

If a cache file is found, the full scan is skipped by default (unless you explicitly want to see the structure delta).
Scanning is then performed **on demand** using the `r` (refresh) key, which updates the cache after the session ends.

Cache file locations:

* Windows: `%LOCALAPPDATA%\.noxdir\cache` (e.g., `C:\Users\{user}\AppData\Local\.noxdir\cache`)
* Linux/macOS: `~/.noxdir/cache`

To clear all cached data, use the `--clear-cache` flag.

## 🎨 Colors Customization

NoxDir supports color schema customization via the `--color-schema` flag. You can provide a JSON configuration to adjust
colors, borders, glyph rendering, and more.

A full example schema (including all default settings) is
available [here](https://github.com/crumbyte/noxdir/blob/main/default-color-schema.json). You can also provide a partial
config that
overrides only specific values.

Example:

```json
{
  "statusBarBorder": false,
  "usageProgressBar": {
    "fullChar": "█",
    "emptyChar": "░"
  }
}
```

In this example, the status bar border is disabled, and the usage progress bar is rendered using ANSI characters (█, ░)
instead of emojis (🟥, 🟩).

## 🚩 Flags

NoxDir accepts flags on a startup. Here's a list of currently available
CLI flags:

```
Usage:
  noxdir [flags]

Flags:
      --clear-cache           Delete all cache files from the application's directory.

                              Example: --clear-cache (provide a flag)

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
      --simple-color          Use a simplified color schema without emojis and glyphs.

                              Example: --simple-color (provide a flag)

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

- **Q:** Why is caching not enabled by default?
- **A:** The caching flow might work relatively slow in some cases (on Darwin, the decompression happens really slow),
  but still much faster than regular scanning. This flow still needs to be polished. After that, the caching will become
  a default behavior.
  <br><br>
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
