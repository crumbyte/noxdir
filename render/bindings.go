package render

import (
	"github.com/charmbracelet/bubbles/key"
)

// bindingKey defines a custom type representing a keyboard's pressed key.
type bindingKey string

func (bk bindingKey) String() string {
	return string(bk)
}

const (
	backspace         bindingKey = "backspace"
	quit              bindingKey = "q"
	cancel            bindingKey = "ctrl+c"
	enter             bindingKey = "enter"
	explore           bindingKey = "e"
	refresh           bindingKey = "r"
	sortTotalCap      bindingKey = "alt+t"
	sortTotalUsed     bindingKey = "alt+u"
	sortTotalFree     bindingKey = "alt+f"
	sortTotalUsedP    bindingKey = "alt+g"
	toggleTopFiles    bindingKey = "ctrl+q"
	toggleTopDirs     bindingKey = "ctrl+e"
	toggleDirsFilter  bindingKey = "."
	toggleFilesFilter bindingKey = ","
	toggleNameFilter  bindingKey = "ctrl+f"
	toggleHelp        bindingKey = "?"
)

var toggleHelpBinding = key.NewBinding(
	key.WithKeys(toggleHelp.String()),
	key.WithHelp(
		bindKeyStyle.Render(toggleHelp.String()),
		helpDescStyle.Render(" - toggle full help"),
	),
)

var navigateKeyMap = [][]key.Binding{
	{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp(
				bindKeyStyle.Render("↑/k"),
				helpDescStyle.Render(" - up"),
			),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp(
				bindKeyStyle.Render("↓/j"),
				helpDescStyle.Render(" - down"),
			),
		),
		key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp(
				bindKeyStyle.Render("g/home"),
				helpDescStyle.Render(" - go to start"),
			),
		),
		key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp(
				bindKeyStyle.Render("G/end"),
				helpDescStyle.Render(" - go to end"),
			),
		),
	},
}

var shortHelp = append(navigateKeyMap[0], toggleHelpBinding)

var drivesKeyMap = [][]key.Binding{
	{
		key.NewBinding(
			key.WithKeys(
				sortTotalCap.String(),
				sortTotalUsed.String(),
				sortTotalFree.String(),
			),
			key.WithHelp(
				bindKeyStyle.Render("alt+(t/f/u/g)"),
				helpDescStyle.Render(" - sort total/free/used/usage"),
			),
		),
		key.NewBinding(
			key.WithKeys(enter.String()),
			key.WithHelp(
				bindKeyStyle.Render(enter.String()),
				helpDescStyle.Render(" - open drive"),
			),
		),
		key.NewBinding(
			key.WithKeys(explore.String()),
			key.WithHelp(
				bindKeyStyle.Render(explore.String()),
				helpDescStyle.Render(" - explore drive"),
			),
		),
		key.NewBinding(
			key.WithKeys(quit.String(), cancel.String()),
			key.WithHelp(
				bindKeyStyle.Render(quit.String()+"/"+cancel.String()),
				helpDescStyle.Render(" - quit"),
			),
		),
	},
	{
		toggleHelpBinding,
		key.NewBinding(
			key.WithKeys(refresh.String()),
			key.WithHelp(
				bindKeyStyle.Render(refresh.String()),
				helpDescStyle.Render(" - refresh"),
			),
		),
	},
}

var dirsKeyMap = [][]key.Binding{
	{
		key.NewBinding(
			key.WithKeys(enter.String()),
			key.WithHelp(
				bindKeyStyle.Render(enter.String()),
				helpDescStyle.Render(" - open dir"),
			),
		),
		key.NewBinding(
			key.WithKeys(backspace.String()),
			key.WithHelp(
				bindKeyStyle.Render(backspace.String()),
				helpDescStyle.Render(" - back"),
			),
		),
		key.NewBinding(
			key.WithKeys(explore.String()),
			key.WithHelp(
				bindKeyStyle.Render(explore.String()),
				helpDescStyle.Render(" - explore dir/file"),
			),
		),
		key.NewBinding(
			key.WithKeys(quit.String(), cancel.String()),
			key.WithHelp(
				bindKeyStyle.Render(quit.String()+"/"+cancel.String()),
				helpDescStyle.Render(" - quit"),
			),
		),
	},
	{
		key.NewBinding(
			key.WithKeys(toggleTopFiles.String()),
			key.WithHelp(
				bindKeyStyle.Render(toggleTopFiles.String()),
				helpDescStyle.Render(" - toggle top files"),
			),
		),
		key.NewBinding(
			key.WithKeys(toggleTopDirs.String()),
			key.WithHelp(
				bindKeyStyle.Render(toggleTopDirs.String()),
				helpDescStyle.Render(" - toggle top dirs"),
			),
		),
		key.NewBinding(
			key.WithKeys(toggleNameFilter.String()),
			key.WithHelp(
				bindKeyStyle.Render(toggleNameFilter.String()),
				helpDescStyle.Render(" - toggle name filter"),
			),
		),
		toggleHelpBinding,
	},
	{
		key.NewBinding(
			key.WithKeys(toggleDirsFilter.String()),
			key.WithHelp(
				bindKeyStyle.Render(toggleDirsFilter.String()),
				helpDescStyle.Render(" - toggle dirs only"),
			),
		),
		key.NewBinding(
			key.WithKeys(toggleFilesFilter.String()),
			key.WithHelp(
				bindKeyStyle.Render(toggleFilesFilter.String()),
				helpDescStyle.Render(" - toggle files only"),
			),
		),
		key.NewBinding(
			key.WithKeys(refresh.String()),
			key.WithHelp(
				bindKeyStyle.Render(refresh.String()),
				helpDescStyle.Render(" - refresh"),
			),
		),
	},
}
