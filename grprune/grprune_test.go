package grprune

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thanm/grvutils/grparser"
	"github.com/thanm/grvutils/testutils"
	"github.com/thanm/grvutils/zgr"
)

const (
	TOOMANY = 10000
)

func doparse(ins string) (*zgr.Graph, error) {
	g := zgr.NewGraph()
	sr := strings.NewReader(ins)
	if err := grparser.ParseGraph(sr, g); err != nil {
		return nil, err
	}
	return g, nil
}

const testgraph = `digraph Y {
   "a" [label="A"]
   "b" [label="B"]
   "c" [label="C"]
   "d" [label="D"]
   "e" [label="E"]
   "f" [label="F"]
   "g" [label="G"]
   "c" -> "b" [x=y]
   "c" -> "d" [z=w]
   "d" -> "e" [q=r]
   "g" -> "c" [q=r]
   "f" -> "c" [b=1]
   "b" -> "a" [b=1]
   "e" -> "f" [b=1]
   "a" -> "g" [b=1]
 }`

type testcase struct {
	mode    string
	exclude string
	depth   int
}

func doTest(g *zgr.Graph, t *testing.T, tmpfile string, tc testcase) string {

	var tf *os.File
	var err error

	// Open temp file file for writing
	tf, err = os.OpenFile(tmpfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Errorf("temp file open for write failed: %v", err)
	}

	// Apply prune
	if err = PruneGraph(g, "c", tc.mode, tc.depth, tc.exclude, tf); err != nil {
		t.Errorf("prune failed: %v", err)
	}

	// Close file
	tf.Close()

	// Re-open for reading
	var rtf *os.File
	rtf, err = os.Open(tmpfile)
	if err != nil {
		t.Errorf("temp file open for read failed: %v", err)
	}

	content, _ := os.ReadFile(tmpfile)
	t.Logf("content:\n%s\n", string(content))

	// Parse, then dump
	pg := zgr.NewGraph()
	if err = grparser.ParseGraph(rtf, pg); err != nil {
		t.Errorf("parsing pruned graph: %v", err)
	}

	return pg.String()
}

func TestBasic(t *testing.T) {
	var inputs = []testcase{
		testcase{"both", "", 0},
		testcase{"fwd", "", 1},
		testcase{"bwd", "", 1},
		testcase{"both", "", 1},
		testcase{"fwd", "", 2},
		testcase{"bwd", "e", 2},
	}
	var expected = []string{

		`N0: '"C"' E: { }`,

		`N0: '"B"' E: { }
	 	 N1: '"C"' E: { 0 2 }
		 N2: '"D"' E: { }`,

		`N0: '"C"' E: { }
 		 N1: '"F"' E: { 0 }
		 N2: '"G"' E: { 0 }`,

		`N0: '"B"' E: { }
		 N1: '"C"' E: { 0 2 }
		 N2: '"D"' E: { }
		 N3: '"F"' E: { 1 }
		 N4: '"G"' E: { 1 }`,

		`N0: '"A"' E: { }
		 N1: '"B"' E: { 0 }
		 N2: '"C"' E: { 1 3 }
		 N3: '"D"' E: { 4 }
		 N4: '"E"' E: { }`,

		`N0: '"A"' E: { 3 }
		 N1: '"C"' E: { }
		 N2: '"F"' E: { 1 }
		 N3: '"G"' E: { 1 }`,
	}
	graph, err := doparse(testgraph)
	if err != nil {
		t.Fatalf("parsing initial graph: %v", err)
	}
	dir, err := ioutil.TempDir("", "prunedir")
	if err != nil {
		t.Errorf("creating TempDir: %v", err)
	}
	defer os.RemoveAll(dir) // clean up
	tmpfn := filepath.Join(dir, "tmpfile")

	tmpfn = "/tmp/tmpfile.txt"

	for pos, tc := range inputs {
		actual := doTest(graph, t, tmpfn, tc)
		td := testutils.Check(actual, expected[pos])
		if td != "" {
			t.Errorf(td)
		}
	}
}
