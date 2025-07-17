package render

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultBorder     = '\ue0b0'
	DefaultBarBGColor = "#353533"
	DynamicWidth      = -1
)

type BarItemWrapper func(data string, limit int) string

// BarItem represents a single status bar item, including its string content,
// background color, and width. The width parameter is optional, and the default
// width equals the content width.
//
// The width value -1 denotes that the item will take all available screen width
// minus the sum of all other elements' widths. If multiple items have a width
// value of -1, the resulting width will be spread equally between them.
type BarItem struct {
	Content string
	BGColor string
	Width   int
	Border  rune
	Wrapper BarItemWrapper
}

// DefaultBarItem returns a new *BarItem instance with default values for
// background color and width.
func DefaultBarItem(content string) *BarItem {
	return &BarItem{
		Content: content,
		BGColor: DefaultBarBGColor,
		Border:  DefaultBorder,
	}
}

// NewBarItem returns a new *BarItem instance based on the provided parameters.
// If the background color is an empty string, a default color will be assigned.
func NewBarItem(content, bgColor string, width int, biw BarItemWrapper) *BarItem {
	if bgColor == "" {
		bgColor = DefaultBarBGColor
	}

	return &BarItem{
		Content: content,
		BGColor: bgColor,
		Width:   width,
		Border:  DefaultBorder,
		Wrapper: biw,
	}
}

type StatusBar struct {
	items []*BarItem
}

func NewStatusBar() *StatusBar {
	return &StatusBar{
		items: make([]*BarItem, 0),
	}
}

func (sb *StatusBar) Add(items []*BarItem) {
	for _, item := range items {
		if item.BGColor == "" {
			item.BGColor = DefaultBarBGColor
		}

		if item.Border == 0 {
			item.Border = DefaultBorder
		}
	}

	sb.items = append(sb.items, items...)
}

func (sb *StatusBar) Clear() {
	sb.items = make([]*BarItem, 0)
}

// Render builds a new status bar based on the provided list of *BarItem
// instances. The total bar width is defined by the totalWidth parameter and all
// bar items will be fit in that width according to their parameters or evenly
// spread for the available width.
//
// NOTE: This implementation does not guarantee that the manually defined element
// sizes will not exceed the totalWidth value.
func (sb *StatusBar) Render(totalWidth int) string {
	styles := make([]lipgloss.Style, len(sb.items))
	renderItems := make([]string, 0, len(sb.items))
	toMaxWidth := make(map[int]struct{}, len(sb.items))

	for i := range sb.items {
		item := sb.items[i]

		if i == len(sb.items)-1 || !borderEnabled() {
			item.Border = 0
		}

		itemStyle := newBarBlockStyle(item)

		if item.Width > 0 {
			itemStyle = itemStyle.Width(item.Width)
		}

		// set the current item border bg color same as next bar item bg color.
		if i+1 < len(sb.items) {
			itemStyle = itemStyle.BorderBackground(
				lipgloss.Color(sb.items[i+1].BGColor),
			)
		}

		widthDiff := lipgloss.Width(itemStyle.Render(item.Content))

		if item.Width == DynamicWidth {
			toMaxWidth[i] = struct{}{}
			widthDiff = 1
		}

		totalWidth -= widthDiff
		styles[i] = itemStyle
	}

	var maxItemWidth int

	if len(toMaxWidth) > 0 {
		maxItemWidth = int(
			math.Ceil(float64(totalWidth) / float64(len(toMaxWidth))),
		)
	}

	for i := range sb.items {
		s := styles[i]

		if _, ok := toMaxWidth[i]; ok {
			s = s.Width(min(totalWidth, maxItemWidth))

			totalWidth -= s.GetWidth()
		}

		if sb.items[i].Wrapper != nil {
			sb.items[i].Content = sb.items[i].Wrapper(
				sb.items[i].Content, s.GetWidth()-10,
			)
		}

		renderItems = append(renderItems, s.Render(sb.items[i].Content))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderItems...)
}

func newBarBlockStyle(bi *BarItem) lipgloss.Style {
	s := style.BarBlock(lipgloss.Color(bi.BGColor))

	if bi.Border != 0 {
		s = s.Border(
			lipgloss.Border{Right: string(bi.Border)},
			false,
			true,
			false,
			false,
		).BorderForeground(lipgloss.Color(bi.BGColor))
	}

	return s
}

func borderEnabled() bool {
	return style.CS().StatusBarBorder
}
