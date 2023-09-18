package binio_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"testing"

	"github.com/KlemensWinter/binio"
	"github.com/stretchr/testify/assert"
)

var (
	le = binary.LittleEndian
)

func pack(args ...any) []byte {
	var buf bytes.Buffer
	for i, itm := range args {
		err := binary.Write(&buf, binary.LittleEndian, itm)
		if err != nil {
			panic(fmt.Errorf("failed to encode arg #%d: %w", i, err))
		}
	}
	return buf.Bytes()
}

type nullReader struct{}

func (*nullReader) Read(p []byte) (n int, err error) {
	n = len(p)
	for i := 0; i < n; i++ { // will be optimized by the compiler... I hope :)
		p[i] = 0
	}
	return
}

// NullReader writes only 0 to the buffer
var NullReader io.Reader = &nullReader{}

func TestIntSize(t *testing.T) {
	testdata := []struct {
		Name string
		Size int
	}{
		{"uint8", 1},
		{"uint16", 2},
		{"uint32", 4},
		{"uint64", 8},
		{"int8", 1},
		{"int16", 2},
		{"int32", 4},
		{"int64", 8},

		{"i8", -1},
		{"uint80", -1},
	}

	for _, tst := range testdata {
		assert.Equal(t, tst.Size, binio.IntSize(tst.Name), "name: %q", tst.Name)
	}
}
