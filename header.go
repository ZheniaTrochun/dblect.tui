package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

const (
	leftTitle = " dblect"
	subTitle  = "database lecture terminal "
)

func renderHeader(width int, hasContinuation bool) string {
	headerLeftTitle := defaultStyle.Foreground(active).Align(lipgloss.Left).Render(leftTitle)
	headerSubTitle := defaultStyle.Foreground(textDim).Align(lipgloss.Right).Render(subTitle)

	// `-2` is needed to compensate borders
	numOfSpaces := width - lipgloss.Width(headerLeftTitle) - lipgloss.Width(headerSubTitle) - 2
	spacer := defaultStyle.Render(strings.Repeat(" ", numOfSpaces))

	var style lipgloss.Style
	if hasContinuation {
		border := lipgloss.NormalBorder()
		border.BottomLeft = border.MiddleLeft
		border.BottomRight = border.MiddleRight
		style = boxWithBorderStyle.Border(border)
	} else {
		style = boxWithBorderStyle
	}

	return style.Render(headerLeftTitle + spacer + headerSubTitle)
}
