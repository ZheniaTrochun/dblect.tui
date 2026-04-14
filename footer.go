package main

import (
	"charm.land/lipgloss/v2"
	"strings"
)

func renderFooter(viewName string, width int) string {

	controls := []string{"k/↑ - вгору", "j/↓ - вниз", "1-4 - обрати варіант", "enter - обрати", "esc - назад", "q - вийти"}

	//controlsText := "\nk/↑ - вгору, j/↓ - вниз, 1/2/3 - перейти на варіант N, enter - обрати, q - вийти\n"

	controlsText := strings.Join(controls, "   ")

	pageKey := defaultStyle.Foreground(active).Render("  " + strings.ToUpper(viewName))
	lang := defaultStyle.Foreground(active).Render("Ukrainian")
	status := defaultStyle.Foreground(okColor).Render("  ●  ")

	controlsWidth := width - lipgloss.Width(pageKey) - lipgloss.Width(status) - lipgloss.Width(lang)
	if controlsWidth < 0 {
		controlsWidth = 0
	}

	renderedControls := defaultStyle.
		Foreground(textDim).
		Width(controlsWidth).
		Height(1).
		Align(lipgloss.Center).
		Render(controlsText)

	//return lipgloss.JoinHorizontal(lipgloss.Top,
	//	pageKey,
	//	renderedControls,
	//	lang,
	//	status,
	//)

	return defaultStyle.
		BorderBottom(true).
		BorderTop(true).
		BorderRight(true).
		BorderLeft(true).
		Render(pageKey + renderedControls + lang + status)
}
