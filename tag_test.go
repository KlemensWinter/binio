package binio_test

import (
	"testing"

	"github.com/KlemensWinter/go-binio"
	"github.com/KlemensWinter/go-binio/expr"
	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {

	testdata := []struct {
		In  string
		Tag binio.Tag
	}{
		{"size=12", binio.Tag{Size: &expr.Const{Value: int64(12)}}},
	}

	for _, tst := range testdata {
		tag, err := binio.ParseTag(tst.In)
		if err != nil {
			t.Error(err)
			continue
		}
		// t.Logf("TAG: %#v", tag)
		assert.NotNil(t, tag)
	}

}

func TestParseTag_vars(t *testing.T) {

	testdata := []struct {
		In  string
		Tag binio.Tag

		VarNames []string
	}{
		{"$foo=12", binio.Tag{Size: expr.NewConst(12)}, nil},
		{"size=12,$foo=333", binio.Tag{Size: expr.NewConst(12)}, []string{"foo"}},
	}

	for _, tst := range testdata {
		tag, err := binio.ParseTag(tst.In)
		if err != nil {
			t.Error(err)
			continue
		}
		// t.Logf("TAG: %#v", tag)
		if tst.VarNames != nil {
			assert.EqualValues(t, tst.VarNames, tag.VarNames())
		}
	}
}

/*
func Test_parseFieldTag(t *testing.T) {

	testdata := []struct {
		In string
	}{
		{"type=dynarray"},
		{"if=1"},
		{"if=1 > 10"},
		{"if=$fooo"},
	}

	for _, tst := range testdata {
		var f field
		parseFieldTag(&f, tst.In)
		t.Logf("RES: %#v", f)
	}
}
*/
