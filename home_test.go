package main

import (
	"testing"
)

func TestHomeScreeWidth(t *testing.T) {

	testLineLengths(t,
		func(width int) string {
			m := homeModel{width: width, height: 50}
			return m.View().Content
		})
}
