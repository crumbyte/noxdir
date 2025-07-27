package render

import (
	"fmt"
	"math"
	"os"
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

// WrapString wraps the string up to the provided limit value. If the string
// reached the limit, it will be appended with the "..." suffix.
func WrapString(data string, limit int) string {
	// wrap the original name and reserve some characters
	wrappedData := lipgloss.NewStyle().MaxWidth(limit - 5).Render(data)

	if lipgloss.Width(wrappedData) == limit-5 {
		wrappedData += "..."
	}

	return wrappedData
}

func FmtUsage(usage, threshold float64, fullWidth int) string {
	// minWidth defines a width of longest possible usage string value 100.00 %.
	minWidth := 8
	usagePercent := max(usage, 0) * 100

	s := lipgloss.NewStyle()

	if usagePercent > threshold {
		s = s.Foreground(lipgloss.Color("#dc2f02"))
	}

	usageStr := strconv.FormatFloat(usagePercent, 'f', 2, 64)
	spacing := strings.Repeat(" ", minWidth-len(usageStr)-2)
	suffix := " %" + strings.Repeat(" ", max(fullWidth-minWidth, 0))

	return s.Render(usageStr + spacing + suffix)
}

func WrapPath(path string, limit int) string {
	if len(path) <= limit || limit < 0 {
		return path
	}

	prefix := "..."

	truncateLength := len(path) - limit

	pathSeparatorIdx := strings.IndexByte(
		path[truncateLength:], os.PathSeparator,
	)

	if pathSeparatorIdx == -1 {
		return prefix + path[truncateLength:]
	}

	return prefix + path[truncateLength:][pathSeparatorIdx:]
}
