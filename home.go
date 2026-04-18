package main

import (
	_ "embed"
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
	selectionStyle = defaultStyle.
			Foreground(lipgloss.Color("#F25D94")).
			Underline(true)

	dialogBoxStyle = defaultStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(5, 5).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
)

var (
	choices = []string{"Лекції", "Рейтинг", "Вправи SQL"}

	choiceToNav = map[string]view{choices[0]: lecturesView, choices[1]: homeView, choices[2]: homeView}

	longestChoice = slices.MaxFunc(choices, func(a, b string) int {
		return len(a) - len(b)
	})
	maxChoiceLength = len(longestChoice)
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
	s := ""
	for i, choice := range choices {
		if i != 0 {
			s += "\n\n"
		}

		orderPrefix := strconv.Itoa(i+1) + ". "

		if m.cursor == i {
			s += "  " + selectionStyle.Align(lipgloss.Center).Render(orderPrefix+choice)
		} else {
			s += orderPrefix + choice
		}
	}

	doc := strings.Builder{}

	header := renderHeader(m.width)

	formattedBanner := formatBanner(banner, m.width)

	navSectionBanner := defaultStyle.
		Align(lipgloss.Left).
		Foreground(active).
		Render(formattedBanner)

	statusBar := buildStatus(m.width)

	chooseList := defaultStyle.Align(lipgloss.Left).Width(maxChoiceLength + 4).Render(s)

	choicesBox := dialogBoxStyle.
		Width(70).
		Align(lipgloss.Left).
		Render(chooseList)

	ui := lipgloss.JoinVertical(lipgloss.Left, header, navSectionBanner, statusBar, choicesBox)

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
