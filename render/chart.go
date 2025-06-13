package render

import (
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	maxSectors      = 7
	chartLabelWidth = 50
	aspectFix       = 2.4
)

var chartColors = []lipgloss.Color{
	lipgloss.Color("#ffbe0b"),
	lipgloss.Color("#fb5607"),
	lipgloss.Color("#ff006e"),
	lipgloss.Color("#8338ec"),
	lipgloss.Color("#3a86ff"),
	lipgloss.Color("#00f5d4"),
	lipgloss.Color("#fef9ef"),
	lipgloss.Color("#ff85a1"),
	lipgloss.Color("#b5838d"),
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

func Chart(width, height, radius int, totalSize int64, raw []RawChartSector) string {
	sb := strings.Builder{}

	sectors := prepareSectors(totalSize, raw)

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
						lipgloss.NewStyle().Foreground(s.color).Render("ø"),
					)

					break
				}
			}
		}

		sb.WriteByte('\n')
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center, sb.String(), legend(sectors, width/2),
	)
}

func prepareSectors(totalSize int64, rawSectors []RawChartSector) []chartSector {
	sectors := make([]chartSector, 0, len(rawSectors))

	others := chartSector{label: "Others"}

	for i, s := range rawSectors {
		usage := float64(s.Size) / float64(totalSize)

		if i > maxSectors {
			others.size += s.Size

			continue
		}

		sectors = append(
			sectors,
			chartSector{
				label: fmtName(s.Label, chartLabelWidth),
				size:  s.Size,
				usage: usage,
			},
		)
	}

	if others.size > 0 {
		others.usage = float64(others.size) / float64(totalSize)
		sectors = append(sectors, others)
	}

	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].size > sectors[j].size
	})

	start := 0.0

	for i := range sectors {
		sectors[i].color = chartColors[i]
		sectors[i].startAngle = start
		sectors[i].endAngle = start + sectors[i].usage*2*math.Pi

		start = sectors[i].endAngle
	}

	return sectors
}

func legend(sectors []chartSector, width int) string {
	l := make([]string, 0, len(sectors))
	listPadding := 5

	for _, s := range sectors {
		label := fmtName(s.label, int(float64(width)*0.6))
		size := FmtSize(s.size, 0)

		padding := strings.Repeat(
			" ",
			max(width-lipgloss.Width(label)-listPadding*2-lipgloss.Width(size), 0),
		)

		row := lipgloss.NewStyle().
			Width(width).
			Foreground(s.color).
			Padding(0, listPadding).
			Render(label + padding + FmtSize(s.size, 0) + "\n")

		l = append(l, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, l...)
}
