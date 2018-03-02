package grparser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/thanm/grvutils/testutils"
	"github.com/thanm/grvutils/zgr"
)

const (
	TOOMANY = 10000
)

func doparse(ins string, t *testing.T) string {
	g := zgr.NewGraph()
	sr := strings.NewReader(ins)
	if err := ParseGraph(sr, g); err != nil {
		return fmt.Sprintf("%v", err)
	}
	return g.String()
}

func TestBasic(t *testing.T) {
	var inputs = []string{

		`digraph X { }`,

		`digraph Y {
           prop=something
           node [foo=bar]
           edge [size=2]
           "a" [label="A"]
           "b" [label="B"]
           "a" -> "b" [shape=box]
         }`,

		`digraph Z {
           prop=something
           node [foo=bar]
           edge [size=2]
           "0x1" [label="one"]
           "0x2" [label="two"]
           "0x3" [label="three"]
           "0x2" -> "0x1" [q=r]
           "0x2" -> "0x3" [q=r]
           "0x1" -> "0x2" [q=r]
           "0x1" -> "0x3" [q=r]
           "0x3" -> "0x2" [q=r]
           "0x3" -> "0x1" [q=r]
         }`,
	}
	var expected = []string{
		"",
		`N0: 'A' E: { 1 }
 	  	 N1: 'B' E: { }`,
		`N0: 'one' E: { 1 2 }
 		 N1: 'two' E: { 0 2 }
		 N2: 'three' E: { 1 0 }`,
	}
	for pos, ins := range inputs {
		actual := doparse(ins, t)
		td := testutils.Check(actual, expected[pos])
		if td != "" {
			t.Errorf(td)
		}
	}
}
