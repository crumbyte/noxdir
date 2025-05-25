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
- 🧰 Minimal dependencies — single binary, portable

## 📸 Preview

Drives list             |  Directories list
:-------------------------:|:-------------------------:
![The San Juan Mountains are beautiful!](/img/drives.png "drives list")  |  ![The San Juan Mountains are beautiful!](/img/dirs.png "directories list")

## 📦 Installation

### Pre-compiled Binaries

Obtain the latest optimized binary from
the [Releases](https://github.com/crumbyte/noxdir/releases) repository. The
application is self-contained and requires no installation process.

### Or build from source (Go 1.24+)

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

The interactive interface initializes immediately without configuration requirements.

## ⚙️ How It Works

- **Windows:** Uses `GetLogicalDrives` and `GetDiskFreeSpaceExW` through direct
  syscalls for optimal performance.
- **Linux/macOS:** Uses `statfs` and parses `/proc/mounts` or `mount` command
  output to find mounted volumes.

## 🧩 Planned Features

- [ ] Real-time filesystem event monitoring and interface updates
- [ ] Exportable reports in various formats (JSON, CSV, HTML)
- [ ] Comprehensive file management capabilities (deletion, renaming, creation operations)
- [ ] Sort directories by usage, free space, etc. (already done for
  drives)
- [ ] Customizable interface aesthetics with theme support

## 🙋 FAQ

- **Q:** Can I use this in scripts or headless environments?
- **A:** Not yet — it's designed for interactive use.
  <br><br>
- **Q:** What are the security implications of running NoxDir?
- **A:** NoxDir operates in a strictly read-only capacity with no file
  modification capabilities in the current release.
  <br><br>
- Q: Does NoxDir support file management operations?
- A: File manipulation features are currently under development and prioritized
  in our roadmap.
  <br><br>
- **Q:** The interface appears to have rendering issues with icons or
  formatting.
- **A:** Visual presentation depends on terminal capabilities and font
  configuration. For optimal experience, a terminal with Unicode and glyph
  support is recommended.
  <br><br>

## 🧪 Contributing

Pull requests are welcome! If you’d like to add features or report bugs, please
open an issue first to discuss.

## 📝 License

MIT © [crumbyte](https://github.com/crumbyte)

---
