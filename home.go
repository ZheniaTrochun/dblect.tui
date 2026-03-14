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
	selectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94")).
			Underline(true)

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(5, 5).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C1C6B2")).
			Background(lipgloss.Color("#353533"))

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Background(lipgloss.Color("#6124DF"))

	docStyle = lipgloss.NewStyle().Padding(0, 0, 0, 0)
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
		case "ctrl+c", "q", "й":
			return m, tea.Quit
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

	grad := applyGradient(
		lipgloss.NewStyle(),
		banner,
		lipgloss.Color("#EDFF82"),
		lipgloss.Color("#F25D94"),
	)

	header := lipgloss.NewStyle().
		Width(70).
		Align(lipgloss.Center).
		Render(grad)

	chooseList := lipgloss.NewStyle().Align(lipgloss.Left).Width(maxChoiceLength + 4).Render(s)

	choicesBox := dialogBoxStyle.
		Width(70).
		Align(lipgloss.Center).
		Render(chooseList)

	ui := lipgloss.JoinVertical(lipgloss.Center, header, choicesBox)

	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ui,
		lipgloss.WithWhitespaceChars("  "),
	)

	doc.WriteString(dialog + "\n\n")

	w := lipgloss.Width

	controlsText := "\nj/↑ - вгору, k/↓ - вниз, 1/2/3 - перейти на варіант N, enter - обрати, q - вийти\n"

	pageKey := statusStyle.Render("\nMAIN\n")
	encoding := encodingStyle.Render("\nUTF-8\n")
	lang := fishCakeStyle.Render("\nUkrainian\n")
	controls := statusText.
		Width(m.width - w(pageKey) - w(encoding) - w(lang)).
		Height(3).
		Align(lipgloss.Center).
		Render(controlsText)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		pageKey,
		controls,
		encoding,
		lang,
	)

	doc.WriteString(statusBarStyle.Width(m.width).Render(bar))

	v := tea.NewView(docStyle.Render(doc.String()))
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
