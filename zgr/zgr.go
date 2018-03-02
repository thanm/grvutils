package zgr

import (
	"errors"
	"fmt"
	"strings"
)

type Node struct {
	id, label string
	adjlist   []int
	idx       int
}

type Graph struct {
	nodes []Node
	ntab  map[string]int
}

func NewGraph() *Graph {
	return &Graph{ntab: make(map[string]int)}
}

func (g *Graph) MakeNode(nid, nlabel string) error {
	if _, ok := g.ntab[nid]; ok {
		return errors.New(fmt.Sprintf("MakeNode: collision on node id %s", nid))
	}
	n := Node{id: nid, label: nlabel, idx: len(g.nodes)}
	g.ntab[nid] = n.idx
	g.nodes = append(g.nodes, n)
	return nil
}

func (g *Graph) AddEdge(src, sink string) error {
	var srcid, sinkid int
	var ok bool
	if srcid, ok = g.ntab[src]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown src %s", src))
	}
	if sinkid, ok = g.ntab[sink]; !ok {
		return errors.New(fmt.Sprintf("AddEdge: unknown sink %s", sink))
	}
	// NB: no check for duplicate edge
	g.nodes[srcid].adjlist = append(g.nodes[srcid].adjlist, sinkid)
	return nil
}

func (n *Node) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("N%d: %s E: {", n.idx, n.label))
	for _, e := range n.adjlist {
		sb.WriteString(fmt.Sprintf(" %d", e))
	}
	sb.WriteString(" }")
	return sb.String()
}

func (g *Graph) String() string {
	var sb strings.Builder
	for _, n := range g.nodes {
		sb.WriteString(n.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

func (g *Graph) Transpose() *Graph {
	tg := &Graph{ntab: make(map[string]int)}
	for _, n := range g.nodes {
		tn := n
		tn.adjlist = []int{}
		tg.nodes = append(tg.nodes, tn)
	}
	for i, n := range g.nodes {
		for _, k := range n.adjlist {
			tg.nodes[k].adjlist = append(tg.nodes[k].adjlist, i)
		}
	}
	return tg
}
