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
	attrs := map[string]string{"prop1": "2", "prop2": "zilch"}
	attrs["label"] = "a"
	g.MakeNode("1", attrs)
	attrs["label"] = "b"
	g.MakeNode("2", attrs)
	attrs["label"] = "c"
	g.MakeNode("3", attrs)
	attrs["label"] = ""
	g.AddEdge("1", "2", attrs)
	g.AddEdge("2", "3", attrs)
	g.AddEdge("3", "1", attrs)
	g.AddEdge("3", "2", attrs)
	return g
}

func TestBasic(t *testing.T) {
	g := makeg()
	exp := `N0: 'a' E: { 1 }
		N1: 'b' E: { 2 }
		N2: 'c' E: { 0 1 }`
	dump := g.String()
	td := testutils.Check(dump, exp)
	if td != "" {
		t.Errorf(td)
	}
}

func TestTranspose(t *testing.T) {
	g := makeg()
	tg := g.Transpose()
	exp := `N0: 'a' E: { 2 }
		N1: 'b' E: { 0 2 }
		N2: 'c' E: { 1 }`
	dump := tg.String()
	td := testutils.Check(dump, exp)
	if td != "" {
		t.Errorf(td)
	}
}

func TestAccess(t *testing.T) {
	g := makeg()
	nn := g.GetNodeCount()
	if nn != 3 {
		t.Errorf("bad GetNodeCount wanted 3 got %d", nn)
	}
	n0 := g.LookupNode("1")
	if n0 == nil {
		t.Errorf("bad LookupNode(1) returned nil")
	}
	es := g.GetEdges(n0)
	if len(es) != 1 {
		t.Errorf("getEdges(n0) len wanted 1 got %d", len(es))
	}
	e0 := g.GetEdge(es[0])
	if e0 == nil {
		t.Errorf("GetEdge(es[0]) returned nil")
	}
	src, sink := g.GetEndpoints(e0)
	if src != 0 || sink != 1 {
		t.Errorf("GetEndpoints returned %d,%d: wanted 0,1", src, sink)
	}
}
