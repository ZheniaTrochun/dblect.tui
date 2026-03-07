package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/rivo/uniseg"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	//tea "charm.land/bubbletea/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

var (
	choices       = []string{"Лекції", "Рейтинг", "Вправи SQL"}
	longestChoice = slices.MaxFunc(choices, func(a, b string) int {
		return len(a) - len(b)
	})
	maxChoiceLength = len(longestChoice)
)

type model struct {
	term    string
	profile string
	width   int
	height  int
	bg      string
	pty     ssh.Pty

	choices  []string
	cursor   int
	selected map[int]struct{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", "space", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := ""
	for i, choice := range m.choices {
		if i != 0 {
			s += "\n\n"
		}

		if m.cursor == i {
			s += "  " + selectionStyle.Align(lipgloss.Center).Render(choice)
		} else {
			s += choice
		}
	}

	doc := strings.Builder{}

	{
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

		chooseList := lipgloss.NewStyle().Align(lipgloss.Left).Width(maxChoiceLength + 2).Render(s)

		choicesBox := dialogBoxStyle.
			Width(70).
			Align(lipgloss.Center).
			Render(chooseList)

		ui := lipgloss.JoinVertical(lipgloss.Center, header, choicesBox)

		dialog := lipgloss.Place(m.pty.Window.Width, m.pty.Window.Height,
			lipgloss.Center, lipgloss.Center,
			ui,
			lipgloss.WithWhitespaceChars("  "),
		)

		doc.WriteString(dialog + "\n\n")
	}

	return docStyle.Render(doc.String())
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

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	renderer := bubbletea.MakeRenderer(s)

	bg := "light"
	if renderer.HasDarkBackground() {
		bg = "dark"
	}

	model := model{
		term:    pty.Term,
		profile: renderer.ColorProfile().Name(),
		bg:      bg,
		pty:     pty,

		width:  pty.Window.Width,
		height: pty.Window.Height,

		choices:  []string{"Лекції", "Рейтинг", "Вправи SQL"},
		cursor:   0,
		selected: make(map[int]struct{}),
	}

	return model, []tea.ProgramOption{tea.WithAltScreen()}
}

const (
	host = "localhost"
	port = "23234"
)

//go:embed banner.txt
var banner string

func main() {

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath("./ssh-keys/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)

	if err != nil {
		fmt.Println("UNRECOVERABLE ERRRRROOOORRRRRR: %s", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server...", "Host", host, "Port", port)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "Error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not shutdown server", "Error", err)
	}
}
