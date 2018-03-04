package zgr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Node struct {
	id, label string
	adjlist   []uint32
	attrs     []uint32
	idx       uint32
}

type Edge struct {
	src, sink uint32
	attrs     []uint32
}

type Attr struct {
	key, val string
}

type Graph struct {
	nodes   []Node
	edges   []Edge
	ntab    map[string]uint32
	attrs   []Attr
	attrtab map[Attr]uint32
}

func NewGraph() *Graph {
	return &Graph{ntab: make(map[string]uint32), attrtab: make(map[Attr]uint32)}
}

func (g *Graph) populateAttrs(attrs map[string]string) []uint32 {
	res := []uint32{}
	for k, v := range attrs {
		a := Attr{key: k, val: v}
		var idx uint32
		var ok bool
		if idx, ok = g.attrtab[a]; !ok {
			idx = uint32(len(g.attrs))
			g.attrs = append(g.attrs, a)
		}
		res = append(res, idx)
	}
	return res
}

func (g *Graph) MakeNode(nid string, attrs map[string]string) error {
	if _, ok := g.ntab[nid]; ok {
		return errors.New(fmt.Sprintf("MakeNode: collision on node id %s", nid))
	}
	res := g.populateAttrs(attrs)
	nlabel := ""
	if lab, ok := attrs["label"]; ok {
		nlabel = lab
	}
	n := Node{id: nid, label: nlabel, idx: uint32(len(g.nodes)), attrs: res}
	g.ntab[nid] = uint32(n.idx)
	g.nodes = append(g.nodes, n)
	return nil
}

// NB: no check for duplicate edge

func (g *Graph) AddEdge(src, sink string, attrs map[string]string) error {
	var srcid, sinkid uint32
	var ok bool
	if srcid, ok = g.ntab[src]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown src %s", src))
	}
	if sinkid, ok = g.ntab[sink]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown sink %s", sink))
	}
	res := g.populateAttrs(attrs)
	e := Edge{src: srcid, sink: sinkid, attrs: res}
	eidx := uint32(len(g.edges))
	g.edges = append(g.edges, e)
	g.nodes[srcid].adjlist = append(g.nodes[srcid].adjlist, eidx)
	return nil
}

func (n *Node) String(g *Graph) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("N%d: '%s' E: {", n.idx, n.label))
	for _, e := range n.adjlist {
		sink := g.edges[e].sink
		sb.WriteString(fmt.Sprintf(" %d", sink))
	}
	sb.WriteString(" }")
	return sb.String()
}

func (g *Graph) String() string {
	var sb strings.Builder
	for _, n := range g.nodes {
		sb.WriteString(n.String(g))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (g *Graph) Transpose() *Graph {
	tg := NewGraph()
	tg.attrs = g.attrs
	tg.attrtab = tg.attrtab
	for _, n := range g.nodes {
		tn := n
		tn.adjlist = []uint32{}
		tg.nodes = append(tg.nodes, tn)
	}
	for i, e := range g.edges {
		tg.nodes[e.sink].adjlist = append(tg.nodes[e.sink].adjlist, uint32(i))
		tmp := e.src
		e.src = e.sink
		e.sink = tmp
		tg.edges = append(tg.edges, e)
	}
	return tg
}

func (g *Graph) GetNode(idx uint32) *Node {
	if idx < uint32(len(g.nodes)) {
		return &g.nodes[idx]
	}
	return nil
}

func (g *Graph) GetEdges(n *Node) []uint32 {
	return n.adjlist
}

func (g *Graph) GetEdge(eidx uint32) *Edge {
	if eidx < uint32(len(g.edges)) {
		return &g.edges[eidx]
	}
	return nil
}

func (g *Graph) GetEndpoints(e *Edge) (uint32, uint32) {
	return e.src, e.sink
}

func (g *Graph) GetNodeIndex(n *Node) uint32 {
	return n.idx
}

func (g *Graph) LookupNode(nid string) *Node {
	if idx, ok := g.ntab[nid]; ok {
		return &g.nodes[idx]
	}
	return nil
}

func (g *Graph) GetNodeCount() uint32 {
	return uint32(len(g.nodes))
}

func (g *Graph) Write(w io.Writer, toinclude map[uint32]bool) error {
	bw := bufio.NewWriter(w)
	bw.Write([]byte("digraph G {\n"))
	panic("not yet impl")
	bw.Write([]byte("}\n"))
	return nil
}
