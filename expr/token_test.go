package expr_test

import (
	"strings"
	"testing"

	e "github.com/KlemensWinter/go-binio/expr"
	"github.com/stretchr/testify/assert"
)

func TestScanner(t *testing.T) {
	testdata := []struct {
		In  string
		Out []e.Token
	}{
		{"", []e.Token{e.EOF}},

		{"true", []e.Token{e.IDENT, e.EOF}},
		{"false", []e.Token{e.IDENT, e.EOF}},
		{"nil", []e.Token{e.IDENT, e.EOF}},

		{"foo", []e.Token{e.IDENT, e.EOF}},
		{"123", []e.Token{e.INT, e.EOF}},
		{"123.5", []e.Token{e.FLOAT, e.EOF}},

		{"1<2.340", []e.Token{e.INT, e.LSS, e.FLOAT, e.EOF}},
		{"1>2.340", []e.Token{e.INT, e.GTR, e.FLOAT, e.EOF}},
		{"1==2.340", []e.Token{e.INT, e.EQL, e.FLOAT, e.EOF}},
		{"1!=2.340", []e.Token{e.INT, e.NEQ, e.FLOAT, e.EOF}},
		{"1<=2.340", []e.Token{e.INT, e.LEQ, e.FLOAT, e.EOF}},
		{"1>=2.340", []e.Token{e.INT, e.GEQ, e.FLOAT, e.EOF}},
		{"true &&false", []e.Token{e.IDENT, e.LAND, e.IDENT, e.EOF}},
		{"false  || true", []e.Token{e.IDENT, e.LOR, e.IDENT, e.EOF}},
	}

	for _, tst := range testdata {
		var s e.Scanner
		s.Init(strings.NewReader(tst.In))
		for i, want := range tst.Out {
			have, err := s.Scan()
			if assert.NoError(t, err) {
				assert.Equal(t, want, have, "expr: %q token %d: want: %s, have: %s", tst.In, i, want, have)
			}
		}
		tok, err := s.Scan()
		if assert.NoError(t, err) {
			assert.Equal(t, e.EOF, tok)
		}
	}
}
