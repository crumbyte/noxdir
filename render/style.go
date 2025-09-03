package render

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleOnce sync.Once
	style     *Style
)

type Style struct {
	cache map[string]*lipgloss.Style
	cs    ColorSchema
}

func InitStyle(cs ColorSchema) *Style {
	styleOnce.Do(func() {
		style = &Style{
			cs:    cs,
			cache: make(map[string]*lipgloss.Style),
		}
	})

	return style
}

func (s *Style) CS() *ColorSchema {
	return &s.cs
}

func (s *Style) TableHeader() *lipgloss.Style {
	cv, ok := s.cache["tableHeader"]
	if !ok {
		cs := lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(s.cs.TableHeaderBorder)).
			BorderBottom(true)

		s.cache["tableHeader"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) TopTableHeader() *lipgloss.Style {
	cv, ok := s.cache["topTable"]
	if !ok {
		cs := lipgloss.NewStyle().
			Inherit(*s.TableHeader()).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder())

		s.cache["topTable"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) SelectedRow() *lipgloss.Style {
	cv, ok := s.cache["selectedRow"]
	if !ok {
		cs := lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.cs.SelectedRowText)).
			Background(lipgloss.Color(s.cs.SelectedRowBG)).
			Bold(false)

		s.cache["selectedRow"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) MarkedRow() *lipgloss.Style {
	cv, ok := s.cache["markedRow"]
	if !ok {
		cs := lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.cs.MarkedRowText)).
			Background(lipgloss.Color(s.cs.MarkedRowBG)).
			Bold(false)

		s.cache["markedRow"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) StatusBar() *lipgloss.Style {
	cv, ok := s.cache["statusBar"]
	if !ok {
		cs := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: s.cs.StatusBar.Text}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: s.cs.StatusBar.BG})

		s.cache["statusBar"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) TopFiles() *lipgloss.Style {
	cv, ok := s.cache["topFiles"]
	if !ok {
		cs := lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.cs.TopFilesText)).
			Bold(true)

		s.cache["topFiles"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) Help() *lipgloss.Style {
	cv, ok := s.cache["help"]
	if !ok {
		cs := lipgloss.NewStyle().Foreground(lipgloss.Color(s.cs.HelpText))

		s.cache["help"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) BindKey() *lipgloss.Style {
	cv, ok := s.cache["bindKey"]
	if !ok {
		cs := lipgloss.NewStyle().Foreground(lipgloss.Color(s.cs.BindingText))

		s.cache["bindKey"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) DialogBox() *lipgloss.Style {
	cv, ok := s.cache["dialogBox"]
	if !ok {
		cs := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(s.cs.DialogBoxBorder))

		s.cache["dialogBox"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) ChartBox() *lipgloss.Style {
	cv, ok := s.cache["chartBox"]
	if !ok {
		cs := lipgloss.NewStyle().
			Inherit(*s.DialogBox()).
			BorderForeground(lipgloss.Color(s.cs.ChartColors.Border)).
			BorderBottom(false)

		s.cache["chartBox"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) ConfirmButton() *lipgloss.Style {
	cv, ok := s.cache["confirmButton"]
	if !ok {
		cs := lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.cs.ConfirmButtonText)).
			Background(lipgloss.Color(s.cs.ConfirmButtonBG)).
			Padding(0, 3).
			Margin(1, 3)

		s.cache["confirmButton"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) ActiveButton() *lipgloss.Style {
	cv, ok := s.cache["activeButton"]
	if !ok {
		cs := s.ConfirmButton().
			Foreground(lipgloss.Color(s.cs.ActiveButtonText)).
			Background(lipgloss.Color(s.cs.ActiveButtonBG)).
			Underline(true)

		s.cache["activeButton"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) BarBlock(bgColor lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.cs.StatusBar.BlockText)).
		Background(bgColor).
		Padding(0, 1)
}

func (s *Style) ChartColors() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color(s.cs.ChartColors.Sector1),
		lipgloss.Color(s.cs.ChartColors.Sector2),
		lipgloss.Color(s.cs.ChartColors.Sector3),
		lipgloss.Color(s.cs.ChartColors.Sector4),
		lipgloss.Color(s.cs.ChartColors.Sector5),
		lipgloss.Color(s.cs.ChartColors.Sector6),
		lipgloss.Color(s.cs.ChartColors.Sector7),
		lipgloss.Color(s.cs.ChartColors.Sector8),
		lipgloss.Color(s.cs.ChartColors.Sector9),
	}
}

func (s *Style) SizeUnit(unit string) *lipgloss.Style {
	ck := fmt.Sprintf("sizeUnit-%s", unit)
	cv, ok := s.cache[ck]
	if !ok {
		sizeUnitColorsMap := map[string]string{
			"GB": s.cs.SizeUnit.GB,
			"TB": s.cs.SizeUnit.TB,
			"PB": s.cs.SizeUnit.PB,
			"EB": s.cs.SizeUnit.EB,
		}
		color := sizeUnitColorsMap[unit]
		cs := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		s.cache[ck] = &cs

		return &cs
	}

	return cv
}

func (s *Style) CmdInputText() *lipgloss.Style {
	cv, ok := s.cache["commandInputText"]
	if !ok {
		cs := lipgloss.NewStyle().Foreground(lipgloss.Color(s.cs.CmdInputText))

		s.cache["commandInputText"] = &cs

		return &cs
	}

	return cv
}

func (s *Style) CmdBarBorder() *lipgloss.Style {
	cv, ok := s.cache["commandBarBorder"]
	if !ok {
		cs := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(s.cs.CmdBarBorder)).
			BorderTop(true)

		s.cache["commandBarBorder"] = &cs

		return &cs
	}

	return cv
}
