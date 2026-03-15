package main

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

var (
	lecturesMap, initialReadErr = readAllLectures(os.DirFS("."))
)

type lectureModel struct {
	width    int
	height   int
	lecture  string
	lang     string
	ready    bool
	viewport viewport.Model
}

func (m lectureModel) Init() tea.Cmd {
	return nil
}

func (m lectureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "esc" || k == "backspace" {
			return m, func() tea.Msg {
				return NavEvent{navTo: lecturesView}
			}
		}

	case OpenLecture:
		if initialReadErr != nil {
			log.Error("Failed to open lecture", "Error", initialReadErr)
			return m, tea.Quit
		}

		m.lecture = msg.name

		m.viewport = m.createViewport()

		if lectureContent, ok := lecturesMap[m.lecture]; ok {
			mdRenderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle("dracula"),
				glamour.WithWordWrap(m.width-10),
			)
			if err != nil {
				log.Error("Failed to create glamour renderer", "Error", err)
				return m, tea.Quit
			}

			renderedLecture, err := mdRenderer.Render(lectureContent)
			if err != nil {
				log.Error("Failed to render lecture content", "Error", err)
				return m, tea.Quit
			}

			m.viewport.SetContent(renderedLecture)
		} else {
			log.Error("Failed to find lecture", "lecture", m.lecture)
			return m, func() tea.Msg {
				return NavEvent{navTo: lecturesView}
			}
		}

		m.ready = true

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = m.createViewport()
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - lipgloss.Height(m.headerView()) - lipgloss.Height(m.footerView()))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m lectureModel) View() tea.View {
	var v tea.View
	v.AltScreen = true                    // use the full size of the terminal in its "alternate screen buffer"
	v.MouseMode = tea.MouseModeCellMotion // turn on mouse support so we can track the mouse wheel

	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
	}
	return v
}

func (m lectureModel) headerView() string {
	title := titleStyle.Render(m.lecture)
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m lectureModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m lectureModel) createViewport() viewport.Model {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	v := viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height-verticalMarginHeight))
	v.YPosition = headerHeight
	v.LeftGutterFunc = func(info viewport.GutterContext) string {
		if info.Soft {
			return "     │ "
		}
		if info.Index >= info.TotalLines {
			return "   ~ │ "
		}
		return fmt.Sprintf("%4d │ ", info.Index+1)
	}
	v.HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("34"))
	v.SelectedHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("47"))

	return v
}

func readAllLectures(fsys fs.FS) (map[string]string, error) {
	dirs, err := fs.ReadDir(fsys, "lectures")

	if err != nil {
		log.Error("Failed to read list of lectures", "err", err)
		return make(map[string]string), err
	}

	names := make([]string, 0)

	for _, dir := range dirs {
		if dir.IsDir() {
			names = append(names, dir.Name())
		}
	}

	result := make(map[string]string)

	for _, name := range names {
		content, err := readLectureContentByName(fsys, name)

		if err != nil {
			log.Error("Failed to read lecture", "name", name, "err", err)
			return make(map[string]string), err
		}

		result[name] = content
	}

	return result, nil
}

func readLectureContentByName(fsys fs.FS, name string) (string, error) {
	bytes, err := fs.ReadFile(fsys, path.Join("lectures", name, "lecture_notes.md"))
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
