package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

var controls = []string{
	"k/↑ - вгору",
	"j/↓ - вниз",
	"1-4 - обрати варіант",
	"enter - обрати",
	"esc - назад",
	"q - вийти",
	"? - help",
	"l - language",
}

var longestControlLen = getLongestControlLen()

func renderFooter(viewName string, width int) string {
	pageKeyText := "  " + strings.ToUpper(viewName) + "   "
	statusText := "  ●  "
	langText := "   Ukrainian"

	pageKeyTextWidth := lipgloss.Width(pageKeyText)
	statusTextWidth := lipgloss.Width(statusText)
	langTextWidth := lipgloss.Width(langText)

	if pageKeyTextWidth+statusTextWidth+langTextWidth+2 >= width {
		statusText = ""
	}

	controlsWidth := width - pageKeyTextWidth - statusTextWidth - langTextWidth - 2
	if controlsWidth < longestControlLen {
		controlsWidth = 0
	}

	controlsText := buildControlsText(controlsWidth)

	renderedControls := boxWithBorderStyle.
		Foreground(textDim).
		Width(controlsWidth).
		Height(2).
		Align(lipgloss.Center).
		BorderLeft(false).
		BorderRight(false).
		Render(controlsText)

	controlsHeight := lipgloss.Height(renderedControls)

	pageKey := boxWithBorderStyle.
		Foreground(active).
		Height(controlsHeight).
		AlignVertical(lipgloss.Center).
		BorderRight(false).
		Render(pageKeyText)

	lang := boxWithBorderStyle.
		Foreground(active).
		Height(controlsHeight).
		AlignVertical(lipgloss.Center).
		BorderLeft(false).
		BorderRight(false).
		Render(langText)

	status := boxWithBorderStyle.
		Foreground(okColor).
		Height(controlsHeight).
		AlignVertical(lipgloss.Center).
		BorderLeft(false).
		Render(statusText)

	return lipgloss.JoinHorizontal(lipgloss.Left, pageKey, renderedControls, lang, status)
}

func buildControlsText(width int) string {
	if width < longestControlLen {
		return ""
	}

	var lines []string
	acc := controls[0]
	separatorWidth := lipgloss.Width(separatorString)

	for _, control := range controls[:1] {
		newWidth := lipgloss.Width(acc) + lipgloss.Width(control) + separatorWidth

		if newWidth > width {
			lines = append(lines, acc)
			acc = control
		} else {
			acc = acc + separatorString + control
		}
	}

	return strings.Join(append(lines, acc), "\n")
}

func getLongestControlLen() int {
	longestControl := 0
	for _, control := range controls {
		if lipgloss.Width(control) > longestControl {
			longestControl = lipgloss.Width(control)
		}
	}

	return longestControl
}
