package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sizeUnits = []string{
	"B", "KB", "MB", "GB", "TB", "PB", "EB",
}

var sizeUnitsSeverity = map[string]string{
	"GB": "#f48c06",
	"TB": "#dc2f02",
	"PB": "#9d0208",
	"EB": "#6a040f",
}

type numeric interface {
	int | uint | uint64 | int64 | int32 | float64 | float32
}

func FmtSize[T numeric](bytesSize T, width int) string {
	size, suffix := fmtSize(bytesSize)
	padding := len(suffix) + 1

	if width > 0 {
		padding = max(width-len(size), padding)
	}

	return fmt.Sprintf("%s%*s", size, padding, suffix)
}

func FmtSizeColor[T numeric](bytesSize T, width, fullWidth int) string {
	size, suffix := fmtSize(bytesSize)
	padding, severity := 1, sizeUnitsSeverity[suffix]

	if width > 0 {
		padding = max(width-len(size)-len(suffix), padding)
	}

	if len(severity) > 0 {
		suffix = lipgloss.NewStyle().Foreground(
			lipgloss.Color(severity),
		).Render(suffix + strings.Repeat(" ", fullWidth))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left, size, strings.Repeat(" ", padding), suffix,
	)
}

func fmtSize[T numeric](bytesSize T) (string, string) {
	size := float64(bytesSize)
	val := size

	suffix := sizeUnits[0]

	if bytesSize > 0 {
		e := math.Floor(math.Log(size) / math.Log(1024))
		suffix = sizeUnits[min(int(e), len(sizeUnits)-1)]

		val = math.Floor(size/math.Pow(1024, e)*10+0.5) / 10

		if int(e) > len(sizeUnits)-1 {
			val = 1024 * float64(int(e)-(len(sizeUnits)-1))
		}
	}

	return fmt.Sprintf("%.2f", val), suffix
}

func unitFmt(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func FmtName(name string, maxWidth int) string {
	nameWrap := lipgloss.NewStyle().MaxWidth(maxWidth - 5).Render(name)

	if lipgloss.Width(nameWrap) == maxWidth-5 {
		nameWrap += "..."
	}

	return nameWrap
}

func FmtUsage(usage, threshold float64, fullWidth int) string {
	minWidth := 8
	usagePercent := usage * 100

	s := lipgloss.NewStyle()

	if usagePercent > threshold {
		s = s.Foreground(lipgloss.Color("#dc2f02"))
	}

	usageFmt := s.Render(
		strconv.FormatFloat(usage*100, 'f', 2, 64) + " %" + strings.Repeat(" ", fullWidth),
	)

	return fmt.Sprintf(
		"%s%s",
		strings.Repeat(" ", max(0, minWidth-lipgloss.Width(usageFmt))),
		usageFmt,
	)
}
