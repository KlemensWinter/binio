package binio_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/KlemensWinter/binio"
	"github.com/stretchr/testify/assert"
)

func TestWriteVarString(t *testing.T) {
	var buf bytes.Buffer
	str := "HelloWorld"
	err := binio.WriteVarString[uint16](&buf, str)
	assert.NoError(t, err)
	assert.Equal(t, len(str)+2, buf.Len())
	n := int(binary.LittleEndian.Uint16(buf.Bytes()[:2]))
	assert.Equal(t, len(str), n)
	assert.Equal(t, str, string(buf.Bytes()[2:]))
}

func TestDecoder_string(t *testing.T) {
	tst := struct {
		Foo string
	}{}

	err := binio.Unmarshal(NullReader, &tst)
	assert.Error(t, err)
}
