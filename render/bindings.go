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
	remove            bindingKey = "!"
	sortTotalCap      bindingKey = "alt+t"
	sortTotalUsed     bindingKey = "alt+u"
	sortTotalFree     bindingKey = "alt+f"
	sortTotalUsedP    bindingKey = "alt+g"
	toggleTopFiles    bindingKey = "ctrl+q"
	toggleTopDirs     bindingKey = "ctrl+e"
	toggleDirsFilter  bindingKey = "."
	toggleFilesFilter bindingKey = ","
	toggleNameFilter  bindingKey = "ctrl+f"
	toggleChart       bindingKey = "ctrl+w"
	toggleHelp        bindingKey = "?"
	left              bindingKey = "left"
	right             bindingKey = "right"
	diff              bindingKey = "+"
	openConfig        bindingKey = "%"
)

func ToggleHelpBinding() key.Binding {
	return key.NewBinding(
		key.WithKeys(toggleHelp.String()),
		key.WithHelp(
			style.BindKey().Render(toggleHelp.String()),
			style.Help().Render(" - toggle full help"),
		),
	)
}

func NavigateKeyMap() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp(
					style.BindKey().Render("↑/k"),
					style.Help().Render(" - up"),
				),
			),
			key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp(
					style.BindKey().Render("↓/j"),
					style.Help().Render(" - down"),
				),
			),
			key.NewBinding(
				key.WithKeys("home", "g"),
				key.WithHelp(
					style.BindKey().Render("g/home"),
					style.Help().Render(" - go to start"),
				),
			),
			key.NewBinding(
				key.WithKeys("end", "G"),
				key.WithHelp(
					style.BindKey().Render("G/end"),
					style.Help().Render(" - go to end"),
				),
			),
		},
	}
}

func ShortHelp() []key.Binding {
	return append(NavigateKeyMap()[0], ToggleHelpBinding())
}

func DrivesKeyMap() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(
				key.WithKeys(
					sortTotalCap.String(),
					sortTotalUsed.String(),
					sortTotalFree.String(),
				),
				key.WithHelp(
					style.BindKey().Render("alt+(t/f/u/g)"),
					style.Help().Render(" - sort total/free/used/usage"),
				),
			),
			key.NewBinding(
				key.WithKeys(enter.String(), right.String()),
				key.WithHelp(
					style.BindKey().Render("→/"+enter.String()),
					style.Help().Render(" - open drive"),
				),
			),
			key.NewBinding(
				key.WithKeys(explore.String()),
				key.WithHelp(
					style.BindKey().Render(explore.String()),
					style.Help().Render(" - explore drive"),
				),
			),
			key.NewBinding(
				key.WithKeys(quit.String(), cancel.String()),
				key.WithHelp(
					style.BindKey().Render(quit.String()+"/"+cancel.String()),
					style.Help().Render(" - quit"),
				),
			),
		},
		{
			ToggleHelpBinding(),
			key.NewBinding(
				key.WithKeys(refresh.String()),
				key.WithHelp(
					style.BindKey().Render(refresh.String()),
					style.Help().Render(" - refresh"),
				),
			),
			key.NewBinding(
				key.WithKeys(openConfig.String()),
				key.WithHelp(
					style.BindKey().Render(openConfig.String()),
					style.Help().Render(" - open config"),
				),
			),
		},
	}
}

//nolint:funlen
func DirsKeyMap() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(
				key.WithKeys(enter.String()),
				key.WithHelp(
					style.BindKey().Render("→/"+enter.String()),
					style.Help().Render(" - open dir"),
				),
			),
			key.NewBinding(
				key.WithKeys(backspace.String(), left.String()),
				key.WithHelp(
					style.BindKey().Render("←/"+backspace.String()),
					style.Help().Render(" - back"),
				),
			),
			key.NewBinding(
				key.WithKeys(explore.String()),
				key.WithHelp(
					style.BindKey().Render(explore.String()),
					style.Help().Render(" - explore dir/file"),
				),
			),
			key.NewBinding(
				key.WithKeys(quit.String(), cancel.String()),
				key.WithHelp(
					style.BindKey().Render(quit.String()+"/"+cancel.String()),
					style.Help().Render(" - quit"),
				),
			),
		},
		{
			key.NewBinding(
				key.WithKeys(toggleTopFiles.String()),
				key.WithHelp(
					style.BindKey().Render(toggleTopFiles.String()),
					style.Help().Render(" - toggle top files"),
				),
			),
			key.NewBinding(
				key.WithKeys(toggleTopDirs.String()),
				key.WithHelp(
					style.BindKey().Render(toggleTopDirs.String()),
					style.Help().Render(" - toggle top dirs"),
				),
			),
			key.NewBinding(
				key.WithKeys(toggleNameFilter.String()),
				key.WithHelp(
					style.BindKey().Render(toggleNameFilter.String()),
					style.Help().Render(" - toggle name filter"),
				),
			),
			key.NewBinding(
				key.WithKeys(toggleChart.String()),
				key.WithHelp(
					style.BindKey().Render(toggleChart.String()),
					style.Help().Render(" - usage chart"),
				),
			),
		},
		{
			key.NewBinding(
				key.WithKeys(toggleDirsFilter.String()),
				key.WithHelp(
					style.BindKey().Render(toggleDirsFilter.String()),
					style.Help().Render(" - toggle dirs only"),
				),
			),
			key.NewBinding(
				key.WithKeys(toggleFilesFilter.String()),
				key.WithHelp(
					style.BindKey().Render(toggleFilesFilter.String()),
					style.Help().Render(" - toggle files only"),
				),
			),
			key.NewBinding(
				key.WithKeys(refresh.String()),
				key.WithHelp(
					style.BindKey().Render(refresh.String()),
					style.Help().Render(" - refresh"),
				),
			),
			key.NewBinding(
				key.WithKeys(remove.String()),
				key.WithHelp(
					style.BindKey().Render(remove.String()),
					style.Help().Render(" - delete"),
				),
			),
		},
		{
			ToggleHelpBinding(),
			key.NewBinding(
				key.WithKeys(diff.String()),
				key.WithHelp(
					style.BindKey().Render(diff.String()),
					style.Help().Render(" - toggle diff"),
				),
			),
			key.NewBinding(
				key.WithKeys(openConfig.String()),
				key.WithHelp(
					style.BindKey().Render(openConfig.String()),
					style.Help().Render(" - open config"),
				),
			),
			key.NewBinding(
				key.WithKeys("/"),
				key.WithHelp(
					style.BindKey().Render("/"),
					style.Help().Render(" - mark row"),
				),
			),
		},
	}
}
