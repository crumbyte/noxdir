package render

import (
	"fmt"
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

var sizeUnits = []string{
	"B", "KB", "MB", "GB", "TB", "PB", "EB",
}

type numeric interface {
	int | uint | uint64 | int64 | int32 | float64 | float32
}

func FmtSize[T numeric](bytesSize T, width int) string {
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

	sizeFmt := fmt.Sprintf("%.2f", val)
	padding := len(suffix) + 1

	if width > 0 {
		padding = max(width-len(sizeFmt), padding)
	}

	return fmt.Sprintf("%s%*s", sizeFmt, padding, suffix)
}

func unitFmt(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func fmtName(name string, maxWidth int) string {
	nameWrap := lipgloss.NewStyle().MaxWidth(maxWidth - 5).Render(name)

	if lipgloss.Width(nameWrap) == maxWidth-5 {
		nameWrap += "..."
	}

	return nameWrap
}
