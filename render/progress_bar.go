package render

import (
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/muesli/termenv"
)

type PGOption func(pg *PG)

type PG struct {
	StartColor   string          `json:"startColor"`
	EndColor     string          `json:"endColor"`
	FullChar     string          `json:"fullChar"`
	EmptyChar    string          `json:"emptyChar"`
	ColorProfile termenv.Profile `json:"colorProfile"`
	HidePercent  bool            `json:"hidePercent"`
}

func (pg *PG) New(width int) progress.Model {
	maxCharLen := max(
		lipgloss.Width(pg.FullChar),
		lipgloss.Width(pg.EmptyChar),
	)

	if maxCharLen < 1 {
		maxCharLen = 1
	}

	opts := []progress.Option{
		progress.WithWidth(width / maxCharLen),
		progress.WithColorProfile(pg.ColorProfile),
		progress.WithGradient(pg.StartColor, pg.EndColor),
	}

	if pg.HidePercent {
		opts = append(opts, progress.WithoutPercentage())
	}

	if len(pg.FullChar) != 0 && len(pg.EmptyChar) != 0 {
		fullChar, _ := utf8.DecodeRuneInString(pg.FullChar)
		emptyChar, _ := utf8.DecodeRuneInString(pg.EmptyChar)

		opts = append(
			opts,
			progress.WithFillCharacters(fullChar, emptyChar),
		)
	}

	return progress.New(opts...)
}
