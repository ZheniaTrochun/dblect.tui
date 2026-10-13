package main

import (
	"charm.land/log/v2"
	_ "embed"
	"io/fs"
	"path"
)

const (
	lecturesFolderName   = "lectures"
	lectureNotesFileName = "lecture_notes.md"
)

//go:embed banner.txt
var banner string

//go:embed test_tip.md
var testTip string

func readAllLectures(fsys fs.FS) ([]string, map[string]string, error) {
	names, err := readAvailableLectures(fsys)

	if err != nil {
		log.Error("Failed to read list of lectures", "err", err)
		return make([]string, 0), make(map[string]string), err
	}

	result := make(map[string]string)

	for _, name := range names {
		content, err := readLectureContentByName(fsys, name)

		if err != nil {
			log.Error("Failed to read lecture", "name", name, "err", err)
			return make([]string, 0), make(map[string]string), err
		}

		if content != "" {
			result[name] = content
		}
	}

	return names, result, nil
}

func readLectureContentByName(fsys fs.FS, name string) (string, error) {
	dirContent, lsErr := fs.ReadDir(fsys, path.Join(lecturesFolderName, name))

	if lsErr != nil {
		return "", lsErr
	}

	var lectureContent = ""

	for _, dirItem := range dirContent {
		if !dirItem.IsDir() && dirItem.Name() == lectureNotesFileName {
			bytes, err := fs.ReadFile(fsys, path.Join(lecturesFolderName, name, lectureNotesFileName))
			if err != nil {
				return "", err
			}

			lectureContent = string(bytes)
		}
	}

	return lectureContent, nil
}

func readAvailableLectures(fsys fs.FS) ([]string, error) {
	lectureDirs, err := fs.ReadDir(fsys, lecturesFolderName)

	if err != nil {
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
