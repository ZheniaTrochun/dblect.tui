package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type view int

type NavEvent struct {
	navTo view
}

const (
	homeView view = iota
	lecturesView
)

type model struct {
	width  int
	height int
	pty    ssh.Pty

	home    tea.Model
	lecture tea.Model

	state view
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		updatedHome, _ := m.home.Update(msg)
		updatedLectures, _ := m.lecture.Update(msg)

		m.home = updatedHome
		m.lecture = updatedLectures
	}

	var cmd tea.Cmd
	var updatedView tea.Model

	switch m.state {
	case homeView:
		updatedView, cmd = m.home.Update(msg)
		m.home = updatedView
	case lecturesView:
		updatedView, cmd = m.lecture.Update(msg)
		m.lecture = updatedView
	}

	switch msg := msg.(type) {
	case NavEvent:
		m.state = msg.navTo
	}

	return m, cmd
}

func (m model) View() tea.View {
	switch m.state {
	case homeView:
		return m.home.View()
	case lecturesView:
		return m.lecture.View()
	}

	return tea.NewView("")
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	lecture := lecturesModel{
		height: pty.Window.Height,
		width:  pty.Window.Width,
	}

	home := homeModel{
		height: pty.Window.Height,
		width:  pty.Window.Width,
		pty:    pty,
		cursor: 0,
	}

	model := model{
		pty: pty,

		width:  pty.Window.Width,
		height: pty.Window.Height,

		state:   homeView,
		lecture: lecture,
		home:    home,
	}

	return model, []tea.ProgramOption{}
}

const (
	host = "localhost"
	port = "23234"
)

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
