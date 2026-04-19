package main

import (
	_ "embed"
	"fmt"
	"github.com/rivo/uniseg"
	"image/color"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ssh"
)

//go:embed banner.txt
var banner string

var (
	choices            = []string{"Лекції", "Рейтинг", "SQL пісочниця"}
	choiceDescriptions = []string{"Конспекти лекцій", "База даних рейтингових балів", "Sandbox бази даних \"кампусу\""}

	choiceToNav = map[string]view{choices[0]: lecturesView, choices[1]: homeView, choices[2]: homeView}

	longestDescription = slices.MaxFunc(choiceDescriptions, func(a, b string) int { return len(a) - len(b) })

	// todo - чогось серйозно не вистачає
	lectureSections      = []string{"Схема БД", "SQL", "Нормалізація", "Індекси", "Транзакції", "NoSQL"}
	lectureSectionsSizes = map[string]int{"Схема БД": 3, "SQL": 4, "Нормалізація": 2, "Індекси": 2, "Транзакції": 1, "NoSQL": 1}
	totalLectures        = 18
	longestLectureName   = slices.MaxFunc(lectureSections, func(a, b string) int { return len(a) - len(b) })

	progressMock = map[string]int{"Схема БД": 2, "SQL": 3, "Нормалізація": 1, "Індекси": 0, "Транзакції": 1, "NoSQL": 0}
)

type homeModel struct {
	term    string
	profile string
	width   int
	height  int
	pty     ssh.Pty

	cursor int
}

func (m homeModel) Init() tea.Cmd {
	return nil
}

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "л":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "о":
			if m.cursor < len(choices)-1 {
				m.cursor++
			}
		case "enter", "space", " ":
			return m, func() tea.Msg {
				return NavEvent{navTo: choiceToNav[choices[m.cursor]]}
			}
		case "1":
			m.cursor = 0
		case "2":
			m.cursor = 1
		case "3":
			m.cursor = 2
		case "esc":
			return m, func() tea.Msg {
				return NavEvent{navTo: homeView}
			}
		}
	}

	return m, nil
}

func (m homeModel) View() tea.View {
	doc := strings.Builder{}

	header := renderHeader(m.width)

	formattedBanner := formatBanner(banner, m.width)

	navSectionBanner := defaultStyle.
		Align(lipgloss.Left).
		Foreground(active).
		Render(formattedBanner)

	statusBar := buildStatus(m.width)

	chooseList := buildOptionsList(m.cursor)
	cursorTopOffset := 14 + m.cursor
	cursorLeftOffset := 5
	selectionCursor := tea.NewCursor(cursorLeftOffset, cursorTopOffset)
	selectionCursor.Color = active
	selectionCursor.Blink = true

	choicesBox := defaultStyle.Width(m.width).PaddingBottom(1).Render("\n" + chooseList)

	horizontalDividerText := " " + strings.Repeat("─", m.width-2) + " "

	horizontalDivider := defaultStyle.Width(m.width).Foreground(defaultBorder).Render(horizontalDividerText)

	//progressSection := buildProgress(m.width / 2)
	progressSection := buildProgress(m.width)

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		navSectionBanner,
		statusBar,
		choicesBox,
		horizontalDivider,
		progressSection,
	)

	dialog := lipgloss.Place(m.width, m.height-5,
		lipgloss.Left, lipgloss.Left,
		ui,
		lipgloss.WithWhitespaceChars("  "),
	)

	doc.WriteString(dialog + "\n\n")

	footer := renderFooter("normal", m.width)

	doc.WriteString(footer)

	v := tea.NewView(defaultStyle.Render(doc.String()))
	v.AltScreen = true
	v.Cursor = selectionCursor

	return v
}

// applyGradient applies a gradient to the given string.
func applyGradient(base lipgloss.Style, input string, from, to color.Color) string {
	// We want to get the graphemes of the input string, which is the number of
	// characters as a human would see them.
	//
	// We definitely don't want to use len(), because that returns the
	// bytes. The rune count would get us closer but there are times, like with
	// emojis, where the rune count is greater than the number of actual
	// characters.
	g := uniseg.NewGraphemes(input)
	var chars []string
	for g.Next() {
		chars = append(chars, g.Str())
	}

	gradient := lipgloss.Blend1D(len(chars), from, to)
	var output strings.Builder
	for i, char := range chars {
		output.WriteString(base.Foreground(gradient[i]).Render(char))
	}
	return output.String()
}

