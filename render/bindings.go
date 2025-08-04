package render

import (
	"strings"
	"sync"

	"github.com/crumbyte/noxdir/config"
	"github.com/crumbyte/noxdir/render/table"

	"github.com/charmbracelet/bubbles/key"
)

var (
	Bindings   KeyMap
	kmSyncOnce sync.Once
)

type DriveKeyMap struct {
	LevelDown key.Binding
	SortKeys  key.Binding
}

type DirsKeyMap struct {
	LevelUp    key.Binding
	LevelDown  key.Binding
	Delete     key.Binding
	TopFiles   key.Binding
	TopDirs    key.Binding
	FilesOnly  key.Binding
	DirsOnly   key.Binding
	NameFilter key.Binding
	Chart      key.Binding
	Diff       key.Binding
	Command    key.Binding
}

type KeyMap struct {
	NavigationKeyMap table.KeyMap
	Drive            DriveKeyMap
	Dirs             DirsKeyMap
	Explore          key.Binding
	Quit             key.Binding
	Refresh          key.Binding
	Help             key.Binding
	Config           key.Binding
	style            *Style
}

func (km *KeyMap) NavigateBindings() [][]key.Binding {
	return [][]key.Binding{
		{
			km.NavigationKeyMap.LineUp,
			km.NavigationKeyMap.LineDown,
			km.NavigationKeyMap.GotoTop,
			km.NavigationKeyMap.GotoBottom,
		},
	}
}

func (km *KeyMap) ShortBindings() []key.Binding {
	return append(km.NavigateBindings()[0], km.Help)
}

func (km *KeyMap) DriveBindings() [][]key.Binding {
	return append(
		km.NavigateBindings(),
		[][]key.Binding{
			{km.Drive.SortKeys, km.Drive.LevelDown, km.Explore, km.Quit},
			{km.Refresh, km.Config},
		}...,
	)
}

func (km *KeyMap) DirBindings() [][]key.Binding {
	return append(
		km.NavigateBindings(),
		[][]key.Binding{
			{km.Dirs.LevelDown, km.Dirs.LevelUp, km.Explore, km.Quit},
			{km.Dirs.TopFiles, km.Dirs.TopDirs, km.Dirs.NameFilter, km.Dirs.Chart},
			{km.Dirs.DirsOnly, km.Dirs.FilesOnly, km.Refresh, km.Dirs.Delete},
			{km.Dirs.Diff, km.Config, km.Refresh, km.Dirs.Delete},
			{km.Dirs.Command},
		}...,
	)
}

//nolint:funlen
func DefaultKeyMap(s *Style) KeyMap {
	return KeyMap{
		style: s,
		NavigationKeyMap: table.KeyMap{
			LineUp: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp(
					s.BindKey().Render("↑/k"),
					s.Help().Render("up"),
				),
			),
			LineDown: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp(
					s.BindKey().Render("↓/j"),
					s.Help().Render("down"),
				),
			),
			GotoTop: key.NewBinding(
				key.WithKeys("home", "g"),
				key.WithHelp(
					s.BindKey().Render("g/home"),
					s.Help().Render("go to start"),
				),
			),
			GotoBottom: key.NewBinding(
				key.WithKeys("end", "G"),
				key.WithHelp(
					s.BindKey().Render("G/end"),
					s.Help().Render("go to end"),
				),
			),
		},
		Drive: DriveKeyMap{
			LevelDown: key.NewBinding(
				key.WithKeys("enter", "right"),
				key.WithHelp(
					s.BindKey().Render("→/enter"),
					s.Help().Render(" - open drive"),
				),
			),
			SortKeys: key.NewBinding(
				key.WithKeys("alt+t", "alt+u", "alt+f", "alt+g"),
				key.WithHelp(
					s.BindKey().Render("alt+(t/f/u/g)"),
					s.Help().Render(" - sort total/free/used/usage"),
				),
			),
		},
		Dirs: DirsKeyMap{
			LevelUp: key.NewBinding(
				key.WithKeys("backspace", "left"),
				key.WithHelp(
					s.BindKey().Render("←/backspace"),
					s.Help().Render(" - back"),
				),
			),
			LevelDown: key.NewBinding(
				key.WithKeys("enter", "right"),
				key.WithHelp(
					s.BindKey().Render("→/enter"),
					s.Help().Render(" - open dir"),
				),
			),
			Delete: key.NewBinding(
				key.WithKeys("!"),
				key.WithHelp(
					s.BindKey().Render("!"),
					s.Help().Render(" - delete"),
				),
			),
			TopFiles: key.NewBinding(
				key.WithKeys("ctrl+q"),
				key.WithHelp(
					s.BindKey().Render("ctrl+q"),
					s.Help().Render(" - toggle top files"),
				),
			),
			TopDirs: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp(
					s.BindKey().Render("ctrl+e"),
					s.Help().Render(" - toggle top dirs"),
				),
			),
			FilesOnly: key.NewBinding(
				key.WithKeys(","),
				key.WithHelp(
					s.BindKey().Render(","),
					s.Help().Render(" - toggle files only"),
				),
			),
			DirsOnly: key.NewBinding(
				key.WithKeys("."),
				key.WithHelp(
					s.BindKey().Render("."),
					s.Help().Render(" - toggle dirs only"),
				),
			),
			NameFilter: key.NewBinding(
				key.WithKeys("ctrl+f"),
				key.WithHelp(
					s.BindKey().Render("ctrl+f"),
					s.Help().Render(" - toggle name filter"),
				),
			),
			Chart: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp(
					s.BindKey().Render("ctrl+w"),
					s.Help().Render(" - usage chart"),
				),
			),
			Diff: key.NewBinding(
				key.WithKeys("+"),
				key.WithHelp(
					s.BindKey().Render("+"),
					s.Help().Render(" - toggle diff"),
				),
			),
			Command: key.NewBinding(
				key.WithKeys(":"),
				key.WithHelp(
					s.BindKey().Render(":"),
					s.Help().Render(" - command"),
				),
			),
		},
		Explore: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp(
				s.BindKey().Render("e"),
				s.Help().Render(" - explore"),
			),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp(
				s.BindKey().Render("q/ctrl+c"),
				s.Help().Render(" - quit"),
			),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp(
				s.BindKey().Render("r"),
				s.Help().Render(" - refresh"),
			),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp(
				s.BindKey().Render("?"),
				s.Help().Render(" - toggle full help"),
			),
		),
		Config: key.NewBinding(
			key.WithKeys("%"),
			key.WithHelp(
				s.BindKey().Render("%"),
				s.Help().Render(" - open config"),
			),
		),
	}
}

