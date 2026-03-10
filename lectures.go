package main

import (
	"charm.land/glamour/v2"
	"charm.land/log/v2"
	_ "embed"
	"fmt"
	//"github.com/charmbracelet/bubbles/viewport"

	//"github.com/charmbracelet/bubbles/viewport"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"strings"
)

//go:embed lecture_note_sample.md
var lectureSample string

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

	mdRenderer, _ = glamour.NewTermRenderer(glamour.WithStandardStyle("dracula"))
)

type lecturesModel struct {
	width    int
	height   int
	lecture  int
	lang     string
	ready    bool
	viewport viewport.Model
}

func (m lecturesModel) Init() tea.Cmd {
	return nil
}

func (m lecturesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height-verticalMarginHeight))
			m.viewport.YPosition = headerHeight
			m.viewport.LeftGutterFunc = func(info viewport.GutterContext) string {
				if info.Soft {
					return "     │ "
				}
				if info.Index >= info.TotalLines {
					return "   ~ │ "
				}
				return fmt.Sprintf("%4d │ ", info.Index+1)
			}
			m.viewport.HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("34"))
			m.viewport.SelectedHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("47"))

			mdRenderer, _ = glamour.NewTermRenderer(
				glamour.WithStandardStyle("dracula"),
				glamour.WithWordWrap(m.width-10),
			)
			renderedLecture, _ := mdRenderer.Render(lectureSample)
			m.viewport.SetContent(renderedLecture)
			//m.viewport.SetContent(lectureSample)
			//m.viewport.SetHighlights(regexp.MustCompile("БД").FindAllStringIndex(lectureSample, -1))
			m.viewport.HighlightNext()

			m.ready = true

			log.Info("ready", "m.ready", m.ready)
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - verticalMarginHeight)
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m lecturesModel) View() tea.View {
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

func (m lecturesModel) headerView() string {
	trimmedLectureContent := strings.TrimSpace(lectureSample)
	withoutH1 := strings.TrimLeft(trimmedLectureContent, "# ")

	title := titleStyle.Render(strings.Split(withoutH1, "\n")[0])
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m lecturesModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
