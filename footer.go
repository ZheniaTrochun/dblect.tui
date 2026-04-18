package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

func renderFooter(viewName string, width int) string {

	controls := []string{
		"k/↑ - вгору",
		"j/↓ - вниз",
		"1-4 - обрати варіант",
		"enter - обрати",
		"esc - назад",
		"q - вийти",
		"? - help",
		"l - language",
	}

	controlsText := strings.Join(controls, separatorString)

	pageKeyText := "  " + strings.ToUpper(viewName) + "   "
	statusText := "  ●  "
	langText := "   Ukrainian"

	controlsWidth := width - len(pageKeyText) - len(statusText) - len(langText)
	if controlsWidth < 0 {
		controlsWidth = 0
	}

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
