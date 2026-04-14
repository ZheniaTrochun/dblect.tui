package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

const (
	leftTitle = " dblect"
	subTitle  = "database lecture terminal "
)

var (
	headerStyle = defaultStyle.
		Border(lipgloss.NormalBorder()).
		BorderForeground(defaultBorder).
		//Padding(0, 1).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
)

func renderHeader(width int) string {
	headerLeftTitle := defaultStyle.Foreground(active).Align(lipgloss.Left).Render(leftTitle)
	headerSubTitle := defaultStyle.Foreground(textDim).Align(lipgloss.Right).Render(subTitle)
	spacer := defaultStyle.Render(strings.Repeat(" ", width-2-lipgloss.Width(headerLeftTitle)-lipgloss.Width(headerSubTitle)))

	return headerStyle.Render(headerLeftTitle + spacer + headerSubTitle)
}
