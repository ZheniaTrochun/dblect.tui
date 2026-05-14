package main

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// todo: extract to a separate file
var (
	lectures, lecturesReadErr = readAvailableLectures()
)

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return "" }

type lecturesModel struct {
	width        int
	height       int
	lectures     []item
	lecturesList list.Model
	lang         string
	cursor       int
}

func newLecturesModel(width, height int) lecturesModel {
	lectureItems := make([]item, len(lectures))
	for i, lecture := range lectures {
		lectureItems[i] = item{title: lecture}
	}

	listLectureItems := make([]list.Item, len(lectures))
	for i, lecture := range lectureItems {
		listLectureItems[i] = lecture
	}

	listModel := list.New(listLectureItems, itemDelegate{}, 0, 0)

	listModel.SetShowHelp(false)
	listModel.SetShowFilter(false)
	listModel.SetHeight(height - 8)
	listModel.SetWidth(width - 10)
	listModel.SetShowTitle(false)
	listModel.SetFilteringEnabled(false)
	listModel.SetShowStatusBar(false)

	return lecturesModel{
		width:        width,
		height:       height,
		lectures:     lectureItems,
		lecturesList: listModel,
	}
}

func readAvailableLectures() ([]string, error) {
	lectureDirs, err := fs.ReadDir(os.DirFS("."), "lectures")

	if err != nil {
		log.Error("Failed to read list of lectures", "err", err)
		return make([]string, 0), err
	}

	var lectureNames []string

	for _, lectureDir := range lectureDirs {
		if lectureDir.IsDir() {
			lectureNames = append(lectureNames, lectureDir.Name())
		}
	}

	return lectureNames, nil
}

func (m lecturesModel) Init() tea.Cmd {
	return nil
}

func (m lecturesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if lecturesReadErr != nil {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavEvent{navTo: homeView}
			}

		case "enter":
			i, ok := m.lecturesList.SelectedItem().(item)
			if ok {
				return m, func() tea.Msg {
					return OpenLecture{i.title}
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.lecturesList.SetSize(msg.Width, msg.Height-10)
	}

	var cmd tea.Cmd
	m.lecturesList, cmd = m.lecturesList.Update(msg)

	return m, cmd
}

func (m lecturesModel) View() tea.View {
	header := renderHeader(m.width, true)

	tableHeader := renderTableHeader(m.width)
	table := m.lecturesList.View()

	footer := renderFooter("lectures", m.width)

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tableHeader,
		table,
		footer,
	)

	v := tea.NewView(ui)
	v.AltScreen = true

	cursorTopOffset := 6 + m.lecturesList.Index()
	cursorLeftOffset := 2
	selectionCursor := tea.NewCursor(cursorLeftOffset, cursorTopOffset)
	selectionCursor.Color = active
	selectionCursor.Blink = true
	v.Cursor = selectionCursor

	v.BackgroundColor = panelBackground

	return v
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var rendered string
	if index == m.Index() {
		rendered = defaultStyle.Background(lipgloss.Color("#131312")).Width(m.Width()).Foreground(active).Render("    " + i.Title())
	} else {
		rendered = defaultStyle.Foreground(textMain).Render("    " + i.Title())
	}

	fmt.Fprint(w, rendered)
}

func renderTableHeader(wight int) string {
	numSize := 9
	topicSize := 15
	dateSize := 7
	statusSize := 9

	titleSize := wight - 6 - numSize - topicSize - dateSize - statusSize

	num := defaultStyle.Foreground(textDim).Align(lipgloss.Left).Width(numSize).Render(" #")
	title := defaultStyle.Foreground(textDim).Align(lipgloss.Left).Width(titleSize).Render(" title")
	topic := defaultStyle.Foreground(textDim).Align(lipgloss.Left).Width(topicSize).Render(" topic")
	date := defaultStyle.Foreground(textDim).Align(lipgloss.Left).Width(dateSize).Render(" date")
	status := defaultStyle.Foreground(textDim).Align(lipgloss.Left).Width(statusSize).Render(" status")

	return boxWithBorderStyle.
		PaddingTop(0).
		BorderTop(false).
		MarginBottom(1).
		Render(num, title, topic, date, status)
}
