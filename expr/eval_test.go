package expr_test

import (
	"io"
	"math"
	"os"
	"testing"

	"github.com/KlemensWinter/binio/expr"
	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	testdata := []struct {
		In   any
		Want bool
	}{
		// bool
		{true, true},
		{false, false},

		{byte(1), true},
		{byte(0), false},

		{rune(11), true},
		{rune(0), false},

		// string
		{"", false},
		{"HelloWorld", true},

		// int
		{int(1), true},
		{int(0), false},
		{int(-10), false},

		{int8(1), true},
		{int8(0), false},
		{int8(-10), false},
		{int16(1), true},
		{int16(0), false},
		{int16(-10), false},
		{int32(1), true},
		{int32(0), false},
		{int32(-10), false},
		{int64(1), true},
		{int64(0), false},
		{int64(-10), false},

		{uint(1), true},
		{uint(0), false},
		{uint8(1), true},
		{uint8(0), false},
		{uint16(1), true},
		{uint16(0), false},
		{uint32(1), true},
		{uint32(0), false},
		{uint64(1), true},
		{uint64(0), false},

		// float
		{float32(1), true},
		{float32(0), false},
		{float32(-10.9), false},
		{float32(math.NaN()), false},
		{float32(math.Inf(1)), true},
		{float32(math.Inf(-1)), false},
		{float64(1), true},
		{float64(0), false},
		{float64(-10.9), false},
		{float64(math.NaN()), false},
		{float64(math.Inf(1)), true},
		{float64(math.Inf(-1)), false},

		{uintptr(12345), true},
		{uintptr(0), false},

		// interface
		{io.Reader(nil), false},
		{io.Reader((io.Reader)(os.Stdout)), true},

		// slice
		{[]byte{}, false},
		{[]byte{1}, true},
	}

	for _, tst := range testdata {
		have := expr.Bool(tst.In)
		assert.Equal(t, tst.Want, have, "input: %T(%#v)", tst.In, tst.In)
	}

	// test errors
	_ = expr.Bool(complex(10, 10))
}

func TestCompare(t *testing.T) {
	testdata := []struct {
		Tok expr.Token
		Lhs any
		Rhs any

		Want bool
	}{
		{expr.LSS, 10, 10, false},
		{expr.LSS, 10, 5, false},
		{expr.LSS, -5, 10.54, true},

		{expr.GTR, 10, 10, false}, // 10 > 10
		{expr.GTR, 11, 10, true},  // 11 > 10
		{expr.GTR, 10.34, 11, false},
		{expr.GTR, 11.34, 11, true},

		{expr.EQL, 64, 64, true},
		{expr.EQL, 64.5, 64, false},
		{expr.EQL, 64, int16(64), true},

		{expr.NEQ, 0, 0, false},
		{expr.NEQ, 0, true, true},
		{expr.NEQ, false, false, false},
		{expr.NEQ, true, false, true},

		{expr.LEQ, 0, 0, true},

		{expr.GEQ, 0, 0, true},
	}

	for _, tst := range testdata {
		have, err := expr.Compare(tst.Tok, tst.Lhs, tst.Rhs)
		if assert.NoError(t, err) {
			assert.Equal(t, tst.Want, have)
		}
	}
}

func TestEval(t *testing.T) {
	testdata := []struct {
		In string

		Want any
	}{
		{"1", int64(1)},
		{"1.234", float64(1.234)},
		{"true", true},
		{"false", false},
		{"nil", nil},

		// unary expr
		{"!true", false},
		{"!!true", true},
		{"!!!true", false},

		{"!1", false},

		{"-1", int64(-1)},
		{"-1234", int64(-1234)},
		{"-123.456", float64(-123.456)},

		// binexpr

		{"1<2", true},
		{"2>1", true},
		{"1>=2", false},
		{"1!=2", true},
		{"1!=1", false},
		{"1==1", true},

		{"true && true", true},
		{"false && false", false},
		{"true && false", false},
		{"false && true", false},

		{"true || true", true},
		{"false || false", false},
		{"true || false", true},
		{"false || true", true},
	}

	ctx := &expr.Context{}

	for _, tst := range testdata {
		ex, err := expr.Parse(tst.In)
		if assert.NoError(t, err, "parse error, input=%q", tst.In) {
			have, err := expr.Eval(ctx, ex)
			if assert.NoError(t, err, "expr: %q", ex) {
				assert.Equal(t, tst.Want, have, "expr: %q", tst.In)
			}
		}
	}
}
