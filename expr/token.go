//go:generate stringer -type Token

package expr

import (
	"fmt"
	"io"
	"text/scanner"
)

const (
	INVALID Token = iota

	IDENT

	INT
	FLOAT
	FIELD

	ADD // +
	SUB // -
	MUL // *
	QUO // /

	REM // %
	DOL // $

	LSS // <
	GTR // >

	EQL // ==
	NEQ // !=
	LEQ // <=
	GEQ // >=

	LAND // &&
	LOR  // ||

	AND // &
	OR  // |

	NOT // !

	LPAREN // (
	//LBRACK // [
	//LBRACE // {
	COMMA // ,

	RPAREN // )
	//	RBRACK // ]
	//	RBRACE // }

	EOF
)

type (
	Token int
)

var opText = map[Token]string{
	// unary
	NOT: "!",
	// binary
	LSS:  "<",
	GTR:  ">",
	EQL:  "==",
	NEQ:  "!=",
	LEQ:  "<=",
	GEQ:  ">=",
	LAND: "&&",
	LOR:  "||",
	// AND:  "&",
	// OR:   "|",
}

type Scanner struct {
	scanner.Scanner
}

func (s *Scanner) Init(rd io.Reader) {
	s.Scanner.Init(rd)
	s.Scanner.Whitespace = scanner.GoWhitespace
	s.Scanner.Mode = scanner.GoTokens
}

func (s *Scanner) Scan() (tok Token, err error) {
	c := s.Scanner.Scan()
	if c == scanner.EOF {
		tok = EOF
		return
	}

	acceptRune := func(r rune) bool {
		if s.Peek() == r {
			s.Next()
			return true
		}
		return false
	}

	switch c {
	case scanner.Ident:
		tok = IDENT
	case scanner.Float:
		tok = FLOAT
	case scanner.Int:
		tok = INT
	case ',':
		tok = COMMA
	case '(':
		tok = LPAREN
	case ')':
		tok = RPAREN
	case '$':
		tok = DOL
	case '-':
		tok = SUB
	case '%':
		tok = REM
	case '=':
		c := s.Next()
		if c != '=' {
			panic("need = here!") // TODO: improve
		}
		tok = EQL
	case '>':
		tok = GTR
		if acceptRune('=') {
			tok = GEQ
		}
	case '<':
		tok = LSS
		if acceptRune('=') {
			tok = LEQ
		}
	case '&':
		tok = AND
		if acceptRune('&') {
			tok = LAND
		}
	case '|':
		tok = OR
		if acceptRune('|') {
			tok = LOR
		}
	case '!':
		tok = NOT
		if acceptRune('=') {
			tok = NEQ
		}
	default:
		return INVALID, fmt.Errorf("invalid token %s", scanner.TokenString(c))
	}
	return
}
