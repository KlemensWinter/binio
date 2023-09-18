package expr

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type (
	Expr interface {
		fmt.Stringer
	}

	BinExpr struct {
		Op  Token
		Lhs Expr
		Rhs Expr
	}

	UnaryExpr struct {
		Op Token
		X  Expr
	}

	Ident struct {
		Name string
	}

	Var struct {
		Name string
	}

	Field struct {
		Name string
	}

	Const struct {
		Value any
	}

	/*
		Float struct {
			Value float64
		}
	*/

	Call struct {
		Name string
		Args []Expr
	}

	parser struct {
		s Scanner

		tok       Token
		tokenText string
	}
)

func NewConst(val any) *Const {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Const{Value: int64(v.Int())}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Const{Value: int64(v.Uint())}
	default:
		panic(fmt.Errorf("invalid const type: %s", v.Kind()))
	}
}

func (expr *BinExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", expr.Lhs, opText[expr.Op], expr.Rhs)
}

func (expr *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", opText[expr.Op], expr.X)
}

func (expr *Const) String() string {
	return fmt.Sprint(expr.Value)
}

func (expr *Var) String() string {
	return fmt.Sprintf("$%s", expr.Name)
}

func (expr *Ident) String() string {
	return expr.Name
}

func (expr *Call) String() string {
	var args []string
	for _, arg := range expr.Args {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", expr.Name, strings.Join(args, ","))
}

func (expr *Field) String() string {
	return fmt.Sprintf("%%%s", expr.Name)
}

func (p *parser) init(rd io.Reader) {
	p.s.Init(rd)
}

func (p *parser) next() (tok Token) {
	tok, err := p.s.Scan()
	p.tokenText = p.s.TokenText()
	if err != nil {
		panic(err)
	}
	p.tok = tok
	return
}

func (p *parser) accept(t Token) bool {
	if p.tok == t {
		p.next()
		return true
	}
	return false
}

func (p *parser) acceptAny(t ...Token) bool {
	for _, tok := range t {
		if p.accept(tok) {
			return true
		}
	}
	return false
}

func (p *parser) expect(t Token) bool {
	if p.accept(t) {
		return true
	}
	panic(fmt.Errorf("unexpected symbol %s", p.tok))
}

func (p *parser) atom() (expr Expr) {
	txt := p.tokenText
	switch {
	case p.accept(IDENT):
		switch txt {
		case "false":
			expr = &Const{Value: false}
		case "true":
			expr = &Const{Value: true}
		case "nil":
			expr = &Const{Value: nil}
		default:
			expr = &Ident{Name: txt}
		}
	case p.accept(REM): // %FIELD
		txt := p.tokenText
		p.expect(IDENT)
		expr = &Field{Name: txt}
	case p.accept(DOL): // $VAR
		txt := p.tokenText
		p.expect(IDENT)
		expr = &Var{Name: txt}
	case p.accept(INT):
		v, err := strconv.ParseInt(txt, 0, 64)
		if err != nil {
			panic(err)
		}
		expr = &Const{Value: v}
	case p.accept(FLOAT):
		v, err := strconv.ParseFloat(txt, 64)
		if err != nil {
			panic(err)
		}
		expr = &Const{Value: v}
	default:
		panic(fmt.Errorf("atom(): invalid token here: %s", p.tok))
	}
	return
}

func (p *parser) unary() Expr {
	switch {
	case p.accept(SUB):
		return &UnaryExpr{
			Op: SUB,
			X:  p.unary(),
		}
	case p.accept(NOT):
		return &UnaryExpr{
			Op: NOT,
			X:  p.unary(),
		}
	}
	return p.atom()
}

func (p *parser) pred1() Expr {
	expr := p.unary()
	if p.accept(LPAREN) { // (
		cexpr := &Call{
			Name: expr.(*Ident).Name,
		}
		expr = cexpr
		if p.accept(RPAREN) {
			// no arguments
		} else {
			e := p.expr()
			cexpr.Args = append(cexpr.Args, e)
			for p.accept(COMMA) {
				cexpr.Args = append(cexpr.Args, p.expr())
			}
			p.expect(RPAREN) // )
		}
	}
	return expr
}

func (p *parser) comp() Expr {
	expr := p.pred1()
	tok := p.tok
	for p.acceptAny(EQL, NEQ, GTR, LSS, GEQ, LEQ) {
		rhs := p.pred1()
		expr = &BinExpr{
			Op:  tok,
			Lhs: expr,
			Rhs: rhs,
		}
		tok = p.tok
	}
	return expr
}

func (p *parser) expr() Expr {
	expr := p.comp()
	tok := p.tok
	for p.acceptAny(LAND, LOR) {
		rhs := p.comp()
		expr = &BinExpr{
			Op:  tok,
			Lhs: expr,
			Rhs: rhs,
		}
		tok = p.tok
	}
	return expr
}

func (p *parser) parse() (expr Expr, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	p.next()
	expr = p.expr()
	p.expect(EOF)
	return
}

func Parse(str string) (expr Expr, err error) {
	var p parser
	p.init(strings.NewReader(str))
	expr, err = p.parse()
	return
}
