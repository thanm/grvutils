package grlex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Token struct {
	Str string
	Tok int
}

// Tokens
const (
	IDENTIFIER = iota
	STRING
	CONST
	EQUAL
	COMMA
	LBRACKET
	RBRACKET
	LCURLY
	RCURLY
	EDGEOPD
	EDGEOPU
	EOF
)

var ctab map[int]string = map[int]string{
	STRING:     "str",
	IDENTIFIER: "id",
	CONST:      "const",
	EQUAL:      "=",
	COMMA:      ",",
	LBRACKET:   "[",
	RBRACKET:   "]",
	LCURLY:     "{",
	RCURLY:     "}",
	EDGEOPD:    "->",
	EDGEOPU:    "--",
	EOF:        "<eof>",
}

func TokenToString(t int) string {
	v, ok := ctab[t]
	if ok {
		return v
	}
	return "<unknown>"
}

func isAlpha(b, bp byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isAlphaNumeric(b, bp byte) bool {
	return isAlpha(b, bp) || b >= '0' && b <= '9'
}

func isNumericConst(b, bp byte) bool {
	return b == '.' || b >= '0' && b <= '9'
}

func isStringConst(b, bp byte) bool {
	return b != '"' || bp == '\\'
}

type Lexer struct {
	rdr    bufio.Reader
	Cur    Token
	lno    uint32
	peekt  Token
	peeke  error
	peeked bool
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{rdr: *bufio.NewReader(r), lno: 1}
}

func (lxr *Lexer) consume1(sb *strings.Builder, eb byte) error {
	b, err := lxr.rdr.ReadByte()
	if err != nil {
		return err
	}
	if b != eb {
		s := fmt.Sprintf("error at line %d: expected char '%c' got '%c'",
			lxr.lno, eb, b)
		return errors.New(s)
	}
	sb.WriteByte(b)
	return nil
}

func (lxr *Lexer) readqual(sb *strings.Builder, f func(b, bp byte) bool) error {
	empty := true
	var bp byte = 0
	for {
		bsl, err := lxr.rdr.Peek(1)
		if err != nil {
			if err == io.EOF && !empty {
				return nil
			}
			return err
		}
		b := bsl[0]
		if f(b, bp) == false {
			return nil
		}
		bp = b
		if err = lxr.consume1(sb, b); err != nil {
			return err
		}
		empty = false
	}
	return nil
}

func (lxr *Lexer) genTok(s string, t int) error {
	for i := 0; i < len(s); i += 1 {
		_, err := lxr.rdr.ReadByte()
		if err != nil {
			return err
		}
	}
	lxr.Cur.Str = s
	lxr.Cur.Tok = t
	return nil
}

func (lxr *Lexer) CurLine() uint32 {
	return lxr.lno
}

func (lxr *Lexer) PeekToken() (Token, error) {
	if lxr.peeked {
		panic("only single token lookahead supported")
	}
	save := lxr.Cur
	lxr.peeke = lxr.GetToken()
	lxr.peekt = lxr.Cur
	lxr.peeked = true
	lxr.Cur = save
	return lxr.peekt, lxr.peeke
}

func (lxr *Lexer) GetToken() error {
	e := lxr.GetToken2()
	return e
}

func (lxr *Lexer) GetToken2() error {
	if lxr.peeked {
		lxr.peeked = false
		lxr.Cur = lxr.peekt
		return lxr.peeke
	}
	var sb strings.Builder
	for {
		bsl, err := lxr.rdr.Peek(1)
		if err != nil {
			if err == io.EOF {
				lxr.Cur.Tok = EOF
				return nil
			}
			return err
		}
		b := bsl[0]

		switch {
		case b == ' ' || b == '\t':
			lxr.rdr.ReadByte()
			continue
		case b == '\n':
			lxr.lno += 1
			lxr.rdr.ReadByte()
			continue
		case isAlpha(b, 0):
			// Identifier
			err = lxr.readqual(&sb, isAlphaNumeric)
			if err != nil {
				return err
			}
			lxr.Cur.Str = sb.String()
			lxr.Cur.Tok = IDENTIFIER
			return nil
		case b >= '0' && b <= '9':
			// Numeric constant
			// Identifier
			err = lxr.readqual(&sb, isNumericConst)
			if err != nil {
				return err
			}
			lxr.Cur.Str = sb.String()
			lxr.Cur.Tok = CONST
			return nil
		case b == '"':
			// quoted string
			if err = lxr.consume1(&sb, '"'); err != nil {
				return err
			}
			err = lxr.readqual(&sb, isStringConst)
			if err != nil {
				return err
			}
			if err = lxr.consume1(&sb, '"'); err != nil {
				return err
			}
			lxr.Cur.Str = sb.String()
			lxr.Cur.Tok = STRING
			return nil
		case b == '=':
			return lxr.genTok("=", EQUAL)
		case b == ',':
			return lxr.genTok(",", COMMA)
		case b == '{':
			return lxr.genTok("{", LCURLY)
		case b == '}':
			return lxr.genTok("}", RCURLY)
		case b == '[':
			return lxr.genTok("[", LBRACKET)
		case b == ']':
			return lxr.genTok("]", RBRACKET)
		case b == '-':
			bsl, err := lxr.rdr.Peek(2)
			if err != nil {
				return err
			}
			b := bsl[1]
			if b == '>' {
				return lxr.genTok("->", EDGEOPD)
			} else if b == '-' {
				return lxr.genTok("->", EDGEOPU)
			} else {
				s := fmt.Sprintf("error at line %d: unknown char: '%c'",
					lxr.lno, b)
				return errors.New(s)
			}
		default:
			s := fmt.Sprintf("error at line %d: unknown char: '%c'",
				lxr.lno, b)
			return errors.New(s)
		}
	}
}
