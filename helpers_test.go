package main

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
)

var testScreenWidths = []int{32, 40, 64, 80, 100, 128, 256}

func plain(s string) string {
	return ansi.Strip(s)
}

func testLineLengths(t *testing.T, buildRows func(int) string) {
	for _, width := range testScreenWidths {
		content := buildRows(width)

		plainContent := plain(content)
		for i, line := range strings.Split(plainContent, "\n") {
			w := lipgloss.Width(line)
			if w > width {
				t.Errorf("line %d: width %d > expected width %d", i, w, width)
			}
		}
	}
}
