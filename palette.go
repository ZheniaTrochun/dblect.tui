package main

import "charm.land/lipgloss/v2"

var (
	toolbarBackground = lipgloss.Color("#161e25")
	panelBackground   = lipgloss.Color("#0f1519")
	bodyBackground    = lipgloss.Color("#090c0f")
	hoverBackground   = lipgloss.Color("#1e2a34")

	defaultBorder = lipgloss.Color("#1a2a38")

	textMain = lipgloss.Color("#c0d8e8")
	textSub  = lipgloss.Color("#7a9ab0")
	textDim  = lipgloss.Color("#3d5a70")

	mainAccent   = lipgloss.Color("#ffc060")
	active       = lipgloss.Color("#e8973a")
	borderAccent = lipgloss.Color("#7a4a12")

	errorColor = lipgloss.Color("#c05050")
	okColor    = lipgloss.Color("#4ab87a")

	defaultStyle = lipgloss.NewStyle().Background(panelBackground)

	boxWithBorderStyle = defaultStyle.
				Border(lipgloss.NormalBorder()).
				BorderForeground(defaultBorder).
				BorderBackground(panelBackground).
				BorderTop(true).
				BorderLeft(true).
				BorderRight(true).
				BorderBottom(true)
)
