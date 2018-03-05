package grparser

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/thanm/grvutils/grlex"
	"github.com/thanm/grvutils/zgr"
)

type pstate struct {
	lxr    *grlex.Lexer
	tok    grlex.Token
	peeked bool
}

func mkerror(p *pstate, s string) error {
	ers := fmt.Sprintf("error: line %d: %s", p.lxr.CurLine(), s)
	return errors.New(ers)
}

func requiredId(p *pstate, name string) error {
	if err := requiredToken(p, grlex.IDENTIFIER); err != nil {
		return err
	}
	if p.tok.Str != name {
		s := fmt.Sprintf("expected token '%s', got identifier '%s'",
			name, p.tok.Str)
		return mkerror(p, s)
	}
	return nil
}

func requiredToken(p *pstate, tok int) error {
	if err := p.GetToken(); err != nil {
		ts := grlex.TokenToString(tok)
		s := fmt.Sprintf("expected %s token, got error: %v", ts, err)
		return mkerror(p, s)
	}
	if p.tok.Tok != tok {
		ets := grlex.TokenToString(tok)
		gts := grlex.TokenToString(p.tok.Tok)
		s := fmt.Sprintf("expected %s token, got token '%s'", ets, gts)
		return mkerror(p, s)
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

func requiredTokenClass(p *pstate, class map[int]bool) error {
	if err := p.GetToken(); err != nil {
		ts := grlex.TokenToString(p.tok.Tok)
		s := fmt.Sprintf("expected token '%s', got error: %v", ts, err)
		return mkerror(p, s)
	}
	if _, ok := class[p.tok.Tok]; !ok {
		ets := tokenClassToStr(class)
		gts := grlex.TokenToString(p.tok.Tok)
		s := fmt.Sprintf("expected token [%s], got token '%s'", ets, gts)
		return mkerror(p, s)
	}
	return nil
}

var attrValClass map[int]bool = map[int]bool{
	grlex.IDENTIFIER: true,
	grlex.STRING:     true,
	grlex.CONST:      true,
}

// Confusingly, some attribute lists seem to have commas and some do not.

func parseAttrList(p *pstate, attrs map[string]string, comma bool) error {
	var err error

	// Attribute lists are optional
	if err = p.PeekToken(); err != nil {
		return err
	}
	if p.tok.Tok != grlex.LBRACKET {
		return nil
	}

	if err = requiredToken(p, grlex.LBRACKET); err != nil {
		return err
	}

	// What's next?
	if err = p.PeekToken(); err != nil {
		return err
	}
	var key, val string
	for {
		if p.tok.Tok != grlex.IDENTIFIER {
			s := fmt.Sprintf("error parsing attr list: expected ID or bracket, got %v", p.tok.Str)
			return errors.New(s)
		}
		key = p.tok.Str

		// Consume X=Y
		if err := requiredToken(p, grlex.IDENTIFIER); err != nil {
			return err
		}
		if err := requiredToken(p, grlex.EQUAL); err != nil {
			return err
		}
		if err := requiredTokenClass(p, attrValClass); err != nil {
			return err
		}
		val = p.tok.Str
		if attrs != nil {
			attrs[key] = val
		}

		// Take a peek at the next token
		if err = p.PeekToken(); err != nil {
			return err
		}

		// End of attrs?
		if p.tok.Tok == grlex.RBRACKET {
			break
		}

		if comma {
			// Expect comma between attributes
			if p.tok.Tok == grlex.COMMA {
				if err = requiredToken(p, grlex.COMMA); err != nil {
					return err
				}
			}
		}
		if err = p.PeekToken(); err != nil {
			return err
		}
	}
	if err := requiredToken(p, grlex.RBRACKET); err != nil {
		return err
	}
	return nil
}

func parseAttribute(p *pstate) error {
	if err := requiredToken(p, grlex.IDENTIFIER); err != nil {
		return err
	}
	if err := requiredToken(p, grlex.EQUAL); err != nil {
		return err
	}
	if err := requiredTokenClass(p, attrValClass); err != nil {
		return err
	}
	return nil
}

func parseNode(p *pstate) error {
	if err := requiredId(p, "node"); err != nil {
		return err
	}
	if err := parseAttrList(p, nil, true); err != nil {
		return err
	}
	return nil
}

func parseEdge(p *pstate) error {
	if err := requiredId(p, "edge"); err != nil {
		return err
	}
	if err := parseAttrList(p, nil, false); err != nil {
		return err
	}
	return nil
}

func parseNodeDef(p *pstate, g *zgr.Graph, id string, pass int) error {

	// ID has already been parsed at this point

	attrs := make(map[string]string)
	if err := parseAttrList(p, attrs, true); err != nil {
		return err
	}
	if pass == 1 {
		if err := g.MakeNode(id, attrs); err != nil {
			return err
		}
	}
	return nil
}

func parseEdgeDef(p *pstate, g *zgr.Graph, src string, pass int) error {

	// ID has already been parsed; now parse "-> dest"
	if err := requiredToken(p, grlex.EDGEOPD); err != nil {
		return err
	}
	if err := requiredToken(p, grlex.STRING); err != nil {
		return err
	}
	sink := p.tok.Str
	attrs := make(map[string]string)
	if err := parseAttrList(p, attrs, false); err != nil {
		return err
	}
	//	fmt.Fprintf(os.Stderr, "parseEdgeDef: %d attrs\n", len(attrs))

	if pass == 2 {
		if err := g.AddEdge(src, sink, attrs); err != nil {
			return err
		}
	}
	return nil
}

func (p *pstate) PeekToken() error {
	var err error
	if p.peeked {
		return nil
	}
	if p.tok, err = p.lxr.PeekToken(); err != nil {
		s := fmt.Sprintf("error %v", err)
		return mkerror(p, s)
	}
	p.peeked = true
	return nil
}

func (p *pstate) GetToken() error {
	if err := p.lxr.GetToken(); err != nil {
		return err
	}
	p.tok = p.lxr.Cur
	p.peeked = false
	return nil
}

func parse(g *zgr.Graph, lxr *grlex.Lexer, pass int) error {

	var err error
	state := pstate{lxr: lxr}
	p := &state

	// Preamble: graph <name> {
	if err := requiredId(p, "digraph"); err != nil {
		return err
	}
	if err := requiredToken(p, grlex.IDENTIFIER); err != nil {
		return err
	}
	if err := requiredToken(p, grlex.LCURLY); err != nil {
		return err
	}

	// Parse a series of node/edge clauses
	done := false
	for !done {

		// Look at first token. Should be either identifier or string.
		if err = p.PeekToken(); err != nil {
			return err
		}

		switch p.tok.Tok {
		case grlex.RCURLY:
			done = true
			break
		case grlex.IDENTIFIER:
			if p.tok.Str == "node" {
				if err = parseNode(p); err != nil {
					return err
				}
				continue
			}
			if p.tok.Str == "edge" {
				if err = parseEdge(p); err != nil {
					return err
				}
				continue
			}

			// Graph attribute
			if err = parseAttribute(p); err != nil {
				return err
			}
			continue
		case grlex.STRING:
			// Consume string
			if err = requiredToken(p, grlex.STRING); err != nil {
				return err
			}
			src := p.tok.Str

			// Edge or node definition; look at next token
			if err = p.PeekToken(); err != nil {
				return err
			}

			// edge def: "foo" -> ...
			if p.tok.Tok == grlex.EDGEOPD {
				if err := parseEdgeDef(p, g, src, pass); err != nil {
					return err
				}
				continue
			}

			// node def: "foo" [attrlist]
			if err = parseNodeDef(p, g, src, pass); err != nil {
				return err
			}
			continue

		default:
			// unknown token
			ts := grlex.TokenToString(p.tok.Tok)
			s := fmt.Sprintf("unexpected token: '%s'", ts)
			return mkerror(p, s)
		}
	}

	if err := requiredToken(p, grlex.RCURLY); err != nil {
		return err
	}

	return nil
}

func ParseGraph(r io.ReadSeeker, g *zgr.Graph) error {
	lxr := grlex.NewLexer(r)

	// Pass 1: collect nodes
	if err := parse(g, lxr, 1); err != nil {
		return err
	}

	lxr.Reset()

	// Pass 2: connect edges
	if err := parse(g, lxr, 2); err != nil {
		return err
	}

	return nil
}
