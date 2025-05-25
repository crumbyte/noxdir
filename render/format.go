package render

import (
	"fmt"
	"strconv"
)

type numeric interface {
	int | uint | uint64 | int64 | int32 | float64 | float32
}

func fmtSize[T numeric](bytesSize T, fmtWidth bool) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	totalLength, unitIdx := 9, 0
	size := float64(bytesSize)

	for ; size > 1024; unitIdx++ {
		size /= 1024
	}

	sizeFmt := fmt.Sprintf("%.2f", size)

	if !fmtWidth {
		totalLength = len(sizeFmt)
	}

	return fmt.Sprintf(
		"%s %*s", sizeFmt, totalLength-len(sizeFmt), units[unitIdx],
	)
}

func unitFmt(val uint64) string {
	return strconv.FormatUint(val, 10)
}
