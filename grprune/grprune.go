package grprune

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/thanm/grvutils/zgr"
)

func walk(g *zgr.Graph, node *zgr.Node, depth int, dcutoff int, inc map[uint32]bool) {
	// Mark this node for inclusion
	inc[g.GetNodeIndex(node)] = true

	// Stop here if we've reached the cutoff
	if depth >= dcutoff {
		return
	}

	// Visit out-edge targets
	outs := g.GetEdges(node)
	for _, eid := range outs {
		e := g.GetEdge(eid)
		_, sinkid := g.GetEndpoints(e)
		sink := g.GetNode(sinkid)
		walk(g, sink, depth+1, dcutoff, inc)
	}
}

func getPrunedSet(g *zgr.Graph, rootid string, mode string, depth int) (error, map[uint32]bool) {

	// Locate the node with the specified ID.
	rn := g.LookupNode(rootid)
	if rn == nil {
		s := fmt.Sprintf("error: unable to locate root node '%s'", rootid)
		return errors.New(s), nil
	}

	// This map is going to hold the indices of the nodes we want to
	// include in the final slice.
	include := make(map[uint32]bool)

	// Forward walk from root
	if mode == "both" || mode == "fwd" {
		walk(g, rn, 0, depth, include)
	}

	// Backwards walk from root
	if mode == "both" || mode == "bwd" {
		tg := g.Transpose()
		walk(tg, rn, 0, depth, include)
	}

	// Debugging: dump include set
	fmt.Fprintf(os.Stderr, "include set:\n")
	for k, _ := range include {
		fmt.Fprintf(os.Stderr, " %d", k)
	}
	fmt.Fprintf(os.Stderr, "\n")

	return nil, include
}

func PruneGraph(g *zgr.Graph, rootid string, mode string, depth int, w io.Writer) error {
	// Collect IDs of nodes to write
	err, include := getPrunedSet(g, rootid, mode, depth)
	if err != nil {
		return err
	}

	// Do the write
	if err := g.Write(w, include); err != nil {
		return err
	}

	return nil
}
