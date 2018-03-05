package grlex

import (
	"fmt"
	"strings"
	"testing"
)

const (
	TOOMANY = 10000
)

func dolex(lxr *Lexer) string {
	var sb strings.Builder
	for i := 0; i < TOOMANY; i += 1 {
		if err := lxr.GetToken(); err != nil {
			sb.WriteString(fmt.Sprintf("error %v at token %d", err, i))
			return sb.String()
		}
		if lxr.Cur.Tok == EOF {
			return sb.String()
		}
		ts := TokenToString(lxr.Cur.Tok)
		sb.WriteString(fmt.Sprintf("(%s '%s')", ts, lxr.Cur.Str))
	}
	return sb.String()
}

func mklexer(ins string) *Lexer {
	sr := strings.NewReader(ins)
	lxr := NewLexer(sr)
	return lxr
}

func tokenize(ins string) string {
	lxr := mklexer(ins)
	return dolex(lxr)
}

func testTok(raw string, expected string) string {
	cooked := tokenize(raw)
	if cooked == expected {
		return ""
	}
	return fmt.Sprintf("raw=%s decoded='%s' wanted '%s'",
		raw, cooked, expected)
}

func TestBasic(t *testing.T) {
	var inputs = []string{
		"",
		"graph foo",
		"101.1",
		"\"foo\"",
		"\"foo \\\"bar\\\" baz\"",
		`digraph n { rankdir="LR"
         node [fontsize=10, shape=box, height=0.25]
         edge [q=r]
         "0x556c43bea3c0" [label="blah"]
         "0x556c42f19ba0" -> "0x556c43bea3c0" [label=" phony"]`,
	}
	var expected = []string{
		"",
		"(id 'graph')(id 'foo')",
		"(const '101.1')",
		`(str '"foo"')`,
		`(str '"foo \"bar\" baz"')`,
		`(id 'digraph')(id 'n')({ '{')(id 'rankdir')(= '=')(str '"LR"')(id 'node')([ '[')(id 'fontsize')(= '=')(const '10')(, ',')(id 'shape')(= '=')(id 'box')(, ',')(id 'height')(= '=')(const '0.25')(] ']')(id 'edge')([ '[')(id 'q')(= '=')(id 'r')(] ']')(str '"0x556c43bea3c0"')([ '[')(id 'label')(= '=')(str '"blah"')(] ']')(str '"0x556c42f19ba0"')(-> '->')(str '"0x556c43bea3c0"')([ '[')(id 'label')(= '=')(str '" phony"')(] ']')`,
	}
	for pos, ins := range inputs {
		td := testTok(ins, expected[pos])
		if td != "" {
			t.Errorf(td)
		}
	}
}

func TestSeek(t *testing.T) {
	var input string = `foo bar baz 1.0 "blix" 22`
	var expected = `(id 'foo')(id 'bar')(id 'baz')(const '1.0')(str '"blix"')(const '22')`
	lxr := mklexer(input)

	// First pass through the string
	res1 := dolex(lxr)

	// Reset to start
	lxr.Reset()

	// Second pass through the string
	res2 := dolex(lxr)

	// Strings should be equal
	if res1 != res1 {
		t.Errorf("expected res1 == res2 got '%s' '%s'", res1, res2)
	}

	// String should be as expected
	if res1 != expected {
		t.Errorf("expected res1 == '%s' got '%s'", expected, res1)
	}
}
