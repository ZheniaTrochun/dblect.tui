package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

const (
	leftTitle = " dblect"
	subTitle  = "database lecture terminal "
)

func renderHeader(width int) string {
	headerLeftTitle := defaultStyle.Foreground(active).Align(lipgloss.Left).Render(leftTitle)
	headerSubTitle := defaultStyle.Foreground(textDim).Align(lipgloss.Right).Render(subTitle)

	// `-2` is needed to compensate borders
	numOfSpaces := width - lipgloss.Width(headerLeftTitle) - lipgloss.Width(headerSubTitle) - 2
	spacer := defaultStyle.Render(strings.Repeat(" ", numOfSpaces))

	return boxWithBorderStyle.Render(headerLeftTitle + spacer + headerSubTitle)
}
