package grparser

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/thanm/grvutils/grlex"
	"github.com/thanm/grvutils/zgr"
)

func mkerror(lxr *grlex.Lexer, s string) error {
	ers := fmt.Sprintf("error: line %d: %s", lxr.CurLine(), s)
	return errors.New(ers)
}

func requiredId(lxr *grlex.Lexer, name string) error {
	if err := requiredToken(lxr, grlex.IDENTIFIER); err != nil {
		return err
	}
	if lxr.Cur.Str != name {
		s := fmt.Sprintf("expected token '%s', got identifier '%s'",
			name, lxr.Cur.Str)
		return mkerror(lxr, s)
	}
	return nil
}

func requiredToken(lxr *grlex.Lexer, tok int) error {
	if err := lxr.GetToken(); err != nil {
		ts := grlex.TokenToString(tok)
		s := fmt.Sprintf("expected %s token, got error: %v", ts, err)
		return mkerror(lxr, s)
	}
	if lxr.Cur.Tok != tok {
		ets := grlex.TokenToString(tok)
		gts := grlex.TokenToString(lxr.Cur.Tok)
		s := fmt.Sprintf("expected %s token, got token '%s'", ets, gts)
		return mkerror(lxr, s)
	}
	return nil
}

func tokenClassToStr(class map[int]bool) string {
	var sb strings.Builder
	first := true
	for k, _ := range class {
		if !first {
			sb.WriteString(",")
		}
		sb.WriteString(grlex.TokenToString(k))
		first = false
	}
	return sb.String()
}

func requiredTokenClass(lxr *grlex.Lexer, class map[int]bool) error {
	if err := lxr.GetToken(); err != nil {
		ts := grlex.TokenToString(lxr.Cur.Tok)
		s := fmt.Sprintf("expected token '%s', got error: %v", ts, err)
		return mkerror(lxr, s)
	}
	if _, ok := class[lxr.Cur.Tok]; !ok {
		ets := tokenClassToStr(class)
		gts := grlex.TokenToString(lxr.Cur.Tok)
		s := fmt.Sprintf("expected token [%s], got token '%s'", ets, gts)
		return mkerror(lxr, s)
	}
	return nil
}

var attrValClass map[int]bool = map[int]bool{
	grlex.IDENTIFIER: true,
	grlex.STRING:     true,
	grlex.CONST:      true,
}

func parseAttrList(lxr *grlex.Lexer, attrs map[string]string) error {
	var err error
	var tok grlex.Token

	if err = requiredToken(lxr, grlex.LBRACKET); err != nil {
		return err
	}

	if tok, err = lxr.PeekToken(); err != nil {
		s := fmt.Sprintf("error %v", err)
		return mkerror(lxr, s)
	}
	needContents := true
	var key, val string
	for {
		if tok.Tok == grlex.RBRACKET {
			if needContents {
				return mkerror(lxr, "parsing attr list: expected ID")
			}
			break
		}
		if tok.Tok != grlex.IDENTIFIER {
			s := fmt.Sprintf("error parsing attr list: expected ID or bracket, got %v", tok.Str)
			return errors.New(s)
		}
		key = tok.Str

		// Consume X=Y
		if err := requiredToken(lxr, grlex.IDENTIFIER); err != nil {
			return err
		}
		if err := requiredToken(lxr, grlex.EQUAL); err != nil {
			return err
		}
		if err := requiredTokenClass(lxr, attrValClass); err != nil {
			return err
		}
		val = lxr.Cur.Str
		if attrs != nil {
			attrs[key] = val
		}
		needContents = false

		// Continue if ","
		if tok, err = lxr.PeekToken(); err != nil {
			return err
		}
		if tok.Tok == grlex.COMMA {
			if err = requiredToken(lxr, grlex.COMMA); err != nil {
				return err
			}
			needContents = true
		}
	}
	if err := requiredToken(lxr, grlex.RBRACKET); err != nil {
		return err
	}
	return nil
}

