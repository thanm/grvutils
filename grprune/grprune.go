package grprune

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/thanm/grvutils/zgr"
)

func walk(g *zgr.Graph, node *zgr.Node, depth int, dcutoff int, inc map[uint32]bool, excl map[uint32]bool) {

	// Bail now if on exclude list
	if _, ok := excl[g.GetNodeIndex(node)]; ok {
		return
	}

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
		walk(g, sink, depth+1, dcutoff, inc, excl)
	}
}

func makeExcludeSet(g *zgr.Graph, toex string, excl map[uint32]bool) error {
	ids := strings.Split(toex, ",")
	for _, id := range ids {
		if id == "" {
			continue
		}
		quid := fmt.Sprintf("\"%s\"", id)
		en := g.LookupNode(quid)
		if en == nil {
			s := fmt.Sprintf("error: unable to locate exclude node '%s'", id)
			return errors.New(s)
		}
		excl[g.GetNodeIndex(en)] = true
	}
	return nil
}

func getPrunedSet(g *zgr.Graph, rootid string, mode string, depth int, toex string) (error, map[uint32]bool) {

	// Locate the root node with the specified ID.
	quid := fmt.Sprintf("\"%s\"", rootid)
	rn := g.LookupNode(quid)
	if rn == nil {
		s := fmt.Sprintf("error: unable to locate root node '%s'", rootid)
		return errors.New(s), nil
	}

	// This map is going to hold the indices of the nodes we want to
	// include in the final slice.
	include := make(map[uint32]bool)

	// This map is going to hold the indices of the nodes we want to
	// exclude from the final slice.
	exclude := make(map[uint32]bool)
	if err := makeExcludeSet(g, toex, exclude); err != nil {
		return err, nil
	}

	// Forward walk from root
	if mode == "both" || mode == "fwd" {
		walk(g, rn, 0, depth, include, exclude)
	}

	// Backwards walk from root
	if mode == "both" || mode == "bwd" {
		tg := g.Transpose()
		quid := fmt.Sprintf("\"%s\"", rootid)
		rn := tg.LookupNode(quid)
		walk(tg, rn, 0, depth, include, exclude)
	}

	return nil, include
}

func PruneGraph(g *zgr.Graph, rootid string, mode string, depth int, exclude string, w io.Writer) error {
	// Collect IDs of nodes to write
	err, include := getPrunedSet(g, rootid, mode, depth, exclude)
	if err != nil {
		return err
	}

	// Do the write
	if err := g.Write(w, include); err != nil {
		return err
	}

	return nil
}