func formatBanner(banner string, width int) string {
	splitted := strings.Split(banner, "\n")

	var formatted strings.Builder

	formatted.WriteString(strings.Repeat(" ", width) + "\n")
	for _, line := range splitted {
		padding := width - lipgloss.Width(line) - 2
		formattedLine := "  " + line + strings.Repeat(" ", padding) + "\n"
		formatted.WriteString(formattedLine)
	}
	formatted.WriteString(strings.Repeat(" ", width))
	//formatted.WriteString(strings.Repeat(" ", width) + "\n")

	return formatted.String()
}

func buildStatus(width int) string {
	connectionLabel := defaultStyle.Foreground(textDim).Render("connected to ")
	connectionName := defaultStyle.Foreground(textMain).Render("databases_lecture_db")
	versionLabel := defaultStyle.Foreground(textDim).Render("postgresql ")
	version := defaultStyle.Foreground(textMain).Render("17.5")
	statusIndicator := defaultStyle.Foreground(okColor).Render("● online")

	statusLine := connectionLabel + connectionName + styledSeparator + versionLabel + version + styledSeparator + statusIndicator

	paddingLeft := "  "
	paddingRightLen := width - lipgloss.Width(statusLine) - 2
	paddingRight := defaultStyle.Render(strings.Repeat(" ", paddingRightLen) + "\n")

	return defaultStyle.Render(paddingLeft + statusLine + paddingRight)
}

func buildOptionsList(cursor int) string {
	acc := ""
	for i, choice := range choices {
		orderPrefix := "  " + strconv.Itoa(i+1)

		var choiceColor color.Color

		if cursor == i {
			choiceColor = active
		} else {
			choiceColor = textMain
		}

		rowMainText := orderPrefix + "    " + choice

		renderedMainText := defaultStyle.Align(lipgloss.Left).Width(25).Foreground(choiceColor).Render(rowMainText)
		renderedDescription := defaultStyle.Align(lipgloss.Right).Width(lipgloss.Width(longestDescription)).Foreground(textDim).Render(choiceDescriptions[i])

		fullRow := lipgloss.JoinHorizontal(lipgloss.Left, renderedMainText, renderedDescription)

		acc += fullRow + "\n"
	}

	return acc
}

func buildProgress(width int) string {
	prefixText := "--progress"
	sectionNameLen := len(longestLectureName) + 4

	progressBarLen := width - sectionNameLen - 7

	var res strings.Builder

	res.WriteString(defaultStyle.Foreground(textDim).Width(width).PaddingLeft(2).Render(prefixText))
	res.WriteString("\n\n")

	totalCompleted := 0

	for _, section := range lectureSections {
		completed := progressMock[section]
		total := lectureSectionsSizes[section]

		totalCompleted += completed

		result := strconv.Itoa(completed) + "/" + strconv.Itoa(total)

		var renderedResult string
		if completed == total {
			renderedResult = defaultStyle.Foreground(okColor).PaddingRight(2).PaddingLeft(2).Render(result)
		} else {
			renderedResult = defaultStyle.Foreground(textDim).PaddingRight(2).PaddingLeft(2).Render(result)
		}

		progressBar := buildProgressBar(total, completed, progressBarLen)

		renderedSectionName := defaultStyle.
			Foreground(textMain).
			Width(sectionNameLen).
			PaddingLeft(2).
			PaddingRight(2).
			Render(section)

		rowText := renderedSectionName + progressBar + renderedResult

		renderedRow := defaultStyle.Foreground(textMain).Width(width).Render(rowText)

		res.WriteString(renderedRow + "\n")
	}

	res.WriteString("\n\n")

	var resultColor color.Color
	if totalCompleted == totalLectures {
		resultColor = okColor
	} else {
		resultColor = textMain
	}

	totalResult := defaultStyle.Foreground(resultColor).Width(width).PaddingRight(2).PaddingLeft(2).Render(fmt.Sprintf("Пройдено %d / 18 лекцій", totalCompleted))
	res.WriteString(totalResult)

	return defaultStyle.Width(width).Render(res.String())
}

func buildProgressBar(steps, completed, length int) string {
	itemsPerSection := length / steps
	doneLen := itemsPerSection * completed
	todoLen := length - doneLen

	//singleItem := "― ━ ▬"
	singleItem := "━"

	doneText := strings.Repeat(singleItem, doneLen)
	todoText := strings.Repeat(singleItem, todoLen)

	done := defaultStyle.Foreground(active).Render(doneText)
	todo := defaultStyle.Foreground(textDim).Render(todoText)

	return done + todo
}