func parseAttribute(lxr *grlex.Lexer) error {
	if err := requiredToken(lxr, grlex.IDENTIFIER); err != nil {
		return err
	}
	if err := requiredToken(lxr, grlex.EQUAL); err != nil {
		return err
	}
	if err := requiredTokenClass(lxr, attrValClass); err != nil {
		return err
	}
	return nil
}

func parseNode(lxr *grlex.Lexer) error {
	if err := requiredId(lxr, "node"); err != nil {
		return err
	}
	if err := parseAttrList(lxr, nil); err != nil {
		return err
	}
	return nil
}

func parseEdge(lxr *grlex.Lexer) error {
	if err := requiredId(lxr, "edge"); err != nil {
		return err
	}
	if err := parseAttrList(lxr, nil); err != nil {
		return err
	}
	return nil
}

func parseNodeDef(lxr *grlex.Lexer, g *zgr.Graph, id string) error {

	// ID has already been parsed at this point

	attrs := make(map[string]string)
	if err := parseAttrList(lxr, attrs); err != nil {
		return err
	}

	var label string
	if v, ok := attrs["label"]; ok {
		label = v
	}
	if err := g.MakeNode(id, label); err != nil {
		return err
	}
	return nil
}

func parseEdgeDef(lxr *grlex.Lexer, g *zgr.Graph, src string) error {

	// ID has already been parsed; now parse "-> dest"
	if err := requiredToken(lxr, grlex.EDGEOPD); err != nil {
		return err
	}
	if err := requiredToken(lxr, grlex.STRING); err != nil {
		return err
	}
	sink := lxr.Cur.Str
	if err := parseAttrList(lxr, nil); err != nil {
		return err
	}

	if err := g.AddEdge(src, sink); err != nil {
		return err
	}
	return nil
}

func ParseGraph(r io.Reader, g *zgr.Graph) error {
	lxr := grlex.NewLexer(r)

	// Preamble: graph <name> {
	if err := requiredId(lxr, "digraph"); err != nil {
		return err
	}
	if err := requiredToken(lxr, grlex.IDENTIFIER); err != nil {
		return err
	}
	if err := requiredToken(lxr, grlex.LCURLY); err != nil {
		return err
	}

	// Parse a series of node/edge clauses
	done := false
	for !done {
		var tok grlex.Token
		var err error

		// Look at first token. Should be either identifier or string.
		if tok, err = lxr.PeekToken(); err != nil {
			return err
		}

		//fmt.Fprintf(os.Stderr, "++ peeked: %s '%s'\n",
		//grlex.TokenToString(tok.Tok), tok.Str)

		switch tok.Tok {
		case grlex.RCURLY:
			done = true
			break
		case grlex.IDENTIFIER:
			if tok.Str == "node" {
				if err = parseNode(lxr); err != nil {
					return err
				}
				continue
			}
			if tok.Str == "edge" {
				if err = parseEdge(lxr); err != nil {
					return err
				}
				continue
			}

			// Graph attribute
			if err = parseAttribute(lxr); err != nil {
				return err
			}
			continue
		case grlex.STRING:
			// Consume string
			if err = requiredToken(lxr, grlex.STRING); err != nil {
				return err
			}
			src := lxr.Cur.Str

			// Edge or node definition; look at next token
			if tok, err = lxr.PeekToken(); err != nil {
				return err
			}

			// node def: "foo" [attrlist]
			if tok.Tok == grlex.LBRACKET {
				if err = parseNodeDef(lxr, g, src); err != nil {
					return err
				}
				continue
			}

			// edge def: "foo" -> ...
			if tok.Tok == grlex.EDGEOPD {
				if err := parseEdgeDef(lxr, g, src); err != nil {
					return err
				}
				continue
			}
		default:
			// unknown token
			ts := grlex.TokenToString(tok.Tok)
			s := fmt.Sprintf("unexpected token: '%s'", ts)
			return mkerror(lxr, s)
		}
	}

	if err := requiredToken(lxr, grlex.RCURLY); err != nil {
		return err
	}

	return nil
}