func InitKeyMap(b *config.Bindings, s *Style) {
	kmSyncOnce.Do(func() {
		Bindings = DefaultKeyMap(s)

		if b == nil {
			return
		}

		Bindings.Config = Bindings.override(Bindings.Config, b.Config)
		Bindings.Refresh = Bindings.override(Bindings.Refresh, b.Refresh)
		Bindings.Quit = Bindings.override(Bindings.Quit, b.Quit)
		Bindings.Help = Bindings.override(Bindings.Help, b.Help)
		Bindings.Explore = Bindings.override(Bindings.Explore, b.Explore)

		Bindings.Drive.LevelDown = Bindings.override(
			Bindings.Drive.LevelDown, b.DriveBindings.LevelDown,
		)
		Bindings.Drive.SortKeys = Bindings.override(
			Bindings.Drive.SortKeys, b.DriveBindings.SortKeys,
		)

		Bindings.Dirs.DirsOnly = Bindings.override(
			Bindings.Dirs.DirsOnly, b.DirBindings.DirsOnly,
		)
		Bindings.Dirs.FilesOnly = Bindings.override(
			Bindings.Dirs.FilesOnly, b.DirBindings.FilesOnly,
		)
		Bindings.Dirs.TopDirs = Bindings.override(
			Bindings.Dirs.TopDirs, b.DirBindings.TopDirs,
		)
		Bindings.Dirs.TopFiles = Bindings.override(
			Bindings.Dirs.TopFiles, b.DirBindings.TopFiles,
		)
		Bindings.Dirs.Delete = Bindings.override(
			Bindings.Dirs.Delete, b.DirBindings.Delete,
		)
		Bindings.Dirs.LevelDown = Bindings.override(
			Bindings.Dirs.LevelDown, b.DirBindings.LevelDown,
		)
		Bindings.Dirs.LevelUp = Bindings.override(
			Bindings.Dirs.LevelUp, b.DirBindings.LevelUp,
		)
		Bindings.Dirs.NameFilter = Bindings.override(
			Bindings.Dirs.NameFilter, b.DirBindings.NameFilter,
		)
		Bindings.Dirs.Diff = Bindings.override(
			Bindings.Dirs.Diff, b.DirBindings.Diff,
		)
		Bindings.Dirs.Chart = Bindings.override(
			Bindings.Dirs.Chart, b.DirBindings.Chart,
		)
	})
}

func (km *KeyMap) override(origin key.Binding, newSettings []string) key.Binding {
	if len(newSettings) == 0 {
		return origin
	}

	origin.SetKeys(newSettings...)

	origin.SetHelp(
		km.style.BindKey().Render(strings.Join(newSettings, "/")),
		origin.Help().Desc,
	)

	return origin
}
