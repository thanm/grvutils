package grparser

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

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

func split(s string) []string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)
	var res []string
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	return res
}

func check(raw string, expected string, t *testing.T) string {
	cooked := doparse(raw, t)
	vcooked := split(cooked)
	vexpected := split(expected)
	reason := ""
	if len(vcooked) != len(vexpected) {
		reason = fmt.Sprintf("lengths differ (have %d want %d)",
			len(vcooked), len(vexpected))
	} else {
		for i := 0; i < len(vcooked); i += 1 {
			if vcooked[i] != vexpected[i] {
				reason = fmt.Sprintf("diff at slot %d (have %s want %s)",
					i, vcooked[i], vexpected[i])
			}
		}
	}
	if reason == "" {
		return ""
	}
	return fmt.Sprintf("%s\nraw=%s decoded='%s' wanted '%s':", reason,
		raw, cooked, expected)
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
           "0x1" [label=" one "]
           "0x2" [label=" two "]
           "0x3" [label=" three "]
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
		`N0: "A" E: { 1 }
 	  	 N1: "B" E: { }`,
		`N0: " one " E: { 1 2 }
 		 N1: " two " E: { 0 2 }
		 N2: " three " E: { 1 0 }`,
	}
	for pos, ins := range inputs {
		td := check(ins, expected[pos], t)
		if td != "" {
			t.Errorf(td)
		}
	}
}
