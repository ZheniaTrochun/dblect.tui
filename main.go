package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"context"
	_ "embed"
	"errors"
	"github.com/charmbracelet/colorprofile"
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

type OpenLecture struct {
	name string
}

const (
	homeView view = iota
	lecturesView
	lectureView
)

type model struct {
	width  int
	height int
	pty    ssh.Pty

	home     tea.Model
	lectures tea.Model
	lecture  tea.Model

	state view
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.home, cmd = m.home.Update(msg)
		cmds = append(cmds, cmd)
		m.lecture, cmd = m.lecture.Update(msg)
		cmds = append(cmds, cmd)
		m.lectures, cmd = m.lectures.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case NavEvent:
		m.state = msg.navTo

	case OpenLecture:
		m.state = lectureView
	}

	switch m.state {
	case homeView:
		m.home, cmd = m.home.Update(msg)
	case lecturesView:
		m.lectures, cmd = m.lectures.Update(msg)
	case lectureView:
		m.lecture, cmd = m.lecture.Update(msg)
	default:
		log.Error("Unexpected navigation state", "State", m.state)
		m.state = homeView
		m.home, cmd = m.home.Update(msg)
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	switch m.state {
	case homeView:
		return m.home.View()
	case lecturesView:
		return m.lectures.View()
	case lectureView:
		return m.lecture.View()
	default:
		log.Error("Unexpected navigation state", "State", m.state)
		return m.home.View()
	}
}

func teaHandler(s ssh.Session) *tea.Program {
	pty, _, ok := s.Pty()
	if !ok {
		log.Error("Client connected without PTY - connection refused", "User", s.User())
		_ = s.Exit(1)
		return nil
	}
	envs := append(s.Environ(), "TERM="+pty.Term, "COLORTERM=truecolor", "CLICOLOR_FORCE=1")

	user := s.User()

	log.Info("New connection established", "User", user)

	lecture := lectureModel{
		height: pty.Window.Height,
		width:  pty.Window.Width,
	}

	lectures := newLecturesModel(pty.Window.Width, pty.Window.Height)

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

		state:    homeView,
		lectures: lectures,
		lecture:  lecture,
		home:     home,
	}

	opts := []tea.ProgramOption{
		tea.WithInput(s),
		tea.WithOutput(s),
		tea.WithEnvironment(envs),
		tea.WithWindowSize(pty.Window.Width, pty.Window.Height),
		tea.WithColorProfile(colorprofile.TrueColor),
		// copied from charm.land/wish/v2@v2.0.0/bubbletea/tea.go:93
		tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
			if _, ok := msg.(tea.SuspendMsg); ok {
				return tea.ResumeMsg{}
			}
			return msg
		}),
	}

	return tea.NewProgram(model, opts...)
}

const (
	host = "0.0.0.0"
	port = "23234"
)

func main() {

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath("./ssh-keys/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)

	if err != nil {
		log.Error("UNRECOVERABLE ERRRRROOOORRRRRR", "Error", err)
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
		os.Exit(1)
	}
}
