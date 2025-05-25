package render

import "github.com/charmbracelet/lipgloss"

var (
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#262626")).
				Background(lipgloss.Color("#ebbd34")).
				Bold(false)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	topFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ebbd34")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#696868"))

	bindKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffbf69"))
)
