package main

import (
	"charm.land/bubbles/v2/list"
	"charm.land/log/v2"
	_ "embed"
	"io/fs"
	"os"

	tea "charm.land/bubbletea/v2"
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
func (i item) FilterValue() string { return i.title }

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

	listModel := list.New(listLectureItems, list.NewDefaultDelegate(), 0, 0)

	listModel.Title = "Lectures"

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
		m.lecturesList.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.lecturesList, cmd = m.lecturesList.Update(msg)

	return m, cmd
}

func (m lecturesModel) View() tea.View {
	v := tea.NewView(m.lecturesList.View())
	v.AltScreen = true
	return v
}
