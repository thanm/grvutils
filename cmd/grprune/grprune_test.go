package main

import (
	"testing"
)

func dotest(inf string, t *testing.T) {
	// Run
}

func TestBasic(t *testing.T) {
	var inputs = []string{"g1.graphvis"}
	for _, inf := range inputs {
		dotest(inf, t)
	}
}
