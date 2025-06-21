package render

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var ansiStyleRegexp = regexp.MustCompile(`\x1b[[\d;]*m`)

func OverlayCenter(fullWidth, fullHeight int, background, overlay string) string {
	row := fullHeight/2 - (lipgloss.Height(overlay) / 2)
	col := fullWidth/2 - (lipgloss.Width(overlay) / 2)

	return Overlay(fullWidth, background, overlay, row, col)
}

// Overlay renders the provided overlay string over the background value. The
// overlay will be rendered starting at the specified row and column. All
// characters from the background will be properly wrapped accordingly without
// breaking the ANSI sequences.
func Overlay(_ int, background, overlay string, row, col int) string {
	// wrappedBG := ansi.Hardwrap(background, fullWidth, true)

	backgroundRows := strings.Split(background, "\n")
	overlayRows := strings.Split(overlay, "\n")

	// replace the background rows with the overlay rows
	for i, overlayRow := range overlayRows {
		// proceed only if there's enough space to render the overlay
		if i+row >= len(backgroundRows) || i+row < 0 {
			break
		}

		bgRow := backgroundRows[i+row]
		if len(bgRow) < col {
			bgRow += strings.Repeat(" ", col-len(bgRow))
		}

		bgLeft := ansi.Truncate(bgRow, col, "")
		bgRight := truncateLeft(bgRow, col+ansi.StringWidth(overlayRow))

		backgroundRows[i+row] = bgLeft + overlayRow + bgRight
	}

	return strings.Join(backgroundRows, "\n")
}

func truncateLeft(line string, padding int) string {
	if strings.Contains(line, "\n") {
		// TODO: get rid of this panic
		panic("line must not contain newline")
	}

	wrapped := strings.Split(ansi.Hardwrap(line, padding, true), "\n")
	if len(wrapped) == 1 {
		return ""
	}

	var ansiStyle string

	ansiStyles := ansiStyleRegexp.FindAllString(wrapped[0], -1)
	if l := len(ansiStyles); l > 0 {
		ansiStyle = ansiStyles[l-1]
	}

	return ansiStyle + strings.Join(wrapped[1:], "")
}
