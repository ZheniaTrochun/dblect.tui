package main

import (
	"testing"
	"testing/fstest"
)

func TestReadAllLecturesIfTheyExist(t *testing.T) {
	testfs := fstest.MapFS{
		"lectures/01 - intro/lecture_notes.md": {Data: []byte("# Intro")},
		"lectures/02 - joins/lecture_notes.md": {Data: []byte("# Joins")},
		"lectures/02 - joins/presentation.pdf": {Data: []byte("pdf")},
		"lectures/README.md":                   {Data: []byte("not a lecture dir")},
	}

	names, contents, err := readAllLectures(testfs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 lectures, got %d", len(names))
	}

	if len(contents) != 2 {
		t.Fatalf("expected 2 contents, got %d", len(contents))
	}

	firstName := names[0]
	secondName := names[1]

	if firstName != "01 - intro" {
		t.Fatalf("expected first lecture to be '01 - intro', got %s", firstName)
	}

	if secondName != "02 - joins" {
		t.Fatalf("expected second lecture to be '02 -joins', got %s", secondName)
	}

	first := contents[firstName]
	second := contents[secondName]

	if first != "# Intro" {
		t.Fatalf("expected first lecture to be '# Intro', got %s", first)
	}

	if second != "# Joins" {
		t.Fatalf("expected second lecture to be '# Joins', got %s", second)
	}
}

func TestReadAllLecturesIfThereIsNone(t *testing.T) {
	testfs := fstest.MapFS{
		"lectures/README.md": {Data: []byte("not a lecture dir")},
	}

	names, contents, err := readAllLectures(testfs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 0 {
		t.Fatalf("expected 0 lectures, got %d", len(names))
	}

	if len(contents) != 0 {
		t.Fatalf("expected 0 lectures, got %d", len(contents))
	}
}

func TestReadAllLecturesIfSomeHasNoNotes(t *testing.T) {
	testfs := fstest.MapFS{
		"lectures/01 - intro/lecture_notes.md": {Data: []byte("# Intro")},
		"lectures/02 - joins/presentation.pdf": {Data: []byte("pdf")},
	}

	names, contents, err := readAllLectures(testfs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 1 {
		t.Fatalf("expected 1 lectures, got %d", len(names))
	}

	if len(contents) != 1 {
		t.Fatalf("expected 1 contents, got %d", len(contents))
	}

	firstName := names[0]

	if firstName != "01 - intro" {
		t.Fatalf("expected first lecture to be '01 - intro', got %s", firstName)
	}

	first := contents[firstName]

	if first != "# Intro" {
		t.Fatalf("expected first lecture to be '# Intro', got %s", first)
	}
}
