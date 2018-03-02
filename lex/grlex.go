package grlex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Token struct {
	s string
	t int
}

// Tokens
const (
	NAME = iota
	STRING
	IDENTIFIER
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
	NAME:       "name",
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
	//	fmt.Fprintf(os.Stderr, "b=%d(%c) bp=%d(%c) rv=%v\n",
	//		b, b, bp, bp, b != '"' || bp == '\\')
	return b != '"' || bp == '\\'
}

type Lexer struct {
	rdr bufio.Reader
	Cur Token
	lno uint32
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{*bufio.NewReader(r), Token{"", 0}, 0}
}

func (lxr *Lexer) consume1(sb *strings.Builder, eb byte) error {
	b, err := lxr.rdr.ReadByte()
	if err != nil {
		return err
	}
	if b != eb {
		return errors.New("lex error")
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
	lxr.Cur.s = s
	lxr.Cur.t = t
	return nil
}

func (lxr *Lexer) GetToken() error {
	var sb strings.Builder
	for {
		bsl, err := lxr.rdr.Peek(1)
		if err != nil {
			if err == io.EOF {
				lxr.Cur.t = EOF
				return nil
			}
			return err
		}
		b := bsl[0]

		//fmt.Fprintf(os.Stderr, "loop: b is %v '%c'\n", b, b)

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
			lxr.Cur.s = sb.String()
			lxr.Cur.t = IDENTIFIER
			return nil
		case b >= '0' && b <= '9':
			// Numeric constant
			// Identifier
			err = lxr.readqual(&sb, isNumericConst)
			if err != nil {
				return err
			}
			lxr.Cur.s = sb.String()
			lxr.Cur.t = CONST
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
			lxr.Cur.s = sb.String()
			lxr.Cur.t = STRING
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
				fmt.Fprintf(os.Stderr, "unknown char: %v\n", b)
				return errors.New("lex error")
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown char: %v\n", b)
			return errors.New("lex error")
		}
	}
}
