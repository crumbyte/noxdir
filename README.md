# 🧹 NoxDir

[![Build](https://github.com/crumbyte/noxdir/actions/workflows/build.yml/badge.svg)](https://github.com/crumbyte/noxdir/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/crumbyte/noxdir)](https://goreportcard.com/report/github.com/crumbyte/noxdir)
![GitHub Release](https://img.shields.io/github/v/release/crumbyte/noxdir)

**NoxDir** is a high-performance, cross-platform command-line tool for
visualizing and exploring your file system usage. It detects mounted drives or
volumes and presents disk usage metrics through a responsive, keyboard-driven
terminal UI. Designed to help you quickly locate space hogs and streamline your
cleanup workflow. Supports: **Windows**, **macOS**, and **Linux**.

[//]: # (![full-preview!]&#40;/img/full-preview.png "full preview"&#41;)

[//]: # ()
[//]: # (![two panes!]&#40;/img/two-panes.png "two panes"&#41;)

![full-preview!](/img/preview.png "full preview")
![full-preview!](/img/linux-preview.png "full preview")
![full-preview!](/img/win-cmd.png "full preview")

## 📦 Installation

### 🍺 Homebrew

Stable release:

```bash
brew tap crumbyte/noxdir
brew install --cask noxdir
```

Nightly release:

```bash
brew tap crumbyte/noxdir
brew uninstall --cask noxdir # If the stable version was installed previously
brew install --cask noxdir-nightly
```

### Arch based Linux distros

```bash
pacman -S noxdir
```

### NixOS / Nix

NoxDir can be installed either directly from nixpkgs or used via flake:

**From nixpkgs (unstable channel):**

```bash
nix-env -iA nixpkgs.noxdir
```

Or in your `configuration.nix`:

```nix
environment.systemPackages = with pkgs; [
  noxdir
];
```

**Using flake:**

```bash
nix run github:crumbyte/noxdir
```

Or add to your `flake.nix`:

```nix
{
  inputs.noxdir.url = "github:crumbyte/noxdir";
}
```

### Other Linux distros

```bash
curl -s https://crumbyte.github.io/noxdir/scripts/install.sh | bash
```

```bash
curl -s https://crumbyte.github.io/noxdir/scripts/install.sh | bash -s -- v0.6.0
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

  -v, --version              Print the application version and exit.
```

## 🔧 Configuration File

On first launch, the application automatically generates a simple configuration file. This file allows you to define
default behaviors without needing to pass flags every time.

The configuration file is created at:

* Windows: `%LOCALAPPDATA%\.noxdir\settings.json` (e.g., `C:\Users\{user}\AppData\Local\.noxdir\settings.json`)
* Linux/macOS: `~/.noxdir/settings.json`

The created configurations file already contains all available settings and has the following structure:

```json
{
  "colorSchema": "",
  "exclude": null,
  "noEmptyDirs": false,
  "noHidden": false,
  "simpleColor": false,
  "useCache": false,
  "bindings": {
    "driveBindings": {
      "levelDown": []
    },
    "dirBindings": {
      "levelUp": null,
      "levelDown": null,
      "delete": null,
      "topFiles": null,
      "topDirs": null,
      "filesOnly": null,
      "dirsOnly": null,
      "nameFilter": null,
      "chart": null,
      "diff": null,
      "toggleSelectAll": null,
      "toggleSelection": null
    },
    "explore": null,
    "quit": null,
    "refresh": null,
    "help": null,
    "diff": null,
    "config": null
  }
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

## 🔍 Viewing Changes (Delta Mode)

NoxDir can display file system changes since your last session. It highlights added or deleted files and directories, as
well as changes in disk space usage. The diff is calculated by comparing the current directory state against its cached version. If no cache exists from the
previous session, no differences will be shown.

To view changes in the current directory, press the `+` key (toggle diff). NoxDir will compare the current state of the
directory with its cached version and display the difference:

![diff!](/img/diff.png "diff")

## ⌨️ Key Bindings

NoxDir provides full support for custom key bindings, allowing users to override nearly all interactive controls.
Bindings are defined in the [configuration file](#-configuration-file). By default, all key binding fields are set to
`null`. When a field is `null` or omitted, the default binding is used.

Default bindings are defined as follows:

```json
{
  "driveBindings": {
    "levelDown":    ["enter", "right"],
  },
  "dirBindings": {
    "levelUp":    ["backspace", "left"],
    "levelDown":  ["enter", "right"],
    "delete":     ["!"],
    "topFiles":   ["ctrl+q"],
    "topDirs":    ["ctrl+e"],
    "filesOnly":  [","],
    "dirsOnly":   ["."],
    "nameFilter": ["ctrl+f"],
    "chart":      ["ctrl+w"],
    "diff":       ["+"]
  },
  "explore": ["e"],
  "quit":    ["q", "ctrl+c"],
  "refresh": ["r"],
  "help":    ["?"],
  "config":  ["%"]
}
```

Each entry maps an action name to one or more key sequences. Bindings support modifiers such as `ctrl`, `alt`, and
`shift`, and are case-sensitive.

**Notes**

- Multiple bindings per action are supported.
- If a binding is explicitly set to `null`, the default will be used.
- Removing a binding entry is equivalent to setting it to `null`.
- Bindings must be declared as arrays of strings, even for single-key bindings.

Custom config example:

```json
{
  "dirBindings": {
    "topFiles": ["t"],
    "topDirs": ["T"]
  }
}
```

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
  support is recommended. The screenshots were made in `WezTerm` using `MesloLGM Nerd Font` font. If your font does not support glyphs consider using `--siimple-color` flag.
  <br><br>
- **Q:** The scanning process is too slow.
- **A:** Consider using caching, exclusion, or running the application only for specific directories. The caching can be enabled with the flag `-c, --use-cache` or in the configuration file. With caching enabled, you choose which directories must be re-scanned with the `r` key. Exclusion flag `-x, --exclude` allows providing a list of directories that must be skipped during scanning, e.g., `.node_modules`. Also, predefined root `-r, --root` will start the application from the specified directory instead of scanning the entire file system.

<p style="text-align: center;">
MIT © <a href="https://github.com/crumbyte">crumbyte</a>
</p>

