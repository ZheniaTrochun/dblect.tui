package main

import (
	"testing"
)

func TestHeaderWidthNoContinuation(t *testing.T) {
	testLineLengths(t,
		func(width int) string {
			return renderHeader(width, false)
		})
}

func TestHeaderWidthWithContinuation(t *testing.T) {
	testLineLengths(t,
		func(width int) string {
			return renderHeader(width, true)
		})
}
