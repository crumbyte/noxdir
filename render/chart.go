package render

import (
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// maxSectors defines the maximum number of sectors on the chart. All sectors
	// that exceed this limit will be merged into a single sector.
	maxSectors = 8

	// chartLabelWidth limits the width of the sector's label.
	chartLabelWidth = 50
)

var defaultChartColors = []lipgloss.Color{
	"#ffbe0b", "#fb5607", "#ff006e", "#8338ec", "#3a86ff",
	"#00f5d4", "#fef9ef", "#ff85a1", "#b5838d",
}

// SectorInfo contains the initial sector info. All other data of the sector will
// be derived during rendering. The color of the sector cannot be set explicitly
// and will be chosen based on the sector's position.
type SectorInfo struct {
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

// Chart represents a chart renderer instance. It contains the initial chart
// settings, including available width and height, the chart radius, and the
// aspect ratio fix, which is responsible for adjusting the circle form depending
// on the ratio of the terminal's font width and height.
type Chart struct {
	colors         []lipgloss.Color
	width          int
	height         int
	radius         int
	aspectRatioFix float64
}

func NewChart(width, height, radius int, aspectRationFix float64, colors []lipgloss.Color) *Chart {
	// the number of colors must not be lower than the number of maxSectors
	// plus one "merged" sector.
	if len(colors) < maxSectors+1 {
		colors = defaultChartColors
	}

	return &Chart{
		colors:         colors,
		width:          width,
		height:         height,
		radius:         radius,
		aspectRatioFix: aspectRationFix,
	}
}

// Render renders a chart window including the chart itself and the corresponding
// legend. The chart form visual correctness depends on the current aspect ratio
// fix value.
func (c *Chart) Render(totalSize int64, raw []SectorInfo) string {
	sb := strings.Builder{}

	sectors := c.prepareSectors(totalSize, raw)

	// The chart center should be placed at the center of the left half of the
	// provided window area
	centerX, centerY := c.width/2/2, c.height/2

	for y := range c.height {
		for x := range c.width / 2 {
			dx := float64(x - centerX)
			dy := float64(y-centerY) * c.aspectRatioFix

			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > float64(c.radius) {
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
		lipgloss.Center, sb.String(), legend(sectors, c.width/2),
	)
}

func (c *Chart) prepareSectors(totalSize int64, si []SectorInfo) []chartSector {
	sectors := make([]chartSector, 0, len(si))

	others := chartSector{label: "Others"}

	for i, s := range si {
		usage := float64(s.Size) / float64(totalSize)

		if i >= maxSectors {
			others.size += s.Size

			continue
		}

		sectors = append(
			sectors,
			chartSector{
				label: WrapString(s.Label, chartLabelWidth),
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
		sectors[i].color = c.colors[i]
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
		label := WrapString(s.label, int(float64(width)*0.6))
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
