package render

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	chartLabelWidth = 50
	aspectFix       = 2.4
)

var chartColors = []lipgloss.Color{
	lipgloss.Color("#ffbe0b"),
	lipgloss.Color("#fb5607"),
	lipgloss.Color("#ff006e"),
	lipgloss.Color("#8338ec"),
	lipgloss.Color("#3a86ff"),
	lipgloss.Color("#fcf6bd"),
	lipgloss.Color("#d0f4de"),
	lipgloss.Color("#a9def9"),
	lipgloss.Color("#e4c1f9"),
	lipgloss.Color("#ffbe0b"),
	lipgloss.Color("#fb5607"),
	lipgloss.Color("#ff006e"),
	lipgloss.Color("#8338ec"),
	lipgloss.Color("#3a86ff"),
	lipgloss.Color("#fcf6bd"),
	lipgloss.Color("#d0f4de"),
	lipgloss.Color("#a9def9"),
	lipgloss.Color("#e4c1f9"),
}

type RawChartSector struct {
	Label string
	Size  int64
}

type chartSector struct {
	color      lipgloss.Color
	label      string
	size       int64
	usage      float64
	startAngle float64
	endAngle   float64
}

func chart(width, height, radius int, totalSize int64, rawSectors []RawChartSector) string {
	sb := strings.Builder{}

	sectors := prepareSectors(totalSize, rawSectors)

	centerX, centerY := width/2/2, height/2

	for y := range height {
		for x := range width / 2 {
			dx := float64(x - centerX)
			dy := float64(y-centerY) * aspectFix

			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > float64(radius) {
				sb.WriteByte(' ')

				continue
			}

			angle := math.Atan2(dy, dx)
			if angle < 0 {
				angle += 2 * math.Pi
			}

			for _, s := range sectors {
				if angle >= s.startAngle && angle < s.endAngle {
					sb.WriteString(
						lipgloss.NewStyle().Foreground(s.color).Render("#"),
					)

					break
				}
			}
		}

		sb.WriteByte('\n')
	}

	ch := sb.String()

	legend := make([]string, 0, len(sectors))

	for _, s := range sectors {
		legend = append(legend, lipgloss.NewStyle().MaxWidth(width).Foreground(s.color).PaddingLeft(10).Render(
			fmt.Sprintf("%s: %s \n", s.label, fmtSize(s.size, false)),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, ch, lipgloss.JoinVertical(lipgloss.Left, legend...))
}

func prepareSectors(totalSize int64, rawSectors []RawChartSector) []chartSector {
	sectors := make([]chartSector, 0, len(rawSectors))

	others := chartSector{label: "Others"}

	for _, s := range rawSectors {
		if float64(s.Size)/float64(totalSize) < 0.04 {
			others.size += s.Size

			continue
		}

		sectors = append(
			sectors,
			chartSector{
				label: fmtName(s.Label, chartLabelWidth),
				size:  s.Size,
			},
		)
	}

	if others.size > 0 {
		sectors = append(sectors, others)
	}

	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].size > sectors[j].size
	})

	start := 0.0

	for i := range sectors {
		sectors[i].color = chartColors[i]
		sectors[i].usage = float64(sectors[i].size) / float64(totalSize)
		sectors[i].startAngle = start
		sectors[i].endAngle = start + sectors[i].usage*2*math.Pi

		start = sectors[i].endAngle
	}

	return sectors
}
