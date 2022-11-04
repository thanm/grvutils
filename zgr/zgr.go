package zgr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Node struct {
	id, label  string
	outadjlist []uint32
	inadjlist  []uint32
	attrs      []uint32
	idx        uint32
}

type npair struct {
	src, sink uint32
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
	etab    map[npair]int
	attrs   []Attr
	attrtab map[Attr]uint32
}

func NewGraph() *Graph {
	return &Graph{
		ntab:    make(map[string]uint32),
		etab:    make(map[npair]int),
		attrtab: make(map[Attr]uint32),
	}
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

func (g *Graph) AddEdge(src, sink string, attrs map[string]string) error {
	var srcid, sinkid uint32
	var ok bool
	if srcid, ok = g.ntab[src]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown src %s", src))
	}
	if sinkid, ok = g.ntab[sink]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown sink %s", sink))
	}
	cand := npair{src: srcid, sink: sinkid}
	if _, ok := g.etab[cand]; ok {
		return errors.New(fmt.Sprintf("duplicate edge %q -> %q", src, sink))
	}
	res := g.populateAttrs(attrs)
	e := Edge{src: srcid, sink: sinkid, attrs: res}
	eidx := uint32(len(g.edges))
	g.etab[cand] = int(eidx)
	g.edges = append(g.edges, e)
	g.nodes[srcid].outadjlist = append(g.nodes[srcid].outadjlist, eidx)
	g.nodes[sinkid].inadjlist = append(g.nodes[sinkid].inadjlist, eidx)
	return nil
}

func (g *Graph) SetEdgeAttrs(src, sink string, attrs map[string]string) error {
	var srcid, sinkid uint32
	var ok bool
	if srcid, ok = g.ntab[src]; !ok {
		return errors.New(fmt.Sprintf("SetEdgeAttrs: unknown src %s", src))
	}
	if sinkid, ok = g.ntab[sink]; !ok {
		return errors.New(fmt.Sprintf("SetEdgeAttrs: unknown sink %s", sink))
	}
	cand := npair{src: srcid, sink: sinkid}
	v, ok := g.etab[cand]
	if !ok {
		return errors.New(fmt.Sprintf("can't locate edge %q -> %q", src, sink))
	}
	g.edges[v].attrs = g.populateAttrs(attrs)
	return nil
}

func (n *Node) Label() string {
	return n.label
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) Idx() uint32 {
	return n.idx
}

func (n *Node) String(g *Graph) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("N%d: '%s' E: {", n.idx, n.label))
	for _, e := range n.outadjlist {
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
	tg.attrtab = g.attrtab
	tg.ntab = g.ntab
	for _, n := range g.nodes {
		tn := n
		tn.inadjlist = []uint32{}
		tn.outadjlist = []uint32{}
		tg.nodes = append(tg.nodes, tn)
	}
	for i, e := range g.edges {
		tg.nodes[e.sink].outadjlist = append(tg.nodes[e.sink].outadjlist, uint32(i))
		tg.nodes[e.src].inadjlist = append(tg.nodes[e.src].inadjlist, uint32(i))
		var te Edge = e
		te.src = e.sink
		te.sink = e.src
		tg.edges = append(tg.edges, te)
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
	return n.outadjlist
}

func (g *Graph) GetInEdges(n *Node) []uint32 {
	return n.inadjlist
}

func (g *Graph) GetEdge(eidx uint32) *Edge {
	if eidx < uint32(len(g.edges)) {
		return &g.edges[eidx]
	}
	return nil
}

func (g *Graph) GetEdgeAttrs(e *Edge) map[string]string {
	res := make(map[string]string)
	for _, at := range e.attrs {
		a := g.attrs[at]
		res[a.key] = a.val
	}
	return res
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

func (g *Graph) writeAttrs(bw *bufio.Writer, attrs []uint32, addcom bool) {
	if len(attrs) == 0 {
		return
	}
	bw.WriteString(" [")
	first := true
	for _, idx := range attrs {
		if !first {
			if addcom {
				bw.WriteString(",")
			}
			bw.WriteString(" ")
		}
		first = false
		a := g.attrs[idx]
		bw.WriteString(fmt.Sprintf("%s=%s", a.key, a.val))
	}
	bw.WriteString("]")
}

func (g *Graph) Write(w io.Writer, toinclude map[uint32]bool) error {
	bw := bufio.NewWriter(w)
	bw.WriteString("digraph G {\n")

	emit := func(x uint32) bool {
		if toinclude == nil {
			return true
		}
		return toinclude[x]
	}

	// Nodes
	for nid, n := range g.nodes {
		if !emit(uint32(nid)) {
			continue
		}
		bw.WriteString(fmt.Sprintf("%s ", n.id))
		g.writeAttrs(bw, n.attrs, true)
		bw.WriteString("\n")
	}

	// Edges
	for nid, n := range g.nodes {
		if !emit(uint32(nid)) {
			continue
		}
		for _, eid := range n.outadjlist {
			e := g.edges[eid]
			if !emit(uint32(e.sink)) {
				continue
			}
			bw.WriteString(fmt.Sprintf("%s -> %s",
				g.nodes[e.src].id, g.nodes[e.sink].id))
			g.writeAttrs(bw, e.attrs, false)
			bw.WriteString("\n")

		}
	}

	bw.WriteString("}\n")
	bw.Flush()
	return nil
}
