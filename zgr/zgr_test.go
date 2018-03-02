package zgr

import (
	"bufio"
	"strings"
	"testing"

	"github.com/thanm/grvutils/testutils"
)

func split(s string) []string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)
	var res []string
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	return res
}

func makeg() *Graph {
	g := NewGraph()
	g.MakeNode("1", "a")
	g.MakeNode("2", "b")
	g.MakeNode("3", "c")
	g.AddEdge("1", "2")
	g.AddEdge("2", "3")
	g.AddEdge("3", "1")
	g.AddEdge("3", "2")
	return g
}

func TestBasic(t *testing.T) {
	g := makeg()
	exp := `N0: a E: { 1 }
		N1: b E: { 2 }
		N2: c E: { 0 1 }`
	dump := g.String()
	td := testutils.Check(dump, exp)
	if td != "" {
		t.Errorf(td)
	}
}

func TestTranspose(t *testing.T) {
	g := makeg()
	tg := g.Transpose()
	exp := `N0: a E: { 2 }
		N1: b E: { 0 2 }
		N2: c E: { 1 }`
	dump := tg.String()
	td := testutils.Check(dump, exp)
	if td != "" {
		t.Errorf(td)
	}
}
