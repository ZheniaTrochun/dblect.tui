package main

import (
	"testing"
)

func TestFooterWidth(t *testing.T) {

	viewName := "Лекції"

	testLineLengths(t,
		func(width int) string {
			return renderFooter(viewName, width)
		})
}
