package expr_test

import (
	"testing"

	"github.com/KlemensWinter/go-binio/expr"
	"github.com/stretchr/testify/assert"
)

func TestNewConst(t *testing.T) {
	testdata := []struct {
		V   any
		Res any
	}{
		{uint8(123), int64(123)},
		{uint16(123), int64(123)},
		{uint32(123), int64(123)},
		{uint64(123), int64(123)},
		{int8(-123), int64(-123)},
		{int16(-123), int64(-123)},
		{int32(-123), int64(-123)},
		{int64(-123), int64(-123)},
	}

	for _, tst := range testdata {
		have := expr.NewConst(tst.V)
		assert.Equal(t, tst.Res, have.Value)
	}

}

func TestParse(t *testing.T) {
	testdata := []struct {
		In        string
		WantError bool
	}{
		{"1", false},
		{"-1", false},

		{"1!=2", false},

		{"!false", false},
		{"!!false", false},

		{"1 > 0.0", false},
		{"sizeof(int16)", false},
		{"dynarray(int16,Foo)", false},
		{"1 > 0 && $foo != 12", false},
		{"1 > 0 && %foo != 12", false},

		{"1 >>> 0", true},
	}
	for _, tst := range testdata {
		e, err := expr.Parse(tst.In)
		if tst.WantError {
			assert.Error(t, err, "expr: %s", tst.In)
		} else {
			assert.NoError(t, err, "expr: %s", tst.In)
		}
		_ = e
	}
}
func TestExpr_String(t *testing.T) {
	testdata := []struct {
		In  string
		Out string
	}{
		{"1", "1"},
		{"1!=2", "(1 != 2)"},
		{"1 > 0.0", "(1 > 0)"},
		{"sizeof(int16      )", "sizeof(int16)"},
		{"dynarray(int16      ,Foo)", "dynarray(int16,Foo)"},

		// unary expr
		{"!true", "(!true)"},
		{"!!true", "(!(!true))"},
		{"!!!true", "(!(!(!true)))"},

		// binary expr
		{"1 > 0   && $foo != 12", "((1 > 0) && ($foo != 12))"},
		{"1   > 0 &&    %foo != 12", "((1 > 0) && (%foo != 12))"},
	}
	for _, tst := range testdata {
		e, err := expr.Parse(tst.In)
		if assert.NoError(t, err, "expr: %q", tst.In) {
			assert.Equal(t, tst.Out, e.String())
		}
	}
}

/*
func TestExprMarshal(t *testing.T) {
	testdata := []struct {
		In        string
		WantError bool
	}{
		{"1", false},

		{"1!=2", false},

		{"!false", false},
		{"!!false", false},

		{"1 > 0.0", false},
		{"sizeof(int16)", false},
		{"dynarray(int16,Foo)", false},
		{"1 > 0 && $foo != 12", false},
		{"1 > 0 && %foo != 12", false},

		{"1 >>> 0", true},
	}

	for _, tst := range testdata {
		ex, err := expr.Parse(tst.In)
		if assert.NoError(t, err) {
			t.Logf("EXPR: %s", ex)
		}
	}
}
*/
