package render

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

var ansiStyleRegexp = regexp.MustCompile(`\x1b[[\d;]*m`)

func OverlayCenter(fullWidth int, background, overlay string, row, col int) string {
	wrappedBG := ansi.Hardwrap(background, fullWidth, true)

	backgroundRows := strings.Split(wrappedBG, "\n")
	overlayRows := strings.Split(overlay, "\n")

	for i, overlayRow := range overlayRows {
		if i+row >= len(backgroundRows) {
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
